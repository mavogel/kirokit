package render

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/mavogel/kirokit/internal/externalskills"
)

// externalSkillURIs returns the skill:// resource of every vendored external
// skill (the names externalskills.Resolve returned), rooted at kiroHome. They
// use the same URI form team-authored skills are listed with.
func externalSkillURIs(kiroHome string, names []string) []string {
	uris := make([]string, 0, len(names))
	for _, name := range names {
		uris = append(uris, "skill://"+path.Join(kiroHome, "skills", name, "SKILL.md"))
	}
	return uris
}

// writeExternalSkills copies each vendored skill into root as skills/<name>/,
// byte for byte. The files are never run through the template engine: a
// vendored file may be named *.tmpl or contain {{ }} and must arrive intact.
// A skill whose name is already taken by a team-authored skill is an error.
func writeExternalSkills(root, teamDir string, names []string) error {
	for _, name := range names {
		if !plainName(name) {
			return fmt.Errorf("render: invalid external skill name %q", name)
		}
		dest := filepath.Join(root, "skills", name)
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("render: external skill %q collides with the team's own skills/%s; rename one of them", name, name)
		}

		src := filepath.Join(externalskills.SkillsDir(teamDir), name)
		err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("%s is not a regular file", p)
			}
			rel, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out := filepath.Join(dest, rel)
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			perm := os.FileMode(0o600)
			if info.Mode().Perm()&0o100 != 0 {
				perm = 0o700
			}
			if err := os.WriteFile(out, data, perm); err != nil {
				return err
			}
			return os.Chmod(out, perm)
		})
		if err != nil {
			return fmt.Errorf("render: external skill %q: %w", name, err)
		}
	}
	return nil
}
