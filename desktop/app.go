package main

import (
	"os"
	"path/filepath"
)

// App is the one service the frontend binds to.
type App struct {
	dataDir string
}

// EnvStatus describes the environment the app runs commands in.
type EnvStatus struct {
	// DefaultKiroHome is the Kiro home folder used until Settings saves one.
	DefaultKiroHome string `json:"defaultKiroHome"`
}

// AppState is what the shell renders at start and after a change.
type AppState struct {
	Settings Settings  `json:"settings"`
	Cloned   bool      `json:"cloned"`
	Env      EnvStatus `json:"env"`
}

// NewApp returns an App that keeps its clone and settings in
// ~/Library/Application Support/kirokit-desktop.
func NewApp() *App {
	return NewAppAt(filepath.Join(homeDir(), "Library", "Application Support", "kirokit-desktop"))
}

// NewAppAt returns an App whose data folder is dir.
func NewAppAt(dir string) *App {
	return &App{dataDir: dir}
}

// State reports the saved settings, whether the repository is cloned and the
// environment status.
func (a *App) State() AppState {
	_, err := os.Stat(filepath.Join(a.dataDir, "repo", ".git"))
	return AppState{
		Settings: a.loadSettings(),
		Cloned:   err == nil,
		Env:      EnvStatus{DefaultKiroHome: filepath.Join(homeDir(), ".kiro")},
	}
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}
