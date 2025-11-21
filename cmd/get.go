// ABOUTME: Get command for retrieving note content from Apple Notes
// ABOUTME: Accepts a note title and returns the raw HTML content

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/harper/notes-mcp/services"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <title>",
	Short: "Get the content of a note from Apple Notes",
	Long:  `Retrieves the full HTML content of a note from Apple Notes by its title.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Create service with real executor
		executor := services.NewOSAScriptExecutor(10 * time.Second)
		notesService := services.NewAppleNotesService(executor)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get the note content
		content, err := notesService.GetNoteContent(ctx, title)
		if err != nil {
			return fmt.Errorf("failed to get note content: %w", err)
		}

		// Output raw content
		fmt.Print(content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
