package output

import (
	"github.com/bentsolheim/gmail-cli/internal/gmail"
)

// Formatter defines the interface for output formatting.
type Formatter interface {
	FormatSearchResults(results []gmail.ThreadSummary) string
	FormatThread(thread *gmail.Thread, savedAttachments map[string]string) string
}

// TextFormatter implements Formatter for plain text output.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}
