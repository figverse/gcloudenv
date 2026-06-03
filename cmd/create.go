package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	createProject string
	createAccount string
)

var createCmd = &cobra.Command{
	Use:   "create <profile>",
	Short: "Create a new gcloud profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ok, err := runner.Exists(name)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf("configuration %q already exists", name)
		}
		if err := runner.Create(name); err != nil {
			return err
		}
		if createAccount != "" {
			if err := runner.SetProperty(name, "account", createAccount); err != nil {
				return err
			}
		}
		if createProject != "" {
			if err := runner.SetProperty(name, "project", createProject); err != nil {
				return err
			}
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created profile %q. Switch to it with: gcloudenv use %s\n", name, name)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createProject, "project", "", "set core/project on the new profile")
	createCmd.Flags().StringVar(&createAccount, "account", "", "set core/account on the new profile")
	rootCmd.AddCommand(createCmd)
}
