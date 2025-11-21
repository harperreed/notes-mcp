// ABOUTME: Attachments command for listing attachments in a note
// ABOUTME: Accepts note title and returns list of attachments with metadata

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var attachmentsCmd = &cobra.Command{
	Use:   "attachments <note-title>",
	Short: "List attachments in a note",
	Long:  `Lists all attachments in a note from Apple Notes, including file paths and metadata.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteTitle := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Get attachments
		attachments, err := notesService.GetNoteAttachments(ctx, noteTitle)
		if err != nil {
			return fmt.Errorf("failed to get attachments: %w", err)
		}

		// Output as JSON
		output, err := json.MarshalIndent(attachments, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format attachments: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(attachmentsCmd)
}
