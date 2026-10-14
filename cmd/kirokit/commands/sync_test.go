package commands

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

func commitFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/f", []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "f")
	git(t, dir, "commit", "-q", "-m", content)
}

func TestUpdateCheckout(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	git(t, origin, "tag", "v1.0.0")
	commitFile(t, origin, "two")
	git(t, origin, "tag", "v1.10.0")
	commitFile(t, origin, "three") // untagged tip of main

	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	chdir(t, clone)

	commitFile(t, origin, "four")
	if err := updateCheckout(false); err != nil {
		t.Fatalf("branch update: %v", err)
	}
	if got, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "HEAD"); got != want {
		t.Errorf("branch update: HEAD = %s, want origin tip %s", got, want)
	}

	if err := updateCheckout(true); err != nil {
		t.Fatalf("latest-tag update: %v", err)
	}
	// v1.10.0 must beat v1.0.0 (version sort, not lexical).
	if got, want := git(t, clone, "rev-parse", "HEAD"), git(t, origin, "rev-parse", "v1.10.0^{commit}"); got != want {
		t.Errorf("latest-tag: HEAD = %s, want v1.10.0 commit %s", got, want)
	}
}

func TestUpdateCheckout_LatestTagWithoutTagsFails(t *testing.T) {
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "main")
	commitFile(t, origin, "one")
	clone := t.TempDir()
	git(t, clone, "clone", "-q", origin, ".")
	chdir(t, clone)

	if err := updateCheckout(true); err == nil || !strings.Contains(err.Error(), "no v* tag") {
		t.Fatalf("err = %v, want a %q error", err, "no v* tag")
	}
}
