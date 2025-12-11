package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
	"github.com/bentsolheim/gmail-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	downloadOutputDir  string
	downloadNoAttach   bool
)

var downloadCmd = &cobra.Command{
	Use:   "download <thread-id>",
	Short: "Download a Gmail thread",
	Long: `Download a complete Gmail thread by its ID.

Thread content is written to stdout. Attachments are saved to --output-dir.

Examples:
  gmail-cli download 18c1234abcd5678 --output-dir ./emails
  gmail-cli download 18c1234abcd5678 --no-attachments`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutputDir, "output-dir", "o", "", "Directory to save attachments")
	downloadCmd.Flags().BoolVar(&downloadNoAttach, "no-attachments", false, "Skip downloading attachments")
	rootCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	threadID := args[0]
	ctx := context.Background()

	client, err := gmail.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Gmail client: %w", err)
	}

	thread, err := client.GetThread(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	savedAttachments := make(map[string]string)

	// Check if thread has attachments
	hasAttachments := false
	for _, msg := range thread.Messages {
		if len(msg.Attachments) > 0 {
			hasAttachments = true
			break
		}
	}

	// Validate output-dir is provided if there are attachments
	if hasAttachments && !downloadNoAttach && downloadOutputDir == "" {
		return fmt.Errorf("thread has attachments; specify --output-dir or use --no-attachments")
	}

	// Download attachments if not disabled
	if !downloadNoAttach && downloadOutputDir != "" {
		for _, msg := range thread.Messages {
			for _, att := range msg.Attachments {
				data, err := client.DownloadAttachment(ctx, msg.ID, att.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to download %s: %v\n", att.Filename, err)
					continue
				}

				savedPath, err := gmail.SaveAttachment(data, downloadOutputDir, att.Filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save %s: %v\n", att.Filename, err)
					continue
				}

				savedAttachments[att.ID] = savedPath
			}
		}
	}

	formatter := output.NewTextFormatter()
	fmt.Print(formatter.FormatThread(thread, savedAttachments))

	return nil
}
