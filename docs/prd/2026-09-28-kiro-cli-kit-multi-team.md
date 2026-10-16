# Kiro CLI Kit — Multi-Team Config Templating & Release Kit

Created: 2026-09-28
Author: info@manuel-vogel.de
Agent: Claude Code
Category: Infrastructure
Status: Final
Research: Standard

## Problem Statement

Multiple teams use Kiro CLI v3 and each needs its own set of agents, hooks, prompts, skills, and steering files, plus a consistent pin on which MCP server versions it runs against. Today there is no shared repository or tooling for this: teams would either hand-author raw Kiro CLI config JSON/Markdown per team with no reuse, or one team's config would silently drift from another's with no way to verify the rendered output is still valid.

This is made worse by an active Kiro CLI bug: Kiro's own `${env:VAR}` / `${VAR}` placeholder substitution inside MCP server config does not reliably resolve (see [kirodotdev/Kiro#10536](https://github.com/kirodotdev/Kiro/issues/10536), where env-var substitution silently fails for MCP auth config and forces users toward hardcoding secrets). Relying on Kiro's native substitution today means either broken MCP server auth or config with hardcoded values. Until that's fixed upstream, this kit needs to own substitution itself: the templating engine must render fully resolved, literal values into each agent's config and env files at build time, so no `${...}` placeholder is ever left for Kiro CLI to fail to resolve at runtime.

We need a single "kit" repository where each team owns its own config sources, can optionally pull in shared/global building blocks (steerings in particular), renders everything through a templating engine, and gets that rendering verified in CI via golden-file tests before every release.

## Core User Flows

### Flow 1: A team updates its own config
1. A team member edits a template under `teams/<team>/` (e.g. adds a new hook, changes an agent's tool list, or updates an agent's own MCP env file under `teams/<team>/agents/<agent>/env/`).
2. They run the kit's CLI locally to render that team's templates into the final Kiro CLI v3 file tree (`kirokit render <team>`).
3. They run the golden test suite for that team (`kirokit test <team>` or `go test ./...`), which diffs the freshly rendered output and resolved template context against the committed golden files.
4. If the diff is intentional, they regenerate goldens (`kirokit update <team>`), review the diff, and commit both the template change and the updated goldens together.
5. They open a PR; CI re-runs the full golden suite across all teams to confirm nothing else broke.

### Flow 2: Onboarding a new team
1. A new team runs `kirokit onboard <new-team>`, which scaffolds `teams/<new-team>/` from the kit's starter layout (a `manifest.yaml` skeleton and a starting `agents/`, `hooks/`, `prompts/`, `skills/`, `steerings/` structure) — equivalent in spirit to copying an existing team's directory, but consistent and scripted; copying an existing team by hand remains just as valid since duplication across teams is expected and accepted.
2. They edit `teams/<new-team>/manifest.yaml` to confirm the team name, and to opt into whichever `global/steerings/*` entries apply to them (or none).
3. They edit/replace the scaffolded agents, hooks, prompts, skills, and steering templates to match their needs.
4. They run `kirokit update <new-team>` to render and generate the initial golden files for the new team, then commit.

### Flow 3: Bumping a shared MCP server version
1. Someone updates the pinned version/image tag for an MCP server type in the central `env/mcp-versions.yaml`.
2. Anyone renders/tests affected teams; golden-file diffs show exactly which agents' rendered `mcpServers` block (embedded in that agent's own config) changes as a result, since each agent's config template pulls the shared version pin plus its own agent-specific env values.
3. Reviewers see the diff per affected agent in the PR and approve or request a team- or agent-specific override.

### Flow 4: Injecting a secret into an agent's MCP server (e.g. a personal Jira/Confluence token)
1. A team adds a secret key (e.g. `JIRA_API_TOKEN`) to that agent's SOPS-encrypted `teams/<team>/agents/<agent>/env/secrets.enc.yaml`, encrypted against the recipients (age/PGP public keys or KMS ARNs) declared for that path in `.sops.yaml`; the ciphertext is committed like any other file.
2. `kirokit render` does not decrypt anything and never writes a plaintext secret into any rendered or golden file. Instead it renders that agent's MCP server `command` as a thin wrapper that runs the real MCP server command through `sops exec-env teams/<team>/agents/<agent>/env/secrets.enc.yaml -- <actual command> <args...>`, so decryption happens at MCP server process-launch time, on whichever machine starts it — sidestepping Kiro CLI's own unreliable `${VAR}` substitution entirely (see Problem Statement).
3. Non-secret env values (e.g. the shared MCP version pin from `env/mcp-versions.yaml`) are still rendered as literal values directly, since they carry no confidentiality requirement.
4. Each engineer (or CI, where a test-only key is used) must already hold a SOPS-compatible decryption key that is a recipient on that file; if they don't, `sops exec-env` fails loudly with a clear error rather than silently resolving to an empty/broken value.
5. Golden tests verify the rendered wrapper command and the encrypted file's path/shape, never a decrypted value — ciphertext is safe to check into golden fixtures.

### Flow 5: Installing/uninstalling a team's config locally
1. An engineer runs `kirokit install <team>`, which renders that team's templates (same rendering path as `render`/`test`) and copies the resulting file tree into the local Kiro CLI config location — `$KIRO_HOME` if set, otherwise Kiro CLI's documented default (`~/.kiro/`), matching Kiro's own `KIRO_HOME` override convention.
2. This writes real, usable agent/hook/steering/skill files (and each agent's `sops exec-env`-wrapped MCP command) directly where Kiro CLI reads them — no manual copying.
3. `kirokit uninstall <team>` removes exactly the files that team's install previously placed under the same `$KIRO_HOME`-aware location, without touching any other team's installed files or anything Kiro CLI itself manages.
4. This is a local, same-machine file placement step, not an automated push to another git repository — it doesn't overlap with the earlier "no auto-sync to downstream repos" scope cut.

### Flow 6: Cutting a release
1. Changes to the kit's own tooling (the Go CLI, templating engine, shared render/test logic) are merged to `main` using Conventional Commits.
2. CI runs the full test suite (unit tests + golden tests across all teams) as a required check before merge.
3. On merge to `main`, semantic-release computes the next version from commit history, updates the changelog, and tags the release; goreleaser then builds and publishes the `kirokit` binary for darwin and linux (no other OS targets in v1) and attaches the archives to the GitHub release — no manual version bumping or manual builds.

### Flow 7: Attaching a local knowledge base to an agent at install time
1. An engineer runs `kirokit install <team>` with a flag identifying an agent, a local folder path to index (the knowledge base source), and a mount point.
2. `install` adds a `knowledgeBase`-type entry to that agent's installed `resources` (per Kiro CLI's `resources` schema: `type`, `source`, `name`, plus Kiro's own `indexType`/`autoUpdate` fields), pointing at the given mount point rather than embedding the raw folder path directly.
3. Instead of copying the folder's contents into the install location, `install` creates the mount point as a symlink to the given source folder — the knowledge base's real content stays wherever the engineer keeps it (e.g. a separate docs checkout) and is not duplicated or vendored into `$KIRO_HOME`.
4. `kirokit uninstall <team>` removes the symlink it created along with the rest of that team's installed files, without touching the real folder the symlink pointed to.
5. This is a per-install, local-machine customization layered on top of the team's rendered config — it does not require the source folder path to exist in, or be committed to, the kit repository, and it is not captured by golden-file tests (which cover the committed, team-authored templates, not ad-hoc local install-time flags).

## Scope

### In Scope
- A Go CLI (`kirokit`) that renders per-team template sources into the final Kiro CLI v3 config tree: one config file per agent under `agents/<agent>/agent.json` (with its `mcpServers` block embedded, per Kiro CLI's schema), `hooks/*.json`, `steering/*.md`, `skills/**/SKILL.md`, and prompt Markdown files referenced from agents via `file://`.
- A consistent template-file naming convention: every templated source file uses a double extension of `<real-extension>.tmpl` (e.g. `agent.json.tmpl`, `hook.json.tmpl`, `steering.md.tmpl`, `SKILL.md.tmpl`, `env/*.env.tmpl`), so the target format stays visible in the filename and the renderer can discover templates via a single `**/*.tmpl` glob and derive each output filename by stripping the trailing `.tmpl`. A non-templated file (e.g. an already-encrypted `secrets.enc.yaml`) keeps its plain extension and is copied through untouched.
- A defined `kirokit` subcommand surface: `onboard <team>` (scaffold a new team directory from the starter layout), `render [team]` (render templates to an output tree), `test [team]` (render and diff against committed golden fixtures, non-zero exit on mismatch — the CI entry point), `update [team]` (render and overwrite the committed golden fixtures), `install <team>` (render and copy the result into the local Kiro CLI config location, honoring `$KIRO_HOME` with Kiro CLI's own default fallback), and `uninstall <team>` (remove exactly what that team's `install` previously placed, same `$KIRO_HOME`-aware location).
- `install` supports attaching a local `knowledgeBase`-type resource to a specified agent at install time, given a source folder path and a mount point: the mount point is created as a symlink to that folder (not a copy), and the agent's installed `resources` entry references the mount point. `uninstall` removes the symlink along with the rest of that team's installed files.
- `kirokit` built and released via goreleaser, producing binaries/archives for darwin and linux only in v1 (no Windows target), attached to each semantic-release GitHub release.
- A `teams/<team>/` tree per team, each owning its own templates end-to-end; copy-paste duplication between teams is an accepted, supported pattern (not something the tooling tries to eliminate).
- Each agent gets its own folder, `teams/<team>/agents/<agent>/`, containing that agent's config template and its own MCP env file(s) (`env/*.env.tmpl`) — since an MCP server is not a standalone/shared resource in this kit, it belongs to and is configured within the single agent that uses it (e.g. env vars loaded into a Docker-container-based MCP server).
- A `teams/<team>/manifest.yaml` per team (registry-style) declaring the team's identity and which `global/*` assets (steerings in particular) it opts into; global steerings live once under `global/steerings/` and are pulled in by reference, not copied, for teams that choose to use them.
- A central `env/mcp-versions.yaml` that pins each MCP server type's version/image tag once for all teams; each agent's `env/*.env.tmpl` template reads the relevant pin from it at render time and combines it with that agent's own env values, so every agent using a given MCP server stays on a consistent version unless it deliberately overrides it.
- A SOPS-based secret injection strategy for agent MCP env values (e.g. a user's personal Jira/Confluence API token): each agent's secrets live in an encrypted `teams/<team>/agents/<agent>/env/secrets.enc.yaml`, safe to commit as ciphertext. The kit renders that agent's MCP server `command` as a wrapper invoking `sops exec-env <that file> -- <real command>`, so decryption happens at MCP server process-launch time on a machine that already holds the right SOPS key — never as a template-render-time step, and never producing a plaintext secret in any rendered or golden file.
- Golden-file tests per team that snapshot both the final rendered output tree and the resolved template context (merged manifest + env values) used to produce it, with a supported "update goldens" workflow.
- CI (GitHub Actions) that runs the golden-file suite (and Go lint/vet/unit tests) on every PR, across all teams, as a required check.
- semantic-release wired to Conventional Commits on `main`, producing versioned releases and a changelog for the kit's own tooling.
- A top-level README documenting: how to add/onboard a team, the manifest schema, how global steerings are opted into, how to bump MCP versions centrally, and how to run/update golden tests locally.

### Explicitly Out of Scope
- Any modification to Kiro CLI itself, or validation that runs the actual `kiro` binary against rendered output — the kit validates rendered file *shape* via golden files, not live Kiro CLI behavior. Why: no CI-available Kiro CLI runtime to validate against; golden files are the agreed-upon proxy for correctness.
- Automatically pushing/syncing rendered config into each team's actual downstream project repo. Why: not requested; teams copy or apply the rendered output themselves for now. Can become a follow-up PRD if teams want continuous sync.
- Generating, distributing, or rotating engineers' own SOPS decryption keys (age/PGP key pairs, cloud KMS access grants). Why: the kit assumes engineers already have SOPS configured with access to the right keys (per the user's stated practice); it only defines where each agent's encrypted secrets file lives and how it's wired into that agent's rendered MCP command, not key lifecycle management.
- Any secrets backend other than SOPS (e.g. direct HashiCorp Vault or AWS Secrets Manager integration, a secrets UI). Why: SOPS file-based encryption is the stated v1 assumption; a different backend is a follow-up if teams need one.
- A web UI or dashboard for managing teams, manifests, or viewing rendered diffs. Why: not requested; git + CLI + PR review is the interface for v1.
- A Python (or any second-language) implementation of the templating engine. Why: Go was chosen as the single implementation language for this kit; introducing a second runtime is unnecessary duplication.
- Automatic/forced inheritance from global config (e.g. a team silently getting new global steerings without updating its manifest). Why: the registry/manifest model is opt-in by design so teams control exactly what they pull in; forced inheritance would reintroduce the drift risk this kit exists to make visible.

## Technical Context

- **Kiro CLI v3 config locations (verified against kiro.dev docs):** project scope reads from `.kiro/agents/*.json`, `.kiro/steering/*.md`, `.kiro/skills/**/SKILL.md`, `.kiro/hooks/*.json`; global scope mirrors this under `~/.kiro/` (overridable via `KIRO_HOME`). Agents follow "closest scope wins" precedence; steering, skills, and hooks merge across scopes. The kit's per-team `teams/<team>/` output trees mirror this same shape so rendered output can be dropped into a team's `.kiro/` directly. `kirokit install`/`uninstall` reuse this exact `KIRO_HOME` convention rather than inventing a separate install-location variable.
- **`kirokit install`/`uninstall` are local, same-machine file operations**, not a push to another repository: `install <team>` writes the team's rendered tree under `$KIRO_HOME` (or Kiro CLI's own documented default when unset) and `uninstall <team>` removes exactly what that install placed there, leaving any other team's installed files and anything Kiro CLI itself manages untouched.
- **Release build target:** `kirokit` is built and published with goreleaser for darwin and linux only in v1; no Windows build is produced or tested.
- **Kiro CLI's `resources` field accepts a `knowledgeBase` entry** with `type: "knowledgeBase"`, a `source` (e.g. `file://./docs`), a `name`, and Kiro-specific `indexType`/`autoUpdate` fields (per the agent configuration reference). `install --knowledge-base`-style attachment targets this existing schema field rather than inventing a new one; symlinking the mount point (vs. copying) keeps the real content out of `$KIRO_HOME` and out of the kit's own git history, since it may be large, local-only, or itself managed elsewhere.
- **MCP servers are per-agent, not a standalone shared resource, by this kit's design.** Kiro CLI's agent schema supports an `mcpServers` field embedded directly in each agent's own JSON config (in addition to an optional project-level `.kiro/settings/mcp.json` that agents can opt into via `includeMcpJson`). This kit deliberately does not use the standalone `settings/mcp.json` path: every MCP server belongs to exactly one agent, so its config — including any env vars a Docker-container-based server needs — is authored and rendered alongside that agent, under `teams/<team>/agents/<agent>/env/`.
- **No native "prompts" directory in Kiro CLI:** prompts are either inline in an agent's `prompt` field or referenced via `file://` to a Markdown file. The kit's `prompts/` template category therefore renders Markdown files that team agent templates reference by path — there is no separate Kiro-side "prompts" config type to target.
- **No documented MCP server version-pinning field in Kiro CLI's own agent schema.** The central `env/mcp-versions.yaml` is a kit-level concept, not a Kiro CLI feature — an agent's own `env/*.env.tmpl` is responsible for injecting the shared version pin into wherever that agent's MCP server's `command`/`args`/`env`/image reference needs it. The exact injection mechanics per server are an implementation detail for `/spec`.
- **Kiro CLI's own `${env:VAR}`/`${VAR}` placeholder substitution in MCP config is currently unreliable** ([kirodotdev/Kiro#10536](https://github.com/kirodotdev/Kiro/issues/10536)): substitution can silently fail, breaking MCP server auth. Constraint: rendered agent configs and env files must contain fully resolved, literal values — the kit's templating engine performs substitution once at render time, and no `${...}` runtime placeholder should remain in committed output for Kiro CLI to resolve on its own.
- **Secrets go through SOPS, not the templating engine's own substitution.** SOPS (`getsops/sops`) encrypts a structured file (YAML/JSON/env) against one or more recipients (age keys, PGP, or cloud KMS) declared per-path in a `.sops.yaml`; the ciphertext is safe to commit, and `sops exec-env <file> -- <command>` decrypts it and runs a child process with the decrypted keys set as env vars, without ever writing plaintext to disk. Using `exec-env` as the MCP server's launch mechanism means Kiro CLI never has to resolve a secret placeholder itself — it just launches the wrapper command the kit rendered, and SOPS handles decryption at that moment. This assumes engineers already have `sops` installed and hold a valid recipient key; the kit does not provision or manage those keys (see Out of Scope).
- **Existing code:** none — `/Users/mavogel/Developer/projects/kiro-cli-kit` is currently an empty, non-git directory. This is a greenfield kit.
- **Constraint:** rendered output must be structurally valid per the Kiro CLI v3 schemas above (agent JSON fields such as `tools`, `hooks`, `resources`, `mcpServers`; steering/skills as Markdown) since golden tests are the only correctness signal without a live Kiro CLI to validate against.

## Key Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| Templating language | Go | Single static-binary CLI, fast in CI, no runtime/venv needed for teams to render locally; user offered Go or Python. |
| Team/global config organization | Registry-style: each team has a `manifest.yaml` declaring which `global/*` assets (steerings) it opts into, alongside its own owned templates | Preserves full copy-paste duplication as a valid, supported pattern while still letting teams that want to stay in sync with shared steerings do so explicitly rather than by copy-drift. |
| Golden test scope | Snapshot both final rendered output *and* resolved template context (merged manifest + env values) | Makes it possible to tell, from a failing golden diff, whether a change came from a template edit, a manifest change, or a central env bump, not just that *something* changed. |
| Release trigger | Standard semantic-release on Conventional Commits merged to `main` | Matches the requested "cut releases via semantic-release" with the conventional, low-maintenance default rather than a custom path-scoped trigger. |
| Prompts handling | Rendered as Markdown files referenced by agent templates via `file://`, no dedicated Kiro-side "prompts" config type | Kiro CLI v3 has no native prompts directory — grounded in current kiro.dev docs rather than assumed. |
| MCP server ownership | Per-agent folder (`teams/<team>/agents/<agent>/env/`), not a standalone/shared MCP config | User confirmed an MCP server always belongs to exactly one agent (e.g. a Docker-container MCP server needs env values scoped to that agent), so config and env files are co-located with the owning agent rather than centralized. |
| Per-agent folder scope | Only the agent config and its own MCP env files live per-agent; prompts, hooks, skills, and steerings stay in team-level shared directories | User's explicit choice — keeps assets that are genuinely reusable across an agent's team shared, while still isolating the agent-specific, potentially sensitive MCP env values. |
| Why a templating engine at all (vs. Kiro's native `${VAR}` substitution) | Kit performs its own literal substitution at render time; no runtime placeholders in committed output | Kiro CLI's own env-var placeholder substitution in MCP config is currently buggy and can silently fail ([kirodotdev/Kiro#10536](https://github.com/kirodotdev/Kiro/issues/10536)); owning substitution in the kit avoids depending on that until it's fixed upstream. |
| Secret injection mechanism | SOPS-encrypted per-agent `env/secrets.enc.yaml`, decrypted only at MCP server process-launch time via a rendered `sops exec-env` wrapper command | Keeps real secret values (e.g. personal Jira/Confluence tokens) out of git and out of every rendered/golden file entirely, while also avoiding Kiro CLI's own buggy substitution path — ciphertext-only output stays fully golden-testable. User confirmed SOPS as the assumed engineering practice. |
| CLI command surface | `onboard`, `render`, `test`, `update`, `install`, `uninstall` as distinct `kirokit` subcommands | User's explicit requirement — separates "scaffold a team," "produce output," "verify against goldens," "refresh goldens," and "place/remove output on this machine" into distinct, individually scriptable operations instead of one command with many flags. |
| `install`/`uninstall` scope | Local, same-machine copy of a team's rendered output into `$KIRO_HOME` (or Kiro CLI's own default), not automated push to a separate downstream repo | User confirmed this is local file placement; keeps it consistent with the earlier decision to exclude cross-repo sync automation from v1, while still giving engineers a one-command way to materialize (and cleanly remove) a team's config locally. |
| Release build tooling | goreleaser, targeting darwin and linux only in v1 | User's explicit requirement; no Windows build is produced, distributed, or tested until requested. |
| Local knowledge-base attachment | `install` flag adds a Kiro `resources` entry of type `knowledgeBase` for a folder path, mounted via a symlink (not a copy) | User's explicit requirement; symlinking avoids vendoring potentially large or local-only doc folders into `$KIRO_HOME` or the kit repo, while still giving Kiro CLI a stable path to index. |
| Template file naming | Double extension, `<real-extension>.tmpl` (e.g. `agent.json.tmpl`), not a bare `.tpl` | Idiomatic for Go's `text/template` tooling, keeps the target format visible for editors/humans, and makes template discovery a single glob (`**/*.tmpl`) whose match strips cleanly to the output filename. |

## Research Findings

- Kiro CLI v3 configuration lives under `.kiro/` (project) and `~/.kiro/` (global, relocatable via `KIRO_HOME`), with dedicated subpaths for `agents/`, `steering/`, `skills/`, `hooks/`, and `settings/mcp.json`; most types merge across scopes, agents/MCP servers use closest-scope precedence. Source: [Kiro configuration scopes](https://kiro.dev/docs/configuration/).
- The Kiro CLI v3 agent config schema (`name`, `description`, `prompt`, `tools`, `allowedTools`, `excludedTools`, `permissions`, `resources`, `mcpServers`, `hooks`, `model`, etc.) has no dedicated "prompts" file type and no documented MCP server version-pinning field — both directly shaped the Technical Context and Scope sections above. Source: [Kiro custom agent configuration reference](https://kiro.dev/docs/custom-agents/configuration-reference/).
