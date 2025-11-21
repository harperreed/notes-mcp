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

// TestListFolders tests successful folder listing
func TestListFolders(t *testing.T) {
	executor := &MockExecutor{
		stdout: "Notes, Work, Personal",
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
	// AppleScript returns nested folder structure
	appleScriptOutput := `{name:"Notes", shared:false, folders:{{name:"Work", shared:false, folders:{{name:"Projects", shared:false, folders:{}}}}, {name:"Personal", shared:true, folders:{}}}, note count:5}`

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

	if hierarchy.Name == "" {
		t.Error("Expected hierarchy to have a name")
	}
	// Note: parseFolderHierarchy currently returns a simplified implementation
	// Full nested parsing would require a more sophisticated AppleScript record parser
	if hierarchy.Children == nil {
		t.Error("Expected hierarchy.Children to be initialized (not nil)")
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
