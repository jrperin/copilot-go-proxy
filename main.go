package main

import (
	"log/slog"
	"os"

	"github.com/jrperin/copilot-go-proxy/cmd"
	"github.com/jrperin/copilot-go-proxy/internal/config"
	"github.com/jrperin/copilot-go-proxy/internal/logger"
)

func main() {
	config.LoadConfig()
	logger.Init()

	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		slog.Error("failed to execute root command", "error", err)
		os.Exit(1)
	}
}
