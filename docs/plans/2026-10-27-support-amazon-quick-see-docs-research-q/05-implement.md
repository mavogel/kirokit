# Support Quick, Kiro IDE and kirocrew; README refresh · the build

## Summary

<!-- ql:summary -->
11 of 13 tasks · 3 of 5 slices

Built: started 27 Oct, 15:08.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | A team renders its Amazon Quick bundle from the same templates | After this slice `kirokit render --quick product-team --out <dir>` writes the `user-story` skill and the `product-owner` Quick agent, with `id:` lines and pinned ids when an ids file knows them. | done |
| 2 | Quick ids come back from Quick's own exports | After this slice `kirokit quick ids <exports>` turns skills exported from Quick (files, folders, zips) into the gitignored `quick-ids.yaml`, so the next render writes the ids back. | task 2 of 2 |
| 3 | The desktop app exports a team for Quick | After this slice each team card that lists Quick content has **Export for Quick**: it runs `kirokit render --quick <team> --out <app data>/quick/<team>` in the kirokit tab, shows the counts and the folder, and opens it once. | done |
| 4 | Kiro IDE and kirocrew are confirmed, and the README starts with what kirokit is | After this slice the README is a short intro (pitch, video, why, install, first steps, Works with, docs index), and its reference sections live in `docs/` pages with every link resolving. | done |
| 5 | The showcase shows Quick, Kiro IDE and kirocrew | After this slice `task showcase` records title, CLI (install product-team, Kiro chat, `render --quick`), Kiro IDE, kirocrew, the app and the end card, checked for privacy, size and a 30 to 90 second length. | task 2 of 2 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Delivered 27 Oct, 16:18

**Built** · 3
- A team renders its Amazon Quick bundle from the same templates · slice 1 · 4 tasks
- The desktop app exports a team for Quick · slice 3 · 2 tasks
- Kiro IDE and kirocrew are confirmed, and the README starts with what kirokit is · slice 4 · 3 tasks

**Verified** · 3
- Project checks passed · 6 of 6 · recorded
- Security review · done · nothing severe in the change
- Second opinion · done · 2 findings

**Left** · 4
- Independent review · Review · not done
- Fix rounds · Fix · not done
- Run the two-pass Quick round trip on your own Quick account, because Quick has no API and only you can import · needs you · T6
- Record the new showcase once Kiro IDE renders again · needs you · T13
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- Slice 1 · Its verify command lost `./desktop/...` (the desktop module is tested from `desktop/` in T7), so QualityLayer did not rerun it; `go test ./internal/golden ./cmd/kirokit/commands` passed when run by hand (96 tests) and `go test ./...` passed after it; nothing you approved changes.
- Slice 1 · `kirokit render --quick` prints the contract's `N skills, M agents` without singular forms ("1 skills, 1 agents"); the desktop card uses singular and plural; nothing you approved changes.
- Slice 1 · A skill folder listed for Quick is copied whole for team skills too, not only for vendored ones, so reference files a team skill ships reach Quick as they reach Kiro; an `id:` line the template already carries is dropped, because ids come only from the ids file; nothing you approved changes.
- Slice 4 · `docs/amazon-quick.md` does not link `docs/research/quick/`, which stays uncommitted: its exports carry real skill ids of your Quick account, which the new docs say never to commit; nothing you approved changes.
- Slice 5 · Kirocrew on a throwaway HOME shows only its setup screen, because kiro-cli's sign-in lives under HOME; `crew-drive.mjs` links (never copies) `~/Library/Application Support/kiro-cli` into the throwaway HOME, as `record.sh` links the SOPS key, and clicks through kirocrew's first-run Privacy and Customize dialogs (its heartbeat is paused for a non-default data folder); nothing you approved changes.
- Slice 5 · The kirocrew scene opens the chat's agent picker (the dashboard's agent list) rather than a separate agents page, and the IDE and kirocrew clips are trimmed from their marks files like the app clip; nothing you approved changes.
- Slice 4 · Kiro IDE also lists kirocrew's own agents once a kirocrew gateway has run on the same Kiro home, because kirocrew writes `kirocrew*.json` into the shared `agents/` folder; the README's "Works with" says so, and the showcase records the IDE before kirocrew; nothing you approved changes.
- Slice 5 · The first full recording hung in the Kiro IDE scene (a blank window, then a close that never returned) and was stopped at your request; `ide-drive.mjs` now bounds every wait, gives up after `KIROKIT_IDE_DEADLINE` seconds (default 300) and kills its Kiro, and retries once on a fresh profile when the chat panel never appears; a test drives it with a Kiro that never starts; nothing you approved changes.
- Slice 5 · `record.sh` now installs playwright-core's own ffmpeg after `npm ci`: since the playwright-core 1.64.1 update (#22) the recorder needed `ffmpeg-1013`, which this Mac's cache did not have, so no video scene could record; nothing you approved changes.
- The Plan · Outline cards kept current while building, no user decision changed: T4 and T7 Files and verify commands match the files that changed (generated goldens listed, desktop tests moved to T7, desktop/app.go added); T5 and T10 Test first describe the evidence actually used (an in-test export instead of docs/research, mutations in a throwaway worktree); T10 adds testdata/golden/platform-team/context.json; T12 adds AGENTS.md (showcase line).
<!-- /ql:decisions -->

## Checks

- T1 · 88f1977 · `go test ./internal/manifest` · exit 0 · 0.3s
- T2 · 88f1977 · `go test ./internal/quick ./internal/render` · exit 0 · 0.7s
- T3 · 88f1977 · `go test ./cmd/kirokit/commands -run 'Quick|KirokitRepo'` · exit 0 · 0.8s
- T4 · 88f1977 · `go test ./internal/golden ./cmd/kirokit/commands` · not run by QualityLayer: changed after approval
- T4 · 88f1977 · `go test ./internal/golden ./cmd/kirokit/commands` · not run by QualityLayer: changed after approval
- T5 · 243c637 · `go test ./internal/quick ./cmd/kirokit/commands -run 'Quick'` · exit 0 · 0.5s
- T7 · 6bed554 · `cd desktop && go test ./...` · exit 0 · 7.3s
- T8 · 6bed554 · `cd desktop/frontend && node --test "src/**/*.test.js"` · exit 0 · 0.3s
- T9 · 62f4623 · `grep -c product-owner docs/plans/2026-10-27-support-amazon-quick-see-docs-research-q/records/check-kirocrew-agents.json` · exit 0 · 0s
- T10 · 62f4623 · `node scripts/check-doc-links.mjs` · exit 0 · 0.2s
- T11 · 62f4623 · `node scripts/check-doc-links.mjs` · reused: nothing changed since it passed at 15:34
- T12 · 179e6e9 · `node --test scripts/showcase/` · exit 0 · 66.4s
- T12 · 179e6e9 · `node --test scripts/showcase/` · exit 1 · 67.5s
- T12 · 179e6e9 · `node --test scripts/showcase/` · exit 0 · 67.7s
- all · 98b2f8c · `task ci` · exit 0 · 11.1s
- all · 98b2f8c · `go vet ./...` · exit 0 · 0.2s
- all · 98b2f8c · `go test ./...` · exit 0 · 8.9s
- all · 98b2f8c · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- all · 98b2f8c · `task desktop:test` · exit 0 · 2.5s
- all · 98b2f8c · `node --test scripts/showcase/` · reused: nothing changed since it passed at 16:17

## Known limits

<!-- ql:limits -->
- T13 · Until task showcase records again, assets/kirokit-showcase.gif and .mp4 are still the old platform-team recording, while the README's alt text already describes the new story (product-team, render --quick, Kiro IDE, kirocrew, app).
<!-- /ql:limits -->

## Deviations

- T4 (tactical): its verify command lost `./desktop/...` (the desktop module is tested from `desktop/` in T7), so QualityLayer did not rerun it; `go test ./internal/golden ./cmd/kirokit/commands` passed when run by hand (96 tests) and `go test ./...` passed after it; nothing you approved changes.
- T4 (tactical): `kirokit render --quick` prints the contract's `N skills, M agents` without singular forms ("1 skills, 1 agents"); the desktop card uses singular and plural; nothing you approved changes.
- T2 (tactical): a skill folder listed for Quick is copied whole for team skills too, not only for vendored ones, so reference files a team skill ships reach Quick as they reach Kiro; an `id:` line the template already carries is dropped, because ids come only from the ids file; nothing you approved changes.
- Slice 1 (tactical): `qualitylayer commit slice 1` refused because it believed `cmd/kirokit/commands/render.go` held another task's uncommitted edits (that task is merged, and the file was clean when this build started), so the slice's files were committed by hand as 243c637 and registered with `commit slice 1 --commit`; nothing you approved changes.
- T10 (tactical): `docs/amazon-quick.md` does not link `docs/research/quick/`, which stays uncommitted: its exports carry real skill ids of your Quick account, which the new docs say never to commit; nothing you approved changes.
- T12 (tactical): kirocrew on a throwaway HOME shows only its setup screen, because kiro-cli's sign-in lives under HOME; `crew-drive.mjs` links (never copies) `~/Library/Application Support/kiro-cli` into the throwaway HOME, as `record.sh` links the SOPS key, and clicks through kirocrew's first-run Privacy and Customize dialogs (its heartbeat is paused for a non-default data folder); nothing you approved changes.
- T12 (tactical): the kirocrew scene opens the chat's agent picker (the dashboard's agent list) rather than a separate agents page, and the IDE and kirocrew clips are trimmed from their marks files like the app clip; nothing you approved changes.
- T9 (tactical): Kiro IDE also lists kirocrew's own agents once a kirocrew gateway has run on the same Kiro home, because kirocrew writes `kirocrew*.json` into the shared `agents/` folder; the README's "Works with" says so, and the showcase records the IDE before kirocrew; nothing you approved changes.
- T12 (tactical): the first full recording hung in the Kiro IDE scene (a blank window, then a close that never returned) and was stopped at your request; `ide-drive.mjs` now bounds every wait, gives up after `KIROKIT_IDE_DEADLINE` seconds (default 300) and kills its Kiro, and retries once on a fresh profile when the chat panel never appears; a test drives it with a Kiro that never starts; nothing you approved changes.
- T12 (tactical): `record.sh` now installs playwright-core's own ffmpeg after `npm ci`: since the playwright-core 1.64.1 update (#22) the recorder needed `ffmpeg-1013`, which this Mac's cache did not have, so no video scene could record; nothing you approved changes.
- T13: not recorded. Kiro IDE 1.4.26 shows only a blank window on this Mac, also on your own profile after a restart, so the approved IDE scene cannot be recorded now; it is a need for Review (run `task showcase` once Kiro renders again) rather than a cut scene, which would change what you approved.
- Plan (amended): Outline cards kept current while building, no user decision changed: T4 and T7 Files and verify commands match the files that changed (generated goldens listed, desktop tests moved to T7, desktop/app.go added); T5 and T10 Test first describe the evidence actually used (an in-test export instead of docs/research, mutations in a throwaway worktree); T10 adds testdata/golden/platform-team/context.json; T12 adds AGENTS.md (showcase line).

## Plan amended

### Plan · amended

Outline cards kept current while building, no user decision changed: T4 and T7 Files and verify commands match the files that changed (generated goldens listed, desktop tests moved to T7, desktop/app.go added); T5 and T10 Test first describe the evidence actually used (an in-test export instead of docs/research, mutations in a throwaway worktree); T10 adds testdata/golden/platform-team/context.json; T12 adds AGENTS.md (showcase line).

Why: each card's Files is what `qualitylayer commit slice` stages, so it has to name every file the card really changed, including regenerated goldens, or a slice commit would leave them out. The verify and Test first lines were corrected to the checks that actually ran, so the reviewer reruns those and not a command that no longer matches. No Done means point, scope item or decision of yours changed.

Amended by the agent on 2026-10-27; changed 04-outline.md.
