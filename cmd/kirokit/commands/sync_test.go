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
	if err := updateCheckout(false); err != nil {
		t.Fatalf("branch update: %v", err)
	}
	if got, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "HEAD"); got != want {
		t.Errorf("branch update: HEAD = %s, want origin tip %s", got, want)
	}

	if err := updateCheckout(true); err != nil {
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
	if want := "uninstalled platform-team from " + home + "\n"; !strings.HasPrefix(got, want) {
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
		prepare func(t *testing.T, clone string)
	}{
		"unknown team": {team: "nonexistent-team", prepare: func(*testing.T, string) {}},
		"failing git pull": {team: "demo", prepare: func(t *testing.T, clone string) {
			// Diverge clone and origin so pull --ff-only fails.
			commitFile(t, clone, "local")
			commitFile(t, origin, "remote")
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

			got, err := runRoot(t, "sync", tc.team, "--repo", clone)
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

	if err := updateCheckout(true); err == nil || !strings.Contains(err.Error(), "no v* tag") {
		t.Fatalf("err = %v, want a %q error", err, "no v* tag")
	}
}
