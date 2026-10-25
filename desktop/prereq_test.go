package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeBin makes a folder holding executable stubs with the given names.
func fakeBin(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, exeName(n)), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func appWithEnv(t *testing.T, vars map[string]string) *App {
	t.Helper()
	a := NewAppAt(t.TempDir())
	a.captureFn = func() EnvCapture { return EnvCapture{Vars: vars} }
	return a
}

func byName(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check named %q in %+v", name, checks)
	return Check{}
}

func TestPrerequisites_ResolveOnCapturedPath(t *testing.T) {
	both := fakeBin(t, "sops", "kiro-cli")
	none := fakeBin(t)
	key := filepath.Join(t.TempDir(), "keys.txt")
	if err := os.WriteFile(key, []byte("AGE-SECRET-KEY-1"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := appWithEnv(t, map[string]string{"PATH": both, "SOPS_AGE_KEY_FILE": key}).Prerequisites("platform-team")
	for _, name := range []string{"SOPS", "Decryption key", "Kiro CLI"} {
		if c := byName(t, got, name); !c.Found {
			t.Errorf("%s not found, want found: %+v", name, c)
		}
	}

	got = appWithEnv(t, map[string]string{"PATH": none, "HOME": t.TempDir(), "SOPS_AGE_KEY_FILE": filepath.Join(t.TempDir(), "absent")}).Prerequisites("platform-team")
	for _, name := range []string{"SOPS", "Decryption key", "Kiro CLI"} {
		if c := byName(t, got, name); c.Found || c.Detail == "" {
			t.Errorf("%s = %+v, want missing with a detail to act on", name, c)
		}
	}
}

func TestPrerequisites_KeyLocationOrder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows reads the key from %AppData%, never XDG_CONFIG_HOME; see TestPlatform_AgeKeyCandidate")
	}
	home := t.TempDir()
	xdg := t.TempDir()
	// sops's default without XDG_CONFIG_HOME: the OS's config folder.
	osDefault := filepath.Join(home, ".config", "sops", "age", "keys.txt")
	if runtime.GOOS == "darwin" {
		osDefault = filepath.Join(home, "Library", "Application Support", "sops", "age", "keys.txt")
	}
	xdgKey := filepath.Join(xdg, "sops", "age", "keys.txt")
	for _, p := range []string{osDefault, xdgKey} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("k"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	vars := map[string]string{"HOME": home, "XDG_CONFIG_HOME": xdg}

	if c := byName(t, appWithEnv(t, vars).Prerequisites("t"), "Decryption key"); c.Detail != xdgKey {
		t.Errorf("with XDG_CONFIG_HOME, key = %q, want %q", c.Detail, xdgKey)
	}

	delete(vars, "XDG_CONFIG_HOME")
	if c := byName(t, appWithEnv(t, vars).Prerequisites("t"), "Decryption key"); c.Detail != osDefault {
		t.Errorf("without XDG_CONFIG_HOME, key = %q, want %q", c.Detail, osDefault)
	}

	vars["SOPS_AGE_KEY_FILE"] = filepath.Join(home, "missing")
	if c := byName(t, appWithEnv(t, vars).Prerequisites("t"), "Decryption key"); c.Found {
		t.Errorf("SOPS_AGE_KEY_FILE points nowhere but the key was found at %q; the variable decides", c.Detail)
	}
}

// "Check again" must see a tool installed since the last look.
func TestPrerequisites_CapturesAfresh(t *testing.T) {
	bin := t.TempDir()
	a := appWithEnv(t, map[string]string{"PATH": bin})
	if byName(t, a.Prerequisites("t"), "SOPS").Found {
		t.Fatal("sops found before it exists")
	}
	if err := os.WriteFile(filepath.Join(bin, exeName("sops")), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !byName(t, a.Prerequisites("t"), "SOPS").Found {
		t.Error("second call did not capture afresh")
	}
}

// The default Kiro home follows the captured KIRO_HOME once it arrives; a
// saved Settings value wins over it.
func TestKiroHome_CapturedThenSaved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := appWithEnv(t, map[string]string{"KIRO_HOME": "/captured/kiro"})

	if got := a.State().Env.DefaultKiroHome; got != filepath.Join(homeDir(), ".kiro") {
		t.Errorf("before capture, default = %q, want ~/.kiro", got)
	}
	a.refreshEnv()
	st := a.State()
	if !st.Env.Ready || st.Env.DefaultKiroHome != "/captured/kiro" {
		t.Errorf("after capture Env = %+v, want ready with /captured/kiro", st.Env)
	}

	saved := filepath.Join(homeDir(), "custom-kiro")
	if err := a.SaveSettings(Settings{KiroHome: saved}); err != nil {
		t.Fatal(err)
	}
	if got := a.kiroHome(); got != saved {
		t.Errorf("kiroHome = %q, want the saved %q", got, saved)
	}
}
