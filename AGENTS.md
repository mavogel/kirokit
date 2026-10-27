# kirokit repository instructions

## Project
`kirokit` is a Go CLI (cobra, `go 1.27`) that renders per-team Kiro CLI v3 config templates into Kiro's real config layout. `README.md` is the short intro and `docs/*.md` the detailed reference; `docs/prd/` and `docs/plans/` hold the product requirements and implementation plan.

- `cmd/kirokit/` — cobra commands (`render`, `test`, `update`, `install`, `uninstall`, `sync`, `skills`, `onboard`)
- `internal/` — `externalskills` (vendored git skills: lock, verify, fetch), `render` (templates + funcs), `quick` (Amazon Quick bundle for `render --quick`, `quick ids`), `golden` (golden-file harness), `install`/`manifest` (install + uninstall tracking), `sopswrap`, `mcpversions`, `kiroconfig` (agent config validation), `scaffold` (`onboard` starter layout)
- `desktop/` — the Wails desktop app: its own Go module (`desktop/go.mod`), `desktop/frontend/` (vite + xterm.js, tested with `node --test`); `task desktop:test` runs it
- Ships for macOS, Linux and Windows (amd64, arm64). Per-OS code lives in build-tagged `*_unix.go`/`*_windows.go` pairs (`desktop/terminal`, `desktop/procgroup_*`, `internal/dirlink`) or in functions taking the OS name (`desktop/platform.go`), so every branch is tested on any machine
- `scripts/linux-docker.sh <command>` — runs a command in a container mirroring CI's Ubuntu runner (also builds the Windows setup programs with `env GOOS=windows task desktop:build`)
- `scripts/showcase/` — `task showcase` re-records the README showcase (`assets/kirokit-showcase.gif` and `.mp4`) from the real CLI and app; see CONTRIBUTING "Showcase video"
- `teams/<team>/` — template sources (`*.tmpl`, double extension); `global/steerings/` — shared steerings; `env/mcp-versions.yaml` — central MCP version pins
- `testdata/golden/<team>/` — committed expected output; compared byte-for-byte

## Commands
```bash
task ci          # build + vet + lint + test, same as .github/workflows/ci.yaml
go test ./...    # includes every team's golden comparison
./kirokit test   # golden diff only (build first: task build)
./kirokit update [team]   # regenerate goldens after an intentional template change
```
Run `kirokit` from the repository root: it resolves `teams/`, `global/`, `env/` relative to the working directory, or relative to `KIROKIT_REPO` when that is set.

## Repository rules
- Template change → run `./kirokit update <team>` and review the golden diff; CI fails on drift. Never hand-edit `testdata/golden/**` or `*.enc.yaml` (SOPS MAC covers exact content).
- Rendering uses the fixed `KiroHome` sentinel `/kiro-home` so goldens match on every machine; only `install` bakes in the real `$KIRO_HOME`.
- External skills: declare them in `manifest.yaml` `externalSkills` and run `kirokit skills fetch <team>`; never hand-edit `teams/*/external/**` (the lock hash covers exact content, and render fails on drift). Fetch is the only networked step.
- Secrets never appear in plaintext in templates, rendered files, or goldens; MCP servers needing them launch via `sopsCommand`/`sopsArgsJSON` (`sops exec-env`) with a `.KiroHome`-absolute secrets path. On Windows sops runs the command through `cmd.exe`, so those arguments may not contain whitespace, `"` or `& | < > ^ % ! ( )`.
- CI runs the CLI and desktop tests on Ubuntu, Windows and macOS; `.gitattributes` keeps checkouts LF and goldens, vendored skills and `*.enc.yaml` untouched. Tests needing a Unix shell or execute bits skip on Windows (`internal/testutil`).
- MCP versions come from `env/mcp-versions.yaml` via `mcpVersion`; an undefined key must fail the render.
- `go test ./internal/golden` also runs `kiro-cli agent validate`; it skips without a logged-in `kiro-cli` unless `KIROKIT_REQUIRE_KIRO_CLI=1`.
- Commits follow Conventional Commits (`cz check` in lefthook; `svu` consumes them). `.claude/` is gitignored tool state.
- Update `README.md` and the affected `docs/*.md` page in the same change when a command, flag, manifest field, or template function changes; `node scripts/check-doc-links.mjs` checks their links and anchors.
- Quick skill ids belong to one Quick account: `quick-ids.yaml` is gitignored and never committed; Quick goldens use the made-up `testdata/golden/quick-ids.yaml`.

## Agent assets
`CLAUDE.md` is only `@AGENTS.md`; edit this file, not the shim. Project skills (none yet) live in `.agents/skills/` and mirror to `.claude/skills/`. Verify parity with `node scripts/sync-agent-assets.mjs --check` (CI runs it); `--write` is recovery only.
