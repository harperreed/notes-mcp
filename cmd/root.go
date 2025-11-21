// ABOUTME: Root command setup for the notes-mcp CLI application
// ABOUTME: Defines the base cobra command structure and configuration

package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "notes-mcp",
	Short: "MCP server for notes management",
	Long:  `A Model Context Protocol server that provides intelligent notes management capabilities.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
