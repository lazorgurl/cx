package profile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadCredentials reads raw bytes from the file at path.
func ReadCredentials(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteCredentials writes data to path atomically by first writing to a
// temporary file in the same directory and then renaming it. The file
// permissions are set to 0600.
func WriteCredentials(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".cred-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting file permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// HashCredentials computes a canonical SHA-256 hex digest of the JSON data.
// It unmarshals and re-marshals the JSON to produce sorted keys for
// canonicalization before hashing.
func HashCredentials(raw []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("unmarshalling credentials for hashing: %w", err)
	}
	canonical, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("marshalling credentials for hashing: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return fmt.Sprintf("%x", sum), nil
}
