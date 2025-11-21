// ABOUTME: Error definitions and detection for Apple Notes operations
// ABOUTME: Maps AppleScript error codes to structured Go errors

package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// Sentinel errors for common Apple Notes failures
var (
	ErrNoteNotFound       = errors.New("note not found")
	ErrFolderNotFound     = errors.New("folder not found")
	ErrNotesAppNotRunning = errors.New("Apple Notes app not running")
	ErrPermissionDenied   = errors.New("permission denied to access Notes")
	ErrScriptTimeout      = errors.New("AppleScript execution timeout")
	ErrInvalidInput       = errors.New("invalid input parameters")
)

// noteNotFoundPattern matches various "note not found" error messages
// Matches "note" followed by anything (non-greedy), then "not found" as a phrase
var noteNotFoundPattern = regexp.MustCompile(`(?i)note.*?\bnot\s+found\b`)

// folderNotFoundPattern matches various "folder not found" error messages
// Matches "folder" followed by anything (non-greedy), then "not found" as a phrase
var folderNotFoundPattern = regexp.MustCompile(`(?i)folder.*?\bnot\s+found\b`)

// DetectError analyzes stderr output and context errors to return structured errors
// Pattern matching:
// - "-1728" or "event not handled" → ErrNotesAppNotRunning
// - "note.*not found" (regex) → ErrNoteNotFound
// - "not allowed" or "-1743" → ErrPermissionDenied
// - context.DeadlineExceeded → ErrScriptTimeout
func DetectError(ctx context.Context, stderr string, err error) error {
	// Check for context deadline exceeded first
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrScriptTimeout
	}

	// If stderr is empty, return original error
	if stderr == "" {
		return err
	}

	stderrLower := strings.ToLower(stderr)

	// Check for Notes app not running
	if strings.Contains(stderrLower, "-1728") || strings.Contains(stderrLower, "event not handled") {
		return ErrNotesAppNotRunning
	}

	// Check for permission denied
	if strings.Contains(stderrLower, "not allowed") || strings.Contains(stderrLower, "-1743") {
		return ErrPermissionDenied
	}

	// Check for note not found
	if noteNotFoundPattern.MatchString(stderr) {
		return ErrNoteNotFound
	}

	// Check for folder not found
	if folderNotFoundPattern.MatchString(stderr) {
		return ErrFolderNotFound
	}

	// Return original error if no pattern matches
	return err
}
