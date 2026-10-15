package externalskills

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/mavogel/kiro-cli-kit/internal/manifest"
)

// Fetch is the only networked operation: it resolves every entry's ref to a
// commit with the machine's git (so SSH keys, credential helpers and the like
// authenticate private sources), extracts the skill directory verbatim, and
// only when every entry succeeded replaces the vendored files and the lock.
// A failure leaves the team directory exactly as it was. Vendored skills with
// no entry are removed.
func Fetch(teamDir string, entries []manifest.ExternalSkill) (lock Lock, err error) {
	// Fetch is the one place user-controlled strings reach git, so it checks
	// each entry itself instead of trusting that the manifest loader did.
	for _, e := range entries {
		if err := e.Validate(); err != nil {
			return Lock{}, fmt.Errorf("external skill %q: %w", e.EffectiveName(), err)
		}
	}

	// git runs in a scratch directory, so every path handed to it must be
	// absolute.
	teamDir, err = filepath.Abs(teamDir)
	if err != nil {
		return Lock{}, fmt.Errorf("externalskills: %w", err)
	}
	externalDir := filepath.Join(teamDir, "external")
	_, statErr := os.Stat(externalDir)
	existed := statErr == nil
	if len(entries) == 0 && !existed {
		return Lock{}, nil
	}

	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return Lock{}, fmt.Errorf("externalskills: %w", err)
	}
	// A failed first fetch must not leave an external/ directory behind.
	defer func() {
		if err != nil && !existed {
			_ = os.RemoveAll(externalDir)
		}
	}()
	staging, err := os.MkdirTemp(externalDir, ".fetch-")
	if err != nil {
		return Lock{}, fmt.Errorf("externalskills: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	claimed := make(map[string]bool) // skill names already staged by an earlier entry
	for _, e := range entries {
		les, err := stage(e, staging, claimed)
		if err != nil {
			return Lock{}, fmt.Errorf("external skill %q (%s @ %s): %w", e.EffectiveName(), e.Source, e.Ref, err)
		}
		lock.Skills = append(lock.Skills, les...)
	}

	skills := SkillsDir(teamDir)
	if err := os.MkdirAll(skills, 0o755); err != nil {
		return Lock{}, fmt.Errorf("externalskills: %w", err)
	}
	keep := make(map[string]bool, len(lock.Skills))
	for _, le := range lock.Skills {
		name := le.Name
		keep[name] = true
		dest := filepath.Join(skills, name)
		if err := os.RemoveAll(dest); err != nil {
			return Lock{}, fmt.Errorf("externalskills: %w", err)
		}
		if err := os.Rename(filepath.Join(staging, name), dest); err != nil {
			return Lock{}, fmt.Errorf("externalskills: install %s: %w", name, err)
		}
	}
	existing, err := os.ReadDir(skills)
	if err != nil {
		return Lock{}, fmt.Errorf("externalskills: %w", err)
	}
	for _, d := range existing {
		if !keep[d.Name()] {
			if err := os.RemoveAll(filepath.Join(skills, d.Name())); err != nil {
				return Lock{}, fmt.Errorf("externalskills: prune %s: %w", d.Name(), err)
			}
		}
	}
	if err := WriteLock(LockPath(teamDir), lock); err != nil {
		return Lock{}, err
	}
	return lock, nil
}

// stage fetches one entry's repository into a scratch repo and checks the
// vendored skill(s) out under stagingRoot/<name>. A single-skill entry yields
// one lock entry; an all: true entry yields one per skill directory under its
// path.
func stage(e manifest.ExternalSkill, stagingRoot string, claimed map[string]bool) ([]LockEntry, error) {
	work, err := os.MkdirTemp("", "kirokit-fetch-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(work) }()

	if _, err := runGit(work, "init", "-q"); err != nil {
		return nil, err
	}
	commit, err := resolve(work, e)
	if err != nil {
		return nil, err
	}

	// read-tree + checkout-index rather than `git archive`: archive honours
	// the source repository's export-ignore and would silently drop files.
	if _, err := runGit(work, "read-tree", commit+":"+e.Path); err != nil {
		return nil, fmt.Errorf("path %q not found as a directory in commit %s", e.Path, commit)
	}

	if !e.IsCollection() {
		name := e.EffectiveName()
		if err := claim(claimed, name); err != nil {
			return nil, err
		}
		le, err := checkoutSkill(work, filepath.Join(stagingRoot, name), LockEntry{
			Name: name, Source: e.Source, Ref: e.Ref, Path: e.Path, Commit: commit,
		})
		if err != nil {
			return nil, err
		}
		return []LockEntry{le}, nil
	}
	return stageCollection(work, stagingRoot, e, commit, claimed)
}

// claim reserves a skill name for the entry being staged; two entries (or a
// collection and an entry) providing the same name is an error whichever of
// them is staged first.
func claim(claimed map[string]bool, name string) error {
	if claimed[name] {
		return fmt.Errorf("skill %q is provided by more than one manifest entry", name)
	}
	claimed[name] = true
	return nil
}

// checkoutSkill checks the index out into dest and completes le with the
// content hash. The directory must hold a SKILL.md at its root.
func checkoutSkill(work, dest string, le LockEntry) (LockEntry, error) {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return le, err
	}
	if _, err := runGit(work, "-c", "core.autocrlf=false", "checkout-index", "-a", "-f", "--prefix="+dest+string(filepath.Separator)); err != nil {
		return le, err
	}
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		return le, fmt.Errorf("%q has no SKILL.md at its root", le.Path)
	}
	sum, err := HashDir(dest)
	if err != nil {
		return le, err
	}
	le.SHA256 = sum
	return le, nil
}

// stageCollection checks out the whole collection directory, then keeps each
// immediate subdirectory that holds a SKILL.md, minus the entry's excludes.
func stageCollection(work, stagingRoot string, e manifest.ExternalSkill, commit string, claimed map[string]bool) ([]LockEntry, error) {
	tree, err := os.MkdirTemp(stagingRoot, ".tree-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tree) }()
	if _, err := runGit(work, "-c", "core.autocrlf=false", "checkout-index", "-a", "-f", "--prefix="+tree+string(filepath.Separator)); err != nil {
		return nil, err
	}

	children, err := os.ReadDir(tree)
	if err != nil {
		return nil, err
	}
	isDir := make(map[string]bool, len(children))
	for _, c := range children {
		if c.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("%q is a symbolic link; symbolic links are not vendored", c.Name())
		}
		isDir[c.Name()] = c.IsDir()
	}
	// A misspelled exclude would silently vendor a skill the user meant to skip.
	for _, x := range e.Exclude {
		if !isDir[x] {
			return nil, fmt.Errorf("exclude names %q, which is not a directory under %q", x, e.Path)
		}
	}

	var les []LockEntry
	for _, c := range children {
		name := c.Name()
		if !c.IsDir() || slices.Contains(e.Exclude, name) {
			continue
		}
		if _, err := os.Stat(filepath.Join(tree, name, "SKILL.md")); err != nil {
			continue // not a skill (for example a docs or template directory)
		}
		if err := claim(claimed, name); err != nil {
			return nil, err
		}
		dest := filepath.Join(stagingRoot, name)
		if err := os.Rename(filepath.Join(tree, name), dest); err != nil {
			return nil, err
		}
		sum, err := HashDir(dest)
		if err != nil {
			return nil, fmt.Errorf("skill %q: %w", name, err)
		}
		les = append(les, LockEntry{
			Name: name, Source: e.Source, Ref: e.Ref, Path: e.Path + "/" + name,
			Commit: commit, SHA256: sum, From: e.Path,
		})
	}
	if len(les) == 0 {
		return nil, fmt.Errorf("no skills found under %q (a skill is a subdirectory with a SKILL.md, and none is left after exclude)", e.Path)
	}
	sort.Slice(les, func(i, j int) bool { return les[i].Name < les[j].Name })
	return les, nil
}

// resolve fetches e.Ref into the scratch repo at work and returns the commit
// it names. It tries a shallow fetch first; hosts that refuse a want of an
// arbitrary commit fall back to fetching all heads and tags.
func resolve(work string, e manifest.ExternalSkill) (string, error) {
	if _, err := runGit(work, "fetch", "--quiet", "--depth", "1", "--", e.Source, e.Ref); err == nil {
		return runGit(work, "rev-parse", "FETCH_HEAD^{commit}")
	}

	if _, err := runGit(work, "fetch", "--quiet", "--", e.Source,
		"+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"); err != nil {
		return "", fmt.Errorf("%w\n(authentication is handled by git: use an SSH URL or a git credential helper)", err)
	}
	for _, candidate := range []string{e.Ref, "refs/remotes/origin/" + e.Ref, "refs/tags/" + e.Ref} {
		if commit, err := runGit(work, "rev-parse", "--verify", "--quiet", candidate+"^{commit}"); err == nil {
			return commit, nil
		}
	}
	return "", fmt.Errorf("ref %q not found in %s", e.Ref, e.Source)
}

// runGit runs git in dir with the caller's environment and stdin, so the
// machine's own authentication applies. Stderr is captured into the error.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
