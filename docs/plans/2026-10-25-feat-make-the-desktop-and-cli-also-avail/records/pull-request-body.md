<!-- ql:links -->
Task: `docs/plans/2026-10-25-feat-make-the-desktop-and-cli-also-avail/`
<!-- /ql:links -->

## Why the change

kirokit and its desktop app only shipped for macOS (the CLI also as a bare Linux tar.gz); this ships both for Linux and Windows on amd64 and arm64, so Linux and Windows team members can install their team's Kiro setup, including MCP servers that read SOPS secrets.

## What a reviewer must know

- **Nothing here has run on Windows or a GitHub runner yet.** The branch is not pushed; the `cmd.exe` launch, the ConPTY terminal, the Windows app start and the split release (which needs `GORELEASER_KEY`) are proved only by the first CI and Desktop runs and a release dry run (`gh workflow run release.yaml --ref <branch> -f dry_run=true`). macOS was run here; Linux in a container.
- **Windows refuses some MCP arguments by design.** sops runs the command through `cmd.exe` there, so Windows installs reject arguments with whitespace, `"` or `& | < > ^ % ! ( )`. platform-team's header became `Authorization:${ATLASSIAN_AUTH}` (one golden line, in two goldens); macOS/Linux output is otherwise byte-identical. The release now builds on three runners and pushes the tag only after all of them succeed; CI keeps a job named `build-vet-lint-test` because the `main` ruleset requires it.
- **Follow-ups from the reviewer:** the release's `dry_run` input is declared but never read (as before: every manual run is a dry run); a symlinked Kiro-home folder is backed up as a link, so the backup points at files the install then overwrites, as on `main`.

## What changed

How a secret-reading MCP server is launched now depends on the OS that installs it; `render` and the goldens always use the POSIX form:

```diff
 kirokit install <team>
-  render.RenderTeam(…, kiroHome)                         POSIX quoting, native paths
+  render.RenderTeamWith(…, Options{
+      KiroHome:    filepath.ToSlash(kiroHome),            C:/Users/…/.kiro on Windows
+      LaunchShell: Cmd on windows, else POSIX })
     sopsArgsJSON -> sopswrap.BuildWrapperFor(shell, file, cmd, args)
       POSIX: "sops exec-env <file> 'npx … Authorization:${X}'"   (sh -c, as before)
+      Cmd:   "sops exec-env <file> npx … Authorization:${X}"     (cmd.exe /C; unsafe token -> render: <team>/<agent>: argument "…" cannot pass …)
```

The desktop app keeps one OS-independent core; what differs per OS lives in build-tagged pairs or in pure functions of the OS name, so every branch is tested on any machine:

```text
desktop/
├── terminal/terminal_unix.go      creack/pty (today's file, unchanged)
├── terminal/terminal_windows.go   pseudo console via golang.org/x/sys/windows; Close = taskkill /T
├── procgroup_{unix,windows}.go    kill the process tree on timeout; no console windows on Windows
├── platform.go                    per OS: data folder, age-key path, hints, tool folders, PATH lookup,
│                                  bundled-CLI location, shell tab (zsh/bash or PowerShell) and its name
└── smoke.go                       KIROKIT_DESKTOP_SMOKE=1: quit once the page loads (CI start check)
internal/dirlink/                  symlink, or a junction on Windows; install, uninstall and backup use it
internal/externalskills/lock.go    + executables: […]   git's exec bits, so Windows can verify the hash
```

The release becomes a split build, one job per OS, merged at the end:

```text
version (svu) ─┬─ app-linux-arm64 (ubuntu-24.04-arm) ──┐
               └─ split: macOS │ Linux │ Windows  ─────┴─ merge: push tag, goreleaser continue --merge
Linux:   kirokit_<v>_linux_<arch>.{deb,rpm,tar.gz} · kirokit-desktop_<v>_linux_<arch>.{deb,rpm} · cask on_linux
Windows: kirokit_<v>_windows_<arch>.zip · Scoop manifest (mavogel/scoop-bucket) · kirokit-desktop_<v>_windows_<arch>-setup.exe
macOS:   unchanged (tar.gz, DMGs, both casks)
```

## Screens compared

<!-- ql:screens -->
- **windows-linux-app#dock** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
- **windows-linux-app#setup** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
7 slices · 24 tasks · 19 commits · 88 files · +3503 −344

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
Taken as it stood with 17 items still open; lanes that ran: Security review, Project checks, Code review and Fix rounds.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. An installed MCP server gets its arguments and decrypted secret unchanged
- Do: run the SOPS launch test, which installs a fixture team and starts its server through the real `sops exec-env`.
- Run: `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v`
- Expect: `--- PASS: TestInstall_SopsLaunchReachesTheServerUnchanged`
- Saw: passes on macOS and in the Linux container — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`

#### 2. macOS installs exactly as before, including through a symlinked agents folder
- Do: install platform-team into a temporary Kiro home whose `agents` folder is a symlink.
- Run: `go build -o kirokit ./cmd/kirokit && T=$(mktemp -d) && mkdir -p $T/real $T/kiro && ln -s $T/real $T/kiro/agents && KIRO_HOME=$T/kiro ./kirokit install platform-team`
- Expect: `installed platform-team into …/kiro`, and the installed files match `main`'s
- Saw: byte-identical to `main` for a fresh home, an existing folder and a symlinked folder — `evidence/it2-truth-macos-install.txt`

#### 3. The app on macOS still names its shell tab zsh and starts
- Do: build the app and smoke-start it, then open it and look at the dock.
- Run: `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?`
- Expect: `0`; in the running app the dock's second tab reads **zsh**
- Saw: exit 0 and the same bundle layout — `evidence/it2-scenario-2-steps-1-5.txt`; the dock's tab reads zsh — `evidence/it1-scenario-2-step-6-dock-shell.png`

### Not verified

- Anything on Windows (the `cmd.exe` launch, the ConPTY terminal, the PowerShell tab, the app start, the setup `.exe` installing) and the CI, Desktop and Release workflows on GitHub's runners: the branch is not pushed, and the build may not push — open item T10.
- The `.deb`/`.rpm`/zip/Scoop/cask release files as GoReleaser's split and merge produce them in CI: needs `GORELEASER_KEY`, which exists only in GitHub. Locally, Linux split snapshots ran and the packages installed on a clean Ubuntu and started; the merge needs the key.
- `brew install --cask mavogel/tap/kirokit` on Linux and `scoop install kirokit` after the first real release.
- The macOS "brew install sops" hint (scenario 2 step 7): sops is installed on this Mac, so the hint never shows; its text is unchanged.

### Docs

- README: install per OS (CLI and app), prerequisites and the age-key path per OS, SmartScreen steps beside the macOS quarantine note, the Windows argument rule, junctions, `executables` in the lock, drive-path sources, the shell tab's name, scheduled sync macOS/Linux only.
- CONTRIBUTING: three-OS CI, Linux/Windows build prerequisites, `scripts/linux-docker.sh`. AGENTS.md: the per-OS layout and rules.

## Security

- Low · `internal/sopswrap/sopswrap.go`: the `cmd.exe` launch refuses whitespace, quotes and `& | < > ^ % ! ( )`, so template values cannot inject commands; secrets stay in the encrypted file and are referenced only as `${VAR}`. No action.
