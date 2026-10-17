# Desktop app from design · the build

## Summary

<!-- ql:summary -->
8 of 24 tasks done · 2 of 7 slices finished · started 18 Oct, 00:00 · 29 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | One team stays installed and kirokit runs from any folder | `KIROKIT_REPO` roots every repository path, `install` and `sync` uninstall any other installed team first, and `sync` without a team only updates the checkout. | done |
| 2 | The app opens with the design's shell and remembers its settings | A Wails window shows the logo, the sidebar with five sections in light and dark mode, the collapsing icon rail, and a Settings page that saves. | done |
| 3 | The pages show the repository's real data | Teams, Agents, Install and Skills and MCP versions read the app's own copy of the repository, and the screens redraw when install records change. | next |
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
- [PASS] scenario 1 step 1 (QualityLayer) · ee57a68 · `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` · exit 0 · as expected
- [PASS] scenario 1 step 2 (QualityLayer) · ee57a68 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` · exit 0 · printed `installed platform-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 3 (QualityLayer) · ee57a68 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `uninstalled platform-team from /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 4 (QualityLayer) · ee57a68 · `ls /tmp/ql-kk-s1/home/.kirokit-installed` · exit 0 · printed `data-team.json` as expected
- [PASS] scenario 1 step 5 (QualityLayer) · ee57a68 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `installed data-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 6 (QualityLayer) · ee57a68 · `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` · exit 0 · printed `rendered platform-team -> /tmp/ql-kk-s1/out` as expected
- [PASS] scenario 1 step 7 (QualityLayer) · ee57a68 · `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 8 (QualityLayer) · ee57a68 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 9 (QualityLayer) · ee57a68 · `/tmp/ql-kk-s1/kirokit sync --kb-agent x` · exit 1 · as expected
- [PASS] scenario 1 step 10 (QualityLayer) · ee57a68 · `go test ./...` · exit 0 · as expected
- slice 1 · 20c83fa · `go build ./...` · exit 0 · 0.3s
- slice 1 · 20c83fa · `go vet ./...` · exit 0 · 0.1s
- slice 1 · 20c83fa · `go test ./...` · exit 0 · 3.2s
- slice 1 · 20c83fa · `golangci-lint run` · exit 0 · 0.5s
- [PASS] scenario 1 step 1 (QualityLayer) · 60fb185 · `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` · exit 0 · as expected
- [PASS] scenario 1 step 2 (QualityLayer) · 60fb185 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` · exit 0 · printed `installed platform-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 3 (QualityLayer) · 60fb185 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `uninstalled platform-team from /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 4 (QualityLayer) · 60fb185 · `ls /tmp/ql-kk-s1/home/.kirokit-installed` · exit 0 · printed `data-team.json` as expected
- [PASS] scenario 1 step 5 (QualityLayer) · 60fb185 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `installed data-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 6 (QualityLayer) · 60fb185 · `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` · exit 0 · printed `rendered platform-team -> /tmp/ql-kk-s1/out` as expected
- [PASS] scenario 1 step 7 (QualityLayer) · 60fb185 · `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 8 (QualityLayer) · 60fb185 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 9 (QualityLayer) · 60fb185 · `/tmp/ql-kk-s1/kirokit sync --kb-agent x` · exit 1 · as expected
- [PASS] scenario 1 step 10 (QualityLayer) · 60fb185 · `go test ./...` · exit 0 · as expected
- [PASS] scenario 1 step 1 (QualityLayer) · f2ee3d4 · `rm -rf /tmp/ql-kk-s1 && mkdir -p /tmp/ql-kk-s1/home && go build -o /tmp/ql-kk-s1/kirokit ./cmd/kirokit` · exit 0 · as expected
- [PASS] scenario 1 step 2 (QualityLayer) · f2ee3d4 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install platform-team` · exit 0 · printed `installed platform-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 3 (QualityLayer) · f2ee3d4 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `uninstalled platform-team from /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 4 (QualityLayer) · f2ee3d4 · `ls /tmp/ql-kk-s1/home/.kirokit-installed` · exit 0 · printed `data-team.json` as expected
- [PASS] scenario 1 step 5 (QualityLayer) · f2ee3d4 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit install data-team` · exit 0 · printed `installed data-team into /tmp/ql-kk-s1/home` as expected
- [PASS] scenario 1 step 6 (QualityLayer) · f2ee3d4 · `cd / && env KIROKIT_REPO=/Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/kirokit render platform-team --out /tmp/ql-kk-s1/out` · exit 0 · printed `rendered platform-team -> /tmp/ql-kk-s1/out` as expected
- [PASS] scenario 1 step 7 (QualityLayer) · f2ee3d4 · `git clone -q /Users/mavogel/Developer/projects/kiro-cli-kit /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 8 (QualityLayer) · f2ee3d4 · `env KIRO_HOME=/tmp/ql-kk-s1/home /tmp/ql-kk-s1/kirokit sync --latest-tag --repo /tmp/ql-kk-s1/clone` · exit 0 · as expected
- [PASS] scenario 1 step 9 (QualityLayer) · f2ee3d4 · `/tmp/ql-kk-s1/kirokit sync --kb-agent x` · exit 1 · as expected
- [PASS] scenario 1 step 10 (QualityLayer) · f2ee3d4 · `go test ./...` · exit 0 · as expected
- T6 · 48002f1 · `task desktop:test` · exit 0 · 2s
- T7 · 48002f1 · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 0.5s
- T8 · 48002f1 · `task desktop:test` · reused: nothing changed since they passed
- slice 2 · 48002f1 · `task desktop:test` · reused: nothing changed since they passed
- slice 2 · 48002f1 · `go build ./...` · exit 0 · 0.4s

## Decided while building

<!-- ql:decisions -->
- Slice 1 · Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- Slice 1 · Installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes.
- Slice 1 · A mistyped team or a failed update in sync removed the installed team first, so sync now updates and checks the new team before removing anything; nothing you approved changes.
- Slice 1 · The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.
- The Plan · App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes.
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

### Checkpoint after slice 1 · run 2

**Runner:** fresh helper
**Revision:** ba2b64db88b601e8d236dc46e5a3063fe7eb702799fd20209ed03ff58af5afe0

- [PASS] scenario 1: steps 1-10 were already run by QualityLayer and recorded PASS (cited, not rerun) — `evidence/cp-1-run-2-sync-unknown-team.txt` covers only the boundary
- [FAIL] boundary: with data-team installed, `kirokit sync nonexistent-team --repo <clone>` printed `uninstalled data-team from <home>`, then failed with `manifest: read teams/nonexistent-team/manifest.yaml: no such file`, exit 1; `.kirokit-installed` was left empty. A failed sync destroys the working install (`uninstallOthers` runs in `cmd/kirokit/commands/sync.go` before the team or the checkout is validated, and before a failing `git pull`). — `evidence/cp-1-run-2-sync-unknown-team.txt`

**Result:** fail

### Checkpoint 1 · run 2 · counted toward the stop

It failed in the program: failure 2 of the 4 after which the checkpoint stays open and the build goes on.

### Checkpoint after slice 1 · run 3

**Runner:** fresh helper
**Revision:** 723e600bb3543944a72cc02279a955bc157cd859d35148320b1deae7c4f82ac9

- [PASS] scenario 1: steps 1-10 were run by QualityLayer and recorded PASS; not re-run. Extra probes below also passed — `evidence/cp-1-run-3-boundary.txt`
- [MISSING] App picture: `desktop/` does not exist in this checkout (slice 1 is CLI-only), so no running App could be captured; scenario 1 has no App steps.
- [PASS] boundary: with platform-team installed, `install nope-team` and `sync nope-team` both exited 1 with a missing-manifest error and left `platform-team.json` installed; `sync platform-team` with an unreachable remote (fetch fails) and with a diverged clone (`git pull --ff-only` fails) also exited 1 and left the team installed with its agents intact.

**Result:** fail

### Checkpoint 1 · run 3 · counted toward the stop

It failed in the program: failure 3 of the 4 after which the checkpoint stays open and the build goes on.

### Checkpoint after slice 1 · run 4

**Runner:** fresh helper
**Revision:** 723e600bb3543944a72cc02279a955bc157cd859d35148320b1deae7c4f82ac9

- [PASS] scenario 1: QualityLayer's recorded PASS for steps 1-10 stands (not rerun). I re-ran the key steps 2, 3, 4 and 5 against a freshly built binary and saw `installed platform-team`, then `uninstalled platform-team` plus `installed data-team`, only `data-team.json` listed, and a repeat install of data-team succeeding — `evidence/cp-1-run-4-terminal-install-replace.txt`
- [PASS] boundary: `install no-such-team` exited 1 with `kirokit install: manifest: read teams/no-such-team/manifest.yaml: ... no such file or directory`, and `.kirokit-installed` still held only `data-team.json`.

**Verdict:** The CLI keeps exactly one team installed: installing a second team removes the first, an unknown team fails with an error and leaves the installed team untouched, and re-installing the same team works. Slice 1 is CLI-only and has no App screens, so no App was seen or run.

**Pictures:**

- `evidence/cp-1-run-4-terminal-install-replace.png` — Terminal output of the CLI (not the App), rendered to PNG from the captured output of the commands (text copy in `evidence/cp-1-run-4-terminal-install-replace.txt`): platform-team replaced by data-team, the unknown-team error with exit 1, and data-team still the only installed team.

**Try it:**

1. Install one team, then another, in a scratch Kiro home, then list the installed marker; from the repository root: `env KIRO_HOME=/tmp/kk-try go run ./cmd/kirokit install platform-team && env KIRO_HOME=/tmp/kk-try go run ./cmd/kirokit install data-team && ls /tmp/kk-try/.kirokit-installed`
2. Look for the `uninstalled platform-team` line and a single `data-team.json`; then run `env KIRO_HOME=/tmp/kk-try go run ./cmd/kirokit install no-such-team` and check it errors and `data-team.json` is still there.

**Result:** pass

## Deviations

- T4 (amended): Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- T2 (found by checkpoint 1): installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes.
- T3 (found by checkpoint 1): a mistyped team or a failed update in sync removed the installed team first, so sync now updates and checks the new team before removing anything; nothing you approved changes.
- T5 (amended): The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.
- T6, T8 (amended): App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes.

## Plan amended

### T4 · amended

Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T5 · amended

The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T6, T8 · amended

App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

## Restart
