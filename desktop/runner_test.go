package main

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mavogel/kirokit/internal/install"
)

// eventLog records every event the app emits, in order.
type eventLog struct {
	mu     sync.Mutex
	events []string
	out    strings.Builder // term:kirokit output
	exits  []termExit
	last   []install.Record
	clone  *cloneResult // the last clone:done
	ref    *refResult   // the last ref:done
}

func (l *eventLog) clearClone() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clone = nil
}

func (l *eventLog) emit(event string, data any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
	switch event {
	case "clone:done":
		res := data.(cloneResult)
		l.clone = &res
	case "ref:done":
		res := data.(refResult)
		l.ref = &res
	case "term:kirokit":
		raw, _ := base64.StdEncoding.DecodeString(data.(termChunk).B64)
		l.out.Write(raw)
	case "term:kirokit:exit":
		l.exits = append(l.exits, data.(termExit))
	case "installs:changed":
		l.last, _ = data.([]install.Record)
	}
}

func (l *eventLog) output() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.String()
}

func (l *eventLog) wait(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s; output %q", what, l.output())
}

func (l *eventLog) waitOutput(t *testing.T, want string) {
	t.Helper()
	l.wait(t, "output "+want, func() bool { return strings.Contains(l.output(), want) })
}

func (l *eventLog) waitExit(t *testing.T) termExit {
	t.Helper()
	l.wait(t, "the exit event", func() bool {
		l.mu.Lock()
		defer l.mu.Unlock()
		return len(l.exits) > 0
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.exits[len(l.exits)-1]
}

// runnerApp's bundled kirokit runs script: sh off Windows, cmd on Windows.
func runnerApp(t *testing.T, script, cmd string) (*App, *eventLog, string) {
	t.Helper()
	a, _, home, _ := shellApp(t, "/bin/sh")
	cli := writeTool(t, t.TempDir(), "kirokit", script, cmd)
	t.Setenv("KIROKIT_DESKTOP_CLI", cli)
	if err := a.SaveSettings(Settings{KiroHome: filepath.Join(home, "kh")}); err != nil {
		t.Fatal(err)
	}
	log := &eventLog{}
	a.emitter = log.emit
	return a, log, filepath.Join(home, "kh")
}

func TestRunner_HeaderThenOutputThenExit(t *testing.T) {
	a, log, kiroHome := runnerApp(t, `echo "args=$*"; echo "repo=$KIROKIT_REPO home=$KIRO_HOME"; exit 4`,
		"echo args=%*\necho repo=%KIROKIT_REPO% home=%KIRO_HOME%\nexit /b 4")

	id, err := a.Run([]string{"install", "data-team"})
	if err != nil {
		t.Fatal(err)
	}
	exit := log.waitExit(t)

	if exit.ID != id || exit.Code != 4 {
		t.Errorf("exit = %+v, want id %d code 4", exit, id)
	}
	out := log.output()
	if !strings.HasPrefix(out, "▸ kirokit install data-team\r\n") {
		t.Errorf("output must start with the header, got %q", out)
	}
	for _, want := range []string{
		"args=install data-team",
		"repo=" + filepath.Join(a.dataDir, "repo") + " home=" + kiroHome,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%q", want, out)
		}
	}
}

func TestRunner_RefusesASecondCommandWhileOneRuns(t *testing.T) {
	a, log, _ := runnerApp(t, `echo ready; read line; echo "got $line"`, "echo ready\nset /p line=\necho got %line%")

	id, err := a.Run([]string{"sync"})
	if err != nil {
		t.Fatal(err)
	}
	log.waitOutput(t, "ready")

	if _, err := a.Run([]string{"sync"}); err == nil || err.Error() != "kirokit is still running a command" {
		t.Fatalf("second Run error = %v, want \"kirokit is still running a command\"", err)
	}

	// RunWrite reaches the running command's input, and the lock is released after the exit.
	a.RunWrite(id, "hello\r")
	log.waitOutput(t, "got hello")
	log.waitExit(t)
	if _, err := a.Run([]string{"version"}); err != nil {
		t.Fatalf("Run after the first ended: %v", err)
	}
}

func TestRunner_WriteForAnotherRunIsIgnored(t *testing.T) {
	a, log, _ := runnerApp(t, `echo ready; read line; echo "got [$line]"`, "echo ready\nset /p line=\necho got [%line%]")

	id, err := a.Run(nil)
	if err != nil {
		t.Fatal(err)
	}
	log.waitOutput(t, "ready")
	a.RunWrite(id+1, "stale\r")
	a.RunWrite(id, "right\r")

	log.waitOutput(t, "got [right]")
}

func TestRunner_MissingCLINamesTheOverride(t *testing.T) {
	a, _, _ := runnerApp(t, "", "")
	t.Setenv("KIROKIT_DESKTOP_CLI", "")
	t.Chdir(t.TempDir())

	_, err := a.Run([]string{"version"})
	if err == nil || !strings.Contains(err.Error(), "KIROKIT_DESKTOP_CLI") {
		t.Fatalf("Run error = %v, want a message naming KIROKIT_DESKTOP_CLI", err)
	}
}

func TestRunner_AnyProgramShowsItsOwnName(t *testing.T) {
	a, log, _ := runnerApp(t, "", "")
	prog := writeTool(t, t.TempDir(), "git", `echo from-git`, `echo from-git`)

	if _, err := a.runProgram("git", prog, []string{"clone", "x"}); err != nil {
		t.Fatal(err)
	}
	log.waitExit(t)
	if out := log.output(); !strings.HasPrefix(out, "▸ git clone x\r\n") || !strings.Contains(out, "from-git") {
		t.Errorf("output = %q, want the git header then its output", out)
	}
}

// A sync can bring a new team or pin, so the cached catalog is rendered again
// when it ends; any other command leaves the cache alone.
func TestRunner_SyncEndRendersTheCatalogAgain(t *testing.T) {
	a, log, _ := runnerApp(t, `d="$KIROKIT_REPO/teams/new-$1"; mkdir -p "$d"; printf 'team: new-%s\nglobalSteerings: []\n' "$1" > "$d/manifest.yaml"`,
		"mkdir \"%KIROKIT_REPO%\\teams\\new-%1\"\n(echo team: new-%1& echo globalSteerings: [])> \"%KIROKIT_REPO%\\teams\\new-%1\\manifest.yaml\"")
	repo := a.repoDir()
	write(t, filepath.Join(repo, "teams", "t1", "manifest.yaml"), "team: t1\nglobalSteerings: []\n")
	write(t, filepath.Join(repo, "env", "mcp-versions.yaml"), "x-mcp: \"1\"\n")
	if cat, err := a.Catalog(); err != nil || len(cat.Teams) != 1 {
		t.Fatalf("first Catalog = %+v, %v", cat, err)
	}

	if _, err := a.Run([]string{"install"}); err != nil {
		t.Fatal(err)
	}
	log.waitExit(t)
	if cat, _ := a.Catalog(); len(cat.Teams) != 1 {
		t.Fatalf("an install must not render again, got %d teams", len(cat.Teams))
	}

	if _, err := a.Run([]string{"sync", "--latest-tag"}); err != nil {
		t.Fatal(err)
	}
	log.wait(t, "the second exit", func() bool {
		log.mu.Lock()
		defer log.mu.Unlock()
		return len(log.exits) == 2
	})
	cat, err := a.Catalog()
	if err != nil || len(cat.Teams) != 3 || cat.Teams[1].Name != "new-sync" {
		t.Fatalf("Catalog after the sync = %+v, %v; want t1, new-install and new-sync", cat, err)
	}
}

// When a command ends the install records are read at once, so a team the
// command installed shows without waiting for the next poll.
func TestRunner_EndChecksTheInstallRecords(t *testing.T) {
	fixture := t.TempDir()
	installFixture(t, fixture)
	a, log, _ := runnerApp(t, `cp -R "`+fixture+`/." "$KIRO_HOME"`, `xcopy /E /I /Y /Q "`+fixture+`" "%KIRO_HOME%"`)
	a.poller = newInstallPoller(a.kiroHome, a.emit, time.Hour)
	t.Cleanup(a.poller.Stop)

	if _, err := a.Run([]string{"install", "t1"}); err != nil {
		t.Fatal(err)
	}
	log.waitExit(t)

	log.wait(t, "installs:changed", func() bool {
		log.mu.Lock()
		defer log.mu.Unlock()
		return len(log.last) == 1 && log.last[0].Team == "t1"
	})
}
