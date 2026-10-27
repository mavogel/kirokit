# Support Quick, Kiro IDE and kirocrew; README refresh · what was checked

## Summary

<!-- ql:summary -->
6 of 7 points assessed · 5 passed

Independent review failed; the recorded findings need attention.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Failed:** Point 6 · A new showcase GIF and video show the current product · The new showcase GIF and video were never recorded. `assets/kirokit-showcase.gif` and `.mp4` are unchanged by this build. The video still shows the old `kirokit install platform-team` story, with no `render --quick`, Kiro IDE or kirocrew scene (T13 is open), while the README's alt text already describes the new recording,
- **Unverified:** The Quick round trip was not run, because it needs your Quick account (Done means 7 is yours to confirm, and T6 is open). Its inputs hold: scenario 2 shows the second-pass skill equals an export carrying the id, and the agent pins that id
- **Unverified:** The Quick docs describe unconfirmed Quick behaviour as fact. `docs/amazon-quick.md` step 5 says a reimport overwrites the agent and keeps the pinned skill, but the Plan lists both as unconfirmed assumptions until your round trip (T6) runs
- **Unverified:** The official app capture could not be taken. `qualitylayer evidence shot` for `team-quick-bundle/index.html#teams-quick` returned `unavailable: Unsafe, changed or oversized regular file`, even without clicks. The Teams pictures are this review's own Playwright captures of the running app
- **Need:** T6 · Run the two-pass Quick round trip on your own Quick account, because Quick has no API and only you can import · queued for Review
- **Need:** T13 · Record the new showcase once Kiro IDE renders again · queued for Review
- **Only you can confirm:** The import in your Quick account
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: reads files exported from Amazon Quick (, folders, ) that the person passes on the command line; it only parses them in memory and writes.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 6 checks
- **Independent review** · fresh helper
  - State: fail
  - Result: failed · 1 of 6 points
- **Second opinion** · Codex · GPT-6.1 Sol · another vendor
  - State: done
  - Result: done · 2 findings fixed
  - Why it runs here: Runs because the change crosses a trust boundary: reads files exported from Amazon Quick (, folders, ) that the person passes on the command line; it only parses them in memory and writes, the Plan says it is risky: amazon Quick is an outside system with no API: its import rules come from experiments on one account (research brief), and agent overwrite and slice 5 is marked risky.
- **Fix rounds** · the build session
  - State: waiting
  - Result: not yet
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
| Done means | Proof | |
|---|---|---|
| 1. A team's chosen skills and agents render into an importable Quick bundle | `render --quick product-team` writes `skills/user-story/SKILL.md` and `agents/product-owner.json`. Name, description and `system_prompt` come from the Kiro agent; welcome message and starter prompts come from `manifest.yaml`. The agent's ten keys match Quick's own export (`docs/research/quick/Daily_Intelligence_Briefing.json`) in name and order. In the running app, Export for Quick rendered the same bundle into `<data>/quick/product-team` and opened the folder once,,. | passed |
| 2. Quick skill ids survive a reimport | `quick ids` read a `SKILL.md`, a folder and a `.zip` export into a gitignored, mode-600 ids file. The next render wrote the id line back (the export equals the rendered file) and pinned the id in the agent's `skills`. | passed |
| 3. Quick output is checked like the Kiro output | `testdata/golden/product-team/quick/` is compared byte for byte by `go test ./...` and `./kirokit test` (every team `ok`), using the made-up ids in `testdata/golden/quick-ids.yaml`. `golden_test.go` covers drift and a stray `quick/` folder. | passed |
| 4. Kiro IDE and kirocrew use what `kirokit install` puts in `KIRO_HOME` | After `kirokit install product-team` into a fresh `KIRO_HOME`, `kiro-cli agent list`, Kiro IDE 1.4.26's agent picker (under GLOBAL) and kirocrew 0.10.0's `/api/agents/installed` each list customer, engineer, product-owner, qa and ux-designer. The README's "Works with" says what each tool picks up, including the IDE's `KIRO_HOME` rules,. | passed |
| 5. The README opens with a new, short intro that names every supported Kiro tool and Quick | The README opens with the logo, a pitch naming Kiro CLI, Kiro IDE, kirocrew and Amazon Quick, the GIF and install, all before any reference material. It has 141 lines (359 before). Every moved section has a `docs/` page, and `check-doc-links` resolves all 54 links. | passed |
| 6. A new showcase GIF and video show the current product | The new showcase GIF and video were never recorded. `assets/kirokit-showcase.gif` and `.mp4` are unchanged by this build. The video still shows the old `kirokit install platform-team` story, with no `render --quick`, Kiro IDE or kirocrew scene (T13 is open), while the README's alt text already describes the new recording,. | failed |
| 7. The Quick round trip works on a real account | Did the rendered user-story skill and product-owner agent import into your Quick account, and does a re-exported skill equal the rendered file apart from its `id:` line? Not run here, because it needs your Quick account. Run the two-pass flow in `docs/amazon-quick.md` and save the exports in the task's records folder under quick-roundtrip. | yours to confirm |
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
- Round 1 · 5 points still open
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
**Second opinion · Codex, GPT-6.1 Sol** · asked for 2 changes · 2 findings: 2 taken, 0 kept.
- `internal/quick/quick.go:93` — Distinct skill folders can share a frontmatter `name`, receive the same id, and produce indistinguishable agent pins. · Taken: `quick.RenderTeam` now fails before publishing with `skills "<a>" and "<b>" share the name "<n>"; Quick matches skills by name`; the case `two folders, one name` in `TestRenderTeam_Errors` failed first and passes now, and `docs/amazon-quick.md` names the error (9b6ff3b).
- `internal/quick/quick.go:158` — Copying every file as `0600` strips executable bits from vendored scripts that the Kiro renderer preserves. · Taken: `writeSkill` keeps the owner-execute bit (0700) as `writeExternalSkills` does, other files stay 0600; `TestRenderTeam_VendoredScriptStaysExecutable` (a vendored `scripts/run.sh` at 0755) failed first and passes now (9b6ff3b).
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `task ci` · exit 0 · 10.3s
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `go vet ./...` · exit 0 · 0.1s
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `go test ./...` · exit 0 · 8.6s
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `task desktop:test` · exit 0 · 2.6s
- all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `node --test scripts/showcase/` · exit 0 · 68.1s
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->

### Security

**Result:** done · nothing severe in the change
**Boundaries:** `kirokit quick ids` reading exported SKILL.md, folders and .zip (files to CLI, parsed in memory); `quick-ids.yaml` write (CLI to disk); desktop `QuickDir`/`OpenFolder` (frontend to Go, then a spawned file manager); `resolvePrompt` file:// reads from the rendered tree.
- Zip entries are never extracted (only read in memory, capped at 256 entries and 1 MiB each, so no zip-slip). Only the frontmatter `name` and a 32-hex `id` leave the export, and the id is validated before it is written, so no YAML or frontmatter injection. The ids file is written 0600 via a temp file and rename.
- Desktop: the frontend passes a team name only; separators, `..`, absolute paths and NUL are refused, and the folder opener is run with an argument vector (no shell) on an absolute path under the data directory.
- `resolvePrompt` refuses absolute, `~` and `..` escapes lexically; its input is the repository's own rendered templates, not an attacker-controlled export.

<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification

### Iteration 1

**Revision:** d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7

**Assignment:** c8323d80-e037-41c2-af19-9a561d13dedc
**Judge:** fresh helper

- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `task ci` · exit 0 · 11.1s
- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `go vet ./...` · exit 0 · 0.2s
- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `go test ./...` · exit 0 · 8.9s
- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `task desktop:test` · exit 0 · 2.5s
- [PASS] check: reused receipt — all · d7e9a9920aae1ae069e984064bc86fddafee11598934497310c7768e03e52ad7 · `node --test scripts/showcase/` · exit 0 · 0s · reused
- [PASS] done 1: `render --quick product-team` writes `skills/user-story/SKILL.md` and `agents/product-owner.json`. Name, description and `system_prompt` come from the Kiro agent; welcome message and starter prompts come from `manifest.yaml`. The agent's ten keys match Quick's own export (`docs/research/quick/Daily_Intelligence_Briefing.json`) in name and order. In the running app, Export for Quick rendered the same bundle into `<data>/quick/product-team` and opened the folder once — `evidence/it1-scenario-1.txt`, `evidence/it1-scenario-3-light-1180.txt`, `evidence/it1-scenario-3-step-3-light-1180.png`
- [PASS] done 2: `quick ids` read a `SKILL.md`, a folder and a `.zip` export into a gitignored, mode-600 ids file. The next render wrote the id line back (the export equals the rendered file) and pinned the id in the agent's `skills` — `evidence/it1-scenario-2.txt`
- [PASS] done 3: `testdata/golden/product-team/quick/` is compared byte for byte by `go test ./...` and `./kirokit test` (every team `ok`), using the made-up ids in `testdata/golden/quick-ids.yaml`. `golden_test.go` covers drift and a stray `quick/` folder — `evidence/it1-scenario-1.txt`
- [PASS] done 4: after `kirokit install product-team` into a fresh `KIRO_HOME`, `kiro-cli agent list`, Kiro IDE 1.4.26's agent picker (under GLOBAL) and kirocrew 0.10.0's `/api/agents/installed` each list customer, engineer, product-owner, qa and ux-designer. The README's "Works with" says what each tool picks up, including the IDE's `KIRO_HOME` rules — `evidence/it1-scenario-4.txt`, `evidence/it1-scenario-4-kiro-ide.png`
- [PASS] done 5: the README opens with the logo, a pitch naming Kiro CLI, Kiro IDE, kirocrew and Amazon Quick, the GIF and install, all before any reference material. It has 141 lines (359 before). Every moved section has a `docs/` page, and `check-doc-links` resolves all 54 links — `evidence/it1-scenario-5.txt`
- [FAIL] done 6: The new showcase GIF and video were never recorded. `assets/kirokit-showcase.gif` and `.mp4` are unchanged by this build. The video still shows the old `kirokit install platform-team` story, with no `render --quick`, Kiro IDE or kirocrew scene (T13 is open), while the README's alt text already describes the new recording — `evidence/it1-scenario-6.txt`, `evidence/it1-scenario-6-old-video-10s.png`
- [USER] done 7: Did the rendered user-story skill and product-owner agent import into your Quick account, and does a re-exported skill equal the rendered file apart from its `id:` line? Not run here, because it needs your Quick account. Run the two-pass flow in `docs/amazon-quick.md` and save the exports in the task's records folder under quick-roundtrip.
- [PASS] scenario 1: all 5 steps as expected. Past the script, rendering with no team skips the non-Quick teams with a note, and a missing `--out`, an unknown team or a directory given as the ids file each fail with a clear message — `evidence/it1-scenario-1.txt`
- [PASS] scenario 2: all 5 steps as expected. Past the script, a zip export, an export without an id, an empty folder, a missing path, a malformed id (the ids file stays unchanged) and the real Quick exports all behave as documented — `evidence/it1-scenario-2.txt`
- [PASS] scenario 3: all 3 steps in the real app. Setup: Wails dev on a throwaway HOME, with the clone at the judged commit. Only product-team has Export for Quick, and the other teams show the note. Pressing Enter runs `render --quick` into `<data>/quick/product-team` and disables every card button while it runs. The card then shows "Quick bundle ready: 1 skill, 1 agent.", the folder and Show folder, and the open shim logged exactly one open. A forced render failure shows the error beside the button and opens nothing. Light, dark and 760 px views match the artifact's Exported state — `evidence/it1-scenario-3-light-1180.txt`, `evidence/it1-scenario-3-failure.txt`, `evidence/it1-scenario-3-step-3-dark-1180.png`, `evidence/it1-scenario-3-artifact-done.png`
- [PASS] scenario 4: all 4 steps. The install printed `installed product-team into /tmp/kirokit-it1-ide/kiro`, and Kiro CLI, Kiro IDE and kirocrew each list the five agents — `evidence/it1-scenario-4.txt`, `evidence/it1-scenario-4-kiro-ide-text.txt`
- [PASS] scenario 5: `check-doc-links` printed `all links resolve (54 checked)`. Read top to bottom, the README puts the intro, GIF, install and first steps before "Works with", and its docs index links all seven `docs/` pages — `evidence/it1-scenario-5.txt`
- [FAIL] scenario 6: The showcase video still shows the old story. Step 1 holds (showcase tests receipt) and step 2 holds (`checks.sh length` passes at 36.4 s). Step 3 fails: the MP4 shows the old platform-team recording, with no Quick, Kiro IDE or kirocrew scene — `evidence/it1-scenario-6.txt`
- [NOTE] scenario 7: The Quick round trip was not run, because it needs your Quick account (Done means 7 is yours to confirm, and T6 is open). Its inputs hold: scenario 2 shows the second-pass skill equals an export carrying the id, and the agent pins that id.
- [PASS] dod T1: without a `quick` section other teams keep no `quick` key in `context.json`; the section is parsed in order; an unknown key (`welcome`) fails and is named; a duplicate agent or skill fails — `evidence/it1-review-probes.txt`
- [PASS] dod T2: skill bytes equal the Kiro render, with `id:` as the last frontmatter line; vendored skills are copied whole; the agent has its 10 keys in order, two-space indent and a trailing newline; a `file://` prompt is resolved; every listed error is tested; `LoadIDs` treats a missing file as empty and a malformed one as an error
- [PASS] dod T3: a named team without content fails with `team data-team lists nothing for Quick`; an all-teams run skips it with a note; a missing id warns on stderr and exits 0; flag over env over repo default is tested; plain render goldens are unchanged
- [PASS] dod T4: `Update` writes `quick/` with the fixture ids; `Compare` reports changed, missing, extra and stray Quick files; the product-team Quick golden carries and pins the fixture id
- [PASS] dod T5: a file, a folder and a zip are read; a merge keeps other names; `skipped <path>: no id` is printed; two ids or a malformed id fails and writes nothing; zip limits are tested; the header comment, mode 600 and `git check-ignore` are shown — `evidence/it1-scenario-2.txt`
- [MISSING] dod T6: Your two-pass Quick round trip has not been run. No quick-roundtrip exports exist in the records folder, so the card's verify `diff` has nothing to compare.
- [PASS] dod T7: in the running app `Team.Quick` holds only for product-team; `QuickDir` stays under `<data>/quick` and refuses bad names; `QuickBundle` counts the files and errors on a missing folder; `openCommandFor` is set per OS; `OpenFolder` refuses a missing folder
- [PASS] dod T8: `quickArgs` has the planned shape; only Quick teams get the button, and the others get the note; counts read in singular and plural with the path and Show folder; the error sits beside the button; only the started run's exit counts, and the folder opens once
- [PASS] dod T9: `grep -c product-owner records/check-kirocrew-agents.json` prints 1, and a re-run on a fresh home confirms the IDE and kirocrew lists — `evidence/it1-scenario-4.txt`
- [PASS] dod T10: seven `docs/` pages hold every moved section; `docs/amazon-quick.md` documents the flags, the env var, `quick ids`, the ids file, the manifest section, the import steps and the "After" Mermaid; the goldens that quote a section were regenerated (`kirokit test` ok) — `evidence/it1-scenario-5.txt`
- [PASS] dod T11: the README is 141 of the old 359 lines, with intro, video and install first; "Works with" covers all four tools, including the IDE's `KIRO_HOME` rules
- [PASS] dod T12: `checks.sh` allows 30 to 90 s and 10 MB GIF / 6 MB MP4; `record.sh` runs the privacy check on `ide-text.txt` and `crew-text.txt`; `ide-drive.mjs` ran here and listed the five agents — `evidence/it1-review-probes.txt`
- [MISSING] dod T13: The new showcase was not recorded or published. The MP4 is still the 36.4 s platform-team recording. Kiro IDE renders on this Mac again (`ide-drive.mjs` succeeded here), so `task showcase` can be retried — `evidence/it1-scenario-6.txt`
- [MISSING] truth oracle: environment: only the Kiro half of the oracle could be shown. The five product-team agents appear in Kiro CLI, in Kiro IDE's agent picker and in kirocrew's list after a fresh install (`evidence/it1-scenario-4.txt`). The Quick two-pass round trip needs your Quick account and was not run (T6).
- [PASS] review: no swallowed errors, null or off-by-one bugs found; no stack drift (`go.mod`, `package.json` unchanged); every changed file sits in a card's Files or the amended Plan; the Security block covers the export reader, the ids file and `OpenFolder`, and nothing was missed; the existing tests reject plausible wrong implementations (exact bytes, key order, id placement); tests: 4 tests added · docs: no doc impact
- [NOTE] review: The Quick docs describe unconfirmed Quick behaviour as fact. `docs/amazon-quick.md` step 5 says a reimport overwrites the agent and keeps the pinned skill, but the Plan lists both as unconfirmed assumptions until your round trip (T6) runs.
- [NOTE] review: The official app capture could not be taken. `qualitylayer evidence shot` for `team-quick-bundle/index.html#teams-quick` returned `unavailable: Unsafe, changed or oversized regular file`, even without clicks. The Teams pictures are this review's own Playwright captures of the running app.

**Pictures:** `evidence/it1-scenario-3-step-1-light-1180.png`, `evidence/it1-scenario-3-step-2-light-1180.png`, `evidence/it1-scenario-3-step-3-light-1180.png`, `evidence/it1-scenario-3-step-3-dark-1180.png`, `evidence/it1-scenario-3-step-3-light-760.png`, `evidence/it1-scenario-3-failure.png`, `evidence/it1-scenario-3-artifact-done.png` (expected state), `evidence/it1-scenario-4-kiro-ide.png`, `evidence/it1-scenario-6-old-video-10s.png`

**Try it:**

1. Build it: `task build`
2. Render the bundle: `./kirokit render --quick product-team --out tmp/q`; look for `rendered Quick bundle for product-team: 1 skills, 1 agents` and one `user-story` warning
3. Record an id: add `id: 0123456789abcdef0123456789abcdef` as the last frontmatter line of a copy of `tmp/q/skills/user-story/SKILL.md`, then run `./kirokit quick ids --quick-ids tmp/ids.yaml <that copy>` and look for `recorded 1 ids`
4. Render again: `./kirokit render --quick product-team --quick-ids tmp/ids.yaml --out tmp/q2`; the skill now ends its frontmatter with the id and `agents/product-owner.json` lists it in `skills`
5. In the app: `task desktop:dev`, open Teams, press **Export for Quick** on product-team; look for "Quick bundle ready: 1 skill, 1 agent." and the folder opening
6. One install, three tools: `env KIRO_HOME=/tmp/k ./kirokit install product-team`, then `env KIRO_HOME=/tmp/k kiro-cli agent list` shows the five agents (Kiro IDE and kirocrew read the same home)
7. Still open: record the showcase with `task showcase` (T13) and run the Quick round trip on your account (T6)

**Result:** fail

### Iteration 2

**Assignment:** 03bd675b-8956-4f89-ac60-9b1a02f97e33
**Judge:** fresh helper

Scope: this round changed `internal/quick/quick.go` (shared-name check, vendored execute bit), its tests, the iteration-1 reviewer's `internal/quick/review_test.go` and `docs/amazon-quick.md`. Everything that runs `quick.RenderTeam` was rerun; the install path (it does not import `internal/quick`), `desktop/`, `scripts/showcase/`, `teams/` and `testdata/` are unchanged, so their earlier passes are carried with that reason. HEAD did not move from the judged commit.

- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `task ci` · exit 0 · 10.3s
- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `go vet ./...` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `go test ./...` · exit 0 · 8.6s
- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `task desktop:test` · exit 0 · 2.6s
- [PASS] check: reused receipt — all · f7f1d05ddeabdfd51c58bf18b2b83657ca8bb1a5d833926b0d261a71ab2f83e7 · `node --test scripts/showcase/` · exit 0 · 68.1s
- [PASS] done 1: rerun on the judged tree, `render --quick product-team` writes `skills/user-story/SKILL.md` and `agents/product-owner.json`, and the skill equals the committed Quick golden apart from its id line. The app half is carried from iteration 1: `desktop/` is unchanged and the bundle it renders for product-team is unchanged (no vendored script, no shared name) — `evidence/it2-scenario-1.txt`, `evidence/it1-scenario-3-light-1180.txt`, `evidence/it1-scenario-3-step-3-light-1180.png`
- [PASS] done 2: `quick ids` recorded the id into a mode-600, gitignored file; the second render's skill equals the export byte for byte and the agent's `skills` holds the id — `evidence/it2-scenario-2.txt`
- [PASS] done 3: `./kirokit test` prints `ok` for every team, and a fresh bundle equals `testdata/golden/product-team/quick/` with the id line removed — `evidence/it2-scenario-1.txt`
- [PASS] done 4: carried from iteration 1 for Kiro IDE and kirocrew; the install path does not import `internal/quick` and nothing it installs changed. Kiro CLI was rerun on a fresh `KIRO_HOME` and lists the five product-team agents — `evidence/it2-oracle.txt`, `evidence/it1-scenario-4.txt`, `evidence/it1-scenario-4-kiro-ide.png`
- [PASS] done 5: `check-doc-links` prints `all links resolve (54 checked)` after this round's docs edit; the README is unchanged at 141 lines, intro, GIF and install first, docs index linking all seven pages — `evidence/it2-scenario-5.txt`
- [FAIL] done 6: The new showcase GIF and video are still not recorded. `assets/kirokit-showcase.gif` and `.mp4` are unchanged by this build (last changed by the earlier showcase PR #19), the MP4 is the 36.4 s platform-team recording, and the README's alt text describes a recording that does not exist yet (T13 open) — `evidence/it2-scenario-6.txt`, `evidence/it1-scenario-6-old-video-10s.png`
- [USER] done 7: Did the rendered user-story skill and product-owner agent import into your Quick account, and does a re-exported skill equal the rendered file apart from its `id:` line? Not run here, because it needs your Quick account, and no `records/quick-roundtrip/` exports exist yet. Run the two-pass flow in `docs/amazon-quick.md` and save the exports there.
- [PASS] scenario 1: all 5 steps as expected on the judged tree. Past the script, `render --quick` with no team and with an empty team name both skip the non-Quick teams with a note and render product-team — `evidence/it2-scenario-1.txt`
- [PASS] scenario 2: all 5 steps as expected. Past the script, an export without an id is skipped, a malformed id fails and leaves the ids file byte-identical, and no path at all fails with a usage error — `evidence/it2-scenario-2.txt`
- [PASS] scenario 3: carried from iteration 1; this round changes no file under `desktop/`, and the bundle the button renders for product-team is unchanged, while the new shared-name error would reach the card through the non-zero-exit path iteration 1 already exercised with a forced failure — `evidence/it1-scenario-3-light-1180.txt`, `evidence/it1-scenario-3-failure.txt`
- [PASS] scenario 4: Kiro CLI step rerun: the install prints `installed product-team into /tmp/kirokit-it2-oracle/kiro` and `kiro-cli agent list` lists the five agents; the Kiro IDE and kirocrew steps are carried from iteration 1 because the install path is untouched — `evidence/it2-oracle.txt`, `evidence/it1-scenario-4.txt`
- [PASS] scenario 5: `check-doc-links` printed `all links resolve (54 checked)`; the README reads intro, GIF, install and first steps before "Works with" — `evidence/it2-scenario-5.txt`
- [FAIL] scenario 6: The showcase video still shows the old story. Steps 1 and 2 hold (showcase receipt; `checks.sh length` exits 0 at 36.4 s); step 3 fails because the MP4 has no Quick, Kiro IDE or kirocrew scene — `evidence/it2-scenario-6.txt`
- [NOTE] scenario 7: The Quick round trip was not run, because it needs your Quick account (done 7, T6). Its inputs still hold: scenario 2 shows the second-pass skill equals an export carrying the id, and the agent pins that id.
- [PASS] dod T1: carried from iteration 1; `internal/manifest` is unchanged in this round
- [PASS] dod T2: the fix, rechecked through the built CLI on a throwaway repository: a vendored `scripts/run.sh` (0755 in the team) lands as `-rwx------` beside `-rw-------` files, as in the Kiro render; two listed folders sharing the frontmatter name `user-story` fail with `skills "user-story" and "copy" share the name "user-story"`, publish nothing, leave no temp folder and keep an existing bundle untouched. Undoing either half of the fix in an isolated copy makes its new test fail — `evidence/it2-review-probes.txt`, `evidence/it2-mutations.txt`
- [PASS] dod T3: a missing id warns on stderr and exits 0, a named team without Quick content fails with `team data-team lists nothing for Quick`, an all-teams run skips it with a note — `evidence/it2-scenario-1.txt`
- [PASS] dod T4: `./kirokit test` is `ok` for every team; no golden changed in this round — `evidence/it2-scenario-1.txt`
- [PASS] dod T5: a file and a folder export are read, a no-id export is skipped, a malformed id fails without writing, the header comment, mode 600 and `git check-ignore` are shown — `evidence/it2-scenario-2.txt`
- [MISSING] dod T6: Your two-pass Quick round trip has still not been run. No `records/quick-roundtrip/` exports exist, so the card's verify `diff` has nothing to compare.
- [PASS] dod T7: carried from iteration 1; no file under `desktop/` changed in this round
- [PASS] dod T8: carried from iteration 1; no file under `desktop/frontend/` changed in this round
- [PASS] dod T9: carried from iteration 1; the install output is unchanged and `records/check-kirocrew-agents.json` still lists product-owner
- [PASS] dod T10: `docs/amazon-quick.md` now names the shared-name error and the kept execute bit, and every link still resolves — `evidence/it2-scenario-5.txt`
- [PASS] dod T11: `README.md` is unchanged in this round (141 lines) and its links resolve — `evidence/it2-scenario-5.txt`
- [PASS] dod T12: carried from iteration 1; `scripts/showcase/` is unchanged in this round and its receipt exits 0
- [MISSING] dod T13: The new showcase was still not recorded or published. The MP4 is the 36.4 s recording from before this build — `evidence/it2-scenario-6.txt`
- [MISSING] truth oracle: environment: only the Kiro half of the oracle can be shown. A fresh install lists the five agents in Kiro CLI (rerun), Kiro IDE and kirocrew (carried, install path untouched) — `evidence/it2-oracle.txt`, `evidence/it1-scenario-4.txt`. The Quick two-pass round trip needs your Quick account and was not run (T6).
- [PASS] review: the fix checks the shared name before publishing (the deferred cleanup removes the temp bundle), returns every new error (`d.Info`, `WriteFile`, `Chmod`), keeps the Kiro renderer's 0700/0600 rule, and on Windows (no execute bits) stays 0600. Files changed sit inside the recorded deviation; no stack drift. The new tests reject plausible wrong code: dropping the name check or writing every file 0600 again fails them, while copying the source mode or dropping `Chmod` are equivalent here (the Kiro tree is already 0700, the target is a fresh file). Iteration 1's note about the docs stating unconfirmed Quick behaviour is resolved: step 5 now says only the skill's id survival was seen in Quick, as `docs/research/quick/QUICK-TARGET-BRIEF.md` records. Tests: 0 tests added (both new behaviours already have tests that fail for the wrong implementations) · docs: updated docs/amazon-quick.md — `evidence/it2-mutations.txt`

**Pictures:** none new — no visual behaviour changed this round; the carried app and Kiro IDE verdicts cite iteration 1's pictures

**Try it:**

1. Build it: `task build`
2. Render the bundle: `env KIROKIT_QUICK_IDS=/nonexistent ./kirokit render --quick product-team --out tmp/q`; look for `rendered Quick bundle for product-team: 1 skills, 1 agents` and one `user-story` warning
3. Record an id: copy `tmp/q/skills/user-story/SKILL.md` to `tmp/export/SKILL.md`, add `id: 0123456789abcdef0123456789abcdef` as its last frontmatter line, run `./kirokit quick ids --quick-ids tmp/ids.yaml tmp/export`; look for `recorded 1 ids`
4. Render again: `./kirokit render --quick product-team --quick-ids tmp/ids.yaml --out tmp/q2`; `diff tmp/export/SKILL.md tmp/q2/skills/user-story/SKILL.md` prints nothing and `agents/product-owner.json` lists the id
5. See the fix: in a team that lists two skill folders whose `SKILL.md` share one `name`, `render --quick` fails with `share the name` and writes nothing; a vendored executable script lands as `-rwx------`
6. One install, three tools: `env KIRO_HOME=/tmp/k ./kirokit install product-team`, then `env KIRO_HOME=/tmp/k kiro-cli agent list` shows the five agents
7. Still open: record the showcase with `task showcase` (T13) and run the Quick round trip on your account (T6)

**Result:** fail
