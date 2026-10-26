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
5 slices · 13 tasks · 14 commits · 42 files · +1618 −40

**Slice 1 · Every team can name owners and a feedback repository** · T1, T2, T3, T4 · 11 files · +519 −5

- [`cmd/kirokit/commands/root_test.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-3e4047a698e242b34a654d196f6f14f0c05d78c1f1e26f75192fa3c6dd51ea88) · T4, T5 · +31 −1
- [`internal/manifest/manifest.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-690e7b459192b78ffe0159d8c095aa41c210de896d67b92b7bf548a8f585f371) · T1 · +95 −2
- [`internal/manifest/manifest_test.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-9fa6b9b6ac53923eed836a50409016adb3eff24f78aee49c601106b960ff57ac) · T1 · +101 −0
- [`internal/render/context.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-ab3233924e21b127260088fa87b31af121c011a8e444b171e6014133a3492ff3) · T2 · +6 −0
- [`internal/render/funcs.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-9b7c3f12a190eda80756dc985f3b6fd6986db9f368e0a93ea25b2b30a5d7fe7a) · T2 · +42 −0
- [`internal/render/render.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-958be64db7636ed99935de3eb080951f6f57f38d5f7c9811de20dd01aaff94d5) · T2, T3 · +111 −0
- [`internal/render/render_test.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-66b32ffffb7326b9d3ae8dea89eaabaf7883f4ad9f046fc7958a7d04da6a7061) · T2, T3 · +124 −0
- [`teams/data-team/manifest.yaml`](https://github.com/mavogel/kirokit/pull/24/files#diff-f29078c1604d858c02e5083b7bea0d5c406efa312c8e92eb727054b76bb8e37a) · T4 · +1 −0
- [`teams/platform-team/manifest.yaml`](https://github.com/mavogel/kirokit/pull/24/files#diff-cf711c3e8e6fa44d82b63f67e54a0f880c210aac463e4bf568ac7c8f0ea581a5) · T4 · +4 −0
- [`testdata/golden/data-team/context.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-9133fa67013be8872d69f2296c2c573b1988a0b662313f988e0ecdae58536606) · T4 · +3 −2
- [`testdata/golden/platform-team/context.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-9c61cb8b262d7641125f02d92349077490fe09829f0ad1d9cfd949a83e054f3a) · T4 · +1 −0

**Slice 2 · A team can install five role agents that consult each other** · T5, T6, T7 · 15 files · +457 −2

- [`desktop/catalog_test.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-378958c2068bc01659303d9140791e1c09ab9606062fe2ee45af7e24cf1ca25a) · T5, T9 · +92 −2
- [`teams/product-team/agents/customer/agent.json.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-fb237d524fa80edaddf8e0c0b9ca3647e9fbd9dfd81a5fa02f29d45a01701d99) · T5, T6, T12 · new · +23 −0
- [`teams/product-team/agents/engineer/agent.json.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-ed347f60e6c0ddd161a01f5ca04dd1c6f392de312add072b56b99839e4ce52e0) · T5, T6, T12 · new · +29 −0
- [`teams/product-team/agents/product-owner/agent.json.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-37af02a405d03fee6c948cb8f7b713f884e1aa14215accd17faca3160147e1b2) · T5, T6, T12 · new · +23 −0
- [`teams/product-team/agents/qa/agent.json.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-79e66618383d37545241504ce67ded516f0d1dd2ccfebbc07a82baefb7e1b7e3) · T5, T6, T12 · new · +23 −0
- [`teams/product-team/agents/ux-designer/agent.json.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-589d369e0dc30ff9a37d6b4fe9a7c8c5170c800847b737b56d8059a550acbfca) · T5, T6, T12 · new · +23 −0
- [`teams/product-team/manifest.yaml`](https://github.com/mavogel/kirokit/pull/24/files#diff-b12311ab55faf81fd9e9f0493f22f2cdb939819842a05eb0a845dc35d130ebac) · T5 · new · +19 −0
- [`teams/product-team/steerings/spec-review-gates.md.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-4df1d4859b2c704e48eb280decde61ac29904e8252c99469ba2d7a6fbd520983) · T6 · new · +34 −0
- [`testdata/golden/product-team/context.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-2d3f7c702c3153680cf25203c85babee669eb944009051df489dd61afe97d421) · T5 · new · +26 −0
- [`testdata/golden/product-team/rendered/agents/customer.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-4f3c43a6fbab1139997243abfd4153d9bb3501cf9be914892596f6556acf44e6) · T5, T6, T12 · new · +25 −0
- [`testdata/golden/product-team/rendered/agents/engineer.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-e487f9aa02b0dbeb8a43cc38ebe78276ed2cd8c3edb12bf17d6814e8ddf0f57e) · T5, T6, T12 · new · +31 −0
- [`testdata/golden/product-team/rendered/agents/product-owner.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-4cd3e835f6bfd9f7a16413c5924d98a1b623e467e552327a872626e99b07ffe1) · T5, T6, T12 · new · +25 −0
- [`testdata/golden/product-team/rendered/agents/qa.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-ae818caeb6534ceac9c8a00737463868922ee884a8102a56bd6a13ecadd3cf51) · T5, T6, T12 · new · +25 −0
- [`testdata/golden/product-team/rendered/agents/ux-designer.json`](https://github.com/mavogel/kirokit/pull/24/files#diff-c4451a00d9752419ce05a3d9db202903773b232c02c4b58ddbe3fedd237a903d) · T5, T6, T12 · new · +25 −0
- [`testdata/golden/product-team/rendered/steering/spec-review-gates.md`](https://github.com/mavogel/kirokit/pull/24/files#diff-38074449f1dffb94c7b8dde82f86178b6109b94ede2917c26c0b2da1e177a8b3) · T6 · new · +34 −0

**Slice 3 · `/feedback` in Kiro drafts a prefilled issue for the owner** · T8 · 2 files · +91 −0

- [`teams/product-team/prompts/feedback.md.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-b8b236651093fa1bf054ec7bf0576643980fd9e1d7a634e3c8cb0cf9f00a99cd) · T8 · new · +44 −0
- [`testdata/golden/product-team/rendered/prompts/feedback.md`](https://github.com/mavogel/kirokit/pull/24/files#diff-0afcf6e5152a78bf85c573985f756e381372e6b9c0c9f5b9937136174bb019fa) · T8, T12 · new · +47 −0

**Slice 4 · The desktop app shows owners and opens a feedback issue** · T9, T10, T11 · 11 files · +368 −31

- [`desktop/catalog.go`](https://github.com/mavogel/kirokit/pull/24/files#diff-0834c61066a97277b759850557ffeffdc25424290e1a9350a78d7a3dccd29491) · T9 · +31 −13
- [`desktop/frontend/index.html`](https://github.com/mavogel/kirokit/pull/24/files#diff-d7021ef4c2c37f5c5a57b6c688d8b2233915c45bab1dd0e9db359c958c4bb3a3) · T11 · +4 −4
- [`desktop/frontend/src/api.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-1f29e5662697e24380064e530fe5825088c24b4d5bc3c757d0d884e5e8a60d4e) · T10 · +27 −0
- [`desktop/frontend/src/api.test.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-9fb087146e29f9abfd08f248f0d02945b9dac48fd111fbe49012d129ebf44c28) · T10 · +36 −0
- [`desktop/frontend/src/main.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-a59db5aebe85dd0ca14d4373fe7f1ffdb3718c69002286f0b33cde422d53f38b) · T11 · +15 −2
- [`desktop/frontend/src/pages/agents.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-f6b04ce648f569814ab7ca3fb0b17aab733baf114fbe24c7fc4ab660c728e851) · T11 · +14 −6
- [`desktop/frontend/src/pages/feedback.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-6e0efe30c2690fe86bd78641b77edef759ffca3219b29a037e02da9bece96da4) · T10, T11 · new · +80 −0
- [`desktop/frontend/src/pages/feedback.test.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-7c0bb9f479e1e872ddf96b7cee58e2355ba9137d6ca0ac772acc02a456e9e438) · T10 · new · +71 −0
- [`desktop/frontend/src/pages/pages.test.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-03c538c0ea841db73cb94e2ab514452b5add1f97253c33e2a34197facd87e52a) · T11 · +71 −0
- [`desktop/frontend/src/pages/skills.js`](https://github.com/mavogel/kirokit/pull/24/files#diff-d46bbf4034fae5f125f140402b50b9e0cc802a4480e967b49494aaca1874facd) · T11 · +16 −6
- [`desktop/frontend/src/style.css`](https://github.com/mavogel/kirokit/pull/24/files#diff-99b8fafca09e3b45f60088c58fc121b6a2cfee2f53fa7a7b6868e66c89de9c6d) · T11 · +3 −0

**Slice 5 · `analyze-session` turns a Kiro chat into owner issues, and the README explains it all** · T12, T13 · 3 files · +183 −2

- [`README.md`](https://github.com/mavogel/kirokit/pull/24/files#diff-b335630551682c19a781afebcf4d07bf978fb1f8ac04c6bf87428ed5106870f5) · T13 · +28 −2
- [`teams/product-team/skills/analyze-session/SKILL.md.tmpl`](https://github.com/mavogel/kirokit/pull/24/files#diff-7d55b546ecea2bee4f5a8154733e8e90e2f61fb70e87dcef458ceea76f7c30c3) · T12 · new · +76 −0
- [`testdata/golden/product-team/rendered/skills/analyze-session/SKILL.md`](https://github.com/mavogel/kirokit/pull/24/files#diff-fb50166837a8ba701271bffd9fcb4e8e0d295ba1b4008093bf90e723a4f812db) · T12 · new · +79 −0
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. A team can install a ready example of cooperating role agents.** Scenario 1: A product owner consults the customer agent, after slice 2 · passed · `evidence/it1-scenario-1.txt`
- **2. The spec workflow has review gates by role.** Scenario 2: A spec's requirements and design are reviewed by role, after slice 2 · passed · `evidence/it1-scenario-2.txt`
- **3. Every agent and skill of a team names its owner.** Scenario 3: Owners are validated and inherited, after slice 1 · Scenario 5: The desktop app shows owners and opens a feedback issue, after slice 4 · passed · `evidence/it1-scenario-5.txt`
- **4. A non-engineer can send feedback on a skill or agent from where they work.** Scenario 4: `/feedback` in Kiro drafts the owner's issue, after slice 3 · Scenario 5: The desktop app shows owners and opens a feedback issue, after slice 4 · passed · `evidence/it1-scenario-4.txt`
- **5. The feedback reaches the owner.** Only you can confirm: the issue reads well to an owner · confirmed by you: Confirmed in the App
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
