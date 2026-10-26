// Package checkout holds the rule for branch and tag names that kirokit hands
// to git, and a read-only view of a config checkout's refs.
package checkout

import (
	"errors"
	"strings"
)

// errRefName is the refusal of a name CheckRefName rejects; callers prefix it
// with the command and flag, as in "kirokit sync: --branch: ...".
var errRefName = errors.New("must be a branch or tag name, not empty, starting with '-' or containing whitespace")

// CheckRefName refuses an empty name, one starting with '-' (git would read it
// as an option), or one with whitespace. Anything else is for git to find or
// not.
func CheckRefName(name string) error {
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, " \t\r\n") {
		return errRefName
	}
	return nil
}
