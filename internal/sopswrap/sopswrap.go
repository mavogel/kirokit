// Package sopswrap implements the verified `sops exec-env` invocation
// shape: exactly two positional arguments (the file to decrypt, then the
// entire real command as one string that sops hands to /bin/sh -c). There
// is no "--" separator — sops counts it as a third positional argument and
// fails with "missing file to decrypt".
package sopswrap

import "strings"

// BuildWrapper returns the command/args pair that launches realCommand
// (with realArgs) through `sops exec-env`, decrypting secretsFilePath at
// process-launch time rather than at template-render time.
func BuildWrapper(secretsFilePath, realCommand string, realArgs []string) (command string, args []string) {
	return "sops", []string{"exec-env", secretsFilePath, shellQuote(realCommand, realArgs)}
}

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
