package config

import (
	"os"
	"path/filepath"
)

const (
	appName          = "gmail-cli"
	credentialsFile  = "credentials.json"
	tokenFile        = "token.json"
)

// ConfigDir returns the path to the config directory.
// Uses XDG_CONFIG_HOME if set, otherwise ~/.config/gmail-cli/
func ConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, appName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", appName)
	}
	return filepath.Join(home, ".config", appName)
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0700)
}

// CredentialsPath returns the path to the OAuth credentials file.
func CredentialsPath() string {
	return filepath.Join(ConfigDir(), credentialsFile)
}

// TokenPath returns the path to the stored OAuth token.
func TokenPath() string {
	return filepath.Join(ConfigDir(), tokenFile)
}
