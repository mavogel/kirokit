# Ship CLI and desktop app for Linux and Windows · the build

## Summary

<!-- ql:summary -->
5 of 24 tasks · 1 of 7 slices

Building now: started 25 Oct, 19:12, 10 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | Windows MCP servers get their SOPS secrets | After this slice, `kirokit install` on Windows writes a `sops exec-env` launch that `cmd.exe` runs unchanged, and refuses an argument it cannot pass. | done |
| 2 | The CLI works on a Windows checkout, and CI tests it on three systems | After this slice, a Windows checkout keeps goldens, lock hashes and SOPS files byte-exact, the CLI's own tests run there, and CI runs build, vet and tests on Ubuntu, Windows and macOS. | next |
| 3 | The desktop terminal runs on Windows | After this slice, the desktop module compiles for Windows: both tabs run on a ConPTY pseudo console there, and env capture and child processes use Windows process handling. | after 2 |
| 4 | The app finds its folders, tools, CLI and shell on each OS | After this slice, the app on Linux and Windows keeps its data where each OS expects, finds sops, kiro-cli and its bundled CLI, opens the right shell named after itself, and shows OS-correct setup hints. | after 3 |
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
- Slice 1 · T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task.
- The Plan · Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment.
<!-- /ql:decisions -->

## Checks

- T1 · a5c5dd3 · `go test ./internal/sopswrap/ && GOOS=windows go vet ./internal/sopswrap/` · exit 0 · 0.4s
- T2 · a5c5dd3 · `go test ./internal/render/` · exit 0 · 1s
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · not run by QualityLayer: changed after approval
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · not run by QualityLayer: changed after approval
- T4 · a5c5dd3 · `go test ./internal/install/ && GOOS=windows go vet ./internal/install/` · exit 0 · 1.1s
- T5 · a5c5dd3 · `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v` · exit 0 · 0.7s
- T3 · a5c5dd3 · `go build -o kirokit ./cmd/kirokit && ./kirokit test platform-team && go test ./internal/golden/` · exit 0 · 5.5s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- T3 (amended): T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task
- Plan (amended): Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment

## Plan amended

### T3 · amended

T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task

Amended by the agent on 2026-10-25; changed 04-outline.md.

### Plan · amended

Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment

Amended by the agent on 2026-10-25; changed 04-outline.md.
