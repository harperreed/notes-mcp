// ABOUTME: Create folder command for creating folders in Apple Notes
// ABOUTME: Accepts folder name and optional parent folder path

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	createFolderParent string
)

var createFolderCmd = &cobra.Command{
	Use:   "create-folder <name>",
	Short: "Create a new folder in Apple Notes",
	Long:  `Creates a new folder in Apple Notes. Optionally specify a parent folder to create a nested folder.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Create the folder
		err := notesService.CreateFolder(ctx, name, createFolderParent)
		if err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}

		// Output success message
		if createFolderParent != "" {
			fmt.Printf("Folder created: %s (under %s)\n", name, createFolderParent)
		} else {
			fmt.Printf("Folder created: %s\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createFolderCmd)

	// Add flags
	createFolderCmd.Flags().StringVar(&createFolderParent, "parent", "", "Parent folder path for nested folder creation")
}
