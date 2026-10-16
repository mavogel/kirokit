package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mavogel/kirokit/internal/install"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var repo string
	var latestTag bool
	var kbAgent, kbSource, kbMount string
	cmd := &cobra.Command{
		Use:   "sync <team>",
		Short: "Update a config checkout from git and reinstall a team, pruning files it no longer ships",
		Long: `Update the kirokit config checkout at --repo, then install <team> from it.

By default the checked-out branch is fast-forwarded from its upstream. With
--latest-tag the newest v* tag is checked out instead (detached HEAD), so only
released changes reach this machine. Files a previous install wrote that the
new version no longer ships are removed.

Run it from a scheduler (see scripts/kirokit-sync-setup.sh).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]
			out := cmd.OutOrStdout()

			kb, err := parseKnowledgeBaseFlags(kbAgent, kbSource, kbMount)
			if err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			kiroHome, err := install.ResolveKiroHome()
			if err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}

			// teams/, global/ and env/ resolve against the working directory.
			if err := os.Chdir(repo); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}

			before, _ := gitOutput("rev-parse", "--short", "HEAD")
			if err := updateCheckout(latestTag); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			after, _ := gitOutput("rev-parse", "--short", "HEAD")
			if before != after {
				_, _ = fmt.Fprintf(out, "updated %s -> %s\n", before, after)
			}

			roots := install.Roots{
				TeamsRoot:       teamsRoot,
				GlobalRoot:      globalRoot,
				EnvVersionsPath: envVersionsPath,
			}
			pruned, err := install.Sync(roots, team, kiroHome, kb)
			if err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			for _, p := range pruned {
				_, _ = fmt.Fprintf(out, "pruned %s\n", p)
			}
			_, _ = fmt.Fprintf(out, "synced %s into %s at %s\n", team, kiroHome, after)
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", ".", "path of the kirokit config checkout to update and install from")
	cmd.Flags().BoolVar(&latestTag, "latest-tag", false, "check out the newest v* tag instead of fast-forwarding the current branch")
	cmd.Flags().StringVar(&kbAgent, "kb-agent", "", "attach a local knowledge base to this installed agent (requires --kb-source and --kb-mount)")
	cmd.Flags().StringVar(&kbSource, "kb-source", "", "local folder to expose as a knowledgeBase resource")
	cmd.Flags().StringVar(&kbMount, "kb-mount", "", "path where --kb-source is symlinked for the agent to reference")
	return cmd
}

func updateCheckout(latestTag bool) error {
	if _, err := gitOutput("fetch", "--tags", "--prune", "--quiet"); err != nil {
		return err
	}
	if !latestTag {
		_, err := gitOutput("pull", "--ff-only", "--quiet")
		return err
	}
	tags, err := gitOutput("tag", "--list", "v*", "--sort=-v:refname")
	if err != nil {
		return err
	}
	tag, _, _ := strings.Cut(tags, "\n")
	if tag == "" {
		return fmt.Errorf("--latest-tag: no v* tag found after fetch")
	}
	_, err = gitOutput("checkout", "--quiet", tag)
	return err
}

func gitOutput(args ...string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, text)
	}
	return text, nil
}
