package golden

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// requireKiroCLIEnv turns this file's skips into failures, for environments
// (e.g. CI) that are expected to provide a logged-in kiro-cli.
const requireKiroCLIEnv = "KIROKIT_REQUIRE_KIRO_CLI"

func skipUnlessRequired(t *testing.T, reason string) {
	t.Helper()
	if os.Getenv(requireKiroCLIEnv) != "" {
		t.Fatalf("%s (%s is set)", reason, requireKiroCLIEnv)
	}
	t.Skip(reason)
}

func kiroCLI(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("kiro-cli")
	if err != nil {
		skipUnlessRequired(t, "kiro-cli is not installed")
	}
	return bin
}

// validateWithKiroCLI runs Kiro's own `kiro-cli agent validate`, which is
// what actually loads an agent config. It catches what a golden diff cannot:
// a config that renders deterministically but that Kiro rejects (invalid
// JSON, wrong field types, an unknown resource type, a prompt file that does
// not resolve from the agent file's directory).
//
// validate exits 0 even for an invalid config: it reports on stderr and is
// silent on success, so the output is what is checked. It does not check
// permission rules, MCP server shape, or that skill:// and file:// resources
// exist.
func validateWithKiroCLI(t *testing.T, bin, agentPath string) {
	t.Helper()
	var out bytes.Buffer
	cmd := exec.Command(bin, "agent", "validate", "--path", agentPath)
	cmd.Stdout, cmd.Stderr = &out, &out
	runErr := cmd.Run()
	got := strings.TrimSpace(out.String())

	if strings.Contains(got, "not logged in") {
		skipUnlessRequired(t, "kiro-cli is not logged in (validate needs a login)")
	}
	if runErr != nil || strings.Contains(got, "Error") {
		t.Errorf("kiro-cli agent validate rejected %s (err=%v):\n%s", agentPath, runErr, got)
	}
}

func TestCommittedGoldenAgentsValidateWithKiroCLI(t *testing.T) {
	bin := kiroCLI(t)

	agents, err := filepath.Glob(filepath.Join("..", "..", "testdata", "golden", "*", "rendered", "agents", "*.json"))
	if err != nil || len(agents) == 0 {
		t.Fatalf("no committed golden agent configs found (err=%v)", err)
	}

	for _, agent := range agents {
		abs, err := filepath.Abs(agent)
		if err != nil {
			t.Fatalf("resolve %s: %v", agent, err)
		}
		team := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(abs))))
		t.Run(team+"/"+filepath.Base(abs), func(t *testing.T) {
			validateWithKiroCLI(t, bin, abs)
		})
	}
}

// TestInstalledAgentWithKnowledgeBaseValidatesWithKiroCLI installs the real
// platform-team with the committed testdata/knowledgebase folder attached to
// devops, which exercises the one config path no golden covers: install
// rewrites the agent JSON to add the knowledgeBase resource.
func TestInstalledAgentWithKnowledgeBaseValidatesWithKiroCLI(t *testing.T) {
	bin := kiroCLI(t)

	kiroHome, mount := installPlatformTeamWithKnowledgeBase(t)

	for _, doc := range []string{"platform-runbook.md", "domain-knowledge.md", "deploy-conventions.md", "aws-account-structure.md"} {
		if _, err := os.Stat(filepath.Join(mount, doc)); err != nil {
			t.Errorf("the mounted knowledge base does not expose the committed fixture %s: %v", doc, err)
		}
	}

	devops := filepath.Join(kiroHome, "agents", "devops.json")
	data, err := os.ReadFile(devops)
	if err != nil {
		t.Fatalf("read installed devops.json: %v", err)
	}
	var agent struct {
		Resources []json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("installed devops.json is not valid JSON: %v", err)
	}
	var found bool
	for _, raw := range agent.Resources {
		var res struct{ Type, Source string }
		if json.Unmarshal(raw, &res) == nil && res.Type == "knowledgeBase" && res.Source == "file://"+mount {
			found = true
		}
	}
	if !found {
		t.Errorf("installed devops.json has no knowledgeBase resource for %s: %s", mount, data)
	}

	for _, name := range []string{"devops", "platformdev"} {
		validateWithKiroCLI(t, bin, filepath.Join(kiroHome, "agents", name+".json"))
	}
}
