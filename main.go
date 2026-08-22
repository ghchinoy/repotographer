package main

import (
	"fmt"
	"os"

	"github.com/ghchinoy/repotographer/internal/cli"
	_ "github.com/ghchinoy/repotographer/internal/mcp"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
