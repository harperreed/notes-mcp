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
	createNote      func(ctx context.Context, title, content string, tags []string) (*services.Note, error)
	searchNotes     func(ctx context.Context, query string) ([]services.Note, error)
	getNoteContent  func(ctx context.Context, title string) (string, error)
	updateNote      func(ctx context.Context, title, content string) error
	deleteNote      func(ctx context.Context, title string) error
	listFolders     func(ctx context.Context) ([]string, error)
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

func (m *mockNotesService) GetNoteContent(ctx context.Context, title string) (string, error) {
	if m.getNoteContent != nil {
		return m.getNoteContent(ctx, title)
	}
	return "", errors.New("not implemented")
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

// Test that createErrorResult properly converts service errors to user-friendly messages
func TestCreateErrorResult(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedText   string
		expectedIsError bool
	}{
		{
			name:           "note not found",
			err:            services.ErrNoteNotFound,
			expectedText:   "Note not found in Apple Notes. Please check the title and try again.",
			expectedIsError: true,
		},
		{
			name:           "notes app not running",
			err:            services.ErrNotesAppNotRunning,
			expectedText:   "Apple Notes app is not running. Please open the Notes app and try again.",
			expectedIsError: true,
		},
		{
			name:           "permission denied",
			err:            services.ErrPermissionDenied,
			expectedText:   "Permission denied to access Notes. Please grant access in System Preferences > Privacy & Security > Automation.",
			expectedIsError: true,
		},
		{
			name:           "script timeout",
			err:            services.ErrScriptTimeout,
			expectedText:   "Apple Notes is not responding (timeout after 10 seconds). Please try again.",
			expectedIsError: true,
		},
		{
			name:           "invalid input",
			err:            services.ErrInvalidInput,
			expectedText:   "Invalid input: invalid input parameters",
			expectedIsError: true,
		},
		{
			name:           "unknown error",
			err:            errors.New("something went wrong"),
			expectedText:   "An error occurred: something went wrong",
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
