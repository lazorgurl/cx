package profile

import (
	"path/filepath"
	"runtime"

	"github.com/lazorgurl/cx/internal/config"
)

// readKeychainCreds and writeKeychainCreds are set by keychain_darwin.go on macOS.
var readKeychainCreds func() ([]byte, error)
var writeKeychainCreds func([]byte) error

// ReadLiveCredentials reads the live Claude Code credentials.
// On macOS, credentials are read from the system Keychain.
// On other platforms, they are read from the file at CredentialsPath().
func ReadLiveCredentials() ([]byte, error) {
	if runtime.GOOS == "darwin" && readKeychainCreds != nil {
		return readKeychainCreds()
	}
	credPath, err := config.CredentialsPath()
	if err != nil {
		return nil, err
	}
	return ReadCredentials(credPath)
}

// WriteLiveCredentials writes the live Claude Code credentials.
// On macOS, credentials are written to the system Keychain.
// On other platforms, they are written to the file at CredentialsPath().
func WriteLiveCredentials(data []byte) error {
	if runtime.GOOS == "darwin" && writeKeychainCreds != nil {
		return writeKeychainCreds(data)
	}
	credPath, err := config.CredentialsPath()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(filepath.Dir(credPath)); err != nil {
		return err
	}
	return WriteCredentials(credPath, data)
}
