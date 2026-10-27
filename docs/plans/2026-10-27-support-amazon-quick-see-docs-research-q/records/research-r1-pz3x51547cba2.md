---
type: research
---

# How kirokit render builds an output tree, and how skills are rendered and copied

## Findings

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

## How it is tested today

- Render unit tests: `internal/render/render_test.go` (fixture `setupFixture` at `:19`, fixture team with `skills/skill1/SKILL.md.tmpl` at `:48-49`) and `internal/render/externalskills_test.go` (vendored copy verbatim, wiring, drift, name collision, collections). They use temp dirs and write a lock with a real hash; no network.
- Golden comparison: `internal/golden/golden.go` (`Compare`, `Update`, `renderForGolden` calls `render.RenderTeam` with the sentinel KiroHome, `:108`); expected trees in `testdata/golden/<team>/rendered` and `context.json`; tests in `internal/golden/golden_test.go`.
- Commands (repository root): one file's package `go test ./internal/render -run TestRenderTeam_VendoredExternalSkillIsCopiedVerbatim`; package `go test ./internal/render`; all `go test ./...` (includes every team's golden comparison, per `AGENTS.md`); CLI golden diff `./kirokit test` after `task build`; full CI `task ci`. Go tests are per package, not per file. I did not run any of them.

## Open questions

None.

## Code references

- `cmd/kirokit/commands/render.go:11-44`, `cmd/kirokit/commands/root.go:74-89`
- `internal/render/render.go:40-155` (pipeline), `:164-205` (publish), `:212-229` (renderTemplate), `:243-291` (inventory, owners), `:337-391` (validate, mapOutputPath)
- `internal/render/context.go:19-68`, `internal/render/funcs.go:15-24`
- `internal/render/globalsteerings.go:19-151` (steering dir, workspace resources, addResources)
- `internal/render/externalskills.go:16-79`
- `internal/externalskills/lock.go:19-27`, `internal/externalskills/verify.go:28-91`, `internal/externalskills/fetch.go:155-164`
- `teams/*/skills/*/SKILL.md.tmpl`, `teams/data-team/agents/godev/agent.json.tmpl:19-21`, `testdata/golden/data-team/rendered/`
- Churn, last 90 days (commit counts): `internal/render` 10, `internal/externalskills` 4, `cmd/kirokit/commands/render.go` 3.
- Exhaustive for the render and external-skill paths named in the questions; `fetch.go` network logic and `install`/`manifest` were not traced.
