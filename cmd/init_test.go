package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestInitWritesSnippetToStdout(t *testing.T) {
	for _, sh := range []string{"bash", "zsh", "fish"} {
		t.Run(sh, func(t *testing.T) {
			stdout := captureStdStreams(t, func() {
				rootCmd.SetArgs([]string{"init", sh})
				t.Cleanup(func() { rootCmd.SetArgs(nil) })
				if err := rootCmd.Execute(); err != nil {
					t.Fatalf("init %s: %v", sh, err)
				}
			})
			if stdout.out == "" {
				t.Errorf("init %s: snippet not written to stdout", sh)
			}
			if stdout.err != "" {
				t.Errorf("init %s: unexpected stderr output: %q", sh, stdout.err)
			}
			if !strings.Contains(stdout.out, "gcloudenv") {
				t.Errorf("init %s: stdout does not look like the shim:\n%s", sh, stdout.out)
			}
		})
	}
}

type capturedStreams struct{ out, err string }

// captureStdStreams redirects os.Stdout and os.Stderr to pipes for the duration
// of fn and returns what was written to each.
func captureStdStreams(t *testing.T, fn func()) capturedStreams {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = origOut, origErr }()

	fn()

	_ = outW.Close()
	_ = errW.Close()
	outBytes, _ := io.ReadAll(outR)
	errBytes, _ := io.ReadAll(errR)
	return capturedStreams{out: string(outBytes), err: string(errBytes)}
}
