package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	quickID      = "0123456789abcdef0123456789abcdef"
	quickIDOther = "fedcba9876543210fedcba9876543210"
)

// quickRepo is a config repository with one team that lists a skill and an
// agent for Quick ("q") and one that lists nothing ("plain").
func quickRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	writeTree(t, repo, map[string]string{
		"env/mcp-versions.yaml":                 "x: \"1\"\n",
		"global/steerings/.keep":                "",
		"teams/q/manifest.yaml":                 "team: q\nglobalSteerings: []\nquick:\n  skills: [story]\n  agents:\n    - name: po\n      skills: [story]\n",
		"teams/q/skills/story/SKILL.md.tmpl":    "---\nname: story\ndescription: Stories.\n---\nBody.\n",
		"teams/q/agents/po/agent.json.tmpl":     `{"name": "po", "description": "d", "prompt": "p", "tools": []}`,
		"teams/plain/manifest.yaml":             "team: plain\nglobalSteerings: []\n",
		"teams/plain/agents/a1/agent.json.tmpl": `{"name": "a1", "description": "d", "prompt": "p", "tools": []}`,
	})
	chdir(t, t.TempDir())
	t.Setenv("KIROKIT_REPO", repo)
	t.Setenv("KIROKIT_QUICK_IDS", "")
	return repo
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRenderQuick_WritesTheBundleAndWarnsAboutMissingIDs(t *testing.T) {
	quickRepo(t)
	out := filepath.Join(t.TempDir(), "bundle")

	got, err := runRoot(t, "render", "--quick", "q", "--out", out)
	if err != nil {
		t.Fatalf("render --quick: %v\n%s", err, got)
	}
	for _, want := range []string{
		"rendered Quick bundle for q: 1 skills, 1 agents -> " + out,
		"warning: no Quick id for skill story yet (first import); run kirokit quick ids after importing",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want it to contain %q", got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "skills", "story", "SKILL.md")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(out, "agents", "po.json")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(out, "skills", "story", "SKILL.md.tmpl")); err == nil {
		t.Error("the bundle holds a template source")
	}
}

func TestRenderQuick_TeamWithoutQuickContent(t *testing.T) {
	quickRepo(t)
	if got, err := runRoot(t, "render", "--quick", "plain", "--out", t.TempDir()); err == nil || !strings.Contains(err.Error(), "team plain lists nothing for Quick") {
		t.Errorf("err = %v, output %q; want the nothing-for-Quick error", err, got)
	}

	all := t.TempDir()
	got, err := runRoot(t, "render", "--quick", "--out", all)
	if err != nil {
		t.Fatalf("render --quick (all teams): %v\n%s", err, got)
	}
	if !strings.Contains(got, "plain: lists nothing for Quick, skipped") || !strings.Contains(got, "rendered Quick bundle for q") {
		t.Errorf("output = %q", got)
	}
	if _, err := os.Stat(filepath.Join(all, "plain")); err == nil {
		t.Error("a skipped team got a folder")
	}
	if _, err := os.Stat(filepath.Join(all, "q", "agents", "po.json")); err != nil {
		t.Error(err)
	}
}

// The ids file comes from --quick-ids, else KIROKIT_QUICK_IDS, else
// quick-ids.yaml at the repository root.
func TestRenderQuick_IDsFileResolution(t *testing.T) {
	repo := quickRepo(t)
	writeIDs := func(path, id string) {
		t.Helper()
		writeTree(t, filepath.Dir(path), map[string]string{filepath.Base(path): "skills:\n  story: " + id + "\n"})
	}
	flagFile := filepath.Join(t.TempDir(), "flag.yaml")
	envFile := filepath.Join(t.TempDir(), "env.yaml")
	writeIDs(filepath.Join(repo, "quick-ids.yaml"), quickID)
	writeIDs(flagFile, quickIDOther)
	writeIDs(envFile, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	render := func(args ...string) string {
		t.Helper()
		out := filepath.Join(t.TempDir(), "b")
		got, err := runRoot(t, append([]string{"render", "--quick", "q", "--out", out}, args...)...)
		if err != nil {
			t.Fatalf("render: %v\n%s", err, got)
		}
		if strings.Contains(got, "warning:") {
			t.Errorf("output = %q, want no warning", got)
		}
		return readFile(t, filepath.Join(out, "skills", "story", "SKILL.md"))
	}

	if got := render(); !strings.Contains(got, "id: "+quickID) {
		t.Errorf("repository default not used:\n%s", got)
	}
	t.Setenv("KIROKIT_QUICK_IDS", envFile)
	if got := render(); !strings.Contains(got, "id: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("KIROKIT_QUICK_IDS not used:\n%s", got)
	}
	if got := render("--quick-ids", flagFile); !strings.Contains(got, "id: "+quickIDOther) {
		t.Errorf("--quick-ids not used:\n%s", got)
	}
}

func TestRenderQuick_MalformedIDsFileFails(t *testing.T) {
	repo := quickRepo(t)
	writeTree(t, repo, map[string]string{"quick-ids.yaml": "skills:\n  story: nope\n"})
	if _, err := runRoot(t, "render", "--quick", "q", "--out", t.TempDir()); err == nil || !strings.Contains(err.Error(), "32 lowercase hex") {
		t.Errorf("err = %v, want the malformed id named", err)
	}
}

// A skill exported from Quick goes through quick ids into the repository's
// ids file, and the next render writes the id back unchanged.
func TestQuickIDs_RecordsExportsForTheNextRender(t *testing.T) {
	repo := quickRepo(t)
	first := filepath.Join(t.TempDir(), "first")
	if _, err := runRoot(t, "render", "--quick", "q", "--out", first); err != nil {
		t.Fatal(err)
	}
	rendered := readFile(t, filepath.Join(first, "skills", "story", "SKILL.md"))
	exportDir := t.TempDir()
	export := strings.Replace(rendered, "\n---\n", "\nid: "+quickID+"\n---\n", 1)
	writeTree(t, exportDir, map[string]string{"story/SKILL.md": export, "other/SKILL.md": "---\nname: other\ndescription: d\n---\n"})

	got, err := runRoot(t, "quick", "ids", exportDir)
	if err != nil {
		t.Fatalf("quick ids: %v\n%s", err, got)
	}
	idsFile := filepath.Join(repo, "quick-ids.yaml")
	if !strings.Contains(got, "recorded 1 ids in "+idsFile) || !strings.Contains(got, "skipped "+filepath.Join(exportDir, "other", "SKILL.md")+": no id") {
		t.Errorf("output = %q", got)
	}

	second := filepath.Join(t.TempDir(), "second")
	got, err = runRoot(t, "render", "--quick", "q", "--out", second)
	if err != nil || strings.Contains(got, "warning:") {
		t.Fatalf("second render: %v\n%s", err, got)
	}
	if again := readFile(t, filepath.Join(second, "skills", "story", "SKILL.md")); again != export {
		t.Errorf("second-pass SKILL.md =\n%s\nwant the export\n%s", again, export)
	}
	if agent := readFile(t, filepath.Join(second, "agents", "po.json")); !strings.Contains(agent, quickID) {
		t.Errorf("agent does not pin the id:\n%s", agent)
	}
}

func TestQuickIDs_MergesIntoTheFlagFileAndKeepsOtherNames(t *testing.T) {
	quickRepo(t)
	idsFile := filepath.Join(t.TempDir(), "ids.yaml")
	writeTree(t, filepath.Dir(idsFile), map[string]string{"ids.yaml": "skills:\n  kept: " + quickIDOther + "\n  story: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"})
	export := filepath.Join(t.TempDir(), "SKILL.md")
	writeTree(t, filepath.Dir(export), map[string]string{"SKILL.md": "---\nname: story\ndescription: d\nid: " + quickID + "\n---\n"})

	if _, err := runRoot(t, "quick", "ids", "--quick-ids", idsFile, export); err != nil {
		t.Fatal(err)
	}
	if got, want := readFile(t, idsFile), "skills:\n  kept: "+quickIDOther+"\n  story: "+quickID+"\n"; !strings.HasSuffix(got, want) {
		t.Errorf("ids file =\n%s\nwant it to end with\n%s", got, want)
	}
}

func TestQuickIDs_NeedsAPathAndWritesNothingOnError(t *testing.T) {
	repo := quickRepo(t)
	if _, err := runRoot(t, "quick", "ids"); err == nil {
		t.Error("quick ids without a path succeeded")
	}
	bad := filepath.Join(t.TempDir(), "SKILL.md")
	writeTree(t, filepath.Dir(bad), map[string]string{"SKILL.md": "---\nname: story\ndescription: d\nid: XYZ\n---\n"})
	if _, err := runRoot(t, "quick", "ids", bad); err == nil {
		t.Error("quick ids with a malformed id succeeded")
	}
	if _, err := os.Stat(filepath.Join(repo, "quick-ids.yaml")); err == nil {
		t.Error("a failed quick ids wrote the ids file")
	}
}

// The ids file is per account and must never be committed.
func TestQuickIDs_FileIsGitignored(t *testing.T) {
	repo := thisRepo(t)
	out, err := exec.Command("git", "-C", repo, "check-ignore", "-q", "quick-ids.yaml").CombinedOutput()
	if err != nil {
		t.Errorf("quick-ids.yaml is not ignored at the repository root: %v %s", err, out)
	}
}

func TestRender_QuickIDsNeedsQuick(t *testing.T) {
	quickRepo(t)
	if _, err := runRoot(t, "render", "q", "--quick-ids", "x.yaml", "--out", t.TempDir()); err == nil || !strings.Contains(err.Error(), "--quick-ids needs --quick") {
		t.Errorf("err = %v", err)
	}
}
