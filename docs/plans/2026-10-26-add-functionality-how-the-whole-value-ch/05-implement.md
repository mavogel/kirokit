# Cross-role agent harness with feedback loop · the build

## Summary

<!-- ql:summary -->
4 of 13 tasks · 1 of 5 slices

Building now: started 26 Oct, 13:07, 4 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | Every team can name owners and a feedback repository | After this slice `manifest.yaml` accepts `owner`, `owners` and `feedback.github`, rejects malformed handles and unknown agent or skill keys, and templates can read the owner table and the feedback repository. | done |
| 2 | A team can install five role agents that consult each other | After this slice `kirokit install product-team` gives the product owner, engineer, customer, QA and UX/UI designer agents, each able to call the other four, with a steering that asks the product owner to review requirements and the engineer to review a design. | next |
| 3 | `/feedback` in Kiro drafts a prefilled issue for the owner | After this slice a person in any `product-team` chat types `/feedback <what went wrong>` and gets a title, a body that @-mentions the owner, and a percent-encoded `issues/new` link in `mavogel/kirokit`, with nothing sent. | after 2 |
| 4 | The desktop app shows owners and opens a feedback issue | After this slice the Agents and Skills pages show each owner and a Feedback button that opens the prefilled GitHub issue in the browser; a team without `feedback.github` shows its buttons off with the reason, and any address that is not a GitHub issues form is refused. | next |
| 5 | `analyze-session` turns a Kiro chat into owner issues, and the README explains it all | After this slice every role agent has the `analyze-session` skill, which reads a past session and reports recommendations grouped by owner with one prefilled issue per owner, and the README documents the example team, owners, feedback and session analysis. | after 3 |
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

- T1 · 381b399 · `go test ./internal/manifest` · exit 0 · 0.4s
- T2 · 381b399 · `go test ./internal/render` · exit 0 · 0.7s
- T3 · 381b399 · `go test ./internal/render` · reused: nothing changed since it passed at 13:10
- T4 · 381b399 · `go test ./cmd/kirokit/... ./internal/golden` · exit 0 · 4.2s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

## Plan amended
