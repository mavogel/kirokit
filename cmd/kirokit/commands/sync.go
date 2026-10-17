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
		Use:   "sync [team]",
		Short: "Update a config checkout from git and reinstall a team, pruning files it no longer ships",
		Long: `Update the kirokit config checkout at --repo, then install <team> from it.
Any other installed team is uninstalled first, so a Kiro home holds one team.

By default the checked-out branch is fast-forwarded from its upstream. With
--latest-tag the newest v* tag is checked out instead (detached HEAD), so only
released changes reach this machine. Files a previous install wrote that the
new version no longer ships are removed.

Without <team> it only updates the checkout and installs nothing; the --kb-*
flags need a team.

Run it from a scheduler (see scripts/kirokit-sync-setup.sh).`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := ""
			if len(args) == 1 {
				team = args[0]
			}
			out := cmd.OutOrStdout()

			var kb *install.KnowledgeBase
			var kiroHome string
			if team == "" {
				for _, f := range []string{"kb-agent", "kb-source", "kb-mount"} {
					if cmd.Flags().Changed(f) {
						return fmt.Errorf("kirokit sync: --%s needs a team", f)
					}
				}
			} else {
				var err error
				if kb, err = parseKnowledgeBaseFlags(kbAgent, kbSource, kbMount); err != nil {
					return fmt.Errorf("kirokit sync: %w", err)
				}
				if kiroHome, err = install.ResolveKiroHome(); err != nil {
					return fmt.Errorf("kirokit sync: %w", err)
				}
				if err := uninstallOthers(out, kiroHome, team); err != nil {
					return fmt.Errorf("kirokit sync: %w", err)
				}
			}

			// An explicit --repo wins, else KIROKIT_REPO, else the working
			// directory. teams/, global/ and env/ then resolve against it.
			dir := repo
			if dir == "" {
				dir = repoBase
			}
			if dir == "" {
				dir = "."
			}
			if err := os.Chdir(dir); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			repoBase = ""

			before, _ := gitOutput("rev-parse", "--short", "HEAD")
			if err := updateCheckout(latestTag); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			after, _ := gitOutput("rev-parse", "--short", "HEAD")
			if before != after {
				_, _ = fmt.Fprintf(out, "updated %s -> %s\n", before, after)
			}

			if team == "" {
				ref, _ := gitOutput("describe", "--tags", "--exact-match", "HEAD")
				if branch, err := gitOutput("symbolic-ref", "--short", "-q", "HEAD"); err == nil {
					ref = branch
				}
				_, _ = fmt.Fprintf(out, "checked out %s at %s\n", ref, after)
				return nil
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
	cmd.Flags().StringVar(&repo, "repo", "", "path of the kirokit config checkout to update and install from (default $KIROKIT_REPO, else the working directory)")
	cmd.Flags().BoolVar(&latestTag, "latest-tag", false, "check out the newest v* tag instead of fast-forwarding the current branch (also without a team)")
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
	cmd := exec.Command("git", args...)
	cmd.Env = withoutRepoGitEnv()
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, text)
	}
	return text, nil
}

// withoutRepoGitEnv drops the variables that redirect git to another repository
// (GIT_DIR, GIT_WORK_TREE, GIT_INDEX_FILE, ...) so git operates on the working
// directory's own repository. Other GIT_* settings are kept.
func withoutRepoGitEnv() []string {
	drop := []string{"GIT_DIR=", "GIT_WORK_TREE=", "GIT_INDEX_FILE=", "GIT_OBJECT_DIRECTORY=", "GIT_ALTERNATE_OBJECT_DIRECTORIES=", "GIT_COMMON_DIR=", "GIT_NAMESPACE=", "GIT_PREFIX="}
	var env []string
next:
	for _, kv := range os.Environ() {
		for _, d := range drop {
			if strings.HasPrefix(kv, d) {
				continue next
			}
		}
		env = append(env, kv)
	}
	return env
}
