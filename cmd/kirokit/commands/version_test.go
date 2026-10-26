package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmdPrintsInjectedValues(t *testing.T) {
	oldV, oldC, oldD := version, commit, date
	t.Cleanup(func() { version, commit, date = oldV, oldC, oldD })
	version, commit, date = "1.2.3", "abc1234", "2026-01-02T03:04:05Z"

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	want := "kirokit 1.2.3\ncommit: abc1234\nbuilt:  2026-01-02T03:04:05Z\n"
	if got := out.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// `kirokit --version` is how the install docs check an installed release.
func TestVersionFlagPrintsTheVersion(t *testing.T) {
	oldV := version
	t.Cleanup(func() { version = oldV })
	version = "1.2.3"

	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "kirokit 1.2.3\n" {
		t.Fatalf("got %q, want %q", got, "kirokit 1.2.3\n")
	}
}

func TestVersionCmdRejectsArgs(t *testing.T) {
	root := newRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"version", "extra"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "unknown command") && !strings.Contains(err.Error(), "accepts 0") {
		t.Fatalf("expected arg error, got %v", err)
	}
}
