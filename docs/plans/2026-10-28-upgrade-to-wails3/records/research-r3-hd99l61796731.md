---
type: research
---

# Showcase recorder and Wails tooling statements in docs

## Findings

### The recorder runs the app in `wails dev` and drives it from headless Chrome over http://localhost:34115
**Answers:** 7

- `scripts/showcase/record.sh:146-155` starts `go tool wails dev` from `desktop/` in its own process group (`set -m`), with a throwaway `HOME`, the real Go/npm caches, and the output in `$out/wails.log`. The only Wails invocation in the script is `npm_config_cache="$npmcache" go tool wails dev` (`scripts/showcase/record.sh:152`).
- Before that, `scripts/showcase/record.sh:85` builds the CLI into `desktop/build/dev-bin/kirokit`, which the app prefers over a bundled CLI in dev mode (`desktop/shell.go:36`, `desktop/shell.go:55`; test at `desktop/shell_test.go:214`).
- Readiness: a 720 x 0.25 s loop (180 s) greps `$out/wails.log` for the literal `Using DevServer URL: http://localhost:34115` and aborts if the wails process dies (`scripts/showcase/record.sh:158-164`). That log line is Wails v2 dev-mode output; the recorder depends on its exact text.
- Ports 34115, 5173 and 34117 are checked free with `lsof` in preflight (`scripts/showcase/record.sh:66-70`). 34115 is the Wails dev server URL the driver opens (`scripts/showcase/app-drive.mjs:23`). 5173 is the vite dev server that `wails.json` launches via `"frontend:dev:watcher": "npm run dev"` with `"frontend:dev:serverUrl": "auto"` (`desktop/wails.json:7-8`; the 5173 default is inferred from vite, and `scripts/showcase/showcase.test.mjs:388-389` says "vite binds [::1]:5173"). 34117 is not the app: it is the kirocrew gateway port (`scripts/showcase/crew-drive.mjs:17`, `scripts/showcase/crew-drive.mjs:31`).
- `scripts/showcase/app-drive.mjs:34` launches installed Google Chrome (`channel: 'chrome'`, headless) through `playwright-core`, opens a 1180x760 context with video recording (`app-drive.mjs:37-41`) and does `page.goto('http://localhost:34115')` (`app-drive.mjs:66`). No Wails window is used; the app is a plain page in Chrome.
- Dependency: the frontend only reaches Go through `window.go.main.App` (bound methods) and `window.runtime` (events, `BrowserOpenURL`), per `desktop/frontend/src/api.js:1-8` and `desktop/frontend/src/api.js:96-104`. Without them `api.js:7` throws "kirokit is not running inside its window". Since the driver's Chrome is not a Wails webview, this works only because the v2 dev server at 34115 serves the page with Wails' injected runtime/IPC bridge (the dev server proxies to vite and adds the bindings over a websocket). This is an inference from the passing recording flow, not read from Wails source. The driver depends on it for: Clone (`#clone-form`), version step, Install, and the dock (xterm rows fed by `term:kirokit` events, `desktop/frontend/src/main.js:293-300`; `app-drive.mjs:117-122` waits on an xterm row `installed platform-team into`). Events are emitted from Go with `runtime.EventsEmit` (`desktop/app.go:78`).
- So the recorder relies on a browser at http://localhost:34115 calling bound Go methods (State, Clone etc.) and receiving events. It also relies on stable DOM ids/selectors (`#clone-address`, `button[data-action="install"]`, `.card .badge.ok`, `#term-kirokit .xterm-rows`; `app-drive.mjs:67-122`).
- After the run, `record.sh:168-170` stops the whole process group (`kill -TERM -- -$wails`) so vite and the dev app die with it.
- `app-drive.mjs` writes `app.webm`, `app-dock.txt` (privacy check input) and `app-marks.json` (`app-drive.mjs:7-10`); `record.sh:256` feeds `app-dock.txt` to `checks.sh privacy`.

### The recorder's tests only cover the failure paths around Wails, not a real dev server
**Answers:** 7

- `scripts/showcase/showcase.test.mjs:394-406`: usage error without outdir (exit 2), and "without a dev server it stops at the welcome step" (skips if port 34115 is busy or Chrome is absent; expects stderr `app-drive: welcome: `).
- `scripts/showcase/showcase.test.mjs:608-624`: occupies 34115 with a plain `net` server and expects `showcase: preflight: port 34115 is busy: close the running wails dev or kirocrew gateway first` (this message text mentions `wails dev`; `record.sh:68` is the source).
- `showcase.test.mjs:572` and `showcase.test.mjs:590` skip when 34115 or 5173 is in use. No test starts `wails dev`, checks the `Using DevServer URL` grep, or exercises the binding bridge. The `Using DevServer URL` readiness path is therefore untested.
- Churn: 2 commits in 90 days touch `record.sh` + `app-drive.mjs`; 4 touch `desktop/frontend/src/api.js` (`git log --oneline --since=90.days`).

### Statements naming Wails, WebKitGTK, NSIS or build outputs in the four docs
**Answers:** 8

`README.md` has no mention of "Wails", `wails dev`, WebKitGTK or NSIS (grep). It names only user-facing outputs:
- `README.md:57-63`: desktop install table: `kirokit-desktop_<version>_darwin_<arch>.dmg`, cask `kirokit-desktop`, `.deb`/`.rpm` (Ubuntu 22.04+, Debian 12+, Fedora 40+), `kirokit-desktop_<version>_windows_<arch>-setup.exe`.
- `README.md:68`: SmartScreen/Setup program note; `README.md:69`: DMG quarantine note. These describe artifacts, not Wails tooling.

`docs/desktop-app.md`:
- `docs/desktop-app.md:7`: "a desktop app (Wails)".
- `docs/desktop-app.md:9-11`: release artifacts: macOS universal `.dmg` (arm64 and amd64 holding the same universal app) and cask; Linux `.deb`/`.rpm` installing `/usr/bin/kirokit-desktop` and `/usr/lib/kirokit-desktop/kirokit`, "depend on GTK 3 and WebKitGTK 4.1 (Ubuntu 22.04+, Debian 12+, Fedora 40+)"; Windows `-setup.exe` installing `kirokit-desktop.exe` with `bin\kirokit.exe`, "uses the WebView2 runtime Windows 11 ships with (the setup instal..." (line truncated in my read; the full sentence needs re-reading by the Plan).
- `docs/desktop-app.md:13`: "`task desktop:build` (on Linux it needs `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`; on Windows `makensis` from NSIS and Git Bash), or run it from source with `task desktop:dev` (`task desktop:test` runs its checks)".

`CONTRIBUTING.md`:
- `CONTRIBUTING.md:63`: install `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` on Linux; NSIS (`makensis`) on Windows, run tasks from Git Bash.
- `CONTRIBUTING.md:65`: the Linux container mirrors CI's runner "(Go, Node, Task, sops, WebKitGTK 4.1, NSIS)"; `env GOOS=windows task desktop:build` "builds the Windows setup programs".
- `CONTRIBUTING.md:74`: layout table, "the Wails desktop app (its own Go module, with a vite + xterm.js frontend)".
- `CONTRIBUTING.md:98`: Desktop workflow runs `task desktop:test`, `task desktop:build`, a smoke start (`KIROKIT_DESKTOP_SMOKE=1`) on all three OSes, and release snapshots with the GoReleaser Pro key.
- `CONTRIBUTING.md:104-113`: showcase section: ports 34115, 5173 and 34117 must be free ("no `task desktop:dev` or kirocrew gateway running").

`AGENTS.md`:
- `AGENTS.md:8`: "the Wails desktop app: its own Go module".
- `AGENTS.md:10`: `scripts/linux-docker.sh` "also builds the Windows setup programs with `env GOOS=windows task desktop:build`".
- `AGENTS.md:29`: CI runs CLI and desktop tests on three OSes (indirect).
- No `go tool wails` or `wails dev` text in `AGENTS.md`, `README.md`, `docs/desktop-app.md` or `CONTRIBUTING.md`; `go tool wails` appears only in `Taskfile.yml` (`desktop:dev`, lines 39-43) and `scripts/showcase/record.sh:152`. The "wails dev" words appear in `record.sh:68` and the test regex at `showcase.test.mjs:622`.

Related non-doc statements the same change would meet: `desktop/go.mod:8` (`wails/v2 v2.17.0`), `desktop/go.mod:113` (`tool github.com/wailsapp/wails/v2/cmd/wails`), `desktop/wails.json` (v2 schema, `$schema` `config.v2.json`), `desktop/main.go:20-41` (v2 `options.App`, `Bind: []any{app}`, `OnStartup`), `desktop/app.go:11,78`.

## How it is tested today

- Recorder: `node --test scripts/showcase/` from the repository root (about 40 s per `CONTRIBUTING.md:113`); one file: `node --test scripts/showcase/showcase.test.mjs`. Tests use scratch dirs, stub `PATH` binaries, and skip without tmux/ffmpeg/Chrome/Kiro IDE or when ports 34115/5173 are in use. A real recording is only `task showcase` (macOS, `Taskfile.yml:51-54`).
- Desktop frontend: `cd desktop/frontend && node --test "src/**/*.test.js"` (`Taskfile.yml:36`); suite `task desktop:test`. Go: `cd desktop && go test ./...`; one file's tests via `go test -run <Name> ./...` (not verified by running).
- I did not execute any of these commands; all findings are from reading source.

## Open questions

Question 7: Whether Wails v3's dev mode (`wails3 dev`) exposes the same HTTP port, the `Using DevServer URL` log line, and a browser-reachable bindings/events bridge is not answered here (it needs the v3 documentation). The v2 behaviour is inferred, not read from Wails source.
Question 7: The default vite port 5173 comes from `desktop/frontend` vite defaults; I did not read `vite.config.js` fully.
Question 8: The tail of `docs/desktop-app.md:11` (Windows WebView2 sentence) was truncated in my read.

## Code references

- `scripts/showcase/record.sh:59-70,84-89,146-170,256`
- `scripts/showcase/app-drive.mjs:23,34-41,66-122`
- `scripts/showcase/showcase.test.mjs:386-406,572,590,608-624`
- `scripts/showcase/crew-drive.mjs:17,31`
- `desktop/frontend/src/api.js:1-8,96-104`; `desktop/frontend/src/main.js:289-300`
- `desktop/main.go:20-41`; `desktop/app.go:73-86`; `desktop/shell.go:36-58`; `desktop/wails.json`; `desktop/go.mod:8,113`
- `Taskfile.yml:36,39-54`
- `README.md:57-69`; `docs/desktop-app.md:7-13`; `CONTRIBUTING.md:63-65,74,98,104-113`; `AGENTS.md:8,10,29`
- Exhaustive for the four named doc files (grep for Wails/WebKit/NSIS/desktop/build outputs); not exhaustive for `.github/workflows`, `.goreleaser.yaml` or `scripts/desktop-*.sh`, which belong to other questions.
