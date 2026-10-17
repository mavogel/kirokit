# Desktop app from design · what was checked

## Summary

<!-- ql:summary -->
Checking has started with the quality pass.
<!-- /ql:summary -->

## How it is checked

<!-- ql:lanes -->
- **Polish** · a fresh agent · Sonnet 5.5
  - State: running
  - Result: done · 1 simplification (uncommitted, see below) · 3 tests added (uncommitted) · docs: AGENTS.md · UI: not captured · 1 screens compared: 0 captured, 0 stale, 1 not captured
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: other processes and two person-entered inputs. The app starts the person's login shell, git and the bundled kirokit..
- **Second opinion** · Codex · GPT-6.1 Sol · another vendor
  - State: fail
  - Result: did not finish
  - Why it runs here: Runs because the change crosses a trust boundary (other processes and two person-entered inputs. The app starts the person's login shell, git and the bundled kirokit.), the Plan says it is risky: new dependencies the app ships (Wails, creack/pty, xterm.js); a changed `kirokit install` and `kirokit sync`, which breaks anyone who syncs two teams on one machine; and a release job that pushes to the Homebrew tap and slice 1, 2, 4 and 7 is marked risky.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: 8 checks pass · 8 reused
- **Independent review** · Claude Code · Opus 5.5 · never saw the build
  - State: running
  - Result: not yet
- **Fix rounds** · one fresh fixer per round
  - State: waiting
  - Result: not yet
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
Not checked yet.
<!-- /ql:proof -->

## Open items

<!-- ql:open-items -->
- **Only you can confirm:** The cask lands in your tap.
- **Only you can confirm:** A real tagged release publishes it.
- **Only you can confirm:** It looks like the approved design.
<!-- /ql:open-items -->

## Quality pass

### Polish

**Result:** done · 1 simplification (uncommitted, see below) · 3 tests added (uncommitted) · docs: AGENTS.md · UI: not captured
- not captured: desktop-app · unavailable: the app is a native Wails window with no loopback URL, and the brief named no surface keys; nothing was compared against the mockup this pass
**Left alone:** the commit steps. The caller said not to commit, so the simplification (`sync.go` now calls `externalskills.WithoutRepoGitEnv`, exported from `fetch.go`, instead of a byte-identical copy), the tests (`TestInstalled_CorruptRecordIsAnErrorNamingTheFile`, `TestInstalled_IgnoresEntriesThatAreNotRecords`, `TestShell_ResizeOutsideTheLimitsStartsNothing`) and the AGENTS.md line on `KIROKIT_REPO` sit in the working tree. README already covers the change.
**Not checked:** screens against the mockup (see above).

### Security

**Result:** done · nothing severe in the change
**Boundaries:** person-entered repository address and Kiro home (Settings, Clone) to git and the file system; the app to the login shell, git and the bundled kirokit (pseudo-terminal child processes); repository content (teams, agent descriptions) to the page; the release workflow and cask.
**Not checked:** the full frontend page code for every escaped field (spot-checked: views build HTML through `esc`), and the third-party Wails and pty libraries.

## Checks

<!-- ql:checks -->
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `go test ./...` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `go vet ./...` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `golangci-lint run` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `go build ./...` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `cd desktop/frontend && npm ci && npm run build` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `cd desktop && go vet ./... && go test ./...` · exit 0 · 0s · reused
- all · 01fddbad4298195c79863f685b36476eecec6115782d2f06befdd5109709611f · `task desktop:build` · exit 0 · 0s · reused
<!-- /ql:checks -->

## Fix rounds

<!-- ql:fix-rounds -->
None so far.
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
**The second opinion on the change could not run:** Codex stopped with an error (tried twice).
<!-- /ql:reviews -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification
