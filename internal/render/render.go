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

	"github.com/mavogel/kirokit/internal/backup"
	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/mavogel/kirokit/internal/kiroconfig"
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

	teamDir := filepath.Join(teamsRoot, team)
	externalNames, err := externalskills.Resolve(teamDir, ctx.Team.ExternalSkills)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	agentResources := append(globalSteeringURIs(ctx), externalSkillURIs(ctx.KiroHome, externalNames)...)
	agentResources = append(agentResources, workspaceResources...)

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
		// `kirokit onboard` leaves .gitkeep placeholders in empty categories;
		// they are not config and must not be validated or rendered.
		if d.Name() == ".gitkeep" {
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
	if err := writeExternalSkills(tmpDir, teamDir, externalNames); err != nil {
		return err
	}

	return publish(tmpDir, outDir)
}

// renameDir is replaced in tests.
var renameDir = os.Rename

// publish makes outDir hold exactly the rendered tree. An existing outDir
// whose content would change is first moved to a timestamped backup beside it
// (see backup.MoveAside), so nothing in it is lost; one that already matches is
// left alone, and an empty one is simply replaced.
func publish(tmpDir, outDir string) error {
	var movedTo string // where the previous outDir went, if it was moved aside
	info, err := os.Lstat(outDir)
	switch {
	case os.IsNotExist(err):
		// nothing to preserve
	case err != nil:
		return fmt.Errorf("render: %w", err)
	case info.IsDir():
		same, err := backup.TreesEqual(outDir, tmpDir)
		if err != nil {
			return fmt.Errorf("render: %w", err)
		}
		if same {
			return nil // the deferred cleanup removes tmpDir
		}
		if empty, err := isEmptyDir(outDir); err != nil {
			return fmt.Errorf("render: %w", err)
		} else if empty {
			if err := os.Remove(outDir); err != nil {
				return fmt.Errorf("render: %w", err)
			}
			break
		}
		fallthrough
	default:
		if movedTo, err = backup.MoveAside(outDir, backup.Stamp()); err != nil {
			return fmt.Errorf("render: %w", err)
		}
	}
	if err := renameDir(tmpDir, outDir); err != nil {
		if movedTo == "" {
			return fmt.Errorf("render: finalize output: %w", err)
		}
		// Put the previous output back rather than leave outDir missing.
		if restoreErr := os.Rename(movedTo, outDir); restoreErr != nil {
			return fmt.Errorf("render: finalize output: %w (the previous output is preserved at %s)", err, movedTo)
		}
		return fmt.Errorf("render: finalize output: %w (the previous output was restored)", err)
	}
	return nil
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) == 0, err
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

	case "external":
		// Vendored external skills and their lock file are copied into the
		// output by writeExternalSkills, never rendered as team sources.
		return "", true, nil

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
