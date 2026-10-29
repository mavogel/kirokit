package main

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App is the one service the frontend binds to (main.App.<Method>).
type App struct {
	dataDir string

	// captureFn reads the login environment; tests replace it.
	captureFn func() EnvCapture
	// startFn starts a program without waiting (OpenFolder); nil starts it
	// for real, tests replace it.
	startFn func(name string, args ...string) error

	mu       sync.Mutex
	env      EnvCapture
	envReady bool

	// emitter is set by setEmitter; before it, emit drops events.
	emitter func(event string, data any)
	cat     catalogCache
	poller  *installPoller
	shell   shellHost
	runner  runHost
}

// EnvStatus describes the environment the app runs commands in.
type EnvStatus struct {
	// DefaultKiroHome is the Kiro home folder used until Settings saves one:
	// the login shell's KIRO_HOME once captured, else ~/.kiro.
	DefaultKiroHome string `json:"defaultKiroHome"`
	// DefaultRepoDir is the clone folder used until Settings saves one.
	DefaultRepoDir string `json:"defaultRepoDir"`
	// Ready is false until the login shell has been read.
	Ready bool `json:"ready"`
	// Note explains a fallback, such as a login shell that was too slow.
	Note string `json:"note"`
	// Shell names the shell tab's shell: zsh, bash, PowerShell.
	Shell string `json:"shell"`
}

// AppState is what the shell renders at start and after a change.
type AppState struct {
	Settings Settings  `json:"settings"`
	Cloned   bool      `json:"cloned"`
	Env      EnvStatus `json:"env"`
}

// NewApp returns an App that keeps its clone and settings in kirokit-desktop
// in the user's config folder: ~/Library/Application Support on macOS,
// ~/.config on Linux, %AppData% on Windows.
func NewApp() *App {
	return NewAppAt(dataDirFor(userConfigDir()))
}

// NewAppAt returns an App whose data folder is dir.
func NewAppAt(dir string) *App {
	return &App{
		dataDir:   dir,
		captureFn: func() EnvCapture { return Capture(os.Getenv("SHELL"), captureTimeout) },
	}
}

// setEmitter sets the function events reach the page through; main.go sets
// Wails' emitter before the app runs, tests set their own.
func (a *App) setEmitter(fn func(event string, data any)) {
	a.mu.Lock()
	a.emitter = fn
	a.mu.Unlock()
}

// ServiceStartup is Wails' service start hook: the login shell is read, the
// catalog rendered and the install records watched in the background.
func (a *App) ServiceStartup(context.Context, application.ServiceOptions) error {
	go a.refreshEnv()
	go a.refreshCatalog()
	a.startInstallPoller()
	return nil
}

// ServiceShutdown is Wails' service shutdown hook. On macOS quitting ends the
// process without returning from Run, so every cleanup belongs here.
func (a *App) ServiceShutdown() error {
	a.mu.Lock()
	p := a.poller
	a.poller = nil
	a.mu.Unlock()
	if p != nil {
		p.Stop()
	}
	a.shell.close()
	a.runner.close()
	return nil
}

func (a *App) emit(event string, data any) {
	a.mu.Lock()
	fn := a.emitter
	a.mu.Unlock()
	if fn != nil {
		fn(event, data)
	}
}

// refreshEnv captures the login environment afresh and keeps it.
func (a *App) refreshEnv() EnvCapture {
	env := a.captureFn()
	a.mu.Lock()
	a.env, a.envReady = env, true
	a.mu.Unlock()
	return env
}

// State reports the saved settings, whether the repository is cloned and the
// environment status.
func (a *App) State() AppState {
	_, err := os.Stat(filepath.Join(a.repoDir(), ".git"))
	a.mu.Lock()
	env, ready := a.env, a.envReady
	a.mu.Unlock()
	shell := a.shellPath()
	note := strings.TrimSpace(env.Note + " " + shellNote(shell))
	return AppState{
		Settings: a.loadSettings(),
		Cloned:   err == nil,
		Env: EnvStatus{
			DefaultKiroHome: defaultKiroHome(env, ready), DefaultRepoDir: a.defaultRepoDir(), Ready: ready, Note: note,
			Shell: shellNameFor(goruntime.GOOS, shell),
		},
	}
}

// defaultRepoDir is where the clone lives until Settings names another folder.
func (a *App) defaultRepoDir() string { return filepath.Join(a.dataDir, "repo") }

// repoDir is the folder the repository is cloned into: the saved Settings
// value, else the default.
func (a *App) repoDir() string {
	if p := a.loadSettings().RepoDir; p != "" {
		return p
	}
	return a.defaultRepoDir()
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}
