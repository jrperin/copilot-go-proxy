package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/jrperin/copilot-go-proxy/internal/translator"
)

// handleChatCompletions proxies OpenAI-format requests directly to Copilot
func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		writeJSONError(w, 400, "failed to read request body")
		return
	}
	defer r.Body.Close()

	// For logging, we need to decode the request body
	var req translator.OpenAIRequest
	_ = json.Unmarshal(body, &req)

	slog.Info("received chat completion request", "model", req.Model, "stream", req.Stream)

	resp, err := s.client.ChatCompletions(string(body))
	if err != nil {
		slog.Error("copilot request failed", "error", err)
		writeJSONError(w, 502, fmt.Sprintf("copilot request failed: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		slog.Error("copilot returned non-200 status", "status_code", resp.StatusCode, "body", string(errBody))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(errBody)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			slog.Error("streaming not supported by client")
			writeJSONError(w, 500, "streaming not supported")
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprint(w, line+"\n")
			flusher.Flush()
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, resp.Body)
	}
}
