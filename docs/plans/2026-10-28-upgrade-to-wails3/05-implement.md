# Upgrade desktop app to Wails v3 · the build

## Summary

<!-- ql:summary -->
5 of 14 tasks · 2 of 6 slices

Building now: started 29 Oct, 08:01, 16 min so far.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | The app starts and works on Wails v3 on macOS | After this slice the desktop module builds against Wails v3.0.0 only, the page reaches Go through `@wailsio/runtime`, and `task desktop:build` produces today's universal `kirokit.app` whose smoke start quits only after `State` answered and the page acknowledged it. | done |
| 2 | The app has a native menu on every system | After this slice the macOS menu bar shows kirokit, Edit, Window and Help, and Windows and Linux show File, Edit, Window and Help in the window. | done |
| 3 | Linux on GTK 4 and Windows setup programs build with Wails v3 | After this slice `task desktop:build` produces `kirokit-desktop` (GTK 4 / WebKitGTK 6.0) on Linux and one setup program per architecture on Windows, under today's names, and the Linux packages declare the GTK 4 dependencies. | next |
| 4 | CI and the release build, smoke-start and check the v3 app on every system | After this slice the Desktop and Release workflows install GTK 4 / WebKitGTK 6.0, build Linux on Ubuntu 24.04 runners, run the installed `.deb`'s smoke start in each candidate distro, and check on Windows that the new setup program installs over the last release. | after 3 |
| 5 | Developer commands and the showcase work on Wails v3 | After this slice `task desktop:dev` opens the app with `wails3 dev` and the freshly built dev CLI, and `task showcase` records the app clip from a loopback-only server build on port 34115 instead of v2's dev server. | next |
| 6 | The docs describe Wails v3, the menu and GTK 4, and no v2 is left | After this slice the docs name `wails3`, the menu and the GTK 4 dependencies and list the measured distros, and nothing outside the task records refers to Wails v2. | after 4, 5 |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Written when the last task is ticked.
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- Slice 1 · The macOS icon is copied as `Contents/Resources/iconfile.icns` by the committed darwin Taskfile's `create:app:bundle` (and the dev bundle) instead of renamed afterwards in `scripts/desktop-build.sh`; the script still checks the file exists. No `Assets.car` is built or shipped (v2 shipped none and kirokit has no Icon Composer source); nothing you approved changes.
- Slice 1 · The committed common Taskfile drops v3's `go mod tidy`, bindings and icon-generation steps, so a release build never rewrites `go.mod` or the committed icons; `icons.icns` was generated once from `desktop/build/appicon.png` with the pinned `wails3 generate icons`; nothing you approved changes.
- Slice 1 · `desktop/frontend/src/pages/pages.test.js` also faked `window.go` for the Settings page, so it moved to the same runtime seam; nothing you approved changes.
- Slice 1 · Adding Wails v3 raised the desktop module's indirect `golang.org/x/term`, `x/text` and `x/tools` to the versions v3.0.0 requires; nothing you approved changes.
<!-- /ql:decisions -->

## Checks

- T1 · 1a3ddd7 · `cd desktop && go vet ./... && go test ./...` · exit 0 · 7.2s
- T2 · 1a3ddd7 · `cd desktop/frontend && npm ci && npm run build && node --test "src/**/*.test.js"` · exit 0 · 1.8s
- T3 · 1a3ddd7 · `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit && scripts/desktop-prebuilt.sh` · exit 0 · 9.9s
- T4 · 5fb955d · `cd desktop && go test -run 'Menu' ./...` · exit 0 · 0.8s
- T5 · 5fb955d · `cd desktop && go vet ./... && GOOS=windows go vet ./... && go test ./...` · exit 0 · 7.9s

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## Deviations

- T3 (tactical): the macOS icon is copied as `Contents/Resources/iconfile.icns` by the committed darwin Taskfile's `create:app:bundle` (and the dev bundle) instead of renamed afterwards in `scripts/desktop-build.sh`; the script still checks the file exists. No `Assets.car` is built or shipped (v2 shipped none and kirokit has no Icon Composer source); nothing you approved changes.
- T3 (tactical): the committed common Taskfile drops v3's `go mod tidy`, bindings and icon-generation steps, so a release build never rewrites `go.mod` or the committed icons; `icons.icns` was generated once from `desktop/build/appicon.png` with the pinned `wails3 generate icons`; nothing you approved changes.
- T2 (tactical): `desktop/frontend/src/pages/pages.test.js` also faked `window.go` for the Settings page, so it moved to the same runtime seam; nothing you approved changes.
- T1 (tactical): adding Wails v3 raised the desktop module's indirect `golang.org/x/term`, `x/text` and `x/tools` to the versions v3.0.0 requires; nothing you approved changes.
- T3 (evidence): the smoke start's sensitivity was checked with a scratch copy whose page calls `state()` instead of `start()`: it exited 1 after 60 s ("the page did not answer within 1m0s"); the real build exits 0 in about 1 s.

## Plan amended
