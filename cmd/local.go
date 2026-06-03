package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/profile"
)

var localCmd = &cobra.Command{
	Use:   "local <profile>",
	Short: "Pin a profile for this directory by writing .gcloudenv",
	Long:  "Write a .gcloudenv file so this directory (and its children) auto-switch to the named profile, like .nvmrc.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ok, err := runner.Exists(name)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no such configuration %q (create it with: gcloudenv create %s)", name, name)
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path, err := profile.WriteLocal(cwd, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s -> %s\n", path, name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}
