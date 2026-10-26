# Check out a config repo branch · the build

## Summary

<!-- ql:summary -->
0 of 11 tasks · 0 of 3 slices

Building now: started 26 Oct, 17:19, 7 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | the CLI switches to a branch or tag and keeps it | After this slice `kirokit sync [team] --branch <name>` switches the checkout to a branch or tag, refusing local changes, local commits and unknown names, and a plain `sync` follows a branch and stays on a tag. | next |
| 2 | the desktop app reads the clone's versions and switches them | After this slice the desktop's Go side can list the clone's branches and tags without the network, say what is checked out, switch through the CLI in the kirokit tab and remember the choice only once the switch worked. | after 1 |
| 3 | the desktop pages pick, keep and show the version | After this slice Settings shows what is in use and a Version list, the first-launch page asks for the version after the clone, every Update and Sync button keeps the choice, and Install and sync names where the templates come from. | after 2 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- The Plan · Checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.
<!-- /ql:decisions -->

## Checks

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- Plan (amended): checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.
- Plan (agreed): Withdrew the agent's earlier change to checkout.Read: the Plan is back to the approved Read(gitPath, dir string) signature. Windows console hiding moves into build-tagged internal/checkout/run_windows.go and run_unix.go (T5 Files). Nothing else in the approved Plan changes.

## Plan amended

### Plan · amended

checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.

Amended by the agent on 2026-10-26; changed 03-plan.md.

### Plan · agreed

Withdrew the agent's earlier change to checkout.Read: the Plan is back to the approved Read(gitPath, dir string) signature. Windows console hiding moves into build-tagged internal/checkout/run_windows.go and run_unix.go (T5 Files). Nothing else in the approved Plan changes.

Agreed with the user on 2026-10-26; changed 03-plan.md, 04-outline.md.

Why: the first amendment changed an approved interface on the agent's own judgment, and the build should rest on what the person approved. The Windows need (no console window when the desktop runs git) is met without touching the contract: `internal/checkout` starts its own git commands, so it can set `HideWindow` and `CREATE_NO_WINDOW` itself in a build-tagged file, the same per-OS pattern as `desktop/procgroup_windows.go` and `internal/dirlink`. The earlier amendment above is superseded and nothing in the build relies on it. The person asked for this record in chat: "record the revert".
