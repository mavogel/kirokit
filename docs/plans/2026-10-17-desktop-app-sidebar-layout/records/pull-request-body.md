<!-- ql:links -->
Task: `docs/plans/2026-10-17-desktop-app-sidebar-layout/`
<!-- /ql:links -->

## Why the change

Non-engineers can now see and click through setting up a kirokit team in a finished mockup, before any desktop app code is written.

## What a reviewer must know

- **The change is one HTML design page and this task's documents.** The page is excluded from git, so the repository diff outside `docs/plans` is empty.
- **The mockup enforces one installed team.** Installing a second team asks to replace the first. The real kirokit CLI still allows several teams.
- **Install state, file counts, times and folder paths are sample values.** Teams, agents, tools, skills and MCP versions are copied from this repository.

## What changed

The page has one install state that every button and typed command goes through, so the pages and the terminal never disagree.

```
sidebar-layout.html
├── sidebar        5 sections, 60 px icon rail, Command-B
├── pages          Teams, Agents, Install and sync, Skills and MCP versions, Settings
├── terminal       docked, Control-backtick, resize 120-480 px, kirokit's real output lines
├── setInstalled   one map; at most one team true
└── dialogs        replace question, then prerequisites check (SOPS, key, Kiro CLI)
```

## Screens compared

<!-- ql:screens -->
- **sidebar-layout** · 1440 px · light · captured · current
  - Polish: No comparison recorded.
- **sidebar-layout** · 1440 px · dark · captured · current
  - Polish: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
1 slice · 3 tasks · 1 commit · 0 files · +0 −0

**Slice 1 · A complete clickable kirokit window with sidebar, terminal and one-team setup** · T1, T2, T3 · no files
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. The mockup opens full size in the QualityLayer App.** Scenario 1: Open the window, visit all five pages and collapse the sidebar, after slice 1 · passed · `evidence/screen-3decee0f3e9494573da4a3f4.png`
- **2. The sidebar lists all five sections, and each opens its page.** Scenario 1: Open the window, visit all five pages and collapse the sidebar, after slice 1 · passed
- **3. The sidebar collapses to an icon rail and expands again.** Scenario 1: Open the window, visit all five pages and collapse the sidebar, after slice 1 · passed
- **4. A non-engineer can click through setting up a team.** Scenario 2: Set up a team through the prerequisites check, after slice 1 · passed
- **5. Engineers get a full terminal docked at the bottom.** Scenario 2: Set up a team through the prerequisites check, after slice 1 · Scenario 3: Only one team is installed at a time, after slice 1 · passed
- **6. The mockup shows kirokit's real teams, agents and versions.** Scenario 1: Open the window, visit all five pages and collapse the sidebar, after slice 1 · passed
- **7. The mockup reads well in light and dark mode.** Only you can confirm: the look fits kirokit · waits for your confirmation
- **8. kirokit's code and behaviour stay unchanged.** Scenario 5: Nothing outside this task changes, after slice 1 · passed
- **9. Only one team is installed at a time.** Scenario 3: Only one team is installed at a time, after slice 1 · passed
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No Go, Wails or build changes.** Nothing under `cmd`, `internal` or the release setup changes, and no binary is produced.
- **No real checks of SOPS, keys or Kiro CLI.** The prerequisites are simulated; detecting them belongs to the later app.
- **No change to kirokit to enforce one team.** The real CLI still allows several installed teams; the mockup's terminal applies the app's rule.
- **No editing of templates or manifests in the mockup.** This is your Discuss decision, and it holds for every page.
- **No phone layout.** The mockup is a desktop window only.
<!-- /ql:not-doing -->

## How it was verified

<!-- ql:checked -->
Polish, Project checks and Code review ran: 8 of 8 points of Done means passed; 1 is yours to confirm.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. Open the mockup and collapse the sidebar
- Do: open the page and press Command-B.
- Run: `qualitylayer design show sidebar-layout`
- Expect: the sidebar shrinks to a rail of five icons and expands again.
- Saw: the sidebar changed between 232 and 60 pixels — `records/it1-browser-run.txt`

#### 2. Set up platform-team through the prerequisites check
- Do: press Install on platform-team, press Check again, then Install.
- Expect: the terminal prints `installed platform-team into /Users/you/.kiro` and both pages show it Installed.
- Saw: the same line and badges — `records/it1-browser-run.txt`

#### 3. Switch teams from the terminal
- Do: type `kirokit install data-team` in the terminal.
- Expect: platform-team is uninstalled first and only data-team shows Installed.
- Saw: both lines printed and one Installed badge — `records/it1-browser-run.txt`

### Not verified

- The 390 pixel width, which the Plan excludes, and whether the look fits kirokit, which only you can judge.

### Docs

- No doc impact.

<!-- optional section here, only when there is content: "Security" -->
