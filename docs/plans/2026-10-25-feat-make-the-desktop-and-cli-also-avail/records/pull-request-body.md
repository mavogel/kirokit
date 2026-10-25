<!-- ql:links -->
Task: `docs/plans/2026-10-25-feat-make-the-desktop-and-cli-also-avail/`
<!-- /ql:links -->

## Why the change

<!-- fill: exactly one sentence: the problem it solves and what becomes possible once it ships -->

## What a reviewer must know

<!-- fill: 1 to 3 bullets (a migration, compatibility, a deliberate omission, a surprising decision), or "None." -->

## What changed

<!-- fill: the change outline in structural views, never a file-by-file list: only the views that explain this change, each after one line that says why it matters -->

## Screens compared

<!-- ql:screens -->
- **windows-linux-app#dock** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
- **windows-linux-app#setup** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
7 slices · 24 tasks · 17 commits · 88 files · +3503 −344

**Slice 1 · Windows MCP servers get their SOPS secrets** · T1, T2, T3, T4, T5 · 17 files · +697 −41

- `.gitleaks.toml` · T5 · new · +9 −0
- `cmd/kirokit/commands/install_sops_test.go` · T5 · new · +181 −0
- `internal/golden/windows_launch_test.go` · T3 · new · +33 −0
- `internal/install/install.go` · T4, T8 · +24 −6
- `internal/install/install_test.go` · T4, T6 · +104 −16
- `internal/render/context.go` · T2 · +4 −0
- `internal/render/funcs.go` · T2 · +15 −9
- `internal/render/render.go` · T2 · +32 −2
- `internal/render/render_test.go` · T2 · +74 −0
- `internal/sopswrap/sopswrap.go` · T1 · +59 −5
- `internal/sopswrap/sopswrap_test.go` · T1 · +83 −0
- `internal/sopswrap/sopswrap_windows_test.go` · T1 · new · +54 −0
- `teams/platform-team/agents/devops/agent.json.tmpl` · T3 · +1 −1
- `testdata/golden/platform-team/installed-with-kb/agents/devops.json` · T3 · +1 −1
- `testdata/golden/platform-team/rendered/agents/devops.json` · T3 · +1 −1
- `testdata/sops/age-test-key.txt` · T5 · new · +6 −0
- `testdata/sops/secrets.enc.yaml` · T5 · new · +16 −0

**Slice 2 · The CLI works on a Windows checkout, and CI tests it on three systems** · T6, T7, T8, T9, T10 · 25 files · +732 −49

- `.gitattributes` · T6 · new · +19 −0
- `.github/workflows/ci.yaml` · T10 · +50 −2
- `cmd/kirokit/commands/root.go` · T9 · +3 −0
- `cmd/kirokit/commands/setup.go` · T9 · +4 −3
- `cmd/kirokit/commands/setup_test.go` · T9 · +17 −0
- `cmd/kirokit/commands/skills_test.go` · T6 · +3 −2
- `cmd/kirokit/commands/sync_test.go` · T6 · +1 −1
- `cmd/kirokit/commands/version_test.go` · T9 · +18 −0
- `internal/backup/backup.go` · T8 · +18 −5
- `internal/backup/backup_test.go` · T6 · +59 −8
- `internal/backup/backup_windows_test.go` · T8 · new · +39 −0
- `internal/dirlink/dirlink.go` · T8 · new · +29 −0
- `internal/dirlink/dirlink_other.go` · T8 · new · +12 −0
- `internal/dirlink/dirlink_test.go` · T8 · new · +59 −0
- `internal/dirlink/dirlink_windows.go` · T8 · new · +50 −0
- `internal/dirlink/dirlink_windows_test.go` · T8 · new · +45 −0
- `internal/externalskills/fetch.go` · T7 · +56 −7
- `internal/externalskills/fetch_test.go` · T6, T7 · +58 −12
- `internal/externalskills/lock.go` · T7 · +20 −1
- `internal/externalskills/lock_test.go` · T6, T7 · +56 −6
- `internal/externalskills/verify.go` · T7 · +17 −1
- `internal/render/externalskills_test.go` · T6 · +2 −1
- `internal/testutil/testutil.go` · T6 · new · +33 −0
- `scripts/linux-docker.Dockerfile` · T10 · new · +29 −0
- `scripts/linux-docker.sh` · T10 · new · +35 −0

**Slice 3 · The desktop terminal runs on Windows** · T11, T12 · 9 files · +480 −44

- `desktop/catalog.go` · T12 · +1 −0
- `desktop/envcapture.go` · T12, T14 · +19 −43
- `desktop/go.mod` · T11 · +1 −1
- `desktop/procgroup_unix.go` · T12 · new · +18 −0
- `desktop/procgroup_windows.go` · T12 · new · +35 −0
- `desktop/terminal/terminal_test.go` · T11 · +14 −0
- `desktop/terminal/terminal_unix.go` (renamed from `desktop/terminal/terminal.go`) · T11 · +2 −0
- `desktop/terminal/terminal_windows.go` · T11 · new · +236 −0
- `desktop/terminal/terminal_windows_test.go` · T11 · new · +154 −0

**Slice 4 · The app finds its folders, tools, CLI and shell on each OS** · T13, T14, T15, T16, T17, T18 · 17 files · +748 −92

- `desktop/app.go` · T13, T15 · +13 −5
- `desktop/clone_test.go` · T18 · +8 −4
- `desktop/envcapture_test.go` · T14, T18 · +8 −3
- `desktop/frontend/src/main.js` · T16 · +4 −2
- `desktop/frontend/src/tabs.js` · T16 · +4 −0
- `desktop/frontend/src/tabs.test.js` · T16 · +9 −1
- `desktop/frontend/src/terminal.js` · T16 · +13 −2
- `desktop/platform.go` · T13, T14, T15 · new · +244 −0
- `desktop/platform_test.go` · T13, T14, T15 · new · +164 −0
- `desktop/prereq.go` · T13 · +4 −15
- `desktop/prereq_test.go` · T13, T18 · +14 −6
- `desktop/runner_test.go` · T18 · +13 −10
- `desktop/settings.go` · T17 · +18 −5
- `desktop/settings_test.go` · T17, T18 · +25 −0
- `desktop/shell.go` · T14, T15 · +25 −26
- `desktop/shell_test.go` · T15, T18 · +128 −13
- `desktop/testtools_test.go` · T18 · new · +54 −0

**Slice 5 · The app builds, packages and starts on Linux and Windows** · T19, T20, T21 · 14 files · +391 −32

- `.github/workflows/desktop.yaml` · T21, T23 · +63 −7
- `Taskfile.yml` · T20 · +3 −2
- `desktop/.gitignore` · T20 · +4 −0
- `desktop/build/linux/kirokit-desktop.desktop` · T20 · new · +9 −0
- `desktop/build/linux/kirokit-desktop.png` · T20 · new · +0 −0
- `desktop/build/windows/icon.ico` · T20 · new · +0 −0
- `desktop/build/windows/info.json` · T20 · new · +15 −0
- `desktop/build/windows/installer/project.nsi` · T20 · new · +132 −0
- `desktop/build/windows/wails.exe.manifest` · T20 · new · +15 −0
- `desktop/main.go` · T19 · +12 −3
- `desktop/smoke.go` · T19 · new · +31 −0
- `desktop/smoke_test.go` · T19 · new · +36 −0
- `lefthook.yml` · T20 · +6 −2
- `scripts/desktop-build.sh` · T20 · +65 −18

**Slice 6 · One release ships every OS from three runners** · T22, T23 · 3 files · +396 −66

- `.github/workflows/release.yaml` · T23 · +231 −46
- `.goreleaser.yaml` · T22 · +110 −0
- `scripts/desktop-prebuilt.sh` · T22 · +55 −20

**Slice 7 · The docs tell every OS how to install and what to expect** · T24 · 3 files · +59 −20

- `AGENTS.md` · T24 · +4 −1
- `CONTRIBUTING.md` · T24 · +7 −2
- `README.md` · T24 · +48 −17
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. Linux users install the kirokit CLI with a package, Homebrew or download.** Scenario 4: One release carries every OS's files, after slice 6 · no proof found · `evidence/it2-blocker-ci.txt`
- **2. Windows users install the kirokit CLI with Scoop or a zip and run it in PowerShell.** Scenario 4: One release carries every OS's files, after slice 6 · no proof found · `evidence/it2-blocker-ci.txt`
- **3. The CLI installs a team on Linux and Windows that Kiro CLI can run.** Scenario 1: A Windows install starts its SOPS-protected MCP server, after slice 2 · no proof found · `evidence/it2-scenario-1-linux.txt`
- **4. Linux users install the desktop app from a .deb or .rpm.** Scenario 3: The app tests pass and the built app starts on Linux and Windows, after slice 5 · no proof found · `evidence/it2-scenario-3-linux-desktop-test.txt`
- **5. Windows users install the desktop app with a setup .exe.** Scenario 3: The app tests pass and the built app starts on Linux and Windows, after slice 5 · no proof found · `evidence/it2-blocker-ci.txt`
- **6. The Mac release keeps working exactly as today.** Scenario 2: macOS users notice nothing, after slice 6 · passed · `evidence/it2-truth-macos-install.txt`
- **7. Every release builds and checks all three platforms automatically.** Scenario 4: One release carries every OS's files, after slice 6 · no proof found · `evidence/it2-blocker-ci.txt`
- **8. The README tells Linux and Windows users how to install and what to expect.** Scenario 5: The README guides Linux and Windows users, after slice 7 · passed · `evidence/it1-scenario-5.txt`
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No signing or notarization.** The README explains SmartScreen and the existing macOS quarantine note.
- **No winget, AppImage, Flathub, Snap or Microsoft Store.**
- **No scheduled sync on Windows.** `scripts/kirokit-sync-setup.sh` stays macOS and Linux only.
- **No auto-update inside the app.**
- **No manual testing on real machines.** CI runners are the proof, as you chose.
- **No change to the showcase recording.** `scripts/showcase/` stays Mac-only.
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
No acceptance recorded; 17 items remain unverified for Review. Lanes that ran: Security review, Project checks, Code review and Fix rounds.
<!-- /ql:checked -->

## Try it yourself

<!-- fill: ### Try it (up to three steps, each a #### block with Do, Run, Expect and Saw lines), ### Not verified and ### Docs; the App shows it in the Approve view -->

<!-- optional section here, only when there is content: "Security" -->
