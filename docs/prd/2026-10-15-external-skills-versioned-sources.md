# External Skills from Versioned Git Sources

Created: 2026-10-15
Author: info@manuel-vogel.de
Agent: Claude Code
Category: Feature
Status: Final
Research: Quick

## Problem Statement

Teams can only ship skills they author themselves, as single templated files under `teams/<team>/skills/`. There is no way to reuse a skill that already exists elsewhere, such as those in [anthropics/skills](https://github.com/anthropics/skills), or a skill kept in a company-private repository. Teams copy files by hand, so nobody records where a copy came from or which revision it is, and nobody can tell when it has drifted. The kit needs a way for a team to declare "use skill X from repository Y at revision Z" so the result is reproducible, reviewable, and works for private repositories that require authentication.

## Core User Flows

### Flow 1: A team owner adds an external skill
1. The owner adds an entry to their team's `manifest.yaml`: the repository URL, a ref (tag, branch, or commit), and the path of one skill directory inside that repository.
2. The owner runs the fetch command. kirokit resolves the ref to an exact commit, copies that skill directory verbatim into the repository, and records the source, resolved commit, and a content hash in a committed lock file.
3. The owner runs `kirokit update <team>` and reviews the diff of vendored files and goldens, then commits.
4. Every agent of that team now lists the skill as a `skill://` resource without any agent template change.

### Flow 2: A team owner moves to a newer version
1. The owner changes the entry's ref (for example a tag, or a new commit when the source publishes no tags), or re-runs fetch to move a branch ref forward.
2. Fetch resolves the new commit, replaces the vendored files, and rewrites the lock entry. Files the new revision no longer contains are removed.
3. The owner reviews the vendored-content diff in the pull request before merging.

### Flow 3: The source is private
1. The owner uses an SSH URL or an HTTPS URL for which the machine's git is already authenticated (SSH agent, `gh auth`, a credential helper).
2. Fetch works exactly as in Flow 1. kirokit stores no credential anywhere in the manifest, lock file, vendored content, or output.
3. If git cannot authenticate, fetch fails with an error naming the source URL and saying authentication failed; no partial files are left.

### Flow 4: A user installs or syncs the team
1. `kirokit install <team>` and `kirokit sync <team>` copy the vendored skills into `$KIRO_HOME/skills/<name>/` together with the team's other files.
2. No network access and no credentials are needed on the user's machine.
3. `uninstall` and `sync` remove or prune the skill's files like any other installed file.

### Flow 5: CI catches drift
1. A manifest entry, lock entry, and vendored content must agree. If someone edits the manifest without re-fetching, edits vendored files by hand, or deletes the lock entry, the offline check run by `kirokit test` / CI fails with a message saying which skill and which of the three disagrees.

## Scope

### In Scope
- A new manifest field on each team for external skills; each entry names repository URL, ref (tag, branch, or commit SHA), and the path of one skill directory in that repository. An optional name override resolves collisions.
- A fetch/update command that resolves refs to commits through the machine's `git`, vendors the skill directory byte-for-byte, and writes the lock file.
- A committed lock file recording, per skill: source, requested ref, resolved commit SHA, and content hash.
- Vendored skills are rendered into the same `skills/<name>/` output location as team-authored skills, including bundled scripts and reference files, and covered by the golden comparison.
- Every agent of a team that declares an external skill gets its `skill://` resource automatically.
- Offline drift check (manifest vs. lock vs. vendored content) in the existing `kirokit test` flow and therefore CI.
- Private sources through the machine's git authentication only.
- A render/fetch error when an external skill's name collides with a team-authored skill or another external skill of the same team.
- README documentation of the manifest field, the fetch command, the lock file, and private-source setup.

### Explicitly Out of Scope
- Fetching at install or sync time — the chosen model keeps install offline, credential-free, and byte-reproducible.
- Token environment variables and SOPS-encrypted tokens for source authentication — authentication is delegated to git, so kirokit never handles credentials.
- Non-git sources (HTTP tarballs, OCI artifacts, marketplaces/registries) — git covers both public and private use cases named.
- Automatic update proposals (scheduled bump PRs, dependency-bot style checks) — updating is an explicit owner action for now.
- Per-agent opt-in or opt-out of an external skill — all agents of a declaring team get it; a team needing a subset can author its own agent-level reference.
- Editing or patching vendored files, or dependency resolution between skills — vendored content is read-only and a fetch overwrites it.
- Validating that a vendored skill's content is correct or safe — the owner reviews the vendored diff in the pull request.

## Technical Context

- **Relevant architecture:** `internal/manifest` parses `teams/<team>/manifest.yaml` (today only `team` and `globalSteerings`); `internal/render` walks `teams/<team>/`, maps `skills/...` paths straight to the same path in the output tree, copies non-`.tmpl` files verbatim, and adds file-backed `resources` to every `agents/*.json` (global steerings already work this way, via `addResources`). `internal/install` writes the rendered tree into `$KIRO_HOME` and tracks every written path in an install manifest, which `uninstall` and `sync` use for removal and pruning. `internal/golden` compares the rendered tree to `testdata/golden/<team>/` byte-for-byte.
- **Constraints:**
  - Rendering must stay offline and deterministic; the only networked step is the explicit fetch command.
  - Vendored files must never be template-processed. The renderer treats any file ending in `.tmpl` as a template, and an external skill could contain such a file or text containing `{{`.
  - The installer currently writes every file with mode `0600`. External skills may ship scripts (for example `skills/pdf/scripts/` in anthropics/skills), so executable bits and the behavior of bundled scripts need an explicit decision in the plan.
  - Agents today reference team skills explicitly as `skill://{{ .KiroHome }}/skills/<name>/SKILL.md`; automatic wiring must produce the same URI form and not duplicate an entry an agent already lists.
  - `anthropics/skills` has no git tags (verified with `git ls-remote --tags`), so "version" must accept a commit SHA and a branch as well as a tag. Each of its skill directories is a folder with `SKILL.md` plus other files, and ships its own `LICENSE.txt`; vendoring must keep those files, and teams are responsible for checking that a source's license permits redistribution inside their repository.
  - Per the repository rules, generated goldens and `*.enc.yaml` are never hand-edited; vendored content gets the same protection (changed only by the fetch command, enforced by the drift check).
- **Existing code:** `internal/manifest/manifest.go`, `internal/render/render.go` (`mapOutputPath`, `RenderTeam`), `internal/render/globalsteerings.go` (`addResources`), `internal/install/install.go`, `internal/golden/golden.go`, `cmd/kirokit/commands/`, `teams/*/skills/`, `README.md`.

## Key Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| When external content is fetched | Vendored into the repo by an explicit fetch command; install and sync stay offline | Keeps render/CI deterministic and goldens meaningful; only the maintainer who updates needs network and credentials; content is reviewed in PRs |
| Private-source authentication | Delegate to the machine's `git` (SSH, `gh auth`, credential helper) | No secrets handled or stored by kirokit; works for any git host; reuses credentials users already have |
| Agent wiring | Manifest declares external skills; every agent of the team gets the resource automatically | Matches how global steerings work; avoids editing every agent template |
| What "version" means | A ref (tag, branch, or commit) resolved to an exact commit SHA recorded in a lock file | Source repos like anthropics/skills publish no tags; the SHA plus content hash is what makes the result reproducible |
| Granularity | One manifest entry vendors one skill directory | A source repo holds many skills; teams should take only what they use |
| Conflicts | Name collision with a team skill or another external skill is an error | Silent override would hide which content an agent actually loads |
