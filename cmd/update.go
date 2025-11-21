// ABOUTME: Update command for updating existing notes in Apple Notes
// ABOUTME: Accepts title and new content via CLI arguments

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <title> <content>",
	Short: "Update an existing note in Apple Notes",
	Long:  `Updates the content of an existing note in Apple Notes identified by its title.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		content := args[1]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Update the note
		err := notesService.UpdateNote(ctx, title, content)
		if err != nil {
			return fmt.Errorf("failed to update note: %w", err)
		}

		// Output success message
		fmt.Printf("Note updated: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
