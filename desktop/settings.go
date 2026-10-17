package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Settings are the two values the person chooses; settings.json beside repo/
// holds them. An empty RepoURL means no repository is configured yet.
type Settings struct {
	RepoURL  string `json:"repoUrl"`
	KiroHome string `json:"kiroHome"`
}

// repoURLPattern accepts https://, ssh:// and file:// addresses, scp-style
// git@host:path and absolute paths. Every alternative starts with a letter or
// "/", so an address can never be read by git as an option.
var repoURLPattern = regexp.MustCompile(`^(?:(?:https|ssh|file)://\S+|git@\S+:\S+|/\S+)$`)

func (a *App) settingsPath() string { return filepath.Join(a.dataDir, "settings.json") }

func (a *App) loadSettings() Settings {
	var s Settings
	data, err := os.ReadFile(a.settingsPath())
	if err != nil {
		return Settings{}
	}
	if json.Unmarshal(data, &s) != nil {
		return Settings{}
	}
	return s
}

// SaveSettings validates both values and stores them. A refusal names each
// bad field on its own line as "KiroHome: ..." or "RepoURL: ...", and saves
// nothing.
func (a *App) SaveSettings(s Settings) error {
	s.RepoURL = strings.TrimSpace(s.RepoURL)
	s.KiroHome = strings.TrimSpace(s.KiroHome)

	var problems []error
	if msg := checkKiroHome(s.KiroHome); msg != "" {
		problems = append(problems, errors.New("KiroHome: "+msg))
	} else {
		s.KiroHome = filepath.Clean(s.KiroHome)
	}
	if s.RepoURL != "" && !repoURLPattern.MatchString(s.RepoURL) {
		problems = append(problems, errors.New("RepoURL: Use an https, ssh or file address, git@host:path, or an absolute path"))
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(a.dataDir, "settings-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // gone after a successful rename
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close() //nolint:errcheck // the chmod error is the one to report
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck // the write error is the one to report
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), a.settingsPath())
}

func checkKiroHome(p string) string {
	if !filepath.IsAbs(p) {
		return "Enter an absolute folder, such as " + filepath.Join(homeDir(), ".kiro")
	}
	clean := filepath.Clean(p)
	if clean == "/" || clean == filepath.Clean(homeDir()) {
		return "Choose a folder inside your home folder, not the home folder itself"
	}
	return ""
}
