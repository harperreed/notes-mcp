// ABOUTME: Folders command for listing folders in Apple Notes
// ABOUTME: Returns a newline-separated list of folder names

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "List all folders in Apple Notes",
	Long:  `Lists all folders in Apple Notes. Returns a newline-separated list of folder names.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// List folders
		folders, err := notesService.ListFolders(ctx)
		if err != nil {
			return fmt.Errorf("failed to list folders: %w", err)
		}

		// Output newline-separated list of folder names
		for _, folder := range folders {
			fmt.Println(folder)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(foldersCmd)
}
