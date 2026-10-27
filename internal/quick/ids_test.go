package quick

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadIDs_ReadsTheSkillsMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quick-ids.yaml")
	writeFiles(t, dir, map[string]string{"quick-ids.yaml": "# comment\nskills:\n  user-story: " + idStory + "\n  other: " + idOther + "\n"})

	ids, err := LoadIDs(path)
	if err != nil {
		t.Fatalf("LoadIDs: %v", err)
	}
	if len(ids) != 2 || ids["user-story"] != idStory || ids["other"] != idOther {
		t.Errorf("ids = %v", ids)
	}
}

func TestLoadIDs_MissingFileIsEmpty(t *testing.T) {
	ids, err := LoadIDs(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil || len(ids) != 0 {
		t.Errorf("LoadIDs = %v, %v; want an empty map", ids, err)
	}
}

func TestLoadIDs_RejectsMalformedFiles(t *testing.T) {
	for name, tc := range map[string]struct{ content, want string }{
		"not yaml":      {"skills: [\n", "quick-ids.yaml"},
		"unknown key":   {"skill:\n  a: " + idStory + "\n", "skill"},
		"bad id":        {"skills:\n  a: 0123\n", "32 lowercase hex"},
		"upper-case id": {"skills:\n  a: " + strings.ToUpper(idStory) + "\n", "32 lowercase hex"},
		"empty name":    {"skills:\n  \"\": " + idStory + "\n", "empty skill name"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{"quick-ids.yaml": tc.content})
			_, err := LoadIDs(filepath.Join(dir, "quick-ids.yaml"))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
