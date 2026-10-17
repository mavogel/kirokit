package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"

	"github.com/mavogel/kirokit/desktop/terminal"
)

// termChunk is the payload of term:shell and term:kirokit: terminal bytes as
// base64, so no escape sequence is mangled in JSON. ID is the run id of a
// kirokit-tab command and absent for the shell.
type termChunk struct {
	ID  int    `json:"id,omitempty"`
	B64 string `json:"b64"`
}

// termExit is the payload of term:kirokit:exit.
type termExit struct {
	ID   int `json:"id"`
	Code int `json:"code"`
}

// BundledCLI returns the kirokit binary the app's terminals use.
// 1. $KIROKIT_DESKTOP_CLI if set. 2. <executable>/../../Resources/bin/kirokit.
// 3. desktop/build/dev-bin/kirokit, which `task desktop:dev` builds.
func BundledCLI() (string, error) {
	exe, _ := os.Executable() //nolint:errcheck // an empty path just skips the bundle lookup
	wd, _ := os.Getwd()       //nolint:errcheck // an empty path just skips the dev lookup
	return resolveCLI(os.Getenv("KIROKIT_DESKTOP_CLI"), exe, wd)
}

func resolveCLI(override, exe, wd string) (string, error) {
	if override != "" {
		return override, nil
	}
	var candidates []string
	if exe != "" {
		macOS := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(macOS, "..", "Resources", "bin", "kirokit"),
			// wails dev runs desktop/build/bin/kirokit.
			filepath.Join(macOS, "..", "dev-bin", "kirokit"))
	}
	if wd != "" {
		candidates = append(candidates, filepath.Join(wd, "build", "dev-bin", "kirokit"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return filepath.Clean(c), nil
		}
	}
	return "", errors.New("kirokit's command-line tool was not found. Set KIROKIT_DESKTOP_CLI to the path of a kirokit binary, or start the app from kirokit.app")
}

// shellPath is the person's login shell: the captured SHELL, else the app's
// own, else zsh.
func (a *App) shellPath() string {
	a.mu.Lock()
	env, ready := a.env, a.envReady
	a.mu.Unlock()
	if ready {
		if s := env.Getenv("SHELL"); s != "" {
			return s
		}
	}
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/zsh"
}

func isZsh(shell string) bool { return filepath.Base(shell) == "zsh" }

// shellNote tells Settings when the bundled CLI cannot be guaranteed first.
func shellNote(shell string) string {
	if isZsh(shell) {
		return ""
	}
	return "Your shell is " + filepath.Base(shell) + ", so the bundled kirokit may not come first on its PATH if your startup files change it."
}

// loginEnv is the captured login environment, read now when no one has yet.
func (a *App) loginEnv() EnvCapture {
	a.mu.Lock()
	env, ready := a.env, a.envReady
	a.mu.Unlock()
	if ready {
		return env
	}
	return a.refreshEnv()
}

// terminalEnv is the environment of every terminal the app starts: the login
// environment, a PATH with the bundled CLI's folder first, and the app's own
// repository and Kiro home. extra entries override the rest.
func (a *App) terminalEnv(cli string, extra map[string]string) []string {
	vars := map[string]string{}
	for k, v := range a.loginEnv().Vars {
		vars[k] = v
	}
	vars["TERM"] = "xterm-256color"
	vars["COLORTERM"] = "truecolor"
	vars["KIROKIT_REPO"] = a.repoDir()
	vars["KIRO_HOME"] = a.kiroHome()
	if cli != "" {
		vars["PATH"] = filepath.Dir(cli) + ":" + vars["PATH"]
	}
	for k, v := range extra {
		vars[k] = v
	}
	env := make([]string, 0, len(vars))
	for k, v := range vars {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return env
}

// wrapperFiles are the zsh startup files of the ZDOTDIR wrapper. Each sources
// the person's own file of the same name, then puts ZDOTDIR back so zsh finds
// the next wrapper file. The last-read files re-assert the app's settings, so
// the person's own startup files cannot override them.
var wrapperFiles = map[string]string{
	".zshenv":   wrapperSource(".zshenv", ""),
	".zprofile": wrapperSource(".zprofile", ""),
	".zshrc":    wrapperSource(".zshrc", wrapperSettings),
	".zlogin":   wrapperSource(".zlogin", wrapperSettings+wrapperRestore),
}

const wrapperSettings = `if [[ -n "$KIROKIT_BIN" ]]; then path=("$KIROKIT_BIN" ${path:#$KIROKIT_BIN}); fi
export KIROKIT_REPO="$KIROKIT_APP_REPO" KIRO_HOME="$KIROKIT_APP_KIRO_HOME"
`

// wrapperRestore leaves the finished shell, and any shell it starts, with the
// person's own ZDOTDIR.
const wrapperRestore = `if [[ -n "${KIROKIT_USER_ZDOTDIR+x}" ]]; then export ZDOTDIR="$KIROKIT_USER_ZDOTDIR"; else unset ZDOTDIR; fi
`

func wrapperSource(name, after string) string {
	return `# Written by kirokit: sources your own ` + name + `, then keeps kirokit's settings last.
__kirokit_wrapper="$ZDOTDIR"
ZDOTDIR="${KIROKIT_USER_ZDOTDIR:-$HOME}"
[[ -r "$ZDOTDIR/` + name + `" ]] && source "$ZDOTDIR/` + name + `"
ZDOTDIR="$__kirokit_wrapper"
unset __kirokit_wrapper
` + after
}

// writeWrapper writes the ZDOTDIR wrapper into dir and returns it.
func writeWrapper(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for name, body := range wrapperFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// shellHost is the zsh tab's one long-lived shell. It starts when the tab is
// first sized, and survives closing the dock.
type shellHost struct {
	mu         sync.Mutex
	sess       *terminal.Session
	cols, rows uint16
	ended      bool // the shell exited; only typing starts a new one
}

// shellCommand builds `$SHELL -il` for the zsh tab. For zsh it starts through
// the ZDOTDIR wrapper; for any other shell the settings are only in the
// start environment.
func (a *App) shellCommand() (*exec.Cmd, error) {
	shell := a.shellPath()
	cli, cliErr := BundledCLI()
	extra := map[string]string{}
	if cliErr == nil {
		extra["KIROKIT_BIN"] = filepath.Dir(cli)
	}
	if isZsh(shell) {
		wrapper, err := writeWrapper(filepath.Join(a.dataDir, "zdotdir"))
		if err != nil {
			return nil, err
		}
		extra["ZDOTDIR"] = wrapper
		extra["KIROKIT_APP_REPO"] = a.repoDir()
		extra["KIROKIT_APP_KIRO_HOME"] = a.kiroHome()
		if z, ok := a.loginEnv().Vars["ZDOTDIR"]; ok {
			extra["KIROKIT_USER_ZDOTDIR"] = z
		}
	}
	cmd := exec.Command(shell, "-il")
	cmd.Env = a.terminalEnv(cli, extra)
	cmd.Dir = a.repoDir()
	if _, err := os.Stat(cmd.Dir); err != nil {
		cmd.Dir = homeDir()
	}
	return cmd, nil
}

// ShellWrite sends keystrokes typed in the zsh tab to the shell.
func (a *App) ShellWrite(data string) {
	h := &a.shell
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sess == nil {
		if h.cols == 0 || !a.startShellLocked() {
			return
		}
	}
	_, _ = h.sess.Write([]byte(data)) //nolint:errcheck // a shell that just ended reports through its exit
}

// ShellResize tells the shell the size of the zsh tab, and starts it the
// first time the tab is sized.
func (a *App) ShellResize(cols, rows int) {
	if cols < 1 || rows < 1 || cols > 1000 || rows > 1000 {
		return
	}
	h := &a.shell
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cols, h.rows = uint16(cols), uint16(rows)
	if h.sess != nil {
		_ = h.sess.Resize(h.cols, h.rows) //nolint:errcheck // the next resize retries
		return
	}
	if !h.ended {
		a.startShellLocked()
	}
}

// startShellLocked starts the shell at the stored size and streams it. It
// reports whether a shell now runs; h.mu is held.
func (a *App) startShellLocked() bool {
	h := &a.shell
	cmd, err := a.shellCommand()
	var sess *terminal.Session
	if err == nil {
		sess, err = terminal.Start(cmd, h.cols, h.rows)
	}
	if err != nil {
		a.emit("term:shell", chunkOf(0, []byte("\x1b[31mcould not start your shell: "+err.Error()+"\x1b[0m\r\n")))
		h.ended = true
		return false
	}
	h.sess, h.ended = sess, false
	go func() {
		stream(sess, func(b []byte) { a.emit("term:shell", chunkOf(0, b)) })
		code, _ := sess.Wait() //nolint:errcheck // a wait error is shown as the exit code
		_ = sess.Close()       //nolint:errcheck // the process has ended
		h.mu.Lock()
		if h.sess == sess {
			h.sess, h.ended = nil, true
		}
		h.mu.Unlock()
		a.emit("term:shell", chunkOf(0, []byte(fmt.Sprintf("\r\n\x1b[2mshell ended (exit %d). Type to start a new one.\x1b[0m\r\n", code))))
	}()
	return true
}

// close ends the shell at app shutdown.
func (h *shellHost) close() {
	h.mu.Lock()
	sess := h.sess
	h.sess = nil
	h.ended = true
	h.mu.Unlock()
	if sess != nil {
		_ = sess.Close() //nolint:errcheck // shutting down
	}
}

func chunkOf(id int, b []byte) termChunk {
	return termChunk{ID: id, B64: base64.StdEncoding.EncodeToString(b)}
}

// stream copies a session's output to send until the terminal closes.
func stream(s *terminal.Session, send func([]byte)) {
	buf := make([]byte, 8192)
	for {
		n, err := s.Read(buf)
		if n > 0 {
			send(append([]byte(nil), buf[:n]...))
		}
		if err != nil {
			return
		}
	}
}
