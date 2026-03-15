package gmail

import (
	"bufio"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "bold text",
			input: "This is **bold** text",
			contains: []string{
				"<strong>bold</strong>",
				"<!DOCTYPE html>",
				"</body>",
			},
		},
		{
			name:  "italic text",
			input: "This is *italic* text",
			contains: []string{
				"<em>italic</em>",
			},
		},
		{
			name:  "heading",
			input: "# Heading 1\n\nSome text",
			contains: []string{
				"<h1>Heading 1</h1>",
				"<p>Some text</p>",
			},
		},
		{
			name:  "bullet list",
			input: "- Item 1\n- Item 2\n- Item 3",
			contains: []string{
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"</ul>",
			},
		},
		{
			name:  "code block",
			input: "```\nfoo := bar\n```",
			contains: []string{
				"<pre><code>",
				"foo := bar",
				"</code></pre>",
			},
		},
		{
			name:  "GFM table",
			input: "| A | B |\n|---|---|\n| 1 | 2 |",
			contains: []string{
				"<table>",
				"<th>A</th>",
				"<td>1</td>",
				"</table>",
			},
		},
		{
			name:  "GFM strikethrough",
			input: "This is ~~deleted~~ text",
			contains: []string{
				"<del>deleted</del>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderMarkdown(tt.input)
			if err != nil {
				t.Fatalf("renderMarkdown() error: %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("renderMarkdown() result missing %q\nGot: %s", want, result)
				}
			}
		})
	}
}

func TestBuildMIMEMessage(t *testing.T) {
	to := []string{"alice@example.com", "bob@example.com"}
	cc := []string{"carol@example.com"}
	bcc := []string{"dave@example.com"}
	subject := "Test Subject"
	plain := "Hello **world**"
	html := "<html><body><p>Hello <strong>world</strong></p></body></html>"

	raw, err := buildMIMEMessage(to, cc, bcc, subject, plain, html)
	if err != nil {
		t.Fatalf("buildMIMEMessage() error: %v", err)
	}

	msg := string(raw)

	// Check headers
	if !strings.Contains(msg, "To: alice@example.com, bob@example.com\r\n") {
		t.Error("missing or incorrect To header")
	}
	if !strings.Contains(msg, "Cc: carol@example.com\r\n") {
		t.Error("missing or incorrect Cc header")
	}
	if !strings.Contains(msg, "Bcc: dave@example.com\r\n") {
		t.Error("missing or incorrect Bcc header")
	}
	if !strings.Contains(msg, "MIME-Version: 1.0\r\n") {
		t.Error("missing MIME-Version header")
	}

	// Parse the multipart body
	_, params, err := mime.ParseMediaType(extractHeader(msg, "Content-Type"))
	if err != nil {
		t.Fatalf("failed to parse Content-Type: %v", err)
	}
	boundary := params["boundary"]
	if boundary == "" {
		t.Fatal("missing boundary in Content-Type")
	}

	// Extract body after blank line
	parts := strings.SplitN(msg, "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatal("could not split headers from body")
	}

	reader := multipart.NewReader(strings.NewReader(parts[1]), boundary)
	var foundPlain, foundHTML bool

	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		ct := part.Header.Get("Content-Type")
		bodyBytes, _ := io.ReadAll(part)
		body := string(bodyBytes)

		if strings.HasPrefix(ct, "text/plain") {
			foundPlain = true
			if !strings.Contains(body, "Hello **world**") {
				t.Errorf("plain text part missing expected content, got: %s", body)
			}
		}
		if strings.HasPrefix(ct, "text/html") {
			foundHTML = true
			if !strings.Contains(body, "<strong>world</strong>") {
				t.Errorf("HTML part missing expected content, got: %s", body)
			}
		}
	}

	if !foundPlain {
		t.Error("no text/plain part found")
	}
	if !foundHTML {
		t.Error("no text/html part found")
	}
}

func TestBuildMIMEMessage_EmptyOptionalFields(t *testing.T) {
	raw, err := buildMIMEMessage([]string{"alice@example.com"}, nil, nil, "Test", "body", "<p>body</p>")
	if err != nil {
		t.Fatalf("buildMIMEMessage() error: %v", err)
	}

	msg := string(raw)

	if strings.Contains(msg, "Cc:") {
		t.Error("Cc header should be omitted when empty")
	}
	if strings.Contains(msg, "Bcc:") {
		t.Error("Bcc header should be omitted when empty")
	}
	if !strings.Contains(msg, "To: alice@example.com\r\n") {
		t.Error("To header missing")
	}
}

func TestBuildMIMEMessage_NonASCIISubject(t *testing.T) {
	raw, err := buildMIMEMessage([]string{"a@b.com"}, nil, nil, "Ærlig talt — norsk emne", "body", "<p>body</p>")
	if err != nil {
		t.Fatalf("buildMIMEMessage() error: %v", err)
	}

	msg := string(raw)

	// Subject should be Q-encoded for non-ASCII
	subjectLine := extractHeader(msg, "Subject")
	if subjectLine == "" {
		t.Fatal("Subject header not found")
	}
	// Decode it back
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(subjectLine)
	if err != nil {
		t.Fatalf("failed to decode subject: %v", err)
	}
	if decoded != "Ærlig talt — norsk emne" {
		t.Errorf("decoded subject = %q, want %q", decoded, "Ærlig talt — norsk emne")
	}
}

// extractHeader extracts a header value from a raw RFC 2822 message string.
func extractHeader(raw string, name string) string {
	headers := textproto.NewReader(bufio.NewReader(strings.NewReader(raw)))
	h, _ := headers.ReadMIMEHeader()
	return h.Get(name)
}
