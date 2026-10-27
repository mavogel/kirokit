package quick

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
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
