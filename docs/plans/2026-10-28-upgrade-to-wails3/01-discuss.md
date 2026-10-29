---
answers: app
title: Upgrade desktop app to Wails v3
request: "upgrade to wails3 . see https://wails.io/migration/v2-to-v3/"
type: feature
depth: light
level: engineering
ui: true
overrides: []
stage: outline
iterations: 0
verification_failures: 0
gate: null
created: 1793213936030
startup_owner: 74d678ebe8123b18
layout: seven
profile: full
status: approved
approved: true
review_digest: 433362293b46adb6f15d4962bcab91795ac6c224a39be8619f7a9061ed2c2e85
review_basis: 6e9e2330c25251f790111e964500f9882ecad656adc2b5487a23cdb263473819
branch: main
base: b12c4a00ed9f9bd674f674b57f04554c7f402a1f
base_ref: main
---

# Upgrade desktop app to Wails v3

## Your request

> upgrade to wails3 . see https://wails.io/migration/v2-to-v3/

## Problem

The kirokit desktop app is built on Wails v2 (v2.17). Wails v3.0.0 is now the stable desktop release and v2 only receives critical fixes, so staying on v2 means the app drifts away from where Wails is maintained and improved. The app should move to Wails v3 without its users noticing anything but the same app.

## Done means

1. **The app runs on Wails v3 with no Wails v2 left.** The desktop module builds against Wails v3 only; nothing in the app, its build scripts or its tests still refers to Wails v2.
2. **Every screen and action works as it does today.** Clone, Settings, the team catalog, prerequisites, install and uninstall runs, branch/tag switching, Export for Quick, Open folder and the "report a problem" link behave as before. (only you can confirm: it feels the same to use)
3. **Both terminal tabs stream live and take input as before.** The kirokit tab shows a run's output as it happens and its exit; the shell tab opens your shell, accepts typing and resizes with the window.
4. **Live updates still arrive without a reload.** Install records changing on disk, a finished clone and a finished branch switch update the window on their own.
5. **Release builds keep their formats and names on all three systems.** macOS still gets a universal `kirokit.app` (13 or later) with the bundled CLI in the DMG, Linux the `.deb`/`.rpm`/archive with the bundled CLI, Windows a setup `.exe` per architecture; CI's smoke start passes on Ubuntu, Windows and macOS.
6. **Existing installs upgrade in place.** Installing the new release over the old one replaces it (Windows: no second entry in Apps & Features, same install folder) and the app keeps its saved settings and clone.
7. **The developer commands and the showcase still work.** `task desktop:dev`, `task desktop:test`, `task desktop:build`, `task ci` and `task showcase` work as before, and the docs that name the Wails tooling say the right thing.
8. **The app has a native menu with the usual entries.** On macOS the menu bar shows kirokit (About, Settings, Quit), Edit (undo, cut, copy, paste, select all), Window and Help (the docs, Report a problem); Windows and Linux show the same menu at the top of the window. Each item does what the matching button or shortcut does in the window. (only you can confirm: the menu reads right on each system)

## Scope

### In this task

- Moving `desktop/` (Go side, frontend glue, build config and packaging) to Wails v3.
- Updating the build script, Taskfile, CI and release workflows wherever they depend on Wails v2 behaviour.
- Keeping `task showcase` working: if v3's dev mode changes how the recorder launches or drives the app, the recorder is updated in this task.
- Adding a native application menu (the v3-only feature you picked).
- Updating `docs/desktop-app.md`, `CONTRIBUTING.md` and `AGENTS.md` where they describe the Wails tooling, and `docs/desktop-app.md` for the menu.

### Not in this task

- A system tray icon, a pop-out terminal window or a separate Settings window.
- Switching the frontend to v3's generated typed bindings (not chosen; the frontend keeps one hand-written module that reaches Go).
- Changes to the kirokit CLI or its rendered output.
- Redesigning any screen.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | Should the upgrade stay like-for-like, or also adopt things v3 offers? | Also add v3-only features (native menu, tray or extra windows) |
| 2 | If v3's dev mode breaks the showcase recorder, fix it in this task? | also update the recorder |
| 3 | Which v3-only features should the upgrade add? | Native application menu |
<!-- /ql:decided -->

## Context pointers

- https://wails.io/migration/v2-to-v3/ (the migration guide the request names; it states v3.0.0 is the stable desktop release)
- Earlier desktop tasks: `docs/plans/2026-10-17-desktop-app-from-design/`, `docs/plans/2026-10-25-feat-make-the-desktop-and-cli-also-avail/`, `docs/plans/2026-10-25-make-a-30-60s-video-showcase-of-how-kiro/`
- Project lesson: Wails v2.17 `BrowserOpenURL` refuses URLs with unescaped `! ' ( ) * ~`; `desktop/frontend/src/pages/feedback.js` encodes them and `feedback.test.js` mirrors Wails' character class.

## First look at the code

| Where | What depends on Wails v2 |
|---|---|
| `desktop/go.mod:8`, `desktop/go.mod:113` | `github.com/wailsapp/wails/v2 v2.17.0` and `tool github.com/wailsapp/wails/v2/cmd/wails` (the CLI is pinned through `go tool wails`) |
| `desktop/main.go:20-43` | `wails.Run(&options.App{...})`: window size, background colour, embedded assets, `OnStartup`/`OnShutdown`, `Bind: app`, and `OnDomReady` for the smoke start |
| `desktop/app.go:76-97` | `startup(ctx)` stores `runtime.EventsEmit(ctx, …)` as the event emitter; `shutdown` stops poller, shell and runner |
| `desktop/smoke.go:15` | `smokeWatch` is shaped as an `OnDomReady` hook taking `runtime.Quit(ctx)` |
| `desktop/frontend/src/api.js:5-9`, `:94-105` | the only frontend module reaching Go: calls `window.go.main.App.<Method>` and `window.runtime.BrowserOpenURL` / `EventsOn`; nothing generated is committed |
| `desktop/wails.json:1-20` | v2 project config (`config.v2.json` schema, frontend commands, product info) |
| `desktop/build/` | v2 packaging assets: `darwin/Info.plist`, `windows/installer/project.nsi`, `windows/info.json`, `wails.exe.manifest`, `linux/kirokit-desktop.desktop` |
| `scripts/desktop-build.sh:47`, `:62`, `:77` | `go tool wails build` per OS (`-platform darwin/universal`, `linux/<arch> -tags webkit2_41`, `windows/<arch> -nsis`) and the installer rename |
| `Taskfile.yml:43` | `desktop:dev` runs `go tool wails dev` |
| `.github/workflows/desktop.yaml:81`, `:119` | smoke start (`KIROKIT_DESKTOP_SMOKE=1`) of the built app and the installed `.deb` |
| `scripts/showcase/record.sh:152-164`, `scripts/showcase/app-drive.mjs:23` | the showcase drives `wails dev`'s browser URL `http://localhost:34115` from headless Chrome |
| Event names in `desktop/clone.go:51`, `desktop/installs.go:66`, `desktop/runner.go:123-163`, `desktop/shell.go:259-273`, `desktop/version.go:58` | `clone:done`, `installs:changed`, `term:kirokit`, `term:kirokit:exit`, `term:shell`, `ref:done`; subscribed in `desktop/frontend/src/main.js:289-317` and `desktop/frontend/src/terminal.js:147-156` |

*Untraced areas:* how v3's packaging (NSIS template, macOS bundle, Linux packages) maps onto `scripts/desktop-prebuilt.sh` and `.goreleaser.yaml`; whether v3's dev mode still lets a plain browser call bound methods (the showcase depends on it); v3's `Browser.OpenURL` URL validation.
