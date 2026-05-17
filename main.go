package main

import (
	"fmt"
	"os"

	"github.com/jrperin/copilot-go-proxy/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
