package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/gcloud"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List gcloud profiles",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		configs, err := runner.List()
		if err != nil {
			return err
		}

		// The shell-selected profile (env) overrides gcloud's own active flag.
		envActive := gcloud.ActiveFromEnv()

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "\tPROFILE\tACCOUNT\tPROJECT")
		for _, c := range configs {
			active := c.IsActive
			if envActive != "" {
				active = c.Name == envActive
			}
			marker := " "
			if active {
				marker = "*"
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, c.Name, dash(c.Account()), dash(c.Project()))
		}
		return w.Flush()
	},
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func init() {
	rootCmd.AddCommand(listCmd)
}
