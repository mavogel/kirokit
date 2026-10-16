package sopswrap

import (
	"bytes"
	"os/exec"
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
