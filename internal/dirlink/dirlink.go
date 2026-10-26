// Package dirlink links a folder at another path without copying it: a
// symbolic link, or on Windows, where creating one needs Developer Mode or
// elevation, a directory junction. Install, uninstall and backup use it to
// recognise the knowledge-base link kirokit made.
package dirlink

import (
	"io/fs"
	"os"
)

// Make links mount to the folder source. mount must not exist.
func Make(source, mount string) error { return makeLink(source, mount) }

// Is reports whether path is a link Make could have made (a symbolic link, or
// on Windows a junction), without following it.
func Is(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return true
	}
	return isJunction(path, info)
}

// Target is the path a link points at.
func Target(path string) (string, error) { return os.Readlink(path) }
