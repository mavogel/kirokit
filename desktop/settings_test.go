package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestApp(t *testing.T) (*App, string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // the home folder Windows reports
	data := filepath.Join(t.TempDir(), "kirokit-desktop")
	return NewAppAt(data), data, home
}

// Saved values must survive a restart, which is a new App on the same folder.
func TestSaveSettings_SurvivesNewApp(t *testing.T) {
	app, data, home := newTestApp(t)
	want := Settings{RepoURL: "git@github.com:acme/kiro-configs.git", KiroHome: filepath.Join(home, "kiro")}

	if err := app.SaveSettings(want); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	if got := NewAppAt(data).State().Settings; got != want {
		t.Errorf("after restart Settings = %+v, want %+v", got, want)
	}
}

// A refused Kiro home must name its field and save nothing, so a bad value
// never replaces a good one.
func TestSaveSettings_RefusesKiroHome(t *testing.T) {
	app, data, home := newTestApp(t)
	good := Settings{KiroHome: filepath.Join(home, "kiro")}
	if err := app.SaveSettings(good); err != nil {
		t.Fatal(err)
	}

	for _, bad := range []string{"", "kiro", "~/.kiro", "/", home, home + "/", "/./"} {
		err := app.SaveSettings(Settings{KiroHome: bad})
		if err == nil || !strings.HasPrefix(err.Error(), "KiroHome: ") {
			t.Errorf("KiroHome %q: err = %v, want a KiroHome: message", bad, err)
		}
	}
	if got := NewAppAt(data).State().Settings; got != good {
		t.Errorf("Settings after refusals = %+v, want unchanged %+v", got, good)
	}
}

func TestSaveSettings_RepoURL(t *testing.T) {
	app, _, home := newTestApp(t)
	kiro := filepath.Join(home, "kiro")

	for _, ok := range []string{
		"https://github.com/acme/kiro-configs.git",
		"ssh://git@example.com/acme/configs.git",
		"git@github.com:acme/kiro-configs.git",
		"file:///tmp/origin",
		"/tmp/origin",
		"", // not configured yet
	} {
		if err := app.SaveSettings(Settings{RepoURL: ok, KiroHome: kiro}); err != nil {
			t.Errorf("RepoURL %q refused: %v", ok, err)
		}
	}
	for _, bad := range []string{
		"-oProxyCommand=x", "--upload-pack=x", "http://example.com/r.git", "ftp://example.com/r",
		"relative/path", "https://", "https://exa mple.com/r", "git@", "/tmp/a b",
	} {
		err := app.SaveSettings(Settings{RepoURL: bad, KiroHome: kiro})
		if err == nil || !strings.HasPrefix(err.Error(), "RepoURL: ") {
			t.Errorf("RepoURL %q: err = %v, want a RepoURL: message", bad, err)
		}
	}
}

func TestSettings_WindowsDrivePaths(t *testing.T) {
	for _, ok := range []string{`C:\repos\team`, `C:/repos/team`, `d:\x`} {
		if !repoURLPattern.MatchString(ok) {
			t.Errorf("address %q refused, want accepted", ok)
		}
	}
	for _, bad := range []string{`C:`, `C:repo`, `-C:\x`, `C:\a b`, `CD:\x`} {
		if repoURLPattern.MatchString(bad) {
			t.Errorf("address %q accepted, want refused", bad)
		}
	}
	for _, tc := range []struct {
		goos, path string
		want       bool
	}{
		{"windows", `C:\`, true}, {"windows", `C:/`, true}, {"windows", `D:\`, true}, {"windows", `C:\\`, true},
		{"windows", `C:\Users\alex\.kiro`, false}, {"windows", `\`, true},
		{"linux", "/", true}, {"darwin", "/", true}, {"linux", "/home/alex", false},
	} {
		if got := isRootFolderFor(tc.goos, tc.path); got != tc.want {
			t.Errorf("isRootFolderFor(%q, %q) = %v, want %v", tc.goos, tc.path, got, tc.want)
		}
	}
}

// Both fields are reported together so the page can mark each one.
func TestSaveSettings_ReportsBothFields(t *testing.T) {
	app, _, _ := newTestApp(t)

	err := app.SaveSettings(Settings{RepoURL: "-x", KiroHome: "/"})

	if err == nil {
		t.Fatal("want an error")
	}
	if lines := strings.Split(err.Error(), "\n"); len(lines) != 2 {
		t.Errorf("errors = %q, want one line per field", lines)
	}
}

func TestSaveSettings_WritesPrivateFile(t *testing.T) {
	app, data, home := newTestApp(t)
	if err := app.SaveSettings(Settings{KiroHome: filepath.Join(home, "kiro")}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(data, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !onWindows && info.Mode().Perm() != 0o600 { // Windows keeps no Unix permission bits
		t.Errorf("settings.json mode = %v, want 0600", info.Mode().Perm())
	}
}

// The clone lives in <app folder>/repo until a folder is saved; a saved folder
// moves where Clone and every page look, and survives a restart.
func TestRepoDir_DefaultsThenFollowsSettings(t *testing.T) {
	app, data, home := newTestApp(t)
	if got, want := app.repoDir(), filepath.Join(data, "repo"); got != want {
		t.Errorf("default repoDir = %q, want %q", got, want)
	}

	dir := filepath.Join(home, "work", "kk")
	if err := app.SetRepoDir(dir + "/"); err != nil {
		t.Fatalf("SetRepoDir: %v", err)
	}
	if got := NewAppAt(data).repoDir(); got != dir {
		t.Errorf("repoDir after restart = %q, want %q", got, dir)
	}

	if err := app.SetRepoDir(""); err != nil {
		t.Fatal(err)
	}
	if got, want := app.repoDir(), filepath.Join(data, "repo"); got != want {
		t.Errorf("repoDir after reset = %q, want %q", got, want)
	}
}

func TestRepoDir_RefusesUnsafeFolders(t *testing.T) {
	app, _, home := newTestApp(t)
	busy := filepath.Join(home, "docs")
	if err := os.MkdirAll(busy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(busy, "a.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, bad := range []string{"repo", "/", home, busy} {
		if err := app.SetRepoDir(bad); err == nil || !strings.HasPrefix(err.Error(), "RepoDir: ") {
			t.Errorf("SetRepoDir(%q) err = %v, want a RepoDir: message", bad, err)
		}
		err := app.SaveSettings(Settings{KiroHome: filepath.Join(home, "kiro"), RepoDir: bad})
		if err == nil || !strings.HasPrefix(err.Error(), "RepoDir: ") {
			t.Errorf("SaveSettings RepoDir %q err = %v, want a RepoDir: message", bad, err)
		}
	}
}
