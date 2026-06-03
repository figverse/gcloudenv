package gcloud

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeGcloud writes a stub "gcloud" script that emits the given JSON for
// `config configurations list` and returns a Runner pointed at it.
func fakeGcloud(t *testing.T, listJSON string) Runner {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub uses a POSIX shell script")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "gcloud")
	script := "#!/bin/sh\ncat <<'EOF'\n" + listJSON + "\nEOF\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return Runner{Bin: path}
}

const sample = `[
  {"name":"staging","is_active":false,"properties":{"core":{"account":"a@x.com","project":"stg"}}},
  {"name":"prod","is_active":true,"properties":{"core":{"account":"b@x.com","project":"prod-1"}}}
]`

func TestListParsesConfigurations(t *testing.T) {
	r := fakeGcloud(t, sample)
	configs, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 2 {
		t.Fatalf("got %d configs, want 2", len(configs))
	}
	if configs[1].Name != "prod" || !configs[1].IsActive {
		t.Fatalf("prod not parsed as active: %+v", configs[1])
	}
	if got := configs[0].Project(); got != "stg" {
		t.Fatalf("staging project = %q, want stg", got)
	}
	if got := configs[1].Account(); got != "b@x.com" {
		t.Fatalf("prod account = %q, want b@x.com", got)
	}
}

func TestExists(t *testing.T) {
	r := fakeGcloud(t, sample)
	if ok, _ := r.Exists("staging"); !ok {
		t.Fatal("staging should exist")
	}
	if ok, _ := r.Exists("nope"); ok {
		t.Fatal("nope should not exist")
	}
}
