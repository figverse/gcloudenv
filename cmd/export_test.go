package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/figverse/gcloudenv/internal/gcloud"
)

// stubGcloud writes a fake gcloud that always lists the given profile, and
// points the package runner at it for the duration of the test.
func stubGcloud(t *testing.T, profileName string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub uses a POSIX shell script")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "gcloud")
	listJSON := `[{"name":"` + profileName + `","is_active":true,"properties":{"core":{"account":"a@x.com","project":"p"}}}]`
	script := "#!/bin/sh\ncat <<'EOF'\n" + listJSON + "\nEOF\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := runner
	runner = gcloud.Runner{Bin: path}
	t.Cleanup(func() { runner = orig })
}

func runExport(t *testing.T, args ...string) string {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs(append([]string{"export"}, args...))
	t.Cleanup(func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil) })
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("export %v: %v", args, err)
	}
	return buf.String()
}

func TestExportExportsADCWhenProfileHasIt(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(gcloud.EnvConfigDir, cfg)
	stubGcloud(t, "staging")

	adcDir := filepath.Join(cfg, "profiles", "staging")
	if err := os.MkdirAll(adcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	adcPath := filepath.Join(adcDir, gcloud.ADCFileName)
	if err := os.WriteFile(adcPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	out := runExport(t, "staging")
	if !strings.Contains(out, gcloud.EnvActiveConfig) {
		t.Errorf("missing active-config export:\n%s", out)
	}
	if !strings.Contains(out, gcloud.EnvADC) || !strings.Contains(out, adcPath) {
		t.Errorf("expected ADC export pointing at %s:\n%s", adcPath, out)
	}
	if strings.Contains(out, "unset") {
		t.Errorf("should not unset ADC when the file exists:\n%s", out)
	}
}

func TestExportUnsetsADCWhenProfileLacksIt(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(gcloud.EnvConfigDir, cfg)
	stubGcloud(t, "staging")

	out := runExport(t, "staging")
	if !strings.Contains(out, "unset "+gcloud.EnvADC) {
		t.Errorf("expected ADC to be unset when no per-profile file exists:\n%s", out)
	}
}
