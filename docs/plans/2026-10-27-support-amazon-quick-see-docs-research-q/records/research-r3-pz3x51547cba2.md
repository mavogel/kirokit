---
type: research
---

# Golden harness comparison and what Kiro IDE 1.4.26 reads from the Kiro home

Bundle path used below: `K` = `/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js` (one 14.7 MB minified file; the cited number is the physical line, and the same lines hold many statements, so each claim also names the minified function). Everything about the IDE is read from that installed bundle (fact) plus kiro.dev docs; nothing was run inside the IDE (gap).

## Findings

### The harness compares two things per team: the `rendered/` tree byte-for-byte and `context.json` modulo surrounding whitespace
**Answers:** 5

- `kirokit test [team]` and `kirokit update [team]` build the same `golden.Inputs` from four fixed relative paths (`teams`, `global/steerings`, `env/mcp-versions.yaml`, `testdata/golden`), prefixed by `KIROKIT_REPO` when set (`cmd/kirokit/commands/root.go:16-19`, `cmd/kirokit/commands/root.go:26-31`; used at `cmd/kirokit/commands/test.go:25-30`, `cmd/kirokit/commands/update.go:25-30`).
- With no argument both commands iterate every directory under `teams/` (`cmd/kirokit/commands/root.go:74-88`); with one argument the name is used as given with no existence check at that point.
- Compared tree: the team is rendered into a fresh `os.MkdirTemp` scratch dir under `<scratch>/rendered` (`internal/golden/golden.go:101-108`) and diffed against `testdata/golden/<team>/rendered` (`internal/golden/golden.go:36-40`). `diffDirs` lists regular files on both sides (directories are skipped, empty directories are not compared, `internal/golden/golden.go:160-180`) and reports three diff kinds: `extra file (not in golden)`, `content differs` (`bytes.Equal`, no normalisation), `missing file (present in golden, not rendered)` (`internal/golden/golden.go:135-156`). A missing golden `rendered/` dir is treated as empty, so every rendered file shows as extra (`internal/golden/golden.go:122-129`).
- Rendering uses `kiroHome == ""`, which `BuildContext` turns into the sentinel `/kiro-home` (`internal/golden/golden.go:91`, `internal/golden/golden.go:108`, `internal/render/context.go:16`, `internal/render/context.go:58-60`).
- `context.json` snapshot: the `*render.Context` returned by `BuildContext` is `json.MarshalIndent`ed with two-space indent and a trailing newline (`internal/golden/golden.go:91-99`). Fields tagged `json:"-"` are omitted: `LaunchShell`, `Agents`, `Skills` (`internal/render/context.go:26`, `:31-32`). What is serialised is `Team` (the manifest), `Env` (the MCP version pins), `Global` (opted-in global steering slug to file content) and `KiroHome`; the committed example is `testdata/golden/data-team/context.json` (`Team`, `Env`, `Global`, `KiroHome: "/kiro-home"`). Comparison trims leading/trailing whitespace on both sides, then requires equality; a mismatch reports only `context.json differs from <path>` (no content diff), and a missing file reports `missing golden context file` (`internal/golden/golden.go:45-53`).
- Result handling: `Compare` sorts the diff strings (`internal/golden/golden.go:55`); `test` prints `<team>: ok` or `<team>: golden mismatch:` plus indented diffs to stderr, continues through all teams, and returns one error at the end if any mismatched (`cmd/kirokit/commands/test.go:32-51`). A render error aborts at that team with `kirokit test: <team>: <err>` (`cmd/kirokit/commands/test.go:34-37`).

### `kirokit update` deletes and rewrites only `rendered/` and `context.json` of the named teams
**Answers:** 5

- `Update` renders as above, then `os.RemoveAll(testdata/golden/<team>/rendered)`, copies the scratch tree in with `copyDir` (files written mode `0o600`, dirs `0o755`), `MkdirAll`s the team dir, and overwrites `context.json` with the marshalled bytes (mode `0o600`) (`internal/golden/golden.go:67-83`, `internal/golden/golden.go:190-212`).
- Anything else under `testdata/golden/<team>/` survives. Observed: `testdata/golden/platform-team/installed-with-kb/` sits beside `rendered/` and `context.json`; it is a separate fixture owned by `internal/golden/knowledgebase_test.go` (golden path built at `internal/golden/knowledgebase_test.go:66`, its own regenerate flag referenced at `:17`), not by `update`. `ls testdata/golden/*` shows three team dirs (data, platform, product) matching `teams/*`.
- `update` prints `updated golden fixtures for <team>` per team (`cmd/kirokit/commands/update.go:36`). It stops at the first failing team (`cmd/kirokit/commands/update.go:33-35`), leaving earlier teams already rewritten. A team whose render fails is not partially written because render happens before `RemoveAll` (`internal/golden/golden.go:61-71`).
- Teams are 29 git-tracked golden files in total (`git ls-files testdata/golden | wc -l`).

### Golden agent JSON uses flat `agents/<name>.json` with `permissions` and no CLI-2.x fields
**Answers:** 5, 6

- The data-team golden agent `testdata/golden/data-team/rendered/agents/godev.json` is a flat file with `name`, `description`, `prompt: "file://../prompts/pipeline-status.md"`, `tools`, `permissions.rules`, `resources` (including `skill:///kiro-home/skills/sql-helper/SKILL.md`, an absolute path using the sentinel) and `mcpServers`; it contains neither `allowedTools` nor `toolsSettings`.
- Rendered output layout (data-team): `agents/`, `hooks/`, `prompts/`, `skills/`, `steering/` (`ls testdata/golden/data-team/rendered`).

### Kiro IDE 1.4.26 honours `KIRO_HOME`, with three rejection rules
**Answers:** 6

- The single `KIRO_HOME` reference in the IDE bundle is where the agent engine resolves its user config dir: `MGp({kiroHome: process.env.KIRO_HOME, homeDir})` (`K:16686`, constructor of the service that owns `userConfigDir`). `MGp` returns `<home>/.kiro` (`_T(homeDir)`, `K:1774`, `rd=".kiro"`) when the variable is empty/blank; otherwise `path.resolve(KIRO_HOME)` with source `"env"` (`K:16686`, function `MGp`, found at char offset 10409319 of the file; the function body is on `K:16686` as well).
- It ignores the variable (falls back to `~/.kiro`, logs `agent.userConfigDir.ignored`) when the value is relative, is a filesystem root, or contains the home directory (reason strings `relative`, `root`, `contains-home`; function `DGp`, same line region). Observed consequence: `KIRO_HOME=$HOME` or a relative path is not honoured by the IDE, while Kiro CLI/kirokit's `ResolveKiroHome` returns any non-empty value unvalidated (`internal/install/install.go:47-58`).
- Every home-scoped reader below receives this `userConfigDir` rather than calling `homedir()` itself (inference from the constructors cited per item). One exception was found: a helper `tF()` computes `os.homedir()/.kiro` directly with no `KIRO_HOME` (`K:16776`, `getHomeKiroPath`) and is used by `rF()` to locate `settings/mcp.json`; which feature calls `rF` was not traced (gap).
- Not verified by running the IDE with `KIRO_HOME` set (gap). `~/.kiro` on this machine contains `agents/`, `skills/`, `steering/`, `prompts/`, `hooks/`, `settings/` (only `cli.json` was seen in settings), plus `crew/`, `powers/`, `extensions/`, `kirokit-secrets/`, `kirokit-steerings/` (`ls ~/.kiro`).

### The IDE reads flat `agents/<name>.json` and `.md`, recursively, but skips files that carry only CLI 2.x trust fields
**Answers:** 6

- Loader `Sxs` (profile loader) is created over `<userConfigDir>/agents` (`K:16686` cache construction via `agentProfileCache.get(join(userConfigDir, "agents"))`; constant `E6="agents"` at `K:1774`; loader at `K:16392`).
- File discovery `wxs` walks the directory recursively, follows symlinks (with cycle guard), accepts any file ending `.md` or `.json`, sorted by name, and uses the path relative to the agents dir without extension as the fallback id, so `agents/godev.json` becomes id `godev` (`K:16392`, function `wxs`).
- `.json` files are parsed with comment/trailing-comma tolerance and a zod object schema with fields `name, description, prompt, tools, excludedTools, model, effortLevel, includeMcpJson, includePowers, mcpServers, resources, permissions, welcomeMessage, dispatchKind, hooks` (`K:2607`, schema `Nxi`; unknown keys are not rejected by the schema, inference from zod object default). The id is `name` when non-empty, else the filename (`K:16392`, `e_p`, and `h=g.frontMatter.name??p`).
- CLI-only rule: a JSON agent is skipped with reason `cli_only_agent` when it contains `allowedTools` or `toolsSettings` and has no `permissions` key (`K:2607`, `Rxi=["allowedTools","toolsSettings"]`, `wjr=["permissions"]`; skip logic `K:16392`). This matches the docs: "A CLI 2.x JSON profile containing `allowedTools` without a `permissions` block is skipped by V3" and `toolsSettings` needs a `permissions` block (kiro.dev/docs/cli/custom-agents/configuration-reference/).
- `prompt: "file://..."` is resolved relative to the realpath of the agent file's dir, absolute paths and `~` are refused, and the resolved realpath must stay inside a containment root (the home dir when `userConfigDir` is under home, otherwise `userConfigDir`) (`K:16392`, `vxs`; root choice `K:16686`). The docs say absolute `file://` URIs are allowed; the installed IDE code rejects them for the prompt field (conflict: the code wins at runtime for IDE 1.4.26; the CLI's behaviour was not tested).
- `resources` entries are parsed into file, skill and knowledge-base resources (`Tae.parse`, `K:2607`); invalid entries are dropped silently (debug log `custom-agent.resource.invalid-dropped`).
- A file watcher on `.kiro/agents` triggers reload (`K:16686`, `watchPaths:[join(".kiro","agents")]`).

### Skills, steering, hooks, prompts and MCP settings under the Kiro home
**Answers:** 6

| Kind | Path under home (`userConfigDir`) | What the IDE does | Evidence |
|---|---|---|---|
| Skills | `skills/<dir>/SKILL.md` | One level of subdirectories (or symlinks to them); each `SKILL.md` parsed; missing file is a warning | `K:16261` `scanSkillsDirectory`/`parseSkillMd`; `p3="skills"`, `Poi="SKILL.md"` at `K:1774` |
| Steering | `steering/` (global, scanned for steering docs; `AGENTS.md` in it too per docs) | Global steering dirs = cloud bases plus `<home>/steering`; on-demand steering requires front matter `inclusion: auto` with a `description` | `K:16261` `globalSteeringDirs`, `parseOnDemandSteeringFile`; docs kiro.dev/docs/steering/ list `always`, `fileMatch`, `manual`, `auto` modes |
| Hooks | `hooks/*.json` | Standalone hook files; only `*.json` in that directory; v2 hooks execution is gated by a feature flag and a trusted workspace | `K:3936` `homeGlobalHooksDir`, `Led=".json"`, `RHr`; docs say standalone `.kiro/hooks/*.json` |
| Prompts | `prompts/` | A watcher on `.kiro/prompts` (home plus workspaces) fires a change callback; the consumer of the files was not traced | `K:16261` (`watchPaths:[join(".kiro","prompts")]`) |
| MCP servers | `settings/mcp.json` | Read and merged with workspace `.kiro/settings/mcp.json` by the MCP config class; global entries are included by agents only when `includeMcpJson` is true (field default `false` in the schema) | `K:14798` (`otn="mcp.json"`, `userConfigBase`), `K:2607` schema default; docs `includeMcpJson` section |
| Agents | `agents/` | see previous finding | above |

- Legacy `*.kiro.hook` files are only read from workspace folders, not the home (`K:` hook manager around char offset 14220966, `_getHooksDirectories` uses `workspace.workspaceFolders`).
- The Kiro CLI `skill://` resource scheme and `file://.kiro/steering/**/*.md` globs in an agent resolve relative to the session workspace; the kirokit goldens use both a `/kiro-home` absolute skill path and workspace globs (`testdata/golden/data-team/rendered/agents/godev.json`).

### Conflicts and sources of truth
**Answers:** 6

- Docs vs code: docs state absolute `file://` prompt paths are used as-is; the IDE loader rejects them (above). The bundle in `/Applications/Kiro.app` (version `1.4.26`, `defaults read .../Info CFBundleShortVersionString`) is what runs locally.
- Docs describe steering at `~/.kiro/steering/` and say nothing about `KIRO_HOME`; only the bundle shows the override and its restrictions.

## How it is tested today

- Golden package (`internal/golden`): fixture-based unit tests build a temp repo layout via `setupFixture` and `mustWrite` (`internal/golden/golden_test.go:11-36`) and cover Compare after Update, detection of a changed template and Update regeneration (`internal/golden/golden_test.go:38`, `:54`, `:73`). `kirocli_test.go` runs `kiro-cli agent validate` over committed goldens and skips unless a logged-in `kiro-cli` exists or `KIROKIT_REQUIRE_KIRO_CLI=1` (`internal/golden/kirocli_test.go:14-60`); `knowledgebase_test.go` compares an installed-agent golden and skips off macOS/Linux (`internal/golden/knowledgebase_test.go:51-69`).
- Commands: `cmd/kirokit/commands/root_test.go` runs `test` against the real repo goldens through `KIROKIT_REPO` (`:110-122`) and requires every team under `teams/` to report `<team>: ok` (`:126-145`). There is no dedicated test file for `update.go` or `test.go`.
- Commands, from the repository root: one file's package `go test ./internal/golden -run TestCompare_ReportsNoDiffsAfterUpdate`; whole suite `go test ./...` (includes every team's golden comparison); CLI golden diff only: `./kirokit test` after `task build`; full CI mirror `task ci` (AGENTS.md). Observed: `go test ./internal/golden ./cmd/kirokit/commands -run 'Golden|Compare|Update|Kirokit'` printed `Go test: 25 passed in 2 packages`.
- Change frequency (90 days): 5 commits touch `internal/golden`, `test.go`, `update.go` together (`git log --oneline --since=90.days -- internal/golden cmd/kirokit/commands/test.go cmd/kirokit/commands/update.go`).
- IDE behaviour has no automated test in this repo; the findings above come from reading the installed bundle only.

## Open questions

- Question 6: Whether the IDE's global `prompts/` directory content is consumed by any feature (only a file watcher on `.kiro/prompts` was found, `K:16261`); not traced further.
- Question 6: Which feature calls `rF()`/`tF()` (the `settings/mcp.json` locator that ignores `KIRO_HOME`, `K:16776`) and therefore whether that code path is live in 1.4.26.
- Question 6: Runtime confirmation (IDE launched with `KIRO_HOME` set) was not done; findings are from static reading of the bundle.

## Code references

Exhaustive for the two harness files and the three command files; selective for the IDE bundle (searched for `KIRO_HOME`, `userConfigDir`, agent/skill/steering/hook/prompt/mcp loaders).

- Harness: `internal/golden/golden.go:19-25` (Inputs), `:29-57` (Compare), `:60-84` (Update), `:90-113` (renderForGolden), `:117-158` (diffDirs), `:160-212` (helpers)
- Commands: `cmd/kirokit/commands/test.go:10-55`, `cmd/kirokit/commands/update.go:10-42`, `cmd/kirokit/commands/root.go:16-31`, `:74-88`
- Context: `internal/render/context.go:16-33` (Context, json tags), `:37-71` (BuildContext)
- Kiro home resolution (kirokit side): `internal/install/install.go:47-58`
- Goldens: `testdata/golden/<team>/context.json`, `testdata/golden/<team>/rendered/**`, `testdata/golden/platform-team/installed-with-kb/`
- Tests: `internal/golden/golden_test.go`, `internal/golden/kirocli_test.go`, `internal/golden/knowledgebase_test.go`, `cmd/kirokit/commands/root_test.go:110-145`
- Kiro IDE (installed, outside repo): `/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js` lines 1774 (path constants), 2607 (agent schema, CLI-only fields), 3936 (hooks), 14798 (MCP config), 16032, 16261 (skills, steering, prompts watcher), 16392 (agent loader), 16686 (`KIRO_HOME`), 16776 (`tF`)
- Docs: https://kiro.dev/docs/cli/custom-agents/configuration-reference/, https://kiro.dev/docs/steering/
