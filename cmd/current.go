package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/gcloud"
	"github.com/figverse/gcloudenv/internal/profile"
)

var currentCmd = &cobra.Command{
	Use:     "current",
	Aliases: []string{"status"},
	Short:   "Show the active profile and where it was selected",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, _ := os.Getwd()
		res := profile.Resolve("", cwd, gcloud.ActiveFromEnv())

		name, source := res.Name, string(res.Source)
		if name == "" {
			// Fall back to gcloud's own active configuration.
			configs, err := runner.List()
			if err != nil {
				return err
			}
			for _, c := range configs {
				if c.IsActive {
					name, source = c.Name, "gcloud default"
				}
			}
		}
		if name == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No active profile.")
			return nil
		}

		c, err := runner.Get(name)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Profile: %s (%s)\n", c.Name, source)
		fmt.Fprintf(out, "Account: %s\n", dash(c.Account()))
		fmt.Fprintf(out, "Project: %s\n", dash(c.Project()))
		if res.Source == profile.SourceLocal {
			fmt.Fprintf(out, "Source:  %s\n", res.Path)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
