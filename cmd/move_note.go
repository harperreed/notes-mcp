// ABOUTME: Move note command for moving notes between folders in Apple Notes
// ABOUTME: Accepts note title and target folder path

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var moveNoteCmd = &cobra.Command{
	Use:   "move-note <note-title> <target-folder>",
	Short: "Move a note to a different folder in Apple Notes",
	Long:  `Moves a note to the specified target folder in Apple Notes.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteTitle := args[0]
		targetFolder := args[1]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Move the note
		err := notesService.MoveNote(ctx, noteTitle, targetFolder)
		if err != nil {
			return fmt.Errorf("failed to move note: %w", err)
		}

		// Output success message
		fmt.Printf("Note '%s' moved to folder '%s'\n", noteTitle, targetFolder)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveNoteCmd)
}
