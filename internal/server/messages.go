package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jrperin/copilot-go-proxy/internal/translator"
)

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAnthropicError(w, 400, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var anthropicReq translator.AnthropicRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		writeAnthropicError(w, 400, fmt.Sprintf("invalid JSON: %s", err.Error()))
		return
	}

	openAIReq := translator.TranslateToOpenAI(anthropicReq)
	openAIBody, _ := json.Marshal(openAIReq)

	isStreaming := anthropicReq.Stream

	// Forward to Copilot
	resp, err := s.client.ChatCompletions(string(openAIBody))
	if err != nil {
		log.Printf("ERROR: copilot request failed: %v", err)
		writeAnthropicError(w, 502, fmt.Sprintf("copilot request failed: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: copilot returned %d: %s", resp.StatusCode, string(errBody))
		writeAnthropicError(w, resp.StatusCode, fmt.Sprintf("copilot error: %s", string(errBody)))
		return
	}

	if isStreaming {
		s.handleStreamingResponse(w, resp, openAIReq.Model)
	} else {
		s.handleNonStreamingResponse(w, resp)
	}
}

func (s *Server) handleNonStreamingResponse(w http.ResponseWriter, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeAnthropicError(w, 502, "failed to read copilot response")
		return
	}

	var openAIResp translator.OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		writeAnthropicError(w, 502, "failed to parse copilot response")
		return
	}

	anthropicResp := translator.TranslateToAnthropic(openAIResp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anthropicResp)
}

func (s *Server) handleStreamingResponse(w http.ResponseWriter, resp *http.Response, model string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAnthropicError(w, 500, "streaming not supported")
		return
	}

	state := translator.NewStreamState(model)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		chunk, err := translator.ParseSSELine(line)
		if err != nil {
			continue // skip unparseable lines
		}
		if chunk == nil {
			// [DONE] marker
			break
		}

		events, err := state.TranslateChunk(*chunk)
		if err != nil {
			log.Printf("ERROR: translation error: %v", err)
			continue
		}

		for _, event := range events {
			fmt.Fprint(w, event.Format())
			flusher.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("ERROR: streaming read error: %v", err)
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "api_error",
		},
	})
}

// writeAnthropicError writes an Anthropic-format error
func writeAnthropicError(w http.ResponseWriter, statusCode int, message string) {
	errResp := translator.TranslateErrorToAnthropic(statusCode, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)
}
