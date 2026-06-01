package copilot

import "time"

// GitHub OAuth device flow
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// Copilot token
type CopilotToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	RefreshIn int    `json:"refresh_in"`
}

type CopilotTokenStore struct {
	GitHubToken string    `json:"github_token"`
	CopilotJWT  string    `json:"copilot_jwt"`
	ExpiresAt   int64     `json:"expires_at"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// Copilot models response
type ModelsResponse struct {
	Data []Model `json:"data"`
}

type Model struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Vendor       string    `json:"vendor"`
	Capabilities ModelCaps `json:"capabilities"`
}

type ModelCaps struct {
	Family   string        `json:"family"`
	Type     string        `json:"type"`
	Limits   ModelLimits   `json:"limits"`
	Supports ModelSupports `json:"supports"`
}

type ModelLimits struct {
	MaxContext int `json:"max_context_window_tokens"`
	MaxOutput  int `json:"max_output_tokens"`
}

type ModelSupports struct {
	ToolCalls         bool `json:"tool_calls"`
	ParallelToolCalls bool `json:"parallel_tool_calls"`
}
