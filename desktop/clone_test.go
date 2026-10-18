package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// cleanGitEnv is the process environment without GIT_* variables, so a hook's
// GIT_DIR cannot redirect the test repositories.
func cleanGitEnv() []string { return withoutGitEnv(os.Environ()) }

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Env = cleanGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newOrigin makes a repository with one team, tagged with tags.
func newOrigin(t *testing.T, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	write(t, filepath.Join(dir, "teams", "t1", "manifest.yaml"), "team: t1\nglobalSteerings: []\n")
	write(t, filepath.Join(dir, "env", "mcp-versions.yaml"), "x-mcp: \"1\"\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "init")
	for _, tag := range tags {
		git(t, dir, "tag", tag)
	}
	return dir
}

var (
	builtOnce sync.Once
	builtCLI  string
	builtErr  error
)

// realKirokit builds the repository's kirokit once for the whole test run.
func realKirokit(t *testing.T) string {
	t.Helper()
	builtOnce.Do(func() {
		dir, err := os.MkdirTemp("", "kirokit-clone-test-")
		if err != nil {
			builtErr = err
			return
		}
		builtCLI = filepath.Join(dir, "kirokit")
		// kirokit is the root module; the desktop module cannot build it.
		cmd := exec.Command("go", "build", "-o", builtCLI, "./cmd/kirokit")
		cmd.Dir = ".."
		if out, err := cmd.CombinedOutput(); err != nil {
			builtErr = &buildError{err, string(out)}
		}
	})
	if builtErr != nil {
		t.Fatal(builtErr)
	}
	return builtCLI
}

type buildError struct {
	err error
	out string
}

func (e *buildError) Error() string { return "go build kirokit: " + e.err.Error() + "\n" + e.out }

func cloneApp(t *testing.T) (*App, *eventLog) {
	t.Helper()
	a, log, _ := runnerApp(t, "")
	t.Setenv("KIROKIT_DESKTOP_CLI", realKirokit(t))
	// The login environment needs git on its PATH and no inherited GIT_*.
	a.captureFn = func() EnvCapture {
		return EnvCapture{Vars: map[string]string{"HOME": t.TempDir(), "PATH": os.Getenv("PATH"), "SHELL": "/bin/sh"}}
	}
	a.refreshEnv()
	return a, log
}

func (l *eventLog) waitClone(t *testing.T) cloneResult {
	t.Helper()
	l.wait(t, "clone:done", func() bool {
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.clone != nil
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	return *l.clone
}

func TestClone_FetchesTheRepositoryAndTheNewestReleaseTag(t *testing.T) {
	origin := newOrigin(t, "v1.0.0")
	git(t, origin, "commit", "-q", "--allow-empty", "-m", "later")
	git(t, origin, "tag", "v1.2.0")
	git(t, origin, "commit", "-q", "--allow-empty", "-m", "untagged")
	a, log := cloneApp(t)

	if err := a.Clone(origin); err != nil {
		t.Fatal(err)
	}
	res := log.waitClone(t)

	if !res.OK || res.Error != "" {
		t.Fatalf("clone:done = %+v, want ok", res)
	}
	repo := filepath.Join(a.dataDir, "repo")
	if tag := git(t, repo, "describe", "--tags", "--exact-match", "HEAD"); tag != "v1.2.0" {
		t.Errorf("clone is on %q, want v1.2.0", tag)
	}
	out := log.output()
	if !strings.Contains(out, "▸ git clone -- "+origin+" "+repo) || !strings.Contains(out, "▸ kirokit sync --latest-tag") || !strings.Contains(out, "checked out v1.2.0 at") {
		t.Errorf("the kirokit tab must show the git clone and the sync, got %q", out)
	}
	if !a.State().Cloned || a.State().Settings.RepoURL != origin {
		t.Errorf("State = %+v, want cloned with the address saved", a.State())
	}
	if cat, err := a.Catalog(); err != nil || len(cat.Teams) != 1 || cat.Tag != "v1.2.0" {
		t.Errorf("Catalog after the clone = %+v, %v; want team t1 on v1.2.0", cat, err)
	}
}

func TestClone_RefusesAnAddressBeforeAnyProcessStarts(t *testing.T) {
	a, log := cloneApp(t)
	for _, address := range []string{"", "-oops", "--upload-pack=touch /tmp/x", "ftp://host/x", "relative/path", "has space /x"} {
		if err := a.Clone(address); err == nil {
			t.Errorf("Clone(%q) was accepted", address)
		}
	}
	if out := log.output(); out != "" {
		t.Errorf("a refused address must start no process, the kirokit tab got %q", out)
	}
	if got := a.State().Settings.RepoURL; got != "" {
		t.Errorf("a refused address must not be saved, got %q", got)
	}
}

func TestClone_RefusesWhileACommandRuns(t *testing.T) {
	a, log := cloneApp(t)
	origin := newOrigin(t, "v1.0.0")
	cli := writeScript(t, t.TempDir(), "kirokit", `echo ready; read line`)
	t.Setenv("KIROKIT_DESKTOP_CLI", cli)
	id, err := a.Run([]string{"sync"})
	if err != nil {
		t.Fatal(err)
	}
	log.waitOutput(t, "ready")

	if err := a.Clone(origin); err == nil || err.Error() != "kirokit is still running a command" {
		t.Fatalf("Clone error = %v, want the still-running message", err)
	}
	a.RunWrite(id, "\r")
	log.waitExit(t)
}

func TestClone_ASecondAddressKeepsTheOldCloneUnderAnotherName(t *testing.T) {
	first := newOrigin(t, "v1.0.0")
	second := newOrigin(t, "v2.0.0")
	a, log := cloneApp(t)
	if err := a.Clone(first); err != nil {
		t.Fatal(err)
	}
	log.waitClone(t)
	log.clearClone()

	if err := a.Clone(second); err != nil {
		t.Fatal(err)
	}
	if res := log.waitClone(t); !res.OK {
		t.Fatalf("clone:done = %+v", res)
	}

	repo := filepath.Join(a.dataDir, "repo")
	if tag := git(t, repo, "describe", "--tags", "--exact-match", "HEAD"); tag != "v2.0.0" {
		t.Errorf("repo is on %q, want v2.0.0", tag)
	}
	olds, _ := filepath.Glob(filepath.Join(a.dataDir, "repo.old-*"))
	if len(olds) != 1 {
		t.Fatalf("old clones = %v, want exactly one repo.old-<time>", olds)
	}
	if tag := git(t, olds[0], "describe", "--tags", "--exact-match", "HEAD"); tag != "v1.0.0" {
		t.Errorf("old clone is on %q, want v1.0.0", tag)
	}
}

func TestClone_AFailedCloneRestoresTheOldOneAndKeepsTheOldAddress(t *testing.T) {
	good := newOrigin(t, "v1.0.0")
	a, log := cloneApp(t)
	if err := a.Clone(good); err != nil {
		t.Fatal(err)
	}
	log.waitClone(t)
	log.clearClone()

	missing := filepath.Join(t.TempDir(), "no-such-origin")
	if err := a.Clone(missing); err != nil {
		t.Fatal(err)
	}
	res := log.waitClone(t)

	if res.OK || res.Error == "" {
		t.Fatalf("clone:done = %+v, want a failure with a message", res)
	}
	repo := filepath.Join(a.dataDir, "repo")
	if tag := git(t, repo, "describe", "--tags", "--exact-match", "HEAD"); tag != "v1.0.0" {
		t.Errorf("the old clone was not restored: repo is on %q", tag)
	}
	if olds, _ := filepath.Glob(filepath.Join(a.dataDir, "repo.old-*")); len(olds) != 0 {
		t.Errorf("a restored clone leaves no repo.old-*, got %v", olds)
	}
	if got := a.State().Settings.RepoURL; got != good {
		t.Errorf("saved address = %q, want the address still in use %q (the failed %q must not be stored)", got, good, missing)
	}
}

func TestClone_AFailedFirstCloneStoresNoAddress(t *testing.T) {
	a, log := cloneApp(t)
	missing := filepath.Join(t.TempDir(), "no-such-origin")
	if err := a.Clone(missing); err != nil {
		t.Fatal(err)
	}
	if res := log.waitClone(t); res.OK {
		t.Fatalf("clone:done = %+v, want a failure", res)
	}
	if got := a.State().Settings.RepoURL; got != "" {
		t.Errorf("saved address = %q, want none after a failed clone", got)
	}
}

func TestClone_AnOriginWithoutAReleaseTagKeepsTheClone(t *testing.T) {
	origin := newOrigin(t)
	a, log := cloneApp(t)

	if err := a.Clone(origin); err != nil {
		t.Fatal(err)
	}
	res := log.waitClone(t)

	want := "This repository has no release yet (no v* tag). Ask your team lead to tag one."
	if !res.OK || res.Error != want {
		t.Errorf("clone:done = %+v, want ok with %q", res, want)
	}
	if !a.State().Cloned {
		t.Error("the clone must be kept")
	}
	if got := a.State().Settings.RepoURL; got != origin {
		t.Errorf("saved address = %q, want %q: the kept clone is the one in use", got, origin)
	}
	if cat, err := a.Catalog(); err != nil || len(cat.Teams) != 1 {
		t.Errorf("Catalog = %+v, %v; the app works from the default branch", cat, err)
	}
}
