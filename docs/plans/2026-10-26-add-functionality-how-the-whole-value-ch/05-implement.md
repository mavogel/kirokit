# Cross-role agent harness with feedback loop · the build

## Summary

<!-- ql:summary -->
11 of 13 tasks · 4 of 5 slices

Building now: started 26 Oct, 13:07, 28 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | Every team can name owners and a feedback repository | After this slice `manifest.yaml` accepts `owner`, `owners` and `feedback.github`, rejects malformed handles and unknown agent or skill keys, and templates can read the owner table and the feedback repository. | done |
| 2 | A team can install five role agents that consult each other | After this slice `kirokit install product-team` gives the product owner, engineer, customer, QA and UX/UI designer agents, each able to call the other four, with a steering that asks the product owner to review requirements and the engineer to review a design. | done |
| 3 | `/feedback` in Kiro drafts a prefilled issue for the owner | After this slice a person in any `product-team` chat types `/feedback <what went wrong>` and gets a title, a body that @-mentions the owner, and a percent-encoded `issues/new` link in `mavogel/kirokit`, with nothing sent. | done |
| 4 | The desktop app shows owners and opens a feedback issue | After this slice the Agents and Skills pages show each owner and a Feedback button that opens the prefilled GitHub issue in the browser; a team without `feedback.github` shows its buttons off with the reason, and any address that is not a GitHub issues form is refused. | done |
| 5 | `analyze-session` turns a Kiro chat into owner issues, and the README explains it all | After this slice every role agent has the `analyze-session` skill, which reads a past session and reports recommendations grouped by owner with one prefilled issue per owner, and the README documents the example team, owners, feedback and session analysis. | next |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- The Plan · D2 and the render contract: Kiro CLI 2.34.0 v3 still prompted for a subagent listed only in trustedAgents (live run, slice 2), while a permissions rule {capability: subagent, match: [agent names], effect: allow} skipped the prompt and left the engineer asking. Each role agent now also carries that rule for its non-writing peers, and render checks the rule's match names against the team's agents like availableAgents and trustedAgents. Affects T3 (check extended) and T5 (agent configs); files internal/render/render.go, internal/render/render_test.go, teams/product-team/agents/*/agent.json.tmpl. The person's decisions are unchanged.
- Slice 1 · Regenerating `data-team`'s golden also updated two stale MCP pins in its `context.json` (`context7-mcp` 4.3.0 → 4.4.2, `mcp-remote` 0.14.3 → 0.15.1), already wrong before this change because `go test ./...` compared only `platform-team`; the new every-team test now catches this; nothing you approved changes.
- Slice 2 · `go test ./desktop/...` cannot run from the root (desktop is its own module); verify lines use `go -C desktop test ./...`; nothing you approved changes.
<!-- /ql:decisions -->

## Checks

- T1 · 381b399 · `go test ./internal/manifest` · exit 0 · 0.4s
- T2 · 381b399 · `go test ./internal/render` · exit 0 · 0.7s
- T3 · 381b399 · `go test ./internal/render` · reused: nothing changed since it passed at 13:10
- T4 · 381b399 · `go test ./cmd/kirokit/... ./internal/golden` · exit 0 · 4.2s
- T3 · ea2bc5f · `go test ./internal/render` · exit 0 · 0.7s
- T5 · ea2bc5f · `go -C desktop test ./... && go test ./cmd/kirokit/... ./internal/golden` · exit 0 · 4.9s
- T6 · ea2bc5f · `go test ./cmd/kirokit/... ./internal/golden` · exit 0 · 0.2s
- T7 · ea2bc5f · `task build` · exit 0 · 0.3s
- slice 1 · ea2bc5f · `go test ./...` · exit 0 · 8.3s
- slice 1 · ea2bc5f · `go vet ./...` · exit 0 · 0.2s
- slice 1 · ea2bc5f · `task lint` · exit 0 · 0.8s
- slice 2 · ea2bc5f · `go test ./...` · reused: nothing changed since it passed at 13:22
- slice 2 · ea2bc5f · `task build` · reused: nothing changed since it passed at 13:22
- T8 · ec8394b · `go test ./cmd/kirokit/...` · exit 0 · 4.4s
- slice 3 · ec8394b · `go test ./...` · exit 0 · 2.2s
- T9 · 2d1b288 · `go -C desktop test ./...` · exit 0 · 0.3s
- T10 · 2d1b288 · `task desktop:test` · exit 0 · 3.1s
- T11 · 2d1b288 · `task desktop:test` · reused: nothing changed since it passed at 13:35
- slice 4 · 2d1b288 · `task desktop:test` · reused: nothing changed since it passed at 13:35

## Known limits

<!-- ql:limits -->
- T11 · The Feedback button's real browser open was run only on macOS; Linux and Windows were not run locally (Docker was off, no Windows machine), and CI's tests there fake BrowserOpenURL.
<!-- /ql:limits -->

## Deviations

- Plan (amended): d2 and the render contract: Kiro CLI 2.34.0 v3 still prompted for a subagent listed only in trustedAgents (live run, slice 2), while a permissions rule {capability: subagent, match: [agent names], effect: allow} skipped the prompt and left the engineer asking. Each role agent now also carries that rule for its non-writing peers, and render checks the rule's match names against the team's agents like availableAgents and trustedAgents. Affects T3 (check extended) and T5 (agent configs); files internal/render/render.go, internal/render/render_test.go, teams/product-team/agents/*/agent.json.tmpl. The person's decisions are unchanged.

- T4 (tactical): regenerating `data-team`'s golden also updated two stale MCP pins in its `context.json` (`context7-mcp` 4.3.0 → 4.4.2, `mcp-remote` 0.14.3 → 0.15.1), already wrong before this change because `go test ./...` compared only `platform-team`; the new every-team test now catches this; nothing you approved changes
- T5 (tactical): `go test ./desktop/...` cannot run from the root (desktop is its own module); verify lines use `go -C desktop test ./...`; nothing you approved changes
- Scenario 2 (tactical): the spec run uses the `engineer` agent, because the product owner has no write tool and cannot write spec files; nothing you approved changes

## Plan amended

### Plan · amended

d2 and the render contract: Kiro CLI 2.34.0 v3 still prompted for a subagent listed only in trustedAgents (live run, slice 2), while a permissions rule {capability: subagent, match: [agent names], effect: allow} skipped the prompt and left the engineer asking. Each role agent now also carries that rule for its non-writing peers, and render checks the rule's match names against the team's agents like availableAgents and trustedAgents. Affects T3 (check extended) and T5 (agent configs); files internal/render/render.go, internal/render/render_test.go, teams/product-team/agents/*/agent.json.tmpl. The person's decisions are unchanged.

Amended by the agent on 2026-10-26; changed 03-plan.md, 04-outline.md.

Reasoning: the Plan left open what a `subagent` permission rule's `match` compares against, and relied on `trustedAgents` to skip the prompt. With `product-team` installed into a scratch Kiro home, `kiro-cli chat --no-interactive --agent product-owner --v3` (kiro-cli 2.34.0) called `customer` and was denied twice with "tool permission approval is not supported in non-interactive mode", although `customer` was in `trustedAgents`. Adding `{ "capability": "subagent", "match": ["customer", "qa", "ux-designer"], "effect": "allow" }` to the scratch copy made the same call complete without a prompt, while a call to `engineer` (not in the rule) was still held for permission. So `match` compares against the target agent's name, and the rule is what implements "consulting non-writers needs no prompt; calling the engineer asks first". `trustedAgents` stays, since Kiro's docs name it and it costs nothing. Render now also checks the rule's names, because a typo there would leave an agent prompting forever.

## Live Kiro runs

- **Scenario 1 (Done means 1):** `kiro-cli chat --no-interactive --agent product-owner --v3 "Ask the customer agent what frustrates them most about onboarding…"` against the committed config: `[tool] Sub-agent: customer` → `Completed`, and the product owner summarised the customer persona's answer. Log: [records/live-s1-product-owner-consults-customer.txt](records/live-s1-product-owner-consults-customer.txt).
- **Scenario 2 (Done means 2):** Kiro CLI v3 runs specs in a chat (`--mode spec`). `kiro-cli chat --no-interactive --agent engineer --v3 --mode spec --trust-tools=fs_read,fs_write "Write a quick spec … export a team's agents as a CSV file"` wrote `requirements.md`, `design.md` and `tasks.md`; after requirements it called `product-owner` (which itself consulted `qa` and `customer`) and folded the gaps in; the engineer reviewed `design.md` itself, as the steering says when the writer is the reviewer. Log: [records/live-s2-spec-mode-review-gates.txt](records/live-s2-spec-mode-review-gates.txt). The Plan's assumption "Kiro CLI v3 runs the spec workflow in a chat" is confirmed.
- **Scenario 4 (Done means 4, 5):** `kiro-cli chat --no-interactive --agent product-owner --v3 "/feedback the customer agent ignored my persona notes …"` printed a title, a body with `**Team:** product-team`, `**Agent:** customer`, `**Owner:** @mavogel`, and `https://github.com/mavogel/kirokit/issues/new?title=…&body=…&labels=feedback,team:product-team,agent:customer`; decoded, title and body round-trip and the body is about 400 characters. The first run wrote "@mavogel (team owner)" on the Owner line; the prompt now asks for the bare handle and the rerun shows it. No tool was called. Log: [records/live-s4-feedback-prompt.txt](records/live-s4-feedback-prompt.txt).
- **Scenario 5 (Done means 3, 4, 5):** `go tool wails dev` with a throwaway `HOME`, cloning a tag-free local copy of this branch (so `product-team` is present), driven by headless Chrome at `http://localhost:34115`. Agents shows `@mavogel` for `product-owner` (its own entry) and `@mavogel team owner` for the rest; Skills shows the team owner on every skill; the four `data-team`/`platform-team` agent buttons are disabled with the note "… names no feedback repository in its manifest.yaml, so its Feedback buttons are off." Pressing Feedback on `product-owner`, and Tab + Enter on `engineer` (focus ring `solid 2px`, visible in dark theme), asked Wails to open `https://github.com/mavogel/kirokit/issues/new?…` with title `Feedback: product-team agent product-owner`, labels `feedback,team:product-team,agent:product-owner` and `**Owner:** @mavogel` in the body. The first run found a defect the unit tests could not: Wails v2.17's `BrowserOpenURL` refused the address ("shell metacharacters not allowed") because `encodeURIComponent` leaves `*` (from `**Team:**`) unescaped; `feedbackURL` now also encodes `! ' ( ) * ~`, a test mirrors Wails' character rule, and the rerun logged no refusal. The browser tab itself was not observed. Evidence: [records/s5/report.json](records/s5/report.json), [records/s5/agents-light.png](records/s5/agents-light.png), [records/s5/skills-light.png](records/s5/skills-light.png), [records/s5/agents-dark-focus.png](records/s5/agents-dark-focus.png).
