package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lazorgurl/cx/internal/config"
)

const (
	tokenEndpoint = "https://platform.claude.com/v1/oauth/token"
	oauthClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	oauthScope    = "user:profile user:inference user:sessions:claude_code user:mcp_servers"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// RefreshCredentials exchanges the refresh token in raw for a new access token
// and returns updated credentials JSON. The refresh token is replaced if the
// server returns a new one (token rotation).
func RefreshCredentials(raw []byte) ([]byte, error) {
	// Unmarshal as a map to preserve unknown fields (MCP entries, etc.).
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	oauthRaw, ok := top["claudeAiOauth"]
	if !ok {
		return nil, fmt.Errorf("credentials missing claudeAiOauth section")
	}

	var oauth map[string]json.RawMessage
	if err := json.Unmarshal(oauthRaw, &oauth); err != nil {
		return nil, fmt.Errorf("parsing claudeAiOauth: %w", err)
	}

	var refreshToken string
	if rt, ok := oauth["refreshToken"]; ok {
		if err := json.Unmarshal(rt, &refreshToken); err != nil {
			return nil, fmt.Errorf("parsing refreshToken: %w", err)
		}
	}
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token found in credentials")
	}

	reqBody, err := json.Marshal(tokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     oauthClientID,
		Scope:        oauthScope,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding token request: %w", err)
	}

	resp, err := httpClient.Post(tokenEndpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("calling token endpoint: %w", err)
	}
	defer resp.Body.Close()

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	if tok.Error != "" {
		if tok.ErrorDesc != "" {
			return nil, fmt.Errorf("token refresh failed: %s (%s)", tok.Error, tok.ErrorDesc)
		}
		return nil, fmt.Errorf("token refresh failed: %s", tok.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned HTTP %d", resp.StatusCode)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned empty access_token")
	}

	// Update fields in the oauth map.
	if v, err := json.Marshal(tok.AccessToken); err == nil {
		oauth["accessToken"] = v
	}
	if tok.RefreshToken != "" {
		if v, err := json.Marshal(tok.RefreshToken); err == nil {
			oauth["refreshToken"] = v
		}
	}
	expiresAt := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second).UnixMilli()
	if v, err := json.Marshal(expiresAt); err == nil {
		oauth["expiresAt"] = v
	}

	newOAuth, err := json.Marshal(oauth)
	if err != nil {
		return nil, fmt.Errorf("re-encoding oauth section: %w", err)
	}
	top["claudeAiOauth"] = newOAuth

	result, err := json.Marshal(top)
	if err != nil {
		return nil, fmt.Errorf("re-encoding credentials: %w", err)
	}
	return result, nil
}

// RefreshProfile refreshes the credentials stored for the named profile.
// If the profile is currently active, the live credentials file is also updated.
func RefreshProfile(name string) error {
	profCredPath, err := config.ProfileCredPath(name)
	if err != nil {
		return err
	}
	raw, err := ReadCredentials(profCredPath)
	if err != nil {
		return fmt.Errorf("reading profile %q: %w", name, err)
	}
	refreshed, err := RefreshCredentials(raw)
	if err != nil {
		return err
	}
	if err := WriteCredentials(profCredPath, refreshed); err != nil {
		return fmt.Errorf("writing refreshed credentials for %q: %w", name, err)
	}

	// Also update live credentials if this profile is currently active.
	state, err := ReadState()
	if err == nil && state.ActiveProfile == name {
		_ = WriteLiveCredentials(refreshed)
	}
	return nil
}
