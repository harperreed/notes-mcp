// ABOUTME: Get attachment command for retrieving attachment content
// ABOUTME: Accepts file path and outputs base64-encoded content or saves to file

package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	attachmentOutput  string
	attachmentMaxSize int64
)

var getAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <file-path>",
	Short: "Get attachment content from Apple Notes",
	Long:  `Retrieves the content of an attachment from Apple Notes. Output as base64 to stdout or save to a file using --output flag.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Create service with real executor
		notesService := newNotesService()

		// Create context with timeout
		ctx, cancel := newCommandContext()
		defer cancel()

		// Get attachment content
		content, err := notesService.GetAttachmentContent(ctx, filePath, attachmentMaxSize)
		if err != nil {
			return fmt.Errorf("failed to get attachment content: %w", err)
		}

		// If output flag is set, save to file
		if attachmentOutput != "" {
			err := os.WriteFile(attachmentOutput, content, 0600)
			if err != nil {
				return fmt.Errorf("failed to write attachment to file: %w", err)
			}
			fmt.Printf("Attachment saved to: %s\n", attachmentOutput)
		} else {
			// Output as base64 to stdout
			encoded := base64.StdEncoding.EncodeToString(content)
			fmt.Println(encoded)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getAttachmentCmd)

	// Add flags
	getAttachmentCmd.Flags().StringVarP(&attachmentOutput, "output", "o", "", "Save attachment to file instead of outputting base64")
	getAttachmentCmd.Flags().Int64Var(&attachmentMaxSize, "max-size", 10*1024*1024, "Maximum attachment size in bytes (default 10MB)")
}
