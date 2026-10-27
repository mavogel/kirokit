---
type: research
status: approved
approved: true
review_digest: 3ac0eff33fa52f8e94a65c00ca75a6793a341534b2604a8cb611a60d0e82aa61
review_basis: 3c5c78ae2d8f2b806a04091b43838de275fe992c620cb3acfbfac19c346c9bac
---

# Research: Support Quick, Kiro IDE and kirocrew; README refresh

## Summary

kirokit's render is one pipeline per team: it walks the team folder, runs every `.tmpl` file through Go templates, copies everything else byte for byte, adds steering and skill entries to each agent, checks the agent JSON shallowly, and writes one tree. Where a file lands is decided in one place, and an unknown top-level folder is an error. Vendored external skills are the only path that never touches the template engine, so text like `{{name}}` survives only there or in a non-`.tmpl` file. A team `SKILL.md.tmpl` containing `{{name}}` fails to parse.

The manifest is read leniently: unknown keys are silently ignored, and only `team` is required. Agent JSON is checked even more loosely. kirokit's own validator looks only at `mcpServers` and `resources`, and `kiro-cli agent validate` accepted an agent with an extra unknown key without complaint. The golden harness compares only `rendered/` and `context.json` per team. Anything else under a team's golden folder is left alone, which already happens for one fixture.

Kiro IDE 1.4.26 and kirocrew 0.10.0 both honour `KIRO_HOME`, with a few rules. Both read the same flat `agents/<name>.json` files kirokit installs, plus `skills/<name>/SKILL.md`. The IDE also reads `steering/`, `hooks/*.json` and `settings/mcp.json`. The IDE skips an agent only when it has CLI-2.x fields without a `permissions` block. Every kirokit agent has `permissions`, so none is skipped by that rule. kirocrew lists kirokit's agents as foreign agents, never rewrites them, and creates its own "skill view" copies beside them in the same `agents/` folder. Each tool reads `settings/mcp.json` from the plain home folder in some code paths, ignoring `KIRO_HOME` there. Nothing about either tool was run; all of it comes from reading the installed apps.

The README showcase is a scripted recording: a tmux terminal for the CLI and headless Chrome for the desktop app, joined with captions and checked for privacy, size and a 30 to 60 second length. No test reads the README. Other files refer to its sections by name rather than by link.

## What the goal puts at risk

- **Reuse:** `mapOutputPath` and the render walk (`internal/render/render.go:352-391`, `:76-155`) decide every output path, and the vendored-copy path (`internal/render/externalskills.go:24-69`) is the existing way to carry `{{ }}` text unchanged. The golden harness (`internal/golden/golden.go:29-113`) and `kirokit update` decide what CI compares.
- **Churn (90 days):** `internal/render` 10 commits, `internal/golden` with `test.go`/`update.go` 5, `internal/manifest` 4, `teams/*/agents` 7, `scripts/showcase` 2, `README.md` 28. The README and render code move the most.
- **Hard to change:** the manifest struct is serialised into every team's `context.json` (`internal/manifest/manifest.go:18-29`), so a new field that is always emitted changes every golden. Agent template edits change committed goldens byte for byte.
- **Lenient parsing hides typos:** neither the manifest (`internal/manifest/manifest.go:148-157`) nor kirokit's agent check (`internal/kiroconfig/validate.go:33-60`) rejects unknown keys. A misspelled new field would quietly do nothing.
- **Look here:** unsure · Kiro IDE resolves `prompt: file://` relative to the agent file and refuses absolute paths (`/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js:16392`, installed Kiro IDE app, not this repository). kirokit's installed agents use relative prompt paths, but also absolute `file://` and `skill://` resources whose IDE resolution was not traced.
- **Look here:** risky · kirocrew writes its own `kirocrew-skill-view-*.json` agent files into the same `agents/` folder kirokit installs to (the installed KiroCrew.app (not this repository) `kiro_crew/acp/skill_projection.py:3773-3797`). kirokit's install and sync record only their own files, so these are never removed by uninstall. kirokit's backup of `agents/` also copies them.
- **Look here:** risky · the showcase has a hard 30 to 60 second length check and runs with a throwaway `HOME` (`scripts/showcase/checks.sh:8-60`, `scripts/showcase/record.sh:125-132`). Two more apps, Kiro IDE (Electron) and kirocrew (a gateway with a web dashboard on port 5476), have no driver today and may need a signed-in profile.
- **Depends on, not covered by the questions:** Quick's own behaviour for agent import and skill ids comes only from the research brief (`docs/research/quick/QUICK-TARGET-BRIEF.md` sections 2.4, 4.3, 5). Earlier README decisions from project memory: the GIF sits at the top under the intro, the MP4 is kept as the sharp master, and the story is a CLI install, then the agent in Kiro, then the same in the app.

## Open questions

Question 6: How Kiro IDE resolves absolute `file://` and `skill://` resources from an installed agent (it stores them as plain paths, `/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js:2607` (installed Kiro IDE app, not this repository) `cqu`), whether its global `prompts/` folder content is used, and which feature reads `settings/mcp.json` from the plain home folder. A runtime check in the IDE is needed.
Question 7: Which `settings/mcp.json` kirocrew uses when `KIRO_HOME` is set depends on the module (the installed KiroCrew.app (not this repository) `kiro_crew/mcp_cleanup.py:249-251` follows it, the installed KiroCrew.app (not this repository) `kiro_crew/dashboard/handlers/mcp.py:120` does not). Nothing in kirocrew was run.
Question 11: Whether pages under `docs/` link to README anchors was not searched; today `docs/` holds only `prd/`, `plans/` and `research/`.

## Findings

Read at 9bfb355.

### `kiro-cli agent validate` accepts an agent JSON with an unknown top-level key
**Answers:** 3

An unchanged golden agent (`testdata/golden/product-team/rendered/agents/qa.json`) passed `kiro-cli agent validate --path` silently. A copy with an extra top-level key `quickWelcome` also passed silently, with no error or warning (run in this Research step with kiro-cli 2.34.2). The Kiro IDE agent schema is a zod object over known fields that does not reject extra keys (`/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js:2607`, `Sjr`/`Nxi`; installed Kiro IDE app, not this repository). kirocrew coerces foreign fields rather than rejecting them (the installed KiroCrew.app (not this repository) `kiro_crew/agent_discovery.py:234-251`).

**Tested today:** no repository test feeds either validator an unknown key (`internal/kiroconfig/validate_test.go:5-73`, `internal/golden/kirocli_test.go:60-78`).

### Kiro IDE splits agent resources by scheme and keeps the path as written
**Answers:** 6

`Tae.parse` sends each string resource through `cqu`. `file://` and `skill://` prefixes are stripped and the rest is kept as a path. Another scheme logs "Unsupported resource URI scheme" and a bare string logs "Invalid resource URI". Object entries become knowledge-base resources (`/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js:2607` (installed Kiro IDE app, not this repository), `cqu`, `lqu`). Entries that fail the schema are dropped with the debug log `custom-agent.resource.invalid-dropped` (`/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js:2607` (installed Kiro IDE app, not this repository), `Oxi`). How the stored paths are later resolved was not traced.

**Tested today:** nothing in this repository runs Kiro IDE.

<!-- ql:folded -->
### `render` walks teams, calls `RenderTeam` per team, and only chooses the output root
**Answers:** 1

- `--out` is required (`cmd/kirokit/commands/render.go:18-20`). With no team argument, every directory under `teams/` is rendered into `<out>/<team>` (`cmd/kirokit/commands/render.go:29-33`, team list from `cmd/kirokit/commands/root.go:74-89`); with a team argument the tree goes directly into `<out>`.
- It calls `render.RenderTeam(...)` with `kiroHome` = `""` (`cmd/kirokit/commands/render.go:34`), which becomes `RenderTeamWith` with POSIX launch shell (`internal/render/render.go:40-42`). An empty KiroHome becomes the sentinel `/kiro-home` (`internal/render/context.go:58-60`, constant at `:16`).

```
render cmd (cmd/kirokit/commands/render.go)
 └ RenderTeam -> RenderTeamWith (internal/render/render.go:45)
    ├ BuildContext: manifest + env pins + opted-in global steerings
    ├ externalskills.Resolve: verify lock/vendored dirs, return names
    ├ teamInventory: ctx.Agents, ctx.Skills
    ├ checkOwnerKeys
    ├ agentResources = global steering URIs + external skill URIs + workspace URIs
    ├ WalkDir(teamDir): map path, render or copy, add resources, validate, write to tmpDir
    ├ writeGlobalSteerings, writeExternalSkills (into tmpDir)
    └ publish(tmpDir -> outDir)
```

### Context is built from manifest, central version pins and opted-in global steerings
**Answers:** 1

- `BuildContext` loads `teams/<team>/manifest.yaml`, `env/mcp-versions.yaml`, and reads `<globalRoot>/<slug>.md` for every slug in the manifest's `globalSteerings`; an unreadable one fails the render (`internal/render/context.go:37-56`).
- `Context` holds Team, Env, Global (slug to content), KiroHome; `LaunchShell`, `Agents`, `Skills` are `json:"-"` and filled by render (`internal/render/context.go:19-33`). Template functions: `mcpVersion`, `globalSteering`, `jsonString`, `sopsCommand`, `sopsArgsJSON`, `owners`, `feedbackRepo` (`internal/render/funcs.go:15-23`). An undefined mcpVersion key errors (`internal/render/funcs.go:69`).

### Templated vs verbatim: the `.tmpl` suffix decides, for every file in the team folder
**Answers:** 1

- The team directory is walked in full; `manifest.yaml` and `.gitkeep` are skipped (`internal/render/render.go:76-91`).
- A file whose path ends in `.tmpl` is parsed with `text/template` and executed against the Context (`internal/render/render.go:107-108`, `:212-229`). Any other file is read and written byte for byte (`internal/render/render.go:112-116`). So a non-`.tmpl` file inside `skills/<skill>/` (a script, a reference doc) is copied verbatim; a `.tmpl` file there is rendered and loses the suffix.
- Output paths are decided only by `mapOutputPath` (`internal/render/render.go:352-391`):

| Source (under `teams/<team>/`) | Output | Evidence |
|---|---|---|
| `agents/<a>/agent.json[.tmpl]` | `agents/<a>.json` | `render.go:370-371` |
| `agents/<a>/env/*` | skipped (consumed by install / sops func) | `render.go:364-368` |
| any other path under `agents/<a>/` | error | `render.go:373` |
| `hooks/**`, `skills/**`, `prompts/**` | same relative path, `.tmpl` stripped | `render.go:380-381` |
| `steerings/**` | `steering/<rest>` (renamed dir), `.tmpl` stripped | `render.go:383-386` |
| `external/**` | skipped (copied later by `writeExternalSkills`) | `render.go:375-378` |
| any other top-level folder | error "unrecognized top-level category" | `render.go:388-389` |

- Rendering goes to a temp dir `.kirokit-render-*` beside the output, written with mode 0600 (`render.go:70`, `:139`), then `publish` moves an existing differing outDir aside to a timestamped backup, leaves an identical one untouched, and renames the temp dir into place (`render.go:164-205`).
- Validation: `agents/*` outputs go through `kiroconfig.ValidateAgentJSON`, `hooks/*` through `ValidateHookJSON` (`render.go:337-346`); agent outputs also pass `checkSubagents` (`render.go:129-133`). Markdown, prompts, skills, steerings are not validated here.

### Agent `resources` are the template's own list, then render appends three groups
**Answers:** 1

- For every output under `agents/`, `addResources` is applied after templating and before validation (`internal/render/render.go:119-124`). The appended list is built once: global steering URIs, then external skill URIs, then the workspace entries (`render.go:64-65`).
  - Global steering: `file://<KiroHome>/kirokit-steerings/<team>/<slug>.md` per manifest slug, in manifest order (`internal/render/globalsteerings.go:19`, `:33-35`, `:45-51`).
  - External skills: `skill://<KiroHome>/skills/<name>/SKILL.md` per resolved name (`internal/render/externalskills.go:16-22`).
  - Workspace: `file://.kiro/steering/**/*.md` and `skill://.kiro/skills/**/SKILL.md` (`globalsteerings.go:27-30`).
- `addResources` edits the JSON text in place, preserving key order and indentation. It appends URIs not already present to an existing `"resources"` array (deduplicated), or creates the array when absent; a non-object config or non-array `resources` is an error (`globalsteerings.go:76-151`).
- Team-authored skills are not appended: the agent template lists them itself, e.g. `teams/data-team/agents/godev/agent.json.tmpl:19-21` has `"skill://{{ .KiroHome }}/skills/sql-helper/SKILL.md"`, rendered to `skill:///kiro-home/skills/sql-helper/SKILL.md` followed by the two workspace entries (`testdata/golden/data-team/rendered/agents/godev.json:19-23`).
- The global steering files themselves are written to `kirokit-steerings/<team>/<slug>.md` in the output after the walk (`globalsteerings.go:55-69`, called at `render.go:147`). Names must be a single path element (`globalsteerings.go:39-41`).
- `teamInventory` takes agent and skill names from immediate subdirectories of `agents/` and `skills/`, appends vendored names, sorts and compacts (`render.go:243-270`); `checkOwnerKeys` rejects manifest owners naming a missing agent/skill (`render.go:275-291`).

### Team skills are `skills/<skill>/SKILL.md.tmpl`, templated, with name/description frontmatter
**Answers:** 2

- Team skills pass through the generic `.tmpl` path above: `skills/<skill>/SKILL.md.tmpl` is executed as a Go template and written as `skills/<skill>/SKILL.md` (`render.go:107-108`, `:380-381`). Nothing in the render code reads or checks skill frontmatter.
- The four existing team skills carry a YAML frontmatter block with exactly `name` and `description` keys (`teams/data-team/skills/sql-helper/SKILL.md.tmpl:1-4`; same shape in `teams/platform-team/skills/jira-search/SKILL.md.tmpl:1-4`, `confluence-lookup`, and the rendered golden `testdata/golden/data-team/rendered/skills/sql-helper/SKILL.md:1-4`). The skills are: data-team/sql-helper, platform-team/confluence-lookup, platform-team/jira-search, product-team/analyze-session (`find teams`).
- Because the file is a Go template, a literal `{{name}}` in a team `.tmpl` skill is parsed as a template action; `name` is not in the FuncMap (`funcs.go:15-23`), so `template.Parse` would fail with a "parse" error (`render.go:217-220`). Inference from the code; I did not run it.

### Vendored external skills are copied byte for byte and never touch the template engine
**Answers:** 2

- Declared in `manifest.yaml` `externalSkills`; vendored into `teams/<team>/external/skills/<name>/` with `external/skills.lock.yaml` (`internal/externalskills/lock.go:19-27`). The walk skips `external/` (`render.go:375-378`).
- Before rendering, `externalskills.Resolve` compares manifest entries with the lock and the vendored folders: missing lock entry, source/ref/path mismatch, duplicate claims, undeclared vendored dirs, missing folder or `SKILL.md`, or a SHA256 mismatch each produce an error that says to run `kirokit skills fetch <team>` (`internal/externalskills/verify.go:28-91`; hash check at `:84-89`). On success it returns the sorted names (`verify.go:26-31` of the function body).
- `writeExternalSkills` copies each file of `external/skills/<name>/` to `skills/<name>/` in the output with `os.ReadFile`/`os.WriteFile`; doc comment: "never run through the template engine: a vendored file may be named *.tmpl or contain {{ }} and must arrive intact" (`internal/render/externalskills.go:24-27`, `:57-69`). Mode is 0600, or 0700 if the source has the owner exec bit (`externalskills.go:65-68`). Non-regular files are an error (`:50-52`). A name that is already an output `skills/<name>` (a team skill of the same name) is an error (`:33-36`). Invalid names are rejected via `plainName` (`:30-32`).
- Frontmatter of vendored skills is whatever the source repo has: render does not parse it. The only structural requirement is a `SKILL.md` at the skill root (`internal/externalskills/fetch.go:163-164`, `verify.go:82-83`). The test fixture uses `name` and `description` (`internal/render/externalskills_test.go:38`).
- Path that keeps `{{name}}` untouched: the vendored path. A `{{ ... }}` in a vendored script and a vendored `notes.md.tmpl` are asserted to arrive as identical bytes, with the `.tmpl` name retained (`internal/render/externalskills_test.go:39-40`, `:64-79`). A non-`.tmpl` file inside a team `skills/` folder is also copied unchanged (`render.go:112-116`) but a team `SKILL.md.tmpl` is not.
- Each vendored skill is wired into every agent as `skill://<KiroHome>/skills/<name>/SKILL.md`, exactly once even if the agent template lists it (`render.go:64`; test `externalskills_test.go:96-116`).

### An agent template is a JSON file with name, description, prompt, tools, permissions, resources, and optionally toolsSettings and mcpServers
**Answers:** 3

All nine templates (`teams/*/agents/*/agent.json.tmpl`: data-team godev, javadev; platform-team devops, platformdev; product-team customer, engineer, product-owner, qa, ux-designer) carry `name`, `description`, `prompt`, `tools`, `permissions`, `resources` (counted with grep over the top-level keys). `toolsSettings` appears in the 5 product-team agents, `mcpServers` in 4 (godev, javadev, devops, platformdev). No template contains a `welcomeMessage`/`welcome` key, nor `model`, `hooks`, `allowedTools` or `includeMcpJson` (grep of `teams/*/agents/*/agent.json.tmpl`). Field shapes, from `teams/platform-team/agents/devops/agent.json.tmpl`:

- `description`: one-sentence string (line 3).
- `prompt`: either a `file://../prompts/<name>.md` reference (line 4; also godev, javadev, platformdev) or an inline string with `\n` escapes (`teams/product-team/agents/qa/agent.json.tmpl:4`).
- `tools`: string array of built-ins (`read`, `write`, `glob`, `knowledge`, `shell`, `subagent`) and `@server` MCP references (`devops/agent.json.tmpl:5`).
- `permissions.rules[]`: objects `{capability, match[], effect}` (`devops/agent.json.tmpl:6-17`).
- `resources`: array of URIs; skills as `skill://{{ .KiroHome }}/skills/<name>/SKILL.md` (`devops/agent.json.tmpl:18-20`); product-team also lists `file://{{ .KiroHome }}/steering/spec-review-gates.md` (`teams/product-team/agents/qa/agent.json.tmpl:20`).
- `mcpServers.<name>`: `command`, `args`, `timeout`. `atlassian` uses template functions `sopsCommand`, `sopsArgsJSON` and `mcpVersion` (`devops/agent.json.tmpl:22-26`); `context7` uses `npx` with `mcpVersion "context7-mcp" | jsonString` (`devops/agent.json.tmpl:27-31`).
- `toolsSettings.subagent.availableAgents/trustedAgents` (`qa/agent.json.tmpl:6-11`).

There is no per-agent `skills` key: skills reach an agent only as `skill://` entries in `resources`. Render appends to every agent's `resources`, in order: opted-in global steerings, vendored external skills, then two workspace globs (`internal/render/render.go:64-65`, applied at `internal/render/render.go:119-120` via `addResources`; globs defined at `internal/render/globalsteerings.go:27-30`). `addResources` edits the JSON text in place, skips URIs already present, and creates `resources` if missing (`internal/render/globalsteerings.go:71-111`).

### Kirokit's own validator is shallow and ignores every key except mcpServers and resources
**Answers:** 3

`ValidateAgentJSON` (`internal/kiroconfig/validate.go:33`) unmarshals into a struct with only `mcpServers` and `resources` (lines 34-37). Go's `encoding/json` drops unknown keys silently, and no `DisallowUnknownFields` exists in the file, so any other key (misspelled, new, or unsupported by Kiro) passes. It checks: JSON validity (38-40); each `mcpServers` entry has a `command` or a `url` pointer non-nil, else `must have either "command" or "url"` (50-52); `resources`, if present, is a JSON array (55-60). Fields `name`, `description`, `prompt`, `tools`, `permissions` are not type-checked. The package doc states the shallowness is intentional: "structural shape only — since golden files remain the primary correctness signal" (`internal/kiroconfig/validate.go:3-5`). `ValidateHookJSON` (68-99) is stricter for hook files: requires `version`, and each hook a `name`, a trigger in an 11-name set (15-27) and `action.type` of `command` or `agent`; it is unused by agent templates.

Call sites: only `validateOutput` in render, for outputs under `agents/` (`internal/render/render.go:337-346`), invoked on the rendered content after `addResources` (`internal/render/render.go:~126`). A second render-time check, `checkSubagents` (`internal/render/render.go:298-335`), parses `toolsSettings.subagent.availableAgents/trustedAgents` and `permissions.rules[].match` for `capability == "subagent"` and fails when a non-glob name is not an agent of the team; it also ignores unknown keys. Tests of the validator: 4 agent cases (valid, mcp server missing command and url, HTTP form, invalid JSON) and 2 hook cases (`internal/kiroconfig/validate_test.go:5-73`); none exercises an unknown key.

### `kiro-cli agent validate` is the only check that sees Kiro's own schema, and it is conditional
**Answers:** 3

`internal/golden/kirocli_test.go` runs `kiro-cli agent validate --path <file>` on every committed golden `testdata/golden/*/rendered/agents/*.json` (`kirocli_test.go:60-78`), and on installed `devops.json` and `platformdev.json` with a knowledge-base resource (`kirocli_test.go:84-120`). Per the comment, validate exits 0 even on an invalid config, reports on stderr and is silent on success, so the test passes/fails on output: it errors when the command fails or the output contains `Error` (`kirocli_test.go:55-57`). The comment lists what it catches (invalid JSON, wrong field types, unknown resource type, unresolvable prompt file) and what it does not (permission rules, MCP server shape, existence of `skill://`/`file://` resources) (`kirocli_test.go:34-43`). The file does not state how validate treats unknown top-level keys, and no test feeds it one; that behaviour is not evidenced in the repository.

The test skips (not fails) when `kiro-cli` is absent from PATH or the output contains `not logged in`, unless `KIROKIT_REQUIRE_KIRO_CLI` is non-empty, which turns skips into failures (`kirocli_test.go:13-32,52-54`). The env var appears in no workflow file (grep over `.github/workflows`, `Taskfile.yml`: no match); it is documented in `AGENTS.md:31` and `CONTRIBUTING.md:52`. `kiro-cli` exists on this machine at `/opt/homebrew/bin/kiro-cli` (`command -v`); I did not run the validate test.

### manifest.yaml is decoded by yaml.v3 into one struct, with no unknown-key rejection
**Answers:** 4

The struct is `Manifest` (`internal/manifest/manifest.go:20-29`): `team`, `owner`, `owners{agents,skills}`, `feedback{github}`, `globalSteerings []string`, `externalSkills []ExternalSkill`. `Load` reads the file and calls `yaml.Unmarshal` into it (`manifest.go:148-157`); there is no `KnownFields(true)` / `yaml.Decoder` anywhere in the repository (grep of `*.go`), so unknown or newly-added keys are silently ignored by default yaml.v3 behaviour, and a typo (e.g. `globalSteering`) yields an empty field with no error. Fields a manifest is required to have: only `team` (`manifest.go:158-160`). After parsing, `Load` runs `validateOwnership` (handle regex `^@[A-Za-z0-9]...(/...)?$`, repo regex `owner/repo`, sorted-key reporting; `manifest.go:66-100`) then `expandExternalSkills` (`manifest.go:164-168`). Optional fields are tagged `json:",omitempty"/",omitzero"` so they stay out of the golden `context.json` when unset (`manifest.go:18-29`).

`globalSteerings` is a plain string slice with no validation in manifest; checks happen at render: `BuildContext` reads `<globalRoot>/<slug>.md` for each slug and errors `team %q opts into global steering %q but ... could not be read` if missing (`internal/render/context.go:48-56`); `writeGlobalSteerings` rejects slugs that are not single path elements (`internal/render/globalsteerings.go:55-58`). Opt-in is per team, listing slugs matching `global/steerings/<slug>.md` (only `security-baseline.md` exists); `[]` opts out (`teams/data-team/manifest.yaml`, `teams/product-team/manifest.yaml`: `globalSteerings: []`; `teams/platform-team/manifest.yaml`: `- security-baseline`). Each opted-in steering is written to `kirokit-steerings/<team>/<slug>.md` and its `file://` URI is added to every agent's `resources` (`globalsteerings.go:19,33-51`).

`Load` callers: `internal/render/context.go:38`, `cmd/kirokit/commands/skills.go:47`, `desktop/catalog.go:120` (desktop lists `GlobalSteerings` as steerings, `desktop/catalog.go:194-198`).

### externalSkills entries are validated and expanded inside Load; the lock check and fetch happen elsewhere
**Answers:** 4

`ExternalSkill` has `name`, `source`, `ref`, `paths`, `all`, `exclude`, plus a yaml-ignored `Path` set by `Expand` (`manifest.go:123-133`). `Validate` (`manifest.go:216-264`) enforces: source, ref and a path present; safe relative slash path (266-275); `name` only with a single path and not with `all`; `exclude` only with `all: true` and a single path, entries as plain directory names; source not starting with `-`, no `ext::`-style helper, scheme in `https/http/ssh/git/file`, no embedded credentials on http(s) (277-311); ref without leading `-`, whitespace or control chars (313-325). `Expand` turns a `paths` entry into one entry per path (172-183); duplicate effective names between single (non-`all`) skills fail `Load` (187-209), while collection name clashes are left to the lock check (198-200). Rendering, lock verification and drift detection live in `internal/externalskills` (`lock.go`, `verify.go`, `fetch.go`) and `internal/render/externalskills.go`; I did not trace them beyond README's statement that `render`/`test`/`install`/`sync` fail on drift and `skills fetch` is the only networked step (`README.md:186,193`).

Documentation: README "manifest.yaml Reference" shows the full schema with comments (`README.md:122-154`) and bullets on steering rendering and owners (`README.md:156-159`); "External Skills" (`README.md:177-197`); the file tree line at `README.md:99`; `CONTRIBUTING.md:128` for the fetch workflow. `AGENTS.md:37` requires README updates when a manifest field changes. `teams/platform-team/manifest.yaml` holds a commented-out `externalSkills` example (lines 12-37) that doubles as in-place documentation.

### Where a change would land: churn and hard spots
**Answers:** 3, 4

Commits in the last 90 days (`git log --oneline --since=90.days -- <path> | wc -l`): `internal/kiroconfig` 1, `internal/manifest` 4, `teams/*/agents` 7, `internal/golden` 4. Hard to change: any agent template edit changes committed goldens, which are compared byte-for-byte and must be regenerated with `./kirokit update <team>` (`AGENTS.md:28-29`); the `Manifest` struct is serialized into each team's `testdata/golden/<team>/context.json`, so a new always-emitted field alters every golden (hence `omitempty/omitzero` on optional ones, `manifest.go:18-29`).

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

### kirocrew resolves one Kiro home, honours KIRO_HOME, and reads agents, skills, settings from it
**Answers:** 7

- `kiro_home()` returns `Path.home()/.kiro` unless `KIRO_HOME` is set (`kiro_crew/config/paths.py:1217-1224`). A `KIRO_HOME` that resolves to a root or system directory is ignored with a warning (`paths.py:1221-1223`). Paths below are inside `/Applications/KiroCrew.app/Contents/Resources/backend-dist/kirocrew-backend-arm64/lib/python3.12/site-packages/`.
- The docstring says `KIRO_HOME` is only partly followed today: the agents directory follows it, but session transcripts and the `settings/mcp.json` registry are read from the host `~/.kiro` in "most" readers, so it "is not yet a supported way to isolate an instance" (`kiro_crew/config/paths.py:1201-1210`). This conflicts with `mcp_cleanup.py:249-251`, which does use `kiro_home() / "settings" / "mcp.json"`, and `dashboard/handlers/mcp.py:120`, which hard-codes `Path.home()/.kiro/settings/mcp.json`. The code in each module decides at runtime, so MCP settings are read from different homes depending on the caller (fact; no single winner).
- Agents: user-level dir is `<kiro home>/agents` (`paths.py:1348`, `1351-1378`). `list_agents` scans it (`kiro_crew/agent_discovery.py:2297`, `2341-2350`). A project's `<project>/.kiro/agents` shadows same-named user agents, with no upward walk (`paths.py:1381-1393`; `agent_discovery.py:2278-2285`).
- Skills: only `skill://` entries in an agent's `resources` count as that agent's mapping (`agent_discovery.py:1257-1260`). A `skill://~/...` URI expands with `Path.home()`, not `kiro_home()` (`agent_discovery.py:1331-1332`), and `skill:///abs/...` is used as is (`1333-1334`). A custom agent with no `skill://` entry inherits `<kiro home>/skills/*/SKILL.md` plus `<project>/.kiro/skills/*/SKILL.md`, unless `chat.disableInheritingDefaultResources` is set (`agent_discovery.py:1595-1601`; `acp/skill_projection.py:1929-1953`, which reads `<kiro home>/settings/cli.json`).
- Steering: kirocrew does not parse steering files. When an agent inherits defaults, its generated view gets the resources `file://<kiro home>/steering/**/*.md`, `file://.kiro/steering/**/*.md` and `file://AGENTS.md` (`acp/skill_projection.py:3911-3919`).
- MCP: `<kiro home>/settings/mcp.json` is treated as belonging to kiro-cli. Writes are refused when `KIROCREW_POD` is set (`paths.py:731-751`). Per-agent MCP server names come from the spec's `mcpServers` keys (`agent_discovery.py:2262-2265`).

### Agents kirocrew did not create are listed, kept as the author wrote them, and given a derived view
**Answers:** 7

- Ownership is by exact filename: `OWNED_KIRO_AGENT_FILES` lists `kirocrew.json`, `kirocrew-lite.json`, `kirocrew-conductor.json` etc. (`kiro_crew/agent_files.py:93-108`; the comment at `agent_files.py:89-92` says user files with the same prefix are never touched). Any other file in `~/.kiro/agents` gets `kirocrew_owned=False` and source `"builtin"`, or `"package"` when the filename is `<package>-<name>.json` (`agent_discovery.py:2209-2237`). The directory is described as shared with "ACP adapters and IDE plugins" (`agent_discovery.py:212-213`).
- Fields in foreign specs are coerced rather than rejected (`agent_discovery.py:234-251`). A spec that cannot be parsed is skipped (`agent_discovery.py:2346-2358`).
- For a session, kirocrew builds a copy of each spec as an alias file (name prefix `kirocrew-skill-view-`) in the same agents folder. The copy drops `skill://` resources, replaces `description` with "Kiro Crew managed view of ...", and adds the `kirocrew-core` MCP server and `skill_search` tool when `can_inherit_skills` allows it (`acp/skill_projection.py:3773-3797`, `3837-3860`; description prefix at `skill_projection.py:278-290`). The source spec file is only read, not rewritten (`skill_projection.py:3763-3773`).
- Agents that exclude `skill_search`, declare `tools: []`, or have no `kirocrew-core` entry with `tools: "*"` or missing get no inherited skills and keep their spawn (`agent_discovery.py:1514-1567`). A mapped agent that withholds `skill_search` is refused a view (`skill_projection.py:3797-3801`).
- Deleting a leftover view needs a sidecar record under kirocrew's data home, or the legacy name pattern plus a positive mark (the `kirocrew-core` server or the steering resource) (`skill_projection.py:2118-2167`). "An operator's own agent could in principle carry this name", so name alone is not enough (`skill_projection.py:2138-2145`).
- kirokit's agents in `~/.kiro/agents` (for example `engineer.json`, `qa.json`, `customer.json`) therefore appear as foreign rows. Observed in `~/.kiro/agents/`: they sit beside `kirocrew*.json` files; kirocrew's own data lives in `~/.kiro/crew` (`paths.py:41-49`).
- kirocrew data home defaults to `~/.kiro/crew` and a recovery pointer is kept outside `~/.kiro` (`paths.py:41-65`).

### `kirokit install` renders with the real KIRO_HOME and writes the rendered tree plus per-agent secrets
**Answers:** 8

- `ResolveKiroHome` uses `$KIRO_HOME` if non-empty, else `~/.kiro` (`internal/install/install.go:49-58`).
- `Install` renders into a scratch folder with `KiroHome` set (slashes normalised) (`install.go:92-102`, `75-77`), plans every rendered file at `<kiroHome>/<relative path>` (`install.go:319-351`), and adds `kirokit-secrets/<team>/<agent>/secrets.enc.yaml` for each agent that has `env/secrets.enc.yaml` (`install.go:353-383`).
- Folders seen in a rendered team: `agents`, `hooks`, `prompts`, `skills`, `steering` (`testdata/golden/data-team/rendered/`). Global steerings go to `kirokit-steerings/<team>/` (`internal/render/globalsteerings.go:19`); team `steerings/` map to `steering/` (`internal/render/render.go:383-386`). README names the backed-up folders as `agents`, `prompts`, `skills`, `steering`, `hooks`, `kirokit-steerings`, `kirokit-secrets` (`README.md:281`).
- Files are written `0600`, or `0700` when the rendered file has the owner-execute bit (`install.go:340-343`); existing file modes are reset with chmod (`install.go:451-454`).
- Rendered agent JSON points at Kiro home absolute paths, for example `skill:///kiro-home/skills/jira-search/SKILL.md` and `file:///kiro-home/kirokit-steerings/platform-team/security-baseline.md` in the golden (`testdata/golden/platform-team/rendered/agents/devops.json`, sentinel `/kiro-home`; `install` replaces it with the real home).
- Optional `--kb-*` flags add a `knowledgeBase` resource to one agent and a symlink at `--kb-mount`, which is also recorded in the manifest (`install.go:120-142`, `213-257`; flags `cmd/kirokit/commands/install.go:50-52`).

### The record is `<KIRO_HOME>/.kirokit-installed/<team>.json`, a list of files and symlinks; uninstall and sync remove only those
**Answers:** 8

- Path: `installManifestPath` = `<kiroHome>/.kirokit-installed/<team>.json` (`install.go:290-292`). Shape: `{"files": [...], "symlinks": [...]}` (`install.go:42-45`), mode `0600` (`install.go:303`). The record file itself is listed in `files` (`install.go:144-145`). It is the install record, distinct from the team's `manifest.yaml` in `internal/manifest`, which parses `team`, `owner`, `owners`, `feedback`, `globalSteerings`, `externalSkills` (`internal/manifest/manifest.go:19-28`).
- `Uninstall` removes recorded symlinks, then files, ignoring ones already gone; missing record is an error (`install.go:262-288`).
- `Sync` installs, then removes paths in the previous record that the new record lacks and removes empty parent folders up to the Kiro home (`internal/install/sync.go:16-62`).
- `Installed` lists `*.json` records sorted by team with file counts and mod-time (`internal/install/installed.go:22-54`); `UninstallOthers` removes every installed team except the one being installed (`installed.go:59-75`). The CLI runs `install.Check` first (render into scratch, nothing written), then `uninstallOthers`, then `Install` (`cmd/kirokit/commands/install.go:36-44`); `sync` does the same with `install.Sync` (`cmd/kirokit/commands/sync.go:111-120`). So one team per Kiro home.
- Observed on this machine: `~/.kiro/.kirokit-installed/` holds `platform-team/` and `product-team.json` (directory listing).
- Backups: before writing, each top-level folder containing a file whose content would change is copied to `<kiroHome>/YYYYMMDD-HHmmss-<folder>.backup`; unchanged folders get none; backups are not in the record (`install.go:402-438`; naming and counter in `internal/backup/backup.go:23-35`, `51`, `120`). Observed: `~/.kiro/20261018-105014-agents.backup` and `20261026-195053-agents.backup`.
- Hand-added files in a folder are not recorded, so uninstall and sync never remove them; files written by `install` over an existing file replace it after the backup (`install.go:441-457`).

### The desktop app chooses the Kiro home as saved Settings, else the login shell's KIRO_HOME, else ~/.kiro, and passes it to the CLI as KIRO_HOME
**Answers:** 8

- `App.kiroHome()` returns `Settings.KiroHome` when saved, else `defaultKiroHome` (captured login-shell `KIRO_HOME` once ready, else `~/.kiro`) (`desktop/settings.go:135-155`). Settings live in `settings.json` in the app data folder (`desktop/settings.go:42`, `107-133`); a saved value must be absolute and not the root or home folder itself (`settings.go:157-166`).
- Every terminal and button command gets `KIROKIT_REPO=<repoDir>` and `KIRO_HOME=<kiroHome>` (`desktop/shell.go:117-118`); the zsh wrapper re-asserts them after the user's startup files (`shell.go:144-146`).
- A poller reads the install records of that home every second and emits `installs:changed` (`desktop/installs.go:11`, `53-68`, `88-93`); `Installs()` serves the list (`installs.go:97-103`).
- The app's data directory is `kirokit-desktop` in the user's config folder; default clone `<dataDir>/repo` (`desktop/app.go:55-60`, `133-143`).

### The repo ignores build output and tool state; paths are overridden by flag or environment variable; the only example-file pattern is the scaffold's `*.tmpl` starter
**Answers:** 9

- `.gitignore` ignores `/dist/`, `/kirokit`, `/tmp/`, `.claude/*` (except `.claude/rules/` and `.claude/skills/`), `demo-runbook-*`, and un-ignores `/docs/plans/*/design/` (`.gitignore:1-12`). The checkout's `.git/info/exclude` also lists `/docs/plans/*/design/` and `/docs/plans/*/artifacts/` (QualityLayer entries; `.git/info/exclude` tail). `.gitleaks.toml` and `.sops.yaml` exist at the root.
- `KIROKIT_REPO`: read in the root command's pre-run, made absolute, stored in `repoBase`; `repoPath(rel)` joins it onto `teams`, `global/steerings`, `env/mcp-versions.yaml`, `testdata/golden`; unset means the working directory (`cmd/kirokit/commands/root.go:12-31`, `44-54`).
- `kirokit sync --repo <dir>` overrides it: the order is flag, then `KIROKIT_REPO`, then `.` (`cmd/kirokit/commands/sync.go:77-89`, `128`). `setup` prints `export KIROKIT_REPO=<dir>` (`cmd/kirokit/commands/setup.go:70`, `85`).
- Other environment variables read in Go: `KIRO_HOME` (`internal/install/install.go:50`), `KIROKIT_DESKTOP_CLI` (`desktop/shell.go:41`), `KIROKIT_DESKTOP_SMOKE` (`desktop/main.go:35`), `SHELL`, and from the captured login environment `SOPS_AGE_KEY_FILE`, `APPDATA`, `XDG_CONFIG_HOME`, `HOME` (`desktop/platform.go:32-46`). The desktop app's own overrides are its saved Settings (`RepoURL`, `KiroHome`, `RepoDir`, `Ref`) (`desktop/settings.go:18-23`).
- No tracked `*.example`, `.env.example` or `*.local` files exist (`git ls-files` filtered for example/sample/local/.env gave only `internal/scaffold/starter/agents/example-agent/{agent.json.tmpl,env/mcp.env.tmpl}`). Those are the `kirokit onboard` starter embedded in the binary (`internal/scaffold/`), not a per-user file pattern. Secrets files are `*.enc.yaml` committed and SOPS-encrypted (`AGENTS.md` repository rules). The SOPS age key is read from `SOPS_AGE_KEY_FILE` or the default path (`scripts/showcase/record.sh:62`).
- Ignored local files in this checkout: `kirokit` binary, `tmp/`, `demo-runbook-v1.md` (confirmed with `git check-ignore -v`).

### The README showcase is recorded by `task showcase` in an isolated /tmp folder from the real CLI and app, then checked and published atomically
**Answers:** 10

- `Taskfile.yml` target `showcase` runs `scripts/showcase/record.sh` (`Taskfile.yml:51-54`). Preflight requires tmux, asciinema, agg, ffmpeg, ffprobe, node, npm, go, git, kiro-cli, lsof, Google Chrome, the SOPS age key, and free ports 34115 and 5173 (`record.sh:58-68`). It works in `/tmp/kirokit-demo`, created `mkdir -m 700` with a pid file, removed on exit unless `--keep` (`record.sh:22`, `70-78`, `32-50`).
- Scene 1 (CLI, `cli-drive.sh`): tmux session with `KIRO_HOME=$demo/kiro` and `KIROKIT_REPO=<repo root>` (`record.sh:89-91`). It types `kirokit install platform-team` and waits for `installed platform-team`, then `kiro-cli chat --agent devops --v3`, Enter on "Trust this workspace", waits for "ask a question", types `/jir`, waits for `jira-search`, Esc, `/quit`, waits for "Session ended"; it never presses Enter on a chat message (`scripts/showcase/cli-drive.sh:1-8`, `50-66`). Recorded with asciinema, rendered to a GIF with agg (`record.sh:92-108`).
- Scene 2 (app, `app-drive.mjs`): `wails dev` with a throwaway `HOME`, `KIRO_HOME` and `KIROKIT_REPO` unset (`record.sh:125-132`); headless Chrome via playwright-core types `https://github.com/mavogel/kirokit.git` into `#clone-address`, clicks Clone, waits for the `platform-team` Install button, clicks it, handles an optional "Replace the installed team?" dialog, clicks `#setup-install`, waits for the `.badge.ok` badge and the dock line `installed platform-team into` (`scripts/showcase/app-drive.mjs:64-115`). It samples all page and terminal text every 200 ms into `app-dock.txt` for the privacy check and writes `app-marks.json` (`app-drive.mjs:44-62`, `118-122`).
- Captions (`caption.mjs`): four captions "Install your team's Kiro config", "Kiro starts with the team's agent and skills", "Or click: clone the config repo", "Install the team, see every command"; a title card and an end card with `brew install --cask mavogel/tap/kirokit` and `...kirokit-desktop` and `github.com/mavogel/kirokit` (`caption.mjs:28-40`). It computes cue times from the cast (idle limit 3 s) and the app marks, cuts the clone wait to 1.5 s, and fails if a caption would show under 2.5 s (`caption.mjs:23-25`, `81-109`).
- Compose: ffmpeg joins title, CLI clip, app clip and end card with fades into `showcase.mp4`, then a 12 fps 1000 px-wide GIF (`record.sh:158-210`).
- Checks (`checks.sh`): `privacy` rejects `$HOME`, the git email and `$USER` in `cli.cast`, `app-dock.txt`, `captions.txt`, except the public slugs `mavogel/kirokit` and `mavogel/tap/`; `size` limits GIF to 8 MB and MP4 to 4 MB; `length` needs 30 to 60 s (`checks.sh:8-60`, `record.sh:215-217`). Publish copies to a temp name in `assets/` then renames to `assets/kirokit-showcase.gif` and `.mp4` (`record.sh:219-226`).
- CONTRIBUTING "Showcase video" says to re-record on a Mac when CLI output or app UI changes, lists the brew packages, the other requirements, and `node --test scripts/showcase/` (about 40 s) (`CONTRIBUTING.md:101-112`). `AGENTS.md:11` points to the same section. The README embeds the GIF and links the MP4 (`README.md:10-11`).
- Last 90 days: 2 commits touched `scripts/showcase` (`git log --oneline --since=90.days -- scripts/showcase`).

### The README has 15 H2/H3 sections; other docs point at it by file or by section name, and no test reads its text
**Answers:** 11

- Headings (`README.md`): H1 `kirokit` (3); `Quickstart` (13) with `using the cli` (15) and `using the desktop app` (70); `Directory Layout` (94); `manifest.yaml Reference` (122); `Role agents and feedback` (161); `External Skills` (177); `Central env/mcp-versions.yaml` (199); `Secrets via SOPS` (211); `Testing and verify by rendering upfront` (247); `Onboarding a New Team` (257); `Install / Uninstall` (265) with `Existing files are backed up` (278); `Setup (initial download)` (289); `Sync (automatic updates)` (303); `Desktop app` (338). Lines starting `#` inside code blocks (32-65) are shell comments, not headings. The file is 359 lines; 28 commits in 90 days.
- Anchor links inside the README: `[reference](#manifestyaml-reference)` (`README.md:172`), `[Role agents and feedback](#role...` (`README.md:352`), `[Desktop app](#desktop-app)` (`README.md:92`).
- `CONTRIBUTING.md` links the README by file only (`CONTRIBUTING.md:81`, `152`) and mentions it at `103`; its own in-page anchors are `#development-setup`, `#checks`, `#commit-messages` (`CONTRIBUTING.md:31-34`). `AGENTS.md` mentions `README.md` at lines 4 and 33 and the showcase at 11, with no anchors.
- Section names are quoted in prose elsewhere: "Secrets via SOPS" in `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9` (rendered into new teams) and `global/steerings/security-baseline.md:9`; "External Skills" in `teams/platform-team/manifest.yaml:12`; the README for `owners` and `feedback` in `teams/platform-team/manifest.yaml:4`; `teams/data-team/steerings/data-conventions.md.tmpl:12`; `tmp/demo-runbook-v0.md:247,264` (ignored folder) names "Directory Layout", "Golden Tests" (no such heading today), "Secrets via SOPS", "External Skills", "Sync (automatic updates)".
- The desktop app and its frontend contain no README references (grep of `desktop/` found only a Wails package homepage URL). The only test hit is `internal/kiroconfig/validate_test.go:11`, which uses `file://README.md` as sample data, not the repo README. `scripts/sync-agent-assets.mjs:286` skips files named `README.md` while syncing skills. `Taskfile.yml:52`, `scripts/showcase/record.sh:2` and `package.json:5` mention it in descriptions.
- Teams' rendered golden files can embed the heading name through the scaffold template, so a rename would show in `kirokit onboard` output (`internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9`); goldens under `testdata/golden/` are not generated from that starter (inference from the starter living in `internal/scaffold`, not `teams/`).
- Rule: update README in the same change when a command, flag, manifest field, or template function changes (`AGENTS.md:33`, `CONTRIBUTING.md:152`).
<!-- /ql:folded -->

## Decided with you

<!-- ql:decided -->
No choices were open.
<!-- /ql:decided -->

## The questions

1. **How does `kirokit render` turn a team into an output tree today?** Trace `cmd/kirokit/commands/render.go` into `internal/render/render.go` (`RenderTeamWith`, `teamInventory`, `BuildContext` in `context.go`): which files are templated, which are copied verbatim, how agent `resources` are built, and where output paths are decided.
2. **How are skills rendered and copied?** Team skills (`teams/<team>/skills/<skill>/SKILL.md.tmpl`) versus vendored external skills (`internal/render/externalskills.go`, `internal/externalskills`): what frontmatter they carry, and which path keeps text such as `{{name}}` untouched.
3. **What does a Kiro agent template hold, and what checks it?** The fields in `teams/*/agents/*/agent.json.tmpl` (prompt, description, welcome message, tools, skills, mcpServers), how `internal/kiroconfig/validate.go` and the `kiro-cli agent validate` test in `internal/golden` treat unknown keys.
4. **How is `manifest.yaml` parsed and validated?** Where the manifest struct lives, how unknown or new fields are handled, and how `externalSkills` and global steering opt-ins are declared and documented.
5. **How does the golden harness compare output?** `internal/golden/golden.go` and `cmd/kirokit/commands/test.go`, `update.go`: which tree is compared, how `context.json` is snapshotted, and what `kirokit update` regenerates.
6. **What does Kiro IDE read from `~/.kiro` (or `$KIRO_HOME`)?** Agents, skills, steering, prompts, hooks and MCP settings: use the Kiro IDE docs (kiro.dev) and the installed app `/Applications/Kiro.app` (version 1.4.26); note whether it honours `KIRO_HOME`, and whether it reads Kiro CLI v3 flat agent files at `agents/<name>.json`.
7. **What does kirocrew read from `~/.kiro`?** In the installed app `/Applications/KiroCrew.app/Contents/Resources/backend-dist/kirocrew-backend-arm64/lib/python3.12/site-packages/kiro_crew/` (`config/paths.py`, `acp/skill_projection.py`): how it finds agents, skills, steering and MCP settings, whether it honours `KIRO_HOME`, and how it treats agents it did not create (such as kirokit's).
8. **What does `kirokit install` write into `KIRO_HOME`, and how is it tracked?** `internal/install`, `internal/manifest`, `cmd/kirokit/commands/install.go`, `sync.go`: the folders written, the `.kirokit-installed` record, backups, and how the desktop app picks the Kiro home (`desktop/app.go`).
9. **Which ignored or per-user files does the repo already support?** `.gitignore`, how paths such as `KIROKIT_REPO` are overridden by flag or environment variable, and any existing example-file pattern.
10. **How is the README showcase recorded and checked?** `scripts/showcase/` (`record.sh`, `cli-drive.sh`, `app-drive.mjs`, `caption.mjs`, `showcase.test.mjs`) and the CONTRIBUTING "Showcase video" section: what each scene shows and what the tests assert.
11. **How is the README organised and what other docs link into it?** Its sections and anchors (`README.md`), anchors referenced from `CONTRIBUTING.md`, `AGENTS.md`, the desktop app and its frontend, and any test that reads README text.

**Key context pointers:**

- `docs/research/quick/QUICK-TARGET-BRIEF.md` and its evidence files in `docs/research/quick/`
- `internal/render/`, `internal/golden/`, `internal/kiroconfig/`, `internal/install/`, `internal/manifest/`, `internal/externalskills/`
- `cmd/kirokit/commands/`
- `teams/*/manifest.yaml`, `teams/*/agents/`, `teams/*/skills/`
- `scripts/showcase/`, `README.md`, `CONTRIBUTING.md`, `AGENTS.md`
- Installed apps: `/Applications/Kiro.app`, `/Applications/KiroCrew.app`, `~/.kiro/`

## How it is tested today

All commands run from the repository root.

- **Everything Go:** `go test ./...` (includes every team's golden comparison); `task ci` adds build, vet and lint, the same as CI.
- **One package or test:** `go test ./internal/render -run TestRenderTeam_VendoredExternalSkillIsCopiedVerbatim`, `go test ./internal/golden -run TestCompare_ReportsNoDiffsAfterUpdate`, `go test ./internal/manifest`, `go test ./internal/install -run <name>`.
- **Golden diff only:** `task build && ./kirokit test`; regenerate with `./kirokit update <team>`. Fixtures: `testdata/golden/<team>/rendered/**` and `context.json`, plus `testdata/golden/platform-team/installed-with-kb/` owned by `internal/golden/knowledgebase_test.go`.
- **Kiro's own schema:** `go test ./internal/golden -run TestCommittedGoldenAgentsValidateWithKiroCLI`, which skips without a logged-in `kiro-cli` unless `KIROKIT_REQUIRE_KIRO_CLI=1`.
- **Desktop app:** `task desktop:test` (Go tests in `desktop/` plus `node --test` for `desktop/frontend`).
- **Showcase:** `node --test scripts/showcase/` (privacy, size, length, caption timing; never records); `task showcase` records for real on a Mac.
- **Not tested at all:** the README text, Kiro IDE and kirocrew behaviour, and anything in Amazon Quick.

## Code references


<!-- ql:folded references -->
- `cmd/kirokit/commands/render.go:11-44`, `cmd/kirokit/commands/root.go:74-89`
- `internal/render/render.go:40-155` (pipeline), `:164-205` (publish), `:212-229` (renderTemplate), `:243-291` (inventory, owners), `:337-391` (validate, mapOutputPath)
- `internal/render/context.go:19-68`, `internal/render/funcs.go:15-24`
- `internal/render/globalsteerings.go:19-151` (steering dir, workspace resources, addResources)
- `internal/render/externalskills.go:16-79`
- `internal/externalskills/lock.go:19-27`, `internal/externalskills/verify.go:28-91`, `internal/externalskills/fetch.go:155-164`
- `teams/*/skills/*/SKILL.md.tmpl`, `teams/data-team/agents/godev/agent.json.tmpl:19-21`, `testdata/golden/data-team/rendered/`
- Churn, last 90 days (commit counts): `internal/render` 10, `internal/externalskills` 4, `cmd/kirokit/commands/render.go` 3.
- Exhaustive for the render and external-skill paths named in the questions; `fetch.go` network logic and `install`/`manifest` were not traced.
- `internal/kiroconfig/validate.go:1-99` (validators), `internal/kiroconfig/validate_test.go:5-73`
- `internal/golden/kirocli_test.go:13-120` (kiro-cli validate tests)
- `internal/render/render.go:64-65,119-132,298-346` (agent resources, validateOutput, checkSubagents)
- `internal/render/globalsteerings.go:19-111`, `internal/render/context.go:35-68`
- `internal/manifest/manifest.go:18-325`, `internal/manifest/manifest_test.go`
- `teams/*/agents/*/agent.json.tmpl` (all nine read via grep; devops and qa read in full), `teams/*/manifest.yaml`, `global/steerings/security-baseline.md`
- `README.md:99,122-197`, `CONTRIBUTING.md:52,128`, `AGENTS.md:28-37`
- Exhaustiveness: complete for validators, manifest.go and the Load/validateOutput call sites (grep over all `*.go`); partial for externalskills internals and render tests.

Exhaustive for the two harness files and the three command files; selective for the IDE bundle (searched for `KIRO_HOME`, `userConfigDir`, agent/skill/steering/hook/prompt/mcp loaders).
- Harness: `internal/golden/golden.go:19-25` (Inputs), `:29-57` (Compare), `:60-84` (Update), `:90-113` (renderForGolden), `:117-158` (diffDirs), `:160-212` (helpers)
- Commands: `cmd/kirokit/commands/test.go:10-55`, `cmd/kirokit/commands/update.go:10-42`, `cmd/kirokit/commands/root.go:16-31`, `:74-88`
- Context: `internal/render/context.go:16-33` (Context, json tags), `:37-71` (BuildContext)
- Kiro home resolution (kirokit side): `internal/install/install.go:47-58`
- Goldens: `testdata/golden/<team>/context.json`, `testdata/golden/<team>/rendered/**`, `testdata/golden/platform-team/installed-with-kb/`
- Tests: `internal/golden/golden_test.go`, `internal/golden/kirocli_test.go`, `internal/golden/knowledgebase_test.go`, `cmd/kirokit/commands/root_test.go:110-145`
- Kiro IDE (installed, outside repo): `/Applications/Kiro.app/Contents/Resources/app/extensions/kiro.kiro-agent/dist/extension.js` lines 1774 (path constants), 2607 (agent schema, CLI-only fields), 3936 (hooks), 14798 (MCP config), 16032, 16261 (skills, steering, prompts watcher), 16392 (agent loader), 16686 (`KIRO_HOME`), 16776 (`tF`)
- Docs: https://kiro.dev/docs/cli/custom-agents/configuration-reference/, https://kiro.dev/docs/steering/

Exhaustive for the files named in the questions; selective for kirocrew (a 5,450-line pair plus targeted greps of the package).
- kirocrew (installed): `kiro_crew/config/paths.py:41-65, 731-793, 1190-1224, 1348-1393`; `kiro_crew/agent_discovery.py:186-257, 1249-1260, 1306-1336, 1514-1602, 2199-2265, 2268-2358`; `kiro_crew/agent_files.py:20-26, 93-108`; `kiro_crew/acp/skill_projection.py:1-8, 255-290, 1929-1953, 2118-2167, 3760-3860, 3911-3919`; `kiro_crew/mcp_cleanup.py:249-251`; `kiro_crew/dashboard/handlers/mcp.py:120-123`. Base: `/Applications/KiroCrew.app/Contents/Resources/backend-dist/kirocrew-backend-arm64/lib/python3.12/site-packages/`.
- Install and records: `internal/install/install.go`, `installed.go`, `sync.go`; `internal/backup/backup.go`; `internal/render/globalsteerings.go:19`, `render.go:341-386`; `internal/manifest/manifest.go:19-28`; `cmd/kirokit/commands/install.go`, `sync.go`, `root.go:12-54`, `setup.go:70,85`.
- Desktop: `desktop/app.go:55-143`, `desktop/settings.go:18-166`, `desktop/installs.go`, `desktop/shell.go:107-146`, `desktop/platform.go:32-46`.
- Ignored files: `.gitignore:1-12`, `.git/info/exclude`, `.gitattributes`, `internal/scaffold/starter/`.
- Showcase: `scripts/showcase/{record.sh,cli-drive.sh,app-drive.mjs,caption.mjs,checks.sh,showcase.test.mjs,package.json}`, `Taskfile.yml:51-54`, `CONTRIBUTING.md:101-112`, `AGENTS.md:11`.
- README links: `README.md` headings (list above), `CONTRIBUTING.md:31-34,81,103,152`, `AGENTS.md:4,11,33`, `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9`, `global/steerings/security-baseline.md:9`, `teams/platform-team/manifest.yaml:4,12`.

**How it is tested today**
- Render unit tests: `internal/render/render_test.go` (fixture `setupFixture` at `:19`, fixture team with `skills/skill1/SKILL.md.tmpl` at `:48-49`) and `internal/render/externalskills_test.go` (vendored copy verbatim, wiring, drift, name collision, collections). They use temp dirs and write a lock with a real hash; no network.
- Golden comparison: `internal/golden/golden.go` (`Compare`, `Update`, `renderForGolden` calls `render.RenderTeam` with the sentinel KiroHome, `:108`); expected trees in `testdata/golden/<team>/rendered` and `context.json`; tests in `internal/golden/golden_test.go`.
- Commands (repository root): one file's package `go test ./internal/render -run TestRenderTeam_VendoredExternalSkillIsCopiedVerbatim`; package `go test ./internal/render`; all `go test ./...` (includes every team's golden comparison, per `AGENTS.md`); CLI golden diff `./kirokit test` after `task build`; full CI `task ci`. Go tests are per package, not per file. I did not run any of them.
- Go test suite from the repository root: `go test ./...` (`Taskfile.yml:19-22` runs `go test ./... -v`; `task ci` adds build, vet, lint). Fixtures for agents are the committed goldens under `testdata/golden/<team>/rendered/` and `installed-with-kb/`, compared byte-for-byte by `internal/golden` (`go test ./internal/golden`).
- Validator: `go test ./internal/kiroconfig` (table-less direct cases, in-line JSON strings; `internal/kiroconfig/validate_test.go`). Single test: `go test ./internal/kiroconfig -run TestValidateAgentJSON_RejectsInvalidJSON`.
- Manifest: `go test ./internal/manifest` (`internal/manifest/manifest_test.go`, 20+ `TestLoad_*` cases, each writing a temp YAML via `writeFixture`; covers valid, missing team, external-skill rejects, owners/feedback). No test asserts behaviour for an unknown manifest key (grep `unknown` in the test file: no such case).
- Kiro's own validator: `go test ./internal/golden -run TestCommittedGoldenAgentsValidateWithKiroCLI` (skips without a logged-in `kiro-cli`; `KIROKIT_REQUIRE_KIRO_CLI=1` forces it). Not run by me.
- Render-time checks (subagent names, owners naming real agents): `go test ./internal/render` (`internal/render/render_test.go`).
- Golden package (`internal/golden`): fixture-based unit tests build a temp repo layout via `setupFixture` and `mustWrite` (`internal/golden/golden_test.go:11-36`) and cover Compare after Update, detection of a changed template and Update regeneration (`internal/golden/golden_test.go:38`, `:54`, `:73`). `kirocli_test.go` runs `kiro-cli agent validate` over committed goldens and skips unless a logged-in `kiro-cli` exists or `KIROKIT_REQUIRE_KIRO_CLI=1` (`internal/golden/kirocli_test.go:14-60`); `knowledgebase_test.go` compares an installed-agent golden and skips off macOS/Linux (`internal/golden/knowledgebase_test.go:51-69`).
- Commands: `cmd/kirokit/commands/root_test.go` runs `test` against the real repo goldens through `KIROKIT_REPO` (`:110-122`) and requires every team under `teams/` to report `<team>: ok` (`:126-145`). There is no dedicated test file for `update.go` or `test.go`.
- Commands, from the repository root: one file's package `go test ./internal/golden -run TestCompare_ReportsNoDiffsAfterUpdate`; whole suite `go test ./...` (includes every team's golden comparison); CLI golden diff only: `./kirokit test` after `task build`; full CI mirror `task ci` (AGENTS.md). Observed: `go test ./internal/golden ./cmd/kirokit/commands -run 'Golden|Compare|Update|Kirokit'` printed `Go test: 25 passed in 2 packages`.
- Change frequency (90 days): 5 commits touch `internal/golden`, `test.go`, `update.go` together (`git log --oneline --since=90.days -- internal/golden cmd/kirokit/commands/test.go cmd/kirokit/commands/update.go`).
- IDE behaviour has no automated test in this repo; the findings above come from reading the installed bundle only.
- Install and backup logic: Go tests beside the code, using temp dirs (`internal/install/install_test.go`, `installed_test.go`, `sync_test.go`, `install_sops_test.go`; `internal/backup/backup_test.go`; `internal/manifest/manifest_test.go`; `cmd/kirokit/commands/sync_test.go`, `root_test.go`). One file: `go test ./internal/install -run TestName`. Suite: `go test ./...` from the repo root, or `task ci` (build, vet, lint, test) (`AGENTS.md` Commands). Tests needing a Unix shell or execute bits skip on Windows (`internal/testutil`).
- Desktop app: own module; `cd desktop && go test ./...` (for example `desktop/settings_test.go`, `installs_test.go`); `task desktop:test` runs it including `node --test` for `desktop/frontend` (`AGENTS.md`). Not run by me.
- Showcase: `node --test scripts/showcase/` runs `scripts/showcase/showcase.test.mjs` (472 lines) for the privacy, size and length checks, caption timing and error paths of `caption.mjs`, `app-drive.mjs`, `cli-drive.sh` and `record.sh` preflight; tests needing Chrome, ffmpeg or tmux skip without them. From the root, one file is `node --test scripts/showcase/showcase.test.mjs`. It never records anything (`CONTRIBUTING.md:112`).
- kirocrew: the installed app ships no tests that I opened; nothing about it was executed. All kirocrew findings are from reading the installed source.
- README: no automated test reads it; `kirokit test` and `go test ./internal/golden` compare only `testdata/golden/**`.
- None of the test commands above were run by me; this research was read-only.
<!-- /ql:folded -->
