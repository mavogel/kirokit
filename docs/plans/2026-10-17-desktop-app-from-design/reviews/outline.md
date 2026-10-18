# Cold read of the Outline

Cold read of 01-discuss, 03-plan and 04-outline for the desktop app (about 2 minutes). Claims were checked against the code; no files were changed by the reader.

## 1. Scenarios 3 and 4 (temporary HOME) probably fail the prerequisite check at Scenario 4 step 6

The login-shell capture reads no `~/.zprofile` under a temporary HOME, so Homebrew and `~/.local/bin` are missing from PATH and SOPS and Kiro CLI show as not found. A temporary HOME may also force Go and npm to re-download their caches.

Answer: Taken · T9 now always appends the Homebrew and local tool folders to the captured PATH; scenarios 3 and 4 pass `GOPATH`, `GOMODCACHE`, `GOCACHE` and the npm cache explicitly.

## 2. T16 (Run) does not say what environment the kirokit tab's commands get

Install, sync and uninstall need `KIROKIT_REPO`, `KIRO_HOME` from Settings and the captured PATH.

Answer: Taken · T16 now requires all three on every run.

## 3. T17 (Clone) never says how `git clone` runs in the kirokit tab

`Run(args)` is kirokit-only; the header, the lock and the sign-in path were unspecified.

Answer: Taken · T16 lets a run start any program with the program's name in the header and one shared lock; T17 runs `git clone` then `kirokit sync --latest-tag` through it.

## 4. T8 in slice 2 needs slice 3

The default Kiro home mentioned the captured `KIRO_HOME`, built in T9.

Answer: Taken · T8 defaults to `~/.kiro`; T9 switches to the captured value and now lists the Settings files it touches.

## 5. Frontend tests are in no project check

`node --test src/` ran only in task verify lines, so CI would miss them.

Answer: Taken · `task desktop:test` now runs `node --test src/` (T6).

## 6. The oracle is not what Scenario 4 step 13 checks

Step 13 compared platform-team after a Sync, not the Plan's data-team replacement, and used an unnamed CLI.

Answer: Taken · new step 9a diffs the Kiro home after the typed install of data-team against a plain uninstall and install with the built CLI; step 13 keeps the sync check with an explicit path, and step 1 builds that CLI.

## 7. Plan decisions with no task or scenario

Bundle id and name, the logo, the window opening before the capture, buttons while a command runs, reinstalling the installed team, and a typed command racing a button.

Answer: Taken · bundle id and name in T6, the logo in T7, the early window in T9, disabled buttons and the reinstall case in T19. Kept · the typed-versus-button race stays unscripted: the kirokit tab is separate from the zsh tab by design (decision d8), so the two cannot mix.

## 8. Files not named by their cards

Settings files for T9 and T15, the module path in T6, and the printed path in `onboard`.

Answer: Taken · T9 lists the Settings files; T6 states the module path. Kept · T15's Settings warning for a non-zsh shell is covered by the T9 Settings files; `onboard` keeps printing the relative path it prints today.

## 9. Release script contract is thin

Flags, tap branch, URL form and the ldflags source were unspecified.

Answer: Taken · T23 names the flags, the URL and the tap target; T21 takes `VERSION` from the environment and the release job passes the tag. Kept · the unconfirmed tap-token assumption has no test; only the first real release can show it, as the Plan's Risks say.

## 10. Scenario 4 step 15 and Done means 3

Step 15 proves a new team but not a new pin.

Answer: Kept · a new pin follows the same refreshed catalog as a new team, and the catalog unit test in T10 reads the pins; a second fixture edit would add steps for no new path.

## 11. Smaller points

Stopping the dev server leaves children; resize is only dragged; Plan text still mentions option A; the team-less sync "no Kiro home" claim has no test.

Answer: Taken · scenario 2 stops the children and checks the port; scenario 3 adds `stty size`. Kept · the Plan's Option A sentence is the approved Plan and stays; the team-less sync is covered by scenario 1 step 8, which runs it with a Kiro home that holds no record for it.

## 12. Commands

No `qualitylayer` commands appear in the scenarios. Scenario 6 step 2 depends on the app built in scenario 5; `ruby -ryaml` is assumed.

Answer: Kept · the scenarios run in order and `ruby` ships with macOS, which is the only build target.
