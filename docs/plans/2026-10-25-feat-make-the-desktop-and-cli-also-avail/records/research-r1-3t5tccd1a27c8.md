---
type: research
---

# Desktop terminal: how processes start and stop, and where it is Unix-only

## Findings

### Every terminal process goes through one wrapper, `terminal.Session`, built on creack/pty
**Answers:** 1

`terminal.Start` is the only place a process gets a pseudo-terminal: it calls `pty.StartWithSize(cmd, &pty.Winsize{...})` (`desktop/terminal/terminal.go:33`). Both callers use it: button commands in the kirokit tab (`desktop/runner.go:114`) and the long-lived `$SHELL -il` tab (`desktop/shell.go:257`).

```
App.Run / runProgram -> launchLocked -> terminal.Start   (desktop/runner.go:45,56,101,114)
App.ShellResize/ShellWrite -> startShellLocked -> shellCommand -> terminal.Start   (desktop/shell.go:233,219,252,190,257)
terminal.Start -> pty.StartWithSize -> (Setsid+Setctty set by creack/pty) -> exec.Cmd.Start
```

The library's non-Windows `StartWithSize` forces `SysProcAttr.Setsid = true` and `Setctty = true` (creack/pty v1.1.24, `start.go:20-23`, module cache). On Windows the same function is `return nil, ErrUnsupported` (`start_windows.go`, same module; `go.mod:6` pins v1.1.24). So Windows fails at runtime in the library even if our own code compiled.

### Stopping is a Unix process-group SIGKILL
**Answers:** 1

`Session.Close` does `syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)` (`desktop/terminal/terminal.go:81`). The negative pid works only because pty made the child a session and group leader (comment at `terminal.go:80`). It then calls `Wait` (reaps via `cmd.Wait`, `terminal.go:63`) and closes the pty file (`terminal.go:84`). `Close` and `Wait` are guarded by `sync.Once` (`terminal.go:22,26`). There is no graceful SIGTERM or SIGHUP step; the kill is immediate.

Callers of `Close`: `finishRun` after output ends (`desktop/runner.go:144`), `runHost.close` at app shutdown (`runner.go:187`), the shell goroutine (`shell.go:268`) and `shellHost.close` (`shell.go:287`). Exit code: `ExitError.ExitCode()`, which is -1 when a signal ended the process (`terminal.go:68`, doc `terminal.go:59-60`).

### EOF detection depends on Unix pty errno EIO
**Answers:** 1

`Session.Read` maps `syscall.EIO` to `io.EOF` (`desktop/terminal/terminal.go:45-46`). `stream` (`desktop/shell.go:296-307`) stops on any read error, so the output loops of `finishRun` and the shell goroutine end this way. `syscall.EIO` itself exists on Windows, but the semantic (closed slave reads as EIO on a master) is Unix pty behaviour.

### Resize uses `pty.Setsize`
**Answers:** 1

`Session.Resize` calls `pty.Setsize` (`desktop/terminal/terminal.go:56`). The library has `winsize_unix.go` and a Windows stub `winsize_unsupported.go` (build-tagged `windows`). I did not read what the stub returns.

### Which files fail to compile with GOOS=windows or GOOS=linux
**Answers:** 1

Commands run in `desktop/`:
- `GOOS=windows go vet ./...` and `GOOS=windows go build ./...` report only `terminal/terminal.go:81:16: undefined: syscall.Kill`. Only that error is reported because the main package depends on `terminal` and is not type-checked after it fails.
- `GOOS=linux go build ./...` and `GOOS=linux go vet ./...` print nothing (clean). `GOOS=darwin go vet ./...` is also clean.
- Inference, with partial evidence: `desktop/envcapture.go` also would not compile on Windows. It uses `&syscall.SysProcAttr{Setpgid: true}` (`envcapture.go:103`) and `syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)` (`envcapture.go:104`). `GOOS=windows go doc syscall.Kill` says "no symbol Kill", and `go doc syscall.SysProcAttr` for windows lists no `Setpgid` field. I did not compile the main package on its own, so these errors are hidden behind the `terminal` failure in the vet output.
- No other non-test file in `desktop/` imports `syscall` or has a `go:build` constraint (grep of `desktop/**/*.go` excluding frontend). `runner.go` and `shell.go` themselves do not import `syscall`; they compile on Linux, and on Windows their only blocker is the dependency on `terminal`.
- `GOOS=windows` also leaves in place runtime assumptions that do not stop compilation, all fact from source:
  - default shell `/bin/zsh` (`desktop/shell.go:85`, `desktop/envcapture.go:95`);
  - PATH joined with `":"` (`desktop/shell.go:122`);
  - login-shell capture via `-ilc ... exec env -0` (`desktop/envcapture.go:100`);
  - zsh `ZDOTDIR` wrapper files (`desktop/shell.go:139-176`, used only when `isZsh`, `shell.go:197`);
  - macOS bundle layout `../Resources/bin/kirokit`, with no `.exe` suffix (`desktop/shell.go:57-58`);
  - fallback PATH `/usr/bin:/bin:/usr/sbin:/sbin` (`desktop/envcapture.go:83`) and tool folders `/opt/homebrew/bin` etc. (`desktop/envcapture.go:48`).
- On Linux the code compiles. `creack/pty` supports linux (`pty_linux.go`). `/bin/zsh` and the `Resources/bin` lookup are macOS-biased defaults, so Linux relies on `$KIROKIT_DESKTOP_CLI` or `build/dev-bin`.

### Change history and ownership
**Answers:** 1

`git log --oneline --since=90.days -- desktop/terminal desktop/runner.go desktop/shell.go` returns a single commit: `6c19579 feat(desktop)!: add the kirokit desktop app ... (#6)`. These files are new and have not been modified since.

## How it is tested today

- `desktop/terminal/terminal_test.go`: four tests, run on a real pty with `/bin/sh` (`exec.Command("/bin/sh", "-c", ...)`, line 44): output and EOF (43), exit code (57), resize (73), close-while-running ends the process (90). Helpers `readAll` and `waitCode` (10s timeout).
- `desktop/runner_test.go` and `desktop/shell_test.go` use `shellApp(t, "/bin/sh")` and generate small `#!/bin/sh` scripts as fake CLIs (`shell_test.go:58`, `runner_test.go:85`). All tests are Unix-only in practice, as they need `/bin/sh` and a pty.
- One package: `cd desktop && go test ./terminal` ; one test: `cd desktop && go test ./terminal -run TestCloseWhileRunningEndsTheProcess`; one file's package: `cd desktop && go test . -run 'TestRunner|TestShell'`. Whole module: `cd desktop && go vet ./... && go test ./...` (this is the last step of `task desktop:test`, `Taskfile.yml:32-37`, which also builds and tests the frontend with `node --test`).
- Not run by me: the tests. I ran only build/vet for linux, darwin and windows.

## Open questions

None.

## Code references

- `desktop/terminal/terminal.go:33` (pty start), `:45` (EIO to EOF), `:56` (Setsize), `:61-74` (Wait), `:77-87` (Close, `syscall.Kill` at 81)
- `desktop/runner.go:101-129` (launchLocked, `terminal.Start` at 114), `:134-165` (finishRun), `:182-189` (close)
- `desktop/shell.go:190-216` (shellCommand), `:252-277` (startShellLocked), `:279-289` (close), `:296-307` (stream)
- `desktop/envcapture.go:100-104` (separate Unix-only `Setpgid`/`syscall.Kill`)
- creack/pty v1.1.24 in the Go module cache: `start.go`, `start_windows.go`, `pty_unsupported.go`
- Exhaustive for syscall/build-tag/`exec.Command` use in non-test, non-frontend Go files under `desktop/`.
