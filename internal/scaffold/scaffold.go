// Package scaffold implements `kirokit onboard`: copying an embedded
// starter team layout into teams/<team>/ for a new team, substituting the
// team name into the scaffolded manifest.yaml.
package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// starterFS embeds the starter team layout, including its empty-category
// .gitkeep placeholders (the "all:" prefix is required for embed to
// include dotfiles it would otherwise skip).
//
//go:embed all:starter
var starterFS embed.FS

const starterRoot = "starter"

// Scaffold copies the embedded starter layout into
// destTeamsDir/teamName, substituting teamName into the scaffolded
// manifest.yaml. It refuses to overwrite an existing team directory.
func Scaffold(destTeamsDir, teamName string) error {
	teamDir := filepath.Join(destTeamsDir, teamName)

	if _, err := os.Stat(teamDir); err == nil {
		return fmt.Errorf("scaffold: team %q already exists at %s", teamName, teamDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("scaffold: %w", err)
	}

	if err := copyStarter(teamDir, teamName); err != nil {
		_ = os.RemoveAll(teamDir)
		return err
	}
	return nil
}

func copyStarter(teamDir, teamName string) error {
	return fs.WalkDir(starterFS, starterRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == starterRoot {
			return nil
		}

		rel := strings.TrimPrefix(path, starterRoot+"/")
		target := filepath.Join(teamDir, filepath.FromSlash(rel))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := starterFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("scaffold: read embedded %s: %w", path, err)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("scaffold: %w", err)
		}

		if rel == "manifest.yaml.tmpl" {
			rendered, err := renderManifest(data, teamName)
			if err != nil {
				return err
			}
			target = filepath.Join(teamDir, "manifest.yaml")
			return os.WriteFile(target, rendered, 0o600)
		}

		return os.WriteFile(target, data, 0o600)
	})
}

func renderManifest(data []byte, teamName string) ([]byte, error) {
	tmpl, err := template.New("manifest.yaml.tmpl").Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("scaffold: parse manifest template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, struct{ Team string }{Team: teamName}); err != nil {
		return nil, fmt.Errorf("scaffold: render manifest template: %w", err)
	}
	return []byte(buf.String()), nil
}
