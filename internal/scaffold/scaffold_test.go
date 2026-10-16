package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffold_CreatesTeamFromStarter(t *testing.T) {
	destTeamsDir := t.TempDir()

	if err := Scaffold(destTeamsDir, "newteam"); err != nil {
		t.Fatalf("Scaffold returned error: %v", err)
	}

	manifest, err := os.ReadFile(filepath.Join(destTeamsDir, "newteam", "manifest.yaml"))
	if err != nil {
		t.Fatalf("expected manifest.yaml: %v", err)
	}
	if got := string(manifest); !containsSub(got, "team: newteam") {
		t.Errorf("manifest.yaml = %q, want it to contain %q", got, "team: newteam")
	}

	for _, p := range []string{
		filepath.Join("agents", "example-agent", "agent.json.tmpl"),
		filepath.Join("agents", "example-agent", "env", "mcp.env.tmpl"),
	} {
		if _, err := os.Stat(filepath.Join(destTeamsDir, "newteam", p)); err != nil {
			t.Errorf("expected scaffolded file %s: %v", p, err)
		}
	}

	for _, dir := range []string{"hooks", "prompts", "skills", "steerings"} {
		info, err := os.Stat(filepath.Join(destTeamsDir, "newteam", dir))
		if err != nil {
			t.Errorf("expected scaffolded directory %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestScaffold_RefusesToOverwriteExistingTeam(t *testing.T) {
	destTeamsDir := t.TempDir()

	if err := Scaffold(destTeamsDir, "newteam"); err != nil {
		t.Fatalf("first Scaffold returned error: %v", err)
	}
	manifestPath := filepath.Join(destTeamsDir, "newteam", "manifest.yaml")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest after first scaffold: %v", err)
	}

	err = Scaffold(destTeamsDir, "newteam")
	if err == nil {
		t.Fatal("second Scaffold returned no error for an already-existing team")
	}

	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest after second scaffold attempt: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("manifest.yaml changed after a rejected overwrite: before=%q after=%q", before, after)
	}
}

func containsSub(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
