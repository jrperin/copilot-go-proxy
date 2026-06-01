package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jrperin/copilot-go-proxy/internal/copilot"
)

func newModelsCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "models",
		Short: "Discover Copilot models and generate opencode.json",
		Long: `Discover all available Copilot models, test their availability, 
and generate an opencode.json configuration file with only working models.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := generateModelsConfig(output); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "opencode.json", "Output file path")
	return cmd
}

type ModelInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Vendor       string                 `json:"vendor"`
	Weight       string                 `json:"weight,omitempty"`
	Context      int                    `json:"context"`
	Output       int                    `json:"output"`
	Type         string                 `json:"type"`
	Working      bool                   `json:"working"`
	MinTokens    int                    `json:"min_tokens"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}

func generateModelsConfig(output string) error {
	log.Println("Fetching models from local proxy...")

	// Fetch models from local proxy
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/models", Port))
	if err != nil {
		return fmt.Errorf("failed to fetch models from proxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("proxy returned status %d", resp.StatusCode)
	}

	var modelsResp struct {
		Data []copilot.Model `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return fmt.Errorf("failed to decode models response: %w", err)
	}

	log.Printf("Found %d models\n", len(modelsResp.Data))

	// Test each model
	var workingModels []ModelInfo
	var notWorkingModels []ModelInfo

	for _, model := range modelsResp.Data {
		if model.Capabilities.Type != "chat" {
			continue // Only test chat models
		}

		info := ModelInfo{
			ID:        model.ID,
			Name:      model.Name,
			Vendor:    model.Vendor,
			Context:   model.Capabilities.Limits.MaxContext,
			Output:    model.Capabilities.Limits.MaxOutput,
			Type:      model.Capabilities.Type,
			MinTokens: 1,
		}

		// Test model availability
		info.Working = testModel(model.ID)

		if info.Working {
			workingModels = append(workingModels, info)
		} else {
			notWorkingModels = append(notWorkingModels, info)
		}
	}

	log.Printf("✅ Working models: %d\n", len(workingModels))
	log.Printf("❌ Not working: %d\n", len(notWorkingModels))

	// Sort by context window (descending)
	sort.Slice(workingModels, func(i, j int) bool {
		return workingModels[i].Context > workingModels[j].Context
	})

	// Generate opencode configuration
	config := map[string]interface{}{
		"$schema":           "https://opencode.ai/config.json",
		"enabled_providers": []string{"copilot"},
		"model":             workingModels[0].ID,
		"provider": map[string]interface{}{
			"copilot": map[string]interface{}{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "GitHub Copilot",
				"options": map[string]string{
					"baseURL": "http://127.0.0.1:4141/v1",
					"apiKey":  "sk-dummy",
				},
				"models": map[string]interface{}{},
			},
		},
	}

	// Add models to config
	for _, model := range workingModels {
		config["provider"].(map[string]interface{})["copilot"].(map[string]interface{})["models"].(map[string]interface{})[model.ID] = map[string]interface{}{
			"name": model.Name,
			"limit": map[string]int{
				"context": model.Context,
				"output":  model.Output,
			},
		}
	}

	// Write output file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(output, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("✅ Configuration written to: %s\n", output)

	// Print summary
	fmt.Println("\n=== Model Summary ===")
	fmt.Printf("Total chat models: %d\n", len(modelsResp.Data))
	fmt.Printf("Working models: %d\n", len(workingModels))
	fmt.Printf("Not working: %d\n", len(notWorkingModels))

	if len(workingModels) > 0 {
		fmt.Println("\nWorking models:")
		for _, model := range workingModels {
			fmt.Printf("  - %s (%s): %dk context, %dk output\n",
				model.ID, model.Vendor, model.Context/1024, model.Output/1024)
		}
	}

	return nil
}

func testModel(modelID string) bool {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("http://127.0.0.1:4141/v1/chat/completions")
	reqBody := fmt.Sprintf(`{
		"model": "%s",
		"messages": [{"role": "user", "content": "Hi"}],
		"max_tokens": 1
	}`, modelID)

	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-dummy")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
