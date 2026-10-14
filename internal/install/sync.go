package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sync installs team like Install, then removes every path the team's
// previous install manifest recorded that the new install no longer writes
// (a template or secrets file deleted from the repository). Install alone
// leaves such files behind. It returns the removed paths. A missing previous
// manifest is a first install, so nothing is pruned.
func Sync(roots Roots, team, kiroHome string, kb *KnowledgeBase) ([]string, error) {
	previous, err := readManifest(installManifestPath(kiroHome, team))
	if err != nil {
		return nil, err
	}

	if err := Install(roots, team, kiroHome, kb); err != nil {
		return nil, err
	}

	current, err := readManifest(installManifestPath(kiroHome, team))
	if err != nil {
		return nil, err
	}
	keep := make(map[string]struct{}, len(current.Files)+len(current.Symlinks))
	for _, p := range current.Files {
		keep[p] = struct{}{}
	}
	for _, p := range current.Symlinks {
		keep[p] = struct{}{}
	}

	var pruned []string
	for _, p := range append(previous.Symlinks, previous.Files...) {
		if _, ok := keep[p]; ok {
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return pruned, fmt.Errorf("install: prune %s: %w", p, err)
		}
		pruned = append(pruned, p)
		removeEmptyParents(filepath.Dir(p), kiroHome)
	}
	return pruned, nil
}

// removeEmptyParents removes dir and its ancestors while they are empty,
// stopping before stop. os.Remove refuses non-empty directories, which ends
// the walk.
func removeEmptyParents(dir, stop string) {
	for dir != stop && strings.HasPrefix(dir, stop+string(filepath.Separator)) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// readManifest returns the zero manifest when path does not exist.
func readManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return m, fmt.Errorf("install: read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("install: parse %s: %w", path, err)
	}
	return m, nil
}
