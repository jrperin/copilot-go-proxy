package copilot

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	editorVersion    = "vscode/1.113.0"
	pluginVersion    = "copilot-chat/0.26.7"
	userAgent        = "GitHubCopilotChat/0.26.7"
	apiVersion       = "2025-04-01"
	vscodeLibVersion = "electron-fetch"
)

// SetCopilotHeaders sets all required headers for Copilot API requests
func SetCopilotHeaders(req *http.Request, copilotJWT string) {
	req.Header.Set("Authorization", "Bearer "+copilotJWT)
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Editor-Plugin-Version", pluginVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Github-Api-Version", apiVersion)
	req.Header.Set("X-Request-Id", uuid.New().String())
	req.Header.Set("Copilot-Integration-Id", "vscode-chat")
	req.Header.Set("Openai-Intent", "conversation-panel")
	req.Header.Set("X-Vscode-User-Agent-Library-Version", vscodeLibVersion)
}
