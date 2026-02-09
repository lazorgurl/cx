package cmd

import (
	"fmt"

	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current credentials as a named profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runSave,
}

func runSave(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := profile.SaveProfile(name); err != nil {
		return err
	}
	fmt.Printf("Profile %q saved.\n", name)
	return nil
}
