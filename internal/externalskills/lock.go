// Package externalskills vendors skills from git repositories into a team's
// directory and keeps a lock file that pins what was vendored. Everything
// except Fetch is offline, so rendering and installing never touch the
// network.
package externalskills

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// SkillsDir is where a team's vendored skills live: one directory per skill.
func SkillsDir(teamDir string) string {
	return filepath.Join(teamDir, "external", "skills")
}

// LockPath is the lock file recording what was vendored.
func LockPath(teamDir string) string {
	return filepath.Join(teamDir, "external", "skills.lock.yaml")
}

// LockEntry pins one vendored skill: where it came from, the commit the ref
// resolved to, and a hash of the vendored files.
type LockEntry struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Ref    string `yaml:"ref"`
	Path   string `yaml:"path"`
	Commit string `yaml:"commit"`
	SHA256 string `yaml:"sha256"`
	// From is the collection path (an all: true entry) that produced this
	// skill; it is empty for a single-skill entry.
	From string `yaml:"from,omitempty"`
	// Executables are the skill's files git records as executable (mode
	// 100755), slash-separated and relative to the skill folder. Windows keeps
	// no execute bit, so it verifies SHA256 with these instead of file modes.
	Executables []string `yaml:"executables,omitempty"`
}

// Lock is the contents of skills.lock.yaml.
type Lock struct {
	Skills []LockEntry `yaml:"skills"`
}

// ReadLock parses the lock file at path. A missing file is an empty lock.
func ReadLock(path string) (Lock, error) {
	var l Lock
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return l, nil
		}
		return l, fmt.Errorf("externalskills: read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &l); err != nil {
		return l, fmt.Errorf("externalskills: parse %s: %w", path, err)
	}
	return l, nil
}

// WriteLock writes l to path sorted by skill name, replacing the file
// atomically.
func WriteLock(path string, l Lock) error {
	sorted := Lock{Skills: append([]LockEntry(nil), l.Skills...)}
	sort.Slice(sorted.Skills, func(i, j int) bool { return sorted.Skills[i].Name < sorted.Skills[j].Name })
	data, err := yaml.Marshal(sorted)
	if err != nil {
		return fmt.Errorf("externalskills: marshal lock: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("externalskills: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".skills.lock-*")
	if err != nil {
		return fmt.Errorf("externalskills: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("externalskills: write lock: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("externalskills: write lock: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return fmt.Errorf("externalskills: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("externalskills: replace lock: %w", err)
	}
	return nil
}

// HashDir returns a deterministic digest of every file under dir: for each
// file in path order, its slash-separated relative path, whether its owner
// execute bit is set, and the SHA-256 of its content. Renames, additions,
// content changes and exec-bit flips therefore all change the digest.
// Symbolic links and other non-regular files are an error.
func HashDir(dir string) (string, error) {
	return hashDirWith(dir, func(_ string, info fs.FileInfo) bool { return info.Mode().Perm()&0o100 != 0 })
}

// hashDirWithExecutables is HashDir with the execute bits taken from
// executables (slash paths relative to dir) instead of the files' modes.
func hashDirWithExecutables(dir string, executables []string) (string, error) {
	set := make(map[string]bool, len(executables))
	for _, rel := range executables {
		set[rel] = true
	}
	return hashDirWith(dir, func(rel string, _ fs.FileInfo) bool { return set[rel] })
}

// hashDirWith is HashDir with execOf deciding each file's execute bit.
func hashDirWith(dir string, execOf func(rel string, info fs.FileInfo) bool) (string, error) {
	var lines []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			return fmt.Errorf("%s is a symbolic link; symbolic links are not vendored", rel)
		case !info.Mode().IsRegular():
			return fmt.Errorf("%s is not a regular file", rel)
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		exec := "-"
		if execOf(rel, info) {
			exec = "x"
		}
		lines = append(lines, rel+"\x00"+exec+"\x00"+hex.EncodeToString(sum[:])+"\n")
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("externalskills: hash %s: %w", dir, err)
	}
	sort.Strings(lines)
	h := sha256.New()
	for _, l := range lines {
		_, _ = h.Write([]byte(l))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
