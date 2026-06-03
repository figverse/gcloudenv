package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrecedence(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, LocalFile), []byte("from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		flag       string
		env        string
		wantName   string
		wantSource Source
	}{
		{"flag wins over everything", "from-flag", "from-env", "from-flag", SourceFlag},
		{"local file beats env", "", "from-env", "from-file", SourceLocal},
		{"env when no flag or file is in dir", "", "from-env", "from-env", SourceEnv},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchDir := dir
			if tt.wantSource == SourceEnv {
				searchDir = t.TempDir() // a dir with no .gcloudenv
			}
			res := Resolve(tt.flag, searchDir, tt.env)
			if res.Name != tt.wantName || res.Source != tt.wantSource {
				t.Fatalf("Resolve = {%q, %q}, want {%q, %q}", res.Name, res.Source, tt.wantName, tt.wantSource)
			}
		})
	}
}

func TestResolveWalksUp(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, LocalFile), []byte("parent-profile"), 0o644); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	res := Resolve("", child, "")
	if res.Name != "parent-profile" || res.Source != SourceLocal {
		t.Fatalf("Resolve from child = {%q, %q}, want parent-profile/local", res.Name, res.Source)
	}
}

func TestResolveNone(t *testing.T) {
	res := Resolve("", t.TempDir(), "")
	if res.Source != SourceNone {
		t.Fatalf("want SourceNone, got %q", res.Source)
	}
}
