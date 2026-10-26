---
type: outline
---

# Outline: Cross-role agent harness with feedback loop

## Summary

Five slices and fourteen tasks: owners and the feedback repository in the manifest and render first, then the `product-team` role agents with the spec review gates, the `/feedback` prompt, the desktop app's Owner column and Feedback button, and last the `analyze-session` skill with the README. The risky parts are the app's first path that opens the system browser (slice 4) and the two live Kiro runs (subagent consultation and a spec in CLI v3, slice 2) that nobody here has done.

## How the build runs

<!-- ql:waves -->
- Wave 1 · Slice 1 (Every team can name owners and a feedback repository) · first
- Wave 2 · Slice 2 (A team can install five role agents that consult each other) · on its own
- Wave 3 · Slice 3 (`/feedback` in Kiro drafts a prefilled issue for the owner), Slice 4 (The desktop app shows owners and opens a feedback issue) · side by side
- Wave 4 · Slice 5 (`analyze-session` turns a Kiro chat into owner issues, and the README explains it all) · on its own

4 waves · longest path 10 of 13 tasks

A fresh agent runs the scenarios after slice 2 (first live run of Kiro's `subagent` tool and of a spec in Kiro CLI v3.) and slice 4 (the app's first call that opens the system browser, with an address built from repository content.) before the build goes on.
<!-- /ql:waves -->

## How each point is proved

<!-- ql:proof-map -->
- Done means 1 · A team can install a ready example of cooperating role agents → "A product owner consults the customer agent" · runs after slice 2
- Done means 2 · The spec workflow has review gates by role → "A spec's requirements and design are reviewed by role" · runs after slice 2
- Done means 3 · Every agent and skill of a team names its owner → "Owners are validated and inherited" · runs after slice 1; "The desktop app shows owners and opens a feedback issue" · runs after slice 4
- Done means 4 · A non-engineer can send feedback on a skill or agent from where they work → "`/feedback` in Kiro drafts the owner's issue" · runs after slice 3; "The desktop app shows owners and opens a feedback issue" · runs after slice 4
- Done means 5 · The feedback reaches the owner → "`/feedback` in Kiro drafts the owner's issue" · runs after slice 3; "The desktop app shows owners and opens a feedback issue" · runs after slice 4
- Done means 6 · A session-analysis skill suggests harness improvements from a Kiro chat → "Session analysis suggests improvements by owner" · runs after slice 5
- Done means 7 · The README explains the role setup and the feedback loop → "The README explains the role setup and the feedback loop" · runs after slice 5
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** With `product-team` installed into a temporary Kiro home, `kiro-cli chat --agent product-owner --v3` consults the customer agent as a subagent, and both `/feedback` and the desktop app's **Feedback** button produce a `https://github.com/mavogel/kirokit/issues/new?…` address whose body names the agent and @-mentions its owner.
- **Truth:** A team manifest without `owner`, `owners` or `feedback` renders byte-for-byte as before; only teams that set them change goldens.

## Decided while outlining

- **Role agents carry their persona as an inline `prompt`, not a `prompts/` file** — every file in `prompts/` becomes a slash command, and `/customer` as a command would confuse a non-engineer.
- **The every-team golden check is a test of `kirokit test` with no team argument** — the command already loops over every team; the test makes `go test ./...` fail on any team's drift.
- **Owner lookup in the desktop catalog uses the agent's folder name** — `OwnerOf` keys on `agents/<name>/`, while a rendered `name` field may differ.

## Slice 1: Every team can name owners and a feedback repository

After this slice `manifest.yaml` accepts `owner`, `owners` and `feedback.github`, rejects malformed handles and unknown agent or skill keys, and templates can read the owner table and the feedback repository. Render also rejects a subagent name that is not an agent of the team, and `go test ./...` checks every team's goldens.

```diff
 internal/
 ├── manifest/manifest.go     ~ Owner, Owners, Feedback, OwnerOf, validation
 └── render/
     ├── context.go           ~ team agent and skill names (json:"-")
     ├── funcs.go             ~ owners, feedbackRepo
     └── render.go            ~ owner key check, subagent name check
 teams/platform-team/manifest.yaml  ~ owner: "@mavogel"
 teams/data-team/manifest.yaml      ~ owner: "@mavogel"
```

- [x] T1 `manifest.yaml` accepts owners and a feedback repository, and rejects malformed values
  - **Objective:** Add `Owner`, `Owners` and `Feedback` to `manifest.Manifest` with field-named validation at load, and `OwnerOf(kind, name)` with the team-owner fallback, so every consumer reads one owner rule.
  - **Design:** "`manifest.Manifest` gains `Owner`, `Owners` and `Feedback`, validated at load", "plan:decision:d5"
  - **Files:** `internal/manifest/manifest.go`, `internal/manifest/manifest_test.go`
  - **Test first:** table test loading fixtures with `owner: "@jane"`, `owners.agents.qa: "jane"`, `feedback.github: "not a repo"` and none of the fields; it fails today because `Manifest` has no such fields and unknown keys are ignored.
  - **Definition of done:**
    - `@user`, `@org/team` load; `jane`, `@`, `@-x`, `@a b` fail with `manifest: <path>: <field>: "<value>" is not a GitHub handle (write @jane)`.
    - `feedback.github` accepts `owner/repo`, rejects `owner`, `a/b/c`, `https://github.com/a/b` and `javascript:x` naming the field.
    - `OwnerOf` returns the override with `fromTeam=false`, else the team owner with `fromTeam=true`, else `""`; a manifest without `owner` loads (d5: optional).
    - verify: `go test ./internal/manifest`

- [x] T2 Render checks owner keys against the team and offers `owners` and `feedbackRepo`
  - **Objective:** Collect the team's agent and skill names (including vendored ones) before the walk, fail on an `owners` key the team does not have, and add the two template functions the `/feedback` prompt uses.
  - **Design:** "Render checks owner keys and subagent names against the team, and offers two template functions"
  - **Files:** `internal/render/context.go`, `internal/render/funcs.go`, `internal/render/render.go`, `internal/render/render_test.go`
  - **Test first:** a fixture team with `owners.agents.ghost` fails render with `owners.agents.ghost: fixture-team has no agent ghost`, and a template calling `{{ range owners }}` and `{{ feedbackRepo }}` renders the rows; both fail today (no check, no functions).
  - **Definition of done:**
    - `owners` returns rows sorted by kind then name, each agent and skill once, vendored skills with the team owner and `FromTeam=true`.
    - `feedbackRepo` on a team without `feedback.github` fails with `feedbackRepo: <team>'s manifest.yaml sets no feedback.github`.
    - An unknown `owners.skills.<name>` key fails the same way as an agent key.
    - The new `Context` fields stay out of `context.json` (existing goldens unchanged).
    - verify: `go test ./internal/render`

- [x] T3 Render rejects a subagent name that is not an agent of the team
  - **Objective:** After rendering an agent, check `toolsSettings.subagent.availableAgents` and `trustedAgents`: every entry without `*`, `?` or `[` must be a team agent, because `kiro-cli agent validate` accepts any name.
  - **Design:** "Render checks owner keys and subagent names against the team, and offers two template functions"
  - **Files:** `internal/render/render.go`, `internal/render/render_test.go`
  - **Test first:** a fixture agent listing `availableAgents: ["agent1", "ghost"]` fails render with `toolsSettings.subagent.availableAgents names "ghost", which is not an agent of fixture-team`; render succeeds today.
  - **Definition of done:**
    - Names of team agents and globs (`*`, `qa-?`, `[ab]*`) pass; an unknown name in `trustedAgents` fails naming that field.
    - An agent without `toolsSettings.subagent` renders as before.
    - verify: `go test ./internal/render`

- [x] T4 The two existing teams name an owner, and `go test ./...` checks every team's goldens
  - **Objective:** Add `owner: "@mavogel"` to `platform-team` and `data-team`, regenerate their goldens, and test `kirokit test` with no argument so any team's drift fails the suite.
  - **Design:** "plan:decision:d4"
  - **Files:** `teams/platform-team/manifest.yaml`, `teams/data-team/manifest.yaml`, `testdata/golden/platform-team/context.json`, `testdata/golden/data-team/context.json`, `cmd/kirokit/commands/root_test.go`
  - **Test first:** the new `kirokit test` (all teams) test fails once the manifests gain `owner` and before `./kirokit update` runs, proving it compares every team.
  - **Definition of done:**
    - Only the two `context.json` files change in `testdata/golden/`; rendered agent files are unchanged.
    - verify: `go test ./cmd/kirokit/... ./internal/golden`

**Validation:** automated: `go test ./...` · `go vet ./...` · `task lint` · manual: none

## Slice 2: A team can install five role agents that consult each other

**Depends on:** Slice 1

**Risky:** first live run of Kiro's `subagent` tool and of a spec in Kiro CLI v3.

After this slice `kirokit install product-team` gives the product owner, engineer, customer, QA and UX/UI designer agents, each able to call the other four, with a steering that asks the product owner to review requirements and the engineer to review a design. The live Kiro runs prove the consultation and record whether specs run in the CLI.

```diff
+teams/product-team/
+├── manifest.yaml                         + owner, owners, feedback.github
+├── agents/{product-owner,engineer,customer,qa,ux-designer}/agent.json.tmpl
+└── steerings/spec-review-gates.md.tmpl   + inclusion: always
+testdata/golden/product-team/            + rendered via ./kirokit update
```

- [x] T5 `product-team` ships five role agents wired as subagents of each other
  - **Objective:** Add the example team with five agents in the config shape the Plan fixes (inline persona prompts, `subagent` tool, `availableAgents` of the other four, `trustedAgents` without the engineer), render its goldens, and update the tests that list exactly two teams.
  - **Design:** "Product-team agent config shape", "plan:decision:d1", "plan:decision:d2"
  - **Files:** `teams/product-team/manifest.yaml`, `teams/product-team/agents/product-owner/agent.json.tmpl`, `teams/product-team/agents/engineer/agent.json.tmpl`, `teams/product-team/agents/customer/agent.json.tmpl`, `teams/product-team/agents/qa/agent.json.tmpl`, `teams/product-team/agents/ux-designer/agent.json.tmpl`, `testdata/golden/product-team/context.json`, `testdata/golden/product-team/rendered/agents/product-owner.json`, `testdata/golden/product-team/rendered/agents/engineer.json`, `testdata/golden/product-team/rendered/agents/customer.json`, `testdata/golden/product-team/rendered/agents/qa.json`, `testdata/golden/product-team/rendered/agents/ux-designer.json`, `desktop/catalog_test.go`, `cmd/kirokit/commands/root_test.go`
  - **Test first:** `TestCatalog_ReadsRealTeams` expecting `data-team`, `platform-team`, `product-team` and a product-owner agent with the `subagent` tool fails today (no such team).
  - **Definition of done:**
    - Each agent's `availableAgents` is exactly the other four; `trustedAgents` is the non-writing three of those; only the engineer has `write` and `shell`, with platform-team's deny rules.
    - The customer prompt says its persona is invented until the team fills it from interviews.
    - `product-team` has no MCP servers and no secrets file; install needs no SOPS key.
    - `go test ./internal/golden` validates the five agents with `kiro-cli agent validate`.
    - verify: `go -C desktop test ./... && go test ./cmd/kirokit/... ./internal/golden`

- [x] T6 The spec review gates steering reaches every role agent
  - **Objective:** Add `spec-review-gates.md.tmpl` (`inclusion: always`) telling the agent writing a spec to ask `product-owner` to review requirements and `engineer` to review the design through `subagent`, and list it in every role agent's `resources`.
  - **Design:** "plan:decision:d3"
  - **Files:** `teams/product-team/steerings/spec-review-gates.md.tmpl`, `teams/product-team/agents/product-owner/agent.json.tmpl`, `teams/product-team/agents/engineer/agent.json.tmpl`, `teams/product-team/agents/customer/agent.json.tmpl`, `teams/product-team/agents/qa/agent.json.tmpl`, `teams/product-team/agents/ux-designer/agent.json.tmpl`, `testdata/golden/product-team/rendered/steering/spec-review-gates.md`, `testdata/golden/product-team/rendered/agents/product-owner.json`, `testdata/golden/product-team/rendered/agents/engineer.json`, `testdata/golden/product-team/rendered/agents/customer.json`, `testdata/golden/product-team/rendered/agents/qa.json`, `testdata/golden/product-team/rendered/agents/ux-designer.json`
  - **Trivial:** golden comparison in `go test ./cmd/kirokit/...` (every team) covers the rendered steering and resources; behaviour is proved by Scenario 2.
  - **Definition of done:**
    - The steering names what each reviewer checks (vision fit, completeness, acceptance criteria, prioritisation; deep modules, information hiding, red flags) and says to fold the answer in before the next phase.
    - verify: `go test ./cmd/kirokit/... ./internal/golden`

- [x] T7 Qualify the live consultation and the spec run in Kiro CLI v3
  - **Objective:** Install `product-team` into a temporary Kiro home, run the product owner and record that it consults the customer agent; run a quick spec and record whether the review gates fire, raising it with the person before Review if specs do not run in the CLI.
  - **Design:** none
  - **Files:** none
  - **Qualification:** Scenarios 1 and 2 run against the built CLI and the logged-in `kiro-cli`; the captured chat output is the evidence, and a missing subagent call or a spec that does not run is the failure witness.
  - **Definition of done:**
    - The chat output shows a `subagent` call to `customer` and the product owner reporting its answer.
    - The spec run's outcome (gates fired, or specs not available in CLI v3) is recorded in the task record.
    - verify: `task build`

**Validation:** automated: `go test ./...` · `task build` · manual: Scenarios 1 and 2 in a logged-in `kiro-cli`

## Slice 3: `/feedback` in Kiro drafts a prefilled issue for the owner

**Depends on:** Slice 2

After this slice a person in any `product-team` chat types `/feedback <what went wrong>` and gets a title, a body that @-mentions the owner, and a percent-encoded `issues/new` link in `mavogel/kirokit`, with nothing sent.

- [x] T8 The `/feedback` prompt renders the owner table and the issue address
  - **Objective:** Add `prompts/feedback.md.tmpl` that renders the team's agents and skills with owners (`owners`) and the issue base address (`feedbackRepo`), and instructs the agent to pick the item, draft title and body (what happened, what was expected, a chat quote, body at most 1500 characters) and print the link with `feedback,team:<team>,<kind>:<name>` labels plus the text to copy.
  - **Design:** "plan:decision:d6", "Render checks owner keys and subagent names against the team, and offers two template functions"
  - **Files:** `teams/product-team/prompts/feedback.md.tmpl`, `testdata/golden/product-team/rendered/prompts/feedback.md`
  - **Trivial:** golden comparison in `go test ./cmd/kirokit/...` covers the rendered table; the behaviour is proved by Scenario 4.
  - **Definition of done:**
    - The rendered prompt lists every agent and skill of the team with its owner, an inherited one marked as the team owner, and an item without owner written as no owner (no @-mention).
    - It tells the agent never to call GitHub or any tool to submit.
    - verify: `go test ./cmd/kirokit/...`

**Validation:** automated: `go test ./...` · manual: Scenario 4 in a logged-in `kiro-cli`

## Slice 4: The desktop app shows owners and opens a feedback issue

**Depends on:** Slice 1

**Risky:** the app's first call that opens the system browser, with an address built from repository content.

After this slice the Agents and Skills pages show each owner and a Feedback button that opens the prefilled GitHub issue in the browser; a team without `feedback.github` shows its buttons off with the reason, and any address that is not a GitHub issues form is refused.

```diff
 desktop/
 ├── catalog.go                    ~ Team.Feedback, Agent/Skill Owner, OwnerFromTeam
 └── frontend/
     ├── index.html                ~ Owner and Feedback columns
     └── src/
         ├── api.js                ~ openURL with the GitHub issues allowlist
+        ├── pages/feedback.js     + feedbackURL
         ├── pages/agents.js       ~ two cells, click delegation
         ├── pages/skills.js       ~ two cells, click delegation
         └── main.js               ~ wires onFeedback to api.openURL
```

- [x] T9 The desktop catalog carries owners and the feedback repository
  - **Objective:** Fill `Team.Feedback` and each agent's and skill's `Owner` and `OwnerFromTeam` from `manifest.OwnerOf`, loading the manifest once in `readTeam`.
  - **Design:** "The desktop catalog carries owners and the feedback repository"
  - **Files:** `desktop/catalog.go`, `desktop/catalog_test.go`
  - **Test first:** a fixture repo with a team owner, one agent override and a vendored skill expects `Owner`/`OwnerFromTeam` per row and `Feedback`; it fails today (no such fields).
  - **Definition of done:**
    - A vendored skill gets the team owner with `ownerFromTeam: true`; a team without `owner` gives `""`; a team without feedback gives `feedback: ""`.
    - The real `product-team` reports `feedback: "mavogel/kirokit"`.
    - verify: `go -C desktop test ./...`

- [x] T10 `feedbackURL` builds the issue address and `api.openURL` opens only GitHub issue forms
  - **Objective:** Add the pure `feedbackURL(repo, team, kind, name, owner)` helper and `api.openURL`, which passes a URL to `window.runtime.BrowserOpenURL` only when it matches `^https://github\.com/[A-Za-z0-9-]+/[A-Za-z0-9._-]+/issues/new\?`.
  - **Design:** "The frontend opens only GitHub issue addresses"
  - **Files:** `desktop/frontend/src/pages/feedback.js`, `desktop/frontend/src/pages/feedback.test.js`, `desktop/frontend/src/api.js`, `desktop/frontend/src/api.test.js`
  - **Test first:** node tests that decode the built URL's title, body and labels, and that feed `javascript:alert(1)`, `https://evil.example/issues/new?x` and `https://github.com/a/b/pulls?x` to `openURL` expecting a rejection and no `BrowserOpenURL` call; both fail today (no module, no function).
  - **Definition of done:**
    - Title `Feedback: <team> <kind> <name>`; body as the artifact's template with `**Owner:** @x` only when an owner exists; labels `feedback,team:<team>,<kind>:<name>`.
    - A name or team with `&`, `#`, spaces or quotes round-trips through `URLSearchParams` unchanged.
    - A refused URL rejects with `refusing to open <url>`.
    - verify: `task desktop:test`

- [x] T11 Agents and Skills pages show the Owner column and a Feedback button
  - **Objective:** Add the Owner and Feedback cells to both tables as the artifact shows, a disabled button and a note under the table for teams without a feedback repository, and wire clicks through `initAgents(section, onFeedback)` / `initSkills(section, onFeedback)` to `api.openURL`, showing a refusal beside the button.
  - **Design:** "The frontend opens only GitHub issue addresses"
  - **Files:** `desktop/frontend/src/pages/agents.js`, `desktop/frontend/src/pages/skills.js`, `desktop/frontend/src/pages/feedback.js`, `desktop/frontend/src/pages/pages.test.js`, `desktop/frontend/src/main.js`, `desktop/frontend/index.html`, `desktop/frontend/src/style.css`
  - **Test first:** `pages.test.js` cases expecting an `Owner` cell with `team owner` for an inherited owner, `No owner` when empty, a `data-feedback` button with an `aria-label` naming the item, and a disabled button plus note for a team without feedback; they fail today.
  - **Definition of done:**
    - Owner text and data attributes are escaped; no page source names a sample team.
    - A refused open shows `refusing to open …` beside that row's button through `actionError`.
    - The button is reachable by keyboard and has a visible focus ring in light and dark themes.
    - verify: `task desktop:test`

**Validation:** automated: `task desktop:test` · manual: Scenario 5 in `task desktop:dev` on macOS; Linux build through `scripts/linux-docker.sh`

## Slice 5: `analyze-session` turns a Kiro chat into owner issues, and the README explains it all

**Depends on:** Slice 3

After this slice every role agent has the `analyze-session` skill, which reads a past session and reports recommendations grouped by owner with one prefilled issue per owner, and the README documents the example team, owners, feedback and session analysis.

- [ ] T12 The `analyze-session` skill reads a Kiro session and groups recommendations by owner
  - **Objective:** Add `skills/analyze-session/SKILL.md.tmpl` that finds a session (`/session-id` or a listing of `<Kiro home>/sessions/*/sess_*/session.json`), reads its `messages.jsonl` for failed or denied tool calls, permission prompts, repeated lookups, unused skills and high context use, maps findings to an agent, skill, steering or hook with its owner from the rendered table, and offers one `/feedback`-format link per owner; list it in every role agent's `resources`.
  - **Design:** "plan:decision:d8"
  - **Files:** `teams/product-team/skills/analyze-session/SKILL.md.tmpl`, `teams/product-team/agents/product-owner/agent.json.tmpl`, `teams/product-team/agents/engineer/agent.json.tmpl`, `teams/product-team/agents/customer/agent.json.tmpl`, `teams/product-team/agents/qa/agent.json.tmpl`, `teams/product-team/agents/ux-designer/agent.json.tmpl`, `testdata/golden/product-team/rendered/skills/analyze-session/SKILL.md`, `testdata/golden/product-team/rendered/agents/product-owner.json`, `testdata/golden/product-team/rendered/agents/engineer.json`, `testdata/golden/product-team/rendered/agents/customer.json`, `testdata/golden/product-team/rendered/agents/qa.json`, `testdata/golden/product-team/rendered/agents/ux-designer.json`
  - **Trivial:** golden comparison and `kiro-cli agent validate` in `go test ./internal/golden` cover the rendered skill and resources; behaviour is proved by Scenario 6.
  - **Definition of done:**
    - The skill names the session fields it reads and tells the person when a session has none of them instead of reporting nothing.
    - It sends nothing itself and uses no shell.
    - verify: `go test ./cmd/kirokit/... ./internal/golden`

- [ ] T13 The README explains role agents, owners, feedback and session analysis
  - **Objective:** Add a "Role agents and feedback" section, the `owner`/`owners`/`feedback` fields in the manifest reference, the three-team "Example teams" bullet, the app's Owner column and Feedback button, and how to copy `/feedback` to another team.
  - **Design:** none
  - **Files:** `README.md`
  - **Trivial:** documentation; checked by Scenario 7 against the code.
  - **Definition of done:**
    - Every field, template function (`owners`, `feedbackRepo`) and command named in the README exists with that name.
    - verify: `node scripts/sync-agent-assets.mjs --check`

**Validation:** automated: `go test ./...` · `task ci` · manual: Scenario 6 in a logged-in `kiro-cli`

## Program design

The Plan's contracts hold the design. The calls left open are listed under *Decided while outlining*.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Owner and feedback fields, `OwnerOf` | `internal/manifest/manifest.go` | `ExternalSkill.Validate` error style |
| Team inventory, owner key and subagent checks | `internal/render/render.go`, `internal/render/context.go` | `validateOutput`, `externalskills.Resolve` |
| `owners`, `feedbackRepo` | `internal/render/funcs.go` | `mcpVersionFunc` |
| Catalog owners | `desktop/catalog.go` | `readTeam`, `manifest.OwnerOf` |
| Issue address and allowlisted open | `desktop/frontend/src/pages/feedback.js`, `desktop/frontend/src/api.js` | `esc`, `actionError` |
| Role team content | `teams/product-team/` | `platform-team` agent, prompt, skill and steering shapes |

## End-to-end scenarios

### Scenario 1: A product owner consults the customer agent

**Runs after:** slice 2

**Proves:** Done means 1

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `task build` | exit 0 |
| 2 | `export KIRO_HOME=$(mktemp -d) && ./kirokit install product-team` | prints `installed product-team into <KIRO_HOME>` |
| 3 | `ls $KIRO_HOME/agents` | `customer.json engineer.json product-owner.json qa.json ux-designer.json` |
| 4 | Run `kiro-cli chat --agent product-owner --v3` and ask "Ask the customer agent what frustrates them most about onboarding, then summarise its answer." | the chat shows a `subagent` call to `customer` without a permission prompt, and the product owner reports the customer's answer |
| 5 | In the same chat ask "Ask the engineer agent whether this is feasible." | Kiro asks permission before calling `engineer` (not in `trustedAgents`) |

### Scenario 2: A spec's requirements and design are reviewed by role

**Runs after:** slice 2

**Proves:** Done means 2

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | With Scenario 1's Kiro home, `cat $KIRO_HOME/steering/spec-review-gates.md` | the steering with `inclusion: always`, naming `product-owner` for requirements and `engineer` for design |
| 2 | In a scratch workspace run `kiro-cli chat --agent engineer --v3` (the role that can write files) and ask it to write a spec for "export a team's agents as CSV" | after `requirements.md` the agent calls `product-owner` as a subagent for review and folds its answer in; after `design.md` it calls `engineer` |
| 3 | If CLI v3 offers no spec workflow, ask "write requirements for …" instead | the steering still triggers the product owner review; the finding is recorded for the person before Review |

### Scenario 3: Owners are validated and inherited

**Runs after:** slice 1

**Proves:** Done means 3

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `go test ./internal/manifest ./internal/render` | ok |
| 2 | Copy `teams/product-team` to a temp teams root with `owners.agents.qa: "jane"` and run `./kirokit render` on it | fails with `owners.agents.qa: "jane" is not a GitHub handle (write @jane)` |
| 3 | `./kirokit test` | every team ok |

### Scenario 4: `/feedback` in Kiro drafts the owner's issue

**Runs after:** slice 3

**Proves:** Done means 4, Done means 5

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Reinstall `product-team` into the Scenario 1 Kiro home and start `kiro-cli chat --agent product-owner --v3` | chat starts |
| 2 | Type `/feedback the customer agent ignored my persona notes` | the agent prints a title, a body naming `product-team` and `customer` with `**Owner:** @mavogel`, and a link starting `https://github.com/mavogel/kirokit/issues/new?title=` with `labels=feedback%2Cteam%3Aproduct-team%2Cagent%3Acustomer` (or the unencoded commas) |
| 3 | Open the link in a browser | GitHub's new-issue form shows the prefilled title and body; nothing was submitted |

### Scenario 5: The desktop app shows owners and opens a feedback issue

**Runs after:** slice 4

**Proves:** Done means 3, Done means 4, Done means 5

**Screens:** `artifacts/agents-skills-feedback.html#agents-page`, `artifacts/agents-skills-feedback.html#skills-page`, `artifacts/agents-skills-feedback.html#opened-issue`

| Step | Action | Expected |
|---|---|---|
| 1 | `task desktop:test` | all tests pass |
| 2 | `task desktop:dev` with this repository as the clone, open **Agents** | each agent row shows its owner; the five product-team agents show `@mavogel`, and inherited owners say *team owner* |
| 3 | Press **Feedback** on `product-owner` | the system browser opens `https://github.com/mavogel/kirokit/issues/new?…` with title `Feedback: product-team agent product-owner`, a body with `**Owner:** @mavogel`, and the three labels |
| 4 | Open **Skills** | `analyze-session` shows its owner and an enabled Feedback button; `platform-team` and `data-team` rows show disabled buttons and the note that their manifest names no feedback repository |
| 5 | Tab to a Feedback button in light and dark theme | the focus ring is visible and Enter opens the issue |

### Scenario 6: Session analysis suggests improvements by owner

**Runs after:** slice 5

**Proves:** Done means 6

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Reinstall `product-team`, run a short chat in which a tool call is denied, then `/quit` | a session folder appears under `$KIRO_HOME/sessions` (or `~/.kiro/sessions`) |
| 2 | Start `kiro-cli chat --agent qa --v3` and ask "Analyze my last session with the analyze-session skill." | the agent lists sessions or uses the given id, reads `messages.jsonl` with file reads only, and reports recommendations grouped by owner, each naming an agent, skill, steering or hook |
| 3 | Accept the offer of issues | one prefilled `issues/new` link per owner in the `/feedback` format; nothing is sent |

### Scenario 7: The README explains the role setup and the feedback loop

**Runs after:** slice 5

**Proves:** Done means 7

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `grep -n "Role agents and feedback" README.md` | the section exists |
| 2 | Read the section and the manifest reference | they explain installing `product-team`, `owner`/`owners`/`feedback.github`, `/feedback`, the app's Owner column and Feedback button, and `analyze-session`, and the "Example teams" bullet names three teams |

## Commands the build runs

<!-- ql:commands -->
- `go test ./...` · slice 1, slice 2, slice 3, slice 5, project checks
- `go vet ./...` · slice 1, project checks
- `task lint` · slice 1, project checks
- `task build` · slice 2, T7, scenario 1, project checks
- `task desktop:test` · slice 4, T10, T11, scenario 5, project checks
- `task ci` · slice 5
- `go test ./internal/manifest` · T1
- `go test ./internal/render` · T2, T3
- `go test ./cmd/kirokit/... ./internal/golden` · T4, T6, T12
- `go -C desktop test ./... && go test ./cmd/kirokit/... ./internal/golden` · T5
- `go test ./cmd/kirokit/...` · T8
- `go -C desktop test ./...` · T9
- `node scripts/sync-agent-assets.mjs --check` · T13, project checks
- `export KIRO_HOME=$(mktemp -d) && ./kirokit install product-team` · scenario 1
- `ls $KIRO_HOME/agents` · scenario 1
- `go test ./internal/manifest ./internal/render` · scenario 3
- `./kirokit test` · scenario 3
- `grep -n "Role agents and feedback" README.md` · scenario 7
<!-- /ql:commands -->
