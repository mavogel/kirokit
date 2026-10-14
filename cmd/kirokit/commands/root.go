// Package commands implements the kirokit CLI's subcommands.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Repository-relative roots every subcommand resolves against the current
// working directory. kirokit has no repo-root auto-discovery in v1, so it
// must always be run from the repository root.
const (
	teamsRoot       = "teams"
	globalRoot      = "global/steerings"
	envVersionsPath = "env/mcp-versions.yaml"
	goldenRoot      = "testdata/golden"
)

// Execute runs the kirokit root command.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "kirokit",
		Short:         "Render, test, and install multi-team Kiro CLI v3 configs",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newRenderCmd())
	root.AddCommand(newTestCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newOnboardCmd())
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newSyncCmd())
	return root
}

// teamNames returns []string{team} when team is non-empty, or every
// directory under teamsRoot otherwise.
func teamNames(team string) ([]string, error) {
	if team != "" {
		return []string{team}, nil
	}
	entries, err := os.ReadDir(teamsRoot)
	if err != nil {
		return nil, fmt.Errorf("kirokit: read %s: %w", teamsRoot, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
