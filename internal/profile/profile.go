// Package profile resolves which gcloud configuration should be active,
// following the same precedence philosophy as nvm/rbenv: an explicit choice
// beats a directory-local file, which beats the environment, which beats the
// configured global default.
package profile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// gitignoreName is the conventional ignore-file name.
const gitignoreName = ".gitignore"

// LocalFile is the per-directory marker, analogous to .nvmrc / .ruby-version.
const LocalFile = ".gcloudenv"

// globalDir and globalFile locate the user-level default at ~/.gcloudenv/global,
// analogous to nvm's default alias / rbenv's global version.
const (
	globalDir  = ".gcloudenv"
	globalFile = "global"
)

// Source describes where a resolved profile name came from, for display.
type Source string

// The possible Source values, in resolution-precedence order.
const (
	SourceFlag   Source = "flag"
	SourceLocal  Source = "local file"
	SourceEnv    Source = "environment"
	SourceGlobal Source = "global file"
	SourceNone   Source = "none"
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
// When none of those apply it consults gcloudenv's own user-level default
// (~/.gcloudenv/global). It does not consult gcloud's own active configuration;
// callers fall back to that when Source is SourceNone.
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
	if name, path := readGlobal(); name != "" {
		return Resolution{Name: name, Source: SourceGlobal, Path: path}
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

// LocalExists reports whether dir already contains a .gcloudenv file.
func LocalExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, LocalFile))
	return err == nil
}

// GlobalPath returns the user-level default file path, ~/.gcloudenv/global.
func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, globalDir, globalFile), nil
}

// WriteGlobal records name as the user-level default profile, creating
// ~/.gcloudenv if needed, and returns the file path it wrote.
func WriteGlobal(name string) (string, error) {
	path, err := GlobalPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(name+"\n"), 0o644)
}

// readGlobal returns the user-level default profile name and the path it was
// read from, or empty strings when the file is absent or unreadable.
func readGlobal() (name, path string) {
	path, err := GlobalPath()
	if err != nil {
		return "", ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	return strings.TrimSpace(string(data)), path
}

// InGitRepo reports whether dir is inside a git working tree, by walking up
// looking for a .git entry (a directory for normal repos, a file for worktrees
// and submodules).
func InGitRepo(dir string) bool {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

// GitignoreHasLocal reports whether dir's .gitignore already lists the
// .gcloudenv pattern. A missing .gitignore counts as "not listed".
func GitignoreHasLocal(dir string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, gitignoreName))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if gitignoreEntryMatches(line) {
			return true, nil
		}
	}
	return false, nil
}

// gitignoreEntryMatches reports whether a single .gitignore line targets the
// local file, ignoring comments, blank lines, and common anchoring syntax
// (leading "/" or "./", trailing "/").
func gitignoreEntryMatches(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return false
	}
	line = strings.TrimPrefix(line, "/")
	line = strings.TrimPrefix(line, "./")
	line = strings.TrimSuffix(line, "/")
	return line == LocalFile
}

// AddToGitignore appends the .gcloudenv pattern to dir's .gitignore, creating
// the file if it does not exist, and returns the .gitignore path.
func AddToGitignore(dir string) (string, error) {
	path := filepath.Join(dir, gitignoreName)

	// Ensure we start the new entry on its own line when the existing file
	// doesn't end in a newline.
	prefix := ""
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		prefix = "\n"
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return path, err
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(prefix + LocalFile + "\n"); err != nil {
		return path, err
	}
	return path, nil
}
