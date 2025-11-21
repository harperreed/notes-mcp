// ABOUTME: Create command for creating notes in Apple Notes
// ABOUTME: Accepts title, content, and optional tags via CLI arguments and flags

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/harper/notes-mcp/services"
	"github.com/spf13/cobra"
)

var (
	createTags []string
)

var createCmd = &cobra.Command{
	Use:   "create <title> <content>",
	Short: "Create a new note in Apple Notes",
	Long:  `Creates a new note in Apple Notes with the specified title and content. Optionally add tags using the --tags flag.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		content := args[1]

		// Create service with real executor
		executor := services.NewOSAScriptExecutor(10 * time.Second)
		notesService := services.NewAppleNotesService(executor)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create the note
		note, err := notesService.CreateNote(ctx, title, content, createTags)
		if err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		// Output success message
		fmt.Printf("Note created: %s\n", note.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Add flags
	createCmd.Flags().StringSliceVar(&createTags, "tags", []string{}, "Comma-separated list of tags")
}
