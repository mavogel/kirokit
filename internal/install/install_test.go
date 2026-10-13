package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

func TestInstall_GlobalSteeringIsInstalledPerTeamAndReferencedByTheAgent(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	const slug = "security-baseline"
	mustWrite(t, filepath.Join(roots.GlobalRoot, slug+".md"), "# Security Baseline\n")
	for _, tm := range []string{team, "other-team"} {
		mustWrite(t, filepath.Join(roots.TeamsRoot, tm, "manifest.yaml"), "team: "+tm+"\nglobalSteerings:\n  - "+slug+"\n")
	}
	mustWrite(t, filepath.Join(roots.TeamsRoot, "other-team", "agents", "agent2", "agent.json.tmpl"), `{"name": "agent2", "mcpServers": {"s": {"command": "c"}}}`)

	for _, tm := range []string{team, "other-team"} {
		if err := Install(roots, tm, kiroHome, nil); err != nil {
			t.Fatalf("Install(%s) returned error: %v", tm, err)
		}
	}

	steering := func(tm string) string { return filepath.Join(kiroHome, "kirokit-steerings", tm, slug+".md") }
	for _, tm := range []string{team, "other-team"} {
		if _, err := os.Stat(steering(tm)); err != nil {
			t.Fatalf("expected %s's installed global steering: %v", tm, err)
		}
	}
	agentJSON, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("read installed agent1.json: %v", err)
	}
	if !strings.Contains(string(agentJSON), "file://"+steering(team)) {
		t.Errorf("installed agent1.json = %s, want a resources entry for %s", agentJSON, steering(team))
	}

	if err := Uninstall(kiroHome, team); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if _, err := os.Stat(steering(team)); !os.IsNotExist(err) {
		t.Errorf("uninstalled team's steering still exists (err=%v)", err)
	}
	if _, err := os.Stat(steering("other-team")); err != nil {
		t.Errorf("uninstalling one team removed another team's steering: %v", err)
	}
}

func TestInstall_WithKnowledgeBaseKeepsExistingResourcesAndWritesTheExactEntry(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	mustWrite(t, filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "agent.json.tmpl"),
		`{"name": "agent1", "resources": ["file:///own.md"], "mcpServers": {"s": {"command": "c"}}}`)
	kbMount := filepath.Join(kiroHome, "kirokit-kb", team, "agent1")

	if err := Install(roots, team, kiroHome, &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: kbMount}); err != nil {
		t.Fatalf("Install with KnowledgeBase returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("read installed agent1.json: %v", err)
	}
	var agent struct {
		Resources []json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("installed agent1.json is not valid JSON: %v\n%s", err, data)
	}
	if len(agent.Resources) != 2 {
		t.Fatalf("resources = %s, want the agent's own entry followed by the knowledge base", data)
	}
	var own string
	if err := json.Unmarshal(agent.Resources[0], &own); err != nil || own != "file:///own.md" {
		t.Errorf("resources[0] = %s, want the agent's own %q entry preserved first", agent.Resources[0], "file:///own.md")
	}
	var kb map[string]any
	if err := json.Unmarshal(agent.Resources[1], &kb); err != nil {
		t.Fatalf("resources[1] = %s, want a knowledgeBase object: %v", agent.Resources[1], err)
	}
	want := map[string]any{
		"type":       "knowledgeBase",
		"source":     "file://" + kbMount,
		"name":       "agent1-kb",
		"indexType":  "best",
		"include":    []any{"**/*.md"},
		"autoUpdate": true,
	}
	if !reflect.DeepEqual(kb, want) {
		t.Errorf("knowledge base entry = %v, want exactly %v", kb, want)
	}
}

func TestAttachKnowledgeBase_RejectsUnusableInputs(t *testing.T) {
	setup := func(t *testing.T) (agentJSON, mount, source string) {
		t.Helper()
		dir := t.TempDir()
		agentJSON = filepath.Join(dir, "agent1.json")
		mustWrite(t, agentJSON, `{"name": "agent1"}`)
		source = filepath.Join(dir, "kb")
		mustWrite(t, filepath.Join(source, "doc.md"), "# Docs\n")
		return agentJSON, filepath.Join(dir, "mount"), source
	}

	t.Run("source folder does not exist", func(t *testing.T) {
		agentJSON, mount, source := setup(t)
		err := AttachKnowledgeBase(agentJSON, mount, filepath.Join(source, "missing"))
		if err == nil || !strings.Contains(err.Error(), "knowledge base source") {
			t.Fatalf("err = %v, want it to name the unusable knowledge base source", err)
		}
	})

	t.Run("source is a file, not a folder", func(t *testing.T) {
		agentJSON, mount, source := setup(t)
		err := AttachKnowledgeBase(agentJSON, mount, filepath.Join(source, "doc.md"))
		if err == nil || !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("err = %v, want a %q error", err, "not a directory")
		}
	})

	t.Run("mount point is a real file and is left untouched", func(t *testing.T) {
		agentJSON, mount, source := setup(t)
		mustWrite(t, mount, "precious\n")
		err := AttachKnowledgeBase(agentJSON, mount, source)
		if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
			t.Fatalf("err = %v, want a %q error", err, "refusing to overwrite")
		}
		if got, _ := os.ReadFile(mount); string(got) != "precious\n" {
			t.Errorf("mount point content = %q, want it untouched", got)
		}
	})

	t.Run("agent config does not exist", func(t *testing.T) {
		agentJSON, mount, source := setup(t)
		missing := filepath.Join(filepath.Dir(agentJSON), "nope.json")
		err := AttachKnowledgeBase(missing, mount, source)
		if err == nil || !strings.Contains(err.Error(), "nope.json") {
			t.Fatalf("err = %v, want it to name the missing agent config", err)
		}
	})
}

func TestAttachKnowledgeBase_ReplacesAStaleSymlinkAtTheMountPoint(t *testing.T) {
	dir := t.TempDir()
	agentJSON := filepath.Join(dir, "agent1.json")
	mustWrite(t, agentJSON, `{"name": "agent1"}`)
	oldSource, newSource := filepath.Join(dir, "old"), filepath.Join(dir, "new")
	mustWrite(t, filepath.Join(oldSource, "a.md"), "old\n")
	mustWrite(t, filepath.Join(newSource, "b.md"), "new\n")
	mount := filepath.Join(dir, "mount")
	if err := os.Symlink(oldSource, mount); err != nil {
		t.Fatalf("seed stale symlink: %v", err)
	}

	if err := AttachKnowledgeBase(agentJSON, mount, newSource); err != nil {
		t.Fatalf("AttachKnowledgeBase returned error: %v", err)
	}

	target, err := os.Readlink(mount)
	if err != nil || target != newSource {
		t.Errorf("mount point -> %q (err=%v), want it repointed to %q", target, err, newSource)
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
