---
title: Desktop app from design
request: now build the desktop app based on the design
type: feature
depth: light
level: engineering
ui: true
overrides: []
stage: review
iterations: 3
verification_failures: 3
gate: null
created: 1792271603793
startup_owner: 4d42f5beaea5d6a1
layout: seven
status: approved
approved: true
review_digest: 4e4d5fb372146939e83bc577f2474c4d35e77775ef98cb856252bfaa8648816d
review_basis: 5bc0fc6289a8b2fa0d4a7d7f3ed845bf37dd1b46ad3df1ec0b75e9206b74708e
branch: ql/desktop-app-from-design
base: 440b841b0183803905935d03e91528a2fe90a2ea
base_ref: main
checkpoint_failures: 0
judge_runs: 3
---

# Desktop app from design

## Your request

> now build the desktop app based on the design


## Problem

**Non-engineers still cannot set up kirokit without a terminal.** The approved design shows a desktop app that fixes this, but it is only a clickable mockup. Nothing a person can install and run exists yet.

**The design's data and terminal are simulated.** Teams, agents and versions were copied into the page by hand, and its terminal only pretends to run commands. The real app must read the repository and run real kirokit.

## Done means

1. **The kirokit desktop app opens a window that matches the design.** It shows the logo, the sidebar with all five sections and the main pane, in light and dark mode. (only you can confirm: it looks like the approved design)
2. **The sidebar collapses to an icon rail and expands again.** The toggle and Command-B both switch it, and the main pane widens while the rail shows.
3. **Teams, Agents, and Skills and MCP versions show the repository's real data.** They are read from the app's own copy of the repository, so a new team or pin appears after an update without an app change.
4. **A non-engineer installs a team with buttons only.** They pick a team, pass a real check for SOPS, the decryption key and Kiro CLI, press Install, and see the team marked installed.
5. **Install writes the same files as `kirokit install`.** The Kiro home folder afterwards holds exactly what the command line would have written.
6. **Install and sync shows the real install state.** Each installed team has working Sync and Uninstall actions, read from kirokit's install records.
7. **A real terminal docks at the bottom, as in VS Code.** It runs the person's own shell, and opens, closes and resizes like the design.
8. **Each button press appears in the terminal as its `kirokit` command.** The command line and its real output show there, for example `installed platform-team into /Users/you/.kiro`.
9. **A command typed in the terminal updates the screens.** A typed `kirokit install data-team` marks data-team installed on both pages.
10. **Settings sets the Kiro home folder and the repository's address.** Both are remembered when the app is closed and opened again.
11. **Only one team is ever installed at a time, in the app and the CLI.** Install on a second team asks to replace the installed one in the app, then uninstalls it first. A typed `kirokit install` of another team uninstalls the installed one first and prints both lines.
12. **One command builds the macOS app, and CI builds it too.** A broken app build fails CI like a broken CLI build does.
13. **The kirokit command line keeps working, apart from the one-team rule.** The Go tests and golden comparisons still pass, and the README describes the new install behaviour.
14. **The app fetches the templates itself, with no git knowledge needed.** On first launch it asks for the repository's address and clones it into its own folder. An Update action brings in the newest release tag, and its command shows in the terminal.
15. **Every kirokit release also ships the macOS app.** The GitHub release carries the app for Apple silicon and Intel, next to the CLI archives. (only you can confirm: a real tagged release publishes it)
16. **A separate `kirokit-desktop` Homebrew cask installs the app, and it opens.** `brew install --cask kirokit-desktop` puts it in Applications, the CLI cask stays as it is, and it opens without macOS calling it damaged, although it is unsigned. (only you can confirm: the cask lands in your tap)

## Scope

### In this task

- **The five pages, the terminal and the setup flow from the design**
- **The one-team rule in the CLI**
- **A macOS app in GitHub releases and Homebrew**

### Not in this task

- **Editing templates or manifests in the app**
- **Signed or notarized releases**
- **Linux or Windows apps**

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | Where does the app get the teams' templates from? | The app clones the repository itself |
| 2 | Should the app keep the design's one-installed-team rule? | Yes, also enforced in the CLI |
| 3 | How far should this task take building and shipping the app? | also attach to GH releases and a homebrew cask |
| 4 | Which operating systems must the first app run on? | macOS only |
<!-- /ql:decided -->

## Reference

**The approved design is the product definition.** It is the clickable `sidebar-layout` mockup from the task "Desktop app sidebar layout", with its Discuss decisions and Plan contracts. Earlier answers there still hold: the app lives in this repository, uses Wails, and gives engineers a full terminal.

## Context pointers

- `docs/plans/2026-10-17-desktop-app-sidebar-layout/design/sidebar-layout.html` — the approved design page (474 lines, inline styles and script)
- `docs/plans/2026-10-17-desktop-app-sidebar-layout/01-discuss.md` — its decisions: Wails, in this repository, collapsible icon rail, full VS Code-style terminal, prerequisites step, one installed team (replace)
- `docs/plans/2026-10-17-desktop-app-sidebar-layout/03-plan.md` — its contracts: terminal output lines, install state map, setup dialog
- `.qualitylayer/designs/desktop-app/` — the project's empty `desktop-app` design stub
- `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md` — v1 PRD, which ruled out a web UI
- `assets/logo.svg` — the project's only visual asset

## First look at the code

There is no desktop app code yet; this task creates it. `kirokit` is a pure-Go CLI with no UI dependency, and its domain logic sits in `internal/` packages a Go app in the same module can call directly.

| What | Where | Why it matters |
|---|---|---|
| CLI entry point | `cmd/kirokit/main.go:12` | The only binary the module builds today |
| Repository roots | `cmd/kirokit/commands/root.go:14-19` | `teams`, `global/steerings`, `env/mcp-versions.yaml` are relative to the working directory; a Finder-launched app has another one |
| Team discovery | `cmd/kirokit/commands/root.go:47` (`teamNames`) | Lists directories under `teams/`; unexported, so the app needs its own or a shared helper |
| Install | `internal/install/install.go:62` (`Install`), `cmd/kirokit/commands/install.go:37` | Takes explicit `Roots` and `kiroHome`, so it is reusable in-process; the CLI prints `installed <team> into <kiroHome>` |
| Kiro home | `internal/install/install.go:46` (`ResolveKiroHome`) | `$KIRO_HOME`, else `~/.kiro` |
| Install records | `internal/install/install.go:260` (`installManifestPath`) | `<kiroHome>/.kirokit-installed/<team>.json`; the install state the app can read and watch |
| Uninstall and sync | `internal/install/install.go:232`, `internal/install/sync.go:16` | Same explicit-arguments shape as `Install` |
| Build and release | `Taskfile.yml`, `.goreleaser.yaml:9-20`, `.github/workflows/` | `CGO_ENABLED=0`, darwin and linux CLI archives only; a webview app needs CGO on macOS |
| Toolchain | local machine | Go 1.27.2, Node 26, pnpm; no `wails` CLI installed |

*Untraced areas:* which Wails major version fits Go 1.27, how a real shell is streamed into the webview, what PATH a Finder-launched app sees for `sops` and `kiro-cli`, and how CI builds a CGO macOS app.
