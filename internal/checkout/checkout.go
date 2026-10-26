// Package checkout holds the rule for branch and tag names that kirokit hands
// to git, and a read-only view of a config checkout's refs.
package checkout

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// errRefName is the refusal of a name CheckRefName rejects; callers prefix it
// with the command and flag, as in "kirokit sync: --branch: ...".
var errRefName = errors.New("must be a branch or tag name, not empty, starting with '-' or containing whitespace")

// CheckRefName refuses an empty name, one starting with '-' (git would read it
// as an option), or one with whitespace. Anything else is for git to find or
// not.
func CheckRefName(name string) error {
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, " \t\r\n") {
		return errRefName
	}
	return nil
}

// Ref is what a checkout's HEAD is on.
type Ref struct {
	Kind   string `json:"kind"` // "branch", "tag", or "" (detached on no tag)
	Name   string `json:"name"`
	Commit string `json:"commit"` // short hash
}

// Refs are a checkout's current version and the versions it can switch to.
type Refs struct {
	Current  Ref      `json:"current"`
	Branches []string `json:"branches"` // refs/remotes/origin/*, without HEAD, sorted by name
	Tags     []string `json:"tags"`     // refs/tags/*, newest version first
	Newest   string   `json:"newest"`   // first v* tag in that order, "" when none
}

// Read lists the refs of the checkout at dir with git at gitPath. Only local
// refs are read, which every sync's fetch refreshes, so it never touches the
// network or meets a sign-in prompt. GIT_* variables are dropped so git reads
// dir itself. A missing or broken checkout returns an error.
func Read(gitPath, dir string) (Refs, error) {
	env := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	run := func(args ...string) (string, error) {
		cmd := exec.Command(gitPath, append([]string{"-C", dir}, args...)...)
		cmd.Env = env
		hideWindow(cmd)
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}
	lines := func(s string) []string {
		out := []string{}
		for l := range strings.SplitSeq(s, "\n") {
			if l = strings.TrimSpace(l); l != "" {
				out = append(out, l)
			}
		}
		return out
	}

	commit, err := run("rev-parse", "--short", "HEAD")
	if err != nil {
		return Refs{}, fmt.Errorf("read the checkout in %s: %w", dir, err)
	}
	refs := Refs{Current: Ref{Commit: commit}, Branches: []string{}, Tags: []string{}}
	// The branch is cut from its full ref: git's short form turns into
	// "heads/<name>" when a tag has the same name.
	if full, err := run("symbolic-ref", "-q", "HEAD"); err == nil {
		refs.Current.Kind, refs.Current.Name = "branch", strings.TrimPrefix(full, "refs/heads/")
	} else if at, err := run("tag", "--points-at", "HEAD", "--sort=-v:refname"); err == nil && at != "" {
		refs.Current.Kind, refs.Current.Name = "tag", lines(at)[0]
	}

	branches, err := run("for-each-ref", "--format=%(refname:lstrip=3)", "refs/remotes/origin/")
	if err != nil {
		return Refs{}, fmt.Errorf("list the branches in %s: %w", dir, err)
	}
	for _, b := range lines(branches) {
		if b != "HEAD" {
			refs.Branches = append(refs.Branches, b)
		}
	}
	sort.Strings(refs.Branches)

	tags, err := run("tag", "--list", "--sort=-v:refname")
	if err != nil {
		return Refs{}, fmt.Errorf("list the tags in %s: %w", dir, err)
	}
	refs.Tags = lines(tags)
	for _, tag := range refs.Tags {
		if strings.HasPrefix(tag, "v") {
			refs.Newest = tag
			break
		}
	}
	return refs, nil
}
