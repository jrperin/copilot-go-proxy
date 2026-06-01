package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jrperin/copilot-go-proxy/internal/config"
)

var (
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"
)

const (
	editorVersion = "vscode/1.113.0"
	pluginVersion = "copilot-chat/0.26.7"
	userAgent     = "GitHubCopilotChat/0.26.7"
	apiVersion    = "2025-04-01"
)

type TokenManager struct {
	mu          sync.Mutex
	githubToken string
	copilotJWT  string
	expiresAt   int64
	stopCh      chan struct{}
}

func NewTokenManager(githubToken string) *TokenManager {
	return &TokenManager{
		githubToken: githubToken,
		stopCh:      make(chan struct{}),
	}
}

func (tm *TokenManager) GetToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.copilotJWT == "" || time.Now().Unix() >= tm.expiresAt-60 {
		slog.Info("copilot token is expired or empty, refreshing...")
		if err := tm.refreshToken(); err != nil {
			slog.Error("failed to refresh copilot token", "error", err)
			return "", fmt.Errorf("refreshing copilot token: %w", err)
		}
		slog.Info("successfully refreshed copilot token")
	}

	return tm.copilotJWT, nil
}

func (tm *TokenManager) refreshToken() error {
	url := copilotTokenURL
	if config.AppConfig != nil && config.AppConfig.CopilotTokenURL != "" {
		url = config.AppConfig.CopilotTokenURL
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+tm.githubToken)
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Editor-Plugin-Version", pluginVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Github-Api-Version", apiVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("copilot token request failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
		RefreshIn int    `json:"refresh_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	tm.copilotJWT = tokenResp.Token
	tm.expiresAt = tokenResp.ExpiresAt

	slog.Debug("token refreshed", "expires_at", time.Unix(tm.expiresAt, 0).Format(time.RFC3339))

	return nil
}

func (tm *TokenManager) StartAutoRefresh() {
	go func() {
		for {
			tm.mu.Lock()
			refreshIn := 600 // default 10 min
			if tm.expiresAt > 0 {
				remaining := tm.expiresAt - time.Now().Unix()
				if remaining > 120 {
					refreshIn = int(remaining) - 60
				}
			}
			tm.mu.Unlock()

			select {
			case <-time.After(time.Duration(refreshIn) * time.Second):
				tm.mu.Lock()
				_ = tm.refreshToken()
				tm.mu.Unlock()
			case <-tm.stopCh:
				return
			}
		}
	}()
}

func (tm *TokenManager) StopAutoRefresh() {
	close(tm.stopCh)
}
