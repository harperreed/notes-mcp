// ABOUTME: AppleScript execution layer for running osascript commands
// ABOUTME: Provides interface and implementation for executing AppleScript with timeout support
package services

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// ScriptExecutor defines the interface for executing AppleScript commands
type ScriptExecutor interface {
	Execute(ctx context.Context, script string) (stdout string, stderr string, err error)
}

// OSAScriptExecutor implements ScriptExecutor using the osascript command
type OSAScriptExecutor struct {
	timeout time.Duration
}

// NewOSAScriptExecutor creates a new OSAScriptExecutor with the specified timeout.
// If timeout is 0 or negative, defaults to 10 seconds.
func NewOSAScriptExecutor(timeout time.Duration) *OSAScriptExecutor {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &OSAScriptExecutor{
		timeout: timeout,
	}
}

// Execute runs the provided AppleScript using osascript and returns stdout, stderr, and any error.
// The execution is subject to the configured timeout and respects context cancellation.
func (e *OSAScriptExecutor) Execute(ctx context.Context, script string) (string, string, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Create command with context for cancellation support
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)

	// Buffers to capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
