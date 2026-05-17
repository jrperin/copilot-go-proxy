package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	dataDirName = ".copilot-proxy"
	tokenFile   = "auth.json"
)

func DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dataDirName)
}

func EnsureDataDir() error {
	return os.MkdirAll(DataDir(), 0700)
}

func tokenFilePath() string {
	return filepath.Join(DataDir(), tokenFile)
}

type StoredAuth struct {
	GitHubToken string `json:"github_token"`
	FetchedAt   string `json:"fetched_at,omitempty"`
}

func SaveGitHubToken(token string) error {
	if err := EnsureDataDir(); err != nil {
		return err
	}

	auth := StoredAuth{
		GitHubToken: token,
	}

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFilePath(), data, 0600)
}

func LoadGitHubToken() (string, error) {
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			// Fallback: try copilot-api token location
			if token, err2 := loadCopilotAPIToken(); err2 == nil {
				return token, nil
			}
			return "", fmt.Errorf("not authenticated. Run: copilot-proxy auth")
		}
		return "", err
	}

	var auth StoredAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		return "", fmt.Errorf("corrupt auth file: %w", err)
	}

	if auth.GitHubToken == "" {
		return "", fmt.Errorf("no GitHub token found. Run: copilot-proxy auth")
	}

	return auth.GitHubToken, nil
}

func IsAuthenticated() bool {
	_, err := LoadGitHubToken()
	return err == nil
}

func RemoveAuth() error {
	return os.Remove(tokenFilePath())
}

// loadCopilotAPIToken tries to load token from copilot-api's default location
func loadCopilotAPIToken() (string, error) {
	home, _ := os.UserHomeDir()
	copilotAPITokenPath := filepath.Join(home, ".local", "share", "copilot-api", "github_token")
	data, err := os.ReadFile(copilotAPITokenPath)
	if err != nil {
		return "", err
	}
	token := string(data)
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	return token, nil
}

// ImportFromCopilotAPI imports token from copilot-api's stored location
func ImportFromCopilotAPI() error {
	token, err := loadCopilotAPIToken()
	if err != nil {
		return fmt.Errorf("copilot-api token not found: %w", err)
	}
	return SaveGitHubToken(token)
}
