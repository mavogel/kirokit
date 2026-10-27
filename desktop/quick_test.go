package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestQuickDir_IsUnderTheAppDataFolder(t *testing.T) {
	dir := t.TempDir()
	a := NewAppAt(dir)
	got, err := a.QuickDir("product-team")
	if err != nil {
		t.Fatalf("QuickDir: %v", err)
	}
	if want := filepath.Join(dir, "quick", "product-team"); got != want {
		t.Errorf("QuickDir = %q, want %q", got, want)
	}
}

// The frontend passes only a team name; anything that could reach outside
// <dataDir>/quick is refused by every call that takes one.
func TestQuick_RefusesNamesThatAreNotPlainTeams(t *testing.T) {
	a := NewAppAt(t.TempDir())
	for _, name := range []string{"", ".", "..", "../x", "a/b", `a\b`, "/abs", "a\x00b"} {
		if _, err := a.QuickDir(name); err == nil {
			t.Errorf("QuickDir(%q) succeeded, want a refusal", name)
		}
		if _, err := a.QuickBundle(name); err == nil {
			t.Errorf("QuickBundle(%q) succeeded, want a refusal", name)
		}
		if err := a.OpenFolder(name); err == nil {
			t.Errorf("OpenFolder(%q) succeeded, want a refusal", name)
		}
	}
}

func writeBundle(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestQuickBundle_CountsSkillsAndAgents(t *testing.T) {
	data := t.TempDir()
	a := NewAppAt(data)
	dir := filepath.Join(data, "quick", "product-team")
	writeBundle(t, dir,
		"skills/user-story/SKILL.md", "skills/user-story/references/a.md", "skills/other/SKILL.md", "skills/empty/notes.md",
		"agents/product-owner.json", "agents/qa.json", "agents/readme.txt")

	got, err := a.QuickBundle("product-team")
	if err != nil {
		t.Fatalf("QuickBundle: %v", err)
	}
	if got != (QuickBundle{Dir: dir, Skills: 2, Agents: 2}) {
		t.Errorf("QuickBundle = %+v, want 2 skills and 2 agents in %s", got, dir)
	}
}

func TestQuickBundle_MissingFolderIsAnError(t *testing.T) {
	if _, err := NewAppAt(t.TempDir()).QuickBundle("product-team"); err == nil || !strings.Contains(err.Error(), "no Quick bundle") {
		t.Errorf("err = %v, want the missing bundle named", err)
	}
}

func TestOpenFolder_OpensOnlyAnExistingBundleFolder(t *testing.T) {
	data := t.TempDir()
	a := NewAppAt(data)
	var ran [][]string
	a.startFn = func(name string, args ...string) error {
		ran = append(ran, append([]string{name}, args...))
		return nil
	}

	if err := a.OpenFolder("product-team"); err == nil {
		t.Error("OpenFolder before an export succeeded")
	}
	dir := filepath.Join(data, "quick", "product-team")
	writeBundle(t, dir, "agents/po.json")
	if err := a.OpenFolder("product-team"); err != nil {
		t.Fatalf("OpenFolder: %v", err)
	}
	if want := [][]string{{openCommandFor(runtime.GOOS), dir}}; !slices.EqualFunc(ran, want, slices.Equal) {
		t.Errorf("ran %v, want %v", ran, want)
	}
}

func TestPlatform_OpenCommand(t *testing.T) {
	for goos, want := range map[string]string{"darwin": "open", "linux": "xdg-open", "freebsd": "xdg-open", "windows": "explorer"} {
		if got := openCommandFor(goos); got != want {
			t.Errorf("openCommandFor(%s) = %q, want %q", goos, got, want)
		}
	}
}

// The repository's product-team lists Quick content; the other teams do not.
func TestCatalog_MarksTeamsWithQuickContent(t *testing.T) {
	cat, err := BuildCatalog("..")
	if err != nil {
		t.Fatalf("BuildCatalog: %v", err)
	}
	for _, tm := range cat.Teams {
		if want := tm.Name == "product-team"; tm.Quick != want {
			t.Errorf("%s Quick = %v, want %v", tm.Name, tm.Quick, want)
		}
	}
}
