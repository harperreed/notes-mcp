// ABOUTME: Folder hierarchy command for displaying folder tree in Apple Notes
// ABOUTME: Returns nested folder structure with note counts

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var folderHierarchyCmd = &cobra.Command{
	Use:   "folder-hierarchy",
	Short: "Display the folder hierarchy in Apple Notes",
	Long:  `Displays the complete folder hierarchy in Apple Notes as a nested structure with note counts.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Get the folder hierarchy
		hierarchy, err := notesService.GetFolderHierarchy(ctx)
		if err != nil {
			return fmt.Errorf("failed to get folder hierarchy: %w", err)
		}

		// Output as JSON
		output, err := json.MarshalIndent(hierarchy, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format hierarchy: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(folderHierarchyCmd)
}
