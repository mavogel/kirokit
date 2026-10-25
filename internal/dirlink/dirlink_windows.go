//go:build windows

package dirlink

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// makeLink tries a symbolic link first, and makes a junction when Windows
// refuses the link for lack of privilege.
func makeLink(source, mount string) error {
	err := os.Symlink(source, mount)
	if err == nil || !errors.Is(err, syscall.ERROR_PRIVILEGE_NOT_HELD) {
		return err
	}
	return makeJunction(source, mount)
}

// makeJunction runs cmd.exe's mklink /J, which needs no privilege. Both paths
// must be absolute; a path with '"' is refused, since cmd.exe cannot quote it.
func makeJunction(source, mount string) error {
	if strings.Contains(source, `"`) || strings.Contains(mount, `"`) {
		return fmt.Errorf("cannot make a junction for a path containing '\"': %s -> %s", mount, source)
	}
	if _, err := os.Lstat(mount); err == nil {
		return fmt.Errorf("make junction %s: %w", mount, fs.ErrExist)
	}
	cmd := exec.Command("cmd.exe", "/C", "mklink", "/J", mount, source)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("make junction %s -> %s: %w: %s", mount, source, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// isJunction reports a mount-point reparse point: Go reports it as irregular,
// not as a symlink, but Readlink resolves it.
func isJunction(path string, info fs.FileInfo) bool {
	if info.Mode()&fs.ModeIrregular == 0 && !info.IsDir() {
		return false
	}
	_, err := os.Readlink(path)
	return err == nil
}
