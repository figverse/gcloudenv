// Package profile resolves which gcloud configuration should be active,
// following the same precedence philosophy as nvm/rbenv: an explicit choice
// beats a directory-local file, which beats the environment, which beats the
// configured global default.
package profile

import (
	"os"
	"path/filepath"
	"strings"
)

// LocalFile is the per-directory marker, analogous to .nvmrc / .ruby-version.
const LocalFile = ".gcloudenv"

// Source describes where a resolved profile name came from, for display.
type Source string

const (
	SourceFlag  Source = "flag"
	SourceLocal Source = "local file"
	SourceEnv   Source = "environment"
	SourceNone  Source = "none"
)

// Resolution is the outcome of resolving the active profile.
type Resolution struct {
	Name   string
	Source Source
	// Path is the .gcloudenv file used, when Source is SourceLocal.
	Path string
}

// Resolve picks a profile name given an explicit flag value (may be empty),
// the starting directory to search upward from, and the current environment.
// It does not consult gcloud's own global default; callers fall back to that
// when Source is SourceNone.
func Resolve(flag, startDir, envValue string) Resolution {
	if flag != "" {
		return Resolution{Name: flag, Source: SourceFlag}
	}
	if name, path := findLocal(startDir); name != "" {
		return Resolution{Name: name, Source: SourceLocal, Path: path}
	}
	if envValue != "" {
		return Resolution{Name: envValue, Source: SourceEnv}
	}
	return Resolution{Source: SourceNone}
}

// findLocal walks up from dir looking for a LocalFile, returning the trimmed
// profile name it contains and the path it was read from.
func findLocal(dir string) (name, path string) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", ""
	}
	for {
		candidate := filepath.Join(dir, LocalFile)
		if data, err := os.ReadFile(candidate); err == nil {
			return strings.TrimSpace(string(data)), candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "" // reached filesystem root
		}
		dir = parent
	}
}

// WriteLocal writes a .gcloudenv file naming the profile in dir.
func WriteLocal(dir, name string) (string, error) {
	path := filepath.Join(dir, LocalFile)
	return path, os.WriteFile(path, []byte(name+"\n"), 0o644)
}
