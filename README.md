# Apple Notes MCP Server (Go Implementation)

A standalone Go implementation of an MCP (Model Context Protocol) server for Apple Notes, providing both MCP protocol integration and CLI tool functionality.

## Features

- **MCP Server Mode**: Integrates with Claude Desktop and other MCP clients
  - **6 Tools**: create_note, search_notes, get_note_content, update_note, delete_note, list_folders
  - **4 Resource Types**: Direct access to notes via URIs (note:///, notes:///recent, notes:///search/{query}, notes:///folder/{folder})
  - **6 Prompt Templates**: One-click workflows for common note operations (daily-review, weekly-summary, meeting-prep, action-items, note-cleanup, quick-note)
- **CLI Tool Mode**: Command-line interface for managing Apple Notes
- **Three-Layer Architecture**: Clean separation between protocol, business logic, and OS interaction
- **Configurable Timeouts**: Environment variable support for large Notes databases
- **Result Limiting**: Automatic limiting of search results to prevent timeouts

## Requirements

- Go 1.21 or later
- macOS with Apple Notes installed
- System permissions to access Apple Notes via AppleScript

## Installation

### Using Make (Recommended)

```bash
make build        # Build the binary
make install      # Install to /usr/local/bin
```

### Using Go directly

```bash
go build -o mcp-apple-notes-go .
```

## Usage

### MCP Server Mode

Run as an MCP server for integration with Claude Desktop:

```bash
./mcp-apple-notes-go mcp
```

### CLI Tool Mode

Use as a command-line tool:

```bash
# Create a note
./mcp-apple-notes-go create "Meeting Notes" "Discussed Q4 roadmap" --tags=work,meeting

# Search notes
./mcp-apple-notes-go search "meeting"

# Get note content
./mcp-apple-notes-go get "Meeting Notes"

# Update a note
./mcp-apple-notes-go update "Meeting Notes" "Updated Q4 roadmap with new timeline"

# Delete a note
./mcp-apple-notes-go delete "Old Note"

# List all folders
./mcp-apple-notes-go folders
```

## Claude Desktop Integration

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "apple-notes-go": {
      "command": "/path/to/notes-mcp/mcp-apple-notes-go",
      "args": ["mcp"],
      "env": {
        "NOTES_MCP_TIMEOUT": "60"
      }
    }
  }
}
```

### Configuration Options

- **NOTES_MCP_TIMEOUT**: Optional timeout in seconds for operations (default: 30). Increase if you have a large Notes database and experience timeouts during searches.
- Search results are automatically limited to 100 notes to prevent timeouts with large result sets.

### MCP Tools

The server provides six tools for Claude to interact with Apple Notes:

1. **create_note** - Create a new note with title, content, and optional tags
2. **search_notes** - Search for notes by title query (limited to 100 results)
3. **get_note_content** - Retrieve the full HTML content of a note
4. **update_note** - Update the content of an existing note
5. **delete_note** - Delete a note by title
6. **list_folders** - List all folders in Apple Notes

### MCP Resources

The server exposes notes as resources for direct access:

- **`note:///{title}`** - Access a specific note by title (e.g., `note:///Meeting%20Notes`)
- **`notes:///recent`** - List 20 most recently modified notes
- **`notes:///search/{query}`** - Search results as a resource (e.g., `notes:///search/meeting`)
- **`notes:///folder/{folder}`** - List notes in a specific folder (e.g., `notes:///folder/Work`)

Resources allow Claude to read note content directly without tool calls, making it more natural to say things like "based on my meeting notes..."

### MCP Prompts

The server provides six pre-built prompt templates for common workflows:

1. **daily-review** - Review today's notes with summary and action items
2. **weekly-summary** - Comprehensive weekly summary by category (optional: `categories`)
3. **meeting-prep** - Prepare for meetings using relevant notes (required: `topic`, optional: `attendees`)
4. **action-items** - Extract and organize action items (required: `search_term`, optional: `status`)
5. **note-cleanup** - Identify notes for archival or deletion (optional: `age_threshold_days`)
6. **quick-note** - Structured templates for rapid note capture (required: `note_type`, `title`)

Prompts are user-triggered and provide Claude with structured instructions for common note operations. They appear in your MCP client's prompt menu for one-click access.

## Development

### Quick Start with Make

```bash
make help         # Show all available commands
make build        # Build the binary
make test         # Run unit tests
make lint         # Run linter
make check        # Run format, lint, and test
make run          # Build and start MCP server
make clean        # Remove build artifacts
```

### Run Tests

```bash
# Using Make
make test                  # Unit tests only
make test-integration      # Integration tests (requires Apple Notes)
make test-all             # All tests
make test-coverage        # Generate coverage report

# Using Go directly
go test ./...
go test -tags=integration ./...
```

### Development Workflow

```bash
make format       # Format code
make lint         # Run linter
make lint-fix     # Auto-fix linting issues
make check        # Full check (format + lint + test)
make pre-commit   # Run all pre-commit hooks
```

### Project Structure

```
notes-mcp/
├── go.mod
├── go.sum
├── main.go                    # CLI entry point with cobra
├── cmd/                       # Subcommand implementations
│   ├── mcp.go                # MCP server subcommand (tools + resources)
│   ├── create.go             # create note subcommand
│   ├── search.go             # search notes subcommand
│   ├── get.go                # get note content subcommand
│   ├── update.go             # update note subcommand
│   ├── delete.go             # delete note subcommand
│   └── folders.go            # list folders subcommand
├── services/                  # Business logic layer
│   ├── notes.go              # NotesService interface & implementation
│   ├── notes_test.go         # Unit tests with mock executor
│   ├── notes_integration_test.go  # Integration tests
│   ├── applescript.go        # ScriptExecutor interface & implementation
│   ├── applescript_test.go   # Executor unit tests
│   └── errors.go             # Custom error types & detection
├── README.md
└── docs/
    └── plans/
        └── 2025-11-20-apple-notes-mcp-design.md
```

## Architecture

The implementation follows a three-layer architecture:

1. **Protocol/CLI Layer**: Handles MCP protocol communication and CLI interface
2. **Service Layer**: Business logic for note operations
3. **Execution Layer**: AppleScript execution and OS interaction

See [docs/plans/2025-11-20-apple-notes-mcp-design.md](docs/plans/2025-11-20-apple-notes-mcp-design.md) for detailed design documentation.

## License

TBD
