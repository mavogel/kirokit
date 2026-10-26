# Cross-role agent harness with feedback loop

<!-- ql:links -->
Task: `docs/plans/2026-10-26-add-functionality-how-the-whole-value-ch/`
<!-- /ql:links -->

## Why the change

kirokit's example teams modelled only engineers working alone, so a product owner or designer unhappy with an agent or skill had no way back to whoever maintains it; now a team can install five role agents that consult each other, name an owner for every agent and skill, and send feedback to that owner as a prefilled GitHub issue from Kiro or the desktop app.

## What a reviewer must know

- **No migration.** `owner`, `owners` and `feedback` are optional; existing config repositories render byte for byte as before (only `context.json` changes where a team adds a line). `go test ./...` now compares every team's goldens, which immediately caught a stale `data-team` golden.
- **Kiro CLI v3 needs permission rules, not just `trustedAgents`.** Live runs showed `trustedAgents` still prompts and loading a skill prompts too, so each role agent carries `subagent` and `skill` allow rules; calling the `engineer` still asks. Render rejects subagent names that are not agents of the team.
- **The app opens the system browser for the first time,** only for `https://github.com/<owner>/<repo>/issues/new?…`. Follow-ups from the independent review: on the app's 720px minimum width the Agents table now overflows by 294px (88px before), reachable by horizontal scroll and Tab; and the spec review gates fire only when an agent writes a spec (`--mode spec`), since a steering cannot force a review.

## What changed

Three new inputs in `manifest.yaml` drive everything else; render checks them against the team's real agents and skills:

```yaml
owner: "@mavogel"              # team owner (optional)
owners:
  agents: { product-owner: "@mavogel" }
  skills: { analyze-session: "@x" }   # vendored skills fall back to the team owner
feedback:
  github: mavogel/kirokit      # where feedback issues open
```

One owner rule, `Manifest.OwnerOf(kind, name)`, feeds both the templates and the desktop app:

```diff
 internal/manifest   + Owner, Owners, Feedback, OwnerOf, validation
 internal/render     + team inventory, owner-key and subagent-name checks
                     + template functions: owners, feedbackRepo
 desktop/catalog.go  + Team.Feedback, Agent/Skill.Owner, OwnerFromTeam
 desktop/frontend    + feedbackURL, api.openURL (GitHub issue forms only)
                     ~ Agents and Skills pages: Owner column, Feedback button
```

The new example team is content only (no MCP servers, no secrets):

```diff
+teams/product-team/
+├── manifest.yaml                     owner, owners, feedback.github
+├── agents/{product-owner,engineer,customer,qa,ux-designer}/agent.json.tmpl
+│                                      subagent tool, peers in availableAgents,
+│                                      allow rules for non-writing peers and the skill
+├── steerings/spec-review-gates.md.tmpl   PO reviews requirements, engineer reviews design
+├── prompts/feedback.md.tmpl          /feedback → prefilled issue link, owner @-mentioned
+└── skills/analyze-session/SKILL.md.tmpl  session → recommendations by owner → issue per owner
```

## Screens compared

<!-- ql:screens -->
- **agents-skills-feedback#agents-page** · 1280 × 900 px · light · captured · stale
  - Independent review: No comparison recorded.
  - Refresh: recorded build changed
- **agents-skills-feedback#skills-page** · 1280 × 900 px · light · unavailable: Page capture timed out: selector #teams-grid .card unavailable · stale
  - Independent review: No comparison recorded.
  - Refresh: recorded build changed
- **agents-skills-feedback#opened-issue** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
5 slices · 13 tasks · 13 commits · 42 files · +1618 −40

**Slice 1 · Every team can name owners and a feedback repository** · T1, T2, T3, T4 · 11 files · +519 −5

- `cmd/kirokit/commands/root_test.go` · T4, T5 · +31 −1
- `internal/manifest/manifest.go` · T1 · +95 −2
- `internal/manifest/manifest_test.go` · T1 · +101 −0
- `internal/render/context.go` · T2 · +6 −0
- `internal/render/funcs.go` · T2 · +42 −0
- `internal/render/render.go` · T2, T3 · +111 −0
- `internal/render/render_test.go` · T2, T3 · +124 −0
- `teams/data-team/manifest.yaml` · T4 · +1 −0
- `teams/platform-team/manifest.yaml` · T4 · +4 −0
- `testdata/golden/data-team/context.json` · T4 · +3 −2
- `testdata/golden/platform-team/context.json` · T4 · +1 −0

**Slice 2 · A team can install five role agents that consult each other** · T5, T6, T7 · 15 files · +457 −2

- `desktop/catalog_test.go` · T5, T9 · +92 −2
- `teams/product-team/agents/customer/agent.json.tmpl` · T5, T6, T12 · new · +23 −0
- `teams/product-team/agents/engineer/agent.json.tmpl` · T5, T6, T12 · new · +29 −0
- `teams/product-team/agents/product-owner/agent.json.tmpl` · T5, T6, T12 · new · +23 −0
- `teams/product-team/agents/qa/agent.json.tmpl` · T5, T6, T12 · new · +23 −0
- `teams/product-team/agents/ux-designer/agent.json.tmpl` · T5, T6, T12 · new · +23 −0
- `teams/product-team/manifest.yaml` · T5 · new · +19 −0
- `teams/product-team/steerings/spec-review-gates.md.tmpl` · T6 · new · +34 −0
- `testdata/golden/product-team/context.json` · T5 · new · +26 −0
- `testdata/golden/product-team/rendered/agents/customer.json` · T5, T6, T12 · new · +25 −0
- `testdata/golden/product-team/rendered/agents/engineer.json` · T5, T6, T12 · new · +31 −0
- `testdata/golden/product-team/rendered/agents/product-owner.json` · T5, T6, T12 · new · +25 −0
- `testdata/golden/product-team/rendered/agents/qa.json` · T5, T6, T12 · new · +25 −0
- `testdata/golden/product-team/rendered/agents/ux-designer.json` · T5, T6, T12 · new · +25 −0
- `testdata/golden/product-team/rendered/steering/spec-review-gates.md` · T6 · new · +34 −0

**Slice 3 · `/feedback` in Kiro drafts a prefilled issue for the owner** · T8 · 2 files · +91 −0

- `teams/product-team/prompts/feedback.md.tmpl` · T8 · new · +44 −0
- `testdata/golden/product-team/rendered/prompts/feedback.md` · T8, T12 · new · +47 −0

**Slice 4 · The desktop app shows owners and opens a feedback issue** · T9, T10, T11 · 11 files · +368 −31

- `desktop/catalog.go` · T9 · +31 −13
- `desktop/frontend/index.html` · T11 · +4 −4
- `desktop/frontend/src/api.js` · T10 · +27 −0
- `desktop/frontend/src/api.test.js` · T10 · +36 −0
- `desktop/frontend/src/main.js` · T11 · +15 −2
- `desktop/frontend/src/pages/agents.js` · T11 · +14 −6
- `desktop/frontend/src/pages/feedback.js` · T10, T11 · new · +80 −0
- `desktop/frontend/src/pages/feedback.test.js` · T10 · new · +71 −0
- `desktop/frontend/src/pages/pages.test.js` · T11 · +71 −0
- `desktop/frontend/src/pages/skills.js` · T11 · +16 −6
- `desktop/frontend/src/style.css` · T11 · +3 −0

**Slice 5 · `analyze-session` turns a Kiro chat into owner issues, and the README explains it all** · T12, T13 · 3 files · +183 −2

- `README.md` · T13 · +28 −2
- `teams/product-team/skills/analyze-session/SKILL.md.tmpl` · T12 · new · +76 −0
- `testdata/golden/product-team/rendered/skills/analyze-session/SKILL.md` · T12 · new · +79 −0
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. A team can install a ready example of cooperating role agents.** Scenario 1: A product owner consults the customer agent, after slice 2 · passed · `evidence/it1-scenario-1.txt`
- **2. The spec workflow has review gates by role.** Scenario 2: A spec's requirements and design are reviewed by role, after slice 2 · passed · `evidence/it1-scenario-2.txt`
- **3. Every agent and skill of a team names its owner.** Scenario 3: Owners are validated and inherited, after slice 1 · Scenario 5: The desktop app shows owners and opens a feedback issue, after slice 4 · passed · `evidence/it1-scenario-5.txt`
- **4. A non-engineer can send feedback on a skill or agent from where they work.** Scenario 4: `/feedback` in Kiro drafts the owner's issue, after slice 3 · Scenario 5: The desktop app shows owners and opens a feedback issue, after slice 4 · passed · `evidence/it1-scenario-4.txt`
- **5. The feedback reaches the owner.** Only you can confirm: the issue reads well to an owner · waits for your confirmation
- **6. A session-analysis skill suggests harness improvements from a Kiro chat.** Scenario 6: Session analysis suggests improvements by owner, after slice 5 · passed · `evidence/it1-scenario-6.txt`
- **7. The README explains the role setup and the feedback loop.** Scenario 7: The README explains the role setup and the feedback loop, after slice 5 · passed · `evidence/it1-scenario-7.txt`
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No QA or UX review gate in the spec steering;** the QA and UX agents are consulted on request.
- **No Kiro hook for spec phases;** Kiro offers none.
- **No CLI `kirokit feedback` command and no GitHub token;** the person submits the issue.
- **No feedback prompt for `platform-team` and `data-team`;** they get an `owner` line only, and the README shows how to copy the prompt.
- **No usage or credit metrics** from sessions.
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
- T11 · The Feedback button's real browser open was run only on macOS; Linux and Windows were not run locally (Docker was off, no Windows machine), and CI's tests there fake BrowserOpenURL.
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
Security review, Project checks and Code review ran: 6 of 6 points of Done means passed; 1 is yours to confirm.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The product owner consults the customer agent without a prompt
- Do: build, install `product-team` into a throwaway Kiro home, and ask the product owner to consult the customer.
- Run: `task build && export KIRO_HOME=$(mktemp -d) && ./kirokit install product-team && kiro-cli chat --agent product-owner --v3 "Ask the customer agent what frustrates them most about onboarding."`
- Expect: a `Sub-agent: customer` call with no permission prompt, then the product owner summarising the customer persona's answer.
- Saw: the call completed and the answer was summarised — `evidence/it1-scenario-1.txt`

#### 2. `/feedback` drafts an issue for the owner
- Do: in that chat, type `/feedback the customer agent ignored my persona notes`, then open the printed link while signed in to GitHub.
- Expect: GitHub's new-issue form in `mavogel/kirokit`, prefilled with a title naming `product-team agent customer`, a body with `**Owner:** @mavogel`, and nothing submitted.
- Saw: the link decodes to that title, body and labels; signed out, GitHub keeps the whole address through its sign-in redirect. The signed-in form itself was not viewed — `evidence/it1-scenario-4.txt`

#### 3. The desktop app shows owners and opens the issue
- Do: run the app from source, clone a repository that has `product-team`, open Agents and press Feedback on `product-owner`.
- Run: `task desktop:dev`
- Expect: owners in the Owner column (*team owner* under inherited ones), Feedback opening the prefilled GitHub issue in your browser, and buttons off for teams without `feedback.github`.
- Saw: owners and disabled buttons as expected, and Wails' real open path received `https://github.com/mavogel/kirokit/issues/new?…` (recorded by a stand-in `open`, so no tab opened) — `evidence/it1-scenario-5.txt`

### Not verified

- How the prefilled issue reads to an owner on GitHub's signed-in form (Done means 5); only you can judge that.
- The real browser open on Linux and Windows (Docker was off, no Windows machine); CI's tests there fake `BrowserOpenURL`.

### Docs

- `README.md`: a new "Role agents and feedback" section, `owner`/`owners`/`feedback` in the manifest reference, the three example teams, and the app's Agents and Skills behaviour.

## Security

- Nothing severe, Medium or Low found. The browser path accepts only a parsed `https://github.com/<owner>/<repo>/issues/new` address, `feedback.github` is validated at load (including `.`/`..`), and the worst reachable outcome is a prefilled issue form the person submits themselves; `analyze-session` only reads the person's own session files and sends nothing.

## Accepted with open items

<!-- ql:accepted -->
None: the build passed its checks.
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- The example team is named `product-team` — it names the cross-functional product group, beside `platform-team` and `data-team`. (Plan)
- Example owners are `@mavogel` only — any other handle in a public example would notify a real GitHub user. (Plan)
- The example feedback repository is `mavogel/kirokit` — the repository the README already uses as the example config source; teams change one line. (Plan)
- Issues carry three labels: `feedback`, `team:<team>` and `agent:<name>` or `skill:<name>` — Done means 5 asks for team and item labels; GitHub applies URL labels only for people who may triage, so the title and body name the team and item too. (Plan)
- `trustedAgents` excludes the engineer — it can write files and run shell commands, so calling it should still ask. (Plan)
- `go test ./...` checks every team's goldens — today it checks only `platform-team`, so a new team's drift would slip through. (Plan)
- The README gains a "Role agents and feedback" section — it covers `product-team`, the `owner`/`owners`/`feedback` fields in the manifest reference, `/feedback`, the app's Owner column and Feedback button, and `analyze-session`, and the "Example teams" bullet lists three teams (Done means 7; AGENTS.md requires README updates with manifest and template-function changes). (Plan)
- `analyze-session` reads files, no shell — the role agents' existing read permissions cover `<Kiro home>/sessions`, and listing sessions needs no `kiro-cli` call. (Plan)
- D2 and the render contract: Kiro CLI 2.34.0 v3 still prompted for a subagent listed only in trustedAgents (live run, slice 2), while a permissions rule {capability: subagent, match: [agent names], effect: allow} skipped the prompt and left the engineer asking. Each role agent now also carries that rule for its non-writing peers, and render checks the rule's match names against the team's agents like availableAgents and trustedAgents. Affects T3 (check extended) and T5 (agent configs); files internal/render/render.go, internal/render/render_test.go, teams/product-team/agents/*/agent.json.tmpl. The person's decisions are unchanged. (the Plan, while building)
- Regenerating `data-team`'s golden also updated two stale MCP pins in its `context.json` (`context7-mcp` 4.3.0 → 4.4.2, `mcp-remote` 0.14.3 → 0.15.1), already wrong before this change because `go test ./...` compared only `platform-team`; the new every-team test now catches this; nothing you approved changes. (slice 1)
- `go test ./desktop/...` cannot run from the root (desktop is its own module); verify lines use `go -C desktop test ./...`; nothing you approved changes. (slice 2)
- Every role agent also carries `{ "capability": "skill", "match": ["analyze-session"], "effect": "allow" }`, because Kiro CLI v3 asked permission to load the skill (live run) and, without it, the agent improvised and never reported; nothing you approved changes. (slice 5)
- `feedbackURL` percent-encodes `! ' ( ) * ~` on top of `encodeURIComponent`, because Wails v2.17's `BrowserOpenURL` refuses them as shell metacharacters (found driving the real app); nothing you approved changes. (slice 4)
- `feedback.github` also rejects `.` or `..` as the repository name, so the address cannot step out of `owner/repo`; nothing you approved changes. (slice 1)
- Outline bookkeeping after the build: T11 Files gains desktop/frontend/src/pages/feedback.js (the Owner and Feedback cells live there, shared by both pages), and T12 Files gains testdata/golden/product-team/rendered/prompts/feedback.md (the /feedback owner table now lists the analyze-session skill). No task, Done means or user decision changes. (the Plan, while building)
- Verification round 1 was recorded while 5 files changed, 5 of them code the run exercised; rerun it or keep it and say why here; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
- Still awaiting your confirmation: The issue reads well to an owner.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Not shipped yet. Branch `feat/add-functionality-how-whole-value`.
<!-- /ql:ship -->
