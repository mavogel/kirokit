package backup

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func fixedClock(t *testing.T, at time.Time) {
	t.Helper()
	prev := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = prev })
}

func put(t *testing.T, path, content string, mode os.FileMode) {
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

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestStamp_IsYYYYMMDDHHmmssInLocalTime(t *testing.T) {
	fixedClock(t, time.Date(2026, 10, 15, 9, 5, 7, 0, time.Local))

	if got := Stamp(); got != "20261015-090507" {
		t.Errorf("Stamp() = %q, want 20261015-090507", got)
	}
	if ok := regexp.MustCompile(`^\d{8}-\d{6}$`).MatchString(Stamp()); !ok {
		t.Errorf("Stamp() = %q does not match YYYYMMDD-HHmmss", Stamp())
	}
}

func TestMoveAside_RenamesTheFolderToATimestampedBackupBesideIt(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "out")
	put(t, filepath.Join(dir, "agents", "devops.json"), "mine", 0o600)

	got, err := MoveAside(dir, "20261015-101530")
	if err != nil {
		t.Fatalf("MoveAside: %v", err)
	}

	want := filepath.Join(parent, "20261015-101530-out.backup")
	if got != want {
		t.Errorf("backup path = %q, want %q", got, want)
	}
	if read(t, filepath.Join(want, "agents", "devops.json")) != "mine" {
		t.Error("the moved folder lost its content")
	}
	if _, err := os.Lstat(dir); !os.IsNotExist(err) {
		t.Errorf("the original path still exists after the move (err=%v)", err)
	}
}

func TestMoveAside_SameSecondCollisionGetsACounter(t *testing.T) {
	parent := t.TempDir()
	stamp := "20261015-101530"
	var paths []string
	for _, content := range []string{"first", "second"} {
		dir := filepath.Join(parent, "out")
		put(t, filepath.Join(dir, "f"), content, 0o600)
		p, err := MoveAside(dir, stamp)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}

	if paths[0] == paths[1] {
		t.Fatalf("both backups went to %q", paths[0])
	}
	if read(t, filepath.Join(paths[0], "f")) != "first" || read(t, filepath.Join(paths[1], "f")) != "second" {
		t.Error("one backup overwrote the other")
	}
	if filepath.Base(paths[1]) != "20261015-101530-2-out.backup" {
		t.Errorf("collision name = %q, want a counter after the timestamp", filepath.Base(paths[1]))
	}
}

func TestCopyAside_CopiesTheWholeFolderAndLeavesTheSourceUntouched(t *testing.T) {
	home := t.TempDir()
	src := filepath.Join(home, "agents")
	put(t, filepath.Join(src, "mine.json"), "mine", 0o600)
	put(t, filepath.Join(src, "nested", "run.sh"), "echo", 0o700)
	if err := os.Symlink("mine.json", filepath.Join(src, "link.json")); err != nil {
		t.Fatal(err)
	}

	got, err := CopyAside(src, "20261015-101530")
	if err != nil {
		t.Fatalf("CopyAside: %v", err)
	}

	if want := filepath.Join(home, "20261015-101530-agents.backup"); got != want {
		t.Errorf("backup path = %q, want %q", got, want)
	}
	if read(t, filepath.Join(got, "mine.json")) != "mine" || read(t, filepath.Join(got, "nested", "run.sh")) != "echo" {
		t.Error("the copy is missing files")
	}
	if info, err := os.Stat(filepath.Join(got, "nested", "run.sh")); err != nil || info.Mode().Perm() != 0o700 {
		t.Errorf("copied script mode = %v (%v), want 700", info.Mode().Perm(), err)
	}
	if target, err := os.Readlink(filepath.Join(got, "link.json")); err != nil || target != "mine.json" {
		t.Errorf("symlink in the copy = %q (%v), want it kept as a link to mine.json", target, err)
	}
	if read(t, filepath.Join(src, "mine.json")) != "mine" {
		t.Error("the source folder was modified")
	}
}

func TestCopyAside_SameSecondCollisionGetsACounter(t *testing.T) {
	home := t.TempDir()
	src := filepath.Join(home, "prompts")
	put(t, filepath.Join(src, "a.md"), "a", 0o600)

	first, err := CopyAside(src, "20261015-101530")
	if err != nil {
		t.Fatal(err)
	}
	second, err := CopyAside(src, "20261015-101530")
	if err != nil {
		t.Fatal(err)
	}

	if first == second || filepath.Base(second) != "20261015-101530-2-prompts.backup" {
		t.Errorf("paths = %q and %q, want two distinct backups with a counter on the second", first, second)
	}
}

func TestDiffers(t *testing.T) {
	dir := t.TempDir()
	same := filepath.Join(dir, "same")
	other := filepath.Join(dir, "other")
	put(t, same, "x", 0o600)
	put(t, other, "y", 0o600)
	if err := os.Symlink(same, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		data string
		want bool
	}{
		{"missing file is not a change to protect", filepath.Join(dir, "missing"), "x", false},
		{"identical content", same, "x", false},
		{"different content", other, "x", true},
		{"symbolic link", filepath.Join(dir, "link"), "x", true},
		{"directory in the way", dir, "x", true},
	}
	for _, tc := range tests {
		if got := Differs(tc.path, []byte(tc.data)); got != tc.want {
			t.Errorf("%s: Differs = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTreesEqual(t *testing.T) {
	build := func(files map[string]string, exec ...string) string {
		root := t.TempDir()
		for rel, content := range files {
			put(t, filepath.Join(root, rel), content, 0o600)
		}
		for _, rel := range exec {
			if err := os.Chmod(filepath.Join(root, rel), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	base := map[string]string{"agents/a.json": "a", "skills/s/SKILL.md": "s"}

	tests := []struct {
		name  string
		other string
		want  bool
	}{
		{"identical", build(base), true},
		{"a byte differs", build(map[string]string{"agents/a.json": "A", "skills/s/SKILL.md": "s"}), false},
		{"extra file", build(map[string]string{"agents/a.json": "a", "skills/s/SKILL.md": "s", "notes.txt": "n"}), false},
		{"missing file", build(map[string]string{"agents/a.json": "a"}), false},
		{"exec bit differs", build(base, "agents/a.json"), false},
	}
	for _, tc := range tests {
		got, err := TreesEqual(build(base), tc.other)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: TreesEqual = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCopyAside_CopiesASingleFile(t *testing.T) {
	home := t.TempDir()
	file := filepath.Join(home, "settings.json")
	put(t, file, "mine", 0o600)

	got, err := CopyAside(file, "20261015-101530")
	if err != nil {
		t.Fatalf("CopyAside: %v", err)
	}

	if want := filepath.Join(home, "20261015-101530-settings.json.backup"); got != want {
		t.Errorf("backup path = %q, want %q", got, want)
	}
	if read(t, got) != "mine" || read(t, file) != "mine" {
		t.Error("the file or its copy has the wrong content")
	}
}

func TestCopyAside_RemovesAHalfWrittenBackupWhenTheCopyFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read an unreadable file")
	}
	home := t.TempDir()
	src := filepath.Join(home, "agents")
	put(t, filepath.Join(src, "a-readable.json"), "ok", 0o600)
	unreadable := filepath.Join(src, "z-unreadable.json")
	put(t, unreadable, "secret", 0o600)
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o600) })

	got, err := CopyAside(src, "20261015-101530")

	if err == nil {
		t.Fatal("CopyAside succeeded although a file could not be read")
	}
	if got != "" {
		t.Errorf("returned backup path %q for a failed copy", got)
	}
	if matches, _ := filepath.Glob(filepath.Join(home, "*.backup")); len(matches) != 0 {
		t.Errorf("a partial backup was left behind: %v", matches)
	}
}
