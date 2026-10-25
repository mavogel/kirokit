package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeShell writes an executable script standing in for the login shell. It
// ignores the -ilc arguments and runs body. Windows reads no login shell, so
// tests that need one skip there.
func fakeShell(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Windows reads no login shell; see TestPlatform_WindowsEnvironmentWithoutAShell")
	}
	p := filepath.Join(t.TempDir(), "fakesh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// Startup files print banners before the marker; only what follows it is the
// environment. Reading everything would turn the banner into a variable name.
func TestCapture_ReadsOnlyAfterMarker(t *testing.T) {
	sh := fakeShell(t, `echo "Welcome back, banner=noise"; printf '%s' '`+captureMarker+`'; exec env -0`)
	t.Setenv("KIRO_HOME", "/captured/kiro")

	got := Capture(sh, 5*time.Second)

	if got.Note != "" {
		t.Errorf("Note = %q, want none for a shell that answered", got.Note)
	}
	if v := got.Getenv("KIRO_HOME"); v != "/captured/kiro" {
		t.Errorf("KIRO_HOME = %q, want /captured/kiro", v)
	}
	if _, ok := got.Vars["Welcome back, banner"]; ok {
		t.Error("startup noise before the marker was read as environment")
	}
}

// A shell that never answers must not hold the app: after the limit Capture
// returns a usable PATH and a note for Settings.
func TestCapture_SlowShellFallsBack(t *testing.T) {
	sh := fakeShell(t, `sleep 10`)

	start := time.Now()
	got := Capture(sh, 200*time.Millisecond)

	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("Capture took %v, want it to give up near the limit", d)
	}
	if got.Note == "" {
		t.Error("Note is empty, want a message Settings can show")
	}
	for _, dir := range toolFoldersFor(runtime.GOOS, os.Getenv("HOME")) {
		if !strings.Contains(got.Getenv("PATH"), dir) {
			t.Errorf("fallback PATH %q lacks %s", got.Getenv("PATH"), dir)
		}
	}
}

// A login shell without startup files reports a bare PATH; Homebrew's folder
// is appended so sops and kiro-cli are still found.
func TestCapture_AppendsToolFolders(t *testing.T) {
	sh := fakeShell(t, `printf '%s' '`+captureMarker+`'; printf 'PATH=/usr/bin:/bin\0'`)

	got := Capture(sh, 5*time.Second)

	parts := strings.Split(got.Getenv("PATH"), ":")
	if parts[0] != "/usr/bin" || parts[1] != "/bin" {
		t.Errorf("PATH = %q, want the shell's folders first", got.Getenv("PATH"))
	}
	for _, dir := range toolFoldersFor(runtime.GOOS, homeDir()) {
		n := 0
		for _, p := range parts {
			if p == dir {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s appears %d times in %q, want once", dir, n, got.Getenv("PATH"))
		}
	}
}
