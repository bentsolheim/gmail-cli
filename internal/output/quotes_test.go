package output

import (
	"strings"
	"testing"
)

func TestStripQuotedContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Gmail style quote",
			input: `Thanks, this looks good!

On Mon, Jan 1, 2025 at 10:00 AM Name <email@example.com> wrote:
> Here is the update you requested.
>
> On Dec 31, 2024 at 9:00 AM Someone <someone@example.com> wrote:
> > Can you send me the update?`,
			expected: "Thanks, this looks good!",
		},
		{
			name: "Apple Mail style quote",
			input: `Got it, thanks!

On 8 Dec 2025, at 14:30, other@example.com wrote:

> Here is the document.`,
			expected: "Got it, thanks!",
		},
		{
			name: "Simple inline quotes",
			input: `I agree with this point.

> Some quoted text
> More quoted text

And this is my response.`,
			expected: `I agree with this point.


And this is my response.`,
		},
		{
			name: "Norwegian date format",
			input: `Takk for svar!

8. des. 2025 kl. 10:30 skrev Ola Nordmann <ola@example.com>:
> Original melding her.`,
			expected: "Takk for svar!",
		},
		{
			name: "Norwegian with weekday abbreviation",
			input: `Yes, we can do that.

man. 1. des. 2025 kl. 12:22 skrev Magnus Young <Magnus.Young@hydro.com>:`,
			expected: "Yes, we can do that.",
		},
		{
			name: "Norwegian with weekday and multiline email",
			input: `Hi Felipe,

Yes, we can select all flows of this category and disable them.

Kind regards
Bent

ons. 12. nov. 2025 kl. 09:53 skrev Felipe Martinez Rodriguez <
felipe.martinez@hydro.com>:`,
			expected: `Hi Felipe,

Yes, we can select all flows of this category and disable them.

Kind regards
Bent`,
		},
		{
			name: "No quoted content",
			input: `This is a simple message with no quotes.

Best regards,
John`,
			expected: `This is a simple message with no quotes.

Best regards,
John`,
		},
		{
			name: "Only quoted content",
			input: `On Mon, Jan 1, 2025 at 10:00 AM Name <email@example.com> wrote:
> This is all quoted content.`,
			expected: "[No new content - forwarded/quoted message only]",
		},
		{
			name: "Forwarded message marker",
			input: `FYI

---------- Forwarded message ---------
From: Someone <someone@example.com>
To: You <you@example.com>
Subject: Important info

Original content here.`,
			expected: "FYI",
		},
		{
			name: "Generic wrote pattern",
			input: `Perfect!

John Doe wrote:
> Some text here`,
			expected: "Perfect!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripQuotedContent(tt.input)
			if result != tt.expected {
				t.Errorf("StripQuotedContent() =\n%q\nwant\n%q", result, tt.expected)
			}
		})
	}
}

func TestStripQuotedContent_PreservesInlineResponses(t *testing.T) {
	// This is a harder case - inline responses mixed with quotes
	// The current implementation will strip inline quotes, which may not be ideal
	// but is a reasonable first implementation
	input := `Thanks for your message.

> Point 1
My response to point 1.

> Point 2
My response to point 2.`

	result := StripQuotedContent(input)

	// Should at least preserve the non-quoted parts
	if !strings.Contains(result, "Thanks for your message.") {
		t.Errorf("Should preserve 'Thanks for your message.' but got: %q", result)
	}
	if !strings.Contains(result, "My response to point 1.") {
		t.Errorf("Should preserve 'My response to point 1.' but got: %q", result)
	}
}
