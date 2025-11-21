// ABOUTME: Export text command for exporting notes to plain text format
// ABOUTME: Accepts note title and returns plain text content

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportTextCmd = &cobra.Command{
	Use:   "export-text <note-title>",
	Short: "Export a note to plain text format",
	Long:  `Exports a note from Apple Notes to plain text format, stripping all HTML formatting.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteTitle := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Export to text
		text, err := notesService.ExportNoteText(ctx, noteTitle)
		if err != nil {
			return fmt.Errorf("failed to export note to text: %w", err)
		}

		// Output text content
		fmt.Print(text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportTextCmd)
}
