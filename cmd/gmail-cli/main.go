package main

import (
	"os"

	"github.com/bentsolheim/gmail-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
