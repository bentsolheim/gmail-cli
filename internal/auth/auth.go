package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/bentsolheim/gmail-cli/internal/config"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GetGmailService returns an authenticated Gmail service.
// If a valid token exists, it uses that. Otherwise, it initiates the OAuth flow.
func GetGmailService(ctx context.Context) (*gmail.Service, error) {
	oauthConfig, err := loadOAuthConfig()
	if err != nil {
		return nil, err
	}

	token, err := getToken(ctx, oauthConfig)
	if err != nil {
		return nil, err
	}

	client := oauthConfig.Client(ctx, token)
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

// ForceReauth performs a fresh OAuth flow, ignoring any existing token.
func ForceReauth(ctx context.Context) (*gmail.Service, error) {
	oauthConfig, err := loadOAuthConfig()
	if err != nil {
		return nil, err
	}

	token, err := performOAuthFlow(ctx, oauthConfig)
	if err != nil {
		return nil, err
	}

	if err := SaveToken(config.TokenPath(), token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	client := oauthConfig.Client(ctx, token)
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

func loadOAuthConfig() (*oauth2.Config, error) {
	credPath := config.CredentialsPath()
	b, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file at %s: %w\n\nTo set up authentication:\n1. Go to https://console.cloud.google.com/\n2. Create a project and enable the Gmail API\n3. Create OAuth 2.0 credentials (Desktop app)\n4. Download the credentials and save as %s", credPath, err, credPath)
	}

	cfg, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	return cfg, nil
}

func getToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	tokenPath := config.TokenPath()
	token, err := LoadToken(tokenPath)
	if err == nil && token.Valid() {
		return token, nil
	}

	// Token doesn't exist or is invalid, need to authenticate
	token, err = performOAuthFlow(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := config.EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := SaveToken(tokenPath, token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

func performOAuthFlow(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	// Start callback server
	server, err := NewCallbackServer()
	if err != nil {
		return nil, err
	}
	defer server.Close()

	// Update config with actual redirect URL
	cfg.RedirectURL = server.RedirectURL()

	// Start server
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	server.Start(cancelCtx)

	// Generate auth URL and open browser
	authURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	if err := browser.OpenURL(authURL); err != nil {
		// Browser failed to open, but we printed the URL so user can copy it
		fmt.Println("Could not open browser automatically. Please copy the URL above.")
	}

	// Wait for callback
	fmt.Println("Waiting for authorization...")
	code, err := server.WaitForCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	// Exchange code for token
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	return token, nil
}
