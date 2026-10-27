package golden

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func setupFixture(t *testing.T) Inputs {
	t.Helper()
	root := t.TempDir()

	in := Inputs{
		TeamsRoot:       filepath.Join(root, "teams"),
		GlobalRoot:      filepath.Join(root, "global", "steerings"),
		EnvVersionsPath: filepath.Join(root, "env", "mcp-versions.yaml"),
		GoldenRoot:      filepath.Join(root, "testdata", "golden"),
	}

	mustWrite(t, in.EnvVersionsPath, "test-mcp: \"1.0.0\"\n")
	mustWrite(t, filepath.Join(in.TeamsRoot, "fixture-team", "manifest.yaml"), "team: fixture-team\nglobalSteerings: []\n")
	mustWrite(t, filepath.Join(in.TeamsRoot, "fixture-team", "hooks", "lint.json.tmpl"),
		`{"version": "v1", "hooks": [{"name": "Lint", "trigger": "FileSave", "action": {"type": "command", "command": "gofmt -l ."}}]}`)

	return in
}

const fixtureQuickID = "0123456789abcdef0123456789abcdef"

// withQuick gives the fixture team a skill and an agent listed for Quick, and
// the golden root the committed-fixture ids file.
func withQuick(t *testing.T, in Inputs) {
	t.Helper()
	team := filepath.Join(in.TeamsRoot, "fixture-team")
	mustWrite(t, filepath.Join(team, "manifest.yaml"),
		"team: fixture-team\nglobalSteerings: []\nquick:\n  skills: [story]\n  agents:\n    - name: po\n      skills: [story]\n")
	mustWrite(t, filepath.Join(team, "skills", "story", "SKILL.md.tmpl"), "---\nname: story\ndescription: Stories.\n---\nBody.\n")
	mustWrite(t, filepath.Join(team, "agents", "po", "agent.json.tmpl"), `{"name": "po", "description": "d", "prompt": "p", "tools": []}`)
	mustWrite(t, filepath.Join(in.GoldenRoot, "quick-ids.yaml"), "skills:\n  story: "+fixtureQuickID+"\n")
}

func TestUpdate_WritesTheQuickBundleWithFixtureIDs(t *testing.T) {
	in := setupFixture(t)
	withQuick(t, in)
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	skill, err := os.ReadFile(filepath.Join(in.GoldenRoot, "fixture-team", "quick", "skills", "story", "SKILL.md"))
	if err != nil {
		t.Fatalf("quick golden skill: %v", err)
	}
	if want := "---\nname: story\ndescription: Stories.\nid: " + fixtureQuickID + "\n---\nBody.\n"; string(skill) != want {
		t.Errorf("quick golden skill = %q, want %q", skill, want)
	}
	if _, err := os.Stat(filepath.Join(in.GoldenRoot, "fixture-team", "quick", "agents", "po.json")); err != nil {
		t.Errorf("quick golden agent: %v", err)
	}
	if diffs, err := Compare(in, "fixture-team"); err != nil || len(diffs) != 0 {
		t.Errorf("Compare after Update = %v, %v; want no diffs", diffs, err)
	}
}

func TestCompare_DetectsQuickDrift(t *testing.T) {
	in := setupFixture(t)
	withQuick(t, in)
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	mustWrite(t, filepath.Join(in.TeamsRoot, "fixture-team", "skills", "story", "SKILL.md.tmpl"), "---\nname: story\ndescription: Stories.\n---\nChanged.\n")
	diffs, err := Compare(in, "fixture-team")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(diffs, "content differs: quick/skills/story/SKILL.md") {
		t.Errorf("diffs = %v, want the Quick skill named", diffs)
	}
}

func TestCompare_QuickFolderWithoutAQuickSectionIsADiff(t *testing.T) {
	in := setupFixture(t)
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(in.GoldenRoot, "fixture-team", "quick")); err == nil {
		t.Fatal("Update wrote quick/ for a team without a quick section")
	}
	mustWrite(t, filepath.Join(in.GoldenRoot, "fixture-team", "quick", "agents", "po.json"), "{}\n")
	diffs, err := Compare(in, "fixture-team")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(diffs, "missing file (present in golden, not rendered): quick/agents/po.json") {
		t.Errorf("diffs = %v, want the stray Quick golden named", diffs)
	}

	// Update removes the stray folder again.
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(in.GoldenRoot, "fixture-team", "quick")); err == nil {
		t.Error("Update kept a quick/ folder for a team without a quick section")
	}
}

func TestCompare_ReportsNoDiffsAfterUpdate(t *testing.T) {
	in := setupFixture(t)

	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	diffs, err := Compare(in, "fixture-team")
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("Compare returned diffs on freshly updated goldens: %v", diffs)
	}
}

func TestCompare_DetectsChangedTemplate(t *testing.T) {
	in := setupFixture(t)
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Mutate the template after goldens were generated.
	mustWrite(t, filepath.Join(in.TeamsRoot, "fixture-team", "hooks", "lint.json.tmpl"),
		`{"version": "v1", "hooks": [{"name": "Lint changed", "trigger": "FileSave", "action": {"type": "command", "command": "gofmt -l ."}}]}`)

	diffs, err := Compare(in, "fixture-team")
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("Compare returned no diffs after a template was mutated")
	}
}

func TestUpdate_RegeneratesGoldensSoCompareThenPasses(t *testing.T) {
	in := setupFixture(t)
	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	mustWrite(t, filepath.Join(in.TeamsRoot, "fixture-team", "hooks", "lint.json.tmpl"),
		`{"version": "v1", "hooks": [{"name": "Lint changed", "trigger": "FileSave", "action": {"type": "command", "command": "gofmt -l ."}}]}`)

	if err := Update(in, "fixture-team"); err != nil {
		t.Fatalf("second Update returned error: %v", err)
	}

	diffs, err := Compare(in, "fixture-team")
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("Compare returned diffs right after Update regenerated goldens: %v", diffs)
	}
}
