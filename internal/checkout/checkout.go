// Package checkout holds the rule for branch and tag names that kirokit hands
// to git, and a read-only view of a config checkout's refs.
package checkout

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
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
	r := reader{git: gitPath, dir: dir}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			r.env = append(r.env, kv)
		}
	}
	current, err := r.current()
	if err != nil {
		return Refs{}, fmt.Errorf("read the checkout in %s: %w", dir, err)
	}
	branches, err := r.lines("for-each-ref", "--format=%(refname:lstrip=3)", "refs/remotes/origin/")
	if err != nil {
		return Refs{}, fmt.Errorf("list the branches in %s: %w", dir, err)
	}
	branches = slices.DeleteFunc(branches, func(b string) bool { return b == "HEAD" })
	sort.Strings(branches)
	tags, err := r.lines("tag", "--list", "--sort=-v:refname")
	if err != nil {
		return Refs{}, fmt.Errorf("list the tags in %s: %w", dir, err)
	}
	refs := Refs{Current: current, Branches: branches, Tags: tags}
	if i := slices.IndexFunc(tags, func(t string) bool { return strings.HasPrefix(t, "v") }); i >= 0 {
		refs.Newest = tags[i]
	}
	return refs, nil
}

// reader runs git on one checkout.
type reader struct {
	git, dir string
	env      []string
}

func (r reader) run(args ...string) (string, error) {
	cmd := exec.Command(r.git, append([]string{"-C", r.dir}, args...)...)
	cmd.Env = r.env
	hideWindow(cmd)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// lines runs git and returns its non-empty output lines, never nil.
func (r reader) lines(args ...string) ([]string, error) {
	out, err := r.run(args...)
	lines := []string{}
	for l := range strings.SplitSeq(out, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, err
}

// current is what HEAD is on: a branch, else the newest tag at HEAD, else
// only its commit. The branch is cut from its full ref, since git's short form
// turns into "heads/<name>" when a tag has the same name.
func (r reader) current() (Ref, error) {
	commit, err := r.run("rev-parse", "--short", "HEAD")
	if err != nil {
		return Ref{}, err
	}
	if full, err := r.run("symbolic-ref", "-q", "HEAD"); err == nil {
		return Ref{Kind: "branch", Name: strings.TrimPrefix(full, "refs/heads/"), Commit: commit}, nil
	}
	if at, err := r.lines("tag", "--points-at", "HEAD", "--sort=-v:refname"); err == nil && len(at) > 0 {
		return Ref{Kind: "tag", Name: at[0], Commit: commit}, nil
	}
	return Ref{Commit: commit}, nil
}
