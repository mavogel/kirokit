package commands

import (
	"fmt"

	"github.com/mavogel/kiro-cli-kit/internal/install"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall <team>",
		Short: "Remove exactly what a prior install of a team wrote from the local Kiro CLI config location",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]

			kiroHome, err := install.ResolveKiroHome()
			if err != nil {
				return fmt.Errorf("kirokit uninstall: %w", err)
			}

			if err := install.Uninstall(kiroHome, team); err != nil {
				return fmt.Errorf("kirokit uninstall: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "uninstalled %s from %s\n", team, kiroHome)
			return nil
		},
	}
	return cmd
}
