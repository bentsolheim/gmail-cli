package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
	"github.com/spf13/cobra"
)

var (
	draftTo      []string
	draftCc      []string
	draftBcc     []string
	draftSubject string
)

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Manage Gmail drafts",
	Long:  `Create and manage Gmail drafts.`,
}

var draftCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Gmail draft from markdown on stdin",
	Long: `Create a new Gmail draft from markdown input on stdin.

The markdown body is read from stdin and converted to both plain text
and HTML for multipart email display. The draft is saved but NOT sent.

Heading convention:
  #    → extracted as email subject (removed from body)
  ##   → primary heading
  ###  → secondary heading
  ####+ → same style as ###

If --subject is provided, it overrides any # heading in the markdown.
If neither --subject nor a # heading is present, an error is returned.

Examples:
  gmail-cli draft create --to a@x.com < email.md
  echo "# Subject\n\nBody" | gmail-cli draft create --to user@example.com
  gmail-cli draft create --to a@x.com --subject "Override" < email.md`,
	RunE: runDraftCreate,
}

func init() {
	draftCreateCmd.Flags().StringSliceVar(&draftTo, "to", nil, "Recipient email address (can be specified multiple times)")
	draftCreateCmd.Flags().StringSliceVar(&draftCc, "cc", nil, "CC recipient (can be specified multiple times)")
	draftCreateCmd.Flags().StringSliceVar(&draftBcc, "bcc", nil, "BCC recipient (can be specified multiple times)")
	draftCreateCmd.Flags().StringVar(&draftSubject, "subject", "", "Email subject line")
	_ = draftCreateCmd.MarkFlagRequired("to")

	draftCmd.AddCommand(draftCreateCmd)
	rootCmd.AddCommand(draftCmd)
}

func runDraftCreate(cmd *cobra.Command, args []string) error {
	// Detect if stdin has piped content
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("failed to check stdin: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("no input on stdin; pipe markdown content, e.g.:\n  echo '**Hello**' | gmail-cli draft create --to user@example.com --subject \"Test\"")
	}

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	markdownBody := strings.TrimSpace(string(body))
	if markdownBody == "" {
		return fmt.Errorf("stdin was empty; provide markdown content for the email body")
	}

	// Extract subject from # heading if --subject not provided
	subject := draftSubject
	if subject == "" {
		extracted, remaining := gmail.ExtractSubject(markdownBody)
		if extracted == "" {
			return fmt.Errorf("no --subject flag and no '# Title' heading found in markdown")
		}
		subject = extracted
		markdownBody = remaining
	} else {
		// --subject provided: still strip the # heading from the body if present
		_, remaining := gmail.ExtractSubject(markdownBody)
		markdownBody = remaining
	}

	ctx := context.Background()
	client, err := gmail.NewClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.CreateDraft(ctx, draftTo, draftCc, draftBcc, subject, markdownBody)
	if err != nil {
		return err
	}

	fmt.Println("Draft created successfully.")
	fmt.Printf("Draft ID: %s\n", result.DraftID)
	fmt.Printf("View in Gmail: https://mail.google.com/mail/u/0/#drafts\n")

	return nil
}
