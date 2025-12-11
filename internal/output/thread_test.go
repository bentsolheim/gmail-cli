package output

import (
	"strings"
	"testing"
	"time"

	"github.com/bentsolheim/gmail-cli/internal/gmail"
)

func TestFormatThread_Reverse(t *testing.T) {
	thread := &gmail.Thread{
		ID:           "thread1",
		Subject:      "Test Subject",
		Participants: []string{"alice@example.com", "bob@example.com"},
		DateRange: gmail.DateRange{
			Start: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC),
		},
		Messages: []gmail.Message{
			{
				ID:   "msg1",
				From: "alice@example.com",
				Date: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				Body: "First message",
			},
			{
				ID:   "msg2",
				From: "bob@example.com",
				Date: time.Date(2025, 12, 2, 10, 0, 0, 0, time.UTC),
				Body: "Second message",
			},
			{
				ID:   "msg3",
				From: "alice@example.com",
				Date: time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC),
				Body: "Third message",
			},
		},
	}

	formatter := NewTextFormatter()

	t.Run("normal order", func(t *testing.T) {
		result := formatter.FormatThread(thread, nil, FormatOptions{})

		// Messages should appear in chronological order
		idx1 := strings.Index(result, "First message")
		idx2 := strings.Index(result, "Second message")
		idx3 := strings.Index(result, "Third message")

		if idx1 > idx2 || idx2 > idx3 {
			t.Errorf("Messages should be in chronological order: First(%d) < Second(%d) < Third(%d)", idx1, idx2, idx3)
		}
	})

	t.Run("reverse order", func(t *testing.T) {
		result := formatter.FormatThread(thread, nil, FormatOptions{Reverse: true})

		// Messages should appear in reverse order
		idx1 := strings.Index(result, "First message")
		idx2 := strings.Index(result, "Second message")
		idx3 := strings.Index(result, "Third message")

		if idx3 > idx2 || idx2 > idx1 {
			t.Errorf("Messages should be in reverse order: Third(%d) < Second(%d) < First(%d)", idx3, idx2, idx1)
		}

		// But message numbers should still be chronological
		if !strings.Contains(result, "Message 1") {
			t.Error("Should still have Message 1 (oldest)")
		}
		if !strings.Contains(result, "Message 3") {
			t.Error("Should still have Message 3 (newest)")
		}
	})
}

func TestFormatThread_MessagesOnly(t *testing.T) {
	thread := &gmail.Thread{
		ID:           "thread1",
		Subject:      "Test Subject",
		Participants: []string{"alice@example.com"},
		DateRange: gmail.DateRange{
			Start: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
		},
		Messages: []gmail.Message{
			{
				ID:   "msg1",
				From: "alice@example.com",
				Date: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				Body: `Thanks!

On Mon, Jan 1, 2025 at 10:00 AM Someone <someone@example.com> wrote:
> Original message here`,
			},
		},
	}

	formatter := NewTextFormatter()

	t.Run("without messages-only", func(t *testing.T) {
		result := formatter.FormatThread(thread, nil, FormatOptions{})

		if !strings.Contains(result, "Original message here") {
			t.Error("Without messages-only, quoted content should be present")
		}
	})

	t.Run("with messages-only", func(t *testing.T) {
		result := formatter.FormatThread(thread, nil, FormatOptions{MessagesOnly: true})

		if strings.Contains(result, "Original message here") {
			t.Error("With messages-only, quoted content should be stripped")
		}
		if !strings.Contains(result, "Thanks!") {
			t.Error("With messages-only, new content should be preserved")
		}
	})
}

func TestFormatThread_CombinedOptions(t *testing.T) {
	thread := &gmail.Thread{
		ID:           "thread1",
		Subject:      "Test Subject",
		Participants: []string{"alice@example.com", "bob@example.com"},
		DateRange: gmail.DateRange{
			Start: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 12, 2, 10, 0, 0, 0, time.UTC),
		},
		Messages: []gmail.Message{
			{
				ID:   "msg1",
				From: "alice@example.com",
				Date: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
				Body: "Original question",
			},
			{
				ID:   "msg2",
				From: "bob@example.com",
				Date: time.Date(2025, 12, 2, 10, 0, 0, 0, time.UTC),
				Body: `Here is the answer!

On Dec 1, 2025, alice@example.com wrote:
> Original question`,
			},
		},
	}

	formatter := NewTextFormatter()
	result := formatter.FormatThread(thread, nil, FormatOptions{
		Reverse:      true,
		MessagesOnly: true,
	})

	// Check reverse order
	idxOriginal := strings.Index(result, "Original question")
	idxAnswer := strings.Index(result, "Here is the answer!")

	if idxAnswer > idxOriginal {
		t.Error("With reverse, answer (newest) should appear before original (oldest)")
	}

	// Check messages-only strips quotes from second message
	// The quoted "Original question" from msg2 should not appear twice
	count := strings.Count(result, "Original question")
	if count != 1 {
		t.Errorf("'Original question' should appear exactly once (not in quote), but found %d times", count)
	}
}
