// ABOUTME: Tests for AppleScript execution layer
// ABOUTME: Validates OSAScriptExecutor with timeout and error handling
package services

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestOSAScriptExecutor_Execute_Success(t *testing.T) {
	executor := NewOSAScriptExecutor(10 * time.Second)

	// Simple AppleScript that prints to stdout
	script := `return "hello world"`

	stdout, stderr, err := executor.Execute(context.Background(), script)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if strings.TrimSpace(stdout) != "hello world" {
		t.Errorf("expected stdout to be 'hello world', got: %q", stdout)
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
}

func TestOSAScriptExecutor_Execute_ScriptError(t *testing.T) {
	executor := NewOSAScriptExecutor(10 * time.Second)

	// Invalid AppleScript that will fail
	script := `error "test error"`

	stdout, stderr, err := executor.Execute(context.Background(), script)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got: %q", stdout)
	}

	if !strings.Contains(stderr, "test error") {
		t.Errorf("expected stderr to contain 'test error', got: %q", stderr)
	}
}

func TestOSAScriptExecutor_Execute_Timeout(t *testing.T) {
	// Create executor with very short timeout
	executor := NewOSAScriptExecutor(100 * time.Millisecond)

	// Script that sleeps longer than timeout
	script := `delay 1`

	ctx := context.Background()
	_, _, err := executor.Execute(ctx, script)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Should get context deadline exceeded or similar timeout error
	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "signal: killed") {
		t.Errorf("expected timeout-related error, got: %v", err)
	}
}

func TestOSAScriptExecutor_Execute_ContextCancellation(t *testing.T) {
	executor := NewOSAScriptExecutor(10 * time.Second)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	script := `return "test"`

	_, _, err := executor.Execute(ctx, script)

	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected 'context canceled' error, got: %v", err)
	}
}

func TestOSAScriptExecutor_Execute_CapturesStdoutAndStderr(t *testing.T) {
	executor := NewOSAScriptExecutor(10 * time.Second)

	// Script that outputs to both stdout (via return) and stderr (via log)
	// Note: AppleScript's 'log' goes to stderr when run via osascript
	script := `
log "this is a log message"
return "this is the result"
`

	stdout, stderr, err := executor.Execute(context.Background(), script)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(stdout, "this is the result") {
		t.Errorf("expected stdout to contain result, got: %q", stdout)
	}

	if !strings.Contains(stderr, "this is a log message") {
		t.Errorf("expected stderr to contain log message, got: %q", stderr)
	}
}

func TestOSAScriptExecutor_DefaultTimeout(t *testing.T) {
	executor := NewOSAScriptExecutor(0)

	// Should use default 10 second timeout
	if executor.timeout != 10*time.Second {
		t.Errorf("expected default timeout of 10s, got: %v", executor.timeout)
	}
}

func TestOSAScriptExecutor_CustomTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	executor := NewOSAScriptExecutor(customTimeout)

	if executor.timeout != customTimeout {
		t.Errorf("expected timeout of %v, got: %v", customTimeout, executor.timeout)
	}
}

func TestScriptExecutor_InterfaceCompliance(t *testing.T) {
	// Verify that OSAScriptExecutor implements ScriptExecutor interface
	var _ ScriptExecutor = (*OSAScriptExecutor)(nil)
}
