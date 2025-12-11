package gmail

import (
	"context"

	"github.com/bentsolheim/gmail-cli/internal/auth"
	"google.golang.org/api/gmail/v1"
)

// Client wraps the Gmail API service.
type Client struct {
	service *gmail.Service
	userID  string
}

// NewClient creates a new authenticated Gmail client.
func NewClient(ctx context.Context) (*Client, error) {
	service, err := auth.GetGmailService(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		service: service,
		userID:  "me",
	}, nil
}

// Service returns the underlying Gmail service for direct API access if needed.
func (c *Client) Service() *gmail.Service {
	return c.service
}
