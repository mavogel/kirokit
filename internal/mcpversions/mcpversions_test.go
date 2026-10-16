package mcpversions

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mcp-versions.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestLoad_ParsesFlatMap(t *testing.T) {
	path := writeFixture(t, "jira-mcp: \"1.4.2\"\nconfluence-mcp: \"2.0.0\"\n")

	v, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if v["jira-mcp"] != "1.4.2" {
		t.Errorf("jira-mcp = %q, want %q", v["jira-mcp"], "1.4.2")
	}
	if v["confluence-mcp"] != "2.0.0" {
		t.Errorf("confluence-mcp = %q, want %q", v["confluence-mcp"], "2.0.0")
	}
}

func TestLoad_NonScalarValueErrors(t *testing.T) {
	path := writeFixture(t, "jira-mcp:\n  nested: true\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load returned no error for a non-scalar value")
	}
}
