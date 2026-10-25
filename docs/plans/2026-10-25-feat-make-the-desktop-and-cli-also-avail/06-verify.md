# Ship CLI and desktop app for Linux and Windows · what was checked

## Summary

<!-- ql:summary -->
8 of 8 points assessed · 2 passed

Independent review failed; the recorded findings need attention.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Unverified:** Point 1 · Linux users install the kirokit CLI with a package, Homebrew or download · No Linux release exists to install from. environment: the branch is still not on GitHub, no CI or release run exists and no `GORELEASER_KEY` is set here, so no `.deb`, `.rpm`, tar.gz or Linux cask was produced; nothing in this round touches the release config,
- **Unverified:** Point 2 · Windows users install the kirokit CLI with Scoop or a zip and run it in PowerShell · No Windows release or Scoop manifest exists to install from. environment: same blocker; there is no Windows host or runner
- **Unverified:** Point 3 · The CLI installs a team on Linux and Windows that Kiro CLI can run · The Windows half has still never run. environment: in the Linux container `kirokit install` plus the real `sops exec-env` launch passes again, as on macOS; no Windows runner has run it,
- **Unverified:** Point 4 · Linux users install the desktop app from a .deb or .rpm · No Linux app package was built. environment: the desktop tests pass again in the Linux container; the `.deb`/`.rpm` need the Pro key's split snapshot, and amd64 needs the runner,
- **Unverified:** Point 5 · Windows users install the desktop app with a setup .exe · The Windows app has never started on Windows. environment: there is no Windows runner; the setup programs were only built in the container in iteration 1,
- **Unverified:** Point 7 · Every release builds and checks all three platforms automatically · No pull-request or release run exists on any runner. environment: the branch is not pushed
- **Unverified:** The release dry-run input still does nothing. Carried from iteration 1: `release.yaml` declares `dry_run` for `workflow_dispatch` but never reads it; this round does not touch it
- **Unverified:** A symlinked folder is backed up only as a link, so the backup keeps no copy of the old files. Install then writes through the link into the folder it points at, and the backup link points at the overwritten files. `main` did exactly the same, so Mac users notice nothing, but the backup protects nothing there
- **Need:** T10 · Push this task's branch and open a pull request, so CI and Desktop run on the windows-latest, ubuntu-latest and macos-latest runners · queued for Review
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 6 checks
- **Independent review** · fresh helper
  - State: fail
  - Result: failed · 6 of 8 points
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers, the Plan says it is risky: yes, The shipped app loads a new dependency (a Windows ConPTY library), secrets cross and slice 3 and 6 is marked risky; say "get a second opinion" to have one.
- **Fix rounds** · the build session
  - State: waiting
  - Result: not yet
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
| Done means | Proof | |
|---|---|---|
| 1. Linux users install the kirokit CLI with a package, Homebrew or download | No Linux release exists to install from. environment: the branch is still not on GitHub, no CI or release run exists and no `GORELEASER_KEY` is set here, so no `.deb`, `.rpm`, tar.gz or Linux cask was produced; nothing in this round touches the release config,. | missing |
| 2. Windows users install the kirokit CLI with Scoop or a zip and run it in PowerShell | No Windows release or Scoop manifest exists to install from. environment: same blocker; there is no Windows host or runner. | missing |
| 3. The CLI installs a team on Linux and Windows that Kiro CLI can run | The Windows half has still never run. environment: in the Linux container `kirokit install` plus the real `sops exec-env` launch passes again, as on macOS; no Windows runner has run it,. | missing |
| 4. Linux users install the desktop app from a .deb or .rpm | No Linux app package was built. environment: the desktop tests pass again in the Linux container; the `.deb`/`.rpm` need the Pro key's split snapshot, and amd64 needs the runner,. | missing |
| 5. Windows users install the desktop app with a setup .exe | The Windows app has never started on Windows. environment: there is no Windows runner; the setup programs were only built in the container in iteration 1,. | missing |
| 6. The Mac release keeps working exactly as today | D1 is fixed. With a symlinked agents folder, a fresh home or an existing agents folder, `kirokit install`'s output and every installed file and link match base byte for byte. The rebuilt `.app` bundles the universal CLI, which installs through a symlinked folder, and smoke-starts with exit 0. No darwin `.goreleaser.yaml` line or `desktop/build/darwin` asset changed. The dock words are carried from iteration 1, since the fix never reaches the frontend,,. | passed |
| 7. Every release builds and checks all three platforms automatically | No pull-request or release run exists on any runner. environment: the branch is not pushed. | missing |
| 8. The README tells Linux and Windows users how to install and what to expect | Carried from iteration 1; this round changes only `internal/backup` and test files, never `README.md`. | passed |
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
- Round 1 · 19 points still open
- Round 2 · 16 points still open
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers, the Plan says it is risky: yes, The shipped app loads a new dependency (a Windows ConPTY library), secrets cross and slice 3 and 6 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go test ./...` · exit 0 · 0.3s
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `golangci-lint run` · exit 0 · 0.5s
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go vet ./...` · exit 0 · 0.2s
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go build ./...` · exit 0 · 0.3s
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `task desktop:test` · exit 0 · 9s
- all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->

### Security

**Result:** done · nothing severe in the change
**Boundaries:** decrypted SOPS secrets to MCP servers via `sops exec-env` through cmd.exe (CLI to child process); vendored external skills fetched from git and verified by lock hash (network to files); Windows ConPTY and junction/taskkill child processes (app to OS); release and CI workflows (secrets, tokens, downloaded sops)
- S1 low · internal/sopswrap/sopswrap.go:66 · cmd.exe launch refuses whitespace, quotes and `& | < > ^ % ! ( )`, so template values cannot inject commands; secrets stay in the encrypted file and are only referenced as `${VAR}` · listed, no action
- Checked and clear: dirlink junction creation refuses `"` in paths and uses argument-vector exec; backup keeps links unfollowed; externalskills rejects git symlinks (mode 120000) and keeps exec bits in the hash; ConPTY command line uses `syscall.EscapeArg`; CI sops download is pinned and checksum-verified; release actions are SHA-pinned and tag values come from svu; the committed age key is test-only and allowlisted for one path.

<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification

### Iteration 1

**Revision:** 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc

**Assignment:** 88cb4311-4210-46dc-88e5-a822c5647d0e
**Judge:** fresh helper

Executed tree: the judged commit (HEAD did not move) plus this review's two unstaged test additions. The diff was read as the whole branch (base `main` to HEAD), which also covers the four commits the assignment's `git show` list leaves out (the slice 4 desktop commit, `kirokit --version`, the Docker arch commit and the first slice 5 commit).

- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `go test ./...` · exit 0 · 0.5s
- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `golangci-lint run` · exit 0 · 0.7s
- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `go vet ./...` · exit 0 · 0.1s
- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `go build ./...` · exit 0 · 0.6s
- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `task desktop:test` · exit 0 · 2.6s
- [PASS] check: reused receipt — all · 5adcd4b329aee1160aad2463a6d12c726840b2ab8062dc423caea76347b10acc · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0s · reused
- [MISSING] done 1: No Linux release exists to install from. environment: the branch is not on GitHub and no `GORELEASER_KEY` is set here, so no `.deb`, `.rpm`, tar.gz or Linux cask was produced; the nfpm and archive config passes `goreleaser check` and `kirokit --version` prints the version — `evidence/it1-blocker-ci.txt`, `evidence/it1-static-cross-os.txt`
- [MISSING] done 2: No Windows release or Scoop manifest exists to install from. environment: same blocker; the windows/amd64 and arm64 CLI builds and the zip override were only checked statically — `evidence/it1-blocker-ci.txt`
- [MISSING] done 3: The Windows half was never run. environment: in the Linux container `kirokit install` plus the real `sops exec-env` launch passes (as on macOS); no Windows runner has run it — `evidence/it1-scenario-1-linux.txt`, `evidence/it1-oracle-macos.txt`
- [MISSING] done 4: No Linux app package was built. environment: in the arm64 container the app builds, links `libwebkit2gtk-4.1` and `libgtk-3`, smoke-starts under `xvfb-run` with exit 0, and the desktop tests pass; the `.deb`/`.rpm` need the Pro key's split snapshot, and amd64 needs the runner — `evidence/it1-t20-linux-windows-build.txt`, `evidence/it1-scenario-3-linux-desktop-test.txt`
- [MISSING] done 5: The Windows app never started on Windows. environment: both setup programs build and list `kirokit-desktop.exe` and `bin/kirokit.exe` (`7z l`); the start, the desktop tests and the PowerShell tab need the Windows runner — `evidence/it1-t20-linux-windows-build.txt`
- [FAIL] done 6: Some Mac users now see install fail. With a symlinked folder in the Kiro home (D1), `kirokit install` stops with a backup error where it used to succeed; otherwise the `.app` layout, bundled universal CLI, smoke start, dock words and install output match today — `evidence/it1-d1-backup-symlink.txt`, `evidence/it1-scenario-2-steps-4-5.txt`
- [MISSING] done 7: No pull-request or release run exists on any runner. environment: CI, Desktop and Release workflows pass `actionlint`, but nothing has run them — `evidence/it1-blocker-ci.txt`
- [PASS] done 8: README names every install command and file per OS, the age-key path per OS, SmartScreen's More info then Run anyway, and keeps the quarantine note — `evidence/it1-scenario-5.txt`
- [MISSING] scenario 1: The CI steps could not run. environment: steps 1-2 need a CI run that does not exist; step 3 passes on macOS and step 4 passes in the Linux container; probe: a `KIRO_HOME` with a space installs, and without sops the test skips, or fails under `KIROKIT_REQUIRE_SOPS=1` — `evidence/it1-oracle-macos.txt`, `evidence/it1-scenario-1-linux.txt`
- [MISSING] scenario 2: Step 7 could not be shown here. environment: the app always adds `/opt/homebrew/bin`, where sops is installed on this Mac, so the hint never appears; the darwin hint string is unchanged in source and its test passes. Steps 1-6 pass: `platform-team: ok`, only the two devops golden lines change, `desktop/build/darwin` untouched, the bundled universal CLI is in place, smoke start exits 0, the dock reads `zsh` and the first line ends `in the zsh tab.`; probe past the script found D1 — `evidence/it1-scenario-2-steps-1-3.txt`, `evidence/it1-scenario-2-steps-4-5.txt`, `evidence/it1-scenario-2-smoke.txt`, `evidence/it1-scenario-2-steps-6-7.txt`
- [MISSING] scenario 3: The Desktop workflow has never run. environment: steps 1-5 need its log; step 6 (`task desktop:test` in the Linux container) passes — `evidence/it1-scenario-3-linux-desktop-test.txt`
- [MISSING] scenario 4: The release dry run could not run. environment: it needs the pushed branch and the GoReleaser Pro key — `evidence/it1-blocker-ci.txt`
- [PASS] scenario 5: all three greps show the Linux and Windows install lines, SmartScreen steps, the quarantine note and the age-key paths — `evidence/it1-scenario-5.txt`
- [MISSING] dod T1: The cmd.exe round trip never ran on Windows. environment: the join, refusal and POSIX bullets pass; `sopswrap_windows_test.go` only compiles here — `evidence/it1-cards.txt`
- [PASS] dod T2: Cmd join, POSIX identity through `RenderTeam`, the named refusal with no output folder — `evidence/it1-cards.txt`
- [PASS] dod T3: one header line changes in each devops golden and nothing else; every team renders for Cmd — `evidence/it1-scenario-2-steps-1-3.txt`
- [MISSING] dod T4: The Windows bullets never ran. environment: on macOS the installed files are byte-identical to base's (also under a path with a space); the `C:/` path and refuse-before-writing tests skip off Windows — `evidence/it1-truth-macos-install.txt`
- [MISSING] dod T5: The Oracle test never ran on Windows. environment: argv, decrypted secret, skip/require and wrong-key bullets hold on macOS and Linux; gitleaks finds the key without `.gitleaks.toml` and nothing with it — `evidence/it1-oracle-macos.txt`, `evidence/it1-cards.txt`
- [PASS] dod T6: `text: unset` for both files, no CRLF index entry; the verify command passes — `evidence/it1-cards.txt`
- [PASS] dod T7: executables come sorted from git's modes and the Windows verification and fetch-again message are tested through `goos`; this review adds the collection case — `evidence/it1-review-tests.txt`
- [MISSING] dod T8: Install no longer behaves as today on macOS and Linux. A Kiro-home folder that is a symlink now fails the backup (D1); the Windows junction bullets never ran — `evidence/it1-d1-backup-symlink.txt`
- [PASS] dod T9: drive paths accepted, `C:`, `C:repo`, `-C:\x`, `C:\a b` refused, earlier addresses kept; `kirokit --version` prints the version — `evidence/it1-cards.txt`
- [PASS] dod T10: three-OS matrix, `fail-fast: false`, pinned checksum-verified sops; `actionlint` and the container suite pass — `evidence/it1-scenario-1-linux.txt`
- [MISSING] dod T11: Every Windows bullet is a Windows-only test that never ran. environment: the Unix file differs from `terminal.go` only by the build tag — `evidence/it1-cards.txt`
- [PASS] dod T12: the desktop module vets and test-compiles for windows/amd64 and arm64; the Unix tests pass — `evidence/it1-static-cross-os.txt`
- [PASS] dod T13: data folder, age-key and hint tables pass, macOS values unchanged — `evidence/it1-cards.txt`
- [PASS] dod T14: tool folders, default shell, PATHEXT lookup and the Path merge pass — `evidence/it1-cards.txt`
- [PASS] dod T15: CLI candidates, missing-CLI words, PowerShell command and `State().Env.Shell` pass — `evidence/it1-cards.txt`
- [PASS] dod T16: node tests pass; the running dock on macOS reads `zsh` with the zsh first line — `evidence/it1-scenario-2-steps-6-7.txt`
- [PASS] dod T17: drive roots refused like `/`, drive addresses accepted, the bad forms refused — `evidence/it1-cards.txt`
- [PASS] dod T18: `task desktop:test` (receipt), the Linux container run and Windows vet pass — `evidence/it1-scenario-3-linux-desktop-test.txt`
- [PASS] dod T19: the smoke start exits 0 and writes no data folder; without the variable the app keeps running — `evidence/it1-scenario-2-smoke.txt`
- [PASS] dod T20: Linux app links WebKitGTK 4.1 beside its CLI; both Windows setups contain `bin/kirokit.exe` — `evidence/it1-t20-linux-windows-build.txt`
- [PASS] dod T21: `actionlint` passes; the smoke step fails on non-zero, runs under `xvfb-run -a`, and the `.deb` step needs the key — `evidence/it1-static-cross-os.txt`
- [PASS] dod T22: `goreleaser check` passes, no darwin line removed, Scoop upload skipped without its token — `evidence/it1-static-cross-os.txt`
- [PASS] dod T23: `actionlint` passes; only `merge` pushes the tag; artifact names carry the run id — `evidence/it1-static-cross-os.txt`
- [PASS] dod T24: README, CONTRIBUTING and AGENTS cover the release layout, the Windows argument rule, `executables` and `scripts/linux-docker.sh` — `evidence/it1-scenario-5.txt`
- [MISSING] truth oracle: The oracle never ran on its Windows target. environment: on macOS and in the Linux container the installed server gets every argument and `Bearer s3cret value`, and a mutation that drops POSIX quoting makes it fail — `evidence/it1-oracle-macos.txt`, `evidence/it1-scenario-1-linux.txt`
- [FAIL] truth 1: macOS install is no longer the same as today for a symlinked Kiro folder (D1). Otherwise the installed files are byte-identical to base's and only the two devops golden lines change — `evidence/it1-truth-macos-install.txt`, `evidence/it1-d1-backup-symlink.txt`
- [FAIL] review: D1 · Install fails when a folder in the Kiro home is a symbolic link. `internal/backup/backup.go` now guards the link case with `p != dir`, so `CopyAside` on a folder that is itself a symlink (a dotfiles-managed `~/.kiro/agents`) falls to "is neither a file, a folder nor a symbolic link" and `kirokit install` exits 1 on macOS and Linux, where base made a link backup and installed. Test `TestCopyAside_AFolderThatIsASymlinkIsBackedUpAsTheLink` passes on base and fails on this change; kept out of the checkout — `evidence/it1-d1-backup-symlink.txt`
- [NOTE] review: The release dry-run input does nothing. `release.yaml` declares `dry_run` for `workflow_dispatch` but never reads it, so a dispatch with `dry_run=false` still builds only a snapshot (safe, but the input misleads)
- [PASS] review: apart from D1, no swallowed errors, races or wrong error paths found in sopswrap, render, install, dirlink, externalskills, the ConPTY backend, platform functions or the workflows; every changed file sits in a Files block or a recorded deviation; stack drift only `golang.org/x/sys` indirect to direct (recorded T11 deviation); the Security block's boundaries hold and D1 is not a security issue; the Oracle rejects an unquoted POSIX join (it then sees `Authorization:Bearer s3cret value` split); tests: 2 tests added (`internal/sopswrap/sopswrap_test.go`: a secrets path with spaces stays sops's own argument, rejecting a check of every token; `internal/externalskills/fetch_test.go`: a fetched collection verifies on Windows with execute bits stripped, rejecting executables recorded with the wrong prefix) · docs: no doc impact — `evidence/it1-review-tests.txt`

**Pictures:** `evidence/it1-scenario-2-step-6-dock-kirokit.png`, `evidence/it1-scenario-2-step-6-dock-shell.png` (macOS dev app in headless Chrome: dock tabs `kirokit` and `zsh`, and the zsh first line)

**Try it:**

1. Build the CLI: `go build -o kirokit ./cmd/kirokit && ./kirokit --version`; look for `kirokit <version>`.
2. Oracle on this machine: `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v`; look for `--- PASS`.
3. D1: `mkdir -p /tmp/k/real /tmp/k/kiro && echo '{}' > /tmp/k/real/devops.json && ln -s /tmp/k/real /tmp/k/kiro/agents && KIRO_HOME=/tmp/k/kiro ./kirokit install platform-team`; today it fails with "is neither a file, a folder nor a symbolic link", and after the fix it should install.
4. Mac app: `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?`; look for `0`.
5. Linux and Windows builds on a Mac: `scripts/linux-docker.sh task desktop:test`, `scripts/linux-docker.sh task desktop:build`, `scripts/linux-docker.sh env GOOS=windows task desktop:build`; look for passing tests, `desktop/build/bin/kirokit-desktop`, and two `*-setup.exe` files.
6. Push the branch and open the pull request; CI and Desktop must pass on `ubuntu-latest`, `windows-latest` and `macos-latest`, then run `gh workflow run release.yaml --ref feat/feat-make-desktop-cli-also` for the dry run.

**Result:** fail

### Iteration 2

**Assignment:** 66d68a6f-ddb8-4a7a-97d6-d66755c8a2ab
**Judge:** fresh helper

Executed tree: the judged commit (HEAD did not move) plus this review's three unstaged test additions (`internal/backup/backup_test.go`, `internal/install/install_test.go`). The fix round changed `internal/backup/backup.go` and committed iteration 1's review tests. This iteration reran everything that change reaches, every verdict iteration 1 did not pass, and the oracle. Verdicts the change cannot reach are carried, each with its reason.

- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go test ./...` · exit 0 · 0.3s
- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `golangci-lint run` · exit 0 · 0.5s
- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go vet ./...` · exit 0 · 0.2s
- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `go build ./...` · exit 0 · 0.3s
- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `task desktop:test` · exit 0 · 9s
- [PASS] check: reused receipt — all · e4ea947de239c32df05f0e6eb79763f33cda342aa32a8e679091bb9ecd36a865 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- [MISSING] done 1: No Linux release exists to install from. environment: the branch is still not on GitHub, no CI or release run exists and no `GORELEASER_KEY` is set here, so no `.deb`, `.rpm`, tar.gz or Linux cask was produced; nothing in this round touches the release config — `evidence/it2-blocker-ci.txt`, `evidence/it1-static-cross-os.txt`
- [MISSING] done 2: No Windows release or Scoop manifest exists to install from. environment: same blocker; there is no Windows host or runner — `evidence/it2-blocker-ci.txt`
- [MISSING] done 3: The Windows half has still never run. environment: in the Linux container `kirokit install` plus the real `sops exec-env` launch passes again, as on macOS; no Windows runner has run it — `evidence/it2-scenario-1-linux.txt`, `evidence/it2-oracle-macos.txt`
- [MISSING] done 4: No Linux app package was built. environment: the desktop tests pass again in the Linux container; the `.deb`/`.rpm` need the Pro key's split snapshot, and amd64 needs the runner — `evidence/it2-scenario-3-linux-desktop-test.txt`, `evidence/it1-t20-linux-windows-build.txt`
- [MISSING] done 5: The Windows app has never started on Windows. environment: there is no Windows runner; the setup programs were only built in the container in iteration 1 — `evidence/it2-blocker-ci.txt`, `evidence/it1-t20-linux-windows-build.txt`
- [PASS] done 6: D1 is fixed. With a symlinked agents folder, a fresh home or an existing agents folder, `kirokit install`'s output and every installed file and link match base byte for byte. The rebuilt `.app` bundles the universal CLI, which installs through a symlinked folder, and smoke-starts with exit 0. No darwin `.goreleaser.yaml` line or `desktop/build/darwin` asset changed. The dock words are carried from iteration 1, since the fix never reaches the frontend — `evidence/it2-truth-macos-install.txt`, `evidence/it2-scenario-2-steps-1-5.txt`, `evidence/it1-scenario-2-steps-6-7.txt`
- [MISSING] done 7: No pull-request or release run exists on any runner. environment: the branch is not pushed — `evidence/it2-blocker-ci.txt`
- [PASS] done 8: carried from iteration 1; this round changes only `internal/backup` and test files, never `README.md` — `evidence/it1-scenario-5.txt`
- [MISSING] scenario 1: The CI steps still cannot run. environment: steps 1-2 need a CI run that does not exist; step 3 passes on macOS and step 4 passes in the Linux container, now including the symlinked-folder tests — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`, `evidence/it2-blocker-ci.txt`
- [MISSING] scenario 2: Step 7 still cannot be shown here. environment: the app always adds `/opt/homebrew/bin`, where sops lives on this Mac, so the hint never appears; the darwin hint string is unchanged. Steps 1-5 pass again: `platform-team: ok`, only the two devops golden lines change, the darwin assets are unchanged and the build script's darwin branch is only re-indented, the bundled CLI is universal (`x86_64 arm64`), and the smoke start exits 0. Step 6 is carried from iteration 1, since the fix never reaches the frontend. Probe past the script: the bundled CLI installs into a home with a symlinked agents folder (exit 0) — `evidence/it2-scenario-2-steps-1-5.txt`, `evidence/it1-scenario-2-steps-6-7.txt`
- [MISSING] scenario 3: The Desktop workflow has never run. environment: steps 1-5 need its log; step 6 (`task desktop:test` in the Linux container) passes again — `evidence/it2-scenario-3-linux-desktop-test.txt`, `evidence/it2-blocker-ci.txt`
- [MISSING] scenario 4: The release dry run cannot run. environment: it needs the pushed branch and the GoReleaser Pro key — `evidence/it2-blocker-ci.txt`
- [PASS] scenario 5: carried from iteration 1; the fix changes only `internal/backup`, which these README greps never read — `evidence/it1-scenario-5.txt`
- [MISSING] dod T1: The cmd.exe round trip has still never run on Windows. environment: the join, refusal, POSIX and spaced-secrets-path tests pass here; `sopswrap_windows_test.go` only compiles — `evidence/it2-cards.txt`
- [PASS] dod T2: carried from iteration 1; render uses `MoveAside`, which this fix does not change — `evidence/it1-cards.txt`
- [PASS] dod T3: `./kirokit test platform-team` ok again; still only the two devops golden lines differ from `main` — `evidence/it2-scenario-2-steps-1-5.txt`
- [MISSING] dod T4: The Windows bullets have still never run. environment: on macOS the installed files match base byte for byte (now also with a symlinked agents folder); the `C:/` and refuse-before-writing tests skip off Windows — `evidence/it2-truth-macos-install.txt`, `evidence/it2-cards.txt`
- [MISSING] dod T5: The Oracle test has still never run on Windows. environment: it passes on macOS and in the Linux container with sops required — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`
- [PASS] dod T6: carried from iteration 1; the fix adds no file and touches no attribute — `evidence/it1-cards.txt`
- [PASS] dod T7: the verify command passes again, including the collection lock that verifies on Windows without execute bits — `evidence/it2-cards.txt`
- [MISSING] dod T8: The Windows junction bullets have still never run. environment: the macOS and Linux bullets now hold (a symlinked folder is backed up as the link and installs as on base, also through the bundled CLI); the junction tests only compile here — `evidence/it2-review-tests.txt`, `evidence/it2-truth-macos-install.txt`, `evidence/it2-cards.txt`
- [PASS] dod T9: carried from iteration 1; `setup.go` and `root.go` are untouched by the fix — `evidence/it1-cards.txt`
- [PASS] dod T10: the container suite with `KIROKIT_REQUIRE_SOPS=1` passes again; `ci.yaml` is unchanged since iteration 1's `actionlint` — `evidence/it2-scenario-1-linux.txt`
- [MISSING] dod T11: Every Windows terminal bullet is a Windows-only test that has still never run. environment: the terminal tests and the windows amd64/arm64 vet pass here — `evidence/it2-cards.txt`
- [PASS] dod T12: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-static-cross-os.txt`
- [PASS] dod T13: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-cards.txt`
- [PASS] dod T14: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-cards.txt`
- [PASS] dod T15: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-cards.txt`
- [PASS] dod T16: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-scenario-2-steps-6-7.txt`
- [PASS] dod T17: carried from iteration 1; the fix changes only `internal/backup`, which this card never runs — `evidence/it1-cards.txt`
- [PASS] dod T18: the `task desktop:test` receipt and the Linux container run pass again — `evidence/it2-scenario-3-linux-desktop-test.txt`
- [PASS] dod T19: the rebuilt app smoke-starts with exit 0; the smoke code is untouched by the fix — `evidence/it2-scenario-2-steps-1-5.txt`
- [PASS] dod T20: the macOS app was rebuilt with the universal CLI inside; the Linux and Windows builds are carried from iteration 1, since the fix touches no build script or asset — `evidence/it2-scenario-2-steps-1-5.txt`, `evidence/it1-t20-linux-windows-build.txt`
- [PASS] dod T21: carried from iteration 1; the fix touches no workflow — `evidence/it1-static-cross-os.txt`
- [PASS] dod T22: no `.goreleaser.yaml` line is removed against `main`; `goreleaser check` is carried from iteration 1, since the config is untouched — `evidence/it2-scenario-2-steps-1-5.txt`, `evidence/it1-static-cross-os.txt`
- [PASS] dod T23: carried from iteration 1; the fix touches no workflow — `evidence/it1-static-cross-os.txt`
- [PASS] dod T24: carried from iteration 1; the fix changes no documented command, flag or field — `evidence/it1-scenario-5.txt`
- [MISSING] truth oracle: The oracle has still never run on its Windows target. environment: on macOS and in the Linux container, with sops required, the installed server gets every argument and `Bearer s3cret value` — `evidence/it2-oracle-macos.txt`, `evidence/it2-scenario-1-linux.txt`
- [PASS] truth 1: macOS install output and every installed file and link match base byte for byte for a fresh home, an existing agents folder and a symlinked agents folder (`KIRO_HOME` with a space); only the two devops golden lines differ — `evidence/it2-truth-macos-install.txt`, `evidence/it2-scenario-2-steps-1-5.txt`
- [NOTE] review: The release dry-run input still does nothing. Carried from iteration 1: `release.yaml` declares `dry_run` for `workflow_dispatch` but never reads it; this round does not touch it
- [NOTE] review: A symlinked folder is backed up only as a link, so the backup keeps no copy of the old files. Install then writes through the link into the folder it points at, and the backup link points at the overwritten files. `main` did exactly the same, so Mac users notice nothing, but the backup protects nothing there — `evidence/it2-truth-macos-install.txt`
- [PASS] review: D1 is fixed at its cause: the `p != dir` guard is gone, so `dir` itself goes through the link case, as on base. The change has no swallowed errors or wrong error paths. The files changed this round sit in T8's Files block plus iteration 1's review tests. There is no stack drift, and no trust boundary is newly crossed. `TestCopyAside_AFolderThatIsASymlinkIsBackedUpAsTheLink` fails on the pre-fix code. tests: 2 tests added (`internal/backup/backup_test.go`: a root-level file that is a symlink is backed up as the link, rejecting a fix that handles only folders; `internal/install/install_test.go`: install into a home whose agents folder is a symlink succeeds, keeps the link, backs it up as the link and writes the agent into the linked folder, rejecting the D1 guard at the user's entry point and any fix that replaces the link with a real folder). Both new tests, and the fix round's own regression test, fail on the pre-fix `backup.go` and pass on the judged commit, on macOS and in the Linux container · docs: no doc impact — `evidence/it2-review-tests.txt`

**Pictures:** none. Only `internal/backup` changed and no screen is reached, so the dock pictures are carried from iteration 1 (`evidence/it1-scenario-2-step-6-dock-kirokit.png`, `evidence/it1-scenario-2-step-6-dock-shell.png`)

**Try it:**

1. Build the CLI: `go build -o kirokit ./cmd/kirokit && ./kirokit --version`; look for `kirokit <version>`.
2. D1 is gone: `mkdir -p /tmp/k/real /tmp/k/kiro && echo '{}' > /tmp/k/real/devops.json && ln -s /tmp/k/real /tmp/k/kiro/agents && KIRO_HOME=/tmp/k/kiro ./kirokit install platform-team; echo $?`; look for `installed platform-team` and `0`, then `ls -l /tmp/k/kiro`, where both `agents` and `*-agents.backup` are links to `/tmp/k/real`.
3. The new tests: `go test ./internal/backup/ ./internal/install/ -run 'Symlink' -v`; look for `--- PASS` on all of them.
4. Oracle on this machine: `KIROKIT_REQUIRE_SOPS=1 go test ./cmd/kirokit/commands/ -run TestInstall_SopsLaunch -v`; look for `--- PASS`.
5. Mac app: `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?`; look for `0`.
6. Push the branch and open the pull request; CI and Desktop must pass on `ubuntu-latest`, `windows-latest` and `macos-latest`. Then run `gh workflow run release.yaml --ref feat/feat-make-desktop-cli-also -f dry_run=true` for the dry run. This is what turns the `environment:` lines into proof.

**Result:** fail
