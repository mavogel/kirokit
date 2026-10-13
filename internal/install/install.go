// Package install implements `kirokit install`/`uninstall`: rendering a
// team with its real $KIRO_HOME baked in, copying the result (plus each
// agent's SOPS secrets file) into that location, and tracking exactly what
// was written so uninstall can remove only those files.
package install

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mavogel/kiro-cli-kit/internal/render"
)

// Roots are the shared repository-relative paths Install needs to render a
// team's templates.
type Roots struct {
	TeamsRoot       string
	GlobalRoot      string
	EnvVersionsPath string
}

// KnowledgeBase describes an ad-hoc local knowledgeBase resource to attach
// to one installed agent: Source is symlinked at Mount (never copied), and
// a corresponding resources entry is added to Agent's installed JSON.
type KnowledgeBase struct {
	Agent  string
	Source string
	Mount  string
}

// manifest records every path a single team's install wrote, so uninstall
// can remove exactly those paths and nothing else.
type manifest struct {
	Files    []string `json:"files"`
	Symlinks []string `json:"symlinks,omitempty"`
}

// ResolveKiroHome returns $KIRO_HOME if set, otherwise Kiro CLI's own
// documented default (~/.kiro).
func ResolveKiroHome() (string, error) {
	if v := os.Getenv("KIRO_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("install: resolve $HOME: %w", err)
	}
	return filepath.Join(home, ".kiro"), nil
}

// Install renders team with KiroHome=kiroHome and copies the rendered
// output plus each agent's SOPS secrets file into kiroHome, then writes an
// install manifest recording every path it wrote. When kb is non-nil, it
// also attaches a local knowledgeBase resource to the named agent (see
// AttachKnowledgeBase).
func Install(roots Roots, team, kiroHome string, kb *KnowledgeBase) error {
	scratch, err := os.MkdirTemp("", "kirokit-install-*")
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	renderedDir := filepath.Join(scratch, "rendered")
	if err := render.RenderTeam(roots.TeamsRoot, team, roots.GlobalRoot, roots.EnvVersionsPath, renderedDir, kiroHome); err != nil {
		return err
	}

	var written []string
	var symlinks []string

	writtenFromRender, err := copyRenderedTree(renderedDir, kiroHome)
	if err != nil {
		return err
	}
	written = append(written, writtenFromRender...)

	writtenSecrets, err := copySecrets(roots.TeamsRoot, team, kiroHome)
	if err != nil {
		return err
	}
	written = append(written, writtenSecrets...)

	if kb != nil {
		agentJSONPath := filepath.Join(kiroHome, "agents", kb.Agent+".json")
		if err := AttachKnowledgeBase(agentJSONPath, kb.Mount, kb.Source); err != nil {
			return err
		}
		symlinks = append(symlinks, kb.Mount)
	}

	manifestPath := installManifestPath(kiroHome, team)
	written = append(written, manifestPath)

	if err := writeManifest(manifestPath, manifest{Files: written, Symlinks: symlinks}); err != nil {
		return err
	}
	return nil
}

// AttachKnowledgeBase creates mountPoint as a symlink to sourcePath (never
// copying its contents) and adds a corresponding knowledgeBase resource
// entry to the agent config at agentJSONPath. The agent name is derived
// from agentJSONPath's base filename.
func AttachKnowledgeBase(agentJSONPath, mountPoint, sourcePath string) error {
	if err := ensureKnowledgeBaseSymlink(mountPoint, sourcePath); err != nil {
		return err
	}
	return addKnowledgeBaseResource(agentJSONPath, mountPoint)
}

func ensureKnowledgeBaseSymlink(mountPoint, sourcePath string) error {
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("install: resolve knowledge base source %s: %w", sourcePath, err)
	}
	info, err := os.Stat(absSource)
	if err != nil {
		return fmt.Errorf("install: knowledge base source %s: %w", sourcePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("install: knowledge base source %s is not a directory", sourcePath)
	}

	if lst, err := os.Lstat(mountPoint); err == nil {
		if lst.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("install: knowledge base mount point %s already exists and is not a symlink kirokit manages; refusing to overwrite", mountPoint)
		}
		if err := os.Remove(mountPoint); err != nil {
			return fmt.Errorf("install: remove stale knowledge base symlink %s: %w", mountPoint, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("install: stat knowledge base mount point %s: %w", mountPoint, err)
	}

	if err := os.MkdirAll(filepath.Dir(mountPoint), 0o755); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := os.Symlink(absSource, mountPoint); err != nil {
		return fmt.Errorf("install: create knowledge base symlink %s -> %s: %w", mountPoint, absSource, err)
	}
	return nil
}

func addKnowledgeBaseResource(agentJSONPath, mountPoint string) error {
	data, err := os.ReadFile(agentJSONPath)
	if err != nil {
		return fmt.Errorf("install: read %s: %w", agentJSONPath, err)
	}

	var agent map[string]json.RawMessage
	if err := json.Unmarshal(data, &agent); err != nil {
		return fmt.Errorf("install: parse %s: %w", agentJSONPath, err)
	}

	var resources []json.RawMessage
	if raw, ok := agent["resources"]; ok {
		if err := json.Unmarshal(raw, &resources); err != nil {
			return fmt.Errorf("install: parse %q in %s: %w", "resources", agentJSONPath, err)
		}
	}

	agentName := strings.TrimSuffix(filepath.Base(agentJSONPath), ".json")
	// indexType and autoUpdate are the documented knowledgeBase resource
	// fields; include (a list of globs, with exclude as its counterpart) is
	// accepted by kiro-cli's own parser but absent from the agent-config docs.
	newResource, err := json.Marshal(map[string]any{
		"type":       "knowledgeBase",
		"source":     "file://" + mountPoint,
		"name":       agentName + "-kb",
		"indexType":  "best",
		"include":    []string{"**/*.md"},
		"autoUpdate": true,
	})
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	resources = append(resources, newResource)

	resourcesJSON, err := json.Marshal(resources)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	agent["resources"] = resourcesJSON

	out, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("install: marshal %s: %w", agentJSONPath, err)
	}
	if err := os.WriteFile(agentJSONPath, out, 0o600); err != nil {
		return fmt.Errorf("install: write %s: %w", agentJSONPath, err)
	}
	return nil
}

// Uninstall removes every path recorded in team's install manifest under
// kiroHome, then the manifest itself. It errors clearly if no manifest
// exists for that team.
func Uninstall(kiroHome, team string) error {
	manifestPath := installManifestPath(kiroHome, team)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("install: no install manifest found for team %q under %s; nothing to uninstall", team, kiroHome)
		}
		return fmt.Errorf("install: read %s: %w", manifestPath, err)
	}

	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("install: parse %s: %w", manifestPath, err)
	}

	for _, symlink := range m.Symlinks {
		if err := os.Remove(symlink); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("install: remove symlink %s: %w", symlink, err)
		}
	}
	for _, f := range m.Files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("install: remove %s: %w", f, err)
		}
	}
	return nil
}

func installManifestPath(kiroHome, team string) string {
	return filepath.Join(kiroHome, ".kirokit-installed", team+".json")
}

func writeManifest(path string, m manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("install: marshal manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("install: write %s: %w", path, err)
	}
	return nil
}

func copyRenderedTree(renderedDir, kiroHome string) ([]string, error) {
	var written []string
	err := filepath.WalkDir(renderedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(renderedDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(kiroHome, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("install: read %s: %w", rel, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("install: %w", err)
		}
		if err := os.WriteFile(target, data, 0o600); err != nil {
			return fmt.Errorf("install: write %s: %w", target, err)
		}
		written = append(written, target)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return written, nil
}

// copySecrets copies every agent's SOPS-encrypted secrets file (when
// present) to kiroHome/kirokit-secrets/<team>/<agent>/secrets.enc.yaml.
func copySecrets(teamsRoot, team, kiroHome string) ([]string, error) {
	agentsDir := filepath.Join(teamsRoot, team, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("install: read %s: %w", agentsDir, err)
	}

	var written []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		agent := e.Name()
		src := filepath.Join(agentsDir, agent, "env", "secrets.enc.yaml")
		data, err := os.ReadFile(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("install: read %s: %w", src, err)
		}

		dst := filepath.Join(kiroHome, "kirokit-secrets", team, agent, "secrets.enc.yaml")
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, fmt.Errorf("install: %w", err)
		}
		if err := os.WriteFile(dst, data, 0o600); err != nil {
			return nil, fmt.Errorf("install: write %s: %w", dst, err)
		}
		written = append(written, dst)
	}
	return written, nil
}
