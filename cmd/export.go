package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/gcloud"
	"github.com/figverse/gcloudenv/internal/profile"
	"github.com/figverse/gcloudenv/internal/shell"
)

var (
	exportShell string
	exportAuto  bool
)

// exportCmd is consumed by the shell shim, not typically run by hand. It prints
// the statement that sets CLOUDSDK_ACTIVE_CONFIG_NAME so the parent shell can
// eval it.
var exportCmd = &cobra.Command{
	Use:    "export [profile]",
	Short:  "Print the shell statement that activates a profile (used by the shim)",
	Args:   cobra.MaximumNArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		flag := ""
		if len(args) == 1 {
			flag = args[0]
		}

		cwd, _ := os.Getwd()
		res := profile.Resolve(flag, cwd, gcloud.ActiveFromEnv())

		// In --auto mode we only emit when a directory-local file selects a
		// profile; otherwise stay silent so `cd` is a no-op.
		if exportAuto && res.Source != profile.SourceLocal {
			return nil
		}
		if res.Name == "" {
			return fmt.Errorf("no profile given and none found in %s or $%s", profile.LocalFile, gcloud.EnvActiveConfig)
		}

		ok, err := runner.Exists(res.Name)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no such configuration %q", res.Name)
		}

		out := cmd.OutOrStdout()
		_, _ = fmt.Fprintln(out, shell.ExportStatement(exportShell, gcloud.EnvActiveConfig, res.Name))

		// Isolate ADC when the profile has its own credentials file; otherwise
		// clear the variable so a previous profile's ADC never leaks into this
		// one (falling back to gcloud's shared well-known ADC).
		if path, err := gcloud.ProfileADCPath(res.Name); err == nil && gcloud.HasProfileADC(res.Name) {
			_, _ = fmt.Fprintln(out, shell.ExportStatement(exportShell, gcloud.EnvADC, path))
		} else {
			_, _ = fmt.Fprintln(out, shell.UnsetStatement(exportShell, gcloud.EnvADC))
		}
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportShell, "shell", "", "target shell syntax (fish|posix); default posix")
	exportCmd.Flags().BoolVar(&exportAuto, "auto", false, "only emit when a .gcloudenv file selects a profile")
	rootCmd.AddCommand(exportCmd)
}
