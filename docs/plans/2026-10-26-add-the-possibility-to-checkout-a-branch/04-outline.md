---
type: outline
---

# Outline: Check out a config repo branch

## Summary

Three slices and eleven tasks: first the CLI (`sync --branch`, the detached plain sync and the README Sync section), then the desktop's Go side (reading the clone's refs, saving the choice, `SwitchRef`), then the desktop pages and the README desktop section. The risky part is slice 1, which runs `git checkout` in a person's checkout; its refusals are tested before anything else builds on it.

## How the build runs

<!-- ql:waves -->
<!-- /ql:waves -->

## How each point is proved

<!-- ql:proof-map -->
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** In a clone of a test repository with a `feature` branch, `v1.0.0` and `v1.1.0`, `kirokit sync <team> --branch feature` leaves HEAD on branch `feature` at `origin/feature` with the team installed from it; a later plain `kirokit sync <team>` after a new commit on `origin/feature` moves HEAD to that commit; and in the desktop app, after `feature` is chosen in Settings, the Install page's Sync button runs `kirokit sync <team> --branch feature`, not `--latest-tag`.
- **Truth:** every `--branch` failure returns before `install.Check`, so the installed team and its records are unchanged.

## Decided while outlining

- `checkout.Read` takes a third argument, a hook run on each git command, so the desktop can hide git's console window on Windows (`hideConsole`); the CLI side never calls `Read`.
- `updateCheckout(latestTag bool, ref string)` is the mode: `ref != ""` switches, `latestTag` picks the newest tag, neither follows; `setup` calls `updateCheckout(true, "")`.
- A local branch is checked out with `git checkout --quiet <name> --`, because `checkout refs/heads/<name>` would detach HEAD; the trailing `--` keeps git from reading the name as a path.
- The list leaves out a tag that has the same name as a branch in the frontend's option builder, so `checkout.Refs` stays a faithful copy of git's refs.
- The source line's badge: a branch or a plain commit is always "not a release"; a tag is "newest release" when the saved choice is Newest release and "pinned · does not follow new releases" otherwise.
- A saved choice that is no longer on `origin` still appears in the list (marked "not on origin") so the select can show it.

## Slice 1: the CLI switches to a branch or tag and keeps it

After this slice `kirokit sync [team] --branch <name>` switches the checkout to a branch or tag, refusing local changes, local commits and unknown names, and a plain `sync` follows a branch and stays on a tag. `setup --branch` shares the name rule, and the README's Sync section describes it.

```diff
 internal/
+├── checkout/checkout.go        + CheckRefName
 cmd/kirokit/commands/
 ├── setup.go                    ~ uses checkout.CheckRefName
 ├── sync.go                     ~ --branch, updateCheckout(latestTag, ref), detached follow, checked out line always
 └── sync_test.go                ~ switch, refusals, detached follow
 README.md                       ~ Sync section
```

**Risky:** it runs `git checkout` in the person's config checkout; the refusals are its first tests.

- [ ] T1 `setup --branch` and `sync --branch` share one name rule
  - **Objective:** Move setup's branch-name check into `internal/checkout.CheckRefName` so `sync` can use the same rule and message.
  - **Design:** "`internal/checkout` holds the name rule and the read-only view of a checkout"
  - **Files:** `internal/checkout/checkout.go`, `internal/checkout/checkout_test.go`, `cmd/kirokit/commands/setup.go`
  - **Test first:** `TestCheckRefName` in `internal/checkout/checkout_test.go` refuses `""`, `-b`, `--upload-pack=x`, `has space`, `a\tb` and accepts `main`, `feature/x`, `v1.2.0`; it fails today because the package does not exist.
  - **Definition of done:**
    - `CheckRefName` refuses an empty name, a leading `-` and any whitespace, and accepts any other name.
    - `kirokit setup --branch` still refuses the same names with `kirokit setup: --branch: must be a branch or tag name, not empty, starting with '-' or containing whitespace` (existing `TestSetup_BranchRefusesBadOrUnknownRef` keeps passing).
    - verify: `go test ./internal/checkout ./cmd/kirokit/commands -run 'TestCheckRefName|TestSetup'`

- [ ] T2 `sync --branch` switches only when nothing local can be lost
  - **Objective:** Add `--branch` to `sync`; with it, `updateCheckout` fetches, refuses tracked changes, local commits and unknown names, then checks out the branch (fast-forwarded) or the tag (detached).
  - **Design:** "`kirokit sync --branch` and the detached plain sync", "plan:decision:d1", "plan:decision:d3"
  - **Files:** `cmd/kirokit/commands/sync.go`, `cmd/kirokit/commands/sync_test.go`, `cmd/kirokit/commands/setup.go`
  - **Test first:** `TestSync_BranchSwitchesToABranchOrTag` (an origin with `feature`, `v1.0.0`, `v1.1.0`: `sync --branch feature` leaves `symbolic-ref HEAD` = `refs/heads/feature` at `origin/feature`; `--branch v1.0.0` detaches at the tag; a branch and a tag both named `both` resolve to the branch) and `TestSync_BranchRefusals` fail today with "unknown flag: --branch".
  - **Definition of done:**
    - A remote-only branch is created tracking `origin/<name>`; an existing local branch behind `origin` is fast-forwarded; a tag leaves a detached HEAD at the tag's commit.
    - A tracked edit refuses with `kirokit sync: local changes in <dir>: commit or discard them, then try again`; HEAD and the edited file are unchanged.
    - A local branch with commits not on `origin/<name>` refuses with `kirokit sync: --branch <name>: local branch <name> has commits that are not on origin/<name>`; HEAD is unchanged.
    - An unknown name refuses with `kirokit sync: --branch <name>: no branch or tag of that name on origin`; `--upload-pack=x`, `-b`, `has space` refuse with the shared name-rule message before any git runs.
    - `--branch` with `--latest-tag` refuses with `kirokit sync: --branch and --latest-tag cannot be used together`.
    - With a team, every refusal leaves the installed team's records unchanged and prints no `uninstalled` line.
    - Untracked files do not block a switch.
    - verify: `go test ./cmd/kirokit/commands -run 'TestSync|TestUpdateCheckout|TestSetup'`

- [ ] T3 A plain `sync` stays on a tag and every sync says where it is
  - **Objective:** On a detached HEAD a plain `sync` skips the pull instead of failing, and `checked out <branch|tag> at <commit>` prints with a team too.
  - **Design:** "`kirokit sync --branch` and the detached plain sync", "plan:decision:d2"
  - **Files:** `cmd/kirokit/commands/sync.go`, `cmd/kirokit/commands/sync_test.go`
  - **Test first:** `TestSync_PlainSyncFollowsTheChosenBranchAndStaysOnATag`: after `--branch feature` and a new commit on `origin/feature`, a plain `sync` moves HEAD to it; after `--branch v1.0.0` a plain `sync` exits 0 and HEAD stays at `v1.0.0` (today git refuses with "You are not currently on a branch").
  - **Definition of done:**
    - A plain sync on a branch fast-forwards as today; on a detached HEAD it does not pull and does not fail.
    - Output order is `updated <before> -> <after>` (when HEAD moved), `checked out <branch or tag> at <commit>`, then for a team the `uninstalled`, `pruned`, `synced` lines; a detached HEAD on no tag prints the commit as the name.
    - `--latest-tag` behaves as today (`TestUpdateCheckout`, `TestUpdateCheckout_LatestTagWithoutTagsFails`).
    - verify: `go test ./cmd/kirokit/commands`

- [ ] T4 README's Sync section documents `--branch`
  - **Objective:** Describe the flag, what a plain sync does on a branch and on a tag, the refusals, the branch-over-tag rule and the `checked out` line with a team.
  - **Design:** "README documents the flag and the version list"
  - **Files:** `README.md`
  - **Trivial:** documentation only; checked by reading the rendered section against T2 and T3's messages.
  - **Definition of done:**
    - The Sync section has a `kirokit sync <team> --branch <name>` example and every refusal message from T2 word for word.
    - verify: `grep -n -e '--branch <name>' -e 'cannot be used together' README.md`

**Validation:** automated: `go test ./internal/checkout ./cmd/kirokit/commands` · `go vet ./...` · `task lint` · manual: none

## Slice 2: the desktop app reads the clone's versions and switches them

**Depends on:** Slice 1

After this slice the desktop's Go side can list the clone's branches and tags without the network, say what is checked out, switch through the CLI in the kirokit tab and remember the choice only once the switch worked. A new repository address resets it to Newest release.

```diff
 internal/checkout/checkout.go   ~ Ref, Refs, Read
 desktop/
 ├── settings.go                 ~ Settings.Ref; SaveSettings keeps it
 ├── clone.go                    ~ Clone resets Ref
+├── version.go                  + App.Checkout, App.SwitchRef, ref:done
+└── version_test.go             + real kirokit against a throwaway origin
```

- [ ] T5 `checkout.Read` lists the clone's branches, tags and current version offline
  - **Objective:** Read what the checkout is on and the `origin` branches and tags from local refs only.
  - **Design:** "`internal/checkout` holds the name rule and the read-only view of a checkout", "plan:decision:d5"
  - **Files:** `internal/checkout/checkout.go`, `internal/checkout/checkout_test.go`
  - **Test first:** `TestRead` on a clone of an origin with `main`, `feature`, `v1.0.0`, `v1.10.0`, `nightly`: wants `Branches` `[feature main]` (no `HEAD`), `Tags` `[v1.10.0 v1.0.0 nightly]`-style `-v:refname` order, `Newest` `v1.10.0`, `Current` `{branch main <short>}`; after `checkout --detach v1.0.0` `{tag v1.0.0 …}`; on an untagged detached commit `{"" "" <short>}`. Fails today: `Read` does not exist.
  - **Definition of done:**
    - A missing folder or a folder that is not a checkout returns an error; a clone with no tags returns empty (non-nil) lists and `Newest` "".
    - Deleting the origin folder after the clone does not change the result (no network).
    - GIT_* variables of the caller do not redirect it.
    - verify: `go test ./internal/checkout`

- [ ] T6 Settings remember the version; a new address resets it
  - **Objective:** Add `Settings.Ref`; `SaveSettings` keeps the saved `Ref` whatever the form sends; a successful `Clone` stores the address and resets `Ref`.
  - **Design:** "Desktop: `Settings.Ref`, `App.Checkout`, `App.SwitchRef`", "plan:decision:d4"
  - **Files:** `desktop/settings.go`, `desktop/settings_test.go`, `desktop/clone.go`, `desktop/clone_test.go`
  - **Test first:** `TestSaveSettings_KeepsTheSavedRef` (write `Ref: "feature"`, call `SaveSettings` with a form without it, read back `feature`) and an assertion in the clone test that a successful clone leaves `Ref` "" after one was saved; both fail because `Ref` does not exist.
  - **Definition of done:**
    - `SaveSettings` overwrites only `RepoURL`, `KiroHome`, `RepoDir`; a refused save changes nothing.
    - A failed clone keeps the old `Ref`.
    - verify: `cd desktop && go test -run 'TestSaveSettings|TestClone|TestRepoDir' .`

- [ ] T7 `App.Checkout` and `App.SwitchRef` switch through the kirokit tab
  - **Objective:** Expose the clone's refs to the pages, and run `kirokit sync [team] --branch <ref>` (or `--latest-tag` for "") under the one-command lock, saving `Ref` only on exit 0 and emitting `ref:done`.
  - **Design:** "Desktop: `Settings.Ref`, `App.Checkout`, `App.SwitchRef`", "plan:decision:d4"
  - **Files:** `desktop/version.go`, `desktop/version_test.go`, `desktop/runner_test.go`
  - **Test first:** `TestSwitchRef_SavesTheChoiceOnlyAfterTheSwitch` with the real kirokit: after a clone of an origin with `feature`, `SwitchRef("feature")` ends with `ref:done {ok:true}`, `Settings.Ref` = `feature`, the clone on branch `feature`, and the tab header `▸ kirokit sync --branch feature`. Fails: `SwitchRef` does not exist.
  - **Definition of done:**
    - With an installed team the command is `sync <team> --branch <ref>`; with `""` it is `sync --latest-tag` and saves `Ref` "".
    - A switch refused by the CLI (unknown name, local changes) emits `ref:done {ok:false, error:<the kirokit sync: line>}` and leaves `Ref` unchanged; a run without such a line reports `kirokit sync failed (exit N). See the kirokit tab.`
    - A bad name (`-b`, `has space`) or a command already running is the returned error, and no process starts.
    - `Checkout()` before a clone returns empty refs and no error; after a clone it returns `Current` and the lists.
    - The catalog re-renders after the switch (it is a `kirokit sync` run).
    - verify: `cd desktop && go test -run 'TestSwitchRef|TestCheckout|TestRunner|TestClone' .`

**Validation:** automated: `go test ./internal/checkout` · `cd desktop && go test ./...` · `task lint` · manual: none

## Slice 3: the desktop pages pick, keep and show the version

**Depends on:** Slice 2

After this slice Settings shows what is in use and a Version list, the first-launch page asks for the version after the clone, every Update and Sync button keeps the choice, and Install and sync names where the templates come from. The README's desktop section says so.

```diff
 desktop/frontend/
 ├── index.html                  ~ Settings Templates: In use line, Version select
 └── src/
     ├── api.js                  ~ checkout(), switchRef(ref)
     ├── main.js                 ~ view.ref, view.checkout, ref:done, first-launch choosing
     ├── style.css               ~ .field select, .source
     └── pages/
+        ├── version.js          + refArgs, versionOptions, sourceLine
         ├── install.js          ~ syncArgs(team, ref), source line
         ├── settings.js         ~ updateArgs(installs, ref), select, Switch label
         ├── welcome.js          ~ choosing state
         └── pages.test.js       ~ args, list, source line, welcome states
 README.md                       ~ Desktop app section
```

- [ ] T8 Update and Sync build their command from the saved choice
  - **Objective:** `refArgs(ref)` gives `['--branch', ref]` or `['--latest-tag']`; `syncArgs(team, ref)` and `updateArgs(installs, ref)` use it, and `main.js` passes `view.ref` (from `State().settings.ref`); `api.js` gains `checkout()` and `switchRef(ref)`.
  - **Design:** "Desktop: `Settings.Ref`, `App.Checkout`, `App.SwitchRef`", "plan:decision:d4"
  - **Files:** `desktop/frontend/src/pages/version.js`, `desktop/frontend/src/pages/install.js`, `desktop/frontend/src/pages/settings.js`, `desktop/frontend/src/api.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** the existing `Buttons run the commands the design names` test gains `syncArgs('alpha', 'feature')` → `['sync', 'alpha', '--branch', 'feature']` and `updateArgs([{Team:'alpha'}], 'v1.0.0')` → `['sync', 'alpha', '--branch', 'v1.0.0']`; it fails because the second argument is ignored.
  - **Definition of done:**
    - With `ref` "" both still send `--latest-tag`; with a ref neither sends `--latest-tag`.
    - verify: `cd desktop/frontend && node --test`

- [ ] T9 Install and Settings show where the templates come from; Settings offers the list
  - **Objective:** `sourceLine(view, lead)` draws "Templates from …" on Install and "In use: …" on Settings; Settings gains the Version select (Newest release, Branches, Tags) and Update switches when another entry is picked.
  - **Design:** "Desktop: `Settings.Ref`, `App.Checkout`, `App.SwitchRef`", "plan:decision:d3", "plan:decision:d5"
  - **Files:** `desktop/frontend/index.html`, `desktop/frontend/src/style.css`, `desktop/frontend/src/pages/version.js`, `desktop/frontend/src/pages/install.js`, `desktop/frontend/src/pages/settings.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** `versionOptions` and `sourceLine` tests: the first option reads `Newest release · follows new releases (now v1.3.0)` (or `Newest release (none yet)`), branches then tags under optgroups, a tag named like a branch left out, values escaped; the line for a branch reads `Templates from branch <b>main</b> at <b>b81d0a7</b>` with `not a release`, for a tag with a saved ref `pinned · does not follow new releases`, for a tag with ref "" `newest release`. They fail: `version.js` does not exist.
  - **Definition of done:**
    - Settings' Update reads `Update` when the select equals the saved choice and `Switch to <name> and update` otherwise; pressing it runs `updateArgs` or calls `switchRef`.
    - A failed switch shows the `ref:done` error under the select (`aria-invalid` on it) and the saved choice stays.
    - The select keeps the person's pick across redraws and is disabled while a command runs.
    - Before a clone no source line is drawn.
    - verify: `cd desktop/frontend && node --test && npm run build`

- [ ] T10 First launch asks for the version after the clone
  - **Objective:** After `clone:done` ok from the welcome page, the page stays (`welcome.choosing`) and shows the Version select with Newest release preselected (the default branch when there is no release); Continue keeps it, another entry switches first.
  - **Design:** "Desktop: `Settings.Ref`, `App.Checkout`, `App.SwitchRef`"
  - **Files:** `desktop/frontend/src/pages/welcome.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** `showWelcome({cloned:true, welcome:{choosing:true}})` is true, and `welcomeView` while choosing renders the select, `Continue` for "", `Switch to main and continue` for `main`, `Switching…` disabled while switching, the error under the select after a failure, and the no-release lede; they fail because there is no choosing state.
  - **Definition of done:**
    - Continue with Newest release ends `choosing` without running anything; another entry calls `switchRef` and ends `choosing` only on `ref:done` ok.
    - A failed switch keeps the page with the reason under the select; the person can pick again or Continue with Newest release.
    - A clone started from Settings never shows this page.
    - verify: `cd desktop/frontend && node --test`

- [ ] T11 README's desktop section describes the version list
  - **Objective:** Describe the Version select on Settings and after the first clone, that Update, Sync and Continue use the saved choice, that new branches appear after the next Update, and the source line; replace the `--latest-tag` wording for Sync and Update with "the chosen version".
  - **Design:** "README documents the flag and the version list"
  - **Files:** `README.md`
  - **Trivial:** documentation only; checked by reading it against T8–T10.
  - **Definition of done:**
    - verify: `grep -n -e 'Version' -e 'chosen version' README.md`

**Validation:** automated: `cd desktop/frontend && node --test` · `cd desktop/frontend && npm run build` · `task desktop:test` · `task ci` · manual: the Settings page, first-launch page and Install page in the running app match the Branch picker artifact.

## Program design

### `checkout.Read` takes a command hook

`func Read(gitPath, dir string, prepare func(*exec.Cmd)) (Refs, error)`: `prepare` (nil allowed) runs on each git command before it starts. The desktop passes `hideConsole` so git opens no console window on Windows; the contract's two-argument form is otherwise unchanged.

### `updateCheckout(latestTag bool, ref string)`

`ref != ""` takes the switch path of decision 1; else `latestTag` checks out the newest `v*` tag; else follow mode (fast-forward on a branch, nothing on a detached HEAD). The flag check makes `latestTag && ref != ""` unreachable.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Name rule, read-only checkout view | `internal/checkout/checkout.go` | `strings`, `os/exec` |
| Switch | `cmd/kirokit/commands/sync.go` (`updateCheckout`, `switchRef`) | `gitOutput` |
| Desktop switch and read | `desktop/version.go` | `reserve`/`runStep`/`release`, `loginEnv`, `writeSettings` |
| Version list and source line | `desktop/frontend/src/pages/version.js` | `esc` from `html.js` |

## End-to-end scenarios

### Scenario 1: the CLI switches to a branch and follows it

**Runs after:** slice 1
**Proves:** Done means 1, 2
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `go test ./cmd/kirokit/commands -run 'TestSync_BranchSwitchesToABranchOrTag|TestSync_PlainSyncFollowsTheChosenBranchAndStaysOnATag' -count=1` | `ok` |
| 2 | Build with `task build`; in a scratch folder make an origin with a team, a `feature` branch, tags `v1.0.0` and `v1.1.0`, clone it, then run `./kirokit sync <team> --repo <clone> --branch feature` with `KIRO_HOME` set to a scratch folder | prints `checked out feature at <commit>` and `synced <team> into <home> at <commit>`; `git -C <clone> symbolic-ref --short HEAD` prints `feature` |
| 3 | Commit to `feature` in the origin, then run `./kirokit sync <team> --repo <clone>` | prints `updated <old> -> <new>` and `checked out feature at <new>` |
| 4 | Run `./kirokit sync <team> --repo <clone> --branch v1.0.0`, then `./kirokit sync <team> --repo <clone>` | both exit 0 and print `checked out v1.0.0 at <commit>` |
| 5 | Run `./kirokit sync <team> --repo <clone> --latest-tag` | prints `checked out v1.1.0 at <commit>` |

### Scenario 2: a bad switch changes nothing

**Runs after:** slice 1
**Proves:** Done means 6
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `go test ./cmd/kirokit/commands -run 'TestSync_BranchRefusals' -count=1` | `ok` |
| 2 | In Scenario 1's clone, edit a tracked file and run `./kirokit sync <team> --repo <clone> --branch main` | exits 1 with `kirokit sync: local changes in <clone>: commit or discard them, then try again`; HEAD, the edit and `<home>/.kirokit-installed` are unchanged |
| 3 | Restore the file and run `./kirokit sync <team> --repo <clone> --branch nope` | exits 1 with `kirokit sync: --branch nope: no branch or tag of that name on origin` |
| 4 | Run `./kirokit sync --repo <clone> --branch --upload-pack=x` | exits 1 with the name-rule message; no git runs |
| 5 | Run `./kirokit sync --repo <clone> --branch main --latest-tag` | exits 1 with `kirokit sync: --branch and --latest-tag cannot be used together` |

### Scenario 3: the desktop app switches, remembers and shows the version

**Runs after:** slice 3
**Proves:** Done means 3, 4, 5
**Screens:** `branch-picker#settings-version`, `branch-picker#install-source`

| Step | Action | Expected |
|---|---|---|
| 1 | `cd desktop && go test -run 'TestSwitchRef|TestCheckout' -count=1 .` | `ok` |
| 2 | `cd desktop/frontend && node --test` | every test passes |
| 3 | Run the app (`task desktop:dev`) on a clone of a repository with a `feature` branch and a team installed; open Settings | "In use: newest release v… at …"; Version lists Newest release, Branches, Tags |
| 4 | Pick `feature`; the button reads `Switch to feature and update`; press it | the kirokit tab shows `▸ kirokit sync <team> --branch feature`; In use reads `branch feature at … not a release` |
| 5 | Open Install and sync, press Sync | the kirokit tab shows `▸ kirokit sync <team> --branch feature`; the line reads `Templates from branch feature at …` with `not a release` |
| 6 | Back in Settings pick Newest release and press Update | the tab shows `--latest-tag`; the line shows `newest release` |

### Scenario 4: first launch asks for the version after the clone

**Runs after:** slice 3
**Proves:** Done means 3
**Screens:** `branch-picker#welcome-version`

| Step | Action | Expected |
|---|---|---|
| 1 | `cd desktop/frontend && node --test --test-name-pattern 'First launch'` | every matching test passes |
| 2 | Start the app with an empty data folder, enter a repository address with a `feature` branch and press Clone | the clone runs in the kirokit tab; the same page then shows Version with Newest release preselected and Continue |
| 3 | Pick `feature` and press `Switch to feature and continue` | the tab shows `▸ kirokit sync --branch feature`; the pages appear and Settings shows `feature` in use |

### Scenario 5: the README describes the flag and the control

**Runs after:** slice 3
**Proves:** Done means 7
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `grep -c -e 'sync <team> --branch <name>' README.md` | `1` or more |
| 2 | Read README *Sync (automatic updates)* and *Desktop app* | they name `--branch`, its refusals, the plain-sync behaviour, the Version select and that Update and Sync use the chosen version |

## Commands the build runs

<!-- ql:commands -->
<!-- /ql:commands -->
