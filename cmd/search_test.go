// ABOUTME: Unit tests for the search command
// ABOUTME: Tests CLI argument parsing and command structure

package cmd

import (
	"io"
	"testing"
)

// TestSearchCommandArgs tests that the search command requires exactly 1 argument
func TestSearchCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{"search"},
			expectError: true,
		},
		{
			name:        "two arguments",
			args:        []string{"search", "query", "extra"},
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
