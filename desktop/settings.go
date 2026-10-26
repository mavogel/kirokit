package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Settings are the values the person chooses; settings.json in the app folder
// holds them. An empty RepoURL means no repository is configured yet; an empty
// RepoDir means the clone lives in <app folder>/repo.
type Settings struct {
	RepoURL  string `json:"repoUrl"`
	KiroHome string `json:"kiroHome"`
	RepoDir  string `json:"repoDir"`
}

// repoURLPattern accepts https://, ssh:// and file:// addresses, scp-style
// git@host:path and absolute paths, including Windows drive paths (C:\... or
// C:/...). Every alternative starts with a letter or "/", so an address can
// never be read by git as an option.
var repoURLPattern = regexp.MustCompile(`^(?:(?:https|ssh|file)://\S+|git@\S+:\S+|/\S+|[A-Za-z]:[\\/]\S+)$`)

// windowsRoot is a drive root (C:\, C:/) or the current drive's root (\).
var windowsRoot = regexp.MustCompile(`^(?:[A-Za-z]:)?[\\/]+$`)

// isRootFolderFor reports whether p is the root of a file system on goos.
func isRootFolderFor(goos, p string) bool {
	if goos == "windows" {
		return windowsRoot.MatchString(p)
	}
	return path.Clean(p) == "/" // a Unix path, judged the same on any machine
}

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

// SaveSettings validates the values and stores them. A refusal names each
// bad field on its own line as "KiroHome: ...", "RepoURL: ..." or
// "RepoDir: ...", and saves nothing.
func (a *App) SaveSettings(s Settings) error {
	s.RepoURL = strings.TrimSpace(s.RepoURL)
	s.KiroHome = strings.TrimSpace(s.KiroHome)
	s.RepoDir = strings.TrimSpace(s.RepoDir)

	var problems []error
	if msg := checkKiroHome(s.KiroHome); msg != "" {
		problems = append(problems, errors.New("KiroHome: "+msg))
	} else {
		s.KiroHome = filepath.Clean(s.KiroHome)
	}
	if s.RepoURL != "" && !repoURLPattern.MatchString(s.RepoURL) {
		problems = append(problems, errors.New("RepoURL: "+repoURLMessage))
	}
	if s.RepoDir != "" {
		if msg := checkRepoDir(s.RepoDir); msg != "" {
			problems = append(problems, errors.New("RepoDir: "+msg))
		} else {
			s.RepoDir = filepath.Clean(s.RepoDir)
		}
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}
	return a.writeSettings(s)
}

// SetRepoDir stores only the clone folder (empty restores the default), for
// the first-launch page, which has no other setting to save. A refusal reads
// "RepoDir: ..." and saves nothing.
func (a *App) SetRepoDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir != "" {
		if msg := checkRepoDir(dir); msg != "" {
			return errors.New("RepoDir: " + msg)
		}
		dir = filepath.Clean(dir)
	}
	s := a.loadSettings()
	s.RepoDir = dir
	return a.writeSettings(s)
}

// repoURLMessage is shown beside an address field that holds a bad address.
const repoURLMessage = "Use an https, ssh or file address, git@host:path, or an absolute path"

// writeSettings stores s as it is, replacing settings.json atomically.
func (a *App) writeSettings(s Settings) error {
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

// defaultKiroHome is the captured KIRO_HOME, else ~/.kiro.
func defaultKiroHome(env EnvCapture, ready bool) string {
	if ready {
		if p := env.Getenv("KIRO_HOME"); p != "" {
			return p
		}
	}
	return filepath.Join(homeDir(), ".kiro")
}

// kiroHome is the Kiro home folder commands act on: the saved Settings value,
// which wins over the shell's own KIRO_HOME, else the default.
func (a *App) kiroHome() string {
	if p := a.loadSettings().KiroHome; p != "" {
		return p
	}
	a.mu.Lock()
	env, ready := a.env, a.envReady
	a.mu.Unlock()
	return defaultKiroHome(env, ready)
}

func checkKiroHome(p string) string {
	if !filepath.IsAbs(p) {
		return "Enter an absolute folder, such as " + filepath.Join(homeDir(), ".kiro")
	}
	clean := filepath.Clean(p)
	if isRootFolderFor(runtime.GOOS, clean) || clean == filepath.Clean(homeDir()) {
		return "Choose a folder inside your home folder, not the home folder itself"
	}
	return ""
}

// checkRepoDir refuses a clone folder that is relative, is the root or home
// folder, or already holds something other than a git checkout, since a new
// clone sets an existing folder aside.
func checkRepoDir(p string) string {
	if !filepath.IsAbs(p) {
		return "Enter an absolute folder, such as " + filepath.Join(homeDir(), "kirokit", "repo")
	}
	clean := filepath.Clean(p)
	if isRootFolderFor(runtime.GOOS, clean) || clean == filepath.Clean(homeDir()) {
		return "Choose a folder inside your home folder, not the home folder itself"
	}
	entries, err := os.ReadDir(clean)
	if err == nil && len(entries) > 0 && !exists(filepath.Join(clean, ".git")) {
		return "That folder already has files in it. Choose a new or empty folder"
	}
	return ""
}
