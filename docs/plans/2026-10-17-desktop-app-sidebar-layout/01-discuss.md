---
title: Desktop app sidebar layout
request: create sidebar layout for desktop app
type: feature
depth: product
level: engineering
ui: true
overrides: []
stage: shipped
iterations: 1
verification_failures: 0
gate: null
created: 1792265605679
startup_owner: 773705b5c336001d
layout: seven
status: verified
approved: true
review_digest: d90f60444b2bec7d376b250bc7f370ce4136bde51cfbd0cbf03ab60783f8339f
review_basis: 3da1a020dda2ace7cc0f935a971ddf6e0602d0beb7bac8d73eaac5b47bed79bf
branch: ql/desktop-app-sidebar-layout
base: 9a919bfb8f018ce028906f5ab1e6a2677bf588b3
base_ref: main
judge_runs: 0
shipped: 2026-10-17
commits: ["1382052","08a59eb","b2e48fd","f183c46"]
touched: [".cz.toml"]
slices: 1
---

# Desktop app sidebar layout

## Your request

> create sidebar layout for desktop app


## Problem

**Non-engineers cannot set up kirokit today.** Every action, from picking a team to installing it, needs a terminal and the repository checked out.

**Nothing defines a desktop app for kirokit yet.** The first product plan ruled out any web interface or dashboard. So this task first shows what the app looks like and what its sidebar holds.

## Success measure

**Non-engineers can set up kirokit without the command line.** Engineers still get a command-line option inside the desktop app.

## Done means

1. **The mockup opens full size in the QualityLayer App.** One design page shows the desktop window with its sidebar and main pane.
2. **The sidebar lists all five sections, and each opens its page.** The sections are Teams, Agents, Install and sync, Skills and MCP versions, and Settings.
3. **The sidebar collapses to an icon rail and expands again.** One toggle switches it, and the main pane widens while the rail shows.
4. **A non-engineer can click through setting up a team.** They pick a team, see its agents, press Install, and see the team marked installed.
5. **Engineers can see the command behind every action.** A command-line panel shows the exact `kirokit` command, such as `kirokit install platform-team`.
6. **The mockup shows kirokit's real teams, agents and versions.** Names and versions come from this repository, not placeholder text.
7. **The mockup reads well in light and dark mode.** It uses colours taken from the kirokit logo. (only you can confirm: the look fits kirokit)
8. **kirokit's code and behaviour stay unchanged.** The change adds only the design page and this task's documents.

> 4. **A non-engineer can click through setting up a team.** They pick a team and pass a prerequisites check for SOPS, the key and Kiro CLI. Then they press Install and see the team marked installed.

> 5. **Engineers get a full terminal docked at the bottom.** It opens, closes and resizes like VS Code's panel. Each button press appears there as its `kirokit` command. A typed command also updates the screens, for example a typed install marks the team installed.

> 9. **Only one team is installed at a time.** The app never shows two teams as installed, whether the install came from a button or the terminal.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | What should the first release of this be? | Clickable design mockup, no app code |
| 2 | How will you know the desktop app was worth it? | non-engineers can setup kirokit without cli, engineer still have a cli option in the desktop app |
| 3 | Which sections should the sidebar list? | Teams + Agents + Install and sync status + Skills and MCP versions + Settings (KIRO_HOME, repo path) |
| 4 | What stays out of the first release? | Editing templates or manifests |
| 5 | Where should the desktop app's code live? | Inside this kirokit repository |
| 6 | Which desktop technology should the app use? | Wails (Go backend, web UI) |
| 7 | How should the sidebar behave? | Collapsible to an icon rail |
<!-- /ql:decided -->

## First release

**The first release is a clickable design mockup, with no app code.** It is one HTML design page that you open and click through in the QualityLayer App. Nothing is built, packaged or released, and kirokit itself does not change.

| Option | What you get | Chosen |
|---|---|---|
| Clickable design mockup | A design page showing the sidebar and its pages | Yes |
| Runnable app shell | A desktop window with the sidebar and placeholder pages | Later |
| Shell plus one working page | The shell, with one page reading real kirokit data | Later |

**A real app becomes a later task.** That task starts from this approved mockup. It lives inside this repository and uses Wails, a Go backend with a web interface, as you decided.

## What the user sees

**A desktop window with a collapsible sidebar on the left.** The kirokit logo sits on top, with the five sections below and a collapse toggle at the bottom.

```
┌──────────────┬──────────────────────────────────────────┐
│ ◆ kirokit    │  Teams                                   │
│              │  ┌──────────────┐  ┌──────────────┐      │
│ ▸ Teams      │  │ platform-team│  │ data-team    │      │
│   Agents     │  │ 2 agents     │  │ 2 agents     │      │
│   Install    │  │ [ Install ]  │  │ Installed ✓  │      │
│   Skills/MCP │  └──────────────┘  └──────────────┘      │
│   Settings   │                                          │
│              ├─ TERMINAL ───────────────────────── ✕ ⌄ ┤
│              │ $ kirokit install platform-team          │
│ « Collapse   │ installed platform-team into ~/.kiro     │
└──────────────┴──────────────────────────────────────────┘
```

- **Teams lists every team with an Install button.** This is where a non-engineer starts.
- **Install starts with a prerequisites check.** It shows SOPS, the decryption key and Kiro CLI as found or missing before Install is enabled.
- **Agents shows each team's agents and their tools.** It is read-only.
- **Install and sync shows what is installed and whether it is current.** Each entry has Sync and Uninstall actions.
- **Skills and MCP versions shows each skill's source and every pinned server.** It is read-only.
- **Settings holds the Kiro home folder and the repository path.** These are the two inputs the command line needs.
- **A terminal panel docks at the bottom, as in VS Code.** Engineers type any `kirokit` command there, and every button press shows up as its command.

## Not doing

- **Editing templates or manifests.** The app shows teams and installs them, but changes stay in git and pull requests.
- **Any app code in this release.** No Wails project, no binary and no release step until the mockup is approved.

<!-- optional section here, only when there is content: "Worth it when" -->

<!-- optional section here, only when there is content: "Reference" -->

<!-- optional section here, only when there is content: "Added while building" -->

## Context pointers

- `.qualitylayer/designs/desktop-app/` — empty design page, purpose "Design for the desktop app with a sidebar layout" (created 2026-10-17, no content yet)
- `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md` — v1 PRD; lists "A web UI or dashboard for managing teams, manifests, or viewing rendered diffs" under *Explicitly Out of Scope*
- `assets/logo.svg`, `assets/logo.png` — the project's only visual assets

## First look at the code

There is no desktop app in this repository today. `kirokit` is a Go CLI (cobra, `go 1.27`, no UI dependency in `go.mod`), so this task would create the app, not change one.

| What | Where | Why it matters |
|---|---|---|
| CLI entry point | `cmd/kirokit/main.go:12` | Only binary the module builds today |
| Command tree | `cmd/kirokit/commands/root.go:26` | `render`, `test`, `update`, `install`, `uninstall`, `sync`, `skills`, `onboard`, `version`: the actions a sidebar could expose |
| Team discovery | `cmd/kirokit/commands/root.go:47` (`teamNames`) | How the CLI lists teams from `teams/`; a "Teams" section would need the same |
| Install and sync | `cmd/kirokit/commands/install.go:13`, `cmd/kirokit/commands/sync.go:18` | Write into `$KIRO_HOME`; any app action here touches the user's real config |
| Domain logic | `internal/render`, `internal/install`, `internal/manifest`, `internal/externalskills` | Reusable from Go directly; reusable from a non-Go app only through the CLI |
| Release | `.goreleaser.yaml`, `.github/workflows/release.yaml` | Builds darwin and linux CLI archives only |
| Out-of-scope note | `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md:86` | v1 ruled out a web UI or dashboard |
| Design stub | `.qualitylayer/designs/desktop-app/index.html` | Empty `<body>`; no colours, type or layout yet |

*Untraced areas:* which desktop framework fits (none is in the repo), whether `internal/*` packages depend on the working directory in ways a long-running app would trip over, and how goreleaser would package a GUI binary.
