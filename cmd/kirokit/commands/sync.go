package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mavogel/kirokit/internal/checkout"
	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/mavogel/kirokit/internal/install"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var repo, ref string
	var latestTag bool
	var kbAgent, kbSource, kbMount string
	cmd := &cobra.Command{
		Use:   "sync [team]",
		Short: "Update a config checkout from git and reinstall a team, pruning files it no longer ships",
		Long: `Update the kirokit config checkout at --repo, then install <team> from it.
Once the checkout is updated and <team> renders, any other installed team is
uninstalled, so a Kiro home holds one team. A failed update or an unknown team
leaves the installed team untouched.

By default the checked-out branch is fast-forwarded from its upstream, and a
checked-out tag (detached HEAD) stays as it is. With --latest-tag the newest
v* tag is checked out instead, so only released changes reach this machine.
With --branch <name> the checkout switches to origin's branch of that name
(fast-forwarded) or, when there is none, to the tag (detached); local changes
to tracked files, local commits on that branch or an unknown name refuse the
switch before anything changes. Files a previous install wrote that the new
version no longer ships are removed.

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

			if cmd.Flags().Changed("branch") {
				if latestTag {
					return errors.New("kirokit sync: --branch and --latest-tag cannot be used together")
				}
				ref = strings.TrimSpace(ref)
				if err := checkout.CheckRefName(ref); err != nil {
					return fmt.Errorf("kirokit sync: --branch: %w", err)
				}
			}

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
			if err := updateCheckout(latestTag, ref); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			after, _ := gitOutput("rev-parse", "--short", "HEAD")
			if before != after {
				_, _ = fmt.Fprintf(out, "updated %s -> %s\n", before, after)
			}
			_, _ = fmt.Fprintf(out, "checked out %s at %s\n", checkedOut(after), after)
			if team == "" {
				return nil
			}

			roots := install.Roots{
				TeamsRoot:       teamsRoot,
				GlobalRoot:      globalRoot,
				EnvVersionsPath: envVersionsPath,
			}
			// Update and validate before removing anything: a failed pull or an
			// unknown team must leave the installed team in place.
			if err := install.Check(roots, team, kiroHome); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
			}
			if err := uninstallOthers(out, kiroHome, team); err != nil {
				return fmt.Errorf("kirokit sync: %w", err)
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
	cmd.Flags().StringVar(&ref, "branch", "", "switch the checkout to this branch or tag of origin first; later plain syncs keep it (a branch wins over a tag of the same name)")
	cmd.Flags().StringVar(&kbAgent, "kb-agent", "", "attach a local knowledge base to this installed agent (requires --kb-source and --kb-mount)")
	cmd.Flags().StringVar(&kbSource, "kb-source", "", "local folder to expose as a knowledgeBase resource")
	cmd.Flags().StringVar(&kbMount, "kb-mount", "", "path where --kb-source is symlinked for the agent to reference")
	return cmd
}

// updateCheckout fetches, then switches to ref when one is named, checks out
// the newest v* tag with latestTag, or else follows: it fast-forwards the
// checked-out branch, and on a detached HEAD (a chosen tag) stays where it is.
func updateCheckout(latestTag bool, ref string) error {
	if _, err := gitOutput("fetch", "--tags", "--prune", "--quiet"); err != nil {
		return err
	}
	if ref != "" {
		return switchTo(ref)
	}
	if !latestTag {
		if !gitOK("symbolic-ref", "-q", "HEAD") {
			return nil
		}
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

// switchTo checks out origin's branch name (fast-forwarded) or, when origin
// has no such branch, the tag name (detached). It refuses before git changes
// anything when tracked files have local changes or the local branch has
// commits origin does not. name has passed checkout.CheckRefName, and reaches
// git only as a full ref or before "--", so it is never read as an option or
// a path.
func switchTo(name string) error {
	changes, err := gitOutput("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return err
	}
	if changes != "" {
		dir, _ := os.Getwd()
		return fmt.Errorf("local changes in %s: commit or discard them, then try again", dir)
	}

	remote, local := "refs/remotes/origin/"+name, "refs/heads/"+name
	switch {
	case gitOK("show-ref", "--verify", "-q", remote) && !gitOK("show-ref", "--verify", "-q", local):
		_, err = gitOutput("checkout", "--quiet", "--track", "-b", name, remote)
	case gitOK("show-ref", "--verify", "-q", remote):
		if !gitOK("merge-base", "--is-ancestor", local, remote) {
			return fmt.Errorf("--branch %s: local branch %s has commits that are not on origin/%s", name, name, name)
		}
		if _, err = gitOutput("checkout", "--quiet", name, "--"); err == nil {
			_, err = gitOutput("merge", "--ff-only", "--quiet", remote)
		}
	case gitOK("show-ref", "--verify", "-q", "refs/tags/"+name):
		_, err = gitOutput("checkout", "--quiet", "--detach", "refs/tags/"+name)
	default:
		return fmt.Errorf("--branch %s: no branch or tag of that name on origin", name)
	}
	return err
}

// checkedOut names what HEAD is on: the branch, else a tag at HEAD, else the
// commit itself. The branch is cut from its full ref, since git's short form
// turns into "heads/<name>" when a tag has the same name.
func checkedOut(commit string) string {
	if full, err := gitOutput("symbolic-ref", "-q", "HEAD"); err == nil {
		return strings.TrimPrefix(full, "refs/heads/")
	}
	if tag, err := gitOutput("describe", "--tags", "--exact-match", "HEAD"); err == nil {
		return tag
	}
	return commit
}

// gitOK reports whether git exits 0, for the commands whose answer is their
// exit code.
func gitOK(args ...string) bool {
	_, err := gitOutput(args...)
	return err == nil
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = externalskills.WithoutRepoGitEnv()
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, text)
	}
	return text, nil
}
