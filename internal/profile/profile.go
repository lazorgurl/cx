package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/lazorgurl/cx/internal/config"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// CredentialInfo holds parsed credential metadata for display purposes.
type CredentialInfo struct {
	SubscriptionType string
	RateLimitTier    string
	ExpiresAt        time.Time
	Scopes           []string
}

// ValidateName checks that the profile name matches the required pattern
// and length constraints.
func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("profile name must not be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("profile name must be at most 64 characters, got %d", len(name))
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("profile name %q is invalid: must start with a letter or digit and contain only letters, digits, hyphens, or underscores", name)
	}
	return nil
}

// SaveProfile saves the current live credentials as a named profile and
// updates the active profile state.
func SaveProfile(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	data, err := ReadLiveCredentials()
	if err != nil {
		return fmt.Errorf("reading live credentials: %w", err)
	}
	profDir, err := config.ProfileDir(name)
	if err != nil {
		return err
	}
	if err := config.EnsureDir(profDir); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}
	profCredPath, err := config.ProfileCredPath(name)
	if err != nil {
		return err
	}
	if err := WriteCredentials(profCredPath, data); err != nil {
		return fmt.Errorf("writing profile credentials: %w", err)
	}
	return WriteState(State{ActiveProfile: name})
}

// UseProfile switches the live credentials to those from the named profile
// and updates the active profile state. Before switching, it flushes the
// current live credentials back to the previously active profile so any
// token refreshes that occurred during use are preserved.
func UseProfile(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	// Flush current live credentials back to the active profile (best-effort)
	// so we never lose a silently-refreshed token when switching.
	state, err := ReadState()
	if err != nil {
		return err
	}
	if state.ActiveProfile != "" && state.ActiveProfile != name {
		if liveData, err := ReadLiveCredentials(); err == nil {
			if prevDir, err := config.ProfileDir(state.ActiveProfile); err == nil {
				if config.EnsureDir(prevDir) == nil {
					if prevCredPath, err := config.ProfileCredPath(state.ActiveProfile); err == nil {
						_ = WriteCredentials(prevCredPath, liveData)
					}
				}
			}
		}
	}

	profCredPath, err := config.ProfileCredPath(name)
	if err != nil {
		return err
	}
	data, err := ReadCredentials(profCredPath)
	if err != nil {
		return fmt.Errorf("reading profile credentials for %q: %w", name, err)
	}
	if err := WriteLiveCredentials(data); err != nil {
		return fmt.Errorf("writing live credentials: %w", err)
	}
	return WriteState(State{ActiveProfile: name})
}

// CredentialsExpired reports whether the parsed credentials have a known,
// non-zero expiry that is in the past.
func CredentialsExpired(raw []byte) (bool, error) {
	info, err := ParseCredentialInfo(raw)
	if err != nil {
		return false, err
	}
	if info.ExpiresAt.IsZero() {
		return false, nil
	}
	return time.Now().After(info.ExpiresAt), nil
}

// ListProfiles returns a sorted list of profile names. If the profiles
// directory does not exist, it returns an empty slice and no error.
func ListProfiles() ([]string, error) {
	dir, err := config.ProfilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// RemoveProfile removes the named profile directory. If it was the active
// profile, the state is cleared.
func RemoveProfile(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	profDir, err := config.ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(profDir); err != nil {
		return fmt.Errorf("removing profile directory: %w", err)
	}
	state, err := ReadState()
	if err != nil {
		return err
	}
	if state.ActiveProfile == name {
		return WriteState(State{})
	}
	return nil
}

// DetectActiveProfile determines which saved profile matches the live
// credentials. Fast path: checks the state hint first. Slow path: iterates
// all profiles comparing hashes. Returns "" with no error if no match.
func DetectActiveProfile() (string, error) {
	liveData, err := ReadLiveCredentials()
	if err != nil {
		// No live credentials available (file missing, keychain empty, etc.).
		return "", nil
	}
	liveHash, err := HashCredentials(liveData)
	if err != nil {
		return "", err
	}

	// Fast path: check state hint.
	state, err := ReadState()
	if err != nil {
		return "", err
	}
	if state.ActiveProfile != "" {
		profCredPath, err := config.ProfileCredPath(state.ActiveProfile)
		if err != nil {
			return "", err
		}
		profData, err := ReadCredentials(profCredPath)
		if err == nil {
			profHash, err := HashCredentials(profData)
			if err == nil && profHash == liveHash {
				return state.ActiveProfile, nil
			}
		}
	}

	// Slow path: iterate all profiles.
	profiles, err := ListProfiles()
	if err != nil {
		return "", err
	}
	for _, name := range profiles {
		profCredPath, err := config.ProfileCredPath(name)
		if err != nil {
			continue
		}
		profData, err := ReadCredentials(profCredPath)
		if err != nil {
			continue
		}
		profHash, err := HashCredentials(profData)
		if err != nil {
			continue
		}
		if profHash == liveHash {
			return name, nil
		}
	}
	return "", nil
}

// ParseCredentialInfo parses minimal fields from the Claude credentials JSON
// for display purposes. Missing fields result in zero values, not errors.
func ParseCredentialInfo(raw []byte) (CredentialInfo, error) {
	var cred struct {
		ClaudeAiOauth struct {
			ExpiresAt        json.Number `json:"expiresAt"`
			Scopes           []string    `json:"scopes"`
			SubscriptionType string      `json:"subscriptionType"`
			RateLimitTier    string      `json:"rateLimitTier"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(raw, &cred); err != nil {
		return CredentialInfo{}, fmt.Errorf("parsing credentials: %w", err)
	}
	info := CredentialInfo{
		SubscriptionType: cred.ClaudeAiOauth.SubscriptionType,
		RateLimitTier:    cred.ClaudeAiOauth.RateLimitTier,
		Scopes:           cred.ClaudeAiOauth.Scopes,
	}
	if s := cred.ClaudeAiOauth.ExpiresAt.String(); s != "" {
		// Try as Unix timestamp (number) first, then RFC3339 string.
		if ms, err := cred.ClaudeAiOauth.ExpiresAt.Int64(); err == nil {
			info.ExpiresAt = time.UnixMilli(ms)
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			info.ExpiresAt = t
		}
	}
	return info, nil
}
