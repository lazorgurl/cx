package cmd

import (
	"fmt"

	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a saved profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := profile.RemoveProfile(name); err != nil {
		return err
	}
	fmt.Printf("Profile %q removed.\n", name)
	return nil
}
