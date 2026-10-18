package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/mavogel/kirokit/desktop/terminal"
)

// errRunning is what Run answers while a command is still going.
var errRunning = errors.New("kirokit is still running a command")

// runSize is the pseudo-terminal size of a button command; the kirokit tab
// wraps what it receives to its own width.
const (
	runCols = 120
	runRows = 30
)

// runHost holds the kirokit tab's one command at a time.
type runHost struct {
	mu       sync.Mutex
	sess     *terminal.Session // nil when no command runs
	id       int               // the last run id handed out
	reserved bool              // a multi-step job (the clone) owns the tab between its commands
}

// runResult is how a command ended: its exit code and the last of its output.
type runResult struct {
	Code int
	Tail string
}

// tailSize is how much of a command's output is kept for the job that waits
// on it.
const tailSize = 4096

// Run starts the bundled kirokit with args in the kirokit tab and returns the
// run id. The tab prints "▸ kirokit <args>" first and streams the output as
// term:kirokit; term:kirokit:exit follows when the command ends. A second Run
// while one runs is refused.
func (a *App) Run(args []string) (int, error) {
	cli, err := BundledCLI()
	if err != nil {
		return 0, err
	}
	return a.runProgram("kirokit", cli, args)
}

// runProgram runs any program in the kirokit tab under the one-command lock;
// name is what the header shows. Every run gets the login environment, the
// bundled kirokit first on PATH, KIROKIT_REPO and the Settings KIRO_HOME.
func (a *App) runProgram(name, program string, args []string) (int, error) {
	h := &a.runner
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sess != nil || h.reserved {
		return 0, errRunning
	}
	id, _, err := a.launchLocked(name, program, args)
	return id, err
}

// reserve takes the kirokit tab for a job of several commands, so no button
// command starts between them. release gives it back.
func (a *App) reserve() error {
	h := &a.runner
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sess != nil || h.reserved {
		return errRunning
	}
	h.reserved = true
	return nil
}

func (a *App) release() {
	h := &a.runner
	h.mu.Lock()
	h.reserved = false
	h.mu.Unlock()
}

// runStep runs one command of a reserved job and waits for it to end.
func (a *App) runStep(name, program string, args []string) (runResult, error) {
	h := &a.runner
	h.mu.Lock()
	_, done, err := a.launchLocked(name, program, args)
	h.mu.Unlock()
	if err != nil {
		return runResult{}, err
	}
	return <-done, nil
}

// launchLocked starts a command with h.mu held and returns its run id and the
// channel its result arrives on.
func (a *App) launchLocked(name, program string, args []string) (int, <-chan runResult, error) {
	h := &a.runner
	if h.sess != nil {
		return 0, nil, errRunning
	}

	cmd := exec.Command(program, args...)
	cli, _ := BundledCLI() //nolint:errcheck // a missing bundled CLI only leaves PATH as it was
	cmd.Env = a.terminalEnv(cli, nil)
	cmd.Dir = a.repoDir()
	if _, err := os.Stat(cmd.Dir); err != nil {
		cmd.Dir = homeDir()
	}
	sess, err := terminal.Start(cmd, runCols, runRows)
	if err != nil {
		return 0, nil, err
	}
	h.id++
	id := h.id
	h.sess = sess

	header := "▸ " + strings.Join(append([]string{name}, args...), " ") + "\r\n"
	a.emit("term:kirokit", chunkOf(id, []byte(header)))
	done := make(chan runResult, 1)
	// A kirokit sync can bring a new team or pin to the checkout.
	isSync := name == "kirokit" && len(args) > 0 && args[0] == "sync"
	go a.finishRun(id, sess, isSync, done)
	return id, done, nil
}

// finishRun streams a command's output and, once it ends, frees the lock,
// renders the catalog again after a sync, reads the install records and
// announces the exit.
func (a *App) finishRun(id int, sess *terminal.Session, isSync bool, done chan<- runResult) {
	var tail []byte
	stream(sess, func(b []byte) {
		tail = append(tail, b...)
		if len(tail) > tailSize {
			tail = tail[len(tail)-tailSize:]
		}
		a.emit("term:kirokit", chunkOf(id, b))
	})
	code, _ := sess.Wait() //nolint:errcheck // a wait error is shown as the exit code
	_ = sess.Close()       //nolint:errcheck // the process has ended

	h := &a.runner
	h.mu.Lock()
	h.sess = nil
	h.mu.Unlock()

	if isSync {
		a.refreshCatalog()
	}

	// The command may have installed or removed a team; show it now rather
	// than at the next poll.
	a.mu.Lock()
	p := a.poller
	a.mu.Unlock()
	if p != nil {
		p.Check()
	}
	a.emit("term:kirokit:exit", termExit{ID: id, Code: code})
	done <- runResult{Code: code, Tail: string(tail)}
}

// RunWrite sends keystrokes to the running command, such as a git sign-in
// answer. Input for any other run id is dropped.
func (a *App) RunWrite(id int, data string) {
	h := &a.runner
	h.mu.Lock()
	sess := h.sess
	current := h.id
	h.mu.Unlock()
	if sess == nil || id != current {
		return
	}
	_, _ = sess.Write([]byte(data)) //nolint:errcheck // a command that just ended reports through its exit
}

// close ends the running command at app shutdown.
func (h *runHost) close() {
	h.mu.Lock()
	sess := h.sess
	h.mu.Unlock()
	if sess != nil {
		_ = sess.Close() //nolint:errcheck // shutting down
	}
}
