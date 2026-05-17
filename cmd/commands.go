package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/jrperin/copilot-go-proxy/internal/auth"
	"github.com/jrperin/copilot-go-proxy/internal/copilot"
	"github.com/jrperin/copilot-go-proxy/internal/process"
	"github.com/jrperin/copilot-go-proxy/internal/server"
)

var Port int

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "copilot-proxy",
		Short: "GitHub Copilot proxy for OpenAI-compatible clients",
		Long:  "A standalone proxy that translates OpenAI-compatible API calls to GitHub Copilot.\nUse with OpenCode, Claude Code, or any OpenAI-compatible client.",
	}

	root.PersistentFlags().IntVarP(&Port, "port", "p", 4141, "Proxy port")

	root.AddCommand(
		newAuthCmd(),
		newImportAuthCmd(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newStatusCmd(),
		newDiagnoseCmd(),
		newConfigCmd(),
	)

	return root
}

func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with GitHub (OAuth device flow)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting GitHub authentication...")

			info, err := auth.RequestDeviceCode()
			if err != nil {
				return fmt.Errorf("requesting device code: %w", err)
			}

			fmt.Printf("\n  Open: %s\n", info.VerificationURI)
			fmt.Printf("  Code: %s\n\n", info.UserCode)
			fmt.Println("Waiting for authorization...")

			token, err := auth.PollForAccessToken(info.DeviceCode, info.Interval)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			if err := auth.SaveGitHubToken(token); err != nil {
				return fmt.Errorf("saving token: %w", err)
			}

			fmt.Println("\nAuthentication successful!")
			fmt.Printf("Token saved to %s\n", auth.DataDir())
			return nil
		},
	}
}

func newImportAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import-auth",
		Short: "Import token from existing copilot-api installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.ImportFromCopilotAPI(); err != nil {
				return err
			}
			fmt.Printf("Token imported from copilot-api to %s\n", auth.DataDir())
			return nil
		},
	}
}

func newStartCmd() *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !foreground {
				// Start as daemon
				if err := process.StartDaemon(); err != nil {
					return err
				}
				fmt.Printf("copilot-proxy started on port %d (daemon)\n", Port)
				return nil
			}

			// Foreground mode — actual server
			return runServer()
		},
	}

	cmd.Flags().BoolVar(&foreground, "foreground", false, "Run in foreground (internal)")
	return cmd
}

func runServer() error {
	githubToken, err := auth.LoadGitHubToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\nRun: copilot-proxy auth", err)
	}

	logFile, err := os.OpenFile(process.LogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}

	tokenManager := auth.NewTokenManager(githubToken)
	tokenManager.StartAutoRefresh()
	defer tokenManager.StopAutoRefresh()

	client := copilot.NewClient(tokenManager)
	srv := server.New(Port, client)

	log.Printf("Starting copilot-proxy on port %d", Port)

	if err := srv.Start(); err != nil {
		log.Printf("Server stopped: %v", err)
		return err
	}

	return nil
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := process.StopDaemon(); err != nil {
				return err
			}
			fmt.Println("copilot-proxy stopped")
			return nil
		},
	}
}

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = process.StopDaemon()
			if err := process.StartDaemon(); err != nil {
				return err
			}
			fmt.Printf("copilot-proxy restarted on port %d\n", Port)
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show proxy status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status := process.GetStatus(Port)

			if status.Running {
				fmt.Printf("copilot-proxy: running (PID %d, port %d)\n", status.PID, status.Port)
			} else {
				fmt.Println("copilot-proxy: stopped")
			}

			if status.Authenticated {
				fmt.Println("auth: ok")
			} else {
				fmt.Println("auth: not authenticated (run: copilot-proxy auth)")
			}

			if status.APIHealthy {
				fmt.Println("api: healthy")
			} else if status.Running {
				fmt.Println("api: not responding")
			}

			return nil
		},
	}
}

func newDiagnoseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Run full diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			diag := process.Diagnose(Port)
			data, _ := json.MarshalIndent(diag, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print provider config for opencode.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := map[string]interface{}{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "GitHub Copilot",
				"options": map[string]string{
					"baseURL": fmt.Sprintf("http://127.0.0.1:%d/v1", Port),
					"apiKey":  "sk-dummy",
				},
				"models": map[string]interface{}{
					"claude-sonnet-4": map[string]interface{}{
						"name": "Claude Sonnet 4 (Copilot)",
						"limit": map[string]int{
							"context": 128000,
							"output":  16384,
						},
					},
					"claude-opus-4": map[string]interface{}{
						"name": "Claude Opus 4 (Copilot)",
						"limit": map[string]int{
							"context": 128000,
							"output":  16384,
						},
					},
					"gpt-4o": map[string]interface{}{
						"name": "GPT-4o (Copilot)",
						"limit": map[string]int{
							"context": 128000,
							"output":  16384,
						},
					},
					"gpt-4o-mini": map[string]interface{}{
						"name": "GPT-4o Mini (Copilot)",
						"limit": map[string]int{
							"context": 128000,
							"output":  16384,
						},
					},
				},
			}

			data, _ := json.MarshalIndent(config, "", "  ")
			fmt.Println(string(data))
			fmt.Println("\n# Add this to your opencode.json under \"provider\" -> \"copilot\"")
			return nil
		},
	}
}
