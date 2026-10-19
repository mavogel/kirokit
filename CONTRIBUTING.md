# Contributing Guidelines

Thank you for your interest in contributing to `kirokit`. Whether it's a bug report, new feature, correction, or additional documentation, feedback and contributions are greatly appreciated.

Please read through this document before submitting any issues or pull requests so we have all the information needed to respond effectively.


## Reporting Bugs/Feature Requests

Use the GitHub issue tracker to report bugs or suggest features. Check existing open, or recently closed, issues first to make sure somebody else hasn't already reported it. Please include as much of the following as you can:

* A reproducible test case or series of steps
* The output of `kirokit version` and `go version`
* The team and template involved, and any modifications you've made relevant to the bug
* Anything unusual about your environment (OS, `KIRO_HOME`, `KIROKIT_REPO`, `kiro-cli` version)

Never paste secrets, tokens, or decrypted SOPS content into an issue.


## Contributing via Pull Requests

Before sending a pull request, please ensure that:

1. You are working against the latest source on the *main* branch.
2. You check existing open, and recently merged, pull requests to make sure someone else hasn't addressed the problem already.
3. You open an issue to discuss any significant work, so your time isn't wasted.

To send a pull request:

1. Fork the repository and create a branch.
2. Set up your environment (see [Development setup](#development-setup)).
3. Modify the source, focused on the specific change you are contributing. Please don't reformat unrelated code; it makes the change hard to review.
4. Run `task ci` and make sure it passes (see [Checks](#checks)).
5. Commit with a [Conventional Commits](#commit-messages) message.
6. Push to your fork and open a pull request, answering the questions in the pull request template.
7. Pay attention to CI failures reported on the pull request and stay involved in the conversation.

GitHub provides additional documentation on [forking a repository](https://help.github.com/articles/fork-a-repo/) and [creating a pull request](https://help.github.com/articles/creating-a-pull-request/).


## Development setup

Requirements:

| Tool | Used for |
|------|----------|
| [Go](https://go.dev) (version in `go.mod`) | building and testing the CLI |
| [Task](https://taskfile.dev) | running the project commands |
| [golangci-lint](https://golangci-lint.run) | linting (CI pins the version in `.github/workflows/ci.yaml`) |
| [lefthook](https://lefthook.dev), [uv](https://docs.astral.sh/uv/), [gitleaks](https://github.com/gitleaks/gitleaks), [commitizen](https://commitizen-tools.github.io/commitizen/) | git hooks |
| [Node.js](https://nodejs.org) | desktop frontend tests and `scripts/sync-agent-assets.mjs` |
| [`kiro-cli`](https://kiro.dev) (optional) | `go test ./internal/golden` also runs `kiro-cli agent validate`; it skips without a logged-in `kiro-cli` unless `KIROKIT_REQUIRE_KIRO_CLI=1` |

```bash
brew install go go-task golangci-lint lefthook uv gitleaks commitizen node
git clone https://github.com/mavogel/kirokit.git && cd kirokit
lefthook install   # once per clone: whitespace/yaml/json fixers, gitleaks, commit message check
task build
```

Run `kirokit` from the repository root: it resolves `teams/`, `global/` and `env/` relative to the working directory, or relative to `KIROKIT_REPO` when that is set.


## Repository layout

| Path | Contents |
|------|----------|
| `cmd/kirokit/` | cobra commands |
| `internal/` | the implementation: `render`, `golden`, `install`, `manifest`, `externalskills`, `sopswrap`, `mcpversions`, `kiroconfig`, `scaffold` |
| `desktop/` | the Wails desktop app (its own Go module, with a vite + xterm.js frontend) |
| `teams/<team>/` | template sources (`*.tmpl`, double extension) |
| `global/steerings/` | steerings shared between teams |
| `env/mcp-versions.yaml` | central MCP server version pins |
| `testdata/golden/<team>/` | committed expected output |
| `docs/prd/`, `docs/plans/` | product requirements and implementation plans |

The [README](README.md) is the detailed reference for commands, manifests and template functions.


## Checks

A [`Taskfile.yml`](https://taskfile.dev) wraps the common commands:

```bash
task build          # go build -o kirokit ./cmd/kirokit
task vet            # go vet ./...
task lint           # golangci-lint run
task test           # go test ./... -v (includes every team's golden-file comparison)
task ci             # build + vet + lint + test, matching .github/workflows/ci.yaml
task desktop:test   # desktop frontend build + node tests, then vet and test the desktop module
```

CI additionally runs `node scripts/sync-agent-assets.mjs --check`. `task ci` must pass before you open a pull request. If you touch `desktop/`, run `task desktop:test` as well (see `.github/workflows/desktop.yaml`).

Add or extend tests for changed behaviour. For bug fixes, a test that fails before your change and passes after it is the best evidence.


## Working with templates and golden files

Rendered output is locked down by golden files in `testdata/golden/`, compared byte-for-byte in CI.

* After an intentional template change, run `./kirokit update <team>` and review the golden diff. CI fails on drift.
* **Never hand-edit** `testdata/golden/**` or `*.enc.yaml` (the SOPS MAC covers the exact content). Regenerate or re-encrypt instead.
* Rendering uses the fixed `KiroHome` sentinel `/kiro-home`, so goldens match on every machine. Only `install` bakes in the real `$KIRO_HOME`.
* MCP server versions come from `env/mcp-versions.yaml` via `mcpVersion`; an undefined key must fail the render.
* **Secrets never appear in plaintext** in templates, rendered files, or goldens. MCP servers that need them launch via `sopsCommand`/`sopsArgsJSON` (`sops exec-env`) with a `.KiroHome`-absolute secrets path.

### External skills

Declare vendored git skills in a team's `manifest.yaml` under `externalSkills`, then run:

```bash
./kirokit skills fetch <team>
```

Never hand-edit `teams/*/external/**`: the lock hash covers the exact content and render fails on drift. Fetch is the only networked step.


## Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org), enforced by `cz check` in the `commit-msg` hook; release versions are derived from them with `svu`. Allowed types are `build`, `bump`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `plan`, `refactor`, `revert`, `ship`, `style` and `test` (see `.cz.toml`).

```
feat(cli): add setup command to clone the config repository
fix(release): snapshot on PRs only
feat(manifest)!: declare external skills with a paths list
```

Mark breaking changes with `!` after the type/scope and explain what users must change in the body.


## Documentation

Update the [README](README.md) in the same pull request when a command, flag, manifest field, or template function changes. Stale docs are a bug.


## Agent assets

`CLAUDE.md` is only a shim that imports `AGENTS.md`; edit `AGENTS.md`. Project skills live in `.agents/skills/` and mirror to `.claude/skills/`. Verify parity with `node scripts/sync-agent-assets.mjs --check` (`--write` is for recovery only).


## Finding contributions to work on

Looking at the existing issues is a great way to find something to contribute to. As this project uses the default GitHub issue labels (enhancement/bug/duplicate/help wanted/invalid/question/wontfix), 'help wanted' issues are a great place to start.


## Code of Conduct

Be respectful and constructive. For questions or concerns, contact `info@manuel-vogel.de`.


## Security issue notifications

If you discover a potential security issue in this project, please do **not** create a public GitHub issue. Report it privately through GitHub's [private vulnerability reporting](https://github.com/mavogel/kirokit/security/advisories/new) or by email to `info@manuel-vogel.de`.


## Licensing

See the [LICENSE](LICENSE) file for the project's license (Apache 2.0). By submitting a contribution you confirm that you have the right to license it under these terms.
