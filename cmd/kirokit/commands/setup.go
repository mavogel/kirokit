package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mavogel/kirokit/internal/checkout"
	"github.com/mavogel/kirokit/internal/externalskills"
	"github.com/spf13/cobra"
)

// sourcePattern is the address set the desktop app accepts for its first
// clone: https://, ssh:// and file:// URLs, scp-style git@host:path and
// absolute paths, including Windows drive paths (C:\... or C:/...). Every
// alternative starts with a letter or "/", so git can never read an address
// as an option.
var sourcePattern = regexp.MustCompile(`^(?:(?:https|ssh|file)://\S+|git@\S+:\S+|/\S+|[A-Za-z]:[\\/]\S+)$`)

const noReleaseNotice = "This repository has no release yet (no v* tag). Ask your team lead to tag one."

func newSetupCmd() *cobra.Command {
	var source, target, branch string
	cmd := &cobra.Command{
		Use:   "setup --source <address> --target <folder>",
		Short: "Clone a kirokit config repository and check out its newest release",
		Long: `Clone the config repository at --source into --target, then check out its
newest v* tag (detached HEAD), the same first step the kirokit desktop app takes.
With --branch, that branch or tag is checked out instead. --target must not
exist or must be an empty folder. Without --branch, a repository with no v* tag
is kept on its default branch and reported as having no release yet.

--source is an https, ssh or file address, git@host:path, or an absolute path.
Afterwards point KIROKIT_REPO at --target (or run kirokit from it), then use
install or sync.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			source = strings.TrimSpace(source)
			if !sourcePattern.MatchString(source) {
				return errors.New("kirokit setup: --source: use an https, ssh or file address, git@host:path, or an absolute path")
			}
			branch = strings.TrimSpace(branch)
			if cmd.Flags().Changed("branch") {
				if err := checkout.CheckRefName(branch); err != nil {
					return fmt.Errorf("kirokit setup: --branch: %w", err)
				}
			}
			dir, err := filepath.Abs(strings.TrimSpace(target))
			if err != nil {
				return fmt.Errorf("kirokit setup: --target: %w", err)
			}

			args := []string{"clone"}
			if branch != "" {
				args = append(args, "--branch", branch)
			}
			clone := exec.Command("git", append(args, "--", source, dir)...)
			clone.Env = externalskills.WithoutRepoGitEnv()
			clone.Stdin, clone.Stdout, clone.Stderr = os.Stdin, out, cmd.ErrOrStderr()
			if err := clone.Run(); err != nil {
				return fmt.Errorf("kirokit setup: git clone: %w", err)
			}

			if branch != "" {
				_, _ = fmt.Fprintf(out, "checked out %s in %s\nnext: export KIROKIT_REPO=%s\n", branch, dir, dir)
				return nil
			}
			if err := os.Chdir(dir); err != nil {
				return fmt.Errorf("kirokit setup: %w", err)
			}
			switch err := updateCheckout(true, ""); {
			case err == nil:
				tag, _ := gitOutput("describe", "--tags", "--exact-match", "HEAD")
				_, _ = fmt.Fprintf(out, "checked out %s in %s\n", tag, dir)
			case strings.Contains(err.Error(), "no v* tag found"):
				_, _ = fmt.Fprintf(out, "cloned into %s\n%s\n", dir, noReleaseNotice)
			default:
				return fmt.Errorf("kirokit setup: %w", err)
			}
			_, _ = fmt.Fprintf(out, "next: export KIROKIT_REPO=%s\n", dir)
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "address of the config repository to clone (required)")
	cmd.Flags().StringVar(&target, "target", "", "folder to clone into; must not exist or be empty (required)")
	cmd.Flags().StringVar(&branch, "branch", "", "branch or tag to check out instead of the newest v* tag")
	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}
