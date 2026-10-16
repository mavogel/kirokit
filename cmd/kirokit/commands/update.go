package commands

import (
	"fmt"

	"github.com/mavogel/kiro-cli-kit/internal/golden"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [team]",
		Short: "Render teams and overwrite their committed golden fixtures",
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
				TeamsRoot:       teamsRoot,
				GlobalRoot:      globalRoot,
				EnvVersionsPath: envVersionsPath,
				GoldenRoot:      goldenRoot,
			}

			for _, t := range teams {
				if err := golden.Update(in, t); err != nil {
					return fmt.Errorf("kirokit update: %s: %w", t, err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "updated golden fixtures for %s\n", t)
			}
			return nil
		},
	}
	return cmd
}
