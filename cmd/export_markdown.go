// ABOUTME: Export markdown command for exporting notes to markdown format
// ABOUTME: Accepts note title and returns markdown-formatted content

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportMarkdownCmd = &cobra.Command{
	Use:   "export-markdown <note-title>",
	Short: "Export a note to markdown format",
	Long:  `Exports a note from Apple Notes to markdown format, converting HTML content to markdown syntax.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteTitle := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Export to markdown
		markdown, err := notesService.ExportNoteMarkdown(ctx, noteTitle)
		if err != nil {
			return fmt.Errorf("failed to export note to markdown: %w", err)
		}

		// Output markdown content
		fmt.Print(markdown)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportMarkdownCmd)
}
