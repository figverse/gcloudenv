// Package gcloud is a thin wrapper around the gcloud CLI. Everything that
// touches gcloud goes through Runner so it can be faked in tests by pointing
// the binary at a stub on PATH.
package gcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EnvActiveConfig is the variable gcloud reads to pick the active configuration
// for a single invocation. Exporting it in a shell is how we switch profiles
// per-shell without mutating gcloud's global state.
const EnvActiveConfig = "CLOUDSDK_ACTIVE_CONFIG_NAME"

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
