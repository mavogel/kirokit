# Support Quick, Kiro IDE and kirocrew; README refresh

<!-- ql:links -->
Task: `docs/plans/2026-10-27-support-amazon-quick-see-docs-research-q/`
<!-- /ql:links -->

## Why the change

Teams could only ship their agents and skills to Kiro CLI; now the same reviewed templates also render an Amazon Quick import bundle (CLI and desktop app), Kiro IDE and kirocrew are confirmed to pick up the same install, and the README is a short intro with its reference material in `docs/`.

## What a reviewer must know

- **The README showcase is not re-recorded yet.** `assets/kirokit-showcase.gif` and `.mp4` are still the old platform-team recording, while the README's alt text describes the new one. Kiro IDE showed only a blank window on this Mac during the build; run `task showcase` once it renders (the recorder now bounds every wait, so it cannot hang).
- **The Quick round trip is unconfirmed on a real account.** Formats come from the research exports and goldens are built from them, but whether Quick overwrites a reimported agent in place and keeps its pinned skill has not been tried; `docs/amazon-quick.md` says so.
- **`quick-ids.yaml` must never be committed** (gitignored at the repo root): ids belong to one Quick account. For the same reason `docs/research/quick/`, whose exports hold real ids, is left uncommitted.

## What changed

A new render mode sits beside the Kiro render and reuses it: the Kiro tree is rendered into scratch, then the listed skills and agents are rebuilt in Quick's format.

```text
kirokit render --quick <team> --out <dir>      kirokit quick ids <export>...
  └ quick.RenderTeam                              └ quick.ReadExports (.md, folders, .zip in memory)
      ├ manifest.Load → Quick (strict decode)     └ quick.WriteIDs → quick-ids.yaml (0600, sorted)
      ├ render.RenderTeamWith → scratch Kiro tree
      ├ quick.LoadIDs(quick-ids.yaml | --quick-ids | KIROKIT_QUICK_IDS)
      ├ writeSkill: SKILL.md + folder, id: line last in frontmatter
      ├ writeAgent: name/description/prompt from the Kiro agent, Quick fields from the manifest
      └ render.Publish → <dir> (backup when it would change)
```

The manifest gains one optional, strictly decoded section; teams without it are unchanged:

```yaml
quick:
  skills: [user-story]
  agents:
    - name: product-owner
      welcomeMessage: "..."
      starterPrompts: ["..."]
      skills: [user-story]     # pinned in Quick; must be in quick.skills
```

The bundle and the golden harness:

```diff
 <out>/skills/user-story/SKILL.md        the Kiro-rendered skill + "id: <hex>" when known
 <out>/agents/product-owner.json         name, description, system_prompt, welcome_message, starter_prompts,
                                         capabilities, connectors, skills (ids), spaces, files
 testdata/golden/<team>/
   rendered/  context.json
+  quick/                                compared byte for byte, rendered with testdata/golden/quick-ids.yaml
```

The desktop app gets one button per Quick team, and three Go calls that take only a team name:

```diff
 Teams card
   Install
+  Export for Quick → Run(['render','--quick',team,'--out',QuickDir(team)])
+                      → on its own exit 0: QuickBundle(team) counts, OpenFolder(team) once
```

The README is now intro, video, why, install, first steps, "Works with" and a docs index (141 lines, was 359); seven `docs/` pages hold the moved reference text, and `scripts/check-doc-links.mjs` checks every link, anchor and quoted section name. The showcase recorder gains Kiro IDE and kirocrew scenes (`ide-drive.mjs`, `crew-drive.mjs`) and a Quick render in the CLI scene.

## Screens compared

<!-- ql:screens -->
- **team-quick-bundle/index.html#teams-quick** · 1180 × 900 px · light · unavailable: Unsafe, changed or oversized regular file · stale
  - Independent review: No comparison recorded.
  - Refresh: recorded build changed
- **team-quick-bundle/index.html#teams-quick** · 1280 × 900 px · light · unavailable: Unsafe, changed or oversized regular file · stale
  - Independent review: No comparison recorded.
  - Refresh: recorded build changed
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
5 slices · 13 tasks · 13 commits · 61 files · +3715 −457

**Slice 1 · A team renders its Amazon Quick bundle from the same templates** · T1, T2, T3, T4 · 22 files · +1834 −24

- `cmd/kirokit/commands/quick_test.go` · T3, T5 · new · +220 −0
- `cmd/kirokit/commands/render.go` · T3 · +55 −2
- `internal/golden/golden.go` · T4 · +86 −18
- `internal/golden/golden_test.go` · T4 · +78 −0
- `internal/manifest/manifest.go` · T1 · +82 −0
- `internal/manifest/manifest_test.go` · T1 · +77 −0
- `internal/quick/ids.go` · T2, T5 · new · +237 −0
- `internal/quick/ids_test.go` · T2, T5 · new · +164 −0
- `internal/quick/quick.go` · T2 · new · +322 −0
- `internal/quick/quick_test.go` · T2 · new · +304 −0
- `internal/render/render.go` · T2 · +5 −0
- `teams/product-team/agents/product-owner/agent.json.tmpl` · T4 · +3 −2
- `teams/product-team/manifest.yaml` · T4 · +14 −0
- `teams/product-team/skills/user-story/SKILL.md.tmpl` · T4 · new · +47 −0
- `testdata/golden/product-team/context.json` · T4 · +19 −1
- `testdata/golden/product-team/quick/agents/product-owner.json` · T4 · new · +17 −0
- `testdata/golden/product-team/quick/skills/user-story/SKILL.md` · T4 · new · +48 −0
- `testdata/golden/product-team/rendered/agents/product-owner.json` · T4 · +2 −1
- `testdata/golden/product-team/rendered/prompts/feedback.md` · T4 · +1 −0
- `testdata/golden/product-team/rendered/skills/analyze-session/SKILL.md` · T4 · +1 −0
- `testdata/golden/product-team/rendered/skills/user-story/SKILL.md` · T4 · new · +47 −0
- `testdata/golden/quick-ids.yaml` · T4 · new · +5 −0

**Slice 2 · Quick ids come back from Quick's own exports** · T5, T6 · 3 files · +62 −0

- `.gitignore` · T5 · +2 −0
- `cmd/kirokit/commands/quick.go` · T5 · new · +59 −0
- `cmd/kirokit/commands/root.go` · T5 · +1 −0

**Slice 3 · The desktop app exports a team for Quick** · T7, T8 · 10 files · +387 −11

- `desktop/app.go` · T7 · +3 −0
- `desktop/catalog.go` · T7 · +4 −0
- `desktop/frontend/src/api.js` · T8 · +7 −0
- `desktop/frontend/src/main.js` · T8 · +50 −3
- `desktop/frontend/src/pages/pages.test.js` · T8 · +55 −1
- `desktop/frontend/src/pages/teams.js` · T8 · +51 −7
- `desktop/frontend/src/style.css` · T8 · +3 −0
- `desktop/platform.go` · T7 · +12 −0
- `desktop/quick.go` · T7 · new · +84 −0
- `desktop/quick_test.go` · T7 · new · +118 −0

**Slice 4 · Kiro IDE and kirocrew are confirmed, and the README starts with what kirokit is** · T9, T10, T11 · 18 files · +619 −329

- `AGENTS.md` · T10, T12 · +5 −4
- `CONTRIBUTING.md` · T10, T12 · +8 −7
- `README.md` · T11 · +92 −310
- `docs/agents-and-skills.md` · T10 · new · +42 −0
- `docs/amazon-quick.md` · T10 · new · +82 −0
- `docs/configuration.md` · T10 · new · +101 −0
- `docs/desktop-app.md` · T10 · new · +27 −0
- `docs/install-and-sync.md` · T10 · new · +77 −0
- `docs/secrets.md` · T10 · new · +39 −0
- `docs/testing.md` · T10 · new · +15 −0
- `global/steerings/security-baseline.md` · T10 · +1 −1
- `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl` · T10 · +1 −1
- `scripts/check-doc-links.mjs` · T10 · new · +122 −0
- `teams/data-team/steerings/data-conventions.md.tmpl` · T10 · +1 −1
- `teams/platform-team/manifest.yaml` · T10 · +3 −2
- `testdata/golden/data-team/rendered/steering/data-conventions.md` · T10 · +1 −1
- `testdata/golden/platform-team/context.json` · T10 · +1 −1
- `testdata/golden/platform-team/rendered/kirokit-steerings/platform-team/security-baseline.md` · T10 · +1 −1

**Slice 5 · The showcase shows Quick, Kiro IDE and kirocrew** · T12, T13 · 7 files · +697 −93

- `scripts/showcase/caption.mjs` · T12 · +48 −16
- `scripts/showcase/checks.sh` · T12 · +6 −6
- `scripts/showcase/cli-drive.sh` · T12 · +20 −6
- `scripts/showcase/crew-drive.mjs` · T12 · new · +174 −0
- `scripts/showcase/ide-drive.mjs` · T12 · new · +191 −0
- `scripts/showcase/record.sh` · T12 · +68 −27
- `scripts/showcase/showcase.test.mjs` · T12 · +190 −38

**Decided by the agent** · 1 file · +116 −0

- `internal/quick/review_test.go` · new · +116 −0
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. A team's chosen skills and agents render into an importable Quick bundle.** Scenario 1: Render the Quick bundle from the CLI, after slice 1 · Scenario 3: Export for Quick in the desktop app, after slice 3 · passed · `evidence/it2-scenario-1.txt`
- **2. Quick skill ids survive a reimport.** Scenario 2: Ids from a Quick export come back on the next render, after slice 2 · passed · `evidence/it2-scenario-2.txt`
- **3. Quick output is checked like the Kiro output.** Scenario 1: Render the Quick bundle from the CLI, after slice 1 · passed · `evidence/it2-scenario-1.txt`
- **4. Kiro IDE and kirocrew use what `kirokit install` puts in `KIRO_HOME`.** Scenario 4: One install, three Kiro tools, after slice 4 · passed · `evidence/it2-oracle.txt`
- **5. The README opens with a new, short intro that names every supported Kiro tool and Quick.** Scenario 5: The README is short and every link resolves, after slice 4 · passed · `evidence/it2-scenario-5.txt`
- **6. A new showcase GIF and video show the current product.** Scenario 6: The new showcase, after slice 5 · failed · `evidence/it2-scenario-6.txt`
- **7. The Quick round trip works on a real account.** Only you can confirm: the import in your Quick account · confirmed by you: Confirmed in the App
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **Installing into Quick**, connectors, spaces, routines, sharing or publishing; Quick has no API for them.
- **Rewriting prompts for Quick.** An agent's Quick prompt is its Kiro prompt; a prompt that mentions Kiro-only abilities (subagents, MCP tools) keeps that text.
- **Steering in Quick agents.** Global and team steerings stay Kiro-only.
- **Validating `KIRO_HOME`** against Kiro IDE's rules (it ignores a relative path, a root folder or one containing the home folder); the docs mention it.
- **Removing kirocrew's own `kirocrew-skill-view-*.json` files** from `agents/`; they are kirocrew's, and kirokit's backup copies them like any other file there.
- **Kiro web.**
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
- T13 · Until task showcase records again, assets/kirokit-showcase.gif and .mp4 are still the old platform-team recording, while the README's alt text already describes the new story (product-team, render --quick, Kiro IDE, kirocrew, app).
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
Taken as it stood with 9 items still open; lanes that ran: Security review, Project checks, Code review, Second opinion and Fix rounds.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The Quick bundle renders from the CLI
- Do: build and render product-team's bundle.
- Run: `task build && ./kirokit render --quick product-team --out tmp/quick-try`
- Expect: `rendered Quick bundle for product-team: 1 skills, 1 agents`, one warning for `user-story`, and `skills/user-story/SKILL.md` plus `agents/product-owner.json` in `tmp/quick-try`.
- Saw: the same line and files (independent review, round 2) — `evidence/it2-scenario-1.txt`

#### 2. Export for Quick in the desktop app
- Do: start `task desktop:dev` on a clone of this branch and press Export for Quick on product-team.
- Expect: the kirokit tab runs `kirokit render --quick product-team --out …/quick/product-team`, then the card shows "Quick bundle ready: 1 skill, 1 agent", the folder and Show folder, and the folder opens once; other teams show "No Quick export".
- Saw: that result, light and dark, and the failure path — `evidence/it1-scenario-3-step-3-light-1180.png`, `evidence/it1-scenario-3-failure.png`

#### 3. The import in your Quick account
- Do: import `tmp/quick-try/skills/user-story` (From folder) and `agents/product-owner.json` (From file), export the skill, run `./kirokit quick ids <export>`, render again, reimport both.
- Expect: one skill and one agent, no duplicates; the re-exported skill equals the rendered file with the same id, and the agent still lists that id in `skills`.
- Saw: not observed; only you can do this (T6).

### Not verified

- The Quick two-pass round trip on a real account (T6, Done means 7): it needs your Quick account.
- The new showcase GIF and video (T13, Done means 6): Kiro IDE showed only a blank window on this Mac during the build, so `task showcase` was stopped; the recorder's own tests pass (39), and the Kiro IDE scene ran on its own once (`records/check-kiro-ide-agents.txt`).

### Docs

- `README.md` rewritten as a short intro; new `docs/configuration.md`, `docs/agents-and-skills.md`, `docs/secrets.md`, `docs/install-and-sync.md`, `docs/testing.md`, `docs/desktop-app.md`, `docs/amazon-quick.md`; `CONTRIBUTING.md` (repository layout, showcase requirements and limits, docs rule) and `AGENTS.md` updated; steering and manifest comments that quoted README sections now point at `docs/`.

## Security

- No Medium or Low findings. The security review read the ids reader (exports parsed in memory, zip entries never extracted, 256 entries and 1 MiB caps), the ids file write (0600, temp file and rename), the desktop folder opener (team name only, argument vector, no shell) and `file://` prompt resolution (stays inside the rendered tree).

## Accepted with open items

<!-- ql:accepted -->
The author accepted this build as it stood, without another judge run. What was left open:

- T6: Run the two-pass Quick round trip on your own Quick account, because Quick has no API and only you can import. Steps: import tmp/quick-pass1 (skill From folder, agent From file); export the skill into the task's records/quick-roundtrip folder; run kirokit quick ids on that folder and render again into tmp/quick-pass2; reimport both and confirm the overwrite; export the skill and agent again into records/quick-roundtrip as user-story-reexport.md and product-owner-reexport.json.
- T13: Record the new showcase once Kiro IDE renders again. Kiro IDE 1.4.26 shows only a blank window on this Mac, also on your own profile after a restart, so its scene cannot be recorded now. When Kiro shows its normal window, run task showcase from the repository root; it publishes assets/kirokit-showcase.gif and .mp4 only after its privacy, size and length checks pass.
- T6 was not built: Your two-pass round trip on your Quick account
- T13 was not built: Record and publish the new showcase
- [FAIL] done 6: The new showcase GIF and video are still not recorded. `assets/kirokit-showcase.gif` and `.mp4` are unchanged by this build (last changed by the earlier showcase PR #19), the MP4 is the 36.4 s platform-team recording, and the README's alt text describes a recording that does not exist yet (T13 open) — `evidence/it2-scenario-6.txt`, `evidence/it1-scenario-6-old-video-10s.png`
- [FAIL] scenario 6: The showcase video still shows the old story. Steps 1 and 2 hold (showcase receipt; `checks.sh length` exits 0 at 36.4 s); step 3 fails because the MP4 has no Quick, Kiro IDE or kirocrew scene — `evidence/it2-scenario-6.txt`
- [MISSING] dod T6: Your two-pass Quick round trip has still not been run. No `records/quick-roundtrip/` exports exist, so the card's verify `diff` has nothing to compare.
- [MISSING] dod T13: The new showcase was still not recorded or published. The MP4 is the 36.4 s recording from before this build — `evidence/it2-scenario-6.txt`
- [MISSING] truth oracle: environment: only the Kiro half of the oracle can be shown. A fresh install lists the five agents in Kiro CLI (rerun), Kiro IDE and kirocrew (carried, install path untouched) — `evidence/it2-oracle.txt`, `evidence/it1-scenario-4.txt`. The Quick two-pass round trip needs your Quick account and was not run (T6).
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- `--quick` is a flag on `render`, not `--target quick` — the research brief records it as your choice and nothing else needs a target list. (Plan)
- The example is a new `user-story` skill and `product-owner` agent in product-team — `analyze-session` reads Kiro session files and means nothing in Quick. (Plan)
- Quick agent JSON keeps Quick's key order and two-space indent and ends with a newline — matches the exports, and goldens need one fixed form. (Plan)
- Missing ids warn and never fail; there is no `--strict-ids` — the brief made it optional and nothing needs it yet. (Plan)
- The `quick` manifest section is decoded strictly (unknown keys fail) while the rest of the manifest stays lenient — a typo there would otherwise silently drop content from the bundle. (Plan)
- No committed `quick-ids.example.yaml` — `kirokit quick ids` writes the file, and `docs/amazon-quick.md` shows its shape. (Plan)
- A skill listed for Quick must have `name` and `description` frontmatter, or render fails — Quick needs both. (Plan)
- Text in a team skill that Quick should see as `{{input}}` is written `{{"{{"}}input}}` in the template, or the skill is vendored — the vendored path already copies text unchanged. (Plan)
- Desktop exports go to `<app data>/quick/<team>`, not into the config clone — the bundle is an output, and the clone stays clean for sync. (Plan)
- The ids file sits at the config repository root and is gitignored there — kirokit already resolves everything from that root, and one person's checkout is one Quick account. (Plan)
- Its verify command lost `./desktop/...` (the desktop module is tested from `desktop/` in T7), so QualityLayer did not rerun it; `go test ./internal/golden ./cmd/kirokit/commands` passed when run by hand (96 tests) and `go test ./...` passed after it; nothing you approved changes. (slice 1)
- `kirokit render --quick` prints the contract's `N skills, M agents` without singular forms ("1 skills, 1 agents"); the desktop card uses singular and plural; nothing you approved changes. (slice 1)
- A skill folder listed for Quick is copied whole for team skills too, not only for vendored ones, so reference files a team skill ships reach Quick as they reach Kiro; an `id:` line the template already carries is dropped, because ids come only from the ids file; nothing you approved changes. (slice 1)
- `docs/amazon-quick.md` does not link `docs/research/quick/`, which stays uncommitted: its exports carry real skill ids of your Quick account, which the new docs say never to commit; nothing you approved changes. (slice 4)
- Kirocrew on a throwaway HOME shows only its setup screen, because kiro-cli's sign-in lives under HOME; `crew-drive.mjs` links (never copies) `~/Library/Application Support/kiro-cli` into the throwaway HOME, as `record.sh` links the SOPS key, and clicks through kirocrew's first-run Privacy and Customize dialogs (its heartbeat is paused for a non-default data folder); nothing you approved changes. (slice 5)
- The kirocrew scene opens the chat's agent picker (the dashboard's agent list) rather than a separate agents page, and the IDE and kirocrew clips are trimmed from their marks files like the app clip; nothing you approved changes. (slice 5)
- Kiro IDE also lists kirocrew's own agents once a kirocrew gateway has run on the same Kiro home, because kirocrew writes `kirocrew*.json` into the shared `agents/` folder; the README's "Works with" says so, and the showcase records the IDE before kirocrew; nothing you approved changes. (slice 4)
- The first full recording hung in the Kiro IDE scene (a blank window, then a close that never returned) and was stopped at your request; `ide-drive.mjs` now bounds every wait, gives up after `KIROKIT_IDE_DEADLINE` seconds (default 300) and kills its Kiro, and retries once on a fresh profile when the chat panel never appears; a test drives it with a Kiro that never starts; nothing you approved changes. (slice 5)
- `record.sh` now installs playwright-core's own ffmpeg after `npm ci`: since the playwright-core 1.64.1 update (#22) the recorder needed `ffmpeg-1013`, which this Mac's cache did not have, so no video scene could record; nothing you approved changes. (slice 5)
- Outline cards kept current while building, no user decision changed: T4 and T7 Files and verify commands match the files that changed (generated goldens listed, desktop tests moved to T7, desktop/app.go added); T5 and T10 Test first describe the evidence actually used (an in-test export instead of docs/research, mutations in a throwaway worktree); T10 adds testdata/golden/platform-team/context.json; T12 adds AGENTS.md (showcase line). (the Plan, while building)
- After verification round 1 judged 98b2f8c, four files changed: the reviewer's own added test (`internal/quick/review_test.go`), and the fix for Codex's two findings (`internal/quick/quick.go`, `internal/quick/quick_test.go`, `docs/amazon-quick.md`, in 9b6ff3b). The fix changes Quick render behaviour, so a second fresh reviewer rechecks it rather than round 1's verdicts being kept for it; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
- The import in your Quick account: confirmed · 27 Oct, 17:32
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Approved · 27 Oct, 17:45. Branch `feat/support-amazon-quick-see-docs`.
<!-- /ql:ship -->
