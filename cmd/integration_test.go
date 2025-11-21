// ABOUTME: Integration tests for CLI commands with real AppleScript execution
// ABOUTME: Only runs with -tags=integration build tag

//go:build integration

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestCreateCommandIntegration tests creating a note end-to-end
func TestCreateCommandIntegration(t *testing.T) {
	// Create a unique note title to avoid conflicts
	title := "CLI Test Note - Integration"
	content := "This is a test note created by integration tests"

	// Set up command
	args := []string{"create", title, content, "--tags=test,integration"}
	rootCmd.SetArgs(args)

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Note created: "+title) {
		t.Errorf("expected success message, got: %s", output)
	}

	// Clean up - try to delete the test note
	// Note: We don't have a delete command yet, so this will remain in Notes
	t.Log("Test note created successfully. Manual cleanup required in Apple Notes.")
}

// TestSearchCommandIntegration tests searching for notes end-to-end
func TestSearchCommandIntegration(t *testing.T) {
	// First create a test note to search for
	createTitle := "CLI Search Test Note"
	createContent := "Searchable content"

	createArgs := []string{"create", createTitle, createContent}
	rootCmd.SetArgs(createArgs)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("failed to create test note: %v", err)
	}

	// Now search for it
	searchArgs := []string{"search", "CLI Search Test"}
	rootCmd.SetArgs(searchArgs)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("search command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, createTitle) {
		t.Errorf("expected to find note in search results, got: %s", output)
	}

	t.Log("Test note created and found in search. Manual cleanup required in Apple Notes.")
}

// TestGetCommandIntegration tests getting note content end-to-end
func TestGetCommandIntegration(t *testing.T) {
	// First create a test note
	createTitle := "CLI Get Test Note"
	createContent := "Content to retrieve"

	createArgs := []string{"create", createTitle, createContent}
	rootCmd.SetArgs(createArgs)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("failed to create test note: %v", err)
	}

	// Now get its content
	getArgs := []string{"get", createTitle}
	rootCmd.SetArgs(getArgs)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}

	output := buf.String()
	// The content should contain the text we put in (possibly wrapped in HTML)
	if !strings.Contains(output, createContent) {
		t.Errorf("expected to find content in output, got: %s", output)
	}

	t.Log("Test note created and content retrieved. Manual cleanup required in Apple Notes.")
}
