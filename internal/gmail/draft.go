package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

// DraftResult holds the result of creating a draft.
type DraftResult struct {
	DraftID   string
	MessageID string
}

// ReplyContext holds threading information for creating reply drafts.
type ReplyContext struct {
	ThreadID   string
	Subject    string // original subject
	InReplyTo  string // Message-ID of the message being replied to
	References string // space-separated chain of Message-IDs

	// Quoted content from the original message
	QuotedFrom    string
	QuotedDate    string
	QuotedPlain   string // plain text body of original
	QuotedHTML    string // HTML body of original
}

// GetReplyContext fetches a thread and builds the context needed for a reply draft.
func (c *Client) GetReplyContext(ctx context.Context, threadID string) (*ReplyContext, error) {
	thread, err := c.GetThread(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread %s: %w", threadID, err)
	}

	if len(thread.Messages) == 0 {
		return nil, fmt.Errorf("thread %s has no messages", threadID)
	}

	lastMsg := thread.Messages[len(thread.Messages)-1]

	// Build References chain from all messages with Message-IDs
	var refIDs []string
	for _, msg := range thread.Messages {
		if msg.MessageID != "" {
			refIDs = append(refIDs, msg.MessageID)
		}
	}

	rc := &ReplyContext{
		ThreadID:    threadID,
		Subject:     thread.Subject,
		InReplyTo:   lastMsg.MessageID,
		References:  strings.Join(refIDs, " "),
		QuotedFrom:  lastMsg.From,
		QuotedDate:  lastMsg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM"),
		QuotedPlain: lastMsg.Body,
		QuotedHTML:  lastMsg.HTMLBody,
	}

	return rc, nil
}

// CreateDraft creates a Gmail draft from markdown content.
// If reply is non-nil, the draft is created as a reply within that thread.
func (c *Client) CreateDraft(ctx context.Context, to, cc, bcc []string, subject, markdownBody string, reply *ReplyContext) (*DraftResult, error) {
	htmlBody, err := renderMarkdown(markdownBody)
	if err != nil {
		return nil, err
	}

	plainText := markdownBody

	// Append quoted original if replying with quote content
	if reply != nil && reply.QuotedPlain != "" {
		attribution := fmt.Sprintf("On %s, %s wrote:", reply.QuotedDate, reply.QuotedFrom)
		plainText += "\n\n---------- " + attribution + " ----------\n\n" + reply.QuotedPlain
		htmlBody = appendQuotedHTML(htmlBody, attribution, reply.QuotedHTML, reply.QuotedPlain)
	}

	rawMsg, err := buildMIMEMessage(to, cc, bcc, subject, plainText, htmlBody, reply)
	if err != nil {
		return nil, err
	}

	draftMsg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(rawMsg),
	}
	if reply != nil {
		draftMsg.ThreadId = reply.ThreadID
	}

	draft := &gmail.Draft{Message: draftMsg}

	created, err := c.service.Users.Drafts.Create(c.userID, draft).Context(ctx).Do()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 403 {
			return nil, fmt.Errorf("insufficient permissions for draft creation.\nRun 'gmail-cli auth' to re-authenticate with the required scopes")
		}
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	return &DraftResult{
		DraftID:   created.Id,
		MessageID: created.Message.Id,
	}, nil
}

// appendQuotedHTML appends quoted original content to a reply HTML body.
// Uses a separator line with attribution rather than a blockquote, for cleaner rendering.
func appendQuotedHTML(replyHTML, attribution, originalHTML, originalPlain string) string {
	// Strip closing </body></html> from reply
	replyHTML = strings.TrimSuffix(replyHTML, "\n</body>\n</html>")

	var quoted strings.Builder
	quoted.WriteString(replyHTML)
	quoted.WriteString("\n<br>\n<div class=\"gmail_quote\">\n")
	quoted.WriteString(fmt.Sprintf("<div style=\"margin: 1em 0 0.5em 0; padding-top: 0.5em; border-top: 1px solid #ccc; color: #555; font-size: 0.9em;\">%s</div>\n", attribution))

	if originalHTML != "" {
		quoted.WriteString(stripOuterHTMLTags(originalHTML))
	} else {
		quoted.WriteString("<pre>")
		quoted.WriteString(originalPlain)
		quoted.WriteString("</pre>")
	}

	quoted.WriteString("\n</div>\n</body>\n</html>")
	return quoted.String()
}

// stripOuterHTMLTags removes the outer html/head/body tags from an HTML document,
// leaving only the inner content suitable for embedding.
func stripOuterHTMLTags(html string) string {
	s := html
	// Remove <!DOCTYPE ...>
	if idx := strings.Index(strings.ToLower(s), "<!doctype"); idx >= 0 {
		if end := strings.Index(s[idx:], ">"); end > 0 {
			s = s[:idx] + s[idx+end+1:]
		}
	}
	// Remove <html...>
	if idx := strings.Index(strings.ToLower(s), "<html"); idx >= 0 {
		if end := strings.Index(s[idx:], ">"); end > 0 {
			s = s[:idx] + s[idx+end+1:]
		}
	}
	s = strings.ReplaceAll(s, "</html>", "")
	// Remove <head>...</head>
	if start := strings.Index(strings.ToLower(s), "<head"); start >= 0 {
		if end := strings.Index(strings.ToLower(s[start:]), "</head>"); end > 0 {
			s = s[:start] + s[start+end+7:]
		}
	}
	// Remove <body...>
	if idx := strings.Index(strings.ToLower(s), "<body"); idx >= 0 {
		if end := strings.Index(s[idx:], ">"); end > 0 {
			s = s[:idx] + s[idx+end+1:]
		}
	}
	s = strings.ReplaceAll(s, "</body>", "")
	return strings.TrimSpace(s)
}

// renderMarkdown converts markdown source to HTML using goldmark with GFM extensions.
func renderMarkdown(source string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}
	return wrapHTML(buf.String()), nil
}

// wrapHTML wraps rendered markdown HTML in an email-friendly HTML document.
func wrapHTML(body string) string {
	return `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: sans-serif; line-height: 1.5;">
<style>
h1, h2 { font-size: 1.2em; margin: 0.5em 0 0 0; }
h3, h4, h5, h6 { font-size: 1.1em; margin: 0.5em 0 0 0; }
h1 + p, h2 + p, h3 + p, h4 + p, h5 + p, h6 + p { margin-top: 0.2em; }
p + ul, p + ol { margin-top: -0.5em; }
ul, ol { padding-left: 1.5em; }
table { border-collapse: collapse; margin: 8px 0; }
th, td { border: 1pt solid #333; padding: 4px 8px; }
th { text-align: center; font-weight: bold; }
pre { background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; }
code { background: #f5f5f5; padding: 2px 4px; border-radius: 2px; }
pre code { background: none; padding: 0; }
blockquote { border-left: 3px solid #ccc; margin: 8px 0; padding-left: 12px; color: #555; }
</style>
` + body + `
</body>
</html>`
}

// ExtractSubject extracts a subject from a leading `# Title` line in markdown.
// Returns the extracted subject and the remaining body with the heading removed.
// If no `# Title` is found, returns empty subject and the original body unchanged.
func ExtractSubject(markdown string) (subject, body string) {
	lines := strings.SplitN(markdown, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "# ") && !strings.HasPrefix(firstLine, "## ") {
		subject = strings.TrimSpace(firstLine[2:])
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
		}
		return subject, body
	}
	return "", markdown
}

// buildMIMEMessage constructs an RFC 2822 multipart/alternative message.
func buildMIMEMessage(to, cc, bcc []string, subject, plainText, htmlBody string, reply *ReplyContext) ([]byte, error) {
	// Build MIME body parts
	var body bytes.Buffer
	mpw := multipart.NewWriter(&body)

	// text/plain part
	plainHeader := textproto.MIMEHeader{}
	plainHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	pw, err := mpw.CreatePart(plainHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create plain text part: %w", err)
	}
	if _, err := pw.Write([]byte(plainText)); err != nil {
		return nil, fmt.Errorf("failed to write plain text: %w", err)
	}

	// text/html part
	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
	hw, err := mpw.CreatePart(htmlHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTML part: %w", err)
	}
	if _, err := hw.Write([]byte(htmlBody)); err != nil {
		return nil, fmt.Errorf("failed to write HTML: %w", err)
	}

	if err := mpw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build complete message with headers
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(to, ", "))
	if len(cc) > 0 {
		fmt.Fprintf(&msg, "Cc: %s\r\n", strings.Join(cc, ", "))
	}
	if len(bcc) > 0 {
		fmt.Fprintf(&msg, "Bcc: %s\r\n", strings.Join(bcc, ", "))
	}
	fmt.Fprintf(&msg, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	if reply != nil && reply.InReplyTo != "" {
		fmt.Fprintf(&msg, "In-Reply-To: %s\r\n", reply.InReplyTo)
	}
	if reply != nil && reply.References != "" {
		fmt.Fprintf(&msg, "References: %s\r\n", reply.References)
	}
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=%q\r\n", mpw.Boundary())
	fmt.Fprintf(&msg, "\r\n")
	msg.Write(body.Bytes())

	return msg.Bytes(), nil
}
