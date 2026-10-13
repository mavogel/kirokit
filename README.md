# kiro-cli-kit

A `kirokit` Go CLI plus repository layout that lets multiple teams own their own [Kiro CLI v3](https://kiro.dev) config sources as templates. Each team's agents, hooks, prompts, skills, and steering docs are rendered deterministically into Kiro's real config layout, MCP server versions are pinned centrally, secrets are injected via [SOPS](https://github.com/getsops/sops) at MCP-launch time (never baked into any committed or rendered file), and the whole rendering pipeline is locked down with golden-file tests that run in CI.

See `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md` for the full product requirements and `docs/plans/2026-09-28-kiro-cli-kit.md` for the implementation plan.

## Quickstart

```bash
go build -o kirokit ./cmd/kirokit

# Render an existing team without touching anything committed:
./kirokit render platform-team --out /tmp/platform-team-rendered

# Verify every team's templates still match their committed golden fixtures
# (this is exactly what CI runs):
./kirokit test

# Scaffold a brand-new team:
./kirokit onboard my-new-team
```

## Taskfile

A [`Taskfile.yml`](https://taskfile.dev) wraps the common Go commands. Install `task` via Homebrew (`brew install go-task`), then:

```bash
task build   # go build -o kirokit ./cmd/kirokit
task vet     # go vet ./...
task lint    # golangci-lint run
task test    # go test ./... -v (includes every team's golden-file comparisons)
task ci      # build + vet + lint + test, matching .github/workflows/ci.yaml
```

## Directory Layout

```
teams/<team>/manifest.yaml              # team identity + global steering opt-ins
teams/<team>/agents/<agent>/agent.json.tmpl
teams/<team>/agents/<agent>/env/secrets.enc.yaml   # optional, SOPS-encrypted
teams/<team>/hooks/*.json.tmpl          # shared across the team's agents
teams/<team>/prompts/*.md.tmpl
teams/<team>/skills/<skill>/SKILL.md.tmpl
teams/<team>/steerings/*.md.tmpl

global/steerings/*.md                   # shared steering docs teams opt into
env/mcp-versions.yaml                   # central MCP server version pins

testdata/golden/<team>/rendered/...     # committed golden output tree
testdata/golden/<team>/context.json     # committed golden template-context snapshot
```

Every templated source file uses a double extension, `<real-extension>.tmpl` (e.g. `agent.json.tmpl`, `steering.md.tmpl`) — the renderer discovers templates via a `**/*.tmpl` glob and strips the trailing `.tmpl` to produce the output filename. Kiro's own layout is flatter than the kit's source layout: an agent's source lives under `agents/<agent>/agent.json.tmpl` but renders to the flat `agents/<agent>.json`, matching Kiro CLI v3's real `.kiro/agents/*.json` schema.

`kirokit` resolves `teams/`, `global/steerings/`, and `env/mcp-versions.yaml` relative to the current working directory — it has no repo-root auto-discovery, so always run it from the repository root.

Two fully-worked example teams are included: `teams/platform-team/` (`devops` + `platformdev`, both with SOPS-secured MCP servers, opted into the shared `security-baseline` global steering) and `teams/data-team/` (`javadev` with a SOPS-secured Snowflake MCP server, `godev` with a plain non-secret MCP server, opted **out** of any global steering). Copy either as a starting point, or use `kirokit onboard` — duplication between teams is expected and fine.

A prompt that no agent uses as its `prompt` is still useful: Kiro CLI v3 lists every file in `$KIRO_HOME/prompts/` as a slash command (`@name` on other engines), and `$ARGUMENTS` / `${1}` in the file are filled in from what you type. For example, `teams/platform-team/prompts/create-steering.md.tmpl` becomes `/create-steering [focus]` after `kirokit install platform-team`: it walks an agent through creating a repository's `.kiro/steering/` files (the `product.md` / `tech.md` / `structure.md` foundation plus focused `fileMatch`, `manual` and `auto` files only where the repo shows a need), and uses the team's opted-in global steerings to tell the agent not to restate them.

## manifest.yaml Reference

```yaml
team: my-team
globalSteerings:
  - security-baseline   # must match a file under global/steerings/<slug>.md
```

`globalSteerings` may be an empty list (`[]`) to explicitly opt out — see `teams/data-team/manifest.yaml`.

Every opted-in global steering is added to the `resources` of **every agent** of the team automatically, as `file://$KIRO_HOME/kirokit-steerings/<team>/<slug>.md` (appended to the agent's own `resources`, created if missing). The steering is rendered to that file, not to Kiro's auto-loaded `steering/` directory, so it applies only to the opted-in team's agents. A global steering file should start with Kiro's steering front matter (e.g. `inclusion: always`, as `global/steerings/security-baseline.md` does). A template can still inline one with `{{ globalSteering "<slug>" }}`, but that inlines the front matter verbatim too, so prefer the automatic resource.

Every agent of every team, opted in or not, also gets `file://.kiro/steering/**/*.md` and `skill://.kiro/skills/**/SKILL.md` in its `resources`, after any global steerings. They are relative paths, so an agent loads the steering and skills of whichever workspace Kiro is started in (nothing is loaded when those folders do not exist). Skills use the `skill://` scheme, so only their metadata is loaded at startup and the rest on demand. A template that already lists an entry is not duplicated.

## Central env/mcp-versions.yaml

A flat map of MCP server name to pinned version/tag:

```yaml
jira-mcp: "1.4.2"
confluence-mcp: "2.1.0"
```

An agent template looks up its own entry via the `mcpVersion` template function, e.g. `{{ mcpVersion "jira-mcp" }}`. A template referencing an undefined key fails the render with a clear error naming the missing key and the offending file. Bumping a shared entry here and re-running `kirokit test` shows exactly which teams' rendered output changes as a result — including teams whose own templates don't reference the changed key, since `context.json` snapshots the *entire* resolved `env/mcp-versions.yaml`.

## Secrets via SOPS

A secret an agent's MCP server needs (a personal Jira/Confluence token, a warehouse credential, ...) never appears as plaintext in any template, rendered file, or golden fixture. Instead:

1. Each agent that needs a secret has its own encrypted `teams/<team>/agents/<agent>/env/secrets.enc.yaml`, a flat SOPS-encrypted YAML file (ciphertext, safe to commit).
2. That agent's `agent.json.tmpl` renders its MCP server's `command`/`args` as a wrapper around `sops exec-env`, using the `sopsCommand` / `sopsArgsJSON` template functions:

   ```
   "command": {{ sopsCommand }},
   "args": {{ sopsArgsJSON (printf "%s/kirokit-secrets/<team>/<agent>/secrets.enc.yaml" .KiroHome) "<real-mcp-command>" "<real-arg>" ... }}
   ```

   The secrets path passed to `sopsArgsJSON` must be `.KiroHome`-relative and absolute — never a bare repo-relative path — because it has to resolve wherever the MCP server actually launches, under the *installed* `$KIRO_HOME`, not the kit's source tree.
3. `sops exec-env <file> "<command>"` decrypts the file and launches the real command with its keys set as env vars, entirely at MCP-server-launch time — never at template-render time, and never through Kiro CLI's own (currently unreliable) `${VAR}` substitution.
4. `kirokit install` copies each agent's encrypted secrets file to `$KIRO_HOME/kirokit-secrets/<team>/<agent>/secrets.enc.yaml` so the baked-in path actually resolves at runtime.

**Remote (HTTP) MCP servers** such as the official Atlassian server (`https://mcp.atlassian.com/v2/mcp`) have no local binary, so they are launched through the pinned `mcp-remote` proxy under the same `sops exec-env` wrapper. The secret holds the full header value (`Bearer <key>` for a service-account API key, or `Basic <base64(email:token)>` for a personal API token) and the template passes `--header 'Authorization:${ATLASSIAN_AUTH}'`; the single quotes keep the placeholder literal in the rendered file, and `mcp-remote` expands it from the environment sops sets at launch. See `teams/platform-team/agents/devops/agent.json.tmpl`. An org admin must first enable the **Allow API token authentication** toggle in [Atlassian Administration](https://admin.atlassian.com/) (Rovo → Rovo MCP server → Authentication); without it every tool call fails with "You don't have permission to connect via API token".

To create the token, use the [MCP token page](https://id.atlassian.com/manage-profile/security/api-tokens?autofillToken&expiryDays=max&appId=mcp-v2&selectedScopes=all). The MCP server needs the "… via Agent Interface" scopes (e.g. `search:jira:agent-interface` for Jira search), which the generic API-token page did not offer. Then base64-encode `email:token` and store it with `Basic` in the secrets file. See Atlassian's [Configure authentication via API token](https://support.atlassian.com/atlassian-rovo-mcp-server/docs/configuring-authentication-via-api-token/) for the details and limitations. Jira Cloud also rejects unbounded JQL, so searches need a restriction such as `project = KEY`.

Decrypting requires a SOPS-compatible key (age, PGP, or KMS) that is a recipient in `.sops.yaml`. Each team has its own `creation_rules` entry in `.sops.yaml` naming that team's age public key(s); a team's key cannot decrypt another team's secrets. Private keys are never committed — keep them outside the repo, e.g. `~/.config/sops/age/<team>.txt`, and point SOPS at the right one:

```bash
SOPS_AGE_KEY_FILE=~/.config/sops/age/platform-team.txt sops teams/platform-team/agents/devops/env/secrets.enc.yaml   # edit
age-keygen -o ~/.config/sops/age/my-team.txt                                                                          # new team key
```

To add an engineer to a team, append their age public key (comma-separated) to that team's rule and run `sops updatekeys <file>` for each of the team's secrets files. The secret values committed here are placeholders; replace them with real tokens via `sops <file>`.

## Golden Tests

```bash
kirokit render [team] --out <dir>   # render only, never touches testdata/golden/
kirokit test [team]                 # render + diff against testdata/golden/<team>/ (CI entry point)
kirokit update [team]               # render + overwrite testdata/golden/<team>/
```

Omitting `[team]` processes every directory under `teams/`. `test` exits non-zero and lists every mismatched file when a team's templates have drifted from its committed goldens. `update` regenerates a team's `testdata/golden/<team>/rendered/` tree and `context.json` snapshot after an intentional template change — review the diff before committing it.

`go test ./internal/golden` also runs `kiro-cli agent validate` on every committed golden agent config, and on the installed `platform-team` agents with the committed `testdata/knowledgebase/` folder attached as `devops`'s knowledge base (the install-time rewrite no golden covers), so a config Kiro itself would reject fails the test even when the golden diff is clean. It skips when `kiro-cli` is not installed or not logged in (`validate` needs a login); set `KIROKIT_REQUIRE_KIRO_CLI=1` to make those skips failures. The complete agent config with that knowledge base attached is also compared to `testdata/golden/platform-team/installed-with-kb/agents/devops.json` (re-indented, with the scratch `KIRO_HOME` replaced by `/kiro-home`); regenerate it with `go test ./internal/golden -run KnowledgeBaseRenderedIntoAgentJSON -update-kb-golden` and review the diff. Note that `validate` exits 0 on error and reports on stderr, and it does not check permission rules, MCP server shape, or that `file://`/`skill://` resources exist.

Rendering always uses the fixed `KiroHome` sentinel `/kiro-home` for `render`/`test`/`update`, so golden fixtures are identical on every machine and in CI regardless of the real `$KIRO_HOME`; only `install` bakes in the real resolved value.

## Onboarding a New Team

```bash
kirokit onboard <team>
```

Scaffolds `teams/<team>/` from an embedded starter layout: a `manifest.yaml` with the given team name, one example agent (`agents/example-agent/`), and empty `hooks/`, `prompts/`, `skills/`, `steerings/` directories. Fails with a clear error if `teams/<team>/` already exists. Hand-copying another team's directory (e.g. `teams/platform-team/`) instead is equally valid.

## Install / Uninstall

```bash
kirokit install <team>
kirokit install <team> --kb-agent <agent> --kb-source <folder> --kb-mount <path>
kirokit uninstall <team>
```

`install` renders a team with its real `$KIRO_HOME` baked in (respecting the `KIRO_HOME` env var, falling back to Kiro CLI's own default `~/.kiro` when unset) and copies the rendered tree plus each agent's encrypted secrets file into that location — real, usable files where Kiro CLI actually reads them. `uninstall` removes exactly what the matching `install` wrote, tracked via an install manifest under `$KIRO_HOME/.kirokit-installed/<team>.json`, leaving every other team's installed files and anything Kiro CLI itself manages untouched.

The three `--kb-*` flags (all-or-nothing) attach a local `knowledgeBase`-type resource to one installed agent: `--kb-source` is a local folder, `--kb-mount` is where it gets symlinked (never copied), and `--kb-agent` is which installed agent's `resources` array gets the new entry. The entry is named `<agent>-kb` and sets `indexType: "best"` (semantic search), `include: ["**/*.md"]` (index only Markdown) and `autoUpdate: true` (re-index when the agent spawns). `include` is accepted by `kiro-cli` but not yet in Kiro's agent-config docs, and Kiro issues report that agent-config knowledge bases are not always indexed automatically, so check `/knowledge show` in a session. `uninstall` removes the symlink but never touches the real source folder's content.

## Git Hooks

[Lefthook](https://lefthook.dev) runs checks on every commit (`brew install lefthook gitleaks commitizen uv`, then once per clone):

```bash
lefthook install
```

- **pre-commit:** `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-json` (from [pre-commit-hooks](https://github.com/pre-commit/pre-commit-hooks), run via `uvx`) and `gitleaks` on the staged changes. The two fixers re-stage what they change but fail the commit, so run it again. They skip `testdata/golden/**` (byte-compared by `kirokit test`) and `*.enc.yaml` (SOPS MAC).
- **commit-msg:** `cz check` enforces Conventional Commits, which `semantic-release` needs. Use `cz commit` for an interactive prompt.

Run everything on demand with `lefthook run pre-commit --all-files`.

## Releases

Releases follow Conventional Commits on `main`:

1. `semantic-release` (configured in `.releaserc.yml`) analyzes commits, computes the next version, generates a changelog, and creates the git tag.
2. Its `@semantic-release/exec` plugin invokes `goreleaser release` (configured in `.goreleaser.yaml`), which builds and publishes the `kirokit` binary for **darwin and linux only** (amd64 + arm64) as GitHub release archives attached to the tag semantic-release just created.
3. `.github/workflows/ci.yaml` runs `go build`, `go vet`, `golangci-lint`, and `go test ./...` (which includes every team's golden-file comparisons) on every pull request and push to `main` — a template change that breaks a golden fixture fails CI.
4. `.github/workflows/release.yaml` runs the release step on every push to `main`.

The first live release requires this repository to be pushed to `github.com/mavogel/kiro-cli-kit` with a real git remote; until then, `goreleaser check` and a `goreleaser release --snapshot --clean` locally, plus `npx semantic-release --dry-run`, are how the release configuration is validated.
