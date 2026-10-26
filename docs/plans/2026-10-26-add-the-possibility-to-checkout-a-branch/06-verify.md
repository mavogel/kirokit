# Check out a config repo branch · what was checked

## Summary

<!-- ql:summary -->
7 of 7 points assessed · 7 passed · 1 fix round

Passed after 2 judge runs and 1 fix round.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Unverified:** Two presses at the same instant leave a stale error after a successful switch. A Playwright `dblclick` on Switch ran the switch once and left "kirokit is still running a command" under the list, while two clicks 60 ms apart did not. Because the fix clears `view.switching` when the second call is refused, the first switch's `ref:done` would also no longer close the first-launch page in that case (found by reading the code, not run),
- **Unverified:** The race fix has no automated test. Only the live probe, which holds the `SwitchRef` answer until after `ref:done`, shows the button returning to "Update",
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 6 checks
- **Independent review** · fresh helper
  - State: done
  - Result: done · 7 of 7 points · 3 screens not captured · 6 check receipts reused
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
- **Fix rounds** · the build session
  - State: done
  - Result: done · 1 round
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
| Done means | Proof | |
|---|---|---|
| 1. The CLI switches an existing config checkout to a named branch or tag | Rerun on the CLI built at the judged commit: `kirokit sync demo --branch feature` fetched, put HEAD on `feature` at `origin/feature` (91dab21), installed demo and printed `checked out feature at 91dab21`; `--branch v1.0.0` detached at the tag. | passed |
| 2. Later CLI syncs keep the choice until the person goes back to releases | After a new commit on `origin/feature` a plain sync printed `updated 91dab21 -> f16a565` and `checked out feature at f16a565`; on `v1.0.0` a plain sync stayed at 5fda641; `--latest-tag` went to `v1.1.0`. | passed |
| 3. The desktop app lets the person pick a branch or tag from a list, on Settings and on the first-launch clone page | ![In the running app, the first-launch page (after the clone, Newest release preselected) and Settings list Newest release, then Branches, then Tags; each switch ran in the kirokit tab, and Teams and Agents went to 2 with the `feature` branch's team `beta`,,](evidence/it2-scenario-4-step-2.png) | passed |
| 4. Desktop Update and Sync keep following the choice | After a reload with `feature` saved, Update ran `kirokit sync demo --branch feature`; Install and sync's Sync ran `--branch feature`, later `--branch v1.0.0`; only picking Newest release ran `--latest-tag`,,. | passed |
| 5. The desktop app shows whether it is on a branch, a tag or the newest release, and which one | ![Settings' In use line and the Install source line read `branch feature at 2f29afd` with "not a release", `tag v1.0.0 at 819bf3b` with "pinned · does not follow new releases", and "newest release" `v1.1.0 at 432c1e1`,,](evidence/it2-done-5-settings-pinned.png) | passed |
| 6. An unknown name or a checkout with local changes fails clearly and changes nothing | In the CLI, local changes, local commits and an unknown name were refused with the planned messages, and HEAD, the edit and the install records stayed the same; in the app, a tracked edit refused the switch to `v1.0.0` with the CLI's message under the list (`aria-invalid`), and HEAD and the saved choice stayed the same after a reload,,. | passed |
| 7. The README describes the new flag and the desktop control | README *Sync* has the `--branch` example and line 313 `kirokit sync <team> --branch <name>`, plain-sync behaviour on a branch and a tag, all five refusals word for word and the branch-over-tag rule; *Desktop app* has the Version list, Update/Sync/Continue on the chosen version and the source line. | passed |
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
- Round 1 · 1 point still open
- Round 2 · everything checked passed
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `go test ./...` · exit 0 · 7.7s
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `task lint` · exit 0 · 0.8s
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `go vet ./...` · exit 0 · 0.1s
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `task ci` · exit 0 · 8.6s
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `task desktop:test` · exit 0 · 8.7s
- all · b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->
### Security

**Result:** done · nothing severe in the change
**Boundaries:** `kirokit sync --branch <name>` (CLI argument to `git` subprocess); `App.SwitchRef` (page to Wails-bound method to bundled kirokit); the `origin` remote's branch and tag names (clone to `git` output to page HTML); `settings.json` `ref` (file to sync arguments); `kirokit setup --branch` (same name rule, now shared)
- Reviewed 0053b15..HEAD, not only the first commit. `SwitchRef` calls `checkout.CheckRefName`, then `runStep` runs `exec.Command(cli, "sync", team, "--branch", ref)` with no shell. The CLI checks the name again, and a ref that starts with `-` cannot become an option. Shell metacharacters such as `;` or `$(` that git allows in ref names are never interpreted.
- `checkout.Read` runs git with an argument list and `-C <repoDir>`, with `GIT_*` stripped. It reads only local refs, takes no user input, and the two `run_*.go` files only hide the console window.
- Ref names from the clone can contain `<`, `>`, `"` and `&`. Every name interpolated into HTML in `version.js` (`versionOptions`, `sourceLine`) and `welcome.js` goes through `esc()`, which escapes `& < > " '`. `optgroup` labels are static, and the Settings button label is set with `textContent`. No injection found.
- `SaveSettings` keeps the stored `Ref`, and `Clone` resets it. A hand-edited `ref` in `settings.json` still reaches the CLI only as the value of `--branch`, where the name rule applies.
- `SwitchRef` echoes the ref into the kirokit terminal header. A caller that bypasses the page could put escape sequences in it (git ref names cannot contain control characters). That only affects the user's own terminal and needs script control of the page, so it is not recorded as a finding.
<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification

### Iteration 1

**Revision:** bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835

**Assignment:** 6a3c8957-a9d0-457d-9697-8801d267b6ac
**Judge:** fresh helper

- [PASS] check: reused receipt — all · 594cacb · `go test ./...` · exit 0 · 7.9s
- [PASS] check: reused receipt — all · 594cacb · `task lint` · exit 0 · 0.7s
- [PASS] check: reused receipt — all · 594cacb · `go vet ./...` · exit 0 · 0.2s
- [PASS] check: reused receipt — all · 594cacb · `task ci` · exit 0 · 8.1s
- [PASS] check: reused receipt — all · 594cacb · `task desktop:test` · exit 0 · 8.9s
- [PASS] check: reused receipt — all · 594cacb · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- [PASS] done 1: `kirokit sync demo --branch feature` fetched, checked out `feature` at `origin/feature`, installed demo and printed `checked out feature at 5716835`; `--branch v1.0.0` detached at the tag — `evidence/it1-scenario-1-2-cli.txt`
- [PASS] done 2: after a new commit on `origin/feature` a plain sync printed `updated 5716835 -> fb3049e` and `checked out feature at fb3049e`; on `v1.0.0` a plain sync stayed at `0e3bc75`; `--latest-tag` went to `v1.1.0` — `evidence/it1-scenario-1-2-cli.txt`
- [PASS] done 3: in the running app, Settings and the first-launch page (after the clone, with Newest release preselected) list Newest release, then Branches, then Tags; the switch ran in the kirokit tab, and Teams and Agents went from 1 to 2 with the `feature` branch's team `beta` (Skills draws from the same catalog) — `evidence/it1-scenario-4-step-2.png`, `evidence/it1-scenario-3-step-4.png`, `evidence/it1-scenario-3-step-4-teams.png`
- [PASS] done 4: after an app restart with `feature` saved, Update ran `kirokit sync demo --branch feature`; Install and sync's Sync ran `--branch feature`, and later `--branch v1.0.0`; only picking Newest release ran `--latest-tag` — `evidence/it1-done-4-5-app.txt`, `evidence/it1-done-4-update-keeps-branch.png`
- [PASS] done 5: Settings' In use line and the Install source line read `branch feature at 322191b` with "not a release", `tag v1.0.0 at 12e901f` with "pinned · does not follow new releases", and "newest release" `v1.1.0 at 501483f` — `evidence/it1-done-5-settings-pinned.png`, `evidence/it1-done-5-install-pinned.png`, `evidence/it1-scenario-3-step-3.png`
- [PASS] done 6: in the CLI, local changes and an unknown name were refused with the planned messages, and HEAD, the edit and the install records stayed the same; in the app, a tracked edit refused the switch to `v1.0.0` with the CLI's message under the list, while the saved choice and HEAD stayed the same — `evidence/it1-scenario-1-2-cli.txt`, `evidence/it1-scenario-3-probe-refused.png`
- [PASS] done 7: README *Sync* has the `--branch` example, plain-sync behaviour on a branch and a tag, all five refusals word for word and the branch-over-tag rule; *Desktop app* has the Version list, Update/Sync/Continue on the chosen version, and the source line — `evidence/it1-scenario-5-readme.txt`
- [PASS] scenario 1: all 5 steps as expected; probes: `--branch ""`, `"   "`, `a b` and a missing `--repo` are refused, untracked files do not block, and a repeated switch is a no-op — `evidence/it1-scenario-1-2-cli.txt`, `evidence/it1-scenario-1-2-tests.txt`
- [PASS] scenario 2: all 5 steps as expected, including `--branch=--upload-pack=x`; a local branch with commits not on origin is refused and HEAD stays on `main` — `evidence/it1-scenario-1-2-cli.txt`
- [PASS] scenario 3: all 6 steps as expected in the running app (Wails dev, driven with Playwright like `scripts/showcase/app-drive.mjs`); probe: a refused switch keeps Newest release after a reload — `evidence/it1-scenario-3-app.txt`
- [PASS] scenario 4: all 3 steps as expected; the clone ran in the kirokit tab, the same page offered Version with Newest release and Continue, `Switch to feature and continue` ran `kirokit sync --branch feature`, and `settings.json` holds `"ref": "feature"` — `evidence/it1-scenario-4-app.txt`
- [FAIL] scenario 5: The README's sync example does not match the scenario's search. Step 1 (`grep -c -e 'sync <team> --branch <name>' README.md`) counts 0, not 1 or more, because the example reads `kirokit sync <team> --repo ~/.kirokit-config --branch <name>`. Step 2 holds. The fix is to make the README example and the step agree — `evidence/it1-scenario-5-readme.txt`
- [PASS] dod T1: `CheckRefName` refuses empty names, a leading `-` and any whitespace, and accepts other names; `setup --branch` keeps its exact message (`TestSetup_BranchRefusesBadOrUnknownRef` compares it)
- [PASS] dod T2: a remote-only branch tracks `origin/<n>`, a local branch behind origin is fast-forwarded, and a tag detaches; all five refusals use the planned words; with a team, the records stay the same (now including local commits); an untracked file survives
- [PASS] dod T3: a branch is fast-forwarded and a tag stays without failing; output order is `updated`, `checked out`, then the team lines; an untagged detached HEAD prints its commit; the `--latest-tag` tests still pass
- [PASS] dod T4: verify grep exits 0; README *Sync* has the example and every T2 refusal word for word
- [PASS] dod T5: `TestRead` covers branches without `HEAD`, `-v:refname` tag order, `Newest`, `Current` on a branch, a tag and an untagged commit, a deleted origin and a foreign `GIT_DIR`; missing or empty folders return an error; a clone with no tags gives empty lists
- [PASS] dod T6: `SaveSettings` keeps `Ref` after both a successful and a refused save; a successful clone resets `Ref` and a failed clone keeps it
- [PASS] dod T7: with a team it runs `sync <team> --branch <ref>`, without one `sync --branch <ref>` (new test), and `--latest-tag` for ""; a refusal emits the CLI's line and keeps `Ref`; a bad name or busy tab returns an error and starts no process; the catalog re-renders
- [PASS] dod T8: "" sends `--latest-tag` and a ref sends `--branch <ref>` without `--latest-tag`; in the live app, Sync and Update sent `--branch feature`
- [PASS] dod T9: Update reads `Update` or `Switch to <name> and update` and switches; a failed switch shows under the select with `aria-invalid` and keeps the choice; controls are disabled while running; there is no source line before a clone
- [PASS] dod T10: Continue on Newest release ends `choosing` without a run; another entry switches and ends on `ref:done` ok; a failed switch keeps the page; a clone from Settings never shows it
- [PASS] dod T11: verify grep exits 0; the desktop section covers the Version list, the chosen version, new branches after the next Update and the source line
- [PASS] truth oracle: `sync demo --branch feature` left HEAD on `feature` at `origin/feature` with demo installed; a plain sync moved to the new commit; in the app, with `feature` chosen in Settings, Install's Sync ran `kirokit sync demo --branch feature`, not `--latest-tag` — `evidence/it1-scenario-1-2-cli.txt`, `evidence/it1-scenario-3-app.txt`
- [PASS] truth 1: every `--branch` refusal (local changes, local commits, unknown name) left the install records the same and printed no `uninstalled` line; each one returns from `updateCheckout` before `install.Check` — `evidence/it1-new-tests.txt`, `evidence/it1-scenario-1-2-cli.txt`
- [PASS] review: I judged the full range `da80102..594cacb` (the assigned `git show` names only its first and last commits). Every changed file is in a Files block or a deviation; there is no new dependency and no public signature beyond the Plan; no errors are swallowed on the switch path; names reach git only as full refs or before `--`; the Security block holds. tests: 5 tests added (`cmd/kirokit/commands/sync_test.go`, `desktop/version_test.go`, `desktop/frontend/src/pages/pages.test.js`) · docs: updated README.md (by the build), no doc impact from the added tests
- [NOTE] review: A switch that finishes before its start call returns would leave the button stuck on "Switching…". `main.js` sets `view.switching` only after `await api.switchRef`, while `startClone` marks itself busy before its call. With a real sync, which fetches first, this never happened in any run

**Pictures:** Playwright captures of the running app (Wails dev, throwaway HOME): `evidence/it1-scenario-4-step-2.png`, `evidence/it1-scenario-4-step-3.png`, `evidence/it1-scenario-3-step-3.png`, `evidence/it1-scenario-3-step-4.png`, `evidence/it1-scenario-3-step-5.png`, `evidence/it1-scenario-3-step-6.png`, `evidence/it1-scenario-3-probe-refused.png`, `evidence/it1-done-5-settings-pinned.png`, `evidence/it1-done-5-install-pinned.png`. They match the Branch picker artifact `evidence/it1-artifact-branch-picker.png` in layout and wording. `qualitylayer evidence shot` could not capture these states: it refuses while the tree has uncommitted source, and with my tests backed out its browser rendered the page without the app's backend data. Its one capture showed an empty Settings form, and waiting for a loaded team card timed out.

**Try it:**

1. Build it: `task build`
2. In a clone of a config repository with a `feature` branch, run `./kirokit sync <team> --repo <clone> --branch feature`; look for `checked out feature at <commit>` and the `synced` line
3. Run `./kirokit sync <team> --repo <clone> --branch nope`; look for `kirokit sync: --branch nope: no branch or tag of that name on origin` and exit 1
4. Run `task desktop:dev`; in Settings → Templates pick `feature`, press "Switch to feature and update", and look for `▸ kirokit sync <team> --branch feature` in the kirokit tab and "In use: branch feature … not a release"
5. On Install and sync press Sync; the tab shows `--branch feature`, not `--latest-tag`

**Result:** fail

### Iteration 2

**Revision:** b2ecb5039bf205c040cdddb9a5744f2b1fcec014abfe88a65f0abc98e53b5067

**Assignment:** caa7c1d1-d81c-42f5-9631-212fb1fbc48e
**Judge:** fresh helper

- [PASS] check: reused receipt — all · 7ae1b97 · `go test ./...` · exit 0 · 7.7s
- [PASS] check: reused receipt — all · 7ae1b97 · `task lint` · exit 0 · 0.8s
- [PASS] check: reused receipt — all · 7ae1b97 · `go vet ./...` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · 7ae1b97 · `task ci` · exit 0 · 8.6s
- [PASS] check: reused receipt — all · 7ae1b97 · `task desktop:test` · exit 0 · 8.7s
- [PASS] check: reused receipt — all · 7ae1b97 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- [PASS] done 1: rerun on the CLI built at the judged commit: `kirokit sync demo --branch feature` fetched, put HEAD on `feature` at `origin/feature` (91dab21), installed demo and printed `checked out feature at 91dab21`; `--branch v1.0.0` detached at the tag — `evidence/it2-scenario-1-2-cli.txt`
- [PASS] done 2: after a new commit on `origin/feature` a plain sync printed `updated 91dab21 -> f16a565` and `checked out feature at f16a565`; on `v1.0.0` a plain sync stayed at 5fda641; `--latest-tag` went to `v1.1.0` — `evidence/it2-scenario-1-2-cli.txt`
- [PASS] done 3: in the running app, the first-launch page (after the clone, Newest release preselected) and Settings list Newest release, then Branches, then Tags; each switch ran in the kirokit tab, and Teams and Agents went to 2 with the `feature` branch's team `beta` — `evidence/it2-scenario-4-step-2.png`, `evidence/it2-scenario-3-step-3.png`, `evidence/it2-scenario-3-step-4-teams.png`
- [PASS] done 4: after a reload with `feature` saved, Update ran `kirokit sync demo --branch feature`; Install and sync's Sync ran `--branch feature`, later `--branch v1.0.0`; only picking Newest release ran `--latest-tag` — `evidence/it2-s3c-app.txt`, `evidence/it2-s3b-app.txt`, `evidence/it2-done-4-update-keeps-branch.png`
- [PASS] done 5: Settings' In use line and the Install source line read `branch feature at 2f29afd` with "not a release", `tag v1.0.0 at 819bf3b` with "pinned · does not follow new releases", and "newest release" `v1.1.0 at 432c1e1` — `evidence/it2-done-5-settings-pinned.png`, `evidence/it2-done-5-install-pinned.png`, `evidence/it2-scenario-3-step-3.png`
- [PASS] done 6: in the CLI, local changes, local commits and an unknown name were refused with the planned messages, and HEAD, the edit and the install records stayed the same; in the app, a tracked edit refused the switch to `v1.0.0` with the CLI's message under the list (`aria-invalid`), and HEAD and the saved choice stayed the same after a reload — `evidence/it2-scenario-1-2-cli.txt`, `evidence/it2-s3-app.txt`, `evidence/it2-scenario-3-probe-refused.png`
- [PASS] done 7: README *Sync* has the `--branch` example and line 313 `kirokit sync <team> --branch <name>`, plain-sync behaviour on a branch and a tag, all five refusals word for word and the branch-over-tag rule; *Desktop app* has the Version list, Update/Sync/Continue on the chosen version and the source line — `evidence/it2-scenario-5-readme.txt`
- [PASS] scenario 1: rerun, all 5 steps as expected; probes: an empty, a blank and a spaced name are refused, a missing `--repo` folder fails clearly, an untracked file is kept, and a repeated switch is a no-op — `evidence/it2-scenario-1-2-cli.txt`
- [PASS] scenario 2: rerun, all 5 steps as expected, including `--branch=--upload-pack=x`; a local branch with commits not on origin is refused and HEAD stays on `main` — `evidence/it2-scenario-1-2-cli.txt`
- [PASS] scenario 3: all 6 steps as expected in the running app (Wails dev, throwaway HOME, driven with Playwright). The first driver run counted step 5's header after it had scrolled away; the rerun with a cleared tab shows Sync running `kirokit sync demo --branch feature`. Probes: a refused switch keeps Newest release after a reload, and a `SwitchRef` answer held until after `ref:done` leaves the button on "Update" and enabled (the old code left it on "Switching…") — `evidence/it2-s3-app.txt`, `evidence/it2-s3b-app.txt`, `evidence/it2-race-app.txt`
- [PASS] scenario 4: all 3 steps as expected: the clone ran in the kirokit tab, the same page offered Version with Newest release and Continue, `Switch to feature and continue` ran `kirokit sync --branch feature`, the pages appeared, and `settings.json` holds `"ref": "feature"` — `evidence/it2-s4-app.txt`
- [PASS] scenario 5: step 1 now counts 1 (README line 313); step 2 holds for both sections — `evidence/it2-scenario-5-readme.txt`
- [PASS] dod T1: carried from iteration 1; the change touches no file of the name rule (`internal/checkout`, `setup.go`), and `go test ./...` passes at the judged commit
- [PASS] dod T2: carried for `sync.go`, which the change does not touch; the new local-commits case in `TestSync_FailureLeavesTheInstalledTeamInPlace` and `TestSync_BranchWithASlashKeepsItsWholeName` pass in the receipt; the refusals were rerun live
- [PASS] dod T3: carried for `sync.go`; the new `TestSync_PlainSyncWithATeamOnATagReinstallsFromIt` pins a plain sync with a team on a tag; live, a plain sync on `v1.0.0` printed `checked out` and then `synced`
- [PASS] dod T4: verify grep exits 0; the Sync section has the example and every T2 refusal word for word
- [PASS] dod T5: carried from iteration 1; `internal/checkout` is untouched by the change
- [PASS] dod T6: carried from iteration 1; `settings.go` and `clone.go` are untouched; live, `settings.json` kept `ref` across reloads
- [PASS] dod T7: `version.go` is untouched; the new `TestSwitchRef_WithoutAnInstalledTeamSwitchesTheCheckoutOnly` pins the team-less command and passes in the receipt; live, the first-launch switch ran `kirokit sync --branch feature`
- [PASS] dod T8: live, Update and Sync sent `--branch feature` or `--branch v1.0.0` with a saved ref, and `--latest-tag` only for Newest release
- [PASS] dod T9: live, Update reads "Update" or "Switch to <name> and update"; a refused switch shows under the select with `aria-invalid` and keeps the choice; with the fix, a late `SwitchRef` answer leaves the button on "Update" and enabled
- [PASS] dod T10: live, Newest release is preselected with Continue after the clone; `Switch to feature and continue` switched and closed the page on `ref:done` ok
- [PASS] dod T11: verify grep exits 0; the desktop section names the Version list, the chosen version, new branches after the next Update and the source line
- [PASS] truth oracle: `sync demo --branch feature` left HEAD on `feature` at `origin/feature` with demo installed; a plain sync moved to the new commit; in the app, with `feature` chosen in Settings, Install's Sync ran `kirokit sync demo --branch feature`, not `--latest-tag` — `evidence/it2-scenario-1-2-cli.txt`, `evidence/it2-s3b-app.txt`
- [PASS] truth 1: every `--branch` refusal (local changes, local commits, unknown name) left HEAD and the install records the same and printed no `uninstalled` line; `sync.go` has not changed since iteration 1 and returns before `install.Check` — `evidence/it2-scenario-1-2-cli.txt`
- [PASS] review: I judged `594cacb..7ae1b97` in full and where it meets the rest. `main.js` now marks the switch before the call and clears the mark when the call is refused. The README line makes scenario 5 hold. The added tests reject plausible wrong implementations: skipping the install on a detached HEAD, cutting a slashed branch name, passing an empty team argument, and Update falling back to `--latest-tag`. Every changed file is in a Files block, there is no new dependency or signature, and the Security block still holds. tests: 0 tests added (the one changed behaviour, `switchVersion` in `main.js`, has no test harness: `main.js` loads xterm and the DOM on import) · docs: no doc impact (README updated by the build)
- [NOTE] review: Two presses at the same instant leave a stale error after a successful switch. A Playwright `dblclick` on Switch ran the switch once and left "kirokit is still running a command" under the list, while two clicks 60 ms apart did not. Because the fix clears `view.switching` when the second call is refused, the first switch's `ref:done` would also no longer close the first-launch page in that case (found by reading the code, not run) — `evidence/it2-race-app.txt`, `evidence/it2-probe-double-press.png`
- [NOTE] review: The race fix has no automated test. Only the live probe, which holds the `SwitchRef` answer until after `ref:done`, shows the button returning to "Update" — `evidence/it2-race-app.txt`, `evidence/it2-probe-late-answer.png`

**Pictures:** Playwright captures of the running app (Wails dev, throwaway HOME, light theme, 1180 px): `evidence/it2-scenario-4-step-2.png`, `evidence/it2-scenario-4-step-3.png`, `evidence/it2-scenario-3-step-3.png`, `evidence/it2-scenario-3-step-4.png`, `evidence/it2-scenario-3-step-5.png`, `evidence/it2-scenario-3-step-6.png`, `evidence/it2-scenario-3-probe-refused.png`, `evidence/it2-done-5-settings-pinned.png`, `evidence/it2-done-5-install-pinned.png`, `evidence/it2-probe-late-answer.png`, `evidence/it2-probe-double-press.png`. The change does not alter layout or wording, so iteration 1's comparison with the Branch picker artifact (`evidence/it1-artifact-branch-picker.png`) still applies. `qualitylayer evidence shot` reported `unavailable` again for `branch-picker#settings-version` (waiting for `#version-source`) and `branch-picker#install-source` (waiting for `#page-install p.source`): its browser opens the page without the app's backend data.

**Try it:**

1. Build it: `task build`
2. In a clone of a config repository with a `feature` branch, run `./kirokit sync <team> --repo <clone> --branch feature`; look for `checked out feature at <commit>` and the `synced` line
3. Run `./kirokit sync <team> --repo <clone> --branch nope`; look for `kirokit sync: --branch nope: no branch or tag of that name on origin` and exit 1
4. Run `task desktop:dev`; in Settings → Templates pick `feature`, press "Switch to feature and update", and look for `▸ kirokit sync <team> --branch feature` in the kirokit tab, "In use: branch feature … not a release", and the button back on "Update"
5. On Install and sync press Sync; the tab shows `--branch feature`, not `--latest-tag`
6. Run `grep -c -e 'sync <team> --branch <name>' README.md`; look for 1

**Result:** pass
