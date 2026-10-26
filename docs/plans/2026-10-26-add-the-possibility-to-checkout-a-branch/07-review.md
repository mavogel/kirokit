# Check out a config repo branch

<!-- ql:links -->
Task: `docs/plans/2026-10-26-add-the-possibility-to-checkout-a-branch/`
<!-- /ql:links -->

## Why the change

Until now kirokit could only follow the newest release (or the branch a checkout happened to be on), and every desktop Update or Sync jumped back to the newest release; after this change `kirokit sync [team] --branch <name>` and a Version list in the desktop app switch to any branch or tag of the config repository and keep that choice until the person goes back to Newest release.

## What a reviewer must know

- **Output change:** `kirokit sync <team>` now prints `checked out <branch or tag> at <commit>` before the `uninstalled`/`synced` lines (it used to print it only without a team), and a plain `sync` on a detached HEAD (a chosen tag) now stays there and exits 0 instead of failing in `git pull`. Scripts that parse the first line of `sync <team>` output would notice.
- **Withdrawn amendment:** *Decided while building* still lists "`checkout.Read` takes a third argument". That change was withdrawn with your agreement; `Read(gitPath, dir)` is as approved and the Windows console hiding sits in `internal/checkout/run_windows.go`. Verification round 1 judged commit `594cacb`; round 2 judged `7ae1b97`, which contains every later code change, so round 1's verdicts on unchanged files carry forward.
- **Follow-ups from the second reviewer (notes, not failures):** (1) two presses of Switch at the same instant leave a stale "kirokit is still running a command" under the list after the switch succeeds, and on the first-launch page the page would then not close on its own (read from the code, not run); a one-line guard in `switchVersion` (`if (view.switching) return;`) would fix it. (2) The `main.js` timing fix has no automated test because `main.js` has no test harness; a live probe shows it working (`evidence/it2-probe-late-answer.png`).

## What changed

Where each new piece lives, and the one call path that crosses from the app into git:

```diff
 internal/checkout/            + new package, shared by CLI and app
+  checkout.go                  CheckRefName (name rule), Read (offline view of a clone's refs)
+  run_windows.go / run_unix.go hide git's console window on Windows
 cmd/kirokit/commands/
~  sync.go                      --branch, updateCheckout(latestTag, ref), switchTo, checkedOut
~  setup.go                     --branch uses checkout.CheckRefName
 desktop/
~  settings.go                  Settings.Ref ("" = Newest release); SaveSettings keeps it
~  clone.go                     a successful clone resets Ref
+  version.go                   App.Checkout, App.SwitchRef, ref:done
 desktop/frontend/src/
+  pages/version.js             refArgs, versionOptions, sourceLine
~  pages/settings.js, welcome.js, install.js, main.js, api.js   the Version select, first-launch choice, source line
```

How `sync --branch <name>` decides, before anything in the checkout moves:

```text
fetch --tags --prune
tracked changes?                 -> refuse "local changes in <dir>"
origin/<name> exists?
  no local <name>                -> checkout --track -b <name> origin/<name>
  local <name> not behind origin -> refuse "local branch <name> has commits that are not on origin/<name>"
  else                           -> checkout <name> --, merge --ff-only origin/<name>
refs/tags/<name> exists?         -> checkout --detach refs/tags/<name>
else                             -> refuse "no branch or tag of that name on origin"
plain sync on a detached HEAD    -> no pull (a chosen tag stays)
```

The desktop's data shapes:

```diff
 Settings { repoUrl, kiroHome, repoDir,
+           ref }                                  // saved only after a switch exited 0
+Refs { current: {kind: branch|tag|"", name, commit},
+       branches[], tags[], newest }              // read from local refs, no network
+App.SwitchRef(ref) -> kirokit sync [team] --branch <ref> | --latest-tag   (kirokit tab)
+event ref:done {ok, error}                       // error = the CLI's "kirokit sync: ..." line
 Update / Sync / first-launch Continue: args from the saved ref, never --latest-tag on their own
```

## Screens compared

<!-- ql:screens -->
- **branch-picker#settings-version** · 1180 × 1100 px · light · unavailable: Page capture timed out: selector #version-source .badge, #version-source unavailable · current
  - Independent review: No comparison recorded.
- **branch-picker#welcome-version** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
- **branch-picker#install-source** · 1180 × 820 px · light · unavailable: Page capture timed out: selector #page-install p.source unavailable · current
  - Independent review: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
3 slices · 11 tasks · 11 commits · 25 files · +1396 −64

**Slice 1 · the CLI switches to a branch or tag and keeps it** · T1, T2, T3, T4 · 7 files · +618 −27

- `README.md` · T4, T11 · +16 −5
- `cmd/kirokit/commands/setup.go` · T1, T2 · +6 −3
- `cmd/kirokit/commands/setup_test.go` · T1 · +6 −1
- `cmd/kirokit/commands/sync.go` · T2, T3 · +89 −13
- `cmd/kirokit/commands/sync_test.go` · T2, T3 · +263 −5
- `internal/checkout/checkout.go` · T1, T5 · new · +117 −0
- `internal/checkout/checkout_test.go` · T1, T5 · new · +121 −0

**Slice 2 · the desktop app reads the clone's versions and switches them** · T5, T6, T7 · 9 files · +389 −6

- `desktop/clone.go` · T6 · +3 −2
- `desktop/clone_test.go` · T6 · +18 −0
- `desktop/runner_test.go` · T7 · +4 −0
- `desktop/settings.go` · T6 · +8 −4
- `desktop/settings_test.go` · T6 · +22 −0
- `desktop/version.go` · T7 · new · +102 −0
- `desktop/version_test.go` · T7 · new · +202 −0
- `internal/checkout/run_unix.go` · T5 · new · +8 −0
- `internal/checkout/run_windows.go` · T5 · new · +22 −0

**Slice 3 · the desktop pages pick, keep and show the version** · T8, T9, T10, T11 · 9 files · +389 −31

- `desktop/frontend/index.html` · T9 · +3 −1
- `desktop/frontend/src/api.js` · T8 · +10 −0
- `desktop/frontend/src/main.js` · T8, T9, T10 · +64 −5
- `desktop/frontend/src/pages/install.js` · T8, T9 · +8 −1
- `desktop/frontend/src/pages/pages.test.js` · T8, T9, T10 · +158 −6
- `desktop/frontend/src/pages/settings.js` · T8, T9 · +49 −11
- `desktop/frontend/src/pages/version.js` · T8, T9 · new · +46 −0
- `desktop/frontend/src/pages/welcome.js` · T10 · +45 −6
- `desktop/frontend/src/style.css` · T9 · +6 −1
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. The CLI switches an existing config checkout to a named branch or tag.** Scenario 1: the CLI switches to a branch and follows it, after slice 1 · passed · `evidence/it2-scenario-1-2-cli.txt`
- **2. Later CLI syncs keep the choice until the person goes back to releases.** Scenario 1: the CLI switches to a branch and follows it, after slice 1 · passed · `evidence/it2-scenario-1-2-cli.txt`
- **3. The desktop app lets the person pick a branch or tag from a list, on Settings and on the first-launch clone page.** Scenario 3: the desktop app switches, remembers and shows the version, after slice 3 · Scenario 4: first launch asks for the version after the clone, after slice 3 · passed · `evidence/it2-scenario-4-step-2.png`
- **4. Desktop Update and Sync keep following the choice.** Scenario 3: the desktop app switches, remembers and shows the version, after slice 3 · passed · `evidence/it2-s3c-app.txt`
- **5. The desktop app shows whether it is on a branch, a tag or the newest release, and which one.** Scenario 3: the desktop app switches, remembers and shows the version, after slice 3 · passed · `evidence/it2-done-5-settings-pinned.png`
- **6. An unknown name or a checkout with local changes fails clearly and changes nothing.** Scenario 2: a bad switch changes nothing, after slice 1 · passed · `evidence/it2-scenario-1-2-cli.txt`
- **7. The README describes the new flag and the desktop control.** Scenario 5: the README describes the flag and the control, after slice 3 · passed · `evidence/it2-scenario-5-readme.txt`
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No change to `kirokit setup --branch`.** It already clones at a branch or tag.
- **No change to `scripts/kirokit-sync-setup.sh`.** A scheduled sync with `--latest-tag` keeps moving to the newest release.
- **No commits, pushes, stashes or new branches.** kirokit only reads and switches.
- **No commit hashes in `--branch`.** Branch and tag names only, like `setup --branch`.
- **No branch list on the first-launch page before the clone.** You chose: clone first, then choose.
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
Security review, Project checks, Code review and Fix rounds ran: 7 of 7 points of Done means passed.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The CLI switches to a branch and a later plain sync follows it
- Do: in a clone of your config repository, switch to a branch, then sync again without the flag.
- Run: `task build && ./kirokit sync --repo <clone> --branch main && ./kirokit sync --repo <clone>`
- Expect: both print `checked out main at <commit>`; `git -C <clone> symbolic-ref --short HEAD` prints `main`.
- Saw: the checking agent ran Scenario 1 on a fresh build against a throwaway origin; the branch was followed, a tag stayed, `--latest-tag` went back to the newest release — `evidence/it2-scenario-1-2-cli.txt`

#### 2. A switch that would lose work is refused and changes nothing
- Do: edit a tracked file in the clone, then try to switch.
- Run: `./kirokit sync --repo <clone> --branch v1.0.0`
- Expect: exit 1 with `kirokit sync: local changes in <clone>: commit or discard them, then try again`; HEAD and your edit unchanged.
- Saw: refused with that message, HEAD and installed team unchanged; the same for an unknown name and for local commits — `evidence/it2-scenario-1-2-cli.txt`

#### 3. The app picks a version, keeps it, and says where the templates come from
- Do: run `task desktop:dev`, open Settings, pick a branch in Version and press "Switch to <branch> and update"; then open Install and sync and press Sync.
- Expect: the kirokit tab shows `▸ kirokit sync <team> --branch <branch>` both times; Settings reads "In use: branch <branch> at … not a release" and Install reads "Templates from branch <branch> at …".
- Saw: driven in the running app with a throwaway home; Sync and Update kept `--branch` even after an app restart, and choosing Newest release ran `--latest-tag` — `evidence/it2-scenario-3-step-5.png`, `evidence/it2-s3c-app.txt`

### Not verified

- QualityLayer's own screen capture (`evidence shot`) could not show the app with data, because its browser opens the page without the app's Go backend; the screen comparisons rest on Playwright captures of the running app instead (`evidence/it2-*.png`, `evidence/it1-artifact-branch-picker.png`).
- The desktop app was run on macOS only; Windows and Linux are covered by vet and the per-OS build tags, not by a live run.
- A double press of Switch at the same instant (see *What a reviewer must know*) was reproduced but not fixed.

### Docs

- `README.md`: *Sync (automatic updates)* documents `--branch`, the plain-sync behaviour, every refusal and the `checked out` line; *Desktop app* documents the Version list, the saved choice and the source line, and no longer says Sync and Update always use `--latest-tag`.

## Accepted with open items

<!-- ql:accepted -->
None: the build passed its checks.
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- Untracked files do not block a switch — they are not lost by a checkout, and git itself refuses when a switch would overwrite one. (Plan)
- The remote is always `origin` — `setup`, the desktop clone and `git clone` all name it so. (Plan)
- `sync` prints `checked out <branch|tag> at <commit>` with a team too — Done means 1 asks for it; today the line appears only without a team. (Plan)
- `--branch` and `--latest-tag` together are refused — each names a different target. (Plan)
- The name rule is `setup --branch`'s rule, shared — not empty, no leading `-`, no whitespace; anything else is for git to find or not. (Plan)
- The Update button's label changes when another entry is picked — so a person can see that pressing it switches, not just updates. (Plan)
- The source lines mark anything but Newest release, judged by the saved choice — a forgotten branch or pin should stand out, and a pinned tag equal to today's newest release still is a pin. (Plan)
- The list sorts tags newest first and branches by name — the same order `--latest-tag` uses for tags. (Plan)
- A switch made by hand in the app's clone is shown, not adopted — the source lines show what git says is checked out, while the buttons follow the choice saved in the app; the next Update or Sync goes back to that choice, which is the one the person made in the app. (Plan)
- `Catalog.Tag` stays as it is — the new checkout read replaces its only purpose for display, and removing it is unrelated cleanup. (Plan)
- Checkout.Read takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (hideConsole lives in package main and cannot be called from internal/). Affects T5, T7; files internal/checkout/checkout.go, desktop/version.go. (the Plan, while building)
- The branch name in `checked out <branch> at <commit>` (and `checkout.Read`'s current branch) is cut from the full ref, because `git symbolic-ref --short` prints `heads/<name>` when a tag has the same name; the test with a branch and a tag both named `both` caught it; nothing you approved changes. (slice 1)
- The Windows console flag is the documented `CREATE_NO_WINDOW` value through `syscall`, because the root module does not depend on `golang.org/x/sys` and one flag does not justify adding it; nothing you approved changes. (slice 2)
- `ref:done`'s message drops terminal escape sequences before it looks for the `kirokit sync:` line, since the kirokit tab runs in a pseudo-terminal; nothing you approved changes. (slice 2)
- Under Newest release, a clone switched by hand to an older tag shows "tag <name> … not the newest release" instead of the newest-release badge, so the line never claims a release the clone is not on; nothing you approved changes. (slice 3)
- Outline only: T1's Files list adds cmd/kirokit/commands/setup_test.go, which T1 extended to assert the unchanged setup --branch message. No Plan text, Done means, scope or user decision changes. (the Plan, while building)
- Verification round 1 was recorded while 5 files changed, 5 of them code the run exercised; rerun it or keep it and say why here; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
Nothing was left for you to settle.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Approved · 26 Oct, 19:12. Branch `feat/add-possibility-checkout-branch-config`.
<!-- /ql:ship -->
