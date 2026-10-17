package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// thisRepo returns the kirokit repository root, found from the package dir.
func thisRepo(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", ".."))
}

func TestKirokitRepo_RenderAllTeamsFromAnUnrelatedFolder(t *testing.T) {
	repo := thisRepo(t)
	chdir(t, t.TempDir())
	t.Setenv("KIROKIT_REPO", repo)
	out := t.TempDir()

	if _, err := runRoot(t, "render", "--out", out); err != nil {
		t.Fatalf("render from an unrelated folder: %v", err)
	}
	for _, team := range []string{"platform-team", "data-team"} {
		if _, err := os.Stat(filepath.Join(out, team)); err != nil {
			t.Errorf("team %s was not rendered: %v", team, err)
		}
	}

	single := t.TempDir()
	got, err := runRoot(t, "render", "platform-team", "--out", single)
	if err != nil {
		t.Fatalf("render platform-team: %v", err)
	}
	if want := "rendered platform-team -> " + single; !strings.Contains(got, want) {
		t.Errorf("output = %q, want it to contain %q", got, want)
	}
}

func TestKirokitRepo_RelativeValueIsMadeAbsoluteOnce(t *testing.T) {
	repo := thisRepo(t)
	chdir(t, filepath.Dir(repo))
	t.Setenv("KIROKIT_REPO", filepath.Base(repo))

	if _, err := runRoot(t, "render", "platform-team", "--out", t.TempDir()); err != nil {
		t.Fatalf("render with a relative KIROKIT_REPO: %v", err)
	}
	if !filepath.IsAbs(repoBase) {
		t.Errorf("repoBase = %q, want an absolute path", repoBase)
	}
}

func TestKirokitRepo_EmptyValueChangesNothing(t *testing.T) {
	chdir(t, thisRepo(t))
	t.Setenv("KIROKIT_REPO", "")

	if _, err := runRoot(t, "render", "platform-team", "--out", t.TempDir()); err != nil {
		t.Fatalf("render with an empty KIROKIT_REPO from the repository root: %v", err)
	}
	if repoBase != "" {
		t.Errorf("repoBase = %q, want empty", repoBase)
	}
	if got := repoPath(teamsRoot); got != teamsRoot {
		t.Errorf("repoPath(%q) = %q, want it unchanged", teamsRoot, got)
	}
}

func TestKirokitRepo_OnboardCreatesTheTeamUnderTheRepository(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	elsewhere := t.TempDir()
	chdir(t, elsewhere)
	t.Setenv("KIROKIT_REPO", repo)

	if _, err := runRoot(t, "onboard", "newteam"); err != nil {
		t.Fatalf("onboard: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "teams", "newteam", "manifest.yaml")); err != nil {
		t.Errorf("team not created under KIROKIT_REPO: %v", err)
	}
	if _, err := os.Stat(filepath.Join(elsewhere, "teams")); err == nil {
		t.Error("onboard wrote teams/ in the working directory")
	}
}

func TestKirokitRepo_SkillsFetchReadsTheManifestUnderTheRepository(t *testing.T) {
	repo := t.TempDir()
	writeTree(t, repo, map[string]string{
		"teams/plain/manifest.yaml": "team: plain\nglobalSteerings: []\n",
	})
	chdir(t, t.TempDir())
	t.Setenv("KIROKIT_REPO", repo)

	got, err := runRoot(t, "skills", "fetch", "plain")
	if err != nil {
		t.Fatalf("skills fetch: %v", err)
	}
	if want := "plain: no external skills declared"; !strings.Contains(got, want) {
		t.Errorf("output = %q, want it to contain %q", got, want)
	}
}

func TestKirokitRepo_TestAndUpdateUseTheRepositoryGoldens(t *testing.T) {
	repo := thisRepo(t)
	chdir(t, t.TempDir())
	t.Setenv("KIROKIT_REPO", repo)

	got, err := runRoot(t, "test", "platform-team")
	if err != nil {
		t.Fatalf("test from an unrelated folder: %v\n%s", err, got)
	}
	if want := "platform-team: ok"; !strings.Contains(got, want) {
		t.Errorf("output = %q, want it to contain %q", got, want)
	}
}

func TestKirokitRepo_SyncWithExplicitRepoIgnoresKirokitRepo(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	writeTree(t, origin, map[string]string{
		"env/mcp-versions.yaml":                "x: \"1\"\n",
		"global/steerings/.keep":               "",
		"teams/demo/manifest.yaml":             "team: demo\nglobalSteerings: []\n",
		"teams/demo/agents/a1/agent.json.tmpl": `{"name": "a1", "description": "d", "tools": [], "mcpServers": {}}`,
	})
	git(t, origin, "add", "-A")
	git(t, origin, "commit", "-q", "-m", "teams")
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")

	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)
	t.Setenv("KIROKIT_REPO", t.TempDir()) // empty repository: must lose to --repo
	chdir(t, t.TempDir())

	got, err := runRoot(t, "sync", "demo", "--repo", clone)
	if err != nil {
		t.Fatalf("sync --repo: %v\n%s", err, got)
	}
	if want := "synced demo into " + home; !strings.Contains(got, want) {
		t.Errorf("output = %q, want it to contain %q", got, want)
	}
}
