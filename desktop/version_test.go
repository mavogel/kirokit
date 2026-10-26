package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/checkout"
)

// waitRef waits for the next ref:done and returns it.
func (l *eventLog) waitRef(t *testing.T) refResult {
	t.Helper()
	l.wait(t, "ref:done", func() bool {
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.ref != nil
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	res := *l.ref
	l.ref = nil
	return res
}

// versionOrigin is a repository whose team t1 installs, tagged v1.0.0 on main,
// with a branch feature that adds a team t2.
func versionOrigin(t *testing.T) string {
	t.Helper()
	origin := newOrigin(t)
	write(t, filepath.Join(origin, "global", "steerings", ".keep"), "")
	write(t, filepath.Join(origin, "teams", "t1", "agents", "a1", "agent.json.tmpl"), `{"name": "a1"}`)
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "agent")
	git(t, origin, "tag", "v1.0.0")
	git(t, origin, "checkout", "-q", "-b", "feature")
	write(t, filepath.Join(origin, "teams", "t2", "manifest.yaml"), "team: t2\nglobalSteerings: []\n")
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "t2")
	git(t, origin, "checkout", "-q", "main")
	return origin
}

// clonedApp is a cloneApp whose clone of origin has finished.
func clonedApp(t *testing.T, origin string) (*App, *eventLog) {
	t.Helper()
	a, log := cloneApp(t)
	if err := a.Clone(origin); err != nil {
		t.Fatal(err)
	}
	if res := log.waitClone(t); !res.OK {
		t.Fatalf("clone:done = %+v", res)
	}
	return a, log
}

func TestCheckout_BeforeACloneIsEmpty(t *testing.T) {
	a, _ := cloneApp(t)
	got, err := a.Checkout()
	if err != nil || got.Current != (checkout.Ref{}) || got.Branches == nil || len(got.Branches) != 0 || got.Tags == nil || len(got.Tags) != 0 {
		t.Errorf("Checkout before a clone = %+v, %v; want empty lists and no error", got, err)
	}
}

func TestCheckout_ListsTheClonesVersions(t *testing.T) {
	a, _ := clonedApp(t, versionOrigin(t))
	got, err := a.Checkout()
	if err != nil {
		t.Fatal(err)
	}
	if got.Current.Kind != "tag" || got.Current.Name != "v1.0.0" || got.Current.Commit == "" {
		t.Errorf("Current = %+v, want tag v1.0.0", got.Current)
	}
	if !reflect.DeepEqual(got.Branches, []string{"feature", "main"}) || !reflect.DeepEqual(got.Tags, []string{"v1.0.0"}) || got.Newest != "v1.0.0" {
		t.Errorf("Checkout = %+v, want branches feature and main, tag and newest v1.0.0", got)
	}
}

func TestSwitchRef_SavesTheChoiceOnlyAfterTheSwitch(t *testing.T) {
	origin := versionOrigin(t)
	a, log := clonedApp(t, origin)
	installFixture(t, a.kiroHome()) // t1 is installed, so the switch reinstalls it
	repo := a.repoDir()

	if err := a.SwitchRef("feature"); err != nil {
		t.Fatal(err)
	}
	if res := log.waitRef(t); !res.OK || res.Error != "" {
		t.Fatalf("ref:done = %+v, want ok", res)
	}
	if !strings.Contains(log.output(), "▸ kirokit sync t1 --branch feature\r\n") {
		t.Errorf("the kirokit tab must show the switch, got %q", log.output())
	}
	if got := a.State().Settings.Ref; got != "feature" {
		t.Errorf("Ref = %q, want feature", got)
	}
	if b := git(t, repo, "symbolic-ref", "HEAD"); b != "refs/heads/feature" {
		t.Errorf("clone is on %q, want refs/heads/feature", b)
	}
	if cat, err := a.Catalog(); err != nil || len(cat.Teams) != 2 {
		t.Errorf("Catalog after the switch = %+v, %v; want t1 and the feature branch's t2", cat, err)
	}

	// A switch the CLI refuses reports its own message and keeps the choice.
	write(t, filepath.Join(repo, "teams", "t1", "manifest.yaml"), "team: t1\nglobalSteerings: [edited]\n")
	if err := a.SwitchRef("v1.0.0"); err != nil {
		t.Fatal(err)
	}
	res := log.waitRef(t)
	if res.OK || !strings.HasPrefix(res.Error, "kirokit sync: local changes in ") {
		t.Errorf("ref:done = %+v, want the CLI's local-changes message", res)
	}
	if got := a.State().Settings.Ref; got != "feature" {
		t.Errorf("Ref after a refused switch = %q, want feature", got)
	}
	git(t, repo, "checkout", "-q", "--", "teams")

	if err := a.SwitchRef("nope"); err != nil {
		t.Fatal(err)
	}
	if res := log.waitRef(t); res.OK || res.Error != "kirokit sync: --branch nope: no branch or tag of that name on origin" {
		t.Errorf("ref:done = %+v, want the unknown-name message", res)
	}

	// Newest release goes back to the newest tag.
	if err := a.SwitchRef(""); err != nil {
		t.Fatal(err)
	}
	if res := log.waitRef(t); !res.OK {
		t.Fatalf("ref:done = %+v, want ok", res)
	}
	if !strings.Contains(log.output(), "▸ kirokit sync t1 --latest-tag\r\n") {
		t.Errorf("Newest release must run --latest-tag, got %q", log.output())
	}
	if got := a.State().Settings.Ref; got != "" {
		t.Errorf("Ref = %q, want Newest release (\"\")", got)
	}
}

func TestSwitchRef_RefusesABadNameOrABusyTabBeforeAnyProcessStarts(t *testing.T) {
	a, log := cloneApp(t)
	for _, ref := range []string{"-b", "--upload-pack=x", "has space"} {
		if err := a.SwitchRef(ref); err == nil || !strings.Contains(err.Error(), "must be a branch or tag name") {
			t.Errorf("SwitchRef(%q) err = %v, want the name rule", ref, err)
		}
	}
	if out := log.output(); out != "" {
		t.Errorf("a refused name must start no process, the kirokit tab got %q", out)
	}

	cli := writeTool(t, t.TempDir(), "kirokit", `echo ready; read line`, "echo ready\nset /p line=")
	t.Setenv("KIROKIT_DESKTOP_CLI", cli)
	id, err := a.Run([]string{"sync"})
	if err != nil {
		t.Fatal(err)
	}
	log.waitOutput(t, "ready")
	if err := a.SwitchRef("main"); err == nil || err.Error() != "kirokit is still running a command" {
		t.Errorf("SwitchRef while a command runs: err = %v, want the still-running message", err)
	}
	a.RunWrite(id, "\r")
	log.waitExit(t)
}

func TestSyncFailure_NamesTheCLIsMessage(t *testing.T) {
	for _, tc := range []struct {
		tail string
		code int
		want string
	}{
		{"▸ kirokit sync --branch x\r\nkirokit sync: --branch x: no branch or tag of that name on origin\r\n", 1, "kirokit sync: --branch x: no branch or tag of that name on origin"},
		{"\x1b[0mkirokit sync: first\r\nkirokit sync: last\x1b[K\r\n", 1, "kirokit sync: last"},
		{"something else broke\r\n", 2, "kirokit sync failed (exit 2). See the kirokit tab."},
	} {
		if got := syncFailure(runResult{Code: tc.code, Tail: tc.tail}); got != tc.want {
			t.Errorf("syncFailure(%q) = %q, want %q", tc.tail, got, tc.want)
		}
	}
}
