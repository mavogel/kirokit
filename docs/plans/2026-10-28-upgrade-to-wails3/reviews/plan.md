# Plan review by Codex (gpt-6.1-sol · high)

REVISE: 4 items

1. **03-plan.md — Trust boundary / Showcase server build:** Setting only the port does not enforce loopback: inherited `WAILS_SERVER_HOST` overrides Wails’ configuration ([official documentation](https://wails.io/guides/server-build/#environment-variables)). Explicitly force loopback and make the first server spike verify that an inherited wildcard host cannot expose `Run` or `ShellWrite`.

   Answer: Taken · `record.sh` now starts the server build with `WAILS_SERVER_HOST=127.0.0.1` and `WAILS_SERVER_PORT=34115` set explicitly, aborts unless `lsof` shows the 34115 listener on loopback only, and `showcase.test.mjs` gains a case that inherits `WAILS_SERVER_HOST=0.0.0.0` and asserts loopback (Decision *The showcase drives a localhost-only server build on the same port*, contract *Showcase server build*, Risks).

2. **03-plan.md — System design contracts / Smoke start:** Quitting directly from `WindowRuntimeReady` contradicts the oracle requiring a successful `State` call. Require a frontend acknowledgement after `State` resolves; verify that a rejected call reaches the failure timeout.

   Answer: Taken · the smoke start no longer uses `WindowRuntimeReady`: `api.start()` calls `State` and emits `smoke:ready` only after it resolves, and Go quits only on that event; `api.test.js` checks a rejecting `State` emits nothing, and `smoke_test.go` keeps the timeout case (Oracle, Decision d3, contract *Smoke start*).

3. **03-plan.md — Windows upgrade check:** The old installation already satisfies the proposed key and file-existence assertions, so a failed replacement can pass; counting only the exact expected key also misses differently named duplicates. Compare installed app/CLI hashes with the new build and enumerate all kirokit uninstall entries.

   Answer: Taken · the check now compares `Get-FileHash` of the installed `kirokit-desktop.exe` and `bin\kirokit.exe` with the new build's files and counts every uninstall entry under HKLM (both views) and HKCU whose `DisplayName` or `InstallLocation` mentions kirokit, requiring exactly one, `mavogelkirokit` (Decision d5, contract *Windows upgrade check*, Risks).

4. **03-plan.md — Linux compatibility checks:** Successful dependency installation cannot establish that the release’s cgo binary runs on that distro; the release currently builds arm64 on Ubuntu 26.04. Run the installed app’s smoke check in each claimed distro and architecture, and choose the build baseline from those results.

   Answer: Taken · Linux builds move to `ubuntu-24.04` / `ubuntu-24.04-arm` runners (the oldest claimed Ubuntu), the Linux container's base follows, and a CI matrix installs the `.deb` and runs the installed app's smoke start in `ubuntu:24.04`, `ubuntu:26.04`, `debian:12` and `debian:13` on amd64 and arm64; the docs list only images that pass (Decision d6, contract *Linux packaging and hosts*, Assumptions, Risks).
