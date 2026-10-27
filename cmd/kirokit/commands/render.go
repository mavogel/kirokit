package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mavogel/kirokit/internal/quick"
	"github.com/mavogel/kirokit/internal/render"
	"github.com/spf13/cobra"
)

func newRenderCmd() *cobra.Command {
	var (
		out, quickIDs string
		toQuick       bool
	)
	cmd := &cobra.Command{
		Use:   "render [team]",
		Short: "Render team template sources into a Kiro CLI v3 config tree, or an Amazon Quick bundle with --quick",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("kirokit render: --out is required")
			}
			if quickIDs != "" && !toQuick {
				return fmt.Errorf("kirokit render: --quick-ids needs --quick")
			}
			team := ""
			if len(args) == 1 {
				team = args[0]
			}
			teams, err := teamNames(team)
			if err != nil {
				return err
			}
			var ids map[string]string
			if toQuick {
				if ids, err = quick.LoadIDs(quickIDsPath(quickIDs)); err != nil {
					return fmt.Errorf("kirokit render: %w", err)
				}
			}
			for _, t := range teams {
				outDir := out
				if team == "" {
					outDir = filepath.Join(out, t)
				}
				if toQuick {
					if err := renderQuick(cmd, t, outDir, ids, team == ""); err != nil {
						return err
					}
					continue
				}
				if err := render.RenderTeam(repoPath(teamsRoot), t, repoPath(globalRoot), repoPath(envVersionsPath), outDir, ""); err != nil {
					return fmt.Errorf("kirokit render: %s: %w", t, err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rendered %s -> %s\n", t, outDir)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output directory for the rendered config tree (required)")
	cmd.Flags().BoolVar(&toQuick, "quick", false, "render the team's Amazon Quick bundle instead of the Kiro tree")
	cmd.Flags().StringVar(&quickIDs, "quick-ids", "", "Quick skill ids file (default <repo>/quick-ids.yaml, or KIROKIT_QUICK_IDS)")
	return cmd
}

// renderQuick renders one team's Quick bundle. When every team is rendered,
// a team that lists nothing for Quick is skipped with a note.
func renderQuick(cmd *cobra.Command, team, outDir string, ids map[string]string, allTeams bool) error {
	res, err := quick.RenderTeam(repoPath(teamsRoot), team, repoPath(globalRoot), repoPath(envVersionsPath), outDir, quick.Options{IDs: ids})
	if allTeams && errors.Is(err, quick.ErrNoQuick) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: lists nothing for Quick, skipped\n", team)
		return nil
	}
	if err != nil {
		return fmt.Errorf("kirokit render: %s: %w", team, err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rendered Quick bundle for %s: %d skills, %d agents -> %s\n", team, res.Skills, res.Agents, outDir)
	for _, name := range res.MissingIDs {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: no Quick id for skill %s yet (first import); run kirokit quick ids after importing\n", name)
	}
	return nil
}

// quickIDsPath is the ids file: the flag value, else KIROKIT_QUICK_IDS, else
// quick-ids.yaml at the config repository root.
func quickIDsPath(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("KIROKIT_QUICK_IDS"); env != "" {
		return env
	}
	return repoPath(quick.IDsFileName)
}
