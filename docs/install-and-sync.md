# Install, setup and sync

Put a team into a Kiro home, take it out again, download the config repository and keep it up to date. Back to the [README](../README.md).

## Install / Uninstall

```bash
kirokit install <team>
kirokit install <team> --kb-agent <agent> --kb-source <folder> --kb-mount <path>
kirokit uninstall <team>
```

- `install` renders with the real `$KIRO_HOME` baked in (`KIRO_HOME` env var, else Kiro's default `~/.kiro`, which is `%USERPROFILE%\.kiro` on Windows) and copies the rendered tree plus each agent's encrypted secrets file there. Paths in the installed configs use forward slashes on every system (`C:/Users/…/.kiro/…` on Windows). Kiro CLI, Kiro IDE and kirocrew all read that Kiro home; see [Works with](../README.md#works-with).
- A Kiro home holds one team. `install <team>` first uninstalls every other installed team and prints `uninstalled <other> from <home>` for each, then `installed <team> into <home>`. Reinstalling the installed team removes nothing. If an uninstall fails, nothing is installed.
- `uninstall` removes exactly what the matching `install` wrote, tracked in `$KIRO_HOME/.kirokit-installed/<team>.json`; other teams' files and Kiro-managed files are untouched.
- `--kb-*` flags (all or none) attach a local `knowledgeBase` resource to one installed agent: `--kb-source` local folder, `--kb-mount` symlink location (never copied), `--kb-agent` agent whose `resources` gets the entry. Entry: name `<agent>-kb`, `indexType: "best"`, `include: ["**/*.md"]`, `autoUpdate: true`. `include` works in `kiro-cli` but is not yet in Kiro's agent-config docs, and Kiro issues report agent-config knowledge bases are not always indexed automatically; check `/knowledge show`. `uninstall` removes the symlink, never the source folder. On Windows, where a symlink needs Developer Mode or an administrator, the mount is a directory junction instead, which reinstall and uninstall treat the same way.
- Nothing is installed into Amazon Quick, which has no API; its bundle is imported by hand (see [Amazon Quick](amazon-quick.md)).

### Existing files are backed up

Before `render --out <dir>`, `install` or `sync` changes something on disk, the folder is kept as `YYYYMMDD-HHmmss-<folder>.backup` (local time) beside it:

- **`render --out <dir>`** (also with `--quick`): an existing `<dir>` whose content would change is renamed to `YYYYMMDD-HHmmss-<dir name>.backup` and a fresh `<dir>` is rendered. A matching or empty `<dir>` is left alone.
- **`install` / `sync`:** each top-level folder in `$KIRO_HOME` holding a file that would change (`agents`, `prompts`, `skills`, `steering`, `hooks`, `kirokit-steerings`, `kirokit-secrets`) is first copied to `$KIRO_HOME/YYYYMMDD-HHmmss-<folder>.backup`. Unchanged folders and Kiro's own data (sessions etc.) are not copied. The backup name differs from the live folder, so Kiro does not load it. Files other tools keep in those folders, such as kirocrew's own `kirocrew*.json` agents, are copied with them.
- A no-change re-install or `sync` makes no backup.
- Same-second collisions add a counter: `20261015-101530-2-agents.backup`.
- Backups are not in the install manifest; `uninstall` neither removes nor restores them.
- Deletion is not backed up: `sync` removes files the team stopped shipping and `uninstall` removes everything the install wrote. Only manifest-listed files are ever removed.

## Setup (initial download)

```bash
brew install --cask mavogel/tap/kirokit   # or see the README's Install for Linux and Windows
kirokit setup --source https://github.com/mavogel/kirokit.git --target ~/.kirokit-config
export KIROKIT_REPO=~/.kirokit-config
```

- `setup` is the command-line twin of the desktop app's first launch: `git clone -- <source> <target>`, then a checkout of the newest `v*` tag (detached HEAD). `--source` and `--target` are required.
- `--branch <name>` checks out that branch or tag instead (`git clone --branch`), skipping the newest-tag rule: `kirokit setup --source <address> --target <folder> --branch main`. It cannot be empty, start with `-` or contain whitespace, and an unknown name fails the clone. It is the same thing `manifest.yaml` calls `ref` for external skills, minus commit hashes.
- `--source` is an `https`, `ssh` or `file` address, `git@host:path`, or an absolute path (on Windows also a drive path, `C:\…` or `C:/…`); an address starting with `-` is refused. `--target` must not exist or must be an empty folder, so an existing clone is never overwritten (use `sync` to update it).
- Without `--branch`, a repository with no `v*` tag is kept on its default branch and `setup` prints that it has no release yet (exit 0).
- It prints the `KIROKIT_REPO` line to use next; then run `install` or `sync`.

## Sync (automatic updates)

```bash
kirokit sync <team> --repo ~/.kirokit-config
kirokit sync <team> --repo ~/.kirokit-config --latest-tag
kirokit sync <team> --repo ~/.kirokit-config --branch <name>   # switch to a branch or tag, then install
kirokit sync --repo ~/.kirokit-config --latest-tag      # no team: update the checkout only
```

- `sync` runs `git fetch --tags`, then a fast-forward `git pull` of the checked-out branch (with `--latest-tag`: checkout of the newest `v*` tag, so only releases reach the machine), then the same install as `kirokit install`. When the checkout is on a tag (detached HEAD), a plain `sync` stays on it and reinstalls from it.
- `kirokit sync <team> --branch <name>` switches the checkout first: to `origin`'s branch of that name, fast-forwarded (a local branch is created to track it when there is none), or, when `origin` has no such branch, to the tag (detached HEAD). Later plain syncs keep the choice: they bring in new commits of the branch and stay on the tag, until `--latest-tag` goes back to releases. When a branch and a tag share a name, the branch wins. Branch and tag names only, no commit hashes.
- A switch changes nothing when it is refused. Untracked files do not block it (so a gitignored `quick-ids.yaml` stays). The refusals:
  - `kirokit sync: local changes in <dir>: commit or discard them, then try again` (tracked files are edited)
  - `kirokit sync: --branch <name>: local branch <name> has commits that are not on origin/<name>`
  - `kirokit sync: --branch <name>: no branch or tag of that name on origin`
  - `kirokit sync: --branch and --latest-tag cannot be used together`
  - `kirokit sync: --branch: must be a branch or tag name, not empty, starting with '-' or containing whitespace` (the rule `setup --branch` uses)
- Every sync prints `updated <before> -> <after>` when HEAD moved, then `checked out <branch or tag> at <commit>` (the commit itself as the name on a detached HEAD with no tag), with or without a team.
- Unlike `install`, it removes files the previous install wrote that the new version no longer ships, plus directories left empty.
- Like `install`, `sync <team>` uninstalls every other installed team (one `uninstalled <other> from <home>` line each), but only after the checkout updated and `<team>` rendered; a failed pull, a refused `--branch` or an unknown team leaves the installed team in place.
- Without a team, `sync` only updates the checkout and installs nothing. It needs no Kiro home, and the `--kb-*` flags are an error (`kirokit sync: --kb-agent needs a team`).
- `--repo` defaults to `$KIROKIT_REPO`, else the cwd; an explicit `--repo` wins over `KIROKIT_REPO`. `--kb-*` flags behave as for `install`; without them a previously attached knowledge base is dropped, so pass them on every run.

**Migration:** one machine now holds one team. A second scheduled sync of a different team (a second run of `scripts/kirokit-sync-setup.sh --team <other>`) replaces the first team on every run instead of installing beside it, so keep one scheduled sync per machine.

`scripts/kirokit-sync-setup.sh` clones the config repo, runs a first sync and registers a launchd agent (macOS) or systemd user timer (Linux); it does not support Windows, where you run `kirokit sync` yourself:

```bash
scripts/kirokit-sync-setup.sh --team platform-team --repo-url git@github.com:org/kiro-config.git --latest-tag
scripts/kirokit-sync-setup.sh --team platform-team --dry-run ...   # print the scheduler definition only
scripts/kirokit-sync-setup.sh --team platform-team --remove        # unregister (checkout and installed config stay)
```

Options: `--dir` (checkout, default `~/.kirokit-config`), `--interval` (hours, default 4). Needs `git` and `kirokit` on `PATH`; logs to `~/.kirokit-sync.log`; each user needs their own SOPS age key. Keep the checkout dedicated: a fast-forward pull fails on local commits and `--latest-tag` leaves a detached HEAD.
