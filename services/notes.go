// ABOUTME: Notes service interface and core business logic
// ABOUTME: Defines the contract for notes management operations

package services

import (
	"context"
)

// NotesService defines the interface for notes management
type NotesService interface {
	// CreateNote creates a new note
	CreateNote(ctx context.Context, title, content string) (string, error)

	// GetNote retrieves a note by ID
	GetNote(ctx context.Context, id string) (Note, error)

	// UpdateNote updates an existing note
	UpdateNote(ctx context.Context, id, title, content string) error

	// DeleteNote removes a note
	DeleteNote(ctx context.Context, id string) error

	// ListNotes returns all notes
	ListNotes(ctx context.Context) ([]Note, error)
}

// Note represents a note entity
type Note struct {
	ID      string
	Title   string
	Content string
}
