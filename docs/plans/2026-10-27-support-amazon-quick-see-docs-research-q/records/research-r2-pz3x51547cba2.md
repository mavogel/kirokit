---
type: research
---

# Agent templates, their validation, and manifest.yaml parsing

## Findings

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

## How it is tested today

- Go test suite from the repository root: `go test ./...` (`Taskfile.yml:19-22` runs `go test ./... -v`; `task ci` adds build, vet, lint). Fixtures for agents are the committed goldens under `testdata/golden/<team>/rendered/` and `installed-with-kb/`, compared byte-for-byte by `internal/golden` (`go test ./internal/golden`).
- Validator: `go test ./internal/kiroconfig` (table-less direct cases, in-line JSON strings; `internal/kiroconfig/validate_test.go`). Single test: `go test ./internal/kiroconfig -run TestValidateAgentJSON_RejectsInvalidJSON`.
- Manifest: `go test ./internal/manifest` (`internal/manifest/manifest_test.go`, 20+ `TestLoad_*` cases, each writing a temp YAML via `writeFixture`; covers valid, missing team, external-skill rejects, owners/feedback). No test asserts behaviour for an unknown manifest key (grep `unknown` in the test file: no such case).
- Kiro's own validator: `go test ./internal/golden -run TestCommittedGoldenAgentsValidateWithKiroCLI` (skips without a logged-in `kiro-cli`; `KIROKIT_REQUIRE_KIRO_CLI=1` forces it). Not run by me.
- Render-time checks (subagent names, owners naming real agents): `go test ./internal/render` (`internal/render/render_test.go`).

## Open questions

Question 3: How `kiro-cli agent validate` treats unknown top-level agent keys is not shown by any file or test in the repository, and I did not run it against a crafted config.
Question 3: Which top-level agent fields Kiro CLI v3 accepts (for example a welcome-message field) is not stated in the repository templates, README or validator; I did not consult Kiro's schema.
Question 4: I did not trace `internal/externalskills` and `internal/render/externalskills.go` line by line beyond what Load, the README and the manifest tests show.

## Code references

- `internal/kiroconfig/validate.go:1-99` (validators), `internal/kiroconfig/validate_test.go:5-73`
- `internal/golden/kirocli_test.go:13-120` (kiro-cli validate tests)
- `internal/render/render.go:64-65,119-132,298-346` (agent resources, validateOutput, checkSubagents)
- `internal/render/globalsteerings.go:19-111`, `internal/render/context.go:35-68`
- `internal/manifest/manifest.go:18-325`, `internal/manifest/manifest_test.go`
- `teams/*/agents/*/agent.json.tmpl` (all nine read via grep; devops and qa read in full), `teams/*/manifest.yaml`, `global/steerings/security-baseline.md`
- `README.md:99,122-197`, `CONTRIBUTING.md:52,128`, `AGENTS.md:28-37`
- Exhaustiveness: complete for validators, manifest.go and the Load/validateOutput call sites (grep over all `*.go`); partial for externalskills internals and render tests.
