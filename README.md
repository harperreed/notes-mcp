# Apple Notes MCP Server (Go Implementation)

[![CI](https://github.com/harperreed/notes-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/harperreed/notes-mcp/actions/workflows/ci.yml)
[![Release](https://github.com/harperreed/notes-mcp/actions/workflows/release.yml/badge.svg)](https://github.com/harperreed/notes-mcp/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/harperreed/notes-mcp)](https://goreportcard.com/report/github.com/harperreed/notes-mcp)

A standalone Go implementation of an MCP (Model Context Protocol) server for Apple Notes, providing both MCP protocol integration and CLI tool functionality.

## Features

- **MCP Server Mode**: Integrates with Claude Desktop and other MCP clients
  - **14 Tools**: Full note lifecycle, folder management, advanced search, attachments, and export
  - **4 Resource Types**: Direct access to notes via URIs (note:///, notes:///recent, notes:///search/{query}, notes:///folder/{folder})
  - **6 Prompt Templates**: One-click workflows for common note operations (daily-review, weekly-summary, meeting-prep, action-items, note-cleanup, quick-note)
  - **Rich Metadata**: All notes include creation/modification dates, folder, sharing status, and ID
- **CLI Tool Mode**: Command-line interface for managing Apple Notes
- **Three-Layer Architecture**: Clean separation between protocol, business logic, and OS interaction
- **Configurable Timeouts**: Environment variable support for large Notes databases
- **Result Limiting**: Automatic limiting of search results to prevent timeouts

## Requirements

- Go 1.21 or later
- macOS with Apple Notes installed
- System permissions to access Apple Notes via AppleScript

## Installation

### Using Homebrew (Recommended)

```bash
brew install harperreed/tap/notes-mcp
```

### Using Make

```bash
make build        # Build the binary
make install      # Install to /usr/local/bin
```

### Using Go directly

```bash
go build -o notes-mcp .
```

## Usage

### MCP Server Mode

Run as an MCP server for integration with Claude Desktop:

```bash
notes-mcp mcp
```

### CLI Tool Mode

Use as a command-line tool:

#### Core Note Operations

```bash
# Create a note
notes-mcp create "Meeting Notes" "Discussed Q4 roadmap" --tags=work,meeting

# Get note content with full metadata
notes-mcp get "Meeting Notes"

# Update a note
notes-mcp update "Meeting Notes" "Updated Q4 roadmap with new timeline"

# Delete a note
notes-mcp delete "Old Note"
```

#### Search and Discovery

```bash
# Basic search by title
notes-mcp search "meeting"

# Advanced search in note body
notes-mcp search-advanced "roadmap" --search-in=body

# Search with folder filter
notes-mcp search-advanced "meeting" --folder="Work"

# Search with date range
notes-mcp search-advanced "project" --date-from="2024-01-01" --date-to="2024-12-31"

# Combine all filters
notes-mcp search-advanced "roadmap" --search-in=both --folder="Work" --date-from="2024-01-01"
```

#### Folder Management

```bash
# List all folders
notes-mcp folders

# Create a folder at root level
notes-mcp create-folder "Work Projects"

# Create a nested folder
notes-mcp create-folder "Active Projects" --parent="Work"

# Move a note to different folder
notes-mcp move-note "Meeting Notes" "Archive"

# Get folder hierarchy with note counts
notes-mcp folder-hierarchy
```

#### Attachments

```bash
# List attachments in a note
notes-mcp attachments "Trip Photos"

# Get attachment content (saves to file or stdout)
notes-mcp get-attachment "x-coredata://..." --output="photo.jpg"

# Get attachment with size limit
notes-mcp get-attachment "x-coredata://..." --max-size=5
```

#### Export

```bash
# Export note as markdown
notes-mcp export-markdown "Design Doc"

# Export note as plain text
notes-mcp export-text "Design Doc"
```

## Claude Desktop Integration

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "apple-notes": {
      "command": "notes-mcp",
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

The server provides 14 tools for Claude to interact with Apple Notes:

#### Core Note Operations

1. **create_note** - Create a new note with title, content, and optional tags
   ```json
   {
     "title": "Meeting Notes",
     "content": "Discussed Q4 roadmap",
     "tags": "work,meeting"
   }
   ```
   Returns full note metadata including creation date, folder, and ID.

2. **get_note_content** - Retrieve the full HTML content of a note with metadata
   ```json
   {
     "title": "Meeting Notes"
   }
   ```
   Returns note with creation_date, modification_date, folder, shared status, and ID.

3. **update_note** - Update the content of an existing note
   ```json
   {
     "title": "Meeting Notes",
     "content": "Updated with action items"
   }
   ```

4. **delete_note** - Delete a note by title
   ```json
   {
     "title": "Old Note"
   }
   ```

#### Search and Discovery

5. **search_notes** - Basic search by title (limited to 100 results)
   ```json
   {
     "query": "meeting"
   }
   ```
   Returns array of notes with full metadata.

6. **search_notes_advanced** - Advanced search with body content, folder, and date filters
   ```json
   {
     "query": "roadmap",
     "search_in": "body",
     "folder": "Work",
     "date_from": "2024-01-01",
     "date_to": "2024-12-31"
   }
   ```
   - `search_in`: "title" (default), "body", or "both"
   - `folder`: Optional - limit search to specific folder
   - `date_from`/`date_to`: Optional - filter by modification date
   - Performance note: Body search may be slow on large databases

#### Folder Management

7. **list_folders** - List all folders in Apple Notes
   ```json
   {}
   ```

8. **create_folder** - Create a new folder with optional parent
   ```json
   {
     "name": "Work Projects",
     "parent_folder": "Work"
   }
   ```
   Omit `parent_folder` to create at root level.

9. **move_note** - Move a note to a different folder
   ```json
   {
     "note_title": "Meeting Notes",
     "target_folder": "Archive"
   }
   ```

10. **get_folder_hierarchy** - Get nested folder structure with note counts
    ```json
    {}
    ```
    Returns tree structure showing all folders, subfolders, and note counts.

#### Attachments

11. **get_note_attachments** - List all attachments in a note
    ```json
    {
      "note_title": "Trip Photos"
    }
    ```
    Returns array of attachments with name, file path, creation date, and ID.

12. **get_attachment_content** - Retrieve attachment content as base64
    ```json
    {
      "attachment_id": "x-coredata://...",
      "max_size_mb": 10
    }
    ```
    Default max size is 10MB. Returns base64-encoded content for small files, error for large files.

#### Export

13. **export_note_markdown** - Export note content as markdown
    ```json
    {
      "note_title": "Design Doc"
    }
    ```
    Converts HTML content to markdown format.

14. **export_note_text** - Export note content as plain text
    ```json
    {
      "note_title": "Design Doc"
    }
    ```
    Returns plain text without HTML formatting.

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
│   ├── mcp.go                # MCP server subcommand (14 tools + resources + prompts)
│   ├── create.go             # create note subcommand
│   ├── search.go             # search notes subcommand
│   ├── get.go                # get note content subcommand
│   ├── update.go             # update note subcommand
│   ├── delete.go             # delete note subcommand
│   ├── folders.go            # list folders subcommand
│   ├── create_folder.go      # create folder subcommand
│   ├── move_note.go          # move note subcommand
│   ├── folder_hierarchy.go   # get folder hierarchy subcommand
│   ├── search_advanced.go    # advanced search subcommand
│   ├── attachments.go        # list attachments subcommand
│   ├── get_attachment.go     # get attachment content subcommand
│   ├── export_markdown.go    # export as markdown subcommand
│   └── export_text.go        # export as plain text subcommand
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
        ├── 2025-11-20-apple-notes-mcp-design.md
        └── 2025-11-20-notes-mcp-enhancements.md
```

## Architecture

The implementation follows a three-layer architecture:

1. **Protocol/CLI Layer**: Handles MCP protocol communication and CLI interface
2. **Service Layer**: Business logic for note operations
3. **Execution Layer**: AppleScript execution and OS interaction

See [docs/plans/2025-11-20-apple-notes-mcp-design.md](docs/plans/2025-11-20-apple-notes-mcp-design.md) for detailed design documentation.

## License

TBD
