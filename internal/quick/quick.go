// Package quick renders a team's Amazon Quick bundle: the skills and agents
// its manifest lists under quick, in the formats Quick imports with
// Customize > Skills > From folder and Customize > Agents > From file.
//
// Quick has no API, so nothing here installs anything. Skill ids are assigned
// by Quick on the first import; quick-ids.yaml (see LoadIDs and RecordIDs)
// carries them back so a reimport overwrites the same skill and keeps the
// agents' pinned skills.
package quick

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/mavogel/kirokit/internal/manifest"
	"github.com/mavogel/kirokit/internal/render"
	"gopkg.in/yaml.v3"
)

// MaxPromptChars is Quick's limit for an agent's instructions.
const MaxPromptChars = 50000

// ErrNoQuick is returned for a team whose manifest lists nothing for Quick.
var ErrNoQuick = errors.New("lists nothing for Quick")

// Options are the per-account inputs of a Quick render.
type Options struct {
	// IDs maps a skill's frontmatter name to the id Quick assigned it.
	IDs map[string]string
}

// Result counts what a render wrote. MissingIDs are the frontmatter names of
// rendered skills with no id yet, which is normal before the first import.
type Result struct {
	Skills, Agents int
	MissingIDs     []string
}

// RenderTeam renders teamsRoot/team's Quick bundle into outDir:
// skills/<dir>/SKILL.md (plus the skill folder's other files) per listed
// skill and agents/<name>.json per listed agent. The Kiro tree is rendered
// into scratch first, with the KiroHome sentinel, so the bundle holds exactly
// the text Kiro gets.
func RenderTeam(teamsRoot, team, globalRoot, envVersionsPath, outDir string, opts Options) (Result, error) {
	m, err := manifest.Load(filepath.Join(teamsRoot, team, "manifest.yaml"))
	if err != nil {
		return Result{}, err
	}
	q := m.Quick
	if q == nil || (len(q.Skills) == 0 && len(q.Agents) == 0) {
		return Result{}, fmt.Errorf("team %s %w", team, ErrNoQuick)
	}
	for name, id := range opts.IDs {
		if err := checkID(name, id); err != nil {
			return Result{}, fmt.Errorf("quick: ids: %w", err)
		}
	}

	scratch, err := os.MkdirTemp("", "kirokit-quick-*")
	if err != nil {
		return Result{}, fmt.Errorf("quick: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	kiroTree := filepath.Join(scratch, "kiro")
	if err := render.RenderTeamWith(teamsRoot, team, globalRoot, envVersionsPath, kiroTree, render.Options{}); err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(outDir), 0o755); err != nil {
		return Result{}, fmt.Errorf("quick: %w", err)
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(outDir), ".kirokit-quick-*")
	if err != nil {
		return Result{}, fmt.Errorf("quick: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	var res Result
	skillNames := make(map[string]string, len(q.Skills)) // skill folder -> frontmatter name
	for _, dir := range q.Skills {
		name, hasID, err := writeSkill(kiroTree, tmpDir, dir, opts.IDs)
		if err != nil {
			return Result{}, fmt.Errorf("quick: skill %q: %w", dir, err)
		}
		skillNames[dir] = name
		if !hasID {
			res.MissingIDs = append(res.MissingIDs, name)
		}
		res.Skills++
	}
	for _, a := range q.Agents {
		if err := writeAgent(kiroTree, tmpDir, a, skillNames, opts.IDs); err != nil {
			return Result{}, fmt.Errorf("quick: agent %q: %w", a.Name, err)
		}
		res.Agents++
	}
	if err := render.Publish(tmpDir, outDir); err != nil {
		return Result{}, err
	}
	return res, nil
}

// writeSkill copies the rendered skill folder dir into the bundle, with the
// id line set from ids. It returns the skill's frontmatter name and whether
// it has an id.
func writeSkill(kiroTree, bundle, dir string, ids map[string]string) (name string, hasID bool, err error) {
	src := filepath.Join(kiroTree, "skills", dir)
	data, err := os.ReadFile(filepath.Join(src, "SKILL.md"))
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, fmt.Errorf("the team has no skill %q (listed in quick.skills)", dir)
	}
	if err != nil {
		return "", false, err
	}
	fm, body, ok := splitFrontmatter(data)
	if !ok {
		return "", false, fmt.Errorf("SKILL.md has no frontmatter; Quick needs name and description")
	}
	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(fm, &meta); err != nil {
		return "", false, fmt.Errorf("SKILL.md frontmatter: %w", err)
	}
	if strings.TrimSpace(meta.Name) == "" || strings.TrimSpace(meta.Description) == "" {
		return "", false, fmt.Errorf("SKILL.md frontmatter needs name and description; Quick needs both")
	}
	id, hasID := ids[meta.Name]

	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if rel == "SKILL.md" {
			content = withID(fm, body, id)
		}
		out := filepath.Join(bundle, "skills", dir, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, content, 0o600)
	})
	return meta.Name, hasID, err
}

// splitFrontmatter splits "---\n<fm>---\n<body>"; fm keeps its last newline.
func splitFrontmatter(data []byte) (fm, body []byte, ok bool) {
	rest, found := bytes.CutPrefix(data, []byte("---\n"))
	if !found {
		return nil, nil, false
	}
	if bytes.HasPrefix(rest, []byte("---\n")) {
		return nil, rest[4:], true
	}
	i := bytes.Index(rest, []byte("\n---\n"))
	if i < 0 {
		if !bytes.HasSuffix(rest, []byte("\n---")) {
			return nil, nil, false
		}
		return rest[:len(rest)-3], nil, true
	}
	return rest[:i+1], rest[i+5:], true
}

// withID rebuilds a SKILL.md with id as its last frontmatter line, where
// Quick writes it on export. A top-level id line the skill already carries is
// dropped: ids belong to one Quick account and come only from the ids file.
func withID(fm, body []byte, id string) []byte {
	var out bytes.Buffer
	out.WriteString("---\n")
	for line := range bytes.Lines(fm) {
		if bytes.HasPrefix(line, []byte("id:")) {
			continue
		}
		out.Write(line)
	}
	if id != "" {
		out.WriteString("id: " + id + "\n")
	}
	out.WriteString("---\n")
	out.Write(body)
	return out.Bytes()
}

// kiroAgent is what a Quick agent takes from its rendered Kiro agent.
type kiroAgent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// quickAgent is Quick's agent file. The field order is the export's order.
type quickAgent struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	SystemPrompt   string   `json:"system_prompt"`
	WelcomeMessage string   `json:"welcome_message"`
	StarterPrompts []string `json:"starter_prompts"`
	Capabilities   []any    `json:"capabilities"`
	Connectors     []any    `json:"connectors"`
	Skills         []string `json:"skills"`
	Spaces         []any    `json:"spaces"`
	Files          []any    `json:"files"`
}

func writeAgent(kiroTree, bundle string, a manifest.QuickAgent, skillNames, ids map[string]string) error {
	data, err := os.ReadFile(filepath.Join(kiroTree, "agents", a.Name+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the team has no agent %q (listed in quick.agents)", a.Name)
	}
	if err != nil {
		return err
	}
	var k kiroAgent
	if err := json.Unmarshal(data, &k); err != nil {
		return fmt.Errorf("rendered agent: %w", err)
	}
	prompt, err := resolvePrompt(kiroTree, k.Prompt)
	if err != nil {
		return err
	}
	if n := utf8.RuneCountInString(prompt); n > MaxPromptChars {
		return fmt.Errorf("prompt has %d characters; Quick allows %d", n, MaxPromptChars)
	}

	pinned := []string{}
	for _, s := range a.Skills {
		name, listed := skillNames[s]
		if !listed {
			return fmt.Errorf("pinned skill %q is not in quick.skills", s)
		}
		if id := ids[name]; id != "" {
			pinned = append(pinned, id)
		}
	}
	starters := slices.Clone(a.StarterPrompts)
	if starters == nil {
		starters = []string{}
	}
	out := quickAgent{
		Name: k.Name, Description: k.Description, SystemPrompt: prompt,
		WelcomeMessage: a.WelcomeMessage, StarterPrompts: starters,
		Capabilities: []any{}, Connectors: []any{}, Skills: pinned, Spaces: []any{}, Files: []any{},
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	path := filepath.Join(bundle, "agents", a.Name+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

// resolvePrompt returns an inline prompt as is, and reads a file:// prompt
// from the rendered tree. A relative path is relative to agents/, as Kiro
// reads it; an absolute one must sit under the KiroHome sentinel. Either way
// the file must stay inside the rendered tree.
func resolvePrompt(kiroTree, prompt string) (string, error) {
	ref, isFile := strings.CutPrefix(prompt, "file://")
	if !isFile {
		return prompt, nil
	}
	var p string
	switch {
	case strings.HasPrefix(ref, render.DefaultKiroHomeSentinel+"/"):
		p = filepath.Join(kiroTree, filepath.FromSlash(strings.TrimPrefix(ref, render.DefaultKiroHomeSentinel+"/")))
	case strings.HasPrefix(ref, "/"), strings.HasPrefix(ref, "~"):
		return "", fmt.Errorf("prompt %s leaves the rendered tree", prompt)
	default:
		p = filepath.Join(kiroTree, "agents", filepath.FromSlash(ref))
	}
	rel, err := filepath.Rel(kiroTree, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("prompt %s leaves the rendered tree", prompt)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("prompt %s: %w", prompt, err)
	}
	return string(data), nil
}
