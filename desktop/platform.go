package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Per-OS paths, words and tools. Each function takes the OS name, so every
// branch is tested on any machine; callers pass runtime.GOOS. macOS keeps
// every value it had before Linux and Windows were supported.

// dataDirFor is the app's data folder (settings, default clone, zsh wrapper)
// inside the user's config folder: ~/Library/Application Support on macOS,
// ~/.config on Linux, %AppData% on Windows.
func dataDirFor(configDir string) string { return filepath.Join(configDir, "kirokit-desktop") }

// userConfigDir is os.UserConfigDir, else the folder it names on Linux.
func userConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return dir
	}
	return filepath.Join(homeDir(), ".config")
}

// ageKeyCandidateFor is the file sops reads the age key from on goos:
// SOPS_AGE_KEY_FILE, else (off Windows) $XDG_CONFIG_HOME/sops/age/keys.txt,
// else the OS's config folder (macOS ~/Library/Application Support, Linux
// ~/.config, Windows %AppData%) plus sops/age/keys.txt. home and appData are
// used when the environment has no HOME or APPDATA.
func ageKeyCandidateFor(goos string, env EnvCapture, home, appData string) string {
	if p := env.Getenv("SOPS_AGE_KEY_FILE"); p != "" {
		return p
	}
	var config string
	switch goos {
	case "windows":
		config = appData
		if v := env.Getenv("APPDATA"); v != "" {
			config = v
		}
	default:
		if xdg := env.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "sops", "age", "keys.txt")
		}
		if v := env.Getenv("HOME"); v != "" {
			home = v
		}
		config = filepath.Join(home, ".config")
		if goos == "darwin" {
			config = filepath.Join(home, "Library", "Application Support")
		}
	}
	return filepath.Join(config, "sops", "age", "keys.txt")
}

// listSepFor is the PATH list separator on goos.
func listSepFor(goos string) string {
	if goos == "windows" {
		return ";"
	}
	return ":"
}

// toolFoldersFor are where package managers and per-user installs put sops
// and kiro-cli on goos; Capture appends the ones PATH lacks.
func toolFoldersFor(goos, home string) []string {
	switch goos {
	case "darwin":
		return []string{"/opt/homebrew/bin", "/opt/homebrew/sbin", "/usr/local/bin", filepath.Join(home, ".local", "bin")}
	case "windows":
		return []string{filepath.Join(home, "scoop", "shims")}
	default:
		return []string{"/home/linuxbrew/.linuxbrew/bin", "/usr/local/bin", filepath.Join(home, ".local", "bin")}
	}
}

// withToolFoldersFor appends the tool folders missing from path, so a login
// shell without startup files (or an app started from a menu) still finds
// the package manager's tools.
func withToolFoldersFor(goos, path, home string) string {
	sep := listSepFor(goos)
	have := map[string]bool{}
	var parts []string
	for p := range strings.SplitSeq(path, sep) {
		if p != "" {
			have[p] = true
			parts = append(parts, p)
		}
	}
	for _, dir := range toolFoldersFor(goos, home) {
		if !have[dir] {
			parts = append(parts, dir)
		}
	}
	return strings.Join(parts, sep)
}

// defaultShellFor is the login shell on goos when none was captured: the
// app's own $SHELL, else zsh on macOS and bash (else sh) on Linux. Windows
// has no login shell to read, so it is empty there.
func defaultShellFor(goos, appShell string, exists func(string) bool) string {
	switch {
	case goos == "windows":
		return ""
	case appShell != "":
		return appShell
	case goos == "darwin":
		return "/bin/zsh"
	case exists("/bin/bash"):
		return "/bin/bash"
	default:
		return "/bin/sh"
	}
}

// environmentFor is the environment the app runs commands in on an OS where it
// reads no login shell (Windows): its own, with Windows' Path spelling folded
// into the one PATH and the tool folders appended.
func environmentFor(goos string, environ []string, home string) map[string]string {
	vars := map[string]string{}
	path := ""
	for _, kv := range environ {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || k == "" {
			continue
		}
		if goos == "windows" && strings.EqualFold(k, "PATH") {
			path = v
			continue
		}
		vars[k] = v
	}
	if goos != "windows" {
		path = vars["PATH"]
	}
	vars["PATH"] = withToolFoldersFor(goos, path, home)
	return vars
}

// lookupFor finds the executable name in dirs as goos would: on Windows by
// trying each PATHEXT suffix (no file mode says "executable" there), else a
// file with an execute bit.
func lookupFor(goos string, dirs []string, pathext, name string) (string, bool) {
	candidates := []string{name}
	if goos == "windows" {
		if pathext == "" {
			pathext = ".COM;.EXE;.BAT;.CMD"
		}
		if filepath.Ext(name) == "" {
			candidates = nil
			for ext := range strings.SplitSeq(pathext, ";") {
				if ext != "" { // Windows file names ignore case
					candidates = append(candidates, name+strings.ToLower(ext))
				}
			}
		}
	}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		for _, c := range candidates {
			p := filepath.Join(dir, c)
			info, err := os.Stat(p)
			if err != nil || info.IsDir() {
				continue
			}
			if goos == "windows" || info.Mode().Perm()&0o111 != 0 {
				return p, true
			}
		}
	}
	return "", false
}

// installedCLIFor is where the app's own package put the bundled CLI, given
// the folder of the app's executable.
func installedCLIFor(goos, exeDir string) []string {
	switch goos {
	case "darwin": // kirokit.app/Contents/MacOS/kirokit -> Contents/Resources/bin/kirokit
		return []string{filepath.Join(exeDir, "..", "Resources", "bin", "kirokit")}
	case "windows": // <install dir>\kirokit-desktop.exe -> <install dir>\bin\kirokit.exe
		return []string{filepath.Join(exeDir, "bin", "kirokit.exe")}
	default: // /usr/bin/kirokit-desktop -> /usr/lib/kirokit-desktop/kirokit
		return []string{filepath.Join(exeDir, "..", "lib", "kirokit-desktop", "kirokit")}
	}
}

// cliMissingHintFor ends the error shown when the bundled CLI is missing.
func cliMissingHintFor(goos string) string {
	switch goos {
	case "darwin":
		return "or start the app from kirokit.app"
	case "windows":
		return "or reinstall kirokit with its setup program"
	default:
		return "or reinstall the kirokit-desktop package"
	}
}

// shellCommandFor is the shell tab's program and arguments: the login shell
// as `<shell> -il`, or on Windows PowerShell 7 (pwsh.exe) when it is on PATH,
// else Windows PowerShell.
func shellCommandFor(goos, shell string, lookPath func(string) (string, error)) (string, []string) {
	if goos != "windows" {
		return shell, []string{"-il"}
	}
	if p, err := lookPath("pwsh.exe"); err == nil {
		return p, []string{"-NoLogo"}
	}
	return "powershell.exe", []string{"-NoLogo"}
}

// shellNameFor is the shell tab's label: PowerShell on Windows, else the
// login shell's name (zsh, bash, fish).
func shellNameFor(goos, shell string) string {
	if goos == "windows" {
		return "PowerShell"
	}
	return filepath.Base(shell)
}

// shellNoteFor tells Settings when the bundled CLI cannot be guaranteed first
// on the shell tab's PATH: a login shell other than zsh may reorder it. On
// Windows PowerShell starts with the app's environment, so there is no note.
func shellNoteFor(goos, shell string) string {
	if goos == "windows" || isZsh(shell) {
		return ""
	}
	return "Your shell is " + filepath.Base(shell) + ", so the bundled kirokit may not come first on its PATH if your startup files change it."
}

// openCommandFor is the program that shows a folder in the file manager.
func openCommandFor(goos string) string {
	switch goos {
	case "darwin":
		return "open"
	case "windows":
		return "explorer"
	default:
		return "xdg-open"
	}
}

// sopsHintFor is the setup check's hint when sops is missing.
func sopsHintFor(goos string) string {
	switch goos {
	case "darwin":
		return "Install it with: brew install sops"
	case "windows":
		return "Install it with: scoop install sops"
	default:
		return "Install it from your package manager, or from https://github.com/getsops/sops/releases"
	}
}
