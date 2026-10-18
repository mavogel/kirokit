# Desktop app from design · the build

## Summary

<!-- ql:summary -->
24 of 24 tasks done · 7 of 7 slices finished · started 18 Oct, 00:00 · 2 h 12 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | One team stays installed and kirokit runs from any folder | `KIROKIT_REPO` roots every repository path, `install` and `sync` uninstall any other installed team first, and `sync` without a team only updates the checkout. | done |
| 2 | The app opens with the design's shell and remembers its settings | A Wails window shows the logo, the sidebar with five sections in light and dark mode, the collapsing icon rail, and a Settings page that saves. | done |
| 3 | The pages show the repository's real data | Teams, Agents, Install and Skills and MCP versions read the app's own copy of the repository, and the screens redraw when install records change. | done |
| 4 | A real terminal docks at the bottom | The dock opens, closes and resizes like VS Code's and holds two tabs: your login shell and the kirokit tab for button commands. | done |
| 5 | A person installs a team with buttons only | First launch asks for the repository's address and clones it, the setup dialog checks the prerequisites, and Install, Sync, Uninstall and Update run their `kirokit` commands in the terminal. | done |
| 6 | One command builds the macOS app and CI builds it too | `task desktop:build` produces a universal `kirokit.app` with the CLI inside, and a new workflow runs the app's tests and build on macOS. | done |
| 7 | Each release attaches the app and a Homebrew cask | A tagged release also builds the macOS app, attaches `kirokit-desktop_<version>_darwin_universal.zip` and pushes the `kirokit-desktop` cask. | done |
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
- T9 · 492219a · `cd desktop && go test -run 'Capture|Prereq' ./...` · exit 0 · 1.5s
- T10 · 492219a · `cd desktop && go test -run Catalog ./...` · exit 0 · 0.7s
- T11 · 492219a · `cd desktop && go test -race -run Installs ./...` · exit 0 · 1.7s
- T9 · 492219a · `cd desktop && go test -run 'Capture|Prereq' ./...` · exit 0 · 0.4s
- T12 · 492219a · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 0.5s
- slice 3 · 492219a · `task desktop:test` · exit 0 · 3.5s
- T13 · 421cad0 · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 0.9s
- T14 · 421cad0 · `cd desktop && go test -race ./terminal/` · exit 0 · 1.5s
- T15 · 421cad0 · `cd desktop && go test -race -run Shell ./...` · exit 0 · 1.9s
- T16 · 421cad0 · `cd desktop && go test -race -run Runner ./...` · exit 0 · 2.6s
- T13 · 421cad0 · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 0.9s
- slice 4 · 421cad0 · `task desktop:test` · exit 0 · 5.4s
- T17 · e10d117 · `cd desktop && go test -race -run Clone ./...` · exit 0 · 4.3s
- T18 · e10d117 · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 1s
- T19 · e10d117 · `cd desktop/frontend && node --test src/ && npm run build` · reused: nothing changed since they passed
- T20 · e10d117 · `task desktop:test` · exit 0 · 8.8s
- slice 5 · e10d117 · `task desktop:test` · exit 0 · 3.4s
- slice 5 · e10d117 · `go test ./...` · exit 0 · 3.6s
- T21 · fd79636 · `task desktop:build` · exit 0 · 5.6s
- T22 · fd79636 · `ruby -ryaml -e 'YAML.load_file(".github/workflows/desktop.yaml")'` · exit 0 · 0.1s
- slice 6 · fd79636 · `task desktop:build` · reused: nothing changed since they passed
- slice 6 · fd79636 · `go test ./...` · exit 0 · 3.6s
- slice 6 · fd79636 · `ruby -ryaml -e 'YAML.load_file(".github/workflows/desktop.yaml")'` · reused: nothing changed since they passed
- T23 · f541fc8 · `bash scripts/desktop-release_test.sh` · exit 1 · 0.1s
- T23 · f541fc8 · `bash scripts/desktop-release_test.sh` · exit 0 · 0.2s
- T24 · f541fc8 · `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")' && bash scripts/desktop-release_test.sh` · exit 0 · 0.2s
- slice 7 · f541fc8 · `ruby -ryaml -e 'YAML.load_file(".github/workflows/release.yaml")'` · exit 0 · 0.1s
- slice 7 · f541fc8 · `bash scripts/desktop-release_test.sh` · exit 0 · 0.1s
- slice 7 · f541fc8 · `go test ./...` · exit 0 · 3.7s
- all · 72e379a · `go test ./...` · exit 0 · 0.5s
- all · 72e379a · `go vet ./...` · exit 0 · 0.1s
- all · 72e379a · `golangci-lint run` · exit 0 · 0.6s
- all · 72e379a · `go build ./...` · exit 0 · 0.6s
- all · 72e379a · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- all · 72e379a · `cd desktop/frontend && npm ci && npm run build` · exit 0 · 1.9s
- all · 72e379a · `cd desktop && go vet ./... && go test ./...` · exit 0 · 6.4s
- all · 72e379a · `task desktop:build` · exit 0 · 5.5s
- all · 72e379a · `go test ./...` · reused: nothing changed since they passed
- all · 72e379a · `go vet ./...` · reused: nothing changed since they passed
- all · 72e379a · `golangci-lint run` · reused: nothing changed since they passed
- all · 72e379a · `go build ./...` · reused: nothing changed since they passed
- all · 72e379a · `node scripts/sync-agent-assets.mjs --check` · reused: nothing changed since they passed
- all · 72e379a · `cd desktop/frontend && npm ci && npm run build` · reused: nothing changed since they passed
- all · 72e379a · `cd desktop && go vet ./... && go test ./...` · reused: nothing changed since they passed
- all · 72e379a · `task desktop:build` · reused: nothing changed since they passed
- all · 5ebc94b · `go test ./...` · exit 0 · 0.6s
- all · 5ebc94b · `go vet ./...` · exit 0 · 0.1s
- all · 5ebc94b · `golangci-lint run` · exit 0 · 0.6s
- all · 5ebc94b · `go build ./...` · exit 0 · 0.3s
- all · 5ebc94b · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- all · 5ebc94b · `cd desktop/frontend && npm ci && npm run build` · exit 0 · 1.9s
- all · 5ebc94b · `cd desktop && go vet ./... && go test ./...` · exit 0 · 6.3s
- all · 5ebc94b · `task desktop:build` · exit 0 · 5.7s
- all · 5ebc94b · `go test ./...` · reused: nothing changed since they passed
- all · 5ebc94b · `go vet ./...` · reused: nothing changed since they passed
- all · 5ebc94b · `golangci-lint run` · reused: nothing changed since they passed
- all · 5ebc94b · `go build ./...` · reused: nothing changed since they passed
- all · 5ebc94b · `node scripts/sync-agent-assets.mjs --check` · reused: nothing changed since they passed
- all · 5ebc94b · `cd desktop/frontend && npm ci && npm run build` · reused: nothing changed since they passed
- all · 5ebc94b · `cd desktop && go vet ./... && go test ./...` · reused: nothing changed since they passed
- all · 5ebc94b · `task desktop:build` · reused: nothing changed since they passed
- T18 · 5ebc94b · `cd desktop/frontend && node --test src/ && npm run build` · exit 0 · 0.9s
- all · 65af338 · `go test ./...` · exit 0 · 3.7s
- all · 65af338 · `go vet ./...` · exit 0 · 0.1s
- all · 65af338 · `golangci-lint run` · exit 0 · 0.6s
- all · 65af338 · `go build ./...` · exit 0 · 0.3s
- all · 65af338 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
- all · 65af338 · `cd desktop/frontend && npm ci && npm run build` · exit 0 · 1.9s
- all · 65af338 · `cd desktop && go vet ./... && go test ./...` · exit 0 · 0.6s
- all · 65af338 · `task desktop:build` · exit 0 · 6s
- all · 65af338 · `go test ./...` · reused: nothing changed since they passed
- all · 65af338 · `go vet ./...` · reused: nothing changed since they passed
- all · 65af338 · `golangci-lint run` · reused: nothing changed since they passed
- all · 65af338 · `go build ./...` · reused: nothing changed since they passed
- all · 65af338 · `node scripts/sync-agent-assets.mjs --check` · reused: nothing changed since they passed
- all · 65af338 · `cd desktop/frontend && npm ci && npm run build` · reused: nothing changed since they passed
- all · 65af338 · `cd desktop && go vet ./... && go test ./...` · reused: nothing changed since they passed
- all · 65af338 · `task desktop:build` · reused: nothing changed since they passed
- T20 · 65af338 · `task desktop:test` · exit 0 · 2.6s

## Decided while building

<!-- ql:decisions -->
- Slice 1 · Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- Slice 1 · Installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes.
- Slice 1 · A mistyped team or a failed update in sync removed the installed team first, so sync now updates and checks the new team before removing anything; nothing you approved changes.
- Slice 1 · The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.
- The Plan · App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes.
- The Plan · Startup hooks, the module requirements, a Settings note, shared page helpers and ignored generated files are needed by the data slice; nothing you approved changes.
- Slice 3 · A placeholder file in the ignored build folder lets a clean checkout compile the app; nothing you approved changes.
- Slice 4 · The build keeps the dist placeholder and the dock markup lives in the page; nothing you approved changes.
- The Plan · Clone logic, first-launch page, dialogs, docs and a keystroke queue need files no card named; oracle commands now set the repository folder.
- Slice 2 · Scenario 2 now clones first, because the pages stay hidden until a clone exists; nothing you approved changes.
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

### Checkpoint after slice 2 · run 1

**Runner:** fresh helper
**Revision:** 4de8c22a2fb7743e33680f790c94b09ad4b91ac432b29e4fc81d513df611bb78

- [PASS] scenario 2: all 8 steps showed their Expected. Dev server answered on :34115; at 1280x800 logo, five sidebar items (Teams, Agents, Install and sync, Skills and MCP versions, Settings) and main pane rendered; toggle took the sidebar 232px -> 60px (main 1220px wide) and back; Cmd-B twice gave 60px -> 232px; dark emulation switched body from rgb(255,255,255)/rgb(29,23,48) to rgb(23,19,31)/rgb(236,233,246); Kiro home "/" was refused with "Choose a folder inside your home folder, not the home folder itself" beside the field (field invalid, no settings.json created); the two valid values showed "Saved on this computer" and wrote settings.json; after stopping wails/vite/app by pid, port 34115 freed, restart with same HOME and reload, Settings showed /tmp/ql-kk-s2/kiro and /tmp/ql-kk-s2/origin — `evidence/cp-2-run-1-expanded.png`, `evidence/cp-2-run-1-rail.png`, `evidence/cp-2-run-1-settings-saved.png`, `evidence/cp-2-run-1-settings-restored.png`
- [PASS] boundary: Kiro home set to "/" (the filesystem root) was refused with a message and nothing was saved. The Teams pane shows only its heading and intro line (no team list yet), which this scenario does not require.
- Finding: `wails dev` generates `desktop/frontend/wailsjs/` and `desktop/frontend/package.json.md5`; both appear as untracked in `git status` and are not git-ignored. The Plan says they must not be committed, so they need .gitignore entries.

**Verdict:** The desktop window opened in the dev server with the designed sidebar and main pane, collapsed to an icon rail by button and Cmd-B, followed dark mode, refused an unsafe Kiro home, and kept its two saved settings across a full restart.

**Pictures:**

- `evidence/cp-2-run-1-expanded.png` — expanded sidebar with logo, five sections and the Teams pane at 1280x800 (dark scheme), proving the layout matches the design.
- `evidence/cp-2-run-1-rail.png` — sidebar collapsed to an icon rail with a wider main pane.
- `evidence/cp-2-run-1-settings-saved.png` — Settings with both values and the "Saved on this computer" confirmation.
- `evidence/cp-2-run-1-settings-restored.png` — Settings after restart and reload with both values still filled in.

**Try it:**

1. Start the app with an empty home folder. `mkdir -p /tmp/kk-try/home && env HOME=/tmp/kk-try/home task desktop:dev`
2. Open http://localhost:34115, toggle the sidebar and press Command-B, set Kiro home to `/` and Save (refused), then to a folder under /tmp and Save, restart and reopen Settings to see the values kept.

**Result:** pass

### Checkpoint after slice 4 · run 1

**Runner:** fresh helper
**Revision:** 0ad2ad0722dfa9b47ff1d2f9ed675bd9d8718f3d61625d180be2f1b638272933

- [PASS] scenario 3: all steps showed their Expected. Teams listed data-team and platform-team, none installed; Agents listed 4 agents with tools; Install and sync empty; Skills and MCP versions pins equal env/mcp-versions.yaml. The dock opened (220), a drag grew it to 420 and main shrank 580 to 380 (synthetic pointer events on the resize handle). `command -v kirokit` printed the CLI given by KIROKIT_DESKTOP_CLI (dev build in scratch). After a drag to a new height, `stty size` printed `11 130`, equal to xterm's rendered 11 rows and about 130 columns. `kirokit install data-team` printed `installed data-team into /tmp/ql-kk-s3/kiro` and Teams and Install and sync showed it installed within 2 s; `kirokit uninstall data-team` printed `uninstalled data-team from /tmp/ql-kk-s3/kiro` and both pages showed none; hide and show kept the zsh output. xterm.js rendered inside the webview. — `evidence/cp-4-run-1-s3-install-teams.png`, `evidence/cp-4-run-1-s3-install-page.png`, `evidence/cp-4-run-1-s3-reopened-dock.png`
- [PASS] boundary: via the bound API, `Run(["sync","--help"])` then immediately `Run(["render","data-team"])` rejected the second with "kirokit is still running a command"; a later `Run(["test"])` was accepted after the first ended.

**Observation (not a scenario failure):** on the first typing attempt through the automation tool the zsh tab received keystrokes out of order (`commadn -v kiirok`); a retry was correct. Possible per-keystroke `ShellWrite` reordering under fast input; seen once, not reproduced, may be a dev-server or harness artefact.

**Verdict:** The running desktop app showed the clone's real teams, agents, skills and MCP pins. The embedded zsh terminal reported the same size as the dock after a resize, and typed `kirokit install` and `uninstall` commands updated the Teams and Install and sync pages within two seconds.

**Pictures:**

- `evidence/cp-4-run-1-s3-install-teams.png` — dock open with zsh output ending `installed data-team into /tmp/ql-kk-s3/kiro` and the data-team card marked installed.
- `evidence/cp-4-run-1-s3-install-page.png` — Install and sync listing data-team as installed.
- `evidence/cp-4-run-1-s3-reopened-dock.png` — dock reopened after closing, earlier zsh output still there.

**Try it:**

1. Start the app with a temporary HOME and KIRO_HOME, open the terminal dock, and type in the zsh tab: `kirokit install data-team`
2. Look for the Teams card for data-team switching to installed within two seconds, then run `kirokit uninstall data-team` and see it revert.

**Result:** pass

## Deviations

- T4 (amended): Git calls in tests and skills fetch now ignore the commit tool's repository variables, so two more files change; nothing you approved changes.
- T2 (found by checkpoint 1): installing a mistyped team name removed the installed team first and then failed, so install now checks the new team renders before removing anything; nothing you approved changes.
- T3 (found by checkpoint 1): a mistyped team or a failed update in sync removed the installed team first, so sync now updates and checks the new team before removing anything; nothing you approved changes.
- T5 (amended): The scheduled-sync setup script gets a header comment about one team per machine; nothing you approved changes.
- T6, T8 (amended): App bundle id and macOS 13 minimum live in two plist files, and the dev server serves the logo; nothing you approved changes.
- T9, T12 (amended): Startup hooks, the module requirements, a Settings note, shared page helpers and ignored generated files are needed by the data slice; nothing you approved changes.
- T9 (amended): A placeholder file in the ignored build folder lets a clean checkout compile the app; nothing you approved changes.
- T13 (amended): The build keeps the dist placeholder and the dock markup lives in the page; nothing you approved changes.
- T17, T18, T19, T20 (amended): Clone logic, first-launch page, dialogs, docs and a keystroke queue need files no card named; oracle commands now set the repository folder.
- T7 (amended): Scenario 2 now clones first, because the pages stay hidden until a clone exists; nothing you approved changes.
- T20 (judged): a failed clone from Settings left its address stored and the status stuck on cloning, so the address is now stored only after a clone succeeds and the failure shows beside it; nothing you approved changes.

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

### T9, T12 · amended

Startup hooks, the module requirements, a Settings note, shared page helpers and ignored generated files are needed by the data slice; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T9 · amended

A placeholder file in the ignored build folder lets a clean checkout compile the app; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T13 · amended

The build keeps the dist placeholder and the dock markup lives in the page; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T17, T18, T19, T20 · amended

Clone logic, first-launch page, dialogs, docs and a keystroke queue need files no card named; oracle commands now set the repository folder.

Amended by the agent on 2026-10-17; changed 04-outline.md.

### T7 · amended

Scenario 2 now clones first, because the pages stay hidden until a clone exists; nothing you approved changes.

Amended by the agent on 2026-10-17; changed 04-outline.md.

## Restart
