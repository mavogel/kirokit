package externalskills

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/mavogel/kirokit/internal/manifest"
)

// Verify checks, offline, that a team's manifest entries, lock file and
// vendored files agree. It returns one error listing every disagreement; each
// line names the skill and how to repair it. A team with no entries and no
// external/ directory is always consistent.
func Verify(teamDir string, entries []manifest.ExternalSkill) error {
	_, err := Resolve(teamDir, entries)
	return err
}

// Resolve runs the checks of Verify and returns the sorted names of the
// skills that are vendored: every single-skill entry plus every lock entry
// that an all: true entry produced. Collection entries list no skill names in
// the manifest, so the lock is where those names come from.
func Resolve(teamDir string, entries []manifest.ExternalSkill) ([]string, error) {
	lock, err := ReadLock(LockPath(teamDir))
	if err != nil {
		return nil, err
	}
	hint := fmt.Sprintf("run `kirokit skills fetch %s`", filepath.Base(teamDir))

	var problems []string
	add := func(name, format string, args ...any) {
		problems = append(problems, fmt.Sprintf("external skill %q: %s; %s", name, fmt.Sprintf(format, args...), hint))
	}

	byName := make(map[string][]LockEntry, len(lock.Skills))
	for _, le := range lock.Skills {
		byName[le.Name] = append(byName[le.Name], le)
	}

	claimed := make(map[string]int) // skill name -> number of manifest entries providing it
	used := make(map[string]bool)   // lock entries accounted for, by name

	for _, e := range entries {
		if e.IsCollection() {
			resolveCollection(teamDir, e, lock, claimed, used, add)
			continue
		}
		name := e.EffectiveName()
		claimed[name]++
		les := byName[name]
		if len(les) == 0 {
			add(name, "declared in manifest.yaml but missing from the lock file")
			continue
		}
		used[name] = true
		le := les[0]
		switch {
		case le.From != "":
			add(name, "the lock file records it as part of the %q collection, not as its own entry", le.From)
		case le.Source != e.Source:
			add(name, "source in manifest.yaml differs from the lock file")
		case le.Ref != e.Ref:
			add(name, "ref in manifest.yaml (%s) differs from the lock file (%s)", e.Ref, le.Ref)
		case le.Path != e.Path:
			add(name, "path in manifest.yaml differs from the lock file")
		default:
			checkVendored(teamDir, le, add)
		}
	}

	for name, n := range claimed {
		if n > 1 {
			add(name, "provided by more than one manifest entry")
		}
	}
	for _, le := range lock.Skills {
		if !used[le.Name] {
			add(le.Name, "in the lock file but not declared in manifest.yaml")
		}
	}
	if dirs, err := os.ReadDir(SkillsDir(teamDir)); err == nil {
		for _, d := range dirs {
			if d.IsDir() && claimed[d.Name()] == 0 {
				add(d.Name(), "vendored directory has no entry in manifest.yaml")
			}
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, fmt.Errorf("%s", strings.Join(problems, "\n"))
	}
	names := make([]string, 0, len(claimed))
	for name := range claimed {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// resolveCollection checks one all: true entry against the lock entries it
// produced (matching collection path, source and ref).
func resolveCollection(teamDir string, e manifest.ExternalSkill, lock Lock, claimed map[string]int, used map[string]bool, add func(string, string, ...any)) {
	var produced, sameCollection []LockEntry
	for _, le := range lock.Skills {
		if le.From != e.Path {
			continue
		}
		sameCollection = append(sameCollection, le)
		if le.Source == e.Source && le.Ref == e.Ref {
			produced = append(produced, le)
		}
	}
	switch {
	case len(sameCollection) == 0:
		add(e.Path, "collection declared in manifest.yaml but missing from the lock file")
		return
	case len(produced) == 0:
		add(e.Path, "source or ref of the collection in manifest.yaml (%s) differs from the lock file (%s)", e.Ref, sameCollection[0].Ref)
		// Those lock entries and their files belong to this collection; do not
		// report them a second time as undeclared.
		for _, le := range sameCollection {
			claimed[le.Name]++
			used[le.Name] = true
		}
		return
	}
	for _, le := range produced {
		claimed[le.Name]++
		used[le.Name] = true
		if slices.Contains(e.Exclude, le.Name) {
			add(le.Name, "vendored from the %q collection but now listed in exclude", e.Path)
			continue
		}
		checkVendored(teamDir, le, add)
	}
}

// checkVendored verifies that a lock entry's vendored directory exists, has a
// SKILL.md, and hashes to the locked value.
func checkVendored(teamDir string, le LockEntry, add func(string, string, ...any)) {
	checkVendoredOn(runtime.GOOS, teamDir, le, add)
}

func checkVendoredOn(goos, teamDir string, le LockEntry, add func(string, string, ...any)) {
	dir := filepath.Join(SkillsDir(teamDir), le.Name)
	if _, err := os.Stat(dir); err != nil {
		add(le.Name, "vendored files are missing")
	} else if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		add(le.Name, "vendored files have no SKILL.md")
	} else if sum, err := vendoredHash(goos, dir, le); err != nil {
		add(le.Name, "%v", err)
	} else if sum != le.SHA256 && goos == "windows" && len(le.Executables) == 0 {
		add(le.Name, "vendored content does not match the lock file, or the lock predates recorded execute bits (Windows keeps none, so the lock must list them)")
	} else if sum != le.SHA256 {
		add(le.Name, "vendored content does not match the lock file")
	}
}

// vendoredHash hashes a vendored skill as the lock does. Windows keeps no
// execute bit, so there the bits come from the lock entry.
func vendoredHash(goos, dir string, le LockEntry) (string, error) {
	if goos == "windows" {
		return hashDirWithExecutables(dir, le.Executables)
	}
	return HashDir(dir)
}
