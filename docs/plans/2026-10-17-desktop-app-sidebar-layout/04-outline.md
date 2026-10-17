---
type: outline
---

# Outline: Desktop app sidebar layout

## Summary

One slice with three tasks finishes the clickable `sidebar-layout` design page. The order is the five pages and the sidebar rail, then the terminal, then the one-team setup flow. Nothing is risky: the page is plain HTML and no kirokit code changes.

## How the build runs

<!-- ql:waves -->
One slice, built on its own.
<!-- /ql:waves -->

## Program design

The Plan's contracts hold the design.

## Shape of the code

No new module.

## Slice 1: A complete clickable kirokit window with sidebar, terminal and one-team setup

After this slice a person opens one page, moves through all five sections, collapses the sidebar and sets up a team through the prerequisites check. Every button prints its `kirokit` command in the docked terminal, and typed commands change the screens too.

```diff
 docs/plans/2026-10-17-desktop-app-sidebar-layout/design/
~└── sidebar-layout.html   ~ five pages with repository data, rail and shortcuts,
                            terminal contract, one-team install state, replace dialog
```

- [x] T1 Five pages show repository data, and the sidebar collapses to a rail
  - **Objective:** Check every name, description, tool, skill and version on the five pages against the repository and fix differences. Make the rail, its tooltips and Command-B match the Plan.
  - **Design:** none
  - **Files:** `docs/plans/2026-10-17-desktop-app-sidebar-layout/design/sidebar-layout.html`
  - **Test first:** Scenario 1 step 5 compares the Agents and MCP rows with the four agent files' lines 3-5 (read-only, for example `teams/platform-team/agents/devops/agent.json.tmpl:3-5`) and `env/mcp-versions.yaml:3-13`. Any difference found in the draft fails it.
  - **Definition of done:**
    - Each of the five sidebar buttons shows only its own page, and the current button has `aria-current="page"`.
    - Rail mode is 60 pixels wide; each icon keeps its name as tooltip and accessible name; Command-B and the toggle switch it both ways.
    - Teams, Agents, Skills and MCP counts equal the repository counts.
    - verify: `qualitylayer design check sidebar-layout`

- [x] T2 The terminal prints kirokit's exact lines and resizes within limits
  - **Objective:** Make typed commands follow the Plan's output table, and keep the terminal's open, close and resize behaviour.
  - **Design:** "The terminal echoes each command and prints kirokit's exact output format"
  - **Files:** `docs/plans/2026-10-17-desktop-app-sidebar-layout/design/sidebar-layout.html`
  - **Test first:** Scenario 3 step 6 types `kirokit version` and expects a `commit: 9a919bf…` line. The draft prints a different commit line, so it fails.
  - **Definition of done:**
    - Each table row prints its stated line; `kirokit test` with no team prints one `<team>: ok` per team.
    - An unknown team prints an error line, and an unknown word prints `zsh: command not found: <word>`.
    - Control-backtick hides and shows the terminal; the input regains focus after each command; the output region has `aria-live="polite"`.
    - The resize handle stops at 120 and 480 pixels, by drag and by arrow keys.
    - verify: `qualitylayer design check sidebar-layout`

- [x] T3 Only one team is ever installed, and setup checks prerequisites first
  - **Objective:** Route every install, uninstall and sync through one `setInstalled` that keeps at most one team `true`. Add the replace question and the prerequisites dialog with focus return.
  - **Design:** "Install state is one map that both pages render", "The setup dialog is a native modal with focus return"
  - **Files:** `docs/plans/2026-10-17-desktop-app-sidebar-layout/design/sidebar-layout.html`
  - **Test first:** Scenario 3 step 3 presses Install on platform-team while data-team is installed and expects the replace question. The draft shows two Installed badges instead, so it fails.
  - **Definition of done:**
    - Install on a second team asks "Replace data-team with platform-team?"; Cancel changes nothing, and Replace runs `kirokit uninstall data-team` then `kirokit install platform-team`.
    - A typed install of another team does the same without the question.
    - Install in the dialog stays disabled with "1 prerequisite missing" until Check again; Escape and Cancel close it and return focus to the opener.
    - Teams and Install and sync show the same badges after every action.
    - verify: `qualitylayer design check sidebar-layout`

**Validation:** automated: `qualitylayer design check sidebar-layout` · `test -z "$(git status --porcelain -- cmd internal teams global env testdata assets)"` · `go test ./...` · manual: none

## End-to-end scenarios

### Scenario 1: Open the window, visit all five pages and collapse the sidebar

**Runs after:** slice 1
**Proves:** Done means 1, 2, 3, 6

| Step | Action | Expected |
|---|---|---|
| 1 | `qualitylayer design check sidebar-layout` | exit code `0` |
| 2 | Open `design/sidebar-layout.html` in a browser at 1440 pixels wide | A window shows the kirokit logo, five sidebar sections, the Teams page and the terminal at the bottom |
| 3 | Click each sidebar section in turn | Teams, Agents, Install and sync, Skills and MCP versions and Settings each show only their own page |
| 4 | Click the Collapse button, then press Command-B | The sidebar becomes a 60-pixel rail with five icons and the main pane widens; Command-B expands it again |
| 5 | Compare the Agents and MCP versions pages with each team's agent files (lines 3-5) and `env/mcp-versions.yaml:3-13` | Every agent description, tool and pinned version matches the repository word for word |

### Scenario 2: Set up a team through the prerequisites check

**Runs after:** slice 1
**Proves:** Done means 4, 5

| Step | Action | Expected |
|---|---|---|
| 1 | On Teams, press Install on platform-team after uninstalling data-team (`kirokit uninstall data-team` in the terminal) | A dialog lists SOPS and Kiro CLI as found and the decryption key as missing |
| 2 | Look at the dialog's Install button | It is disabled and the dialog reads "1 prerequisite missing" |
| 3 | Press Check again, then Install | The button reads "Installing…", then the dialog closes |
| 4 | Read the terminal | It shows `kirokit install platform-team` and `installed platform-team into /Users/you/.kiro` |
| 5 | Look at Teams and Install and sync | platform-team is marked Installed on both pages |
| 6 | Reopen the dialog and press Escape | The dialog closes and focus is on the button that opened it |

### Scenario 3: Only one team is installed at a time

**Runs after:** slice 1
**Proves:** Done means 5, 9

| Step | Action | Expected |
|---|---|---|
| 1 | Reload the page | data-team is Installed and platform-team is Not installed |
| 2 | On Teams, press Install on platform-team | The dialog asks "Replace data-team with platform-team?" |
| 3 | Press Cancel | Nothing changes and the terminal prints no new line |
| 4 | Press Install again, then Replace, finish the prerequisites check and press Install | The terminal prints `uninstalled data-team from /Users/you/.kiro`, then `installed platform-team into /Users/you/.kiro` |
| 5 | Look at both pages | Only platform-team shows Installed |
| 6 | Type `kirokit install data-team` in the terminal | The terminal prints the uninstall line for platform-team, then `installed data-team into /Users/you/.kiro`; only data-team shows Installed |
| 7 | Type `kirokit uninstall data-team`, then `kirokit install nope` | The first marks data-team Not installed on both pages; the second prints an error line and changes nothing |
| 8 | Type `kirokit version` | The terminal prints `kirokit v0.4.0` and a `commit: 9a919bf…` line |
| 9 | Drag the terminal's top edge up and down past both limits, then press Control-backtick twice | Its height stops at 120 and 480 pixels, and it hides and shows again |

### Scenario 4: The window reads well in light and dark

**Runs after:** slice 1
**Proves:** Done means 7

| Step | Action | Expected |
|---|---|---|
| 1 | `qualitylayer design shot sidebar-layout --width 1440 --theme light` | A picture file is written and shows the Teams page with readable text |
| 2 | `qualitylayer design shot sidebar-layout --width 1440 --theme dark` | A picture file is written and shows the same page with dark surfaces and a visible violet accent |

### Scenario 5: Nothing outside this task changes

**Runs after:** slice 1
**Proves:** Done means 8

| Step | Action | Expected |
|---|---|---|
| 1 | `test -z "$(git status --porcelain -- cmd internal teams global env testdata assets)"` | exit code `0` |
| 2 | `go test ./...` | exit code `0` |

## Oracle

- **Oracle:** In a browser, one scripted run of the design page collapses and expands the sidebar, opens all five pages, installs platform-team through the prerequisites check and sees the terminal print `installed platform-team into /Users/you/.kiro`. Then a typed `kirokit uninstall platform-team` flips both badges back, and at no point do two teams show Installed, while `git status` shows no change outside this task's documents.

## Decided while outlining

- **Checked for gaps: what prevents it, cancel paths, boundary inputs, two at once** · 3 scenario steps added: cancel in the replace question, an unknown team name, and the terminal's size limits.
- **One slice holds all three tasks.** The Plan sizes it at one slice, and all three edit the same page, so they build in order.
- **The tasks are checked by the page validator and the scenarios.** The page has no test framework, and adding one would add a dependency the Plan rules out.

## How each point is proved

<!-- ql:proof-map -->
- Done means 1 · The mockup opens full size in the QualityLayer App → "Open the window, visit all five pages and collapse the sidebar" · runs after slice 1
- Done means 2 · The sidebar lists all five sections, and each opens its page → "Open the window, visit all five pages and collapse the sidebar" · runs after slice 1
- Done means 3 · The sidebar collapses to an icon rail and expands again → "Open the window, visit all five pages and collapse the sidebar" · runs after slice 1
- Done means 4 · A non-engineer can click through setting up a team → "Set up a team through the prerequisites check" · runs after slice 1
- Done means 5 · Engineers get a full terminal docked at the bottom → "Set up a team through the prerequisites check" · runs after slice 1; "Only one team is installed at a time" · runs after slice 1
- Done means 6 · The mockup shows kirokit's real teams, agents and versions → "Open the window, visit all five pages and collapse the sidebar" · runs after slice 1
- Done means 7 · The mockup reads well in light and dark mode → "The window reads well in light and dark" · runs after slice 1
- Done means 8 · kirokit's code and behaviour stay unchanged → "Nothing outside this task changes" · runs after slice 1
- Done means 9 · Only one team is installed at a time → "Only one team is installed at a time" · runs after slice 1
<!-- /ql:proof-map -->

## Checked by a cold read

<!-- ql:reviews -->
Not needed: one code slice and nothing risky, so the Outline was not read cold.
<!-- /ql:reviews -->

## Commands the build runs

<!-- ql:commands -->
- `qualitylayer design check sidebar-layout` · slice 1, T1, T2, T3, scenario 1
- `test -z "$(git status --porcelain -- cmd internal teams global env testdata assets)"` · slice 1, scenario 5
- `go test ./...` · slice 1, scenario 5, project checks
- `qualitylayer design shot sidebar-layout --width 1440 --theme light` · scenario 4
- `qualitylayer design shot sidebar-layout --width 1440 --theme dark` · scenario 4
- `golangci-lint run` · project checks
- `go vet ./...` · project checks
- `go build ./...` · project checks
- `node scripts/sync-agent-assets.mjs --check` · project checks
<!-- /ql:commands -->
