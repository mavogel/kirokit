// Package sopswrap implements the verified `sops exec-env` invocation
// shape: exactly two positional arguments (the file to decrypt, then the
// entire real command as one string that sops hands to a shell: /bin/sh -c
// on macOS and Linux, cmd.exe /C on Windows). There is no "--" separator —
// sops counts it as a third positional argument and fails with "missing file
// to decrypt".
package sopswrap

import (
	"errors"
	"strconv"
	"strings"
)

// Shell is the shell sops exec-env hands the command string to.
type Shell int

const (
	// POSIX is /bin/sh -c, which sops uses on macOS and Linux.
	POSIX Shell = iota
	// Cmd is cmd.exe /C, which sops uses on Windows.
	Cmd
)

// ErrUnsafeForCmd is wrapped by every UnsafeArgError.
var ErrUnsafeForCmd = errors.New("cannot pass through cmd.exe")

// UnsafeArgError names the first token cmd.exe cannot pass unchanged.
type UnsafeArgError struct {
	Arg string
}

func (e *UnsafeArgError) Error() string {
	return "argument " + strconv.Quote(e.Arg) + " cannot pass through sops exec-env on Windows (cmd.exe): use a form without spaces or quotes, such as Authorization:${VAR}"
}

func (e *UnsafeArgError) Unwrap() error { return ErrUnsafeForCmd }

// BuildWrapper returns the command/args pair that launches realCommand
// (with realArgs) through `sops exec-env`, decrypting secretsFilePath at
// process-launch time rather than at template-render time. The command string
// is quoted for /bin/sh.
func BuildWrapper(secretsFilePath, realCommand string, realArgs []string) (command string, args []string) {
	return "sops", []string{"exec-env", secretsFilePath, shellQuote(realCommand, realArgs)}
}

// BuildWrapperFor is BuildWrapper for the shell sops will use. For Cmd the
// tokens are joined with single spaces and no quoting: sops passes the string
// through Go's Windows argument escaping, which turns every '"' into '\"', a
// form cmd.exe does not read, and batch files such as npx.cmd parse their
// arguments a second time. So a token with whitespace, '"' or a cmd.exe
// special character is refused with an *UnsafeArgError instead.
func BuildWrapperFor(sh Shell, secretsFilePath, realCommand string, realArgs []string) (command string, args []string, err error) {
	if sh != Cmd {
		command, args = BuildWrapper(secretsFilePath, realCommand, realArgs)
		return command, args, nil
	}
	tokens := append([]string{realCommand}, realArgs...)
	for _, tok := range tokens {
		if tok == "" || strings.ContainsAny(tok, cmdUnsafeChars) {
			return "", nil, &UnsafeArgError{Arg: tok}
		}
	}
	return "sops", []string{"exec-env", secretsFilePath, strings.Join(tokens, " ")}, nil
}

// cmdUnsafeChars cannot reach a program unchanged through cmd.exe /C and a
// batch-file shim.
const cmdUnsafeChars = " \t\r\n\v\f\"&|<>^%!()"

// shellQuote joins realCommand and realArgs into one POSIX-shell-safe
// string, since sops exec-env passes its command argument to `/bin/sh -c`.
func shellQuote(realCommand string, realArgs []string) string {
	parts := make([]string, 0, len(realArgs)+1)
	parts = append(parts, realCommand)
	parts = append(parts, realArgs...)

	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = quoteIfNeeded(p)
	}
	return strings.Join(quoted, " ")
}

// shellSpecialChars are characters that require single-quote protection in
// a POSIX shell command line.
const shellSpecialChars = " \t\n'\"\\$`!*?[]{}();&|<>~#"

func quoteIfNeeded(s string) string {
	if s != "" && !strings.ContainsAny(s, shellSpecialChars) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
