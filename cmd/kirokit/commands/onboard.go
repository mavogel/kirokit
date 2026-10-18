package commands

import (
	"fmt"
	"path/filepath"

	"github.com/mavogel/kirokit/internal/scaffold"
	"github.com/spf13/cobra"
)

func newOnboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onboard <team>",
		Short: "Scaffold a new team directory from the starter layout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]
			if err := scaffold.Scaffold(repoPath(teamsRoot), team); err != nil {
				return fmt.Errorf("kirokit onboard: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "onboarded %s at %s\n", team, filepath.Join(repoPath(teamsRoot), team))
			return nil
		},
	}
	return cmd
}
