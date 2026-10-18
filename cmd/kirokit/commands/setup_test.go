package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetup_ClonesSourceAndChecksOutTheNewestTag(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v1.10.0")
	commitFile(t, origin, "three") // untagged tip of main

	target := filepath.Join(t.TempDir(), "repo")
	chdir(t, t.TempDir())

	got, err := runRoot(t, "setup", "--source", origin, "--target", target)
	if err != nil {
		t.Fatalf("setup: %v\n%s", err, got)
	}
	if head, want := git(t, target, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "v1.10.0^{commit}"); head != want {
		t.Errorf("HEAD = %s, want v1.10.0 commit %s", head, want)
	}
	if !strings.Contains(got, "checked out v1.10.0") || !strings.Contains(got, "KIROKIT_REPO="+target) {
		t.Errorf("output = %q, want the checked-out tag and the KIROKIT_REPO hint", got)
	}
}

func TestSetup_KeepsTheCloneWhenTheSourceHasNoRelease(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")

	target := filepath.Join(t.TempDir(), "repo")
	chdir(t, t.TempDir())

	got, err := runRoot(t, "setup", "--source", origin, "--target", target)
	if err != nil {
		t.Fatalf("setup: %v\n%s", err, got)
	}
	if _, err := os.Stat(filepath.Join(target, ".git")); err != nil {
		t.Errorf("clone was not kept: %v", err)
	}
	if !strings.Contains(got, "no release yet") {
		t.Errorf("output = %q, want the no-release notice", got)
	}
}

func TestSetup_BranchChecksOutThatBranchOrTag(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	git(t, origin, "branch", "feature")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v2.0.0")

	for ref, want := range map[string]string{
		"feature": git(t, origin, "rev-parse", "feature"),
		"v1.0.0":  git(t, origin, "rev-parse", "v1.0.0^{commit}"),
		"main":    git(t, origin, "rev-parse", "main"),
	} {
		t.Run(ref, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "repo")
			chdir(t, t.TempDir())
			got, err := runRoot(t, "setup", "--source", origin, "--target", target, "--branch", ref)
			if err != nil {
				t.Fatalf("setup --branch %s: %v\n%s", ref, err, got)
			}
			if head := git(t, target, "rev-parse", "HEAD"); head != want {
				t.Errorf("HEAD = %s, want %s", head, want)
			}
			if !strings.Contains(got, "checked out "+ref) {
				t.Errorf("output = %q, want it to name %s", got, ref)
			}
		})
	}
}

func TestSetup_BranchRefusesBadOrUnknownRef(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	for _, ref := range []string{"--upload-pack=x", "-b", "has space", "nonexistent"} {
		t.Run(ref, func(t *testing.T) {
			chdir(t, t.TempDir())
			if got, err := runRoot(t, "setup", "--source", origin, "--target", filepath.Join(t.TempDir(), "r"), "--branch", ref); err == nil {
				t.Fatalf("setup --branch %q succeeded:\n%s", ref, got)
			}
		})
	}
}

func TestSetup_RefusesBadInput(t *testing.T) {
	existing := t.TempDir()
	if err := os.WriteFile(filepath.Join(existing, "keep"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]string{
		"option as source": {"--source", "--upload-pack=x", "--target", filepath.Join(t.TempDir(), "r")},
		"relative source":  {"--source", "some/dir", "--target", filepath.Join(t.TempDir(), "r")},
		"non-empty target": {"--source", t.TempDir(), "--target", existing},
		"missing --source": {"--target", filepath.Join(t.TempDir(), "r")},
		"missing --target": {"--source", t.TempDir()},
		"target is a file": {"--source", t.TempDir(), "--target", filepath.Join(existing, "keep")},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			chdir(t, t.TempDir())
			if got, err := runRoot(t, append([]string{"setup"}, args...)...); err == nil {
				t.Fatalf("setup %v succeeded:\n%s", args, got)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(existing, "keep")); err != nil {
		t.Errorf("existing target content was touched: %v", err)
	}
}
