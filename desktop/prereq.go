package main

import (
	"os"
	"path/filepath"
)

// Check is one prerequisite and whether it was found. Detail is the path that
// was found, or what to do when it is missing.
type Check struct {
	Name   string `json:"name"`
	Found  bool   `json:"found"`
	Detail string `json:"detail"`
}

// Prerequisites captures the login environment afresh, so "Check again" sees
// what was installed since the last look, and reports SOPS, the decryption
// key and Kiro CLI. team is the team the setup dialog is about; the checks
// are the same for every team.
func (a *App) Prerequisites(team string) []Check {
	_ = team
	env := a.refreshEnv()

	checks := make([]Check, 0, 3)
	checks = append(checks, toolCheck(env, "SOPS", "sops", "Install it with: brew install sops"))
	if key, ok := findAgeKey(env); ok {
		checks = append(checks, Check{Name: "Decryption key", Found: true, Detail: key})
	} else {
		checks = append(checks, Check{Name: "Decryption key", Detail: "Ask your team lead for the age key and save it as " + ageKeyCandidate(env)})
	}
	checks = append(checks, toolCheck(env, "Kiro CLI", "kiro-cli", "Install Kiro CLI from kiro.dev, then sign in"))
	return checks
}

func toolCheck(env EnvCapture, name, bin, hint string) Check {
	if p, ok := env.Lookup(bin); ok {
		return Check{Name: name, Found: true, Detail: p}
	}
	return Check{Name: name, Detail: hint}
}

// ageKeyCandidate is the file the key is read from: SOPS_AGE_KEY_FILE, else
// $XDG_CONFIG_HOME/sops/age/keys.txt, else the macOS default.
func ageKeyCandidate(env EnvCapture) string {
	if p := env.Getenv("SOPS_AGE_KEY_FILE"); p != "" {
		return p
	}
	if xdg := env.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "sops", "age", "keys.txt")
	}
	home := env.Getenv("HOME")
	if home == "" {
		home = homeDir()
	}
	return filepath.Join(home, "Library", "Application Support", "sops", "age", "keys.txt")
}

func findAgeKey(env EnvCapture) (string, bool) {
	p := ageKeyCandidate(env)
	info, err := os.Stat(p)
	return p, err == nil && !info.IsDir()
}
