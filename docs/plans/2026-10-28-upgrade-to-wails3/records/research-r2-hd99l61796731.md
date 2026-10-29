---
type: research
---

# Desktop build, packaging and installed identity today (Wails v2)

## Findings

### One script builds per OS, and each consumer reads a fixed path under desktop/build/bin/
**Answers:** 4

`scripts/desktop-build.sh` switches on `go env GOOS` (`scripts/desktop-build.sh:34`). `task desktop:build` only calls `bash scripts/desktop-build.sh` (`Taskfile.yml:45-49`). It derives the version from `$VERSION` or `git describe`, strips a leading `v` (`scripts/desktop-build.sh:23-24`), and injects it into the CLI only via ldflags `-X ...commands.version/commit/date` (`scripts/desktop-build.sh:28-29`). The Wails app itself gets no version flag; its product version stays `0.0.0` from `desktop/wails.json:12`.

```
desktop/build/bin/
  macOS:   kirokit.app/Contents/MacOS/kirokit            (Wails, universal)
           kirokit.app/Contents/Resources/bin/kirokit   (CLI, lipo universal)
  Linux:   kirokit-desktop  +  kirokit                  (CLI beside it)
  Windows: cli/windows_<arch>/kirokit.exe               (staged CLI)
           kirokit-desktop-<arch>.exe                   (Wails app)
           kirokit-<arch>-installer.exe  -> mv ->  installers/kirokit-desktop_<v>_windows_<arch>-setup.exe
```

- macOS: builds the CLI for amd64 and arm64 with `CGO_ENABLED=0`, joins them with `lipo` (`scripts/desktop-build.sh:36-40`). Exports `-mmacosx-version-min=13.0` in CGO flags (`:43-44`). Runs `go tool wails build -platform darwin/universal -clean` (`:47`). The Wails convention supplies the output `kirokit.app` (from `outputfilename` in `desktop/wails.json:4`). `-clean` empties `build/bin`, so the CLI is copied into `Contents/Resources/bin/kirokit` afterwards (`:46-51`). Then `codesign --force --deep --sign -` re-seals ad hoc (`:53`).
- Linux: builds the CLI for `$GOARCH`, then `wails build -platform linux/$arch -tags webkit2_41 -clean -o kirokit-desktop` (`:62`). Installs the CLI as `desktop/build/bin/kirokit` (`:63`). Wails v2 conventions in play: `-platform`, `-tags webkit2_41` (WebKitGTK 4.1 ABI), `-o`.
- Windows: `rm -rf desktop/build/bin`, then per arch (default `amd64 arm64`, `:71`) stages the CLI at `bin/cli/windows_<arch>/kirokit.exe` (`:73-74`). Runs `wails build -platform windows/$arch -nsis -o kirokit-desktop-$arch.exe` without `-clean` and with GOOS/GOARCH unset for the tool (`:77`). Renames Wails' `kirokit-<arch>-installer.exe` to `installers/kirokit-desktop_<v>_windows_<arch>-setup.exe` (`:78`). The installer name comes from the template `OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"` (`desktop/build/windows/installer/project.nsi:78`), which is Wails v2's NSIS convention. The Windows build needs `makensis` on PATH (`scripts/desktop-build.sh:11`).

### Consumers of those paths
**Answers:** 4

- `scripts/desktop-prebuilt.sh` (staging for GoReleaser prebuilt builders), under `desktop/build/bin/prebuilt/`:
  - darwin: reads `kirokit.app/Contents/MacOS/kirokit`, `Contents/Info.plist`, `Contents/Resources/iconfile.icns`, `Contents/Resources/bin/kirokit` (`scripts/desktop-prebuilt.sh:34,42-44`). Writes `kirokit_darwin_{amd64,arm64}/kirokit`, removes the signature and re-signs ad hoc (`:32-40`), plus a `bundle/` folder with the plist, icon and CLI (`:30,42-44`).
  - linux: requires `desktop/build/bin/{kirokit-desktop,kirokit}` (`:49-51`). Writes `kirokit-desktop_linux_<GOARCH>/{kirokit-desktop,kirokit}` (`:52-56`).
  - windows: no-op (`:59-61`).
- `.goreleaser.yaml`:
  - Builds: `kirokit` (CLI, `CGO_ENABLED=0`, 3 OS x 2 arch, `:16-32`). `kirokit-desktop-bin` prebuilt, darwin amd64/arm64, path `desktop/build/bin/prebuilt/kirokit_{{ .Os }}_{{ .Arch }}/kirokit` (`:37-46`). `kirokit-desktop-linux-amd64` and `-arm64` prebuilt (`:52-71`). The arm64 build is skipped when `KIROKIT_SKIP_LINUX_ARM64_APP` is set (`:65`).
  - `app_bundles` wraps the binary as `kirokit.app` with bundle id `io.github.mavogel.kirokit`. `extra_files` copy `prebuilt/bundle/Contents/{Info.plist, Resources/iconfile.icns, Resources/bin/kirokit}` (`:75-87`).
  - `dmg` uses `appbundle`, named `kirokit-desktop_{{ .Version }}_darwin_{{ .Arch }}` (`:89-95`).
  - Archives: tar.gz, zip for windows, CLI only (`:97-107`). nfpm `kirokit` (CLI `.deb`/`.rpm`, `/usr/bin`, `:112-126`). nfpm `kirokit-desktop` (`:128-161`): the app at `/usr/bin/kirokit-desktop`, its CLI copy from `prebuilt/kirokit-desktop_linux_{{ .Arch }}/kirokit` to `/usr/lib/kirokit-desktop/kirokit`, `desktop/build/linux/kirokit-desktop.desktop` and `.png` to the menu and `hicolor/512x512` (`:143-151`). Deps are `libgtk-3-0` + `libwebkit2gtk-4.1-0` (deb), `gtk3` + `webkit2gtk4.1` (rpm) (`:152-161`).
  - Casks: `kirokit-desktop` with `app: kirokit.app` and a quarantine-clearing postflight (`:220-235`). Release `extra_files` glob `./desktop/build/bin/installers/*-setup.exe` (`:244-245`).
  - Windows setup programs are not GoReleaser artifacts; they are release extras.
- `.github/workflows/release.yaml`: jobs `version`, `app-linux-arm64` (runner `ubuntu-26.04-arm`, `:74-77`), a `split` matrix over macos/ubuntu/windows (`:117-130`), and `merge`. Each build job runs `task desktop:build` then `scripts/desktop-prebuilt.sh` (`:102-107`, `:169-172`). The arm64 Linux staging folder is uploaded and downloaded into `desktop/build/bin/prebuilt/kirokit-desktop_linux_arm64` (`:113`, `:174-179`). GoReleaser Pro runs `release --clean --split` (`:191-196`). Windows uploads `desktop/build/bin/installers/` as artifact `installers-<run>` (`:219-226`), which `merge` restores to the same path (`:256-260`) before `continue --merge` (`:269-275`). NSIS is installed with choco (`:161-165`). Linux installs `libgtk-3-dev libwebkit2gtk-4.1-dev` (`:155-159`).
- `.github/workflows/desktop.yaml`: per OS runs `task desktop:test`, `task desktop:build` (`:71-75`). Smoke start with `KIROKIT_DESKTOP_SMOKE=1` of `desktop/build/bin/kirokit.app/Contents/MacOS/kirokit`, `desktop/build/bin/kirokit-desktop` (under xvfb) or `desktop/build/bin/kirokit-desktop-$(go env GOARCH).exe` (`:78-87`). Windows' smoke exe is the pre-NSIS `-o kirokit-desktop-<arch>.exe` file, so it depends on the `-o` naming in `scripts/desktop-build.sh:77`. On PRs with `GORELEASER_KEY` it stages and runs a snapshot, and on Linux installs the built `.deb`s and checks `/usr/lib/kirokit-desktop/kirokit` and `/usr/bin/kirokit-desktop` (`:92-125`).
- `Taskfile.yml`: `desktop:test` (frontend `npm ci && npm run build`, `node --test`, `go vet`/`go test` in `desktop`, `:32-37`), `desktop:dev` (builds the CLI to `desktop/build/dev-bin/kirokit`, then `cd desktop && go tool wails dev`, `:39-43`), `desktop:build` (`:45-49`).
- `scripts/showcase/record.sh` builds the CLI to `desktop/build/dev-bin/kirokit` (`:85`) and runs `go tool wails dev` in `desktop` (`:141-148`), waiting for the log line `Using DevServer URL: http://localhost:34115` (`:152-157`). That is Wails v2 dev-server output and port.

### Which names are Wails v2 conventions versus project-chosen
**Answers:** 4

| Name / flag | Origin |
|---|---|
| `go tool wails build`, `-platform darwin/universal`, `linux/<arch>`, `windows/<arch>`, `-clean`, `-nsis`, `-o` | Wails v2 CLI (tool pinned `desktop/go.mod:113`, `wails/v2 v2.17.0` at `:8`) |
| `-tags webkit2_41` | Wails v2 build tag for WebKitGTK 4.1 |
| `kirokit.app` | `outputfilename: kirokit` in `desktop/wails.json:4`, bundle template in `desktop/build/darwin/` |
| `<project>-<arch>-installer.exe` in `build/bin` | Wails NSIS template (`project.nsi:78`) |
| `build/bin/`, `build/darwin/`, `build/windows/`, `build/linux/` | Wails v2 build directory layout |
| `iconfile.icns`, `Contents/Info.plist` | Wails v2 macOS bundle (read by `scripts/desktop-prebuilt.sh:42-43`) |
| `kirokit-desktop` binary, `bin/cli/windows_<arch>/`, `installers/`, `prebuilt/`, `dev-bin/`, `*-setup.exe` | project-chosen |

### The v2 assets customise identity, the CLI placement and the Windows installer
**Answers:** 5

- macOS: `desktop/build/darwin/Info.plist` hard-codes `CFBundleIdentifier` `io.github.mavogel.kirokit` (`:10-11`), minimum macOS 13.0 (`:20-21`), icon `iconfile` (`:18-19`), and fills name/version/copyright from Wails templates (`{{.Info.ProductName}}`, `{{.OutputFilename}}`, `:6-17`). `Info.dev.plist` is identical plus `NSAppTransportSecurity`/`NSAllowsLocalNetworking` (`:26-30`) for the dev server. GoReleaser sets the same bundle id (`.goreleaser.yaml:80`) but takes the plist from the built app (`scripts/desktop-prebuilt.sh:42`).
- `desktop/wails.json`: name `kirokit`, `outputfilename` `kirokit`, company `mavogel`, product name `kirokit`, version `0.0.0`, copyright "Copyright mavogel", frontend commands `npm install`/`npm run build`/`npm run dev` (`:3-15`).
- Windows `project.nsi` overrides only `PRODUCT_EXECUTABLE` as `kirokit-desktop.exe` (`:35`), so the app does not collide with the `kirokit` CLI on PATH (`:33-34`). Defaults from `wails_tools.nsh`: project `kirokit`, company `mavogel`, product `kirokit` (`wails_tools.nsh:7-14`). Uninstall key name is `${COMPANY}${PRODUCT}` = `mavogelkirokit` (`:25-26`), so the registry key is `Software\Microsoft\Windows\CurrentVersion\Uninstall\mavogelkirokit` (`:28`). `project.nsi` does not define `WAILS_INSTALL_SCOPE`, so the `!else` branch writes it under HKLM with `RequestExecutionLevel admin` (`wails_tools.nsh:134-140`, `:30-34`); the registry view is 64-bit (`:117`). Install dir with no scope defined is `$PROGRAMFILES64\mavogel\kirokit` (`project.nsi:86`). Shortcuts `$SMPROGRAMS\kirokit.lnk` and `$DESKTOP\kirokit.lnk` point to `kirokit-desktop.exe` (`:109-110`). Uninstall removes `$AppData\kirokit-desktop.exe` (WebView2 data) and `$INSTDIR` (`:121-123`). The installer bootstraps WebView2 (`wails.webview2runtime`, `:97`).
- `windows/info.json` and `wails.exe.manifest` are Wails templates: version resource (company, product name, comments, copyright) and assembly identity `com.wails.{{.Name}}` plus Common-Controls 6 and per-monitor DPI (`manifest:3,6,11-12`).
- Linux `kirokit-desktop.desktop`: `Name=kirokit`, `Exec=kirokit-desktop`, `Icon=kirokit-desktop`, `StartupWMClass=kirokit`, `Categories=Development;` (`:1-9`). Installed by nfpm only (`.goreleaser.yaml:148-151`).

### The bundled CLI is placed per OS and found at run time by a fixed search order
**Answers:** 5

Placement:
- macOS: `Contents/Resources/bin/kirokit` (`scripts/desktop-build.sh:50-51`, `.goreleaser.yaml:86-87`).
- Windows: `$INSTDIR\bin\kirokit.exe` from `File "/oname=kirokit.exe" "..\..\bin\cli\windows_${ARCH}\kirokit.exe"` (`project.nsi:105-106`).
- Linux: `/usr/lib/kirokit-desktop/kirokit` (`.goreleaser.yaml:144-145`).

Run time: `BundledCLI` in `desktop/shell.go:38-42` calls `resolveCLIFor` (`:44-75`). Order:
1. `$KIROKIT_DESKTOP_CLI` returned unchecked (`:45-47`).
2. Under `wails dev`, i.e. when the executable is under `<cwd>/build/bin/`, `<cwd>/build/dev-bin/kirokit(.exe)` goes first (`:57-62`).
3. `installedCLIFor(goos, filepath.Dir(exe))` (`desktop/platform.go:179-188`): darwin `exeDir/../Resources/bin/kirokit`, windows `exeDir/bin/kirokit.exe`, other `exeDir/../lib/kirokit-desktop/kirokit`.
4. `<cwd>/build/dev-bin/kirokit(.exe)` (`shell.go:66-68`).
The first existing non-directory file wins (`:69-73`); otherwise it errors with a per-OS hint (`:74`, `platform.go:191-199`). `desktop/runner.go:108` puts the CLI's folder first on PATH for terminals. Note the macOS and Windows paths depend on the Wails v2 executable location (`Contents/MacOS/<outputfilename>`, install folder next to `kirokit-desktop.exe`).

### Churn and hard-to-change points
**Answers:** 4, 5

- Fact: `git log --since=90.days` shows 2 commits touching `desktop/build` and 13 touching the build script, `.goreleaser.yaml` and `release.yaml`.
- The coupling is by literal path strings across `scripts/desktop-build.sh`, `scripts/desktop-prebuilt.sh`, `.goreleaser.yaml`, `desktop.yaml` and `release.yaml` (all listed above), plus `desktop/platform.go:181-186`. `desktop/shell_test.go:201-246` hard-codes the same layouts.
- A reply to the batch question could not be recorded (`question reply` refused: "read the current answer before recording what you did"). The poll answer ("also add v3-only features") concerns scope, not these questions, and is not used here.

## How it is tested today

- Go tests for CLI resolution and per-OS paths: `desktop/shell_test.go` (cases at `:175-246`, covering darwin/linux/windows and dev-bin precedence), `desktop/runner_test.go`. One file or test: `cd desktop && go test -run <Name> ./...`. Suite: `task desktop:test` (`Taskfile.yml:32-37`) from the repository root, or `cd desktop && go vet ./... && go test ./...` after building the frontend.
- Frontend: `cd desktop/frontend && node --test "src/**/*.test.js"`.
- Build and packaging have no unit tests: they are exercised in CI by `.github/workflows/desktop.yaml` (build, smoke start with `KIROKIT_DESKTOP_SMOKE=1`, GoReleaser snapshot, `.deb` install on Linux) on Ubuntu, Windows and macOS. `scripts/linux-docker.sh <command>` reproduces the Linux runner locally (AGENTS.md).
- Not run by me: no build or test was executed in this research; everything is from reading source.

## Open questions

None.

## Code references

- `scripts/desktop-build.sh:18-88`, `scripts/desktop-prebuilt.sh:19-69`, `Taskfile.yml:32-54`
- `.goreleaser.yaml:15-161,200-245`
- `.github/workflows/release.yaml:29-307`, `.github/workflows/desktop.yaml:21-125`
- `desktop/wails.json:1-20`, `desktop/go.mod:8,113`
- `desktop/build/darwin/Info.plist:10-21`, `desktop/build/darwin/Info.dev.plist:26-30`
- `desktop/build/windows/installer/project.nsi:35,78,86,101-123`, `desktop/build/windows/installer/wails_tools.nsh:7-34,114-140`
- `desktop/build/windows/info.json`, `desktop/build/windows/wails.exe.manifest:3-12`, `desktop/build/linux/kirokit-desktop.desktop:1-9`
- `desktop/shell.go:32-75`, `desktop/platform.go:177-199`, `desktop/runner.go:108`, `desktop/shell_test.go:175-246`
- `scripts/showcase/record.sh:85,141-157`
- Exhaustiveness: the files named in the question were all read in full; `ci.yaml` and `app-drive.mjs` were not opened, and docs under `docs/` were not audited.
