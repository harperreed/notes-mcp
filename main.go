// ABOUTME: Main entry point for the notes-mcp CLI application
// ABOUTME: Sets up cobra command structure and executes root command

package main

import (
	"os"

	"github.com/harper/notes-mcp/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
