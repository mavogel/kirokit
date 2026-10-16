package render

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mavogel/kirokit/internal/manifest"
	"github.com/mavogel/kirokit/internal/mcpversions"
)

// DefaultKiroHomeSentinel is the fixed, portable placeholder used for
// render/test/update so golden fixtures never depend on a real machine's
// home directory. Only install resolves and bakes in the real $KIRO_HOME.
const DefaultKiroHomeSentinel = "/kiro-home"

// Context is the data every team template renders against.
type Context struct {
	Team     manifest.Manifest
	Env      mcpversions.Versions
	Global   map[string]string // opted-in global steering slug -> file content
	KiroHome string
}

// BuildContext loads a team's manifest, the central MCP version pins, and
// the content of every global steering that team's manifest opts into.
func BuildContext(teamsRoot, team, globalRoot, envVersionsPath, kiroHome string) (*Context, error) {
	m, err := manifest.Load(filepath.Join(teamsRoot, team, "manifest.yaml"))
	if err != nil {
		return nil, err
	}

	versions, err := mcpversions.Load(envVersionsPath)
	if err != nil {
		return nil, err
	}

	global := make(map[string]string, len(m.GlobalSteerings))
	for _, slug := range m.GlobalSteerings {
		path := filepath.Join(globalRoot, slug+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("render: team %q opts into global steering %q but %s could not be read: %w", team, slug, path, err)
		}
		global[slug] = string(data)
	}

	if kiroHome == "" {
		kiroHome = DefaultKiroHomeSentinel
	}

	return &Context{
		Team:     *m,
		Env:      versions,
		Global:   global,
		KiroHome: kiroHome,
	}, nil
}
