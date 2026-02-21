//go:build darwin

package profile

import (
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

const keychainService = "Claude Code-credentials"

func init() {
	readKeychainCreds = readKeychainCredentials
	writeKeychainCreds = writeKeychainCredentials
}

func readKeychainCredentials() ([]byte, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", keychainService, "-w")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("reading keychain: %w", err)
	}
	return bytes.TrimRight(out, "\n"), nil
}

func writeKeychainCredentials(data []byte) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}
	cmd := exec.Command("security", "add-generic-password", "-U",
		"-s", keychainService,
		"-a", u.Username,
		"-w", string(data))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("writing keychain: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
