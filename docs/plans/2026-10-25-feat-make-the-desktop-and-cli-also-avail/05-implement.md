# Ship CLI and desktop app for Linux and Windows · the build

## Summary

<!-- ql:summary -->
13 of 24 tasks · 3 of 7 slices

Building now: started 25 Oct, 19:12, 44 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | Windows MCP servers get their SOPS secrets | After this slice, `kirokit install` on Windows writes a `sops exec-env` launch that `cmd.exe` runs unchanged, and refuses an argument it cannot pass. | done |
| 2 | The CLI works on a Windows checkout, and CI tests it on three systems | After this slice, a Windows checkout keeps goldens, lock hashes and SOPS files byte-exact, the CLI's own tests run there, and CI runs build, vet and tests on Ubuntu, Windows and macOS. | done |
| 3 | The desktop terminal runs on Windows | After this slice, the desktop module compiles for Windows: both tabs run on a ConPTY pseudo console there, and env capture and child processes use Windows process handling. | done |
| 4 | The app finds its folders, tools, CLI and shell on each OS | After this slice, the app on Linux and Windows keeps its data where each OS expects, finds sops, kiro-cli and its bundled CLI, opens the right shell named after itself, and shows OS-correct setup hints. | task 2 of 6 |
| 5 | The app builds, packages and starts on Linux and Windows | After this slice, `task desktop:build` builds the app on each OS (Linux with WebKitGTK 4.1, Windows with an NSIS setup carrying `bin\kirokit.exe`), the app has a smoke start, and the Desktop workflow proves the build starts on all three runners. | after 4 |
| 6 | One release ships every OS from three runners | After this slice, GoReleaser builds Windows CLI zips, Linux `.deb` and `.rpm` for the CLI and the app, a Scoop manifest and the Windows setup files, and the release workflow builds on macOS, Linux and Windows runners and pushes the tag only after all of them succeed. | after 5 |
| 7 | The docs tell every OS how to install and what to expect | After this slice, README and CONTRIBUTING describe installing and running kirokit and its app on macOS, Linux and Windows, including prerequisites and the unsigned-app warnings. | after 6 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- Slice 2 · "renormalize changes nothing" is checked with `git ls-files --eol` (no index entry is CRLF) instead of `git add --renormalize .`, which would stage files; T6 also adds `internal/testutil` (execute-bit and symlink skips shared by the CLI tests); nothing you approved changes.
- Slice 2 · `skills fetch` also refuses a symbolic link that git's index records (mode 120000), since a Windows checkout writes one as a plain file the old file-mode check cannot see; nothing you approved changes.
- Slice 2 · T8 also edits `internal/install/install_test.go` (accepts a junction, checks the `file://` path, covers reinstall and a real folder at the mount) and adds `internal/backup/backup_windows_test.go`; the knowledge-base `file://` source uses forward slashes like every other rendered path; nothing you approved changes.
- Slice 2 · The `main` ruleset requires a check named `build-vet-lint-test`, so the OS matrix runs as `test (<os>)` and a job keeping that name passes only when all three pass; nothing you approved changes.
- Slice 1 · T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task.
- The Plan · Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment.
- The Plan · T6 owns the shared internal/testutil helper; T8 owns its Windows-only junction tests for dirlink and backup.
- Slice 3 · The Windows terminal backend calls the pseudo-console API through golang.org/x/sys/windows (v0.46.0, already in the desktop module via Wails) instead of github.com/UserExistsError/conpty v0.1.4; the Session API and taskkill tree kill are unchanged.
<!-- /ql:decisions -->

## Checks

- T1 · a5c5dd3 · `go test ./internal/sopswrap/ && GOOS=windows go vet ./internal/sopswrap/` · exit 0 · 0.4s
- T2 · a5c5dd3 · `go test ./internal/render/` · exit 0 · 1s
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · not run by QualityLayer: changed after approval
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · not run by QualityLayer: changed after approval
- T4 · a5c5dd3 · `go test ./internal/install/ && GOOS=windows go vet ./internal/install/` · exit 0 · 1.1s
- T5 · a5c5dd3 · `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v` · exit 0 · 0.7s
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · exit 0 · 5.5s
- T6 · f73c655 · `go test ./... && GOOS=windows go vet ./...` · exit 0 · 7.3s
- T7 · f73c655 · `go test ./internal/externalskills/ ./internal/render/` · exit 0 · 7.2s
- T8 · f73c655 · `go test ./internal/dirlink/ ./internal/install/ ./internal/backup/ && GOOS=windows go vet ./internal/...` · exit 0 · 0.6s
- T9 · f73c655 · `go test ./cmd/kirokit/commands/ -run Setup` · exit 0 · 1.4s
- T10 · f73c655 · `actionlint && scripts/linux-docker.sh env KIROKIT_REQUIRE_SOPS=1 go test ./...` · exit 0 · 2.9s
- T11 · 3718d63 · `cd desktop && go test ./terminal/ && GOOS=windows GOARCH=amd64 go vet ./terminal/ && GOOS=windows GOARCH=arm64 go vet ./terminal/` · exit 0 · 3.5s
- T12 · 3718d63 · `cd desktop && go test ./... && GOOS=windows GOARCH=amd64 go build ./... && GOOS=windows GOARCH=arm64 go build ./... && GOOS=linux go vet ./...` · exit 0 · 4.1s
- T13 · b9daa54 · `cd desktop && go test -run 'Platform|Prereq|DataDir' ./...` · exit 0 · 0.6s
- T14 · b9daa54 · `cd desktop && go test -run 'Capture|Lookup|ToolFolders|DefaultShell|TerminalEnv' ./...` · exit 0 · 1s
- T15 · b9daa54 · `cd desktop && go test -run 'BundledCLI|ResolveCLI|Shell|State' ./...` · exit 0 · 0.8s
- T16 · b9daa54 · `cd desktop/frontend && node --test "src/**/*.test.js"` · exit 0 · 0.4s
- T17 · b9daa54 · `cd desktop && go test -run Settings ./...` · exit 0 · 0.6s
- T18 · b9daa54 · `task desktop:test && scripts/linux-docker.sh task desktop:test && cd desktop && GOOS=windows go vet ./...` · exit 0 · 14.4s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- T6 (tactical): "renormalize changes nothing" is checked with `git ls-files --eol` (no index entry is CRLF) instead of `git add --renormalize .`, which would stage files; T6 also adds `internal/testutil` (execute-bit and symlink skips shared by the CLI tests); nothing you approved changes
- T7 (tactical): `skills fetch` also refuses a symbolic link that git's index records (mode 120000), since a Windows checkout writes one as a plain file the old file-mode check cannot see; nothing you approved changes
- T8 (tactical): T8 also edits `internal/install/install_test.go` (accepts a junction, checks the `file://` path, covers reinstall and a real folder at the mount) and adds `internal/backup/backup_windows_test.go`; the knowledge-base `file://` source uses forward slashes like every other rendered path; nothing you approved changes
- T10 (tactical): the `main` ruleset requires a check named `build-vet-lint-test`, so the OS matrix runs as `test (<os>)` and a job keeping that name passes only when all three pass; nothing you approved changes

- T3 (amended): T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task
- Plan (amended): Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment
- T6, T8 (amended): T6 owns the shared internal/testutil helper; T8 owns its Windows-only junction tests for dirlink and backup
- T11 (amended): The Windows terminal backend calls the pseudo-console API through golang.org/x/sys/windows (v0.46.0, already in the desktop module via Wails) instead of github.com/UserExistsError/conpty v0.1.4; the Session API and taskkill tree kill are unchanged

## Plan amended

### T3 · amended

T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task

Amended by the agent on 2026-10-25; changed 04-outline.md.

### Plan · amended

Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment

Amended by the agent on 2026-10-25; changed 04-outline.md.

### T6, T8 · amended

T6 owns the shared internal/testutil helper; T8 owns its Windows-only junction tests for dirlink and backup

Amended by the agent on 2026-10-25; changed 04-outline.md.

### T11 · amended

The Windows terminal backend calls the pseudo-console API through golang.org/x/sys/windows (v0.46.0, already in the desktop module via Wails) instead of github.com/UserExistsError/conpty v0.1.4; the Session API and taskkill tree kill are unchanged

Why: reading conpty v0.1.4's source showed two defects for this app. It keeps the console's own
pipe ends open in our process, so after the process exits `Read` never returns `io.EOF`, and the
kirokit tab's `finishRun` and the shell goroutine would wait forever. Its `Close` also closes our
read end together with the console, so output still in the pipe is lost. With `x/sys/windows` the
console's ends are closed right after `CreatePseudoConsole`, and the console is closed when the
process exits, so `Read` drains the last output and then ends. The child is started with
`STARTF_USESTDHANDLES` and no handles, so it never inherits the app's own redirected handles. No
third-party module is added (`golang.org/x/sys` moves from indirect to direct in `desktop/go.mod`).
The Plan's "Decisions made for you" line naming the library is replaced by this one.

Amended by the agent on 2026-10-25; changed 04-outline.md.
