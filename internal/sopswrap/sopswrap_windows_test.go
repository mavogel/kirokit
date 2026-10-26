//go:build windows

package sopswrap

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// TestHelperArgv is not a test: run by the round trip below, it prints the
// arguments after "--" as JSON.
func TestHelperArgv(t *testing.T) {
	if os.Getenv("KIROKIT_SOPSWRAP_HELPER") != "1" {
		t.Skip("helper process")
	}
	for i, a := range os.Args {
		if a == "--" {
			b, _ := json.Marshal(os.Args[i+1:]) //nolint:errchkjson // a []string always marshals
			fmt.Print(string(b))
			os.Exit(0)
		}
	}
	os.Exit(2)
}

// The string BuildWrapperFor(Cmd, …) returns is run exactly as sops runs it on
// Windows, exec.Command("cmd.exe", "/C", s), and must reach the program as the
// original arguments.
func TestBuildWrapperFor_CmdRoundTripsThroughCmdExe(t *testing.T) {
	realArgs := []string{"-test.run=TestHelperArgv", "--",
		"-y", "mcp-remote@0.15.1", "https://mcp.example.com/v2/mcp?team=x", "--header", "Authorization:${ATLASSIAN_AUTH}",
		"user@example.com", "a=b", `C:\tools\x`, "C:/kiro/s.enc.yaml", "--port", "8080"}
	_, args, err := BuildWrapperFor(Cmd, "s.enc.yaml", os.Args[0], realArgs)
	if err != nil {
		t.Fatalf("BuildWrapperFor: %v (test binary at %s)", err, os.Args[0])
	}
	cmd := exec.Command("cmd.exe", "/C", args[2])
	cmd.Env = append(os.Environ(), "KIROKIT_SOPSWRAP_HELPER=1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("cmd.exe /C %s: %v", args[2], err)
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("helper output %q: %v", out, err)
	}
	if want := realArgs[2:]; !reflect.DeepEqual(got, want) {
		t.Errorf("argv through cmd.exe = %q, want %q", got, want)
	}
}
