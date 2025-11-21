// ABOUTME: Unit tests for NotesService implementation
// ABOUTME: Tests all service methods with mock executor for isolation

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// SequentialMockExecutor implements ScriptExecutor for testing with multiple sequential responses
type SequentialMockExecutor struct {
	responses []struct {
		stdout string
		stderr string
		err    error
	}
	callIndex int
}

func (m *SequentialMockExecutor) Execute(ctx context.Context, script string) (string, string, error) {
	if m.callIndex >= len(m.responses) {
		return "", "", fmt.Errorf("unexpected call to Execute (call %d, only %d responses configured)", m.callIndex, len(m.responses))
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return resp.stdout, resp.stderr, resp.err
}

// TestNoteJSONMarshaling verifies Note struct has correct JSON tags
func TestNoteJSONMarshaling(t *testing.T) {
	now := time.Now()
	note := Note{
		ID:                "123456789",
		Title:             "Test Note",
		Content:           "Test content",
		Tags:              []string{"tag1", "tag2"},
		Created:           now,
		Modified:          now,
		CreationDate:      now,
		ModificationDate:  now,
		Folder:            "Work",
		Shared:            true,
		PasswordProtected: false,
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("Failed to marshal note: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{
		"\"id\"", "\"title\"", "\"content\"", "\"tags\"",
		"\"created\"", "\"modified\"",
		"\"creation_date\"", "\"modification_date\"",
		"\"folder\"", "\"shared\"", "\"password_protected\"",
	}
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
	if decoded.Folder != note.Folder {
		t.Errorf("Unmarshaled folder = %q, want %q", decoded.Folder, note.Folder)
	}
	if decoded.Shared != note.Shared {
		t.Errorf("Unmarshaled shared = %v, want %v", decoded.Shared, note.Shared)
	}
	if decoded.PasswordProtected != note.PasswordProtected {
		t.Errorf("Unmarshaled password_protected = %v, want %v", decoded.PasswordProtected, note.PasswordProtected)
	}
}

// TestAttachmentJSONMarshaling verifies Attachment struct has correct JSON tags
func TestAttachmentJSONMarshaling(t *testing.T) {
	now := time.Now()
	attachment := Attachment{
		Name:              "document.pdf",
		FilePath:          "/Users/test/document.pdf",
		ContentIdentifier: "x-coredata://12345",
		CreationDate:      now,
		ModificationDate:  now,
		ID:                "att-123",
	}

	data, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("Failed to marshal attachment: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{
		"\"name\"", "\"file_path\"", "\"content_identifier\"",
		"\"creation_date\"", "\"modification_date\"", "\"id\"",
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field %s. Got: %s", field, jsonStr)
		}
	}

	// Verify we can unmarshal back
	var decoded Attachment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal attachment: %v", err)
	}

	if decoded.Name != attachment.Name {
		t.Errorf("Unmarshaled name = %q, want %q", decoded.Name, attachment.Name)
	}
	if decoded.FilePath != attachment.FilePath {
		t.Errorf("Unmarshaled file_path = %q, want %q", decoded.FilePath, attachment.FilePath)
	}
	if decoded.ContentIdentifier != attachment.ContentIdentifier {
		t.Errorf("Unmarshaled content_identifier = %q, want %q", decoded.ContentIdentifier, attachment.ContentIdentifier)
	}
}

// TestFolderNodeJSONMarshaling verifies FolderNode struct has correct JSON tags
func TestFolderNodeJSONMarshaling(t *testing.T) {
	folder := FolderNode{
		Name:      "Work",
		Shared:    true,
		NoteCount: 42,
		Children: []FolderNode{
			{
				Name:      "Projects",
				Shared:    false,
				NoteCount: 15,
				Children:  []FolderNode{},
			},
			{
				Name:      "Archive",
				Shared:    false,
				NoteCount: 27,
				Children:  nil,
			},
		},
	}

	data, err := json.Marshal(folder)
	if err != nil {
		t.Fatalf("Failed to marshal folder: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{
		"\"name\"", "\"shared\"", "\"note_count\"", "\"children\"",
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing expected field %s. Got: %s", field, jsonStr)
		}
	}

	// Verify we can unmarshal back
	var decoded FolderNode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal folder: %v", err)
	}

	if decoded.Name != folder.Name {
		t.Errorf("Unmarshaled name = %q, want %q", decoded.Name, folder.Name)
	}
	if decoded.Shared != folder.Shared {
		t.Errorf("Unmarshaled shared = %v, want %v", decoded.Shared, folder.Shared)
	}
	if decoded.NoteCount != folder.NoteCount {
		t.Errorf("Unmarshaled note_count = %d, want %d", decoded.NoteCount, folder.NoteCount)
	}
	if len(decoded.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(decoded.Children))
	}
	if decoded.Children[0].Name != "Projects" {
		t.Errorf("First child name = %q, want %q", decoded.Children[0].Name, "Projects")
	}
}

// TestSearchOptionsStructure verifies SearchOptions struct fields
func TestSearchOptionsStructure(t *testing.T) {
	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := SearchOptions{
		Query:    "meeting notes",
		SearchIn: "both",
		Folder:   "Work",
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	// Verify all fields are accessible
	if opts.Query != "meeting notes" {
		t.Errorf("Query = %q, want %q", opts.Query, "meeting notes")
	}
	if opts.SearchIn != "both" {
		t.Errorf("SearchIn = %q, want %q", opts.SearchIn, "both")
	}
	if opts.Folder != "Work" {
		t.Errorf("Folder = %q, want %q", opts.Folder, "Work")
	}
	if opts.DateFrom == nil {
		t.Fatal("DateFrom should not be nil")
	}
	if opts.DateTo == nil {
		t.Fatal("DateTo should not be nil")
	}
	if !opts.DateFrom.Equal(dateFrom) {
		t.Errorf("DateFrom = %v, want %v", opts.DateFrom, dateFrom)
	}
	if !opts.DateTo.Equal(dateTo) {
		t.Errorf("DateTo = %v, want %v", opts.DateTo, dateTo)
	}
}

// TestSearchOptionsNilDates verifies SearchOptions with nil date pointers
func TestSearchOptionsNilDates(t *testing.T) {
	opts := SearchOptions{
		Query:    "test",
		SearchIn: "title",
		Folder:   "",
		DateFrom: nil,
		DateTo:   nil,
	}

	// Verify all fields are properly set
	if opts.Query != "test" {
		t.Errorf("Query = %q, want %q", opts.Query, "test")
	}
	if opts.SearchIn != "title" {
		t.Errorf("SearchIn = %q, want %q", opts.SearchIn, "title")
	}
	if opts.Folder != "" {
		t.Errorf("Folder = %q, want empty string", opts.Folder)
	}
	if opts.DateFrom != nil {
		t.Error("DateFrom should be nil")
	}
	if opts.DateTo != nil {
		t.Error("DateTo should be nil")
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

// TestCreateNote tests successful note creation with full metadata retrieval
func TestCreateNote(t *testing.T) {
	// Mock metadata response from GetNoteMetadata call
	metadataResponse := `{id:"x-coredata://12345", name:"Test Note", creation date:date "Monday, January 1, 2024 at 10:00:00 AM", modification date:date "Monday, January 1, 2024 at 11:30:00 AM", container:"Work", shared:true, password protected:false}`

	executor := &SequentialMockExecutor{
		responses: []struct {
			stdout string
			stderr string
			err    error
		}{
			{stdout: "note created", stderr: "", err: nil},   // CreateNote response
			{stdout: metadataResponse, stderr: "", err: nil}, // GetNoteMetadata response
		},
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

	// Verify basic fields
	if note.Title != title {
		t.Errorf("Note title = %q, want %q", note.Title, title)
	}

	if note.Content != content {
		t.Errorf("Note content = %q, want %q", note.Content, content)
	}

	if len(note.Tags) != len(tags) {
		t.Errorf("Note tags length = %d, want %d", len(note.Tags), len(tags))
	}

	// Verify metadata fields are populated
	if note.ID == "" || note.ID == fmt.Sprintf("%d", time.Now().UnixMilli()) {
		t.Error("Note ID should be populated from metadata, not generated")
	}

	if note.Folder != "Work" {
		t.Errorf("Note folder = %q, want %q", note.Folder, "Work")
	}

	if !note.Shared {
		t.Error("Note shared should be true from metadata")
	}

	if note.PasswordProtected {
		t.Error("Note password_protected should be false from metadata")
	}

	// Verify timestamps are set from metadata
	if note.Created.IsZero() || note.CreationDate.IsZero() {
		t.Error("Note creation timestamps should be set from metadata")
	}

	if note.Modified.IsZero() || note.ModificationDate.IsZero() {
		t.Error("Note modification timestamps should be set from metadata")
	}

	// Verify both timestamp fields are synchronized
	if !note.Created.Equal(note.CreationDate) {
		t.Errorf("Created and CreationDate should be equal: %v != %v", note.Created, note.CreationDate)
	}

	if !note.Modified.Equal(note.ModificationDate) {
		t.Errorf("Modified and ModificationDate should be equal: %v != %v", note.Modified, note.ModificationDate)
	}
}

// TestCreateNoteWithSpecialCharacters tests escaping in note creation with metadata
func TestCreateNoteWithSpecialCharacters(t *testing.T) {
	metadataResponse := `{id:"x-coredata://67890", name:"Note with \"quotes\"", creation date:date "Monday, January 1, 2024 at 10:00:00 AM", modification date:date "Monday, January 1, 2024 at 10:00:00 AM", container:"Notes", shared:false, password protected:false}`

	executor := &SequentialMockExecutor{
		responses: []struct {
			stdout string
			stderr string
			err    error
		}{
			{stdout: "note created", stderr: "", err: nil},
			{stdout: metadataResponse, stderr: "", err: nil},
		},
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

	// Verify metadata was populated
	if note.ID == "" {
		t.Error("Note ID should be populated from metadata")
	}

	if note.Folder == "" {
		t.Error("Note folder should be populated from metadata")
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

// TestSearchNotes tests successful note search with full metadata
func TestSearchNotes(t *testing.T) {
	executor := &SequentialMockExecutor{
		responses: []struct {
			stdout string
			stderr string
			err    error
		}{
			// Search response
			{stdout: "Meeting Notes|||Project Ideas|||Random Thoughts", stderr: "", err: nil},
			// GetNoteMetadata responses for each note
			{stdout: `{id:"x-coredata://1", name:"Meeting Notes", creation date:date "Monday, January 1, 2024 at 10:00:00 AM", modification date:date "Monday, January 1, 2024 at 11:00:00 AM", container:"Work", shared:false, password protected:false}`, stderr: "", err: nil},
			{stdout: `{id:"x-coredata://2", name:"Project Ideas", creation date:date "Monday, January 1, 2024 at 12:00:00 PM", modification date:date "Monday, January 1, 2024 at 1:00:00 PM", container:"Personal", shared:true, password protected:false}`, stderr: "", err: nil},
			{stdout: `{id:"x-coredata://3", name:"Random Thoughts", creation date:date "Monday, January 1, 2024 at 2:00:00 PM", modification date:date "Monday, January 1, 2024 at 3:00:00 PM", container:"Notes", shared:false, password protected:true}`, stderr: "", err: nil},
		},
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

	// Verify each note has full metadata
	expectedData := []struct {
		title             string
		folder            string
		shared            bool
		passwordProtected bool
	}{
		{"Meeting Notes", "Work", false, false},
		{"Project Ideas", "Personal", true, false},
		{"Random Thoughts", "Notes", false, true},
	}

	for i, note := range notes {
		expected := expectedData[i]

		if note.Title != expected.title {
			t.Errorf("Note %d title = %q, want %q", i, note.Title, expected.title)
		}

		// Content should still be empty in search results
		if note.Content != "" {
			t.Errorf("Note %d content should be empty, got %q", i, note.Content)
		}

		// Metadata fields should be populated
		if note.ID == "" {
			t.Errorf("Note %d ID should not be empty", i)
		}

		if note.Folder != expected.folder {
			t.Errorf("Note %d folder = %q, want %q", i, note.Folder, expected.folder)
		}

		if note.Shared != expected.shared {
			t.Errorf("Note %d shared = %v, want %v", i, note.Shared, expected.shared)
		}

		if note.PasswordProtected != expected.passwordProtected {
			t.Errorf("Note %d password_protected = %v, want %v", i, note.PasswordProtected, expected.passwordProtected)
		}

		// Timestamps should be set
		if note.Created.IsZero() || note.CreationDate.IsZero() {
			t.Errorf("Note %d creation timestamps should be set", i)
		}

		if note.Modified.IsZero() || note.ModificationDate.IsZero() {
			t.Errorf("Note %d modification timestamps should be set", i)
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

// TestListFolders tests successful folder listing
func TestListFolders(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Notes|||Work|||Personal",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	folders, err := service.ListFolders(ctx)
	if err != nil {
		t.Fatalf("ListFolders failed: %v", err)
	}

	if len(folders) != 3 {
		t.Fatalf("Expected 3 folders, got %d", len(folders))
	}

	expectedFolders := []string{"Notes", "Work", "Personal"}
	for i, folder := range folders {
		if folder != expectedFolders[i] {
			t.Errorf("Folder %d = %q, want %q", i, folder, expectedFolders[i])
		}
	}
}

// TestListFoldersEmpty tests empty folder list
func TestListFoldersEmpty(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	folders, err := service.ListFolders(ctx)
	if err != nil {
		t.Fatalf("ListFolders failed: %v", err)
	}

	if len(folders) != 0 {
		t.Errorf("Expected empty result, got %d folders", len(folders))
	}
}

// TestListFoldersWithError tests error handling in folder listing
func TestListFoldersWithError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "Apple Notes app not running",
		err:    ErrNotesAppNotRunning,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	folders, err := service.ListFolders(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify error is detected correctly (unwrap the wrapped error)
	if !strings.Contains(err.Error(), "Apple Notes app not running") {
		t.Errorf("Expected error containing 'Apple Notes app not running', got %v", err)
	}

	// Should return empty folders on error
	if len(folders) != 0 {
		t.Errorf("Expected empty folders on error, got %d", len(folders))
	}
}

// TestGetNoteMetadata tests retrieval of full note metadata including dates, folder, and sharing info
func TestGetNoteMetadata(t *testing.T) {
	// AppleScript returns metadata as JSON-like structured output
	appleScriptOutput := `{id:"x-coredata://12345", name:"Test Note", creation date:date "Monday, January 1, 2024 at 10:00:00 AM", modification date:date "Monday, January 15, 2024 at 3:30:00 PM", container:"Work", shared:true, password protected:false}`

	executor := &MockExecutor{
		stdout: appleScriptOutput,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	note, err := service.GetNoteMetadata(ctx, "Test Note")
	if err != nil {
		t.Fatalf("GetNoteMetadata failed: %v", err)
	}

	if note.ID == "" {
		t.Error("Expected note ID to be populated")
	}
	if note.Title != "Test Note" {
		t.Errorf("Title = %q, want %q", note.Title, "Test Note")
	}
	if note.Folder != "Work" {
		t.Errorf("Folder = %q, want %q", note.Folder, "Work")
	}
	if !note.Shared {
		t.Error("Expected Shared to be true")
	}
	if note.PasswordProtected {
		t.Error("Expected PasswordProtected to be false")
	}

	// Verify timestamps are synchronized
	if note.Created.IsZero() {
		t.Error("Created timestamp should not be zero")
	}
	if note.CreationDate.IsZero() {
		t.Error("CreationDate timestamp should not be zero")
	}
	if !note.Created.Equal(note.CreationDate) {
		t.Errorf("Created (%v) and CreationDate (%v) should be equal", note.Created, note.CreationDate)
	}

	if note.Modified.IsZero() {
		t.Error("Modified timestamp should not be zero")
	}
	if note.ModificationDate.IsZero() {
		t.Error("ModificationDate timestamp should not be zero")
	}
	if !note.Modified.Equal(note.ModificationDate) {
		t.Errorf("Modified (%v) and ModificationDate (%v) should be equal", note.Modified, note.ModificationDate)
	}
}

// TestGetNoteMetadataNotFound tests note not found error for metadata retrieval
func TestGetNoteMetadataNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.GetNoteMetadata(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestParseAppleScriptDate tests AppleScript date parsing
func TestParseAppleScriptDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "standard AppleScript date",
			input:       "date \"Monday, January 1, 2024 at 10:00:00 AM\"",
			shouldError: false,
		},
		{
			name:        "AppleScript date without prefix",
			input:       "Monday, January 1, 2024 at 10:00:00 AM",
			shouldError: false,
		},
		{
			name:        "invalid date format",
			input:       "not a date",
			shouldError: true,
		},
	}

	service := NewAppleNotesService(&MockExecutor{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseAppleScriptDate(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result.IsZero() {
					t.Errorf("Expected non-zero time for input %q", tt.input)
				}
			}
		})
	}
}

// TestCreateFolder tests creating a folder at root level
func TestCreateFolder(t *testing.T) {
	executor := &MockExecutor{
		stdout: "folder created",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.CreateFolder(ctx, "New Folder", "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}
}

// TestCreateFolderNested tests creating a folder within a parent folder
func TestCreateFolderNested(t *testing.T) {
	executor := &MockExecutor{
		stdout: "folder created",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.CreateFolder(ctx, "Subfolder", "Work")
	if err != nil {
		t.Fatalf("CreateFolder (nested) failed: %v", err)
	}
}

// TestCreateFolderParentNotFound tests error when parent folder doesn't exist
func TestCreateFolderParentNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "folder 'NonExistent' not found",
		err:    ErrFolderNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.CreateFolder(ctx, "Subfolder", "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "folder not found") {
		t.Errorf("Expected error containing 'folder not found', got %v", err)
	}
}

// TestMoveNote tests moving a note to a different folder
func TestMoveNote(t *testing.T) {
	executor := &MockExecutor{
		stdout: "note moved",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.MoveNote(ctx, "Test Note", "Archive")
	if err != nil {
		t.Fatalf("MoveNote failed: %v", err)
	}
}

// TestMoveNoteNotFound tests error when note doesn't exist
func TestMoveNoteNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.MoveNote(ctx, "NonExistent", "Archive")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestMoveNoteFolderNotFound tests error when target folder doesn't exist
func TestMoveNoteFolderNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "folder 'NonExistent' not found",
		err:    ErrFolderNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	err := service.MoveNote(ctx, "Test Note", "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "folder not found") {
		t.Errorf("Expected error containing 'folder not found', got %v", err)
	}
}

// TestGetFolderHierarchy tests retrieval of folder hierarchy
func TestGetFolderHierarchy(t *testing.T) {
	// AppleScript returns nested folder structure with the actual format from GetFolderHierarchy
	appleScriptOutput := `{name:"iCloud", shared:false, noteCount:0, children:{{name:"Work", shared:false, noteCount:3, children:{{name:"Projects", shared:false, noteCount:2, children:{}}}}, {name:"Personal", shared:true, noteCount:5, children:{}}}}`

	executor := &MockExecutor{
		stdout: appleScriptOutput,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	hierarchy, err := service.GetFolderHierarchy(ctx)
	if err != nil {
		t.Fatalf("GetFolderHierarchy failed: %v", err)
	}

	// Verify root node
	if hierarchy.Name != "iCloud" {
		t.Errorf("Expected hierarchy name 'iCloud', got '%s'", hierarchy.Name)
	}
	if hierarchy.Shared {
		t.Error("Expected root to not be shared")
	}
	if hierarchy.NoteCount != 0 {
		t.Errorf("Expected root noteCount 0, got %d", hierarchy.NoteCount)
	}

	// Verify children were parsed
	if len(hierarchy.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(hierarchy.Children))
	}

	// Verify first child (Work)
	work := hierarchy.Children[0]
	if work.Name != "Work" {
		t.Errorf("Expected first child name 'Work', got '%s'", work.Name)
	}
	if work.Shared {
		t.Error("Expected Work folder to not be shared")
	}
	if work.NoteCount != 3 {
		t.Errorf("Expected Work noteCount 3, got %d", work.NoteCount)
	}
	if len(work.Children) != 1 {
		t.Fatalf("Expected Work to have 1 child, got %d", len(work.Children))
	}

	// Verify nested child (Projects)
	projects := work.Children[0]
	if projects.Name != "Projects" {
		t.Errorf("Expected nested child name 'Projects', got '%s'", projects.Name)
	}
	if projects.NoteCount != 2 {
		t.Errorf("Expected Projects noteCount 2, got %d", projects.NoteCount)
	}
	if len(projects.Children) != 0 {
		t.Errorf("Expected Projects to have 0 children, got %d", len(projects.Children))
	}

	// Verify second child (Personal)
	personal := hierarchy.Children[1]
	if personal.Name != "Personal" {
		t.Errorf("Expected second child name 'Personal', got '%s'", personal.Name)
	}
	if !personal.Shared {
		t.Error("Expected Personal folder to be shared")
	}
	if personal.NoteCount != 5 {
		t.Errorf("Expected Personal noteCount 5, got %d", personal.NoteCount)
	}
	if len(personal.Children) != 0 {
		t.Errorf("Expected Personal to have 0 children, got %d", len(personal.Children))
	}
}

// TestGetNoteAttachments tests retrieval of attachments for a note
func TestGetNoteAttachments(t *testing.T) {
	// AppleScript returns attachment list
	appleScriptOutput := `{id:"x-coredata://att1", name:"document.pdf", contents:"file:///Users/test/document.pdf", creation date:date "Monday, January 1, 2024 at 10:00:00 AM", modification date:date "Monday, January 15, 2024 at 3:30:00 PM"}
{id:"x-coredata://att2", name:"image.png", contents:"file:///Users/test/image.png", creation date:date "Monday, January 2, 2024 at 11:00:00 AM", modification date:date "Monday, January 16, 2024 at 4:30:00 PM"}`

	executor := &MockExecutor{
		stdout: appleScriptOutput,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	attachments, err := service.GetNoteAttachments(ctx, "Test Note")
	if err != nil {
		t.Fatalf("GetNoteAttachments failed: %v", err)
	}

	if len(attachments) == 0 {
		t.Fatal("Expected attachments to be returned")
	}

	// Check first attachment
	att := attachments[0]
	if att.ID == "" {
		t.Error("Expected attachment ID to be populated")
	}
	if att.Name == "" {
		t.Error("Expected attachment name to be populated")
	}
	if att.FilePath == "" {
		t.Error("Expected attachment file path to be populated")
	}
	if att.CreationDate.IsZero() {
		t.Error("Expected attachment creation date to be populated")
	}
	if att.ModificationDate.IsZero() {
		t.Error("Expected attachment modification date to be populated")
	}
}

// TestGetNoteAttachmentsNoAttachments tests note with no attachments
func TestGetNoteAttachmentsNoAttachments(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	attachments, err := service.GetNoteAttachments(ctx, "Test Note")
	if err != nil {
		t.Fatalf("GetNoteAttachments failed: %v", err)
	}

	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments, got %d", len(attachments))
	}
}

// TestGetNoteAttachmentsNoteNotFound tests error when note doesn't exist
func TestGetNoteAttachmentsNoteNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.GetNoteAttachments(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestSearchNotesAdvanced_TitleOnly tests searching in title only
func TestSearchNotesAdvanced_TitleOnly(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Meeting Notes|||Project Ideas",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "meeting",
		SearchIn: "title",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}

	expectedTitles := []string{"Meeting Notes", "Project Ideas"}
	for i, note := range notes {
		if note.Title != expectedTitles[i] {
			t.Errorf("Note %d title = %q, want %q", i, note.Title, expectedTitles[i])
		}
	}
}

// TestSearchNotesAdvanced_BodyOnly tests searching in body only
func TestSearchNotesAdvanced_BodyOnly(t *testing.T) {
	// Mock returns list of note titles that match body search
	executor := &MockExecutor{
		stdout: "Design Doc|||Implementation Plan",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "architecture",
		SearchIn: "body",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}

	expectedTitles := []string{"Design Doc", "Implementation Plan"}
	for i, note := range notes {
		if note.Title != expectedTitles[i] {
			t.Errorf("Note %d title = %q, want %q", i, note.Title, expectedTitles[i])
		}
	}
}

// TestSearchNotesAdvanced_Both tests searching in both title and body
func TestSearchNotesAdvanced_Both(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Meeting Notes|||Design Doc|||Project Ideas",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "project",
		SearchIn: "both",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 3 {
		t.Fatalf("Expected 3 notes, got %d", len(notes))
	}
}

// TestSearchNotesAdvanced_FolderFilter tests filtering by folder
func TestSearchNotesAdvanced_FolderFilter(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Work Note 1|||Work Note 2",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "note",
		SearchIn: "title",
		Folder:   "Work",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}
}

// TestSearchNotesAdvanced_DateRangeFilter tests filtering by date range
func TestSearchNotesAdvanced_DateRangeFilter(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Recent Note 1|||Recent Note 2",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := SearchOptions{
		Query:    "note",
		SearchIn: "title",
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}
}

// TestSearchNotesAdvanced_CombinedFilters tests folder and date filters together
func TestSearchNotesAdvanced_CombinedFilters(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Work Note Q1",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	opts := SearchOptions{
		Query:    "note",
		SearchIn: "title",
		Folder:   "Work",
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	if notes[0].Title != "Work Note Q1" {
		t.Errorf("Note title = %q, want %q", notes[0].Title, "Work Note Q1")
	}
}

// TestSearchNotesAdvanced_BodySearchWithFilters tests body search with folder/date filters applied first
func TestSearchNotesAdvanced_BodySearchWithFilters(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Filtered Note",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	opts := SearchOptions{
		Query:    "architecture",
		SearchIn: "body",
		Folder:   "Work",
		DateFrom: &dateFrom,
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}
}

// TestSearchNotesAdvanced_InvalidSearchIn tests error handling for invalid SearchIn value
func TestSearchNotesAdvanced_InvalidSearchIn(t *testing.T) {
	executor := &MockExecutor{}
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "test",
		SearchIn: "invalid",
	}

	_, err := service.SearchNotesAdvanced(ctx, opts)
	if err == nil {
		t.Fatal("Expected error for invalid SearchIn value, got nil")
	}

	if !strings.Contains(err.Error(), "invalid SearchIn value") {
		t.Errorf("Expected error containing 'invalid SearchIn value', got %v", err)
	}
}

// TestSearchNotesAdvanced_EmptySearchIn tests default to title search when SearchIn is empty
func TestSearchNotesAdvanced_EmptySearchIn(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Note 1|||Note 2",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "note",
		SearchIn: "", // Empty should default to title search
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}
}

// TestSearchNotesAdvanced_TitlesWithCommas tests note titles containing commas are parsed correctly
func TestSearchNotesAdvanced_TitlesWithCommas(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Design, Implementation, Testing|||Meeting Notes, Q1 2024|||Budget: Revenue, Expenses, Profit",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "test",
		SearchIn: "title",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 3 {
		t.Fatalf("Expected 3 notes, got %d", len(notes))
	}

	expectedTitles := []string{
		"Design, Implementation, Testing",
		"Meeting Notes, Q1 2024",
		"Budget: Revenue, Expenses, Profit",
	}
	for i, note := range notes {
		if note.Title != expectedTitles[i] {
			t.Errorf("Note %d title = %q, want %q", i, note.Title, expectedTitles[i])
		}
	}
}

// TestSearchNotesAdvanced_EmptyResults tests empty search results
func TestSearchNotesAdvanced_EmptyResults(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "nonexistent",
		SearchIn: "title",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Expected empty result, got %d notes", len(notes))
	}
}

// TestSearchNotesAdvanced_ExecutorError tests error handling from executor
func TestSearchNotesAdvanced_ExecutorError(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "Apple Notes app not running",
		err:    ErrNotesAppNotRunning,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	opts := SearchOptions{
		Query:    "test",
		SearchIn: "title",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "Apple Notes app not running") {
		t.Errorf("Expected error containing 'Apple Notes app not running', got %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Expected empty notes on error, got %d", len(notes))
	}
}

// TestGetAttachmentContent tests successful retrieval of attachment content
func TestGetAttachmentContent(t *testing.T) {
	// Create a temporary file to simulate an attachment
	tmpFile := t.TempDir() + "/test_attachment.txt"
	expectedContent := []byte("This is test attachment content")
	if err := writeTestFile(tmpFile, expectedContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	content, err := service.GetAttachmentContent(ctx, tmpFile, 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachmentContent failed: %v", err)
	}

	if string(content) != string(expectedContent) {
		t.Errorf("Content = %q, want %q", string(content), string(expectedContent))
	}
}

// TestGetAttachmentContentFileNotFound tests error when attachment file doesn't exist
func TestGetAttachmentContentFileNotFound(t *testing.T) {
	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	_, err := service.GetAttachmentContent(ctx, "/nonexistent/path/file.txt", 10*1024*1024)
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read attachment file") {
		t.Errorf("Expected error containing 'failed to read attachment file', got %v", err)
	}
}

// TestGetAttachmentContentExceedsMaxSize tests error when file exceeds maxSize limit
func TestGetAttachmentContentExceedsMaxSize(t *testing.T) {
	// Create a file larger than the maxSize limit
	tmpFile := t.TempDir() + "/large_attachment.txt"
	largeContent := make([]byte, 1024) // 1KB file
	if err := writeTestFile(tmpFile, largeContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	// Set maxSize to 512 bytes, which is smaller than the 1KB file
	_, err := service.GetAttachmentContent(ctx, tmpFile, 512)
	if err == nil {
		t.Fatal("Expected error for file exceeding maxSize, got nil")
	}

	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("Expected error containing 'exceeds maximum size', got %v", err)
	}
}

// TestGetAttachmentContentEmptyFile tests reading an empty file
func TestGetAttachmentContentEmptyFile(t *testing.T) {
	tmpFile := t.TempDir() + "/empty_attachment.txt"
	if err := writeTestFile(tmpFile, []byte{}); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	content, err := service.GetAttachmentContent(ctx, tmpFile, 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachmentContent failed for empty file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("Expected empty content, got %d bytes", len(content))
	}
}

// TestGetAttachmentContentBinaryData tests reading binary data (e.g., image)
func TestGetAttachmentContentBinaryData(t *testing.T) {
	tmpFile := t.TempDir() + "/binary_attachment.bin"
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
	if err := writeTestFile(tmpFile, binaryContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	content, err := service.GetAttachmentContent(ctx, tmpFile, 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachmentContent failed for binary file: %v", err)
	}

	if !bytesEqual(content, binaryContent) {
		t.Errorf("Binary content mismatch")
	}
}

// TestGetAttachmentContentDefaultMaxSize tests default maxSize parameter
func TestGetAttachmentContentDefaultMaxSize(t *testing.T) {
	tmpFile := t.TempDir() + "/test_attachment.txt"
	expectedContent := []byte("Test content")
	if err := writeTestFile(tmpFile, expectedContent); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewAppleNotesService(&MockExecutor{})
	ctx := context.Background()

	// Call with 10MB default maxSize (standard default from design)
	content, err := service.GetAttachmentContent(ctx, tmpFile, 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachmentContent failed: %v", err)
	}

	if string(content) != string(expectedContent) {
		t.Errorf("Content = %q, want %q", string(content), string(expectedContent))
	}
}

// Helper function to write test files
func writeTestFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// Helper function to compare byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestExportNoteText tests successful plain text export
func TestExportNoteText(t *testing.T) {
	expectedText := "This is plain text content without HTML formatting"
	executor := &MockExecutor{
		stdout: expectedText,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	text, err := service.ExportNoteText(ctx, "Test Note")
	if err != nil {
		t.Fatalf("ExportNoteText failed: %v", err)
	}

	if text != expectedText {
		t.Errorf("Text = %q, want %q", text, expectedText)
	}
}

// TestExportNoteTextNotFound tests error when note doesn't exist
func TestExportNoteTextNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.ExportNoteText(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestExportNoteTextWithSpecialCharacters tests text export with special characters
func TestExportNoteTextWithSpecialCharacters(t *testing.T) {
	expectedText := `Text with "quotes" and special chars: <>&`
	executor := &MockExecutor{
		stdout: expectedText,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	text, err := service.ExportNoteText(ctx, `Note with "quotes"`)
	if err != nil {
		t.Fatalf("ExportNoteText failed: %v", err)
	}

	if text != expectedText {
		t.Errorf("Text = %q, want %q", text, expectedText)
	}
}

// TestExportNoteMarkdown tests successful markdown export
func TestExportNoteMarkdown(t *testing.T) {
	// HTML body from AppleScript
	htmlBody := "<div>Test note with <b>bold</b> and <i>italic</i> text</div>"
	executor := &MockExecutor{
		stdout: htmlBody,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	markdown, err := service.ExportNoteMarkdown(ctx, "Test Note")
	if err != nil {
		t.Fatalf("ExportNoteMarkdown failed: %v", err)
	}

	// Verify markdown conversion happened
	if markdown == "" {
		t.Error("Expected non-empty markdown")
	}

	// Basic conversion should handle bold and italic
	if !strings.Contains(markdown, "bold") {
		t.Error("Expected markdown to contain 'bold' text")
	}
}

// TestExportNoteMarkdownNotFound tests error when note doesn't exist
func TestExportNoteMarkdownNotFound(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "note 'NonExistent' not found",
		err:    ErrNoteNotFound,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	_, err := service.ExportNoteMarkdown(ctx, "NonExistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "note not found") {
		t.Errorf("Expected error containing 'note not found', got %v", err)
	}
}

// TestExportNoteMarkdownComplexHTML tests markdown conversion with various HTML elements
func TestExportNoteMarkdownComplexHTML(t *testing.T) {
	htmlBody := `<div>
		<h1>Title</h1>
		<p>Paragraph with <b>bold</b>, <i>italic</i>, and <a href="https://example.com">link</a></p>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
	</div>`

	executor := &MockExecutor{
		stdout: htmlBody,
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	markdown, err := service.ExportNoteMarkdown(ctx, "Complex Note")
	if err != nil {
		t.Fatalf("ExportNoteMarkdown failed: %v", err)
	}

	// Verify markdown conversion happened
	if markdown == "" {
		t.Error("Expected non-empty markdown")
	}

	// Should contain the text content at minimum
	if !strings.Contains(markdown, "Title") {
		t.Error("Expected markdown to contain 'Title'")
	}
}

// TestExportNoteMarkdownEmptyBody tests markdown export with empty body
func TestExportNoteMarkdownEmptyBody(t *testing.T) {
	executor := &MockExecutor{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	service := NewAppleNotesService(executor)
	ctx := context.Background()

	markdown, err := service.ExportNoteMarkdown(ctx, "Empty Note")
	if err != nil {
		t.Fatalf("ExportNoteMarkdown failed: %v", err)
	}

	// Empty body should return empty markdown
	if markdown != "" {
		t.Errorf("Expected empty markdown for empty body, got %q", markdown)
	}
}
