// Command kirokit renders, tests, and installs multi-team Kiro CLI v3
// configs from templated sources.
package main

import (
	"fmt"
	"os"

	"github.com/mavogel/kiro-cli-kit/cmd/kirokit/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
