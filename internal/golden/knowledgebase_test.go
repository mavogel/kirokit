package golden

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/install"
)

var updateKnowledgeBaseGolden = flag.Bool("update-kb-golden", false,
	"rewrite testdata/golden/platform-team/installed-with-kb instead of comparing against it")

// installPlatformTeamWithKnowledgeBase installs the real platform-team into a
// scratch KIRO_HOME with the committed testdata/knowledgebase folder attached
// to its devops agent, and returns that home and the knowledge base's mount point.
func installPlatformTeamWithKnowledgeBase(t *testing.T) (kiroHome, mount string) {
	t.Helper()
	repo := filepath.Join("..", "..")
	roots := install.Roots{
		TeamsRoot:       filepath.Join(repo, "teams"),
		GlobalRoot:      filepath.Join(repo, "global", "steerings"),
		EnvVersionsPath: filepath.Join(repo, "env", "mcp-versions.yaml"),
	}
	kiroHome = t.TempDir()
	mount = filepath.Join(kiroHome, "kirokit-kb", "platform-team", "devops")
	kb := &install.KnowledgeBase{Agent: "devops", Source: filepath.Join(repo, "testdata", "knowledgebase"), Mount: mount}

	if err := install.Install(roots, "platform-team", kiroHome, kb); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	return kiroHome, mount
}

// TestKnowledgeBaseRenderedIntoAgentJSONMatchesGolden compares the complete
// agent config a user ends up with once a knowledge base is attached: the
// rendered agent plus the knowledgeBase resource that install adds. The
// ordinary goldens cannot cover this, because attaching happens at install
// time, with a real KIRO_HOME.
//
// install writes compact JSON, so the config is re-indented for a readable
// golden, and the scratch KIRO_HOME is replaced by the same /kiro-home
// sentinel the render goldens use. Regenerate with:
//
//	go test ./internal/golden -run KnowledgeBaseRenderedIntoAgentJSON -update-kb-golden
func TestKnowledgeBaseRenderedIntoAgentJSONMatchesGolden(t *testing.T) {
	kiroHome, _ := installPlatformTeamWithKnowledgeBase(t)

	raw, err := os.ReadFile(filepath.Join(kiroHome, "agents", "devops.json"))
	if err != nil {
		t.Fatalf("read installed devops.json: %v", err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		t.Fatalf("installed devops.json is not valid JSON: %v", err)
	}
	pretty.WriteByte('\n')
	got := strings.ReplaceAll(pretty.String(), kiroHome, "/kiro-home")

	goldenPath := filepath.Join("..", "..", "testdata", "golden", "platform-team", "installed-with-kb", "agents", "devops.json")
	if *updateKnowledgeBaseGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
			t.Fatalf("write %s: %v", goldenPath, err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v\nrun: go test ./internal/golden -run KnowledgeBaseRenderedIntoAgentJSON -update-kb-golden", goldenPath, err)
	}
	if got != string(want) {
		t.Errorf("installed devops.json with a knowledge base differs from %s\n--- got ---\n%s\n--- want ---\n%s\nIf the change is intended, regenerate with -update-kb-golden and review the diff.", goldenPath, got, want)
	}
}
