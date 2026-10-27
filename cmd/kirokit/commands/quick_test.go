package commands

import (
	"os"
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

func TestRender_QuickIDsNeedsQuick(t *testing.T) {
	quickRepo(t)
	if _, err := runRoot(t, "render", "q", "--quick-ids", "x.yaml", "--out", t.TempDir()); err == nil || !strings.Contains(err.Error(), "--quick-ids needs --quick") {
		t.Errorf("err = %v", err)
	}
}
