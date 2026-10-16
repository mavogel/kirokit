package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupFixture builds a minimal team tree (one agent, one hook, one
// steering opting into a global steering, one skill) plus the shared
// global/env inputs, and returns the roots RenderTeam needs.
func setupFixture(t *testing.T) (teamsRoot, globalRoot, envVersionsPath string) {
	t.Helper()
	root := t.TempDir()

	teamsRoot = filepath.Join(root, "teams")
	globalRoot = filepath.Join(root, "global", "steerings")
	envVersionsPath = filepath.Join(root, "env", "mcp-versions.yaml")

	mustWrite := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	mustWrite(filepath.Join(globalRoot, "security-baseline.md"), "# Security Baseline\n\nFollow least privilege.\n")
	mustWrite(envVersionsPath, "test-mcp: \"1.2.3\"\n")

	teamDir := filepath.Join(teamsRoot, "fixture-team")
	mustWrite(filepath.Join(teamDir, "manifest.yaml"), "team: fixture-team\nglobalSteerings:\n  - security-baseline\n")
	mustWrite(filepath.Join(teamDir, "agents", "agent1", "agent.json.tmpl"),
		`{"name": "agent1", "mcpServers": {"test-mcp": {"command": "test-mcp-server", "args": ["--version", {{ mcpVersion "test-mcp" | jsonString }}]}}}`)
	mustWrite(filepath.Join(teamDir, "hooks", "lint.json.tmpl"),
		`{"version": "v1", "hooks": [{"name": "Lint", "trigger": "FileSave", "action": {"type": "command", "command": "gofmt -l ."}}]}`)
	mustWrite(filepath.Join(teamDir, "steerings", "conventions.md.tmpl"),
		"# Team Conventions\n\n{{ globalSteering \"security-baseline\" }}\n")
	mustWrite(filepath.Join(teamDir, "skills", "skill1", "SKILL.md.tmpl"),
		"---\nname: skill1\ndescription: A fixture skill.\n---\n\nDo the thing.\n")

	return teamsRoot, globalRoot, envVersionsPath
}

func TestRenderTeam_ProducesFlatOutputTree(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	outDir := filepath.Join(t.TempDir(), "out")

	if err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, outDir, ""); err != nil {
		t.Fatalf("RenderTeam returned error: %v", err)
	}

	agentJSON, err := os.ReadFile(filepath.Join(outDir, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("expected agents/agent1.json: %v", err)
	}
	got := string(agentJSON)
	if !contains(got, `"1.2.3"`) {
		t.Errorf("rendered agent1.json = %s, want it to contain the resolved mcpVersion \"1.2.3\"", got)
	}
	if contains(got, "{{") {
		t.Errorf("rendered agent1.json still contains an unresolved template action: %s", got)
	}

	if _, err := os.ReadFile(filepath.Join(outDir, "hooks", "lint.json")); err != nil {
		t.Fatalf("expected hooks/lint.json: %v", err)
	}

	steering, err := os.ReadFile(filepath.Join(outDir, "steering", "conventions.md"))
	if err != nil {
		t.Fatalf("expected steering/conventions.md: %v", err)
	}
	if !contains(string(steering), "Follow least privilege") {
		t.Errorf("rendered steering/conventions.md = %s, want the global steering content resolved in", string(steering))
	}

	if _, err := os.ReadFile(filepath.Join(outDir, "skills", "skill1", "SKILL.md")); err != nil {
		t.Fatalf("expected skills/skill1/SKILL.md: %v", err)
	}

	// No .tmpl suffix should remain anywhere in the output tree.
	_ = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && contains(path, ".tmpl") {
			t.Errorf("output tree still contains a .tmpl file: %s", path)
		}
		return nil
	})
}

func TestRenderTeam_UndefinedMcpVersionFailsWithClearError(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	// Reference a version pin that doesn't exist in env/mcp-versions.yaml.
	badAgent := filepath.Join(teamsRoot, "fixture-team", "agents", "agent1", "agent.json.tmpl")
	if err := os.WriteFile(badAgent, []byte(`{"mcpServers": {"x": {"command": "c", "args": [{{ mcpVersion "does-not-exist" | jsonString }}]}}}`), 0o600); err != nil {
		t.Fatalf("overwrite fixture: %v", err)
	}
	outDir := filepath.Join(t.TempDir(), "out")

	err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, outDir, "")
	if err == nil {
		t.Fatal("RenderTeam returned no error for an undefined mcpVersion key")
	}
	if !contains(err.Error(), "does-not-exist") {
		t.Errorf("error = %v, want it to name the missing key %q", err, "does-not-exist")
	}
	if !contains(err.Error(), "agent.json.tmpl") {
		t.Errorf("error = %v, want it to name the offending file", err)
	}
}

func TestRenderTeam_MalformedAgentJSONFailsValidation(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	badAgent := filepath.Join(teamsRoot, "fixture-team", "agents", "agent1", "agent.json.tmpl")
	// mcpServers entry with neither command nor url -> Task 2 validator rejects it.
	if err := os.WriteFile(badAgent, []byte(`{"mcpServers": {"x": {"timeout": 1}}}`), 0o600); err != nil {
		t.Fatalf("overwrite fixture: %v", err)
	}
	outDir := filepath.Join(t.TempDir(), "out")

	if err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, outDir, ""); err == nil {
		t.Fatal("RenderTeam returned no error for a rendered agent.json that fails validation")
	}
}

func TestRenderTeam_SopsArgsJSONProducesValidJSONArgsArray(t *testing.T) {
	teamsRoot, globalRoot, envVersionsPath := setupFixture(t)
	agentPath := filepath.Join(teamsRoot, "fixture-team", "agents", "agent1", "agent.json.tmpl")
	content := `{"mcpServers": {"test-mcp": {"command": {{ sopsCommand }}, "args": {{ sopsArgsJSON "agents/agent1/env/secrets.enc.yaml" "test-mcp-server" "--port" "8080" }}}}}`
	if err := os.WriteFile(agentPath, []byte(content), 0o600); err != nil {
		t.Fatalf("overwrite fixture: %v", err)
	}
	outDir := filepath.Join(t.TempDir(), "out")

	if err := RenderTeam(teamsRoot, "fixture-team", globalRoot, envVersionsPath, outDir, ""); err != nil {
		t.Fatalf("RenderTeam returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("expected agents/agent1.json: %v", err)
	}

	var decoded struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("rendered output is not valid JSON: %v\ncontent: %s", err, string(data))
	}
	server := decoded.MCPServers["test-mcp"]
	if server.Command != "sops" {
		t.Errorf("command = %q, want %q", server.Command, "sops")
	}
	want := []string{"exec-env", "agents/agent1/env/secrets.enc.yaml", "test-mcp-server --port 8080"}
	if len(server.Args) != len(want) {
		t.Fatalf("args = %v, want %v", server.Args, want)
	}
	for i := range want {
		if server.Args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, server.Args[i], want[i])
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
