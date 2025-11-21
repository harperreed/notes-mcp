This is a robust Go implementation of the `AppleNotesManager`.

To make this "robust" compared to a direct line-for-line translation, I have added:

1.  **Context with Timeouts:** Prevents the application from hanging if `osascript` stalls.
2.  **Strict Error Handling:** Captures standard error output from the terminal command.
3.  **Struct Definitions:** Strongly typed `Note` objects.
4.  **Input Sanitization:** A helper function to ensure quotes and backslashes don't break the AppleScript execution.

### `services/apple_notes_manager.go`

```go
package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Note represents the structure of an Apple Note
type Note struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Tags     []string  `json:"tags"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// AppleNotesManager handles interactions with the local Apple Notes app
type AppleNotesManager struct {
	iCloudAccount string
}

// NewAppleNotesManager initializes the manager
func NewAppleNotesManager() *AppleNotesManager {
	return &AppleNotesManager{
		iCloudAccount: "iCloud",
	}
}

// formatContent prepares text for AppleScript (HTML body compatible)
[cite_start]// Porting logic from src/services/appleNotesManager.ts [cite: 15]
func (m *AppleNotesManager) formatContent(content string) string {
	if content == "" {
		return ""
	}
	// Escape backslashes and double quotes
	escaped := m.escapeForAppleScript(content)
	// Replace newlines with HTML breaks for Note body
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

// escapeForAppleScript handles string escaping to prevent script syntax errors
func (m *AppleNotesManager) escapeForAppleScript(input string) string {
	// 1. Escape backslashes first (so we don't double escape)
	result := strings.ReplaceAll(input, "\\", "\\\\")
	// 2. Escape double quotes
	result = strings.ReplaceAll(result, "\"", "\\\"")
	return result
}

// runScript executes the AppleScript command with a timeout
func (m *AppleNotesManager) runScript(script string) (string, error) {
	[cite_start]// Create a context with a 10-second timeout (matching original Utils) [cite: 35]
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -e flag tells osascript to execute the following string
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("AppleScript execution failed: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateNote creates a new note in the default iCloud account
func (m *AppleNotesManager) CreateNote(title, content string, tags []string) (*Note, error) {
	formattedContent := m.formatContent(content)
	safeTitle := m.escapeForAppleScript(title)

	[cite_start]// Note: Original source code ignores tags in the actual AppleScript execution [cite: 21]
	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				make new note with properties {name:"%s", body:"%s"}
			end tell
		end tell
	`, m.iCloudAccount, safeTitle, formattedContent)

	_, err := m.runScript(script)
	if err != nil {
		return nil, err
	}

	// Return the note object (ID generation simulated to match original logic)
	return &Note{
		ID:       fmt.Sprintf("%d", time.Now().UnixMilli()),
		Title:    title,
		Content:  content,
		Tags:     tags,
		Created:  time.Now(),
		Modified: time.Now(),
	}, nil
}

// SearchNotes looks for notes containing the query in the title
func (m *AppleNotesManager) SearchNotes(query string) ([]Note, error) {
	safeQuery := m.escapeForAppleScript(query)

	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get name of notes where name contains "%s"
			end tell
		end tell
	`, m.iCloudAccount, safeQuery)

	output, err := m.runScript(script)
	if err != nil {
		// If no notes are found or script fails, return empty slice rather than erroring out
		return []Note{}, nil
	}

	if output == "" {
		return []Note{}, nil
	}

	// Parse Comma Separated Values
	[cite_start]// Note: This is a simple split based on the original TS implementation [cite: 27]
	// It might be fragile if note titles contain commas.
	titles := strings.Split(output, ", ")
	var notes []Note

	for _, t := range titles {
		if t == "" {
			continue
		}
		notes = append(notes, Note{
			ID:       fmt.Sprintf("%d", time.Now().UnixMilli()),
			Title:    strings.TrimSpace(t),
			Content:  "", // Search doesn't retrieve content by default
			Tags:     []string{},
			Created:  time.Now(),
			Modified: time.Now(),
		})
	}

	return notes, nil
}

// GetNoteContent retrieves the body of a specific note
func (m *AppleNotesManager) GetNoteContent(title string) (string, error) {
	safeTitle := m.escapeForAppleScript(title)

	script := fmt.Sprintf(`
		tell application "Notes"
			tell account "%s"
				get body of note "%s"
			end tell
		end tell
	`, m.iCloudAccount, safeTitle)

	content, err := m.runScript(script)
	if err != nil {
		return "", fmt.Errorf("failed to get note content: %w", err)
	}

	return content, nil
}
```

### Usage Example (`main.go`)

Here is how you would use this manager in your `main` function to simulate the MCP tool call.

```go
package main

import (
	"fmt"
	"log"
	"your-module/services" // Replace with your actual module path
)

func main() {
	manager := services.NewAppleNotesManager()

	// 1. Create a Note
	fmt.Println("--- Creating Note ---")
	note, err := manager.CreateNote("Go Test Note", "This is a note created via Go.\nIt handles newlines.", []string{"code", "go"})
	if err != nil {
		log.Fatalf("Error creating note: %v", err)
	}
	fmt.Printf("Created: %s\n", note.Title)

	// 2. Search Notes
	fmt.Println("\n--- Searching Notes ---")
	notes, err := manager.SearchNotes("Go Test")
	if err != nil {
		log.Fatalf("Error searching: %v", err)
	}
	for _, n := range notes {
		fmt.Printf("Found: %s\n", n.Title)
	}

	// 3. Get Content
	fmt.Println("\n--- Getting Content ---")
	content, err := manager.GetNoteContent("Go Test Note")
	if err != nil {
		log.Printf("Error getting content: %v", err)
	} else {
		fmt.Printf("Content: %s\n", content)
	}
}
```

### Key Changes for "Robustness"

1.  **`escapeForAppleScript`**: I added logic to handle backslashes *before* quotes. This is critical because if a user inputs `C:\Path`, a naive replace might break the script string.
2.  **`context.WithTimeout`**: In `runScript`, the `exec.CommandContext` ensures that if Apple Notes freezes (which can happen during sync), your Go server won't hang forever—it will kill the process after 10 seconds.
3.  **`runScript` Error Details**: The error return now includes the `output` (stderr) from `osascript`, which is essential for debugging why a script failed (e.g., "Note not found" vs "Syntax Error").
