package output

import (
	"fmt"
	"strings"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
)

// FormatSearchResults formats search results for display.
// Output format:
// [1] Dec 11 | Felipe Garcia | Re: Conversion factors (3 messages, 2 attachments)
func (f *TextFormatter) FormatSearchResults(results []gmail.ThreadSummary) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	for i, r := range results {
		// Format date as "Dec 11"
		date := r.LastMessageDate.Format("Jan 2")

		// Format participants (join with ", ", truncate if too long)
		participants := strings.Join(r.Participants, ", ")
		if len(participants) > 30 {
			participants = participants[:27] + "..."
		}

		// Format message/attachment counts
		msgText := "message"
		if r.MessageCount != 1 {
			msgText = "messages"
		}

		var countParts []string
		countParts = append(countParts, fmt.Sprintf("%d %s", r.MessageCount, msgText))

		if r.AttachmentCount > 0 {
			attText := "attachment"
			if r.AttachmentCount != 1 {
				attText = "attachments"
			}
			countParts = append(countParts, fmt.Sprintf("%d %s", r.AttachmentCount, attText))
		}

		counts := strings.Join(countParts, ", ")

		// Write the line
		fmt.Fprintf(&sb, "[%d] %s | %s | %s | %s (%s)\n",
			i+1, r.ID, date, participants, r.Subject, counts)
	}

	return sb.String()
}

