//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// killTreeOnCancel puts cmd and everything it starts in one process group, so
// cancelling cmd's context ends all of it and releases its output pipe.
func killTreeOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
}

// hideConsole is a no-op off Windows, where a helper process opens no window.
func hideConsole(*exec.Cmd) {}
