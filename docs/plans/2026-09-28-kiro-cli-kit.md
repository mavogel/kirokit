# Kiro CLI Kit Implementation Plan

Created: 2026-09-28
Author: info@manuel-vogel.de
Agent: Claude Code
Status: VERIFIED
Approved: Yes
Iterations: 0
Worktree: No
Type: Feature

## Summary

**Goal:** A `kirokit` Go CLI plus a `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md`-driven repository layout that lets multiple teams own their own Kiro CLI v3 config sources as templates, renders them deterministically into Kiro's real config layout, pins MCP server versions centrally, injects secrets via SOPS at MCP-launch time (never at template-render time), locks the rendering down with golden-file tests run in CI, and ships darwin/linux binaries via semantic-release + goreleaser.

## Out of Scope

- Actually pushing this repository to GitHub or triggering a real release run. This plan adds and locally validates the release configuration; the first live `semantic-release` run happens once the user pushes to `github.com/mavogel/kiro-cli-kit` and merges a Conventional Commit to `main`.
- Managing real SOPS recipient keys for real teams/engineers — only a test-only fixture age key shared by both example teams/CI is created (Task 9). Onboarding a real team's real secrets is documented in the README but not performed here.
- Example teams beyond the two built here (`platform-team`, `data-team`) — together they exercise every kit feature (per-agent SOPS secrets, a non-secret MCP server, shared hooks/prompts/skills, global-steering opt-in *and* opt-out) across 4 agents, which is enough to prove the pipeline without indefinite examples. Additional teams are created by users via `onboard`.

## Approach

**Chosen:** A single Go module (`github.com/mavogel/kiro-cli-kit`, binary `kirokit`) with layered internal packages — `manifest`/`mcpversions` (config parsing) → `kiroconfig` (output validation) → `render` (templating) → `sopswrap` (SOPS wrapper generation) → `golden` (test harness) → `scaffold`/`install` (onboard/install commands) — wired to a `cobra`-based CLI, plus two fully-worked example teams whose golden fixtures prove the whole pipeline, and goreleaser/semantic-release release plumbing.
**Why:** Matches the PRD's explicit architecture decisions (Go, per-agent SOPS env files, registry-style manifests, `.tmpl` naming) with a straightforward layered dependency chain — each package only depends on packages before it, so tasks can be built and tested bottom-up without rework.

## Global Constraints

- Go module path: `github.com/mavogel/kiro-cli-kit` (user-confirmed); Go version floor: `go 1.23`; binary name: `kirokit`.
- Kiro CLI v3 verified output paths (project scope, all under the target `.kiro/` root): `agents/<name>.json` (flat — **not** a per-agent subfolder), `hooks/<name>.json`, `skills/<skill>/SKILL.md`, `steering/<name>.md`.
- `mcpServers` stdio entry fields: `command` (string), `args` ([]string), `env` (map[string]string), `timeout` (int ms, default 60000) — no `disabled` field.
- `resources` `knowledgeBase` entry fields: required `type: "knowledgeBase"`, `source`, `name`; optional `description`, `indexType` (`"best"` default | `"fast"`), `autoUpdate`.
- Verified `sops exec-env` syntax: **exactly two positional arguments** — `sops exec-env <file> "<entire command as one shell string>"`. No `--` separator (a `--` is counted as a third positional arg and produces `error: missing file to decrypt`). The command string is passed to `/bin/sh -c` on both darwin and linux.
- Template file naming: every templated source file uses `<real-extension>.tmpl` (e.g. `agent.json.tmpl`); the renderer strips the trailing `.tmpl` to produce the output filename.
- `KIRO_HOME` env var overrides the whole `~/.kiro` directory; default is `~/.kiro` when unset.
- No `${VAR}` / `${env:VAR}` placeholder may appear in any rendered output — every value is resolved to a literal at render/install time (Kiro CLI's own substitution is undocumented/unreliable, per kirodotdev/Kiro#10536).
- goreleaser: `version: 2` is a required top-level key; use current field names `archives.ids` (not deprecated `archives.builds`) and `archives.formats` (not deprecated `archives.format`).
- Release action refs (verified current as of this plan against goreleaser.com/ci/actions): `actions/checkout@v7` (with `fetch-depth: 0`), `actions/setup-go@v7` (`go-version: stable`), `goreleaser/goreleaser-action@v7`. Re-confirm these major versions still exist immediately before Task 11/12 if significant time has passed, since action majors move independently of this plan.
- `teams/`, `global/steerings/`, and `env/mcp-versions.yaml` are always resolved relative to the process's current working directory — `kirokit` has no repo-root auto-discovery (e.g. walking up for `go.mod`) in v1, so every command must be run from the repository root. This is documented as a hard requirement in the README (Task 13).
- semantic-release plugin set: `@semantic-release/commit-analyzer`, `@semantic-release/release-notes-generator`, `@semantic-release/changelog`, `@semantic-release/git`, `@semantic-release/exec` (goreleaser invoked from `exec.publishCmd`) — deliberately **no** `@semantic-release/github`, so goreleaser is the one tool that creates the GitHub release from the tag semantic-release creates.

## Context for Implementer

`render.Context.KiroHome` exists purely to keep golden-file tests deterministic while still letting `install` bake real, fully-resolved absolute paths into rendered output. Kiro CLI's `mcpServers` command/args are not documented as shell-expanded (see Assumptions), so the kit never relies on `$VAR`/`~` shortcuts inside rendered JSON — every path is a literal string baked in at render time. `render`/`test`/`update` (Task 5) always use the fixed sentinel `/kiro-home` for `KiroHome`; only `install` (Task 7) resolves the real `$KIRO_HOME`/`~/.kiro` and re-renders with that value. This is why golden fixtures never contain a real machine's home directory path, and why `install`'s output is never compared against `testdata/golden/`.

## Assumptions

- Kiro CLI invokes an `mcpServers` entry's `command`/`args` directly (no shell expansion of `$VAR`/`~`) — Tasks 4, 7, 9, 10 depend on this. Low-risk if wrong: every path the kit renders into these fields is already a literal, fully-resolved string, never a variable reference, so the assumption being wrong would not break the rendered output either way.
- `resources.knowledgeBase.source` with a `file://` URI pointing at a symlinked directory is indexed by Kiro following the symlink — this is undocumented (research found no mention of symlinks either way). Task 8 depends on this for the feature to work as intended; if wrong, the rendered JSON is still valid, only Kiro's own indexing behavior might not follow the link.
- The PRD's Scope section describes agent output as "one config file per agent under `agents/<agent>/agent.json`" (a per-agent subfolder), while its own Technical Context section cites the verified kiro.dev docs showing a flat `.kiro/agents/*.json` layout — the PRD is internally inconsistent on this point. This plan follows the flat layout (Global Constraints, Task 3) because it matches the directly verified kiro.dev schema, not the PRD's own illustrative Scope wording; the per-agent *source* folder (`teams/<team>/agents/<agent>/`) still exists for organizing each agent's template + its own SOPS env files, it just renders to a flat output filename.

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Kiro CLI's real invocation model for `mcpServers.command` (shell vs. direct exec) is undocumented; a wrong assumption could break MCP server auth for every installed team | Medium | High | Every path baked into rendered `command`/`args` is always a literal, fully-resolved absolute string (never a `$VAR`/`~` shortcut), so correctness does not depend on which model Kiro actually uses (see Assumptions) |
| `knowledgeBase` symlink `source` support is undocumented; Kiro's indexer may not follow the link | Medium | Medium | Implement exactly as requested (Task 8) and document the caveat in the README (Task 13) — no local CI can validate real Kiro CLI indexing behavior |

## Progress Tracking

- [x] Task 1: Repo bootstrap, git init, manifest & central env parsing
- [x] Task 2: Kiro output schema validation (agent.json / hook.json)
- [x] Task 3: Template rendering engine
- [x] Task 4: SOPS exec-env wrapper generator
- [x] Task 5: Golden-file test harness + render/test/update CLI commands
- [x] Task 6: `onboard` command (new-team scaffolding)
- [x] Task 7: `install`/`uninstall` commands
- [x] Task 8: `install --knowledge-base` symlinked resource attachment
- [x] Task 9: Example team "platform-team" (devops + platformdev) with shared SOPS/global-steering infra
- [x] Task 10: Example team "data-team" (javadev + godev)
- [x] Task 11: CI workflow (build, vet, lint, test)
- [x] Task 12: goreleaser + semantic-release release wiring
- [x] Task 13: README documentation

## Implementation Tasks

### Task 1: Repo bootstrap, git init, manifest & central env parsing

**Objective:** Initialize the git repository (`git init` + an initial commit) and the Go module (`github.com/mavogel/kiro-cli-kit`, Go 1.23 floor), and implement typed parsing for a team's `manifest.yaml` (team name + list of global steerings it opts into) and the central `env/mcp-versions.yaml` (flat map of MCP server name → pinned version/tag string). This is the foundational data layer every later package depends on.

**Files:**

- Create: `go.mod`
- Create: `go.sum`
- Create: `.gitignore`
- Create: `internal/manifest/manifest.go`
- Create: `internal/manifest/manifest_test.go`
- Create: `internal/mcpversions/mcpversions.go`
- Create: `internal/mcpversions/mcpversions_test.go`
- Create: `env/mcp-versions.yaml`
- Test: `internal/manifest/manifest_test.go`
- Test: `internal/mcpversions/mcpversions_test.go`

**Key Decisions / Notes:**

- `manifest.yaml` schema: `team: <name>`, `globalSteerings: [<slug>, ...]` (slugs match filenames under `global/steerings/` without extension; an empty list is valid and means opting out). Parse with `gopkg.in/yaml.v3`.
- `internal/manifest`: `type Manifest struct { Team string; GlobalSteerings []string }`, `func Load(path string) (*Manifest, error)` — errors clearly when `team` is empty/missing.
- `internal/mcpversions`: `type Versions map[string]string`, `func Load(path string) (Versions, error)` — errors on a non-scalar YAML value for any key.
- `env/mcp-versions.yaml` seeded with one placeholder entry (e.g. `example-mcp: "1.0.0"`); Tasks 9 and 10 add the real entries their example agents use.
- `.gitignore`: ignore build output (`/dist/`, `/kirokit`) — **not** `testdata/golden/`, which must be committed.

**Definition of Done:**

- [ ] `go build ./...` succeeds
- [ ] `manifest.Load` parses a valid fixture manifest.yaml and returns a clear error for one missing `team`
- [ ] `mcpversions.Load` parses `env/mcp-versions.yaml` into a flat map and errors on a fixture with a non-scalar value
- [ ] Verify: `go test ./internal/manifest/... ./internal/mcpversions/... -v`

### Task 2: Kiro output schema validation (agent.json / hook.json)

**Objective:** Implement minimal structural validation for the two standalone JSON shapes the kit renders — agent configs (`agents/<name>.json`) and standalone hook files (`hooks/<name>.json`) — so the render pipeline catches malformed output beyond mere golden-diff drift, per the PRD's constraint that rendered output must be structurally valid since no live Kiro CLI exists in CI to check against.

**Files:**

- Create: `internal/kiroconfig/validate.go`
- Create: `internal/kiroconfig/validate_test.go`

**Key Decisions / Notes:**

- `ValidateAgentJSON(data []byte) error`: must unmarshal as a JSON object; if an `mcpServers` key is present, each entry must have either a `command` (string) key (stdio form) or a `url` (string) key (HTTP form) — reject entries with neither; if a `resources` key is present, it must be a JSON array (elements not deeply validated).
- `ValidateHookJSON(data []byte) error`: top-level `version` (string) and `hooks` (array) required; each hook entry requires `name`, a `trigger` matching one of the verified values (`PromptSubmit`, `AgentStop`, `SessionStart`, `AgentSpawn`, `PreToolUse`, `PostToolUse`, `FileCreate`, `FileSave`, `FileDelete`, `PreTaskExecution`, `PostTaskExecution`), and an `action` object with `type` of `"command"` or `"agent"`.
- Validation is intentionally shallow (structural shape only) — golden files remain the primary correctness signal; this only catches gross malformation.

**Definition of Done:**

- [ ] `ValidateAgentJSON` accepts a well-formed fixture and rejects one whose `mcpServers` entry has neither `command` nor `url`
- [ ] `ValidateHookJSON` accepts a well-formed fixture and rejects one with an unrecognized `trigger` value
- [ ] Verify: `go test ./internal/kiroconfig/... -v`

### Task 3: Template rendering engine

**Objective:** Implement the core rendering pipeline: given a team directory, walk every `*.tmpl` file under `teams/<team>/`, render it with Go's `text/template` against a context built from that team's `manifest.yaml`, the central `env/mcp-versions.yaml`, and available global steerings, then write output mapped to Kiro CLI's real flat layout. Non-`.tmpl` files (e.g. a SOPS-encrypted secrets file) are copied through byte-for-byte. Rendered `agents/**.json` and `hooks/**.json` outputs are passed through Task 2's validators before a render is considered successful.

**Files:**

- Create: `internal/render/render.go`
- Create: `internal/render/context.go`
- Create: `internal/render/funcs.go`
- Create: `internal/render/render_test.go`

**Key Decisions / Notes:**

- Path mapping, source → output (Kiro's real layout is flatter than the kit's source layout):
  - `teams/<team>/agents/<agent>/agent.json.tmpl` → `agents/<agent>.json` (flat — no per-agent subfolder in Kiro's own layout)
  - `teams/<team>/agents/<agent>/env/*.env.tmpl` → **not written to the output tree at all**; consumed only by the `sopsArgsJSON` template func (Task 4) baked into that agent's `mcpServers` entry
  - `teams/<team>/hooks/*.json.tmpl` → `hooks/<name>.json`
  - `teams/<team>/steerings/*.md.tmpl` → `steering/<name>.md`
  - `teams/<team>/skills/<skill>/SKILL.md.tmpl` → `skills/<skill>/SKILL.md`
  - `teams/<team>/prompts/*.md.tmpl` → `prompts/<name>.md` (referenced by hand from an agent template's `resources`/`prompt` field via `file://prompts/<name>.md` — the renderer does not auto-wire this reference)
- **Deliberate divergence from the PRD's illustrative description:** the PRD describes the central MCP version pin flowing through "each agent's `env/*.env.tmpl` template," but that file type is reserved here specifically for SOPS-wrapped secret command generation (Task 4). The version pin itself is looked up directly inside `agent.json.tmpl` via `{{ mcpVersion "<name>" }}` — an agent with no secret (e.g. Task 10's `godev`) has no `env/` directory at all and still gets its version pin this way. The PRD explicitly leaves exact injection mechanics as an implementation detail for `/spec` (PRD Technical Context), so this is a permitted choice; documented here and in the README (Task 13) so a reader doesn't expect `env/*.env.tmpl` to carry version-pin logic it doesn't.
- `render.Context` struct: `{ Team manifest.Manifest; Env mcpversions.Versions; Global map[string]string /* steering slug → content */; KiroHome string }`. `KiroHome` defaults to the literal sentinel `/kiro-home` (see Context for Implementer).
- `text/template` `FuncMap`: `mcpVersion "<name>"` (looks up `.Env`, errors if missing), `globalSteering "<slug>"` (looks up `.Global`, errors if the team's manifest didn't opt in), `jsonString` (marshals a Go string to a properly quoted+escaped JSON string literal for embedding in a template-authored JSON file).
- Discover templates via `filepath.WalkDir` + `strings.HasSuffix(path, ".tmpl")`; map source path to output path via an explicit function (not generic suffix-stripping, since the agents/hooks/skills shapes are not uniform — see mapping above).
- Render to a temp directory first; only move/copy into the final output directory on full success (validation included) — no partial/corrupt output left in place on failure.

**Definition of Done:**

- [ ] Rendering a fixture team (one agent using `mcpVersion`/`jsonString`, one hook, one steering doc opting into a global steering, one skill) produces the correct flat output tree with every `{{ }}` resolved and no `.tmpl` suffix remaining
- [ ] A template referencing an undefined `mcpVersion` key fails the render with an error naming the missing key and offending file
- [ ] A rendered agent.json that fails Task 2's validator fails the render with a clear error
- [ ] Verify: `go test ./internal/render/... -v`

### Task 4: SOPS exec-env wrapper generator

**Objective:** Implement the verified `sops exec-env` invocation shape as a reusable Go helper wired into the template `FuncMap`, so agent templates render a correct MCP server `command`/`args` pair that decrypts secrets at MCP-launch time. This corrects an earlier informal `sops exec-env <file> -- <cmd> <args>` assumption — the verified contract is exactly two positional arguments: the file to decrypt, then the entire command as one string that `sops` hands to `/bin/sh -c`.

**Files:**

- Create: `internal/sopswrap/sopswrap.go`
- Create: `internal/sopswrap/sopswrap_test.go`
- Modify: `internal/render/funcs.go`
- Modify: `internal/render/render_test.go`

**Key Decisions / Notes:**

- `func BuildWrapper(secretsFilePath, realCommand string, realArgs []string) (command string, args []string)` returns `command = "sops"`, `args = []string{"exec-env", secretsFilePath, shellQuote(realCommand, realArgs)}` — exactly 3 elements in `args`, never a `--` separator.
- `shellQuote` joins `realCommand`+`realArgs` into one POSIX-shell-safe string via single-quote wrapping with `'\''`-escaping for embedded single quotes (the joined string is later interpreted by `/bin/sh -c`).
- Two new template funcs in `internal/render/funcs.go`, backed by `internal/sopswrap`: `sopsCommand` (returns the literal string `"sops"`) and `sopsArgsJSON "<secretsFilePath>" "<realCommand>" "<realArg1>" ...` (returns the already-JSON-array-formatted string `["exec-env", "<path>", "<quoted command>"]`, ready to paste directly into a template's `"args": {{ sopsArgsJSON ... }}`).

**Definition of Done:**

- [ ] `BuildWrapper` produces exactly `["exec-env", "<path>", "<cmd> <args...>"]` for a command with no special characters
- [ ] `shellQuote`'s output round-trips correctly through an actual `/bin/sh -c` subprocess for an argument containing a single quote and a space (assert the printed argv matches the original unquoted values)
- [ ] `sopsArgsJSON` is wired into `internal/render/funcs.go` and covered by a `render_test.go` case producing a syntactically valid JSON `args` array
- [ ] Verify: `go test ./internal/sopswrap/... ./internal/render/... -v`

### Task 5: Golden-file test harness + render/test/update CLI commands

**Objective:** Implement the golden-file comparison harness (render a team, diff its output tree plus a JSON snapshot of the resolved template context against `testdata/golden/<team>/`, with an update mode that overwrites fixtures) and wire it to three `kirokit` subcommands: `render [team]` (render only, no golden comparison), `test [team]` (render + diff — the CI entry point, non-zero exit on mismatch), and `update [team]` (render + overwrite goldens). Omitting `[team]` processes every directory under `teams/`.

**Files:**

- Create: `internal/golden/golden.go`
- Create: `internal/golden/golden_test.go`
- Create: `cmd/kirokit/main.go`
- Create: `cmd/kirokit/commands/root.go`
- Create: `cmd/kirokit/commands/render.go`
- Create: `cmd/kirokit/commands/test.go`
- Create: `cmd/kirokit/commands/update.go`

**Key Decisions / Notes:**

- `golden.Compare(team string) (diffs []string, err error)`: renders the team (Task 3) into a temp dir, does an exact byte-for-byte comparison against `testdata/golden/<team>/rendered/` per file (missing/extra files are diffs too), plus a normalized JSON-marshal comparison of the `render.Context` used against `testdata/golden/<team>/context.json`.
- `golden.Update(team string) error`: same render, but writes the output tree to `testdata/golden/<team>/rendered/` and context to `testdata/golden/<team>/context.json`, removing stale files no longer produced.
- CLI built with `github.com/spf13/cobra`; root command `kirokit`; each subcommand takes an optional positional `[team]`, defaulting to every directory under `teams/`.
- `test` prints a per-team, per-file diff summary to stderr and exits non-zero on any mismatch.

**Definition of Done:**

- [ ] `kirokit render <team>` writes a correctly rendered tree for a fixture team to a caller-specified `--out` directory without touching `testdata/golden/`
- [ ] `kirokit test <team>` exits 0 when goldens are current, and fails listing the changed file after a deliberate template mutation
- [ ] `kirokit update <team>` regenerates goldens such that a subsequent `kirokit test <team>` passes
- [ ] Verify: `go build ./... && go test ./... -v`

### Task 6: `onboard` command (new-team scaffolding)

**Objective:** Implement `kirokit onboard <team>`, which scaffolds `teams/<team>/` from an embedded starter layout (one example agent with an MCP server template, plus empty `hooks/`, `prompts/`, `skills/`, `steerings/` directories), substituting the given team name into the scaffolded `manifest.yaml`. Gives new teams a consistent starting point per the PRD's onboarding flow; hand-copying another team's directory remains equally valid and unaffected.

**Files:**

- Create: `internal/scaffold/scaffold.go`
- Create: `internal/scaffold/starter/manifest.yaml.tmpl`
- Create: `internal/scaffold/starter/agents/example-agent/agent.json.tmpl`
- Create: `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl`
- Create: `internal/scaffold/starter/hooks/.gitkeep`
- Create: `internal/scaffold/starter/prompts/.gitkeep`
- Create: `internal/scaffold/starter/skills/.gitkeep`
- Create: `internal/scaffold/starter/steerings/.gitkeep`
- Create: `internal/scaffold/scaffold_test.go`
- Create: `cmd/kirokit/commands/onboard.go`

**Key Decisions / Notes:**

- Embed the starter directory via `//go:embed starter` (`embed.FS`) in `internal/scaffold`, so `onboard` works regardless of the current working directory.
- `Scaffold(destTeamsDir, teamName string) error` copies the embedded tree to `teams/<teamName>/`, refuses to overwrite an existing `teams/<teamName>/` (clear "team already exists" error), and does simple `{{ .Team }}` substitution on `manifest.yaml.tmpl` only — the starter's other `.tmpl` files are left as literal templates for the team to fill in, not pre-rendered.
- `.gitkeep` placeholders let `onboard` materialize genuinely empty category directories (git does not track empty directories).

**Definition of Done:**

- [ ] Running `kirokit onboard` against a scratch teams directory with a test team name creates that team's `manifest.yaml` (with the given team name), `agents/example-agent/agent.json.tmpl` + `env/mcp.env.tmpl`, and empty `hooks/`, `prompts/`, `skills/`, `steerings/` directories
- [ ] Running `kirokit onboard` a second time for the same team name fails with a clear "already exists" error and leaves the existing directory unmodified
- [ ] Verify: `go test ./internal/scaffold/... -v`

### Task 7: `install`/`uninstall` commands

**Objective:** Implement `kirokit install <team>` and `kirokit uninstall <team>`, which render a team's config with `KiroHome` resolved to `$KIRO_HOME` (or `~/.kiro` when unset) and copy the rendered tree plus each agent's encrypted SOPS secrets file(s) into that location, tracking exactly what was written so `uninstall` removes only those files.

**Files:**

- Create: `internal/install/install.go`
- Create: `internal/install/install_test.go`
- Create: `cmd/kirokit/commands/install.go`
- Create: `cmd/kirokit/commands/uninstall.go`

**Key Decisions / Notes:**

- `ResolveKiroHome() string`: reads `$KIRO_HOME`; if unset, returns `filepath.Join(os.UserHomeDir(), ".kiro")`.
- `Install(team, kiroHome string) error`: renders the team (Task 3) with `Context.KiroHome = kiroHome`, writes the output tree under `kiroHome` (e.g. `kiroHome/agents/<agent>.json`), copies each `teams/<team>/agents/<agent>/env/secrets.enc.yaml` (if present) to `kiroHome/kirokit-secrets/<team>/<agent>/secrets.enc.yaml`, and writes an install manifest `kiroHome/.kirokit-installed/<team>.json` listing every absolute path this install wrote (rendered files + copied secrets + the manifest's own path).
- `Uninstall(team, kiroHome string) error`: reads `kiroHome/.kirokit-installed/<team>.json`, removes every listed path, then removes the manifest itself; returns a clear error (not a panic) if no manifest exists for that team.

**Definition of Done:**

- [ ] `kirokit install <team>` (with `KIRO_HOME` pointed at a temp dir) writes `agents/<agent>.json` whose MCP `args` entry points at `<temp-kiro-home>/kirokit-secrets/<team>/<agent>/secrets.enc.yaml`, and that file exists there with the same bytes as `teams/<team>/agents/<agent>/env/secrets.enc.yaml`
- [ ] `kirokit uninstall <team>` removes exactly the files the preceding `install` wrote (diff the temp KIRO_HOME's file listing before/after) and leaves an unrelated pre-existing file under that KIRO_HOME untouched
- [ ] Verify: `go test ./internal/install/... -v`

### Task 8: `install --knowledge-base` symlinked resource attachment

**Objective:** Extend `install` with flags to attach a local `knowledgeBase`-type resource to a specified agent at install time — given a source folder path and a mount point, `install` creates the mount point as a symlink to that folder (never copying its contents) and adds a corresponding `resources` entry to that agent's installed `agents/<agent>.json`. `uninstall` removes the symlink along with the rest of that team's installed files.

**Files:**

- Modify: `internal/install/install.go`
- Modify: `internal/install/install_test.go`
- Modify: `cmd/kirokit/commands/install.go`

**Key Decisions / Notes:**

- New `install` flags, required together (omitting all three, the common case, skips this step): `--kb-agent <name>`, `--kb-source <path>` (must exist and be a directory), `--kb-mount <path>` (absolute path outside the kit repo, e.g. `$KIRO_HOME/kirokit-kb/<team>/<agent>/`).
- `AttachKnowledgeBase(agentJSONPath, mountPoint, sourcePath string) error`: creates the symlink via `os.Symlink(sourcePath, mountPoint)` (erroring if `mountPoint` exists and is not a symlink this tool previously created, checked via `os.Lstat`/`os.Readlink`), then appends `{"type":"knowledgeBase","source":"file://<mountPoint>","name":"<agent>-kb"}` to that agent's `resources` array (creating the array if absent) and re-marshals the JSON back to the same file.
- The symlink path is recorded in that team's install manifest under a distinct `symlinks` list, so `uninstall` calls `os.Remove` on the link itself (never following it) rather than treating it as an ordinary file to delete.

**Definition of Done:**

- [ ] `kirokit install <team> --kb-agent <agent> --kb-source <dir> --kb-mount <path>` creates a real symlink at `<path>` pointing at `<dir>`, and `agents/<agent>.json` gains a `resources` entry whose `source` is `file://<path>`
- [ ] `kirokit uninstall <team>` removes the symlink (`os.Lstat` confirms it's gone) but leaves the real source directory and its contents completely intact
- [ ] Verify: `go test ./internal/install/... -v`

### Task 9: Example team "platform-team" (devops + platformdev) with shared SOPS/global-steering infra

**Objective:** Build the first of two complete example teams, `platform-team`, with two agents — `devops` and `platformdev` — each with its own SOPS-secured MCP server (echoing the PRD's own Jira/Confluence motivating example), plus team-shared hooks, prompts, skills, and a steering doc that opts into a new `global/steerings/security-baseline.md`. This task also creates the shared test-only SOPS infrastructure (`.sops.yaml`, test age key) that Task 10's team reuses. Golden fixtures are generated via `kirokit update platform-team`.

**Files:**

- Create: `global/steerings/security-baseline.md`
- Create: `.sops.yaml`
- Create: `testdata/sops-test-key/age-test-key.txt`
- Create: `teams/platform-team/manifest.yaml`
- Create: `teams/platform-team/agents/devops/agent.json.tmpl`
- Create: `teams/platform-team/agents/devops/env/secrets.enc.yaml`
- Create: `teams/platform-team/agents/platformdev/agent.json.tmpl`
- Create: `teams/platform-team/agents/platformdev/env/secrets.enc.yaml`
- Create: `teams/platform-team/hooks/lint-on-save.json.tmpl`
- Create: `teams/platform-team/prompts/jira-triage.md.tmpl`
- Create: `teams/platform-team/prompts/confluence-search.md.tmpl`
- Create: `teams/platform-team/skills/jira-search/SKILL.md.tmpl`
- Create: `teams/platform-team/skills/confluence-lookup/SKILL.md.tmpl`
- Create: `teams/platform-team/steerings/team-conventions.md.tmpl`
- Create: `testdata/golden/platform-team/rendered/agents/devops.json`
- Create: `testdata/golden/platform-team/rendered/agents/platformdev.json`
- Create: `testdata/golden/platform-team/rendered/hooks/lint-on-save.json`
- Create: `testdata/golden/platform-team/rendered/prompts/jira-triage.md`
- Create: `testdata/golden/platform-team/rendered/prompts/confluence-search.md`
- Create: `testdata/golden/platform-team/rendered/skills/jira-search/SKILL.md`
- Create: `testdata/golden/platform-team/rendered/skills/confluence-lookup/SKILL.md`
- Create: `testdata/golden/platform-team/rendered/steering/team-conventions.md`
- Create: `testdata/golden/platform-team/context.json`
- Modify: `env/mcp-versions.yaml`

**Key Decisions / Notes:**

- `testdata/sops-test-key/age-test-key.txt` is an age identity generated solely for this repo's own tests/CI; its public key is the sole recipient in `.sops.yaml`'s `creation_rules` for `path_regex: teams/.*/env/secrets\.enc\.yaml$` (a single rule that also covers Task 10's `data-team` secrets file). Documented as test-only (file header comment + README, Task 13) — never reused for a real secret.
- Both agents' `secrets.enc.yaml` are SOPS-encrypted, flat key/value, `.yaml` suffix (format inferred by suffix per the verified rule): `devops` holds `JIRA_API_TOKEN: "test-jira-token"`, `platformdev` holds `CONFLUENCE_API_TOKEN: "test-confluence-token"` — both committed as ciphertext.
- `devops/agent.json.tmpl` uses `{{ mcpVersion "jira-mcp" }}` and `{{ sopsCommand }}` / `{{ sopsArgsJSON "agents/devops/env/secrets.enc.yaml" "jira-mcp-server" "--port" "8080" }}` for its MCP `command`/`args`; `platformdev/agent.json.tmpl` does the analogous thing with `confluence-mcp` and its own secrets path. Both entries added to `env/mcp-versions.yaml` in this task.
- `steerings/team-conventions.md.tmpl` uses `{{ globalSteering "security-baseline" }}`; `manifest.yaml` sets `globalSteerings: [security-baseline]` — demonstrating the global-steering **opt-in** path (Task 10's team demonstrates opting out).
- Each agent's `resources`/`prompt` field references its matching prompt file via `file://prompts/<name>.md.tmpl`'s rendered path; each agent also references its matching skill via `skill://skills/<skill>/SKILL.md`.

**Definition of Done:**

- [ ] `kirokit test platform-team` passes cleanly against the golden fixtures generated in this task
- [ ] Rendered `agents/devops.json` and `agents/platformdev.json` contain no `${...}` placeholder anywhere, and each `mcpServers` entry's `args` matches the exact `["exec-env", "<path>", "<command>"]` shape from Task 4
- [ ] `SOPS_AGE_KEY_FILE=testdata/sops-test-key/age-test-key.txt sops -d teams/platform-team/agents/devops/env/secrets.enc.yaml` successfully decrypts to the fake `JIRA_API_TOKEN` value, proving the fixture is real SOPS ciphertext
- [ ] Verify: `kirokit test platform-team && go test ./... -v`

### Task 10: Example team "data-team" (javadev + godev)

**Objective:** Build the second example team, `data-team`, with two agents — `javadev` (a SOPS-secured Snowflake MCP server, directly echoing the Kiro issue #10536 scenario that motivates this whole kit) and `godev` (an MCP server with **no** secret, just a plain version-pinned `command`/`args`/`env`, proving the render pipeline handles both secret and non-secret MCP entries correctly) — plus its own hook, prompts, skill, and a steering doc that deliberately does **not** opt into any global steering. Reuses the shared `.sops.yaml`/test age key from Task 9. Golden fixtures generated via `kirokit update data-team`.

**Files:**

- Create: `teams/data-team/manifest.yaml`
- Create: `teams/data-team/agents/javadev/agent.json.tmpl`
- Create: `teams/data-team/agents/javadev/env/secrets.enc.yaml`
- Create: `teams/data-team/agents/godev/agent.json.tmpl`
- Create: `teams/data-team/hooks/format-on-save.json.tmpl`
- Create: `teams/data-team/prompts/warehouse-query.md.tmpl`
- Create: `teams/data-team/prompts/pipeline-status.md.tmpl`
- Create: `teams/data-team/skills/sql-helper/SKILL.md.tmpl`
- Create: `teams/data-team/steerings/data-conventions.md.tmpl`
- Create: `testdata/golden/data-team/rendered/agents/javadev.json`
- Create: `testdata/golden/data-team/rendered/agents/godev.json`
- Create: `testdata/golden/data-team/rendered/hooks/format-on-save.json`
- Create: `testdata/golden/data-team/rendered/prompts/warehouse-query.md`
- Create: `testdata/golden/data-team/rendered/prompts/pipeline-status.md`
- Create: `testdata/golden/data-team/rendered/skills/sql-helper/SKILL.md`
- Create: `testdata/golden/data-team/rendered/steering/data-conventions.md`
- Create: `testdata/golden/data-team/context.json`
- Modify: `env/mcp-versions.yaml`

**Key Decisions / Notes:**

- `manifest.yaml` sets `team: data-team`, `globalSteerings: []` — an explicit, valid opt-out; `steerings/data-conventions.md.tmpl` contains standalone content with no `{{ globalSteering ... }}` call.
- `javadev/env/secrets.enc.yaml`: SOPS-encrypted (reusing Task 9's `.sops.yaml` rule and test age key) holding `SNOWFLAKE_API_TOKEN: "test-snowflake-token"`. `agent.json.tmpl` uses `{{ mcpVersion "snowflake-mcp" }}` and `{{ sopsCommand }}` / `{{ sopsArgsJSON "agents/javadev/env/secrets.enc.yaml" "snowflake-mcp-server" "--warehouse" "analytics" }}`.
- `godev/agent.json.tmpl` has **no** `env/` directory (no secret needed) and renders a plain `mcpServers` entry directly: `"command": "pipeline-mcp-server", "args": ["--mode", "readonly"], "env": {"PIPELINE_MCP_VERSION": {{ mcpVersion "pipeline-mcp" | jsonString }}}` — no `sopsCommand`/`sopsArgsJSON` call at all, proving those funcs are opt-in per agent, not mandatory.
- `skills/sql-helper/SKILL.md.tmpl` is referenced (via `skill://skills/sql-helper/SKILL.md`) by **both** `javadev` and `godev` — one shared skill used by both of this team's agents, demonstrating that a team-level skill is not tied to a single agent, complementing `platform-team`'s pattern of one skill per agent.
- `snowflake-mcp` and `pipeline-mcp` entries added to `env/mcp-versions.yaml` in this task (alongside Task 9's `jira-mcp`/`confluence-mcp`).

**Definition of Done:**

- [ ] `kirokit test data-team` passes cleanly against the golden fixtures generated in this task
- [ ] Rendered `agents/godev.json` contains no `sops`/`exec-env` reference anywhere (confirms the non-secret MCP path never invokes the SOPS wrapper), while `agents/javadev.json`'s `mcpServers` entry matches the `["exec-env", "<path>", "<command>"]` shape from Task 4
- [ ] `SOPS_AGE_KEY_FILE=testdata/sops-test-key/age-test-key.txt sops -d teams/data-team/agents/javadev/env/secrets.enc.yaml` successfully decrypts to the fake `SNOWFLAKE_API_TOKEN` value
- [ ] Verify: `kirokit test data-team && go test ./... -v`

### Task 11: CI workflow (build, vet, lint, test)

**Objective:** Add a GitHub Actions workflow running on every pull request and push to `main`: `go build ./...`, `go vet ./...`, `golangci-lint run`, and `go test ./...` (which includes the golden-file comparisons from Tasks 5/9/10 as ordinary Go tests) — the CI entry point that fails when a template change breaks a golden fixture, per the PRD's "templating shall be tested in the ci" requirement.

**Files:**

- Create: `.github/workflows/ci.yaml`
- Create: `.golangci.yml`

**Key Decisions / Notes:**

- `actions/checkout@v7`, `actions/setup-go@v7` (`go-version: stable`), `golangci/golangci-lint-action` (current major) for lint.
- `.golangci.yml`: default linter set (`govet`, `staticcheck`, `errcheck`, `unused`) — no bespoke rule tuning.
- Triggers: `pull_request` and `push: branches: [main]`; a single job (no OS matrix needed — goreleaser, not this workflow, handles cross-platform builds).

**Definition of Done:**

- [ ] Workflow YAML is valid and its job runs build, vet, lint, and test in that order
- [ ] Verify: `go build ./... && go vet ./... && go test ./... -v` (the local equivalent of the CI job; `golangci-lint` verified by config review if not installed locally)

### Task 12: goreleaser + semantic-release release wiring

**Objective:** Wire automated releases: Conventional Commits merged to `main` drive `semantic-release` to compute the next version, generate a changelog, create the git tag, and (via `@semantic-release/exec`) invoke `goreleaser release`, which builds and publishes the `kirokit` binary for darwin and linux (amd64+arm64) as GitHub release archives.

**Files:**

- Create: `.goreleaser.yaml`
- Create: `.releaserc.yml`
- Create: `package.json`
- Create: `.github/workflows/release.yaml`

**Key Decisions / Notes:**

- `.goreleaser.yaml`: `version: 2`; `builds` with `main: ./cmd/kirokit`, `goos: [darwin, linux]`, `goarch: [amd64, arm64]`, `env: [CGO_ENABLED=0]`; `archives` using `ids:` (not `builds:`) and `formats: [tar.gz]` (not `format:`); `checksum`; `changelog: {use: github}`; `release: {github: {owner: mavogel, name: kiro-cli-kit}}`.
- `.releaserc.yml`: plugin list from Global Constraints; `@semantic-release/exec`'s `publishCmd` runs `echo "${nextRelease.notes}" > /tmp/release-notes.md && goreleaser release --release-notes /tmp/release-notes.md --clean`.
- `package.json` declares `semantic-release` + the plugins above as devDependencies only, so `npx semantic-release` resolves them in CI.
- `.github/workflows/release.yaml`: triggers on `push: branches: [main]`; single job; `actions/checkout@v7` (`fetch-depth: 0`), `actions/setup-go@v7`, `actions/setup-node@v5`, a `goreleaser/goreleaser-action@v7` step with `install-only: true` (so the `goreleaser` binary is on `PATH` for the `exec` plugin to invoke), then `npx semantic-release`; `permissions: {contents: write}`.

**Definition of Done:**

- [ ] `go run github.com/goreleaser/goreleaser/v2@latest check` validates `.goreleaser.yaml` with no errors
- [ ] `npx --yes -p semantic-release -p @semantic-release/commit-analyzer -p @semantic-release/release-notes-generator -p @semantic-release/changelog -p @semantic-release/git -p @semantic-release/exec semantic-release --dry-run` runs without a config-parsing error
- [ ] Verify: both commands above complete without a config/parse error (a full dry run needs real git history/CI env vars this local check won't have — see Out of Scope)

### Task 13: README documentation

**Objective:** Document the kit end to end: onboarding a team, the `manifest.yaml` schema, opting into (or out of) global steerings, bumping MCP versions centrally, setting up `.sops.yaml`/an engineer's own age key for secrets, running/updating golden tests locally, `install`/`uninstall` usage (including `--kb-agent`/`--kb-source`/`--kb-mount`), and how the semantic-release + goreleaser release process works — referencing the `platform-team`/`data-team` examples throughout.

**Files:**

- Create: `README.md`

**Key Decisions / Notes:**

- Structure: Quickstart → Directory layout → `manifest.yaml` reference → Central `env/mcp-versions.yaml` → Secrets via SOPS (setup + `exec-env` mechanism, explicitly flagging Task 9's key as test-only/non-production) → Golden tests (render/test/update) → install/uninstall (incl. knowledge-base flags) → Releases (Conventional Commits → semantic-release → goreleaser). Point to `teams/platform-team/` and `teams/data-team/` as worked examples of secret and non-secret MCP servers and of opting in/out of global steerings.
- Reference exact command/flag names implemented in Tasks 5–8; don't invent new ones.
- Trivial: pure documentation, reviewed by reading the rendered Markdown and cross-checking every named command/flag against `cmd/kirokit/commands/*.go`.

**Definition of Done:**

- [ ] README.md has a working section for each of: onboard, render, test, update, install (with `--kb-*` flags), uninstall, and the release pipeline, with no leftover placeholder or unfinished-section text
- [ ] Verify: `grep -c "^## " README.md` shows at least 8 sections, and a manual cross-check confirms every `kirokit` command/flag named matches an actual cobra command/flag in `cmd/kirokit/commands/`

## Deviations

- Task 7 (tactical): `install`/`uninstall` take an explicit `Roots{TeamsRoot, GlobalRoot, EnvVersionsPath}` struct (mirroring `golden.Inputs`) rather than the plan's shorthand `Install(team, kiroHome string)` — `install` must actually render team templates via `render.RenderTeam`, and those roots are not otherwise reachable from the `internal/install` package.
- Tasks 3/9/10 (tactical): a secrets path passed to `sopsArgsJSON` must be `.KiroHome`-relative and absolute (e.g. `{{ sopsArgsJSON (printf "%s/kirokit-secrets/<team>/<agent>/secrets.enc.yaml" .KiroHome) "<cmd>" "<args...>" }}`), not the bare repo-relative literal shown in Task 9/10's original Key Decisions (`"agents/jira-agent/env/secrets.enc.yaml"`). This was discovered building and testing Task 7's `install`: the secrets path baked into a rendered `mcpServers` entry must resolve at Kiro CLI runtime (under the installed `$KIRO_HOME`), and the plan's own Context for Implementer already establishes that every such path must be a literal, fully-resolved value — a bare repo-relative string never resolves correctly once installed. Tasks 9 and 10's agent templates use the corrected pattern.

## Goal Verification

### Truths

1. No plaintext secret value appears in any file under version control (rendered golden output, source templates, or `.sops.yaml`) — the only committed secret material anywhere in the repo is SOPS ciphertext (spans Tasks 3, 4, 9, 10).
2. `kirokit test` produces byte-identical results regardless of the machine or `$KIRO_HOME` it runs on, because golden rendering always uses the fixed `/kiro-home` sentinel and never the caller's real environment (spans Tasks 3, 5, 7).
