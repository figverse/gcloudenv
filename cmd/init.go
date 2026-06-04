package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/shell"
)

var initCmd = &cobra.Command{
	Use:       "init <shell>",
	Short:     "Print the shell integration to source in your rc file",
	Long:      "Print the shell integration snippet. Add to your shell rc:\n  bash/zsh:  eval \"$(gcloudenv init zsh)\"\n  fish:      gcloudenv init fish | source",
	Args:      cobra.ExactArgs(1),
	ValidArgs: shell.Supported,
	RunE: func(cmd *cobra.Command, args []string) error {
		snippet, err := shell.Init(strings.ToLower(args[0]))
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), snippet)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
