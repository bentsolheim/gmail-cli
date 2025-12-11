package cli

import (
	"context"
	"fmt"

	"github.com/bentsolheim/gmail-cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Gmail",
	Long:  `Triggers the OAuth authentication flow with Gmail. Opens a browser for authorization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fmt.Println("Starting authentication...")
		service, err := auth.ForceReauth(ctx)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Get user's email to confirm authentication
		profile, err := service.Users.GetProfile("me").Do()
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}

		fmt.Printf("Successfully authenticated as: %s\n", profile.EmailAddress)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
