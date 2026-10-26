package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "core.hooksPath=" + os.DevNull, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	cmd.Env = cleanGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// cleanGitEnv returns the process environment without any GIT_* variable, so a
// foreign GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE cannot redirect test repos.
func cleanGitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	return env
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

func commitFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/f", []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "f")
	git(t, dir, "commit", "-q", "-m", content)
}

func TestUpdateCheckout(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v1.10.0")
	commitFile(t, origin, "three") // untagged tip of main

	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	chdir(t, clone)

	commitFile(t, origin, "four")
	if err := updateCheckout(false, ""); err != nil {
		t.Fatalf("branch update: %v", err)
	}
	if got, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "HEAD"); got != want {
		t.Errorf("branch update: HEAD = %s, want origin tip %s", got, want)
	}

	if err := updateCheckout(true, ""); err != nil {
		t.Fatalf("latest-tag update: %v", err)
	}
	// v1.10.0 must beat v1.0.0 (version sort, not lexical).
	if got, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "v1.10.0^{commit}"); got != want {
		t.Errorf("latest-tag: HEAD = %s, want v1.10.0 commit %s", got, want)
	}
}

func TestSync_WithoutATeamUpdatesTheCheckoutOnly(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	before := git(t, clone, "rev-parse", "--short", "HEAD")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v1.1.0")
	after := git(t, origin, "rev-parse", "--short", "HEAD")

	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)
	chdir(t, t.TempDir())

	got, err := runRoot(t, "sync", "--latest-tag", "--repo", clone)
	if err != nil {
		t.Fatalf("sync without a team: %v\n%s", err, got)
	}
	want := "updated " + before + " -> " + after + "\nchecked out v1.1.0 at " + after + "\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".kirokit-installed")); err == nil {
		t.Error("team-less sync installed something")
	}
}

func TestSync_KnowledgeBaseFlagsNeedATeam(t *testing.T) {
	for _, flag := range []string{"--kb-agent", "--kb-source", "--kb-mount"} {
		_, err := runRoot(t, "sync", flag, "x")
		if want := "kirokit sync: " + flag + " needs a team"; err == nil || err.Error() != want {
			t.Errorf("%s: err = %v, want %q", flag, err, want)
		}
	}
}

func TestSync_WithATeamRemovesOtherInstalledTeamsFirst(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)
	t.Setenv("KIROKIT_REPO", thisRepo(t))
	chdir(t, t.TempDir())
	if _, err := runRoot(t, "install", "platform-team"); err != nil {
		t.Fatal(err)
	}

	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	writeTree(t, origin, map[string]string{
		"env/mcp-versions.yaml":                "x: \"1\"\n",
		"global/steerings/.keep":               "",
		"teams/demo/manifest.yaml":             "team: demo\nglobalSteerings: []\n",
		"teams/demo/agents/a1/agent.json.tmpl": `{"name": "a1"}`,
	})
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "teams")
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")

	got, err := runRoot(t, "sync", "demo", "--repo", clone)
	if err != nil {
		t.Fatalf("sync demo: %v\n%s", err, got)
	}
	short := git(t, clone, "rev-parse", "--short", "HEAD")
	if want := "checked out main at " + short + "\nuninstalled platform-team from " + home + "\n"; !strings.HasPrefix(got, want) {
		t.Errorf("output = %q, want it to start with %q", got, want)
	}
	if !strings.Contains(got, "synced demo into "+home) {
		t.Errorf("output = %q, want today's synced line", got)
	}
	if recs := installedRecords(t, home); !reflect.DeepEqual(recs, []string{"demo.json"}) {
		t.Errorf("records = %v, want only demo.json", recs)
	}
}

func TestSync_FailureLeavesTheInstalledTeamInPlace(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	writeTree(t, origin, map[string]string{
		"env/mcp-versions.yaml":                "x: \"1\"\n",
		"global/steerings/.keep":               "",
		"teams/demo/manifest.yaml":             "team: demo\nglobalSteerings: []\n",
		"teams/demo/agents/a1/agent.json.tmpl": `{"name": "a1"}`,
	})
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "teams")

	cases := map[string]struct {
		team    string
		flags   []string
		prepare func(t *testing.T, clone string)
	}{
		"unknown team": {team: "nonexistent-team", prepare: func(*testing.T, string) {}},
		"failing git pull": {team: "demo", prepare: func(t *testing.T, clone string) {
			// Diverge clone and origin so pull --ff-only fails.
			commitFile(t, clone, "local")
			commitFile(t, origin, "remote")
		}},
		"unknown --branch": {team: "demo", flags: []string{"--branch", "nope"}, prepare: func(*testing.T, string) {}},
		"--branch with local changes": {team: "demo", flags: []string{"--branch", "main"}, prepare: func(t *testing.T, clone string) {
			writeTree(t, clone, map[string]string{"teams/demo/manifest.yaml": "team: demo\nglobalSteerings: [edited]\n"})
		}},
		"--branch with local commits": {team: "demo", flags: []string{"--branch", "main"}, prepare: func(t *testing.T, clone string) {
			commitFile(t, clone, "local work")
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("KIRO_HOME", home)
			t.Setenv("KIROKIT_REPO", thisRepo(t))
			chdir(t, t.TempDir())
			if _, err := runRoot(t, "install", "platform-team"); err != nil {
				t.Fatal(err)
			}
			want := installedRecords(t, home)

			clone := t.TempDir()
			git(t, clone, "clone", "-q", origin, ".")
			tc.prepare(t, clone)

			got, err := runRoot(t, append([]string{"sync", tc.team, "--repo", clone}, tc.flags...)...)
			if err == nil {
				t.Fatalf("sync %s succeeded:\n%s", tc.team, got)
			}
			if strings.Contains(got, "uninstalled") {
				t.Errorf("output = %q, must not uninstall anything", got)
			}
			if recs := installedRecords(t, home); !reflect.DeepEqual(recs, want) {
				t.Errorf("records = %v, want %v left in place", recs, want)
			}
		})
	}
}

func TestUpdateCheckout_LatestTagWithoutTagsFails(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	chdir(t, clone)

	if err := updateCheckout(true, ""); err == nil || !strings.Contains(err.Error(), "no v* tag") {
		t.Fatalf("err = %v, want a %q error", err, "no v* tag")
	}
}

// demoTree is a config repository with one team, demo.
var demoTree = map[string]string{
	"env/mcp-versions.yaml":                "x: \"1\"\n",
	"global/steerings/.keep":               "",
	"teams/demo/manifest.yaml":             "team: demo\nglobalSteerings: []\n",
	"teams/demo/agents/a1/agent.json.tmpl": `{"name": "a1"}`,
}

// branchOrigin is a config repository on main with tags v1.0.0 and v1.1.0, a
// feature branch one commit past v1.0.0, and a branch "both" at v1.0.0 next to
// a tag "both" on main's tip.
func branchOrigin(t *testing.T) string {
	t.Helper()
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	writeTree(t, origin, demoTree)
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "teams")
	git(t, origin, "tag", "v1.0.0")
	git(t, origin, "branch", "both")
	git(t, origin, "checkout", "-q", "-b", "feature")
	commitFile(t, origin, "feature-1")
	git(t, origin, "checkout", "-q", "main")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v1.1.0")
	commitFile(t, origin, "three")
	git(t, origin, "tag", "both")
	return origin
}

// cloneOf clones origin into a fresh folder and leaves the test in another one.
func cloneOf(t *testing.T, origin string) string {
	t.Helper()
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	chdir(t, t.TempDir())
	return clone
}

// headBranch is the branch HEAD is on, or "HEAD" when HEAD is detached.
func headBranch(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimPrefix(git(t, dir, "rev-parse", "--symbolic-full-name", "HEAD"), "refs/heads/")
}

func TestSync_BranchSwitchesToABranchOrTag(t *testing.T) {
	origin := branchOrigin(t)
	clone := cloneOf(t, origin)
	commitFile(t, origin, "four") // main moves on after the clone
	if err := os.WriteFile(filepath.Join(clone, "untracked.txt"), []byte("mine"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		ref, branch, commit string
	}{
		{"feature", "feature", "refs/heads/feature"}, // only on origin: a tracking branch is made
		{"v1.0.0", "HEAD", "v1.0.0^{commit}"},        // a tag: detached
		{"main", "main", "refs/heads/main"},          // the local main, fast-forwarded
		{"both", "both", "refs/heads/both"},          // a branch wins over a tag of the same name
	} {
		got, err := runRoot(t, "sync", "--repo", clone, "--branch", tc.ref)
		if err != nil {
			t.Fatalf("sync --branch %s: %v\n%s", tc.ref, err, got)
		}
		if head, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", tc.commit); head != want {
			t.Errorf("--branch %s: HEAD = %s, want %s", tc.ref, head, want)
		}
		if b := headBranch(t, clone); b != tc.branch {
			t.Errorf("--branch %s: on branch %q, want %q", tc.ref, b, tc.branch)
		}
		short := git(t, clone, "rev-parse", "--short", "HEAD")
		if line := "checked out " + tc.ref + " at " + short + "\n"; !strings.HasSuffix(got, line) {
			t.Errorf("--branch %s: output = %q, want it to end with %q", tc.ref, got, line)
		}
	}
	if b := git(t, clone, "rev-parse", "--abbrev-ref", "feature@{upstream}"); b != "origin/feature" {
		t.Errorf("feature tracks %q, want origin/feature", b)
	}
	if _, err := os.Stat(filepath.Join(clone, "untracked.txt")); err != nil {
		t.Errorf("an untracked file must survive the switches: %v", err)
	}
}

func TestSync_BranchRefusals(t *testing.T) {
	origin := branchOrigin(t)
	const rule = "kirokit sync: --branch: must be a branch or tag name, not empty, starting with '-' or containing whitespace"

	t.Run("bad names refuse before any git runs", func(t *testing.T) {
		chdir(t, t.TempDir())
		missing := filepath.Join(t.TempDir(), "no-checkout")
		for _, ref := range []string{"-b", "--upload-pack=x", "has space", ""} {
			if _, err := runRoot(t, "sync", "--repo", missing, "--branch", ref); err == nil || err.Error() != rule {
				t.Errorf("--branch %q: err = %v, want %q", ref, err, rule)
			}
		}
	})

	t.Run("with --latest-tag", func(t *testing.T) {
		chdir(t, t.TempDir())
		want := "kirokit sync: --branch and --latest-tag cannot be used together"
		if _, err := runRoot(t, "sync", "--repo", t.TempDir(), "--branch", "main", "--latest-tag"); err == nil || err.Error() != want {
			t.Errorf("err = %v, want %q", err, want)
		}
	})

	t.Run("unknown name", func(t *testing.T) {
		clone := cloneOf(t, origin)
		head := git(t, clone, "rev-parse", "HEAD")
		want := "kirokit sync: --branch nope: no branch or tag of that name on origin"
		if _, err := runRoot(t, "sync", "--repo", clone, "--branch", "nope"); err == nil || err.Error() != want {
			t.Errorf("err = %v, want %q", err, want)
		}
		if got := git(t, clone, "rev-parse", "HEAD"); got != head {
			t.Errorf("HEAD moved to %s", got)
		}
	})

	t.Run("local changes", func(t *testing.T) {
		clone := cloneOf(t, origin)
		head := git(t, clone, "rev-parse", "HEAD")
		edited := filepath.Join(clone, "f")
		if err := os.WriteFile(edited, []byte("edited"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := runRoot(t, "sync", "--repo", clone, "--branch", "feature")
		if err == nil || !strings.HasPrefix(err.Error(), "kirokit sync: local changes in ") ||
			!strings.HasSuffix(err.Error(), filepath.Base(clone)+": commit or discard them, then try again") {
			t.Errorf("err = %v, want the local-changes refusal naming the checkout", err)
		}
		if got := git(t, clone, "rev-parse", "HEAD"); got != head || headBranch(t, clone) != "main" {
			t.Errorf("HEAD moved to %s on %q", got, headBranch(t, clone))
		}
		if data, _ := os.ReadFile(edited); string(data) != "edited" {
			t.Errorf("the edit was lost: f = %q", data)
		}
	})

	t.Run("local commits", func(t *testing.T) {
		clone := cloneOf(t, origin)
		if got, err := runRoot(t, "sync", "--repo", clone, "--branch", "feature"); err != nil {
			t.Fatalf("first switch: %v\n%s", err, got)
		}
		commitFile(t, clone, "local work")
		git(t, clone, "checkout", "-q", "main")
		head := git(t, clone, "rev-parse", "HEAD")
		want := "kirokit sync: --branch feature: local branch feature has commits that are not on origin/feature"
		if _, err := runRoot(t, "sync", "--repo", clone, "--branch", "feature"); err == nil || err.Error() != want {
			t.Errorf("err = %v, want %q", err, want)
		}
		if got := git(t, clone, "rev-parse", "HEAD"); got != head || headBranch(t, clone) != "main" {
			t.Errorf("HEAD moved to %s on %q", got, headBranch(t, clone))
		}
	})
}

func TestSync_PlainSyncFollowsTheChosenBranchAndStaysOnATag(t *testing.T) {
	origin := branchOrigin(t)
	clone := cloneOf(t, origin)

	if got, err := runRoot(t, "sync", "--repo", clone, "--branch", "feature"); err != nil {
		t.Fatalf("switch: %v\n%s", err, got)
	}
	git(t, origin, "checkout", "-q", "feature")
	commitFile(t, origin, "feature-2")
	git(t, origin, "checkout", "-q", "main")
	got, err := runRoot(t, "sync", "--repo", clone)
	if err != nil {
		t.Fatalf("plain sync on a branch: %v\n%s", err, got)
	}
	if head, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "feature"); head != want {
		t.Errorf("plain sync: HEAD = %s, want origin's new feature tip %s", head, want)
	}
	short := git(t, clone, "rev-parse", "--short", "HEAD")
	if !strings.HasPrefix(got, "updated ") || !strings.HasSuffix(got, "checked out feature at "+short+"\n") {
		t.Errorf("output = %q, want the updated line, then checked out feature", got)
	}

	if got, err := runRoot(t, "sync", "--repo", clone, "--branch", "v1.0.0"); err != nil {
		t.Fatalf("switch to a tag: %v\n%s", err, got)
	}
	got, err = runRoot(t, "sync", "--repo", clone)
	if err != nil {
		t.Fatalf("plain sync on a tag must not fail: %v\n%s", err, got)
	}
	if head, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "v1.0.0^{commit}"); head != want {
		t.Errorf("plain sync on a tag: HEAD = %s, want it to stay on v1.0.0 %s", head, want)
	}
	if want := "checked out v1.0.0 at " + git(t, clone, "rev-parse", "--short", "HEAD") + "\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	// A detached HEAD on no tag names its commit.
	git(t, clone, "checkout", "-q", "--detach", "origin/feature~1")
	short = git(t, clone, "rev-parse", "--short", "HEAD")
	if got, err := runRoot(t, "sync", "--repo", clone); err != nil || got != "checked out "+short+" at "+short+"\n" {
		t.Errorf("plain sync on an untagged commit = %q, %v; want the commit as the name", got, err)
	}
}

// A pinned tag is reinstalled from, not skipped: a plain sync with a team on a
// detached HEAD must still install the team from that tag.
func TestSync_PlainSyncWithATeamOnATagReinstallsFromIt(t *testing.T) {
	origin := branchOrigin(t)
	clone := cloneOf(t, origin)
	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)

	if got, err := runRoot(t, "sync", "demo", "--repo", clone, "--branch", "v1.0.0"); err != nil {
		t.Fatalf("switch to a tag: %v\n%s", err, got)
	}
	commitFile(t, origin, "after the pin") // a new commit on main must not reach the pin
	got, err := runRoot(t, "sync", "demo", "--repo", clone)
	if err != nil {
		t.Fatalf("plain sync on a tag with a team: %v\n%s", err, got)
	}
	short := git(t, origin, "rev-parse", "--short", "v1.0.0^{commit}")
	want := "checked out v1.0.0 at " + short + "\nsynced demo into " + home + " at " + short + "\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if recs := installedRecords(t, home); !reflect.DeepEqual(recs, []string{"demo.json"}) {
		t.Errorf("records = %v, want demo.json", recs)
	}
}

// A branch name with a slash is a full name, not a path: the tracking branch
// keeps the whole name and the output names it whole.
func TestSync_BranchWithASlashKeepsItsWholeName(t *testing.T) {
	origin := branchOrigin(t)
	git(t, origin, "branch", "team/try-x", "feature")
	clone := cloneOf(t, origin)

	got, err := runRoot(t, "sync", "--repo", clone, "--branch", "team/try-x")
	if err != nil {
		t.Fatalf("sync --branch team/try-x: %v\n%s", err, got)
	}
	if b := headBranch(t, clone); b != "team/try-x" {
		t.Errorf("on branch %q, want team/try-x", b)
	}
	if u := git(t, clone, "rev-parse", "--abbrev-ref", "team/try-x@{upstream}"); u != "origin/team/try-x" {
		t.Errorf("team/try-x tracks %q, want origin/team/try-x", u)
	}
	short := git(t, origin, "rev-parse", "--short", "feature")
	if want := "checked out team/try-x at " + short + "\n"; !strings.HasSuffix(got, want) {
		t.Errorf("output = %q, want it to end with %q", got, want)
	}
}
