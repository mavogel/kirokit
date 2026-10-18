// Package commands implements the kirokit CLI's subcommands.
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Repository-relative roots. Every subcommand passes them through repoPath,
// which roots them at KIROKIT_REPO, or at the working directory when it is
// unset.
const (
	teamsRoot       = "teams"
	globalRoot      = "global/steerings"
	envVersionsPath = "env/mcp-versions.yaml"
	goldenRoot      = "testdata/golden"
)

// repoBase is the absolute KIROKIT_REPO, or "" when unset or empty.
var repoBase string

// repoPath joins repoBase and rel; with no base it returns rel unchanged.
func repoPath(rel string) string {
	if repoBase == "" {
		return rel
	}
	return filepath.Join(repoBase, rel)
}

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
		PersistentPreRunE: func(*cobra.Command, []string) error {
			repoBase = ""
			if env := os.Getenv("KIROKIT_REPO"); env != "" {
				abs, err := filepath.Abs(env)
				if err != nil {
					return fmt.Errorf("kirokit: KIROKIT_REPO: %w", err)
				}
				repoBase = abs
			}
			return nil
		},
	}
	root.AddCommand(newRenderCmd())
	root.AddCommand(newTestCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newOnboardCmd())
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newSetupCmd())
	root.AddCommand(newSyncCmd())
	root.AddCommand(newSkillsCmd())
	root.AddCommand(newVersionCmd())
	return root
}

// teamNames returns []string{team} when team is non-empty, or every
// directory under teamsRoot otherwise.
func teamNames(team string) ([]string, error) {
	if team != "" {
		return []string{team}, nil
	}
	entries, err := os.ReadDir(repoPath(teamsRoot))
	if err != nil {
		return nil, fmt.Errorf("kirokit: read %s: %w", repoPath(teamsRoot), err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
