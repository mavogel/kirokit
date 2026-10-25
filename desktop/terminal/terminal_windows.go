//go:build windows

// Package terminal runs a process on a pseudo-terminal and streams its
// output, so the app's dock can show a real shell or a kirokit command. On
// Windows the pseudo-terminal is a pseudo console (ConPTY).
package terminal

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Session is one process on its own pseudo console. Read returns its output,
// Write sends it keystrokes and Wait returns its exit code once it has ended.
type Session struct {
	hpc     windows.Handle
	process windows.Handle
	pid     uint32
	in      *os.File // keystrokes to the console
	out     *os.File // what the console draws

	exited  chan struct{} // closed once the process has ended and code is set
	code    int
	waitErr error

	consoleMu     sync.Mutex
	consoleClosed bool

	closeOnce sync.Once
	closeErr  error
}

// Start runs cmd on a new pseudo console of the given size. cmd.Path,
// cmd.Args, cmd.Dir and cmd.Env are honoured; its other fields are not.
func Start(cmd *exec.Cmd, cols, rows uint16) (*Session, error) {
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	line := commandLine(cmd)
	lineW, err := windows.UTF16PtrFromString(line)
	if err != nil {
		return nil, err
	}
	var dirW *uint16
	if cmd.Dir != "" {
		if dirW, err = windows.UTF16PtrFromString(cmd.Dir); err != nil {
			return nil, err
		}
	}

	// consoleIn/consoleOut are the console's own ends; it keeps duplicates, so
	// ours are closed once it exists. Then the output pipe ends (Read returns
	// io.EOF) when the console is closed, after the process has exited.
	var consoleIn, inW, outR, consoleOut windows.Handle
	if err := windows.CreatePipe(&consoleIn, &inW, nil, 0); err != nil {
		return nil, err
	}
	if err := windows.CreatePipe(&outR, &consoleOut, nil, 0); err != nil {
		closeHandles(consoleIn, inW)
		return nil, err
	}
	var hpc windows.Handle
	err = windows.CreatePseudoConsole(windows.Coord{X: int16(cols), Y: int16(rows)}, consoleIn, consoleOut, 0, &hpc)
	closeHandles(consoleIn, consoleOut)
	if err != nil {
		closeHandles(inW, outR)
		return nil, err
	}

	pi, err := startAttached(hpc, lineW, dirW, cmd.Env)
	if err != nil {
		windows.ClosePseudoConsole(hpc)
		closeHandles(inW, outR)
		return nil, err
	}
	_ = windows.CloseHandle(pi.Thread)

	s := &Session{
		hpc:     hpc,
		process: pi.Process,
		pid:     pi.ProcessId,
		in:      os.NewFile(uintptr(inW), "conpty-in"),
		out:     os.NewFile(uintptr(outR), "conpty-out"),
		exited:  make(chan struct{}),
	}
	go s.watch()
	return s, nil
}

// commandLine quotes the program and its arguments the way the C runtime and
// Go programs read them back.
func commandLine(cmd *exec.Cmd) string {
	parts := make([]string, 0, len(cmd.Args))
	parts = append(parts, syscall.EscapeArg(cmd.Path))
	if len(cmd.Args) > 1 {
		for _, a := range cmd.Args[1:] {
			parts = append(parts, syscall.EscapeArg(a))
		}
	}
	return strings.Join(parts, " ")
}

// startAttached creates the process with the pseudo console as its console.
// STARTF_USESTDHANDLES with no handles keeps it from inheriting this
// process's own (possibly redirected) standard handles instead.
func startAttached(hpc windows.Handle, line, dir *uint16, env []string) (*windows.ProcessInformation, error) {
	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return nil, err
	}
	defer attrs.Delete()
	// The attribute's value is the console handle itself, not a pointer to it.
	if err := attrs.Update(windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, *(*unsafe.Pointer)(unsafe.Pointer(&hpc)), unsafe.Sizeof(hpc)); err != nil {
		return nil, err
	}
	var si windows.StartupInfoEx
	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = windows.STARTF_USESTDHANDLES
	si.ProcThreadAttributeList = attrs.List()

	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT)
	var envW *uint16
	if env != nil {
		flags |= windows.CREATE_UNICODE_ENVIRONMENT
		envW = envBlock(env)
	}
	var pi windows.ProcessInformation
	if err := windows.CreateProcess(nil, line, nil, nil, false, flags, envW, dir, &si.StartupInfo, &pi); err != nil {
		return nil, err
	}
	return &pi, nil
}

// envBlock is env as CreateProcess wants it: NUL-separated, NUL-terminated
// UTF-16.
func envBlock(env []string) *uint16 {
	var b strings.Builder
	for _, kv := range env {
		if strings.IndexByte(kv, 0) >= 0 {
			continue
		}
		b.WriteString(kv)
		b.WriteByte(0)
	}
	if len(env) == 0 {
		b.WriteByte(0)
	}
	b.WriteByte(0)
	return &utf16.Encode([]rune(b.String()))[0]
}

// watch waits for the process, keeps its exit code and closes the console, so
// the output pipe drains and ends.
func (s *Session) watch() {
	_, err := windows.WaitForSingleObject(s.process, windows.INFINITE)
	if err == nil {
		var code uint32
		if err = windows.GetExitCodeProcess(s.process, &code); err == nil {
			s.code = int(code)
		}
	}
	if err != nil {
		s.code, s.waitErr = -1, err
	}
	close(s.exited)
	s.closeConsole()
}

func (s *Session) closeConsole() {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()
	if !s.consoleClosed {
		s.consoleClosed = true
		windows.ClosePseudoConsole(s.hpc)
	}
}

// Read returns the process's output. It returns io.EOF once the process has
// ended and its last output has been read.
func (s *Session) Read(p []byte) (int, error) { return s.out.Read(p) }

// Write sends keystrokes to the process.
func (s *Session) Write(p []byte) (int, error) { return s.in.Write(p) }

// Resize changes the size the process sees.
func (s *Session) Resize(cols, rows uint16) error {
	s.consoleMu.Lock()
	defer s.consoleMu.Unlock()
	if s.consoleClosed {
		return errors.New("terminal: the process has ended")
	}
	return windows.ResizePseudoConsole(s.hpc, windows.Coord{X: int16(cols), Y: int16(rows)})
}

// Wait blocks until the process ends and returns its exit code. It may be
// called more than once.
func (s *Session) Wait() (int, error) {
	<-s.exited
	return s.code, s.waitErr
}

// Close ends the process and everything it started if it still runs, then
// releases the console.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		select {
		case <-s.exited:
		default:
			// taskkill /T reaches the whole tree the shell started.
			kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.FormatUint(uint64(s.pid), 10))
			kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
			if kill.Run() != nil {
				_ = windows.TerminateProcess(s.process, 1)
			}
			<-s.exited
		}
		s.closeConsole()
		s.closeErr = errors.Join(s.in.Close(), s.out.Close(), windows.CloseHandle(s.process))
	})
	return s.closeErr
}

func closeHandles(hs ...windows.Handle) {
	for _, h := range hs {
		_ = windows.CloseHandle(h)
	}
}
