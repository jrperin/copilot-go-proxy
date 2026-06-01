package translator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TranslateToAnthropic converts an OpenAI response to Anthropic format
func TranslateToAnthropic(resp OpenAIResponse) AnthropicResponse {
	anthropic := AnthropicResponse{
		ID:    resp.ID,
		Type:  "message",
		Role:  "assistant",
		Model: resp.Model,
		Usage: translateUsage(resp.Usage),
	}

	if len(resp.Choices) == 0 {
		stopReason := "end_turn"
		anthropic.StopReason = &stopReason
		return anthropic
	}

	choice := resp.Choices[0]

	// Stop reason
	if choice.FinishReason != nil {
		sr := mapStopReason(*choice.FinishReason)
		anthropic.StopReason = &sr
	}

	// Content
	if choice.Message != nil {
		anthropic.Content = translateMessageContent(choice.Message)
	}

	return anthropic
}

func translateMessageContent(msg *OpenAIMessage) []ContentBlock {
	var blocks []ContentBlock

	// Text content
	if msg.Content != nil {
		text := extractTextContent(msg.Content)
		if text != "" {
			blocks = append(blocks, ContentBlock{
				Type: "text",
				Text: text,
			})
		}
	}

	// Tool calls
	for _, tc := range msg.ToolCalls {
		input := map[string]interface{}{}
		if tc.Function.Arguments != "" {
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
		}
		blocks = append(blocks, ContentBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: input,
		})
	}

	if len(blocks) == 0 {
		blocks = append(blocks, ContentBlock{Type: "text", Text: ""})
	}

	return blocks
}

func extractTextContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, part := range v {
			if m, ok := part.(map[string]interface{}); ok {
				if m["type"] == "text" {
					if text, ok := m["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func mapStopReason(openAIReason string) string {
	switch openAIReason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	case "content_filter":
		return "end_turn"
	default:
		return "end_turn"
	}
}

func translateUsage(usage OpenAIUsage) AnthropicUsage {
	cached := 0
	if usage.PromptTokensDetails != nil {
		cached = usage.PromptTokensDetails.CachedTokens
	}
	return AnthropicUsage{
		InputTokens:          usage.PromptTokens - cached,
		OutputTokens:         usage.CompletionTokens,
		CacheReadInputTokens: cached,
	}
}

// TranslateErrorToAnthropic creates an Anthropic-format error response
func TranslateErrorToAnthropic(statusCode int, message string) AnthropicError {
	var errType string
	switch statusCode {
	case 400:
		errType = "invalid_request_error"
	case 401:
		errType = "authentication_error"
	case 403:
		errType = "permission_error"
	case 404:
		errType = "not_found_error"
	case 429:
		errType = "rate_limit_error"
	case 500, 502, 503:
		errType = "api_error"
	default:
		errType = "api_error"
	}

	err := AnthropicError{Type: "error"}
	err.Error.Type = errType
	err.Error.Message = message
	return err
}

// GenerateMessageID creates a unique message ID
func GenerateMessageID() string {
	return fmt.Sprintf("msg_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:24])
}
