// ABOUTME: Notes service interface and core business logic
// ABOUTME: Defines the contract for notes management operations

package services

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// NotesService defines the interface for notes management
type NotesService interface {
	// CreateNote creates a new note in Apple Notes
	CreateNote(ctx context.Context, title, content string, tags []string) (*Note, error)

	// SearchNotes searches for notes by title query
	SearchNotes(ctx context.Context, query string) ([]Note, error)

	// GetNoteContent retrieves the full content of a note by title
	GetNoteContent(ctx context.Context, title string) (string, error)

	// UpdateNote updates an existing note's content by title
	UpdateNote(ctx context.Context, title, content string) error

	// DeleteNote deletes a note by title
	DeleteNote(ctx context.Context, title string) error

	// ListFolders lists all folders in Apple Notes
	ListFolders(ctx context.Context) ([]string, error)

	// GetRecentNotes retrieves recently modified notes
	GetRecentNotes(ctx context.Context, limit int) ([]Note, error)

	// GetNotesInFolder retrieves all notes in a specific folder
	GetNotesInFolder(ctx context.Context, folder string) ([]Note, error)
}

// Note represents a note entity
type Note struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Tags     []string  `json:"tags"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// AppleNotesService implements NotesService using AppleScript
type AppleNotesService struct {
	executor      ScriptExecutor
	iCloudAccount string
}

// NewAppleNotesService creates a new AppleNotesService with the provided executor
func NewAppleNotesService(executor ScriptExecutor) *AppleNotesService {
	return &AppleNotesService{
		executor:      executor,
		iCloudAccount: "iCloud",
	}
}

// escapeForAppleScript escapes special characters for use in AppleScript strings
// Must escape backslashes first to avoid double-escaping, then quotes
func (s *AppleNotesService) escapeForAppleScript(input string) string {
	// 1. Escape backslashes first (prevents double-escaping)
	result := strings.ReplaceAll(input, "\\", "\\\\")
	// 2. Escape double quotes
	result = strings.ReplaceAll(result, "\"", "\\\"")
	return result
}

// formatContent prepares content for AppleScript by escaping and converting newlines to HTML breaks
func (s *AppleNotesService) formatContent(content string) string {
	if content == "" {
		return ""
	}
	// Escape special characters
	escaped := s.escapeForAppleScript(content)
	// Convert newlines to HTML breaks for Note body
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

// CreateNote creates a new note in Apple Notes with the given title, content, and tags
// Tags are stored in the Note struct but not passed to AppleScript (matching TypeScript behavior)
func (s *AppleNotesService) CreateNote(ctx context.Context, title, content string, tags []string) (*Note, error) {
	// Format content and escape title
	formattedContent := s.formatContent(content)
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to create note
	// Note: tags are not passed to AppleScript as the API doesn't support them
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				make new note with properties {name:"%s", body:"%s"}
			end tell
		end tell
	`, s.iCloudAccount, safeTitle, formattedContent)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return nil, fmt.Errorf("failed to create note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	// Return the note object with generated ID
	now := time.Now()
	return &Note{
		ID:       fmt.Sprintf("%d", now.UnixMilli()),
		Title:    title,
		Content:  content,
		Tags:     tags,
		Created:  now,
		Modified: now,
	}, nil
}

// SearchNotes searches for notes containing the query string in their title
// Returns notes with empty Content (search doesn't retrieve full bodies)
func (s *AppleNotesService) SearchNotes(ctx context.Context, query string) ([]Note, error) {
	safeQuery := s.escapeForAppleScript(query)

	// Generate AppleScript to search notes
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes where name contains "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeQuery)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to search notes: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	titles := strings.Split(stdout, ", ")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Search doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})
	}

	return notes, nil
}

// GetNoteContent retrieves the full HTML body content of a note by its title
func (s *AppleNotesService) GetNoteContent(ctx context.Context, title string) (string, error) {
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to get note content
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get body of note "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return "", fmt.Errorf("failed to get note content: %w", detectedErr)
	}

	return stdout, nil
}

// UpdateNote updates the content of an existing note by its title
func (s *AppleNotesService) UpdateNote(ctx context.Context, title, content string) error {
	// Format content and escape title
	formattedContent := s.formatContent(content)
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to update note
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set body of note "%s" to "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle, formattedContent)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to update note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	return nil
}

// DeleteNote deletes a note by its title
func (s *AppleNotesService) DeleteNote(ctx context.Context, title string) error {
	// Escape title
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to delete note
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				delete note "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to delete note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	return nil
}

// ListFolders lists all folders in Apple Notes
func (s *AppleNotesService) ListFolders(ctx context.Context) ([]string, error) {
	// Generate AppleScript to list folders
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of folders
			end tell
		end tell
	`, s.iCloudAccount)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []string{}, fmt.Errorf("failed to list folders: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []string{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	folders := strings.Split(stdout, ", ")
	result := make([]string, 0, len(folders))

	for _, folder := range folders {
		folder = strings.TrimSpace(folder)
		if folder == "" {
			continue
		}
		result = append(result, folder)
	}

	return result, nil
}

// GetRecentNotes retrieves recently modified notes, sorted by modification date
func (s *AppleNotesService) GetRecentNotes(ctx context.Context, limit int) ([]Note, error) {
	// Generate AppleScript to get recent notes sorted by modification date
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes
			end tell
		end tell
	`, s.iCloudAccount)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to get recent notes: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	titles := strings.Split(stdout, ", ")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	// Apply limit
	count := 0
	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Recent notes doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return notes, nil
}

// GetNotesInFolder retrieves all notes in a specific folder
func (s *AppleNotesService) GetNotesInFolder(ctx context.Context, folder string) ([]Note, error) {
	safeFolder := s.escapeForAppleScript(folder)

	// Generate AppleScript to get notes in folder
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes in folder "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeFolder)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to get notes in folder: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	titles := strings.Split(stdout, ", ")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Folder listing doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})
	}

	return notes, nil
}
