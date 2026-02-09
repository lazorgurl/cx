package cmd

import (
	"fmt"

	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	profiles, err := profile.ListProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		fmt.Println("No profiles saved yet.")
		return nil
	}
	active, _ := profile.DetectActiveProfile()
	for _, name := range profiles {
		if name == active {
			fmt.Printf("* %s (active)\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}
