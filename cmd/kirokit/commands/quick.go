package commands

import (
	"fmt"
	"maps"

	"github.com/mavogel/kirokit/internal/quick"
	"github.com/spf13/cobra"
)

func newQuickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick",
		Short: "Work with Amazon Quick bundles (see kirokit render --quick)",
	}
	cmd.AddCommand(newQuickIDsCmd())
	return cmd
}

func newQuickIDsCmd() *cobra.Command {
	var idsPath string
	cmd := &cobra.Command{
		Use:   "ids <export>...",
		Short: "Record the skill ids Amazon Quick assigned, from skills exported from Quick",
		Long: `Read skills exported from Amazon Quick and record each skill's id in the ids
file, so the next kirokit render --quick writes the id back and a reimport
overwrites the same skill. Each argument is an exported SKILL.md (or other .md),
a .zip export, or a folder searched for SKILL.md files and .zip files.

The ids file is quick-ids.yaml at the config repository root, or
KIROKIT_QUICK_IDS, or --quick-ids. Ids belong to one Quick account: the file
is gitignored and must never be committed. Names already in it are updated,
others kept.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := quickIDsPath(idsPath)
			ids, err := quick.LoadIDs(path)
			if err != nil {
				return fmt.Errorf("kirokit quick ids: %w", err)
			}
			found, skipped, err := quick.ReadExports(args)
			if err != nil {
				return fmt.Errorf("kirokit quick ids: %w", err)
			}
			out := cmd.OutOrStdout()
			for _, s := range skipped {
				_, _ = fmt.Fprintf(out, "skipped %s: no id\n", s)
			}
			maps.Copy(ids, found)
			if err := quick.WriteIDs(path, ids); err != nil {
				return fmt.Errorf("kirokit quick ids: %w", err)
			}
			_, _ = fmt.Fprintf(out, "recorded %d ids in %s\n", len(found), path)
			return nil
		},
	}
	cmd.Flags().StringVar(&idsPath, "quick-ids", "", "Quick skill ids file (default <repo>/quick-ids.yaml, or KIROKIT_QUICK_IDS)")
	return cmd
}
