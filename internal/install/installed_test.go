package install

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// installTeams installs the fixture team under each given name into kiroHome.
func installTeams(t *testing.T, kiroHome string, names ...string) {
	t.Helper()
	roots, _, _ := setupFixture(t)
	for _, name := range names {
		mustWrite(t, filepath.Join(roots.TeamsRoot, name, "manifest.yaml"), "team: "+name+"\nglobalSteerings: []\n")
		mustWrite(t, filepath.Join(roots.TeamsRoot, name, "agents", "agent-"+name, "agent.json.tmpl"),
			`{"name": "agent-`+name+`"}`)
	}
	for _, name := range names {
		if err := Install(roots, name, kiroHome, nil); err != nil {
			t.Fatalf("Install %s: %v", name, err)
		}
	}
}

func TestInstalled_ListsRecordsSortedWithFileCountAndTime(t *testing.T) {
	kiroHome := t.TempDir()
	installTeams(t, kiroHome, "zeta-team", "alpha-team")

	got, err := Installed(kiroHome)
	if err != nil {
		t.Fatalf("Installed: %v", err)
	}
	if len(got) != 2 || got[0].Team != "alpha-team" || got[1].Team != "zeta-team" {
		t.Fatalf("Installed = %+v, want alpha-team then zeta-team", got)
	}
	for _, r := range got {
		if r.Files < 1 {
			t.Errorf("%s: Files = %d, want at least the agent file", r.Team, r.Files)
		}
		info, err := os.Stat(filepath.Join(kiroHome, ".kirokit-installed", r.Team+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !r.Modified.Equal(info.ModTime()) {
			t.Errorf("%s: Modified = %v, want the record's mtime %v", r.Team, r.Modified, info.ModTime())
		}
		if r.Modified.After(time.Now().Add(time.Minute)) {
			t.Errorf("%s: Modified %v is in the future", r.Team, r.Modified)
		}
	}
}

func TestInstalled_MissingRecordsFolderIsEmpty(t *testing.T) {
	got, err := Installed(t.TempDir())
	if err != nil {
		t.Fatalf("Installed on a fresh home: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Installed = %+v, want none", got)
	}
}

func TestUninstallOthers_RemovesEveryTeamButTheKeptOne(t *testing.T) {
	kiroHome := t.TempDir()
	installTeams(t, kiroHome, "a-team", "b-team", "c-team")

	removed, err := UninstallOthers(kiroHome, "b-team")
	if err != nil {
		t.Fatalf("UninstallOthers: %v", err)
	}
	if want := []string{"a-team", "c-team"}; !reflect.DeepEqual(removed, want) {
		t.Errorf("removed = %v, want %v", removed, want)
	}
	left, err := Installed(kiroHome)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 1 || left[0].Team != "b-team" {
		t.Errorf("Installed after = %+v, want only b-team", left)
	}
	if _, err := os.Stat(filepath.Join(kiroHome, "agents", "agent-b-team.json")); err != nil {
		t.Errorf("kept team's agent was touched: %v", err)
	}
	if _, err := os.Stat(filepath.Join(kiroHome, "agents", "agent-a-team.json")); !os.IsNotExist(err) {
		t.Errorf("removed team's agent still exists (err=%v)", err)
	}
}

func TestUninstallOthers_KeepingTheOnlyTeamRemovesNothing(t *testing.T) {
	kiroHome := t.TempDir()
	installTeams(t, kiroHome, "a-team")

	removed, err := UninstallOthers(kiroHome, "a-team")
	if err != nil || len(removed) != 0 {
		t.Fatalf("UninstallOthers = %v, %v; want none", removed, err)
	}
}

func TestUninstallOthers_FailedUninstallStopsAndReturnsTheError(t *testing.T) {
	kiroHome := t.TempDir()
	installTeams(t, kiroHome, "a-team", "b-team", "c-team")
	// A corrupt record makes the first uninstall fail.
	mustWrite(t, filepath.Join(kiroHome, ".kirokit-installed", "a-team.json"), "{not json")

	removed, err := UninstallOthers(kiroHome, "c-team")
	if err == nil {
		t.Fatal("UninstallOthers returned nil for a corrupt record")
	}
	if len(removed) != 0 {
		t.Errorf("removed = %v, want none before the failure", removed)
	}
	if _, statErr := os.Stat(filepath.Join(kiroHome, "agents", "agent-b-team.json")); statErr != nil {
		t.Errorf("b-team was uninstalled after a-team failed: %v", statErr)
	}
}
