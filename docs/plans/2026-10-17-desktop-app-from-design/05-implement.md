# Desktop app from design · the build

## Summary

<!-- ql:summary -->
5 of 24 tasks done · 1 of 7 slices finished · started 18 Oct, 00:00 · 16 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | One team stays installed and kirokit runs from any folder | `KIROKIT_REPO` roots every repository path, `install` and `sync` uninstall any other installed team first, and `sync` without a team only updates the checkout. | done |
| 2 | The app opens with the design's shell and remembers its settings | A Wails window shows the logo, the sidebar with five sections in light and dark mode, the collapsing icon rail, and a Settings page that saves. | next |
| 3 | The pages show the repository's real data | Teams, Agents, Install and Skills and MCP versions read the app's own copy of the repository, and the screens redraw when install records change. | after 2 |
| 4 | A real terminal docks at the bottom | The dock opens, closes and resizes like VS Code's and holds two tabs: your login shell and the kirokit tab for button commands. | after 3 |
| 5 | A person installs a team with buttons only | First launch asks for the repository's address and clones it, the setup dialog checks the prerequisites, and Install, Sync, Uninstall and Update run their `kirokit` commands in the terminal. | after 4 |
| 6 | One command builds the macOS app and CI builds it too | `task desktop:build` produces a universal `kirokit.app` with the CLI inside, and a new workflow runs the app's tests and build on macOS. | after 5 |
| 7 | Each release attaches the app and a Homebrew cask | A tagged release also builds the macOS app, attaches `kirokit-desktop_<version>_darwin_universal.zip` and pushes the `kirokit-desktop` cask. | after 6 |
<!-- /ql:slices -->

## Checks

- T1 · 445c884 · `go test ./cmd/kirokit/commands/ ./internal/golden/` · exit 0 · 2.3s
- T1 · 445c884 · `go test ./cmd/kirokit/commands/ ./internal/golden/` · reused: nothing changed since they passed
- T2 · 445c884 · `go test ./internal/install/` · exit 0 · 0.6s
- T3 · 445c884 · `go test ./cmd/kirokit/commands/ -run Install` · exit 0 · 0.9s
- T4 · 445c884 · `go test ./cmd/kirokit/commands/ -run Sync` · exit 0 · 1.3s
- T5 · 445c884 · `go test ./...` · exit 0 · 8.8s
- slice 1 · 445c884 · `go build ./...` · exit 0 · 0.4s
- slice 1 · 445c884 · `go vet ./...` · exit 0 · 0.1s
- slice 1 · 445c884 · `go test ./...` · reused: nothing changed since they passed
- slice 1 · 445c884 · `golangci-lint run` · exit 0 · 0.5s
- [PASS] scenario 1 step 1 (QualityLayer) · 445c884 · `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` · exit 0 · as expected
- [PASS] scenario 1 step 2 (QualityLayer) · 445c884 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` · exit 0 · printed `installed platform-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 3 (QualityLayer) · 445c884 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `uninstalled platform-team from /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 4 (QualityLayer) · 445c884 · `ls /tmp/ql-kk-s1/home/.kirokit-installed` · exit 0 · printed `data-team.json` as expected
- [PASS] scenario 1 step 5 (QualityLayer) · 445c884 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `installed data-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 6 (QualityLayer) · 445c884 · `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` · exit 0 · printed `rendered platform-team -> /tmp/ql-kk-s1/out` as expected
- [PASS] scenario 1 step 7 (QualityLayer) · 445c884 · `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 8 (QualityLayer) · 445c884 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 9 (QualityLayer) · 445c884 · `/tmp/ql-kk-s1/kirokit sync --kb-agent x` · exit 1 · as expected
- [PASS] scenario 1 step 10 (QualityLayer) · 445c884 · `go test ./...` · exit 0 · as expected
- slice 1 · 2357174 · `go build ./...` · exit 0 · 0.4s
- slice 1 · 2357174 · `go vet ./...` · exit 0 · 0.2s
- slice 1 · 2357174 · `go test ./...` · exit 0 · 8.6s
- slice 1 · 2357174 · `golangci-lint run` · exit 0 · 0.6s

## Decided while building

<!-- ql:decisions -->
- Slice 1 · Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- Slice 1 · The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.
<!-- /ql:decisions -->

## Checkpoints

### Checkpoint after slice 1 · run 1

**Runner:** fresh helper
**Revision:** 22d8da6b26dc7eb0792ca5cc2267da3a7a6e0dca1f88467fa7494a51e4ff7b6d

- [PASS] scenario 1: steps 1-10 were run by QualityLayer and recorded PASS (cited, not rerun); I additionally confirmed the built binary carries the new code: `render` from `/` fails without `KIROKIT_REPO` and succeeds with it — `evidence/cp-1-run-1-scenario1-probe.txt`
- [FAIL] boundary: `install no-such-team` while `data-team` was installed printed `uninstalled data-team from /tmp/ql-kk-s1/home`, then failed with `kirokit install: manifest: read .../teams/no-such-team/manifest.yaml: ... no such file or directory` (exit 1); `.kirokit-installed` was left empty, so a mistyped team name leaves the machine with no team installed — `evidence/cp-1-run-1-scenario1-probe.txt`

**Result:** fail

### Checkpoint 1 · run 1 · counted toward the stop

It failed in the program: failure 1 of the 4 after which the checkpoint stays open and the build goes on.

## Deviations

- T4 (amended): Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- T2 (found by checkpoint 1): installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes.
- T5 (amended): The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.

## Plan amended

### T4 · amended

Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T5 · amended

The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

## Restart
