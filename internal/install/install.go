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
	"sort"
	"strings"

	"github.com/mavogel/kiro-cli-kit/internal/backup"
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

	planned, err := planRenderedTree(renderedDir, kiroHome)
	if err != nil {
		return err
	}
	secrets, err := planSecrets(roots.TeamsRoot, team, kiroHome)
	if err != nil {
		return err
	}
	planned = append(planned, secrets...)

	// The knowledge-base resource is part of what will be on disk, so it must be
	// in the plan: otherwise every re-install would see the agent file as
	// changed (its attached resource missing from the plan) and back it up.
	if kb != nil {
		if planned, err = withKnowledgeBase(planned, filepath.Join(kiroHome, "agents", kb.Agent+".json"), kb); err != nil {
			return err
		}
	}

	// Everything is planned before anything is written, so the backup below
	// sees the folders exactly as the user left them.
	if err := backupChangedFolders(kiroHome, planned); err != nil {
		return err
	}
	writtenFiles, err := writePlanned(planned)
	if err != nil {
		return err
	}
	written = append(written, writtenFiles...)

	if kb != nil {
		if err := ensureKnowledgeBaseSymlink(kb.Mount, kb.Source); err != nil {
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
	agentName := strings.TrimSuffix(filepath.Base(agentJSONPath), ".json")
	out, err := withKnowledgeBaseResource(data, agentName, mountPoint)
	if err != nil {
		return fmt.Errorf("install: %s: %w", agentJSONPath, err)
	}
	if err := os.WriteFile(agentJSONPath, out, 0o600); err != nil {
		return fmt.Errorf("install: write %s: %w", agentJSONPath, err)
	}
	return nil
}

// withKnowledgeBaseResource returns the agent config data with a local
// knowledgeBase resource for mountPoint appended to its resources. It is a
// pure function of its input, so install can apply it before deciding whether
// anything on disk would change.
func withKnowledgeBaseResource(data []byte, agentName, mountPoint string) ([]byte, error) {
	var agent map[string]json.RawMessage
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var resources []json.RawMessage
	if raw, ok := agent["resources"]; ok {
		if err := json.Unmarshal(raw, &resources); err != nil {
			return nil, fmt.Errorf("parse %q: %w", "resources", err)
		}
	}

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
		return nil, err
	}
	resources = append(resources, newResource)

	resourcesJSON, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}
	agent["resources"] = resourcesJSON

	out, err := json.Marshal(agent)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return out, nil
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

// plannedFile is one file install is about to write.
type plannedFile struct {
	target string
	data   []byte
	perm   os.FileMode
}

// planRenderedTree lists every file of the rendered tree with its destination
// under kiroHome. Files stay private to the user (0600), but a script (for
// example one a vendored skill ships) keeps its owner-execute bit.
func planRenderedTree(renderedDir, kiroHome string) ([]plannedFile, error) {
	var planned []plannedFile
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
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("install: read %s: %w", rel, err)
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("install: stat %s: %w", rel, err)
		}
		perm := os.FileMode(0o600)
		if info.Mode().Perm()&0o100 != 0 {
			perm = 0o700
		}
		planned = append(planned, plannedFile{target: filepath.Join(kiroHome, rel), data: data, perm: perm})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return planned, nil
}

// planSecrets lists every agent's SOPS-encrypted secrets file (when present)
// with its destination kiroHome/kirokit-secrets/<team>/<agent>/secrets.enc.yaml.
func planSecrets(teamsRoot, team, kiroHome string) ([]plannedFile, error) {
	agentsDir := filepath.Join(teamsRoot, team, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("install: read %s: %w", agentsDir, err)
	}

	var planned []plannedFile
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
		planned = append(planned, plannedFile{
			target: filepath.Join(kiroHome, "kirokit-secrets", team, agent, "secrets.enc.yaml"),
			data:   data,
			perm:   0o600,
		})
	}
	return planned, nil
}

// withKnowledgeBase returns planned with the knowledge-base resource added to
// the planned content of the agent config at agentJSONPath.
func withKnowledgeBase(planned []plannedFile, agentJSONPath string, kb *KnowledgeBase) ([]plannedFile, error) {
	for i, f := range planned {
		if f.target != agentJSONPath {
			continue
		}
		out, err := withKnowledgeBaseResource(f.data, kb.Agent, kb.Mount)
		if err != nil {
			return nil, fmt.Errorf("install: %s: %w", agentJSONPath, err)
		}
		planned[i].data = out
		return planned, nil
	}
	return nil, fmt.Errorf("install: knowledge base agent %q is not an agent of this team (no %s)", kb.Agent, agentJSONPath)
}

// backupChangedFolders copies each top-level folder of kiroHome (agents,
// prompts, skills, ...) in which an existing file would change to
// kiroHome/YYYYMMDD-HHmmss-<folder>.backup before anything is written (a file
// written directly into kiroHome is copied as itself). Folders
// with no change, and everything else in kiroHome, are not copied, so an
// unchanged re-install (a scheduled sync) makes no backup. The backups are
// not recorded in the install manifest, so uninstall leaves them alone.
func backupChangedFolders(kiroHome string, planned []plannedFile) error {
	changed := map[string]bool{}
	for _, f := range planned {
		if !backup.Differs(f.target, f.data) {
			continue
		}
		rel, err := filepath.Rel(kiroHome, f.target)
		if err != nil {
			return fmt.Errorf("install: %w", err)
		}
		// The first path element: a folder, or (for a file written straight into
		// kiroHome) the file itself, which CopyAside also handles.
		folder, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
		changed[folder] = true
	}

	folders := make([]string, 0, len(changed))
	for folder := range changed {
		folders = append(folders, folder)
	}
	sort.Strings(folders)

	stamp := backup.Stamp()
	for _, folder := range folders {
		if _, err := backup.CopyAside(filepath.Join(kiroHome, folder), stamp); err != nil {
			return fmt.Errorf("install: %w", err)
		}
	}
	return nil
}

// writePlanned writes every planned file, creating folders as needed, and
// returns the paths written.
func writePlanned(planned []plannedFile) ([]string, error) {
	written := make([]string, 0, len(planned))
	for _, f := range planned {
		if err := os.MkdirAll(filepath.Dir(f.target), 0o755); err != nil {
			return nil, fmt.Errorf("install: %w", err)
		}
		if err := os.WriteFile(f.target, f.data, f.perm); err != nil {
			return nil, fmt.Errorf("install: write %s: %w", f.target, err)
		}
		// WriteFile leaves the mode of an existing file alone.
		if err := os.Chmod(f.target, f.perm); err != nil {
			return nil, fmt.Errorf("install: chmod %s: %w", f.target, err)
		}
		written = append(written, f.target)
	}
	return written, nil
}
