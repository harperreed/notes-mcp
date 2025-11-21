// ABOUTME: Delete command for deleting notes in Apple Notes
// ABOUTME: Accepts title via CLI argument

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <title>",
	Short: "Delete a note from Apple Notes",
	Long:  `Deletes a note from Apple Notes identified by its title.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Delete the note
		err := notesService.DeleteNote(ctx, title)
		if err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}

		// Output success message
		fmt.Printf("Note deleted: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
