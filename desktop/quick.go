package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

// QuickBundle is a team's exported Amazon Quick bundle: where it is and how
// many skills and agents it holds.
type QuickBundle struct {
	Dir    string `json:"dir"`
	Skills int    `json:"skills"`
	Agents int    `json:"agents"`
}

// QuickDir is the folder Export for Quick renders team's bundle into:
// <data folder>/quick/<team>. The frontend passes only the team name, never a
// path, so nothing outside <data folder>/quick can be named.
func (a *App) QuickDir(team string) (string, error) {
	if team == "" || team == "." || team == ".." || strings.ContainsAny(team, "/\\\x00") || filepath.IsAbs(team) {
		return "", fmt.Errorf("%q is not a team name", team)
	}
	return filepath.Join(a.dataDir, "quick", team), nil
}

// QuickBundle counts skills/*/SKILL.md and agents/*.json in team's bundle.
func (a *App) QuickBundle(team string) (QuickBundle, error) {
	dir, err := a.existingQuickDir(team)
	if err != nil {
		return QuickBundle{}, err
	}
	skills, err := filepath.Glob(filepath.Join(dir, "skills", "*", "SKILL.md"))
	if err != nil {
		return QuickBundle{}, err
	}
	agents, err := filepath.Glob(filepath.Join(dir, "agents", "*.json"))
	if err != nil {
		return QuickBundle{}, err
	}
	return QuickBundle{Dir: dir, Skills: len(skills), Agents: len(agents)}, nil
}

// OpenFolder shows team's bundle folder in the system file manager.
func (a *App) OpenFolder(team string) error {
	dir, err := a.existingQuickDir(team)
	if err != nil {
		return err
	}
	start := a.startFn
	if start == nil {
		start = startDetached
	}
	return start(openCommandFor(goruntime.GOOS), dir)
}

// existingQuickDir is QuickDir, failing when nothing was exported there yet.
func (a *App) existingQuickDir(team string) (string, error) {
	dir, err := a.QuickDir(team)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !info.IsDir()) {
		return "", fmt.Errorf("%s has no Quick bundle yet; export it first", team)
	}
	return dir, err
}

// startDetached starts a program without waiting for it; the file manager
// keeps running after the call (and explorer exits 1 even when it worked).
func startDetached(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open folder: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
