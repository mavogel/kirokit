// Package backup keeps a copy of a folder before kirokit changes it. A backup
// sits next to the original and is named YYYYMMDD-HHmmss-<name>.backup (local
// time), for example 20261015-101530-agents.backup.
package backup

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mavogel/kirokit/internal/dirlink"
)

// now is replaced in tests.
var now = time.Now

// Stamp returns the current local time as YYYYMMDD-HHmmss. Take one stamp per
// operation so all the backups it makes share a timestamp.
func Stamp() string { return now().Format("20060102-150405") }

// backupPath returns <parent>/<stamp>-<name>.backup, or, when that exists
// already (two backups within one second), <stamp>-<n>-<name>.backup.
func backupPath(path, stamp string) string {
	dir, base := filepath.Split(filepath.Clean(path))
	p := filepath.Join(dir, stamp+"-"+base+".backup")
	for n := 2; ; n++ {
		if _, err := os.Lstat(p); os.IsNotExist(err) {
			return p
		}
		p = filepath.Join(dir, fmt.Sprintf("%s-%d-%s.backup", stamp, n, base))
	}
}

// MoveAside renames path (a folder, normally) to its backup name and returns
// that name. The original path no longer exists afterwards.
func MoveAside(path, stamp string) (string, error) {
	dst := backupPath(path, stamp)
	if err := os.Rename(path, dst); err != nil {
		return "", fmt.Errorf("backup: move %s aside: %w", path, err)
	}
	return dst, nil
}

// CopyAside copies the folder dir, with its permissions and symbolic links,
// to its backup name beside it and returns that name. The original is left
// untouched.
func CopyAside(dir, stamp string) (string, error) {
	dst := backupPath(dir, stamp)
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case p != dir && dirlink.Is(p):
			// A symbolic link, or a Windows junction (which reads as a folder):
			// kept as a link, never followed into what it points at.
			link, err := dirlink.Target(p)
			if err != nil {
				return err
			}
			if info.Mode()&fs.ModeSymlink != 0 {
				err = os.Symlink(link, target)
			} else {
				err = dirlink.Make(link, target)
			}
			if err == nil && d.IsDir() {
				return fs.SkipDir
			}
			return err
		case d.IsDir():
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		case info.Mode().IsRegular():
			return copyFile(p, target, info.Mode().Perm())
		default:
			return fmt.Errorf("%s is neither a file, a folder nor a symbolic link", p)
		}
	})
	if err != nil {
		_ = os.RemoveAll(dst) // never leave a half-written backup behind
		return "", fmt.Errorf("backup: copy %s: %w", dir, err)
	}
	return dst, nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// Differs reports whether writing data to path would change something that is
// already there: an existing regular file with different content, a symbolic
// link, or a folder. A missing path, or a file that already holds exactly
// data, is not a change.
func Differs(path string, data []byte) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return true
	}
	existing, err := os.ReadFile(path)
	return err != nil || !bytes.Equal(existing, data)
}

// TreesEqual reports whether folders a and b hold the same files with the same
// content and owner-execute bits. A symbolic link or other non-regular file in
// either makes them unequal.
func TreesEqual(a, b string) (bool, error) {
	ta, err := snapshot(a)
	if err != nil {
		return false, err
	}
	tb, err := snapshot(b)
	if err != nil {
		return false, err
	}
	if len(ta) != len(tb) {
		return false, nil
	}
	for rel, fa := range ta {
		fb, ok := tb[rel]
		if !ok || fa.exec != fb.exec || !bytes.Equal(fa.data, fb.data) {
			return false, nil
		}
	}
	return true, nil
}

type fileState struct {
	data []byte
	exec bool
}

func snapshot(root string) (map[string]fileState, error) {
	files := make(map[string]fileState)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			files[rel] = fileState{data: []byte("non-regular:" + rel + fmt.Sprint(info.Mode()))}
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[rel] = fileState{data: data, exec: info.Mode().Perm()&0o100 != 0}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("backup: read %s: %w", root, err)
	}
	return files, nil
}
