package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GitHubToken         string
	CopilotAPIURL       string
	CopilotTokenURL     string
	GitHubClientID      string
	GitHubDeviceCodeURL string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using fallback values")
	}

	AppConfig = &Config{
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		CopilotAPIURL:       getEnv("COPILOT_API_URL", "https://github.com/login/oauth/access_token"),
		CopilotTokenURL:     getEnv("COPILOT_TOKEN_URL", "https://api.github.com/copilot_internal/v2/token"),
		GitHubClientID:      os.Getenv("GITHUB_CLIENT_ID"),
		GitHubDeviceCodeURL: getEnv("GITHUB_DEVICE_CODE_URL", "https://github.com/login/device/code"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
