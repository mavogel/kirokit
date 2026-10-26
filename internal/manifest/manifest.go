// Package manifest parses a team's manifest.yaml: its identity, which
// global steerings it opts into, and the external skills it vendors.
package manifest

import (
	"fmt"
	"maps"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest is a team's manifest.yaml. Owner, Owners and Feedback are
// optional; they stay out of the golden context.json when unset.
type Manifest struct {
	Team string `yaml:"team"`
	// Owner is the team owner, a GitHub handle (@user or @org/team), who owns
	// every agent and skill without an entry in Owners.
	Owner           string          `yaml:"owner" json:",omitempty"`
	Owners          Owners          `yaml:"owners" json:",omitzero"`
	Feedback        Feedback        `yaml:"feedback" json:",omitzero"`
	GlobalSteerings []string        `yaml:"globalSteerings"`
	ExternalSkills  []ExternalSkill `yaml:"externalSkills" json:",omitempty"`
}

// Owners overrides the team owner for single agents and skills, by name.
type Owners struct {
	Agents map[string]string `yaml:"agents" json:",omitempty"`
	Skills map[string]string `yaml:"skills" json:",omitempty"`
}

// Feedback names where feedback on the team's agents and skills goes:
// GitHub is the owner/repo whose issues receive it.
type Feedback struct {
	GitHub string `yaml:"github" json:",omitempty"`
}

// The kinds OwnerOf takes.
const (
	KindAgent = "agent"
	KindSkill = "skill"
)

// OwnerOf returns the owner of the agent or skill name (kind is KindAgent or
// KindSkill) and whether that owner is the team owner. It is "" when neither
// the item nor the team names one.
func (m Manifest) OwnerOf(kind, name string) (owner string, fromTeam bool) {
	overrides := m.Owners.Agents
	if kind == KindSkill {
		overrides = m.Owners.Skills
	}
	if o, ok := overrides[name]; ok {
		return o, false
	}
	if m.Owner == "" {
		return "", false
	}
	return m.Owner, true
}

var (
	githubHandle = regexp.MustCompile(`^@[A-Za-z0-9][A-Za-z0-9-]{0,38}(?:/[A-Za-z0-9._-]+)?$`)
	githubRepo   = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`)
)

// validateOwnership checks every handle and the feedback repository, naming
// the field that is wrong. Map keys are checked in sorted order so the same
// manifest always reports the same error.
func (m Manifest) validateOwnership() error {
	checkHandle := func(field, v string) error {
		if !githubHandle.MatchString(v) {
			return fmt.Errorf("%s: %q is not a GitHub handle (write @%s)", field, v, suggestHandle(v))
		}
		return nil
	}
	if m.Owner != "" {
		if err := checkHandle("owner", m.Owner); err != nil {
			return err
		}
	}
	for _, group := range []struct {
		field string
		names map[string]string
	}{{"owners.agents", m.Owners.Agents}, {"owners.skills", m.Owners.Skills}} {
		for _, name := range slices.Sorted(maps.Keys(group.names)) {
			if err := checkHandle(group.field+"."+name, group.names[name]); err != nil {
				return err
			}
		}
	}
	if r := m.Feedback.GitHub; r != "" && (!githubRepo.MatchString(r) || path.Base(r) == "." || path.Base(r) == "..") {
		return fmt.Errorf("feedback.github: %q is not a GitHub repository (write owner/repo)", r)
	}
	return nil
}

// suggestHandle is the example shown after "write @": the value without
// leading @ signs, or "jane" when nothing usable is left.
func suggestHandle(v string) string {
	s := strings.TrimLeft(v, "@")
	if !githubHandle.MatchString("@" + s) {
		return "jane"
	}
	return s
}

// ExternalSkill declares skill directories of a git repository that the team
// vendors: Source is a git URL, Ref a tag, branch or commit, and Paths the
// skills' directories inside the repository.
//
// Load expands an entry into one entry per path, each sharing Source and Ref,
// so code past Load sees only Path (and never Paths). Name overrides the
// default name (the last element of Path) and needs a single path.
//
// With All set an expanded entry is a collection: Path is a directory of
// skills, and every immediate subdirectory holding a SKILL.md is vendored,
// except those named in Exclude (which needs a single path).
type ExternalSkill struct {
	Name    string   `yaml:"name"`
	Source  string   `yaml:"source"`
	Ref     string   `yaml:"ref"`
	Paths   []string `yaml:"paths" json:",omitempty"`
	All     bool     `yaml:"all" json:",omitempty"`
	Exclude []string `yaml:"exclude" json:",omitempty"`

	// Path is set by Expand, never read from the YAML.
	Path string `yaml:"-"`
}

// IsCollection reports whether the entry vendors every skill under Path.
func (s ExternalSkill) IsCollection() bool { return s.All }

// EffectiveName is the skill's directory name in the output: Name, or the
// last element of Path.
func (s ExternalSkill) EffectiveName() string {
	if s.Name != "" {
		return s.Name
	}
	return path.Base(s.Path)
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
	if err := m.validateOwnership(); err != nil {
		return nil, fmt.Errorf("manifest: %s: %w", path, err)
	}
	expanded, err := expandExternalSkills(m.ExternalSkills)
	if err != nil {
		return nil, fmt.Errorf("manifest: %s: %w", path, err)
	}
	m.ExternalSkills = expanded
	return &m, nil
}

// Expand returns the entries a Paths entry stands for, one per path, or the
// entry itself when it has no paths.
func (s ExternalSkill) Expand() []ExternalSkill {
	if len(s.Paths) == 0 {
		return []ExternalSkill{s}
	}
	out := make([]ExternalSkill, 0, len(s.Paths))
	for _, p := range s.Paths {
		out = append(out, ExternalSkill{Name: s.Name, Source: s.Source, Ref: s.Ref, Path: p, All: s.All, Exclude: s.Exclude})
	}
	return out
}

// expandExternalSkills validates every entry, expands paths entries, and
// rejects two single skills sharing a name.
func expandExternalSkills(skills []ExternalSkill) ([]ExternalSkill, error) {
	seen := make(map[string]bool, len(skills))
	var out []ExternalSkill
	for i, s := range skills {
		label := fmt.Sprintf("externalSkills[%d]", i)
		if err := s.Validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", label, err)
		}
		for _, e := range s.Expand() {
			out = append(out, e)
			if e.IsCollection() {
				// A collection's skill names are known only after a fetch; the
				// lock check reports clashes between them.
				continue
			}
			name := e.EffectiveName()
			if seen[name] {
				return nil, fmt.Errorf("%s: duplicate skill name %q", label, name)
			}
			seen[name] = true
		}
	}
	return out, nil
}

// Validate checks one entry on its own: required fields, a safe path and
// name, a source and ref that are safe to hand to git, and no credentials in
// the source URL. Load calls it for every entry; code that acts on an entry
// (such as a fetch) calls it again rather than trusting its caller.
func (s ExternalSkill) Validate() error {
	if len(s.Paths) > 0 {
		if len(s.Paths) > 1 && s.Name != "" {
			return fmt.Errorf("name needs a single path (each skill keeps its directory name)")
		}
		if len(s.Paths) > 1 && len(s.Exclude) > 0 {
			return fmt.Errorf("exclude needs a single path")
		}
		for _, e := range s.Expand() {
			if err := e.Validate(); err != nil {
				return err
			}
		}
		return nil
	}
	for field, value := range map[string]string{"source": s.Source, "ref": s.Ref, "paths": s.Path} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("missing required field %q", field)
		}
	}
	if err := validateSkillPath(s.Path); err != nil {
		return err
	}
	name := s.EffectiveName()
	if s.IsCollection() {
		if s.Name != "" {
			return fmt.Errorf("name cannot be set on an entry with all: true (each skill keeps its directory name)")
		}
		for _, x := range s.Exclude {
			if x == "" || x == "." || x == ".." || strings.ContainsAny(x, `/\`) {
				return fmt.Errorf("invalid exclude %q (exclude names skill directories, not paths)", x)
			}
		}
	} else {
		if len(s.Exclude) > 0 {
			return fmt.Errorf("exclude requires all: true")
		}
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
			return fmt.Errorf("invalid skill name %q (name must be a single path element)", name)
		}
	}
	if err := validateSource(s.Source); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := validateRef(s.Ref); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func validateSkillPath(p string) error {
	if path.IsAbs(p) || strings.Contains(p, `\`) {
		return fmt.Errorf("path %q must be a relative, slash-separated path", p)
	}
	clean := path.Clean(p)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path %q must name a directory inside the repository", p)
	}
	return nil
}

// gitSchemes are the URL schemes a source may use. Anything else, notably
// git's transport helpers (ext::, fd::), is refused.
var gitSchemes = map[string]bool{"https": true, "http": true, "ssh": true, "git": true, "file": true}

// validateSource keeps a source safe to pass to git: it must not be parsed as
// an option (leading dash), must not use a transport helper (`<name>::...`),
// and must not carry credentials (any userinfo on http(s), or a password on
// any other scheme). Schemeless sources (scp-style git@host:org/repo, or a
// local path) are otherwise accepted. The offending value is not echoed when
// it may hold a credential.
func validateSource(source string) error {
	if strings.HasPrefix(source, "-") {
		return fmt.Errorf("source %q must not start with a dash", source)
	}
	if helper, _, ok := strings.Cut(source, "::"); ok && !strings.ContainsAny(helper, "/@ ") {
		return fmt.Errorf("source uses the git transport helper %q, which is not allowed", helper+"::")
	}
	if !strings.Contains(source, "://") {
		return nil
	}
	u, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("source is not a valid URL")
	}
	if !gitSchemes[u.Scheme] {
		return fmt.Errorf("source scheme %q is not allowed (use https, ssh, git, or file)", u.Scheme)
	}
	if u.User == nil {
		return nil
	}
	if _, hasPassword := u.User.Password(); hasPassword || u.Scheme == "http" || u.Scheme == "https" {
		return fmt.Errorf("source must not embed credentials; authenticate through git (SSH or a credential helper)")
	}
	return nil
}

// validateRef keeps a ref from being parsed by git as an option and rejects
// characters that cannot appear in a ref name.
func validateRef(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("ref %q must not start with a dash", ref)
	}
	for _, r := range ref {
		if r <= ' ' || r == 0x7f {
			return fmt.Errorf("ref %q contains whitespace or a control character", ref)
		}
	}
	return nil
}
