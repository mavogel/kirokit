---
type: outline
---

# Outline: Support Quick, Kiro IDE and kirocrew; README refresh

## Summary

Five slices and fourteen tasks: the Quick bundle renders first (with ids and goldens), then the ids helper, the desktop button, the Kiro IDE and kirocrew check with the README and docs pages, and finally the longer showcase. The risky parts are Amazon Quick's untested agent overwrite (checked by your round trip after slice 2) and the showcase's two new app scenes.

## How the build runs

<!-- ql:waves -->
<!-- /ql:waves -->

## How each point is proved

<!-- ql:proof-map -->
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** On a fresh Kiro home, `kirokit install product-team` makes the five product-team agents appear in Kiro CLI, in Kiro IDE's agent picker and in kirocrew's agent list, and the `user-story` skill and `product-owner` agent from `kirokit render --quick product-team` survive the full two-pass round trip in Quick: after `kirokit quick ids` and a second render, the reimport overwrites both in place, the skill re-exports identical to the rendered file with the same id, and the re-exported agent still lists that id in `skills`.

## Decided while outlining

- Render reads the ids file in slice 1 and the `quick ids` helper comes in slice 2, because the goldens need both id-less and id-carrying output from the first slice, while the helper is an independent reader of Quick exports.
- The product-owner Kiro agent also gets the new `user-story` skill (resource and permission), because the showcase's CLI scene types `/user` in a product-owner chat and waits for `user-story`.
- The README link check is a small repository script (`scripts/check-doc-links.mjs`) run by hand and in the docs slice's validation, not wired into CI, because CI wiring is outside the agreed scope.

## Slice 1: A team renders its Amazon Quick bundle from the same templates

After this slice `kirokit render --quick product-team --out <dir>` writes the `user-story` skill and the `product-owner` Quick agent, with `id:` lines and pinned ids when an ids file knows them. The Quick bundle has goldens compared byte for byte in `go test ./...`.

```diff
 internal/
   manifest/manifest.go        ~ Quick section, strict decode
+  quick/quick.go              + RenderTeam: Kiro render to scratch, then the bundle
+  quick/ids.go                + read quick-ids.yaml
   golden/golden.go            ~ also renders quick/ for teams with a quick section
 cmd/kirokit/commands/render.go ~ --quick, --quick-ids
 teams/product-team/
+  skills/user-story/SKILL.md.tmpl
   manifest.yaml               ~ quick: section
   agents/product-owner/agent.json.tmpl ~ user-story resource + permission
+testdata/golden/quick-ids.yaml  + fixture ids
 testdata/golden/product-team/ ~ regenerated, + quick/
```

- [ ] T1 manifest.yaml accepts a strict `quick` section
  - **Objective:** Teams list the skills and agents meant for Quick, with each agent's welcome message, starter prompts and pinned skills; a typo in that section fails the load instead of silently dropping content.
  - **Design:** "Manifest `quick` section"
  - **Files:** `internal/manifest/manifest.go`, `internal/manifest/manifest_test.go`
  - **Test first:** `TestLoad_Quick*` in `manifest_test.go` loads a manifest with a `quick` section and expects the parsed skills and agents; it fails today because `Manifest` has no `Quick` field (compile error).
  - **Definition of done:**
    - A manifest without `quick` loads with `Quick == nil`, and its JSON form has no `quick` key.
    - Skills, agents, welcome message, starter prompts and agent skills are parsed in order.
    - An unknown key anywhere under `quick` (for example `quick.agents[0].welcome`) fails with an error naming the key.
    - Two agents with one name, or one skill listed twice, fail.
    - verify: `go test ./internal/manifest`

- [ ] T2 `internal/quick` builds the bundle, with ids when known
  - **Objective:** A new package renders the team's Kiro tree into scratch, then writes `skills/<name>/SKILL.md` and `agents/<name>.json` in Quick's format, inserts `id:` lines and pinned ids from the ids map, and publishes the folder like render does.
  - **Design:** "`render` gains `--quick` and `--quick-ids`; output is the Quick bundle only", "Quick skill and agent shapes", "`quick-ids.yaml` and `kirokit quick ids`"
  - **Files:** `internal/quick/quick.go`, `internal/quick/ids.go`, `internal/quick/quick_test.go`, `internal/quick/ids_test.go`, `internal/render/render.go`
  - **Test first:** `TestRenderTeam_*` in `quick_test.go` renders a temp team with one skill and one agent and expects the two files with exact bytes; it fails because the package does not exist.
  - **Definition of done:**
    - Skill bytes equal the Kiro-rendered `SKILL.md`; with an id, `id: <hex>` is the last frontmatter line; vendored skills are copied whole.
    - Agent JSON has the ten keys in the contract order, two-space indent, trailing newline; `system_prompt` is the inline prompt or the rendered `file://` prompt file; `skills` lists only pinned skills that have ids, in manifest order.
    - `Result` reports skill and agent counts and the skills without an id.
    - Errors: no `quick` section, a listed skill or agent missing from the team, a pinned skill not in `quick.skills`, a skill without `name` or `description`, a `file://` prompt leaving the rendered tree, a prompt over 50,000 characters, an id that is not 32 lowercase hex.
    - `LoadIDs` reads the `skills:` map; a missing file is an empty map, an unreadable or malformed one is an error.
    - verify: `go test ./internal/quick ./internal/render`

- [ ] T3 `kirokit render --quick` writes the bundle and warns about missing ids
  - **Objective:** The CLI exposes the bundle with `--quick`, finds the ids file by flag, `KIROKIT_QUICK_IDS` or `<repo>/quick-ids.yaml`, prints one line per team and one warning per missing id.
  - **Design:** "`render` gains `--quick` and `--quick-ids`; output is the Quick bundle only"
  - **Files:** `cmd/kirokit/commands/render.go`, `cmd/kirokit/commands/quick_test.go`
  - **Test first:** `TestRenderQuick_*` in `quick_test.go` runs `render --quick product-team --out <tmp>` through `runRoot` and expects `rendered Quick bundle for product-team: 1 skills, 1 agents`; it fails because the flag is unknown.
  - **Definition of done:**
    - A named team with no `quick` section fails with `team <t> lists nothing for Quick`; rendering all teams skips such teams with a note.
    - A missing id prints `warning: no Quick id for skill <name> yet (first import); run kirokit quick ids after importing` to stderr and exits 0.
    - `--quick-ids` wins over `KIROKIT_QUICK_IDS`, which wins over `<repo>/quick-ids.yaml`.
    - Plain `render` (no `--quick`) output is unchanged.
    - verify: `go test ./cmd/kirokit/commands -run 'Quick|KirokitRepo'`

- [ ] T4 product-team ships `user-story` for Kiro and Quick, and its Quick goldens are checked
  - **Objective:** The example team gets a `user-story` skill and a `quick` section for it and the product-owner agent; the golden harness renders `quick/` with the committed fixture ids and fails on drift or on a stray `quick/` folder.
  - **Design:** "Golden harness covers the Quick bundle", "plan:decision:d2"
  - **Files:** `teams/product-team/skills/user-story/SKILL.md.tmpl`, `teams/product-team/manifest.yaml`, `teams/product-team/agents/product-owner/agent.json.tmpl`, `internal/golden/golden.go`, `internal/golden/golden_test.go`, `testdata/golden/quick-ids.yaml`, `testdata/golden/product-team/context.json`, `testdata/golden/product-team/rendered/agents/product-owner.json`, `testdata/golden/product-team/rendered/skills/user-story/SKILL.md`, `testdata/golden/product-team/quick/skills/user-story/SKILL.md`, `testdata/golden/product-team/quick/agents/product-owner.json`, `desktop/catalog_test.go`
  - **Test first:** `TestCompare_Quick*` in `golden_test.go` sets up a fixture team with a `quick` section, runs `Update` then `Compare`, and expects a `quick/` golden and no diffs, then a diff after a skill change; it fails because the harness ignores Quick.
  - **Definition of done:**
    - `Update` writes `testdata/golden/<team>/quick/` for a team with a `quick` section, using `testdata/golden/quick-ids.yaml`.
    - `Compare` reports a changed, missing or extra Quick file, and a `quick/` folder for a team without the section.
    - The committed product-team Quick golden has the fixture id line and the pinned id; only product-team's goldens change.
    - verify: `go test ./internal/golden ./desktop/... ./cmd/kirokit/commands`

**Validation:** automated: `go test ./...` · `task build` · `./kirokit render --quick product-team --out tmp/quick-check` · manual: none

## Slice 2: Quick ids come back from Quick's own exports

**Depends on:** Slice 1

After this slice `kirokit quick ids <exports>` turns skills exported from Quick (files, folders, zips) into the gitignored `quick-ids.yaml`, so the next render writes the ids back. Your round trip on your Quick account then checks the full two-pass flow.

- [ ] T5 `kirokit quick ids` records ids from exported skills
  - **Objective:** A new `quick ids` subcommand reads exported `SKILL.md` files, folders and `.zip` files in memory, merges their `name`/`id` pairs into the ids file (mode 0600, keys sorted) and the repository ignores that file.
  - **Design:** "`quick-ids.yaml` and `kirokit quick ids`"
  - **Files:** `internal/quick/ids.go`, `internal/quick/ids_test.go`, `cmd/kirokit/commands/quick.go`, `cmd/kirokit/commands/root.go`, `cmd/kirokit/commands/quick_test.go`, `.gitignore`
  - **Test first:** `TestQuickIDs_*` in `quick_test.go` runs `quick ids` on an exported skill from `docs/research/quick/SKILL-minimal-exported.md` and expects `recorded 1 ids in <file>` with that id; it fails because the command does not exist.
  - **Definition of done:**
    - A `SKILL.md`, a folder searched recursively, and a `.zip` with `SKILL.md` entries are all read; existing names are updated and others kept.
    - An export with no `id:` prints `skipped <path>: no id`; two different ids for one name in one run fail; a malformed id fails.
    - A zip entry over 1 MiB or a zip with more than 256 entries fails without writing.
    - The written file starts with the "never commit" comment, is mode 0600, and `git check-ignore quick-ids.yaml` matches.
    - verify: `go test ./internal/quick ./cmd/kirokit/commands -run 'Quick'`

- [ ] T6 Your two-pass round trip on your Quick account
  - **Objective:** Confirm on a real account what no test can: the bundle imports, `quick ids` plus a second render make the reimport overwrite both in place, the skill re-exports identical with the same id, and the agent still pins it.
  - **Design:** "plan:decision:d3"
  - **Files:** none
  - **Qualification:** the build records a need for you with the exact steps; your exports are saved under `records/` and compared with `diff` against the rendered files and with `grep` for the id in the agent's `skills`.
  - **Definition of done:**
    - The re-exported `SKILL.md` equals the second-pass rendered file byte for byte.
    - The re-exported agent JSON lists the same id in `skills`.
    - A mismatch amends the Plan before slice 3 describes the flow.
    - verify: `diff records/quick-roundtrip/user-story-reexport.md tmp/quick-pass2/skills/user-story/SKILL.md`

**Validation:** automated: `go test ./...` · `task lint` · manual: your Quick round trip (T6)

## Slice 3: The desktop app exports a team for Quick

**Depends on:** Slice 1

After this slice each team card that lists Quick content has **Export for Quick**: it runs `kirokit render --quick <team> --out <app data>/quick/<team>` in the kirokit tab, shows the counts and the folder, and opens it once. A team without Quick content shows a one-line note.

- [ ] T7 The app knows which teams have Quick content and can open their bundle folder
  - **Objective:** The catalog marks Quick teams, and three bound calls return the export folder, count the bundle and open the folder, never taking a path from the frontend.
  - **Design:** "Desktop: catalog flag, Export for Quick, open folder"
  - **Files:** `desktop/catalog.go`, `desktop/catalog_test.go`, `desktop/quick.go`, `desktop/quick_test.go`, `desktop/platform.go`, `desktop/platform_test.go`
  - **Test first:** `TestQuickDir_*` and `TestQuickBundle_*` in `desktop/quick_test.go` and `TestOpenCommandFor` in `platform_test.go`; they fail because the functions do not exist.
  - **Definition of done:**
    - `Team.Quick` is true for product-team and false for the others.
    - `QuickDir` returns `<dataDir>/quick/<team>` and refuses a name with a separator, `..` or empty.
    - `QuickBundle` counts `skills/*/SKILL.md` and `agents/*.json`; a missing folder is an error.
    - `openCommandFor` gives `open` on darwin, `xdg-open` on linux, `explorer` on windows; `OpenFolder` refuses a team whose folder does not exist.
    - verify: `cd desktop && go test ./...`

- [ ] T8 Teams page has Export for Quick with a result line
  - **Objective:** The card shows the button (or a note), disables it while a command runs, and after a successful run shows "Quick bundle ready: N skills, M agents", the folder and Show folder; a failed run shows the error beside the button.
  - **Design:** "Desktop: catalog flag, Export for Quick, open folder"
  - **Files:** `desktop/frontend/src/pages/teams.js`, `desktop/frontend/src/pages/pages.test.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/api.js`, `desktop/frontend/src/style.css`
  - **Test first:** new cases in `pages.test.js` render `teamsView` with a Quick team, a non-Quick team, a busy view, a ready result and an error, and check the HTML; they fail because the button does not exist.
  - **Definition of done:**
    - `quickArgs(team, dir)` is `['render', '--quick', team, '--out', dir]`.
    - A Quick team's card has `data-action="quick"`; a team without Quick content has the note and no button.
    - A ready result shows the counts with singular and plural forms, the folder path and Show folder; an error for `quick:<team>` shows beside the button.
    - main.js stores the result only for the run id it started, and opens the folder once.
    - verify: `cd desktop/frontend && node --test "src/**/*.test.js"`

**Validation:** automated: `task desktop:test` · manual: Export for Quick in `task desktop:dev` (Scenario 3)

## Slice 4: Kiro IDE and kirocrew are confirmed, and the README starts with what kirokit is

**Depends on:** Slice 2

After this slice the README is a short intro (pitch, video, why, install, first steps, Works with, docs index), and its reference sections live in `docs/` pages with every link resolving. Kiro IDE and kirocrew are confirmed to list the installed product-team agents, and the README says what each picks up.

- [ ] T9 Kiro IDE and kirocrew list the installed product-team agents
  - **Objective:** Repeat the IDE and kirocrew probes on a fresh Kiro home with this build's product-team (now with `user-story`), to confirm the docs' "Works with" claims.
  - **Design:** "plan:decision:d4"
  - **Files:** none
  - **Qualification:** install product-team into `/tmp/kirokit-ide-check/kiro` with `./kirokit install`, drive Kiro IDE with playwright-core and kirocrew's `/api/agents/installed`, save both outputs under `records/`.
  - **Definition of done:**
    - The IDE agent selector lists customer, engineer, product-owner, qa and ux-designer.
    - kirocrew's installed-agents response lists the same five names.
    - Nothing is written outside `/tmp/kirokit-ide-check`.
    - verify: `grep -c product-owner docs/plans/2026-10-27-support-amazon-quick-see-docs-research-q/records/check-kirocrew-agents.json`

- [ ] T10 Reference sections move to `docs/` pages and every link resolves
  - **Objective:** Each README reference section becomes a page under `docs/`, a new `docs/amazon-quick.md` explains the two-pass flow, and every file that names a README section points at its new page.
  - **Design:** "plan:decision:d6"
  - **Files:** `docs/configuration.md`, `docs/agents-and-skills.md`, `docs/secrets.md`, `docs/install-and-sync.md`, `docs/testing.md`, `docs/desktop-app.md`, `docs/amazon-quick.md`, `scripts/check-doc-links.mjs`, `global/steerings/security-baseline.md`, `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl`, `teams/platform-team/manifest.yaml`, `teams/data-team/steerings/data-conventions.md.tmpl`, `testdata/golden/platform-team/rendered/kirokit-steerings/platform-team/security-baseline.md`, `testdata/golden/data-team/rendered/steering/data-conventions.md`, `CONTRIBUTING.md`, `AGENTS.md`
  - **Test first:** `node scripts/check-doc-links.mjs` lists every relative Markdown link and anchor in `README.md`, `CONTRIBUTING.md`, `AGENTS.md` and `docs/*.md` and every quoted README section name, and fails on one that does not resolve; it fails at first on the moved sections.
  - **Definition of done:**
    - Each README section in the Plan's table has a home page; no content is dropped (heading inventory compared before and after).
    - `docs/amazon-quick.md` documents `render --quick`, `--quick-ids`, `KIROKIT_QUICK_IDS`, `quick ids`, the ids file shape, the manifest `quick` section and the import steps, with the Plan's "After" Mermaid block.
    - Goldens that quote a section name are regenerated with `./kirokit update` and match.
    - verify: `node scripts/check-doc-links.mjs`

- [ ] T11 The README is a short intro in the QualityLayer README's order
  - **Objective:** Rewrite `README.md` to logo, one-line pitch naming Kiro CLI, Kiro IDE, kirocrew and Amazon Quick, the GIF linking the MP4, "Why kirokit", install for CLI and app, first steps, "Works with" (with the Mermaid block) and a docs index.
  - **Design:** "plan:decision:d6", "plan:decision:d4"
  - **Files:** `README.md`
  - **Trivial:** `node scripts/check-doc-links.mjs` covers every link and anchor; the length is checked with `wc -l`.
  - **Definition of done:**
    - Intro, video and install come before any reference material; the README is under half of today's 359 lines.
    - "Works with" says what Kiro CLI, Kiro IDE, kirocrew and Amazon Quick each pick up, including the IDE's `KIRO_HOME` rules.
    - verify: `node scripts/check-doc-links.mjs`

**Validation:** automated: `node scripts/check-doc-links.mjs` · `go test ./...` · `node scripts/sync-agent-assets.mjs --check` · manual: none

## Slice 5: The showcase shows Quick, Kiro IDE and kirocrew

**Depends on:** Slice 4

**Risky:** the recording drives two more desktop apps (Kiro IDE through Electron, kirocrew in headless Chrome) on this Mac.

After this slice `task showcase` records title, CLI (install product-team, Kiro chat, `render --quick`), Kiro IDE, kirocrew, the app and the end card, checked for privacy, size and a 30 to 90 second length. The README's GIF and MP4 are the new recording.

- [ ] T12 Showcase scripts drive the IDE and kirocrew scenes and the longer limits
  - **Objective:** Update the CLI scene, add `ide-drive.mjs` and `crew-drive.mjs`, compose them in `record.sh`, raise the limits and add the captions, with tests for each new limit and failure path.
  - **Design:** "Showcase scenes", "plan:decision:d5"
  - **Files:** `scripts/showcase/cli-drive.sh`, `scripts/showcase/ide-drive.mjs`, `scripts/showcase/crew-drive.mjs`, `scripts/showcase/record.sh`, `scripts/showcase/checks.sh`, `scripts/showcase/caption.mjs`, `scripts/showcase/showcase.test.mjs`, `CONTRIBUTING.md`
  - **Test first:** new cases in `showcase.test.mjs` expect a 75 s video to pass and 95 s to fail the length check, a 9 MB GIF to pass, the privacy check to read `ide-text.txt` and `crew-text.txt`, and the new captions to be timed; they fail on today's limits.
  - **Definition of done:**
    - `checks.sh length` accepts 30 to 90 s; `size` accepts up to 10 MB GIF and 6 MB MP4.
    - The privacy check fails when `$HOME`, the git email or `$USER` appears in either new text file.
    - Both drivers exit non-zero with a clear message when their app is missing or the five names do not appear within the timeout; the crew driver stops the gateway on every exit.
    - The crew token never appears in `crew-text.txt`.
    - verify: `node --test scripts/showcase/`

- [ ] T13 Record and publish the new showcase
  - **Objective:** Run `task showcase` on this Mac and publish the new `assets/kirokit-showcase.gif` and `.mp4`.
  - **Design:** "Showcase scenes"
  - **Files:** `assets/kirokit-showcase.gif`, `assets/kirokit-showcase.mp4`
  - **Trivial:** `task showcase` runs `checks.sh` (privacy, size, length) before it publishes.
  - **Definition of done:**
    - The recording shows the CLI install, Kiro chat, `render --quick` and its bundle, Kiro IDE's agent list, kirocrew's agent list and the app.
    - verify: `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4`

**Validation:** automated: `node --test scripts/showcase/` · manual: watch the published MP4

## Program design

The Plan's contracts hold the design.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Quick bundle render | `internal/quick/quick.go` | `render.RenderTeamWith`, render's publish-by-rename |
| Quick ids file and export reader | `internal/quick/ids.go` | `yaml.v3`, `archive/zip` |
| `quick ids` command | `cmd/kirokit/commands/quick.go` | the `skills` command group's shape |
| Desktop Quick calls | `desktop/quick.go` | `openCommandFor` in `desktop/platform.go` |
| Doc link check | `scripts/check-doc-links.mjs` | Node standard library only |

## End-to-end scenarios

### Scenario 1: Render the Quick bundle from the CLI

**Runs after:** slice 1
**Proves:** Done means 1, Done means 3
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `task build` | exits 0 |
| 2 | `env KIROKIT_QUICK_IDS=/nonexistent ./kirokit render --quick product-team --out tmp/scenario1` | prints `rendered Quick bundle for product-team: 1 skills, 1 agents` and a warning for `user-story` |
| 3 | `rtk proxy find tmp/scenario1 -type f` | lists `tmp/scenario1/skills/user-story/SKILL.md` and `tmp/scenario1/agents/product-owner.json` |
| 4 | `./kirokit render --quick data-team --out tmp/scenario1-data` | fails with `team data-team lists nothing for Quick` |
| 5 | `./kirokit test` | every team prints `ok` |

### Scenario 2: Ids from a Quick export come back on the next render

**Runs after:** slice 2
**Proves:** Done means 2
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Copy `tmp/scenario1/skills/user-story/SKILL.md` to `tmp/export/SKILL.md` and add `id: 0123456789abcdef0123456789abcdef` as the last frontmatter line | file exists |
| 2 | `./kirokit quick ids --quick-ids tmp/ids.yaml tmp/export` | prints `recorded 1 ids in tmp/ids.yaml` |
| 3 | `./kirokit render --quick product-team --quick-ids tmp/ids.yaml --out tmp/scenario2` | prints the bundle line and no warning |
| 4 | `diff tmp/export/SKILL.md tmp/scenario2/skills/user-story/SKILL.md` | no output, exit 0 |
| 5 | `grep -c 0123456789abcdef0123456789abcdef tmp/scenario2/agents/product-owner.json` | `1` |

### Scenario 3: Export for Quick in the desktop app

**Runs after:** slice 3
**Proves:** Done means 1
**Screens:** artifacts/team-quick-bundle/index.html#teams-quick

| Step | Action | Expected |
|---|---|---|
| 1 | Start the app with `task desktop:dev` on a cloned config repo and open Teams | product-team's card has Export for Quick; data-team and platform-team show the note |
| 2 | Click Export for Quick on product-team | the kirokit tab runs `kirokit render --quick product-team --out …/quick/product-team`; buttons are disabled while it runs |
| 3 | Wait for the run to end | the card shows "Quick bundle ready: 1 skill, 1 agent", the folder path and Show folder; the folder opens once |

### Scenario 4: One install, three Kiro tools

**Runs after:** slice 4
**Proves:** Done means 4
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `env KIRO_HOME=/tmp/kirokit-ide-check/kiro ./kirokit install product-team` | prints `installed product-team into /tmp/kirokit-ide-check/kiro` |
| 2 | `env KIRO_HOME=/tmp/kirokit-ide-check/kiro kiro-cli agent list` | lists the five product-team agents |
| 3 | Launch Kiro IDE with that `KIRO_HOME` and a fresh profile, open the agent selector | the five agents are listed |
| 4 | Start `kirocrew gateway` on a throwaway home with that `KIRO_HOME`, read `/api/agents/installed` | the five agents are listed |

### Scenario 5: The README is short and every link resolves

**Runs after:** slice 4
**Proves:** Done means 5
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `node scripts/check-doc-links.mjs` | prints `all links resolve` |
| 2 | Read `README.md` top to bottom | intro naming Kiro CLI, Kiro IDE, kirocrew and Amazon Quick, the GIF, install and first steps come before any reference material; the docs index links every `docs/` page |

### Scenario 6: The new showcase

**Runs after:** slice 5
**Proves:** Done means 6
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `node --test scripts/showcase/` | all tests pass |
| 2 | `scripts/showcase/checks.sh length assets/kirokit-showcase.mp4` | passes (30 to 90 s) |
| 3 | Watch `assets/kirokit-showcase.mp4` | CLI install and Kiro chat, `render --quick`, Kiro IDE's agents, kirocrew's agents, the app |

### Scenario 7: The Quick round trip on your account

**Runs after:** slice 2
**Proves:** Done means 7 (only you can confirm: the import in your Quick account)
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Import `skills/user-story` (From folder) and `agents/product-owner.json` (From file) from a first-pass bundle | both appear in Quick |
| 2 | Export the skill, run `kirokit quick ids <export>`, render again, reimport both and confirm the overwrite | one skill and one agent, no duplicates |
| 3 | Export the skill and agent again | the skill equals the rendered file with the same id; the agent's `skills` holds that id |

## Commands the build runs

<!-- ql:commands -->
<!-- /ql:commands -->
