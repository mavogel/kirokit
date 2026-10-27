// Package golden implements the kirokit golden-file test harness: it
// renders a team and diffs the result — both the output tree and a JSON
// snapshot of the resolved template context — against committed fixtures
// under testdata/golden/<team>/.
package golden

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mavogel/kirokit/internal/quick"
	"github.com/mavogel/kirokit/internal/render"
)

// Inputs are the shared roots every render/golden operation needs.
type Inputs struct {
	TeamsRoot       string
	GlobalRoot      string
	EnvVersionsPath string
	GoldenRoot      string
}

// Compare renders team and diffs it against testdata/golden/<team>/. An
// empty result means the team's committed goldens are up to date.
func Compare(in Inputs, team string) ([]string, error) {
	r, err := renderForGolden(in, team)
	if err != nil {
		return nil, err
	}
	defer r.cleanup()
	renderedDir, contextJSON := r.renderedDir, r.contextJSON

	goldenTeamDir := filepath.Join(in.GoldenRoot, team)
	goldenRenderedDir := filepath.Join(goldenTeamDir, "rendered")
	goldenContextPath := filepath.Join(goldenTeamDir, "context.json")

	diffs, err := diffDirs(renderedDir, goldenRenderedDir)
	if err != nil {
		return nil, err
	}
	// The Quick bundle is compared the same way; a team without a quick
	// section is compared with an empty folder, so a stray quick/ golden fails.
	quickDiffs, err := diffDirs(r.quickDir, filepath.Join(goldenTeamDir, quickGoldenDir))
	if err != nil {
		return nil, err
	}
	for _, d := range quickDiffs {
		diffs = append(diffs, prefixPath(d, quickGoldenDir+"/"))
	}

	goldenContext, err := os.ReadFile(goldenContextPath)
	switch {
	case os.IsNotExist(err):
		diffs = append(diffs, fmt.Sprintf("missing golden context file: %s", goldenContextPath))
	case err != nil:
		return nil, fmt.Errorf("golden: read %s: %w", goldenContextPath, err)
	case !bytes.Equal(bytes.TrimSpace(goldenContext), bytes.TrimSpace(contextJSON)):
		diffs = append(diffs, fmt.Sprintf("context.json differs from %s", goldenContextPath))
	}

	sort.Strings(diffs)
	return diffs, nil
}

// Update renders team and overwrites its committed golden fixtures.
func Update(in Inputs, team string) error {
	r, err := renderForGolden(in, team)
	if err != nil {
		return err
	}
	defer r.cleanup()
	renderedDir, contextJSON := r.renderedDir, r.contextJSON

	goldenTeamDir := filepath.Join(in.GoldenRoot, team)
	goldenRenderedDir := filepath.Join(goldenTeamDir, "rendered")
	goldenQuickDir := filepath.Join(goldenTeamDir, quickGoldenDir)
	goldenContextPath := filepath.Join(goldenTeamDir, "context.json")

	if err := os.RemoveAll(goldenRenderedDir); err != nil {
		return fmt.Errorf("golden: %w", err)
	}
	if err := copyDir(renderedDir, goldenRenderedDir); err != nil {
		return fmt.Errorf("golden: %w", err)
	}
	if err := os.RemoveAll(goldenQuickDir); err != nil {
		return fmt.Errorf("golden: %w", err)
	}
	if r.hasQuick {
		if err := copyDir(r.quickDir, goldenQuickDir); err != nil {
			return fmt.Errorf("golden: %w", err)
		}
	}
	if err := os.MkdirAll(goldenTeamDir, 0o755); err != nil {
		return fmt.Errorf("golden: %w", err)
	}
	if err := os.WriteFile(goldenContextPath, contextJSON, 0o600); err != nil {
		return fmt.Errorf("golden: write %s: %w", goldenContextPath, err)
	}
	return nil
}

// quickGoldenDir holds a team's Quick bundle golden beside rendered/.
const quickGoldenDir = "quick"

// QuickIDsFixture is the committed ids file Quick goldens render with, under
// the golden root; a real account's ids are never read by the harness.
const QuickIDsFixture = "quick-ids.yaml"

// goldenRender is one team rendered for comparison. quickDir is an empty
// folder when the team has no quick section (hasQuick false).
type goldenRender struct {
	renderedDir, quickDir string
	hasQuick              bool
	contextJSON           []byte
	cleanup               func()
}

// renderForGolden renders team into a scratch directory (using the fixed
// KiroHome sentinel, per render.DefaultKiroHomeSentinel), renders its Quick
// bundle with the fixture ids, and marshals its resolved context. The caller
// must call cleanup once done with the result.
func renderForGolden(in Inputs, team string) (goldenRender, error) {
	ctx, err := render.BuildContext(in.TeamsRoot, team, in.GlobalRoot, in.EnvVersionsPath, "")
	if err != nil {
		return goldenRender{}, err
	}
	contextJSON, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return goldenRender{}, fmt.Errorf("golden: marshal context: %w", err)
	}
	contextJSON = append(contextJSON, '\n')

	scratch, err := os.MkdirTemp("", "kirokit-golden-*")
	if err != nil {
		return goldenRender{}, fmt.Errorf("golden: %w", err)
	}
	r := goldenRender{
		renderedDir: filepath.Join(scratch, "rendered"),
		quickDir:    filepath.Join(scratch, quickGoldenDir),
		contextJSON: contextJSON,
		cleanup:     func() { _ = os.RemoveAll(scratch) },
	}
	if err := render.RenderTeam(in.TeamsRoot, team, in.GlobalRoot, in.EnvVersionsPath, r.renderedDir, ""); err != nil {
		r.cleanup()
		return goldenRender{}, err
	}

	ids, err := quick.LoadIDs(filepath.Join(in.GoldenRoot, QuickIDsFixture))
	if err != nil {
		r.cleanup()
		return goldenRender{}, fmt.Errorf("golden: %w", err)
	}
	_, err = quick.RenderTeam(in.TeamsRoot, team, in.GlobalRoot, in.EnvVersionsPath, r.quickDir, quick.Options{IDs: ids})
	switch {
	case errors.Is(err, quick.ErrNoQuick):
		err = os.MkdirAll(r.quickDir, 0o755)
	case err == nil:
		r.hasQuick = true
	}
	if err != nil {
		r.cleanup()
		return goldenRender{}, err
	}
	return r, nil
}

// prefixPath puts prefix before the file path that ends a diffDirs line.
func prefixPath(diff, prefix string) string {
	i := strings.LastIndex(diff, ": ")
	if i < 0 {
		return diff
	}
	return diff[:i+2] + prefix + diff[i+2:]
}

// diffDirs compares two rendered output trees file-by-file, byte-for-byte.
// A file present in only one side, or differing content, is a diff.
func diffDirs(actualDir, goldenDir string) ([]string, error) {
	actualFiles, err := listFiles(actualDir)
	if err != nil {
		return nil, err
	}
	goldenFiles, err := listFiles(goldenDir)
	if err != nil {
		if os.IsNotExist(err) {
			goldenFiles = nil
		} else {
			return nil, err
		}
	}

	goldenSet := toSet(goldenFiles)
	actualSet := toSet(actualFiles)

	var diffs []string
	for _, f := range actualFiles {
		if !goldenSet[f] {
			diffs = append(diffs, fmt.Sprintf("extra file (not in golden): %s", f))
			continue
		}
		a, err := os.ReadFile(filepath.Join(actualDir, f))
		if err != nil {
			return nil, fmt.Errorf("golden: read %s: %w", f, err)
		}
		g, err := os.ReadFile(filepath.Join(goldenDir, f))
		if err != nil {
			return nil, fmt.Errorf("golden: read %s: %w", f, err)
		}
		if !bytes.Equal(a, g) {
			diffs = append(diffs, fmt.Sprintf("content differs: %s", f))
		}
	}
	for _, f := range goldenFiles {
		if !actualSet[f] {
			diffs = append(diffs, fmt.Sprintf("missing file (present in golden, not rendered): %s", f))
		}
	}
	return diffs, nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func toSet(files []string) map[string]bool {
	m := make(map[string]bool, len(files))
	for _, f := range files {
		m[f] = true
	}
	return m
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
}
