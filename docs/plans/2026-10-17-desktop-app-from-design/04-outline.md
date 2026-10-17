---
type: outline
---

# Outline: Desktop app from design

## Summary

Seven slices and 24 tasks, built in one chain: the CLI changes first, then the app shell, real data, the terminal, the install flow, the build and CI, and last the release. The risky parts are the new dependencies the app ships (Wails, creack/pty, xterm.js), the changed `kirokit install` and `kirokit sync`, and the release job that pushes to your Homebrew tap, which only a real tagged release exercises.

## How the build runs

<!-- ql:waves -->
- Wave 1 · Slice 1 (One team stays installed and kirokit runs from any folder) · first
- Wave 2 · Slice 2 (The app opens with the design's shell and remembers its settings) · on its own
- Wave 3 · Slice 3 (The pages show the repository's real data) · on its own
- Wave 4 · Slice 4 (A real terminal docks at the bottom) · on its own
- Wave 5 · Slice 5 (A person installs a team with buttons only) · on its own
- Wave 6 · Slice 6 (One command builds the macOS app and CI builds it too) · on its own
- Wave 7 · Slice 7 (Each release attaches the app and a Homebrew cask) · on its own

7 waves · longest path 24 of 24 tasks

A fresh agent runs the scenarios after slice 1 (`install` and `sync` now uninstall other teams for every CLI user.) and slice 2 (new dependency the shipped app loads: Wails v2.14.0 in a nested module.) and slice 4 (new dependencies the shipped app loads: creack/pty and xterm.js 6, and xterm.js inside the Wails webview is not yet proven.) and slice 7 (an external system: the release job pushes to your Homebrew tap and runs only on a real release.) before the build goes on.
<!-- /ql:waves -->

## Program design

### Repository root resolution

```go
// repoBase is the absolute KIROKIT_REPO, or "" when unset or empty.
var repoBase string
// repoPath joins repoBase and rel; with no base it returns rel unchanged.
func repoPath(rel string) string
```

`newRootCmd` sets `repoBase` once in `PersistentPreRunE`, made absolute with `filepath.Abs`. The relative names `teamsRoot`, `globalRoot`, `envVersionsPath` and `goldenRoot` stay as constants, so existing tests that use them still compile. Every caller passes them through `repoPath`. `sync` resolves its repository as the explicit `--repo`, else `repoBase`, else `.`, changes into it and sets `repoBase = ""`, so an explicit `--repo` roots every path as it does today.

### The app finds its bundled kirokit in a fixed order

```go
// BundledCLI returns the kirokit binary the app's terminals use.
// 1. $KIROKIT_DESKTOP_CLI if set. 2. <executable>/../../Resources/bin/kirokit.
// 3. desktop/build/dev-bin/kirokit, which `task desktop:dev` builds.
func BundledCLI() (string, error)
```

### An old clone is renamed, never deleted

```
Clone(address):
  validate address                      refuse with a message beside the field
  if <app folder>/repo exists:          rename it to repo.old-<unix time>
  run: git clone -- <address> <app folder>/repo      in the kirokit tab
  run: kirokit sync --latest-tag                     with KIROKIT_REPO set
  on a clone failure:                   restore the renamed folder, keep the address, emit clone:done {ok:false}
  on "no v* tag":                       keep the clone, emit clone:done {ok:true, error:"This repository has no release yet (no v* tag). Ask your team lead to tag one."}
```

### The frontend reaches Go through one module

`frontend/src/api.js` wraps `window.go.main.App.*` and `window.runtime.EventsOn`. No generated bindings are committed, so `npm run build` works before any Wails run. Pages import `api.js` only.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Repository root and one-team rule | `cmd/kirokit/commands/root.go`, `internal/install/installed.go` | `install.Uninstall`, `install.ResolveKiroHome` |
| App service and settings | `desktop/app.go`, `desktop/settings.go` | `internal/install`, `internal/render` |
| Catalog, prerequisites, login environment | `desktop/catalog.go`, `desktop/prereq.go`, `desktop/envcapture.go` | `render.RenderTeam`, `internal/manifest`, `internal/mcpversions` |
| Terminals | `desktop/terminal/`, `desktop/shell.go`, `desktop/runner.go` | creack/pty |
| Pages and dock | `desktop/frontend/src/` | the design page's tokens and markup |
| Release files | `.github/workflows/desktop.yaml`, `scripts/desktop-release.sh` | `release.yaml`, the CLI cask |

## Slice 1: One team stays installed and kirokit runs from any folder

`KIROKIT_REPO` roots every repository path, `install` and `sync` uninstall any other installed team first, and `sync` without a team only updates the checkout. After this slice the CLI side of the app's contract works and is documented.

```diff
 cmd/kirokit/commands/
+├── root_test.go          + KIROKIT_REPO from another working directory
 ├── root.go               ~ repoBase, repoPath, teamNames
 ├── install.go            ~ removes other teams, prints both lines
 ├── sync.go               ~ team-less form, removals first
 └── render.go, test.go, update.go, onboard.go, skills.go   ~ roots through repoPath
 internal/install/
+├── installed.go          + Installed, UninstallOthers
+└── installed_test.go
 README.md                 ~ KIROKIT_REPO, team-less sync, one-team rule, migration note
```

- [ ] T1 Every command reads its folders from `KIROKIT_REPO`
  - **Objective:** Add `repoPath` and use it everywhere the four root constants are read, so a command works from any folder.
  - **Design:** "Repository root resolution", "`KIROKIT_REPO` roots every repository path"
  - **Files:** `cmd/kirokit/commands/root.go`, `cmd/kirokit/commands/install.go`, `cmd/kirokit/commands/sync.go`, `cmd/kirokit/commands/test.go`, `cmd/kirokit/commands/update.go`, `cmd/kirokit/commands/render.go`, `cmd/kirokit/commands/onboard.go`, `cmd/kirokit/commands/skills.go`, `cmd/kirokit/commands/root_test.go`
  - **Test first:** In `root_test.go`, run `render`, `onboard`, `skills fetch` argument errors and `teamNames` from a temp working directory with `KIROKIT_REPO` set to this repository; it fails today because the folders resolve against the working directory.
  - **Definition of done:**
    - With `KIROKIT_REPO` set, `render platform-team --out <dir>` works from an unrelated folder and `teamNames("")` lists both teams.
    - `onboard` creates the team under `$KIROKIT_REPO/teams`, and `skills fetch` reads `$KIROKIT_REPO/teams/<team>/manifest.yaml`.
    - A relative `KIROKIT_REPO` is made absolute once; an empty one changes nothing.
    - `sync` with an explicit `--repo` roots every path at it, even when `KIROKIT_REPO` is set.
    - verify: `go test ./cmd/kirokit/commands/ ./internal/golden/`

- [ ] T2 The install library lists installed teams and removes the others
  - **Objective:** Add `Installed` and `UninstallOthers` beside the install library without changing `Install` or `Uninstall`.
  - **Design:** "kirokit keeps one team installed and prints each removal"
  - **Files:** `internal/install/installed.go`, `internal/install/installed_test.go`
  - **Test first:** In `installed_test.go`, with `setupFixture(t)` install both fixture teams into a temp Kiro home and call `Installed`; it fails because the function does not exist.
  - **Definition of done:**
    - `Installed` returns the teams sorted by name with file count and the record's modification time.
    - A missing install-records folder returns an empty list, not an error.
    - `UninstallOthers(home, keep)` removes every installed team except `keep` and returns their names in order; the kept team is untouched.
    - A failed uninstall returns its error and stops.
    - verify: `go test ./internal/install/`

- [ ] T3 `kirokit install` replaces the installed team and says so
  - **Objective:** Before installing, remove every other installed team and print one `uninstalled` line each, then the existing `installed` line.
  - **Design:** "kirokit keeps one team installed and prints each removal"
  - **Files:** `cmd/kirokit/commands/install.go`, `cmd/kirokit/commands/install_test.go`
  - **Test first:** In `install_test.go`, run the install command for `platform-team` then `data-team` with `KIROKIT_REPO` and a temp `KIRO_HOME`; it fails because both teams stay installed.
  - **Definition of done:**
    - Installing `data-team` after `platform-team` prints `uninstalled platform-team from <home>` then `installed data-team into <home>`, and only `data-team.json` remains.
    - Reinstalling the installed team prints only the `installed` line and removes nothing.
    - A failing uninstall prints no `installed` line and returns `kirokit install: …`.
    - verify: `go test ./cmd/kirokit/commands/ -run Install`

- [ ] T4 `kirokit sync` removes other teams and works without a team
  - **Objective:** `sync <team>` applies the same removals first; `sync` without a team only updates the checkout.
  - **Design:** "kirokit sync without a team updates the checkout only"
  - **Files:** `cmd/kirokit/commands/sync.go`, `cmd/kirokit/commands/sync_test.go`
  - **Test first:** In `sync_test.go`, build a temp origin repository with a `v*` tag and run team-less sync in a clone; it fails because `sync` demands exactly one argument.
  - **Definition of done:**
    - Without a team and with `--latest-tag`, it prints `updated <before> -> <after>` when HEAD moved, then `checked out <tag> at <after>`, and installs nothing.
    - `--kb-agent`, `--kb-source` or `--kb-mount` without a team fail with `kirokit sync: --kb-agent needs a team`.
    - With a team, other installed teams are removed first with one `uninstalled` line each, then today's output follows.
    - `Use`, `Long` and the `--latest-tag` help describe the team-less form.
    - verify: `go test ./cmd/kirokit/commands/ -run Sync`

- [ ] T5 The README describes the new CLI behaviour
  - **Objective:** Document `KIROKIT_REPO`, team-less sync and the one-team rule, with a migration note.
  - **Design:** none
  - **Files:** `README.md`, `scripts/kirokit-sync-setup.sh:1`
  - **Test first:** **Trivial:** documentation only; `go test ./...` and a read of the rendered sections cover it.
  - **Definition of done:**
    - The Install / Uninstall section states that installing a team replaces the installed one.
    - A migration note says two scheduled syncs of different teams now replace each other, pointing at the scheduled-sync setup script.
    - `KIROKIT_REPO` and `sync` without a team appear in the command reference.
    - verify: `go test ./...`

**Validation:** automated: `go build ./...` · `go vet ./...` · `go test ./...` · `golangci-lint run` · manual: none

**Risky:** `install` and `sync` now uninstall other teams for every CLI user.

## Slice 2: The app opens with the design's shell and remembers its settings

**Depends on:** Slice 1

A Wails window shows the logo, the sidebar with five sections in light and dark mode, the collapsing icon rail, and a Settings page that saves. After this slice `task desktop:dev` runs a real app and Done means 2 and 10 hold.

```diff
 Taskfile.yml                 ~ desktop:dev, desktop:test
+desktop/
+├── go.mod, main.go, app.go, settings.go, settings_test.go, wails.json, .gitignore
+└── frontend/
+    ├── package.json, package-lock.json, vite.config.js, index.html
+    └── src/ api.js, main.js, sidebar.js, sidebar.test.js, style.css, pages/settings.js
```

- [ ] T6 A nested Wails module opens an empty window
  - **Objective:** Create `desktop/` as its own Go module with a `replace` to the repository, a bound `App`, and the two Taskfile entries.
  - **Design:** "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/go.mod`, `desktop/go.sum`, `desktop/main.go`, `desktop/app.go`, `desktop/app_test.go`, `desktop/wails.json`, `desktop/.gitignore`, `desktop/build/appicon.png`, `desktop/frontend/package.json`, `desktop/frontend/package-lock.json`, `desktop/frontend/index.html`, `desktop/frontend/vite.config.js`, `Taskfile.yml`
  - **Test first:** In `app_test.go`, `TestState_Defaults` calls `NewApp().State()`; it fails because the module does not exist.
  - **Definition of done:**
    - `State()` returns empty settings, no clone and an env status field.
    - Root `go build ./...` and `go test ./...` still ignore `desktop/`.
    - The module path is `github.com/mavogel/kirokit/desktop` with a `replace` to the repository root, so kirokit's internal packages import.
    - `wails.json` names the app `kirokit` with bundle id `io.github.mavogel.kirokit`; `build/appicon.png` is made from the project logo.
    - `task desktop:test` builds the frontend, runs `node --test src/`, then vets and tests the module, so CI runs the frontend tests too; `task desktop:dev` runs `wails dev` at the pinned v2.14.0.
    - `main.go` embeds `all:frontend/dist` and sets the minimum macOS to 13.
    - verify: `task desktop:test`

- [ ] T7 The sidebar, rail and pages match the design
  - **Objective:** Port the design's tokens, layout, five pages and shortcut handler into the Vite frontend, with light and dark themes.
  - **Design:** "The frontend reaches Go through one module"
  - **Files:** `assets/logo.svg`, `desktop/frontend/src/main.js`, `desktop/frontend/src/style.css`, `desktop/frontend/src/sidebar.js`, `desktop/frontend/src/sidebar.test.js`, `desktop/frontend/src/api.js`, `desktop/frontend/index.html`, `desktop/frontend/package.json`
  - **Test first:** `sidebar.test.js` (node's built-in test runner) toggles the state and sends Command-B; it fails because `sidebar.js` does not exist.
  - **Definition of done:**
    - The toggle and Command-B both collapse the sidebar to the icon rail and expand it again; the main pane widens while the rail shows.
    - Plain B and Control-B do nothing.
    - The five sidebar buttons switch pages; the active one has `aria-current="page"`.
    - `prefers-color-scheme: dark` switches the tokens.
    - The logo in the sidebar is the project's logo asset.
    - verify: `cd desktop/frontend && node --test src/ && npm run build`

- [ ] T8 Settings validates and remembers both values
  - **Objective:** Store `RepoURL` and `KiroHome` in `settings.json` and show them on the Settings page.
  - **Design:** "Data shapes the pages render"
  - **Files:** `desktop/settings.go`, `desktop/settings_test.go`, `desktop/app.go`, `desktop/frontend/src/pages/settings.js`, `desktop/frontend/src/main.js`
  - **Test first:** In `settings_test.go`, save then load settings from a temp data folder; it fails because `settings.go` does not exist.
  - **Definition of done:**
    - Saved values come back after a new `App` is created on the same folder.
    - Kiro home refuses a relative path, `/`, and the user's home folder, with a message beside the field.
    - A repository address must match `https://…`, `ssh://…`, `git@host:path`, `file://…` or an absolute path and must not start with `-`.
    - Default Kiro home is `~/.kiro`; slice 3 replaces it with the captured `KIRO_HOME` when one exists.
    - verify: `task desktop:test`

**Validation:** automated: `task desktop:test` · `go build ./...` · manual: `task desktop:dev` opens a window with the design's shell

**Risky:** new dependency the shipped app loads: Wails v2.14.0 in a nested module.

## Slice 3: The pages show the repository's real data

**Depends on:** Slice 2

Teams, Agents, Install and Skills and MCP versions read the app's own copy of the repository, and the screens redraw when install records change. After this slice Done means 3 holds and the prerequisites are real.

```diff
 desktop/
+├── envcapture.go, envcapture_test.go   + login shell environment
+├── prereq.go, prereq_test.go           + SOPS, key, Kiro CLI checks
+├── catalog.go, catalog_test.go         + teams, agents, skills, pins
+├── installs.go, installs_test.go       + records poller, installs:changed
 ├── app.go                              ~ Catalog, Installs, Prerequisites
 └── frontend/src/pages/                 + teams.js, agents.js, install.js, skills.js
```

- [ ] T9 The login shell's environment decides the prerequisite checks
  - **Objective:** Capture the person's login environment once at start and answer the three checks from it.
  - **Design:** "Prerequisites come from the captured login environment"
  - **Files:** `desktop/envcapture.go`, `desktop/envcapture_test.go`, `desktop/prereq.go`, `desktop/prereq_test.go`, `desktop/app.go`, `desktop/settings.go`, `desktop/frontend/src/pages/settings.js`, `desktop/frontend/src/api.js`
  - **Test first:** `envcapture_test.go` runs `Capture` with a fake shell script that prints a marker then `env -0`; it fails because the package has no `Capture`.
  - **Definition of done:**
    - Only the output after the marker is read, so shell startup noise is ignored.
    - A shell that sleeps past five seconds makes `Capture` return a fallback PATH with the Homebrew and local tool folders and a note shown in Settings.
    - The key is found at `$SOPS_AGE_KEY_FILE`, else `$XDG_CONFIG_HOME/sops/age/keys.txt`, else `~/Library/Application Support/sops/age/keys.txt`.
    - `Prerequisites` reports SOPS and Kiro CLI by resolving them on the captured PATH, using fake PATH folders in tests.
    - "Check again" captures afresh.
    - The captured PATH always gets the Homebrew and local tool folders appended when missing, so a login shell without startup files still finds them.
    - The window opens before the capture finishes; the default Kiro home becomes the captured `KIRO_HOME` once it arrives.
    - verify: `cd desktop && go test -run 'Capture|Prereq' ./...`

- [ ] T10 The catalog reads teams, agents, skills and pins from the clone
  - **Objective:** Render each team into a temporary folder and read the result into `Catalog`.
  - **Design:** "Data shapes the pages render"
  - **Files:** `desktop/catalog.go`, `desktop/catalog_test.go`, `desktop/app.go`
  - **Test first:** `catalog_test.go` builds the catalog from this repository's team folders; it fails because `BuildCatalog` does not exist.
  - **Definition of done:**
    - Both real teams appear with agent names, descriptions, tools and servers.
    - `NeedsKey` is true exactly for agents that ship an encrypted secrets file.
    - A skill shows "Written by the team" or "<source> @ <ref>", and pins come from the central MCP version file.
    - A render error is returned on the page's data with its message, not a crash.
    - `Catalog()` serves the cached result until a clone or sync ends.
    - verify: `cd desktop && go test -run Catalog ./...`

- [ ] T11 Install records drive a polled event
  - **Objective:** Check the records every second and after each command, emitting `installs:changed` only when they differ.
  - **Design:** "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/installs.go`, `desktop/installs_test.go`, `desktop/app.go`
  - **Test first:** `installs_test.go` installs a fixture team into a temp home behind a poller with a fake emitter; it fails because there is no poller.
  - **Definition of done:**
    - A new record, a removed record and a changed file time each emit once; an unchanged state emits nothing.
    - The Kiro home folder comes from Settings and a Settings change takes effect on the next check.
    - Stopping the poller ends its goroutine.
    - verify: `cd desktop && go test -race -run Installs ./...`

- [ ] T12 The four pages show real data
  - **Objective:** Replace the design's hard-coded teams, agents, installs and pins with bound calls and the event.
  - **Design:** "Data shapes the pages render"
  - **Files:** `desktop/frontend/src/pages/teams.js`, `desktop/frontend/src/pages/agents.js`, `desktop/frontend/src/pages/install.js`, `desktop/frontend/src/pages/skills.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/api.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** `pages.test.js` renders each page's view function from a sample `Catalog` and installs list; it fails because the view functions do not exist.
  - **Definition of done:**
    - Teams shows an "installed" badge from the records, with the time of the last change.
    - The Install page lists installed teams with their file counts.
    - A render error replaces the table with its message.
    - Nothing in the pages names `platform-team` or `data-team` in code.
    - verify: `cd desktop/frontend && node --test src/ && npm run build`

**Validation:** automated: `task desktop:test` · manual: none

## Slice 4: A real terminal docks at the bottom

**Depends on:** Slice 3

The dock opens, closes and resizes like VS Code's and holds two tabs: your login shell and the kirokit tab for button commands. After this slice typed commands run for real and update the screens.

```diff
 desktop/
+├── terminal/terminal.go, terminal_test.go   + pty session
+├── shell.go, shell_test.go                  + zsh tab, ZDOTDIR wrapper
+├── runner.go, runner_test.go                + kirokit tab, Run, RunWrite
 ├── app.go                                   ~ ShellWrite, ShellResize, Run, RunWrite
 └── frontend/src/terminal.js, tabs.js, tabs.test.js   + xterm.js dock
```

- [ ] T13 The dock renders xterm.js inside the webview
  - **Objective:** Port the design's dock with xterm.js 6 and a fit addon, with tab state in a testable module; this is the first proof that xterm.js works inside Wails.
  - **Design:** "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/frontend/src/terminal.js`, `desktop/frontend/src/tabs.js`, `desktop/frontend/src/tabs.test.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/style.css`, `desktop/frontend/package.json`, `desktop/frontend/package-lock.json`
  - **Test first:** `tabs.test.js` opens, closes and switches tabs and the dock height; it fails because `tabs.js` does not exist.
  - **Definition of done:**
    - The dock opens, closes and is resized by dragging, within a minimum and maximum height.
    - Two tabs, "zsh" and "kirokit", switch without losing either's output.
    - The terminal refits when the window or dock resizes and reports its columns and rows.
    - verify: `cd desktop/frontend && node --test src/ && npm run build`

- [ ] T14 A pseudo-terminal package streams a process
  - **Objective:** Add `terminal.Session` over creack/pty with start, read, write, resize and a reaped exit.
  - **Design:** "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/terminal/terminal.go`, `desktop/terminal/terminal_test.go`, `desktop/go.mod`, `desktop/go.sum`
  - **Test first:** `terminal_test.go` starts `/bin/sh -c 'echo hi'`; it fails because the package does not exist.
  - **Definition of done:**
    - Output reaches the reader and the read loop ends on EOF.
    - `Resize` changes the size the process sees (`stty size` prints the new values).
    - The exit code is returned and the child is reaped.
    - `Close` while running ends the process.
    - verify: `cd desktop && go test -race ./terminal/`

- [ ] T15 The zsh tab runs your shell with the bundled kirokit first
  - **Objective:** Start `$SHELL -il` through the `ZDOTDIR` wrapper, stream it as `term:shell` and wire the tab to the dock.
  - **Design:** "The zsh tab starts through a wrapper that keeps the app's settings last", "The app finds its bundled kirokit in a fixed order"
  - **Files:** `desktop/shell.go`, `desktop/shell_test.go`, `desktop/app.go`, `desktop/frontend/src/terminal.js`, `desktop/frontend/src/api.js`
  - **Test first:** `shell_test.go` starts the shell with a fake bundled folder and runs `command -v kirokit`; it fails because `shell.go` does not exist.
  - **Definition of done:**
    - `command -v kirokit` in the tab prints the bundled path, even when the person's `.zshrc` prepends another folder.
    - `echo $KIROKIT_REPO $KIRO_HOME` prints `<app folder>/repo` and the Settings value, even when the person's files set `KIRO_HOME`.
    - For a shell other than zsh the same variables are set in the start environment and Settings says the bundled CLI may not come first.
    - Keystrokes typed in the tab reach the shell and the dock resize reaches `ShellResize`.
    - verify: `cd desktop && go test -race -run Shell ./...`

- [ ] T16 Buttons run in their own kirokit tab, one command at a time
  - **Objective:** `Run` starts the bundled CLI in a fresh pseudo-terminal, prints `▸ kirokit <args>`, streams `term:kirokit` and forwards `RunWrite`.
  - **Design:** "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/runner.go`, `desktop/runner_test.go`, `desktop/app.go`, `desktop/frontend/src/terminal.js`
  - **Test first:** `runner_test.go` runs a fake CLI script twice at once; it fails because `Run` does not exist.
  - **Definition of done:**
    - The first output is `▸ kirokit <args>`, then the command's real output.
    - A second `Run` while one runs returns "kirokit is still running a command".
    - `term:kirokit:exit` carries the run id and exit code, and the tab's input is live only while a command runs.
    - `RunWrite` reaches the running command's input.
    - A missing bundled CLI returns a message naming `KIROKIT_DESKTOP_CLI`.
    - Every run gets `KIROKIT_REPO=<app folder>/repo`, `KIRO_HOME` from Settings and the captured login environment, so `git`, `sops` and `kiro-cli` resolve.
    - A run can start any program, not only kirokit: the header shows the program's name (`▸ kirokit …` or `▸ git …`), and all runs share the one-command lock.
    - verify: `cd desktop && go test -race -run Runner ./...`

**Validation:** automated: `task desktop:test` · manual: `task desktop:dev` shows xterm.js rendering in the window and the zsh tab accepting input

**Risky:** new dependencies the shipped app loads: creack/pty and xterm.js 6, and xterm.js inside the Wails webview is not yet proven.

## Slice 5: A person installs a team with buttons only

**Depends on:** Slice 4

First launch asks for the repository's address and clones it, the setup dialog checks the prerequisites, and Install, Sync, Uninstall and Update run their `kirokit` commands in the terminal. After this slice Done means 4 to 6, 8, 9, 11 and 14 hold.

```diff
 desktop/
+├── clone.go, clone_test.go                 + validation, rename, clone, first sync
 ├── app.go                                  ~ Clone, State, catalog refresh after sync
 └── frontend/src/
+    ├── pages/welcome.js                    + first launch
+    └── dialogs.js, dialogs.test.js         + setup and replace dialogs
```

- [ ] T17 Clone fetches the repository and the newest release tag
  - **Objective:** Validate the address, keep any old clone, clone with `git`, then run `kirokit sync --latest-tag`.
  - **Design:** "An old clone is renamed, never deleted", "kirokit sync without a team updates the checkout only"
  - **Files:** `desktop/clone.go`, `desktop/clone_test.go`, `desktop/app.go`, `desktop/runner.go`
  - **Test first:** `clone_test.go` clones a temp origin with a `v1.0.0` tag through `Clone`; it fails because `Clone` does not exist.
  - **Definition of done:**
    - A local path address clones into `<app folder>/repo` and checks out the newest `v*` tag; `clone:done` is `{ok:true}`.
    - An address starting with `-` or not matching an allowed form is refused before any process starts.
    - A second address renames the old folder to `repo.old-<time>` first, and a failed clone puts it back and keeps the address.
    - An origin with no `v*` tag keeps the clone and returns the no-release message.
    - The address reaches git as `git clone -- <address> <folder>`, never through a shell; it runs in the kirokit tab as `▸ git clone -- …`, then `kirokit sync --latest-tag` runs after it under the same lock.
    - verify: `cd desktop && go test -race -run Clone ./...`

- [ ] T18 First launch asks for the address and shows the clone
  - **Objective:** Show the welcome page until a clone exists, run `Clone` from it and load the catalog when `clone:done` arrives.
  - **Design:** "An old clone is renamed, never deleted", "The app's Go side binds one service and emits five events"
  - **Files:** `desktop/frontend/src/pages/welcome.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/api.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** `pages.test.js` renders the welcome state from a `State` without a clone and with a failed `clone:done`; it fails because `welcome.js` does not exist.
  - **Definition of done:**
    - The sidebar pages are hidden until a clone exists; the field is prefilled from Settings.
    - A refused or failed address stays in the field with the error beside it.
    - A sign-in prompt appears in the kirokit tab and can be answered there.
    - After success the Teams page shows.
    - verify: `cd desktop/frontend && node --test src/ && npm run build`

- [ ] T19 Install runs after a real check and replaces the installed team
  - **Objective:** The setup dialog shows the three checks with "Check again", Install runs `kirokit install <team>`, and the replace dialog appears when another team is installed.
  - **Design:** "Prerequisites come from the captured login environment"
  - **Files:** `desktop/frontend/src/dialogs.js`, `desktop/frontend/src/dialogs.test.js`, `desktop/frontend/src/pages/teams.js`, `desktop/frontend/src/main.js`
  - **Test first:** `dialogs.test.js` builds the dialog state from checks with one missing; it fails because `dialogs.js` does not exist.
  - **Definition of done:**
    - Install stays disabled while a check fails, and the dialog's status names how many fail.
    - Pressing Install runs `kirokit install <team>` and the kirokit tab shows `▸ kirokit install <team>` and `installed <team> into <home>`.
    - With another team installed the replace dialog names both; confirming runs the same command, whose output shows both lines.
    - Cancelling the replace dialog runs nothing; installing the team that is already installed skips it.
    - While a command runs, Install, Sync, Uninstall and Update are disabled, and a "still running" error shows beside the button.
    - verify: `cd desktop/frontend && node --test src/ && npm run build`

- [ ] T20 Sync, Uninstall and Update run their commands and refresh
  - **Objective:** Wire Sync, Uninstall and Settings' Update to `Run`, and refresh the catalog and install state when a command ends.
  - **Design:** "`kirokit sync` without a team updates the checkout only"
  - **Files:** `desktop/frontend/src/pages/install.js`, `desktop/frontend/src/pages/settings.js`, `desktop/runner.go`, `desktop/runner_test.go`, `desktop/app.go`
  - **Test first:** `runner_test.go` ends a fake `sync` run and checks the catalog cache; it fails because the cache is not invalidated.
  - **Definition of done:**
    - Sync runs `kirokit sync <team> --latest-tag` and Uninstall runs `kirokit uninstall <team>`.
    - Update runs `kirokit sync --latest-tag` when no team is installed and `kirokit sync <team> --latest-tag` otherwise.
    - After a `sync` ends the next `Catalog()` call renders again and shows a new team or pin.
    - verify: `task desktop:test`

**Validation:** automated: `task desktop:test` · `go test ./...` · manual: none

## Slice 6: One command builds the macOS app and CI builds it too

**Depends on:** Slice 5

`task desktop:build` produces a universal `kirokit.app` with the CLI inside, and a new workflow runs the app's tests and build on macOS. After this slice Done means 12 holds.

```diff
 scripts/desktop-build.sh               + builds both CLIs, lipo, wails build, copies CLI in
 Taskfile.yml                           ~ desktop:build
 .github/workflows/desktop.yaml         + macOS tests and build
```

- [ ] T21 `task desktop:build` makes a universal app with the CLI inside
  - **Objective:** Build the CLI for both architectures with the release ldflags, join them with `lipo`, build the app and copy the CLI to `Contents/Resources/bin/kirokit`.
  - **Design:** "The app bundle and release artifacts"
  - **Files:** `scripts/desktop-build.sh`, `Taskfile.yml`, `desktop/.gitignore`
  - **Test first:** **Trivial:** a build script has no unit seam; `task desktop:build` followed by `lipo -archs` on both binaries is the check.
  - **Definition of done:**
    - `desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` and `Contents/Resources/bin/kirokit` both report `x86_64 arm64`.
    - The bundled `kirokit version` prints the version, commit and date set by ldflags; the script takes `VERSION` from its environment, defaulting to `git describe`, and the release job passes the new tag.
    - `task desktop:dev` builds the dev CLI at `desktop/build/dev-bin/kirokit` first.
    - verify: `task desktop:build`

- [ ] T22 CI builds the app on macOS
  - **Objective:** Add `desktop.yaml`, limited to pushes to `main` and pull requests touching the app or what it imports.
  - **Design:** "The app bundle and release artifacts"
  - **Files:** `.github/workflows/desktop.yaml`
  - **Test first:** **Trivial:** a workflow cannot run locally; parsing the YAML and reading its triggers is the check.
  - **Definition of done:**
    - It runs on `macos-latest`, executes `task desktop:test` and `task desktop:build`, and pins actions to the same commit hashes as the existing CI workflow.
    - Its path filter lists the app folder, the command and internal packages, the Go module files and the Taskfile.
    - verify: `ruby -ryaml -e 'YAML.load_file(".github/workflows/desktop.yaml")'`

**Validation:** automated: `task desktop:build` · `go test ./...` · `ruby -ryaml -e 'YAML.load_file(".github/workflows/desktop.yaml")'` · manual: none

## Slice 7: Each release attaches the app and a Homebrew cask

**Depends on:** Slice 6

A tagged release also builds the macOS app, attaches `kirokit-desktop_<version>_darwin_universal.zip` and pushes the `kirokit-desktop` cask. After this slice the release workflow is ready and its dry run works.

```diff
 scripts/desktop-release.sh             + zip, sha256, cask text, optional upload and push
 scripts/desktop-release_test.sh        + fake app, dry run output
 .github/workflows/release.yaml         ~ outputs, dispatch dry run, desktop job
 README.md, AGENTS.md                   ~ desktop app, layout
```

- [ ] T23 A release script zips the app and renders the cask
  - **Objective:** Create the zip with `ditto`, compute its checksum and print the cask; upload and push only when told to.
  - **Design:** "The app bundle and release artifacts"
  - **Files:** `scripts/desktop-release.sh`, `scripts/desktop-release_test.sh`
  - **Test first:** `desktop-release_test.sh` runs the script on a fake `kirokit.app` folder in dry-run mode; it fails because the script does not exist.
  - **Definition of done:**
    - The zip is named `kirokit-desktop_<version>_darwin_universal.zip` and keeps the app folder as its top level.
    - The printed cask has the version, the zip's sha256, the release URL, `app "kirokit.app"` and a postflight that clears `com.apple.quarantine`.
    - Flags are `--version`, `--app`, `--out` and `--publish`; only `--publish` runs `gh release upload v<version>` and pushes the cask file to the tap repository's default branch.
    - The cask URL is `https://github.com/mavogel/kiro-cli-kit/releases/download/v<version>/kirokit-desktop_<version>_darwin_universal.zip`.
    - Without `--publish` nothing is uploaded or pushed, and the script says so.
    - A missing app folder or version exits non-zero with a message.
    - verify: `bash scripts/desktop-release_test.sh`

- [ ] T24 The release workflow gains a macOS desktop job, and the docs follow
  - **Objective:** Add the `desktop` job after `release`, a manual dry run, and document the app in `README.md` and `AGENTS.md`.
  - **Design:** "The app bundle and release artifacts"
  - **Files:** `.github/workflows/release.yaml`, `README.md`, `AGENTS.md`
  - **Test first:** **Trivial:** the job runs only on a real release; YAML parsing and the script's dry-run test are the check.
  - **Definition of done:**
    - `release` runs on push only and exposes `release` and `tag` outputs; `desktop` needs it and runs on a released push or a manual dispatch.
    - On a push it builds at the new tag, runs `gh release upload`, and pushes the cask with the tap token and the message `kirokit-desktop <version>`; on a dispatch it only prints the cask.
    - The README describes the desktop app, `brew install --cask mavogel/tap/kirokit-desktop` and the unsigned-app note; `AGENTS.md` names `desktop/` in its layout.
    - verify: `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")' && bash scripts/desktop-release_test.sh`

**Validation:** automated: `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")'` · `bash scripts/desktop-release_test.sh` · `go test ./...` · manual: the first real tagged release publishes the zip and the cask

**Risky:** an external system: the release job pushes to your Homebrew tap and runs only on a real release.

## End-to-end scenarios

### Scenario 1: The CLI keeps one team installed and runs from any folder

**Runs after:** slice 1
**Proves:** Done means 11, 13

| Step | Action | Expected |
|---|---|---|
| 1 | `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` | exit code `0` |
| 2 | `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` | `installed platform-team into /tmp/ql-kk-s1/home` |
| 3 | `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` | two lines: `uninstalled platform-team from /tmp/ql-kk-s1/home`, then `installed data-team into /tmp/ql-kk-s1/home` |
| 4 | `ls /tmp/ql-kk-s1/home/.kirokit-installed` | `data-team.json` |
| 5 | `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` | `installed data-team into /tmp/ql-kk-s1/home` |
| 6 | `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` | `rendered platform-team -> /tmp/ql-kk-s1/out` |
| 7 | `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` | exit code `0` |
| 8 | `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` | exit code `0`; the output ends with a line starting `checked out v`; `ls /tmp/ql-kk-s1/home/.kirokit-installed` still shows only `data-team.json` |
| 9 | `/tmp/ql-kk-s1/kirokit sync --kb-agent x` | exit code `1`; the error says `kirokit sync: --kb-agent needs a team` |
| 10 | `go test ./...` | exit code `0` |

### Scenario 2: The window matches the design and remembers its settings

**Runs after:** slice 2
**Proves:** Done means 2, 10

| Step | Action | Expected |
|---|---|---|
| 1 | In a terminal run `rm -rf /tmp/ql-kk-s2 && mkdir -p /tmp/ql-kk-s2/home` then `env HOME=/tmp/ql-kk-s2/home task desktop:dev` in the background and wait for `http://localhost:34115` | the dev server answers |
| 2 | Open `http://localhost:34115` in a browser at 1280 by 800 | the logo, the sidebar with Teams, Agents, Install and sync, Skills and MCP versions, Settings, and the main pane |
| 3 | Click the sidebar toggle | the sidebar becomes an icon rail and the main pane is wider |
| 4 | Click the toggle again, then press Command-B twice | expanded, rail, expanded |
| 5 | Emulate `prefers-color-scheme: dark` | the page background and text switch to the dark tokens |
| 6 | Open Settings, set Kiro home to `/` and press Save | refused, with a message beside the field, and nothing saved |
| 7 | Set Kiro home to `/tmp/ql-kk-s2/kiro` and the repository address to `/tmp/ql-kk-s2/origin`, then press Save | a saved confirmation |
| 8 | Stop `task desktop:dev` and its `wails` and `vite` children by their process ids, check port 34115 is free, run it again with the same `HOME`, reload the page and open Settings | both values are still filled in |

### Scenario 3: The pages show the clone's real data and follow typed changes

**Runs after:** slice 4
**Proves:** Done means 3, 7, 9

| Step | Action | Expected |
|---|---|---|
| 1 | Run `rm -rf /tmp/ql-kk-s3 && mkdir -p "/tmp/ql-kk-s3/home/Library/Application Support/kirokit-desktop" /tmp/ql-kk-s3/kiro && git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit "/tmp/ql-kk-s3/home/Library/Application Support/kirokit-desktop/repo"`, write `settings.json` there with `{"RepoURL":"/Users/mavogel/Developer/projects/kiro-cli-kit","KiroHome":"/tmp/ql-kk-s3/kiro"}`, then run `env HOME=/tmp/ql-kk-s3/home GOPATH="$(go env GOPATH)" GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" npm_config_cache="$HOME/.npm" task desktop:dev` in the background | the app starts on `http://localhost:34115` |
| 2 | Open the page and visit Teams, Agents, Install and sync and Skills and MCP versions | Teams lists `data-team` and `platform-team`; Agents lists each agent with its tools; Skills and MCP versions shows the pins from `env/mcp-versions.yaml`; no team is marked installed |
| 3 | Open the terminal dock with its toggle and drag its top edge upward | the dock opens, grows, and the main pane shrinks |
| 4 | In the zsh tab type `command -v kirokit` and Enter | the path of the bundled `kirokit` |
| 4a | Drag the dock's top edge to a new height, then type `stty size` and Enter | rows and columns equal the dock's new size in the tab |
| 5 | In the zsh tab type `kirokit install data-team` and Enter | the output ends with `installed data-team into /tmp/ql-kk-s3/kiro`; within two seconds `data-team` shows installed on Teams and on Install and sync |
| 6 | In the zsh tab type `kirokit uninstall data-team` and Enter | `uninstalled data-team from /tmp/ql-kk-s3/kiro`; both pages show no installed team |
| 7 | Close the dock with its toggle and open it again | the zsh tab still shows the earlier output |

### Scenario 4: Buttons install a team, and the app writes what the command line writes

**Runs after:** slice 5
**Proves:** Done means 4, 5, 6, 8, 11, 14

| Step | Action | Expected |
|---|---|---|
| 1 | Run `rm -rf /tmp/ql-kk-s4 && mkdir -p /tmp/ql-kk-s4/home /tmp/ql-kk-s4/kiro && git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s4/origin && git -C /tmp/ql-kk-s4/origin tag v99.0.0 && : > /tmp/ql-kk-s4/keys.txt`, then `go build -o /tmp/ql-kk-s4/kirokit ./cmd/kirokit`, then `env HOME=/tmp/ql-kk-s4/home GOPATH="$(go env GOPATH)" GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" npm_config_cache="$HOME/.npm" SOPS_AGE_KEY_FILE=/tmp/ql-kk-s4/keys.txt task desktop:dev` in the background | the app starts |
| 2 | Open `http://localhost:34115` | the page "Set up kirokit" with an address field |
| 3 | Enter `-oops` and press Clone | refused with a message beside the field; no process starts |
| 4 | Enter `/tmp/ql-kk-s4/origin` and press Clone | the kirokit tab shows `git clone` output; then the Teams page lists `data-team` and `platform-team` |
| 5 | In Settings set Kiro home to `/tmp/ql-kk-s4/kiro` and Save | saved |
| 6 | On Teams press Install on `platform-team` | the setup dialog shows SOPS, Decryption key and Kiro CLI as found and Install is enabled |
| 7 | Press Install in the dialog | the kirokit tab shows `▸ kirokit install platform-team` and `installed platform-team into /tmp/ql-kk-s4/kiro`; Teams marks `platform-team` installed |
| 8 | Open Install and sync | it lists `platform-team` with its file count and Sync and Uninstall buttons |
| 9 | In the zsh tab type `kirokit install data-team` and Enter | two lines `uninstalled platform-team from /tmp/ql-kk-s4/kiro` and `installed data-team into /tmp/ql-kk-s4/kiro`; both pages show only `data-team` installed |
| 9a | Run `cp -R /tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/snapshot && env KIRO_HOME=/tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/kirokit uninstall data-team && env KIRO_HOME=/tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/kirokit install data-team && diff -r /tmp/ql-kk-s4/snapshot /tmp/ql-kk-s4/kiro` | exit code `0`: the oracle holds, the app's install and a plain command-line install write the same bytes |
| 10 | On Teams press Install on `platform-team` | the replace dialog names `data-team` and `platform-team`; press Cancel and nothing runs |
| 11 | Press Install again and confirm the replace dialog, then the setup dialog | the kirokit tab shows both `uninstalled data-team` and `installed platform-team` lines |
| 12 | On Install and sync press Sync | the tab shows `▸ kirokit sync platform-team --latest-tag` and a line starting `synced platform-team into /tmp/ql-kk-s4/kiro at` |
| 13 | Run `rm -rf /tmp/ql-kk-s4/snapshot && cp -R /tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/snapshot && env KIRO_HOME=/tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/kirokit uninstall platform-team && env KIRO_HOME=/tmp/ql-kk-s4/kiro /tmp/ql-kk-s4/kirokit install platform-team && diff -r /tmp/ql-kk-s4/snapshot /tmp/ql-kk-s4/kiro` | exit code `0`: what the Sync button left is what a plain install writes |
| 14 | On Install and sync press Uninstall for `platform-team` | the tab shows `uninstalled platform-team from /tmp/ql-kk-s4/kiro`; no team is marked installed |
| 15 | Run `git -C /tmp/ql-kk-s4/origin commit -q --allow-empty -m "feat: release" && git -C /tmp/ql-kk-s4/origin tag v99.1.0 && cp -R /tmp/ql-kk-s4/origin/teams/data-team /tmp/ql-kk-s4/origin/teams/qa-team && git -C /tmp/ql-kk-s4/origin add -A && git -C /tmp/ql-kk-s4/origin commit -q -m "feat: qa team" && git -C /tmp/ql-kk-s4/origin tag v99.2.0` and press Update in Settings | the tab shows `▸ kirokit sync --latest-tag` and `checked out v99.2.0 at`; Teams now lists `qa-team` |

### Scenario 5: One command builds the universal app, and CI covers it

**Runs after:** slice 6
**Proves:** Done means 12

| Step | Action | Expected |
|---|---|---|
| 1 | `task desktop:build` | exit code `0` |
| 2 | `lipo -archs desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` | `x86_64 arm64` |
| 3 | `lipo -archs desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit` | `x86_64 arm64` |
| 4 | `desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit version` | exit code `0` |
| 5 | `ruby -ryaml -e 'puts YAML.load_file(".github/workflows/desktop.yaml")["jobs"].values.first["runs-on"]'` | `macos-latest` |
| 6 | `go build ./...` | exit code `0`: the root build still leaves out `desktop/` |

### Scenario 6: The release job's dry run builds the zip and the cask

**Runs after:** slice 7
**Proves:** Done means 15, 16

| Step | Action | Expected |
|---|---|---|
| 1 | `bash scripts/desktop-release_test.sh` | exit code `0` |
| 2 | `bash scripts/desktop-release.sh --version 0.0.0 --app desktop/build/bin/kirokit.app --out /tmp/ql-kk-s6` | exit code `0`; the output names `kirokit-desktop_0.0.0_darwin_universal.zip`, the sha256 and `Casks/kirokit-desktop.rb`; it says nothing was uploaded or pushed |
| 3 | `unzip -l /tmp/ql-kk-s6/kirokit-desktop_0.0.0_darwin_universal.zip` | the first entries are under `kirokit.app/` |
| 4 | `ruby -ryaml -e 'y=YAML.load_file(".github/workflows/release.yaml"); puts y["jobs"]["desktop"]["needs"]'` | `release` |

## Oracle

- **Oracle:** The app runs with a temporary home and Kiro home folder. It clones this repository from a local path, installs platform-team through the setup dialog, and a typed `kirokit install data-team` in its zsh tab then replaces it on both pages. The files in that Kiro home then match, byte for byte, what a plain `kirokit install data-team` writes into the same folder after a `kirokit uninstall data-team`.

## Decided while outlining

- **Checked for gaps: what prevents it, cancel paths, boundary inputs, two at once** · 6 scenario steps and 6 definition-of-done bullets added: a refused address, a cancelled replace dialog, Kiro home `/`, an empty `KIROKIT_REPO`, a second command while one runs, and a clone that fails or has no release tag.
- **The root constants stay and `repoPath` wraps them** — the Plan said to replace them, but keeping the names leaves every existing test untouched and behaves the same.
- **No generated Wails bindings are committed** — the frontend calls `window.go.main.App` through `api.js`, so `npm run build` works before any Wails run and CI needs no generation step.
- **Frontend logic is tested with node's built-in test runner** — it adds no test dependency; rendered behaviour is checked in the browser through `wails dev`.
- **A new address renames the old clone to `repo.old-<time>`** — the app keeps one fixed `repo` path as the Plan's contracts say, and still deletes nothing.
- **The Outline is 7 slices and 24 tasks, as sized** — no more.
- **Scenario 6 proves the release only as far as a dry run can** — the real upload, cask push and Homebrew install need a tagged release and stay yours to confirm.

## How each point is proved

<!-- ql:proof-map -->
- Done means 1 · The kirokit desktop app opens a window that matches the design → only you can confirm it
- Done means 2 · The sidebar collapses to an icon rail and expands again → "The window matches the design and remembers its settings" · runs after slice 2
- Done means 3 · Teams, Agents, and Skills and MCP versions show the repository's real data → "The pages show the clone's real data and follow typed changes" · runs after slice 4
- Done means 4 · A non-engineer installs a team with buttons only → "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 5 · Install writes the same files as `kirokit install` → "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 6 · Install and sync shows the real install state → "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 7 · A real terminal docks at the bottom, as in VS Code → "The pages show the clone's real data and follow typed changes" · runs after slice 4
- Done means 8 · Each button press appears in the terminal as its `kirokit` command → "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 9 · A command typed in the terminal updates the screens → "The pages show the clone's real data and follow typed changes" · runs after slice 4
- Done means 10 · Settings sets the Kiro home folder and the repository's address → "The window matches the design and remembers its settings" · runs after slice 2
- Done means 11 · Only one team is ever installed at a time, in the app and the CLI → "The CLI keeps one team installed and runs from any folder" · runs after slice 1; "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 12 · One command builds the macOS app, and CI builds it too → "One command builds the universal app, and CI covers it" · runs after slice 6
- Done means 13 · The kirokit command line keeps working, apart from the one-team rule → "The CLI keeps one team installed and runs from any folder" · runs after slice 1
- Done means 14 · The app fetches the templates itself, with no git knowledge needed → "Buttons install a team, and the app writes what the command line writes" · runs after slice 5
- Done means 15 · Every kirokit release also ships the macOS app → "The release job's dry run builds the zip and the cask" · runs after slice 7
- Done means 16 · A separate `kirokit-desktop` Homebrew cask installs the app, and it opens → "The release job's dry run builds the zip and the cask" · runs after slice 7
<!-- /ql:proof-map -->

## Checked by a cold read

<!-- ql:reviews -->
**Cold read** · no gaps: the Outline can be built as it is.
<!-- /ql:reviews -->

## Commands the build runs

<!-- ql:commands -->
- `go build ./...` · slice 1, slice 2, scenario 5, project checks
- `go vet ./...` · slice 1, project checks
- `go test ./...` · slice 1, slice 5, slice 6, slice 7, T5, scenario 1, project checks
- `golangci-lint run` · slice 1, project checks
- `task desktop:test` · slice 2, slice 3, slice 4, slice 5, T6, T8, T20
- `task desktop:build` · slice 6, T21, scenario 5, project checks
- `ruby -ryaml -e 'YAML.load_file(".github/workflows/desktop.yaml")'` · slice 6, T22
- `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")'` · slice 7
- `bash scripts/desktop-release_test.sh` · slice 7, T23, scenario 6
- `go test ./cmd/kirokit/commands/ ./internal/golden/` · T1
- `go test ./internal/install/` · T2
- `go test ./cmd/kirokit/commands/ -run Install` · T3
- `go test ./cmd/kirokit/commands/ -run Sync` · T4
- `cd desktop/frontend && node --test src/ && npm run build` · T7, T12, T13, T18, T19
- `cd desktop && go test -run 'Capture|Prereq' ./...` · T9
- `cd desktop && go test -run Catalog ./...` · T10
- `cd desktop && go test -race -run Installs ./...` · T11
- `cd desktop && go test -race ./terminal/` · T14
- `cd desktop && go test -race -run Shell ./...` · T15
- `cd desktop && go test -race -run Runner ./...` · T16
- `cd desktop && go test -race -run Clone ./...` · T17
- `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")' && bash scripts/desktop-release_test.sh` · T24
- `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` · scenario 1
- `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` · scenario 1
- `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · scenario 1
- `ls /tmp/ql-kk-s1/home/.kirokit-installed` · scenario 1
- `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` · scenario 1
- `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` · scenario 1
- `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` · scenario 1
- `/tmp/ql-kk-s1/kirokit sync --kb-agent x` · scenario 1
- `lipo -archs desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` · scenario 5
- `lipo -archs desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit` · scenario 5
- `desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit version` · scenario 5
- `ruby -ryaml -e 'puts YAML.load_file(".github/workflows/desktop.yaml")["jobs"].values.first["runs-on"]'` · scenario 5
- `bash scripts/desktop-release.sh --version 0.0.0 --app desktop/build/bin/kirokit.app --out /tmp/ql-kk-s6` · scenario 6
- `unzip -l /tmp/ql-kk-s6/kirokit-desktop_0.0.0_darwin_universal.zip` · scenario 6
- `ruby -ryaml -e 'y=YAML.load_file(".github/workflows/release.yaml"); puts y["jobs"]["desktop"]["needs"]'` · scenario 6
- `node scripts/sync-agent-assets.mjs --check` · project checks
- `cd desktop/frontend && npm ci && npm run build` · project checks
- `cd desktop && go vet ./... && go test ./...` · project checks
<!-- /ql:commands -->
