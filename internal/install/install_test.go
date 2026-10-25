package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/dirlink"
	"github.com/mavogel/kirokit/internal/sopswrap"
	"github.com/mavogel/kirokit/internal/testutil"
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
	// Rendered paths use forward slashes on every OS, which sops, Kiro and
	// file:// resources accept on Windows too.
	if !strings.Contains(string(agentJSON), `"`+filepath.ToSlash(wantSecretsDst)+`"`) {
		t.Errorf("installed agent1.json = %s, want it to reference the installed secrets path %s", agentJSON, filepath.ToSlash(wantSecretsDst))
	}
	if strings.Contains(string(agentJSON), `\\`) {
		t.Errorf("installed agent1.json = %s, want no backslash in its paths", agentJSON)
	}
}

func TestLaunchShellFor_CmdOnlyOnWindows(t *testing.T) {
	for goos, want := range map[string]sopswrap.Shell{"windows": sopswrap.Cmd, "linux": sopswrap.POSIX, "darwin": sopswrap.POSIX} {
		if got := launchShellFor(goos); got != want {
			t.Errorf("launchShellFor(%q) = %v, want %v", goos, got, want)
		}
	}
}

func TestInstall_WindowsRefusesAnArgumentCmdExeCannotPassAndWritesNothing(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("cmd.exe launches are written only on Windows")
	}
	roots, team, _ := setupFixture(t)
	mustWrite(t, filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "agent.json.tmpl"),
		`{"mcpServers": {"test-mcp": {"command": {{ sopsCommand }}, "args": {{ sopsArgsJSON (printf "%s/kirokit-secrets/fixture-team/agent1/secrets.enc.yaml" .KiroHome) "test-mcp-server" "--header" "Authorization: x" }}}}}`)
	kiroHome := filepath.Join(t.TempDir(), "kiro")

	err := Install(roots, team, kiroHome, nil)
	if !errors.Is(err, sopswrap.ErrUnsafeForCmd) || !strings.Contains(err.Error(), "fixture-team/agent1") {
		t.Fatalf("Install error = %v, want a cmd.exe refusal naming fixture-team/agent1", err)
	}
	if _, statErr := os.Stat(kiroHome); !os.IsNotExist(statErr) {
		t.Errorf("KIRO_HOME was written after a refused install: %v", statErr)
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
	if len(agent.Resources) != 4 {
		t.Fatalf("resources = %s, want the agent's own entry, the two workspace globs, then the knowledge base", data)
	}
	var own string
	if err := json.Unmarshal(agent.Resources[0], &own); err != nil || own != "file:///own.md" {
		t.Errorf("resources[0] = %s, want the agent's own %q entry preserved first", agent.Resources[0], "file:///own.md")
	}
	for i, want := range []string{"file://.kiro/steering/**/*.md", "skill://.kiro/skills/**/SKILL.md"} {
		var got string
		if err := json.Unmarshal(agent.Resources[1+i], &got); err != nil || got != want {
			t.Errorf("resources[%d] = %s, want the workspace glob %q", 1+i, agent.Resources[1+i], want)
		}
	}
	var kb map[string]any
	if err := json.Unmarshal(agent.Resources[3], &kb); err != nil {
		t.Fatalf("resources[3] = %s, want a knowledgeBase object: %v", agent.Resources[3], err)
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
	testutil.Symlink(t, oldSource, mount)

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

	// A symlink, or on Windows without the symlink privilege a junction.
	if !dirlink.Is(kbMount) {
		t.Fatalf("%s is not a link", kbMount)
	}
	target, err := dirlink.Target(kbMount)
	if err != nil {
		t.Fatalf("readlink %s: %v", kbMount, err)
	}
	absSource, _ := filepath.Abs(kbSource)
	if filepath.Clean(target) != absSource {
		t.Errorf("link target = %q, want %q", target, absSource)
	}

	agentJSON, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil {
		t.Fatalf("read installed agent1.json: %v", err)
	}
	if !strings.Contains(string(agentJSON), `"type":"knowledgeBase"`) && !strings.Contains(string(agentJSON), `"type": "knowledgeBase"`) {
		t.Errorf("installed agent1.json = %s, want a knowledgeBase resource entry", agentJSON)
	}
	if !strings.Contains(string(agentJSON), `"file://`+filepath.ToSlash(kbMount)+`"`) {
		t.Errorf("installed agent1.json = %s, want it to reference the mount point file://%s", agentJSON, filepath.ToSlash(kbMount))
	}

	// A reinstall replaces kirokit's own link rather than refusing it.
	if err := Install(roots, team, kiroHome, kb); err != nil {
		t.Fatalf("reinstall with KnowledgeBase: %v", err)
	}
	if !dirlink.Is(kbMount) {
		t.Errorf("%s is not a link after a reinstall", kbMount)
	}
}

func TestInstall_KnowledgeBaseRefusesARealFolderAtTheMountPoint(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	kbMount := filepath.Join(kiroHome, "kirokit-kb", team, "agent1")
	mustWrite(t, filepath.Join(kbMount, "mine.md"), "mine\n")

	err := Install(roots, team, kiroHome, &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: kbMount})
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("Install = %v, want a refusal to overwrite the real folder", err)
	}
	if _, err := os.Stat(filepath.Join(kbMount, "mine.md")); err != nil {
		t.Errorf("the real folder's file is gone: %v", err)
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

// installTree plans and writes a rendered tree the way Install does, without
// the backup step.
func installTree(renderedDir, kiroHome string) ([]string, error) {
	planned, err := planRenderedTree(renderedDir, kiroHome)
	if err != nil {
		return nil, err
	}
	return writePlanned(planned)
}

func TestCopyRenderedTree_PreservesOwnerExecuteBit(t *testing.T) {
	testutil.SkipWithoutExecBits(t)
	rendered := t.TempDir()
	kiroHome := t.TempDir()
	script := filepath.Join(rendered, "skills", "pdf", "scripts", "run.sh")
	plain := filepath.Join(rendered, "skills", "pdf", "SKILL.md")
	mustWrite(t, script, "echo hi\n")
	mustWrite(t, plain, "skill\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := installTree(rendered, kiroHome); err != nil {
		t.Fatalf("copyRenderedTree: %v", err)
	}

	modeOf := func(p string) os.FileMode {
		t.Helper()
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		return info.Mode().Perm()
	}
	if got := modeOf(filepath.Join(kiroHome, "skills", "pdf", "scripts", "run.sh")); got != 0o700 {
		t.Errorf("installed script mode = %o, want 700", got)
	}
	if got := modeOf(filepath.Join(kiroHome, "skills", "pdf", "SKILL.md")); got != 0o600 {
		t.Errorf("installed plain file mode = %o, want 600", got)
	}
}

func TestCopyRenderedTree_ReinstallMakesAPreviouslyPlainFileExecutable(t *testing.T) {
	testutil.SkipWithoutExecBits(t)
	rendered := t.TempDir()
	kiroHome := t.TempDir()
	script := filepath.Join(rendered, "skills", "pdf", "run.sh")
	installed := filepath.Join(kiroHome, "skills", "pdf", "run.sh")
	mustWrite(t, script, "echo hi\n")
	if _, err := installTree(rendered, kiroHome); err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(installed); info.Mode().Perm() != 0o600 {
		t.Fatalf("precondition: first install mode = %o, want 600", info.Mode().Perm())
	}

	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := installTree(rendered, kiroHome); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(installed); err != nil || info.Mode().Perm() != 0o700 {
		t.Errorf("after re-install mode = %v (%v), want 700", info.Mode().Perm(), err)
	}
}

func backupDirs(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".backup") {
			names = append(names, e.Name())
		}
	}
	return names
}

func backupOf(t *testing.T, root, folder string) string {
	t.Helper()
	for _, n := range backupDirs(t, root) {
		if strings.HasSuffix(n, "-"+folder+".backup") {
			return filepath.Join(root, n)
		}
	}
	return ""
}

func TestInstall_BacksUpTheFoldersItWouldChangeBeforeWritingAnything(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	mustWrite(t, filepath.Join(kiroHome, "agents", "agent1.json"), `{"name": "my-own-agent"}`)
	mustWrite(t, filepath.Join(kiroHome, "agents", "other.json"), `{"name": "someone else's"}`)
	mustWrite(t, filepath.Join(kiroHome, "kirokit-secrets", team, "agent1", "secrets.enc.yaml"), "previous ciphertext\n")
	mustWrite(t, filepath.Join(kiroHome, "sessions", "log.txt"), "kiro's own data")
	mustWrite(t, filepath.Join(kiroHome, "skills", "unrelated", "SKILL.md"), "a skill kirokit does not touch")

	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	agents := backupOf(t, kiroHome, "agents")
	secrets := backupOf(t, kiroHome, "kirokit-secrets")
	if agents == "" || secrets == "" {
		t.Fatalf("backup folders = %v, want one for agents and one for kirokit-secrets", backupDirs(t, kiroHome))
	}
	if got, _ := os.ReadFile(filepath.Join(agents, "agent1.json")); string(got) != `{"name": "my-own-agent"}` {
		t.Errorf("agents backup holds agent1.json = %q, want the previous content", got)
	}
	if got, _ := os.ReadFile(filepath.Join(agents, "other.json")); string(got) != `{"name": "someone else's"}` {
		t.Errorf("agents backup is missing the file kirokit does not manage: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(secrets, team, "agent1", "secrets.enc.yaml")); string(got) != "previous ciphertext\n" {
		t.Errorf("secrets backup = %q, want the previous ciphertext", got)
	}
	for _, untouched := range []string{"sessions", "skills"} {
		if b := backupOf(t, kiroHome, untouched); b != "" {
			t.Errorf("%s was backed up (%s) although no file in it changed", untouched, b)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(kiroHome, "agents", "other.json")); string(got) != `{"name": "someone else's"}` {
		t.Errorf("the live agents/other.json changed: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json")); strings.Contains(string(got), "my-own-agent") {
		t.Errorf("agent1.json still holds the old content: %s", got)
	}
}

func TestInstall_NoBackupFileEverLandsInsideALiveFolder(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	mustWrite(t, filepath.Join(kiroHome, "agents", "agent1.json"), `{"name": "mine"}`)

	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatal(err)
	}

	for _, live := range []string{"agents", "kirokit-secrets"} {
		matches, _ := filepath.Glob(filepath.Join(kiroHome, live, "*.backup"))
		if len(matches) != 0 {
			t.Errorf("%s/ contains backup entries %v; a backup there could be loaded as live config", live, matches)
		}
	}
}

func TestInstall_BackupsAreNotTrackedSoUninstallLeavesThem(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	mustWrite(t, filepath.Join(kiroHome, "agents", "agent1.json"), `{"name": "mine"}`)
	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatal(err)
	}
	recorded, err := os.ReadFile(installManifestPath(kiroHome, team))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(recorded), ".backup") {
		t.Errorf("install manifest lists a backup: %s", recorded)
	}

	if err := Uninstall(kiroHome, team); err != nil {
		t.Fatal(err)
	}

	backup := backupOf(t, kiroHome, "agents")
	if got, err := os.ReadFile(filepath.Join(backup, "agent1.json")); err != nil || string(got) != `{"name": "mine"}` {
		t.Errorf("backup after uninstall = %q (%v), want the user's file preserved", got, err)
	}
}

func TestInstall_ReinstallingUnchangedConfigCreatesNoBackups(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()

	for i := 0; i < 3; i++ { // what a scheduled sync does
		if err := Install(roots, team, kiroHome, nil); err != nil {
			t.Fatalf("Install #%d: %v", i+1, err)
		}
	}

	if got := backupDirs(t, kiroHome); len(got) != 0 {
		t.Errorf("repeated installs of unchanged config created backups: %v", got)
	}
}

func TestSync_BacksUpOnlyWhenTheShippedConfigChanged(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	if _, err := Sync(roots, team, kiroHome, nil); err != nil {
		t.Fatal(err)
	}
	if got := backupDirs(t, kiroHome); len(got) != 0 {
		t.Fatalf("first sync created backups: %v", got)
	}

	agentTmpl := filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "agent.json.tmpl")
	mustWrite(t, agentTmpl, `{"name": "agent1-v2"}`)
	if _, err := Sync(roots, team, kiroHome, nil); err != nil {
		t.Fatal(err)
	}

	agents := backupOf(t, kiroHome, "agents")
	if agents == "" {
		t.Fatalf("a changed agent did not produce an agents backup: %v", backupDirs(t, kiroHome))
	}
	if got, _ := os.ReadFile(filepath.Join(agents, "agent1.json")); strings.Contains(string(got), "agent1-v2") {
		t.Errorf("the backup holds the new content: %s", got)
	}
}

func TestInstall_ReinstallingWithAKnowledgeBaseCreatesNoBackups(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	mustWrite(t, filepath.Join(kbSource, "doc.md"), "# Docs\n")
	kb := &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: filepath.Join(kiroHome, "kirokit-kb", team, "agent1")}

	for i := 0; i < 3; i++ { // a scheduled sync with --kb-* flags
		if err := Install(roots, team, kiroHome, kb); err != nil {
			t.Fatalf("Install #%d: %v", i+1, err)
		}
	}

	if got := backupDirs(t, kiroHome); len(got) != 0 {
		t.Errorf("re-installing an unchanged config with a knowledge base created backups: %v", got)
	}
	agentJSON, err := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if err != nil || strings.Count(string(agentJSON), "knowledgeBase") != 1 {
		t.Errorf("agent1.json = %s (%v), want exactly one knowledgeBase resource after repeated installs", agentJSON, err)
	}
}

func TestInstall_AChangedAgentWithAKnowledgeBaseIsBackedUpWithItsPreviousKnowledgeBase(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	kbSource := t.TempDir()
	mustWrite(t, filepath.Join(kbSource, "doc.md"), "# Docs\n")
	kb := &KnowledgeBase{Agent: "agent1", Source: kbSource, Mount: filepath.Join(kiroHome, "kirokit-kb", team, "agent1")}
	if err := Install(roots, team, kiroHome, kb); err != nil {
		t.Fatal(err)
	}

	mustWrite(t, filepath.Join(roots.TeamsRoot, team, "agents", "agent1", "agent.json.tmpl"), `{"name": "agent1-v2"}`)
	if err := Install(roots, team, kiroHome, kb); err != nil {
		t.Fatal(err)
	}

	agents := backupOf(t, kiroHome, "agents")
	if agents == "" {
		t.Fatalf("a changed agent produced no agents backup: %v", backupDirs(t, kiroHome))
	}
	old, _ := os.ReadFile(filepath.Join(agents, "agent1.json"))
	if !strings.Contains(string(old), "knowledgeBase") || strings.Contains(string(old), "agent1-v2") {
		t.Errorf("backup = %s, want the previous agent with its knowledge base attached", old)
	}
	now, _ := os.ReadFile(filepath.Join(kiroHome, "agents", "agent1.json"))
	if !strings.Contains(string(now), "agent1-v2") || !strings.Contains(string(now), "knowledgeBase") {
		t.Errorf("installed agent = %s, want the new agent with the knowledge base attached", now)
	}
}

// A Kiro home whose agents folder is a symbolic link (a dotfiles manager
// linking ~/.kiro/agents) installs as it did before this change: the backup is
// the link, the folder stays a link, and the new agent lands in the folder it
// points at.
func TestInstall_AnAgentsFolderThatIsASymlinkStillInstalls(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	dotfiles := filepath.Join(t.TempDir(), "kiro-agents")
	mustWrite(t, filepath.Join(dotfiles, "agent1.json"), `{"name": "my-own-agent"}`)
	testutil.Symlink(t, dotfiles, filepath.Join(kiroHome, "agents"))

	if err := Install(roots, team, kiroHome, nil); err != nil {
		t.Fatalf("Install into a Kiro home with a symlinked agents folder: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(kiroHome, "*-agents.backup"))
	if len(matches) != 1 {
		t.Fatalf("agents backups = %v, want one", matches)
	}
	if got, err := os.Readlink(matches[0]); err != nil || got != dotfiles {
		t.Errorf("agents backup = %q (%v), want a link to %q", got, err, dotfiles)
	}
	if !dirlink.Is(filepath.Join(kiroHome, "agents")) {
		t.Error("the agents folder is no longer a link")
	}
	if got, _ := os.ReadFile(filepath.Join(dotfiles, "agent1.json")); strings.Contains(string(got), "my-own-agent") || !strings.Contains(string(got), "test-mcp") {
		t.Errorf("agent1.json in the linked folder = %s, want the installed agent", got)
	}
}

func TestBackupChangedFolders_ARootLevelFileIsBackedUpAsItself(t *testing.T) {
	kiroHome := t.TempDir()
	mustWrite(t, filepath.Join(kiroHome, "settings.json"), "mine")

	err := backupChangedFolders(kiroHome, []plannedFile{{target: filepath.Join(kiroHome, "settings.json"), data: []byte("new"), perm: 0o600}})
	if err != nil {
		t.Fatalf("backupChangedFolders: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(kiroHome, "*-settings.json.backup"))
	if len(matches) != 1 {
		t.Fatalf("backups = %v, want the root-level file backed up as itself", matches)
	}
	if got, _ := os.ReadFile(matches[0]); string(got) != "mine" {
		t.Errorf("backup = %q, want the previous content", got)
	}
}
