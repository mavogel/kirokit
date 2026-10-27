# Contributing Guidelines

Thank you for your interest in contributing to `kirokit`. Whether it's a bug report, new feature, correction, or additional documentation, feedback and contributions are greatly appreciated.

Please read through this document before submitting any issues or pull requests so we have all the information needed to respond effectively.


## Reporting Bugs/Feature Requests

Use the GitHub issue tracker to report bugs or suggest features. Check existing open, or recently closed, issues first to make sure somebody else hasn't already reported it. Please include as much of the following as you can:

* A reproducible test case or series of steps
* The output of `kirokit version` and `go version`, and your system (macOS, Linux distribution, Windows version)
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

The CLI builds and tests on macOS, Linux and Windows alike. For the desktop app on Linux, install `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (Debian/Ubuntu names); on Windows, install [NSIS](https://nsis.sourceforge.io) (`makensis`) and run the tasks from Git Bash. The repository's `.gitattributes` checks every text file out with LF, so goldens, lock hashes and SOPS files stay byte-exact on Windows; keep it that way.

On a Mac with Docker or OrbStack, `scripts/linux-docker.sh <command>` runs a command in a container that mirrors CI's Ubuntu runner (Go, Node, Task, sops, WebKitGTK 4.1, NSIS), for example `scripts/linux-docker.sh task desktop:test` or `scripts/linux-docker.sh env GOOS=windows task desktop:build` (which builds the Windows setup programs). `LINUX_DOCKER_ARCH=amd64` picks amd64 on Apple silicon.


## Repository layout

| Path | Contents |
|------|----------|
| `cmd/kirokit/` | cobra commands |
| `internal/` | the implementation: `render`, `quick` (Amazon Quick bundle and ids), `golden`, `install`, `manifest`, `externalskills`, `sopswrap`, `mcpversions`, `kiroconfig`, `scaffold` |
| `desktop/` | the Wails desktop app (its own Go module, with a vite + xterm.js frontend) |
| `teams/<team>/` | template sources (`*.tmpl`, double extension) |
| `global/steerings/` | steerings shared between teams |
| `env/mcp-versions.yaml` | central MCP server version pins |
| `testdata/golden/<team>/` | committed expected output (`rendered/`, `context.json`, and `quick/` for teams with a Quick bundle) |
| `docs/*.md` | the user documentation the README links to |
| `docs/prd/`, `docs/plans/` | product requirements and implementation plans |

The [README](README.md) introduces kirokit; the pages under [`docs/`](docs/) are the detailed reference for commands, manifests and template functions.


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

CI runs build, vet and tests on Ubuntu, Windows and macOS (lint and `node scripts/sync-agent-assets.mjs --check` on Ubuntu), with sops installed and `KIROKIT_REQUIRE_SOPS=1`, so the SOPS launch test (`cmd/kirokit/commands/install_sops_test.go`) runs everywhere; locally it skips without `sops`. The Desktop workflow runs `task desktop:test`, `task desktop:build` and a smoke start (`KIROKIT_DESKTOP_SMOKE=1` makes the app quit once its page has loaded) on all three, and with the GoReleaser Pro key also each system's part of the release as a snapshot. `task ci` must pass before you open a pull request. If you touch `desktop/`, run `task desktop:test` as well. Tests that need a Unix shell, zsh or execute bits skip on Windows; give a Windows counterpart where the behaviour exists there.

Add or extend tests for changed behaviour. For bug fixes, a test that fails before your change and passes after it is the best evidence.

### Showcase video

The README's showcase (`assets/kirokit-showcase.gif`, with the sharper `assets/kirokit-showcase.mp4`) is recorded from the real CLI and desktop app. When the CLI output or the app's UI changes, record it again on a Mac:

```bash
brew install tmux asciinema agg ffmpeg node
task showcase   # scripts/showcase/record.sh; --keep leaves /tmp/kirokit-demo for inspection
```

It also needs Google Chrome in `/Applications`, a signed-in `kiro-cli`, the SOPS age key at its default location (or `SOPS_AGE_KEY_FILE`), network access to github.com, and free ports 34115 and 5173 (no `task desktop:dev` running). The run types into a real Kiro session but never sends a chat message. It refuses to write `assets/` if the recorded text shows your home path, user name or git email, if the GIF exceeds 8 MB or the MP4 4 MB, or if the cut falls outside 30 to 60 seconds.

`node --test scripts/showcase/` tests the recorder's own checks, caption timing and error paths without recording anything (about 40 s; the tests that need tmux, ffmpeg or Chrome skip without them).


## Working with templates and golden files

Rendered output is locked down by golden files in `testdata/golden/`, compared byte-for-byte in CI.

* After an intentional template change, run `./kirokit update <team>` and review the golden diff. CI fails on drift.
* **Never hand-edit** `testdata/golden/**` or `*.enc.yaml` (the SOPS MAC covers the exact content). Regenerate or re-encrypt instead.
* Rendering uses the fixed `KiroHome` sentinel `/kiro-home`, so goldens match on every machine. Only `install` bakes in the real `$KIRO_HOME`.
* MCP server versions come from `env/mcp-versions.yaml` via `mcpVersion`; an undefined key must fail the render.
* **Secrets never appear in plaintext** in templates, rendered files, or goldens. MCP servers that need them launch via `sopsCommand`/`sopsArgsJSON` (`sops exec-env`) with a `.KiroHome`-absolute secrets path.
* Arguments of a `sopsArgsJSON` command must also pass `cmd.exe` on Windows: no spaces, tabs, `"` or `& | < > ^ % ! ( )` (write headers as `Authorization:${VAR}`). `go test ./internal/golden` renders every team for Windows and fails on such an argument.

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

Update the [README](README.md) and the affected page under [`docs/`](docs/) in the same pull request when a command, flag, manifest field, or template function changes. Stale docs are a bug. `node scripts/check-doc-links.mjs` checks that every link and anchor between them still resolves.


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
