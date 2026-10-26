---
type: research
---

# Desktop app: environment capture, paths, shell tab, Wails build needs and platform wording

## Findings

### The login shell is run as `$SHELL -ilc`, and the capture is a NUL-separated `env -0` after a marker
**Answers:** 2

The app starts a capture in a goroutine at startup (`desktop/app.go:74`) and again on every `Prerequisites` call (`desktop/prereq.go:22` via `refreshEnv`, `desktop/app.go:102-108`). `NewAppAt` wires `captureFn` to `Capture(os.Getenv("SHELL"), 5s)` (`desktop/app.go:62`, `desktop/envcapture.go:19`).

```
Capture(shell, timeout)               desktop/envcapture.go:75
 └ runLoginShell                      :93
    ├ shell == "" -> "/bin/zsh"       :94-96
    ├ exec <shell> -ilc "printf '%s' '__KIROKIT_ENV_BEGIN__'; exec env -0"   :100
    ├ Setpgid + Cancel = kill(-pid, SIGKILL), WaitDelay 1s                   :103-105
    ├ bytes.Cut at marker; split on 0x00; Cut each at first "="              :115-124
    └ on failure/timeout -> nil + Note
 fallback: os.Environ() with PATH="/usr/bin:/bin:/usr/sbin:/sbin"            :77-84
 always: PATH = withToolFolders(PATH, HOME)                                  :89
```

- Fact: output before the marker (startup-file chatter) is discarded (`desktop/envcapture.go:115`; test `desktop/envcapture_test.go:24`). Stdin is not set and stderr is not captured (`desktop/envcapture.go:106-108`).
- Fact: the shell is chosen only from `$SHELL` of the app process, with `/bin/zsh` as the default (`desktop/envcapture.go:94-96`). The `-ilc` flag set is POSIX-shell style. `exec env -0` needs an `env` that supports `-0`, a GNU/BSD feature; the macOS `env` has it. The command string is passed to any shell, so fish or nushell would receive a POSIX-syntax string (`printf ...; exec env -0`) — inference, untested.
- macOS assumptions: `syscall.Setpgid` and `syscall.Kill(-pid, SIGKILL)` (`desktop/envcapture.go:10,103-104`) are Unix-only. `desktop/terminal/terminal.go:11,63,87` also uses `syscall.EIO`/`Kill`, and `github.com/creack/pty` is the pty layer (`desktop/terminal/terminal.go:13,33`). No build tags or `_windows.go` files exist in `desktop/` (`find` listing shows only `desktop/*.go` and `desktop/terminal/terminal.go`), so `desktop/` as written cannot compile for GOOS=windows (inference from the Unix-only `syscall` members; not compiled).
- `toolFolders` hardcodes `/opt/homebrew/bin`, `/opt/homebrew/sbin`, `/usr/local/bin` and `~/.local/bin` and appends any missing to `PATH` using `":"` as separator (`desktop/envcapture.go:47-68`). `Lookup` uses `filepath.SplitList` but never adds `.exe`/`PATHEXT` (`desktop/envcapture.go:33-44`); it accepts any non-directory with an execute bit (`:39`), a Unix permission test.

### Tool lookup and age-key path: PATH search, three tools, and a macOS default key location
**Answers:** 2

`Prerequisites` returns three `Check`s: SOPS (`sops`), "Decryption key", Kiro CLI (`kiro-cli`) (`desktop/prereq.go:25-31`). Hints are macOS/Homebrew wording: `"Install it with: brew install sops"` (`desktop/prereq.go:25`).

`ageKeyCandidate` order (`desktop/prereq.go:44-56`): `SOPS_AGE_KEY_FILE`, else `$XDG_CONFIG_HOME/sops/age/keys.txt`, else `$HOME/Library/Application Support/sops/age/keys.txt` (the macOS default; `HOME` falls back to `os.UserHomeDir()` via `homeDir`, `desktop/app.go:137-143`). The existence check is `os.Stat` non-directory (`desktop/prereq.go:58-62`). There is no `runtime.GOOS` branch anywhere in `desktop/*.go` (grep for `runtime.GOOS` found none), so Linux without `XDG_CONFIG_HOME` is pointed at the macOS path, and Windows is not represented. SOPS itself reads `SOPS_AGE_KEY_FILE`, else `os.UserConfigDir()` + `sops/age/keys.txt`, handling `XDG_CONFIG_HOME` by hand on macOS (sops `age/keysource.go`, `SopsAgeKeyUserConfigPath`, fetched 2026-10-25), so its defaults are `~/.config/sops/age/keys.txt` on Linux and `%AppData%\sops\age\keys.txt` on Windows (Go `os.UserConfigDir` semantics). Test `desktop/prereq_test.go:62` pins the order and the mac path (`:65`).

### The data folder is a hardcoded macOS Application Support path; the CLI is found only inside a `.app` bundle, the dev-bin folder, or by override
**Answers:** 3

- Data folder: `NewApp` = `<home>/Library/Application Support/kirokit-desktop` (`desktop/app.go:54-56`). Under it: `settings.json` (`desktop/settings.go:26`), `repo/` default clone (`desktop/app.go:126`), `zdotdir/` wrapper (`desktop/shell.go:198`). `NewAppAt(dir)` accepts any folder (tests use it, `desktop/app.go:59`). `homeDir()` returns `"/"` if `UserHomeDir` fails (`desktop/app.go:137-143`).
- `repoDir()`: saved `Settings.RepoDir` else the default (`desktop/app.go:130-135`). `kiroHome()`: saved `Settings.KiroHome`, else captured `KIRO_HOME`, else `<home>/.kiro` (`desktop/settings.go:117-137`). `State()` stats `<repoDir>/.git` for `Cloned` (`desktop/app.go:113`).
- Path validators use `filepath.IsAbs`, `clean == "/"` and the home folder (`desktop/settings.go:139-166`); `"/"` as the root test is Unix-only. `repoURLPattern` accepts `/\S+` absolute paths and not Windows drive paths (`desktop/settings.go:24`). Examples in messages say `~/.kiro`-style (`desktop/settings.go:141,155`).
- `installs.go` has no path logic of its own: the poller calls `install.Installed(a.kiroHome)` every second (`desktop/installs.go:32,54,89`).
- `resolveCLI` order (`desktop/shell.go:41-69`): (1) `$KIROKIT_DESKTOP_CLI` returned as is, not stat'ed (`:42-44`); (2) if the executable lies under `<cwd>/build/bin/`, `<cwd>/build/dev-bin/kirokit` first (`:53-55`); (3) `<dir(exe)>/../Resources/bin/kirokit` — the macOS bundle layout `kirokit.app/Contents/MacOS/kirokit` -> `Contents/Resources/bin/kirokit` (`:56-59`); (4) `<cwd>/build/dev-bin/kirokit` (`:60-62`). A candidate must exist and not be a directory (`:63-66`). The error text says "start the app from kirokit.app" (`:68`).
- Consequences today: the literal name is `kirokit` with no `.exe` (`:48,58`); a Linux or Windows non-bundle layout (binary beside `kirokit`, or `bin/` next to the exe) would not be found without `KIROKIT_DESKTOP_CLI` or a working directory containing `build/dev-bin/kirokit`. `desktop/build/dev-bin/kirokit` is built by `task desktop:dev` (`Taskfile.yml:42`). Test `TestBundledCLI_FixedOrder` pins the order with `kirokit.app/Contents/...` paths (`desktop/shell_test.go:170-194`).
- Callers: `runner.go` (`desktop/runner.go:46,108`), `clone.go` (`desktop/clone.go:30`), `shell.go:192`. A missing CLI makes `runner`/`clone` fail, but the shell tab only skips `KIROKIT_BIN` (`desktop/shell.go:194-196`).
- Terminal PATH is built as `dir(cli) + ":" + PATH` (`desktop/shell.go:122`): the `:` separator is Unix-only.

### The shell tab runs `$SHELL -il` on a pty; zsh gets a ZDOTDIR wrapper, other shells get only environment variables
**Answers:** 4

```
frontend tab "zsh" (index.html:88)  -> terminal.js onData -> api.shellWrite -> App.ShellWrite   (shell.go:219)
                                       onResize -> api.shellResize -> App.ShellResize           (shell.go:233)
ShellResize: size 1..1000 only; first sizing starts the shell (startShellLocked, shell.go:252)
startShellLocked -> shellCommand (shell.go:190) -> terminal.Start(creack/pty) -> stream -> event "term:shell" (base64)
shell exit -> prints "shell ended (exit N). Type to start a new one." ; typing restarts (shell.go:224,274)
```

- Shell choice: captured `SHELL`, else the app's `SHELL`, else `/bin/zsh` (`desktop/shell.go:73-86`). Command: `exec.Command(shell, "-il")`, `Dir` = repoDir else home (`:209-214`).
- Environment for every terminal: the whole login env + `TERM=xterm-256color`, `COLORTERM=truecolor`, `KIROKIT_REPO`, `KIRO_HOME`, bundled-CLI folder first on `PATH` (`desktop/shell.go:112-133`).
- zsh only (`isZsh` = base name `zsh`, `:88`): `writeWrapper` writes `.zshenv`, `.zprofile`, `.zshrc`, `.zlogin` into `<data>/zdotdir` (`:139-144,166-176`). Each sources the user's file of the same name from `KIROKIT_USER_ZDOTDIR` or `$HOME`, then restores `ZDOTDIR` (`:155-163`). `.zshrc` and `.zlogin` then re-prepend `$KIROKIT_BIN` to `path` and re-export `KIROKIT_REPO` and `KIRO_HOME` (`:146-148`); `.zlogin` also restores the user's `ZDOTDIR` (`:152`). Env vars set: `ZDOTDIR`, `KIROKIT_APP_REPO`, `KIROKIT_APP_KIRO_HOME`, `KIROKIT_USER_ZDOTDIR` (when the login env has one) (`:202-207`), plus `KIROKIT_BIN` (`:195`). The wrapper uses zsh syntax (`path=(...)`, `${path:#...}`).
- Non-zsh shells: nothing is injected beyond the env vars above; the person's startup files may reorder `PATH` or reset `KIRO_HOME`. `shellNote` returns "Your shell is <name>, so the bundled kirokit may not come first on its PATH..." shown in Settings via `State().Env.Note` (`desktop/shell.go:91-96`, `desktop/app.go:117`).
- Where "zsh" is named in user-visible text: tab label `zsh` (`desktop/frontend/index.html:88`); dock hint "Type your own commands in the zsh tab." (`desktop/frontend/src/terminal.js:60`); comments at `terminal.js:2`, `api.js:52`; README (`README.md:288`). The internal tab id is `shell` (`desktop/frontend/src/tabs.js:5`), and Go names the host `shellHost`; the "zsh" naming is in the label and prose only. Error text for failed start: "could not start your shell" (`desktop/shell.go:260`).
- Pty layer is `creack/pty` + `syscall` (`desktop/terminal/terminal.go:13,63,87`), Unix only.
- Dependency note: `desktop/go.mod` has `creack/pty v1.1.24` and `wails/v2 v2.17.0` (`desktop/go.mod:6,8`); `go-webview2` is already an indirect dependency (`desktop/go.mod:89`).

### Wails v2.17 builds on Linux and Windows but needs per-platform packages, tags, and assets that the repository lacks
**Answers:** 9

Repository today: `desktop/wails.json` (name, `outputfilename: kirokit`, npm commands, `info` block; `desktop/wails.json:1-20`), `desktop/build/appicon.png`, `desktop/build/darwin/Info.plist` and `Info.dev.plist`, nothing under `desktop/build/windows/` or `desktop/build/linux/` (`find desktop/build -type f` shows only appicon.png and `darwin/*`, besides ignored output under `bin/` and `dev-bin/`). `desktop/main.go` has no `Windows`/`Linux`/`Mac` options (`desktop/main.go:18-31`). Only macOS is built in CI and release: `runs-on: macos-latest` (`.github/workflows/desktop.yaml:23`, `.github/workflows/release.yaml:27`); `scripts/desktop-build.sh:34` passes `-platform darwin/universal`.

From Wails docs (cited; not run here):
- Platforms: Windows 10/11 AMD64/ARM64, Linux AMD64/ARM64; `-platform` takes `windows/amd64`, `windows/arm64`, `linux/amd64`, `linux/arm64` (https://wails.io/docs/gettingstarted/installation, https://wails.io/docs/reference/cli). Go 1.21+ and npm required.
- Linux: build needs GTK3 and WebKit2GTK dev packages, gcc, pkg-config, npm (apt package list shown on https://wails.io/docs/guides/linux-distro-support). Runtime: Debian 12/Ubuntu 22.04+ `libgtk-3-0 libwebkit2gtk-4.1-0`; ABI 4.1 needs `-tags webkit2_41`, ABI 4.0 (Debian 11, Ubuntu 20.04, RHEL 8-9) needs `-tags webkit2_40`. `wails doctor` reports missing packages.
- Windows: runtime needs the WebView2 runtime (preinstalled on Windows 11); `wails build -webview2 download|embed|browser|error` (default `download`) (https://wails.io/docs/guides/windows). `-nsis` builds an NSIS installer into `build/bin`, needs `makensis` on PATH, and reads `build/windows/installer` generated at `wails init` (https://wails.io/docs/guides/windows-installer); this project did not generate those files. Docs also note spawned child programs show a console window unless `SysProcAttr{HideWindow, CreationFlags: 0x08000000}` is set. The `-windowsconsole` flag keeps a console.
- Icons/manifest: Wails projects normally keep `build/windows/icon.ico`, `wails.exe.manifest` and `info.json` for Windows; this is not on the pages fetched, so it is a gap/inference, not confirmed here.
- arm64: `-platform linux/arm64` or `windows/arm64`. Cross-compiling Wails is not supported by the pages read; the docs' cross-platform guide builds each OS on its own runner (matrix `ubuntu-latest`, `windows-latest`, `macos-latest`) (https://wails.io/docs/guides/crossplatform-build). cgo is needed on Linux (GTK/WebKit); `scripts/desktop-build.sh` sets `CGO_CFLAGS`/`CGO_LDFLAGS` with `-mmacosx-version-min=13.0` (`scripts/desktop-build.sh:30-31`), which is macOS-only. No existing `webkit2_41` tag is passed anywhere (grep of `scripts/`, `Taskfile.yml`, workflows: none).
- Separately, the CLI target list in `.goreleaser.yaml` is `darwin` and `linux`, `amd64`/`arm64` (`.goreleaser.yaml:12-17`), no windows; the desktop builder is `prebuilt` for `darwin` only (`.goreleaser.yaml:30-40`); `app_bundles`, `dmg`, and `homebrew_casks` with `xattr` follow (`:44-64,98-134`).

### Docs, release config, and showcase scripts call the app macOS-only in many places
**Answers:** 10

| Where | What it says |
|---|---|
| `README.md:17,20-21,244` | Homebrew install of the CLI (`brew install --cask mavogel/tap/kirokit`); `brew install sops` (`:17`) |
| `README.md:64,67` | "The macOS app bundles `kirokit`"; `brew install --cask mavogel/tap/kirokit-desktop` |
| `README.md:271` | sync-setup registers launchd (macOS) or systemd user timer (Linux): the only place naming Linux |
| `README.md:283` | "a macOS app (Wails)"; DMG names `darwin_arm64`/`darwin_amd64`; macOS 13+; unsigned, cask clears quarantine, manual `xattr -dr com.apple.quarantine kirokit.app` |
| `README.md:285` | default repo folder `~/Library/Application Support/kirokit-desktop/repo` |
| `README.md:288` | the "zsh" tab |
| `CONTRIBUTING.md:55` | `brew install go go-task golangci-lint lefthook uv gitleaks commitizen node` |
| `CONTRIBUTING.md:99-108` | showcase "on a Mac"; `brew install tmux asciinema agg ffmpeg node`; needs Google Chrome in `/Applications`, SOPS age key at its default location |
| `scripts/showcase/record.sh:43,61,62,121` | `Library/Application Support/sops/age/keys.txt`; `/Applications/Google Chrome.app` check |
| `scripts/showcase/showcase.test.mjs:19,287` | Chrome path `/Applications/Google Chrome.app`; caption text includes the brew cask line |
| `scripts/showcase/caption.mjs:37-38` | caption lines `brew install --cask mavogel/tap/kirokit` and `...kirokit-desktop` |
| `Taskfile.yml:46` | `desktop:build` described as "universal, macOS 13 or later" |
| `.goreleaser.yaml:73,102-134` | `xattr` quarantine steps in both casks; `homebrew_casks` is macOS-oriented |
| `scripts/desktop-build.sh`, `scripts/desktop-prebuilt.sh` | `lipo`, `codesign`, `.app` layout; macOS only |
| `desktop/build/darwin/Info.plist:20` | `LSMinimumSystemVersion` 13.0 |

Fact: README line 21 and the context-pointer lines (`README.md:21,67,244,283`) are Homebrew install lines or the macOS app paragraph. Changes in `README.md` happened 25 times in 90 days, the most volatile of these files.

### Churn and hard-to-change spots
**Answers:** 2, 3, 4, 9, 10

`git log --oneline --since=90.days -- <path>` counts: every file in `desktop/*.go`, `desktop/frontend/src`, `desktop/terminal`, `desktop/build`: 1 commit each (the app arrived in one change); `scripts/desktop-build.sh`: 2; `.goreleaser.yaml`: 7; `README.md`: 25. Callers already used by others: `BundledCLI` (runner, clone, shell), `terminalEnv`, `Capture`/`refreshEnv`/`loginEnv`, `homeDir()` (`desktop/app.go:137`, used in `settings.go`, `envcapture.go`, `prereq.go`, `shell.go`). Hardest to change: `terminal.Start/Close` and `Capture` (Unix `syscall`), `.goreleaser.yaml` (7 commits, prebuilt/DMG/cask chain), the zsh wrapper (tested with real zsh, `desktop/shell_test.go:87,116`).

## How it is tested today

- Go (desktop module): `cd desktop && go vet ./... && go test ./...` (`Taskfile.yml:37`). One file: `cd desktop && go test -run TestCapture ./ -count=1` with the file's tests such as `envcapture_test.go` (`TestCapture_ReadsOnlyAfterMarker`, `_SlowShellFallsBack`, `_AppendsToolFolders`; fake shell scripts via temp dirs), `prereq_test.go` (`TestPrerequisites_*`, `TestKiroHome_*`), `shell_test.go` (wrapper and `TestBundledCLI_FixedOrder`, real zsh, skipped when `zsh` is absent, `desktop/shell_test.go:89,118`). Tests use `NewAppAt(t.TempDir())`-style isolation and `writeScript` fixtures (`desktop/shell_test.go:70`), and some assert mac paths (`desktop/prereq_test.go:65`, `desktop/shell_test.go:172-176`, `desktop/envcapture_test.go:55,73`).
- Frontend: `cd desktop/frontend && npm ci && npm run build` then `node --test "src/**/*.test.js"` (`Taskfile.yml:35-36`; `desktop/frontend/package.json` `test` script). One file: `cd desktop/frontend && node --test src/tabs.test.js`.
- Full desktop check from the repository root: `task desktop:test` (`Taskfile.yml:32-37`); CI runs it on `macos-latest` only (`.github/workflows/desktop.yaml:23,46-47`); the main CI is `ubuntu-latest` (`.github/workflows/ci.yaml:13`).
- Showcase scripts: `node --test scripts/showcase/` (`CONTRIBUTING.md:108`). CLI/golden: `go test ./...`. I did not run any of these commands.

## Open questions

- Question 9: Windows icon/manifest file names (`build/windows/icon.ico`, `wails.exe.manifest`, `installer/*`) are not stated on the fetched pages; confirm by running `wails init` for v2.17.0 or reading its template.
- Question 9: whether `webkit2_41` is the default WebKit ABI in v2.17.0 (the docs page says "default for most modern distros" without stating the build default) was not verified.
- Question 2: behaviour of `-ilc ... exec env -0` under fish/nushell/PowerShell was not run; stated as inference only.

## Code references

- `desktop/envcapture.go:16-126` (Capture, runLoginShell, Lookup, toolFolders); `desktop/prereq.go:20-62`; `desktop/app.go:52-143`; `desktop/settings.go:24-166`; `desktop/installs.go:30-103`; `desktop/shell.go:31-307`; `desktop/terminal/terminal.go:11-95`; `desktop/runner.go:46,100-125`; `desktop/clone.go:30`; `desktop/main.go:15-35`; `desktop/wails.json`; `desktop/build/` (only `appicon.png`, `darwin/Info.plist`, `darwin/Info.dev.plist` tracked inputs)
- `desktop/frontend/index.html:88`; `desktop/frontend/src/{tabs.js:5,terminal.js:2,60-73,138,api.js:52}`
- `scripts/desktop-build.sh:23-41`; `scripts/desktop-prebuilt.sh`; `Taskfile.yml:32-48`; `.goreleaser.yaml:12-134`; `.github/workflows/{desktop,release,ci}.yaml`
- `README.md:17-21,64-67,244,271,283-288`; `CONTRIBUTING.md:55,97-108`; `scripts/showcase/{record.sh:43,61-62,121; caption.mjs:37-38; showcase.test.mjs:19,287}`
- Wails docs: https://wails.io/docs/gettingstarted/installation, /reference/cli, /guides/windows, /guides/windows-installer, /guides/linux-distro-support, /guides/crossplatform-build
- Exhaustiveness: all of `desktop/*.go` non-test files and the listed scripts and docs were read; test files were read only by name and the lines cited; `scripts/showcase/*.mjs` were searched by keyword, not read in full.
