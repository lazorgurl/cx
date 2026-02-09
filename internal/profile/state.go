package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lazorgurl/cx/internal/config"
)

// State holds the cx application state, including the currently active profile.
type State struct {
	ActiveProfile string `json:"activeProfile"`
}

// ReadState reads and unmarshals the state file. If the file does not exist,
// it returns a zero-value State and no error.
func ReadState() (State, error) {
	path, err := config.StatePath()
	if err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("reading state file: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("parsing state file: %w", err)
	}
	return s, nil
}

// WriteState marshals and atomically writes the state to the state file,
// ensuring the parent directory exists.
func WriteState(s State) error {
	path, err := config.StatePath()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("ensuring state directory: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}
	data = append(data, '\n')

	// Atomic write: write to temp file, then rename.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp state file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp state file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting state file permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp state file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp state file: %w", err)
	}
	return nil
}
