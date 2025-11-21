// ABOUTME: Advanced search command for finding notes with filters in Apple Notes
// ABOUTME: Supports searching in title/body/both, folder filtering, and date range filtering

package cmd

import (
	"fmt"
	"time"

	"github.com/harper/notes-mcp/services"
	"github.com/spf13/cobra"
)

var (
	searchIn     string
	searchFolder string
	dateFrom     string
	dateTo       string
)

var searchAdvancedCmd = &cobra.Command{
	Use:   "search-advanced <query>",
	Short: "Advanced search for notes with filters",
	Long:  `Searches for notes in Apple Notes with advanced filtering options including search location (title/body/both), folder, and date range.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		// Parse date flags if provided
		var dateFromPtr, dateToPtr *time.Time
		if dateFrom != "" {
			t, err := time.Parse("2006-01-02", dateFrom)
			if err != nil {
				return fmt.Errorf("invalid date-from format (use YYYY-MM-DD): %w", err)
			}
			dateFromPtr = &t
		}
		if dateTo != "" {
			t, err := time.Parse("2006-01-02", dateTo)
			if err != nil {
				return fmt.Errorf("invalid date-to format (use YYYY-MM-DD): %w", err)
			}
			dateToPtr = &t
		}

		// Build search options
		opts := services.SearchOptions{
			Query:    query,
			SearchIn: searchIn,
			Folder:   searchFolder,
			DateFrom: dateFromPtr,
			DateTo:   dateToPtr,
		}

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Search for notes
		notes, err := notesService.SearchNotesAdvanced(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		// Limit results to prevent timeouts with large result sets
		totalNotes := len(notes)
		if totalNotes > maxSearchResults {
			notes = notes[:maxSearchResults]
		}

		// Output newline-separated list of titles
		for _, note := range notes {
			fmt.Println(note.Title)
		}

		// Add indicator if results were limited
		if totalNotes > maxSearchResults {
			//nolint:errcheck // stderr write failure is non-critical
			fmt.Fprintf(cmd.ErrOrStderr(), "\n(Showing first %d of %d matching notes)\n", maxSearchResults, totalNotes)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchAdvancedCmd)

	// Add flags
	searchAdvancedCmd.Flags().StringVar(&searchIn, "search-in", "title", "Where to search: title, body, or both")
	searchAdvancedCmd.Flags().StringVar(&searchFolder, "folder", "", "Limit search to specific folder")
	searchAdvancedCmd.Flags().StringVar(&dateFrom, "date-from", "", "Filter by creation date from (YYYY-MM-DD)")
	searchAdvancedCmd.Flags().StringVar(&dateTo, "date-to", "", "Filter by creation date to (YYYY-MM-DD)")
}
