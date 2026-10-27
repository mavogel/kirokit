# Support Quick, Kiro IDE and kirocrew; README refresh · the build

## Summary

<!-- ql:summary -->
7 of 13 tasks · 2 of 5 slices

Building now: started 27 Oct, 15:08, 18 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | A team renders its Amazon Quick bundle from the same templates | After this slice `kirokit render --quick product-team --out <dir>` writes the `user-story` skill and the `product-owner` Quick agent, with `id:` lines and pinned ids when an ids file knows them. | done |
| 2 | Quick ids come back from Quick's own exports | After this slice `kirokit quick ids <exports>` turns skills exported from Quick (files, folders, zips) into the gitignored `quick-ids.yaml`, so the next render writes the ids back. | task 2 of 2 |
| 3 | The desktop app exports a team for Quick | After this slice each team card that lists Quick content has **Export for Quick**: it runs `kirokit render --quick <team> --out <app data>/quick/<team>` in the kirokit tab, shows the counts and the folder, and opens it once. | done |
| 4 | Kiro IDE and kirocrew are confirmed, and the README starts with what kirokit is | After this slice the README is a short intro (pitch, video, why, install, first steps, Works with, docs index), and its reference sections live in `docs/` pages with every link resolving. | after 2 |
| 5 | The showcase shows Quick, Kiro IDE and kirocrew | After this slice `task showcase` records title, CLI (install product-team, Kiro chat, `render --quick`), Kiro IDE, kirocrew, the app and the end card, checked for privacy, size and a 30 to 90 second length. | after 4 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
None so far: the build follows the Plan as approved.
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

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

## Plan amended
