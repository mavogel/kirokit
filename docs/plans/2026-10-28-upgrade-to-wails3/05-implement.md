# Upgrade desktop app to Wails v3 · the build

## Summary

<!-- ql:summary -->
14 of 14 tasks · 6 of 6 slices

Built: started 29 Oct, 08:01.
<!-- /ql:summary -->

## Slices

<!-- ql:slices live -->
| # | Slice | In plain words | Now |
|---|---|---|---|
| 1 | The app starts and works on Wails v3 on macOS | After this slice the desktop module builds against Wails v3.0.0 only, the page reaches Go through `@wailsio/runtime`, and `task desktop:build` produces today's universal `kirokit.app` whose smoke start quits only after `State` answered and the page acknowledged it. | done |
| 2 | The app has a native menu on every system | After this slice the macOS menu bar shows kirokit, Edit, Window and Help, and Windows and Linux show File, Edit, Window and Help in the window. | done |
| 3 | Linux on GTK 4 and Windows setup programs build with Wails v3 | After this slice `task desktop:build` produces `kirokit-desktop` (GTK 4 / WebKitGTK 6.0) on Linux and one setup program per architecture on Windows, under today's names, and the Linux packages declare the GTK 4 dependencies. | done |
| 4 | CI and the release build, smoke-start and check the v3 app on every system | After this slice the Desktop and Release workflows install GTK 4 / WebKitGTK 6.0, build Linux on Ubuntu 24.04 runners, run the installed `.deb`'s smoke start in each candidate distro, and check on Windows that the new setup program installs over the last release. | done |
| 5 | Developer commands and the showcase work on Wails v3 | After this slice `task desktop:dev` opens the app with `wails3 dev` and the freshly built dev CLI, and `task showcase` records the app clip from a loopback-only server build on port 34115 instead of v2's dev server. | done |
| 6 | The docs describe Wails v3, the menu and GTK 4, and no v2 is left | After this slice the docs name `wails3`, the menu and the GTK 4 dependencies and list the measured distros, and nothing outside the task records refers to Wails v2. | done |
<!-- /ql:slices -->

## Delivered

<!-- ql:delivered -->
Delivered 29 Oct, 08:49

**Built** · 6
- The app starts and works on Wails v3 on macOS · slice 1 · 3 tasks
- The app has a native menu on every system · slice 2 · 2 tasks
- Linux on GTK 4 and Windows setup programs build with Wails v3 · slice 3 · 2 tasks
- CI and the release build, smoke-start and check the v3 app on every system · slice 4 · 2 tasks
- Developer commands and the showcase work on Wails v3 · slice 5 · 2 tasks
- 1 more slices · 6 of 6 built

**Verified** · 3
- Project checks passed · 6 of 6 · recorded
- Security review · done · nothing severe in the change
- Second opinion · done · 1 finding

**Left** · 4
- Independent review · Review · not done
- Fix rounds · Fix · not done
- Open each menu of the built macOS app (desktop/build/bin/kirokit.app) and check: kirokit (About, Settings…, Services, Hide, Hide Others, Show All, Quit), Edit, Window, Help (kirokit documentation, Report a problem with kirokit…); ⌘, opens Settings; Help entries open the docs page and the new-issue form · needs you · T5
- Push this branch and open its pull request so the Desktop workflow runs (Ubuntu 24.04, Windows, macOS smoke starts, the linux-distros job and the Windows install-over-last-release check), then run the Release workflow by hand as a dry run · needs you · T8
<!-- /ql:delivered -->

## Decided while building

<!-- ql:decisions -->
- Slice 1 · The macOS icon is copied as `Contents/Resources/iconfile.icns` by the committed darwin Taskfile's `create:app:bundle` (and the dev bundle) instead of renamed afterwards in `scripts/desktop-build.sh`; the script still checks the file exists. No `Assets.car` is built or shipped (v2 shipped none and kirokit has no Icon Composer source); nothing you approved changes.
- Slice 1 · The committed common Taskfile drops v3's `go mod tidy`, bindings and icon-generation steps, so a release build never rewrites `go.mod` or the committed icons; `icons.icns` was generated once from `desktop/build/appicon.png` with the pinned `wails3 generate icons`; nothing you approved changes.
- Slice 1 · `desktop/frontend/src/pages/pages.test.js` also faked `window.go` for the Settings page, so it moved to the same runtime seam; nothing you approved changes.
- Slice 1 · Adding Wails v3 raised the desktop module's indirect `golang.org/x/term`, `x/text` and `x/tools` to the versions v3.0.0 requires; nothing you approved changes.
- Slice 3 · T6's smoke start runs as a regular user in a container with seccomp and /proc unconfined: WebKitGTK 6.0 always sandboxes its web process with bubblewrap, which Docker's defaults forbid (records/linux-ubuntu2404-smoke.txt). The Linux container moves to Ubuntu 24.04 because Wails v3 needs GTK 4.14 and Debian 12 has 4.8 (records/linux-debian12-gtk4.txt); debian:12 leaves the distro list as the Plan's measure-and-drop rule says (records/linux-distro-smoke.txt).
- Slice 3 · The setup program's uninstall key and folder were read from `makensis -PPO` of the committed `project.nsi` (`HKLM\…\Uninstall\mavogelkirokit`, `$PROGRAMFILES64\mavogel\kirokit`) because this container's 7-Zip cannot decompile NSIS scripts; the packed `kirokit-desktop.exe` and `bin\kirokit.exe` are byte-identical to the build (records/windows-installer.txt); nothing you approved changes.
- Slice 3 · V2's `desktop/build/windows/installer/` moved to v3's `desktop/build/windows/nsis/`; v3's `wails_tools.nsh` is committed with v2's product values (it is not regenerated by the build); nothing you approved changes.
- Slice 4 · The `linux-distros` job builds the app on each runner (ubuntu-24.04, ubuntu-24.04-arm) and installs exactly the dependencies the `.deb` declares (read from `.goreleaser.yaml` with `yq`) in each image, instead of installing the snapshot `.deb`; the snapshot needs the GoReleaser Pro key and builds amd64 only, so this way both architectures and fork pull requests are checked. The installed-`.deb` start stays in the `desktop` job. Images are `ubuntu:24.04`, `ubuntu:26.04`, `debian:13`; `debian:12` was measured out (records/linux-distro-smoke.txt); nothing you approved changes.
- Slice 4 · The Windows upgrade check counts uninstall entries whose `DisplayName`, `InstallLocation` or `UninstallString` mention kirokit (v2's installer writes no `InstallLocation`, so `UninstallString` is what names its folder); nothing you approved changes.
- Slice 5 · `task desktop:dev` failed twice before it worked, and both causes are fixed: a universal `desktop/bin/kirokit` left by `task desktop:build` cannot be overwritten by Go (`scripts/desktop-build.sh` now removes `desktop/bin/` once it has staged its output), and vite bound only `[::1]:9245` while Wails' dev proxy dials `127.0.0.1` (the dev task now passes `--host 127.0.0.1`). The dev window then showed the Teams page from the test clone (records/macos-desktop-dev.png); nothing you approved changes.
- Slice 5 · `task desktop:dev` builds the pinned `wails3` into `desktop/build/dev-bin/tools/` (already gitignored) and puts it first on `PATH`; the staged release app in `desktop/build/bin/` now uses its own bundled CLI even when started from `desktop/`, since `wails3 dev` no longer runs from there; nothing you approved changes.
- Slice 5 · `app-server.sh` also fails when the port's listener is another process (it checks the listener's pid, not only its address), so a stray server on 34115 cannot pass for the app; its server build was driven in headless Chrome to the first-launch page with no page errors; nothing you approved changes.
- Slice 5 · T10's Files add desktop/build/Taskfile.yml (vite binds 127.0.0.1, which Wails' dev proxy dials) and scripts/desktop-build.sh (removes Wails' intermediate desktop/bin after staging, since Go cannot overwrite the universal binary there when wails3 dev builds). The 'Decided while outlining' note on menu:settings now says main.js checks setup() itself: sidebar.select does not refuse while the first-launch page shows.
<!-- /ql:decisions -->

## Checks

- T1 · 1a3ddd7 · `cd desktop && go vet ./... && go test ./...` · exit 0 · 7.2s
- T2 · 1a3ddd7 · `cd desktop/frontend && npm ci && npm run build && node --test "src/**/*.test.js"` · exit 0 · 1.8s
- T3 · 1a3ddd7 · `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit && scripts/desktop-prebuilt.sh` · exit 0 · 9.9s
- T4 · 5fb955d · `cd desktop && go test -run 'Menu' ./...` · exit 0 · 0.8s
- T5 · 5fb955d · `cd desktop && go vet ./... && GOOS=windows go vet ./... && go test ./...` · exit 0 · 7.9s
- T6 · c2eacbd · `scripts/linux-docker.sh task desktop:build && docker run --rm --security-opt seccomp=unconfined --security-opt systempaths=unconfined -v "$PWD":/src:ro kirokit-linux-check bash -c 'useradd -m tester && su tester -c "cd && KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /src/desktop/build/bin/kirokit-desktop"'` · not run by QualityLayer: changed after approval
- T6 · c2eacbd · `scripts/linux-docker.sh task desktop:build && docker run --rm --security-opt seccomp=unconfined --security-opt systempaths=unconfined -v "$PWD":/src:ro kirokit-linux-check bash -c 'useradd -m tester && su tester -c "cd && KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /src/desktop/build/bin/kirokit-desktop"'` · exit 0 · 9.5s
- T7 · c2eacbd · `scripts/linux-docker.sh env GOOS=windows task desktop:build` · exit 0 · 9.5s
- T8 · bd2a0a6 · `node -e "require('node:fs'); const y=require('node:fs').readFileSync('.github/workflows/desktop.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-latest/.test(y)) process.exit(1)"` · exit 0 · 0.1s
- T9 · bd2a0a6 · `node -e "const y=require('node:fs').readFileSync('.github/workflows/release.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-26\.04-arm/.test(y)) process.exit(1)"` · exit 0 · 0.1s
- T10 · 31f16ca · `cd desktop && go test -run 'CLI|Resolve' ./...` · exit 0 · 0.9s
- T11 · 31f16ca · `node --test scripts/showcase/` · exit 0 · 77.5s
- T12 · f61b5e7 · `node scripts/check-doc-links.mjs && node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.3s
- T13 · f61b5e7 · `cd desktop/frontend && node --test src/pages/feedback.test.js` · exit 0 · 0.3s
- T14 · f61b5e7 · `! git grep -nE 'wailsapp/wails/v2|window\.go\b|window\.runtime|go tool wails |wails dev|wails build' -- ':!docs/plans/' && task ci && task desktop:test` · exit 0 · 41.6s
- all · a4dbabc · `task desktop:test` · exit 0 · 10.2s
- all · a4dbabc · `task ci` · exit 0 · 9.8s
- all · a4dbabc · `node --test scripts/showcase/` · exit 0 · 76.9s
- all · a4dbabc · `task desktop:build` · exit 0 · 9.4s
- all · a4dbabc · `node scripts/check-doc-links.mjs` · exit 0 · 0.1s
- all · a4dbabc · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s

## Known limits

<!-- ql:limits -->
- T8 · Untested here: whether stock Ubuntu 24.04's AppArmor lets WebKitGTK 6's bubblewrap sandbox run for kirokit-desktop; CI's Linux smoke start proves it, and a failure there means the .deb needs an AppArmor profile allowing userns.
<!-- /ql:limits -->

## Deviations

- T3 (tactical): the macOS icon is copied as `Contents/Resources/iconfile.icns` by the committed darwin Taskfile's `create:app:bundle` (and the dev bundle) instead of renamed afterwards in `scripts/desktop-build.sh`; the script still checks the file exists. No `Assets.car` is built or shipped (v2 shipped none and kirokit has no Icon Composer source); nothing you approved changes.
- T3 (tactical): the committed common Taskfile drops v3's `go mod tidy`, bindings and icon-generation steps, so a release build never rewrites `go.mod` or the committed icons; `icons.icns` was generated once from `desktop/build/appicon.png` with the pinned `wails3 generate icons`; nothing you approved changes.
- T2 (tactical): `desktop/frontend/src/pages/pages.test.js` also faked `window.go` for the Settings page, so it moved to the same runtime seam; nothing you approved changes.
- T1 (tactical): adding Wails v3 raised the desktop module's indirect `golang.org/x/term`, `x/text` and `x/tools` to the versions v3.0.0 requires; nothing you approved changes.
- T3 (evidence): the smoke start's sensitivity was checked with a scratch copy whose page calls `state()` instead of `start()`: it exited 1 after 60 s ("the page did not answer within 1m0s"); the real build exits 0 in about 1 s.
- T6 (amended): T6's smoke start runs as a regular user in a container with seccomp and /proc unconfined: WebKitGTK 6.0 always sandboxes its web process with bubblewrap, which Docker's defaults forbid (records/linux-ubuntu2404-smoke.txt). The Linux container moves to Ubuntu 24.04 because Wails v3 needs GTK 4.14 and Debian 12 has 4.8 (records/linux-debian12-gtk4.txt); debian:12 leaves the distro list as the Plan's measure-and-drop rule says (records/linux-distro-smoke.txt).

- T7 (tactical): the setup program's uninstall key and folder were read from `makensis -PPO` of the committed `project.nsi` (`HKLM\…\Uninstall\mavogelkirokit`, `$PROGRAMFILES64\mavogel\kirokit`) because this container's 7-Zip cannot decompile NSIS scripts; the packed `kirokit-desktop.exe` and `bin\kirokit.exe` are byte-identical to the build (records/windows-installer.txt); nothing you approved changes.
- T7 (tactical): v2's `desktop/build/windows/installer/` moved to v3's `desktop/build/windows/nsis/`; v3's `wails_tools.nsh` is committed with v2's product values (it is not regenerated by the build); nothing you approved changes.

- T8 (tactical): the `linux-distros` job builds the app on each runner (ubuntu-24.04, ubuntu-24.04-arm) and installs exactly the dependencies the `.deb` declares (read from `.goreleaser.yaml` with `yq`) in each image, instead of installing the snapshot `.deb`; the snapshot needs the GoReleaser Pro key and builds amd64 only, so this way both architectures and fork pull requests are checked. The installed-`.deb` start stays in the `desktop` job. Images are `ubuntu:24.04`, `ubuntu:26.04`, `debian:13`; `debian:12` was measured out (records/linux-distro-smoke.txt); nothing you approved changes.
- T8 (tactical): the Windows upgrade check counts uninstall entries whose `DisplayName`, `InstallLocation` or `UninstallString` mention kirokit (v2's installer writes no `InstallLocation`, so `UninstallString` is what names its folder); nothing you approved changes.

- T10 (tactical): `task desktop:dev` failed twice before it worked, and both causes are fixed: a universal `desktop/bin/kirokit` left by `task desktop:build` cannot be overwritten by Go (`scripts/desktop-build.sh` now removes `desktop/bin/` once it has staged its output), and vite bound only `[::1]:9245` while Wails' dev proxy dials `127.0.0.1` (the dev task now passes `--host 127.0.0.1`). The dev window then showed the Teams page from the test clone (records/macos-desktop-dev.png); nothing you approved changes.
- T10 (tactical): `task desktop:dev` builds the pinned `wails3` into `desktop/build/dev-bin/tools/` (already gitignored) and puts it first on `PATH`; the staged release app in `desktop/build/bin/` now uses its own bundled CLI even when started from `desktop/`, since `wails3 dev` no longer runs from there; nothing you approved changes.
- T11 (tactical): `app-server.sh` also fails when the port's listener is another process (it checks the listener's pid, not only its address), so a stray server on 34115 cannot pass for the app; its server build was driven in headless Chrome to the first-launch page with no page errors; nothing you approved changes.
- T10 (amended): T10's Files add desktop/build/Taskfile.yml (vite binds 127.0.0.1, which Wails' dev proxy dials) and scripts/desktop-build.sh (removes Wails' intermediate desktop/bin after staging, since Go cannot overwrite the universal binary there when wails3 dev builds). The 'Decided while outlining' note on menu:settings now says main.js checks setup() itself: sidebar.select does not refuse while the first-launch page shows.

## Plan amended

### T6 · amended

T6's smoke start runs as a regular user in a container with seccomp and /proc unconfined: WebKitGTK 6.0 always sandboxes its web process with bubblewrap, which Docker's defaults forbid (records/linux-ubuntu2404-smoke.txt). The Linux container moves to Ubuntu 24.04 because Wails v3 needs GTK 4.14 and Debian 12 has 4.8 (records/linux-debian12-gtk4.txt); debian:12 leaves the distro list as the Plan's measure-and-drop rule says (records/linux-distro-smoke.txt).

Why: the first Linux build in the old Debian 12 container failed to compile (`gtk_file_dialog_*`, `gtk_css_provider_load_from_string`, `gdk_monitor_get_scale`: GTK 4.10–4.14 APIs), and the Ubuntu 24.04 build then fails on Debian 12 with `undefined symbol: gdk_monitor_get_scale`. Ubuntu 24.04 (GTK 4.14.5) is therefore both the oldest system the app runs on and the release's build host, as the Plan's Assumptions expected (Ubuntu 24.04+ and Debian 13). The smoke start as root in a default container crashes with `bwrap: Creating new namespace failed` and then `Can't mount proc`, because WebKitGTK 6.0 cannot turn its sandbox off; running as a regular user with seccomp and `/proc` unconfined (what a desktop session has) passes in 1 s with the sandbox on. Disabling WebKit's sandbox was not used. The CI distro job uses the same container settings.

Amended by the agent on 2026-10-29; changed 04-outline.md.

### T10 · amended

T10's Files add desktop/build/Taskfile.yml (vite binds 127.0.0.1, which Wails' dev proxy dials) and scripts/desktop-build.sh (removes Wails' intermediate desktop/bin after staging, since Go cannot overwrite the universal binary there when wails3 dev builds). The 'Decided while outlining' note on menu:settings now says main.js checks setup() itself: sidebar.select does not refuse while the first-launch page shows.

Amended by the agent on 2026-10-29; changed 04-outline.md.
