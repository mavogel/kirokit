package dirlink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMake_LinksAFolderThatIsThenRecognisedAndRemovable(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "kb")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a.md"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mount := filepath.Join(dir, "mount")

	if err := Make(source, mount); err != nil {
		t.Fatalf("Make: %v", err)
	}
	if !Is(mount) {
		t.Fatalf("Is(%s) = false after Make", mount)
	}
	if got, err := Target(mount); err != nil || filepath.Clean(got) != source {
		t.Errorf("Target = %q (%v), want %q", got, err, source)
	}
	if b, err := os.ReadFile(filepath.Join(mount, "a.md")); err != nil || string(b) != "a\n" {
		t.Errorf("reading through the link = %q (%v)", b, err)
	}

	if err := os.Remove(mount); err != nil {
		t.Fatalf("remove link: %v", err)
	}
	if _, err := os.Stat(filepath.Join(source, "a.md")); err != nil {
		t.Errorf("removing the link removed the source: %v", err)
	}
}

func TestIs_FalseForAnythingButALink(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{dir, file, filepath.Join(dir, "missing")} {
		if Is(p) {
			t.Errorf("Is(%s) = true, want false", p)
		}
	}
}

func TestMake_RefusesAnExistingMountPoint(t *testing.T) {
	dir := t.TempDir()
	if err := Make(dir, dir); err == nil {
		t.Fatal("Make over an existing folder succeeded")
	}
}
