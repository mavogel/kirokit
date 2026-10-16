package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestLoad_ValidManifest(t *testing.T) {
	path := writeFixture(t, "team: platform-team\nglobalSteerings:\n  - security-baseline\n")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if m.Team != "platform-team" {
		t.Errorf("Team = %q, want %q", m.Team, "platform-team")
	}
	if len(m.GlobalSteerings) != 1 || m.GlobalSteerings[0] != "security-baseline" {
		t.Errorf("GlobalSteerings = %v, want [security-baseline]", m.GlobalSteerings)
	}
}

func TestLoad_EmptyGlobalSteeringsIsValidOptOut(t *testing.T) {
	path := writeFixture(t, "team: data-team\nglobalSteerings: []\n")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(m.GlobalSteerings) != 0 {
		t.Errorf("GlobalSteerings = %v, want empty", m.GlobalSteerings)
	}
}

func TestLoad_MissingTeamErrors(t *testing.T) {
	path := writeFixture(t, "globalSteerings: []\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load returned no error for missing team field")
	}
}
