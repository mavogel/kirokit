package install

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSync_PrunesFilesNoLongerShippedAndKeepsUnrelatedOnes(t *testing.T) {
	roots, team, _ := setupFixture(t)
	kiroHome := t.TempDir()
	agent2 := filepath.Join(roots.TeamsRoot, team, "agents", "agent2", "agent.json.tmpl")
	mustWrite(t, agent2, `{"name": "agent2"}`)
	mustWrite(t, filepath.Join(filepath.Dir(agent2), "env", "secrets.enc.yaml"), "ciphertext\n")

	unrelated := filepath.Join(kiroHome, "agents", "unrelated.json")
	mustWrite(t, unrelated, `{"name": "unrelated"}`)

	if _, err := Sync(roots, team, kiroHome, nil); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	installed := filepath.Join(kiroHome, "agents", "agent2.json")
	if _, err := os.Stat(installed); err != nil {
		t.Fatalf("expected %s after first sync: %v", installed, err)
	}

	if err := os.RemoveAll(filepath.Dir(agent2)); err != nil {
		t.Fatalf("remove agent2 template: %v", err)
	}
	pruned, err := Sync(roots, team, kiroHome, nil)
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}

	if !reflect.DeepEqual(pruned, []string{installed, filepath.Join(kiroHome, "kirokit-secrets", team, "agent2", "secrets.enc.yaml")}) {
		t.Errorf("pruned = %v, want agent2.json and its secrets file", pruned)
	}
	if _, err := os.Stat(filepath.Join(kiroHome, "kirokit-secrets", team, "agent2")); !os.IsNotExist(err) {
		t.Errorf("empty secrets directory of the pruned agent should be removed (err=%v)", err)
	}
	if _, err := os.Stat(installed); !os.IsNotExist(err) {
		t.Errorf("agent2.json still installed after its template was removed (err=%v)", err)
	}
	for _, keep := range []string{unrelated, filepath.Join(kiroHome, "agents", "agent1.json")} {
		if _, err := os.Stat(keep); err != nil {
			t.Errorf("%s should survive sync: %v", keep, err)
		}
	}
}

func TestSync_FirstInstallPrunesNothing(t *testing.T) {
	roots, team, _ := setupFixture(t)

	pruned, err := Sync(roots, team, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(pruned) != 0 {
		t.Errorf("pruned = %v, want none on a first install", pruned)
	}
}
