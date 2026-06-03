// Package shell holds the sourced shim templates and helpers for emitting
// shell-appropriate environment statements.
package shell

import (
	"embed"
	"fmt"
)

//go:embed templates/*
var templates embed.FS

// Supported shells for `gcloudenv init`.
var Supported = []string{"bash", "zsh", "fish"}

// Init returns the integration snippet to be sourced for the given shell.
// bash and zsh share a template.
func Init(name string) (string, error) {
	var file string
	switch name {
	case "bash", "zsh":
		file = "templates/bash.sh"
	case "fish":
		file = "templates/fish.fish"
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", name)
	}
	data, err := templates.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExportStatement renders "set VAR to value" in the syntax of the target
// shell. shellName "fish" uses `set -gx`; everything else uses POSIX `export`.
func ExportStatement(shellName, key, value string) string {
	if shellName == "fish" {
		return fmt.Sprintf("set -gx %s %q", key, value)
	}
	return fmt.Sprintf("export %s=%q", key, value)
}
