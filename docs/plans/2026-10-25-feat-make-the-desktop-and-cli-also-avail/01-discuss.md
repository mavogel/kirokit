---
answers: app
title: Ship CLI and desktop app for Linux and Windows
request: "feat: make the desktop and cli also available for linux and windows"
type: feature
depth: light
level: engineering
ui: true
overrides: []
stage: verify
iterations: 1
verification_failures: 1
gate: null
created: 1792949198136
startup_owner: a419c65251def52e
layout: seven
profile: full
status: approved
approved: true
review_digest: 4ce5ce0400d5ec2043fc8ac1cd5868caefbf2a682dfa3436d7a6936d41bdf3da
review_basis: affc22105c9b08df56a31d27f260573576e42c859ad3b526f0b33b7549966754
branch: feat/feat-make-desktop-cli-also
base: eaa4b319dc596336caf91ffb7fc8221bfd5ad765
base_ref: main
judge_runs: 1
---

# Ship CLI and desktop app for Linux and Windows

## Your request

> feat: make the desktop and cli also available for linux and windows


## Problem

kirokit and its desktop app can only be installed and used on a Mac. Team members on Linux or Windows
cannot get the app at all, and Windows users have no `kirokit` CLI either (Linux users only have a
bare download). Kiro CLI itself runs on macOS, Linux and Windows 11, so these people are left out of
the team setup that kirokit gives Mac users.

## Done means

1. **Linux users install the kirokit CLI with a package, Homebrew or download.** Each release
   carries `.deb` and `.rpm` packages and a tar.gz for amd64 and arm64, and `brew install`
   from `mavogel/tap` works on Linux; installed, `kirokit --version` prints the release version.
2. **Windows users install the kirokit CLI with Scoop or a zip and run it in PowerShell.**
   Each release carries a zip for amd64 and arm64, and `scoop install kirokit` from the mavogel
   bucket works; `kirokit --version` prints the release version.
3. **The CLI installs a team on Linux and Windows that Kiro CLI can run.** On Linux and Windows CI
   runners, `kirokit install <team>` writes the team into `$KIRO_HOME`, and an installed MCP server
   that reads SOPS secrets starts through `sops exec-env` and receives its secret.
4. **Linux users install the desktop app from a .deb or .rpm.** Each release carries both for
   amd64 and arm64; on a Linux CI runner the built app starts, and the desktop tests (terminals,
   shell tab, install and sync through the bundled CLI) pass.
5. **Windows users install the desktop app with a setup .exe.** Each release carries an installer
   for amd64 and arm64; on a Windows CI runner the built app starts, the desktop tests pass, and the
   shell tab opens PowerShell.
6. **The Mac release keeps working exactly as today.** The DMGs, `.app` and Homebrew casks are
   unchanged; Mac users notice nothing.
7. **Every release builds and checks all three platforms automatically.** The release on `main`
   publishes the Linux and Windows files next to the Mac ones, and pull requests run the CLI and
   desktop tests on Linux, Windows and macOS.
8. **The README tells Linux and Windows users how to install and what to expect.** Install
   commands per platform, prerequisites (SOPS, age key, Kiro CLI), and the unsigned-app warnings:
   Windows SmartScreen ("More info", then "Run anyway") as well as the existing macOS note.

## Scope

### In this task

- Release builds of the CLI and the desktop app for Linux and Windows, amd64 and arm64.
- The code changes that make the CLI and the app work on Windows and Linux (terminal, shell tab,
  data folders, prerequisite hints, SOPS-launched MCP servers, knowledge-base links).
- CI and release workflow changes; README and CONTRIBUTING updates.
- Publishing a Scoop manifest to the existing `mavogel/scoop-bucket` repository. You create a
  fine-grained token for that repository and add it as the `SCOOP_BUCKET_GITHUB_TOKEN` secret;
  the build records this as a step only you can do.
- On Windows, `kirokit install` writes the SOPS launch command quoted for `cmd.exe`; macOS and
  Linux output and the goldens stay as today.

### Not in this task

- Code signing or notarization on any platform.
- winget, AppImage, auto-update inside the app.
- Scheduled sync on Windows: `scripts/kirokit-sync-setup.sh` stays macOS and Linux only.
- Submission to curated stores (Microsoft Store, Flathub, Snap Store).
- Manual testing on real Linux or Windows machines; CI runners are the proof.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | How should Windows be supported? | Native Windows 11, CLI and app |
| 2 | How should Linux users install? | .deb and .rpm packages + Homebrew on Linux (CLI) + Plain tar.gz download |
| 3 | How should Windows users install? | Zip plus setup .exe on release + Scoop bucket |
| 4 | Sign the Windows builds? | Unsigned, README explains SmartScreen |
| 5 | Who checks it on real machines? | CI runners only |
<!-- /ql:decided -->

<!-- optional section here, only when there is content: "First release" -->

<!-- optional section here, only when there is content: "What the user sees" -->

<!-- optional section here, only when there is content: "Launch" -->

<!-- optional section here, only when there is content: "Not doing" -->

<!-- optional section here, only when there is content: "Worth it when" -->

<!-- optional section here, only when there is content: "Reference" -->

<!-- optional section here, only when there is content: "Added while building" -->

## Context pointers

- Request: "feat: make the desktop and cli also available for linux and windows" (no ticket given).
- Kiro CLI system requirements (macOS, Windows 11 PowerShell, Linux glibc 2.34+ or musl):
  https://kiro.dev/docs/cli/installation/

## First look at the code

- **CLI release**: `.goreleaser.yaml:15` builds the CLI for `darwin` and `linux` only (no
  `windows`); `.goreleaser.yaml:65` packs tar.gz archives; `.goreleaser.yaml:94` publishes two
  Homebrew casks; `.github/workflows/release.yaml:27` runs the whole release on one `macos-latest`
  job.
- **Desktop release**: `scripts/desktop-build.sh:27-40` builds a universal macOS app (`lipo`,
  `wails build -platform darwin/universal`, ad-hoc `codesign`); `scripts/desktop-prebuilt.sh` and
  `.goreleaser.yaml:30` stage it for GoReleaser Pro's prebuilt builder, `app_bundles` and `dmg`.
  `.github/workflows/desktop.yaml:23` tests the app on macOS only; `.github/workflows/ci.yaml:13`
  tests the CLI on Ubuntu only.
- **Desktop code with Unix-only assumptions**: `desktop/terminal/terminal.go:13,33,81`
  (`creack/pty`, Unix process-group kill); `desktop/envcapture.go:95-104` (login-shell capture,
  `/bin/zsh`, `Setpgid`); `desktop/shell.go:32-62` (bundled CLI at `../Resources/bin/kirokit`
  inside a `.app`), `desktop/shell.go:85-142` (zsh shell tab, ZDOTDIR wrapper);
  `desktop/app.go:54` (`~/Library/Application Support`); `desktop/prereq.go:26,55` (`brew install
  sops` hint, macOS age-key path); `desktop/frontend/src/tabs.js:5` (`kirokit` and `shell` tabs).
- **CLI code to check on Windows**: `internal/sopswrap/sopswrap.go:14-30` (MCP servers start via
  `sops exec-env`, quoting for `/bin/sh -c`); `internal/install/install.go:146-176` (knowledge base
  mounted with `os.Symlink`); `internal/install/install.go:324,434` and
  `internal/render/externalskills.go:67-72` (Unix file modes); `testdata/golden/**` compared byte
  for byte (line endings on a Windows checkout).
- **Docs**: `README.md:21,67,244,283` (Homebrew cask install, "macOS app" wording).
- **Similar patterns**: the existing GoReleaser Pro prebuilt flow (`.goreleaser.yaml:30-63`) is the
  model for importing Linux and Windows app builds made on their own runners.
- *Untraced areas*: how `kiro-cli` locates `KIRO_HOME` and launches MCP commands on Windows; how
  `sops exec-env` runs its command on Windows; whether Homebrew casks install on Linux; Wails v2
  Linux system libraries and Windows installer output; `desktop/runner.go` and `desktop/catalog.go`
  process handling on Windows.
