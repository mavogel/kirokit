package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Set at release time through GoReleaser ldflags (see .goreleaser.yaml).
// Unset for `go build` / `go install`, where buildInfo fills them in.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the kirokit version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, c, d := buildInfo()
			s := "kirokit " + v + "\n"
			if c != "" {
				s += "commit: " + c + "\n"
			}
			if d != "" {
				s += "built:  " + d + "\n"
			}
			_, err := fmt.Fprint(cmd.OutOrStdout(), s)
			return err
		},
	}
}

// buildInfo returns the ldflags values, falling back to the module version and
// VCS stamp the Go toolchain embeds when they were not injected.
func buildInfo() (v, c, d string) {
	v, c, d = version, commit, date
	if v != "dev" {
		return v, c, d
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v, c, d
	}
	if mv := info.Main.Version; mv != "" && mv != "(devel)" {
		v = mv
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if c == "" {
				c = s.Value
			}
		case "vcs.time":
			if d == "" {
				d = s.Value
			}
		}
	}
	return v, c, d
}
