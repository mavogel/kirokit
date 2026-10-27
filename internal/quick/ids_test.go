package quick

import (
	"archive/zip"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
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

// exported is a skill as Quick exports it: the imported file plus an id line.
func exported(name, id string) string {
	s := "---\nname: " + name + "\ndescription: \"Test skill.\"\n"
	if id != "" {
		s += "id: " + id + "\n"
	}
	return s + "---\n\n## Overview\nBody with {{input}}.\n"
}

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, content := range entries {
		e, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReadExports_FilesFoldersAndZips(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"single/SKILL-v4.md":     exported("context7", idStory),
		"folder/a/SKILL.md":      exported("story", idOther),
		"folder/deep/b/SKILL.md": exported("noid", ""),
		"folder/ignored.md":      exported("ignored", "11111111111111111111111111111111"),
	})
	writeZip(t, filepath.Join(dir, "folder", "export-v4.zip"), map[string]string{"SKILL.md": exported("zipped", "22222222222222222222222222222222")})

	got, skipped, err := ReadExports([]string{filepath.Join(dir, "single", "SKILL-v4.md"), filepath.Join(dir, "folder")})
	if err != nil {
		t.Fatalf("ReadExports: %v", err)
	}
	want := map[string]string{"context7": idStory, "story": idOther, "zipped": "22222222222222222222222222222222"}
	if !maps.Equal(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
	if len(skipped) != 1 || !strings.HasSuffix(skipped[0], filepath.Join("deep", "b", "SKILL.md")) {
		t.Errorf("skipped = %v, want the id-less export", skipped)
	}
}

func TestReadExports_Rejects(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"a/SKILL.md":   exported("story", idStory),
		"b/SKILL.md":   exported("story", idOther),
		"bad/SKILL.md": exported("story", "nothex"),
		"nofm.md":      "no frontmatter\n",
	})
	many := map[string]string{}
	for i := range MaxZipEntries + 1 {
		many[fmt.Sprintf("s%d/SKILL.md", i)] = exported(fmt.Sprintf("s%d", i), "")
	}
	writeZip(t, filepath.Join(dir, "many.zip"), many)
	writeZip(t, filepath.Join(dir, "big.zip"), map[string]string{"SKILL.md": exported("big", idStory) + strings.Repeat("x", MaxZipEntryBytes)})

	for name, tc := range map[string]struct {
		paths []string
		want  string
	}{
		"two ids for one name": {[]string{filepath.Join(dir, "a"), filepath.Join(dir, "b")}, `two different ids for skill "story"`},
		"malformed id":         {[]string{filepath.Join(dir, "bad")}, "32 lowercase hex"},
		"no frontmatter":       {[]string{filepath.Join(dir, "nofm.md")}, "frontmatter"},
		"too many entries":     {[]string{filepath.Join(dir, "many.zip")}, "more than 256"},
		"entry too large":      {[]string{filepath.Join(dir, "big.zip")}, "larger than"},
		"missing path":         {[]string{filepath.Join(dir, "nope")}, "nope"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := ReadExports(tc.paths); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestWriteIDs_MergesSortsAndKeepsItPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quick-ids.yaml")
	if err := WriteIDs(path, map[string]string{"zeta": idOther, "alpha": idStory}); err != nil {
		t.Fatalf("WriteIDs: %v", err)
	}
	want := "# Amazon Quick skill ids for one Quick account. Written by `kirokit quick ids`; never commit.\nskills:\n  alpha: " + idStory + "\n  zeta: " + idOther + "\n"
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("file =\n%s\nwant\n%s", got, want)
	}
	if info, err := os.Stat(path); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o600) {
		t.Errorf("mode = %v, %v; want 0600", info.Mode(), err)
	}
	ids, err := LoadIDs(path)
	if err != nil || ids["alpha"] != idStory {
		t.Errorf("LoadIDs after WriteIDs = %v, %v", ids, err)
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
