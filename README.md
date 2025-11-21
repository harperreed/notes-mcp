# Apple Notes MCP Server (Go Implementation)

A standalone Go implementation of an MCP (Model Context Protocol) server for Apple Notes, providing both MCP protocol integration and CLI tool functionality.

## Features

- **MCP Server Mode**: Integrates with Claude Desktop and other MCP clients
- **CLI Tool Mode**: Command-line interface for managing Apple Notes
- **Three-Layer Architecture**: Clean separation between protocol, business logic, and OS interaction
- **Feature Parity**: Matches the TypeScript implementation functionality

## Requirements

- Go 1.21 or later
- macOS with Apple Notes installed
- System permissions to access Apple Notes via AppleScript

## Installation

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
```

## Claude Desktop Integration

Add to your Claude Desktop configuration:

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

## Development

### Run Tests

```bash
# Unit tests only
go test ./...

# Include integration tests (requires Apple Notes running)
go test -tags=integration ./...
```

### Project Structure

```
notes-mcp/
├── go.mod
├── go.sum
├── main.go                    # CLI entry point with cobra
├── cmd/                       # Subcommand implementations
│   ├── mcp.go                # MCP server subcommand
│   ├── create.go             # create note subcommand
│   ├── search.go             # search notes subcommand
│   └── get.go                # get note content subcommand
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
