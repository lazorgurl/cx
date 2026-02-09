package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/lazorgurl/cx/internal/config"
	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active profile details and credential info",
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	configDir, err := config.ClaudeConfigDir()
	if err != nil {
		return err
	}
	credPath, err := config.CredentialsPath()
	if err != nil {
		return err
	}
	raw, err := profile.ReadCredentials(credPath)
	if err != nil {
		return fmt.Errorf("no credentials found: %w", err)
	}
	activeName, err := profile.DetectActiveProfile()
	if err != nil {
		return err
	}
	info, err := profile.ParseCredentialInfo(raw)
	if err != nil {
		return err
	}

	profileDisplay := "(none)"
	if activeName != "" {
		profileDisplay = activeName
	}

	expiresDisplay := "unknown"
	if !info.ExpiresAt.IsZero() {
		expiresDisplay = info.ExpiresAt.UTC().Format("2006-01-02 15:04:05 UTC")
	}

	scopesDisplay := "(none)"
	if len(info.Scopes) > 0 {
		scopesDisplay = strings.Join(info.Scopes, ", ")
	}

	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Profile:", profileDisplay)
	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Subscription:", info.SubscriptionType)
	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Rate limit:", info.RateLimitTier)
	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Expires:", expiresDisplay)
	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Scopes:", scopesDisplay)
	fmt.Fprintf(os.Stdout, "%-14s%s\n", "Config dir:", configDir)

	return nil
}
