package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/figverse/gcloudenv/internal/gcloud"
)

// runner is the gcloud interop used by all commands. Tests can override
// runner.Bin to point at a stub.
var runner gcloud.Runner

// version is set at build time via -ldflags (see .goreleaser.yaml).
var version = "dev"

var rootCmd = &cobra.Command{
	Use:           "gcloudenv",
	Short:         "Manage gcloud configurations like nvm/rbenv manage versions",
	Long:          "gcloudenv switches the active gcloud configuration per-shell, sets a global default, and auto-switches based on a .gcloudenv file.",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gcloudenv:", err)
		os.Exit(1)
	}
}
