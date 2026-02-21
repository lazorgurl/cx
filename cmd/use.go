package cmd

import (
	"fmt"
	"os"

	"github.com/lazorgurl/cx/internal/config"
	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a named credential profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	// If the target profile's token is expired, try to refresh it automatically
	// before switching. Surface a clear error if that fails (e.g. the refresh
	// token was rotated away since the profile was last saved).
	if profCredPath, err := config.ProfileCredPath(name); err == nil {
		if raw, err := profile.ReadCredentials(profCredPath); err == nil {
			if expired, err := profile.CredentialsExpired(raw); err == nil && expired {
				fmt.Fprintf(os.Stderr, "Token for %q is expired, attempting refresh...\n", name)
				if err := profile.RefreshProfile(name); err != nil {
					fmt.Fprintf(os.Stderr, "Auto-refresh failed: %v\n", err)
					fmt.Fprintf(os.Stderr, "You will need to log in when Claude Code starts.\n")
					fmt.Fprintf(os.Stderr, "Afterwards, run: cx save %s\n", name)
				} else {
					fmt.Fprintf(os.Stderr, "Token refreshed.\n")
				}
			}
		}
	}

	if err := profile.UseProfile(name); err != nil {
		return err
	}
	fmt.Printf("Switched to profile %q.\n", name)
	return nil
}
