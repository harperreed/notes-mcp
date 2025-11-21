// ABOUTME: MCP server subcommand that starts the Model Context Protocol server
// ABOUTME: Implements stdio-based MCP server with six tools: create_note, search_notes, get_note_content, update_note, delete_note, list_folders

package cmd

import (
	"context"
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

		// Return success result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Note created: %s", note.Title),
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_note",
		Description: "Creates a new note in Apple Notes with the specified title, content, and optional tags. Returns confirmation of note creation.",
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

		// Format the results as newline-separated list of titles
		var titles []string
		for _, note := range notes {
			titles = append(titles, note.Title)
		}
		result := strings.Join(titles, "\n")

		// Add indicator if results were limited
		if totalNotes > maxSearchResults {
			result = fmt.Sprintf("%s\n\n(Showing first %d of %d matching notes)", result, maxSearchResults, totalNotes)
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
		Name:        "search_notes",
		Description: "Searches for notes in Apple Notes by title. Returns a list of matching note titles, or a message if no notes are found.",
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

		// Call the service
		content, err := notesService.GetNoteContent(opCtx, input.Title)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Return success result with raw HTML body
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: content,
				},
			},
		}, nil, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note_content",
		Description: "Retrieves the full content of a note from Apple Notes by its title. Returns the raw HTML body of the note.",
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
