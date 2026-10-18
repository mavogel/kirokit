package commands

import (
	"fmt"

	"github.com/mavogel/kirokit/internal/golden"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [team]",
		Short: "Render teams and diff the result against committed golden fixtures",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := ""
			if len(args) == 1 {
				team = args[0]
			}
			teams, err := teamNames(team)
			if err != nil {
				return err
			}

			in := golden.Inputs{
				TeamsRoot:       repoPath(teamsRoot),
				GlobalRoot:      repoPath(globalRoot),
				EnvVersionsPath: repoPath(envVersionsPath),
				GoldenRoot:      repoPath(goldenRoot),
			}

			mismatched := false
			for _, t := range teams {
				diffs, err := golden.Compare(in, t)
				if err != nil {
					return fmt.Errorf("kirokit test: %s: %w", t, err)
				}
				if len(diffs) > 0 {
					mismatched = true
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: golden mismatch:\n", t)
					for _, d := range diffs {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  %s\n", d)
					}
					continue
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: ok\n", t)
			}
			if mismatched {
				return fmt.Errorf("kirokit test: one or more teams have golden fixture mismatches")
			}
			return nil
		},
	}
	return cmd
}
