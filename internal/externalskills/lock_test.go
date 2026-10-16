package externalskills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mavogel/kirokit/internal/manifest"
)

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func mustHash(t *testing.T, dir string) string {
	t.Helper()
	h, err := HashDir(dir)
	if err != nil {
		t.Fatalf("HashDir(%s): %v", dir, err)
	}
	return h
}

func TestHashDir_IsStableAndSensitiveToEveryChange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "skill\n", 0o644)
	writeFile(t, filepath.Join(dir, "scripts", "run.sh"), "echo hi\n", 0o644)
	base := mustHash(t, dir)

	if again := mustHash(t, dir); again != base {
		t.Fatalf("hash not stable: %s vs %s", base, again)
	}

	mutations := map[string]func(dir string){
		"a byte changes":  func(d string) { writeFile(t, filepath.Join(d, "SKILL.md"), "skill!\n", 0o644) },
		"a file is added": func(d string) { writeFile(t, filepath.Join(d, "extra.md"), "x", 0o644) },
		"a file is renamed": func(d string) {
			_ = os.Rename(filepath.Join(d, "scripts", "run.sh"), filepath.Join(d, "scripts", "go.sh"))
		},
		"exec bit flips": func(d string) { _ = os.Chmod(filepath.Join(d, "scripts", "run.sh"), 0o755) },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			d := t.TempDir()
			writeFile(t, filepath.Join(d, "SKILL.md"), "skill\n", 0o644)
			writeFile(t, filepath.Join(d, "scripts", "run.sh"), "echo hi\n", 0o644)
			mutate(d)
			if got := mustHash(t, d); got == base {
				t.Errorf("hash unchanged after: %s", name)
			}
		})
	}
}

func TestHashDir_RejectsSymlinks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "skill\n", 0o644)
	if err := os.Symlink("/etc/passwd", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := HashDir(dir); err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("err = %v, want a symbolic link error", err)
	}
}

func TestLock_RoundTripSortedAndMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "external", "skills.lock.yaml")

	empty, err := ReadLock(path)
	if err != nil || len(empty.Skills) != 0 {
		t.Fatalf("ReadLock(missing) = %+v, %v; want empty lock and no error", empty, err)
	}

	in := Lock{Skills: []LockEntry{
		{Name: "zeta", Source: "s", Ref: "r", Path: "p/zeta", Commit: "c2", SHA256: "h2"},
		{Name: "alpha", Source: "s", Ref: "r", Path: "p/alpha", Commit: "c1", SHA256: "h1"},
	}}
	if err := WriteLock(path, in); err != nil {
		t.Fatalf("WriteLock: %v", err)
	}
	out, err := ReadLock(path)
	if err != nil {
		t.Fatalf("ReadLock: %v", err)
	}
	if len(out.Skills) != 2 || out.Skills[0].Name != "alpha" || out.Skills[1] != in.Skills[0] {
		t.Errorf("round trip = %+v, want sorted by name with fields intact", out.Skills)
	}
}

// consistent builds a team directory with one vendored skill, its lock entry
// and the matching manifest entry.
func consistent(t *testing.T) (teamDir string, entries []manifest.ExternalSkill) {
	t.Helper()
	teamDir = filepath.Join(t.TempDir(), "demo-team")
	skill := filepath.Join(SkillsDir(teamDir), "pdf")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "skill\n", 0o644)
	writeFile(t, filepath.Join(skill, "scripts", "run.sh"), "echo\n", 0o755)

	entries = []manifest.ExternalSkill{{Source: "https://h/r.git", Ref: "v1", Path: "skills/pdf"}}
	err := WriteLock(LockPath(teamDir), Lock{Skills: []LockEntry{{
		Name: "pdf", Source: "https://h/r.git", Ref: "v1", Path: "skills/pdf",
		Commit: "abc", SHA256: mustHash(t, skill),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	return teamDir, entries
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		break_    func(t *testing.T, teamDir string, entries []manifest.ExternalSkill) []manifest.ExternalSkill
		wantInErr []string
	}{
		{"consistent", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill { return e }, nil},
		{"vendored byte modified", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			writeFile(t, filepath.Join(SkillsDir(d), "pdf", "SKILL.md"), "tampered\n", 0o644)
			return e
		}, []string{`"pdf"`, "content", "kirokit skills fetch demo-team"}},
		{"ref changed without fetch", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			e[0].Ref = "v2"
			return e
		}, []string{`"pdf"`, "ref"}},
		{"source changed without fetch", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			e[0].Source = "https://h/other.git"
			return e
		}, []string{`"pdf"`, "source"}},
		{"no lock entry", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			if err := WriteLock(LockPath(d), Lock{}); err != nil {
				t.Fatal(err)
			}
			return e
		}, []string{`"pdf"`, "lock"}},
		{"vendored dir missing", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			if err := os.RemoveAll(filepath.Join(SkillsDir(d), "pdf")); err != nil {
				t.Fatal(err)
			}
			return e
		}, []string{`"pdf"`, "missing"}},
		{"SKILL.md missing", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			if err := os.Remove(filepath.Join(SkillsDir(d), "pdf", "SKILL.md")); err != nil {
				t.Fatal(err)
			}
			return e
		}, []string{`"pdf"`, "SKILL.md"}},
		{"vendored dir without manifest entry", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			writeFile(t, filepath.Join(SkillsDir(d), "stray", "SKILL.md"), "x", 0o644)
			return e
		}, []string{`"stray"`, "manifest"}},
		{"lock entry without manifest entry", func(_ *testing.T, _ string, _ []manifest.ExternalSkill) []manifest.ExternalSkill {
			return nil
		}, []string{`"pdf"`, "manifest"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamDir, entries := consistent(t)
			entries = tc.break_(t, teamDir, entries)

			err := Verify(teamDir, entries)
			if tc.wantInErr == nil {
				if err != nil {
					t.Fatalf("Verify = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Verify = nil, want an error")
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestVerify_TeamWithoutExternalSkillsIsClean(t *testing.T) {
	if err := Verify(filepath.Join(t.TempDir(), "plain-team"), nil); err != nil {
		t.Fatalf("Verify = %v, want nil when a team has no external skills and no external/ directory", err)
	}
}

const collectionSource = "https://h/r.git"

// vendorCollection vendors skills/<name> for each name into teamDir the way a
// fetch of an all: true entry would, and returns that manifest entry.
func vendorCollection(t *testing.T, teamDir string, names ...string) manifest.ExternalSkill {
	t.Helper()
	lock, err := ReadLock(LockPath(teamDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		dir := filepath.Join(SkillsDir(teamDir), name)
		writeFile(t, filepath.Join(dir, "SKILL.md"), name+"\n", 0o644)
		lock.Skills = append(lock.Skills, LockEntry{
			Name: name, Source: collectionSource, Ref: "v1", Path: "skills/" + name,
			From: "skills", Commit: "abc", SHA256: mustHash(t, dir),
		})
	}
	if err := WriteLock(LockPath(teamDir), lock); err != nil {
		t.Fatal(err)
	}
	return manifest.ExternalSkill{Source: collectionSource, Ref: "v1", Path: "skills", All: true, Exclude: []string{"docx"}}
}

func TestResolve_ReturnsExplicitAndCollectionSkillNames(t *testing.T) {
	teamDir, entries := consistent(t) // explicit pdf
	entries = append(entries, vendorCollection(t, teamDir, "b", "a"))

	names, err := Resolve(teamDir, entries)
	if err != nil {
		t.Fatalf("Resolve = %v, want nil", err)
	}
	if got := strings.Join(names, ","); got != "a,b,pdf" {
		t.Errorf("names = %q, want the sorted union a,b,pdf", got)
	}
	if err := Verify(teamDir, entries); err != nil {
		t.Errorf("Verify = %v, want nil for the same consistent team", err)
	}
}

func TestVerify_Collections(t *testing.T) {
	tests := []struct {
		name      string
		break_    func(t *testing.T, teamDir string, entries []manifest.ExternalSkill) []manifest.ExternalSkill
		wantInErr []string
	}{
		{"collection without lock entries", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			if err := WriteLock(LockPath(d), Lock{}); err != nil {
				t.Fatal(err)
			}
			return e
		}, []string{`"skills"`, "missing from the lock file"}},
		{"vendored skill that the manifest now excludes", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			e[0].Exclude = []string{"a"}
			return e
		}, []string{`"a"`, "exclude"}},
		{"lock entries without a manifest entry", func(_ *testing.T, _ string, _ []manifest.ExternalSkill) []manifest.ExternalSkill {
			return nil
		}, []string{`"a"`, "not declared in manifest.yaml"}},
		{"name produced by an explicit entry and the collection", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			return append(e, manifest.ExternalSkill{Source: collectionSource, Ref: "v1", Path: "skills/a"})
		}, []string{`"a"`, "more than one"}},
		{"collection-derived file modified", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			writeFile(t, filepath.Join(SkillsDir(d), "a", "SKILL.md"), "tampered\n", 0o644)
			return e
		}, []string{`"a"`, "does not match the lock file"}},
		{"ref changed without a fetch", func(_ *testing.T, _ string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			e[0].Ref = "v2"
			return e
		}, []string{`"skills"`, "ref"}},
		{"vendored directory the collection did not produce", func(t *testing.T, d string, e []manifest.ExternalSkill) []manifest.ExternalSkill {
			writeFile(t, filepath.Join(SkillsDir(d), "stray", "SKILL.md"), "x", 0o644)
			return e
		}, []string{`"stray"`, "no entry in manifest.yaml"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teamDir := filepath.Join(t.TempDir(), "demo-team")
			entries := []manifest.ExternalSkill{vendorCollection(t, teamDir, "a", "b")}
			entries = tc.break_(t, teamDir, entries)

			err := Verify(teamDir, entries)
			if err == nil {
				t.Fatal("Verify = nil, want an error")
			}
			for _, want := range tc.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestVerify_CollectionRefDriftIsReportedOnceNotAlsoAsUndeclared(t *testing.T) {
	teamDir := filepath.Join(t.TempDir(), "demo-team")
	entry := vendorCollection(t, teamDir, "a", "b")
	entry.Ref = "v2"

	err := Verify(teamDir, []manifest.ExternalSkill{entry})
	if err == nil || !strings.Contains(err.Error(), `"skills"`) || !strings.Contains(err.Error(), "ref") {
		t.Fatalf("err = %v, want the collection's source/ref drift", err)
	}
	if strings.Contains(err.Error(), "not declared in manifest.yaml") {
		t.Errorf("err = %q also reports the same lock entries as undeclared", err)
	}
}
