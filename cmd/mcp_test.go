// ABOUTME: Tests for the MCP server command
// ABOUTME: Verifies tool registration and error handling

package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/harper/notes-mcp/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockNotesService is a simple mock for testing tool handlers
type mockNotesService struct {
	createNote           func(ctx context.Context, title, content string, tags []string) (*services.Note, error)
	searchNotes          func(ctx context.Context, query string) ([]services.Note, error)
	searchNotesAdvanced  func(ctx context.Context, opts services.SearchOptions) ([]services.Note, error)
	getNoteContent       func(ctx context.Context, title string) (string, error)
	getNoteMetadata      func(ctx context.Context, title string) (*services.Note, error)
	updateNote           func(ctx context.Context, title, content string) error
	deleteNote           func(ctx context.Context, title string) error
	listFolders          func(ctx context.Context) ([]string, error)
	getRecentNotes       func(ctx context.Context, limit int) ([]services.Note, error)
	getNotesInFolder     func(ctx context.Context, folder string) ([]services.Note, error)
	createFolder         func(ctx context.Context, name string, parentFolder string) error
	moveNote             func(ctx context.Context, noteTitle string, targetFolder string) error
	getFolderHierarchy   func(ctx context.Context) (*services.FolderNode, error)
	getNoteAttachments   func(ctx context.Context, noteTitle string) ([]services.Attachment, error)
	getAttachmentContent func(ctx context.Context, filePath string, maxSize int64) ([]byte, error)
	exportNoteMarkdown   func(ctx context.Context, noteTitle string) (string, error)
	exportNoteText       func(ctx context.Context, noteTitle string) (string, error)
}

func (m *mockNotesService) CreateNote(ctx context.Context, title, content string, tags []string) (*services.Note, error) {
	if m.createNote != nil {
		return m.createNote(ctx, title, content, tags)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) SearchNotes(ctx context.Context, query string) ([]services.Note, error) {
	if m.searchNotes != nil {
		return m.searchNotes(ctx, query)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) SearchNotesAdvanced(ctx context.Context, opts services.SearchOptions) ([]services.Note, error) {
	if m.searchNotesAdvanced != nil {
		return m.searchNotesAdvanced(ctx, opts)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) GetNoteContent(ctx context.Context, title string) (string, error) {
	if m.getNoteContent != nil {
		return m.getNoteContent(ctx, title)
	}
	return "", errors.New("not implemented")
}

func (m *mockNotesService) GetNoteMetadata(ctx context.Context, title string) (*services.Note, error) {
	if m.getNoteMetadata != nil {
		return m.getNoteMetadata(ctx, title)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) UpdateNote(ctx context.Context, title, content string) error {
	if m.updateNote != nil {
		return m.updateNote(ctx, title, content)
	}
	return errors.New("not implemented")
}

func (m *mockNotesService) DeleteNote(ctx context.Context, title string) error {
	if m.deleteNote != nil {
		return m.deleteNote(ctx, title)
	}
	return errors.New("not implemented")
}

func (m *mockNotesService) ListFolders(ctx context.Context) ([]string, error) {
	if m.listFolders != nil {
		return m.listFolders(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) GetRecentNotes(ctx context.Context, limit int) ([]services.Note, error) {
	if m.getRecentNotes != nil {
		return m.getRecentNotes(ctx, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) GetNotesInFolder(ctx context.Context, folder string) ([]services.Note, error) {
	if m.getNotesInFolder != nil {
		return m.getNotesInFolder(ctx, folder)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) CreateFolder(ctx context.Context, name string, parentFolder string) error {
	if m.createFolder != nil {
		return m.createFolder(ctx, name, parentFolder)
	}
	return errors.New("not implemented")
}

func (m *mockNotesService) MoveNote(ctx context.Context, noteTitle string, targetFolder string) error {
	if m.moveNote != nil {
		return m.moveNote(ctx, noteTitle, targetFolder)
	}
	return errors.New("not implemented")
}

func (m *mockNotesService) GetFolderHierarchy(ctx context.Context) (*services.FolderNode, error) {
	if m.getFolderHierarchy != nil {
		return m.getFolderHierarchy(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) GetNoteAttachments(ctx context.Context, noteTitle string) ([]services.Attachment, error) {
	if m.getNoteAttachments != nil {
		return m.getNoteAttachments(ctx, noteTitle)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) GetAttachmentContent(ctx context.Context, filePath string, maxSize int64) ([]byte, error) {
	if m.getAttachmentContent != nil {
		return m.getAttachmentContent(ctx, filePath, maxSize)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNotesService) ExportNoteMarkdown(ctx context.Context, noteTitle string) (string, error) {
	if m.exportNoteMarkdown != nil {
		return m.exportNoteMarkdown(ctx, noteTitle)
	}
	return "", errors.New("not implemented")
}

func (m *mockNotesService) ExportNoteText(ctx context.Context, noteTitle string) (string, error) {
	if m.exportNoteText != nil {
		return m.exportNoteText(ctx, noteTitle)
	}
	return "", errors.New("not implemented")
}

// Test that createErrorResult properly converts service errors to user-friendly messages
func TestCreateErrorResult(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedText    string
		expectedIsError bool
	}{
		{
			name:            "note not found",
			err:             services.ErrNoteNotFound,
			expectedText:    "Note not found in Apple Notes. Please check the title and try again.",
			expectedIsError: true,
		},
		{
			name:            "notes app not running",
			err:             services.ErrNotesAppNotRunning,
			expectedText:    "Apple Notes app is not running. Please open the Notes app and try again.",
			expectedIsError: true,
		},
		{
			name:            "permission denied",
			err:             services.ErrPermissionDenied,
			expectedText:    "Permission denied to access Notes. Please grant access in System Preferences > Privacy & Security > Automation.",
			expectedIsError: true,
		},
		{
			name:            "script timeout",
			err:             services.ErrScriptTimeout,
			expectedText:    "Apple Notes is not responding (timeout after 10 seconds). Please try again.",
			expectedIsError: true,
		},
		{
			name:            "invalid input",
			err:             services.ErrInvalidInput,
			expectedText:    "Invalid input: invalid input parameters",
			expectedIsError: true,
		},
		{
			name:            "unknown error",
			err:             errors.New("something went wrong"),
			expectedText:    "An error occurred: something went wrong",
			expectedIsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createErrorResult(tt.err)

			if result.IsError != tt.expectedIsError {
				t.Errorf("expected IsError=%v, got %v", tt.expectedIsError, result.IsError)
			}

			if len(result.Content) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(result.Content))
			}

			// Type assertion to TextContent
			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("expected TextContent, got %T", result.Content[0])
			}

			if textContent.Text != tt.expectedText {
				t.Errorf("expected text %q, got %q", tt.expectedText, textContent.Text)
			}
		})
	}
}

// TestMockServiceIntegration verifies the service interface works correctly with handlers
func TestMockServiceIntegration(t *testing.T) {
	// This test verifies that our mock service implementation is compatible
	// with the NotesService interface
	var _ services.NotesService = (*mockNotesService)(nil)

	mock := &mockNotesService{
		createNote: func(ctx context.Context, title, content string, tags []string) (*services.Note, error) {
			return &services.Note{
				ID:       "123",
				Title:    title,
				Content:  content,
				Tags:     tags,
				Created:  time.Now(),
				Modified: time.Now(),
			}, nil
		},
	}

	note, err := mock.CreateNote(context.Background(), "Test", "Content", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", note.Title)
	}
}

// TestNoteResourceHandler tests the note:///{title} resource handler
func TestNoteResourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uri            string
		mockContent    string
		mockError      error
		expectError    bool
		expectNotFound bool
	}{
		{
			name:        "successful note retrieval",
			uri:         "note:///Test%20Note",
			mockContent: "<div>Test content</div>",
			mockError:   nil,
			expectError: false,
		},
		{
			name:           "note not found",
			uri:            "note:///Nonexistent",
			mockError:      services.ErrNoteNotFound,
			expectError:    true,
			expectNotFound: true,
		},
		{
			name:        "invalid URI format",
			uri:         "invalid://test",
			expectError: true,
		},
		{
			name:        "empty title",
			uri:         "note:///",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{
				getNoteContent: func(ctx context.Context, title string) (string, error) {
					if tt.mockError != nil {
						return "", tt.mockError
					}
					return tt.mockContent, nil
				},
			}

			handler := createNoteResourceHandler(mock)
			result, err := handler(context.Background(), &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{URI: tt.uri},
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Contents) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(result.Contents))
			}

			if result.Contents[0].Text != tt.mockContent {
				t.Errorf("expected content %q, got %q", tt.mockContent, result.Contents[0].Text)
			}

			if result.Contents[0].MIMEType != "text/html" {
				t.Errorf("expected MIME type 'text/html', got %q", result.Contents[0].MIMEType)
			}
		})
	}
}

// TestRecentNotesResourceHandler tests the notes:///recent resource handler
func TestRecentNotesResourceHandler(t *testing.T) {
	tests := []struct {
		name        string
		mockNotes   []services.Note
		mockError   error
		expectError bool
		expectText  string
	}{
		{
			name: "successful recent notes retrieval",
			mockNotes: []services.Note{
				{Title: "Note 1"},
				{Title: "Note 2"},
				{Title: "Note 3"},
			},
			expectText: "Note 1\nNote 2\nNote 3",
		},
		{
			name:       "no notes found",
			mockNotes:  []services.Note{},
			expectText: "No notes found.",
		},
		{
			name:        "service error",
			mockError:   errors.New("service error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{
				getRecentNotes: func(ctx context.Context, limit int) ([]services.Note, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockNotes, nil
				},
			}

			handler := createRecentNotesResourceHandler(mock)
			result, err := handler(context.Background(), &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{URI: "notes:///recent"},
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Contents[0].Text != tt.expectText {
				t.Errorf("expected text %q, got %q", tt.expectText, result.Contents[0].Text)
			}
		})
	}
}

// TestSearchNotesResourceHandler tests the notes:///search/{query} resource handler
func TestSearchNotesResourceHandler(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		mockNotes   []services.Note
		mockError   error
		expectError bool
		expectText  string
	}{
		{
			name: "successful search",
			uri:  "notes:///search/test",
			mockNotes: []services.Note{
				{Title: "Test Note 1"},
				{Title: "Test Note 2"},
			},
			expectText: "Test Note 1\nTest Note 2",
		},
		{
			name:       "no results",
			uri:        "notes:///search/nonexistent",
			mockNotes:  []services.Note{},
			expectText: "No notes found matching the query.",
		},
		{
			name:        "empty query",
			uri:         "notes:///search/",
			expectError: true,
		},
		{
			name:        "invalid URI",
			uri:         "invalid:///search/test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{
				searchNotes: func(ctx context.Context, query string) ([]services.Note, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockNotes, nil
				},
			}

			handler := createSearchNotesResourceHandler(mock)
			result, err := handler(context.Background(), &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{URI: tt.uri},
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Contents[0].Text != tt.expectText {
				t.Errorf("expected text %q, got %q", tt.expectText, result.Contents[0].Text)
			}
		})
	}
}

// TestFolderNotesResourceHandler tests the notes:///folder/{folder} resource handler
func TestFolderNotesResourceHandler(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		mockNotes   []services.Note
		mockError   error
		expectError bool
		expectText  string
	}{
		{
			name: "successful folder notes retrieval",
			uri:  "notes:///folder/Work",
			mockNotes: []services.Note{
				{Title: "Work Note 1"},
				{Title: "Work Note 2"},
			},
			expectText: "Work Note 1\nWork Note 2",
		},
		{
			name:       "empty folder",
			uri:        "notes:///folder/Empty",
			mockNotes:  []services.Note{},
			expectText: "No notes found in folder 'Empty'.",
		},
		{
			name:        "empty folder name",
			uri:         "notes:///folder/",
			expectError: true,
		},
		{
			name:        "invalid URI",
			uri:         "invalid:///folder/Work",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{
				getNotesInFolder: func(ctx context.Context, folder string) ([]services.Note, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockNotes, nil
				},
			}

			handler := createFolderNotesResourceHandler(mock)
			result, err := handler(context.Background(), &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{URI: tt.uri},
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Contents[0].Text != tt.expectText {
				t.Errorf("expected text %q, got %q", tt.expectText, result.Contents[0].Text)
			}
		})
	}
}

// TestDailyReviewPrompt tests the daily-review prompt handler
func TestDailyReviewPrompt(t *testing.T) {
	mock := &mockNotesService{}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerDailyReviewPrompt(server, mock)

	// The prompt should be registered without error
	// We can't easily test the handler directly without exposing it,
	// but we verify registration succeeds
}

// TestWeeklySummaryPrompt tests the weekly-summary prompt handler
func TestWeeklySummaryPrompt(t *testing.T) {
	tests := []struct {
		name             string
		categories       string
		expectCategories bool
	}{
		{
			name:             "without categories",
			categories:       "",
			expectCategories: false,
		},
		{
			name:             "with categories",
			categories:       "meetings,ideas,todos",
			expectCategories: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{}
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

			registerWeeklySummaryPrompt(server, mock)

			// Verify registration succeeds
		})
	}
}

// TestMeetingPrepPrompt tests the meeting-prep prompt handler
func TestMeetingPrepPrompt(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		attendees string
	}{
		{
			name:      "without attendees",
			topic:     "Project Planning",
			attendees: "",
		},
		{
			name:      "with attendees",
			topic:     "Sprint Review",
			attendees: "Alice,Bob,Carol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{}
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

			registerMeetingPrepPrompt(server, mock)

			// Verify registration succeeds
		})
	}
}

// TestActionItemsPrompt tests the action-items prompt handler
func TestActionItemsPrompt(t *testing.T) {
	tests := []struct {
		name       string
		searchTerm string
		status     string
	}{
		{
			name:       "with default status",
			searchTerm: "TODO",
			status:     "",
		},
		{
			name:       "with open status",
			searchTerm: "action",
			status:     "open",
		},
		{
			name:       "with all status",
			searchTerm: "project",
			status:     "all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{}
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

			registerActionItemsPrompt(server, mock)

			// Verify registration succeeds
		})
	}
}

// TestNoteCleanupPrompt tests the note-cleanup prompt handler
func TestNoteCleanupPrompt(t *testing.T) {
	tests := []struct {
		name            string
		ageThreshold    string
		expectedDefault string
	}{
		{
			name:            "with default threshold",
			ageThreshold:    "",
			expectedDefault: "90",
		},
		{
			name:            "with custom threshold",
			ageThreshold:    "30",
			expectedDefault: "30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{}
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

			registerNoteCleanupPrompt(server, mock)

			// Verify registration succeeds
		})
	}
}

// TestQuickNotePrompt tests the quick-note prompt handler
func TestQuickNotePrompt(t *testing.T) {
	tests := []struct {
		name     string
		noteType string
		title    string
	}{
		{
			name:     "meeting note",
			noteType: "meeting",
			title:    "Team Standup",
		},
		{
			name:     "idea note",
			noteType: "idea",
			title:    "New Feature Concept",
		},
		{
			name:     "todo note",
			noteType: "todo",
			title:    "Weekly Tasks",
		},
		{
			name:     "journal note",
			noteType: "journal",
			title:    "Daily Journal Entry",
		},
		{
			name:     "general note",
			noteType: "general",
			title:    "Random Thoughts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotesService{}
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

			registerQuickNotePrompt(server, mock)

			// Verify registration succeeds
		})
	}
}

// TestRegisterPromptsIntegration tests that all prompts can be registered together
func TestRegisterPromptsIntegration(t *testing.T) {
	mock := &mockNotesService{}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	// This should register all 6 prompts without error
	registerPrompts(server, mock)

	// If we get here without panic, registration succeeded
}

// TestRegisterCreateFolderTool tests the create_folder tool registration
func TestRegisterCreateFolderTool(t *testing.T) {
	mock := &mockNotesService{
		createFolder: func(ctx context.Context, name string, parentFolder string) error {
			return nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerCreateFolderTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterMoveNoteTool tests the move_note tool registration
func TestRegisterMoveNoteTool(t *testing.T) {
	mock := &mockNotesService{
		moveNote: func(ctx context.Context, noteTitle string, targetFolder string) error {
			return nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerMoveNoteTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterGetFolderHierarchyTool tests the get_folder_hierarchy tool registration
func TestRegisterGetFolderHierarchyTool(t *testing.T) {
	mock := &mockNotesService{
		getFolderHierarchy: func(ctx context.Context) (*services.FolderNode, error) {
			return &services.FolderNode{
				Name:      "Root",
				NoteCount: 5,
				Children: []services.FolderNode{
					{Name: "Child1", NoteCount: 2},
					{Name: "Child2", NoteCount: 3},
				},
			}, nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerGetFolderHierarchyTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterSearchNotesAdvancedTool tests the search_notes_advanced tool registration
func TestRegisterSearchNotesAdvancedTool(t *testing.T) {
	mock := &mockNotesService{
		searchNotesAdvanced: func(ctx context.Context, opts services.SearchOptions) ([]services.Note, error) {
			return []services.Note{
				{Title: "Advanced Note 1"},
				{Title: "Advanced Note 2"},
			}, nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerSearchNotesAdvancedTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterGetNoteAttachmentsTool tests the get_note_attachments tool registration
func TestRegisterGetNoteAttachmentsTool(t *testing.T) {
	mock := &mockNotesService{
		getNoteAttachments: func(ctx context.Context, noteTitle string) ([]services.Attachment, error) {
			return []services.Attachment{
				{
					Name:     "file1.pdf",
					FilePath: "/path/to/file1.pdf",
					ID:       "123",
				},
			}, nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerGetNoteAttachmentsTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterGetAttachmentContentTool tests the get_attachment_content tool registration
func TestRegisterGetAttachmentContentTool(t *testing.T) {
	mock := &mockNotesService{
		getAttachmentContent: func(ctx context.Context, filePath string, maxSize int64) ([]byte, error) {
			return []byte("test content"), nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerGetAttachmentContentTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterExportNoteMarkdownTool tests the export_note_markdown tool registration
func TestRegisterExportNoteMarkdownTool(t *testing.T) {
	mock := &mockNotesService{
		exportNoteMarkdown: func(ctx context.Context, noteTitle string) (string, error) {
			return "# Markdown Content", nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerExportNoteMarkdownTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestRegisterExportNoteTextTool tests the export_note_text tool registration
func TestRegisterExportNoteTextTool(t *testing.T) {
	mock := &mockNotesService{
		exportNoteText: func(ctx context.Context, noteTitle string) (string, error) {
			return "Plain text content", nil
		},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	registerExportNoteTextTool(server, mock)
	// If we get here without panic, registration succeeded
}

// TestAllToolsRegistrationIntegration tests that all tools can be registered together
func TestAllToolsRegistrationIntegration(t *testing.T) {
	mock := &mockNotesService{}
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)

	// Register all tools (original 6 + new 8 = 14 total)
	registerCreateNoteTool(server, mock)
	registerSearchNotesTool(server, mock)
	registerGetNoteContentTool(server, mock)
	registerUpdateNoteTool(server, mock)
	registerDeleteNoteTool(server, mock)
	registerListFoldersTool(server, mock)
	registerCreateFolderTool(server, mock)
	registerMoveNoteTool(server, mock)
	registerGetFolderHierarchyTool(server, mock)
	registerSearchNotesAdvancedTool(server, mock)
	registerGetNoteAttachmentsTool(server, mock)
	registerGetAttachmentContentTool(server, mock)
	registerExportNoteMarkdownTool(server, mock)
	registerExportNoteTextTool(server, mock)

	// If we get here without panic, all registrations succeeded
}
