// Package manifest parses a team's manifest.yaml: its identity, which
// global steerings it opts into, and the external skills it vendors.
package manifest

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest is a team's manifest.yaml.
type Manifest struct {
	Team            string          `yaml:"team"`
	GlobalSteerings []string        `yaml:"globalSteerings"`
	ExternalSkills  []ExternalSkill `yaml:"externalSkills" json:",omitempty"`
}

// ExternalSkill declares one skill directory of a git repository that the
// team vendors: Source is a git URL, Ref a tag, branch or commit, and Path
// the skill's directory inside the repository. Name overrides the default
// name (the last element of Path).
//
// With All set the entry is a collection: Path is a directory of skills, and
// every immediate subdirectory holding a SKILL.md is vendored, except those
// named in Exclude.
type ExternalSkill struct {
	Name    string   `yaml:"name"`
	Source  string   `yaml:"source"`
	Ref     string   `yaml:"ref"`
	Path    string   `yaml:"path"`
	All     bool     `yaml:"all" json:",omitempty"`
	Exclude []string `yaml:"exclude" json:",omitempty"`
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
	if err := validateExternalSkills(m.ExternalSkills); err != nil {
		return nil, fmt.Errorf("manifest: %s: %w", path, err)
	}
	return &m, nil
}

func validateExternalSkills(skills []ExternalSkill) error {
	seen := make(map[string]bool, len(skills))
	for i, s := range skills {
		label := fmt.Sprintf("externalSkills[%d]", i)
		if err := s.Validate(); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
		if s.IsCollection() {
			// A collection's skill names are known only after a fetch; the
			// lock check reports clashes between them.
			continue
		}
		name := s.EffectiveName()
		if seen[name] {
			return fmt.Errorf("%s: duplicate skill name %q", label, name)
		}
		seen[name] = true
	}
	return nil
}

// Validate checks one entry on its own: required fields, a safe path and
// name, a source and ref that are safe to hand to git, and no credentials in
// the source URL. Load calls it for every entry; code that acts on an entry
// (such as a fetch) calls it again rather than trusting its caller.
func (s ExternalSkill) Validate() error {
	for field, value := range map[string]string{"source": s.Source, "ref": s.Ref, "path": s.Path} {
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
