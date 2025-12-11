package gmail

import "time"

// ThreadSummary contains summary information about a thread for search results.
type ThreadSummary struct {
	ID              string
	Subject         string
	Participants    []string
	LastMessageDate time.Time
	MessageCount    int
	AttachmentCount int
}

// Thread contains the full content of an email thread.
type Thread struct {
	ID           string
	Subject      string
	Participants []string
	DateRange    DateRange
	Messages     []Message
}

// DateRange represents the time span of a thread.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// Message represents a single email message within a thread.
type Message struct {
	ID          string
	From        string
	Date        time.Time
	Body        string
	Attachments []Attachment
}

// Attachment represents an email attachment.
type Attachment struct {
	ID        string
	MessageID string
	Filename  string
	MimeType  string
	Size      int64
}
