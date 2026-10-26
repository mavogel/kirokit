package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestPlatform_DataDirIsKirokitDesktopInTheUserConfigDir(t *testing.T) {
	if got, want := dataDirFor("/Users/alex/Library/Application Support"), filepath.Join("/Users/alex/Library/Application Support", "kirokit-desktop"); got != want {
		t.Errorf("dataDirFor = %q, want %q", got, want)
	}
}

// NewApp keeps macOS's data folder exactly where it is today.
func TestPlatform_NewAppDataFolderOnThisOS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	var want string
	switch runtime.GOOS {
	case "darwin":
		want = filepath.Join(home, "Library", "Application Support", "kirokit-desktop")
	case "linux":
		want = filepath.Join(home, ".config", "kirokit-desktop")
	case "windows":
		appData := t.TempDir()
		t.Setenv("AppData", appData)
		want = filepath.Join(appData, "kirokit-desktop")
	default:
		t.Skip("no expectation for " + runtime.GOOS)
	}
	if got := NewApp().dataDir; got != want {
		t.Errorf("NewApp().dataDir = %q, want %q", got, want)
	}
}

func TestPlatform_AgeKeyCandidate(t *testing.T) {
	const home, appData = "/home/alex", `C:\Users\alex\AppData\Roaming`
	env := func(kv ...string) EnvCapture {
		vars := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			vars[kv[i]] = kv[i+1]
		}
		return EnvCapture{Vars: vars}
	}
	tests := []struct {
		name, goos string
		env        EnvCapture
		want       string
	}{
		{"darwin default", "darwin", env("HOME", home), filepath.Join(home, "Library", "Application Support", "sops", "age", "keys.txt")},
		{"darwin XDG", "darwin", env("HOME", home, "XDG_CONFIG_HOME", "/x"), filepath.Join("/x", "sops", "age", "keys.txt")},
		{"linux default", "linux", env("HOME", home), filepath.Join(home, ".config", "sops", "age", "keys.txt")},
		{"linux XDG", "linux", env("HOME", home, "XDG_CONFIG_HOME", "/x"), filepath.Join("/x", "sops", "age", "keys.txt")},
		{"windows default", "windows", env("APPDATA", appData), filepath.Join(appData, "sops", "age", "keys.txt")},
		{"windows ignores XDG", "windows", env("APPDATA", appData, "XDG_CONFIG_HOME", "/x"), filepath.Join(appData, "sops", "age", "keys.txt")},
		{"windows without APPDATA", "windows", env(), filepath.Join("/fallback-config", "sops", "age", "keys.txt")},
		{"variable wins everywhere", "windows", env("SOPS_AGE_KEY_FILE", "/k.txt", "APPDATA", appData), "/k.txt"},
		{"variable wins on linux", "linux", env("SOPS_AGE_KEY_FILE", "/k.txt", "XDG_CONFIG_HOME", "/x"), "/k.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ageKeyCandidateFor(tc.goos, tc.env, "/fallback-home", "/fallback-config"); got != tc.want {
				t.Errorf("ageKeyCandidateFor = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPlatform_ToolFolders(t *testing.T) {
	const home = "/h"
	want := map[string][]string{
		"darwin":  {"/opt/homebrew/bin", "/opt/homebrew/sbin", "/usr/local/bin", filepath.Join(home, ".local", "bin")},
		"linux":   {"/home/linuxbrew/.linuxbrew/bin", "/usr/local/bin", filepath.Join(home, ".local", "bin")},
		"windows": {filepath.Join(home, "scoop", "shims")},
	}
	for goos, w := range want {
		if got := toolFoldersFor(goos, home); !slices.Equal(got, w) {
			t.Errorf("toolFoldersFor(%q) = %q, want %q", goos, got, w)
		}
	}
}

func TestPlatform_DefaultShell(t *testing.T) {
	none := func(string) bool { return false }
	hasBash := func(p string) bool { return p == "/bin/bash" }
	tests := []struct {
		goos, appShell string
		exists         func(string) bool
		want           string
	}{
		{"darwin", "", none, "/bin/zsh"},
		{"darwin", "/opt/homebrew/bin/fish", none, "/opt/homebrew/bin/fish"},
		{"linux", "/usr/bin/zsh", hasBash, "/usr/bin/zsh"},
		{"linux", "", hasBash, "/bin/bash"},
		{"linux", "", none, "/bin/sh"},
		{"windows", `C:\x\sh.exe`, hasBash, ""},
	}
	for _, tc := range tests {
		if got := defaultShellFor(tc.goos, tc.appShell, tc.exists); got != tc.want {
			t.Errorf("defaultShellFor(%q, %q) = %q, want %q", tc.goos, tc.appShell, got, tc.want)
		}
	}
}

// On Windows the app reads no login shell: its own environment is the user's.
// A Path key (Windows' spelling) becomes the one PATH, with the tool folders.
func TestPlatform_WindowsEnvironmentWithoutAShell(t *testing.T) {
	got := environmentFor("windows", []string{`Path=C:\Windows;C:\bin`, "USERPROFILE=" + `C:\Users\alex`, "KIRO_HOME=" + `C:\k`}, `C:\Users\alex`)
	if _, dup := got["Path"]; dup {
		t.Errorf("Path kept beside PATH: %q", got)
	}
	want := strings.Join([]string{`C:\Windows`, `C:\bin`, filepath.Join(`C:\Users\alex`, "scoop", "shims")}, ";")
	if got["PATH"] != want {
		t.Errorf("PATH = %q, want %q", got["PATH"], want)
	}
	if got["KIRO_HOME"] != `C:\k` {
		t.Errorf("KIRO_HOME = %q, want it kept", got["KIRO_HOME"])
	}
}

func TestPlatform_WindowsLookupTriesPathext(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	for _, p := range []string{filepath.Join(a, "sops.exe"), filepath.Join(b, "kiro-cli.cmd"), filepath.Join(a, "git.txt")} {
		if err := os.WriteFile(p, nil, 0o600); err != nil { // no execute bit: Windows has none
			t.Fatal(err)
		}
	}
	dirs := []string{a, b}
	const pathext = ".COM;.EXE;.BAT;.CMD"
	if p, ok := lookupFor("windows", dirs, pathext, "sops"); !ok || p != filepath.Join(a, "sops.exe") {
		t.Errorf("sops = %q, %v; want %s", p, ok, filepath.Join(a, "sops.exe"))
	}
	if p, ok := lookupFor("windows", dirs, pathext, "kiro-cli"); !ok || p != filepath.Join(b, "kiro-cli.cmd") {
		t.Errorf("kiro-cli = %q, %v; want %s", p, ok, filepath.Join(b, "kiro-cli.cmd"))
	}
	if p, ok := lookupFor("windows", dirs, pathext, "git"); ok {
		t.Errorf("git found at %q, want not found (.txt is not in PATHEXT)", p)
	}
	if _, ok := lookupFor("windows", dirs, "", "sops"); !ok {
		t.Error("an empty PATHEXT must fall back to Windows' default list")
	}
	// Off Windows a file without an execute bit is not a tool.
	if p, ok := lookupFor("linux", dirs, pathext, "sops.exe"); ok {
		t.Errorf("linux found the non-executable %q", p)
	}
}

func TestPlatform_SopsHint(t *testing.T) {
	want := map[string]string{
		"darwin":  "Install it with: brew install sops",
		"linux":   "Install it from your package manager, or from https://github.com/getsops/sops/releases",
		"windows": "Install it with: scoop install sops",
	}
	for goos, w := range want {
		if got := sopsHintFor(goos); got != w {
			t.Errorf("sopsHintFor(%q) = %q, want %q", goos, got, w)
		}
	}
}
