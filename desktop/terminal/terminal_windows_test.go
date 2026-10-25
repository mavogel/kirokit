//go:build windows

package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHelperReport is not a test: started on a pseudo console by the test
// below, it prints its arguments, folder and one variable as JSON between
// markers, so the console's own escape sequences around it do not matter.
func TestHelperReport(t *testing.T) {
	if os.Getenv("KIROKIT_TERMINAL_HELPER") != "1" {
		t.Skip("helper process")
	}
	var args []string
	for i, a := range os.Args {
		if a == "--" {
			args = os.Args[i+1:]
		}
	}
	wd, _ := os.Getwd()
	b, _ := json.Marshal(map[string]any{"args": args, "dir": wd, "var": os.Getenv("KIROKIT_TERMINAL_VAR")}) //nolint:errchkjson // plain values
	fmt.Printf("<<%s>>\n", b)
	os.Exit(0)
}

func TestWindows_ArgumentsDirAndEnvReachTheProcess(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperReport", "--", "a b", `q"uote`, "")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "KIROKIT_TERMINAL_HELPER=1", "KIROKIT_TERMINAL_VAR=v 1")
	s, err := Start(cmd, 250, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	out := readAll(t, s)
	start, end := strings.Index(out, "<<"), strings.Index(out, ">>")
	if start < 0 || end < start {
		t.Fatalf("no report in output %q", out)
	}
	var got struct {
		Args []string `json:"args"`
		Dir  string   `json:"dir"`
		Var  string   `json:"var"`
	}
	if err := json.Unmarshal([]byte(out[start+2:end]), &got); err != nil {
		t.Fatalf("report %q: %v", out[start+2:end], err)
	}
	if want := []string{"a b", `q"uote`, ""}; strings.Join(got.Args, "|") != strings.Join(want, "|") || len(got.Args) != len(want) {
		t.Errorf("args = %q, want %q", got.Args, want)
	}
	if !strings.EqualFold(filepath.Clean(got.Dir), filepath.Clean(dir)) {
		t.Errorf("dir = %q, want %q", got.Dir, dir)
	}
	if got.Var != "v 1" {
		t.Errorf("KIROKIT_TERMINAL_VAR = %q, want %q", got.Var, "v 1")
	}
	if code := waitCode(t, s); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestWindows_OutputReachesTheReaderAndEndsOnEOF(t *testing.T) {
	s, err := Start(exec.Command("cmd.exe", "/C", "echo hi"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	if out := readAll(t, s); !strings.Contains(out, "hi") {
		t.Fatalf("output = %q, want it to contain hi", out)
	}
	if code := waitCode(t, s); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestWindows_ExitCodeIsReturned(t *testing.T) {
	s, err := Start(exec.Command("cmd.exe", "/C", "exit 3"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	readAll(t, s)
	if code := waitCode(t, s); code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
	if code := waitCode(t, s); code != 3 {
		t.Fatalf("second wait = %d, want 3", code)
	}
}

func TestWindows_ResizeChangesTheSizeTheProcessSees(t *testing.T) {
	script := `[void][Console]::ReadLine(); $s = $Host.UI.RawUI.WindowSize; Write-Output ('size {0} {1}' -f $s.Height, $s.Width)`
	s, err := Start(exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-Command", script), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	if err := s.Resize(120, 40); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write([]byte("go\r")); err != nil {
		t.Fatal(err)
	}
	if out := readAll(t, s); !strings.Contains(out, "size 40 120") {
		t.Fatalf("output = %q, want rows and columns 40 120", out)
	}
}

func TestWindows_CloseWhileRunningEndsTheProcess(t *testing.T) {
	s, err := Start(exec.Command("ping.exe", "-n", "60", "127.0.0.1"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _, _ = readAllQuiet(s) }()
	time.Sleep(500 * time.Millisecond) // let ping start
	done := make(chan error, 1)
	go func() { done <- s.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("close: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Close did not end the process")
	}
	if code := waitCode(t, s); code == 0 {
		t.Fatal("a killed process must not report success")
	}
	if err := s.Close(); err != nil {
		t.Errorf("second close = %v, want the first result", err)
	}
}

func readAllQuiet(s *Session) (string, error) {
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := s.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			return b.String(), err
		}
	}
}
