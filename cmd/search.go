// ABOUTME: Search command for finding notes in Apple Notes
// ABOUTME: Accepts a query string and returns matching note titles

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for notes in Apple Notes",
	Long:  `Searches for notes in Apple Notes by title. Returns a newline-separated list of matching note titles.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Search for notes
		notes, err := notesService.SearchNotes(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		// Output newline-separated list of titles
		for _, note := range notes {
			fmt.Println(note.Title)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
