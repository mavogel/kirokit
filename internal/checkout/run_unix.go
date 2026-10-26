//go:build !windows

package checkout

import "os/exec"

// hideWindow does nothing off Windows, where git opens no window.
func hideWindow(*exec.Cmd) {}
