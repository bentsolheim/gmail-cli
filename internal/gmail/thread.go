package gmail

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

// GetThread retrieves the full content of a thread.
func (c *Client) GetThread(ctx context.Context, threadID string) (*Thread, error) {
	gmailThread, err := c.service.Users.Threads.Get(c.userID, threadID).
		Format("full").
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	thread := &Thread{
		ID:       threadID,
		Messages: make([]Message, 0, len(gmailThread.Messages)),
	}

	participantSet := make(map[string]struct{})
	var earliestDate, latestDate time.Time

	for _, gmailMsg := range gmailThread.Messages {
		msg := Message{
			ID: gmailMsg.Id,
		}

		// Extract headers
		for _, header := range gmailMsg.Payload.Headers {
			switch header.Name {
			case "Subject":
				if thread.Subject == "" {
					thread.Subject = header.Value
				}
			case "From":
				msg.From = header.Value
				email := extractEmail(header.Value)
				participantSet[email] = struct{}{}
			case "To", "Cc":
				for _, addr := range parseAddressList(header.Value) {
					participantSet[addr] = struct{}{}
				}
			case "Date":
				if t, err := parseDate(header.Value); err == nil {
					msg.Date = t
					if earliestDate.IsZero() || t.Before(earliestDate) {
						earliestDate = t
					}
					if t.After(latestDate) {
						latestDate = t
					}
				}
			}
		}

		// Extract body
		msg.Body = extractBody(gmailMsg.Payload)

		// Extract attachments
		msg.Attachments = extractAttachments(gmailMsg.Id, gmailMsg.Payload)

		thread.Messages = append(thread.Messages, msg)
	}

	// Convert participant set to slice
	for p := range participantSet {
		thread.Participants = append(thread.Participants, p)
	}

	thread.DateRange = DateRange{
		Start: earliestDate,
		End:   latestDate,
	}

	return thread, nil
}

// extractBody recursively extracts the text body from a message part.
// Prefers text/plain over text/html.
func extractBody(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}

	// If this part is text/plain, decode and return it
	if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}

	// Check nested parts for text/plain first
	for _, p := range part.Parts {
		if p.MimeType == "text/plain" {
			if body := extractBody(p); body != "" {
				return body
			}
		}
	}

	// If no text/plain, try text/html as fallback
	if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return stripHTML(string(decoded))
		}
	}

	// Check nested parts for text/html
	for _, p := range part.Parts {
		if body := extractBody(p); body != "" {
			return body
		}
	}

	return ""
}

// stripHTML removes HTML tags and decodes common entities for basic readability.
func stripHTML(html string) string {
	// Very basic HTML stripping - just remove tags
	var result strings.Builder
	inTag := false

	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}

	// Decode common HTML entities
	text := result.String()
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return text
}

// extractAttachments recursively finds all attachments in a message part.
func extractAttachments(messageID string, part *gmail.MessagePart) []Attachment {
	var attachments []Attachment

	if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
		attachments = append(attachments, Attachment{
			ID:        part.Body.AttachmentId,
			MessageID: messageID,
			Filename:  part.Filename,
			MimeType:  part.MimeType,
			Size:      part.Body.Size,
		})
	}

	for _, p := range part.Parts {
		attachments = append(attachments, extractAttachments(messageID, p)...)
	}

	return attachments
}

// extractEmail extracts the email address from a "Name <email>" string.
func extractEmail(from string) string {
	if start := strings.Index(from, "<"); start >= 0 {
		if end := strings.Index(from[start:], ">"); end > 0 {
			return from[start+1 : start+end]
		}
	}
	return strings.TrimSpace(from)
}

// parseAddressList parses a comma-separated list of email addresses.
func parseAddressList(list string) []string {
	var addresses []string
	for _, addr := range strings.Split(list, ",") {
		if email := extractEmail(strings.TrimSpace(addr)); email != "" {
			addresses = append(addresses, email)
		}
	}
	return addresses
}
