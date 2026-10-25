package main

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	cli := writeTool(t, bundled, "kirokit", `echo bundled-kirokit`, `echo bundled-kirokit`)
	t.Setenv("KIROKIT_DESKTOP_CLI", cli)

	a := NewAppAt(filepath.Join(t.TempDir(), "data"))
	a.captureFn = func() EnvCapture {
		return EnvCapture{Vars: loginVars(map[string]string{
			"HOME": home, "USER": "tester", "SHELL": shell,
			"PATH": "/usr/bin:/bin:/usr/sbin:/sbin", "KIRO_HOME": "/captured/wrong",
		})}
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

// A size outside 1..1000 is dropped: it neither starts the shell nor reaches it.
func TestShell_ResizeOutsideTheLimitsStartsNothing(t *testing.T) {
	a, rec, _, _ := shellApp(t, "/bin/sh")

	for _, size := range [][2]int{{0, 24}, {80, 0}, {-1, 24}, {1001, 24}, {80, 1001}} {
		a.ShellResize(size[0], size[1])
		a.ShellWrite("echo should-not-run\r")
	}

	a.shell.mu.Lock()
	started := a.shell.sess != nil
	a.shell.mu.Unlock()
	if started {
		t.Error("an out-of-range size started the shell")
	}
	if got := rec.String(); got != "" {
		t.Errorf("the shell tab received output: %q", got)
	}
}

// A shell other than zsh gets the settings and the bundled CLI's folder only in
// its start environment; its startup files (Debian's /etc/profile resets
// PATH) may reorder PATH, which is what the note says.
func TestShell_OtherShellGetsTheVariablesAndANote(t *testing.T) {
	if onWindows {
		t.Skip("needs /bin/sh; see TestShell_WindowsOpensPowerShellWithTheSettings")
	}
	a, rec, home, cli := shellApp(t, "/bin/sh")
	if err := a.SaveSettings(Settings{KiroHome: filepath.Join(home, "kh")}); err != nil {
		t.Fatal(err)
	}
	a.refreshEnv()

	if note := a.State().Env.Note; !strings.Contains(note, "may not come first") {
		t.Errorf("Env.Note = %q, want it to say the bundled CLI may not come first", note)
	}

	a.ShellResize(200, 24)
	a.ShellWrite("echo \"[$KIROKIT_REPO] [$KIRO_HOME] [$KIROKIT_BIN]\"\r")

	want := "[" + filepath.Join(a.dataDir, "repo") + "] [" + filepath.Join(home, "kh") + "] [" + filepath.Dir(cli) + "]"
	rec.waitFor(t, want)
}

// On Windows the shell tab is PowerShell, started with the app's settings in
// its environment and the bundled CLI's folder first on PATH.
func TestShell_WindowsOpensPowerShellWithTheSettings(t *testing.T) {
	if !onWindows {
		t.Skip("PowerShell is the shell tab on Windows only")
	}
	a, rec, home, cli := shellApp(t, "")
	if err := a.SaveSettings(Settings{KiroHome: filepath.Join(home, "kh")}); err != nil {
		t.Fatal(err)
	}
	a.refreshEnv()
	if note := a.State().Env.Note; note != "" {
		t.Errorf("Env.Note = %q, want none on Windows", note)
	}

	a.ShellResize(400, 30)
	a.ShellWrite(`Write-Output ("[{0}] [{1}] [{2}] [{3}]" -f $env:KIROKIT_REPO, $env:KIRO_HOME, ($env:PATH -split ';')[0], $PSVersionTable.PSVersion.Major)` + "\r")

	want := "[" + filepath.Join(a.dataDir, "repo") + "] [" + filepath.Join(home, "kh") + "] [" + filepath.Dir(cli) + "] ["
	rec.waitFor(t, want)
}

func TestBundledCLI_FixedOrder(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "kirokit.app", "Contents", "MacOS", "kirokit")
	inApp := writeScript(t, filepath.Join(root, "kirokit.app", "Contents", "Resources", "bin"), "kirokit", "")
	dev := writeScript(t, filepath.Join(root, "work", "build", "dev-bin"), "kirokit", "")
	devApp := filepath.Join(root, "work", "build", "bin", "kirokit.app", "Contents", "MacOS", "kirokit")
	writeScript(t, filepath.Join(root, "work", "build", "bin", "kirokit.app", "Contents", "Resources", "bin"), "kirokit", "")
	override := writeScript(t, filepath.Join(root, "custom"), "kirokit", "")
	work := filepath.Join(root, "work")

	for _, tc := range []struct{ name, override, exe, wd, want string }{
		{"override wins", override, app, work, override},
		{"app resources", "", app, work, inApp},
		{"dev-bin by working directory", "", filepath.Join(root, "elsewhere", "kirokit"), work, dev},
		// wails dev runs the app that an earlier desktop:build left in build/bin,
		// bundled CLI included; the dev-bin build is the one under test.
		{"wails dev prefers dev-bin over a stale bundled CLI", "", devApp, work, dev},
	} {
		got, err := resolveCLIFor("darwin", tc.override, tc.exe, tc.wd)
		if err != nil || got != tc.want {
			t.Errorf("%s: resolveCLI = %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}

	_, err := resolveCLIFor("darwin", "", filepath.Join(root, "elsewhere", "kirokit"), filepath.Join(root, "empty"))
	if err == nil || !strings.Contains(err.Error(), "KIROKIT_DESKTOP_CLI") || !strings.HasSuffix(err.Error(), "or start the app from kirokit.app") {
		t.Errorf("a missing CLI must name KIROKIT_DESKTOP_CLI and kirokit.app, got %v", err)
	}
}

// The .deb/.rpm put the CLI in /usr/lib/kirokit-desktop beside
// /usr/bin/kirokit-desktop; the Windows setup puts bin\kirokit.exe beside
// kirokit-desktop.exe.
func TestBundledCLI_LinuxAndWindowsLayouts(t *testing.T) {
	root := t.TempDir()
	linuxExe := filepath.Join(root, "usr", "bin", "kirokit-desktop")
	linuxCLI := writeScript(t, filepath.Join(root, "usr", "lib", "kirokit-desktop"), "kirokit", "")
	winExe := filepath.Join(root, "Programs", "kirokit", "kirokit-desktop.exe")
	winCLI := writeScript(t, filepath.Join(root, "Programs", "kirokit", "bin"), "kirokit.exe", "")
	work := filepath.Join(root, "work")
	winDev := writeScript(t, filepath.Join(work, "build", "dev-bin"), "kirokit.exe", "")
	winDevApp := filepath.Join(work, "build", "bin", "kirokit-desktop.exe")
	writeScript(t, filepath.Join(work, "build", "bin", "bin"), "kirokit.exe", "")

	for _, tc := range []struct{ name, goos, exe, wd, want string }{
		{"linux package", "linux", linuxExe, root, linuxCLI},
		{"windows setup", "windows", winExe, root, winCLI},
		{"windows wails dev prefers dev-bin", "windows", winDevApp, work, winDev},
		{"windows dev-bin by working directory", "windows", filepath.Join(root, "elsewhere", "kirokit-desktop.exe"), work, winDev},
	} {
		got, err := resolveCLIFor(tc.goos, "", tc.exe, tc.wd)
		if err != nil || got != tc.want {
			t.Errorf("%s: resolveCLI = %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}

	for goos, hint := range map[string]string{
		"linux":   "or reinstall the kirokit-desktop package",
		"windows": "or reinstall kirokit with its setup program",
	} {
		_, err := resolveCLIFor(goos, "", filepath.Join(root, "elsewhere", "x"), filepath.Join(root, "empty"))
		if err == nil || !strings.HasSuffix(err.Error(), hint) {
			t.Errorf("%s: missing CLI error = %v, want it to end with %q", goos, err, hint)
		}
	}
}

func TestShellCommandFor_PerOS(t *testing.T) {
	pwsh := func(name string) (string, error) {
		if name == "pwsh.exe" {
			return `C:\Program Files\PowerShell\7\pwsh.exe`, nil
		}
		return "", exec.ErrNotFound
	}
	noPwsh := func(string) (string, error) { return "", exec.ErrNotFound }
	for _, tc := range []struct {
		name, goos, shell string
		look              func(string) (string, error)
		wantPath          string
		wantArgs          []string
		wantName          string
	}{
		{"macOS zsh", "darwin", "/bin/zsh", noPwsh, "/bin/zsh", []string{"-il"}, "zsh"},
		{"linux bash", "linux", "/usr/bin/bash", noPwsh, "/usr/bin/bash", []string{"-il"}, "bash"},
		{"windows pwsh", "windows", "", pwsh, `C:\Program Files\PowerShell\7\pwsh.exe`, []string{"-NoLogo"}, "PowerShell"},
		{"windows powershell", "windows", "/usr/bin/bash", noPwsh, "powershell.exe", []string{"-NoLogo"}, "PowerShell"},
	} {
		path, args := shellCommandFor(tc.goos, tc.shell, tc.look)
		if path != tc.wantPath || strings.Join(args, " ") != strings.Join(tc.wantArgs, " ") {
			t.Errorf("%s: shellCommandFor = %q %q, want %q %q", tc.name, path, args, tc.wantPath, tc.wantArgs)
		}
		if got := shellNameFor(tc.goos, tc.shell); got != tc.wantName {
			t.Errorf("%s: shellNameFor = %q, want %q", tc.name, got, tc.wantName)
		}
	}
	if note := shellNoteFor("windows", "/usr/bin/bash"); note != "" {
		t.Errorf("windows shell note = %q, want none (PowerShell gets the settings in its environment)", note)
	}
	if note := shellNoteFor("linux", "/usr/bin/bash"); !strings.Contains(note, "bash") {
		t.Errorf("linux bash note = %q, want it to name bash", note)
	}
}

func TestState_NamesTheShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		a := appWithEnv(t, map[string]string{})
		a.refreshEnv()
		if got := a.State().Env.Shell; got != "PowerShell" {
			t.Errorf("Env.Shell = %q, want PowerShell", got)
		}
		return
	}
	a := appWithEnv(t, map[string]string{"SHELL": "/bin/zsh"})
	a.refreshEnv()
	if got := a.State().Env.Shell; got != "zsh" {
		t.Errorf("Env.Shell = %q, want zsh", got)
	}
}
