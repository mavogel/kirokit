package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// captureMarker precedes the environment in the login shell's output, so
// whatever the startup files print before it is ignored.
const captureMarker = "__KIROKIT_ENV_BEGIN__"

// captureTimeout is how long the login shell may take before the app moves on.
const captureTimeout = 5 * time.Second

// EnvCapture is the environment the person's login shell produces. A GUI app
// starts with launchd's minimal PATH, so tools are looked up here instead.
type EnvCapture struct {
	Vars map[string]string
	// Note is set when the capture fell back; Settings shows it.
	Note string
}

// Getenv returns a captured variable, empty when unset.
func (e EnvCapture) Getenv(name string) string { return e.Vars[name] }

// Lookup resolves an executable on the captured PATH.
func (e EnvCapture) Lookup(name string) (string, bool) {
	for _, dir := range filepath.SplitList(e.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0 {
			return p, true
		}
	}
	return "", false
}

// toolFolders are where Homebrew and per-user installs put sops and kiro-cli.
func toolFolders(home string) []string {
	return []string{"/opt/homebrew/bin", "/opt/homebrew/sbin", "/usr/local/bin", filepath.Join(home, ".local", "bin")}
}

// withToolFolders appends the tool folders missing from path, so a login shell
// without startup files still finds Homebrew's tools.
func withToolFolders(path, home string) string {
	have := map[string]bool{}
	var parts []string
	for _, p := range filepath.SplitList(path) {
		if p != "" {
			have[p] = true
			parts = append(parts, p)
		}
	}
	for _, dir := range toolFolders(home) {
		if !have[dir] {
			parts = append(parts, dir)
		}
	}
	return strings.Join(parts, ":")
}

// Capture runs shell as a login shell, reads the environment it prints after
// captureMarker and returns it. When the shell fails or takes longer than
// timeout, it returns the app's own environment with a fallback PATH and a
// Note saying so.
func Capture(shell string, timeout time.Duration) EnvCapture {
	vars, note := runLoginShell(shell, timeout)
	if vars == nil {
		vars = map[string]string{}
		for _, kv := range os.Environ() {
			if k, v, ok := strings.Cut(kv, "="); ok {
				vars[k] = v
			}
		}
		vars["PATH"] = "/usr/bin:/bin:/usr/sbin:/sbin"
	}
	home := vars["HOME"]
	if home == "" {
		home = homeDir()
	}
	vars["PATH"] = withToolFolders(vars["PATH"], home)
	return EnvCapture{Vars: vars, Note: note}
}

func runLoginShell(shell string, timeout time.Duration) (map[string]string, string) {
	if shell == "" {
		shell = "/bin/zsh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-ilc", "printf '%s' '"+captureMarker+"'; exec env -0")
	// The shell and everything it starts form one group, so a timeout ends all
	// of it and releases the output pipe.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, "Your login shell took more than " + timeout.String() + " to start, so kirokit is using a default PATH. Tools outside it may not be found."
		}
		return nil, "Your login shell could not be read (" + err.Error() + "), so kirokit is using a default PATH."
	}
	_, env, ok := bytes.Cut(out.Bytes(), []byte(captureMarker))
	if !ok {
		return nil, "Your login shell printed no environment, so kirokit is using a default PATH."
	}
	vars := map[string]string{}
	for _, kv := range bytes.Split(env, []byte{0}) {
		if k, v, ok := bytes.Cut(kv, []byte("=")); ok && len(k) > 0 {
			vars[string(k)] = string(v)
		}
	}
	return vars, ""
}
