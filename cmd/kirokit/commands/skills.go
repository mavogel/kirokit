package commands

import (
	"fmt"
	"path/filepath"

	"github.com/mavogel/kiro-cli-kit/internal/externalskills"
	"github.com/mavogel/kiro-cli-kit/internal/manifest"
	"github.com/spf13/cobra"
)

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage external skills vendored from git repositories",
	}
	cmd.AddCommand(newSkillsFetchCmd())
	return cmd
}

func newSkillsFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch [team]",
		Short: "Vendor a team's external skills from their git sources and update the lock file",
		Long: `Resolve each externalSkills entry of a team's manifest.yaml to a commit with
the machine's git, copy the skill directory into teams/<team>/external/skills/,
and rewrite teams/<team>/external/skills.lock.yaml. Private sources
authenticate through git itself (an SSH URL, or a git credential helper);
kirokit never reads or stores a credential.

Without a team, every team that declares external skills is fetched. Review
the vendored diff, then run kirokit update <team>.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := ""
			if len(args) == 1 {
				team = args[0]
			}
			teams, err := teamNames(team)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fetched := false
			for _, t := range teams {
				m, err := manifest.Load(filepath.Join(teamsRoot, t, "manifest.yaml"))
				if err != nil {
					return fmt.Errorf("kirokit skills fetch: %s: %w", t, err)
				}
				if len(m.ExternalSkills) == 0 {
					if team != "" {
						_, _ = fmt.Fprintf(out, "%s: no external skills declared in manifest.yaml\n", t)
					}
					continue
				}

				lock, err := externalskills.Fetch(filepath.Join(teamsRoot, t), m.ExternalSkills)
				if err != nil {
					return fmt.Errorf("kirokit skills fetch: %s: %w", t, err)
				}
				for _, s := range lock.Skills {
					_, _ = fmt.Fprintf(out, "%s: %s  %s @ %s -> %s\n", t, s.Name, s.Source, s.Ref, shortCommit(s.Commit))
				}
				_, _ = fmt.Fprintf(out, "%s: review the vendored files, then run `kirokit update %s`\n", t, t)
				fetched = true
			}
			if !fetched && team == "" {
				_, _ = fmt.Fprintln(out, "no team declares external skills")
			}
			return nil
		},
	}
}

func shortCommit(c string) string {
	if len(c) > 7 {
		return c[:7]
	}
	return c
}
