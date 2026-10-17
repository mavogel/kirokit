// Package terminal runs a process on a pseudo-terminal and streams its
// output, so the app's dock can show a real shell or a kirokit command.
package terminal

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/creack/pty"
)

// Session is one process on its own pseudo-terminal. Read returns its output,
// Write sends it keystrokes and Wait returns its exit code once it is reaped.
type Session struct {
	cmd *exec.Cmd
	pty *os.File

	waitOnce sync.Once
	code     int
	waitErr  error

	closeOnce sync.Once
	closeErr  error
}

// Start runs cmd on a new pseudo-terminal of the given size. The child leads
// its own session, so Close can end everything it started.
func Start(cmd *exec.Cmd, cols, rows uint16) (*Session, error) {
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, err
	}
	return &Session{cmd: cmd, pty: f}, nil
}

// Read returns the process's output. It returns io.EOF when the process and
// everything holding the terminal have closed it.
func (s *Session) Read(p []byte) (int, error) {
	n, err := s.pty.Read(p)
	// A closed slave side reads as EIO on a pseudo-terminal master.
	if errors.Is(err, syscall.EIO) {
		err = io.EOF
	}
	return n, err
}

// Write sends keystrokes to the process.
func (s *Session) Write(p []byte) (int, error) { return s.pty.Write(p) }

// Resize changes the size the process sees.
func (s *Session) Resize(cols, rows uint16) error {
	return pty.Setsize(s.pty, &pty.Winsize{Cols: cols, Rows: rows})
}

// Wait blocks until the process ends, reaps it and returns its exit code, or
// -1 when a signal ended it. It may be called more than once.
func (s *Session) Wait() (int, error) {
	s.waitOnce.Do(func() {
		err := s.cmd.Wait()
		var exit *exec.ExitError
		switch {
		case err == nil:
		case errors.As(err, &exit):
			s.code = exit.ExitCode()
		default:
			s.code, s.waitErr = -1, err
		}
	})
	return s.code, s.waitErr
}

// Close ends the process if it still runs, reaps it and releases the terminal.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		if s.cmd.Process != nil {
			// The child leads its own process group; a negative pid reaches all of it.
			_ = syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // already gone is fine
		}
		_, _ = s.Wait() //nolint:errcheck // the code is read through Wait
		s.closeErr = s.pty.Close()
	})
	return s.closeErr
}
