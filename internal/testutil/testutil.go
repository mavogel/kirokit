// Package testutil holds test helpers for behaviour that differs by OS.
package testutil

import (
	"os"
	"runtime"
	"testing"
)

// ExecBitsKept is false on Windows, whose file modes carry no execute bit
// (nor any other Unix permission bit, such as 0600's group and other bits).
const ExecBitsKept = runtime.GOOS != "windows"

// SkipWithoutExecBits skips a test that checks a file's execute bit: Windows
// keeps none, so there is nothing to check there.
func SkipWithoutExecBits(t testing.TB) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Windows keeps no execute bit")
	}
}

// Symlink creates newname as a symbolic link to oldname. On Windows, where
// creating one needs Developer Mode or elevation, it skips the test when the
// link cannot be made; elsewhere a failure fails the test.
func Symlink(t testing.TB, oldname, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symbolic links are not permitted here: %v", err)
		}
		t.Fatal(err)
	}
}
