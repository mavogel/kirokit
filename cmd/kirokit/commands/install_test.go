package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseKnowledgeBaseFlags_AllOrNone(t *testing.T) {
	tests := []struct {
		name                  string
		agent, source, mount  string
		wantNil, wantTogether bool
	}{
		{"none given", "", "", "", true, false},
		{"all given", "devops", "/kb", "/mnt", false, false},
		{"only agent", "devops", "", "", false, true},
		{"only source", "", "/kb", "", false, true},
		{"only mount", "", "", "/mnt", false, true},
		{"agent and source", "devops", "/kb", "", false, true},
		{"agent and mount", "devops", "", "/mnt", false, true},
		{"source and mount", "", "/kb", "/mnt", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kb, err := parseKnowledgeBaseFlags(tc.agent, tc.source, tc.mount)

			if tc.wantTogether {
				if err == nil || !strings.Contains(err.Error(), "must be given together") {
					t.Fatalf("err = %v, want a %q error", err, "must be given together")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if kb != nil {
					t.Errorf("kb = %+v, want nil when no flag is given", kb)
				}
				return
			}
			if kb == nil || kb.Agent != tc.agent || kb.Source != tc.source || kb.Mount != tc.mount {
				t.Errorf("kb = %+v, want agent=%q source=%q mount=%q", kb, tc.agent, tc.source, tc.mount)
			}
		})
	}
}

func installedRecords(t *testing.T, home string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".kirokit-installed"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestInstall_ReplacesTheInstalledTeamAndSaysSo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)
	t.Setenv("KIROKIT_REPO", thisRepo(t))
	chdir(t, t.TempDir())

	if got, err := runRoot(t, "install", "platform-team"); err != nil || got != "installed platform-team into "+home+"\n" {
		t.Fatalf("first install = %q, %v", got, err)
	}

	got, err := runRoot(t, "install", "data-team")
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	want := "uninstalled platform-team from " + home + "\ninstalled data-team into " + home + "\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if recs := installedRecords(t, home); !reflect.DeepEqual(recs, []string{"data-team.json"}) {
		t.Errorf("records = %v, want only data-team.json", recs)
	}

	got, err = runRoot(t, "install", "data-team")
	if err != nil || got != "installed data-team into "+home+"\n" {
		t.Errorf("reinstall = %q, %v; want only the installed line", got, err)
	}
}

func TestInstall_FailingUninstallPrintsNoInstalledLine(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KIRO_HOME", home)
	t.Setenv("KIROKIT_REPO", thisRepo(t))
	chdir(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(home, ".kirokit-installed"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".kirokit-installed", "old-team.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := runRoot(t, "install", "data-team")
	if err == nil || !strings.HasPrefix(err.Error(), "kirokit install: ") {
		t.Fatalf("err = %v, want a %q error", err, "kirokit install: ")
	}
	if strings.Contains(got, "installed data-team") {
		t.Errorf("output = %q, want no installed line", got)
	}
}
