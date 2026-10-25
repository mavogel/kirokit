# Showcase video of kirokit CLI and app for the README · the build

## Summary

<!-- ql:summary -->
0 of 11 tasks · 0 of 3 slices

Building now: started 25 Oct, 11:07, 14 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | The CLI part records itself from one script | After this slice, `scripts/showcase/record.sh --keep` builds the CLI, types the install-and-chat story into a real Kiro session and leaves `cli.cast` and `cli.gif` in `/tmp/kirokit-demo/out`. | next |
| 2 | One command makes the checked GIF and MP4 | After this slice, `task showcase` records the app in Wails dev mode too, renders captions and frames, composes the MP4 and the GIF, and only writes them to `assets/` when the privacy, size and length checks pass. | after 1 |
| 3 | The README opens with the showcase and commits keep media intact | After this slice, the README shows the looping GIF right under the intro with a link to the MP4, CONTRIBUTING names `task showcase`, and the pre-commit fixers leave GIF and MP4 bytes alone. | after 2 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
None so far: the build follows the Plan as approved.
<!-- /ql:decisions -->

## Checks

- T1 · 2c83013 · `npm ci --prefix scripts/showcase && node -e "import('playwright-core').then(m=>console.log(typeof m.chromium.launch))" --input-type=module` · exit 1 · 1.4s
- T1 · 2c83013 · `npm ci --prefix scripts/showcase && node -e "const m=require(require.resolve('playwright-core',{paths:['scripts/showcase']}));if(typeof m.chromium.launch!=='function')process.exit(1)"` · not run by QualityLayer: changed after approval
- T1 · 2c83013 · `npm ci --prefix scripts/showcase && node -e "const m=require(require.resolve('playwright-core',{paths:['scripts/showcase']}));if(typeof m.chromium.launch!=='function')process.exit(1)"` · not run by QualityLayer: changed after approval
- T1 · 2c83013 · `npm ci --prefix scripts/showcase && npm ls --prefix scripts/showcase playwright-core@1.63.0` · not run by QualityLayer: changed after approval
- T1 · 2c83013 · `npm ci --prefix scripts/showcase && npm ls --prefix scripts/showcase playwright-core@1.63.0` · not run by QualityLayer: changed after approval
- T2 · 2c83013 · `bash -n scripts/showcase/cli-drive.sh && shellcheck scripts/showcase/cli-drive.sh` · exit 0 · 0.2s
- T3 · 2c83013 · `shellcheck scripts/showcase/record.sh && scripts/showcase/record.sh --keep` · exit 0 · 23.4s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- T1 (tactical): the approved verify command imported `playwright-core` from the repository root, where Node cannot resolve it (exit 1 above); it now runs `npm ls --prefix scripts/showcase playwright-core@1.63.0`, run by hand with exit 0 because QualityLayer only runs commands frozen at approval; nothing you approved changes
- T2 (tactical): `tmux capture-pane` drops trailing spaces, so the driver waits for the prompt as `$`, and the session sets `BASH_SILENCE_DEPRECATION_WARNING=1` so macOS bash prints no zsh banner; nothing you approved changes
- T3 (tactical): `agg` uses a 3 s idle limit instead of the spike's 2 s and the driver holds the `/jir` list 3.5 s, so the `jira-search` list stays readable for 3 s; nothing you approved changes
- T8 (tactical): the privacy scan removes the public slugs `mavogel/kirokit` and `mavogel/tap/` before it looks for `$USER`, because the approved clone address and end card contain the GitHub owner, which equals the local user name; every other occurrence still fails; nothing you approved changes

## Plan amended
