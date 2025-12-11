package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
	"github.com/bentsolheim/gmail-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	interactive bool
	outputDir   string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Gmail threads",
	Long: `Search for Gmail threads matching the query.

Uses Gmail's search syntax. Examples:
  gmail-cli search "from:felipe subject:conversion"
  gmail-cli search "after:2025/12/01 has:attachment"
  gmail-cli search "is:unread from:me"

With --interactive, prompts to select and download a thread after search.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode: select and download a thread after search")
	searchCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for attachments (used with --interactive)")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	ctx := context.Background()

	client, err := gmail.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Gmail client: %w", err)
	}

	results, err := client.SearchThreads(ctx, query, 25)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	formatter := output.NewTextFormatter()
	fmt.Print(formatter.FormatSearchResults(results))

	if !interactive || len(results) == 0 {
		return nil
	}

	// Interactive mode: prompt for selection
	selection, err := promptSelection(len(results))
	if err != nil {
		return err
	}

	if selection == 0 {
		// User quit
		return nil
	}

	selectedThread := results[selection-1]

	// Download the selected thread
	return downloadThread(ctx, client, selectedThread.ID, outputDir, false)
}

func promptSelection(max int) (int, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\nEnter number (1-%d) to download, or 'q' to quit: ", max)
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" {
			return 0, nil
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > max {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", max)
			continue
		}

		return num, nil
	}
}

func downloadThread(ctx context.Context, client *gmail.Client, threadID, outDir string, noAttachments bool) error {
	thread, err := client.GetThread(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	savedAttachments := make(map[string]string)

	// Download attachments if not disabled
	if !noAttachments {
		for _, msg := range thread.Messages {
			for _, att := range msg.Attachments {
				if outDir == "" {
					// Prompt for output dir if there are attachments and none specified
					fmt.Printf("Thread has attachments. Specify output directory (or press Enter to skip): ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					outDir = strings.TrimSpace(input)
					if outDir == "" {
						noAttachments = true
						break
					}
				}

				if noAttachments {
					break
				}

				data, err := client.DownloadAttachment(ctx, msg.ID, att.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to download %s: %v\n", att.Filename, err)
					continue
				}

				savedPath, err := gmail.SaveAttachment(data, outDir, att.Filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save %s: %v\n", att.Filename, err)
					continue
				}

				savedAttachments[att.ID] = savedPath
			}
			if noAttachments {
				break
			}
		}
	}

	formatter := output.NewTextFormatter()
	fmt.Print(formatter.FormatThread(thread, savedAttachments))

	return nil
}
