// Package mcpversions parses the central env/mcp-versions.yaml file that
// pins each MCP server type's version/image tag once for all teams.
package mcpversions

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Versions maps an MCP server name to its pinned version/tag string.
type Versions map[string]string

// Load parses the flat env/mcp-versions.yaml map at path.
func Load(path string) (Versions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("mcpversions: read %s: %w", path, err)
	}

	v := make(Versions)
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("mcpversions: parse %s: %w", path, err)
	}
	return v, nil
}
