---
type: research
---

# kirokit sync, setup --branch and their git tests today

## Findings

### sync updates the checkout with fetch, then pull --ff-only or a tag checkout, and fails before any uninstall
**Answers:** 1

Order in `RunE` (`cmd/kirokit/commands/sync.go:36-114`):
1. Without a team, any `--kb-*` flag changed returns `--<flag> needs a team` (`sync.go:45-50`). With a team: parse KB flags and resolve KIRO_HOME (`sync.go:51-58`).
2. Resolve dir (`--repo`, else `repoBase`/KIROKIT_REPO, else ".") and `os.Chdir` into it (`sync.go:63-72`); `repoBase` is cleared (`sync.go:73`).
3. `git rev-parse --short HEAD` as "before" (error ignored, `sync.go:75`), then `updateCheckout(latestTag)` (`sync.go:76`).
4. `updateCheckout` (`sync.go:125-143`): always `git fetch --tags --prune --quiet` (`:126`). Without `--latest-tag`: `git pull --ff-only --quiet` (`:130`). With it: `git tag --list v* --sort=-v:refname` (`:133`), first line is the tag; empty gives `--latest-tag: no v* tag found after fetch` (`:138-139`); else `git checkout --quiet <tag>` (`:141`).
5. Prints `updated <before> -> <after>` only if short HEADs differ (`sync.go:79-82`).
6. No team: prints `checked out <ref> at <after>` and returns; ref is the exact-match tag, overridden by the branch name when `symbolic-ref --short -q HEAD` succeeds (`sync.go:84-90`). Nothing installed.
7. With a team: `install.Check` renders into a scratch dir and writes nothing to KIRO_HOME (`internal/install/install.go:63-70`), then `uninstallOthers` (prints `uninstalled X from home`, `cmd/kirokit/commands/install.go:58-64`), then `install.Sync` (prints `pruned <p>` per file) and `synced <team> into <home> at <after>` (`sync.go:100-113`).

Failure boundary: a failed fetch/pull/checkout returns at `sync.go:76-78`; an unknown or broken team fails in `install.Check` at `sync.go:100-102`. Both precede `uninstallOthers` at `:103`. The Long help states this (`sync.go:21-24`).

Git invocation: `gitOutput` runs `git` with combined output, trimmed, env stripped of GIT_DIR etc. via `externalskills.WithoutRepoGitEnv` (`sync.go:145-153`, `internal/externalskills/fetch.go:324-336`). Errors read `git <args>: <err>: <output>` (`sync.go:151`).

Situations (the code has no special handling for any of them; results are git's):
- Detached HEAD, default mode: `pull --ff-only` fails. Run: git prints "You are not currently on a branch." and exits 1 (ran on a clone checked out at a tag). So plain `sync` on a detached checkout errors and stops before install. `--latest-tag` works on a detached HEAD because it checks out by tag.
- Local changes: no stash/reset/clean is issued; `pull --ff-only` and `checkout` only fail where git refuses (diverged history, or overwriting dirty files). Inference from git semantics; the only test of failure is a diverged branch (`sync_test.go:179-183`).
- Unknown ref: sync has no branch/ref flag; the only ref it names is the computed newest `v*` tag (`sync.go:117-121`, flags). No-tag case is the error above.

### The no-team path reports the checkout; HEAD short hash before/after drives the "updated" line
**Answers:** 1

See steps 5-6. Output asserted exactly in `sync_test.go:107`.

### setup --branch validates by a small rule set and passes the name to git clone --branch
**Answers:** 3

- `--source` must match `sourcePattern` (https/ssh/file URL, `git@host:path`, absolute path, Windows drive path), else `kirokit setup: --source: use an https, ssh or file address, git@host:path, or an absolute path` (`cmd/kirokit/commands/setup.go:21`, `:43-45`).
- `--branch` is trimmed (`setup.go:46`). Only if the flag was passed (`Changed`), it is refused when empty, starting with `-`, or containing space/tab/CR/LF, with `kirokit setup: --branch: must be a branch or tag name, not empty, starting with '-' or containing whitespace` (`setup.go:47-49`). No other git ref-format checks (e.g. `..`, `~`) exist.
- Clone: `git clone [--branch <b>] -- <source> <dir>` (`setup.go:55-59`); stdout goes to the command output, stderr to the command stderr (`:61`). A clone failure returns `kirokit setup: git clone: <err>` (`:62-64`). `--target` must be empty/non-existent: enforced by git clone itself (test `setup_test.go:124,127`); no code check in setup.go.
- git behaviour: `--branch` accepts a branch or tag; a tag yields a detached HEAD. Ran: tag `v1` clones with "detached HEAD" notice; unknown name gives `fatal: Remote branch nope not found in upstream origin`, surfaced through the wrapped clone error plus git's stderr.
- With `--branch` success prints `checked out <b> in <dir>` and `next: export KIROKIT_REPO=<dir>`, with no tag lookup (`setup.go:66-69`).
- No-release case (no `--branch`): after clone, chdir, `updateCheckout(true)` (`setup.go:70-73`); if the error contains `no v* tag found` it prints `cloned into <dir>` plus `noReleaseNotice` ("This repository has no release yet (no v* tag). Ask your team lead to tag one.", `setup.go:23`) and exits successfully, clone kept on the default branch (`:77-78`); other errors fail (`:79-80`). The match is string-based on the error from `sync.go:139`, so the two files are coupled.

Reuse: setup shares `updateCheckout` and `gitOutput` with sync (`setup.go:73,75`); these are the only callers outside sync.go.

## How it is tested today

Both files are package `commands`, plain `go test`, real `git` subprocesses, no mocks. Commands from the repository root:
- One file's tests: `go test ./cmd/kirokit/commands -run 'TestSetup|TestSync|TestUpdateCheckout'` (the package has other test files; choose names with `-run`).
- Suite: `go test ./...` (or `task ci`).

Helpers (defined in `sync_test.go`, shared by setup_test.go): `git(t, dir, args...)` runs git with hooksPath=/dev/null, fixed identity, gpgsign off and GIT_* env stripped (`sync_test.go:12-34`); `chdir` with cleanup restore (`:36-46`); `commitFile` writes file `f` and commits (`:48-55`). Others live elsewhere: `runRoot` (`skills_test.go:14`), `writeTree` (`skills_test.go:25`), `thisRepo` (`root_test.go:11`), `installedRecords` (`install_test.go:52`).

Fixtures: origin is `t.TempDir()` with `git init -b main`, commits/tags created through `commitFile`/`git tag` (`sync_test.go:58-64`, `setup_test.go:11-17`); the remote is a local path, clone via `git clone origin .`. Branches via `git branch feature` (`setup_test.go:59`). Team fixtures use `writeTree` with a minimal demo team (`sync_test.go:136-141`). KIRO_HOME and KIROKIT_REPO set via `t.Setenv`.

Coverage: `updateCheckout` branch update plus version-sort tag (v1.10.0 over v1.0.0) (`sync_test.go:57-85`), no-tags error (`:214-225`); team-less sync output (`:87-114`); KB flags need a team (`:116-123`); other teams uninstalled first (`:125-160`); failure leaves installed team for unknown team and diverged pull (`:162-212`). Setup: newest tag (`setup_test.go:10`), no release (`:34`), `--branch` for branch, tag, main (`:54-83`), refused `--upload-pack=x`, `-b`, `has space`, `nonexistent` (asserts only err != nil, not text, `:85-97`), source pattern (`:99-114`), bad input (`:116-140`). Not covered: detached-HEAD plain sync, dirty tree, empty `--branch ""` text.

Windows: neither file references `runtime.GOOS`, `testutil` or any skip (grep returned no matches), so no test in them skips on Windows. AGENTS.md says Unix-shell tests skip via `internal/testutil`, not used here.

## Open questions

None.

## Code references

- `cmd/kirokit/commands/sync.go:14-154` (whole command, `updateCheckout` :125, `gitOutput` :145)
- `cmd/kirokit/commands/setup.go:16-92`
- `cmd/kirokit/commands/install.go:58-64` (uninstallOthers)
- `internal/install/install.go:63-70` (Check), `internal/install/sync.go:16` (Sync), `internal/install/installed.go:59` (UninstallOthers)
- `internal/externalskills/fetch.go:324-336`
- `cmd/kirokit/commands/sync_test.go`, `setup_test.go`, `skills_test.go:14,25`, `root_test.go:11`, `install_test.go:52`
- Churn, last 90 days: sync.go 3 commits, setup.go 2, the two test files 4 combined. Exhaustive for the three questions; desktop and README pointers in the brief were not read.
