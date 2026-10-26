---
type: research
---

# Desktop Settings page, Go-to-JS bindings, and where `--latest-tag` / Update / Sync are referenced

## Findings

### The Settings page is one static HTML section driven by `initSettings`, with per-field error spans

**Answers:** 6

- `desktop/frontend/index.html:69-79` holds `page-settings`. Each of the three fields (`kirohome`, `repo`, `repodir`) is a `.field` with `<input>`, `<span class="hint">` and `<span class="error" role="alert">`, tied together with `aria-describedby` (`desktop/frontend/index.html:71-73`).
- Below the fields sit `env-note` (a `role="status"` paragraph, line 74), the Save button with `save-status` (line 75), and the "Templates" block: a static note "Update fetches the newest release of your team's repository, and re-installs the installed team from it." (line 77) plus the Update button `#update` and `#update-error` (line 78). The Update copy is hard-coded HTML, not generated in JS.
- `initSettings(onUpdate, onClone)` (`desktop/frontend/src/pages/settings.js:39`) looks the elements up by id and returns `{ load, paint, cloneDone }` (line 138).
  - `show(input, errEl, message)` sets the error text and toggles `aria-invalid` (lines 52-56).
  - `load()` calls `api.state()`. It fills KiroHome and RepoDir unless the person has typed (`edited` / `dirEdited`), fills the address only if the field is empty, remembers `saved.repoUrl`, shows `env.note`, and re-polls every 300 ms until `env.ready` (lines 69-81).
  - The Save click handler clears the errors, builds the form, calls `saveAndMaybeClone(api.saveSettings, ...)`, and sets the status to "Saved. Cloning the new repository..." or "Saved on this computer" (lines 83-102).
  - On error it runs `parseFieldErrors`, moves a clone-start refusal (no field prefix) to the address field, shows the field errors, puts any other text in the status line, and focuses the first bad field (lines 103-117).
  - `paint(view)` disables Update while `view.running` and shows `view.actionError.message` when `where === 'update'` (lines 120-124).
  - `cloneDone({ok,error})` shows a failure beside the address, or "Cloned. The new repository is in use." / the no-release message in the status line (lines 128-136).
- `parseFieldErrors` splits the error message on newlines and maps the prefixes `KiroHome: `, `RepoDir: `, `RepoURL: ` to fields; any other non-blank line goes to `other` (`desktop/frontend/src/pages/settings.js:5-14`).
- `saveAndMaybeClone` saves the OLD address together with the form, and calls `clone(form.repoUrl)` only when the address is non-empty and differs from `saved.repoUrl` (lines 22-27). `Clone` itself stores the new address after success (`desktop/clone.go:41-49`).
- Update's arguments are built in JS: `updateArgs(installs)` returns `['sync', team, '--latest-tag']`, or `['sync', '--latest-tag']` with no install (`desktop/frontend/src/pages/settings.js:31-34`). `main.js` wires it: `initSettings(() => runCommand(updateArgs(view.installs), 'update'), cloneFromSettings)` (`desktop/frontend/src/main.js:140`). `runCommand` calls `api.run(args)`; a refusal becomes `view.actionError = {where:'update', message}` (`main.js:69-79`).
- Go side: `Settings{RepoURL, KiroHome, RepoDir}` with JSON keys `repoUrl`, `kiroHome`, `repoDir` (`desktop/settings.go:17-21`). The settings file has no branch or ref field today.
- `SaveSettings` trims every field and collects per-field problems as `errors.Join` of `"KiroHome: ..."`, `"RepoURL: ..."`, `"RepoDir: ..."` lines, saving nothing on any problem (`desktop/settings.go:57-82`). The RepoURL rule is `repoURLPattern` (line 27), whose message is `repoURLMessage` (line 101). An empty RepoURL is allowed. Persistence is an atomic temp-file rename with mode 0600 (`writeSettings`, lines 104-129). `SetRepoDir` stores only the clone folder (lines 87-98).
- Layering: the three field names are a shared contract between `settings.go` error prefixes and `parseFieldErrors`; adding a field needs both.
- Churn in the last 90 days (`git log --oneline --since=90.days`): `desktop/settings.go` 2 commits, `settings.js` 1, `index.html` 2, `README.md` 27.

### Update and Sync run `kirokit sync ... --latest-tag` as a PTY command through `App.Run`; the clone does the same internally

**Answers:** 6, 7, 8

- Sync button args: `syncArgs(team)` = `['sync', team, '--latest-tag']` (`desktop/frontend/src/pages/install.js:26`); the button is rendered per installed team at `install.js:20` and dispatched via `main.js:145-146`.
- `Run(args)` starts the bundled CLI in the kirokit tab (`desktop/runner.go:45-51`). A second run is refused with `errRunning` = "kirokit is still running a command" (`runner.go:14`, `:60-62`). The command runs in `repoDir()`, falling back to home (`runner.go:109-113`). When the first arg is `sync` and the program is `kirokit`, the catalog is re-rendered after exit (`runner.go:126`, `:151-153`).
- `Clone(address)` validates against `repoURLPattern`, reserves the tab, then `cloneAndSync` runs `git clone -- <address> <repo>` and then hard-codes `kirokit sync --latest-tag` (`desktop/clone.go:82`, `:90`). It treats output containing "no v* tag found" as OK-with-message (`clone.go:96-97`). The failure path restores the old clone (`clone.go:70-76`).

### Bound methods are the exported methods on `*App`; `api.js` is the only JS module that touches them

**Answers:** 7

- `main.go` binds one value: `Bind: []any{app}` (`desktop/main.go:32`), with `OnStartup: app.startup` and `OnShutdown: app.shutdown` (lines 30-31). Wails exposes every exported method of `App` as `window.go.main.App.<Method>`. Nothing generated is committed.
- `api.js` resolves `window.go.main.App` lazily and throws 'kirokit is not running inside its window' if absent (`desktop/frontend/src/api.js:5-9`). It has one thin wrapper per method:
  - `state` -> `State()` (line 11)
  - `setRepoDir` -> `SetRepoDir(dir)` (line 15)
  - `saveSettings` -> `SaveSettings(settings)` (line 17)
  - `catalog` -> `Catalog()` (line 19)
  - `installs` -> `Installs()` (line 21)
  - `prerequisites(team)` (line 24)
  - `run(args)` -> `Run(args)` (line 27)
  - `clone(address)` (line 32)
  - `runWrite` and `shellWrite` go through ordered "lines" (lines 38-53)
  - `shellResize` (line 55)
- Pages import only `api.js` (header comment, lines 1-3). Events come back through `on(event, handler)` (`api.js:85-88`). The events are `installs:changed`, `term:kirokit`, `term:kirokit:exit` and `clone:done`, emitted from Go via `a.emit` (`desktop/app.go:96-103`) and subscribed in `main.js:195-217`.
- Errors: a Go `error` return becomes a rejected promise whose `.message` is the Go error text (Wails behaviour; the code relies on it). Multi-field errors are `errors.Join` text split on newlines by `parseFieldErrors`; plain errors such as `errRunning` are shown verbatim via `message(e)` (`main.js:64`). Structured results such as `State` and `Catalog` are Go structs with JSON tags (`desktop/app.go:34-53`). Asynchronous outcomes are not returned but emitted as events (`clone:done` payload `cloneResult{ok,error}`, `desktop/clone.go:14-17`).
- A new bound method needs: (1) an exported method on `App` in `desktop/*.go`, (2) a wrapper in `api.js`, (3) a Go test that calls it on `NewAppAt(tmp)`, and (4) a JS test where the page logic depends on it. Frontend tests fake the binding by setting `globalThis.window = { go: { main: { App: {...} } } }` (`desktop/frontend/src/pages/pages.test.js:194-199`, `desktop/frontend/src/api.test.js:21-26`). Nothing checks that `api.js` method names match the Go methods; that wiring is unverified outside a running window (the smoke start only loads the page, `desktop/main.go:35-40`). Inference, not run.
- Churn: `api.js` 2 commits, `runner.go` 1, `clone.go` 1 in 90 days.

### Documentation and callers that name `--latest-tag`, `setup --branch`, Update and Sync

**Answers:** 8

Code (non-test):
- `cmd/kirokit/commands/sync.go:16`, `:27`, `:118`, `:125-139` define the flag; its help text says "check out the newest v* tag instead of fast-forwarding the current branch (also without a team)".
- `cmd/kirokit/commands/setup.go:25`, `:32-33`, `:48`, `:57` define and validate `--branch`; `setup` has no desktop caller.
- `desktop/clone.go:90`, `desktop/frontend/src/pages/settings.js:33`, `desktop/frontend/src/pages/install.js:26`.
- `scripts/kirokit-sync-setup.sh:17` (usage), `:32` (`--latest-tag) latest_tag="--latest-tag"`) passes the flag through to the scheduled sync.
- `desktop/frontend/index.html:77-78` is the Update copy; `desktop/frontend/src/pages/html.js:42` comments on the "update" error location.

Docs:
- `README.md:297` (setup description, newest `v*` tag), `:298` (`--branch`), `:300` (no `v*` tag -> default branch, exit 0), `:307-308` (sync examples), `:311` (sync semantics with `--latest-tag`), `:322` (script example with `--latest-tag`), `:327` (dedicated-checkout warning about detached HEAD), `:341` (First launch: `git clone --` then `kirokit sync --latest-tag`; Settings address change), `:344` (Sync = `kirokit sync <team> --latest-tag`; Settings Update = `kirokit sync --latest-tag`, with or without team), `:346` (`KIROKIT_REPO` + `sync --latest-tag`).
- `CONTRIBUTING.md` has no mention of these flags or buttons (only the generic rule at `CONTRIBUTING.md:152`; `AGENTS.md:33` repeats it). The showcase scripts (`scripts/showcase/*`) contain no `--latest-tag`, `--branch`, Update or Sync button references (grep over the folder returned nothing for these terms; their "update/sync" hits are unrelated identifiers).

Tests asserting these arguments or texts:
- `desktop/frontend/src/pages/pages.test.js:110-113` (`syncArgs`, `updateArgs`) and `:102-107` (Sync/Uninstall buttons).
- `desktop/clone_test.go:126` asserts the tab shows `▸ kirokit sync --latest-tag`; `:106-135` asserts the clone ends on the newest tag (v1.2.0).
- `desktop/runner_test.go:204` runs `sync --latest-tag` through `Run` and checks the catalog is re-rendered.
- `cmd/kirokit/commands/setup_test.go:71`, `:92-93` (`--branch` accepted/refused), `cmd/kirokit/commands/sync_test.go:79-103` (`--latest-tag`).
- No test asserts the Update note text in `index.html` (static HTML, not tested) or the `README` text.

## How it is tested today

- Frontend (node test runner, no DOM library): `cd desktop/frontend && node --test "src/**/*.test.js"` (`package.json` script `test`); one file: `cd desktop/frontend && node --test src/pages/pages.test.js`. Page logic is split into pure view functions (`teamsView`, `installView`, `welcomeView`, `updateArgs`, `saveAndMaybeClone`) tested by string match on HTML or deepEqual. `initSettings` is tested with a hand-rolled fake `document`/`window.go.main.App` (`desktop/frontend/src/pages/pages.test.js:183-205`), covering the failed-clone and finished-clone status behaviour (`:207-241`). `parseFieldErrors` is exported but has no direct test in this file (inferred from the imports at `:7`). `index.html` is untested.
- Go desktop module: `cd desktop && go test ./...` (module in `desktop/go.mod`); one test: `cd desktop && go test -run TestSaveSettings_RepoURL ./`. Whole desktop gate from root: `task desktop:test` (`Taskfile` lines 32-37: builds frontend, node tests, `go vet`, `go test`). Root suite: `go test ./...` or `task ci`.
- Go fixtures: `newTestApp` (`desktop/settings_test.go:10-17`) uses temp HOME and `NewAppAt(tmp)` and calls `SaveSettings`/`State` directly, including the "survives a new App" and per-field error prefix checks (`:20-51`, `:106`). Runner and clone tests use `runnerApp` (`desktop/runner_test.go:84-96`), which installs a fake bundled `kirokit` script via `KIROKIT_DESKTOP_CLI` and an `eventLog` as `a.emitter` to capture events. Clone tests build the real CLI once (`realKirokit`, `desktop/clone_test.go:70-87`) and use real git origins from `newOrigin` (`:29-40`) plus `waitClone`. No test goes through Wails itself.
- CLI tests: `go test ./cmd/kirokit/commands -run TestSetup` style, using real git repos (`setup_test.go`, `sync_test.go`).

## Open questions

None.

## Code references

- Settings UI: `desktop/frontend/index.html:69-79`, `desktop/frontend/src/pages/settings.js:5-139`, `desktop/frontend/src/main.js:64-79,105-160,195-217`
- Go settings: `desktop/settings.go:17-129`, `desktop/app.go:14-53,96-143`
- Binding: `desktop/main.go:17-44`, `desktop/frontend/src/api.js:1-88`
- Run / clone: `desktop/runner.go:14-165`, `desktop/clone.go:14-101`
- Install page: `desktop/frontend/src/pages/install.js:6-46`
- Tests: `desktop/frontend/src/pages/pages.test.js`, `desktop/frontend/src/api.test.js`, `desktop/settings_test.go`, `desktop/clone_test.go`, `desktop/runner_test.go`, `desktop/app_test.go`
- Docs and scripts: `README.md:297-346`, `scripts/kirokit-sync-setup.sh:17,32`, `CONTRIBUTING.md:152`
- CLI: `cmd/kirokit/commands/sync.go`, `cmd/kirokit/commands/setup.go`

Exhaustive for the files named in the questions; the grep for the flag and button terms covered the whole repository except `node_modules`, `.git`, `dist`, `testdata`, `teams`, `docs` and `.claude`, so mentions under `docs/` were not searched.
