package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/profile"
)

var globalCmd = &cobra.Command{
	Use:   "global <profile>",
	Short: "Set the user-level default profile by writing ~/.gcloudenv/global",
	Long: `Set the user-level default profile.

Writes ~/.gcloudenv/global so gcloudenv falls back to the named profile when no
flag, no .gcloudenv file, and no per-shell selection is in effect. This is
gcloudenv's own default and does not change gcloud's active configuration; use
"gcloudenv use --global" for that.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ok, err := runner.Exists(name)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no such configuration %q (create it with: gcloudenv create %s)", name, name)
		}

		path, err := profile.WriteGlobal(name)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s -> %s\n", path, name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(globalCmd)
}
