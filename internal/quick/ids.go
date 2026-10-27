package quick

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// IDsFileName is the ids file's name at the config repository root.
const IDsFileName = "quick-ids.yaml"

// idPattern is the shape of a server-assigned Quick skill id.
var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// idsFile is the on-disk shape of quick-ids.yaml.
type idsFile struct {
	Skills map[string]string `yaml:"skills"`
}

// checkID fails unless id looks like a Quick skill id.
func checkID(name, id string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("empty skill name")
	}
	if !idPattern.MatchString(id) {
		return fmt.Errorf("skill %q: id %q is not 32 lowercase hex characters", name, id)
	}
	return nil
}

// Limits for a .zip passed to ReadExports, which reads it in memory.
const (
	MaxZipEntries    = 256
	MaxZipEntryBytes = 1 << 20
)

// idsHeader opens every ids file WriteIDs writes.
const idsHeader = "# Amazon Quick skill ids for one Quick account. Written by `kirokit quick ids`; never commit.\n"

// ReadExports reads the frontmatter name and id of skills exported from
// Quick. Each path is a .md file, a .zip (each SKILL.md entry read in memory)
// or a folder searched for SKILL.md files and .zip files. Exports without an
// id are returned in skipped. Two different ids for one name, a malformed id
// or an export without frontmatter is an error. The files are only parsed;
// nothing in them is run or written.
func ReadExports(paths []string) (ids map[string]string, skipped []string, err error) {
	ids = map[string]string{}
	add := func(where string, data []byte) error {
		name, id, err := exportedID(data)
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if id == "" {
			skipped = append(skipped, where)
			return nil
		}
		if err := checkID(name, id); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if prev, seen := ids[name]; seen && prev != id {
			return fmt.Errorf("%s: two different ids for skill %q (%s and %s)", where, name, prev, id)
		}
		ids[name] = id
		return nil
	}
	readOne := func(p string) error {
		if strings.EqualFold(filepath.Ext(p), ".zip") {
			return readZip(p, add)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return add(p, data)
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, nil, err
		}
		if !info.IsDir() {
			if err := readOne(p); err != nil {
				return nil, nil, err
			}
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if d.Name() == "SKILL.md" || strings.EqualFold(filepath.Ext(path), ".zip") {
				return readOne(path)
			}
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return ids, skipped, nil
}

// readZip passes each SKILL.md entry of the zip at path to add.
func readZip(path string, add func(where string, data []byte) error) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer func() { _ = r.Close() }()
	if len(r.File) > MaxZipEntries {
		return fmt.Errorf("%s: has more than %d entries", path, MaxZipEntries)
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() || pathpkg.Base(f.Name) != "SKILL.md" {
			continue
		}
		where := path + ":" + f.Name
		if f.UncompressedSize64 > MaxZipEntryBytes {
			return fmt.Errorf("%s: larger than %d bytes", where, MaxZipEntryBytes)
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		// The header's size can lie; read at most one byte past the limit.
		data, err := io.ReadAll(io.LimitReader(rc, MaxZipEntryBytes+1))
		_ = rc.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if len(data) > MaxZipEntryBytes {
			return fmt.Errorf("%s: larger than %d bytes", where, MaxZipEntryBytes)
		}
		if err := add(where, data); err != nil {
			return err
		}
	}
	return nil
}

// exportedID reads name and id from an exported SKILL.md's frontmatter.
func exportedID(data []byte) (name, id string, err error) {
	fm, _, ok := splitFrontmatter(data)
	if !ok {
		return "", "", fmt.Errorf("no frontmatter; not an exported skill")
	}
	var meta struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}
	if err := yaml.Unmarshal(fm, &meta); err != nil {
		return "", "", fmt.Errorf("frontmatter: %w", err)
	}
	if strings.TrimSpace(meta.Name) == "" {
		return "", "", fmt.Errorf("frontmatter has no name")
	}
	return meta.Name, meta.ID, nil
}

// WriteIDs writes ids to path, sorted by name, readable only by the owner.
// It writes a temporary file beside path and renames it into place, so a
// failed write leaves the previous file whole.
func WriteIDs(path string, ids map[string]string) error {
	for name, id := range ids {
		if err := checkID(name, id); err != nil {
			return fmt.Errorf("quick: %w", err)
		}
	}
	var buf bytes.Buffer
	buf.WriteString(idsHeader)
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(idsFile{Skills: ids}); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".quick-ids-*")
	if err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("quick: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("quick: %w", err)
	}
	return nil
}

// LoadIDs reads the skill-name-to-id map of an ids file. A missing file is
// an empty map, because the first render runs before Quick assigned any id.
func LoadIDs(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("quick: read ids file: %w", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f idsFile
	if err := dec.Decode(&f); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("quick: %s: %w", path, err)
	}
	ids := make(map[string]string, len(f.Skills))
	for name, id := range f.Skills {
		if err := checkID(name, id); err != nil {
			return nil, fmt.Errorf("quick: %s: %w", path, err)
		}
		ids[name] = id
	}
	return ids, nil
}
