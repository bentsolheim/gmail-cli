package gmail

import (
	"context"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

// SearchThreads searches for threads matching the query and returns summaries.
func (c *Client) SearchThreads(ctx context.Context, query string, maxResults int64) ([]ThreadSummary, error) {
	resp, err := c.service.Users.Threads.List(c.userID).
		Q(query).
		MaxResults(maxResults).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	summaries := make([]ThreadSummary, 0, len(resp.Threads))
	for _, t := range resp.Threads {
		summary, err := c.getThreadSummary(ctx, t.Id)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (c *Client) getThreadSummary(ctx context.Context, threadID string) (ThreadSummary, error) {
	thread, err := c.service.Users.Threads.Get(c.userID, threadID).
		Format("metadata").
		MetadataHeaders("From", "Subject", "Date").
		Context(ctx).
		Do()
	if err != nil {
		return ThreadSummary{}, err
	}

	summary := ThreadSummary{
		ID:           threadID,
		MessageCount: len(thread.Messages),
	}

	participantSet := make(map[string]struct{})
	var latestDate time.Time

	for _, msg := range thread.Messages {
		// Count attachments
		summary.AttachmentCount += countAttachments(msg.Payload)

		// Extract headers
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "Subject":
				if summary.Subject == "" {
					summary.Subject = header.Value
				}
			case "From":
				name := extractName(header.Value)
				participantSet[name] = struct{}{}
			case "Date":
				if t, err := parseDate(header.Value); err == nil {
					if t.After(latestDate) {
						latestDate = t
					}
				}
			}
		}
	}

	// Convert participant set to slice
	for p := range participantSet {
		summary.Participants = append(summary.Participants, p)
	}

	summary.LastMessageDate = latestDate

	return summary, nil
}

func countAttachments(part *gmail.MessagePart) int {
	count := 0
	if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
		count++
	}
	for _, p := range part.Parts {
		count += countAttachments(p)
	}
	return count
}

// extractName extracts a display name from an email address like "Name <email@example.com>"
func extractName(from string) string {
	// Handle "Name <email>" format
	if idx := strings.Index(from, "<"); idx > 0 {
		name := strings.TrimSpace(from[:idx])
		name = strings.Trim(name, "\"")
		if name != "" {
			return name
		}
	}

	// Handle "email@example.com" format - extract local part
	if idx := strings.Index(from, "@"); idx > 0 {
		email := strings.TrimPrefix(from, "<")
		return email[:idx]
	}

	return from
}

// parseDate attempts to parse various email date formats.
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
		"2 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}
