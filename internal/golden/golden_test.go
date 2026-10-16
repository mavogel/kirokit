package golden

import (
	"os"
	"path/filepath"
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
