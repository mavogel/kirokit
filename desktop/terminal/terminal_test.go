package terminal

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func readAll(t *testing.T, s *Session) string {
	t.Helper()
	out, err := io.ReadAll(s)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(out)
}

func waitCode(t *testing.T, s *Session) int {
	t.Helper()
	type result struct {
		code int
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		code, err := s.Wait()
		ch <- result{code, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("wait: %v", r.err)
		}
		return r.code
	case <-time.After(10 * time.Second):
		t.Fatal("process did not end")
		return 0
	}
}

func TestOutputReachesTheReaderAndEndsOnEOF(t *testing.T) {
	s, err := Start(exec.Command("/bin/sh", "-c", "echo hi"), 80, 24)
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

func TestExitCodeIsReturned(t *testing.T) {
	s, err := Start(exec.Command("/bin/sh", "-c", "exit 3"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	readAll(t, s)
	if code := waitCode(t, s); code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
	// Waiting again returns the same result: the child was reaped once.
	if code := waitCode(t, s); code != 3 {
		t.Fatalf("second wait = %d, want 3", code)
	}
}

func TestResizeChangesTheSizeTheProcessSees(t *testing.T) {
	s, err := Start(exec.Command("/bin/sh", "-c", "read x; stty size"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close() //nolint:errcheck // cleanup
	if err := s.Resize(120, 40); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write([]byte("go\n")); err != nil {
		t.Fatal(err)
	}
	if out := readAll(t, s); !strings.Contains(out, "40 120") {
		t.Fatalf("output = %q, want rows and columns 40 120", out)
	}
}

func TestCloseWhileRunningEndsTheProcess(t *testing.T) {
	s, err := Start(exec.Command("/bin/sh", "-c", "sleep 60"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if code := waitCode(t, s); code == 0 {
		t.Fatal("a killed process must not report success")
	}
}
