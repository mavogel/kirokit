# Plan review by Sonnet 5.5 (sonnet · high)

The first review, by Codex, could not run (Codex stopped with an error); Sonnet 5.5 read it instead.

REVISE: 8 items

1. **Plan, "Patterns to follow → Repository roots" and "`KIROKIT_REPO` roots every repository path".** The list of users of the root constants is incomplete. `cmd/kirokit/commands/onboard.go:17,20` and `skills.go:47,58` also use `teamsRoot`, and neither is listed. The plan says "every user goes through it", but the build would miss `kirokit onboard` and `kirokit skills fetch`. Add both to the list and to the tests. The plan also leaves open what happens when `KIROKIT_REPO` is set and `sync --repo` points somewhere else, because roots would follow `KIROKIT_REPO` while git runs in `--repo`. Either make `--repo` override the roots or reject the mix.

   Answer: Taken · Repository roots now list `onboard.go:17-20` and `skills.go:47,58`, each with a `KIROKIT_REPO` test; an explicit `sync --repo` wins and roots every path at it.

2. **Plan, "How we will know it works" (the oracle).** The byte-for-byte comparison cannot pass as written. Install bakes the real Kiro home into the rendered files: all four agent templates use `.KiroHome`, and the install record stores absolute paths. Installing into a temporary app Kiro home and into "a second empty home" produces different bytes. Normalise the home path before comparing, or install both times into the same path in sequence. As written, the oracle fails for the wrong reason.

   Answer: Taken · The oracle now compares the app's install with a plain `kirokit install data-team` into the same Kiro home after `kirokit uninstall data-team`, so the baked-in home path is identical.

3. **Plan, "Engineering decisions" (first-launch clone, sync) vs `sync.go:80-98`.** First launch checks out the newest tag, which leaves a detached HEAD. The Sync button in the mockup runs plain `kirokit sync data-team`, which does `git pull --ff-only` and fails on a detached HEAD. The Sync button's command is never specified. Decide whether the app's Sync always passes `--latest-tag`, and whether `sync` should detect a detached HEAD. Separately, `--latest-tag` errors when the repository has no `v*` tag. Name that failure in the first-launch error path, since the Assumptions list only calls it "unconfirmed".

   Answer: Taken · The app's Sync and Update always pass `--latest-tag` (its clone sits on a tag). A repository without a `v*` tag keeps the clone, shows a named message on first launch and works from the default branch.

4. **Plan, "The app bundles its own kirokit CLI, first on its terminals' PATH".** Done #9 depends on this, and the plan does not prove it. The zsh tab runs `-il`, so `.zprofile` and `.zshrc` run after the app sets PATH. Typical Homebrew setup (`brew shellenv`) can put `/opt/homebrew/bin` ahead of the bundled folder. A typed `kirokit` would then run an older CLI that lacks `KIROKIT_REPO` and the one-team rule. The spike only covered `env -i … zsh -ilc env` and a pty shell, not command resolution. Add a check that `command -v kirokit` in the shell tab resolves to the bundled copy. Re-prepend after startup if it does not, for example with a `precmd` hook or a `ZDOTDIR` wrapper.

   Answer: Taken · The zsh tab now starts through a `ZDOTDIR` wrapper that sources the person's own files and then puts the bundled CLI first and re-exports the app's settings; proven in `records/spike-zdotdir-wrapper.txt`, with a `command -v kirokit` test.

5. **Plan, "Settings" and "`KIROKIT_REPO` roots every repository path".** The plan says the app sets `KIROKIT_REPO` in its terminals but never says it also exports `KIRO_HOME` from Settings. If Settings holds a Kiro home other than `~/.kiro`, a typed `kirokit install` in the zsh tab writes elsewhere. The records the app polls do not change, so Done #5, #6 and #9 fail silently. State that every terminal and every `Run` gets `KIRO_HOME=<Settings.KiroHome>`. Also state what happens when the login-shell environment already exports a different `KIRO_HOME`.

   Answer: Taken · The contracts state that both terminals get `KIRO_HOME` from Settings, that Settings defaults to the captured `KIRO_HOME` else `~/.kiro`, and that once saved it wins over the shell's own value (re-exported by the wrapper).

6. **Plan, "A macOS release job adds the app and its cask" and "Risks".** The release workflow has no job-level outputs, only step outputs (`release.yaml:35-47`), so the new job needs `needs: release` and an `outputs:` block. The tag is also needed. The `workflow_dispatch` dry-run needs a new `on:` trigger. That trigger would also run the existing `release` job on dispatch, which must be gated. The plan lists none of this. The sentence "The release job runs only in CI, on a tagged release" is unclear. A macOS job that does a universal Wails build on every PR is also slow and costly. Consider running it only on pushes to `main`, or on changes under `desktop/` or `cmd/`.

   Answer: Taken · The release contract now shows the `outputs:` block, `needs: release`, a `workflow_dispatch` dry run with the existing job gated to pushes, and a separate `desktop.yaml` CI workflow limited to pushes to main and pull requests touching Go or desktop paths.

7. **Plan, "The one-team rule…" and "Risky".** The plan says this changes `install` and `sync` for every CLI user, but not who breaks. `scripts/kirokit-sync-setup.sh` schedules `sync` per team. Anyone syncing two teams from cron would have each run uninstall the other. Name that in Risky and the README migration note, and consider a CHANGELOG entry. The `--latest-tag` flag text and the `sync` Long help also need updating for the team-less form.

   Answer: Taken · The Risky line and decision 2's Look-here name the two-scheduled-syncs case; the README gets a migration note, the commit carries `!`, and `sync`'s `Use`, `Long` and `--latest-tag` help are updated. The project keeps no CHANGELOG file; release notes come from commits.

8. **Plan, "Trust boundary" and "Engineering decisions".** The line omits that the repository address is arbitrary and that its templates are installed into `~/.kiro` as agent configs, which can launch commands. It also omits the `KiroHome` setting, a person-entered path under which uninstall deletes files. For `git clone`, "one argument, never through a shell" is not enough. Require a `--` separator (an address starting with `-` would otherwise be read as an option) and restrict the allowed schemes or forms.

   Answer: Taken · The trust boundary now covers the address (`git clone -- <address>`, allowed forms, no leading `-`, templates become agent configs) and the Kiro home setting (absolute, not `/` or the home folder; uninstall still removes only recorded paths).

Smaller points:
- `npm ci` and `npm run build` are named, but the project's toolchain note says pnpm. Pick one and commit its lockfile.

  Answer: Kept · the repository has no JavaScript toolchain or pnpm note; npm with a committed `package-lock.json` is Wails' default, now stated under Build order.
- `go vet` and `go test` in `desktop/` fail locally without `frontend/dist`, because of the Wails embed. Add a placeholder file or document the build order.

  Answer: Taken · Build order is documented: `frontend/dist` is git-ignored and `task desktop:test` builds the frontend before vet and test.
- Calling `Catalog()` renders every team into a temporary folder each time. Say when it runs, for example only on clone, update and start.

  Answer: Taken · `Catalog()` renders at start, after a clone and after each sync, and returns the cached result otherwise.
- Option A of d8 is described only as replacing `Run`. Say whether the kirokit tab and the `term:kirokit` events are removed in that case.

  Answer: Taken · Option A drops the kirokit tab, `Run`, `RunWrite` and the `term:kirokit` events.

Not checked: the spike records in `records/*.txt` and the research records `research-r1` to `r3`, and `desktop-app-build.html` beyond what the screenshot shows. I also did not check the claims about Wails v2.14.0, creack/pty or the Finder PATH against external sources. I viewed only the light 1440 screenshot of the main window. I did not see the First launch screen, the A and B toggle states or the dark theme.
