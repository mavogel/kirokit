# Cross-role agent harness with feedback loop · what was checked

## Summary

<!-- ql:summary -->
6 of 7 points assessed · 6 passed

Passed after 1 judge run.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Unverified:** On narrow windows the new Owner and Feedback columns of the Agents table sit off-screen. At the app's 720px minimum the main area overflows by 294px (88px before the change), and at 1000px the first Feedback button ends at 1001px; it stays reachable by horizontal scroll and Tab,
- **Unverified:** The review gates fire only when an agent writes a spec. Asked to draft requirements outside spec mode, the `qa` agent did not consult `product-owner`; the README already says a steering cannot force the review
- **Only you can confirm:** The issue reads well to an owner · confirmed · 26 Oct, 14:39
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: the desktop app opens the system browser at an address built from content of a cloned repository; the skill reads the person's local Kiro.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 6 checks · 2 reused
- **Independent review** · fresh helper
  - State: done
  - Result: done · 6 of 6 points · 1 yours to confirm · 2 screens stale
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: the desktop app opens the system browser at an address built from content of a cloned repository; the skill reads the person's local Kiro, the Plan says it is risky: the desktop app gains its first path that opens the system browser, with an address built from repository content; and the review gates and slice 2 and 4 is marked risky; say "get a second opinion" to have one.
- **Fix rounds** · the build session
  - State: done
  - Result: done · none needed
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
| Done means | Proof | |
|---|---|---|
| 1. A team can install a ready example of cooperating role agents | `kirokit install product-team` into a temporary Kiro home gave `customer`, `engineer`, `product-owner`, `qa`, `ux-designer`; `kiro-cli chat --agent product-owner --v3` ran `Sub-agent: customer` → Completed and reported Sam's answer. | passed |
| 2. The spec workflow has review gates by role | With `product-team` installed, `kiro-cli chat --agent engineer --v3 --mode spec` wrote `requirements.md`, had `product-owner` review it as a subagent and folded its gaps in before `design.md`, then reviewed `design.md` itself as the engineer, as `spec-review-gates.md` tells the reviewing agent. | passed |
| 3. Every agent and skill of a team names its owner | `owner` and `owners` are validated (a bare `jane` and an unknown key fail render naming the field); vendored skills fall back to the team owner (render and catalog tests in the receipt); the running app's Agents and Skills pages show `@mavogel` and *team owner* under inherited owners, as the artifact draws them. QualityLayer's own capture ran but its browser got no catalog from the `wails dev` backend, so the pictures are Playwright captures of the same running app,,,. | passed |
| 4. A non-engineer can send feedback on a skill or agent from where they work | `/feedback` in Kiro drafted a report naming the `customer` agent; the app's Feedback button on `product-owner` and on `analyze-session` opened prefilled issues naming the item,. | passed |
| 5. The feedback reaches the owner | Does the prefilled issue read well to you as its owner? Both entry points produce a `https://github.com/mavogel/kirokit/issues/new?…` link labelled `feedback,team:product-team,<kind>:<name>` whose body carries `Owner: @mavogel`, and nothing is sent (GitHub answered the anonymous link with a sign-in redirect that keeps the whole address). Open the link in signed in to GitHub and judge the form; do not submit unless you want the issue. | confirmed by you |
| 6. A session-analysis skill suggests harness improvements from a Kiro chat | `analyze-session`, run by `qa` on a session with two denied `engineer` calls, read it with file tools only, reported two recommendations under `@mavogel` naming `teams/product-team/agents/product-owner/agent.json.tmpl`, and printed one prefilled issue link; an unknown session id is reported as missing with the list of sessions. | passed |
| 7. The README explains the role setup and the feedback loop | The README's "Role agents and feedback" section and manifest reference cover `product-team`, `owner`/`owners`/`feedback.github`, `/feedback`, the app's Owner column and Feedback button, and `analyze-session`. | passed |
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
- Round 1 · everything checked passed
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: the desktop app opens the system browser at an address built from content of a cloned repository; the skill reads the person's local Kiro, the Plan says it is risky: the desktop app gains its first path that opens the system browser, with an address built from repository content; and the review gates and slice 2 and 4 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `go test ./...` · exit 0 · 0s · reused
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task lint` · exit 0 · 1s
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `go vet ./...` · exit 0 · 0.1s
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task build` · exit 0 · 0.6s
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task desktop:test` · exit 0 · 2.7s
- all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0s · reused
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->
### Security

**Result:** done · nothing severe in the change
**Boundaries:** manifest.yaml of a cloned repository to the desktop app's system-browser open (repository to app to OS); the analyze-session skill reading local Kiro session files (agent, read-only).
- Feedback address: `feedback.github` is validated at load (`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`), then the frontend `openURL` allows only a parsed `https://github.com/<owner>/<repo>/issues/new` with no userinfo. A `..` or `.` repo segment is normalised by `new URL` and refused. Title, body and labels are percent-encoded, including the characters Wails refuses. The worst reachable result is a prefilled GitHub issue form that the person submits themselves.
- analyze-session: the template only reads the person's own session files and sends nothing.
<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification

### Iteration 1

**Revision:** a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08

**Assignment:** 9afc1170-f5ed-43f3-a31b-4f34caee8632
**Judge:** fresh helper

- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `go test ./...` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task lint` · exit 0 · 1s
- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `go vet ./...` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task build` · exit 0 · 0.6s
- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `task desktop:test` · exit 0 · 2.7s
- [PASS] check: reused receipt — all · a67a2d303474d7e6490b7e208d938045ad5f453fd53e08cbd08317dc01b68c08 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0s · reused
- [PASS] done 1: `kirokit install product-team` into a temporary Kiro home gave `customer`, `engineer`, `product-owner`, `qa`, `ux-designer`; `kiro-cli chat --agent product-owner --v3` ran `Sub-agent: customer` → Completed and reported Sam's answer — `evidence/it1-scenario-1.txt`
- [PASS] done 2: with `product-team` installed, `kiro-cli chat --agent engineer --v3 --mode spec` wrote `requirements.md`, had `product-owner` review it as a subagent and folded its gaps in before `design.md`, then reviewed `design.md` itself as the engineer, as `spec-review-gates.md` tells the reviewing agent — `evidence/it1-scenario-2.txt`
- [PASS] done 3: `owner` and `owners` are validated (a bare `jane` and an unknown key fail render naming the field); vendored skills fall back to the team owner (render and catalog tests in the receipt); the running app's Agents and Skills pages show `@mavogel` and *team owner* under inherited owners, as the artifact draws them. QualityLayer's own capture ran but its browser got no catalog from the `wails dev` backend, so the pictures are Playwright captures of the same running app — `evidence/it1-scenario-5.txt`, `evidence/it1-scenario-5-step-2-agents-light.png`, `evidence/it1-scenario-5-step-4-skills-light.png`, `evidence/it1-artifact-agents-page.png`
- [PASS] done 4: `/feedback` in Kiro drafted a report naming the `customer` agent; the app's **Feedback** button on `product-owner` and on `analyze-session` opened prefilled issues naming the item — `evidence/it1-scenario-4.txt`, `evidence/it1-scenario-5.txt`
- [USER] done 5: Does the prefilled issue read well to you as its owner? Both entry points produce a `https://github.com/mavogel/kirokit/issues/new?…` link labelled `feedback,team:product-team,<kind>:<name>` whose body carries `**Owner:** @mavogel`, and nothing is sent (GitHub answered the anonymous link with a sign-in redirect that keeps the whole address). Open the link in `evidence/it1-scenario-4.txt` signed in to GitHub and judge the form; do not submit unless you want the issue.
- [PASS] done 6: `analyze-session`, run by `qa` on a session with two denied `engineer` calls, read it with file tools only, reported two recommendations under `@mavogel` naming `teams/product-team/agents/product-owner/agent.json.tmpl`, and printed one prefilled issue link; an unknown session id is reported as missing with the list of sessions — `evidence/it1-scenario-6.txt`
- [PASS] done 7: the README's "Role agents and feedback" section and manifest reference cover `product-team`, `owner`/`owners`/`feedback.github`, `/feedback`, the app's Owner column and Feedback button, and `analyze-session` — `evidence/it1-scenario-7.txt`
- [PASS] scenario 1: all 5 steps as expected (engineer call held for permission, denied in non-interactive mode); probe: an unknown team exits 1 and writes nothing — `evidence/it1-scenario-1.txt`
- [PASS] scenario 2: steps 1–2 as expected; step 3 does not apply because CLI v3 runs specs — `evidence/it1-scenario-2.txt`
- [PASS] scenario 3: all 3 steps as expected; probes: an unknown `owners` key and a `javascript:` feedback repository fail naming the field — `evidence/it1-scenario-3.txt`
- [PASS] scenario 4: steps 1–2 as expected; step 3 checked by decoding the link and GitHub's sign-in redirect (the signed-in form view is done 5's question); probe: `/feedback` with no argument asks which item — `evidence/it1-scenario-4.txt`
- [PASS] scenario 5: all 5 steps in the running `wails dev` app driven by headless Chrome, with a recording `open` shim so Wails' real `BrowserOpenURL` path ran without opening a tab; probe: the running app refuses `https://evil.example/…`, `javascript:` and `/../` addresses with no open — `evidence/it1-scenario-5.txt`, `evidence/it1-scenario-5-step-5-agents-dark-focus.png`
- [PASS] scenario 6: all 3 steps as expected; probe: an unknown session id — `evidence/it1-scenario-6.txt`
- [PASS] scenario 7: both steps as expected — `evidence/it1-scenario-7.txt`
- [PASS] dod T1: field-named validation of handles and `feedback.github`, `OwnerOf` fallback, a manifest without `owner` loads
- [PASS] dod T2: `owners` rows sorted with vendored skills from the team, `feedbackRepo` error text, unknown skill key fails, new `Context` fields stay out of `context.json`
- [PASS] dod T3: team agents and globs pass, unknown names in `availableAgents`, `trustedAgents` and a subagent rule's `match` fail naming the field, agents without subagent settings render
- [PASS] dod T4: among existing goldens only the two `context.json` files changed; `go test ./...` now runs `kirokit test` over every team
- [PASS] dod T5: each agent lists exactly the other four; the trusted set and allow rule leave out `engineer`; only `engineer` has `write` and `shell`, with platform-team's deny rules; invented-persona notice; no MCP servers or secrets
- [PASS] dod T6: the steering names each reviewer's checks and says to change the file before the next phase; every role agent lists it
- [PASS] dod T7: the consultation and the spec run are recorded in `05-implement.md` and re-run here
- [PASS] dod T8: the rendered owner table lists every agent and skill with *(team owner)*, empty owners render as *no owner*, and the prompt forbids submitting
- [PASS] dod T9: catalog owners and feedback per the contract; `product-team` reports `mavogel/kirokit`
- [PASS] dod T10: title, body and labels per the contract, special characters round-trip, refusal text `refusing to open <url>`
- [PASS] dod T11: escaped cells, refusal beside its row, keyboard reachable with a visible focus outline (captured in dark; the same `:focus-visible` rule and `--focus` token cover light)
- [PASS] dod T12: the skill names the fields it reads, reports missing types, reads files only and sends nothing
- [PASS] dod T13: every README name exists; `sync-agent-assets --check` receipt
- [PASS] truth oracle: `product-team` in a temporary Kiro home: `product-owner` consulted `customer` as a subagent, and `/feedback` and the app's **Feedback** button both produced `https://github.com/mavogel/kirokit/issues/new?…` addresses whose body names the agent and @-mentions `@mavogel` — `evidence/it1-oracle.txt`
- [PASS] truth 1: `platform-team` and `data-team` from before the change render byte-identically with the old and the new binary; goldens change only for teams that set the fields (plus the recorded pre-existing `data-team` pin drift) — `evidence/it1-truth-1.txt`
- [PASS] review: no swallowed errors or races found, no stack drift (no `go.mod` or `package.json` change), every changed file is in a Files block or a recorded deviation, the Security block holds and nothing was missed; tests: 5 tests added · docs: no doc impact (`README.md` was updated by the build)
- [NOTE] review: On narrow windows the new Owner and Feedback columns of the Agents table sit off-screen. At the app's 720px minimum the main area overflows by 294px (88px before the change), and at 1000px the first Feedback button ends at 1001px; it stays reachable by horizontal scroll and Tab — `evidence/it1-scenario-5.txt`, `evidence/it1-scenario-5-probe-agents-720.png`
- [NOTE] review: The review gates fire only when an agent writes a spec. Asked to draft requirements outside spec mode, the `qa` agent did not consult `product-owner`; the README already says a steering cannot force the review — `evidence/it1-scenario-2.txt`

Tests added (unstaged, each run once and passing — `evidence/it1-new-tests.txt`): `TestLoad_OwnerHandleBoundariesAndEmptyOverride` (`internal/manifest/manifest_test.go`; rejects a handle check with no 39-character limit, one that skips empty overrides, or a repo check that allows `.`), `TestRenderTeam_OwnerKeyForAVendoredSkillIsAccepted` (`internal/render/render_test.go`; rejects an owner-key check that ignores vendored skills), `TestCatalog_OwnerIsKeyedByTheAgentFolder` (`desktop/catalog_test.go`; rejects an owner lookup by the file's `name` field), `A Feedback press reports the pressed item…` (`desktop/frontend/src/pages/feedback.test.js`; rejects swapped data attributes or firing on a disabled button), `openURL outside the app window…` (`desktop/frontend/src/api.test.js`; rejects a TypeError instead of the readable message).

**Pictures:** `evidence/it1-scenario-5-step-2-agents-light.png`, `evidence/it1-scenario-5-step-4-skills-light.png`, `evidence/it1-scenario-5-step-5-agents-dark-focus.png`, `evidence/it1-scenario-5-probe-agents-720.png`, compared with `evidence/it1-artifact-agents-page.png`, `evidence/it1-artifact-skills-page.png`, `evidence/it1-artifact-opened-issue.png`

**Try it:**

1. Build it: `task build`
2. Install the example team into a scratch Kiro home: `export KIRO_HOME=$(mktemp -d) && ./kirokit install product-team`; look for `installed product-team into …`
3. Ask the product owner to consult the customer: `kiro-cli chat --agent product-owner --v3`, then "Ask the customer agent what frustrates them most about onboarding, then summarise its answer."; look for `Sub-agent: customer` without a permission prompt
4. In the same chat type `/feedback the customer agent ignored my persona notes`; look for a `https://github.com/mavogel/kirokit/issues/new?…` link whose body has `**Owner:** @mavogel`, open it and do not submit
5. Start the app with `task desktop:dev`, open **Agents**, and press **Feedback** on `product-owner`; your browser opens the prefilled GitHub issue form
6. Ask any role agent "Analyze my last session with the analyze-session skill"; look for recommendations grouped under `@mavogel` and one issue link

**Result:** pass
