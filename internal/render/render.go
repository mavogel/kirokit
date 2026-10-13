// Package render implements the kirokit template rendering pipeline: it
// walks a team's *.tmpl sources, renders them against that team's
// manifest/env/global-steering context, and writes output mapped to Kiro
// CLI v3's real (flatter) config layout.
package render

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mavogel/kiro-cli-kit/internal/kiroconfig"
)

// RenderTeam renders teamsRoot/team into outDir. globalRoot is the root of
// global/steerings; envVersionsPath is the central env/mcp-versions.yaml.
// kiroHome overrides Context.KiroHome; an empty string uses
// DefaultKiroHomeSentinel.
func RenderTeam(teamsRoot, team, globalRoot, envVersionsPath, outDir, kiroHome string) error {
	ctx, err := BuildContext(teamsRoot, team, globalRoot, envVersionsPath, kiroHome)
	if err != nil {
		return err
	}
	funcMap := FuncMap(ctx)
	agentResources := append(globalSteeringURIs(ctx), workspaceResources...)

	teamDir := filepath.Join(teamsRoot, team)

	if err := os.MkdirAll(filepath.Dir(outDir), 0o755); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(outDir), ".kirokit-render-*")
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	walkErr := filepath.WalkDir(teamDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path == filepath.Join(teamDir, "manifest.yaml") {
			return nil
		}

		relPath, err := filepath.Rel(teamDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		outRel, skip, err := mapOutputPath(relPath)
		if err != nil {
			return fmt.Errorf("render: %s: %w", relPath, err)
		}
		if skip {
			return nil
		}

		var content []byte
		if strings.HasSuffix(path, ".tmpl") {
			content, err = renderTemplate(path, relPath, funcMap, ctx)
			if err != nil {
				return err
			}
		} else {
			content, err = os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("render: read %s: %w", relPath, err)
			}
		}

		if strings.HasPrefix(outRel, "agents/") {
			content, err = addResources(content, agentResources)
			if err != nil {
				return fmt.Errorf("render: %s: add steering to resources: %w", relPath, err)
			}
		}

		if err := validateOutput(outRel, content); err != nil {
			return fmt.Errorf("render: %s: %w", relPath, err)
		}

		outPath := filepath.Join(tmpDir, filepath.FromSlash(outRel))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("render: %w", err)
		}
		if err := os.WriteFile(outPath, content, 0o600); err != nil {
			return fmt.Errorf("render: write %s: %w", outRel, err)
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if err := writeGlobalSteerings(tmpDir, ctx); err != nil {
		return err
	}

	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if err := os.Rename(tmpDir, outDir); err != nil {
		return fmt.Errorf("render: finalize output: %w", err)
	}
	return nil
}

func renderTemplate(path, relPath string, funcMap template.FuncMap, ctx *Context) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("render: read %s: %w", relPath, err)
	}
	tmpl, err := template.New(relPath).Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("render: parse %s: %w", relPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("render: execute %s: %w", relPath, err)
	}
	return buf.Bytes(), nil
}

func validateOutput(outRel string, content []byte) error {
	switch {
	case strings.HasPrefix(outRel, "agents/"):
		return kiroconfig.ValidateAgentJSON(content)
	case strings.HasPrefix(outRel, "hooks/"):
		return kiroconfig.ValidateHookJSON(content)
	default:
		return nil
	}
}

// mapOutputPath maps a team-relative source path (as it appears under
// teams/<team>/) to its Kiro-real-layout output path. skip is true for
// paths that are consumed elsewhere (agent-scoped env files) rather than
// written to the output tree.
func mapOutputPath(relPath string) (outRelPath string, skip bool, err error) {
	segments := strings.Split(relPath, "/")
	if len(segments) == 0 {
		return "", false, fmt.Errorf("unexpected empty path")
	}

	switch segments[0] {
	case "agents":
		if len(segments) < 3 {
			return "", false, fmt.Errorf("unexpected path under agents/: %s", relPath)
		}
		agent := segments[1]
		if segments[2] == "env" {
			// Agent-scoped env files (secrets, version pins) are consumed
			// by the sopsArgsJSON template func / by install, never
			// written to the rendered output tree directly.
			return "", true, nil
		}
		if len(segments) == 3 && strings.TrimSuffix(segments[2], ".tmpl") == "agent.json" {
			return "agents/" + agent + ".json", false, nil
		}
		return "", false, fmt.Errorf("unexpected path under agents/%s/: %s", agent, relPath)

	case "hooks", "skills", "prompts":
		return strings.TrimSuffix(relPath, ".tmpl"), false, nil

	case "steerings":
		rest := strings.TrimPrefix(relPath, "steerings/")
		rest = strings.TrimSuffix(rest, ".tmpl")
		return "steering/" + rest, false, nil

	default:
		return "", false, fmt.Errorf("unrecognized top-level category %q in %s", segments[0], relPath)
	}
}
