package translator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// StreamState tracks stateful translation during SSE streaming
type StreamState struct {
	MessageStartSent  bool
	ContentBlockIndex int
	ContentBlockOpen  bool
	CurrentBlockType  string // "text" or "tool_use"
	ToolCalls         map[int]*ToolCallTracker
	MessageID         string
	Model             string
	InputTokens       int
	OutputTokens      int
	CachedTokens      int
}

type ToolCallTracker struct {
	ID                string
	Name              string
	AnthropicBlockIdx int
	ArgsStarted       bool
}

func NewStreamState(model string) *StreamState {
	return &StreamState{
		ToolCalls: make(map[int]*ToolCallTracker),
		MessageID: GenerateMessageID(),
		Model:     model,
	}
}

// TranslateChunk converts an OpenAI SSE chunk to Anthropic SSE events
func (s *StreamState) TranslateChunk(chunk OpenAIChunk) ([]SSEEvent, error) {
	var events []SSEEvent

	if len(chunk.Choices) == 0 {
		// Might be usage-only chunk at the end
		if chunk.Usage != nil {
			s.InputTokens = chunk.Usage.PromptTokens
			s.OutputTokens = chunk.Usage.CompletionTokens
			if chunk.Usage.PromptTokensDetails != nil {
				s.CachedTokens = chunk.Usage.PromptTokensDetails.CachedTokens
			}
		}
		return events, nil
	}

	choice := chunk.Choices[0]

	// Send message_start if not sent yet
	if !s.MessageStartSent {
		events = append(events, s.makeMessageStart())
		s.MessageStartSent = true
	}

	// Process delta content
	if choice.Delta != nil {
		delta := choice.Delta

		// Text content
		if delta.Content != nil {
			text := extractTextContent(delta.Content)
			if text != "" {
				// Close tool block if open
				if s.ContentBlockOpen && s.CurrentBlockType == "tool_use" {
					events = append(events, s.makeContentBlockStop())
					s.ContentBlockOpen = false
				}
				// Open text block if needed
				if !s.ContentBlockOpen {
					events = append(events, s.makeContentBlockStart("text"))
					s.ContentBlockOpen = true
					s.CurrentBlockType = "text"
				}
				events = append(events, s.makeTextDelta(text))
			}
		}

		// Tool calls
		for _, tc := range delta.ToolCalls {
			idx := tc.Index

			if _, exists := s.ToolCalls[idx]; !exists {
				// Close text block if open
				if s.ContentBlockOpen && s.CurrentBlockType == "text" {
					events = append(events, s.makeContentBlockStop())
					s.ContentBlockOpen = false
				}

				// New tool call
				blockIdx := s.ContentBlockIndex
				s.ToolCalls[idx] = &ToolCallTracker{
					ID:                tc.ID,
					Name:              tc.Function.Name,
					AnthropicBlockIdx: blockIdx,
				}
				s.ContentBlockIndex++

				events = append(events, SSEEvent{
					Event: "content_block_start",
					Data: map[string]interface{}{
						"type":  "content_block_start",
						"index": blockIdx,
						"content_block": map[string]interface{}{
							"type":  "tool_use",
							"id":    tc.ID,
							"name":  tc.Function.Name,
							"input": map[string]interface{}{},
						},
					},
				})
				s.ContentBlockOpen = true
				s.CurrentBlockType = "tool_use"
			}

			// Tool arguments delta
			if tc.Function.Arguments != "" {
				tracker := s.ToolCalls[idx]
				events = append(events, SSEEvent{
					Event: "content_block_delta",
					Data: map[string]interface{}{
						"type":  "content_block_delta",
						"index": tracker.AnthropicBlockIdx,
						"delta": map[string]interface{}{
							"type":         "input_json_delta",
							"partial_json": tc.Function.Arguments,
						},
					},
				})
			}
		}
	}

	// Handle finish
	if choice.FinishReason != nil {
		// Close any open content block
		if s.ContentBlockOpen {
			events = append(events, s.makeContentBlockStop())
			s.ContentBlockOpen = false
		}

		// Update usage from chunk if available
		if chunk.Usage != nil {
			s.InputTokens = chunk.Usage.PromptTokens
			s.OutputTokens = chunk.Usage.CompletionTokens
			if chunk.Usage.PromptTokensDetails != nil {
				s.CachedTokens = chunk.Usage.PromptTokensDetails.CachedTokens
			}
		}

		stopReason := mapStopReason(*choice.FinishReason)
		events = append(events, s.makeMessageDelta(stopReason))
		events = append(events, s.makeMessageStop())
	}

	return events, nil
}

type SSEEvent struct {
	Event string
	Data  interface{}
}

func (e SSEEvent) Format() string {
	b, _ := json.Marshal(e.Data)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", e.Event, string(b))
}

func (s *StreamState) makeMessageStart() SSEEvent {
	return SSEEvent{
		Event: "message_start",
		Data: map[string]interface{}{
			"type": "message_start",
			"message": map[string]interface{}{
				"id":            s.MessageID,
				"type":          "message",
				"role":          "assistant",
				"content":       []interface{}{},
				"model":         s.Model,
				"stop_reason":   nil,
				"stop_sequence": nil,
				"usage": map[string]interface{}{
					"input_tokens":              s.InputTokens,
					"output_tokens":             0,
					"cache_read_input_tokens":   s.CachedTokens,
				},
			},
		},
	}
}

func (s *StreamState) makeContentBlockStart(blockType string) SSEEvent {
	block := map[string]interface{}{
		"type": blockType,
		"text": "",
	}
	return SSEEvent{
		Event: "content_block_start",
		Data: map[string]interface{}{
			"type":          "content_block_start",
			"index":         s.ContentBlockIndex,
			"content_block": block,
		},
	}
}

func (s *StreamState) makeContentBlockStop() SSEEvent {
	idx := s.ContentBlockIndex
	if s.CurrentBlockType == "tool_use" {
		// Find the tracker for current block
		for _, t := range s.ToolCalls {
			if t.AnthropicBlockIdx == s.ContentBlockIndex {
				idx = t.AnthropicBlockIdx
				break
			}
		}
	} else {
		idx = s.ContentBlockIndex
		s.ContentBlockIndex++
	}
	return SSEEvent{
		Event: "content_block_stop",
		Data: map[string]interface{}{
			"type":  "content_block_stop",
			"index": idx,
		},
	}
}

func (s *StreamState) makeTextDelta(text string) SSEEvent {
	return SSEEvent{
		Event: "content_block_delta",
		Data: map[string]interface{}{
			"type":  "content_block_delta",
			"index": s.ContentBlockIndex,
			"delta": map[string]interface{}{
				"type": "text_delta",
				"text": text,
			},
		},
	}
}

func (s *StreamState) makeMessageDelta(stopReason string) SSEEvent {
	return SSEEvent{
		Event: "message_delta",
		Data: map[string]interface{}{
			"type": "message_delta",
			"delta": map[string]interface{}{
				"stop_reason":   stopReason,
				"stop_sequence": nil,
			},
			"usage": map[string]interface{}{
				"input_tokens":              s.InputTokens,
				"output_tokens":             s.OutputTokens,
				"cache_read_input_tokens":   s.CachedTokens,
			},
		},
	}
}

func (s *StreamState) makeMessageStop() SSEEvent {
	return SSEEvent{
		Event: "message_stop",
		Data: map[string]interface{}{
			"type": "message_stop",
		},
	}
}

// ParseSSELine parses a single SSE data line into an OpenAI chunk
func ParseSSELine(line string) (*OpenAIChunk, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data: ") {
		return nil, nil
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "[DONE]" {
		return nil, nil // end of stream
	}

	var chunk OpenAIChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}
