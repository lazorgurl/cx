package cmd

import (
	"fmt"

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
	if err := profile.UseProfile(name); err != nil {
		return err
	}
	fmt.Printf("Switched to profile %q.\n", name)
	return nil
}
