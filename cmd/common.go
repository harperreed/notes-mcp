// ABOUTME: Common helper functions for CLI commands
// ABOUTME: Provides factories for creating services and contexts with consistent configuration

package cmd

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/harper/notes-mcp/services"
)

// Command-level timeouts
const (
	// osascriptTimeout is the timeout for individual AppleScript invocations
	osascriptTimeout = 10 * time.Second
	// commandTimeout is the timeout for the entire command execution
	commandTimeout = 30 * time.Second
	// maxSearchResults limits search results to prevent timeouts with large result sets
	maxSearchResults = 100
)

// getOperationTimeout returns the operation timeout, checking NOTES_MCP_TIMEOUT env var first
func getOperationTimeout() time.Duration {
	if timeoutStr := os.Getenv("NOTES_MCP_TIMEOUT"); timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return commandTimeout
}

// newNotesService creates an AppleNotesService with a configured OSAScriptExecutor
func newNotesService() *services.AppleNotesService {
	executor := services.NewOSAScriptExecutor(osascriptTimeout)
	return services.NewAppleNotesService(executor)
}

// newCommandContext creates a context with a timeout for command execution
func newCommandContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), commandTimeout)
}
