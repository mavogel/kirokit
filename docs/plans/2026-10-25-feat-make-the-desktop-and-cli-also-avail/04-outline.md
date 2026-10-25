---
type: outline
---

# Outline: Ship CLI and desktop app for Linux and Windows

## Summary

Seven slices and 24 tasks: the CLI first (secrets through `cmd.exe`, then Windows checkouts and a
three-OS CI), then the desktop app (Windows terminal, per-OS paths and shell, packaging and smoke
start), then the split release and the docs. The risky parts are the ConPTY terminal and the
`cmd.exe` launch, which this Mac can only compile; their runtime proof is the Windows CI runner.

## How the build runs

<!-- ql:waves -->
<!-- /ql:waves -->

## How each point is proved

<!-- ql:proof-map -->
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** On a Windows runner, an MCP server that `kirokit install` wrote into a fresh `KIRO_HOME` starts through the real `sops exec-env`, exactly as Kiro CLI would launch it, and receives every argument and its decrypted secret unchanged. The same test runs on Linux and macOS.
- **Truth:** On macOS, `kirokit render`, `kirokit install`, the DMGs, the `.app` layout and every word the app shows stay byte-for-byte as today; the only golden that changes is the platform-team `Authorization:` line the Plan's first decision names.

## Decided while outlining

- **Vendored execute bits are recorded in the lock** — Windows reports no execute bit, so the lock hash of a skill with a script could never verify there; `skills.lock.yaml` entries gain an `executables` list that Windows uses in place of the file mode (see *Program design*).
- **Knowledge-base links and backups share one link helper** — a Windows junction is not a symlink to Go, so install, uninstall and backup all ask `internal/dirlink` whether a path is kirokit's link.
- **`kirokit setup --source` accepts Windows drive paths too** — the same address rule as the app's Settings, so a Windows user can point either at `C:\…`.
- **Only `catalog.go` needs `hideConsole`** — `clone.go` runs git through the terminal session, which already owns a pseudo console, so it opens no window.
- **Per-OS values are pure functions of the OS name** — every Linux and Windows branch is then tested on any machine, not only on its own runner.
- **Local Linux checks run in Docker** — `scripts/linux-docker.sh` mirrors the Ubuntu runner (Go, Node, WebKitGTK 4.1, NSIS) on OrbStack, so Linux tests and the Windows installer build run on this Mac; Windows runtime tests run only on GitHub's runner.

## Slice 1: Windows MCP servers get their SOPS secrets

After this slice, `kirokit install` on Windows writes a `sops exec-env` launch that `cmd.exe` runs
unchanged, and refuses an argument it cannot pass. The Oracle test runs on every OS that has sops.

```diff
 internal/
 ├── sopswrap/sopswrap.go         ~ BuildWrapperFor(POSIX|Cmd), ErrUnsafeForCmd
 ├── render/{render,context,funcs}.go ~ Options{KiroHome, LaunchShell}, RenderTeamWith
 └── install/install.go           ~ forward-slash KiroHome, Cmd on Windows
 teams/platform-team/agents/devops/agent.json.tmpl ~ Authorization:${ATLASSIAN_AUTH}
 cmd/kirokit/commands/
+└── install_sops_test.go         + the Oracle test
 testdata/
+└── sops/                        + test-only age key and encrypted fixture secrets
```

- [ ] T1 `sopswrap` builds a `cmd.exe` launch string and refuses what it cannot pass
  - **Objective:** Add `BuildWrapperFor(Cmd, …)`, which joins tokens with single spaces and returns `ErrUnsafeForCmd` naming the first token with whitespace, `"` or `& | < > ^ % ! ( )`; `BuildWrapper` stays the POSIX form.
  - **Design:** "`sopswrap` gains a `cmd.exe` flavour that refuses what it cannot pass"
  - **Files:** `internal/sopswrap/sopswrap.go`, `internal/sopswrap/sopswrap_test.go`, `internal/sopswrap/sopswrap_windows_test.go`
  - **Test first:** a table test calling `BuildWrapperFor(Cmd, …)` with plain tokens, each refused character and a space; it fails to compile today because `BuildWrapperFor` and `Cmd` do not exist.
  - **Definition of done:**
    - `Cmd` with `npx -y mcp-remote@0.15.1 https://h/v2/mcp --header Authorization:${X}` returns `exec-env`, the file, and those tokens joined by single spaces, with no quotes added.
    - Each of space, tab, `"`, `&`, `|`, `<`, `>`, `^`, `%`, `!`, `(`, `)` in any token returns an error that `errors.Is(err, ErrUnsafeForCmd)` and quotes that token; an empty token is refused too.
    - `BuildWrapper` and `BuildWrapperFor(POSIX, …)` return today's output byte for byte; the `/bin/sh -c` round-trip test runs off Windows only.
    - On Windows, the joined string run as `cmd.exe /C <string>` (sops's exact call) starts a helper that reports exactly the original argv.
    - verify: `go test ./internal/sopswrap/ && GOOS=windows go vet ./internal/sopswrap/`

- [ ] T2 Render takes the launch shell as an option and names the failing agent
  - **Objective:** Add `render.Options{KiroHome, LaunchShell}` and `RenderTeamWith`; `sopsArgsJSON` builds through the context's shell, and a refusal reads `render: <team>/<agent>: argument "<arg>" cannot pass through sops exec-env on Windows (cmd.exe): …`.
  - **Design:** "Render takes the launch shell as an option; goldens always use POSIX"
  - **Files:** `internal/render/render.go`, `internal/render/context.go`, `internal/render/funcs.go`, `internal/render/render_test.go`
  - **Test first:** render a fixture agent whose `sopsArgsJSON` has the argument `Authorization: x` with `LaunchShell: sopswrap.Cmd`; it fails to compile today because `RenderTeamWith` does not exist.
  - **Definition of done:**
    - With `Cmd`, the fixture's `args[2]` is the plain space-joined string; with POSIX (and through `RenderTeam`), the output is byte-identical to today's.
    - A refused argument fails the render with the error above, naming `fixture-team/agent1` and the argument, and writes no output folder.
    - `RenderTeam(…, kiroHome)` keeps its signature and callers (`kirokit render`, goldens, the app's catalog).
    - verify: `go test ./internal/render/`

- [ ] T3 platform-team's Atlassian header passes through `cmd.exe`
  - **Objective:** Change the devops agent's header argument to mcp-remote's documented `Authorization:${ATLASSIAN_AUTH}` and regenerate that golden.
  - **Design:** "plan:decision:d1"
  - **Files:** `teams/platform-team/agents/devops/agent.json.tmpl`, `testdata/golden/platform-team/rendered/agents/devops.json`
  - **Test first:** **Trivial:** `go test ./internal/golden/` fails after the template edit until `./kirokit update platform-team` rewrites the golden.
  - **Definition of done:**
    - `git diff testdata/golden` shows exactly one changed line, `--header 'Authorization: ${ATLASSIAN_AUTH}'` becoming `--header 'Authorization:${ATLASSIAN_AUTH}'`.
    - Rendering platform-team with `LaunchShell: Cmd` succeeds (no other platform-team or data-team argument is refused).
    - verify: `go build -o kirokit ./cmd/kirokit && ./kirokit test && go test ./internal/golden/`

- [ ] T4 `kirokit install` bakes forward-slash paths and the `cmd.exe` launch on Windows
  - **Objective:** `Install` and `Check` call `RenderTeamWith` with `KiroHome: filepath.ToSlash(kiroHome)` and `sopswrap.Cmd` when `runtime.GOOS == "windows"`, POSIX elsewhere.
  - **Design:** "Render takes the launch shell as an option; goldens always use POSIX"
  - **Files:** `internal/install/install.go`, `internal/install/install_test.go`
  - **Test first:** on Windows, installing a fixture whose MCP argument has a space must fail naming the agent, and the installed secrets path in `args[1]` must contain no `\`; both fail today (POSIX quotes, `filepath.Join` separators).
  - **Definition of done:**
    - On macOS and Linux the installed agent JSON is byte-identical to today's for the same `KIRO_HOME`.
    - On Windows the installed `args[1]` is `C:/…/kirokit-secrets/<team>/<agent>/secrets.enc.yaml` and that file exists.
    - On Windows an unsafe argument fails `install` before anything is written to `KIRO_HOME` (the manifest folder stays absent).
    - verify: `go test ./internal/install/ && GOOS=windows go vet ./internal/install/`

- [ ] T5 The Oracle test launches an installed MCP server through the real sops
  - **Objective:** Add a test that installs a fixture team into a fresh `KIRO_HOME`, then runs the installed agent's `command` and `args` as Kiro CLI does, with a committed test-only age key, and checks the argv and the decrypted secret the server receives.
  - **Design:** "The Oracle fixture"
  - **Files:** `cmd/kirokit/commands/install_sops_test.go`, `testdata/sops/age-test-key.txt`, `testdata/sops/secrets.enc.yaml`, `.gitleaks.toml`
  - **Test first:** the Oracle test itself; on Windows it fails before T1–T4 because the POSIX-quoted string reaches `cmd.exe` with literal `'` quotes, and the `argv-echo` server reports mangled arguments.
  - **Definition of done:**
    - The server receives exactly `https://example.com/mcp?team=x`, `--header`, `Authorization:${TEST_AUTH}`, `user@example.com`, `a=b`, `--port`, `8080`, and `TEST_AUTH` equals the decrypted `Bearer s3cret value`.
    - Without `sops` on `PATH` the test skips; with `KIROKIT_REQUIRE_SOPS=1` it fails instead of skipping.
    - A wrong key (`SOPS_AGE_KEY_FILE` pointing at an empty file) makes the launch exit non-zero, and the test asserts that too.
    - `gitleaks git --staged` passes with the test key staged, and only `testdata/sops/age-test-key.txt` is allowlisted.
    - verify: `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v`

**Validation:** automated: `go test ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` · `go build -o kirokit ./cmd/kirokit && ./kirokit test` · manual: none

## Slice 2: The CLI works on a Windows checkout, and CI tests it on three systems

After this slice, a Windows checkout keeps goldens, lock hashes and SOPS files byte-exact, the CLI's
own tests run there, and CI runs build, vet and tests on Ubuntu, Windows and macOS.

**Depends on:** Slice 1

- [ ] T6 Windows checkouts stay byte-exact, and Unix-only CLI tests skip or adapt
  - **Objective:** Add `.gitattributes` (LF everywhere, `-text` for goldens, vendored skills and `*.enc.yaml`, binary for media); tests use `os.DevNull` and skip mode-bit and unprivileged-symlink checks on Windows.
  - **Design:** "plan:decision:d6"
  - **Files:** `.gitattributes`, `cmd/kirokit/commands/sync_test.go`, `cmd/kirokit/commands/skills_test.go`, `internal/externalskills/fetch_test.go`, `internal/externalskills/lock_test.go`, `internal/install/install_test.go`, `internal/render/externalskills_test.go`, `internal/backup/backup_test.go`
  - **Trivial:** `go test ./...` on macOS and in the Linux container, and `GOOS=windows go vet ./...`; the Windows runner (T10) is the runtime check.
  - **Definition of done:**
    - `git add --renormalize . && git status --short` changes no tracked file.
    - `git check-attr -a testdata/golden/platform-team/rendered/agents/devops.json teams/platform-team/agents/devops/env/secrets.enc.yaml` reports `text: unset` for both.
    - verify: `go test ./... && GOOS=windows go vet ./...`

- [ ] T7 The skills lock records execute bits, so Windows verifies vendored skills
  - **Objective:** `LockEntry` gains `executables` (slash paths with the execute bit), written by `skills fetch` from git's index; on Windows, `HashDir` for verification takes execute bits from it instead of the file mode.
  - **Design:** "Vendored execute bits on Windows"
  - **Files:** `internal/externalskills/lock.go`, `internal/externalskills/verify.go`, `internal/externalskills/fetch.go`, `internal/externalskills/lock_test.go`, `internal/externalskills/fetch_test.go`
  - **Test first:** hash a skill folder with a `0o755` script through the Windows path (files read without execute bits, `executables` listing the script) and compare with the macOS hash; it fails today because the flag comes only from the file mode.
  - **Definition of done:**
    - The hash on macOS and Linux is unchanged for every existing lock; a lock without `executables` still verifies there.
    - `skills fetch` writes `executables` sorted, omitted when empty, from the fetched tree's git modes (`100755`), on every OS.
    - On Windows, a lock whose hash needs execute bits but has no `executables` fails with a message saying to run `kirokit skills fetch <team>` again; a content edit still fails as drift.
    - verify: `go test ./internal/externalskills/ ./internal/render/`

- [ ] T8 The knowledge-base link falls back to a directory junction on Windows
  - **Objective:** A new `internal/dirlink` makes a symlink, or on Windows a junction (`mklink /J`) when symlinks need a privilege, and recognises either as kirokit's link; install, uninstall and backup use it.
  - **Design:** "Directory links on Windows"
  - **Files:** `internal/dirlink/dirlink.go`, `internal/dirlink/dirlink_windows.go`, `internal/dirlink/dirlink_other.go`, `internal/dirlink/dirlink_test.go`, `internal/install/install.go`, `internal/backup/backup.go`
  - **Test first:** `dirlink.Make` then `dirlink.Is` on a temp folder, and on Windows a junction made with `mklink /J`; it fails to compile today because the package does not exist.
  - **Definition of done:**
    - On macOS and Linux, `Make` creates a symlink exactly as `os.Symlink` did; install and uninstall behave as today.
    - On Windows, a reinstall replaces kirokit's existing junction, uninstall removes the junction and never the source folder, and a real folder at the mount point is still refused.
    - A backup of a folder holding the link copies it as a link, not as the target's files.
    - verify: `go test ./internal/dirlink/ ./internal/install/ ./internal/backup/ && GOOS=windows go vet ./internal/...`

- [ ] T9 `kirokit setup --source` accepts Windows drive paths
  - **Objective:** The source pattern also accepts `C:\…` and `C:/…`, which still cannot be read as a git option.
  - **Design:** "plan:decision:d4"
  - **Files:** `cmd/kirokit/commands/setup.go`, `cmd/kirokit/commands/setup_test.go`
  - **Test first:** `setup --source C:\repos\team` must pass the address check; it fails today with the `--source` error.
  - **Definition of done:**
    - `C:\repos\team`, `C:/repos/team` and `d:\x` pass the check; `C:`, `C:repo`, `-C:\x` and `C:\a b` are refused.
    - Every address accepted today is still accepted.
    - verify: `go test ./cmd/kirokit/commands/ -run Setup`

- [ ] T10 CI runs the CLI checks on Ubuntu, Windows and macOS, and Linux checks run locally in Docker
  - **Objective:** `ci.yaml` gets an OS matrix that installs sops and sets `KIROKIT_REQUIRE_SOPS=1`, with lint and the agent-asset check on Ubuntu only; `scripts/linux-docker.sh` runs a command in an Ubuntu-like container for local Linux checks.
  - **Design:** "plan:decision:d6"
  - **Files:** `.github/workflows/ci.yaml`, `scripts/linux-docker.sh`, `scripts/linux-docker.Dockerfile`
  - **Trivial:** `actionlint .github/workflows/ci.yaml`, and `scripts/linux-docker.sh go test ./...` passing with `KIROKIT_REQUIRE_SOPS=1`.
  - **Definition of done:**
    - The matrix is `ubuntu-latest`, `windows-latest`, `macos-latest`, `fail-fast: false`; each job builds, vets and runs `go test ./...` with sops installed at a pinned version.
    - `scripts/linux-docker.sh` builds its image once, mounts the checkout, and passes its arguments as the command; it exits with the command's code.
    - verify: `actionlint && scripts/linux-docker.sh env KIROKIT_REQUIRE_SOPS=1 go test ./...`

**Validation:** automated: `go test ./...` · `GOOS=windows go vet ./...` · `actionlint` · `scripts/linux-docker.sh env KIROKIT_REQUIRE_SOPS=1 go test ./...` · manual: none

## Slice 3: The desktop terminal runs on Windows

After this slice, the desktop module compiles for Windows: both tabs run on a ConPTY pseudo console
there, and env capture and child processes use Windows process handling. macOS and Linux keep
creack/pty unchanged.

**Depends on:** Slice 2
**Risky:** a new dependency (`github.com/UserExistsError/conpty` v0.1.4) that only the Windows runner exercises at runtime.

- [ ] T11 `terminal.Session` gets a ConPTY backend behind the same API
  - **Objective:** Move today's `terminal.go` unchanged to `terminal_unix.go` (`//go:build !windows`) and add `terminal_windows.go` with the same exported shape.
  - **Design:** "Terminal backends share one API"
  - **Files:** `desktop/terminal/terminal.go`, `desktop/terminal/terminal_unix.go`, `desktop/terminal/terminal_windows.go`, `desktop/terminal/terminal_test.go`, `desktop/terminal/terminal_windows_test.go`, `desktop/go.mod`, `desktop/go.sum`
  - **Test first:** Windows tests for output then `io.EOF`, exit code 3 returned twice, resize seen by PowerShell, and `Close` ending a running `ping -n 60` with a non-zero code; they cannot compile today (`undefined: syscall.Kill`).
  - **Red check:** `cd desktop && GOOS=windows go vet ./terminal/`
  - **Definition of done:**
    - The Unix file's code is byte-identical to today's `terminal.go` apart from the build tag; the Unix tests run off Windows only.
    - On Windows, `Start` honours `cmd.Path`, `Args`, `Dir` and `Env`; arguments with spaces reach the child intact.
    - After the process exits, `Read` returns `io.EOF` (not an error), so `stream` ends.
    - `Close` ends the process tree with `taskkill /T /F /PID <pid>` and releases the console; a second `Close` returns the first result.
    - verify: `cd desktop && go test ./terminal/ && GOOS=windows GOARCH=amd64 go vet ./terminal/ && GOOS=windows GOARCH=arm64 go vet ./terminal/`

- [ ] T12 Env capture and helper processes use per-OS process handling
  - **Objective:** `procgroup_unix.go` and `procgroup_windows.go` hold `killTreeOnCancel(cmd)` and `hideConsole(cmd)`; `envcapture.go` and `catalog.go`'s `repoTag` use them, so the whole desktop module builds for Windows.
  - **Design:** "Terminal backends share one API"
  - **Files:** `desktop/procgroup_unix.go`, `desktop/procgroup_windows.go`, `desktop/envcapture.go`, `desktop/catalog.go`
  - **Test first:** **Trivial:** `cd desktop && GOOS=windows go build ./...` fails today on `Setpgid` and `syscall.Kill`; the existing env-capture timeout test proves the Unix kill still ends the group.
  - **Definition of done:**
    - On Unix, a login shell that hangs past the timeout is still killed with its children (existing test passes unchanged).
    - On Windows, `hideConsole` sets `CREATE_NO_WINDOW`; `killTreeOnCancel` ends the tree with `taskkill /T /F`.
    - verify: `cd desktop && go test ./... && GOOS=windows GOARCH=amd64 go build ./... && GOOS=windows GOARCH=arm64 go build ./... && GOOS=linux go vet ./...`

**Validation:** automated: `task desktop:test` · `cd desktop && GOOS=windows go vet ./...` · manual: none

## Slice 4: The app finds its folders, tools, CLI and shell on each OS

After this slice, the app on Linux and Windows keeps its data where each OS expects, finds sops,
kiro-cli and its bundled CLI, opens the right shell named after itself, and shows OS-correct setup
hints. Every word and path on macOS stays as today.

**Depends on:** Slice 3

- [ ] T13 Data folder, age-key path and setup hints follow the OS
  - **Objective:** `NewApp` uses `os.UserConfigDir()/kirokit-desktop`; the age-key default follows sops's rule; the SOPS hint uses the Interface's *Words per OS* table.
  - **Design:** "The app's paths, tools and shell per OS", "Per-OS values are pure functions of GOOS"
  - **Files:** `desktop/platform.go`, `desktop/platform_test.go`, `desktop/app.go`, `desktop/prereq.go`, `desktop/prereq_test.go`
  - **Test first:** table tests for `dataDirFor`, `ageKeyCandidateFor` and `sopsHintFor` over `darwin`, `linux` and `windows`; they fail to compile today because the functions do not exist.
  - **Definition of done:**
    - darwin: `~/Library/Application Support/kirokit-desktop`, key `~/Library/Application Support/sops/age/keys.txt`, hint `Install it with: brew install sops`, exactly as today.
    - linux: `~/.config/kirokit-desktop`, key `$XDG_CONFIG_HOME/sops/age/keys.txt` else `~/.config/sops/age/keys.txt`, hint `Install it from your package manager, or from https://github.com/getsops/sops/releases`.
    - windows: `%AppData%\kirokit-desktop`, key `%AppData%\sops\age\keys.txt` (no XDG step), hint `Install it with: scoop install sops`; `SOPS_AGE_KEY_FILE` wins everywhere.
    - verify: `cd desktop && go test -run 'Platform|Prereq|DataDir' ./...`

- [ ] T14 Env capture, PATH and tool lookup follow the OS
  - **Objective:** On Windows, `Capture` uses the app's own environment with no shell, `Lookup` tries `PATHEXT` suffixes without a mode check, and PATH joins with the OS separator; Linux's login shell defaults to `$SHELL`, `/bin/bash`, then `/bin/sh`; tool folders follow the OS.
  - **Design:** "The app's paths, tools and shell per OS"
  - **Files:** `desktop/envcapture.go`, `desktop/envcapture_test.go`, `desktop/platform.go`, `desktop/platform_test.go`, `desktop/shell.go`
  - **Test first:** tests for `toolFoldersFor`, `defaultShellFor` and a Windows `Lookup` that finds `sops.exe` and `kiro-cli.cmd` with `PATHEXT=.COM;.EXE;.BAT;.CMD`; they fail today (no such functions, `Lookup` demands an execute bit).
  - **Definition of done:**
    - darwin keeps `/opt/homebrew/bin`, `/opt/homebrew/sbin`, `/usr/local/bin`, `~/.local/bin` and `/bin/zsh`; linux adds `/home/linuxbrew/.linuxbrew/bin`, `/usr/local/bin`, `~/.local/bin`; windows adds `%USERPROFILE%\scoop\shims`.
    - `terminalEnv` puts the CLI folder first with `os.PathListSeparator`; on Windows a `Path` key is merged with `PATH` rather than duplicated.
    - On Unix, a file without an execute bit is still not found.
    - verify: `cd desktop && go test -run 'Capture|Lookup|ToolFolders|DefaultShell|TerminalEnv' ./...`

- [ ] T15 The bundled CLI and the shell tab follow the OS, and the state names the shell
  - **Objective:** `resolveCLI` looks beside the exe per OS, with OS-specific error words; the shell tab runs `pwsh.exe -NoLogo` (else `powershell.exe -NoLogo`) on Windows; `EnvStatus.Shell` reports `zsh`, `bash`, `PowerShell`.
  - **Design:** "The app's paths, tools and shell per OS"
  - **Files:** `desktop/shell.go`, `desktop/shell_test.go`, `desktop/app.go`, `desktop/platform.go`, `desktop/platform_test.go`
  - **Test first:** `cliCandidatesFor` and `shellCommandFor` tables over three OSes, and `State().Env.Shell`; they fail today (no such functions or field).
  - **Definition of done:**
    - Candidates: darwin `<exe>/../Resources/bin/kirokit`; linux `<exe>/../lib/kirokit-desktop/kirokit`; windows `<exe dir>\bin\kirokit.exe`; `build/dev-bin/kirokit(.exe)` stays first under `wails dev` on every OS.
    - The CLI-missing error ends with `start the app from kirokit.app`, `reinstall the kirokit-desktop package` or `reinstall kirokit with its setup program`.
    - On Windows no zsh wrapper is written and `shellNote` is empty; on macOS and Linux the zsh wrapper and the non-zsh note stay as today.
    - verify: `cd desktop && go test -run 'BundledCLI|ResolveCLI|Shell|State' ./...`

- [ ] T16 The dock names the shell tab after the running shell
  - **Objective:** The frontend sets the shell tab's label and the kirokit tab's first line from `state.env.shell`, keeping `zsh` until the state arrives.
  - **Design:** "The app's paths, tools and shell per OS"
  - **Files:** `desktop/frontend/src/tabs.js`, `desktop/frontend/src/tabs.test.js`, `desktop/frontend/src/terminal.js`, `desktop/frontend/src/main.js`
  - **Test first:** `node --test` for `shellLabel(state)`: `PowerShell`, `bash`, and `zsh` for a missing or empty value; it fails today because `shellLabel` does not exist.
  - **Definition of done:**
    - With no state yet, the tab reads `zsh` and the first line reads `… Type your own commands in the zsh tab.`, as today.
    - Once `State()` reports `PowerShell`, the tab label changes to `PowerShell`; the first line printed afterwards names it too.
    - verify: `cd desktop/frontend && node --test "src/**/*.test.js"`

- [ ] T17 Settings accept Windows folders and addresses
  - **Objective:** The root check uses `filepath.VolumeName`, and `repoURLPattern` accepts `C:\…` and `C:/…`.
  - **Design:** "plan:decision:d4"
  - **Files:** `desktop/settings.go`, `desktop/settings_test.go`
  - **Test first:** `isRootFolder` for `/`, `C:\`, `C:/`, `D:\` and `repoURLPattern` for `C:\repos\team`; they fail today.
  - **Definition of done:**
    - A drive root is refused like `/`; `C:\Users\alex\.kiro` passes on Windows; today's accepted and refused addresses keep their answers.
    - `C:repo`, `-C:\x` and `C:\a b` are refused as addresses.
    - verify: `cd desktop && go test -run Settings ./...`

- [ ] T18 Desktop tests run on Linux and Windows
  - **Objective:** Fake tools are `.cmd` files on Windows, Unix paths in tests come from `t.TempDir()`, and tests needing `/bin/sh`, zsh or a Unix pty skip on Windows with a Windows counterpart where the behaviour exists.
  - **Design:** "plan:decision:d6"
  - **Files:** `desktop/shell_test.go`, `desktop/runner_test.go`, `desktop/envcapture_test.go`, `desktop/prereq_test.go`, `desktop/clone_test.go`, `desktop/settings_test.go`, `desktop/catalog_test.go`, `desktop/testtools_test.go`
  - **Trivial:** `task desktop:test` on macOS and in the Linux container, `GOOS=windows go vet ./...` here; the Windows runner (T21) runs them.
  - **Definition of done:**
    - On Windows, the runner test runs a fake `kirokit.cmd` through the bundled-CLI path and streams its output to the kirokit tab; the shell test starts PowerShell.
    - No test is skipped on macOS or Linux that runs there today.
    - verify: `task desktop:test && scripts/linux-docker.sh task desktop:test && cd desktop && GOOS=windows go vet ./...`

**Validation:** automated: `task desktop:test` · `scripts/linux-docker.sh task desktop:test` · `cd desktop && GOOS=windows go vet ./...` · manual: none

## Slice 5: The app builds, packages and starts on Linux and Windows

After this slice, `task desktop:build` builds the app on each OS (Linux with WebKitGTK 4.1, Windows
with an NSIS setup carrying `bin\kirokit.exe`), the app has a smoke start, and the Desktop workflow
proves the build starts on all three runners.

**Depends on:** Slice 4

- [ ] T19 With `KIROKIT_DESKTOP_SMOKE=1` the app quits once its page has loaded
  - **Objective:** Wails' `OnDomReady` quits with exit code 0 when the variable is set; a timer exits 1 when the page has not loaded within 60 seconds.
  - **Design:** "plan:decision:d7"
  - **Files:** `desktop/main.go`, `desktop/smoke.go`, `desktop/smoke_test.go`
  - **Test first:** a test of `smokeWatch` with a fake quit and a short timeout: DOM-ready before the deadline quits with 0, none exits 1; it fails to compile today.
  - **Definition of done:**
    - Without the variable the app starts exactly as today.
    - The smoke start never writes settings or clones (it quits before any page action).
    - verify: `cd desktop && go test -run Smoke ./... && task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit`

- [ ] T20 `task desktop:build` builds the Linux and Windows apps
  - **Objective:** `scripts/desktop-build.sh` branches on `go env GOOS`: darwin unchanged, linux builds `kirokit-desktop` with `-tags webkit2_41` and the CLI beside it, windows builds both arches with `-nsis` and `bin\kirokit.exe` in the installer; the Windows and Linux assets are committed.
  - **Design:** "plan:decision:d5"
  - **Files:** `scripts/desktop-build.sh`, `Taskfile.yml`, `desktop/build/windows/icon.ico`, `desktop/build/windows/info.json`, `desktop/build/windows/wails.exe.manifest`, `desktop/build/windows/installer/project.nsi`, `desktop/build/windows/installer/wails_tools.nsh`, `desktop/build/linux/kirokit-desktop.desktop`, `desktop/build/linux/kirokit-desktop.png`, `.gitignore`
  - **Test first:** **Trivial:** `scripts/linux-docker.sh task desktop:build` fails today (the darwin-only script calls `lipo`); after the change it and `GOOS=windows` in the container produce the artifacts below.
  - **Definition of done:**
    - On macOS, the built `kirokit.app` has the same layout and `Info.plist` as before this change.
    - Linux: `desktop/build/bin/kirokit-desktop` and `desktop/build/bin/kirokit`; `kirokit-desktop` links `libwebkit2gtk-4.1`.
    - Windows: `desktop/build/bin/installers/kirokit-desktop_<v>_windows_{amd64,arm64}-setup.exe`, each containing `bin\kirokit.exe` (listed by `7z l`).
    - verify: `scripts/linux-docker.sh task desktop:build && scripts/linux-docker.sh env GOOS=windows task desktop:build`

- [ ] T21 The Desktop workflow tests, builds and smoke-starts the app on three runners
  - **Objective:** `desktop.yaml` runs `task desktop:test`, `task desktop:build` and the smoke start on Ubuntu (under `xvfb-run`), Windows and macOS, and on Linux installs the snapshot's `.deb` with `apt install` and smoke-starts `/usr/bin/kirokit-desktop`.
  - **Design:** "plan:decision:d7"
  - **Files:** `.github/workflows/desktop.yaml`
  - **Trivial:** `actionlint .github/workflows/desktop.yaml`; the pull request's run is the runtime check.
  - **Definition of done:**
    - The smoke step fails the job on a non-zero exit; on Linux it runs under `xvfb-run -a`.
    - The `.deb` install and smoke step runs only when the split snapshot ran (`GORELEASER_KEY` present).
    - verify: `actionlint .github/workflows/desktop.yaml`

**Validation:** automated: `task desktop:build` · `scripts/linux-docker.sh task desktop:build` · `actionlint` · manual: none

## Slice 6: One release ships every OS from three runners

After this slice, GoReleaser builds Windows CLI zips, Linux `.deb` and `.rpm` for the CLI and the
app, a Scoop manifest and the Windows setup files, and the release workflow builds on macOS, Linux and
Windows runners and pushes the tag only after all of them succeed.

**Depends on:** Slice 5
**Risky:** the split and merge publishing order runs for real only on `main`.

- [ ] T22 GoReleaser config and staging cover Linux and Windows
  - **Objective:** `.goreleaser.yaml` gains `partial.by: goos`, the Windows CLI target and zip override, the two `nfpms`, the `kirokit-desktop-linux` prebuilt build, `scoops` (skipped without its token) and `release.extra_files`; `scripts/desktop-prebuilt.sh` stages per OS.
  - **Design:** "Release layout and file names"
  - **Files:** `.goreleaser.yaml`, `scripts/desktop-prebuilt.sh`
  - **Test first:** **Trivial:** `goreleaser check` passes, and `scripts/linux-docker.sh scripts/desktop-prebuilt.sh` stages `prebuilt/kirokit-desktop_linux_<arch>/kirokit-desktop`, which today's script refuses (no `.app`).
  - **Definition of done:**
    - The darwin parts (builds, `app_bundles`, `dmg`, both casks) are unchanged lines.
    - The desktop `.deb` lists `/usr/bin/kirokit-desktop`, `/usr/lib/kirokit-desktop/kirokit`, the `.desktop` file and the icon, and depends on `libgtk-3-0, libwebkit2gtk-4.1-0` (`dpkg-deb -c` / `-f` in the container on a split snapshot when a key is available, else on CI).
    - With `SCOOP_BUCKET_GITHUB_TOKEN` empty, the scoop publish is skipped, not failed.
    - verify: `goreleaser check`

- [ ] T23 The release builds on three runners and pushes the tag last
  - **Objective:** `release.yaml` becomes `version` → `app-linux-arm64` → `split` (macOS, Linux, Windows) → `merge`, which pushes the tag and runs `goreleaser continue --merge`; the dry run prints the casks and the Scoop manifest; `desktop.yaml` runs the split snapshot on pull requests.
  - **Design:** "Release workflow jobs"
  - **Files:** `.github/workflows/release.yaml`, `.github/workflows/desktop.yaml`
  - **Trivial:** `actionlint`; the `workflow_dispatch` dry run on the branch is the runtime check.
  - **Definition of done:**
    - No job before `merge` runs `git push`; every split job tags only locally.
    - Each split job uploads `dist/<os>` under a run-specific artifact name, and `merge` downloads all of them into `dist/`.
    - The dry run never pushes, uploads to a release or publishes to the tap or bucket.
    - verify: `actionlint`

**Validation:** automated: `goreleaser check` · `actionlint` · manual: none

## Slice 7: The docs tell every OS how to install and what to expect

After this slice, README and CONTRIBUTING describe installing and running kirokit and its app on
macOS, Linux and Windows, including prerequisites and the unsigned-app warnings.

**Depends on:** Slice 6

- [ ] T24 README, CONTRIBUTING and AGENTS cover Linux and Windows
  - **Objective:** Install commands per OS (cask on macOS and Linux, `.deb`/`.rpm`, tar.gz, Scoop, zip, setup `.exe`), prerequisites (SOPS, age key path per OS, Kiro CLI), the SmartScreen steps beside the macOS quarantine note, the shell tab's name, the Windows argument rule, and the three-OS CI and Docker checks for contributors.
  - **Design:** none
  - **Files:** `README.md`, `CONTRIBUTING.md`, `AGENTS.md`
  - **Trivial:** `node scripts/sync-agent-assets.mjs --check`, and a read-through against the Done means 8 checklist.
  - **Definition of done:**
    - README names `More info` then `Run anyway` for SmartScreen, and keeps the existing macOS quarantine note.
    - README's install section names each command and file of the release layout, and the age-key path per OS.
    - Every changed command, flag, manifest field or path in this task (`executables`, the Windows argument rule, `scripts/linux-docker.sh`) is documented where the reader looks.
    - verify: `node scripts/sync-agent-assets.mjs --check`

**Validation:** automated: `node scripts/sync-agent-assets.mjs --check` · manual: a reader checks the README install section per OS

## Program design

### Per-OS values are pure functions of GOOS

`desktop/platform.go` holds one function per concern taking the OS name and the inputs it needs:
`dataDirFor(goos, home, configDir)`, `ageKeyCandidateFor(goos, env, home, configDir)`,
`sopsHintFor(goos)`, `toolFoldersFor(goos, home, userProfile)`, `defaultShellFor(goos, env)`,
`cliCandidatesFor(goos, exe, wd)`, `cliMissingHintFor(goos)` and `shellCommandFor(goos, …)`. Callers
pass `runtime.GOOS`; tests pass each of `darwin`, `linux`, `windows`. Only Windows-API calls
(`taskkill`, `CREATE_NO_WINDOW`, ConPTY, junctions) live in build-tagged files.

### Vendored execute bits on Windows

`LockEntry` gains `Executables []string \`yaml:"executables,omitempty"\``: the slash paths, relative
to the skill folder, whose git mode is `100755` in the fetched tree. `HashDir` keeps its digest
format; a new `hashDirWith(dir, execOf func(rel string, info fs.FileInfo) bool)` lets verification on
Windows read the flag from `Executables`. The digest of a skill therefore matches on every OS, and
locks written before this change verify unchanged on macOS and Linux.

### Directory links on Windows

`internal/dirlink` exposes `Make(source, mount) error`, `Is(path) bool` and `Remove(path) error`.
On Windows, `Make` tries `os.Symlink` and, on `ERROR_PRIVILEGE_NOT_HELD`, runs `cmd.exe /C mklink /J
<mount> <source>` (both paths absolute and refused if they contain `"`); `Is` is true for a symlink
or a junction (`os.Readlink` succeeds on the reparse point). Elsewhere it is `os.Symlink` and the
`ModeSymlink` test.

### The Oracle fixture

The test writes a temporary repository (`teams/sops-team` with one agent, `global/`,
`env/mcp-versions.yaml`) whose agent launches `argv-echo` through `sopsArgsJSON` with the arguments
in T5. `argv-echo` is the test binary itself copied to `<tmp>/bin/argv-echo(.exe)` and put first on
`PATH`; when `KIROKIT_ARGV_ECHO_OUT` is set, `TestMain` writes `os.Args[1:]` and `TEST_AUTH` as JSON
to that file and exits. `testdata/sops/secrets.enc.yaml` is encrypted to the public half of
`testdata/sops/age-test-key.txt`, a key used nowhere else; `.gitleaks.toml` extends gitleaks'
defaults and allowlists that one path.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| `cmd.exe` launch string and its refusal | `internal/sopswrap/sopswrap.go` | `BuildWrapper`'s token list |
| Launch shell as a render option | `internal/render/render.go`, `context.go` | `RenderTeam`'s walk |
| Directory link or junction | `internal/dirlink/` | `os.Symlink`, `os.Readlink` |
| Execute bits in the lock | `internal/externalskills/lock.go` | `HashDir`'s digest lines |
| ConPTY terminal backend | `desktop/terminal/terminal_windows.go` | `Session` API |
| Per-OS process helpers | `desktop/procgroup_{unix,windows}.go` | `exec.Cmd.Cancel` |
| Per-OS paths, words and shell | `desktop/platform.go` | `os.UserConfigDir` |
| Smoke start | `desktop/smoke.go` | Wails `OnDomReady` |
| Local Linux runner | `scripts/linux-docker.sh` | OrbStack / Docker |

## End-to-end scenarios

### Scenario 1: A Windows install starts its SOPS-protected MCP server

**Runs after:** slice 2
**Proves:** Done means 3
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | On a pushed branch, open the latest CI run: `gh run list --workflow CI --branch feat/linux-windows --limit 1` | one run, status `completed`, conclusion `success` |
| 2 | `gh run view <run-id> --log \| grep -E 'TestInstall_SopsLaunch'` | `--- PASS: TestInstall_SopsLaunch` in the `windows-latest`, `ubuntu-latest` and `macos-latest` jobs, with no `SKIP` |
| 3 | Locally: `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v` | `--- PASS: TestInstall_SopsLaunch` |
| 4 | `scripts/linux-docker.sh env KIROKIT_REQUIRE_SOPS=1 go test ./... ` | exit 0, every package `ok` |

### Scenario 2: macOS users notice nothing

**Runs after:** slice 6
**Proves:** Done means 6

| Step | Action | Expected |
|---|---|---|
| 1 | `go build -o kirokit ./cmd/kirokit && ./kirokit test` | every team passes |
| 2 | `git diff main --stat -- testdata/golden` | one file changed, `testdata/golden/platform-team/rendered/agents/devops.json`, 1 line in and 1 out |
| 3 | `git diff main -- desktop/build/darwin scripts/desktop-build.sh \| grep -c '^[-+][^-+]' ` | only additions in the non-darwin branches of the build script; `desktop/build/darwin` unchanged |
| 4 | `task desktop:build && ls desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit` | the bundled CLI path exists, as today |
| 5 | `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?` | `0` |
| 6 | `task desktop:dev`, open the app, open the dock | the second tab reads `zsh`; the kirokit tab's first line ends `Type your own commands in the zsh tab.` |
| 7 | Open Install on a team with SOPS missing from `PATH` | the hint reads `Install it with: brew install sops` |

### Scenario 3: The app tests pass and the built app starts on Linux and Windows

**Runs after:** slice 5
**Proves:** Done means 4, Done means 5
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `gh run list --workflow Desktop --branch feat/linux-windows --limit 1` | conclusion `success` |
| 2 | `gh run view <run-id> --log \| grep -E 'ok +github.com/mavogel/kirokit/desktop'` | `ok` for `desktop` and `desktop/terminal` in the `ubuntu-latest`, `windows-latest` and `macos-latest` jobs |
| 3 | `gh run view <run-id> --log \| grep -E 'TestShellCommand.*windows\|--- PASS: TestShellTab_StartsPowerShell'` | the Windows job passes the PowerShell shell-tab test |
| 4 | Same log, smoke steps | each OS's smoke step exits 0; Linux runs under `xvfb-run` |
| 5 | Same log, Linux `.deb` step | `apt install ./dist/linux/kirokit-desktop_*_linux_amd64.deb` succeeds and `/usr/bin/kirokit-desktop` smoke-starts with exit 0 |
| 6 | `scripts/linux-docker.sh task desktop:test` | exit 0 |

### Scenario 4: One release carries every OS's files

**Runs after:** slice 6
**Proves:** Done means 1, Done means 2, Done means 7
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `gh workflow run release.yaml --ref feat/linux-windows -f dry_run=true`, then `gh run watch` | the `version`, `app-linux-arm64`, three `split` and `merge` jobs succeed; no tag is created (`git ls-remote --tags origin` unchanged) |
| 2 | Merge job log, `ls dist` | `kirokit_<v>_linux_{amd64,arm64}.tar.gz`, `.deb`, `.rpm`; `kirokit_<v>_windows_{amd64,arm64}.zip`; `kirokit-desktop_<v>_linux_{amd64,arm64}.deb` and `.rpm`; `kirokit-desktop_<v>_windows_{amd64,arm64}-setup.exe`; both darwin DMGs and tar.gz files as before |
| 3 | Merge job log, printed casks and Scoop manifest | the `kirokit` cask has `on_linux` URLs for both arches; `kirokit.json` names the two Windows zips |
| 4 | Download the run's Linux artifact, then `scripts/linux-docker.sh sh -c 'apt-get install -y ./dist/linux/kirokit_*_linux_arm64.deb && kirokit --version'` | the snapshot version |
| 5 | Pull request run: `gh pr checks` | CI and Desktop pass on Ubuntu, Windows and macOS |
| 6 | After the first release on `main`: `scoop install kirokit` and `brew install --cask mavogel/tap/kirokit` on Linux | each prints the release version for `kirokit --version` |

### Scenario 5: The README guides Linux and Windows users

**Runs after:** slice 7
**Proves:** Done means 8
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `grep -n 'scoop install kirokit\|\.deb\|\.rpm\|setup.exe\|brew install --cask mavogel/tap/kirokit' README.md` | install lines for Linux (`.deb`, `.rpm`, tar.gz, cask) and Windows (Scoop, zip, setup `.exe`) |
| 2 | `grep -n 'More info\|Run anyway\|xattr -dr com.apple.quarantine' README.md` | the SmartScreen steps and the existing macOS quarantine note |
| 3 | `grep -n 'keys.txt' README.md` | the age-key path for macOS, Linux and Windows |

## Commands the build runs

<!-- ql:commands -->
<!-- /ql:commands -->
