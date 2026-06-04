// Package gcloud is a thin wrapper around the gcloud CLI. Everything that
// touches gcloud goes through Runner so it can be faked in tests by pointing
// the binary at a stub on PATH.
package gcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvActiveConfig is the variable gcloud reads to pick the active configuration
// for a single invocation. Exporting it in a shell is how we switch profiles
// per-shell without mutating gcloud's global state.
const EnvActiveConfig = "CLOUDSDK_ACTIVE_CONFIG_NAME"

// EnvADC is the variable client libraries (the Go/Python/etc. SDKs, Terraform,
// ...) read to locate Application Default Credentials, ahead of gcloud's
// well-known file. Pointing it at a per-profile file is how we isolate ADC,
// which gcloud configurations alone do not.
const EnvADC = "GOOGLE_APPLICATION_CREDENTIALS"

// EnvConfigDir overrides gcloud's configuration directory for an invocation.
const EnvConfigDir = "CLOUDSDK_CONFIG"

// ADCFileName is gcloud's well-known Application Default Credentials filename.
const ADCFileName = "application_default_credentials.json"

// Runner executes gcloud commands. The zero value runs the "gcloud" binary
// found on PATH.
type Runner struct {
	// Bin is the gcloud executable to run. Defaults to "gcloud".
	Bin string
}

func (r Runner) bin() string {
	if r.Bin != "" {
		return r.Bin
	}
	return "gcloud"
}

// Configuration mirrors a single entry from
// `gcloud config configurations list --format=json`.
type Configuration struct {
	Name       string                       `json:"name"`
	IsActive   bool                         `json:"is_active"`
	Properties map[string]map[string]string `json:"properties"`
}

// Account returns the core/account property, or "" if unset.
func (c Configuration) Account() string { return c.Properties["core"]["account"] }

// Project returns the core/project property, or "" if unset.
func (c Configuration) Project() string { return c.Properties["core"]["project"] }

// run executes gcloud with args and returns stdout, surfacing stderr on error.
func (r Runner) run(args ...string) ([]byte, error) {
	cmd := exec.Command(r.bin(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("gcloud %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.Bytes(), nil
}

// List returns all gcloud configurations.
func (r Runner) List() ([]Configuration, error) {
	out, err := r.run("config", "configurations", "list", "--format=json")
	if err != nil {
		return nil, err
	}
	var configs []Configuration
	if err := json.Unmarshal(out, &configs); err != nil {
		return nil, fmt.Errorf("parsing gcloud output: %w", err)
	}
	return configs, nil
}

// Exists reports whether a configuration with the given name is present.
func (r Runner) Exists(name string) (bool, error) {
	configs, err := r.List()
	if err != nil {
		return false, err
	}
	for _, c := range configs {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Get returns the named configuration.
func (r Runner) Get(name string) (Configuration, error) {
	configs, err := r.List()
	if err != nil {
		return Configuration{}, err
	}
	for _, c := range configs {
		if c.Name == name {
			return c, nil
		}
	}
	return Configuration{}, fmt.Errorf("no such configuration %q", name)
}

// Create creates a new (empty) configuration.
func (r Runner) Create(name string) error {
	_, err := r.run("config", "configurations", "create", name)
	return err
}

// Activate sets the gcloud global default configuration.
func (r Runner) Activate(name string) error {
	_, err := r.run("config", "configurations", "activate", name)
	return err
}

// SetProperty sets a config property (e.g. "project", "account") on a
// specific configuration without changing the active one.
func (r Runner) SetProperty(config, key, value string) error {
	_, err := r.run("config", "set", key, value, "--configuration", config)
	return err
}

// ActiveFromEnv returns the profile selected via the environment, if any.
func ActiveFromEnv() string { return os.Getenv(EnvActiveConfig) }

// ConfigDir returns gcloud's active configuration directory, mirroring gcloud's
// own resolution: $CLOUDSDK_CONFIG if set, else %APPDATA%\gcloud on Windows,
// else ~/.config/gcloud.
func ConfigDir() (string, error) {
	if dir := os.Getenv(EnvConfigDir); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "gcloud"), nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gcloud"), nil
}

// ProfileADCDir returns the per-profile directory under the gcloud config dir
// that holds an isolated ADC file: <config>/profiles/<name>.
func ProfileADCDir(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles", name), nil
}

// ProfileADCPath returns the per-profile ADC file path:
// <config>/profiles/<name>/application_default_credentials.json.
func ProfileADCPath(name string) (string, error) {
	dir, err := ProfileADCDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ADCFileName), nil
}

// HasProfileADC reports whether a per-profile ADC file exists for name.
func HasProfileADC(name string) bool {
	path, err := ProfileADCPath(name)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// LoginProfileADC runs the interactive `gcloud auth application-default login`
// isolated to a throwaway config dir, then installs the resulting credentials
// at the per-profile ADC path, creating <config>/profiles/<name> if needed.
// Isolating the login keeps it from clobbering gcloud's shared well-known ADC
// file or other profiles. It returns the destination path on success.
func (r Runner) LoginProfileADC(name string) (string, error) {
	dest, err := ProfileADCPath(name)
	if err != nil {
		return "", err
	}
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", err
	}

	// Run the login against a temp config dir on the same filesystem as the
	// destination so the credential it writes can be moved with a plain rename
	// and never touches the user's real gcloud home.
	tmp, err := os.MkdirTemp(destDir, ".adc-login-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	cmd := exec.Command(r.bin(), "auth", "application-default", "login")
	cmd.Env = append(os.Environ(), EnvConfigDir+"="+tmp)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gcloud auth application-default login: %w", err)
	}

	src := filepath.Join(tmp, ADCFileName)
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("login did not produce %s", ADCFileName)
	}
	if err := moveFile(src, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// moveFile renames src to dest, falling back to a copy when the two live on
// different filesystems.
func moveFile(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
