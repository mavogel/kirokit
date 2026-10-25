package sopswrap

import (
	"bytes"
	"errors"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestBuildWrapper_SimpleCommandNoSpecialCharacters(t *testing.T) {
	command, args := BuildWrapper("agents/devops/env/secrets.enc.yaml", "jira-mcp-server", []string{"--port", "8080"})

	if command != "sops" {
		t.Errorf("command = %q, want %q", command, "sops")
	}
	want := []string{"exec-env", "agents/devops/env/secrets.enc.yaml", "jira-mcp-server --port 8080"}
	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildWrapper_NeverUsesDoubleDashSeparator(t *testing.T) {
	_, args := BuildWrapper("path", "cmd", []string{"a"})
	for _, a := range args {
		if a == "--" {
			t.Fatalf("args contains a bare \"--\" separator, which sops exec-env counts as a third positional arg: %v", args)
		}
	}
	if len(args) != 3 {
		t.Fatalf("args has %d elements, want exactly 3 (exec-env, file, command)", len(args))
	}
}

func TestShellQuote_RoundTripsThroughRealShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sops runs the command with cmd.exe on Windows; see the Cmd round trip")
	}
	command, args := BuildWrapper("secrets.yaml", "printf", []string{"%s", "it's a test with spaces"})
	quotedCommandString := args[2]

	out, err := exec.Command("/bin/sh", "-c", quotedCommandString).Output()
	if err != nil {
		t.Fatalf("running the quoted command through /bin/sh -c failed: %v", err)
	}
	if got := string(bytes.TrimRight(out, "\n")); got != "it's a test with spaces" {
		t.Errorf("shell round-trip produced %q, want %q", got, "it's a test with spaces")
	}
	_ = command
}

func TestBuildWrapperFor_POSIXIsBuildWrapper(t *testing.T) {
	realArgs := []string{"-y", "mcp-remote@0.15.1", "--header", "Authorization: ${ATLASSIAN_AUTH}", ""}
	wantCommand, wantArgs := BuildWrapper("/kiro-home/s.enc.yaml", "npx", realArgs)
	command, args, err := BuildWrapperFor(POSIX, "/kiro-home/s.enc.yaml", "npx", realArgs)
	if err != nil {
		t.Fatalf("POSIX: %v", err)
	}
	if command != wantCommand || !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("POSIX = %q %q, want %q %q", command, args, wantCommand, wantArgs)
	}
}

func TestBuildWrapperFor_CmdJoinsPlainTokens(t *testing.T) {
	command, args, err := BuildWrapperFor(Cmd, "C:/Users/a/.kiro/kirokit-secrets/t/a/secrets.enc.yaml", "npx",
		[]string{"-y", "mcp-remote@0.15.1", "https://h/v2/mcp?x=1", "--header", "Authorization:${X}", `C:\tools\x`, "a=b", "user@host"})
	if err != nil {
		t.Fatalf("Cmd: %v", err)
	}
	want := []string{"exec-env", "C:/Users/a/.kiro/kirokit-secrets/t/a/secrets.enc.yaml",
		`npx -y mcp-remote@0.15.1 https://h/v2/mcp?x=1 --header Authorization:${X} C:\tools\x a=b user@host`}
	if command != "sops" || !reflect.DeepEqual(args, want) {
		t.Errorf("Cmd = %q %q, want %q %q", command, args, "sops", want)
	}
}

func TestBuildWrapperFor_CmdRefusesWhatCmdExeCannotPass(t *testing.T) {
	bad := map[string]string{
		"space":        "Authorization: x",
		"tab":          "a\tb",
		"newline":      "a\nb",
		"double quote": `say"hi`,
		"ampersand":    "a&b",
		"pipe":         "a|b",
		"less":         "a<b",
		"greater":      "a>b",
		"caret":        "a^b",
		"percent":      "%PATH%",
		"bang":         "a!b",
		"open paren":   "a(b",
		"close paren":  "a)b",
		"empty":        "",
	}
	for name, arg := range bad {
		t.Run(name, func(t *testing.T) {
			_, _, err := BuildWrapperFor(Cmd, "s.enc.yaml", "server", []string{"--ok", arg})
			if !errors.Is(err, ErrUnsafeForCmd) {
				t.Fatalf("err = %v, want ErrUnsafeForCmd", err)
			}
			var unsafe *UnsafeArgError
			if !errors.As(err, &unsafe) || unsafe.Arg != arg {
				t.Fatalf("err = %#v, want an UnsafeArgError for %q", err, arg)
			}
			if !strings.Contains(err.Error(), strconv.Quote(arg)) {
				t.Errorf("message %q does not quote the argument", err.Error())
			}
		})
	}
	if _, _, err := BuildWrapperFor(Cmd, "s.enc.yaml", "my server", nil); !errors.Is(err, ErrUnsafeForCmd) {
		t.Errorf("a command with a space: err = %v, want ErrUnsafeForCmd", err)
	}
}
