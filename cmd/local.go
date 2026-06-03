package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

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

		existed := profile.LocalExists(cwd)
		path, err := profile.WriteLocal(cwd, name)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(out, "Wrote %s -> %s\n", path, name)

		return maybeOfferGitignore(cmd, cwd, existed)
	},
}

// maybeOfferGitignore prompts to add .gcloudenv to .gitignore, but only when
// the file was just created, we're inside a git repo, and .gitignore doesn't
// already list it. It stays silent (and does nothing) in non-interactive runs.
func maybeOfferGitignore(cmd *cobra.Command, dir string, existed bool) error {
	if existed || !profile.InGitRepo(dir) {
		return nil
	}
	has, err := profile.GitignoreHasLocal(dir)
	if err != nil {
		return err
	}
	if has || !isInteractive() {
		return nil
	}

	out := cmd.OutOrStdout()
	if !confirm(cmd.InOrStdin(), out, fmt.Sprintf("Add %s to .gitignore?", profile.LocalFile)) {
		return nil
	}
	gi, err := profile.AddToGitignore(dir)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Added %s to %s\n", profile.LocalFile, gi)
	return nil
}

// isInteractive reports whether stdin is a terminal, so we never block waiting
// for input in scripts, pipes, or CI.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// confirm asks a yes/no question, defaulting to yes on empty input. EOF or any
// other answer is treated as no.
func confirm(in io.Reader, out io.Writer, question string) bool {
	_, _ = fmt.Fprintf(out, "%s [Y/n] ", question)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return false // EOF / read error with no input
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(localCmd)
}
