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
	if info.Mode().Perm() != 0o600 {
		t.Errorf("settings.json mode = %v, want 0600", info.Mode().Perm())
	}
}
