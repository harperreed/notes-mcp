// ABOUTME: MCP server subcommand that starts the Model Context Protocol server
// ABOUTME: Implements stdio-based MCP server with three tools: create_note, search_notes, get_note_content

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

	// Register the three tools
	registerCreateNoteTool(server, notesService)
	registerSearchNotesTool(server, notesService)
	registerGetNoteContentTool(server, notesService)

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
		opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
		Description: "Creates a new note in Apple Notes",
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
		opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Call the service
		notes, err := notesService.SearchNotes(opCtx, input.Query)
		if err != nil {
			return createErrorResult(err), nil, nil
		}

		// Format the results as newline-separated list of titles
		var titles []string
		for _, note := range notes {
			titles = append(titles, note.Title)
		}
		result := strings.Join(titles, "\n")

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
		Description: "Searches for notes by title query",
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
		opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
		Description: "Retrieves the full content of a note by title",
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
