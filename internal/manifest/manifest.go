// Package manifest parses a team's manifest.yaml: its identity and which
// global steerings it opts into.
package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest is a team's manifest.yaml.
type Manifest struct {
	Team            string   `yaml:"team"`
	GlobalSteerings []string `yaml:"globalSteerings"`
}

// Load parses the manifest.yaml at path.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: read %s: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest: parse %s: %w", path, err)
	}
	if m.Team == "" {
		return nil, fmt.Errorf("manifest: %s: missing required field %q", path, "team")
	}
	return &m, nil
}
