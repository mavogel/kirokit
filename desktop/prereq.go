package main

import (
	"os"
	"runtime"
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
	checks = append(checks, toolCheck(env, "SOPS", "sops", sopsHintFor(runtime.GOOS)))
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

// ageKeyCandidate is the file sops reads the key from on this OS.
func ageKeyCandidate(env EnvCapture) string {
	return ageKeyCandidateFor(runtime.GOOS, env, homeDir(), userConfigDir())
}

func findAgeKey(env EnvCapture) (string, bool) {
	p := ageKeyCandidate(env)
	info, err := os.Stat(p)
	return p, err == nil && !info.IsDir()
}
