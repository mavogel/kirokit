# Ship CLI and desktop app for Linux and Windows · what was checked

## Summary

<!-- ql:summary -->
Security review is running.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Need:** T10 · Push this task's branch and open a pull request, so CI and Desktop run on the windows-latest, ubuntu-latest and macos-latest runners · queued for Review
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: running
  - Result: not yet
  - Why it runs here: The change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers.
- **Project checks** · QualityLayer itself · no agent
  - State: waiting
  - Result: not yet
- **Independent review** · Claude Code · Opus 5.5 · never saw the build
  - State: waiting
  - Result: not yet
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers, the Plan says it is risky: yes, The shipped app loads a new dependency (a Windows ConPTY library), secrets cross and slice 3 and 6 is marked risky; say "get a second opinion" to have one.
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
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: secrets and other processes, Decrypted SOPS values reach MCP servers, the Plan says it is risky: yes, The shipped app loads a new dependency (a Windows ConPTY library), secrets cross and slice 3 and 6 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
Not run yet.
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->

<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification
