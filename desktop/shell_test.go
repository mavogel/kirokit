package main

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// recorder collects the term:* output the app emits.
type recorder struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (r *recorder) emit(event string, data any) {
	if event != "term:shell" {
		return
	}
	chunk := data.(termChunk)
	raw, _ := base64.StdEncoding.DecodeString(chunk.B64)
	r.mu.Lock()
	r.buf.Write(raw)
	r.mu.Unlock()
}

func (r *recorder) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func (r *recorder) waitFor(t *testing.T, want string) string {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if s := r.String(); strings.Contains(s, want) {
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("output never contained %q; got %q", want, r.String())
	return ""
}

// writeScript writes an executable script at dir/name.
func writeScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// shellApp returns an App whose data folder, home folder and bundled CLI are
// temporary, with a login environment that points HOME at the temp home.
func shellApp(t *testing.T, shell string) (*App, *recorder, string, string) {
	t.Helper()
	home := t.TempDir()
	bundled := filepath.Join(t.TempDir(), "bundle")
	cli := writeScript(t, bundled, "kirokit", `echo bundled-kirokit`)
	t.Setenv("KIROKIT_DESKTOP_CLI", cli)

	a := NewAppAt(filepath.Join(t.TempDir(), "data"))
	a.captureFn = func() EnvCapture {
		return EnvCapture{Vars: map[string]string{
			"HOME": home, "USER": "tester", "SHELL": shell,
			"PATH": "/usr/bin:/bin:/usr/sbin:/sbin", "KIRO_HOME": "/captured/wrong",
		}}
	}
	rec := &recorder{}
	a.emitter = rec.emit
	t.Cleanup(func() { a.shutdown(context.Background()) })
	return a, rec, home, cli
}

func TestShell_BundledKirokitComesFirstAndSettingsWin(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not installed")
	}
	a, rec, home, cli := shellApp(t, zsh)

	// The person's own startup files put another kirokit first and set KIRO_HOME.
	other := filepath.Join(t.TempDir(), "older")
	writeScript(t, other, "kirokit", `echo older-kirokit`)
	rc := "export PATH=" + other + ":$PATH\nexport KIRO_HOME=/from-zshrc\nPROMPT='%# '\n"
	if err := os.WriteFile(filepath.Join(home, ".zshrc"), []byte(rc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.SaveSettings(Settings{KiroHome: filepath.Join(home, "kh")}); err != nil {
		t.Fatal(err)
	}

	wantVars := "[" + filepath.Join(a.dataDir, "repo") + "] [" + filepath.Join(home, "kh") + "]"

	a.ShellResize(100, 30)
	a.ShellWrite("command -v kirokit; echo \"[$KIROKIT_REPO] [$KIRO_HOME]\"\r")

	out := rec.waitFor(t, wantVars)
	if !strings.Contains(out, cli+"\r\n") {
		t.Errorf("command -v kirokit did not print %s:\n%q", cli, out)
	}
}

func TestShell_ResizeReachesTheShell(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not installed")
	}
	a, rec, _, _ := shellApp(t, zsh)

	a.ShellResize(100, 30)
	a.ShellResize(90, 20)
	a.ShellWrite("stty size\r")

	rec.waitFor(t, "20 90\r\n")
}

func TestShell_OtherShellGetsTheVariablesAndANote(t *testing.T) {
	a, rec, home, cli := shellApp(t, "/bin/sh")
	if err := a.SaveSettings(Settings{KiroHome: filepath.Join(home, "kh")}); err != nil {
		t.Fatal(err)
	}
	a.refreshEnv()

	if note := a.State().Env.Note; !strings.Contains(note, "may not come first") {
		t.Errorf("Env.Note = %q, want it to say the bundled CLI may not come first", note)
	}

	a.ShellResize(80, 24)
	a.ShellWrite("echo \"[$KIROKIT_REPO] [$KIRO_HOME]\"; command -v kirokit\r")

	out := rec.waitFor(t, filepath.Dir(cli)+"/kirokit\r\n")
	want := "[" + filepath.Join(a.dataDir, "repo") + "] [" + filepath.Join(home, "kh") + "]"
	if !strings.Contains(out, want) {
		t.Errorf("variables: want %s in:\n%q", want, out)
	}
}

func TestBundledCLI_FixedOrder(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "kirokit.app", "Contents", "MacOS", "kirokit")
	inApp := writeScript(t, filepath.Join(root, "kirokit.app", "Contents", "Resources", "bin"), "kirokit", "")
	dev := writeScript(t, filepath.Join(root, "work", "build", "dev-bin"), "kirokit", "")
	override := writeScript(t, filepath.Join(root, "custom"), "kirokit", "")
	work := filepath.Join(root, "work")

	for _, tc := range []struct{ name, override, exe, wd, want string }{
		{"override wins", override, app, work, override},
		{"app resources", "", app, work, inApp},
		{"dev-bin by working directory", "", filepath.Join(root, "elsewhere", "kirokit"), work, dev},
	} {
		got, err := resolveCLI(tc.override, tc.exe, tc.wd)
		if err != nil || got != tc.want {
			t.Errorf("%s: resolveCLI = %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}

	_, err := resolveCLI("", filepath.Join(root, "elsewhere", "kirokit"), filepath.Join(root, "empty"))
	if err == nil || !strings.Contains(err.Error(), "KIROKIT_DESKTOP_CLI") {
		t.Errorf("a missing CLI must name KIROKIT_DESKTOP_CLI, got %v", err)
	}
}
