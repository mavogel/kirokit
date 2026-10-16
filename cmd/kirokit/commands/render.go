package commands

import (
	"fmt"
	"path/filepath"

	"github.com/mavogel/kirokit/internal/render"
	"github.com/spf13/cobra"
)

func newRenderCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "render [team]",
		Short: "Render team template sources into a Kiro CLI v3 config tree",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("kirokit render: --out is required")
			}
			team := ""
			if len(args) == 1 {
				team = args[0]
			}
			teams, err := teamNames(team)
			if err != nil {
				return err
			}
			for _, t := range teams {
				outDir := out
				if team == "" {
					outDir = filepath.Join(out, t)
				}
				if err := render.RenderTeam(teamsRoot, t, globalRoot, envVersionsPath, outDir, ""); err != nil {
					return fmt.Errorf("kirokit render: %s: %w", t, err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rendered %s -> %s\n", t, outDir)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output directory for the rendered config tree (required)")
	return cmd
}
