# Ship CLI and desktop app for Linux and Windows

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

## Accepted with open items

<!-- ql:accepted -->
The author accepted this build as it stood, without another judge run. What was left open:

- T10: Push this task's branch and open a pull request, so CI and Desktop run on the windows-latest, ubuntu-latest and macos-latest runners. The Windows runtime proof (the SOPS launch test through cmd.exe, the ConPTY terminal tests, the app smoke start) can only run there, and the build may not push.
- [MISSING] done 1: No Linux release exists to install from. environment: the branch is still not on GitHub, no CI or release run exists and no `GORELEASER_KEY` is set here, so no `.deb`, `.rpm`, tar.gz or Linux cask was produced; nothing in this round touches the release config — `evidence/it2-blocker-ci.txt`, `evidence/it1-static-cross-os.txt`
- [MISSING] done 2: No Windows release or Scoop manifest exists to install from. environment: same blocker; there is no Windows host or runner — `evidence/it2-blocker-ci.txt`
- [MISSING] done 3: The Windows half has still never run. environment: in the Linux container `kirokit install` plus the real `sops exec-env` launch passes again, as on macOS; no Windows runner has run it — `evidence/it2-scenario-1-linux.txt`, `evidence/it2-oracle-macos.txt`
- [MISSING] done 4: No Linux app package was built. environment: the desktop tests pass again in the Linux container; the `.deb`/`.rpm` need the Pro key's split snapshot, and amd64 needs the runner — `evidence/it2-scenario-3-linux-desktop-test.txt`, `evidence/it1-t20-linux-windows-build.txt`
- [MISSING] done 5: The Windows app has never started on Windows. environment: there is no Windows runner; the setup programs were only built in the container in iteration 1 — `evidence/it2-blocker-ci.txt`, `evidence/it1-t20-linux-windows-build.txt`
- [MISSING] done 7: No pull-request or release run exists on any runner. environment: the branch is not pushed — `evidence/it2-blocker-ci.txt`
- [MISSING] scenario 1: The CI steps still cannot run. environment: steps 1-2 need a CI run that does not exist; step 3 passes on macOS and step 4 passes in the Linux container, now including the symlinked-folder tests — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`, `evidence/it2-blocker-ci.txt`
- [MISSING] scenario 2: Step 7 still cannot be shown here. environment: the app always adds `/opt/homebrew/bin`, where sops lives on this Mac, so the hint never appears; the darwin hint string is unchanged. Steps 1-5 pass again: `platform-team: ok`, only the two devops golden lines change, the darwin assets are unchanged and the build script's darwin branch is only re-indented, the bundled CLI is universal (`x86_64 arm64`), and the smoke start exits 0. Step 6 is carried from iteration 1, since the fix never reaches the frontend. Probe past the script: the bundled CLI installs into a home with a symlinked agents folder (exit 0) — `evidence/it2-scenario-2-steps-1-5.txt`, `evidence/it1-scenario-2-steps-6-7.txt`
- [MISSING] scenario 3: The Desktop workflow has never run. environment: steps 1-5 need its log; step 6 (`task desktop:test` in the Linux container) passes again — `evidence/it2-scenario-3-linux-desktop-test.txt`, `evidence/it2-blocker-ci.txt`
- [MISSING] scenario 4: The release dry run cannot run. environment: it needs the pushed branch and the GoReleaser Pro key — `evidence/it2-blocker-ci.txt`
- [MISSING] dod T1: The cmd.exe round trip has still never run on Windows. environment: the join, refusal, POSIX and spaced-secrets-path tests pass here; `sopswrap_windows_test.go` only compiles — `evidence/it2-cards.txt`
- [MISSING] dod T4: The Windows bullets have still never run. environment: on macOS the installed files match base byte for byte (now also with a symlinked agents folder); the `C:/` and refuse-before-writing tests skip off Windows — `evidence/it2-truth-macos-install.txt`, `evidence/it2-cards.txt`
- [MISSING] dod T5: The Oracle test has still never run on Windows. environment: it passes on macOS and in the Linux container with sops required — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`
- [MISSING] dod T8: The Windows junction bullets have still never run. environment: the macOS and Linux bullets now hold (a symlinked folder is backed up as the link and installs as on base, also through the bundled CLI); the junction tests only compile here — `evidence/it2-review-tests.txt`, `evidence/it2-truth-macos-install.txt`, `evidence/it2-cards.txt`
- [MISSING] dod T11: Every Windows terminal bullet is a Windows-only test that has still never run. environment: the terminal tests and the windows amd64/arm64 vet pass here — `evidence/it2-cards.txt`
- [MISSING] truth oracle: The oracle has still never run on its Windows target. environment: on macOS and in the Linux container, with sops required, the installed server gets every argument and `Bearer s3cret value` — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- ConPTY library: `github.com/UserExistsError/conpty` v0.1.4 — small, Windows-only, maps one-to-one onto `Session`; it compiled into the app for windows/amd64 and windows/arm64 in the spike. (Plan)
- creack/pty stays for macOS and Linux — the Mac behaviour must not change; only a build tag is added. (Plan)
- Windows shell: `pwsh.exe` when on `PATH`, else `powershell.exe`, started with `-NoLogo` — PowerShell is where Kiro CLI runs on Windows. (Plan)
- Windows environment: the app's own environment, no login-shell capture — Windows apps inherit the user's full environment; there is no login shell to read. (Plan)
- Linux login shell: `$SHELL`, else `/bin/bash`, else `/bin/sh`; macOS keeps `/bin/zsh` — Linux desktops do not default to zsh. (Plan)
- Tool folders per OS — macOS keeps the Homebrew folders; Linux adds `/home/linuxbrew/.linuxbrew/bin`, `/usr/local/bin`, `~/.local/bin`; Windows adds `%USERPROFILE%\scoop\shims`. (Plan)
- Rendered paths use forward slashes on every OS — `install` renders with `filepath.ToSlash(kiroHome)`, so Windows configs read `C:/Users/…/.kiro/…`, which Windows, sops and `file://` resources accept; macOS and Linux paths do not change. (Plan)
- Knowledge-base link on Windows — `os.Symlink` first; when Windows refuses it for lack of privilege, a directory junction (`mklink /J`), which uninstall and reinstall recognise as kirokit's own. (Plan)
- Child processes on Windows open no console window — `git` and other helpers get `CREATE_NO_WINDOW`. (Plan)
- App arm64 builds — Windows arm64 cross-compiles on the Windows runner (no cgo there); Linux arm64 builds on GitHub's `ubuntu-24.04-arm` runner, free for this public repository. (Plan)
- WebKitGTK 4.1 (`-tags webkit2_41`) — the ABI of Ubuntu 22.04+, Debian 12+ and Fedora 40+, which also covers Kiro CLI's glibc 2.34+ floor. (Plan)
- Scoop publishes the CLI only — the app ships as the setup `.exe`; the Scoop step is skipped when `SCOOP_BUCKET_GITHUB_TOKEN` is empty, so a missing secret cannot fail a release after tagging. (Plan)
- The Linux CLI cask needs no new config — GoReleaser already generates its `on_linux` block; the README's `brew install --cask mavogel/tap/kirokit` line is documented for Linux too. (Plan)
- Settings paths accept Windows drive paths — the root check uses `filepath.VolumeName`; repository addresses accept `C:\…` and `C:/…`. (Plan)
- "renormalize changes nothing" is checked with `git ls-files --eol` (no index entry is CRLF) instead of `git add --renormalize .`, which would stage files; T6 also adds `internal/testutil` (execute-bit and symlink skips shared by the CLI tests); nothing you approved changes. (slice 2)
- `withToolFoldersFor` keeps the person's own PATH entries as they are (no de-duplication), exactly as before on macOS; nothing you approved changes. (slice 4)
- The non-zsh shell test checks `$KIROKIT_BIN` instead of `command -v kirokit`, because Debian's `/etc/profile` resets PATH for a login `sh` (the case the Settings note already describes); nothing you approved changes. (slice 4)
- The Windows app is built as `kirokit-desktop-<arch>.exe` and installed as `kirokit-desktop.exe`; the installers carry Wails' default product version 0.0.0 in their Windows "Apps" entry (the file names carry the release version); nothing you approved changes. (slice 5)
- The smoke step relies on the app's own 60-second limit plus a 5-minute step limit, since macOS runners have no `timeout`; nothing you approved changes. (slice 5)
- The Linux app has one prebuilt build per architecture, and `KIROKIT_SKIP_LINUX_ARM64_APP` leaves arm64 out of pull-request snapshots (GoReleaser filters architectures only when splitting by target, and the Plan splits by OS); nothing you approved changes. (slice 6)
- The dry run merges the snapshot parts too; their recorded context (`Snapshot: true`, `Skips: publish, announce`) keeps the merge from publishing, and the tap and bucket get placeholder tokens; the artifact actions are pinned at upload-artifact v7.0.2 and download-artifact v8.0.2; nothing you approved changes. (slice 6)
- `skills fetch` also refuses a symbolic link that git's index records (mode 120000), since a Windows checkout writes one as a plain file the old file-mode check cannot see; nothing you approved changes. (slice 2)
- T8 also edits `internal/install/install_test.go` (accepts a junction, checks the `file://` path, covers reinstall and a real folder at the mount) and adds `internal/backup/backup_windows_test.go`; the knowledge-base `file://` source uses forward slashes like every other rendered path; nothing you approved changes. (slice 2)
- The `main` ruleset requires a check named `build-vet-lint-test`, so the OS matrix runs as `test (<os>)` and a job keeping that name passes only when all three pass; nothing you approved changes. (slice 2)
- T3 also regenerates the installed-with-kb golden (same one-line header change) and adds internal/golden/windows_launch_test.go, which renders every team in teams/ with the cmd.exe launch shell; its check runs './kirokit test platform-team' because data-team's context.json golden is already stale on main (Renovate bump dabb3eb), unrelated to this task. (slice 1)
- Slice 1 validation and Scenario 2 step 1 run './kirokit test platform-team', and Scenario 2 step 2 expects the two platform-team devops goldens, for the same reason as T3's amendment. (the Plan, while building)
- T6 owns the shared internal/testutil helper; T8 owns its Windows-only junction tests for dirlink and backup. (the Plan, while building)
- The Windows terminal backend calls the pseudo-console API through golang.org/x/sys/windows (v0.46.0, already in the desktop module via Wails) instead of github.com/UserExistsError/conpty v0.1.4; the Session API and taskkill tree kill are unchanged. (slice 3)
- T9 also adds 'kirokit --version' (cobra's version flag, printing 'kirokit <version>' like the first line of 'kirokit version'), because Done means 1 and 2 check installs with 'kirokit --version', which was an unknown flag. (slice 2)
- T20 keeps Wails' regenerated wails_tools.nsh and the downloaded WebView2 bootstrapper out of git (desktop/.gitignore) and tracks only project.nsi, the icon, info.json and the manifest; Taskfile runs the build script through bash so Windows (Git Bash) runs it. (slice 5)
- T20 also excludes *.ico and *.icns from lefthook's whitespace fixers, which corrupted the Windows icon on the first commit attempt. (slice 5)
- T19's check runs the smoke tests in a subshell, so the smoke start's desktop/build path is relative to the repository root. (slice 5)
- The scenarios' gh commands name this checkout's branch, feat/feat-make-desktop-cli-also, instead of the placeholder feat/linux-windows. (the Plan, while building)
- Verification round 1 is kept although 2 files changed during it, because the only changes were the reviewer's own new tests (`internal/sopswrap/sopswrap_test.go`, `internal/externalskills/fetch_test.go`), which add checks and change no code its verdicts judged; nothing you approved changes. (the Plan, while building)
- Verification round 2 was recorded while 2 files changed, 2 of them code the run exercised; rerun it or keep it and say why here; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
- Still awaiting your answer: The release dry-run input still does nothing. Carried from iteration 1: `release.yaml` declares `dry_run` for `workflow_dispatch` but never reads it; this round does not touch it.
- Still awaiting your answer: A symlinked folder is backed up only as a link, so the backup keeps no copy of the old files. Install then writes through the link into the folder it points at, and the backup link points at the overwritten files. `main` did exactly the same, so Mac users notice nothing, but the backup protects nothing there.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Not shipped yet. Branch `feat/feat-make-desktop-cli-also`.
<!-- /ql:ship -->
