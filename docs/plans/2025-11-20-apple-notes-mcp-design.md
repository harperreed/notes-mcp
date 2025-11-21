# Apple Notes MCP Server - Go Implementation Design

**Date:** 2025-11-20
**Status:** Approved
**Author:** Claude (with Doctor Biz)

## Overview

A standalone Go implementation of an MCP server for Apple Notes, providing both MCP protocol integration and CLI tool functionality. Achieves feature parity with the existing TypeScript implementation while adding command-line interface capabilities.

## Architecture

### Three-Layer Architecture

**1. Protocol/CLI Layer (main.go, cmd/)**
- Handles MCP protocol communication using `github.com/modelcontextprotocol/go-sdk`
- Provides CLI interface using `github.com/spf13/cobra`
- Registers tool handlers for `create_note`, `search_notes`, `get_note_content`
- Marshals/unmarshals JSON requests and responses
- Catches panics and converts them to MCP error responses

**2. Service Layer (services/notes.go)**
- Defines `NotesService` interface with business operations
- Implements graceful error handling with custom error types
- Formats content for AppleScript (HTML escaping, newline conversion)
- Parses AppleScript output into structured Note types

**3. Execution Layer (services/applescript.go)**
- Defines `ScriptExecutor` interface for running AppleScript commands
- Concrete `OSAScriptExecutor` implementation using `exec.CommandContext`
- 10-second timeout on all commands
- Captures stdout/stderr for error messages

### Benefits
- Clean boundaries between MCP protocol, business logic, and OS interaction
- Easy to mock `ScriptExecutor` for unit tests without running real AppleScript
- Service layer reusable in both MCP and CLI contexts
- Repository pattern readiness if we later add SQLite caching

## Data Models & Interfaces

### Core Domain Model

```go
type Note struct {
    ID       string    `json:"id"`
    Title    string    `json:"title"`
    Content  string    `json:"content"`
    Tags     []string  `json:"tags"`
    Created  time.Time `json:"created"`
    Modified time.Time `json:"modified"`
}
```

### Service Interface

```go
type NotesService interface {
    CreateNote(ctx context.Context, title, content string, tags []string) (*Note, error)
    SearchNotes(ctx context.Context, query string) ([]Note, error)
    GetNoteContent(ctx context.Context, title string) (string, error)
}
```

### Executor Interface

```go
type ScriptExecutor interface {
    Execute(ctx context.Context, script string) (stdout string, stderr string, err error)
}
```

### Implementation Structs

```go
type AppleNotesService struct {
    executor      ScriptExecutor
    iCloudAccount string
    logger        *slog.Logger
}

type OSAScriptExecutor struct {
    timeout time.Duration
}
```

### Key Design Decisions
- Context throughout for cancellation and timeout propagation
- Tags stored in `Note` struct but not used by AppleScript (matching TS behavior)
- `SearchNotes` returns `[]Note` with empty Content (search doesn't retrieve bodies)
- ID generation uses `time.Now().UnixMilli()` as string (simple, monotonic)

## Error Handling

### Custom Error Types

```go
var (
    ErrNoteNotFound       = errors.New("note not found")
    ErrNotesAppNotRunning = errors.New("Apple Notes app not running")
    ErrPermissionDenied   = errors.New("permission denied to access Notes")
    ErrScriptTimeout      = errors.New("AppleScript execution timeout")
    ErrInvalidInput       = errors.New("invalid input parameters")
)
```

### Error Detection Strategy

AppleScript returns cryptic numeric codes. Pattern-match stderr output:
- `-1728` or "event not handled" → `ErrNotesAppNotRunning`
- `note.*not found` (regex) → `ErrNoteNotFound`
- `not allowed` or `-1743` → `ErrPermissionDenied`
- `context.DeadlineExceeded` → `ErrScriptTimeout`

### MCP Error Response Mapping

Convert custom errors to user-friendly messages:
- `ErrNoteNotFound` → "Note 'Meeting Notes' not found in Apple Notes"
- `ErrPermissionDenied` → "Please grant access to Notes in System Preferences > Privacy & Security"
- `ErrScriptTimeout` → "Apple Notes is not responding (10s timeout)"

### Structured Logging

Use `log/slog` for structured logging with context:
- Every AppleScript execution logs: command, duration, success/failure
- Errors log: error type, raw stderr, sanitized user message

## MCP Tool Handlers

### Tool Registration

Three tools matching the TypeScript implementation:

```go
server.AddTool(mcp.Tool{
    Name:        "create_note",
    Description: "Creates a new note in Apple Notes",
    InputSchema: mcp.ToolInputSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "title":   map[string]string{"type": "string", "description": "The title of the note"},
            "content": map[string]string{"type": "string", "description": "The content of the note"},
            "tags":    map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
        },
        Required: []string{"title", "content"},
    },
})
```

### Handler Implementation Pattern

Each handler follows this flow:
1. Unmarshal MCP arguments into a typed struct
2. Validate required fields (return `ErrInvalidInput` if missing)
3. Call the service layer with context
4. Convert service errors to MCP error responses with user-friendly messages
5. Format successful responses as MCP TextContent

### Response Formats
- `create_note` → "Note created: {title}"
- `search_notes` → Newline-separated list of titles (matching TS output)
- `get_note_content` → Raw HTML body from Apple Notes

### Context Management

Each MCP request gets a background context with 30-second timeout (3x the AppleScript timeout) to allow for proper cleanup.

## CLI Interface

### Subcommands

Using `cobra` for clean subcommand structure:

```bash
# MCP server mode
mcp-apple-notes-go mcp

# CLI tool mode
mcp-apple-notes-go create "Meeting Notes" "Discussed Q4 roadmap" --tags=work,meeting
mcp-apple-notes-go search "meeting"
mcp-apple-notes-go get "Meeting Notes"
```

### Subcommand Specifications

**create:**
- Args: `<title> <content>`
- Flags: `--tags` (comma-separated list)
- Output: "Note created: {title}" or error message

**search:**
- Args: `<query>`
- Output: Newline-separated list of matching note titles

**get:**
- Args: `<title>`
- Output: Raw note content (HTML) or error message

### Shared Service Layer

Both MCP and CLI modes use the same `NotesService` implementation, ensuring consistent behavior and reducing duplication.

## Testing Strategy

### Unit Tests (services/notes_test.go)

Mock the `ScriptExecutor` interface to test business logic without AppleScript:

```go
type MockExecutor struct {
    stdout string
    stderr string
    err    error
}

func (m *MockExecutor) Execute(ctx context.Context, script string) (string, string, error) {
    return m.stdout, m.stderr, m.err
}
```

Test cases:
- CreateNote: Verify AppleScript generation, content escaping, note object creation
- SearchNotes: Test CSV parsing, empty results, special characters in titles
- GetNoteContent: Test HTML body retrieval, note not found handling
- Error detection: Verify mapping of stderr patterns to custom errors
- Escaping: Test backslashes, quotes, newlines in content

### Integration Tests (services/notes_integration_test.go)

Use `//go:build integration` tag. Requires Apple Notes running:

```go
//go:build integration

func TestCreateNoteIntegration(t *testing.T) {
    executor := NewOSAScriptExecutor(10 * time.Second)
    service := NewAppleNotesService(executor)

    note, err := service.CreateNote(ctx, "Test Note", "Content", nil)
    // Verify note appears in Apple Notes
    // Clean up: delete test note
}
```

### Table-Driven Tests

Use table-driven pattern for error detection:

```go
tests := []struct {
    name   string
    stderr string
    want   error
}{
    {"note not found", "note 'Foo' not found", ErrNoteNotFound},
    {"permission denied", "not allowed (-1743)", ErrPermissionDenied},
    // ...
}
```

## Project Structure

```
notes-mcp/
├── mcp-apple-notes/           # TypeScript reference (untouched)
├── go.mod
├── go.sum
├── main.go                    # CLI entry point with cobra
├── cmd/
│   ├── mcp.go                # MCP server subcommand
│   ├── create.go             # create note subcommand
│   ├── search.go             # search notes subcommand
│   └── get.go                # get note content subcommand
├── services/
│   ├── notes.go              # NotesService interface & implementation
│   ├── notes_test.go         # Unit tests with mock executor
│   ├── notes_integration_test.go  # Integration tests (build tag)
│   ├── applescript.go        # ScriptExecutor interface & implementation
│   ├── applescript_test.go   # Executor unit tests
│   └── errors.go             # Custom error types & detection
├── README.md
└── docs/
    └── plans/
        └── 2025-11-20-apple-notes-mcp-design.md
```

## Dependencies

```go
module github.com/yourusername/notes-mcp

go 1.21

require (
    github.com/modelcontextprotocol/go-sdk v0.1.0
    github.com/spf13/cobra v1.8.0
)
```

## Build & Usage

### Build

```bash
go build -o mcp-apple-notes-go .
```

### Run Tests

```bash
go test ./...                        # Unit tests only
go test -tags=integration ./...      # Include integration tests
```

### MCP Server Mode

```bash
./mcp-apple-notes-go mcp
```

### CLI Tool Mode

```bash
./mcp-apple-notes-go create "My Note" "Note content" --tags=personal,ideas
./mcp-apple-notes-go search "meeting"
./mcp-apple-notes-go get "My Note"
```

### Claude Desktop Integration

```json
{
  "mcpServers": {
    "apple-notes-go": {
      "command": "/path/to/notes-mcp/mcp-apple-notes-go",
      "args": ["mcp"]
    }
  }
}
```

## Implementation Notes

### AppleScript Escaping

Critical for security and correctness:
1. Escape backslashes first (prevent double-escaping)
2. Escape double quotes second
3. Convert newlines to `<br>` for HTML body

### iCloud Account

Hard-coded to `"iCloud"` matching TypeScript implementation. Future enhancement: make configurable via environment variable or flag.

### Tag Handling

Tags are accepted and stored in the Note struct but NOT passed to AppleScript (AppleScript doesn't support tags in the create command). This matches the TypeScript behavior and maintains API compatibility.

## Future Enhancements

- Update/delete note operations
- List all notes
- Folder/account selection
- Tag management (if Apple Notes API supports it)
- SQLite caching layer for faster searches
- Retry logic for transient failures
