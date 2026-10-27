// Package render implements the kirokit template rendering pipeline: it
// walks a team's *.tmpl sources, renders them against that team's
// manifest/env/global-steering context, and writes output mapped to Kiro
// CLI v3's real (flatter) config layout.
package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/mavogel/kirokit/internal/backup"
	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/mavogel/kirokit/internal/kiroconfig"
	"github.com/mavogel/kirokit/internal/manifest"
	"github.com/mavogel/kirokit/internal/sopswrap"
)

// Options are the per-machine inputs of a render.
type Options struct {
	// KiroHome overrides Context.KiroHome; empty uses DefaultKiroHomeSentinel.
	KiroHome string
	// LaunchShell is the shell sops exec-env runs MCP commands with; install
	// picks sopswrap.Cmd on Windows. Goldens always use POSIX (the zero value).
	LaunchShell sopswrap.Shell
}

// RenderTeam renders teamsRoot/team into outDir. globalRoot is the root of
// global/steerings; envVersionsPath is the central env/mcp-versions.yaml.
// kiroHome overrides Context.KiroHome; an empty string uses
// DefaultKiroHomeSentinel. MCP launches are built for a POSIX shell.
func RenderTeam(teamsRoot, team, globalRoot, envVersionsPath, outDir, kiroHome string) error {
	return RenderTeamWith(teamsRoot, team, globalRoot, envVersionsPath, outDir, Options{KiroHome: kiroHome, LaunchShell: sopswrap.POSIX})
}

// RenderTeamWith is RenderTeam with every per-machine input in opts.
func RenderTeamWith(teamsRoot, team, globalRoot, envVersionsPath, outDir string, opts Options) error {
	ctx, err := BuildContext(teamsRoot, team, globalRoot, envVersionsPath, opts.KiroHome)
	if err != nil {
		return err
	}
	ctx.LaunchShell = opts.LaunchShell
	funcMap := FuncMap(ctx)

	teamDir := filepath.Join(teamsRoot, team)
	externalNames, err := externalskills.Resolve(teamDir, ctx.Team.ExternalSkills)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if ctx.Agents, ctx.Skills, err = teamInventory(teamDir, externalNames); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	if err := checkOwnerKeys(ctx); err != nil {
		return err
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
		if strings.HasPrefix(outRel, "agents/") {
			if err := checkSubagents(content, ctx); err != nil {
				return fmt.Errorf("render: %s: %w", relPath, err)
			}
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

// Publish moves tmpDir into place at outDir the way RenderTeam publishes its
// own output: an existing outDir that would change is moved to a timestamped
// backup first. tmpDir must sit beside outDir (same file system).
func Publish(tmpDir, outDir string) error { return publish(tmpDir, outDir) }

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
		var unsafe *sopswrap.UnsafeArgError
		if errors.As(err, &unsafe) {
			return nil, fmt.Errorf("render: %s/%s: %w", ctx.Team.Team, agentOf(relPath), unsafe)
		}
		return nil, fmt.Errorf("render: execute %s: %w", relPath, err)
	}
	return buf.Bytes(), nil
}

// agentOf is the agent a team-relative source path belongs to
// (agents/<agent>/...), else the path itself.
func agentOf(relPath string) string {
	if segments := strings.Split(relPath, "/"); len(segments) > 2 && segments[0] == "agents" {
		return segments[1]
	}
	return relPath
}

// teamInventory returns the sorted names of the team's agents (folders under
// agents/) and skills (folders under skills/ plus the vendored names).
func teamInventory(teamDir string, externalNames []string) (agents, skills []string, err error) {
	dirs := func(category string) ([]string, error) {
		entries, err := os.ReadDir(filepath.Join(teamDir, category))
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		return names, nil
	}
	if agents, err = dirs("agents"); err != nil {
		return nil, nil, err
	}
	if skills, err = dirs("skills"); err != nil {
		return nil, nil, err
	}
	skills = append(skills, externalNames...)
	slices.Sort(agents)
	slices.Sort(skills)
	return agents, slices.Compact(skills), nil
}

// checkOwnerKeys fails on an owners entry that names an agent or skill the
// team does not have, which is most likely a typo that would leave the real
// item with the team owner.
func checkOwnerKeys(ctx *Context) error {
	for _, group := range []struct {
		kind   string
		owners map[string]string
		have   []string
	}{
		{manifest.KindAgent, ctx.Team.Owners.Agents, ctx.Agents},
		{manifest.KindSkill, ctx.Team.Owners.Skills, ctx.Skills},
	} {
		for _, name := range slices.Sorted(maps.Keys(group.owners)) {
			if !slices.Contains(group.have, name) {
				return fmt.Errorf("render: manifest.yaml: owners.%ss.%s: %s has no %s %s", group.kind, name, ctx.Team.Team, group.kind, name)
			}
		}
	}
	return nil
}

// checkSubagents fails when an agent's subagent settings, or the match of a
// subagent permission rule (which Kiro compares with the agent name), name an
// agent the team does not have. kiro-cli accepts any name there, so a typo
// would only show up as a subagent that never runs or keeps prompting.
// Entries with glob characters are patterns and are not checked.
func checkSubagents(content []byte, ctx *Context) error {
	var doc struct {
		ToolsSettings struct {
			Subagent struct {
				AvailableAgents []string `json:"availableAgents"`
				TrustedAgents   []string `json:"trustedAgents"`
			} `json:"subagent"`
		} `json:"toolsSettings"`
		Permissions struct {
			Rules []struct {
				Capability string   `json:"capability"`
				Match      []string `json:"match"`
			} `json:"rules"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(content, &doc); err != nil {
		return err
	}
	type field struct {
		name  string
		names []string
	}
	sub := doc.ToolsSettings.Subagent
	fields := []field{{"toolsSettings.subagent.availableAgents", sub.AvailableAgents}, {"toolsSettings.subagent.trustedAgents", sub.TrustedAgents}}
	for i, r := range doc.Permissions.Rules {
		if r.Capability == "subagent" {
			fields = append(fields, field{fmt.Sprintf("permissions.rules[%d].match", i), r.Match})
		}
	}
	for _, field := range fields {
		for _, n := range field.names {
			if !strings.ContainsAny(n, "*?[") && !slices.Contains(ctx.Agents, n) {
				return fmt.Errorf("%s names %q, which is not an agent of %s", field.name, n, ctx.Team.Team)
			}
		}
	}
	return nil
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
