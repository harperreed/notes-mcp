// ABOUTME: Common helper functions for CLI commands
// ABOUTME: Provides factories for creating services and contexts with consistent configuration

package cmd

import (
	"context"
	"time"

	"github.com/harper/notes-mcp/services"
)

// Command-level timeouts
const (
	// osascriptTimeout is the timeout for individual AppleScript invocations
	osascriptTimeout = 10 * time.Second
	// commandTimeout is the timeout for the entire command execution
	commandTimeout = 30 * time.Second
)

// newNotesService creates an AppleNotesService with a configured OSAScriptExecutor
func newNotesService() *services.AppleNotesService {
	executor := services.NewOSAScriptExecutor(osascriptTimeout)
	return services.NewAppleNotesService(executor)
}

// newCommandContext creates a context with a timeout for command execution
func newCommandContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), commandTimeout)
}
