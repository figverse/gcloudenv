package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/gcloud"
)

var adcCmd = &cobra.Command{
	Use:   "adc",
	Short: "Manage per-profile Application Default Credentials (ADC)",
	Long: `Manage per-profile Application Default Credentials (ADC).

gcloud configurations isolate the CLI's account and project but share a single
ADC file, so client libraries (the Go/Python/etc. SDKs, Terraform, ...) cannot
tell profiles apart. These commands give each profile its own ADC file under
<gcloud-config>/profiles/<profile>; switching to the profile then exports
GOOGLE_APPLICATION_CREDENTIALS so SDKs pick up the right identity.`,
}

var adcLoginCmd = &cobra.Command{
	Use:   "login <profile>",
	Short: "Run application-default login and store the credentials for a profile",
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

		dest, err := runner.LoginProfileADC(name)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Stored ADC for profile %q at:\n  %s\nIt takes effect on: gcloudenv use %s\n", name, dest, name)
		return nil
	},
}

var adcStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which profiles have isolated ADC",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		configs, err := runner.List()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "PROFILE\tADC")
		for _, c := range configs {
			status := "shared (gcloud default)"
			if gcloud.HasProfileADC(c.Name) {
				path, _ := gcloud.ProfileADCPath(c.Name)
				status = path
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\n", c.Name, status)
		}
		return w.Flush()
	},
}

func init() {
	adcCmd.AddCommand(adcLoginCmd)
	adcCmd.AddCommand(adcStatusCmd)
	rootCmd.AddCommand(adcCmd)
}
