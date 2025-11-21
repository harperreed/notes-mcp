// ABOUTME: Integration tests for NotesService with real Apple Notes
// ABOUTME: Only runs with -tags=integration build tag

//go:build integration

package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Integration test helper to create unique test names
func uniqueTestName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixMilli())
}

// TestCreateNoteIntegration tests creating a note with real AppleScript
func TestCreateNoteIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	title := uniqueTestName("IntTest_Create")
	content := "Integration test note created by automated tests"
	tags := []string{"integration", "test"}

	note, err := service.CreateNote(ctx, title, content, tags)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	// Verify the note was created with full metadata
	if note.Title != title {
		t.Errorf("Note title = %q, want %q", note.Title, title)
	}
	if note.ID == "" {
		t.Error("Note ID should be populated")
	}
	if note.Folder == "" {
		t.Error("Note folder should be populated")
	}
	if note.Created.IsZero() {
		t.Error("Note creation date should be set")
	}
	if note.Modified.IsZero() {
		t.Error("Note modification date should be set")
	}

	t.Logf("Created test note: %s (ID: %s, Folder: %s)", title, note.ID, note.Folder)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestSearchNotesIntegration tests searching for notes with real AppleScript
func TestSearchNotesIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// First create a test note to search for
	title := uniqueTestName("IntTest_Search")
	content := "Searchable integration test content"

	_, err := service.CreateNote(ctx, title, content, []string{"integration"})
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	// Wait a moment for the note to be indexed
	time.Sleep(1 * time.Second)

	// Now search for it
	notes, err := service.SearchNotes(ctx, "IntTest_Search")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	// Verify we found at least one note (there may be multiple from previous test runs)
	if len(notes) == 0 {
		t.Error("Expected to find at least one note in search results")
	}

	// Verify the notes have metadata
	found := false
	for _, note := range notes {
		if strings.Contains(note.Title, "IntTest_Search") {
			found = true
			if note.ID == "" {
				t.Error("Search result should have ID")
			}
			if note.Folder == "" {
				t.Error("Search result should have folder")
			}
			if note.Created.IsZero() {
				t.Error("Search result should have creation date")
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find our test note in search results")
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestGetNoteContentIntegration tests retrieving note content with real AppleScript
func TestGetNoteContentIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	title := uniqueTestName("IntTest_GetContent")
	content := "Content to retrieve in integration test"

	_, err := service.CreateNote(ctx, title, content, nil)
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	// Wait a moment for the note to be available
	time.Sleep(1 * time.Second)

	// Get the content
	retrievedContent, err := service.GetNoteContent(ctx, title)
	if err != nil {
		t.Fatalf("GetNoteContent failed: %v", err)
	}

	// Content should contain the text we put in (possibly wrapped in HTML)
	if !strings.Contains(retrievedContent, content) {
		t.Errorf("Retrieved content should contain %q, got: %q", content, retrievedContent)
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestCreateFolderIntegration tests creating folders with real AppleScript
func TestCreateFolderIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a root-level folder
	folderName := uniqueTestName("IntTest_Folder")

	err := service.CreateFolder(ctx, folderName, "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	t.Logf("Created test folder: %s", folderName)

	// Verify the folder exists by listing folders
	folders, err := service.ListFolders(ctx)
	if err != nil {
		t.Fatalf("ListFolders failed: %v", err)
	}

	found := false
	for _, folder := range folders {
		if folder == folderName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find folder %q in folder list", folderName)
	}

	t.Log("Manual cleanup required in Apple Notes (delete the test folder)")
}

// TestCreateNestedFolderIntegration tests creating nested folders with real AppleScript
func TestCreateNestedFolderIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a parent folder first
	parentName := uniqueTestName("IntTest_Parent")
	err := service.CreateFolder(ctx, parentName, "")
	if err != nil {
		t.Fatalf("CreateFolder (parent) failed: %v", err)
	}

	// Wait a moment for the folder to be available
	time.Sleep(1 * time.Second)

	// Create a nested folder
	childName := uniqueTestName("IntTest_Child")
	err = service.CreateFolder(ctx, childName, parentName)
	if err != nil {
		t.Fatalf("CreateFolder (nested) failed: %v", err)
	}

	t.Logf("Created nested folder: %s/%s", parentName, childName)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestMoveNoteIntegration tests moving notes between folders with real AppleScript
func TestMoveNoteIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a target folder
	targetFolder := uniqueTestName("IntTest_TargetFolder")
	err := service.CreateFolder(ctx, targetFolder, "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Create a test note
	noteTitle := uniqueTestName("IntTest_MoveNote")
	_, err = service.CreateNote(ctx, noteTitle, "Test note to move", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Move the note to the target folder
	err = service.MoveNote(ctx, noteTitle, targetFolder)
	if err != nil {
		t.Fatalf("MoveNote failed: %v", err)
	}

	// Verify the note was moved by getting its metadata
	time.Sleep(1 * time.Second)
	metadata, err := service.GetNoteMetadata(ctx, noteTitle)
	if err != nil {
		t.Fatalf("GetNoteMetadata failed: %v", err)
	}

	if metadata.Folder != targetFolder {
		t.Errorf("Note folder = %q, want %q", metadata.Folder, targetFolder)
	}

	t.Logf("Moved note %q to folder %q", noteTitle, targetFolder)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestGetFolderHierarchyIntegration tests retrieving folder hierarchy with real AppleScript
func TestGetFolderHierarchyIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	hierarchy, err := service.GetFolderHierarchy(ctx)
	if err != nil {
		t.Fatalf("GetFolderHierarchy failed: %v", err)
	}

	// Verify the hierarchy has a root
	if hierarchy == nil {
		t.Fatal("Expected non-nil hierarchy")
	}

	if hierarchy.Name == "" {
		t.Error("Hierarchy root should have a name")
	}

	// The root should have at least some folders (iCloud account typically has default folders)
	if len(hierarchy.Children) == 0 {
		t.Log("Warning: No folders found in hierarchy (this may be normal for a fresh account)")
	}

	t.Logf("Folder hierarchy root: %s (note count: %d, children: %d)",
		hierarchy.Name, hierarchy.NoteCount, len(hierarchy.Children))

	// Log the folder structure for verification
	logFolderHierarchy(t, hierarchy, 0)
}

// Helper function to log folder hierarchy
func logFolderHierarchy(t *testing.T, node *FolderNode, depth int) {
	indent := strings.Repeat("  ", depth)
	t.Logf("%s- %s (notes: %d, shared: %v, children: %d)",
		indent, node.Name, node.NoteCount, node.Shared, len(node.Children))

	for _, child := range node.Children {
		logFolderHierarchy(t, &child, depth+1)
	}
}

// TestSearchNotesAdvanced_TitleIntegration tests advanced title search with real AppleScript
func TestSearchNotesAdvanced_TitleIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	title := uniqueTestName("IntTest_AdvSearch_Title")
	_, err := service.CreateNote(ctx, title, "Content for advanced search test", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Search with advanced options - title only
	opts := SearchOptions{
		Query:    "IntTest_AdvSearch_Title",
		SearchIn: "title",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced failed: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected to find at least one note")
	}

	// Verify notes have metadata
	for _, note := range notes {
		if note.ID == "" {
			t.Error("Search result should have ID")
		}
		if note.Folder == "" {
			t.Error("Search result should have folder")
		}
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestSearchNotesAdvanced_BodyIntegration tests advanced body search with real AppleScript
func TestSearchNotesAdvanced_BodyIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note with unique content
	title := uniqueTestName("IntTest_BodySearch")
	uniqueWord := fmt.Sprintf("UNIQUE_BODY_CONTENT_%d", time.Now().UnixMilli())
	content := fmt.Sprintf("This note contains a unique word: %s", uniqueWord)

	_, err := service.CreateNote(ctx, title, content, nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(2 * time.Second) // Body search may need more time to index

	// Search with advanced options - body only
	opts := SearchOptions{
		Query:    uniqueWord,
		SearchIn: "body",
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced (body) failed: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected to find at least one note with body search")
	}

	found := false
	for _, note := range notes {
		if note.Title == title {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find note %q in body search results", title)
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestSearchNotesAdvanced_FolderFilterIntegration tests filtering by folder with real AppleScript
func TestSearchNotesAdvanced_FolderFilterIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test folder
	folderName := uniqueTestName("IntTest_SearchFolder")
	err := service.CreateFolder(ctx, folderName, "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Create a note in that folder
	noteTitle := uniqueTestName("IntTest_FolderSearch_Note")
	_, err = service.CreateNote(ctx, noteTitle, "Content in specific folder", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Move note to the test folder
	err = service.MoveNote(ctx, noteTitle, folderName)
	if err != nil {
		t.Fatalf("MoveNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Search with folder filter
	opts := SearchOptions{
		Query:    "IntTest_FolderSearch",
		SearchIn: "title",
		Folder:   folderName,
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced (folder filter) failed: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected to find at least one note in folder")
	}

	// Verify all notes are in the specified folder
	for _, note := range notes {
		if note.Folder != folderName {
			t.Errorf("Note folder = %q, want %q", note.Folder, folderName)
		}
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestSearchNotesAdvanced_DateFilterIntegration tests filtering by date range with real AppleScript
func TestSearchNotesAdvanced_DateFilterIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	noteTitle := uniqueTestName("IntTest_DateFilter")
	_, err := service.CreateNote(ctx, noteTitle, "Content for date filter test", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Search with date filter - from 1 hour ago to now
	now := time.Now()
	dateFrom := now.Add(-1 * time.Hour)

	opts := SearchOptions{
		Query:    "IntTest_DateFilter",
		SearchIn: "title",
		DateFrom: &dateFrom,
		DateTo:   &now,
	}

	notes, err := service.SearchNotesAdvanced(ctx, opts)
	if err != nil {
		t.Fatalf("SearchNotesAdvanced (date filter) failed: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected to find at least one note within date range")
	}

	// Verify notes are within date range
	for _, note := range notes {
		if note.ModificationDate.Before(dateFrom) {
			t.Errorf("Note modification date %v is before dateFrom %v", note.ModificationDate, dateFrom)
		}
		if note.ModificationDate.After(now) {
			t.Errorf("Note modification date %v is after dateTo %v", note.ModificationDate, now)
		}
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestGetNoteAttachmentsIntegration tests retrieving attachments from a note
// Note: This test requires manual setup - create a note with attachments before running
func TestGetNoteAttachmentsIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// This test requires a note with attachments to exist
	// We'll create a note, but it won't have attachments automatically
	noteTitle := uniqueTestName("IntTest_Attachments")
	_, err := service.CreateNote(ctx, noteTitle, "Note for attachment testing", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Get attachments (should be empty for a newly created note)
	attachments, err := service.GetNoteAttachments(ctx, noteTitle)
	if err != nil {
		t.Fatalf("GetNoteAttachments failed: %v", err)
	}

	// New notes won't have attachments
	if len(attachments) > 0 {
		t.Logf("Found %d attachments (unexpected for new note)", len(attachments))

		// Verify attachment structure
		for i, att := range attachments {
			if att.ID == "" {
				t.Errorf("Attachment %d should have ID", i)
			}
			if att.Name == "" {
				t.Errorf("Attachment %d should have name", i)
			}
			if att.FilePath == "" {
				t.Errorf("Attachment %d should have file path", i)
			}
			t.Logf("Attachment %d: %s (%s)", i, att.Name, att.FilePath)
		}
	} else {
		t.Log("No attachments found (expected for new note)")
		t.Log("To fully test attachments, manually add files to a note and re-run this test")
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestGetAttachmentContentIntegration tests retrieving attachment content
// Note: This test requires manual setup - create a note with attachments before running
func TestGetAttachmentContentIntegration(t *testing.T) {
	t.Skip("Skipping attachment content test - requires manual setup of note with attachments")

	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// This would require a real attachment file path
	// Manual test: create a note with an attachment, get its path, and test here
	filePath := "/path/to/real/attachment/file"

	content, err := service.GetAttachmentContent(ctx, filePath, 10*1024*1024)
	if err != nil {
		t.Fatalf("GetAttachmentContent failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty content")
	}

	t.Logf("Retrieved %d bytes of attachment content", len(content))
}

// TestExportNoteTextIntegration tests exporting note to plain text with real AppleScript
func TestExportNoteTextIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	noteTitle := uniqueTestName("IntTest_ExportText")
	noteContent := "This is plain text content for export testing"

	_, err := service.CreateNote(ctx, noteTitle, noteContent, nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Export as text
	text, err := service.ExportNoteText(ctx, noteTitle)
	if err != nil {
		t.Fatalf("ExportNoteText failed: %v", err)
	}

	// Verify the text contains our content
	if !strings.Contains(text, noteContent) {
		t.Errorf("Exported text should contain %q, got: %q", noteContent, text)
	}

	t.Logf("Exported text (%d chars): %s", len(text), text)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestExportNoteMarkdownIntegration tests exporting note to markdown with real AppleScript
func TestExportNoteMarkdownIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note with some formatting
	noteTitle := uniqueTestName("IntTest_ExportMarkdown")
	noteContent := "This is content with bold and italic formatting"

	_, err := service.CreateNote(ctx, noteTitle, noteContent, nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Export as markdown
	markdown, err := service.ExportNoteMarkdown(ctx, noteTitle)
	if err != nil {
		t.Fatalf("ExportNoteMarkdown failed: %v", err)
	}

	// Verify the markdown contains our content
	if markdown == "" {
		t.Error("Expected non-empty markdown")
	}

	// The content should be present in the markdown output
	if !strings.Contains(markdown, "content") {
		t.Errorf("Exported markdown should contain the word 'content', got: %q", markdown)
	}

	t.Logf("Exported markdown (%d chars): %s", len(markdown), markdown)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestGetNotesInFolderIntegration tests retrieving all notes in a specific folder
func TestGetNotesInFolderIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test folder
	folderName := uniqueTestName("IntTest_GetNotesFolder")
	err := service.CreateFolder(ctx, folderName, "")
	if err != nil {
		t.Fatalf("CreateFolder failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Create notes in the folder
	note1Title := uniqueTestName("IntTest_Note1")
	note2Title := uniqueTestName("IntTest_Note2")

	_, err = service.CreateNote(ctx, note1Title, "First note", nil)
	if err != nil {
		t.Fatalf("CreateNote (note1) failed: %v", err)
	}

	_, err = service.CreateNote(ctx, note2Title, "Second note", nil)
	if err != nil {
		t.Fatalf("CreateNote (note2) failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Move both notes to the folder
	err = service.MoveNote(ctx, note1Title, folderName)
	if err != nil {
		t.Fatalf("MoveNote (note1) failed: %v", err)
	}

	err = service.MoveNote(ctx, note2Title, folderName)
	if err != nil {
		t.Fatalf("MoveNote (note2) failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Get all notes in the folder
	notes, err := service.GetNotesInFolder(ctx, folderName)
	if err != nil {
		t.Fatalf("GetNotesInFolder failed: %v", err)
	}

	// Should find at least our 2 notes
	if len(notes) < 2 {
		t.Errorf("Expected at least 2 notes in folder, got %d", len(notes))
	}

	// Verify notes have metadata and are in the correct folder
	foundNote1 := false
	foundNote2 := false

	for _, note := range notes {
		if note.Folder != folderName {
			t.Errorf("Note %q folder = %q, want %q", note.Title, note.Folder, folderName)
		}
		if note.ID == "" {
			t.Errorf("Note %q should have ID", note.Title)
		}

		if note.Title == note1Title {
			foundNote1 = true
		}
		if note.Title == note2Title {
			foundNote2 = true
		}
	}

	if !foundNote1 {
		t.Errorf("Expected to find note %q in folder", note1Title)
	}
	if !foundNote2 {
		t.Errorf("Expected to find note %q in folder", note2Title)
	}

	t.Logf("Found %d notes in folder %q", len(notes), folderName)
	t.Log("Manual cleanup required in Apple Notes")
}

// TestUpdateNoteIntegration tests updating a note with real AppleScript
func TestUpdateNoteIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	noteTitle := uniqueTestName("IntTest_Update")
	originalContent := "Original content"

	_, err := service.CreateNote(ctx, noteTitle, originalContent, nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Update the note
	updatedContent := "Updated content - modified by integration test"
	err = service.UpdateNote(ctx, noteTitle, updatedContent)
	if err != nil {
		t.Fatalf("UpdateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Verify the update
	content, err := service.GetNoteContent(ctx, noteTitle)
	if err != nil {
		t.Fatalf("GetNoteContent failed: %v", err)
	}

	if !strings.Contains(content, updatedContent) {
		t.Errorf("Content should contain updated text, got: %q", content)
	}

	t.Log("Manual cleanup required in Apple Notes")
}

// TestDeleteNoteIntegration tests deleting a note with real AppleScript
func TestDeleteNoteIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a test note
	noteTitle := uniqueTestName("IntTest_Delete")

	_, err := service.CreateNote(ctx, noteTitle, "Note to be deleted", nil)
	if err != nil {
		t.Fatalf("CreateNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Delete the note
	err = service.DeleteNote(ctx, noteTitle)
	if err != nil {
		t.Fatalf("DeleteNote failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Verify the note is gone
	_, err = service.GetNoteContent(ctx, noteTitle)
	if err == nil {
		t.Error("Expected error when getting deleted note, got nil")
	}

	// Should get a "note not found" error
	if !strings.Contains(err.Error(), "note not found") && !strings.Contains(err.Error(), "doesn't understand") {
		t.Logf("Got error (expected): %v", err)
	}

	t.Log("Note successfully deleted - no manual cleanup needed for this test")
}

// TestGetRecentNotesIntegration tests retrieving recently modified notes
func TestGetRecentNotesIntegration(t *testing.T) {
	executor := NewOSAScriptExecutor(30 * time.Second)
	service := NewAppleNotesService(executor)
	ctx := context.Background()

	// Create a couple of test notes
	note1Title := uniqueTestName("IntTest_Recent1")
	note2Title := uniqueTestName("IntTest_Recent2")

	_, err := service.CreateNote(ctx, note1Title, "Recent note 1", nil)
	if err != nil {
		t.Fatalf("CreateNote (note1) failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	_, err = service.CreateNote(ctx, note2Title, "Recent note 2", nil)
	if err != nil {
		t.Fatalf("CreateNote (note2) failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Get recent notes
	notes, err := service.GetRecentNotes(ctx, 10)
	if err != nil {
		t.Fatalf("GetRecentNotes failed: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected at least some recent notes")
	}

	// Our newly created notes should be in the recent list
	foundRecent := 0
	for _, note := range notes {
		if note.Title == note1Title || note.Title == note2Title {
			foundRecent++
		}

		// Verify metadata
		if note.ID == "" {
			t.Errorf("Recent note %q should have ID", note.Title)
		}
		if note.ModificationDate.IsZero() {
			t.Errorf("Recent note %q should have modification date", note.Title)
		}
	}

	if foundRecent < 2 {
		t.Errorf("Expected to find both recent notes, found %d", foundRecent)
	}

	t.Logf("Retrieved %d recent notes", len(notes))
	t.Log("Manual cleanup required in Apple Notes")
}
