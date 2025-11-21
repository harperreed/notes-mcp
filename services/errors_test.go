// ABOUTME: Unit tests for error detection logic
// ABOUTME: Uses table-driven approach to verify stderr pattern matching

package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDetectError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		err    error
		want   error
	}{
		// Notes app not running cases
		{
			name:   "notes app not running with -1728",
			stderr: "execution error: Apple Notes got an error: Connection invalid. (-1728)",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
		{
			name:   "notes app not running with event not handled",
			stderr: "error: event not handled by application",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
		{
			name:   "notes app not running case insensitive",
			stderr: "ERROR: Event Not Handled",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
		{
			name:   "notes app not running with -1728 mixed case",
			stderr: "Error -1728 occurred",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},

		// Note not found cases
		{
			name:   "note not found exact match",
			stderr: "note 'Meeting Notes' not found",
			err:    errors.New("script failed"),
			want:   ErrNoteNotFound,
		},
		{
			name:   "note not found case insensitive",
			stderr: "Note Not Found in database",
			err:    errors.New("script failed"),
			want:   ErrNoteNotFound,
		},
		{
			name:   "note not found with additional text",
			stderr: "The requested note was not found",
			err:    errors.New("script failed"),
			want:   ErrNoteNotFound,
		},
		{
			name:   "note not found mixed case pattern",
			stderr: "ERROR: Note 'Test' Not Found",
			err:    errors.New("script failed"),
			want:   ErrNoteNotFound,
		},

		// Permission denied cases
		{
			name:   "permission denied with not allowed",
			stderr: "operation not allowed",
			err:    errors.New("script failed"),
			want:   ErrPermissionDenied,
		},
		{
			name:   "permission denied with -1743",
			stderr: "error: access not allowed (-1743)",
			err:    errors.New("script failed"),
			want:   ErrPermissionDenied,
		},
		{
			name:   "permission denied case insensitive",
			stderr: "Access Not Allowed by system",
			err:    errors.New("script failed"),
			want:   ErrPermissionDenied,
		},
		{
			name:   "permission denied with -1743 only",
			stderr: "execution error: -1743",
			err:    errors.New("script failed"),
			want:   ErrPermissionDenied,
		},

		// Timeout cases
		{
			name:   "context deadline exceeded",
			stderr: "",
			err:    context.DeadlineExceeded,
			want:   ErrScriptTimeout,
		},
		{
			name:   "context deadline exceeded with stderr",
			stderr: "some output before timeout",
			err:    context.DeadlineExceeded,
			want:   ErrScriptTimeout,
		},

		// No match cases - should return original error
		{
			name:   "unknown error",
			stderr: "some unknown error message",
			err:    errors.New("original error"),
			want:   errors.New("original error"),
		},
		{
			name:   "empty stderr",
			stderr: "",
			err:    errors.New("original error"),
			want:   errors.New("original error"),
		},
		{
			name:   "random error code",
			stderr: "error -9999 occurred",
			err:    errors.New("script failed"),
			want:   errors.New("script failed"),
		},

		// Edge cases
		{
			name:   "multiple error patterns - first match wins (app not running)",
			stderr: "event not handled and note not found",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
		{
			name:   "whitespace in stderr",
			stderr: "  \n  event not handled  \n  ",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
		{
			name:   "error code in larger message",
			stderr: "The operation failed with error code -1728 and could not complete",
			err:    errors.New("script failed"),
			want:   ErrNotesAppNotRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got := DetectError(ctx, tt.stderr, tt.err)

			// For custom errors, use errors.Is
			if errors.Is(got, tt.want) {
				return
			}

			// For non-sentinel errors, compare messages
			if got.Error() != tt.want.Error() {
				t.Errorf("DetectError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectErrorWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Context canceled should return original error, not timeout
	err := errors.New("script failed")
	got := DetectError(ctx, "", err)

	if got.Error() != err.Error() {
		t.Errorf("DetectError() with canceled context = %v, want %v", got, err)
	}
}

func TestDetectErrorWithTimedOutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	// Should detect timeout
	got := DetectError(ctx, "", context.DeadlineExceeded)

	if !errors.Is(got, ErrScriptTimeout) {
		t.Errorf("DetectError() with timed out context = %v, want %v", got, ErrScriptTimeout)
	}
}

func TestSentinelErrorIdentity(t *testing.T) {
	// Verify that sentinel errors are distinct
	tests := []struct {
		name string
		err1 error
		err2 error
		want bool
	}{
		{
			name: "ErrNoteNotFound == ErrNoteNotFound",
			err1: ErrNoteNotFound,
			err2: ErrNoteNotFound,
			want: true,
		},
		{
			name: "ErrNoteNotFound != ErrNotesAppNotRunning",
			err1: ErrNoteNotFound,
			err2: ErrNotesAppNotRunning,
			want: false,
		},
		{
			name: "ErrPermissionDenied != ErrScriptTimeout",
			err1: ErrPermissionDenied,
			err2: ErrScriptTimeout,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err1, tt.err2)
			if got != tt.want {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err1, tt.err2, got, tt.want)
			}
		})
	}
}

func TestNoteNotFoundPattern(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"note not found", true},
		{"Note not found", true},
		{"NOTE NOT FOUND", true},
		{"note 'Test' not found", true},
		{"The note was not found", true},
		{"notes not found", true}, // plural also matches
		{"notification received", false},
		{"noted and logged", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := noteNotFoundPattern.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("noteNotFoundPattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
