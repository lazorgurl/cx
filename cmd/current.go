package cmd

import (
	"fmt"
	"os"

	"github.com/lazorgurl/cx/internal/profile"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print the active profile name",
	Args:  cobra.NoArgs,
	RunE:  runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	name, err := profile.DetectActiveProfile()
	if err != nil {
		return err
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "No active profile.")
		os.Exit(1)
	}
	fmt.Println(name)
	return nil
}
