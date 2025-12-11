package output

import (
	"github.com/bentsolheim/gmail-cli/internal/gmail"
)

// FormatOptions controls how thread output is formatted.
type FormatOptions struct {
	// Reverse displays messages in reverse order (newest first)
	Reverse bool
	// MessagesOnly strips quoted content, showing only new message text
	MessagesOnly bool
}

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatSearchResults(results []gmail.ThreadSummary) string
	FormatThread(thread *gmail.Thread, savedAttachments map[string]string, opts FormatOptions) string
}

// TextFormatter implements Formatter for plain text output.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}
