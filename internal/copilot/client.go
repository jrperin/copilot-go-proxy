package copilot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/jrperin/copilot-go-proxy/internal/auth"
)

const (
	CopilotBaseIndividual  = "https://api.githubcopilot.com"
	CopilotBaseBusiness    = "https://api.business.githubcopilot.com"
	CopilotBaseEnterprise  = "https://api.enterprise.githubcopilot.com"
)

type Client struct {
	tokenManager *auth.TokenManager
	baseURL      string
	models       []Model
	modelsMu     sync.RWMutex
}

func NewClient(tokenManager *auth.TokenManager) *Client {
	return &Client{
		tokenManager: tokenManager,
		baseURL:      CopilotBaseIndividual,
	}
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	token, err := c.tokenManager.GetToken()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	SetCopilotHeaders(req, token)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return http.DefaultClient.Do(req)
}

// ChatCompletions sends a JSON body to the Copilot chat completions endpoint
func (c *Client) ChatCompletions(jsonBody string) (*http.Response, error) {
	return c.doRequest("POST", "/chat/completions", strings.NewReader(jsonBody))
}

// ListModels fetches available models from Copilot
func (c *Client) ListModels() ([]Model, error) {
	c.modelsMu.RLock()
	if len(c.models) > 0 {
		models := c.models
		c.modelsMu.RUnlock()
		return models, nil
	}
	c.modelsMu.RUnlock()

	resp, err := c.doRequest("GET", "/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("models request failed (%d): %s", resp.StatusCode, string(body))
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, err
	}

	c.modelsMu.Lock()
	c.models = modelsResp.Data
	c.modelsMu.Unlock()

	return modelsResp.Data, nil
}

// InvalidateModelsCache clears the cached models list
func (c *Client) InvalidateModelsCache() {
	c.modelsMu.Lock()
	c.models = nil
	c.modelsMu.Unlock()
}
