// ABOUTME: Unit tests for NotesService implementation
// ABOUTME: Tests all service methods with mock executor for isolation

package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// MockExecutor implements ScriptExecutor for testing
type MockExecutor struct {
	stdout string
	stderr string
	err    error
}

func (m *MockExecutor) Execute(ctx context.Context, script string) (string, string, error) {
	return m.stdout, m.stderr, m.err
}

// TestNoteJSONMarshaling verifies Note struct has correct JSON tags
func TestNoteJSONMarshaling(t *testing.T) {
	now := time.Now()
	note := Note{
		ID:       "123456789",
		Title:    "Test Note",
		Content:  "Test content",
		Tags:     []string{"tag1", "tag2"},
		Created:  now,
		Modified: now,
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("Failed to marshal note: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{"\"id\"", "\"title\"", "\"content\"", "\"tags\"", "\"created\"", "\"modified\""}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field %s. Got: %s", field, jsonStr)
		}
	}

	// Verify we can unmarshal back
	var decoded Note
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal note: %v", err)
	}

	if decoded.ID != note.ID || decoded.Title != note.Title || decoded.Content != note.Content {
		t.Errorf("Unmarshaled note doesn't match original")
	}
}

// TestEscapeForAppleScript tests the escaping logic
func TestEscapeForAppleScript(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "double quotes",
			input:    `He said "hello"`,
			expected: `He said \"hello\"`,
		},
		{
			name:     "backslashes",
			input:    `C:\Path\To\File`,
			expected: `C:\\Path\\To\\File`,
		},
		{
			name:     "backslashes and quotes",
			input:    `Path is "C:\Program Files"`,
			expected: `Path is \"C:\\Program Files\"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	executor := &MockExecutor{}
	service := NewAppleNotesService(executor)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.escapeForAppleScript(tt.input)
			if result != tt.expected {
				t.Errorf("escapeForAppleScript(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatContent tests content formatting for AppleScript
func TestFormatContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "newlines to br",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1<br>Line 2<br>Line 3",
		},
		{
			name:     "quotes and newlines",
			input:    "\"Quote\"\nNext line",
			expected: `\"Quote\"<br>Next line`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	executor := &MockExecutor{}
	service := NewAppleNotesService(executor)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatContent(tt.input)
			if result != tt.expected {
				t.Errorf("formatContent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCreateNote tests successful note creation
func TestCreateNote(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note created",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := "Test Note"
	content := "Test content with\nnewlines"
	tags := []string{"tag1", "tag2"}

	note, err := service.CreateNote(ctx, title, content, tags)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	if note.Title != title {
		t.Errorf("Note title = %q, want %q", note.Title, title)
	}

	if note.Content != content {
		t.Errorf("Note content = %q, want %q", note.Content, content)
	}

	if len(note.Tags) != len(tags) {
		t.Errorf("Note tags length = %d, want %d", len(note.Tags), len(tags))
	}

	if note.ID == "" {
		t.Error("Note ID should not be empty")
	}

	// Verify timestamps are set
	if note.Created.IsZero() {
		t.Error("Note created timestamp should be set")
	}

	if note.Modified.IsZero() {
		t.Error("Note modified timestamp should be set")
	}
}

// TestCreateNoteWithSpecialCharacters tests escaping in note creation
func TestCreateNoteWithSpecialCharacters(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note created",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := `Note with "quotes"`
	content := `Content with "quotes" and\nbackslashes`

	note, err := service.CreateNote(ctx, title, content, nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	if note.Title != title {
		t.Errorf("Note title = %q, want %q", note.Title, title)
	}
}

// TestCreateNoteError tests error handling in note creation
func TestCreateNoteError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "not allowed (-1743)",
		err:    ErrPermissionDenied,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.CreateNote(ctx, "Test", "Content", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Expected error containing 'permission denied', got %v", err)
	}
}

// TestSearchNotes tests successful note search
func TestSearchNotes(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Meeting Notes, Project Ideas, Random Thoughts",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	notes, err := service.SearchNotes(ctx, "notes")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(notes) != 3 {
		t.Fatalf("Expected 3 notes, got %d", len(notes))
	}

	expectedTitles := []string{"Meeting Notes", "Project Ideas", "Random Thoughts"}
	for i, note := range notes {
		if note.Title != expectedTitles[i] {
			t.Errorf("Note %d title = %q, want %q", i, note.Title, expectedTitles[i])
		}

		// Content should be empty in search results
		if note.Content != "" {
			t.Errorf("Note %d content should be empty, got %q", i, note.Content)
		}

		// ID should be set
		if note.ID == "" {
			t.Errorf("Note %d ID should not be empty", i)
		}
	}
}

// TestSearchNotesEmpty tests empty search results
func TestSearchNotesEmpty(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	notes, err := service.SearchNotes(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Expected empty result, got %d notes", len(notes))
	}
}

// TestSearchNotesWithError tests error handling in search
func TestSearchNotesWithError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "Apple Notes app not running",
		err:    ErrNotesAppNotRunning,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	notes, err := service.SearchNotes(ctx, "test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "Apple Notes app not running") {
		t.Errorf("Expected error containing 'Apple Notes app not running', got %v", err)
	}

	// Should return empty notes on error
	if len(notes) != 0 {
		t.Errorf("Expected empty notes on error, got %d", len(notes))
	}
}

// TestGetNoteContent tests successful content retrieval
func TestGetNoteContent(t *testing.T) {
	expectedContent := "<div>Test note content with <b>formatting</b></div>"
	executor := &MockExecutor{
		stdout: expectedContent,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	content, err := service.GetNoteContent(ctx, "Test Note")
	if err != nil {
		t.Fatalf("GetNoteContent failed: %v", err)
	}

	if content != expectedContent {
		t.Errorf("Content = %q, want %q", content, expectedContent)
	}
}

// TestGetNoteContentNotFound tests note not found error
func TestGetNoteContentNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.GetNoteContent(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestNewAppleNotesService tests constructor
func TestNewAppleNotesService(t *testing.T) {
	executor := &MockExecutor{}
	service := NewAppleNotesService(executor)

	if service == nil {
		t.Fatal("NewAppleNotesService returned nil")
	}

	// Verify default iCloud account is set
	if service.iCloudAccount != "iCloud" {
		t.Errorf("Expected iCloudAccount to be 'iCloud', got %q", service.iCloudAccount)
	}

	// Verify executor is set
	if service.executor == nil {
		t.Error("Executor should not be nil")
	}
}

// TestUpdateNote tests successful note update
func TestUpdateNote(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note updated",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := "Test Note"
	content := "Updated content with\nnewlines"

	err := service.UpdateNote(ctx, title, content)
	if err != nil {
		t.Fatalf("UpdateNote failed: %v", err)
	}
}

// TestUpdateNoteWithSpecialCharacters tests escaping in note update
func TestUpdateNoteWithSpecialCharacters(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note updated",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := `Note with "quotes"`
	content := `Updated content with "quotes" and\nbackslashes`

	err := service.UpdateNote(ctx, title, content)
	if err != nil {
		t.Fatalf("UpdateNote failed: %v", err)
	}
}

// TestUpdateNoteError tests error handling in note update
func TestUpdateNoteError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "not allowed (-1743)",
		err:    ErrPermissionDenied,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.UpdateNote(ctx, "Test", "Content")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Expected error containing 'permission denied', got %v", err)
	}
}

// TestUpdateNoteNotFound tests note not found error during update
func TestUpdateNoteNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.UpdateNote(ctx, "NonExistent", "New content")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestDeleteNote tests successful note deletion
func TestDeleteNote(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note deleted",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := "Test Note"

	err := service.DeleteNote(ctx, title)
	if err != nil {
		t.Fatalf("DeleteNote failed: %v", err)
	}
}

// TestDeleteNoteWithSpecialCharacters tests escaping in note deletion
func TestDeleteNoteWithSpecialCharacters(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note deleted",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := `Note with "quotes"`

	err := service.DeleteNote(ctx, title)
	if err != nil {
		t.Fatalf("DeleteNote failed: %v", err)
	}
}

// TestDeleteNoteError tests error handling in note deletion
func TestDeleteNoteError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "not allowed (-1743)",
		err:    ErrPermissionDenied,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.DeleteNote(ctx, "Test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Expected error containing 'permission denied', got %v", err)
	}
}

// TestDeleteNoteNotFound tests note not found error during deletion
func TestDeleteNoteNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.DeleteNote(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}
