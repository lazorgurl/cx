package config

import (
	"os"
	"path/filepath"
)

// ClaudeConfigDir returns the Claude configuration directory.
// It uses the CLAUDE_CONFIG_DIR environment variable if set,
// otherwise defaults to ~/.claude.
func ClaudeConfigDir() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// CredentialsPath returns the path to the live Claude credentials file.
// On macOS, Claude Code stores live credentials in the system Keychain instead
// of this file. This path is used as a fallback on non-macOS platforms.
func CredentialsPath() (string, error) {
	dir, err := ClaudeConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".credentials.json"), nil
}

// CxConfigDir returns the cx configuration directory (~/.config/cx).
func CxConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "cx"), nil
}

// ProfilesDir returns the directory containing all saved profiles.
func ProfilesDir() (string, error) {
	dir, err := CxConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles"), nil
}

// ProfileDir returns the directory for a specific named profile.
func ProfileDir(name string) (string, error) {
	dir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// ProfileCredPath returns the credentials file path for a specific named profile.
func ProfileCredPath(name string) (string, error) {
	dir, err := ProfileDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// StatePath returns the path to the cx state file.
func StatePath() (string, error) {
	dir, err := CxConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

// EnsureDir creates the directory at path (and any parents) with 0700 permissions.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}
