# kiro-cli-kit repository instructions

## Project
`kirokit` is a Go CLI (cobra, `go 1.23`) that renders per-team Kiro CLI v3 config templates into Kiro's real config layout. `README.md` is the detailed reference; `docs/prd/` and `docs/plans/` hold the product requirements and implementation plan.

- `cmd/kirokit/` — cobra commands (`render`, `test`, `update`, `install`, `uninstall`, `onboard`)
- `internal/` — `render` (templates + funcs), `golden` (golden-file harness), `install`/`manifest` (install + uninstall tracking), `sopswrap`, `mcpversions`, `kiroconfig` (agent config validation), `scaffold` (`onboard` starter layout)
- `teams/<team>/` — template sources (`*.tmpl`, double extension); `global/steerings/` — shared steerings; `env/mcp-versions.yaml` — central MCP version pins
- `testdata/golden/<team>/` — committed expected output; compared byte-for-byte

## Commands
```bash
task ci          # build + vet + lint + test, same as .github/workflows/ci.yaml
go test ./...    # includes every team's golden comparison
./kirokit test   # golden diff only (build first: task build)
./kirokit update [team]   # regenerate goldens after an intentional template change
```
Run `kirokit` from the repository root: it resolves `teams/`, `global/`, `env/` relative to the working directory.

## Repository rules
- Template change → run `./kirokit update <team>` and review the golden diff; CI fails on drift. Never hand-edit `testdata/golden/**` or `*.enc.yaml` (SOPS MAC covers exact content).
- Rendering uses the fixed `KiroHome` sentinel `/kiro-home` so goldens match on every machine; only `install` bakes in the real `$KIRO_HOME`.
- Secrets never appear in plaintext in templates, rendered files, or goldens; MCP servers needing them launch via `sopsCommand`/`sopsArgsJSON` (`sops exec-env`) with a `.KiroHome`-absolute secrets path.
- MCP versions come from `env/mcp-versions.yaml` via `mcpVersion`; an undefined key must fail the render.
- `go test ./internal/golden` also runs `kiro-cli agent validate`; it skips without a logged-in `kiro-cli` unless `KIROKIT_REQUIRE_KIRO_CLI=1`.
- Commits follow Conventional Commits (`cz check` in lefthook; `semantic-release` consumes them). `.claude/` is gitignored tool state.
- Update `README.md` in the same change when a command, flag, manifest field, or template function changes.

## Agent assets
`CLAUDE.md` is only `@AGENTS.md`; edit this file, not the shim. Project skills (none yet) live in `.agents/skills/` and mirror to `.claude/skills/`. Verify parity with `node scripts/sync-agent-assets.mjs --check` (CI runs it); `--write` is recovery only.
