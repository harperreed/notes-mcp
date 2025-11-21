# Apple Notes MCP Server - Major Enhancements

**Date:** 2025-11-20
**Status:** Approved for Implementation

## Overview

This design extends the Apple Notes MCP server with comprehensive folder management, attachment support, rich metadata, advanced search, and export capabilities.

## Goals

1. Enable full folder hierarchy management and note organization
2. Support attachment discovery and content retrieval
3. Expose rich note metadata (dates, sharing status, location)
4. Provide advanced search with body content, date ranges, and folder filters
5. Export notes to markdown and plain text formats

## Non-Goals

- Batch operations (users can call tools multiple times)
- Note link parsing and graph building
- Attachment uploads or modifications
- Folder deletion or renaming
- Sharing management or permission controls

## Data Model Changes

### Enhanced Note Structure

```go
type Note struct {
    Title            string    `json:"title"`
    Content          string    `json:"content,omitempty"`
    CreationDate     time.Time `json:"creation_date"`
    ModificationDate time.Time `json:"modification_date"`
    Folder           string    `json:"folder"`
    Shared           bool      `json:"shared"`
    PasswordProtected bool     `json:"password_protected"`
    ID               string    `json:"id"`
}
```

**Rationale:** All fields available via AppleScript properties. Backwards compatible - existing code works unchanged.

### New Structures

```go
type Attachment struct {
    Name             string    `json:"name"`
    FilePath         string    `json:"file_path"`
    ContentIdentifier string   `json:"content_identifier"`
    CreationDate     time.Time `json:"creation_date"`
    ModificationDate time.Time `json:"modification_date"`
    ID               string    `json:"id"`
}

type FolderNode struct {
    Name      string       `json:"name"`
    Shared    bool         `json:"shared"`
    Children  []FolderNode `json:"children,omitempty"`
    NoteCount int          `json:"note_count"`
}

type SearchOptions struct {
    Query      string
    SearchIn   string     // "title", "body", "both"
    Folder     string     // optional: limit to folder
    DateFrom   *time.Time
    DateTo     *time.Time
}
```

**Rationale:**
- `Attachment` captures AppleScript-exposed properties
- `FolderNode` enables hierarchical folder trees
- `SearchOptions` supports incremental performance optimization

## Service Layer (services/notes.go)

### New Methods

```go
// Folder operations
CreateFolder(ctx context.Context, name string, parentFolder string) error
MoveNote(ctx context.Context, noteTitle string, targetFolder string) error
GetFolderHierarchy(ctx context.Context) (*FolderNode, error)

// Enhanced search
SearchNotesAdvanced(ctx context.Context, opts SearchOptions) ([]Note, error)

// Attachments
GetNoteAttachments(ctx context.Context, noteTitle string) ([]Attachment, error)
GetAttachmentContent(ctx context.Context, attachmentID string, maxSize int64) ([]byte, error)

// Export
ExportNoteMarkdown(ctx context.Context, noteTitle string) (string, error)
ExportNoteText(ctx context.Context, noteTitle string) (string, error)
```

### Key Design Decisions

**Folder Creation:**
- `parentFolder=""` creates at root level
- Non-empty parent enables nested folder creation
- AppleScript: `make new folder with properties {name:"Subfolder", container:folder "Parent"}`

**Note Movement:**
- Move via AppleScript: `move note "Title" to folder "Target"`
- Error if note or folder doesn't exist

**Advanced Search Performance:**
- Title search: Fast (current implementation)
- Body search: Slow on large databases (AppleScript timeout observed)
- Strategy: Filter by folder/date first, then search body in subset
- Result limiting: Keep 100-result cap

**Attachment Content:**
- Hybrid approach: Return file path + optional base64 content
- `maxSize` parameter (default 10MB) prevents OOM on large files
- Base64 encoding for small files enables inline analysis

**Export:**
- Returns string content (caller decides what to do)
- Markdown: Convert HTML body to markdown (basic conversion)
- Text: Use `plaintext` property from AppleScript
- Stateless (no filesystem writes from server)

## MCP Tools

### New Tools (8)

1. **create_folder**
   - Params: `name` (string), `parent_folder` (string, optional)
   - Creates folder at root or nested under parent

2. **move_note**
   - Params: `note_title` (string), `target_folder` (string)
   - Moves note to specified folder

3. **get_folder_hierarchy**
   - Returns: Nested FolderNode structure with note counts
   - Useful for organization overview

4. **search_notes_advanced**
   - Params: `query`, `search_in` (title|body|both), `folder`, `date_from`, `date_to`
   - Replaces basic search_notes with optional advanced filters
   - Backwards compatible: defaults to title-only search

5. **get_note_attachments**
   - Params: `note_title` (string)
   - Returns: Array of Attachment objects with paths

6. **get_attachment_content**
   - Params: `attachment_id` (string), `max_size_mb` (int, default 10)
   - Returns: Base64-encoded content or error if too large

7. **export_note_markdown**
   - Params: `note_title` (string)
   - Returns: Markdown-formatted note content

8. **export_note_text**
   - Params: `note_title` (string)
   - Returns: Plain text note content

### Enhanced Existing Tools

All tools that return notes now include full metadata:
- `create_note` returns Note with all fields
- `search_notes` returns Note[] with metadata
- `get_note_content` returns Note with full details

## CLI Commands

All MCP tools get corresponding CLI commands:

```bash
./mcp-apple-notes-go create-folder "Work Projects" --parent="Work"
./mcp-apple-notes-go move-note "Meeting Notes" "Archive"
./mcp-apple-notes-go folder-hierarchy
./mcp-apple-notes-go search-advanced "roadmap" --in=body --folder="Work"
./mcp-apple-notes-go attachments "Trip Photos"
./mcp-apple-notes-go export "Design Doc" --format=markdown
```

## AppleScript Implementation Notes

### Folder Operations

```applescript
-- Create nested folder
tell application "Notes"
    set parentFld to folder "Parent"
    make new folder with properties {name:"Child", container:parentFld}
end tell

-- Move note
tell application "Notes"
    set targetFld to folder "Target"
    set theNote to note "Title"
    move theNote to targetFld
end tell

-- Get hierarchy
tell application "Notes"
    repeat with fld in folders
        {name of fld, folders of fld}
    end repeat
end tell
```

### Advanced Search

```applescript
-- Fast: Title only
notes whose name contains "query"

-- Slow: Body search (timeout risk)
notes whose body contains "query"

-- Optimized: Folder + date filter first
set fld to folder "Work"
set startDate to date "2024-01-01"
set filtered to notes of fld whose modification date > startDate
repeat with n in filtered
    if body of n contains "query" then
        -- match
    end if
end repeat
```

### Attachments

```applescript
tell application "Notes"
    set theNote to note "Title"
    set attList to attachments of theNote
    repeat with att in attList
        {name of att, contents of att, id of att}
    end repeat
end tell
```

**Key Properties:**
- `name`: Filename
- `contents`: POSIX path (file:// format)
- `id`: x-coredata:// identifier
- `modification date`, `creation date`: Timestamps

## Implementation Strategy

### Group 1 - Foundation (parallel)
1. Data models - Add fields to Note, new Attachment/FolderNode/SearchOptions structs
2. AppleScript helpers - Scripts for metadata, folders, attachments

### Group 2 - Core Service Methods (parallel)
3. Folder operations - CreateFolder, MoveNote, GetFolderHierarchy
4. Advanced search - SearchNotesAdvanced with all filters
5. Attachments - GetNoteAttachments, GetAttachmentContent
6. Export - ExportNoteMarkdown, ExportNoteText

### Group 3 - Integration (parallel)
7. MCP tools - 8 new tool handlers in cmd/mcp.go
8. CLI commands - 8 new subcommands in cmd/
9. Enhanced existing - Update create/search/get to return full metadata

### Group 4 - Testing & Documentation
10. Unit tests - All new service methods
11. Integration tests - End-to-end with real Notes
12. Documentation - Update README with new features

## Testing Strategy

### Unit Tests
- Mock ScriptExecutor for all new service methods
- Test error handling (note not found, folder not found, timeout)
- Test data parsing (dates, nested folders, attachments)

### Integration Tests
- Create/move/list folders
- Advanced search scenarios (body, date, folder filters)
- Attachment listing and content retrieval
- Export formats

### Performance Tests
- Body search on large databases (verify timeout handling)
- Large attachment handling (verify size limits)
- Deep folder hierarchy traversal

## Risks and Mitigations

**Risk:** Body search timeouts on large databases
**Mitigation:** Incremental filtering (folder/date first), result limits, clear user warnings

**Risk:** Large attachments cause OOM
**Mitigation:** maxSize parameter with 10MB default, document limits clearly

**Risk:** Nested folder creation if parent doesn't exist
**Mitigation:** Validate parent exists first, return clear error

**Risk:** Move note fails if already in target folder
**Mitigation:** Check current folder, skip move if already there

## Success Metrics

- All 8 new tools working in Claude Desktop
- Full test coverage maintained (>95%)
- No performance regressions on existing operations
- Clear documentation for all new features
- All pre-commit hooks passing

## Future Enhancements (Deferred)

- Batch operations (bulk create/update/delete)
- Note link parsing and backlink discovery
- Attachment uploads
- Folder renaming and deletion
- Sharing management and permissions
- Resource subscriptions (real-time updates)
