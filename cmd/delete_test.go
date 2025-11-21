// ABOUTME: Unit tests for the delete command
// ABOUTME: Tests CLI argument parsing and command execution with mocked service

package cmd

import (
	"io"
	"testing"
)

// TestDeleteCommandArgs tests that the delete command requires exactly 1 argument
func TestDeleteCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{"delete"},
			expectError: true,
		},
		{
			name:        "two arguments",
			args:        []string{"delete", "title", "extra"},
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
