//go:build windows

package backup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mavogel/kirokit/internal/dirlink"
)

// A knowledge-base junction inside a backed-up folder is copied as a link, not
// as the files it points at.
func TestCopyAside_KeepsAJunctionAsALink(t *testing.T) {
	home := t.TempDir()
	kb := filepath.Join(home, "kb")
	put(t, filepath.Join(kb, "doc.md"), "doc", 0o600)
	src := filepath.Join(home, "agents")
	put(t, filepath.Join(src, "a.json"), "a", 0o600)
	if out, err := exec.Command("cmd.exe", "/C", "mklink", "/J", filepath.Join(src, "kb"), kb).CombinedOutput(); err != nil {
		t.Fatalf("mklink /J: %v: %s", err, out)
	}

	got, err := CopyAside(src, "20261015-101530")
	if err != nil {
		t.Fatalf("CopyAside: %v", err)
	}
	if !dirlink.Is(filepath.Join(got, "kb")) {
		t.Fatal("the junction was not copied as a link")
	}
	if target, err := dirlink.Target(filepath.Join(got, "kb")); err != nil || filepath.Clean(target) != kb {
		t.Errorf("copied link -> %q (%v), want %q", target, err, kb)
	}
	if _, err := os.Stat(filepath.Join(kb, "doc.md")); err != nil {
		t.Errorf("the source the junction points at changed: %v", err)
	}
}
