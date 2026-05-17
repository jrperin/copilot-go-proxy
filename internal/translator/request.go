package translator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TranslateToOpenAI converts an Anthropic request to OpenAI format
func TranslateToOpenAI(req AnthropicRequest) OpenAIRequest {
	openAI := OpenAIRequest{
		Model:       collapseModelVersion(req.Model),
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	if req.Metadata != nil && req.Metadata.UserID != "" {
		openAI.User = req.Metadata.UserID
	}

	if len(req.StopSequences) > 0 {
		openAI.Stop = req.StopSequences
	}

	// System → first system message
	if req.System != nil {
		systemText := extractSystemText(req.System)
		if systemText != "" {
			openAI.Messages = append(openAI.Messages, OpenAIMessage{
				Role:    "system",
				Content: systemText,
			})
		}
	}

	// Translate messages
	for _, msg := range req.Messages {
		openAI.Messages = append(openAI.Messages, translateMessage(msg)...)
	}

	// Tools
	for _, tool := range req.Tools {
		openAI.Tools = append(openAI.Tools, OpenAITool{
			Type: "function",
			Function: OpenAIFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	// Tool choice
	if req.ToolChoice != nil {
		openAI.ToolChoice = translateToolChoice(req.ToolChoice)
	}

	return openAI
}

func extractSystemText(system interface{}) string {
	switch v := system.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, block := range v {
			if m, ok := block.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n\n")
	}
	return ""
}

func translateMessage(msg AnthropicMessage) []OpenAIMessage {
	var result []OpenAIMessage

	switch content := msg.Content.(type) {
	case string:
		if content != "" {
			result = append(result, OpenAIMessage{
				Role:    msg.Role,
				Content: content,
			})
		}
	case []interface{}:
		var toolResults []OpenAIMessage
		var textParts []string
		var toolCalls []OpenAIToolCall
		var imageParts []map[string]interface{}

		for _, block := range content {
			b, _ := json.Marshal(block)
			var cb ContentBlock
			json.Unmarshal(b, &cb)

			switch cb.Type {
			case "text":
				if cb.Text != "" {
					textParts = append(textParts, cb.Text)
				}

			case "tool_use":
				args := "{}"
				if cb.Input != nil {
					if a, err := json.Marshal(cb.Input); err == nil {
						args = string(a)
					}
				}
				toolCalls = append(toolCalls, OpenAIToolCall{
					ID:   cb.ID,
					Type: "function",
					Function: OpenAIFunctionCall{
						Name:      cb.Name,
						Arguments: args,
					},
				})

			case "tool_result":
				resultContent := ""
				if cb.Content != nil {
					switch v := cb.Content.(type) {
					case string:
						resultContent = v
					case []interface{}:
						var parts []string
						for _, part := range v {
							if m, ok := part.(map[string]interface{}); ok {
								if text, ok := m["text"].(string); ok {
									parts = append(parts, text)
								}
							}
						}
						resultContent = strings.Join(parts, "\n")
					}
				}
				toolResults = append(toolResults, OpenAIMessage{
					Role:       "tool",
					ToolCallID: cb.ToolUseID,
					Content:    resultContent,
				})

			case "image":
				if cb.Source != nil {
					imageParts = append(imageParts, map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:%s;base64,%s", cb.Source.MediaType, cb.Source.Data),
						},
					})
				}
			}
		}

		// Tool results come first (as separate messages before user content)
		result = append(result, toolResults...)

		// Convert imageParts to []interface{}
		var imagePartsIface []interface{}
		for _, ip := range imageParts {
			imagePartsIface = append(imagePartsIface, ip)
		}

		// Build content parts for the main message
		if msg.Role == "assistant" && len(toolCalls) > 0 {
			assistantMsg := OpenAIMessage{
				Role:      "assistant",
				ToolCalls: toolCalls,
			}
			if len(textParts) > 0 {
				assistantMsg.Content = strings.Join(textParts, "\n")
			}
			result = append(result, assistantMsg)
		} else if len(textParts) > 0 || len(imageParts) > 0 {
			var parts []interface{}
			for _, text := range textParts {
				parts = append(parts, map[string]string{"type": "text", "text": text})
			}
			parts = append(parts, imagePartsIface...)
			if len(parts) == 1 {
				if m, ok := parts[0].(map[string]interface{}); ok {
					if m["type"] == "text" {
						result = append(result, OpenAIMessage{
							Role:    msg.Role,
							Content: m["text"],
						})
						break
					}
				}
			}
			result = append(result, OpenAIMessage{
				Role:    msg.Role,
				Content: parts,
			})
		}
	}

	if len(result) == 0 {
		result = append(result, OpenAIMessage{
			Role:    msg.Role,
			Content: "",
		})
	}

	return result
}

func translateToolChoice(tc interface{}) interface{} {
	switch v := tc.(type) {
	case map[string]interface{}:
		typ, _ := v["type"].(string)
		switch typ {
		case "auto":
			return "auto"
		case "any":
			return "required"
		case "tool":
			if name, ok := v["name"].(string); ok {
				return map[string]interface{}{
					"type": "function",
					"function": map[string]string{"name": name},
				}
			}
		}
	case string:
		return v
	}
	return "auto"
}

// collapseModelVersion strips date suffixes from Anthropic model names
func collapseModelVersion(model string) string {
	for _, prefix := range []string{"claude-sonnet-4-", "claude-opus-4-", "claude-haiku-3.5-"} {
		if strings.HasPrefix(model, prefix) {
			return strings.TrimSuffix(model, model[len(prefix):])
		}
	}
	// Generic: remove -YYYYMMDD suffix
	if len(model) > 9 {
		suffix := model[len(model)-9:]
		if suffix[0] == '-' && isAllDigits(suffix[1:]) {
			return model[:len(model)-9]
		}
	}
	return model
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
