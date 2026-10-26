# Check out a config repo branch · what was checked

## Summary

<!-- ql:summary -->
Independent review is running.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
Nothing is open.
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
- **Independent review** · Claude Code · Opus 5.5 · never saw the build
  - State: running
  - Result: not yet
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
- **Fix rounds** · the build session
  - State: waiting
  - Result: not yet
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
Not checked yet.
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
None so far.
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `go test ./...` · exit 0 · 7.9s
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `task lint` · exit 0 · 0.7s
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `go vet ./...` · exit 0 · 0.2s
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `task ci` · exit 0 · 8.1s
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `task desktop:test` · exit 0 · 8.9s
- all · bf3c8e3152cd83f444dcb7c96d451ac2a14fb56db1b9b5142000320db7ff3835 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
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
