# Check out a config repo branch · the build

## Summary

<!-- ql:summary -->
11 of 11 tasks · 3 of 3 slices

Building now: started 26 Oct, 17:19, 28 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | the CLI switches to a branch or tag and keeps it | After this slice `kirokit sync [team] --branch <name>` switches the checkout to a branch or tag, refusing local changes, local commits and unknown names, and a plain `sync` follows a branch and stays on a tag. | done |
| 2 | the desktop app reads the clone's versions and switches them | After this slice the desktop's Go side can list the clone's branches and tags without the network, say what is checked out, switch through the CLI in the kirokit tab and remember the choice only once the switch worked. | done |
| 3 | the desktop pages pick, keep and show the version | After this slice Settings shows what is in use and a Version list, the first-launch page asks for the version after the clone, every Update and Sync button keeps the choice, and Install and sync names where the templates come from. | done |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- The Plan · Checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.
- Slice 1 · The branch name in `checked out <branch> at <commit>` (and `checkout.Read`'s current branch) is cut from the full ref, because `git symbolic-ref --short` prints `heads/<name>` when a tag has the same name; the test with a branch and a tag both named `both` caught it; nothing you approved changes.
- Slice 2 · The Windows console flag is the documented `CREATE_NO_WINDOW` value through `syscall`, because the root module does not depend on `golang.org/x/sys` and one flag does not justify adding it; nothing you approved changes.
- Slice 2 · `ref:done`'s message drops terminal escape sequences before it looks for the `kirokit sync:` line, since the kirokit tab runs in a pseudo-terminal; nothing you approved changes.
- Slice 3 · Under Newest release, a clone switched by hand to an older tag shows "tag <name> … not the newest release" instead of the newest-release badge, so the line never claims a release the clone is not on; nothing you approved changes.
- The Plan · Outline only: T1's Files list adds cmd/kirokit/commands/setup_test.go, which T1 extended to assert the unchanged setup --branch message. No Plan text, Done means, scope or user decision changes.
<!-- /ql:decisions -->

## Checks

- T1 · 0053b15 · `go test ./internal/checkout ./cmd/kirokit/commands -run 'TestCheckRefName|TestSetup'` · exit 0 · 1.2s
- T2 · 0053b15 · `go test ./cmd/kirokit/commands -run 'TestSync|TestUpdateCheckout|TestSetup'` · exit 0 · 5.6s
- T3 · 0053b15 · `go test ./cmd/kirokit/commands` · exit 0 · 7.2s
- T4 · 0053b15 · `grep -n -e '--branch <name>' -e 'cannot be used together' README.md` · exit 0 · 0s
- T5 · 0a8792a · `go test ./internal/checkout` · exit 0 · 0.9s
- T6 · 0a8792a · `cd desktop && go test -run 'TestSaveSettings|TestClone|TestRepoDir' .` · exit 0 · 3.1s
- T7 · 0a8792a · `cd desktop && go test -run 'TestSwitchRef|TestCheckout|TestRunner|TestClone' .` · exit 0 · 5.3s
- T8 · faafd4d · `cd desktop/frontend && node --test` · exit 0 · 0.3s
- T9 · faafd4d · `cd desktop/frontend && node --test && npm run build` · exit 0 · 0.8s
- T10 · faafd4d · `cd desktop/frontend && node --test` · reused: nothing changed since it passed at 17:42
- T11 · faafd4d · `grep -n -e 'Version' -e 'chosen version' README.md` · exit 0 · 0s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- Plan (amended): checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.
- Plan (agreed): Withdrew the agent's earlier change to checkout.Read: the Plan is back to the approved Read(gitPath, dir string) signature. Windows console hiding moves into build-tagged internal/checkout/run_windows.go and run_unix.go (T5 Files). Nothing else in the approved Plan changes.
- T3 (tactical): the branch name in `checked out <branch> at <commit>` (and `checkout.Read`'s current branch) is cut from the full ref, because `git symbolic-ref --short` prints `heads/<name>` when a tag has the same name; the test with a branch and a tag both named `both` caught it; nothing you approved changes.
- T5 (tactical): the Windows console flag is the documented `CREATE_NO_WINDOW` value through `syscall`, because the root module does not depend on `golang.org/x/sys` and one flag does not justify adding it; nothing you approved changes.
- T7 (tactical): `ref:done`'s message drops terminal escape sequences before it looks for the `kirokit sync:` line, since the kirokit tab runs in a pseudo-terminal; nothing you approved changes.
- T9 (tactical): under Newest release, a clone switched by hand to an older tag shows "tag <name> … not the newest release" instead of the newest-release badge, so the line never claims a release the clone is not on; nothing you approved changes.
- Slices 2 and 3 (tactical): `qualitylayer commit slice` refused, saying `desktop/clone.go` and `desktop/frontend/src/pages/welcome.js` held uncommitted edits of the showcase-video task; their diffs held only this task's hunks, so the slice files were committed with plain git and the slice recorded afterwards; nothing you approved changes.
- Plan (amended): Outline only: T1's Files list adds cmd/kirokit/commands/setup_test.go, which T1 extended to assert the unchanged setup --branch message. No Plan text, Done means, scope or user decision changes.

## Plan amended

### Plan · amended

checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go.

Amended by the agent on 2026-10-26; changed 03-plan.md.

### Plan · agreed

Withdrew the agent's earlier change to checkout.Read: the Plan is back to the approved Read(gitPath, dir string) signature. Windows console hiding moves into build-tagged internal/checkout/run_windows.go and run_unix.go (T5 Files). Nothing else in the approved Plan changes.

Agreed with the user on 2026-10-26; changed 03-plan.md, 04-outline.md.

Why: the first amendment changed an approved interface on the agent's own judgment, and the build should rest on what the person approved. The Windows need (no console window when the desktop runs git) is met without touching the contract: `internal/checkout` starts its own git commands, so it can set `HideWindow` and `CREATE_NO_WINDOW` itself in a build-tagged file, the same per-OS pattern as `desktop/procgroup_windows.go` and `internal/dirlink`. The earlier amendment above is superseded and nothing in the build relies on it. The person asked for this record in chat: "record the revert".

### Plan · amended

Outline only: T1's Files list adds cmd/kirokit/commands/setup_test.go, which T1 extended to assert the unchanged setup --branch message. No Plan text, Done means, scope or user decision changes.

Amended by the agent on 2026-10-26; changed 04-outline.md.

Why: T1's Definition of done promises that `setup --branch` keeps its exact refusal message after the rule moved to `internal/checkout`. The existing `TestSetup_BranchRefusesBadOrUnknownRef` only checked that setup failed, so T1 extended it to compare the message, and the card's Files must name every source path the task edits.
