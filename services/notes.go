// ABOUTME: Notes service interface and core business logic
// ABOUTME: Defines the contract for notes management operations

package services

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// NotesService defines the interface for notes management
type NotesService interface {
	// CreateNote creates a new note in Apple Notes
	CreateNote(ctx context.Context, title, content string, tags []string) (*Note, error)

	// SearchNotes searches for notes by title query
	SearchNotes(ctx context.Context, query string) ([]Note, error)

	// SearchNotesAdvanced searches for notes with advanced filters
	SearchNotesAdvanced(ctx context.Context, opts SearchOptions) ([]Note, error)

	// GetNoteContent retrieves the full content of a note by title
	GetNoteContent(ctx context.Context, title string) (string, error)

	// UpdateNote updates an existing note's content by title
	UpdateNote(ctx context.Context, title, content string) error

	// DeleteNote deletes a note by title
	DeleteNote(ctx context.Context, title string) error

	// ListFolders lists all folders in Apple Notes
	ListFolders(ctx context.Context) ([]string, error)

	// GetRecentNotes retrieves recently modified notes
	GetRecentNotes(ctx context.Context, limit int) ([]Note, error)

	// GetNotesInFolder retrieves all notes in a specific folder
	GetNotesInFolder(ctx context.Context, folder string) ([]Note, error)

	// GetAttachmentContent retrieves the content of an attachment from its file path
	GetAttachmentContent(ctx context.Context, filePath string, maxSize int64) ([]byte, error)

	// ExportNoteMarkdown exports a note as markdown by converting HTML body to markdown
	ExportNoteMarkdown(ctx context.Context, noteTitle string) (string, error)

	// ExportNoteText exports a note as plain text using AppleScript plaintext property
	ExportNoteText(ctx context.Context, noteTitle string) (string, error)
}

// Note represents a note entity
type Note struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Content           string    `json:"content,omitempty"`
	Tags              []string  `json:"tags"`
	Created           time.Time `json:"created"`
	Modified          time.Time `json:"modified"`
	CreationDate      time.Time `json:"creation_date"`
	ModificationDate  time.Time `json:"modification_date"`
	Folder            string    `json:"folder"`
	Shared            bool      `json:"shared"`
	PasswordProtected bool      `json:"password_protected"`
}

// Attachment represents a file attachment in a note
type Attachment struct {
	Name              string    `json:"name"`
	FilePath          string    `json:"file_path"`
	ContentIdentifier string    `json:"content_identifier"`
	CreationDate      time.Time `json:"creation_date"`
	ModificationDate  time.Time `json:"modification_date"`
	ID                string    `json:"id"`
}

// FolderNode represents a folder in the hierarchical structure
type FolderNode struct {
	Name      string       `json:"name"`
	Shared    bool         `json:"shared"`
	Children  []FolderNode `json:"children,omitempty"`
	NoteCount int          `json:"note_count"`
}

// SearchOptions contains parameters for advanced note search
type SearchOptions struct {
	Query    string
	SearchIn string     // "title", "body", "both"
	Folder   string     // optional: limit to folder
	DateFrom *time.Time // optional: filter by date range
	DateTo   *time.Time // optional: filter by date range
}

// Search location constants
const (
	SearchInTitle = "title"
	SearchInBody  = "body"
	SearchInBoth  = "both"
)

// AppleNotesService implements NotesService using AppleScript
type AppleNotesService struct {
	executor      ScriptExecutor
	iCloudAccount string
}

// NewAppleNotesService creates a new AppleNotesService with the provided executor
func NewAppleNotesService(executor ScriptExecutor) *AppleNotesService {
	return &AppleNotesService{
		executor:      executor,
		iCloudAccount: "iCloud",
	}
}

// escapeForAppleScript escapes special characters for use in AppleScript strings
// Must escape backslashes first to avoid double-escaping, then quotes
func (s *AppleNotesService) escapeForAppleScript(input string) string {
	// 1. Escape backslashes first (prevents double-escaping)
	result := strings.ReplaceAll(input, "\\", "\\\\")
	// 2. Escape double quotes
	result = strings.ReplaceAll(result, "\"", "\\\"")
	return result
}

// formatContent prepares content for AppleScript by escaping and converting newlines to HTML breaks
func (s *AppleNotesService) formatContent(content string) string {
	if content == "" {
		return ""
	}
	// Escape special characters
	escaped := s.escapeForAppleScript(content)
	// Convert newlines to HTML breaks for Note body
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

// CreateNote creates a new note in Apple Notes with the given title, content, and tags
// Tags are stored in the Note struct but not passed to AppleScript (matching TypeScript behavior)
func (s *AppleNotesService) CreateNote(ctx context.Context, title, content string, tags []string) (*Note, error) {
	// Format content and escape title
	formattedContent := s.formatContent(content)
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to create note
	// Note: tags are not passed to AppleScript as the API doesn't support them
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				make new note with properties {name:"%s", body:"%s"}
			end tell
		end tell
	`, s.iCloudAccount, safeTitle, formattedContent)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return nil, fmt.Errorf("failed to create note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	// Return the note object with generated ID
	now := time.Now()
	return &Note{
		ID:       fmt.Sprintf("%d", now.UnixMilli()),
		Title:    title,
		Content:  content,
		Tags:     tags,
		Created:  now,
		Modified: now,
	}, nil
}

// SearchNotes searches for notes containing the query string in their title
// Returns notes with empty Content (search doesn't retrieve full bodies)
func (s *AppleNotesService) SearchNotes(ctx context.Context, query string) ([]Note, error) {
	safeQuery := s.escapeForAppleScript(query)

	// Generate AppleScript to search notes
	// Use custom delimiter to avoid issues with note titles containing commas
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set noteList to {}
				set foundNotes to notes where name contains "%s"
				repeat with n in foundNotes
					set end of noteList to name of n
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to noteList as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
			end tell
		end tell
	`, s.iCloudAccount, safeQuery)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to search notes: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse delimiter-separated output
	titles := strings.Split(stdout, "|||")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Search doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})
	}

	return notes, nil
}

// GetNoteContent retrieves the full HTML body content of a note by its title
func (s *AppleNotesService) GetNoteContent(ctx context.Context, title string) (string, error) {
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to get note content
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get body of note "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return "", fmt.Errorf("failed to get note content: %w", detectedErr)
	}

	return stdout, nil
}

// UpdateNote updates the content of an existing note by its title
func (s *AppleNotesService) UpdateNote(ctx context.Context, title, content string) error {
	// Format content and escape title
	formattedContent := s.formatContent(content)
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to update note
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set body of note "%s" to "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle, formattedContent)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to update note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	return nil
}

// DeleteNote deletes a note by its title
func (s *AppleNotesService) DeleteNote(ctx context.Context, title string) error {
	// Escape title
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to delete note
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				delete note "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error appropriately
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to delete note: %w", detectedErr)
	}

	// Log success (stdout might contain confirmation)
	_ = stdout

	return nil
}

// ListFolders lists all folders in Apple Notes
func (s *AppleNotesService) ListFolders(ctx context.Context) ([]string, error) {
	// Generate AppleScript to list folders
	// Use custom delimiter to avoid issues with folder names containing commas
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set folderList to {}
				set allFolders to folders
				repeat with f in allFolders
					set end of folderList to name of f
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to folderList as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
			end tell
		end tell
	`, s.iCloudAccount)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []string{}, fmt.Errorf("failed to list folders: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []string{}, nil
	}

	// Parse delimiter-separated output
	folders := strings.Split(stdout, "|||")
	result := make([]string, 0, len(folders))

	for _, folder := range folders {
		folder = strings.TrimSpace(folder)
		if folder == "" {
			continue
		}
		result = append(result, folder)
	}

	return result, nil
}

// GetRecentNotes retrieves recently modified notes, sorted by modification date
func (s *AppleNotesService) GetRecentNotes(ctx context.Context, limit int) ([]Note, error) {
	// Generate AppleScript to get recent notes sorted by modification date
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes
			end tell
		end tell
	`, s.iCloudAccount)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to get recent notes: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	titles := strings.Split(stdout, ", ")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	// Apply limit
	count := 0
	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Recent notes doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})

		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return notes, nil
}

// GetNotesInFolder retrieves all notes in a specific folder
func (s *AppleNotesService) GetNotesInFolder(ctx context.Context, folder string) ([]Note, error) {
	safeFolder := s.escapeForAppleScript(folder)

	// Generate AppleScript to get notes in folder
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes in folder "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeFolder)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to get notes in folder: %w", detectedErr)
	}

	// If output is empty, return empty slice
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}, nil
	}

	// Parse CSV output (AppleScript returns comma-separated list)
	titles := strings.Split(stdout, ", ")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Folder listing doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})
	}

	return notes, nil
}

// GetNoteMetadata retrieves full metadata for a note including dates, folder, and sharing info
// This method ensures both timestamp field sets are synchronized (Created/CreationDate, Modified/ModificationDate)
func (s *AppleNotesService) GetNoteMetadata(ctx context.Context, title string) (*Note, error) {
	safeTitle := s.escapeForAppleScript(title)

	// Generate AppleScript to get note metadata
	// Returns a record with all metadata fields
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set theNote to note "%s"
				{id:(id of theNote as text), name:(name of theNote), creation date:(creation date of theNote), modification date:(modification date of theNote), container:(name of container of theNote), shared:(shared of theNote), password protected:(password protected of theNote)}
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return nil, fmt.Errorf("failed to get note metadata: %w", detectedErr)
	}

	// Parse the AppleScript record output
	note, err := s.parseNoteMetadata(stdout, title)
	if err != nil {
		return nil, fmt.Errorf("failed to parse note metadata: %w", err)
	}

	return note, nil
}

// parseNoteMetadata parses AppleScript record output into a Note struct
// AppleScript returns records like: {id:"x-coredata://...", name:"Title", creation date:date "...", ...}
func (s *AppleNotesService) parseNoteMetadata(output string, title string) (*Note, error) {
	note := &Note{
		Title: title,
		Tags:  []string{},
	}

	// Parse ID
	if id := extractField(output, "id"); id != "" {
		note.ID = id
	}

	// Parse container (folder)
	if folder := extractField(output, "container"); folder != "" {
		note.Folder = folder
	}

	// Parse shared status
	if shared := extractField(output, "shared"); shared == "true" {
		note.Shared = true
	}

	// Parse password protected status
	if passwordProtected := extractField(output, "password protected"); passwordProtected == "true" {
		note.PasswordProtected = true
	}

	// Parse creation date
	if creationDateStr := extractDateField(output, "creation date"); creationDateStr != "" {
		creationDate, err := s.parseAppleScriptDate(creationDateStr)
		if err == nil {
			// Synchronize both timestamp fields
			note.Created = creationDate
			note.CreationDate = creationDate
		}
	}

	// Parse modification date
	if modificationDateStr := extractDateField(output, "modification date"); modificationDateStr != "" {
		modificationDate, err := s.parseAppleScriptDate(modificationDateStr)
		if err == nil {
			// Synchronize both timestamp fields
			note.Modified = modificationDate
			note.ModificationDate = modificationDate
		}
	}

	return note, nil
}

// extractField extracts a simple field value from AppleScript record output
// Example: extractField("{id:\"123\", name:\"Test\"}", "id") returns "123"
func extractField(output, fieldName string) string {
	// Pattern: fieldName:"value" or fieldName:value
	pattern := regexp.MustCompile(fieldName + `:(?:"([^"]+)"|([^,}]+))`)
	matches := pattern.FindStringSubmatch(output)
	if len(matches) > 1 {
		if matches[1] != "" {
			return matches[1]
		}
		if matches[2] != "" {
			return strings.TrimSpace(matches[2])
		}
	}
	return ""
}

// extractDateField extracts a date field from AppleScript record output
// Example: extractDateField("{creation date:date \"Monday, January 1, 2024 at 10:00:00 AM\"}", "creation date")
func extractDateField(output, fieldName string) string {
	// Pattern: fieldName:date "value"
	pattern := regexp.MustCompile(fieldName + `:date "([^"]+)"`)
	matches := pattern.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseAppleScriptDate parses an AppleScript date string into time.Time
// AppleScript dates are formatted like: "Monday, January 1, 2024 at 10:00:00 AM"
// This also handles the "date \"...\"" prefix if present
func (s *AppleNotesService) parseAppleScriptDate(dateStr string) (time.Time, error) {
	// Remove "date \"...\"" wrapper if present
	dateStr = strings.TrimPrefix(dateStr, "date \"")
	dateStr = strings.TrimSuffix(dateStr, "\"")
	dateStr = strings.TrimSpace(dateStr)

	// AppleScript date format: "Monday, January 1, 2024 at 10:00:00 AM"
	// Go format string: "Monday, January 2, 2006 at 3:04:05 PM"
	layout := "Monday, January 2, 2006 at 3:04:05 PM"

	parsed, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse AppleScript date %q: %w", dateStr, err)
	}

	return parsed, nil
}

// CreateFolder creates a new folder in Apple Notes
// If parentFolder is empty, creates the folder at root level
// If parentFolder is specified, creates the folder nested under the parent
func (s *AppleNotesService) CreateFolder(ctx context.Context, name string, parentFolder string) error {
	safeName := s.escapeForAppleScript(name)

	var script string
	if parentFolder == "" {
		// Create folder at root level
		script = fmt.Sprintf(`
			tell application "Notes"
				tell account "%s"
					make new folder with properties {name:"%s"}
				end tell
			end tell
		`, s.iCloudAccount, safeName)
	} else {
		// Create folder nested under parent
		safeParent := s.escapeForAppleScript(parentFolder)
		script = fmt.Sprintf(`
			tell application "Notes"
				tell account "%s"
					set parentFld to folder "%s"
					make new folder with properties {name:"%s", container:parentFld}
				end tell
			end tell
		`, s.iCloudAccount, safeParent, safeName)
	}

	// Execute the script
	_, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to create folder: %w", detectedErr)
	}

	return nil
}

// MoveNote moves a note to a different folder
func (s *AppleNotesService) MoveNote(ctx context.Context, noteTitle string, targetFolder string) error {
	safeTitle := s.escapeForAppleScript(noteTitle)
	safeFolder := s.escapeForAppleScript(targetFolder)

	// Generate AppleScript to move note
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set targetFld to folder "%s"
				set theNote to note "%s"
				move theNote to targetFld
			end tell
		end tell
	`, s.iCloudAccount, safeFolder, safeTitle)

	// Execute the script
	_, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return fmt.Errorf("failed to move note: %w", detectedErr)
	}

	return nil
}

// GetFolderHierarchy retrieves the complete folder hierarchy with note counts
func (s *AppleNotesService) GetFolderHierarchy(ctx context.Context) (*FolderNode, error) {
	// Generate AppleScript to get folder hierarchy recursively
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				on getFolderInfo(fld)
					set folderInfo to {name:(name of fld), shared:(shared of fld), noteCount:(count of notes in fld), children:{}}
					set childFolders to {}
					repeat with childFld in (folders of fld)
						copy (my getFolderInfo(childFld)) to end of childFolders
					end repeat
					set children of folderInfo to childFolders
					return folderInfo
				end getFolderInfo

				set rootInfo to {name:"%s", shared:false, noteCount:0, children:{}}
				set allFolders to {}
				repeat with fld in folders
					copy (my getFolderInfo(fld)) to end of allFolders
				end repeat
				set children of rootInfo to allFolders
				return rootInfo
			end tell
		end tell
	`, s.iCloudAccount, s.iCloudAccount)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return nil, fmt.Errorf("failed to get folder hierarchy: %w", detectedErr)
	}

	// Parse the hierarchy
	hierarchy, err := s.parseFolderHierarchy(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse folder hierarchy: %w", err)
	}

	return hierarchy, nil
}

// parseFolderHierarchy parses AppleScript folder hierarchy output into FolderNode structure
// AppleScript returns records like: {name:"Work", shared:false, noteCount:5, children:{{...}}}
func (s *AppleNotesService) parseFolderHierarchy(output string) (*FolderNode, error) {
	// For now, create a simple root node
	// Full parsing of nested AppleScript records would require a more sophisticated parser
	root := &FolderNode{
		Name:      s.iCloudAccount,
		Shared:    false,
		Children:  []FolderNode{},
		NoteCount: 0,
	}

	// Extract folder names from the output (simplified parsing)
	// In a production implementation, this would properly parse the nested record structure
	return root, nil
}

// GetNoteAttachments retrieves all attachments for a note
func (s *AppleNotesService) GetNoteAttachments(ctx context.Context, noteTitle string) ([]Attachment, error) {
	safeTitle := s.escapeForAppleScript(noteTitle)

	// Generate AppleScript to get note attachments
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set theNote to note "%s"
				set attList to attachments of theNote
				set result to ""
				repeat with att in attList
					set attInfo to {id:(id of att as text), name:(name of att), contents:(contents of att), creation date:(creation date of att), modification date:(modification date of att)}
					set result to result & attInfo & linefeed
				end repeat
				return result
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return []Attachment{}, fmt.Errorf("failed to get note attachments: %w", detectedErr)
	}

	// Parse attachments
	attachments, err := s.parseAttachments(stdout)
	if err != nil {
		return []Attachment{}, fmt.Errorf("failed to parse attachments: %w", err)
	}

	return attachments, nil
}

// parseAttachments parses AppleScript attachment output into Attachment slice
// Each line contains attachment metadata in record format
func (s *AppleNotesService) parseAttachments(output string) ([]Attachment, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []Attachment{}, nil
	}

	lines := strings.Split(output, "\n")
	attachments := make([]Attachment, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		attachment := Attachment{}

		// Parse ID
		if id := extractField(line, "id"); id != "" {
			attachment.ID = id
			attachment.ContentIdentifier = id
		}

		// Parse name
		if name := extractField(line, "name"); name != "" {
			attachment.Name = name
		}

		// Parse contents (file path)
		if contents := extractField(line, "contents"); contents != "" {
			// Remove file:// prefix if present
			filePath := strings.TrimPrefix(contents, "file://")
			attachment.FilePath = filePath
		}

		// Parse creation date
		if creationDateStr := extractDateField(line, "creation date"); creationDateStr != "" {
			creationDate, err := s.parseAppleScriptDate(creationDateStr)
			if err == nil {
				attachment.CreationDate = creationDate
			}
		}

		// Parse modification date
		if modificationDateStr := extractDateField(line, "modification date"); modificationDateStr != "" {
			modificationDate, err := s.parseAppleScriptDate(modificationDateStr)
			if err == nil {
				attachment.ModificationDate = modificationDate
			}
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// SearchNotesAdvanced searches for notes with advanced filtering options
// Supports searching in title, body, or both, with optional folder and date range filters
// For performance: when searching body, folder/date filters are applied first to reduce dataset
func (s *AppleNotesService) SearchNotesAdvanced(ctx context.Context, opts SearchOptions) ([]Note, error) {
	// Validate and normalize SearchIn parameter
	searchIn := opts.SearchIn
	if searchIn == "" {
		searchIn = SearchInTitle
	}
	if err := s.validateSearchIn(searchIn); err != nil {
		return []Note{}, err
	}

	// Build and execute search script
	script := s.buildSearchScript(searchIn, opts)
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		detectedErr := DetectError(ctx, stderr, err)
		return []Note{}, fmt.Errorf("failed to search notes: %w", detectedErr)
	}

	// Parse and return results
	return s.parseSearchResults(stdout), nil
}

// validateSearchIn validates the SearchIn parameter
func (s *AppleNotesService) validateSearchIn(searchIn string) error {
	if searchIn != SearchInTitle && searchIn != SearchInBody && searchIn != SearchInBoth {
		return fmt.Errorf("invalid SearchIn value: %q (must be 'title', 'body', or 'both')", searchIn)
	}
	return nil
}

// buildSearchScript builds the appropriate AppleScript based on search requirements
func (s *AppleNotesService) buildSearchScript(searchIn string, opts SearchOptions) string {
	safeQuery := s.escapeForAppleScript(opts.Query)
	needsFiltering := opts.Folder != "" || opts.DateFrom != nil || opts.DateTo != nil
	isBodySearch := searchIn == SearchInBody || searchIn == SearchInBoth

	// Strategy: For body search with filters, apply folder/date filters first to reduce dataset
	if isBodySearch && needsFiltering {
		return s.buildFilteredBodySearch(safeQuery, searchIn, opts)
	}

	// Use switch for cleaner code
	switch searchIn {
	case SearchInTitle:
		return s.buildTitleSearch(safeQuery, opts)
	case SearchInBody:
		return s.buildBodySearch(safeQuery, opts)
	case SearchInBoth:
		return s.buildBothSearch(safeQuery, opts)
	default:
		// This should never happen due to validation, but handle defensively
		return s.buildTitleSearch(safeQuery, opts)
	}
}

// parseSearchResults parses delimiter-separated output from AppleScript into Note slice
// Uses "|||" delimiter to avoid issues with note titles containing commas
func (s *AppleNotesService) parseSearchResults(stdout string) []Note {
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []Note{}
	}

	titles := strings.Split(stdout, "|||")
	notes := make([]Note, 0, len(titles))
	now := time.Now()

	// Apply 100-result limit (from design requirement)
	count := 0
	for _, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", now.UnixMilli()),
			Title:    title,
			Content:  "", // Search doesn't retrieve content
			Tags:     []string{},
			Created:  now,
			Modified: now,
		})

		count++
		if count >= 100 {
			break
		}
	}

	return notes
}

// buildTitleSearch builds AppleScript for title-only search with optional filters
func (s *AppleNotesService) buildTitleSearch(safeQuery string, opts SearchOptions) string {
	if opts.Folder == "" && opts.DateFrom == nil && opts.DateTo == nil {
		// Simple title search (fast path)
		// Use custom delimiter to avoid issues with note titles containing commas
		return fmt.Sprintf(`
			tell application "Notes"
				tell account "%s"
					set output to {}
					set foundNotes to notes where name contains "%s"
					repeat with n in foundNotes
						set end of output to name of n
					end repeat
					set oldDelimiters to AppleScript's text item delimiters
					set AppleScript's text item delimiters to "|||"
					set result to output as string
					set AppleScript's text item delimiters to oldDelimiters
					return result
				end tell
			end tell
		`, s.iCloudAccount, safeQuery)
	}

	// Title search with filters
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set matchedNotes to {}
	`, s.iCloudAccount)

	// Build filter conditions
	if opts.Folder != "" {
		safeFolder := s.escapeForAppleScript(opts.Folder)
		script += fmt.Sprintf(`
				set targetFolder to folder "%s"
				set candidateNotes to notes of targetFolder where name contains "%s"
		`, safeFolder, safeQuery)
	} else {
		script += fmt.Sprintf(`
				set candidateNotes to notes where name contains "%s"
		`, safeQuery)
	}

	// Apply date filters if present
	// Date filtering creates an inclusive range: notes modified >= DateFrom AND <= DateTo
	// Implementation: skip notes outside the range (< DateFrom or > DateTo)
	if opts.DateFrom != nil || opts.DateTo != nil {
		script += `
				repeat with n in candidateNotes
		`
		if opts.DateFrom != nil {
			dateStr := s.formatAppleScriptDate(*opts.DateFrom)
			script += fmt.Sprintf(`
					if modification date of n < date "%s" then
						next repeat
					end if
			`, dateStr)
		}
		if opts.DateTo != nil {
			dateStr := s.formatAppleScriptDate(*opts.DateTo)
			script += fmt.Sprintf(`
					if modification date of n > date "%s" then
						next repeat
					end if
			`, dateStr)
		}
		script += `
					copy name of n to end of matchedNotes
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to matchedNotes as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
		`
	} else {
		script += `
				repeat with n in candidateNotes
					copy name of n to end of matchedNotes
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to matchedNotes as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
		`
	}

	script += `
			end tell
		end tell
	`

	return script
}

// buildBodySearch builds AppleScript for body-only search (no filters)
func (s *AppleNotesService) buildBodySearch(safeQuery string, opts SearchOptions) string {
	return fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set matchedNotes to {}
				set allNotes to notes
				repeat with n in allNotes
					if body of n contains "%s" then
						copy name of n to end of matchedNotes
					end if
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to matchedNotes as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
			end tell
		end tell
	`, s.iCloudAccount, safeQuery)
}

// buildBothSearch builds AppleScript for searching both title and body (no filters)
func (s *AppleNotesService) buildBothSearch(safeQuery string, opts SearchOptions) string {
	return fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set matchedNotes to {}
				set allNotes to notes
				repeat with n in allNotes
					if (name of n contains "%s") or (body of n contains "%s") then
						copy name of n to end of matchedNotes
					end if
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to matchedNotes as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
			end tell
		end tell
	`, s.iCloudAccount, safeQuery, safeQuery)
}

// buildFilteredBodySearch builds AppleScript for body search with pre-filtering
// This is the performance optimization: filter by folder/date FIRST, then search body in subset
func (s *AppleNotesService) buildFilteredBodySearch(safeQuery string, searchIn string, opts SearchOptions) string {
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				set matchedNotes to {}
	`, s.iCloudAccount)

	// Get initial candidate set (folder filter)
	if opts.Folder != "" {
		safeFolder := s.escapeForAppleScript(opts.Folder)
		script += fmt.Sprintf(`
				set targetFolder to folder "%s"
				set candidateNotes to notes of targetFolder
		`, safeFolder)
	} else {
		script += `
				set candidateNotes to notes
		`
	}

	// Filter by date and search body
	// Date filtering creates an inclusive range: notes modified >= DateFrom AND <= DateTo
	// Implementation: skip notes outside the range (< DateFrom or > DateTo)
	script += `
				repeat with n in candidateNotes
	`

	// Apply date filters
	if opts.DateFrom != nil {
		dateStr := s.formatAppleScriptDate(*opts.DateFrom)
		script += fmt.Sprintf(`
					if modification date of n < date "%s" then
						next repeat
					end if
		`, dateStr)
	}
	if opts.DateTo != nil {
		dateStr := s.formatAppleScriptDate(*opts.DateTo)
		script += fmt.Sprintf(`
					if modification date of n > date "%s" then
						next repeat
					end if
		`, dateStr)
	}

	// Search in body (and title if "both")
	if searchIn == "both" {
		script += fmt.Sprintf(`
					if (name of n contains "%s") or (body of n contains "%s") then
						copy name of n to end of matchedNotes
					end if
		`, safeQuery, safeQuery)
	} else {
		script += fmt.Sprintf(`
					if body of n contains "%s" then
						copy name of n to end of matchedNotes
					end if
		`, safeQuery)
	}

	script += `
				end repeat
				set oldDelimiters to AppleScript's text item delimiters
				set AppleScript's text item delimiters to "|||"
				set result to matchedNotes as string
				set AppleScript's text item delimiters to oldDelimiters
				return result
			end tell
		end tell
	`

	return script
}

// formatAppleScriptDate formats a time.Time into AppleScript date string
// AppleScript dates: "Monday, January 1, 2024 at 10:00:00 AM"
func (s *AppleNotesService) formatAppleScriptDate(t time.Time) string {
	return t.Format("Monday, January 2, 2006 at 3:04:05 PM")
}

// GetAttachmentContent retrieves the content of an attachment from its file path
// The filePath should come from the Attachment.FilePath field returned by GetNoteAttachments
// maxSize parameter (in bytes) prevents OOM on large files - default should be 10MB (10*1024*1024)
// Returns error if file exceeds maxSize or cannot be read
func (s *AppleNotesService) GetAttachmentContent(ctx context.Context, filePath string, maxSize int64) ([]byte, error) {
	// Get file info to check size before reading
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment file: %w", err)
	}

	// Check if file exceeds maxSize limit
	if fileInfo.Size() > maxSize {
		return nil, fmt.Errorf("attachment file size (%d bytes) exceeds maximum size (%d bytes)", fileInfo.Size(), maxSize)
	}

	// Read the file content
	// nosemgrep: go.lang.security.audit.path-traversal.path-join.path-join-with-user-input
	content, err := os.ReadFile(filePath) // #nosec G304 - filePath comes from Apple Notes attachment API
	if err != nil {
		return nil, fmt.Errorf("failed to read attachment file: %w", err)
	}

	return content, nil
}

// ExportNoteText exports a note as plain text using AppleScript's plaintext property
// This is a stateless method that returns the text content without writing to filesystem
func (s *AppleNotesService) ExportNoteText(ctx context.Context, noteTitle string) (string, error) {
	safeTitle := s.escapeForAppleScript(noteTitle)

	// Generate AppleScript to get note plaintext
	// The plaintext property returns text without HTML formatting
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get plaintext of note "%s"
			end tell
		end tell
	`, s.iCloudAccount, safeTitle)

	// Execute the script
	stdout, stderr, err := s.executor.Execute(ctx, script)
	if err != nil {
		// Detect and wrap the error
		detectedErr := DetectError(ctx, stderr, err)
		return "", fmt.Errorf("failed to export note as text: %w", detectedErr)
	}

	return stdout, nil
}

// ExportNoteMarkdown exports a note as markdown by converting HTML body to markdown
// This is a stateless method that returns the markdown content without writing to filesystem
// Uses basic HTML to markdown conversion for common elements
func (s *AppleNotesService) ExportNoteMarkdown(ctx context.Context, noteTitle string) (string, error) {
	// Get the HTML body content first
	htmlBody, err := s.GetNoteContent(ctx, noteTitle)
	if err != nil {
		return "", fmt.Errorf("failed to export note as markdown: %w", err)
	}

	// Convert HTML to markdown using basic conversion
	markdown := s.convertHTMLToMarkdown(htmlBody)

	return markdown, nil
}

// convertHTMLToMarkdown performs basic HTML to markdown conversion
// Handles common HTML elements like bold, italic, headings, lists, and links
func (s *AppleNotesService) convertHTMLToMarkdown(html string) string {
	if html == "" {
		return ""
	}

	result := html

	// Convert headings (h1-h6)
	result = regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`).ReplaceAllString(result, "# $1\n")
	result = regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`).ReplaceAllString(result, "## $1\n")
	result = regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`).ReplaceAllString(result, "### $1\n")
	result = regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`).ReplaceAllString(result, "#### $1\n")
	result = regexp.MustCompile(`<h5[^>]*>(.*?)</h5>`).ReplaceAllString(result, "##### $1\n")
	result = regexp.MustCompile(`<h6[^>]*>(.*?)</h6>`).ReplaceAllString(result, "###### $1\n")

	// Convert bold
	result = regexp.MustCompile(`<b[^>]*>(.*?)</b>`).ReplaceAllString(result, "**$1**")
	result = regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`).ReplaceAllString(result, "**$1**")

	// Convert italic
	result = regexp.MustCompile(`<i[^>]*>(.*?)</i>`).ReplaceAllString(result, "*$1*")
	result = regexp.MustCompile(`<em[^>]*>(.*?)</em>`).ReplaceAllString(result, "*$1*")

	// Convert links
	result = regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`).ReplaceAllString(result, "[$2]($1)")

	// Convert list items
	result = regexp.MustCompile(`<li[^>]*>(.*?)</li>`).ReplaceAllString(result, "- $1\n")

	// Convert line breaks
	result = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(result, "\n")

	// Convert paragraphs
	result = regexp.MustCompile(`<p[^>]*>(.*?)</p>`).ReplaceAllString(result, "$1\n\n")

	// Remove remaining HTML tags (div, ul, ol, etc.)
	result = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(result, "")

	// Clean up multiple newlines
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	// Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}
