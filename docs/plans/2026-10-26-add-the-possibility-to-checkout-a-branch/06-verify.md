# Check out a config repo branch · what was checked

## Summary

<!-- ql:summary -->
Security review is running.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
Nothing is open.
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: running
  - Result: not yet
  - Why it runs here: The change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 6 checks
- **Independent review** · Claude Code · Opus 5.5 · never saw the build
  - State: waiting
  - Result: not yet
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
- **Fix rounds** · the build session
  - State: waiting
  - Result: not yet
<!-- /ql:lanes -->

## Done means, checked

<!-- ql:proof -->
Not checked yet.
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
None so far.
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: a branch or tag name now crosses into a subprocess as an argument, from the flag and from the desktop's bound method (whose value comes, the Plan says it is risky: **Look here:** risky · it runs in the person's config checkout, where local edits or local commits could be lost; the switch refuses both and slice 1 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `go test ./...` · exit 0 · 7.1s
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `task lint` · exit 0 · 0.5s
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `go vet ./...` · exit 0 · 0.2s
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `task ci` · exit 0 · 8s
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `task desktop:test` · exit 0 · 2.6s
- all · 101e9dff61cb06643d4c7269ef0098e913c6ecbfad140af6f563675c9b4f49a7 · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.1s
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->
### Security

**Result:** done · nothing severe in the change
**Boundaries:** `kirokit sync --branch <name>` (CLI argument to `git` subprocess); `App.SwitchRef` (page to Wails-bound method to bundled kirokit); the `origin` remote's branch and tag names (clone to `git` output to page HTML); `settings.json` `ref` (file to sync arguments); `kirokit setup --branch` (same name rule, now shared)
- Reviewed 0053b15..HEAD, not only the first commit. `SwitchRef` calls `checkout.CheckRefName`, then `runStep` runs `exec.Command(cli, "sync", team, "--branch", ref)` with no shell. The CLI checks the name again, and a ref that starts with `-` cannot become an option. Shell metacharacters such as `;` or `$(` that git allows in ref names are never interpreted.
- `checkout.Read` runs git with an argument list and `-C <repoDir>`, with `GIT_*` stripped. It reads only local refs, takes no user input, and the two `run_*.go` files only hide the console window.
- Ref names from the clone can contain `<`, `>`, `"` and `&`. Every name interpolated into HTML in `version.js` (`versionOptions`, `sourceLine`) and `welcome.js` goes through `esc()`, which escapes `& < > " '`. `optgroup` labels are static, and the Settings button label is set with `textContent`. No injection found.
- `SaveSettings` keeps the stored `Ref`, and `Clone` resets it. A hand-edited `ref` in `settings.json` still reaches the CLI only as the value of `--branch`, where the name rule applies.
- `SwitchRef` echoes the ref into the kirokit terminal header. A caller that bypasses the page could put escape sequences in it (git ref names cannot contain control characters). That only affects the user's own terminal and needs script control of the page, so it is not recorded as a finding.
<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification
