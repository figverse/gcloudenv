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

func TestLocalExists(t *testing.T) {
	dir := t.TempDir()
	if LocalExists(dir) {
		t.Fatal("empty dir should not have a local file")
	}
	if _, err := WriteLocal(dir, "p"); err != nil {
		t.Fatal(err)
	}
	if !LocalExists(dir) {
		t.Fatal("local file should exist after WriteLocal")
	}
}

func TestGitignoreHasLocal(t *testing.T) {
	tests := []struct {
		name     string
		contents string // "" means no .gitignore file
		want     bool
	}{
		{"no gitignore", "", false},
		{"unrelated entries", "node_modules/\n*.log\n", false},
		{"plain entry", "*.log\n.gcloudenv\n", true},
		{"anchored entry", "/.gcloudenv\n", true},
		{"trailing slash", ".gcloudenv/\n", true},
		{"commented out", "# .gcloudenv\n", false},
		{"substring is not a match", ".gcloudenv.bak\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.contents != "" {
				if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(tt.contents), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got, err := GitignoreHasLocal(dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("GitignoreHasLocal(%q) = %v, want %v", tt.contents, got, tt.want)
			}
		})
	}
}

func TestAddToGitignore(t *testing.T) {
	t.Run("creates file when missing", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := AddToGitignore(dir); err != nil {
			t.Fatal(err)
		}
		has, _ := GitignoreHasLocal(dir)
		if !has {
			t.Fatal("entry not present after AddToGitignore")
		}
	})

	t.Run("appends with separating newline", func(t *testing.T) {
		dir := t.TempDir()
		// No trailing newline on the existing content.
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := AddToGitignore(dir); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(data); got != "*.log\n.gcloudenv\n" {
			t.Fatalf("got %q, want %q", got, "*.log\n.gcloudenv\n")
		}
	})
}

func TestInGitRepo(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if !InGitRepo(child) {
		t.Fatal("child of a repo should be detected as in-repo")
	}
	if InGitRepo(t.TempDir()) {
		t.Fatal("a plain temp dir should not be in a repo")
	}
}
