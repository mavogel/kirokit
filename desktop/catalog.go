package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/mavogel/kirokit/internal/manifest"
	"github.com/mavogel/kirokit/internal/mcpversions"
	"github.com/mavogel/kirokit/internal/render"
)

// Catalog is what the Teams, Agents and Skills pages render. Tag is the
// release the clone sits on, empty when it is on a branch.
type Catalog struct {
	Teams []Team            `json:"teams"`
	Pins  map[string]string `json:"pins"`
	Tag   string            `json:"tag"`
}

// Team is one folder under teams/, as rendered. Feedback is the owner/repo
// whose GitHub issues receive feedback, empty when the manifest names none.
type Team struct {
	Name      string   `json:"name"`
	Agents    []Agent  `json:"agents"`
	Skills    []Skill  `json:"skills"`
	Prompts   int      `json:"prompts"`
	Steerings []string `json:"steerings"`
	Feedback  string   `json:"feedback"`
	// Quick is true when the manifest lists skills or agents for Amazon
	// Quick, so the card offers Export for Quick.
	Quick bool `json:"quick"`
}

// Agent is one rendered agent file. NeedsKey is true when the agent ships an
// encrypted secrets file. Owner is its GitHub handle, empty when neither the
// agent nor the team names one; OwnerFromTeam is true for the team owner.
type Agent struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
	Servers     []string `json:"servers"`
	// Pins are the keys of env/mcp-versions.yaml the agent's template resolves
	// with mcpVersion; a server name need not equal the key of its pin.
	Pins          []string `json:"pins"`
	NeedsKey      bool     `json:"needsKey"`
	Owner         string   `json:"owner"`
	OwnerFromTeam bool     `json:"ownerFromTeam"`
}

// Skill's Source is "Written by the team" or "<source> @ <ref>". Owner and
// OwnerFromTeam are as on Agent; a vendored skill has the team owner.
type Skill struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	Owner         string `json:"owner"`
	OwnerFromTeam bool   `json:"ownerFromTeam"`
}

const teamSkillSource = "Written by the team"

// BuildCatalog renders every team of the repository at repo into a temporary
// folder and reads the result. A team that fails to render fails the whole
// catalog with an error naming it, so a page shows the message instead of a
// partial table.
func BuildCatalog(repo string) (Catalog, error) {
	teamsRoot := filepath.Join(repo, "teams")
	globalRoot := filepath.Join(repo, "global", "steerings")
	versionsPath := filepath.Join(repo, "env", "mcp-versions.yaml")

	cat := Catalog{Teams: []Team{}, Pins: map[string]string{}, Tag: repoTag(repo)}
	if pins, err := mcpversions.Load(versionsPath); err == nil {
		cat.Pins = pins
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Catalog{}, err
	}

	entries, err := os.ReadDir(teamsRoot)
	if err != nil {
		return Catalog{}, fmt.Errorf("catalog: %w", err)
	}
	scratch, err := os.MkdirTemp("", "kirokit-catalog-*")
	if err != nil {
		return Catalog{}, fmt.Errorf("catalog: %w", err)
	}
	defer os.RemoveAll(scratch) //nolint:errcheck // a leftover temp folder is harmless

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(teamsRoot, e.Name(), "manifest.yaml")); err != nil {
			continue
		}
		out := filepath.Join(scratch, e.Name())
		if err := render.RenderTeam(teamsRoot, e.Name(), globalRoot, versionsPath, out, ""); err != nil {
			return Catalog{}, fmt.Errorf("team %s: %w", e.Name(), err)
		}
		team, err := readTeam(teamsRoot, e.Name(), out)
		if err != nil {
			return Catalog{}, fmt.Errorf("team %s: %w", e.Name(), err)
		}
		cat.Teams = append(cat.Teams, team)
	}
	return cat, nil
}

func readTeam(teamsRoot, name, rendered string) (Team, error) {
	teamDir := filepath.Join(teamsRoot, name)
	team := Team{Name: name, Agents: []Agent{}, Skills: []Skill{}, Steerings: []string{}}
	// The render before this call has already loaded this manifest, so an
	// error here means the clone changed in between.
	m, err := manifest.Load(filepath.Join(teamDir, "manifest.yaml"))
	if err != nil {
		return team, err
	}
	team.Feedback = m.Feedback.GitHub
	team.Quick = m.Quick != nil && (len(m.Quick.Skills) > 0 || len(m.Quick.Agents) > 0)

	agentFiles, _ := filepath.Glob(filepath.Join(rendered, "agents", "*.json"))
	sort.Strings(agentFiles)
	for _, f := range agentFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return team, err
		}
		var doc struct {
			Name        string                     `json:"name"`
			Description string                     `json:"description"`
			Tools       []string                   `json:"tools"`
			McpServers  map[string]json.RawMessage `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			return team, fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
		file := strings.TrimSuffix(filepath.Base(f), ".json")
		a := Agent{Name: doc.Name, Description: doc.Description, Tools: nonNil(doc.Tools), Servers: []string{}}
		if a.Name == "" {
			a.Name = file
		}
		for s := range doc.McpServers {
			a.Servers = append(a.Servers, s)
		}
		sort.Strings(a.Servers)
		a.Pins = templatePins(filepath.Join(teamDir, "agents", file, "agent.json.tmpl"))
		_, statErr := os.Stat(filepath.Join(teamDir, "agents", file, "env", "secrets.enc.yaml"))
		a.NeedsKey = statErr == nil
		// Owners are keyed by the agent's folder name, which the name field
		// in the file may differ from.
		a.Owner, a.OwnerFromTeam = m.OwnerOf(manifest.KindAgent, file)
		team.Agents = append(team.Agents, a)
	}

	lock, err := externalskills.ReadLock(externalskills.LockPath(teamDir))
	if err != nil {
		return team, err
	}
	external := map[string]string{}
	for _, le := range lock.Skills {
		external[le.Name] = le.Source + " @ " + le.Ref
	}
	if dirs, err := os.ReadDir(filepath.Join(rendered, "skills")); err == nil {
		for _, d := range dirs {
			if !d.IsDir() {
				continue
			}
			src, ok := external[d.Name()]
			if !ok {
				src = teamSkillSource
			}
			s := Skill{Name: d.Name(), Source: src}
			s.Owner, s.OwnerFromTeam = m.OwnerOf(manifest.KindSkill, d.Name())
			team.Skills = append(team.Skills, s)
		}
	}

	prompts, _ := filepath.Glob(filepath.Join(rendered, "prompts", "*"))
	team.Prompts = len(prompts)

	seen := map[string]bool{}
	if files, err := os.ReadDir(filepath.Join(rendered, "steering")); err == nil {
		for _, f := range files {
			n := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			seen[n] = true
			team.Steerings = append(team.Steerings, n)
		}
	}
	for _, g := range m.GlobalSteerings {
		if !seen[g] {
			team.Steerings = append(team.Steerings, g)
		}
	}
	sort.Strings(team.Steerings)
	return team, nil
}

var mcpVersionCall = regexp.MustCompile(`mcpVersion\s+"([^"]+)"`)

// templatePins returns the sorted, distinct keys the agent template passes to
// mcpVersion. An unreadable template has none.
func templatePins(path string) []string {
	pins := []string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return pins
	}
	for _, m := range mcpVersionCall.FindAllStringSubmatch(string(data), -1) {
		if !slices.Contains(pins, m[1]) {
			pins = append(pins, m[1])
		}
	}
	sort.Strings(pins)
	return pins
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// repoTag is the release tag the clone is on, empty on a branch or when git
// cannot say. GIT_* variables are dropped so git reads the clone itself.
func repoTag(repo string) string {
	cmd := exec.Command("git", "-C", repo, "describe", "--tags", "--exact-match", "HEAD")
	cmd.Env = withoutGitEnv(os.Environ())
	hideConsole(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func withoutGitEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if !strings.HasPrefix(kv, "GIT_") {
			out = append(out, kv)
		}
	}
	return out
}

// catalogCache holds the last build until a clone or sync ends.
type catalogCache struct {
	mu    sync.Mutex
	built bool
	cat   Catalog
	err   error
}

// Catalog returns the teams as rendered from the clone. The result is cached:
// it is built the first time it is asked for and again only after a clone or
// a sync ends. Before a clone it is empty. A render error comes back as the
// error, with its message.
func (a *App) Catalog() (Catalog, error) {
	a.cat.mu.Lock()
	defer a.cat.mu.Unlock()
	if !a.cat.built {
		a.cat.build(a.repoDir())
	}
	return a.cat.cat, a.cat.err
}

// refreshCatalog rebuilds the cached catalog; the code that ends a clone or a
// sync calls it.
func (a *App) refreshCatalog() {
	a.cat.mu.Lock()
	defer a.cat.mu.Unlock()
	a.cat.build(a.repoDir())
}

func (c *catalogCache) build(repo string) {
	c.built = true
	if _, err := os.Stat(filepath.Join(repo, "teams")); err != nil {
		c.cat, c.err = Catalog{Teams: []Team{}, Pins: map[string]string{}}, nil
		return
	}
	c.cat, c.err = BuildCatalog(repo)
}
