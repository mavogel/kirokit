# Desktop app from design

<!-- ql:links -->
Task: `docs/plans/2026-10-17-desktop-app-from-design/`
<!-- /ql:links -->

## Why the change

People without a terminal can now install and update a team's Kiro config from a macOS app that runs the real `kirokit` commands, and each release can ship that app through its own Homebrew cask.

## What a reviewer must know

- **`kirokit install` and `kirokit sync` now keep one team per machine.** Installing or syncing a team first uninstalls any other installed team and prints a line for it. Two scheduled syncs of different teams on one machine now replace each other, which the README's migration note explains; the commit type needs a `!` so the release notes flag it.
- **The release job has never run for real.** Uploading the app, pushing the `kirokit-desktop` cask to your tap and `brew install` can only be tried by a tagged release. The workflow and the zip script were checked locally in dry-run mode only.
- **Two small gaps remain in Settings.** A Kiro home change saved in the same click as a failed address clone is saved without a confirmation. If the settings file cannot be written after a successful clone, the app reports a failure while the new clone is already in use; that path has no test. Follow-up, by your decision: the clone folder is fixed at `~/Library/Application Support/kirokit-desktop/repo` and is not configurable yet.

## What changed

The app is a separate Go module that reuses kirokit's own packages and bundles the CLI, so buttons and typed commands run the same code.

```
desktop/                      new Wails v2 module, outside the root build
├── app.go, settings.go       bound service, settings.json
├── catalog.go, prereq.go     teams, agents, skills, pins; SOPS, key, Kiro CLI checks
├── clone.go, runner.go       first-launch clone; one kirokit command at a time
├── shell.go, terminal/       zsh tab through a ZDOTDIR wrapper; pty sessions
└── frontend/                 Vite, xterm.js, pages and dialogs
```

The CLI changed so the app can drive it from any folder and keep one team:

```diff
 kirokit install <team>
+  for each other installed team: uninstall it, print "uninstalled <other> from <home>"
+  (only after the new team renders, so a mistyped name removes nothing)
   install <team>, print "installed <team> into <home>"
+kirokit sync                 (no team: only update the checkout)
+KIROKIT_REPO=<folder>        roots teams/, global/, env/ for every command
```

Releases gain a macOS job after the CLI release: build, zip without AppleDouble files, upload, write the cask.

## Screens compared

<!-- ql:screens -->
- **desktop-app-build** · 1200 px · light · captured · current
  - Polish: No comparison recorded.
- **desktop-app-build** · 1200 px · dark · captured · current
  - Polish: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
7 slices · 24 tasks · 31 commits · 77 files · +7287 −42

**Slice 1 · One team stays installed and kirokit runs from any folder** · T1, T2, T3, T4, T5 · 18 files · +778 −41

- `README.md` · T5, T20, T24 · +20 −2
- `cmd/kirokit/commands/install.go` · T1, T3 · +22 −3
- `cmd/kirokit/commands/install_test.go` · T3 · +86 −0
- `cmd/kirokit/commands/onboard.go` · T1 · +3 −2
- `cmd/kirokit/commands/render.go` · T1 · +1 −1
- `cmd/kirokit/commands/root.go` · T1 · +28 −5
- `cmd/kirokit/commands/root_test.go` · T1 · new · +150 −0
- `cmd/kirokit/commands/skills.go` · T1 · +2 −2
- `cmd/kirokit/commands/sync.go` · T1, T4 · +62 −15
- `cmd/kirokit/commands/sync_test.go` · T4 · +143 −1
- `cmd/kirokit/commands/test.go` · T1 · +4 −4
- `cmd/kirokit/commands/update.go` · T1 · +4 −4
- `internal/externalskills/fetch.go` · T4 · +19 −0
- `internal/externalskills/fetch_test.go` · T4 · +2 −1
- `internal/install/install.go` · T1, T3 · +12 −0
- `internal/install/installed.go` · T2 · new · +75 −0
- `internal/install/installed_test.go` · T2 · new · +143 −0
- `scripts/kirokit-sync-setup.sh` · T5 · +2 −1

**Slice 2 · The app opens with the design's shell and remembers its settings** · T6, T7, T8 · 23 files · +2634 −0

- `Taskfile.yml` · T6, T21 · +18 −0
- `desktop/.gitignore` · T6, T9, T21 · new · +9 −0
- `desktop/app.go` · T6, T8, T9, T10, T11, T15, T16, T17, T20 · new · +131 −0
- `desktop/app_test.go` · T6 · new · +21 −0
- `desktop/build/appicon.png` · T6 · new · +0 −0
- `desktop/build/darwin/Info.dev.plist` · T6 · new · +32 −0
- `desktop/build/darwin/Info.plist` · T6 · new · +27 −0
- `desktop/frontend/index.html` · T6, T7, T9, T13, T19 · new · +120 −0
- `desktop/frontend/package-lock.json` · T6, T13 · new · +1183 −0
- `desktop/frontend/package.json` · T6, T7, T13 · new · +19 −0
- `desktop/frontend/src/api.js` · T7, T9, T12, T15, T18 · new · +57 −0
- `desktop/frontend/src/main.js` · T7, T8, T12, T13, T18, T19 · new · +192 −0
- `desktop/frontend/src/pages/settings.js` · T8, T9, T20 · new · +130 −0
- `desktop/frontend/src/sidebar.js` · T7 · new · +38 −0
- `desktop/frontend/src/sidebar.test.js` · T7 · new · +66 −0
- `desktop/frontend/src/style.css` · T7, T13, T19 · new · +150 −0
- `desktop/frontend/vite.config.js` · T6, T13 · new · +24 −0
- `desktop/go.mod` · T6, T9, T14 · new · +43 −0
- `desktop/go.sum` · T6, T9, T14 · new · +92 −0
- `desktop/main.go` · T6, T9 · new · +35 −0
- `desktop/settings.go` · T8, T9, T17 · new · +122 −0
- `desktop/settings_test.go` · T8 · new · +105 −0
- `desktop/wails.json` · T6 · new · +20 −0

**Slice 3 · The pages show the repository's real data** · T9, T10, T11, T12 · 15 files · +1575 −0

- `desktop/catalog.go` · T10 · new · +269 −0
- `desktop/catalog_test.go` · T10 · new · +205 −0
- `desktop/envcapture.go` · T9 · new · +126 −0
- `desktop/envcapture_test.go` · T9 · new · +84 −0
- `desktop/frontend/dist/.gitkeep` · T9 · new · +0 −0
- `desktop/frontend/src/pages/agents.js` · T12 · new · +34 −0
- `desktop/frontend/src/pages/html.js` · T12, T18 · new · +50 −0
- `desktop/frontend/src/pages/install.js` · T12, T20 · new · +46 −0
- `desktop/frontend/src/pages/pages.test.js` · T12, T18 · new · +247 −0
- `desktop/frontend/src/pages/skills.js` · T12 · new · +49 −0
- `desktop/frontend/src/pages/teams.js` · T12, T19 · new · +37 −0
- `desktop/installs.go` · T11 · new · +103 −0
- `desktop/installs_test.go` · T11 · new · +134 −0
- `desktop/prereq.go` · T9 · new · +62 −0
- `desktop/prereq_test.go` · T9 · new · +129 −0

**Slice 4 · A real terminal docks at the bottom** · T13, T14, T15, T16 · 9 files · +1380 −0

- `desktop/frontend/src/tabs.js` · T13 · new · +50 −0
- `desktop/frontend/src/tabs.test.js` · T13 · new · +62 −0
- `desktop/frontend/src/terminal.js` · T13, T15, T16 · new · +152 −0
- `desktop/runner.go` · T16, T17, T20 · new · +189 −0
- `desktop/runner_test.go` · T16, T20 · new · +234 −0
- `desktop/shell.go` · T15 · new · +307 −0
- `desktop/shell_test.go` · T15 · new · +198 −0
- `desktop/terminal/terminal.go` · T14 · new · +87 −0
- `desktop/terminal/terminal_test.go` · T14 · new · +101 −0

**Slice 5 · A person installs a team with buttons only** · T17, T18, T19, T20 · 7 files · +589 −1

- `AGENTS.md` · T20, T24 · +2 −1
- `desktop/clone.go` · T17 · new · +106 −0
- `desktop/clone_test.go` · T17 · new · +262 −0
- `desktop/frontend/src/api.test.js` · T18 · new · +44 −0
- `desktop/frontend/src/dialogs.js` · T19 · new · +89 −0
- `desktop/frontend/src/dialogs.test.js` · T19 · new · +47 −0
- `desktop/frontend/src/pages/welcome.js` · T18 · new · +39 −0

**Slice 6 · One command builds the macOS app and CI builds it too** · T21, T22 · 2 files · +85 −0

- `.github/workflows/desktop.yaml` · T22 · new · +44 −0
- `scripts/desktop-build.sh` · T21 · new · +41 −0

**Slice 7 · Each release attaches the app and a Homebrew cask** · T23, T24 · 3 files · +246 −0

- `.github/workflows/release.yaml` · T24 · +59 −0
- `scripts/desktop-release.sh` · T23 · new · +98 −0
- `scripts/desktop-release_test.sh` · T23 · new · +89 −0
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. The kirokit desktop app opens a window that matches the design.** Only you can confirm: it looks like the approved design · waits for your confirmation
- **2. The sidebar collapses to an icon rail and expands again.** Scenario 2: The window matches the design and remembers its settings, after slice 2 · passed · `evidence/it2-scenario-2.txt`
- **3. Teams, Agents, and Skills and MCP versions show the repository's real data.** Scenario 3: The pages show the clone's real data and follow typed changes, after slice 4 · passed · `evidence/it2-scenario-3.txt`
- **4. A non-engineer installs a team with buttons only.** Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-scenario-4-step-6.png`
- **5. Install writes the same files as `kirokit install`.** Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-oracle.txt`
- **6. Install and sync shows the real install state.** Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-scenario-4.txt`
- **7. A real terminal docks at the bottom, as in VS Code.** Scenario 3: The pages show the clone's real data and follow typed changes, after slice 4 · passed · `evidence/it2-scenario-3.txt`
- **8. Each button press appears in the terminal as its `kirokit` command.** Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-scenario-4.txt`
- **9. A command typed in the terminal updates the screens.** Scenario 3: The pages show the clone's real data and follow typed changes, after slice 4 · passed · `evidence/it2-scenario-4-step-9.png`
- **10. Settings sets the Kiro home folder and the repository's address.** Scenario 2: The window matches the design and remembers its settings, after slice 2 · passed · `evidence/it2-scenario-2.txt`
- **11. Only one team is ever installed at a time, in the app and the CLI.** Scenario 1: The CLI keeps one team installed and runs from any folder, after slice 1 · Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-scenarios-1-5-6.txt`
- **12. One command builds the macOS app, and CI builds it too.** Scenario 5: One command builds the universal app, and CI covers it, after slice 6 · passed · `evidence/it2-scenarios-1-5-6.txt`
- **13. The kirokit command line keeps working, apart from the one-team rule.** Scenario 1: The CLI keeps one team installed and runs from any folder, after slice 1 · passed · `evidence/it2-scenarios-1-5-6.txt`
- **14. The app fetches the templates itself, with no git knowledge needed.** Scenario 4: Buttons install a team, and the app writes what the command line writes, after slice 5 · passed · `evidence/it2-scenario-4.txt`
- **15. Every kirokit release also ships the macOS app.** Only you can confirm: a real tagged release publishes it · waits for your confirmation
- **16. A separate `kirokit-desktop` Homebrew cask installs the app, and it opens.** Only you can confirm: the cask lands in your tap · waits for your confirmation
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No change to the install library's own behaviour.** The one-team rule stays in the commands.
- **No signing, notarization or Linux app.** The cask clears the quarantine flag instead.
- **No change to the existing `kirokit` cask or CLI archives.**
<!-- /ql:not-doing -->

## How it was verified

<!-- ql:checked -->
Taken as it stood with 1 item still open; lanes that ran: Polish, Security review, Project checks, Code review, Second opinion and Fix rounds.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The CLI keeps one team installed
- Do: install one team, then another, into a scratch Kiro home.
- Run: `KIRO_HOME=/tmp/try kirokit install platform-team && KIRO_HOME=/tmp/try kirokit install data-team`
- Expect: the second command prints `uninstalled platform-team from /tmp/try` and then `installed data-team into /tmp/try`.
- Saw: both lines, and a mistyped team left the installed one in place — `evidence/it2-scenarios-1-5-6.txt`, `evidence/cp-1-run-3-boundary.txt`

#### 2. Buttons install a team and match the command line
- Do: run `task desktop:dev`, enter a repository address, clone, then install a team from Teams.
- Expect: the kirokit tab shows `▸ kirokit install <team>` and the Teams card turns to installed; the Kiro home matches a plain `kirokit install`.
- Saw: the buttons, a typed install in the zsh tab and a byte-for-byte `diff -r` all matched — `evidence/it2-scenario-4.txt`, `evidence/it2-oracle.txt`

#### 3. A new address in Settings clones fresh
- Do: save a different valid repository address in Settings.
- Expect: the old clone is kept as `repo.old-<time>` and Teams shows the new teams; a missing address shows the git error beside the field.
- Saw: both cases in the running app — `evidence/polish4-settings-clone-failed.png`, `evidence/it3-scenario-2.txt`

### Not verified

- Whether the window looks like the approved design (point 1), a real tagged release and the Homebrew cask landing in your tap (points 15 and 16): these are yours to confirm.
- The page screenshots through QualityLayer's capture tool show only the app shell; the other pages were checked in a normal browser tab.
- Scenario 2 step 7 was amended after the last check run and not re-judged: it now creates the new address's origin first.
- The second opinion from Codex failed twice and did not run.

### Docs

- `README.md` covers the desktop app, `KIROKIT_REPO`, team-less sync, the one-team rule and its migration note; `AGENTS.md` names `desktop/`.

<!-- optional section here, only when there is content: "Security" -->

## Accepted with open items

<!-- ql:accepted -->
Checking ran twice without settling, so Verify ended and the build went to review. What was left open:

- Judge, round 3: [FAIL] scenario 2: Step 7 expects a saved confirmation, but its address does not exist, so none shows. Step 1 removes `/tmp/ql-kk-s2`, so `/tmp/ql-kk-s2/origin` is missing; since Save clones a new address, the app now rightly shows "git clone failed (exit 128)…" beside the field with an empty status (Kiro home is saved silently). The scenario was not amended for the clone-on-save behaviour. Everything else held: first-launch page, clone, toggle and Command-B 232 -> 60 -> 232 px, Control-B nothing, dark tokens, `/` refused with nothing saved; a retry Save ran a second `▸ git clone`; with a valid origin created at that path Save showed "Saved. Cloning…" then "Cloned. The new repository is in use." and stored it; after a restart by pid both values were filled in. Amend step 7 (create the origin first, or expect the failure beside the field) — `evidence/it3-scenario-2.txt`, `evidence/it3-scenario-2-step-7-failed.png`
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- The app is named kirokit, with bundle id `io.github.mavogel.kirokit` — the design already says kirokit, and the cask name tells it apart from the CLI. (Plan)
- The app bundles its own kirokit CLI, first on its terminals' PATH — typed commands match the app's version, with no separate install. (Plan)
- Its clone and settings live in `~/Library/Application Support/kirokit-desktop/` — macOS's place for an app's own data. (Plan)
- A new repository address clones into a fresh folder and keeps the old one — nothing is deleted without asking. (Plan)
- The frontend is plain JavaScript with Vite and xterm.js 6 — the design is already plain HTML. (Plan)
- Agents are read by rendering each team into a temporary folder — their tools and servers exist only after rendering. (Plan)
- Last change is the install record's file time — kirokit's record holds no time of its own. (Plan)
- Wails runs through `go run` at a pinned version, and the minimum macOS is 13 — nobody installs a Wails command, and Go 1.27 needs macOS 13. (Plan)
- Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes. (slice 1)
- Installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes. (slice 1)
- A mistyped team or a failed update in sync removed the installed team first, so sync now updates and checks the new team before removing anything; nothing you approved changes. (slice 1)
- The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes. (slice 1)
- App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes. (the Plan, while building)
- Startup hooks, the module requirements, a Settings note, shared page helpers and ignored generated files are needed by the data slice; nothing you approved changes. (the Plan, while building)
- A placeholder file in the ignored build folder lets a clean checkout compile the app; nothing you approved changes. (slice 3)
- The build keeps the dist placeholder and the dock markup lives in the page; nothing you approved changes. (slice 4)
- Clone logic, first-launch page, dialogs, docs and a keystroke queue need files no card named; oracle commands now set the repository folder. (the Plan, while building)
- Scenario 2 now clones first, because the pages stay hidden until a clone exists; nothing you approved changes. (slice 2)
- Scenario 2 now creates the new address's origin first, because saving an address clones it; nothing you approved changes. (slice 2)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
Nothing was left for you to settle.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Approved · 18 Oct, 02:49. Branch `ql/desktop-app-from-design`.
<!-- /ql:ship -->
