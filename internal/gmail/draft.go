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

// CreateDraft creates a Gmail draft from markdown content.
func (c *Client) CreateDraft(ctx context.Context, to, cc, bcc []string, subject, markdownBody string) (*DraftResult, error) {
	htmlBody, err := renderMarkdown(markdownBody)
	if err != nil {
		return nil, err
	}

	rawMsg, err := buildMIMEMessage(to, cc, bcc, subject, markdownBody, htmlBody)
	if err != nil {
		return nil, err
	}

	draft := &gmail.Draft{
		Message: &gmail.Message{
			Raw: base64.URLEncoding.EncodeToString(rawMsg),
		},
	}

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
func buildMIMEMessage(to, cc, bcc []string, subject, plainText, htmlBody string) ([]byte, error) {
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
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=%q\r\n", mpw.Boundary())
	fmt.Fprintf(&msg, "\r\n")
	msg.Write(body.Bytes())

	return msg.Bytes(), nil
}
