---
answers: app
title: Check out a config repo branch
request: add the possibility to checkout a branch of the config repo in the cli and desktop-app
type: feature
depth: light
level: engineering
ui: true
overrides: []
stage: shipped
iterations: 2
verification_failures: 0
gate: null
created: 1793030325145
startup_owner: 4208fa65abb54844
layout: seven
profile: full
status: verified
approved: true
review_digest: dae46be72ca02773927087d2433228a59acc891f03e03bb622da72a76d1b39cd
review_basis: 107bbdc41f9f96b6732c872f5bee96b898c8e41d1b2ff4f803128f1ac304ad60
branch: feat/add-possibility-checkout-branch-config
base: da801025453cef7f572bf1659c3366d36bce0cfd
base_ref: main
judge_runs: 0
shipped: 2026-10-26
commits: ["e46c042","0053b15","0a8792a","faafd4d","af08a2a","185e2ed","51635fa","594cacb","7ae1b97","eb8ab2e","b808b43"]
touched: ["README.md","cmd/kirokit/","desktop/clone.go","desktop/clone_test.go","desktop/frontend/","desktop/runner_test.go","desktop/settings.go","desktop/settings_test.go","desktop/version.go","desktop/version_test.go","internal/checkout/"]
slices: 3
---

# Check out a config repo branch

## Your request

> add the possibility to checkout a branch of the config repo in the cli and desktop-app

## Problem

Someone using kirokit can only get team configs from the newest release of the config repository, or from the branch they cloned at the very start. To try configs that are not released yet (a colleague's change, a `main` ahead of the last tag), they have to go into the checkout and run `git` by hand. In the desktop app it is worse: every Update and Sync button jumps back to the newest release, so a branch they switched to by hand is lost on the next click.

## Done means

1. **The CLI switches an existing config checkout to a named branch or tag.** `kirokit sync [team] --branch <name>` fetches, checks out the branch or tag and, given a team, installs that team from it; it prints which branch or tag and commit it is on.
2. **Later CLI syncs keep the choice until the person goes back to releases.** A plain `kirokit sync` brings in new commits of a chosen branch and stays on a chosen tag; `--latest-tag` returns to the newest release.
3. **The desktop app lets the person pick a branch or tag from a list, on Settings and on the first-launch clone page.** The list shows "Newest release" and the repository's branches and tags; on the first-launch page the clone runs first (as today), then the same page shows the list with "Newest release" preselected, and picking another entry switches before setup ends. The command runs in the kirokit tab, and Teams, Agents and Skills then show the templates from that choice.
4. **Desktop Update and Sync keep following the choice.** Neither button silently returns to the newest release; choosing "Newest release" in the list is the way back.
5. **The desktop app shows whether it is on a branch, a tag or the newest release, and which one.** The person can tell at a glance where the installed templates come from.
6. **An unknown name or a checkout with local changes fails clearly and changes nothing.** The message names the problem; the checkout and the installed team stay as they were.
7. **The README describes the new flag and the desktop control.**

## Scope

### In this task

- Switching an existing config checkout between a branch, a tag and the newest release, in the CLI and in the desktop app.
- Choosing a branch or tag for the desktop app's first clone.
- Making the desktop Update and Sync buttons respect that choice.

### Not in this task

- Changing `kirokit setup --branch` (it already clones at a branch).
- The scheduled sync script `scripts/kirokit-sync-setup.sh`.
- Committing, pushing or creating branches from kirokit.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | How should the CLI switch an existing checkout to a branch? | A --branch flag on sync |
| 2 | What can be chosen: branches only, or branches and tags? | Branches and tags (pin an older release too) |
| 3 | How does the person pick the branch in the desktop app? | A list of the repository's branches |
| 4 | Where in the desktop app can the branch be chosen? | Settings and the first-launch clone page |
<!-- /ql:decided -->

## Context pointers

- Request only; no tickets or links were given.

## First look at the code

- `cmd/kirokit/commands/sync.go:14` — `sync [team]`: `updateCheckout` (`:125`) runs `git fetch --tags --prune`, then `git pull --ff-only` on the current branch, or with `--latest-tag` checks out the newest `v*` tag (detached HEAD). There is no flag to change branch. Without a team it prints `checked out <tag or branch> at <commit>` (`:84`).
- `cmd/kirokit/commands/setup.go:25` — `setup --branch <name>` clones at a branch or tag; its name validation (`:46`) is a pattern to reuse.
- `desktop/clone.go:57` — first clone runs `git clone`, then always `kirokit sync --latest-tag`.
- `desktop/frontend/src/pages/settings.js:31` — `updateArgs`: Update always passes `--latest-tag`.
- `desktop/frontend/src/pages/install.js:26` — `syncArgs`: the Sync button always passes `--latest-tag`.
- `desktop/catalog.go:231` — `repoTag` reads the release tag the clone is on (empty on a branch); the frontend keeps it in `view.catalog.tag` but no page shows it.
- `desktop/settings.go:17` — `Settings` holds `repoUrl`, `kiroHome`, `repoDir`; nothing about a branch.
- `desktop/frontend/index.html:69` — the Settings page: Kiro home, repository address, repository folder, and the Templates → Update button.
- `README.md:297` (setup), `README.md:303` (Sync), `README.md:329` (Desktop app) — the docs to update.
- Untraced areas: how the desktop app's tests drive `git` (fixtures in `desktop/clone_test.go`, `desktop/runner_test.go`), and how the frontend tests in `desktop/frontend/src/pages/pages.test.js` assert the command arguments.
