package commands

import (
	"fmt"
	"io"

	"github.com/mavogel/kirokit/internal/install"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var kbAgent, kbSource, kbMount string
	cmd := &cobra.Command{
		Use:   "install <team>",
		Short: "Render a team and copy the result into the local Kiro CLI config location",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]

			kb, err := parseKnowledgeBaseFlags(kbAgent, kbSource, kbMount)
			if err != nil {
				return fmt.Errorf("kirokit install: %w", err)
			}

			kiroHome, err := install.ResolveKiroHome()
			if err != nil {
				return fmt.Errorf("kirokit install: %w", err)
			}

			roots := install.Roots{
				TeamsRoot:       repoPath(teamsRoot),
				GlobalRoot:      repoPath(globalRoot),
				EnvVersionsPath: repoPath(envVersionsPath),
			}
			// A team that cannot render must not cost the machine its installed team.
			if err := install.Check(roots, team, kiroHome); err != nil {
				return fmt.Errorf("kirokit install: %w", err)
			}

			if err := uninstallOthers(cmd.OutOrStdout(), kiroHome, team); err != nil {
				return fmt.Errorf("kirokit install: %w", err)
			}
			if err := install.Install(roots, team, kiroHome, kb); err != nil {
				return fmt.Errorf("kirokit install: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "installed %s into %s\n", team, kiroHome)
			return nil
		},
	}
	cmd.Flags().StringVar(&kbAgent, "kb-agent", "", "attach a local knowledge base to this installed agent (requires --kb-source and --kb-mount)")
	cmd.Flags().StringVar(&kbSource, "kb-source", "", "local folder to expose as a knowledgeBase resource")
	cmd.Flags().StringVar(&kbMount, "kb-mount", "", "path where --kb-source is symlinked for the agent to reference")
	return cmd
}

// uninstallOthers removes every installed team except keep, printing one
// line per removal, so a Kiro home holds one team.
func uninstallOthers(out io.Writer, kiroHome, keep string) error {
	removed, err := install.UninstallOthers(kiroHome, keep)
	for _, other := range removed {
		_, _ = fmt.Fprintf(out, "uninstalled %s from %s\n", other, kiroHome)
	}
	return err
}

// parseKnowledgeBaseFlags validates that --kb-agent/--kb-source/--kb-mount
// are either all empty (no knowledge base attachment) or all set together.
func parseKnowledgeBaseFlags(agent, source, mount string) (*install.KnowledgeBase, error) {
	set := 0
	for _, v := range []string{agent, source, mount} {
		if v != "" {
			set++
		}
	}
	switch set {
	case 0:
		return nil, nil
	case 3:
		return &install.KnowledgeBase{Agent: agent, Source: source, Mount: mount}, nil
	default:
		return nil, fmt.Errorf("--kb-agent, --kb-source, and --kb-mount must be given together")
	}
}
