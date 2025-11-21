// ABOUTME: MCP server subcommand that starts the Model Context Protocol server
// ABOUTME: Implements stdio-based MCP server with six tools and four resource types for direct note access

package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/harper/notes-mcp/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server",
	Long:  `Starts the Model Context Protocol server for Apple Notes integration over stdio.`,
	Run:   runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

// Tool input argument structs with JSON schema annotations

type CreateNoteArgs struct {
	Title   string   `json:"title" jsonschema:"The title of the note"`
	Content string   `json:"content" jsonschema:"The content of the note"`
	Tags    []string `json:"tags,omitempty" jsonschema:"Optional tags for the note"`
}

type SearchNotesArgs struct {
	Query string `json:"query" jsonschema:"The search query to find notes by title"`
}

type GetNoteContentArgs struct {
	Title string `json:"title" jsonschema:"The title of the note to retrieve"`
}

type UpdateNoteArgs struct {
	Title   string `json:"title" jsonschema:"The title of the note to update"`
	Content string `json:"content" jsonschema:"The new content for the note"`
}

type DeleteNoteArgs struct {
	Title string `json:"title" jsonschema:"The title of the note to delete"`
}

type CreateFolderArgs struct {
	Name         string `json:"name" jsonschema:"The name of the folder to create"`
	ParentFolder string `json:"parent_folder,omitempty" jsonschema:"Optional parent folder name for nested folders"`
}

type MoveNoteArgs struct {
	NoteTitle    string `json:"note_title" jsonschema:"The title of the note to move"`
	TargetFolder string `json:"target_folder" jsonschema:"The target folder to move the note to"`
}

type SearchNotesAdvancedArgs struct {
	Query    string `json:"query" jsonschema:"The search query"`
	SearchIn string `json:"search_in,omitempty" jsonschema:"Where to search: 'title', 'body', or 'both' (default: 'title')"`
	Folder   string `json:"folder,omitempty" jsonschema:"Optional folder name to limit search scope"`
	DateFrom string `json:"date_from,omitempty" jsonschema:"Optional start date filter (YYYY-MM-DD format)"`
	DateTo   string `json:"date_to,omitempty" jsonschema:"Optional end date filter (YYYY-MM-DD format)"`
}

type GetNoteAttachmentsArgs struct {
	NoteTitle string `json:"note_title" jsonschema:"The title of the note to get attachments from"`
}

type GetAttachmentContentArgs struct {
	FilePath  string `json:"file_path" jsonschema:"The file path of the attachment to retrieve"`
	MaxSizeMB int    `json:"max_size_mb,omitempty" jsonschema:"Maximum file size in MB (default: 10)"`
}

type ExportNoteMarkdownArgs struct {
	NoteTitle string `json:"note_title" jsonschema:"The title of the note to export as markdown"`
}

type ExportNoteTextArgs struct {
	NoteTitle string `json:"note_title" jsonschema:"The title of the note to export as plain text"`
}

// runMCPServer starts the MCP server in stdio mode
func runMCPServer(cmd *cobra.Command, args []string) {
	// Create the notes service
	executor := services.NewOSAScriptExecutor(10 * time.Second)
	notesService := services.NewAppleNotesService(executor)

	// Create the MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "apple-notes-go",
			Version: "1.0.0",
		},
		nil,
	)

	// Register the tools
	registerCreateNoteTool(server, notesService)
	registerSearchNotesTool(server, notesService)
	registerGetNoteContentTool(server, notesService)
	registerUpdateNoteTool(server, notesService)
	registerDeleteNoteTool(server, notesService)
	registerListFoldersTool(server, notesService)
	registerCreateFolderTool(server, notesService)
	registerMoveNoteTool(server, notesService)
	registerGetFolderHierarchyTool(server, notesService)
	registerSearchNotesAdvancedTool(server, notesService)
	registerGetNoteAttachmentsTool(server, notesService)
	registerGetAttachmentContentTool(server, notesService)
	registerExportNoteMarkdownTool(server, notesService)
	registerExportNoteTextTool(server, notesService)

	// Register resources
	registerResources(server, notesService)

	// Register prompts
	registerPrompts(server, notesService)

	// Run the server over stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}

// registerCreateNoteTool registers the create_note tool
func registerCreateNoteTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input CreateNoteArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Title == "" {
			return nil, nil, fmt.Errorf("%w: title is required", services.ErrInvalidInput)
		}
		if input.Content == "" {
			return nil, nil, fmt.Errorf("%w: content is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		note, err := notesService.CreateNote(opCtx, input.Title, input.Content, input.Tags)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Marshal note to JSON for structured output with full metadata
		noteJSON, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return createErrorResult(fmt.Errorf("failed to format note: %w", err)), nil, nil
		}

		// Return success result with full note details
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(noteJSON),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_note",
		Description: "Creates a new note in Apple Notes with the specified title, content, and optional tags. Returns the created note with full metadata including creation/modification dates, folder, and sharing status as JSON.",
	}, handler)
}

// registerSearchNotesTool registers the search_notes tool
func registerSearchNotesTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input SearchNotesArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Query == "" {
			return nil, nil, fmt.Errorf("%w: query is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		notes, err := notesService.SearchNotes(opCtx, input.Query)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Handle empty results
		if len(notes) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "No notes found matching the query.",
					},
				},
			}, nil, nil
		}

		// Limit results to prevent timeouts with large result sets
		totalNotes := len(notes)
		if totalNotes > maxSearchResults {
			notes = notes[:maxSearchResults]
		}

		// Format results with metadata as JSON
		result, err := formatSearchResults(notes, totalNotes)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result with full note metadata
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: result,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_notes",
		Description: "Searches for notes in Apple Notes by title. Returns a list of matching notes with full metadata including creation/modification dates, folder, and sharing status as JSON.",
	}, handler)
}

// registerGetNoteContentTool registers the get_note_content tool
func registerGetNoteContentTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input GetNoteContentArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Title == "" {
			return nil, nil, fmt.Errorf("%w: title is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Get note metadata
		note, err := notesService.GetNoteMetadata(opCtx, input.Title)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Get note content
		content, err := notesService.GetNoteContent(opCtx, input.Title)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Populate content field
		note.Content = content

		// Marshal note to JSON for structured output with full metadata
		noteJSON, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return createErrorResult(fmt.Errorf("failed to format note: %w", err)), nil, nil
		}

		// Return success result with full note details
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(noteJSON),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note_content",
		Description: "Retrieves the full content and metadata of a note from Apple Notes by its title. Returns the note with all fields including creation/modification dates, folder, sharing status, and content as JSON.",
	}, handler)
}

// registerUpdateNoteTool registers the update_note tool
func registerUpdateNoteTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input UpdateNoteArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Title == "" {
			return nil, nil, fmt.Errorf("%w: title is required", services.ErrInvalidInput)
		}
		if input.Content == "" {
			return nil, nil, fmt.Errorf("%w: content is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		err := notesService.UpdateNote(opCtx, input.Title, input.Content)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Note updated: %s", input.Title),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_note",
		Description: "Updates the content of an existing note in Apple Notes by its title. Returns confirmation of note update.",
	}, handler)
}

// registerDeleteNoteTool registers the delete_note tool
func registerDeleteNoteTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input DeleteNoteArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Title == "" {
			return nil, nil, fmt.Errorf("%w: title is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		err := notesService.DeleteNote(opCtx, input.Title)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Note deleted: %s", input.Title),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_note",
		Description: "Deletes a note from Apple Notes by its title. Returns confirmation of note deletion.",
	}, handler)
}

// registerListFoldersTool registers the list_folders tool
func registerListFoldersTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (
		*mcp.CallToolResult, any, error) {

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		folders, err := notesService.ListFolders(opCtx)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Handle empty results
		if len(folders) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "No folders found.",
					},
				},
			}, nil, nil
		}

		// Format the results as newline-separated list of folder names
		result := strings.Join(folders, "\n")

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: result,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_folders",
		Description: "Lists all folders in Apple Notes. Returns a list of folder names, or a message if no folders are found.",
	}, handler)
}

// registerCreateFolderTool registers the create_folder tool
func registerCreateFolderTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input CreateFolderArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Name == "" {
			return nil, nil, fmt.Errorf("%w: name is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		err := notesService.CreateFolder(opCtx, input.Name, input.ParentFolder)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Format success message
		var message string
		if input.ParentFolder == "" {
			message = fmt.Sprintf("Folder created: %s", input.Name)
		} else {
			message = fmt.Sprintf("Folder created: %s (in %s)", input.Name, input.ParentFolder)
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: message,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_folder",
		Description: "Creates a new folder in Apple Notes. Can create at root level or nested under a parent folder.",
	}, handler)
}

// registerMoveNoteTool registers the move_note tool
func registerMoveNoteTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input MoveNoteArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.NoteTitle == "" {
			return nil, nil, fmt.Errorf("%w: note_title is required", services.ErrInvalidInput)
		}
		if input.TargetFolder == "" {
			return nil, nil, fmt.Errorf("%w: target_folder is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		err := notesService.MoveNote(opCtx, input.NoteTitle, input.TargetFolder)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Note '%s' moved to folder '%s'", input.NoteTitle, input.TargetFolder),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_note",
		Description: "Moves a note to a different folder in Apple Notes. Returns confirmation of note movement.",
	}, handler)
}

// registerGetFolderHierarchyTool registers the get_folder_hierarchy tool
func registerGetFolderHierarchyTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (
		*mcp.CallToolResult, any, error) {

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		hierarchy, err := notesService.GetFolderHierarchy(opCtx)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Marshal to JSON for structured output
		hierarchyJSON, err := json.MarshalIndent(hierarchy, "", "  ")
		if err != nil {
			return createErrorResult(fmt.Errorf("failed to format hierarchy: %w", err)), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(hierarchyJSON),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_folder_hierarchy",
		Description: "Retrieves the complete folder hierarchy from Apple Notes with note counts. Returns nested folder structure as JSON.",
	}, handler)
}

// parseDateFilter parses a date string in YYYY-MM-DD format
// Returns nil pointer and nil error when dateStr is empty (valid case for optional dates)
func parseDateFilter(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		var nilTime *time.Time
		return nilTime, nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date format, use YYYY-MM-DD", services.ErrInvalidInput)
	}
	return &t, nil
}

// formatSearchResults formats notes as JSON with optional result limiting message
func formatSearchResults(notes []services.Note, totalNotes int) (string, error) {
	notesJSON, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format results: %w", err)
	}

	result := string(notesJSON)
	if totalNotes > maxSearchResults {
		result = fmt.Sprintf("%s\n\n(Showing first %d of %d matching notes)", result, maxSearchResults, totalNotes)
	}
	return result, nil
}

// registerSearchNotesAdvancedTool registers the search_notes_advanced tool
func registerSearchNotesAdvancedTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input SearchNotesAdvancedArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.Query == "" {
			return nil, nil, fmt.Errorf("%w: query is required", services.ErrInvalidInput)
		}

		// Set defaults
		if input.SearchIn == "" {
			input.SearchIn = "title"
		}

		// Parse date filters
		dateFrom, err := parseDateFilter(input.DateFrom)
		if err != nil {
			return nil, nil, err
		}
		dateTo, err := parseDateFilter(input.DateTo)
		if err != nil {
			return nil, nil, err
		}

		// Create search options
		opts := services.SearchOptions{
			Query:    input.Query,
			SearchIn: input.SearchIn,
			Folder:   input.Folder,
			DateFrom: dateFrom,
			DateTo:   dateTo,
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		notes, err := notesService.SearchNotesAdvanced(opCtx, opts)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Handle empty results
		if len(notes) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "No notes found matching the search criteria.",
					},
				},
			}, nil, nil
		}

		// Limit results to prevent timeouts with large result sets
		totalNotes := len(notes)
		if totalNotes > maxSearchResults {
			notes = notes[:maxSearchResults]
		}

		// Format results with metadata
		result, err := formatSearchResults(notes, totalNotes)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: result,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_notes_advanced",
		Description: "Searches for notes with advanced filters including body search, folder filtering, and date ranges. Returns notes with full metadata as JSON.",
	}, handler)
}

// registerGetNoteAttachmentsTool registers the get_note_attachments tool
func registerGetNoteAttachmentsTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input GetNoteAttachmentsArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.NoteTitle == "" {
			return nil, nil, fmt.Errorf("%w: note_title is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		attachments, err := notesService.GetNoteAttachments(opCtx, input.NoteTitle)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Handle empty results
		if len(attachments) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Note '%s' has no attachments.", input.NoteTitle),
					},
				},
			}, nil, nil
		}

		// Marshal to JSON for structured output
		attachmentsJSON, err := json.MarshalIndent(attachments, "", "  ")
		if err != nil {
			return createErrorResult(fmt.Errorf("failed to format attachments: %w", err)), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(attachmentsJSON),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note_attachments",
		Description: "Retrieves all attachments for a note in Apple Notes. Returns attachment metadata including file paths as JSON.",
	}, handler)
}

// registerGetAttachmentContentTool registers the get_attachment_content tool
func registerGetAttachmentContentTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input GetAttachmentContentArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.FilePath == "" {
			return nil, nil, fmt.Errorf("%w: file_path is required", services.ErrInvalidInput)
		}

		// Set default max size to 10MB
		maxSizeMB := input.MaxSizeMB
		if maxSizeMB <= 0 {
			maxSizeMB = 10
		}
		maxSizeBytes := int64(maxSizeMB) * 1024 * 1024

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		content, err := notesService.GetAttachmentContent(opCtx, input.FilePath, maxSizeBytes)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Base64 encode the content
		encoded := base64.StdEncoding.EncodeToString(content)

		// Return success result with base64-encoded content
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: encoded,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_attachment_content",
		Description: "Retrieves the content of an attachment from Apple Notes. Returns base64-encoded content. Limited by max_size_mb parameter (default: 10MB).",
	}, handler)
}

// registerExportNoteMarkdownTool registers the export_note_markdown tool
func registerExportNoteMarkdownTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input ExportNoteMarkdownArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.NoteTitle == "" {
			return nil, nil, fmt.Errorf("%w: note_title is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		markdown, err := notesService.ExportNoteMarkdown(opCtx, input.NoteTitle)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: markdown,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_note_markdown",
		Description: "Exports a note from Apple Notes as markdown format. Returns the note content converted to markdown.",
	}, handler)
}

// registerExportNoteTextTool registers the export_note_text tool
func registerExportNoteTextTool(server *mcp.Server, notesService services.NotesService) {
	handler := func(ctx context.Context, req *mcp.CallToolRequest, input ExportNoteTextArgs) (
		*mcp.CallToolResult, any, error) {

		// Validate required fields
		if input.NoteTitle == "" {
			return nil, nil, fmt.Errorf("%w: note_title is required", services.ErrInvalidInput)
		}

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Call the service
		plainText, err := notesService.ExportNoteText(opCtx, input.NoteTitle)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: plainText,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_note_text",
		Description: "Exports a note from Apple Notes as plain text. Returns the note content as plain text without formatting.",
	}, handler)
}

// createErrorResult converts service errors to user-friendly MCP error responses
func createErrorResult(err error) *mcp.CallToolResult {
	var message string

	// Map service errors to user-friendly messages
	switch {
	case errors.Is(err, services.ErrNoteNotFound):
		message = "Note not found in Apple Notes. Please check the title and try again."
	case errors.Is(err, services.ErrNotesAppNotRunning):
		message = "Apple Notes app is not running. Please open the Notes app and try again."
	case errors.Is(err, services.ErrPermissionDenied):
		message = "Permission denied to access Notes. Please grant access in System Preferences > Privacy & Security > Automation."
	case errors.Is(err, services.ErrScriptTimeout):
		message = "Apple Notes is not responding (timeout after 10 seconds). Please try again."
	case errors.Is(err, services.ErrInvalidInput):
		message = fmt.Sprintf("Invalid input: %v", err)
	default:
		// Include the error message for unexpected errors
		message = fmt.Sprintf("An error occurred: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: message,
			},
		},
		IsError: true,
	}
}

// registerResources registers MCP resources for direct note access
func registerResources(server *mcp.Server, notesService services.NotesService) {
	// Register resource template for individual notes: note:///{title}
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "note:///{title}",
			Name:        "note",
			Title:       "Apple Note",
			Description: "Access a specific note by title. Returns the note content in HTML format.",
			MIMEType:    "text/html",
		},
		createNoteResourceHandler(notesService),
	)

	// Register static resource for recent notes: notes:///recent
	server.AddResource(
		&mcp.Resource{
			URI:         "notes:///recent",
			Name:        "recent-notes",
			Title:       "Recent Notes",
			Description: "List of recently modified notes in Apple Notes. Returns note titles sorted by modification date.",
			MIMEType:    "text/plain",
		},
		createRecentNotesResourceHandler(notesService),
	)

	// Register resource template for search: notes:///search/{query}
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "notes:///search/{query}",
			Name:        "search-notes",
			Title:       "Search Notes",
			Description: "Search for notes by title query. Returns matching note titles.",
			MIMEType:    "text/plain",
		},
		createSearchNotesResourceHandler(notesService),
	)

	// Register resource template for folder notes: notes:///folder/{folder}
	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "notes:///folder/{folder}",
			Name:        "folder-notes",
			Title:       "Folder Notes",
			Description: "Access notes in a specific folder. Returns note titles in the folder.",
			MIMEType:    "text/plain",
		},
		createFolderNotesResourceHandler(notesService),
	)
}

// createNoteResourceHandler creates a handler for note:///{title} resources
func createNoteResourceHandler(notesService services.NotesService) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract title from URI (format: note:///{title})
		uri := req.Params.URI
		if !strings.HasPrefix(uri, "note:///") {
			return nil, fmt.Errorf("invalid note URI: %s", uri)
		}

		title := strings.TrimPrefix(uri, "note:///")
		if title == "" {
			return nil, fmt.Errorf("note title is required")
		}

		// URL decode the title
		title = strings.ReplaceAll(title, "%20", " ")

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Get note content
		content, err := notesService.GetNoteContent(opCtx, title)
		if err != nil {
			if errors.Is(err, services.ErrNoteNotFound) {
				return nil, mcp.ResourceNotFoundError(uri)
			}
			return nil, fmt.Errorf("failed to get note content: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "text/html",
					Text:     content,
				},
			},
		}, nil
	}
}

// createRecentNotesResourceHandler creates a handler for notes:///recent resource
func createRecentNotesResourceHandler(notesService services.NotesService) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Get recent notes by searching for all notes (AppleScript returns them sorted)
		notes, err := notesService.GetRecentNotes(opCtx, 20)
		if err != nil {
			return nil, fmt.Errorf("failed to get recent notes: %w", err)
		}

		// Format the results as newline-separated list of titles
		var titles []string
		for _, note := range notes {
			titles = append(titles, note.Title)
		}
		result := strings.Join(titles, "\n")

		if result == "" {
			result = "No notes found."
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "text/plain",
					Text:     result,
				},
			},
		}, nil
	}
}

// createSearchNotesResourceHandler creates a handler for notes:///search/{query} resources
func createSearchNotesResourceHandler(notesService services.NotesService) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract query from URI (format: notes:///search/{query})
		uri := req.Params.URI
		if !strings.HasPrefix(uri, "notes:///search/") {
			return nil, fmt.Errorf("invalid search URI: %s", uri)
		}

		query := strings.TrimPrefix(uri, "notes:///search/")
		if query == "" {
			return nil, fmt.Errorf("search query is required")
		}

		// URL decode the query
		query = strings.ReplaceAll(query, "%20", " ")

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Search for notes
		notes, err := notesService.SearchNotes(opCtx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to search notes: %w", err)
		}

		// Limit results to prevent timeouts
		totalNotes := len(notes)
		if totalNotes > maxSearchResults {
			notes = notes[:maxSearchResults]
		}

		// Format the results as newline-separated list of titles
		var titles []string
		for _, note := range notes {
			titles = append(titles, note.Title)
		}
		result := strings.Join(titles, "\n")

		if result == "" {
			result = "No notes found matching the query."
		} else if totalNotes > maxSearchResults {
			result = fmt.Sprintf("%s\n\n(Showing first %d of %d matching notes)", result, maxSearchResults, totalNotes)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "text/plain",
					Text:     result,
				},
			},
		}, nil
	}
}

// createFolderNotesResourceHandler creates a handler for notes:///folder/{folder} resources
func createFolderNotesResourceHandler(notesService services.NotesService) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract folder from URI (format: notes:///folder/{folder})
		uri := req.Params.URI
		if !strings.HasPrefix(uri, "notes:///folder/") {
			return nil, fmt.Errorf("invalid folder URI: %s", uri)
		}

		folder := strings.TrimPrefix(uri, "notes:///folder/")
		if folder == "" {
			return nil, fmt.Errorf("folder name is required")
		}

		// URL decode the folder name
		folder = strings.ReplaceAll(folder, "%20", " ")

		// Create a context with timeout for the operation
		opCtx, cancel := context.WithTimeout(ctx, getOperationTimeout())
		defer cancel()

		// Get notes in folder
		notes, err := notesService.GetNotesInFolder(opCtx, folder)
		if err != nil {
			return nil, fmt.Errorf("failed to get notes in folder: %w", err)
		}

		// Format the results as newline-separated list of titles
		var titles []string
		for _, note := range notes {
			titles = append(titles, note.Title)
		}
		result := strings.Join(titles, "\n")

		if result == "" {
			result = fmt.Sprintf("No notes found in folder '%s'.", folder)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: "text/plain",
					Text:     result,
				},
			},
		}, nil
	}
}

// registerPrompts registers all prompt templates for workflow assistance
func registerPrompts(server *mcp.Server, notesService services.NotesService) {
	registerDailyReviewPrompt(server, notesService)
	registerWeeklySummaryPrompt(server, notesService)
	registerMeetingPrepPrompt(server, notesService)
	registerActionItemsPrompt(server, notesService)
	registerNoteCleanupPrompt(server, notesService)
	registerQuickNotePrompt(server, notesService)
}

// registerDailyReviewPrompt registers the daily-review prompt
func registerDailyReviewPrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "daily-review",
		Description: "Review notes from today with summary and action items. Analyzes recent notes to provide a daily overview.",
		Arguments:   []*mcp.PromptArgument{},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		// Get today's date
		today := time.Now().Format("2006-01-02")

		instructions := fmt.Sprintf(`Review my notes from today (%s) and provide:

1. A brief summary of key topics and themes
2. Action items extracted from the notes
3. Any follow-up tasks or decisions that need attention
4. Patterns or insights from today's work

Use the notes:///recent resource to access recent notes. Focus on notes created or modified today.`, today)

		return &mcp.GetPromptResult{
			Description: "Daily review prompt with instructions for summarizing today's notes",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}

// registerWeeklySummaryPrompt registers the weekly-summary prompt
func registerWeeklySummaryPrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "weekly-summary",
		Description: "Summarize notes from the past week by category. Provides a high-level overview of weekly activity.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "categories",
				Description: "Optional comma-separated list of categories to focus on (e.g., 'meetings,ideas,todos')",
				Required:    false,
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		categories := req.Params.Arguments["categories"]

		var categoryInstructions string
		if categories != "" {
			categoryInstructions = fmt.Sprintf("\nFocus on these categories: %s", categories)
		}

		instructions := fmt.Sprintf(`Review my notes from the past week and provide a comprehensive summary:

1. Group notes by category or theme
2. Highlight key accomplishments and progress
3. List outstanding action items and decisions
4. Identify any recurring themes or patterns
5. Note any areas that need more attention%s

Use the notes:///recent resource to access recent notes from the past week. Organize the summary by categories to make it easy to review.`, categoryInstructions)

		return &mcp.GetPromptResult{
			Description: "Weekly summary prompt with instructions for reviewing the past week's notes",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}

// registerMeetingPrepPrompt registers the meeting-prep prompt
func registerMeetingPrepPrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "meeting-prep",
		Description: "Prepare for a meeting using relevant notes. Gathers context and talking points for upcoming meetings.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "topic",
				Description: "The meeting topic or project name to search for relevant notes",
				Required:    true,
			},
			{
				Name:        "attendees",
				Description: "Optional comma-separated list of attendees to consider when gathering context",
				Required:    false,
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		topic := req.Params.Arguments["topic"]
		attendees := req.Params.Arguments["attendees"]

		var attendeeContext string
		if attendees != "" {
			attendeeContext = fmt.Sprintf("\nAttendees: %s", attendees)
		}

		instructions := fmt.Sprintf(`Prepare me for a meeting about: %s%s

Please provide:

1. Summary of relevant notes and past discussions on this topic
2. Key points and decisions from previous meetings
3. Outstanding action items or open questions
4. Suggested talking points or agenda items
5. Any background context I should review

Use the notes:///search/%s resource to find relevant notes. Also check notes:///recent for any recent updates related to this topic.`, topic, attendeeContext, topic)

		return &mcp.GetPromptResult{
			Description: "Meeting preparation prompt with instructions for gathering relevant context",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}

// registerActionItemsPrompt registers the action-items prompt
func registerActionItemsPrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "action-items",
		Description: "Extract action items from notes with a specific tag or search term. Consolidates todos and tasks.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "search_term",
				Description: "Tag name or search term to filter notes (e.g., 'TODO', 'action', project name)",
				Required:    true,
			},
			{
				Name:        "status",
				Description: "Optional status filter: 'open', 'completed', or 'all' (default: 'open')",
				Required:    false,
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		searchTerm := req.Params.Arguments["search_term"]
		status := req.Params.Arguments["status"]
		if status == "" {
			status = "open"
		}

		instructions := fmt.Sprintf(`Extract and organize action items from notes matching: %s

Please provide:

1. List of all action items, organized by priority or category
2. For each item, include:
   - The action description
   - Source note title
   - Any deadlines or time constraints mentioned
   - Current status (%s items only)
3. Group items by urgency (urgent, soon, later)
4. Flag any overdue or blocked items

Use the notes:///search/%s resource to find relevant notes. Look for action items, todos, tasks, or similar indicators in the note content.`, searchTerm, status, searchTerm)

		return &mcp.GetPromptResult{
			Description: "Action items extraction prompt with instructions for finding and organizing tasks",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}

// registerNoteCleanupPrompt registers the note-cleanup prompt
func registerNoteCleanupPrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "note-cleanup",
		Description: "Identify notes that should be archived or deleted. Helps maintain a clean and organized notes database.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "age_threshold_days",
				Description: "Optional minimum age in days for notes to consider for cleanup (default: 90)",
				Required:    false,
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		ageThreshold := req.Params.Arguments["age_threshold_days"]
		if ageThreshold == "" {
			ageThreshold = "90"
		}

		instructions := fmt.Sprintf(`Review my notes and identify candidates for archival or deletion.

Please analyze notes and suggest:

1. Notes older than %s days that may be outdated
2. Duplicate or redundant notes
3. Notes with completed action items that can be archived
4. Empty or placeholder notes
5. Notes that should be consolidated or merged

For each suggestion, provide:
- Note title
- Reason for archival/deletion
- Any important content that should be preserved elsewhere

Use the notes:///recent resource to get an overview of notes. Be conservative - only suggest cleanup for notes that are clearly outdated or redundant.`, ageThreshold)

		return &mcp.GetPromptResult{
			Description: "Note cleanup prompt with instructions for identifying notes to archive or delete",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}

// registerQuickNotePrompt registers the quick-note prompt
func registerQuickNotePrompt(server *mcp.Server, notesService services.NotesService) {
	prompt := &mcp.Prompt{
		Name:        "quick-note",
		Description: "Quick capture template for creating structured notes. Provides a consistent format for new notes.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "note_type",
				Description: "Type of note to create: 'meeting', 'idea', 'todo', 'journal', or 'general'",
				Required:    true,
			},
			{
				Name:        "title",
				Description: "Title for the note",
				Required:    true,
			},
		},
	}

	handler := func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		noteType := req.Params.Arguments["note_type"]
		title := req.Params.Arguments["title"]

		var template string
		switch noteType {
		case "meeting":
			template = `Create a meeting note with the following structure:

Title: %s
Date: %s

## Attendees
- [List attendees]

## Agenda
- [Topics to discuss]

## Discussion
- [Key points discussed]

## Decisions
- [Decisions made]

## Action Items
- [ ] [Task 1]
- [ ] [Task 2]

## Next Steps
- [Follow-up items]`
		case "idea":
			template = `Create an idea note with the following structure:

Title: %s
Date: %s

## Overview
[Brief description of the idea]

## Context
[Why this idea matters, what problem it solves]

## Details
[Expanded thoughts and details]

## Next Steps
- [ ] [What to do next]

## Related Notes
- [Links to related notes or resources]`
		case "todo":
			template = `Create a todo note with the following structure:

Title: %s
Date: %s

## Tasks
- [ ] [Task 1]
- [ ] [Task 2]
- [ ] [Task 3]

## Priority
[High/Medium/Low]

## Context
[Why these tasks matter]

## Deadline
[Target completion date]`
		case "journal":
			template = `Create a journal entry with the following structure:

Title: %s
Date: %s

## Mood
[How are you feeling?]

## Accomplishments
- [What went well today]

## Challenges
- [What was difficult]

## Learnings
- [What you learned]

## Tomorrow
- [What to focus on tomorrow]`
		default:
			template = `Create a note with the following structure:

Title: %s
Date: %s

## Summary
[Brief overview]

## Details
[Main content]

## Tags
[Relevant tags for organization]

## References
- [Related notes or links]`
		}

		instructions := fmt.Sprintf("Please create a new note using the create_note tool with this template:\n\n"+template,
			title, time.Now().Format("2006-01-02"))

		return &mcp.GetPromptResult{
			Description: "Quick note creation prompt with a structured template",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: instructions,
					},
				},
			},
		}, nil
	}

	server.AddPrompt(prompt, handler)
}
