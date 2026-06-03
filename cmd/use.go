package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var useGlobal bool

var useCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch the active gcloud profile",
	Long: `Switch the active gcloud profile.

Without --global this affects only the current shell (via the shim). When run
as the bare binary it can only set the global default, so it prints a hint to
source the shim for per-shell switching.`,
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

		if useGlobal {
			if err := runner.Activate(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set global default profile to %q\n", name)
			return nil
		}

		// Reaching here means the shim isn't installed (otherwise `use` is a
		// shell function that never calls this branch).
		fmt.Fprintf(cmd.OutOrStdout(),
			"Per-shell switching needs the shell integration. Either:\n"+
				"  eval \"$(gcloudenv export %s)\"   # this shell, now\n"+
				"  gcloudenv use %s --global         # change the gcloud default\n"+
				"Install the shim once with: eval \"$(gcloudenv init zsh)\"\n",
			name, name)
		return nil
	},
}

func init() {
	useCmd.Flags().BoolVarP(&useGlobal, "global", "g", false, "set the gcloud global default instead of just this shell")
	rootCmd.AddCommand(useCmd)
}
