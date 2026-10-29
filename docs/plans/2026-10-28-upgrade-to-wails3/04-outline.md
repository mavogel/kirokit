---
type: outline
---

# Outline: Upgrade desktop app to Wails v3

## Summary

Six slices and fourteen tasks: first the thinnest working v3 app on macOS (Go start-up, page bridge, smoke start and the build adapter's macOS branch), then the native menu, then the Linux (GTK 4) and Windows (NSIS) builds, then CI and the release workflow, then `task desktop:dev` and the showcase's server build, and last the docs and a sweep for any v2 leftover. The riskiest spot is the build adapter (`scripts/desktop-build.sh`) staging v3's differently named output under today's paths for six consumers, and the Windows installer keeping v2's identity; both are proved locally where a container can build them and finally only by the CI and release runs.

## Where the change lands

<!-- ql:landing -->
- `desktop` · 4 new, 9 changed, 1 removed · slices 1, 2, 3, 5 · risky: outside system — GTK 4 / WebKitGTK 6.0 distro packages and NSIS under v3's templates, built here only in the Linux container.
- `desktop/build` · 8 new, 4 changed, 1 removed · slices 1, 3, 5 · risky: outside system — GTK 4 / WebKitGTK 6.0 distro packages and NSIS under v3's templates, built here only in the Linux container.
- `.` · 6 changed · slices 3, 5, 6 · risky: outside system — GTK 4 / WebKitGTK 6.0 distro packages and NSIS under v3's templates, built here only in the Linux container.
- `desktop/frontend` · 5 changed · slices 1, 2, 6
- `scripts/showcase` · 1 new, 2 changed · slice 5
- `.github/workflows` · 2 changed · slice 4
- `scripts` · 2 changed · slices 1, 3, 5 · risky: outside system — GTK 4 / WebKitGTK 6.0 distro packages and NSIS under v3's templates, built here only in the Linux container.
- `docs` · 1 changed · slice 6

Tests: 4 test files, named on the cards.
<!-- /ql:landing -->

## Program design

### The showcase's launch step is its own script
`scripts/showcase/app-server.sh <out-dir> <home>` builds the frontend and the server build into `<out-dir>`, starts it in its own process group with `WAILS_SERVER_HOST=127.0.0.1 WAILS_SERVER_PORT=34115` set explicitly and the given `HOME`, waits up to 180 s for `http://localhost:34115/`, aborts unless every 34115 listener is on loopback, and prints the server's pid. `record.sh` calls it and kills that process group in its cleanup; `showcase.test.mjs` calls it alone with `WAILS_SERVER_HOST=0.0.0.0` inherited, which is how the Plan's loopback test reaches "record.sh's launch step" without running the whole recorder.

### Menu actions are plain functions; the Wails adapter only wires them
`menuFor(goos)` returns data (Plan contract *Menu data*). `(*App).menuAction(action string, open func(string) error) func()` maps `settings` to `a.emit("menu:settings", nil)` and `docs`/`report` to `open(<constant URL>)`, logging an open error; an unknown action returns nil. `buildMenu(wapp, groups, a)` in `menu_wails.go` is the only code touching `application.Menu`; it is proved by the built app, not by unit tests.

### `start()` runs once, on the page's first load
`main.js`'s first `refresh()` uses `loadCatalog(api.start)`; later refreshes keep calling `api.state()`, so `smoke:ready` is sent once per page load.

## How each point is proved

<!-- ql:proof-map -->
- Done means 1 · The app runs on Wails v3 with no Wails v2 left → "The built macOS app starts on Wails v3 and the bridge answers both ways" · runs after slice 1
- Done means 2 · Every screen and action works as it does today → "Every page, both terminals and live updates work in the v3 app" · runs after slice 1
- Done means 3 · Both terminal tabs stream live and take input as before → "Every page, both terminals and live updates work in the v3 app" · runs after slice 1
- Done means 4 · Live updates still arrive without a reload → "Every page, both terminals and live updates work in the v3 app" · runs after slice 1
- Done means 5 · Release builds keep their formats and names on all three systems → "The built macOS app starts on Wails v3 and the bridge answers both ways" · runs after slice 1; "Linux and Windows packages build under today's names and install in place" · runs after slice 4
- Done means 6 · Existing installs upgrade in place → "Linux and Windows packages build under today's names and install in place" · runs after slice 4
- Done means 7 · The developer commands and the showcase still work → "The developer commands and the showcase work" · runs after slice 6
- Done means 8 · The app has a native menu with the usual entries → "The native menu reads right and its entries work" · runs after slice 2
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** the release workflow's macOS, Linux and Windows builds from this branch each produce today's file names, and each built app passes CI's smoke start, which now quits only after the page has loaded through Wails v3, received the answer to one bound call (`State`) and acknowledged it back to Go — a v2 leftover or a broken bridge in either direction fails it even when every unit test passes.
- **Truth:** `desktop/go.mod`, `desktop/go.sum` and every tracked file outside `docs/plans/` name no `wailsapp/wails/v2`, `window.go`, `window.runtime` or `wails dev`/`wails build` command after the last slice.

## Decided while outlining

- The showcase launch step moves into `scripts/showcase/app-server.sh` so the loopback test can run it alone (see *Program design*).
- `menu:settings` while the first-launch page shows does nothing (`main.js` checks `setup()` before selecting Settings) — the same as the window, which has no Settings button on that page.
- v3's own nfpm/AppImage tasks are not used: GoReleaser keeps building the `.deb`/`.rpm`, so the committed Linux Taskfile keeps only the build task.
- `desktop/build/windows/installer/` moves to v3's `desktop/build/windows/nsis/`; the v2 folder and its gitignored `tmp/` and `wails_tools.nsh` entries are removed with it.
- The Windows upgrade check and the distro matrix run only in CI; locally the Windows installer is built in the Linux container and its contents and uninstall identity are read back with 7-Zip and `strings`.

## Slice 1: The app starts and works on Wails v3 on macOS

After this slice the desktop module builds against Wails v3.0.0 only, the page reaches Go through `@wailsio/runtime`, and `task desktop:build` produces today's universal `kirokit.app` whose smoke start quits only after `State` answered and the page acknowledged it. Linux and Windows builds are not yet converted.

```diff
 desktop/
~├── go.mod                       ~ wails/v3 v3.0.0 and the wails3 tool replace v2
~├── go.sum                       ~ follows go.mod
~├── main.go                      ~ application.New + one window; smoke listens for smoke:ready
~├── app.go                       ~ App is a v3 service: ServiceStartup/ServiceShutdown, setEmitter
~├── smoke.go                     ~ smokeWatch takes a func() quit
~├── smoke_test.go                ~ new signature, same two cases
~├── shell_test.go                ~ calls ServiceShutdown
-├── wails.json                   - replaced by build/config.yml
+├── Taskfile.yml                 + v3 project Taskfile, APP_NAME kirokit
~├── .gitignore                   ~ ignores v3's bin/, drops v2's wailsjs entries
 ├── build/
+│   ├── config.yml               + product info (mavogel, kirokit, 0.0.0)
+│   ├── Taskfile.yml             + v3 common tasks, no bindings step
 │   └── darwin/
+│       ├── Taskfile.yml         + darwin build/package, minimum macOS 13.0
~│       ├── Info.plist           ~ v3 keys, CFBundleIconFile iconfile, minimum 13.0
~│       ├── Info.dev.plist       ~ v3 dev plist
+│       └── icons.icns           + app icon in v3's place
 └── frontend/
~    ├── package.json             ~ @wailsio/runtime 3.0.0
~    ├── package-lock.json        ~ follows package.json
     └── src/
~        ├── api.js               ~ Call.ByName / Events.On / Browser.OpenURL, start()
~        ├── api.test.js          ~ fake runtime seam; start() and on() cases
~        └── main.js              ~ first load calls start()
 scripts/
~└── desktop-build.sh             ~ darwin: pinned wails3 on PATH, v3 package task, stage under today's names
```

- [x] T1 Go starts the app through Wails v3 and quits a smoke start only on `smoke:ready`
  - **Objective:** Replace `wails.Run` with `application.New`, one window and `App` as a service, and make the smoke start wait for the page's `smoke:ready` event, so the desktop module no longer depends on Wails v2.
  - **Design:** "Go start-up: `application.New` + one window + menu, `App` as service", "Smoke start", "plan:decision:d1", "plan:decision:d3", "plan:decision:d9"
  - **Files:** `desktop/go.mod`, `desktop/go.sum`, `desktop/main.go`, `desktop/app.go`, `desktop/smoke.go`, `desktop/smoke_test.go`, `desktop/shell_test.go`
  - **Test first:** `desktop/smoke_test.go` calls `smokeWatch(timeout, func(){…}, fail)` with the new `func()` quit; it fails to compile against today's `func(context.Context)` signature.
  - **Definition of done:**
    - `smokeWatch`'s ready function quits once however often it runs, and a missing `smoke:ready` within the timeout calls fail and never quit.
    - `ServiceStartup` keeps an emitter a test already set; `ServiceShutdown` stops the poller, shell and runner (the existing `shell_test.go` cleanup still passes).
    - `go list -m all` in `desktop/` names `github.com/wailsapp/wails/v3 v3.0.0` and no `wails/v2`; `go tool wails3 version` prints `v3.0.0`.
    - verify: `cd desktop && go vet ./... && go test ./...`

- [x] T2 The page reaches Go through `@wailsio/runtime` and acknowledges its first `State`
  - **Objective:** Move `api.js` to `Call.ByName("main.App.<Method>")`, `Events.On` (handing handlers `event.data`) and `Browser.OpenURL`, add `start()` that emits `smoke:ready` after `State` resolves, and call it on the page's first load.
  - **Design:** "Page bridge: `api.js` exports unchanged; one runtime seam", "Smoke start", "`start()` runs once, on the page's first load", "plan:decision:d2"
  - **Files:** `desktop/frontend/package.json`, `desktop/frontend/package-lock.json`, `desktop/frontend/src/api.js`, `desktop/frontend/src/api.test.js`, `desktop/frontend/src/main.js`, `desktop/frontend/src/pages/pages.test.js`
  - **Test first:** `api.test.js` installs `globalThis.__kirokitRuntime` with fake `Call`, `Events` and `Browser`, then asserts `shellWrite('c')` calls `Call.ByName('main.App.ShellWrite', 'c')`; it fails because `api.js` still reads `window.go`.
  - **Definition of done:**
    - Every bound call goes to `main.App.<Method>` with its arguments; keystrokes keep their per-terminal order and a failed one does not hold back the next.
    - `on(event, h)` hands `h` the event's `data`, not the envelope, and returns the unsubscribe function.
    - `start()` emits `smoke:ready` only after `State` resolved and resolves to the state; a rejecting `State` rejects `start()` and emits nothing.
    - `openURL` keeps refusing every non-issue-form address; outside the window (no runtime object) a call or `openURL` rejects with `kirokit is not running inside its window`.
    - A Go error's message reaches the page unchanged (`RuntimeError` with `err.Error()`), so `main.js`'s `RepoDir: ` strip still applies.
    - verify: `cd desktop/frontend && npm ci && npm run build && node --test "src/**/*.test.js"`

- [x] T3 `task desktop:build` builds the universal macOS app with Wails v3 under today's names
  - **Objective:** Commit v3's project files (config, Taskfiles, darwin assets) and make `scripts/desktop-build.sh`'s darwin branch build the pinned `wails3`, put it first on `PATH`, run `darwin:package:universal`, and stage the result as `desktop/build/bin/kirokit.app` with `iconfile.icns`, the bundled CLI and an ad-hoc signature.
  - **Design:** "Build adapter: `scripts/desktop-build.sh`", "plan:decision:d4"
  - **Files:** `desktop/wails.json`, `desktop/Taskfile.yml`, `desktop/.gitignore`, `desktop/build/config.yml`, `desktop/build/Taskfile.yml`, `desktop/build/darwin/Taskfile.yml`, `desktop/build/darwin/Info.plist`, `desktop/build/darwin/Info.dev.plist`, `desktop/build/darwin/icons.icns`, `scripts/desktop-build.sh`
  - **Test first:** **Trivial:** packaging config; covered by `task desktop:build` followed by the smoke start and the bundle checks below, which fail today because `go tool wails` no longer exists once T1 removed v2.
  - **Definition of done:**
    - `desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` is universal (`lipo -archs`: `x86_64 arm64`) with minimum macOS 13.0 (`vtool -show-build`), `Contents/Resources/iconfile.icns` and `Contents/Resources/bin/kirokit` exist, and `codesign --verify --deep --strict` passes.
    - A missing expected v3 output stops the script with `desktop-build: missing <path>` and a non-zero exit.
    - `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` exits 0; `scripts/desktop-prebuilt.sh` stages the bundle unchanged.
    - verify: `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit && scripts/desktop-prebuilt.sh`

**Validation:** automated: `task desktop:test` · `task desktop:build` · `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` · manual: open the built app and click through the pages and both terminal tabs

## Slice 2: The app has a native menu on every system

**Depends on:** Slice 1

After this slice the macOS menu bar shows kirokit, Edit, Window and Help, and Windows and Linux show File, Edit, Window and Help in the window. Settings… (⌘, / Ctrl+,) opens the Settings page and Help opens kirokit's docs and new-issue form.

```diff
 desktop/
+├── menu.go                      + menuFor(goos) data and App.menuAction
+├── menu_test.go                 + per-OS menu shape and actions, on any machine
+├── menu_wails.go                + buildMenu: data -> application.Menu with roles
~├── main.go                      ~ sets the application menu, window uses it
 └── frontend/src/
~    └── main.js                  ~ menu:settings selects the Settings page
```

- [x] T4 `menuFor(goos)` describes each system's menu and its actions
  - **Objective:** Add the menu as data per OS and the action mapping, tested for darwin, linux and windows on any machine.
  - **Design:** "Menu data", "Menu actions are plain functions; the Wails adapter only wires them", "plan:decision:d7"
  - **Files:** `desktop/menu.go`, `desktop/menu_test.go`
  - **Test first:** `menu_test.go` asserts darwin's group titles are `kirokit, Edit, Window, Help` and windows/linux's are `File, Edit, Window, Help`; it fails because `menuFor` does not exist.
  - **Definition of done:**
    - darwin: `kirokit` holds About, Settings… (`CmdOrCtrl+,`), Services, Hide, Hide Others, Show All and Quit in that order with separators; Help holds the docs and report entries and no About.
    - windows and linux: `File` holds Settings… (`CmdOrCtrl+,`) and Quit; Help ends with About; no group uses Wails' Help role.
    - Settings is the only item with a shortcut; every non-role item has a known action.
    - `settings` emits `menu:settings`; `docs` and `report` open exactly the two fixed addresses; an open error is logged and not raised; an unknown action yields nil.
    - verify: `cd desktop && go test -run 'Menu' ./...`

- [x] T5 The menu appears in the app and Settings… opens the Settings page
  - **Objective:** Turn the data into a Wails menu with roles, set it as the application menu (shown in the window on Windows and Linux), and make the page answer `menu:settings` by selecting Settings.
  - **Design:** "Menu data", "Go start-up: `application.New` + one window + menu, `App` as service", "Page bridge: `api.js` exports unchanged; one runtime seam"
  - **Files:** `desktop/menu_wails.go`, `desktop/main.go`, `desktop/frontend/src/main.js`
  - **Test first:** **Trivial:** wiring only; `menu_test.go` (T4) covers the data and actions, and the built app's menu is checked by Scenario 3.
  - **Definition of done:**
    - The built macOS app's menu bar shows kirokit, Edit, Window, Help; Settings… and ⌘, open the Settings page; Edit's copy/paste work in the Settings fields; Help's entries open the docs page and the new-issue form in the browser.
    - The app still builds for linux and windows (`GOOS=windows go vet ./...` in `desktop/`).
    - verify: `cd desktop && go vet ./... && GOOS=windows go vet ./... && go test ./...`

**Validation:** automated: `cd desktop && go test ./...` · `task desktop:build` · manual: the menu on macOS (Scenario 3)

## Slice 3: Linux on GTK 4 and Windows setup programs build with Wails v3

**Depends on:** Slice 1
**Risky:** outside system — GTK 4 / WebKitGTK 6.0 distro packages and NSIS under v3's templates, built here only in the Linux container.

After this slice `task desktop:build` produces `kirokit-desktop` (GTK 4 / WebKitGTK 6.0) on Linux and one setup program per architecture on Windows, under today's names, and the Linux packages declare the GTK 4 dependencies. The Windows setup program keeps v2's install folder, executable names and uninstall key.

```diff
 desktop/build/
 ├── linux/
+│   └── Taskfile.yml             + linux build task (GTK 4, no gtk3 tag)
 └── windows/
+    ├── Taskfile.yml             + windows build + NSIS package tasks
~    ├── info.json                ~ v3 version info, same company/product
~    ├── wails.exe.manifest       ~ v3 manifest
     ├── nsis/
+    │   ├── project.nsi          + v3 template with kirokit-desktop.exe and bin\kirokit.exe
+    │   └── wails_tools.nsh      + v3 NSIS helpers
     └── installer/
-        └── project.nsi          - v2 template
 scripts/
~├── desktop-build.sh             ~ linux + windows branches on v3 tasks, staged under today's names
~└── linux-docker.Dockerfile      ~ oldest claimed Debian/Ubuntu base, GTK 4 / WebKitGTK 6.0 dev
~.goreleaser.yaml                 ~ deb/rpm depend on GTK 4 / WebKitGTK 6.0
```

- [x] T6 Linux builds `kirokit-desktop` against GTK 4 / WebKitGTK 6.0
  - **Objective:** Add v3's Linux build task and the script's Linux branch, move the Linux container to a base with WebKitGTK 6.0 dev packages, and change the packages' declared dependencies.
  - **Design:** "Build adapter: `scripts/desktop-build.sh`", "Linux packaging and hosts", "plan:decision:d6"
  - **Files:** `desktop/build/linux/Taskfile.yml`, `desktop/Taskfile.yml`, `scripts/desktop-build.sh`, `scripts/linux-docker.Dockerfile`, `.goreleaser.yaml`
  - **Test first:** **Trivial:** build config; `scripts/linux-docker.sh task desktop:build` fails today (the container lacks WebKitGTK 6.0 and the script still calls v2) and passes after.
  - **Definition of done:**
    - In the container, `desktop/build/bin/kirokit-desktop` and `desktop/build/bin/kirokit` exist, `ldd kirokit-desktop` lists `libwebkitgtk-6.0` and `libgtk-4`, and no `webkit2gtk-4.1`.
    - `KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a desktop/build/bin/kirokit-desktop` exits 0 for a regular user in the container (WebKitGTK 6.0's sandbox needs the namespaces a desktop session has: seccomp and `/proc` unconfined).
    - `.goreleaser.yaml` declares deb `libgtk-4-1, libwebkitgtk-6.0-4` and rpm `gtk4, webkitgtk6.0`; `goreleaser check` passes when GoReleaser is installed.
    - verify: `scripts/linux-docker.sh task desktop:build && docker run --rm --security-opt seccomp=unconfined --security-opt systempaths=unconfined -v "$PWD":/src:ro kirokit-linux-check bash -c 'useradd -m tester && su tester -c "cd && KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /src/desktop/build/bin/kirokit-desktop"'`

- [x] T7 Windows builds a setup program per architecture that installs over v2's
  - **Objective:** Commit v3's Windows assets with kirokit's NSIS customisations and make the script's Windows branch build each architecture's app and setup program, staged under today's names.
  - **Design:** "Build adapter: `scripts/desktop-build.sh`", "plan:decision:d5"
  - **Files:** `desktop/build/windows/Taskfile.yml`, `desktop/build/windows/info.json`, `desktop/build/windows/wails.exe.manifest`, `desktop/build/windows/nsis/project.nsi`, `desktop/build/windows/nsis/wails_tools.nsh`, `desktop/build/windows/installer/project.nsi`, `desktop/.gitignore`, `scripts/desktop-build.sh`
  - **Test first:** **Trivial:** packaging config; `scripts/linux-docker.sh env GOOS=windows task desktop:build` fails today and passes after, and the 7-Zip listing below checks what the installer holds.
  - **Definition of done:**
    - `desktop/build/bin/installers/kirokit-desktop_<v>_windows_{amd64,arm64}-setup.exe` and `desktop/build/bin/kirokit-desktop-{amd64,arm64}.exe` exist.
    - `7z l` of each setup program lists `kirokit-desktop.exe` and `bin/kirokit.exe`; the installer's strings carry the uninstall key `mavogelkirokit` and the folder `mavogel\kirokit`.
    - A missing expected v3 output stops the script with `desktop-build: missing <path>`.
    - verify: `scripts/linux-docker.sh env GOOS=windows task desktop:build`

**Validation:** automated: `scripts/linux-docker.sh task desktop:build` · `scripts/linux-docker.sh env GOOS=windows task desktop:build` · manual: none

## Slice 4: CI and the release build, smoke-start and check the v3 app on every system

**Depends on:** Slice 3

After this slice the Desktop and Release workflows install GTK 4 / WebKitGTK 6.0, build Linux on Ubuntu 24.04 runners, run the installed `.deb`'s smoke start in each candidate distro, and check on Windows that the new setup program installs over the last release. Its proof is the CI run of this branch, which needs the branch pushed.

```diff
 .github/workflows/
~├── desktop.yaml                 ~ GTK 4 deps, ubuntu-24.04, distro matrix, Windows upgrade check
~└── release.yaml                 ~ GTK 4 deps, ubuntu-24.04 / ubuntu-24.04-arm runners
```

- [x] T8 The Desktop workflow builds v3 on pinned runners and checks distros and Windows upgrades
  - **Objective:** Update the Desktop workflow's Linux dependencies and runner, add the distro smoke matrix and the Windows upgrade-in-place check.
  - **Design:** "Linux packaging and hosts", "Windows upgrade check", "plan:decision:d5", "plan:decision:d6"
  - **Files:** `.github/workflows/desktop.yaml`
  - **Test first:** **Trivial:** workflow config; checked locally with `actionlint` when installed and a YAML parse, and proved by the branch's CI run.
  - **Definition of done:**
    - The Linux desktop job runs on `ubuntu-24.04` and installs `libgtk-4-dev libwebkitgtk-6.0-dev`.
    - The distro job installs the snapshot's `.deb` in `ubuntu:24.04`, `ubuntu:26.04`, `debian:12`, `debian:13` on amd64 and arm64 and runs `KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /usr/bin/kirokit-desktop`.
    - The Windows check installs the latest release, then the new setup program silently, compares `Get-FileHash` of the installed app and CLI with the new build, requires exactly one uninstall entry mentioning kirokit (`mavogelkirokit`), and skips with a notice when no release exists.
    - verify: `node -e "require('node:fs'); const y=require('node:fs').readFileSync('.github/workflows/desktop.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-latest/.test(y)) process.exit(1)"`

- [x] T9 The Release workflow builds Linux on the oldest claimed Ubuntu with GTK 4
  - **Objective:** Point the release's Linux builds at `ubuntu-24.04` / `ubuntu-24.04-arm` and the GTK 4 / WebKitGTK 6.0 dev packages.
  - **Design:** "Linux packaging and hosts", "plan:decision:d6"
  - **Files:** `.github/workflows/release.yaml`
  - **Test first:** **Trivial:** workflow config; proved by the branch's release dry run.
  - **Definition of done:**
    - `app-linux-arm64` runs on `ubuntu-24.04-arm` and the linux `split` entry on `ubuntu-24.04`; both install `libgtk-4-dev libwebkitgtk-6.0-dev`.
    - verify: `node -e "const y=require('node:fs').readFileSync('.github/workflows/release.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-26\.04-arm/.test(y)) process.exit(1)"`

**Validation:** automated: `actionlint .github/workflows/desktop.yaml .github/workflows/release.yaml` (when installed) · manual: the branch's Desktop run and a Release dry run (`workflow_dispatch`) pass on GitHub — needs the branch pushed

## Slice 5: Developer commands and the showcase work on Wails v3

**Depends on:** Slice 1

After this slice `task desktop:dev` opens the app with `wails3 dev` and the freshly built dev CLI, and `task showcase` records the app clip from a loopback-only server build on port 34115 instead of v2's dev server.

```diff
 Taskfile.yml                     ~ desktop:dev builds the pinned wails3 and runs wails3 dev
 desktop/
~├── shell.go                     ~ dev lookup follows v3's desktop/bin
~└── shell_test.go                ~ dev-bin wins under desktop/bin
 scripts/showcase/
+├── app-server.sh                + builds and starts the loopback-only server build
~├── record.sh                    ~ app step uses app-server.sh; preflight drops 5173
~├── app-drive.mjs                ~ comment names the server build
~└── showcase.test.mjs            ~ loopback case; busy-port message
```

- [x] T10 `task desktop:dev` runs Wails v3's dev mode with the dev CLI
  - **Objective:** Make `desktop:dev` build the pinned `wails3` onto `PATH` and run `wails3 dev`, and make the bundled-CLI lookup prefer `build/dev-bin` when the app runs from v3's `desktop/bin`.
  - **Design:** "plan:decision:d4"
  - **Files:** `Taskfile.yml`, `desktop/shell.go`, `desktop/shell_test.go`, `desktop/build/Taskfile.yml`, `scripts/desktop-build.sh`
  - **Test first:** a `shell_test.go` case with the executable at `<wd>/bin/kirokit` and both a bundled CLI and `build/dev-bin/kirokit` present expects the dev-bin one; it fails because the lookup only knows `<wd>/build/bin`.
  - **Definition of done:**
    - Under `<wd>/bin/` the dev-bin CLI comes first on every OS name; an installed app's lookup order is unchanged.
    - `task desktop:dev` opens the kirokit window and its kirokit tab runs `desktop/build/dev-bin/kirokit`.
    - verify: `cd desktop && go test -run 'CLI|Resolve' ./...`

- [x] T11 The showcase records the app from a loopback-only server build
  - **Objective:** Move the app step's launch into `app-server.sh` (server build, explicit loopback host and port, readiness poll, loopback check) and call it from `record.sh`.
  - **Design:** "Showcase server build", "The showcase's launch step is its own script", "plan:decision:d8"
  - **Files:** `scripts/showcase/app-server.sh`, `scripts/showcase/record.sh`, `scripts/showcase/app-drive.mjs`, `scripts/showcase/showcase.test.mjs`
  - **Test first:** `showcase.test.mjs` runs `app-server.sh` with `WAILS_SERVER_HOST=0.0.0.0` inherited and asserts every 34115 listener is `127.0.0.1`; it fails because the script does not exist.
  - **Definition of done:**
    - With `WAILS_SERVER_HOST=0.0.0.0` inherited the server still listens only on loopback, and `http://localhost:34115/` serves the page.
    - A server that exits before it answers fails naming its log; a non-loopback listener aborts the run and stops the server.
    - Preflight checks 34115 and 34117 only, and the busy-port message names "a running kirokit app server or kirocrew gateway".
    - verify: `node --test scripts/showcase/`

**Validation:** automated: `node --test scripts/showcase/` · `cd desktop && go test ./...` · manual: `task desktop:dev` opens the window; `task showcase` produces both assets

## Slice 6: The docs describe Wails v3, the menu and GTK 4, and no v2 is left

**Depends on:** Slice 4, Slice 5

After this slice the docs name `wails3`, the menu and the GTK 4 dependencies and list the measured distros, and nothing outside the task records refers to Wails v2.

```diff
 docs/desktop-app.md              ~ menu, GTK 4 deps, measured distros, dev mode
 CONTRIBUTING.md                  ~ Linux deps, container contents
 AGENTS.md                        ~ desktop module line
 README.md                        ~ Linux distro list
 renovate.json                    ~ wails3 tool pin guard
 desktop/frontend/src/pages/
~├── feedback.js                  ~ comment cites v3
~└── feedback.test.js             ~ comment cites v3's validator
```

- [x] T12 The docs say Wails v3, describe the menu and list the Linux systems that pass
  - **Objective:** Update every doc that names Wails tooling, the Linux dependencies or supported distros, and describe the menu.
  - **Design:** none
  - **Files:** `docs/desktop-app.md`, `CONTRIBUTING.md`, `AGENTS.md`, `README.md`
  - **Test first:** **Trivial:** documentation; `node scripts/check-doc-links.mjs` checks links and anchors.
  - **Definition of done:**
    - The Linux distro list in `README.md`, `docs/desktop-app.md` and the `.goreleaser.yaml` comment is the same and matches the distro check's passing images.
    - verify: `node scripts/check-doc-links.mjs && node scripts/sync-agent-assets.mjs --check`

- [x] T13 Renovate and the remaining comments track Wails v3
  - **Objective:** Point Renovate's tool-pin guard at `wails/v3/cmd/wails3` and update the comments that cite v2's URL validator.
  - **Design:** none
  - **Files:** `renovate.json`, `desktop/frontend/src/pages/feedback.js`, `desktop/frontend/src/pages/feedback.test.js`
  - **Test first:** **Trivial:** config and comments; `feedback.test.js` keeps passing.
  - **Definition of done:**
    - verify: `cd desktop/frontend && node --test src/pages/feedback.test.js`

- [x] T14 Nothing outside the task records refers to Wails v2
  - **Objective:** Prove the Oracle's *Truth* line across the repository.
  - **Design:** none
  - **Files:** none
  - **Read only:** every tracked file
  - **Qualification:** `git grep` for v2 module paths, `window.go`, `window.runtime`, `go tool wails ` and `wails dev`, excluding `docs/plans/`, returns nothing; `task ci` and `task desktop:test` pass.
  - **Definition of done:**
    - verify: `! git grep -nE 'wailsapp/wails/v2|window\.go\b|window\.runtime|go tool wails |wails dev|wails build' -- ':!docs/plans/' && task ci && task desktop:test`

**Validation:** automated: `node scripts/check-doc-links.mjs` · `node scripts/sync-agent-assets.mjs --check` · `task ci` · `task desktop:test` · manual: none

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Menu as per-OS data and its actions | `desktop/menu.go` | `installedCLIFor(goos, …)` pattern in `desktop/platform.go` |
| Menu → Wails adapter | `desktop/menu_wails.go` | `application.NewMenu`, `AddRole` |
| Runtime seam for page tests | `desktop/frontend/src/api.js` (`globalThis.__kirokitRuntime`) | the `globalThis` fakes in `api.test.js` |
| Showcase launch step | `scripts/showcase/app-server.sh` | `record.sh`'s process-group cleanup |

## End-to-end scenarios

### Scenario 1: The built macOS app starts on Wails v3 and the bridge answers both ways

**Runs after:** slice 1
**Proves:** Done means 1, Done means 5
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `cd desktop && go list -m all \| grep -c wailsapp/wails/v2` | `0` |
| 2 | `task desktop:build` | exit 0; prints `built desktop/build/bin/kirokit.app` |
| 3 | `lipo -archs desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` | `x86_64 arm64` |
| 4 | `test -x desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit && test -f desktop/build/bin/kirokit.app/Contents/Resources/iconfile.icns && echo ok` | `ok` |
| 5 | `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?` | `0` within 60 s |
| 6 | `scripts/desktop-prebuilt.sh` | prints `staged …/desktop/build/bin/prebuilt` |

### Scenario 2: Every page, both terminals and live updates work in the v3 app

**Runs after:** slice 1
**Proves:** Done means 2, Done means 3, Done means 4
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Open `desktop/build/bin/kirokit.app` with a fresh `HOME` (`HOME=$(mktemp -d) desktop/build/bin/kirokit.app/Contents/MacOS/kirokit`) | The first-launch clone page shows |
| 2 | Clone `https://github.com/mavogel/kirokit` and pick Newest release | The clone finishes and the Teams page appears without a reload (`clone:done`, `ref:done`) |
| 3 | Install a team from Teams | The kirokit tab streams the run's output live and shows its exit; the Install page updates on its own (`installs:changed`) |
| 4 | Open the shell tab, type `echo hi`, press Enter, then resize the window | `hi` prints; the prompt re-wraps to the new width |
| 5 | Open Settings, switch to another tag, then back | The switch runs in the kirokit tab and the version line updates |
| 6 | Export for Quick on a team, then Open folder | The bundle renders and Finder opens its folder |
| 7 | On the Agents page, use "Report a problem" | The browser opens a GitHub new-issue form for the repository |

### Scenario 3: The native menu reads right and its entries work

**Runs after:** slice 2
**Proves:** Done means 8
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `cd desktop && go test -run Menu ./...` | `ok` |
| 2 | Start the built macOS app and open each menu | kirokit (About, Settings…, Services, Hide, Hide Others, Show All, Quit), Edit, Window, Help (kirokit documentation, Report a problem with kirokit…) |
| 3 | Press ⌘, | The Settings page opens |
| 4 | Copy text from a Settings field with ⌘C and paste it into another with ⌘V | The text is pasted |
| 5 | Help → kirokit documentation | The browser opens `docs/desktop-app.md` on GitHub |
| 6 | On Linux (container with a display) or Windows, start the app | A File, Edit, Window, Help menu bar shows at the top of the window |

### Scenario 4: Linux and Windows packages build under today's names and install in place

**Runs after:** slice 4
**Proves:** Done means 5, Done means 6
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `scripts/linux-docker.sh task desktop:build && docker run --rm --security-opt seccomp=unconfined --security-opt systempaths=unconfined -v "$PWD":/src:ro kirokit-linux-check bash -c 'useradd -m tester && su tester -c "cd && KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /src/desktop/build/bin/kirokit-desktop"'` | exit 0 |
| 2 | `scripts/linux-docker.sh env GOOS=windows task desktop:build` | exit 0; lists both `kirokit-desktop_*_windows_{amd64,arm64}-setup.exe` |
| 3 | Push the branch and open its pull request | The Desktop workflow passes on Ubuntu, Windows and macOS, including the distro matrix and the Windows upgrade check (one uninstall entry `mavogelkirokit`, same install folder) |
| 4 | Run the Release workflow by hand with dry run | The printed release files keep today's names (DMGs, `.deb`/`.rpm`, setup programs) |
| 5 | Install the new macOS app over the previous release and start it | The saved settings and clone are still there |

### Scenario 5: The developer commands and the showcase work

**Runs after:** slice 6
**Proves:** Done means 7
**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | `task desktop:test` | exit 0 |
| 2 | `task ci` | exit 0 |
| 3 | `node --test scripts/showcase/` | every test passes or skips for a missing tool |
| 4 | `task desktop:dev` | The kirokit window opens; its kirokit tab runs `desktop/build/dev-bin/kirokit` |
| 5 | `task showcase` | `assets/kirokit-showcase.gif` and `.mp4` are rewritten |
| 6 | `node scripts/check-doc-links.mjs && node scripts/sync-agent-assets.mjs --check` | exit 0 |

## Commands the build runs

<!-- ql:commands -->
- `task desktop:test` · slice 1, slice 6, scenario 5, project checks
- `task desktop:build` · slice 1, slice 2, scenario 1, project checks
- `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` · slice 1
- `cd desktop && go test ./...` · slice 2, slice 5
- `scripts/linux-docker.sh task desktop:build` · slice 3
- `scripts/linux-docker.sh env GOOS=windows task desktop:build` · slice 3, T7, scenario 4
- `actionlint .github/workflows/desktop.yaml .github/workflows/release.yaml` · slice 4
- `node --test scripts/showcase/` · slice 5, T11, scenario 5, project checks
- `node scripts/check-doc-links.mjs` · slice 6, project checks
- `node scripts/sync-agent-assets.mjs --check` · slice 6, project checks
- `task ci` · slice 6, scenario 5, project checks
- `cd desktop && go vet ./... && go test ./...` · T1
- `cd desktop/frontend && npm ci && npm run build && node --test "src/**/*.test.js"` · T2
- `task desktop:build && KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit && scripts/desktop-prebuilt.sh` · T3
- `cd desktop && go test -run 'Menu' ./...` · T4
- `cd desktop && go vet ./... && GOOS=windows go vet ./... && go test ./...` · T5
- `scripts/linux-docker.sh task desktop:build && docker run --rm --security-opt seccomp=unconfined --security-opt systempaths=unconfined -v "$PWD":/src:ro kirokit-linux-check bash -c 'useradd -m tester && su tester -c "cd && KIROKIT_DESKTOP_SMOKE=1 xvfb-run -a /src/desktop/build/bin/kirokit-desktop"'` · T6, scenario 4
- `node -e "require('node:fs'); const y=require('node:fs').readFileSync('.github/workflows/desktop.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-latest/.test(y)) process.exit(1)"` · T8
- `node -e "const y=require('node:fs').readFileSync('.github/workflows/release.yaml','utf8'); if(/libgtk-3|webkit2gtk-4\.1|ubuntu-26\.04-arm/.test(y)) process.exit(1)"` · T9
- `cd desktop && go test -run 'CLI|Resolve' ./...` · T10
- `node scripts/check-doc-links.mjs && node scripts/sync-agent-assets.mjs --check` · T12, scenario 5
- `cd desktop/frontend && node --test src/pages/feedback.test.js` · T13
- `! git grep -nE 'wailsapp/wails/v2|window\.go\b|window\.runtime|go tool wails |wails dev|wails build' -- ':!docs/plans/' && task ci && task desktop:test` · T14
- `lipo -archs desktop/build/bin/kirokit.app/Contents/MacOS/kirokit` · scenario 1
- `test -x desktop/build/bin/kirokit.app/Contents/Resources/bin/kirokit && test -f desktop/build/bin/kirokit.app/Contents/Resources/iconfile.icns && echo ok` · scenario 1
- `KIROKIT_DESKTOP_SMOKE=1 desktop/build/bin/kirokit.app/Contents/MacOS/kirokit; echo $?` · scenario 1
- `scripts/desktop-prebuilt.sh` · scenario 1
- `cd desktop && go test -run Menu ./...` · scenario 3
- `task desktop:dev` · scenario 5
- `task showcase` · scenario 5
<!-- /ql:commands -->
