# Showcase video of kirokit CLI and app for the README · what was checked

## Summary

<!-- ql:summary -->
Security review is running.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Only you can confirm:** The pacing feels right
- **Only you can confirm:** It looks right on the rendered GitHub page
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: running
  - Result: not yet
  - Why it runs here: The change crosses a trust boundary: the person's SOPS age key — creates a symlink to (or when set) inside, It never reads, copies or prints the key, The link is removed in the.
- **Project checks** · QualityLayer itself · no agent
  - State: waiting
  - Result: not yet
- **Independent review** · Claude Code · Opus 5.5 · never saw the build
  - State: waiting
  - Result: not yet
- **Second opinion** · another vendor's agent
  - State: skipped
  - Result: not run
  - Why it runs here: Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: the person's SOPS age key — creates a symlink to (or when set) inside, It never reads, copies or prints the key, The link is removed in the, the Plan says it is risky: the recording drives outside tools (a logged-in, a GitHub clone over the network, Chrome via a new dev-only dependency) and links the and slice 2 is marked risky; say "get a second opinion" to have one.
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
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: the person's SOPS age key — creates a symlink to (or when set) inside, It never reads, copies or prints the key, The link is removed in the, the Plan says it is risky: the recording drives outside tools (a logged-in, a GitHub clone over the network, Chrome via a new dev-only dependency) and links the and slice 2 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
Not run yet.
<!-- /ql:checks -->

## Quality pass

<!-- quality-record:security -->
### Security

**Result:** done · nothing severe in the change
**Boundaries:** the SOPS age key (record.sh links it into /tmp/kirokit-demo/home; never read, copied or printed, link removed in the EXIT trap, rm -rf does not follow it); the clone of the project's own public repository over HTTPS (fixed address in app-drive.mjs); the Wails dev server on localhost:34115 (wails 2.17.0 checks the websocket origin); the real kiro-cli session (cli-drive.sh never sends a chat message); local files written under /tmp and assets/.
<!-- /quality-record:security -->

<!-- quality-record:security-recheck -->

<!-- /quality-record:security-recheck -->

<!-- optional section here, only when there is content: "Independent review" -->

## Verification
