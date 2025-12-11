package output

import (
	"fmt"
	"strings"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
)

// FormatThread formats a thread for display.
// Output format:
//
// Subject: Re: Conversion factors
// Participants: felipe@example.com, you@gmail.com
// Date Range: Dec 9-11, 2025
//
// --- Message 1 (Dec 9, 10:30 AM) ---
// From: Felipe <felipe@example.com>
// <body>
//
// Attachments:
// - conversion_factors.xlsx (saved to: /path/to/output/conversion_factors.xlsx)
func (f *TextFormatter) FormatThread(thread *gmail.Thread, savedAttachments map[string]string) string {
	var sb strings.Builder

	// Header
	fmt.Fprintf(&sb, "Subject: %s\n", thread.Subject)
	fmt.Fprintf(&sb, "Participants: %s\n", strings.Join(thread.Participants, ", "))
	fmt.Fprintf(&sb, "Date Range: %s\n", formatDateRange(thread.DateRange))
	sb.WriteString("\n")

	// Messages
	for i, msg := range thread.Messages {
		// Message header
		date := msg.Date.Format("Jan 2, 3:04 PM")
		fmt.Fprintf(&sb, "--- Message %d (%s) ---\n", i+1, date)
		fmt.Fprintf(&sb, "From: %s\n", msg.From)

		// Message body
		body := strings.TrimSpace(msg.Body)
		if body != "" {
			sb.WriteString(body)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Attachments section
	allAttachments := collectAllAttachments(thread)
	if len(allAttachments) > 0 {
		sb.WriteString("Attachments:\n")
		for _, att := range allAttachments {
			if savedPath, ok := savedAttachments[att.ID]; ok {
				fmt.Fprintf(&sb, "- %s (saved to: %s)\n", att.Filename, savedPath)
			} else {
				fmt.Fprintf(&sb, "- %s (not downloaded)\n", att.Filename)
			}
		}
	}

	return sb.String()
}

// formatDateRange formats a date range like "Dec 9-11, 2025" or "Dec 9, 2025" if same day.
func formatDateRange(dr gmail.DateRange) string {
	if dr.Start.IsZero() && dr.End.IsZero() {
		return "Unknown"
	}

	if dr.Start.IsZero() {
		return dr.End.Format("Jan 2, 2006")
	}

	if dr.End.IsZero() {
		return dr.Start.Format("Jan 2, 2006")
	}

	// Same day
	if dr.Start.Year() == dr.End.Year() &&
		dr.Start.Month() == dr.End.Month() &&
		dr.Start.Day() == dr.End.Day() {
		return dr.Start.Format("Jan 2, 2006")
	}

	// Same month and year
	if dr.Start.Year() == dr.End.Year() && dr.Start.Month() == dr.End.Month() {
		return fmt.Sprintf("%s-%d, %d",
			dr.Start.Format("Jan 2"),
			dr.End.Day(),
			dr.End.Year())
	}

	// Same year
	if dr.Start.Year() == dr.End.Year() {
		return fmt.Sprintf("%s - %s, %d",
			dr.Start.Format("Jan 2"),
			dr.End.Format("Jan 2"),
			dr.End.Year())
	}

	// Different years
	return fmt.Sprintf("%s - %s",
		dr.Start.Format("Jan 2, 2006"),
		dr.End.Format("Jan 2, 2006"))
}

// collectAllAttachments gathers all attachments from all messages in a thread.
func collectAllAttachments(thread *gmail.Thread) []gmail.Attachment {
	var attachments []gmail.Attachment
	for _, msg := range thread.Messages {
		attachments = append(attachments, msg.Attachments...)
	}
	return attachments
}
