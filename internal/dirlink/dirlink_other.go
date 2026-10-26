//go:build !windows

package dirlink

import (
	"io/fs"
	"os"
)

func makeLink(source, mount string) error { return os.Symlink(source, mount) }

func isJunction(string, fs.FileInfo) bool { return false }
