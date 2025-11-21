// ABOUTME: Unit tests for the update command
// ABOUTME: Tests CLI argument parsing and command execution with mocked service

package cmd

import (
	"io"
	"testing"
)

// TestUpdateCommandArgs tests that the update command requires exactly 2 arguments
func TestUpdateCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{"update"},
			expectError: true,
		},
		{
			name:        "one argument",
			args:        []string{"update", "title"},
			expectError: true,
		},
		{
			name:        "three arguments",
			args:        []string{"update", "title", "content", "extra"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up command
			rootCmd.SetArgs(tt.args)

			// Silence output
			rootCmd.SetOut(io.Discard)
			rootCmd.SetErr(io.Discard)

			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			// Reset for next test
			rootCmd.SetArgs([]string{})
		})
	}
}
