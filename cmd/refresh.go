package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/lazorgurl/cx/internal/config"
	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var refreshWatch bool

var refreshCmd = &cobra.Command{
	Use:   "refresh [<name>...]",
	Short: "Refresh OAuth tokens for profiles (all profiles if none specified)",
	Long: `Refresh OAuth tokens for one or more saved profiles.

With --watch, runs continuously and refreshes each profile when its token
is within 1 hour of expiry. Use this to keep all profiles warm in the background:

  cx refresh --watch &`,
	RunE: runRefresh,
}

func init() {
	refreshCmd.Flags().BoolVarP(&refreshWatch, "watch", "w", false, "Continuously refresh tokens as they near expiry")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	if refreshWatch {
		return runRefreshWatch(args)
	}
	return refreshOnce(args)
}

func refreshOnce(names []string) error {
	if len(names) == 0 {
		var err error
		names, err = profile.ListProfiles()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No profiles found.")
			return nil
		}
	}

	var failed int
	for _, name := range names {
		if err := profile.RefreshProfile(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error refreshing %q: %v\n", name, err)
			failed++
		} else {
			fmt.Printf("Refreshed %q.\n", name)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d profile(s) failed to refresh", failed)
	}
	return nil
}

// runRefreshWatch loops indefinitely, refreshing each profile when its token
// is within 1 hour of expiry.
func runRefreshWatch(names []string) error {
	const checkInterval = 15 * time.Minute
	const refreshWindow = time.Hour

	fmt.Println("Watching profiles for token expiry (Ctrl-C to stop)...")

	for {
		targets := names
		if len(targets) == 0 {
			var err error
			targets, err = profile.ListProfiles()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: listing profiles: %v\n", err)
			}
		}

		for _, name := range targets {
			needsRefresh, err := profileNeedsRefresh(name, refreshWindow)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: checking %q: %v\n", name, err)
				continue
			}
			if !needsRefresh {
				continue
			}
			if err := profile.RefreshProfile(name); err != nil {
				fmt.Fprintf(os.Stderr, "Error refreshing %q: %v\n", name, err)
			} else {
				fmt.Printf("[%s] Refreshed %q.\n", time.Now().Format("15:04:05"), name)
			}
		}

		time.Sleep(checkInterval)
	}
}

// profileNeedsRefresh reports whether the named profile's token expires within
// the given window (or is already expired).
func profileNeedsRefresh(name string, window time.Duration) (bool, error) {
	credPath, err := config.ProfileCredPath(name)
	if err != nil {
		return false, err
	}
	raw, err := profile.ReadCredentials(credPath)
	if err != nil {
		return false, err
	}
	info, err := profile.ParseCredentialInfo(raw)
	if err != nil {
		return false, err
	}
	if info.ExpiresAt.IsZero() {
		return false, nil // no expiry info, skip
	}
	return time.Until(info.ExpiresAt) < window, nil
}
