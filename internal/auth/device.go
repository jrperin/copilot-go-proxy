package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jrperin/copilot-go-proxy/internal/config"
)

const (
	githubScopes = "read:user"
)

type DeviceCodeInfo struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       int
	Interval        int
}

func RequestDeviceCode() (*DeviceCodeInfo, error) {
	data := url.Values{
		"client_id": {config.AppConfig.GitHubClientID},
		"scope":     {githubScopes},
	}

	req, err := http.NewRequest("POST", config.AppConfig.GitHubDeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var raw struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if raw.DeviceCode == "" {
		return nil, fmt.Errorf("no device code in response: %s", string(body))
	}

	return &DeviceCodeInfo{
		DeviceCode:      raw.DeviceCode,
		UserCode:        raw.UserCode,
		VerificationURI: raw.VerificationURI,
		ExpiresIn:       raw.ExpiresIn,
		Interval:        raw.Interval,
	}, nil
}

func PollForAccessToken(deviceCode string, interval int) (string, error) {
	ticker := time.NewTicker(time.Duration(interval+1) * time.Second)
	defer ticker.Stop()

	timeout := time.After(15 * time.Minute)

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("authentication timed out after 15 minutes")
		case <-ticker.C:
			token, err := pollOnce(deviceCode)
			if err != nil {
				continue // retry
			}
			if token != "" {
				return token, nil
			}
			// authorization_pending — keep polling
		}
	}
}

func pollOnce(deviceCode string) (string, error) {
	data := url.Values{
		"client_id":   {config.AppConfig.GitHubClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	req, err := http.NewRequest("POST", config.AppConfig.CopilotAPIURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error == "authorization_pending" {
		return "", nil // not yet
	}

	if result.Error != "" {
		return "", fmt.Errorf("oauth error: %s", result.Error)
	}

	return result.AccessToken, nil
}
