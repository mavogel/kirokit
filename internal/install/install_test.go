package install

import (
	"os"
	"path/filepath"
	"strings"
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

func setupFixture(t *testing.T) (roots Roots, team, agentSecretsSrc string) {
	t.Helper()
	root := t.TempDir()

	roots = Roots{
		TeamsRoot:       filepath.Join(root, "teams"),
		GlobalRoot:      filepath.Join(root, "global", "steerings"),
		EnvVersionsPath: filepath.Join(root, "env", "mcp-versions.yaml"),
	}
	team = "fixture-team"

	mustWrite(t, roots.EnvVersionsPath, "test-mcp: \"1.0.0\"\n")
	mustWrite(t, filepath.Join(roots.TeamsRoot, team, "manifest.yaml"), "team: fixture-team\nglobalSteerings: []\n")
	mustWrite(t, filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "agent.json.tmpl"),
		`{"mcpServers": {"test-mcp": {"command": {{ sopsCommand }}, "args": {{ sopsArgsJSON (printf "%s/kirokit-secrets/fixture-team/agent1/secrets.enc.yaml" .KiroHome) "test-mcp-server" "--port" "8080" }}}}}`)

	agentSecretsSrc = filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "env", "secrets.enc.yaml")
	mustWrite(t, agentSecretsSrc, "sops-ciphertext-placeholder\n")

	return roots, team, agentSecretsSrc
}

func TestInstall_WritesSecretsFileAtExpectedPathAndBakesItIntoAgentArgs(t *testing.T) {
	roots, team, agentSecretsSrc := setupFixture(t)
	kiroHome := t.TempDir()

	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	wantSecretsDst := filepath.Join(kiroHome, "kirokit-secrets", team, "agent1", "secrets.enc.yaml")
	gotSecrets, err := os.ReadFile(wantSecretsDst)
	if err != nil {
		t.Fatalf("expected copied secrets file at %s: %v", wantSecretsDst, err)
	}
	wantSecrets, err := os.ReadFile(agentSecretsSrc)
	if err != nil {
		t.Fatalf("read source secrets file: %v", err)
	}
	if string(gotSecrets) != string(wantSecrets) {
		t.Errorf("copied secrets content = %q, want %q", gotSecrets, wantSecrets)
	}

	agentJSON, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("expected installed agents/agent1.json: %v", err)
	}
	if !strings.Contains(string(agentJSON), wantSecretsDst) {
		t.Errorf("installed agent1.json = %s, want it to reference the installed secrets path %s", agentJSON, wantSecretsDst)
	}
}

func TestUninstall_RemovesExactlyWhatInstallWroteAndNothingElse(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()

	unrelated := filepath.Join(kiroHome, "agents", "unrelated.json")
	mustWrite(t, unrelated, `{"name": "unrelated"}`)

	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	installedAgent := filepath.Join(kiroHome, "agents", "agent1.json")
	if _, err := os.Stat(installedAgent); err != nil {
		t.Fatalf("expected installed agent1.json before uninstall: %v", err)
	}

	if err := Uninstall(kiroHome, team); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	if _, err := os.Stat(installedAgent); !os.IsNotExist(err) {
		t.Errorf("installed agent1.json still exists after uninstall (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(kiroHome, "kirokit-secrets", team, "agent1", "secrets.enc.yaml")); !os.IsNotExist(err) {
		t.Errorf("installed secrets file still exists after uninstall (err=%v)", err)
	}

	if _, err := os.Stat(unrelated); err != nil {
		t.Errorf("uninstall removed an unrelated pre-existing file: %v", err)
	}
}

func TestInstall_WithKnowledgeBaseCreatesSymlinkAndResourceEntry(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	mustWrite(t, filepath.Join(kbSource, "doc.md"), "# Docs\n")
	kbMount := filepath.Join(kiroHome, "kirokit-kb", team, "agent1")

	kb := &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: kbMount}
	if err := Install(roots, team, kiroHome, kb); err != nil {
		t.Fatalf("Install with KnowledgeBase returned error: %v", err)
	}

	linkInfo, err := os.Lstat(kbMount)
	if err != nil {
		t.Fatalf("expected symlink at mount point %s: %v", kbMount, err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", kbMount)
	}
	target, err := os.Readlink(kbMount)
	if err != nil {
		t.Fatalf("readlink %s: %v", kbMount, err)
	}
	absSource, _ := filepath.Abs(kbSource)
	if target != absSource {
		t.Errorf("symlink target = %q, want %q", target, absSource)
	}

	agentJSON, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("read installed agent1.json: %v", err)
	}
	if !strings.Contains(string(agentJSON), `"type":"knowledgeBase"`) && !strings.Contains(string(agentJSON), `"type": "knowledgeBase"`) {
		t.Errorf("installed agent1.json = %s, want a knowledgeBase resource entry", agentJSON)
	}
	if !strings.Contains(string(agentJSON), kbMount) {
		t.Errorf("installed agent1.json = %s, want it to reference the mount point %s", agentJSON, kbMount)
	}
}

func TestUninstall_WithKnowledgeBaseRemovesSymlinkButNotRealSource(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	kbFile := filepath.Join(kbSource, "doc.md")
	mustWrite(t, kbFile, "# Docs\n")
	kbMount := filepath.Join(kiroHome, "kirokit-kb", team, "agent1")

	kb := &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: kbMount}
	if err := Install(roots, team, kiroHome, kb); err != nil {
		t.Fatalf("Install with KnowledgeBase returned error: %v", err)
	}

	if err := Uninstall(kiroHome, team); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	if _, err := os.Lstat(kbMount); !os.IsNotExist(err) {
		t.Errorf("symlink %s still exists after uninstall (err=%v)", kbMount, err)
	}
	if _, err := os.Stat(kbFile); err != nil {
		t.Errorf("uninstall removed the real knowledge base source content: %v", err)
	}
}
