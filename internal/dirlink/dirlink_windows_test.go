//go:build windows

package dirlink

import (
	"os"
	"path/filepath"
	"testing"
)

// A junction is what Make falls back to without the symlink privilege; it must
// be recognised and removed like a symlink, never followed into its source.
func TestJunction_IsRecognisedAndRemovingItKeepsTheSource(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "kb")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a.md"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mount := filepath.Join(dir, "mount")
	if err := makeJunction(source, mount); err != nil {
		t.Fatalf("makeJunction: %v", err)
	}
	if !Is(mount) {
		t.Fatal("Is = false for a junction")
	}
	if got, err := Target(mount); err != nil || filepath.Clean(got) != source {
		t.Errorf("Target = %q (%v), want %q", got, err, source)
	}
	if err := os.Remove(mount); err != nil {
		t.Fatalf("remove junction: %v", err)
	}
	if _, err := os.Stat(filepath.Join(source, "a.md")); err != nil {
		t.Errorf("removing the junction removed the source: %v", err)
	}
}

func TestMakeJunction_RefusesAQuoteInAPath(t *testing.T) {
	dir := t.TempDir()
	if err := makeJunction(dir, filepath.Join(dir, `a"b`)); err == nil {
		t.Fatal(`makeJunction accepted a path with '"'`)
	}
}
