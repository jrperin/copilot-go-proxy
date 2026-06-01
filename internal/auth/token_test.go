package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jrperin/copilot-go-proxy/internal/config"
)

func init() {
	config.AppConfig = &config.Config{
		CopilotTokenURL: "https://api.github.com/copilot_internal/v2/token",
	}
}

func TestRefreshToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token test-github-token" {
			t.Errorf("expected authorization header, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":      "test-copilot-jwt",
			"expires_at": time.Now().Unix() + 3600,
		})
	}))
	defer server.Close()

	tm := &TokenManager{
		githubToken: "test-github-token",
		copilotJWT:  "",
		expiresAt:   0,
	}

	// Temporarily override the config URL to point to our test server
	originalURL := config.AppConfig.CopilotTokenURL
	config.AppConfig.CopilotTokenURL = server.URL
	defer func() { config.AppConfig.CopilotTokenURL = originalURL }()

	err := tm.refreshToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tm.copilotJWT != "test-copilot-jwt" {
		t.Errorf("expected copilotJWT to be 'test-copilot-jwt', got '%s'", tm.copilotJWT)
	}
	if tm.expiresAt == 0 {
		t.Error("expected expiresAt to be set")
	}
}

func TestRefreshToken_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Close the server immediately to cause a network error
	server.Close()

	tm := &TokenManager{
		githubToken: "test-github-token",
	}

	originalURL := config.AppConfig.CopilotTokenURL
	config.AppConfig.CopilotTokenURL = server.URL
	defer func() { config.AppConfig.CopilotTokenURL = originalURL }()

	err := tm.refreshToken()
	if err == nil {
		t.Fatal("expected a network error, but got nil")
	}
}

func TestRefreshToken_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal server error")
	}))
	defer server.Close()

	tm := &TokenManager{
		githubToken: "test-github-token",
	}

	originalURL := config.AppConfig.CopilotTokenURL
	config.AppConfig.CopilotTokenURL = server.URL
	defer func() { config.AppConfig.CopilotTokenURL = originalURL }()

	err := tm.refreshToken()
	if err == nil {
		t.Fatal("expected an error for 500 status, but got nil")
	}

	expectedError := "copilot token request failed (500): internal server error"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token": "test-jwt", "expires_at": "not-a-number"}`)
	}))
	defer server.Close()

	tm := &TokenManager{
		githubToken: "test-github-token",
	}

	originalURL := config.AppConfig.CopilotTokenURL
	config.AppConfig.CopilotTokenURL = server.URL
	defer func() { config.AppConfig.CopilotTokenURL = originalURL }()

	err := tm.refreshToken()
	if err == nil {
		t.Fatal("expected a JSON parsing error, but got nil")
	}

	// Error may be wrapped, just check that it mentions parsing
	if !contains(err.Error(), "parsing token response") {
		t.Errorf("expected error to mention 'parsing token response', got '%s'", err.Error())
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
