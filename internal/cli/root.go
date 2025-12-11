package cli

import (
	"github.com/bentsolheim/gmail-cli/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gmail-cli",
	Short:   "A read-only Gmail CLI tool",
	Long:    `gmail-cli is a read-only Gmail CLI tool designed for agent/LLM consumption. It enables searching Gmail, listing results, and downloading complete email threads with attachments.`,
	Version: version.String(),
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
}
