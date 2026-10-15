# External Skills from Versioned Git Sources Implementation Plan

Created: 2026-10-15
Author: info@manuel-vogel.de
Agent: Claude Code
Status: VERIFIED
Approved: Yes
Iterations: 1
Worktree: No
Type: Feature

## Summary

**Goal:** A team can declare `externalSkills` in its `manifest.yaml` (git URL + tag/branch/commit + path of one skill directory); `kirokit skills fetch` vendors the pinned content into the repo with a lock file, and every render, test, install and sync then ships the skill to all of that team's agents with no network access.

## Out of Scope

- Fetching at install/sync time, and token or SOPS-based source authentication: authentication is delegated to the machine's `git`, so kirokit never handles a credential.
- Non-git sources (tarballs, OCI, registries); automatic update PRs; per-agent opt-in; patching or editing vendored files; skill-to-skill dependency resolution; Git LFS content in a source repo.
- Committing a sample external skill: tests build local git fixtures, and `README.md` documents `anthropics/skills` as an example. Existing teams, goldens and `context.json` stay unchanged.

## Approach

**Chosen:** A new `internal/externalskills` package (lock file, content hash, offline verify, git-backed fetch) wired into `RenderTeam` (`internal/render/render.go`), which already adds `file://` resources to every agent through `addResources` for global steerings.
**Why:** Vendoring under `teams/<team>/external/` keeps render, `kirokit test` and install byte-deterministic and offline, and verification inside `RenderTeam` makes drift impossible to install, not just detectable in CI. It costs a second tree of committed files per team that must be excluded from the whitespace fixers and never hand-edited.

## Global Constraints

- Go 1.23 module: no `t.Chdir` (use a local chdir helper like `cmd/kirokit/commands/sync_test.go`).
- `go test ./...` and `golangci-lint run` stay green; existing teams' `testdata/golden/**` and `context.json` must not change (new manifest field is `omitempty` in JSON).
- Vendored files are copied byte-for-byte: never template-rendered, never edited by hand, symlinks rejected.
- No credentials in manifest, lock, vendored content, output, or error text: a source URL is rejected when it has any userinfo on `http`/`https` (covers `https://TOKEN@host/...` and `user:pass@`), or a password on any other scheme. `ssh://git@host/...` and `git@host:org/repo` stay valid.
- Commits follow Conventional Commits; `README.md` is updated in the same change as any command or manifest field.

## Context for Implementer

Layout decided here: vendored skill files live in `teams/<team>/external/skills/<name>/` and the lock file in `teams/<team>/external/skills.lock.yaml`. `RenderTeam` walks all of `teams/<team>/` and today errors on an unknown top-level category, so `external` must be skipped there (like `agents/*/env`) and its files injected into the output as `skills/<name>/...` separately. The effective skill name is the manifest `name` or, when absent, the last element of `path`. Content hash = SHA-256 over the sorted list of `relative-path NUL exec-flag NUL sha256(content)` lines, so renames and exec-bit changes are detected.

## Progress Tracking

- [x] Task 1: Manifest `externalSkills` field and validation
- [x] Task 2: Lock file, content hash and offline verification
- [x] Task 3: Git-backed fetch and vendoring
- [x] Task 4: Installer preserves the executable bit
- [x] Task 5: Render integration (inject, wire to agents, enforce lock)
- [x] Task 6: `kirokit skills fetch` command and end-to-end flow test
- [x] Task 7: Hook exclusions and documentation
- [x] Task 8: Manifest `all` / `exclude` fields and validation
- [x] Task 9: Lock records collection origin; Verify/Resolve and render use resolved skill names
- [x] Task 10: Fetch expands a collection entry into every skill under its path
- [x] Task 11: Document and exemplify `all: true`

## Implementation Tasks

### Task 1: Manifest externalSkills field and validation

**Objective:** Extend the team manifest with an optional `externalSkills` list so a team can declare a source repository, a ref and the path of one skill directory. Invalid declarations fail at load with an error naming the entry, so every later stage can trust the data.

**Files:**

- Modify: `internal/manifest/manifest.go`
- Test: `internal/manifest/manifest_test.go`

**Key Decisions / Notes:**

- Add `ExternalSkill{Name, Source, Ref, Path string}` (yaml `name`, `source`, `ref`, `path`) and `Manifest.ExternalSkills []ExternalSkill` tagged `yaml:"externalSkills" json:",omitempty"` so existing teams' `context.json` goldens are byte-identical.
- Add method `(ExternalSkill) EffectiveName()`: `Name`, else `path.Base(Path)`.
- `Load` validation: `source`, `ref`, `path` required; `path` relative, no `..`, not `.`; effective name is a single plain path element; names unique within the team; `source` rejected under the credential rule in Global Constraints (parse with `net/url`; scp-style `git@host:org/repo` has no scheme and is accepted). Follow the error style at `manifest.go:29`.

**Definition of Done:**

- [ ] A manifest with a valid entry loads, and `EffectiveName()` returns `pdf` for `path: skills/pdf` and the override when `name` is set.
- [ ] Each of these fails `Load` with an error naming the offending entry: missing `source`/`ref`/`path`, `path: ../x`, absolute `path`, duplicate effective names, name containing `/`, `https://u:p@host/r.git`, `https://TOKEN@host/r.git`, `ssh://u:p@host/r.git`; while `ssh://git@host/r.git` and `git@host:org/r.git` load.
- [ ] Manifests without `externalSkills` load unchanged.
- [ ] Verify: `go test ./internal/manifest -run . -count=1`

### Task 2: Lock file, content hash and offline verification

**Objective:** Create the `internal/externalskills` package's offline half: the lock file format, a deterministic directory hash and a verifier that checks manifest, lock and vendored content agree. Verification is what makes drift fail the render, so it is built and tested before anything consumes it.

**Files:**

- Create: `internal/externalskills/lock.go`
- Create: `internal/externalskills/verify.go`
- Test: `internal/externalskills/lock_test.go`

**Key Decisions / Notes:**

- Paths: `SkillsDir(teamDir)` = `<teamDir>/external/skills`, `LockPath(teamDir)` = `<teamDir>/external/skills.lock.yaml`.
- `Lock{Skills []LockEntry}`, `LockEntry{Name, Source, Ref, Path, Commit, SHA256}`; `ReadLock` returns an empty lock when the file does not exist; `WriteLock` writes sorted by name via `yaml.v3` (already a dependency, see `internal/manifest/manifest.go:9`).
- `HashDir(dir)` follows the definition in Context for Implementer; it errors on symlinks and non-regular files; exec flag is `mode&0o100 != 0`.
- `Verify(teamDir, entries []manifest.ExternalSkill)` returns an error listing every problem, each naming the skill and ending with the hint to run `kirokit skills fetch <team>`: no lock entry; lock `source`/`ref`/`path` differs from the manifest; vendored dir missing or lacking `SKILL.md`; hash mismatch; a lock entry or a `external/skills/*` directory with no manifest entry.
- With no entries and no `external/` directory, `Verify` returns nil (the common case for existing teams).

**Definition of Done:**

- [ ] `HashDir` is stable across runs and changes when a byte changes, a file is added, a file is renamed, or an exec bit flips; it errors on a symlink.
- [ ] `Verify` passes for a consistent fixture and fails with a message naming the skill for each case listed above.
- [ ] `ReadLock`/`WriteLock` round-trip, and `ReadLock` on a missing file returns an empty lock.
- [ ] Verify: `go test ./internal/externalskills -count=1`

### Task 3: Git-backed fetch and vendoring

**Objective:** Implement the single networked operation: for each manifest entry, resolve the ref to a commit through the machine's `git`, extract the skill directory verbatim into a staging area, and only if every entry succeeds replace the vendored files and rewrite the lock. A failure of any entry leaves the repository exactly as it was.

**Files:**

- Create: `internal/externalskills/fetch.go`
- Test: `internal/externalskills/fetch_test.go`

**Key Decisions / Notes:**

- `Fetch(teamDir string, entries []manifest.ExternalSkill) (Lock, error)`. Per entry: temp dir, `git init -q`, `git fetch --depth 1 <source> <ref>` and `git rev-parse FETCH_HEAD^{commit}` (peels annotated tags). If the shallow fetch fails (some hosts refuse a commit-SHA want), retry with a full fetch of heads and tags and resolve `<ref>^{commit}`.
- Extract with `git read-tree <commit>:<path>` then `git checkout-index -a -f --prefix=<staging>/` (run with `-c core.autocrlf=false`), not `git archive`: archive honours `export-ignore` from the source repo and would silently drop files. Then walk the staging dir: reject symlinks and non-regular files, require `SKILL.md` at the root, compute `HashDir`.
- Auth is delegated: run `git` with the caller's environment, stdin and stderr attached so SSH agent, credential helpers and prompts work. On git failure return an error naming the source URL, the ref and git's trimmed output, plus a one-line note that authentication is handled by git.
- Swap in: stage all entries under `teamDir/external/.fetch-*`, then replace each `external/skills/<name>` and remove vendored directories that have no manifest entry, then write the lock atomically (temp file + rename). Remove the staging area on every path.
- Preserve the executable bit from the checkout; do not touch file contents.

**Definition of Done:**

- [ ] Against a local git fixture repo (no network): a tag pin vendors that tag's tree and records its commit; a commit-SHA pin works; a test that forces the shallow fetch to fail (fixture repo configured to refuse unadvertised SHA wants, fetched with `protocol.version=0`, or an injected failure seam if git still serves it) proves the full-fetch fallback resolves the SHA pin; a branch ref follows a branch that moved on a second `Fetch`; re-running with no upstream change yields an identical lock and tree.
- [ ] A nested script file with the executable bit is vendored with the bit set and hashes differently from the same file without it.
- [ ] A missing `path`, a missing `SKILL.md`, a symlink in the tree, and an unreachable source each fail with a message naming the skill and source, and leave existing vendored files and the lock byte-identical.
- [ ] A fixture whose `.gitattributes` marks a vendored file `export-ignore` still vendors that file.
- [ ] Verify: `go test ./internal/externalskills -count=1`

### Task 4: Installer preserves the executable bit

**Objective:** `kirokit install` currently writes every file with mode 0600, which would strip the executable bit from vendored scripts such as `skills/pdf/scripts/*`. Preserve the owner-execute bit when copying the rendered tree into `$KIRO_HOME`.

**Files:**

- Modify: `internal/install/install.go`
- Test: `internal/install/install_test.go`

**Key Decisions / Notes:**

- In `copyRenderedTree` (`install.go`, the `WalkDir` callback) read the source file's mode: if `mode&0o100 != 0` write `0o700`, else `0o600`; call `os.Chmod` after `WriteFile` because `WriteFile` does not change the mode of an existing file (a re-install of a previously 0600 file must become executable).
- Group/other bits stay unset: the existing private-by-default behaviour is kept.
- Secrets copying (`copySecrets`) is unchanged.

**Definition of Done:**

- [ ] Installing a rendered tree containing an owner-executable file yields an installed file with owner-execute set; a non-executable file stays `0600`.
- [ ] Re-installing over an existing `0600` copy of a now-executable file makes it executable.
- [ ] Existing install, uninstall, sync and knowledge-base tests still pass.
- [ ] Verify: `go test ./internal/install -count=1`

### Task 5: Render integration: inject, wire to agents, enforce lock

**Objective:** Make `RenderTeam` verify the team's external skills against the lock, copy the vendored files into the output as `skills/<name>/...` verbatim, and add `skill://<KiroHome>/skills/<name>/SKILL.md` to every agent's resources. This is what turns vendored content into installed, referenced skills and what blocks rendering when anything has drifted.

**Files:**

- Modify: `internal/render/render.go`
- Create: `internal/render/externalskills.go`
- Test: `internal/render/externalskills_test.go`

**Key Decisions / Notes:**

- In `RenderTeam` (`render.go:23`): after `BuildContext`, call `externalskills.Verify(teamDir, ctx.Team.ExternalSkills)` and return its error; extend `agentResources` (`render.go:29`) to `global steerings + externalSkillURIs(ctx) + workspaceResources`; after the walk and `writeGlobalSteerings` (`render.go:108`) call a new `writeExternalSkills(tmpDir, teamDir, ctx)`.
- `mapOutputPath` (`render.go:152`): add `case "external": return "", true, nil` so the walk ignores the vendored tree and lock file instead of failing on an unknown category.
- `writeExternalSkills` copies each file byte-for-byte (never through `renderTemplate`, so a vendored `*.tmpl` file or `{{` text is preserved), writes `0o700` for owner-executable sources and `0o600` otherwise, and errors when `skills/<name>` already exists in the output (collision with a team-authored skill) naming both sources.
- URI form matches existing team skills: `skill://` + `path.Join(ctx.KiroHome, "skills", name, "SKILL.md")`, deduplicated by `addResources` (`globalsteerings.go:76`) when an agent template already lists it.
- Follow the `globalsteerings.go` structure (`globalSteeringURIs`, `writeGlobalSteerings`, `plainName`) for naming and path-escape checks.

**Definition of Done:**

- [ ] For a fixture team with one vendored skill (built offline with `externalskills` helpers): `skills/<name>/SKILL.md` and its other files appear in the output byte-identical, including a `*.tmpl` file and text containing `{{ }}`; an executable source file renders with the owner-execute bit.
- [ ] Every rendered agent lists `skill://<KiroHome>/skills/<name>/SKILL.md` exactly once, including an agent whose template already lists it and an agent with no `resources` key.
- [ ] Rendering fails with an error naming the skill for: a modified vendored file, a manifest entry with no lock entry, a vendored directory with no manifest entry, and a name that collides with a team-authored `skills/<name>`.
- [ ] A team without `externalSkills` renders exactly as before: all existing render and golden tests pass with no golden diff.
- [ ] Verify: `go test ./internal/render ./internal/golden -count=1`

### Task 6: `kirokit skills fetch` command and end-to-end flow test

**Objective:** Expose the fetch operation as a command and prove the complete user flow offline-after-fetch: fetch from a git source, render, install into a scratch `KIRO_HOME`, and uninstall, with the skill and its script present and removed correctly.

**Files:**

- Create: `cmd/kirokit/commands/skills.go`
- Modify: `cmd/kirokit/commands/root.go`
- Test: `cmd/kirokit/commands/skills_test.go`

**Key Decisions / Notes:**

- Cobra group `skills` with subcommand `fetch [team]` (all teams that declare entries when omitted, via `teamNames` at `root.go:44`); resolves `teams/` relative to the working directory like every other command. Load each manifest with `manifest.Load`, skip teams with no entries, call `externalskills.Fetch`, print one line per skill (`name  source@ref -> <short commit>`), and end with a reminder to run `kirokit update <team>` and review the diff.
- Reuse the `git`, `commitFile` and `chdir` helpers already in `cmd/kirokit/commands/sync_test.go` for the fixture repo; register the command next to `newSyncCmd()` in `root.go`.
- The test runs `fetch` through the cobra command, then `install.Install` with a scratch `KIRO_HOME` while the source repo is deleted, to prove install needs no source access.

**Definition of Done:**

- [ ] `kirokit skills fetch <team>` vendors the fixture skill, writes the lock, prints the resolved commit, and a second run changes nothing.
- [ ] After fetch, render succeeds; after deleting the source repo, `install.Install` still writes `skills/<name>/SKILL.md` and an executable script under the scratch `KIRO_HOME`, every installed agent references the skill, and `install.Uninstall` removes them.
- [ ] `fetch` on a team with no `externalSkills` exits 0 with a clear message; on an unreachable source it exits non-zero and leaves the team directory untouched.
- [ ] Verify: `go test ./cmd/... -count=1`

### Task 7: Hook exclusions and documentation

**Objective:** Keep the repo's own tooling from corrupting vendored content and document the feature. The whitespace fixers would rewrite upstream files and silently invalidate the content hash, so they must skip the vendored tree; the README and AGENTS.md must describe the new field, command, layout and private-source setup.

**Files:**

- Modify: `lefthook.yml`
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `teams/platform-team/manifest.yaml`

**Key Decisions / Notes:**

- `lefthook.yml`: add `teams/*/external/**` to the `exclude` lists of `trailing-whitespace` and `end-of-file-fixer`, and to `check-yaml`/`check-json` (upstream files are not ours to lint); extend the header comment.
- `README.md`: add `externalSkills` to the manifest reference near `README.md:63`; new "External skills" section covering the entry fields, `kirokit skills fetch`, the `teams/<team>/external/` layout and lock file, the update flow (change `ref`, fetch, review the diff, `kirokit update`), private sources via SSH or a git credential helper (no credentials stored), that refs without tags (as in `anthropics/skills`) use a commit SHA or branch, and a licence note (each source skill may carry its own `LICENSE.txt`, which is vendored).
- `AGENTS.md`: add `skills` to the `cmd/kirokit/` command list and a repository rule: never hand-edit `teams/*/external/**`; change `externalSkills` and re-run `kirokit skills fetch`.

**Definition of Done:**

- [ ] `lefthook run pre-commit --all-files` does not modify a file under `teams/*/external/` that has trailing whitespace and no final newline (checked on a scratch file).
- [ ] The README documents every manifest key, the command and its example, the layout, the update flow and private-source setup; command names match `kirokit skills fetch --help` output.
- [ ] The agent-asset parity check (the `--check` mode CI runs, implemented at scripts/sync-agent-assets.mjs:1) passes after the `AGENTS.md` edit.
- [ ] Verify: `go build -o kirokit ./cmd/kirokit && ./kirokit skills fetch --help`, then run the parity check described above

### Task 8: Manifest all / exclude fields and validation

**Objective:** Let one `externalSkills` entry stand for every skill under a directory, as requested at the review gate. Add `all` and `exclude` to the entry and make validation understand that a collection entry names a directory of skills, not one skill.

**Files:**

- Modify: `internal/manifest/manifest.go`
- Test: `internal/manifest/manifest_test.go`

**Key Decisions / Notes:**

- Add `All bool` (`yaml:"all" json:",omitempty"`) and `Exclude []string` (`yaml:"exclude" json:",omitempty"`) to `ExternalSkill` so existing goldens stay byte-identical (same approach as `ExternalSkills` at `manifest.go:19`).
- In `Validate()`: with `all: true`, `name` must be empty and the per-skill name checks are skipped; `exclude` is valid only with `all` and each element must be a plain single path element; `path` rules are unchanged (relative, not `.`), `source`/`ref` rules are unchanged.
- In `validateExternalSkills`: only single-skill entries take part in the duplicate-name check (a collection's skill names are known only after fetch; duplicates are caught in Task 9/10 against the lock).
- Add `(ExternalSkill).IsCollection()` returning `All`.

**Definition of Done:**

- [ ] `all: true` with `path: skills` and `exclude: [docx, pdf]` loads; `IsCollection()` is true and `EffectiveName()` is not used for it.
- [ ] `all: true` with `name` set, `exclude` without `all`, and an `exclude` element containing `/` or empty each fail `Load` naming the entry.
- [ ] Manifests without `all` load exactly as before, and existing teams' golden `context.json` files are unchanged.
- [ ] Verify: `go test ./internal/manifest ./internal/golden -count=1`

### Task 9: Lock records collection origin; Verify/Resolve and render use resolved skill names

**Objective:** A collection entry expands to skills the manifest does not list, so offline verification and rendering must learn the skill names from the lock. Add the origin of each lock entry, make `Verify` understand collections, and have render copy and wire the resolved skills instead of the manifest entries.

**Files:**

- Modify: `internal/externalskills/lock.go`
- Modify: `internal/externalskills/verify.go`
- Test: `internal/externalskills/lock_test.go`
- Modify: `internal/render/render.go`
- Modify: `internal/render/externalskills.go`
- Test: `internal/render/externalskills_test.go`

**Key Decisions / Notes:**

- `LockEntry` gains `From string` (`yaml:"from,omitempty"`): the collection `path` that produced it; empty for single-skill entries.
- New `Resolve(teamDir string, entries []manifest.ExternalSkill) ([]string, error)`: runs the existing checks and returns the sorted skill names that are vendored (explicit entries plus lock entries whose `From`, `Source` and `Ref` match a collection entry). `Verify` becomes a thin wrapper that discards the names.
- Collection checks: a collection entry with no lock entries is "missing from the lock file"; a lock entry whose `From` matches no collection entry, or whose name is in that entry's `exclude`, or whose source/ref differ, is drift; a name produced twice (explicit and collection, or two collections) is an error naming the skill; vendored directories not in the resolved set are drift.
- Render (`render.go:23`, `externalskills.go`): replace the `ctx.Team.ExternalSkills` loops in `externalSkillURIs` and `writeExternalSkills` with the names returned by `Resolve`, computed once in `RenderTeam` after `BuildContext`.

**Definition of Done:**

- [ ] `Resolve` returns the explicit names and every collection-derived name for a consistent fixture, and `Verify` still passes it.
- [ ] `Verify` fails naming the skill or collection for: collection with no lock entries, lock entry whose skill is in `exclude`, lock entry with no matching manifest entry, name produced by two entries, and a modified collection-derived file.
- [ ] A fixture team with one collection entry (three skills, one excluded in the manifest) renders `skills/<name>/` for exactly the lock's skills and adds one `skill://` resource per skill to every agent.
- [ ] All Task 2 and Task 5 tests still pass unchanged.
- [ ] Verify: `go test ./internal/externalskills ./internal/render ./internal/golden -count=1`

### Task 10: Fetch expands a collection entry into every skill under its path

**Objective:** Make `kirokit skills fetch` handle `all: true`: one git fetch of the collection directory, then every immediate subdirectory containing a `SKILL.md` (minus `exclude`) is vendored and locked as its own skill, with the same all-or-nothing behaviour as single skills.

**Files:**

- Modify: `internal/externalskills/fetch.go`
- Test: `internal/externalskills/fetch_test.go`
- Test: `cmd/kirokit/commands/skills_test.go`

**Key Decisions / Notes:**

- In `stage` (`fetch.go`): split the shared part (fetch, resolve commit, `read-tree` + `checkout-index` of `commit:path`) from the single-skill checks. For a collection, walk the checked-out tree's immediate subdirectories; keep those with a `SKILL.md`, drop `exclude`d ones, skip other directories, hash each, and emit one `LockEntry` per skill with `Path` = `<collection path>/<name>` and `From` = the collection path.
- Errors: a collection yielding no skills, an `exclude` naming a directory that does not exist under the path (a typo would silently vendor a skill the user meant to skip), a skill name produced twice across entries, and a symbolic link anywhere in the tree all fail the whole fetch with the skill or entry named, leaving the repository untouched.
- `Fetch` input validation (`Validate()`) and the `--` guards are unchanged; the swap-in and pruning loop works on the expanded skill list instead of the raw entries.
- The CLI (`cmd/kirokit/commands/skills.go`) already prints one line per lock entry, so it needs no change; the command test proves the collection flow end to end.

**Definition of Done:**

- [ ] Against a local git fixture with `skills/a`, `skills/b`, `skills/c` and a `skills/notes` directory without `SKILL.md`: one `all: true` entry vendors a, b, c, locks each with `From: skills` and the same commit, and ignores `notes`.
- [ ] `exclude: [b]` vendors only a and c, and removes a previously vendored b on re-fetch; an `exclude` naming a non-existent skill fails naming it and changes nothing.
- [ ] A collection that yields no skills, a name clash with an explicit entry, and a symbolic link in the tree each fail naming the entry and leave existing vendored files and the lock byte-identical.
- [ ] Through the CLI: `skills fetch` on a team with a collection entry, then render and install (source repository deleted) installs every vendored skill and every agent references each of them; uninstall removes them.
- [ ] Verify: `go test ./internal/externalskills ./cmd/... -count=1`

### Task 11: Document and exemplify all: true

**Objective:** Tell users how to take every skill from a repository, including the licence caveat for repositories that mix licences, and show it in the platform-team manifest example.

**Files:**

- Modify: `README.md`
- Modify: `teams/platform-team/manifest.yaml`

**Key Decisions / Notes:**

- `README.md`: add `all` and `exclude` to the manifest reference and the External Skills section (what `all` does, that `path` is the directory of skills, that new upstream skills appear only after re-fetching, that `exclude` names must exist, that the lock records each skill with its collection, and that name clashes are errors); keep the statement that a single entry without `all` vendors one skill. State that some skills in a collection may carry their own licence terms (for example `anthropics/skills`), so `exclude` them or check before vendoring everything.
- `teams/platform-team/manifest.yaml`: add a commented `all: true` example with `exclude` next to the existing single-skill example (comments only).

**Definition of Done:**

- [ ] The README documents `all`, `exclude`, the lock `from` field, the re-fetch behaviour and the licence caveat, and its manifest example matches the field names accepted by `Load`.
- [ ] The commented example in `teams/platform-team/manifest.yaml`, uncommented in a scratch project, loads and fetches the pinned `anthropics/skills` commit as a collection with `exclude`.
- [ ] `kirokit test` still passes with no golden change.
- [ ] Verify: `go build -o kirokit ./cmd/kirokit && ./kirokit test`

## Deviations

- Task 3 (tactical): the Task 6 end-to-end test showed `Fetch` mis-extracted files when given the CLI's relative `teams/<team>` path (git runs in a scratch directory) and left an empty `external/` after a failed first fetch → `Fetch` now resolves `teamDir` to an absolute path and removes an `external/` it created when the fetch fails; regression tests added to `internal/externalskills/fetch_test.go` (no new files).

- Tasks 1 and 3 (review-driven, tactical): the changes review found that a manifest `source` or `ref` starting with `-` reached `git fetch` as an option (reproduced: `--upload-pack=<cmd>` executed the command) → `ExternalSkill.Validate()` in `internal/manifest/manifest.go` now rejects option-like and transport-helper sources, non-git schemes, and unsafe refs; `Fetch` re-validates every entry and passes `--` before the positional arguments in `internal/externalskills/fetch.go`; regression tests added to `internal/manifest/manifest_test.go` and `internal/externalskills/fetch_test.go`; the README private-sources bullet lists the accepted source forms (`README.md`). Git's default already blocks `ext::`, so that part is defense in depth.

- Task 7 (user-agreed): at the review gate the user asked for a commented-out `anthropics/skills` example in the platform-team manifest → added a commented `externalSkills` example (public SHA-pinned entry and a private SSH entry) to `teams/platform-team/manifest.yaml`; comments only, so rendering and goldens are unaffected.

- Tasks 8-11 (user-agreed): at the review gate the user asked whether all skills of a repository can be taken and chose to add an `all: true` collection mode with an optional `exclude` list → added Tasks 8-11 (manifest fields, lock `from` field plus `Resolve`, fetch expansion, docs); `internal/manifest/manifest.go`, `internal/externalskills/{lock,verify,fetch}.go`, `internal/render/{render,externalskills}.go`, their tests, `cmd/kirokit/commands/skills_test.go`, `README.md` and `teams/platform-team/manifest.yaml` are in the affected Tasks' `Files:` blocks.

- Tasks 9 and 10 (review-driven, tactical): the iteration-1 changes review found that a single-skill entry listed before a colliding `all: true` entry failed with a raw staging `rename ... file exists` error attributed to the wrong entry, and that a collection ref drift was also reported as undeclared lock entries → `stage` now reserves skill names through a shared set so a clash is reported as `skill "a" is provided by more than one manifest entry` in either order, and `resolveCollection` marks a drifted collection's lock entries as accounted for; regression tests added to `internal/externalskills/fetch_test.go` and `internal/externalskills/lock_test.go` (files already in Tasks 9 and 10).

## Goal Verification

### Truths

1. In a scratch copy of the repository, declaring `anthropics/skills` at a pinned commit with `path: skills/skill-creator` for a team, then running `kirokit skills fetch <team>`, `kirokit update <team>`, and `kirokit install <team>` with a scratch `KIRO_HOME` and the network disabled for the last two, installs `skills/skill-creator/SKILL.md` and its bundled files, with scripts executable, and every installed agent lists the skill. Pointing the same entry at a private repository over SSH succeeds when `git ls-remote` on that URL succeeds, with no token appearing in any repository file.
2. Editing one vendored byte, or changing `ref` in the manifest without re-running fetch, makes `kirokit test` and `go test ./internal/golden` fail naming the skill.
3. A single `all: true` entry for `anthropics/skills` at a pinned commit with an `exclude` list vendors every other skill under `skills/`, each with its own lock entry; install (network off) places them all and every agent references each; re-fetching with a longer `exclude` removes the newly excluded skills.

## Verification Gaps

| Gap | Type | Severity | Affected Files | Fix Description |
|-----|------|----------|----------------|-----------------|
| No way to vendor every skill in a repository with one entry | Requirement (user-requested at review gate) | should_fix | `internal/manifest/manifest.go`, `internal/externalskills/*.go`, `internal/render/*.go`, `README.md`, `teams/platform-team/manifest.yaml` | Tasks 8-11: `all: true` / `exclude` collection entries |

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| A host refuses a shallow fetch by commit SHA | Medium | Fetch fails for SHA pins on that host | Task 3 retries with a full heads+tags fetch; its fallback is forced by a fixture test, and the Goal Verification run against `anthropics/skills` at a pinned commit exercises a real host |
| Upstream `.gitattributes` `export-ignore` drops files | Low | Vendored skill silently incomplete | Extract with `read-tree` + `checkout-index`, not `git archive`; Task 3 DoD fixture |
| Pre-commit fixers rewrite vendored files | High without mitigation | Hash mismatch after commit | Task 7 excludes `teams/*/external/**`; drift check fails render if it still happens |
| Vendored script loses its executable bit on install | High without mitigation | Skill scripts fail to run | Task 4 preserves the bit; Task 6 asserts it on the installed file |
