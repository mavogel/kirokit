# Upgrade desktop app to Wails v3

<!-- ql:links -->
Task: `docs/plans/2026-10-28-upgrade-to-wails3/`
<!-- /ql:links -->

## Risk at merge

<!-- ql:risk -->
**Higher than planned.**

- 4 dependency files changed: `desktop/frontend/package-lock.json`, `desktop/frontend/package.json`, `desktop/go.mod` and 1 more
- 2 amendments to the approved Plan
- Points 2, 5, 6, 7, 8 not verified
<!-- /ql:risk -->

## Why the change

The desktop app moves from Wails v2.17, which now gets only critical fixes, to Wails v3.0.0, keeping every screen, release file name and install path, and gains a native application menu.

## What a reviewer must know

- **Higher than planned, look first at Linux support:** Wails v3 needs GTK 4.14, so the Linux app now supports Ubuntu 24.04+ and Debian 13+ only. Ubuntu 22.04 and Debian 12 drop out (measured: Debian 12 fails with `undefined symbol: gdk_monitor_get_scale`, `records/linux-debian12-gtk4.txt`). The commit is marked `feat(desktop)!`.
- **Nothing ran in CI yet:** the branch is not pushed. CI's smoke starts on all three systems, the new `linux-distros` job, the Windows install-over-v0.8.1 check and a Release dry run are the remaining proof. One open risk: stock Ubuntu 24.04 restricts the bubblewrap sandbox WebKitGTK 6 always uses; if CI's Linux smoke start fails with a bwrap permission error, the `.deb` needs an AppArmor profile.
- **Follow-ups from the reviewer:** Renovate does not group `github.com/wailsapp/wails/v3` with `@wailsio/runtime`, so either can be bumped alone; the app inside each DMG fails `codesign --verify --strict`, as the v0.8.1 DMG already does.

## What changed

The Go side and the page each keep one place that talks to Wails; only those and the build adapter changed shape:

```diff
 Go (desktop/)
-  wails.Run(options.App{Bind: app, OnStartup, OnShutdown, OnDomReady: smoke})
+  application.New(Services: NewService(app)) + one window + Menu.Set(buildMenu(menuFor(goos)))
-  App.startup(ctx) / App.shutdown(ctx), runtime.EventsEmit(ctx, …)
+  App.ServiceStartup / ServiceShutdown, setEmitter(wapp.Event.Emit)
+  menu.go: menuFor(goos) → data; App.menuAction("settings"|"docs"|"report")
 Page (desktop/frontend/src/api.js)
-  window.go.main.App.<Method>(…), window.runtime.EventsOn / BrowserOpenURL
+  @wailsio/runtime: Call.ByName("main.App.<Method>", …), Events.On(e => h(e.data)), Browser.OpenURL
+  start(): State, then Events.Emit("smoke:ready")   (first page load)
```

The smoke start now proves the bridge works in both directions, not just that the page loaded:

```text
KIROKIT_DESKTOP_SMOKE=1:  page → Call main.App.State → Go answers → page emits smoke:ready → Go quits 0
                          no smoke:ready within 60 s → exit 1   (checked with a page that skips it: exit 1 after 60 s)
```

The build adapter keeps every consumer's paths; Wails v3 writes to `desktop/bin/`, the script stages today's names:

```text
go tool-pinned wails3 (on PATH) → wails3 task darwin:package:universal | linux:build | windows:package ARCH=…
  → desktop/bin/*  → scripts/desktop-build.sh stages →  desktop/build/bin/kirokit.app
                                                          desktop/build/bin/kirokit-desktop (+ kirokit)
                                                          desktop/build/bin/kirokit-desktop-<arch>.exe
                                                          desktop/build/bin/installers/kirokit-desktop_<v>_windows_<arch>-setup.exe
  → desktop-prebuilt.sh, .goreleaser.yaml, desktop.yaml, release.yaml (unchanged paths)
```

New project files from Wails v3, committed and hand-edited (do not run `wails3 update build-assets`):

```text
desktop/
├── Taskfile.yml                 v3 tasks, APP_NAME kirokit
└── build/
    ├── config.yml               product info (mavogel / kirokit / 0.0.0), dev mode
    ├── Taskfile.yml             frontend build, no bindings step
    ├── darwin/                  Taskfile (macOS 13.0), Info.plist, icons.icns
    ├── linux/Taskfile.yml       GTK 4 build only (GoReleaser packages it)
    └── windows/                 Taskfile, nsis/project.nsi + wails_tools.nsh (v2's identity: mavogelkirokit)
```

Linux, CI and the showcase:

- Linux build hosts move to Ubuntu 24.04 (`ubuntu-24.04`, `ubuntu-24.04-arm`, and the local container), with `libgtk-4-dev libwebkitgtk-6.0-dev`; the `.deb` depends on `libgtk-4-1, libwebkitgtk-6.0-4`, the `.rpm` on `gtk4, webkitgtk6.0`.
- `.github/workflows/desktop.yaml` adds `linux-distros` (amd64 and arm64 × `ubuntu:24.04`, `ubuntu:26.04`, `debian:13`) and a Windows step that installs v0.8.1, then the new setup program, and checks the files and a single `mavogelkirokit` uninstall entry.
- `task showcase` drives a server build (`scripts/showcase/app-server.sh`) bound to 127.0.0.1:34115 instead of v2's `wails dev`.

## Screens compared

<!-- ql:screens -->
- **app-menu** · 1280 px · light · not captured · current
  - Independent review: No comparison recorded.
<!-- /ql:screens -->

## Built, slice by slice

<!-- ql:slices -->
6 slices · 14 tasks · 10 commits · 51 files · +2136 −513

**Slice 1 · The app starts and works on Wails v3 on macOS** · T1, T2, T3 · 23 files · +987 −359

- `desktop/.gitignore` · T3, T7 · +9 −7
- `desktop/Taskfile.yml` · T3, T6 · +40 −0
- `desktop/app.go` · T1 · +16 −10
- `desktop/build/Taskfile.yml` · T3, T10 · +45 −1
- `desktop/build/config.yml` · T3 · new · +44 −0
- `desktop/build/darwin/Info.dev.plist` · T3 · +6 −6
- `desktop/build/darwin/Info.plist` · T3 · +6 −6
- `desktop/build/darwin/Taskfile.yml` · T3 · new · +84 −0
- `desktop/build/darwin/icons.icns` · T3 · new · +0 −0
- `desktop/frontend/package-lock.json` · T2 · +7 −0
- `desktop/frontend/package.json` · T2 · +1 −0
- `desktop/frontend/src/api.js` · T2 · +55 −31
- `desktop/frontend/src/api.test.js` · T2 · +164 −16
- `desktop/frontend/src/main.js` · T2, T5 · +10 −3
- `desktop/frontend/src/pages/pages.test.js` · T2 · +5 −5
- `desktop/go.mod` · T1 · +106 −61
- `desktop/go.sum` · T1 · +261 −132
- `desktop/main.go` · T1, T5 · +29 −22
- `desktop/shell_test.go` · T1, T10 · +42 −10
- `desktop/smoke.go` · T1 · +7 −7
- `desktop/smoke_test.go` · T1 · +6 −7
- `desktop/wails.json` · T3 · deleted · +0 −20
- `scripts/desktop-build.sh` · T3, T6, T7, T10 · +44 −15

**Slice 2 · The app has a native menu on every system** · T4, T5 · 3 files · +337 −0

- `desktop/menu.go` · T4 · new · +79 −0
- `desktop/menu_test.go` · T4 · +202 −0
- `desktop/menu_wails.go` · T5 · new · +56 −0

**Slice 3 · Linux on GTK 4 and Windows setup programs build with Wails v3** · T6, T7 · 10 files · +459 −69

- `.goreleaser.yaml` · T6 · +15 −6
- `desktop/build/linux/Taskfile.yml` · T6 · new · +33 −0
- `desktop/build/linux/apparmor/kirokit-desktop` · T6 · new · +20 −0
- `desktop/build/linux/apparmor/postinstall.sh` · T6 · new · +13 −0
- `desktop/build/windows/Taskfile.yml` · T7 · new · +64 −0
- `desktop/build/windows/info.json` · T7 · +7 −7
- `desktop/build/windows/nsis/project.nsi` (renamed from `desktop/build/windows/installer/project.nsi`) · T7 · +14 −44
- `desktop/build/windows/nsis/wails_tools.nsh` · T7 · +272 −5
- `desktop/build/windows/wails.exe.manifest` · T7 · +8 −1
- `scripts/linux-docker.Dockerfile` · T6 · +13 −6

**Slice 4 · CI and the release build, smoke-start and check the v3 app on every system** · T8, T9 · 2 files · +152 −17

- `.github/workflows/desktop.yaml` · T8 · +145 −12
- `.github/workflows/release.yaml` · T9 · +7 −5

**Slice 5 · Developer commands and the showcase work on Wails v3** · T10, T11 · 6 files · +176 −46

- `Taskfile.yml` · T10 · +7 −3
- `desktop/shell.go` · T10 · +4 −5
- `scripts/showcase/app-drive.mjs` · T11 · +3 −2
- `scripts/showcase/app-server.sh` · T11 · new · +69 −0
- `scripts/showcase/record.sh` · T11 · +12 −32
- `scripts/showcase/showcase.test.mjs` · T11 · +81 −4

**Slice 6 · The docs describe Wails v3, the menu and GTK 4, and no v2 is left** · T12, T13, T14 · 7 files · +25 −22

- `AGENTS.md` · T12 · +1 −1
- `CONTRIBUTING.md` · T12 · +6 −6
- `README.md` · T12 · +1 −1
- `desktop/frontend/src/pages/feedback.js` · T13 · +3 −3
- `desktop/frontend/src/pages/feedback.test.js` · T13 · +4 −3
- `docs/desktop-app.md` · T12 · +6 −4
- `renovate.json` · T13 · +4 −4
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. The app runs on Wails v3 with no Wails v2 left.** Scenario 1: The built macOS app starts on Wails v3 and the bridge answers both ways, after slice 1 · passed · `evidence/it1-truth-1.txt`
- **2. Every screen and action works as it does today.** Only you can confirm: it feels the same to use · not confirmed
- **3. Both terminal tabs stream live and take input as before.** Scenario 2: Every page, both terminals and live updates work in the v3 app, after slice 1 · passed · `evidence/it1-scenario-2.txt`
- **4. Live updates still arrive without a reload.** Scenario 2: Every page, both terminals and live updates work in the v3 app, after slice 1 · passed · `evidence/it1-native-events-before.png`
- **5. Release builds keep their formats and names on all three systems.** Scenario 1: The built macOS app starts on Wails v3 and the bridge answers both ways, after slice 1 · Scenario 4: Linux and Windows packages build under today's names and install in place, after slice 4 · no proof found · `evidence/it1-macos-dmg.txt`
- **6. Existing installs upgrade in place.** Scenario 4: Linux and Windows packages build under today's names and install in place, after slice 4 · no proof found · `evidence/it1-release-config.txt`
- **7. The developer commands and the showcase still work.** Scenario 5: The developer commands and the showcase work, after slice 6 · no proof found · `evidence/it1-desktop-dev.txt`
- **8. The app has a native menu with the usual entries.** Only you can confirm: the menu reads right on each system · not confirmed
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- **No generated typed bindings.** You chose like-for-like plus the menu; the page keeps calling Go by name through one module.
- **No tray icon, extra windows or separate Settings window.** Not picked.
- **No GTK 3 fallback build.** You chose GTK 4 now; a second Linux variant is not built.
- **No signing or notarisation changes.** Releases stay unsigned as today.
- **No v3 dev-mode browser workflow.** `task desktop:dev` opens the app's own window; a plain browser on the dev server has no bindings in v3.
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
- T8 · Untested here: whether stock Ubuntu 24.04's AppArmor lets WebKitGTK 6's bubblewrap sandbox run for kirokit-desktop; CI's Linux smoke start proves it, and a failure there means the .deb needs an AppArmor profile allowing userns.
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
12 items remain unverified. Lanes that ran: Security review, Project checks, Code review and Second opinion.

- T5: Open each menu of the built macOS app (desktop/build/bin/kirokit.app) and check: kirokit (About, Settings…, Services, Hide, Hide Others, Show All, Quit), Edit, Window, Help (kirokit documentation, Report a problem with kirokit…); ⌘, opens Settings; Help entries open the docs page and the new-issue form. This session has no macOS accessibility permission, so it can read the menu bar only from a screenshot, not open the menus or press shortcuts.
- T8: Push this branch and open its pull request so the Desktop workflow runs (Ubuntu 24.04, Windows, macOS smoke starts, the linux-distros job and the Windows install-over-last-release check), then run the Release workflow by hand as a dry run. This session may not push.
- [MISSING] done 5: environment: CI's smoke starts and the Linux packages were not run here. The branch is not pushed (main is 8 commits ahead), so the Desktop and Release workflows never ran on it, and no `.deb`/`.rpm` was packaged. Shown locally: the universal `kirokit.app` (minimum 13.0, CLI inside) in both `kirokit-desktop_<v>_darwin_<arch>.dmg` from a GoReleaser snapshot, with a smoke start exit 0 from the DMG (`evidence/it1-macos-dmg.txt`); `kirokit-desktop` and `kirokit` for Linux with smoke exit 0 (`evidence/it1-scenario-4-linux.txt`); one `kirokit-desktop_<v>_windows_<arch>-setup.exe` per architecture (`evidence/it1-scenario-4-windows.txt`). Push the branch and let the Desktop workflow and a Release dry run finish.
- [MISSING] done 6: environment: installing over the last Windows release runs only in CI, which has not run. The new installer's identity matches v2's in the source (key `mavogelkirokit` under HKLM, 64-bit view, folder `$PROGRAMFILES64\mavogel\kirokit`, `kirokit-desktop.exe`), and `v0.8.1` has a Windows amd64 setup asset for the check to use (`evidence/it1-release-config.txt`). On macOS the data folder is unchanged and the new app opened straight on a saved clone and settings, but no install over the v0.8.1 app was done.
- [MISSING] done 7: environment: `task showcase` was not run, because it needs your signed-in Kiro IDE and kiro-cli, kirocrew and your SOPS age key. The rest holds: `task desktop:test`, `task desktop:build` and `task ci` receipts pass, and `task desktop:dev` built the pinned `wails3`, started vite and opened the window from `desktop/bin/kirokit.dev.app` with the fresh dev-bin CLI (`evidence/it1-desktop-dev.txt`, `evidence/it1-desktop-dev.png`). The changed showcase step, `app-server.sh`, served the app on 127.0.0.1:34115 and Chrome drove it end to end (`evidence/it1-scenario-2.txt`). The docs name `wails3`, the menu, GTK 4 and the measured distros.
- [MISSING] scenario 3: environment: menus could not be opened and ⌘, could not be sent here (osascript: "not allowed assistive access", `AXIsProcessTrusted` false). Step 1 passed (`go test -run Menu`), step 2 shows the menu bar only, step 6 shows Linux's File, Edit, Window, Help bar; steps 3 to 5 need someone at the Mac — `evidence/it1-scenario-3.txt`
- [MISSING] scenario 4: environment: the branch is not pushed, so the Desktop workflow, a Release dry run and an install over a real release could not run. Steps 1 and 2 pass: the Linux build links GTK 4 / WebKitGTK 6.0 and its smoke start as a regular user exits 0; both Windows setup programs are built and hold the app and CLI byte for byte — `evidence/it1-scenario-4-linux.txt`, `evidence/it1-scenario-4-windows.txt`
- [MISSING] scenario 5: environment: `task showcase` needs your Kiro sign-in, Kiro IDE, kirocrew and SOPS key and was not run. Steps 1, 2, 3 and 6 are receipts; step 4, `task desktop:dev`, opened the window — `evidence/it1-desktop-dev.txt`
- [MISSING] dod T5: environment: Settings…, ⌘, , Edit's copy/paste and the Help entries could not be driven in the built app here (no accessibility access). The menu bar shows kirokit, Edit, Window, Help and `GOOS=windows go vet ./...` passes — `evidence/it1-scenario-3.txt`
- [MISSING] truth oracle: environment: the release workflow and CI's smoke starts have not run on this branch, which is not pushed. Local stand-ins pass: the macOS app and both DMGs with smoke exit 0, the Linux build's smoke exit 0 in Ubuntu 24.04, both Windows setup programs under today's names. Windows' own smoke start and the CI matrix remain — `evidence/it1-macos-dmg.txt`, `evidence/it1-scenario-4-linux.txt`, `evidence/it1-scenario-4-windows.txt`
- Every screen and action works as it does today: Every action in scenario 2 worked as before (clone, version step, Teams, both terminals, tag switch, Export for Quick, Open folder, the report link); only you can judge the feel, so use the built app for a while.
- The app has a native menu with the usual entries: On macOS the menu bar reads kirokit, Edit, Window, Help (`evidence/it1-menu-bar.png`); on Linux the window shows File, Edit, Window, Help (`evidence/it1-scenario-3-linux-menu.png`). The menus could not be opened here, so open each one on macOS, Windows and Linux, press ⌘, / Ctrl+, and try both Help entries.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The built macOS app starts and its menu works
- Do: build the app, open it, then open each menu, press ⌘, and try both Help entries.
- Run: `task desktop:build && open desktop/build/bin/kirokit.app`
- Expect: menu bar kirokit · Edit · Window · Help; ⌘, opens Settings; Help opens `docs/desktop-app.md` and a new kirokit issue on GitHub.
- Saw: the build passes and the smoke start exits 0 in about 1 s; the menu bar reads kirokit, Edit, Window, Help. The menus themselves could not be opened here (no accessibility permission) — `evidence/it1-menu-bar.png`, `records/macos-menu-bar.png`

#### 2. The app works as before
- Do: use the built app for a while: Teams, Install, both terminal tabs (type and resize), Settings and a version switch.
- Expect: everything behaves as in v0.8.1, and live updates (install records, clone and switch results) arrive without a reload.
- Saw: the reviewer drove clone, version step, Teams, both terminals, a tag switch, Export for Quick, Open folder and the report link; an install made outside the app updated the native window on its own — `evidence/it1-scenario-2.txt`, `evidence/it1-native-events-after.png`

#### 3. CI proves the release on all three systems
- Do: push the branch, open a pull request, then run the Release workflow by hand as a dry run.
- Expect: the Desktop workflow passes on Ubuntu 24.04, Windows and macOS, including `linux-distros` and "Install over the last release (Windows)"; the dry run prints today's file names.
- Saw: not run — this session may not push. Locally: the Linux app's smoke start passes on Ubuntu 24.04/26.04 and Debian 13 (arm64), both Windows setup programs build with v2's install identity — `records/linux-distro-smoke.txt`, `records/windows-installer.txt`

### Not verified

- Everything in CI: the three smoke starts, `linux-distros` on amd64, the Windows install over v0.8.1, `.deb`/`.rpm` packaging and the Release dry run (branch not pushed).
- Whether stock Ubuntu 24.04's AppArmor lets WebKitGTK 6's sandbox run; CI's Linux smoke start shows it.
- Opening the native menus, ⌘, / Ctrl+, and the Help entries (no macOS accessibility permission here); the menu on Windows.
- `task showcase` end to end (needs your Kiro sign-in, Kiro IDE, kirocrew and SOPS key); its new launch step was tested on its own.
- The menu against the `app-menu` artifact: no screen comparison was recorded.
- The last two commits after the judged one (`a4dbabc`): the reviewer's added tests (passing) and the Codex fix to the Windows upgrade step (actionlint clean, its `gh` queries run against mavogel/kirokit); no app code changed after the judged commit.

### Docs

- `docs/desktop-app.md` (Wails v3, the menu, GTK 4 dependencies and supported Linux systems), `README.md` (Linux systems), `CONTRIBUTING.md` (Linux packages, the Ubuntu 24.04 container and its sandbox note, CI checks, showcase ports), `AGENTS.md` (desktop module layout).

## Security

- **Low:** while `task showcase` records, the server build on 127.0.0.1:34115 answers every bound method without authentication, including `ShellWrite`, so another local process on the same machine could drive the shell tab as the recording user (`scripts/showcase/app-server.sh:36`). It only listens on loopback (enforced and tested), only exists in the recorder's temporary folder, and matches what v2's `wails dev` exposed on the same port; browsers are kept out by Wails v3's Host and Origin checks.

## Accepted with open items

<!-- ql:accepted -->
These items remain unverified for Review:

- T5: Open each menu of the built macOS app (desktop/build/bin/kirokit.app) and check: kirokit (About, Settings…, Services, Hide, Hide Others, Show All, Quit), Edit, Window, Help (kirokit documentation, Report a problem with kirokit…); ⌘, opens Settings; Help entries open the docs page and the new-issue form. This session has no macOS accessibility permission, so it can read the menu bar only from a screenshot, not open the menus or press shortcuts.
- T8: Push this branch and open its pull request so the Desktop workflow runs (Ubuntu 24.04, Windows, macOS smoke starts, the linux-distros job and the Windows install-over-last-release check), then run the Release workflow by hand as a dry run. This session may not push.
- [MISSING] done 5: environment: CI's smoke starts and the Linux packages were not run here. The branch is not pushed (main is 8 commits ahead), so the Desktop and Release workflows never ran on it, and no `.deb`/`.rpm` was packaged. Shown locally: the universal `kirokit.app` (minimum 13.0, CLI inside) in both `kirokit-desktop_<v>_darwin_<arch>.dmg` from a GoReleaser snapshot, with a smoke start exit 0 from the DMG (`evidence/it1-macos-dmg.txt`); `kirokit-desktop` and `kirokit` for Linux with smoke exit 0 (`evidence/it1-scenario-4-linux.txt`); one `kirokit-desktop_<v>_windows_<arch>-setup.exe` per architecture (`evidence/it1-scenario-4-windows.txt`). Push the branch and let the Desktop workflow and a Release dry run finish.
- [MISSING] done 6: environment: installing over the last Windows release runs only in CI, which has not run. The new installer's identity matches v2's in the source (key `mavogelkirokit` under HKLM, 64-bit view, folder `$PROGRAMFILES64\mavogel\kirokit`, `kirokit-desktop.exe`), and `v0.8.1` has a Windows amd64 setup asset for the check to use (`evidence/it1-release-config.txt`). On macOS the data folder is unchanged and the new app opened straight on a saved clone and settings, but no install over the v0.8.1 app was done.
- [MISSING] done 7: environment: `task showcase` was not run, because it needs your signed-in Kiro IDE and kiro-cli, kirocrew and your SOPS age key. The rest holds: `task desktop:test`, `task desktop:build` and `task ci` receipts pass, and `task desktop:dev` built the pinned `wails3`, started vite and opened the window from `desktop/bin/kirokit.dev.app` with the fresh dev-bin CLI (`evidence/it1-desktop-dev.txt`, `evidence/it1-desktop-dev.png`). The changed showcase step, `app-server.sh`, served the app on 127.0.0.1:34115 and Chrome drove it end to end (`evidence/it1-scenario-2.txt`). The docs name `wails3`, the menu, GTK 4 and the measured distros.
- [MISSING] scenario 3: environment: menus could not be opened and ⌘, could not be sent here (osascript: "not allowed assistive access", `AXIsProcessTrusted` false). Step 1 passed (`go test -run Menu`), step 2 shows the menu bar only, step 6 shows Linux's File, Edit, Window, Help bar; steps 3 to 5 need someone at the Mac — `evidence/it1-scenario-3.txt`
- [MISSING] scenario 4: environment: the branch is not pushed, so the Desktop workflow, a Release dry run and an install over a real release could not run. Steps 1 and 2 pass: the Linux build links GTK 4 / WebKitGTK 6.0 and its smoke start as a regular user exits 0; both Windows setup programs are built and hold the app and CLI byte for byte — `evidence/it1-scenario-4-linux.txt`, `evidence/it1-scenario-4-windows.txt`
- [MISSING] scenario 5: environment: `task showcase` needs your Kiro sign-in, Kiro IDE, kirocrew and SOPS key and was not run. Steps 1, 2, 3 and 6 are receipts; step 4, `task desktop:dev`, opened the window — `evidence/it1-desktop-dev.txt`
- [MISSING] dod T5: environment: Settings…, ⌘, , Edit's copy/paste and the Help entries could not be driven in the built app here (no accessibility access). The menu bar shows kirokit, Edit, Window, Help and `GOOS=windows go vet ./...` passes — `evidence/it1-scenario-3.txt`
- [MISSING] truth oracle: environment: the release workflow and CI's smoke starts have not run on this branch, which is not pushed. Local stand-ins pass: the macOS app and both DMGs with smoke exit 0, the Linux build's smoke exit 0 in Ubuntu 24.04, both Windows setup programs under today's names. Windows' own smoke start and the CI matrix remain — `evidence/it1-macos-dmg.txt`, `evidence/it1-scenario-4-linux.txt`, `evidence/it1-scenario-4-windows.txt`
- Every screen and action works as it does today: Every action in scenario 2 worked as before (clone, version step, Teams, both terminals, tag switch, Export for Quick, Open folder, the report link); only you can judge the feel, so use the built app for a while.
- The app has a native menu with the usual entries: On macOS the menu bar reads kirokit, Edit, Window, Help (`evidence/it1-menu-bar.png`); on Linux the window shows File, Edit, Window, Help (`evidence/it1-scenario-3-linux-menu.png`). The menus could not be opened here, so open each one on macOS, Windows and Linux, press ⌘, / Ctrl+, and try both Help entries.
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- Wails v3 pinned at v3.0.0 and its CLI through `tool github.com/wailsapp/wails/v3/cmd/wails3` in `desktop/go.mod` — keeps the existing `go tool` pinning pattern, no global install. Because v3's task files call a bare `wails3`, `desktop-build.sh` and `task desktop:dev` first build the pinned CLI into a folder they put first on `PATH` (the spike's gotcha 2). (Plan)
- The frontend keeps its own plain `vite.config.js`; Wails' vite plugin is not added — it requires generated bindings (spike gotcha 3). (Plan)
- macOS minimum stays 13.0, set in the committed darwin Taskfile and `Info.plist` (v3's template says 12.0) — matches today and Go 1.27's own minimum. (Plan)
- `desktop/wails.json` is replaced by v3's `desktop/build/config.yml` carrying the same product info (company `mavogel`, product `kirokit`, version `0.0.0`) — v3 reads its project info there; same values keep installer identity. (Plan)
- v3's default build step that generates bindings is removed from the committed Taskfile — nothing generated is committed or needed. (Plan)
- `feedback.js` keeps percent-encoding `! ' ( ) * ~` even though v3 no longer refuses them; the test's comment cites v3's validator — extra encoding is harmless and keeps old issue addresses identical. (Plan)
- Shortcut ⌘, (macOS) / Ctrl+, (Windows, Linux) opens Settings; all other shortcuts are the system's — the only new shortcut follows platform convention. (Plan)
- Help → "kirokit documentation" opens `https://github.com/mavogel/kirokit/blob/main/docs/desktop-app.md`, "Report a problem with kirokit…" opens `https://github.com/mavogel/kirokit/issues/new` — fixed addresses, no user input reaches them. (Plan)
- The dev-mode CLI lookup follows v3's dev output folder (`desktop/bin/`) instead of v2's `desktop/build/bin/` — keeps `task desktop:dev` using the freshly built `dev-bin` CLI. (Plan)
- The showcase preflight stops checking port 5173 — the server build serves the built page itself, so vite's dev port is no longer used. (Plan)
- The macOS icon is copied as `Contents/Resources/iconfile.icns` by the committed darwin Taskfile's `create:app:bundle` (and the dev bundle) instead of renamed afterwards in `scripts/desktop-build.sh`; the script still checks the file exists. No `Assets.car` is built or shipped (v2 shipped none and kirokit has no Icon Composer source); nothing you approved changes. (slice 1)
- The committed common Taskfile drops v3's `go mod tidy`, bindings and icon-generation steps, so a release build never rewrites `go.mod` or the committed icons; `icons.icns` was generated once from `desktop/build/appicon.png` with the pinned `wails3 generate icons`; nothing you approved changes. (slice 1)
- `desktop/frontend/src/pages/pages.test.js` also faked `window.go` for the Settings page, so it moved to the same runtime seam; nothing you approved changes. (slice 1)
- Adding Wails v3 raised the desktop module's indirect `golang.org/x/term`, `x/text` and `x/tools` to the versions v3.0.0 requires; nothing you approved changes. (slice 1)
- T6's smoke start runs as a regular user in a container with seccomp and /proc unconfined: WebKitGTK 6.0 always sandboxes its web process with bubblewrap, which Docker's defaults forbid (records/linux-ubuntu2404-smoke.txt). The Linux container moves to Ubuntu 24.04 because Wails v3 needs GTK 4.14 and Debian 12 has 4.8 (records/linux-debian12-gtk4.txt); debian:12 leaves the distro list as the Plan's measure-and-drop rule says (records/linux-distro-smoke.txt). (slice 3)
- The setup program's uninstall key and folder were read from `makensis -PPO` of the committed `project.nsi` (`HKLM\…\Uninstall\mavogelkirokit`, `$PROGRAMFILES64\mavogel\kirokit`) because this container's 7-Zip cannot decompile NSIS scripts; the packed `kirokit-desktop.exe` and `bin\kirokit.exe` are byte-identical to the build (records/windows-installer.txt); nothing you approved changes. (slice 3)
- V2's `desktop/build/windows/installer/` moved to v3's `desktop/build/windows/nsis/`; v3's `wails_tools.nsh` is committed with v2's product values (it is not regenerated by the build); nothing you approved changes. (slice 3)
- The `linux-distros` job builds the app on each runner (ubuntu-24.04, ubuntu-24.04-arm) and installs exactly the dependencies the `.deb` declares (read from `.goreleaser.yaml` with `yq`) in each image, instead of installing the snapshot `.deb`; the snapshot needs the GoReleaser Pro key and builds amd64 only, so this way both architectures and fork pull requests are checked. The installed-`.deb` start stays in the `desktop` job. Images are `ubuntu:24.04`, `ubuntu:26.04`, `debian:13`; `debian:12` was measured out (records/linux-distro-smoke.txt); nothing you approved changes. (slice 4)
- The Windows upgrade check counts uninstall entries whose `DisplayName`, `InstallLocation` or `UninstallString` mention kirokit (v2's installer writes no `InstallLocation`, so `UninstallString` is what names its folder); nothing you approved changes. (slice 4)
- `task desktop:dev` failed twice before it worked, and both causes are fixed: a universal `desktop/bin/kirokit` left by `task desktop:build` cannot be overwritten by Go (`scripts/desktop-build.sh` now removes `desktop/bin/` once it has staged its output), and vite bound only `[::1]:9245` while Wails' dev proxy dials `127.0.0.1` (the dev task now passes `--host 127.0.0.1`). The dev window then showed the Teams page from the test clone (records/macos-desktop-dev.png); nothing you approved changes. (slice 5)
- `task desktop:dev` builds the pinned `wails3` into `desktop/build/dev-bin/tools/` (already gitignored) and puts it first on `PATH`; the staged release app in `desktop/build/bin/` now uses its own bundled CLI even when started from `desktop/`, since `wails3 dev` no longer runs from there; nothing you approved changes. (slice 5)
- `app-server.sh` also fails when the port's listener is another process (it checks the listener's pid, not only its address), so a stray server on 34115 cannot pass for the app; its server build was driven in headless Chrome to the first-launch page with no page errors; nothing you approved changes. (slice 5)
- T10's Files add desktop/build/Taskfile.yml (vite binds 127.0.0.1, which Wails' dev proxy dials) and scripts/desktop-build.sh (removes Wails' intermediate desktop/bin after staging, since Go cannot overwrite the universal binary there when wails3 dev builds). The 'Decided while outlining' note on menu:settings now says main.js checks setup() itself: sidebar.select does not refuse while the first-launch page shows. (slice 5)
- Verification round 1 was recorded while 4 files changed, 4 of them code the run exercised; rerun it or keep it and say why here; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
- Still awaiting your confirmation: The menu reads right on each system.
- Still awaiting your confirmation: It feels the same to use.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Not shipped yet. Branch `main`.
<!-- /ql:ship -->
