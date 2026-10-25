package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/install"
	"github.com/mavogel/kirokit/internal/testutil"
)

func runRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func writeTree(t *testing.T, base string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// skillSource is a git repository with skills/pdf (a SKILL.md and an
// executable script).
func skillSource(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	git(t, repo, "init", "-q", "-b", "main")
	writeTree(t, repo, map[string]string{
		"skills/pdf/SKILL.md":         "---\nname: pdf\ndescription: PDFs.\n---\n",
		"skills/pdf/scripts/run.sh":   "#!/bin/sh\necho run\n",
		"skills/pdf/reference/api.md": "api\n",
	})
	if err := os.Chmod(filepath.Join(repo, "skills/pdf/scripts/run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "pdf")
	return repo
}

// projectRoot builds a kit checkout with three teams and chdirs into it:
// demo (one external skill), plain (none) and broken (an unreachable source).
func projectRoot(t *testing.T, source string) {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"env/mcp-versions.yaml":                  "x: \"1\"\n",
		"teams/demo/manifest.yaml":               "team: demo\nglobalSteerings: []\nexternalSkills:\n  - source: " + source + "\n    ref: main\n    paths: [skills/pdf]\n",
		"teams/demo/agents/a1/agent.json.tmpl":   `{"name": "a1"}`,
		"teams/plain/manifest.yaml":              "team: plain\nglobalSteerings: []\n",
		"teams/plain/agents/a1/agent.json.tmpl":  `{"name": "a1"}`,
		"teams/broken/manifest.yaml":             "team: broken\nglobalSteerings: []\nexternalSkills:\n  - source: " + filepath.Join(source, "missing") + "\n    ref: main\n    paths: [skills/pdf]\n",
		"teams/broken/agents/a1/agent.json.tmpl": `{"name": "a1"}`,
	})
	if err := os.MkdirAll(filepath.Join(root, "global", "steerings"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)
}

func TestSkillsFetch_VendorsThenInstallsAndUninstallsWithoutTheSource(t *testing.T) {
	source := skillSource(t)
	projectRoot(t, source)

	out, err := runRoot(t, "skills", "fetch", "demo")
	if err != nil {
		t.Fatalf("skills fetch: %v\n%s", err, out)
	}
	commit := git(t, source, "rev-parse", "HEAD")
	if !strings.Contains(out, "pdf") || !strings.Contains(out, commit[:7]) {
		t.Errorf("output = %q, want the skill name and the resolved short commit %s", out, commit[:7])
	}
	if !strings.Contains(out, "kirokit update demo") {
		t.Errorf("output = %q, want a reminder to run kirokit update", out)
	}
	lock, err := os.ReadFile("teams/demo/external/skills.lock.yaml")
	if err != nil {
		t.Fatalf("expected the lock file: %v", err)
	}

	if _, err := runRoot(t, "skills", "fetch", "demo"); err != nil {
		t.Fatalf("second skills fetch: %v", err)
	}
	if again, _ := os.ReadFile("teams/demo/external/skills.lock.yaml"); string(again) != string(lock) {
		t.Error("a second fetch with no upstream change altered the lock")
	}

	// From here on the source must not be needed.
	if err := os.RemoveAll(source); err != nil {
		t.Fatal(err)
	}
	kiroHome := t.TempDir()
	roots := install.Roots{TeamsRoot: teamsRoot, GlobalRoot: globalRoot, EnvVersionsPath: envVersionsPath}
	if err := install.Install(roots, "demo", kiroHome, nil); err != nil {
		t.Fatalf("Install without the source repository: %v", err)
	}

	skillMD := filepath.Join(kiroHome, "skills", "pdf", "SKILL.md")
	script := filepath.Join(kiroHome, "skills", "pdf", "scripts", "run.sh")
	if _, err := os.Stat(skillMD); err != nil {
		t.Errorf("expected installed %s: %v", skillMD, err)
	}
	if info, err := os.Stat(script); err != nil || (testutil.ExecBitsKept && info.Mode().Perm()&0o100 == 0) {
		t.Errorf("installed script = %v (%v), want owner-executable", info, err)
	}
	agent, err := os.ReadFile(filepath.Join(kiroHome, "agents", "a1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "skill://" + skillMD; !strings.Contains(string(agent), want) {
		t.Errorf("installed agent = %s, want it to reference %s", agent, want)
	}

	if err := install.Uninstall(kiroHome, "demo"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	for _, p := range []string{skillMD, script} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s survived uninstall (err=%v)", p, err)
		}
	}
}

func TestSkillsFetch_TeamWithoutExternalSkillsSucceedsAndChangesNothing(t *testing.T) {
	projectRoot(t, skillSource(t))

	out, err := runRoot(t, "skills", "fetch", "plain")
	if err != nil {
		t.Fatalf("skills fetch plain: %v\n%s", err, out)
	}
	if !strings.Contains(out, "no external skills") {
		t.Errorf("output = %q, want a message that the team declares no external skills", out)
	}
	if _, err := os.Stat("teams/plain/external"); !os.IsNotExist(err) {
		t.Errorf("external/ created for a team with no external skills (err=%v)", err)
	}
}

func TestSkillsFetch_UnreachableSourceFailsAndLeavesTheTeamUntouched(t *testing.T) {
	projectRoot(t, skillSource(t))

	out, err := runRoot(t, "skills", "fetch", "broken")
	if err == nil {
		t.Fatalf("skills fetch succeeded, want an error\n%s", out)
	}
	if !strings.Contains(err.Error(), `"pdf"`) {
		t.Errorf("err = %v, want it to name the skill", err)
	}
	entries, _ := os.ReadDir("teams/broken")
	for _, e := range entries {
		if e.Name() == "external" {
			t.Error("teams/broken/external was left behind by a failed fetch")
		}
	}
}

func TestSkillsFetch_WithoutATeamFetchesEveryTeamThatDeclaresSkills(t *testing.T) {
	source := skillSource(t)
	projectRoot(t, source)
	// Drop the unreachable team so the all-teams run can succeed.
	if err := os.RemoveAll("teams/broken"); err != nil {
		t.Fatal(err)
	}

	out, err := runRoot(t, "skills", "fetch")
	if err != nil {
		t.Fatalf("skills fetch: %v\n%s", err, out)
	}
	if _, err := os.Stat("teams/demo/external/skills/pdf/SKILL.md"); err != nil {
		t.Errorf("demo was not fetched: %v", err)
	}
	if _, err := os.Stat("teams/plain/external"); !os.IsNotExist(err) {
		t.Errorf("plain gained an external/ directory (err=%v)", err)
	}
}

// collectionSource is a git repository with skills/{pdf,sql,docx}: three skills
// under one directory, plus a docs directory that is not a skill.
func collectionSource(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	git(t, repo, "init", "-q", "-b", "main")
	writeTree(t, repo, map[string]string{
		"skills/pdf/SKILL.md":       "pdf\n",
		"skills/pdf/scripts/run.sh": "#!/bin/sh\necho pdf\n",
		"skills/sql/SKILL.md":       "sql\n",
		"skills/docx/SKILL.md":      "docx\n",
		"skills/docs/guide.md":      "not a skill\n",
	})
	if err := os.Chmod(filepath.Join(repo, "skills/pdf/scripts/run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "skills")
	return repo
}

func TestSkillsFetch_AllSkillsOfARepositoryWithExcludeEndToEnd(t *testing.T) {
	source := collectionSource(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"env/mcp-versions.yaml":                "x: \"1\"\n",
		"teams/demo/agents/a1/agent.json.tmpl": `{"name": "a1"}`,
		"teams/demo/manifest.yaml": "team: demo\nglobalSteerings: []\nexternalSkills:\n  - source: " + source +
			"\n    ref: main\n    paths: [skills]\n    all: true\n    exclude: [docx]\n",
	})
	if err := os.MkdirAll(filepath.Join(root, "global", "steerings"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	out, err := runRoot(t, "skills", "fetch", "demo")
	if err != nil {
		t.Fatalf("skills fetch: %v\n%s", err, out)
	}
	for _, want := range []string{"pdf", "sql"} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want a line for %s", out, want)
		}
	}
	if strings.Contains(out, "docx") || strings.Contains(out, "docs") {
		t.Errorf("output = %q, want neither the excluded docx nor the non-skill docs", out)
	}

	if err := os.RemoveAll(source); err != nil {
		t.Fatal(err)
	}
	kiroHome := t.TempDir()
	roots := install.Roots{TeamsRoot: teamsRoot, GlobalRoot: globalRoot, EnvVersionsPath: envVersionsPath}
	if err := install.Install(roots, "demo", kiroHome, nil); err != nil {
		t.Fatalf("Install without the source repository: %v", err)
	}

	agent, err := os.ReadFile(filepath.Join(kiroHome, "agents", "a1.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"pdf", "sql"} {
		skillMD := filepath.Join(kiroHome, "skills", name, "SKILL.md")
		if _, err := os.Stat(skillMD); err != nil {
			t.Errorf("expected installed %s: %v", skillMD, err)
		}
		if !strings.Contains(string(agent), "skill://"+skillMD) {
			t.Errorf("installed agent = %s, want it to reference %s", agent, skillMD)
		}
	}
	if info, err := os.Stat(filepath.Join(kiroHome, "skills", "pdf", "scripts", "run.sh")); err != nil || (testutil.ExecBitsKept && info.Mode().Perm()&0o100 == 0) {
		t.Errorf("installed script = %v (%v), want owner-executable", info, err)
	}
	for _, name := range []string{"docx", "docs"} {
		if _, err := os.Stat(filepath.Join(kiroHome, "skills", name)); err == nil {
			t.Errorf("skills/%s was installed; it is excluded or not a skill", name)
		}
	}

	if err := install.Uninstall(kiroHome, "demo"); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	for _, name := range []string{"pdf", "sql"} {
		if _, err := os.Stat(filepath.Join(kiroHome, "skills", name, "SKILL.md")); !os.IsNotExist(err) {
			t.Errorf("skills/%s survived uninstall (err=%v)", name, err)
		}
	}
}
