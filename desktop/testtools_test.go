package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const onWindows = runtime.GOOS == "windows"

// exeName is name as an installed program is called on this OS.
func exeName(name string) string {
	if onWindows {
		return name + ".exe"
	}
	return name
}

// writeTool writes a stand-in program into dir and returns its path: the sh
// script off Windows, the cmd.exe batch file (name.cmd) on Windows.
func writeTool(t *testing.T, dir, name, sh, cmd string) string {
	t.Helper()
	if !onWindows {
		return writeScript(t, dir, name, sh)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name+".cmd")
	body := "@echo off\r\n" + strings.ReplaceAll(cmd, "\n", "\r\n") + "\r\n"
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// loginVars is the login environment a test app captures. On Windows the app
// reads no login shell and runs with its own environment, so the test does
// the same (cmd.exe needs SystemRoot, PATH and the like).
func loginVars(vars map[string]string) map[string]string {
	if !onWindows {
		return vars
	}
	base := environmentFor("windows", os.Environ(), homeDir())
	for k, v := range vars {
		if k == "PATH" || k == "SHELL" {
			continue
		}
		base[k] = v
	}
	return base
}
