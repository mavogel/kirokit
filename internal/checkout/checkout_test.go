package checkout

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCheckRefName(t *testing.T) {
	for _, bad := range []string{"", "-b", "--upload-pack=x", "has space", "a\tb", "a\nb"} {
		if err := CheckRefName(bad); err == nil {
			t.Errorf("CheckRefName(%q) accepted, want refused", bad)
		}
	}
	for _, ok := range []string{"main", "feature/x", "v1.2.0", "release-1"} {
		if err := CheckRefName(ok); err != nil {
			t.Errorf("CheckRefName(%q) = %v, want accepted", ok, err)
		}
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "core.hooksPath=" + os.DevNull, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	git(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
}

func TestRead(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commit(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	git(t, origin, "tag", "nightly")
	git(t, origin, "branch", "feature/x")
	commit(t, origin, "two")
	git(t, origin, "tag", "-a", "-m", "annotated", "v1.10.0")
	commit(t, origin, "three")

	clone := filepath.Join(t.TempDir(), "clone")
	git(t, t.TempDir(), "clone", "-q", origin, clone)
	// The read uses local refs only: the origin may be gone.
	if err := os.RemoveAll(origin); err != nil {
		t.Fatal(err)
	}
	// A foreign GIT_DIR must not redirect the read.
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "elsewhere"))

	short := func() string { return git(t, clone, "rev-parse", "--short", "HEAD") }
	got, err := Read("git", clone)
	if err != nil {
		t.Fatal(err)
	}
	want := Refs{
		Current:  Ref{Kind: "branch", Name: "main", Commit: short()},
		Branches: []string{"feature/x", "main"},
		Tags:     []string{"v1.10.0", "v1.0.0", "nightly"},
		Newest:   "v1.10.0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("on main: Read = %+v, want %+v", got, want)
	}

	for _, tc := range []struct{ at, kind, name string }{
		{"v1.0.0", "tag", "v1.0.0"},   // two tags here: the newest by version wins
		{"v1.10.0", "tag", "v1.10.0"}, // an annotated tag
		{"main", "", ""},              // main's tip has no tag
	} {
		git(t, clone, "checkout", "-q", "--detach", tc.at)
		got, err := Read("git", clone)
		if err != nil {
			t.Fatal(err)
		}
		if want := (Ref{Kind: tc.kind, Name: tc.name, Commit: short()}); got.Current != want {
			t.Errorf("detached at %s: Current = %+v, want %+v", tc.at, got.Current, want)
		}
	}
}

func TestRead_ACloneWithoutTagsHasEmptyLists(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commit(t, origin, "one")
	clone := filepath.Join(t.TempDir(), "clone")
	git(t, t.TempDir(), "clone", "-q", origin, clone)

	got, err := Read("git", clone)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tags == nil || len(got.Tags) != 0 || got.Newest != "" || !reflect.DeepEqual(got.Branches, []string{"main"}) {
		t.Errorf("Read = %+v, want branches [main], tags [] and no newest", got)
	}
}

func TestRead_RefusesAFolderThatIsNoCheckout(t *testing.T) {
	for name, dir := range map[string]string{
		"missing": filepath.Join(t.TempDir(), "gone"),
		"empty":   t.TempDir(),
	} {
		if got, err := Read("git", dir); err == nil {
			t.Errorf("%s folder: Read = %+v, want an error", name, got)
		}
	}
}
