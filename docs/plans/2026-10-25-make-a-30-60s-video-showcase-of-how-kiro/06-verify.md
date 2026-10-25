# Showcase video of kirokit CLI and app for the README · what was checked

## Summary

<!-- ql:summary -->
6 of 8 points assessed · 5 passed

Independent review failed; the recorded findings need attention.
<!-- /ql:summary -->

## Open items

<!-- ql:open-items -->
- **Failed:** Point 3 · The app part shows the same job done with buttons · The app clip's rounded frame is broken in the committed showcase. The real app (Wails dev) clone, Install, setup dialog and the dock's `installed platform-team into` are all there, but every clip frame, CLI and app alike, has a second rounded top edge at y≈693 of 896 and no bottom border or corners. This is not a plain rounded frame (D1); a run whose `frame.png` rendered intact shows it right:
- **Unverified:** A failed run can delete another run's work folder. If a second run fails preflight while one is already running, its EXIT trap still removes the fixed `/tmp/kirokit-demo` and kills the `kirokit-showcase` tmux server, which belong to the other run
- **Unverified:** The GIF shows a Kiro session resume id. The CLI clip ends on `Resume with: kiro-cli --resume-id sess_…` for an empty chat session. It is not a credential, but the privacy check does not consider it
- **Unverified:** One check prints a different failure message than documented. `checks.sh length` on a non-media file exits 1 with only ffprobe's error, not `showcase: length: …`; it still fails closed
- **Unverified:** The new recorder tests are not part of `task ci` or CI. They run with `node --test scripts/showcase/`, as CONTRIBUTING now says
- **Only you can confirm:** The pacing feels right
- **Only you can confirm:** It looks right on the rendered GitHub page
<!-- /ql:open-items -->

## How it is checked

<!-- ql:lanes -->
- **Security review** · a fresh agent · read-only
  - State: done
  - Result: done · nothing severe in the change
  - Why it runs here: The change crosses a trust boundary: the person's SOPS age key — creates a symlink to (or when set) inside, It never reads, copies or prints the key, The link is removed in the.
- **Project checks** · QualityLayer itself · no agent
  - State: done
  - Result: done · 7 checks
- **Independent review** · fresh helper
  - State: fail
  - Result: failed · 1 of 6 points
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
| Done means | Proof | |
|---|---|---|
| 1. A 30–60 second showcase autoplays at the top of the GitHub README | Does the GIF loop by itself right under the intro on github.com, with both screenshots below? `README.md:10` sits between the intro and `## Quickstart`, and the GIF's loop count is 0 (forever); push the branch and open the repository page to confirm. | yours to confirm |
| 2. The CLI part shows a team's config installed and its agent starting | ![GIF frames show `kirokit install platform-team` printing `installed platform-team into /tmp/kirokit-demo/kiro`, `kiro-cli chat --agent devops --v3` starting, and `/jir` listing `/jira-search` and `/jira-triage (file prompt)`,](evidence/it1-scenario-1-step-4-install.png) | passed |
| 3. The app part shows the same job done with buttons | ![The app clip's rounded frame is broken in the committed showcase. The real app (Wails dev) clone, Install, setup dialog and the dock's `installed platform-team into` are all there, but every clip frame, CLI and app alike, has a second rounded top edge at y≈693 of 896 and no bottom border or corners. This is not a plain rounded frame (D1); a run whose `frame.png` rendered intact shows it right:](evidence/it1-done-3-committed-seam.png) | failed |
| 4. A viewer can follow every step without sound | Does each caption stay long enough to read without sound? There are four numbered captions plus title and end cards, and `caption.mjs` refuses any caption shorter than 2.5 s (fresh runs: caption 2 at 5.1 s of a 21.7 s CLI clip, caption 4 at 6.6 s of an 11.4 s app clip); watch the GIF and judge the pacing. D1 can also drop caption 1 (1 in 20 renders). | yours to confirm |
| 5. No private details are on screen | The privacy check passes on two fresh runs' cast, dock text and captions; independent greps find no home path, user name or git email in the cast; `strings` on both committed files finds none; the frames show only `/tmp/kirokit-demo` paths and the public GitHub slug,. | passed |
| 6. The video can be made again with one command when the UI changes | `scripts/showcase/` holds the recorder; two fresh `record.sh --keep` runs of the judged tree recorded the real CLI and app and rebuilt both files in about 52 s; CONTRIBUTING's "Showcase video" names `task showcase`,. | passed |
| 7. The README GIF is small enough for GitHub to show it reliably | The GIF is 2395833 bytes (gate 8 MiB, well under ~10 MB), 1000×700, 12 fps, loops forever. | passed |
| 8. A sharp MP4 of the same showcase sits beside the GIF | `assets/kirokit-showcase.mp4` is 813601 bytes, H.264 `yuv420p` 1280×896 at 30 fps, 36.1 s; the GIF is made from that MP4 in the same run. | passed |
<!-- /ql:proof -->

## Fix rounds

<!-- ql:fix-rounds -->
- Round 1 · 3 points still open
<!-- /ql:fix-rounds -->

## Second opinion

<!-- ql:reviews -->
Not run: Second opinion is off in Settings, though it would run here because the change crosses a trust boundary: the person's SOPS age key — creates a symlink to (or when set) inside, It never reads, copies or prints the key, The link is removed in the, the Plan says it is risky: the recording drives outside tools (a logged-in, a GitHub clone over the network, Chrome via a new dev-only dependency) and links the and slice 2 is marked risky; say "get a second opinion" to have one.
<!-- /ql:reviews -->

## Checks

<!-- ql:checks -->
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `go test ./...` · exit 0 · 3.8s
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `golangci-lint run` · exit 0 · 0.7s
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `go vet ./...` · exit 0 · 0.1s
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `task build` · exit 0 · 0.4s
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.2s
- all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `lefthook run pre-commit` · exit 0 · 0.1s
- all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `task showcase` · exit 0 · 45.1s
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

### Iteration 1

**Revision:** 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c

**Assignment:** 244b2621-ea9f-4a7f-8211-c2cf33cec7bb
**Judge:** fresh helper

- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `go test ./...` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `golangci-lint run` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `go vet ./...` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `task build` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `lefthook run pre-commit` · exit 0 · 0s · reused
- [PASS] check: reused receipt — all · 054d6fbf92a6ae58031ed416b27885400b6e7d338c5df9d82fc32bf2cb66f00c · `task showcase` · exit 0 · 45.1s
- [USER] done 1: Does the GIF loop by itself right under the intro on github.com, with both screenshots below? `README.md:10` sits between the intro and `## Quickstart`, and the GIF's loop count is 0 (forever) — `evidence/it1-scenario-1.txt`; push the branch and open the repository page to confirm.
- [PASS] done 2: GIF frames show `kirokit install platform-team` printing `installed platform-team into /tmp/kirokit-demo/kiro`, `kiro-cli chat --agent devops --v3` starting, and `/jir` listing `/jira-search` and `/jira-triage (file prompt)` — `evidence/it1-scenario-1-step-4-install.png`, `evidence/it1-scenario-1-step-4-jir.png`
- [FAIL] done 3: The app clip's rounded frame is broken in the committed showcase. The real app (Wails dev) clone, **Install**, setup dialog and the dock's `installed platform-team into` are all there, but every clip frame, CLI and app alike, has a second rounded top edge at y≈693 of 896 and no bottom border or corners. This is not a plain rounded frame (D1) — `evidence/it1-done-3-committed-seam.png`; a run whose `frame.png` rendered intact shows it right: `evidence/it1-done-3-fresh-run-no-seam.png`
- [USER] done 4: Does each caption stay long enough to read without sound? There are four numbered captions plus title and end cards, and `caption.mjs` refuses any caption shorter than 2.5 s (fresh runs: caption 2 at 5.1 s of a 21.7 s CLI clip, caption 4 at 6.6 s of an 11.4 s app clip) — `evidence/it1-dod-keep2.txt`; watch the GIF and judge the pacing. D1 can also drop caption 1 (1 in 20 renders).
- [PASS] done 5: the privacy check passes on two fresh runs' cast, dock text and captions; independent greps find no home path, user name or git email in the cast; `strings` on both committed files finds none; the frames show only `/tmp/kirokit-demo` paths and the public GitHub slug — `evidence/it1-scenario-2.txt`, `evidence/it1-scenario-1.txt`
- [PASS] done 6: `scripts/showcase/` holds the recorder; two fresh `record.sh --keep` runs of the judged tree recorded the real CLI and app and rebuilt both files in about 52 s; CONTRIBUTING's "Showcase video" names `task showcase` — `evidence/it1-dod-keep2.txt`, `evidence/it1-keep-run.txt`
- [PASS] done 7: the GIF is 2395833 bytes (gate 8 MiB, well under ~10 MB), 1000×700, 12 fps, loops forever — `evidence/it1-scenario-1.txt`
- [PASS] done 8: `assets/kirokit-showcase.mp4` is 813601 bytes, H.264 `yuv420p` 1280×896 at 30 fps, 36.1 s; the GIF is made from that MP4 in the same run — `evidence/it1-scenario-1.txt`
- [FAIL] scenario 1: The committed showcase's rounded frame is broken. Step 1 is the receipt (`task showcase` exit 0). Steps 2–4 and 6 hold: 36.1 s, GIF 2395833 B, MP4 813601 B, the install and the `/jir` list with `jira-search` and `jira-triage` show, and `/tmp/kirokit-demo` is absent. Step 5 shows the Clone form and the dock's `installed platform-team into`, but the frame around them is broken (D1) — `evidence/it1-scenario-1.txt`, `evidence/it1-scenario-1-step-5-clone.png`, `evidence/it1-scenario-1-step-5-installed.png`
- [PASS] scenario 2: all 4 steps as expected: home path and git email exit 1 naming the file, the public slugs exit 0, and the kept `cli.cast`, `app-dock.txt` and `captions.txt` from a `record.sh --keep` run of the judged tree exit 0. Probes: an empty file passes, an upper-cased user name fails, no files fails — `evidence/it1-scenario-2.txt`
- [PASS] scenario 3: SIGINT to the run's process group during the app step: exit 130, `/tmp/kirokit-demo` gone, no wails/vite process and no 34115/5173 listener, the key's inode and mtime and the `assets/` shasums identical — `evidence/it1-scenario-3.txt`; probe: SIGINT to the script's PID alone waits for the driver, then cleans up the same way after 9 s — `evidence/it1-scenario-3-probe-pid.txt`
- [PASS] scenario 4: in an isolated temp repo the staged GIF and MP4 keep SHA-1 `fcf2436…`/`e32ddaf…` through `lefthook run pre-commit` (exit 0); a control with the pre-change `lefthook.yml` rewrites both (exit 1) — `evidence/it1-scenario-4.txt`
- [PASS] scenario 5: steps 1–3 as expected: the GIF (width 800, linked to the MP4) sits between the intro and `## Quickstart`, both screenshots remain, the CLI sentence names `jira-search` as the skill and `jira-triage` as a prompt, and CONTRIBUTING has "Showcase video"; step 4 is done 1 — `evidence/it1-scenario-5.txt`
- [PASS] dod T1: `npm ls --prefix scripts/showcase playwright-core@1.63.0` exit 0; `node_modules` is ignored and absent from `git status` — `evidence/it1-dod-keep2.txt`
- [PASS] dod T2: fresh runs reach `Session ended`; new test: a shell without `kirokit` exits 1 with `timeout waiting for installed platform-team` in 32.7 s; shellcheck and `bash -n` exit 0 — `evidence/it1-new-tests.txt`
- [PASS] dod T3: the kept `cli.cast` holds `installed platform-team` and `jira-search` and no model turn; new test: a busy 34115 fails with `close the running wails dev first`; without `--keep` the folder is gone (Scenario 3) — `evidence/it1-dod-keep2.txt`
- [PASS] dod T4: writes `app.webm`, `app-dock.txt` (holds `installed platform-team into`) and `app-marks.json`; with no server it exits 1 with `app-drive: welcome:` (new test); with `data-team` installed beforehand it confirmed "Replace the installed team?" and installed platform-team — `evidence/it1-dod-t4-replace-frames.png`, `evidence/it1-dod-t4-replace.txt`
- [PASS] dod T5: a kept run leaves the three app files; after SIGINT no dev server survives, the link is gone and the key's inode and mtime are unchanged; the 180 s DevServer timeout was read, not run (a 720 × 0.25 s loop, then `fail`) — `evidence/it1-scenario-3.txt`
- [PASS] dod T6: new tests: `cap-1…4.png` are 1280×96 RGBA with a transparent corner, the caption 2 cue falls after `installed platform-team`, and a missing `cli.cast` exits 1 naming it (the `frame.png` flakiness is D1) — `evidence/it1-new-tests.txt`
- [PASS] dod T7: fresh `showcase.mp4` is 1280×896 `yuv420p`, 37.8 s, and the GIF is 1000 px wide; `caption.mjs` enforces 2.5 s per caption (new test) — `evidence/it1-dod-keep2.txt`
- [PASS] dod T8: new tests: the public slugs pass; user name, home path and email fail naming the file; the 8/4 MiB and 30/60 s boundaries hold; shellcheck exit 0; `task showcase` receipt exit 0 — `evidence/it1-new-tests.txt`
- [PASS] dod T9: both fixers exclude `**/*.gif` and `**/*.mp4`; staged media SHA-1 unchanged (Scenario 4); `check-yaml` exit 0 — `evidence/it1-scenario-4.txt`
- [PASS] dod T10: GIF 2.4 MB and MP4 0.8 MB, both 36.1 s; frames show the CLI install, `/jir` with `jira-search`, the Clone form and the dock's `installed platform-team into` — `evidence/it1-scenario-1.txt`
- [PASS] dod T11: the README shows the GIF under the intro with the MP4 link and the `jira-search` wording, both screenshots stay; CONTRIBUTING has "Showcase video"; AGENTS.md has the pointer; `grep -c` gives 2 — `evidence/it1-scenario-5.txt`
- [PASS] truth oracle: the committed GIF is 36.1 s, and its frames show `kirokit install platform-team`, `/jir` listing `jira-search` and the dock's `installed platform-team into …`, with no home path or user name in frames or text (rendering on github.com is done 1) — `evidence/it1-scenario-1-step-5-installed.png`, `evidence/it1-scenario-1-step-4-jir.png`
- [PASS] truth 1: an interrupted run left `assets/` byte-identical, removed `/tmp/kirokit-demo` and left the age key's inode and mtime unchanged; a failed preflight leaves no folder (new test) — `evidence/it1-scenario-3.txt`
- [FAIL] review: D1 · The rounded frame image often renders wrongly, and the checks publish it anyway. `caption.mjs` reuses one Chrome page, resizing the viewport from 1280×896 to 1180×760 before the `frame.png` screenshot. In 20 identical runs it drew a broken `frame.png` 13 times (top rows repeated from y≈673, bottom border and corners lost) and a fully blank `cap-1.png` once. The committed media and a fresh full `record.sh --keep` run both carry the seam, and no check looks at the pictures — `evidence/it1-d1-frame-repeat.txt`, `evidence/it1-d1-frame-variants.png`, `evidence/it1-d1-cap1-blank.png`, `evidence/it1-d1-rerun-seam.png`
- [PASS] review: no swallowed errors, no stack drift (`playwright-core` 1.63.0 as planned); every file is inside a Files block or the AGENTS.md amendment; the trust boundaries match the Security block (key linked, never read, link removed in the trap; fixed public clone address); tests: 23 tests added · docs: CONTRIBUTING.md
- [NOTE] review: A failed run can delete another run's work folder. If a second run fails preflight while one is already running, its EXIT trap still removes the fixed `/tmp/kirokit-demo` and kills the `kirokit-showcase` tmux server, which belong to the other run.
- [NOTE] review: The GIF shows a Kiro session resume id. The CLI clip ends on `Resume with: kiro-cli --resume-id sess_…` for an empty chat session. It is not a credential, but the privacy check does not consider it.
- [NOTE] review: One check prints a different failure message than documented. `checks.sh length` on a non-media file exits 1 with only ffprobe's error, not `showcase: length: …`; it still fails closed — `evidence/it1-scenario-2.txt`
- [NOTE] review: The new recorder tests are not part of `task ci` or CI. They run with `node --test scripts/showcase/`, as CONTRIBUTING now says.

**Pictures:** `evidence/it1-scenario-1-step-4-install.png`, `evidence/it1-scenario-1-step-4-jir.png`, `evidence/it1-scenario-1-step-5-clone.png`, `evidence/it1-scenario-1-step-5-installed.png`, `evidence/it1-done-3-committed-seam.png`, `evidence/it1-done-3-fresh-run-no-seam.png`, `evidence/it1-d1-frame-variants.png`, `evidence/it1-d1-cap1-blank.png`, `evidence/it1-d1-rerun-seam.png`, `evidence/it1-dod-t4-replace-frames.png`

**Try it:**

1. Run the recorder's own tests: `node --test scripts/showcase/`; look for 23 passing (about 40 s).
2. Record the showcase: `task showcase`; look for `showcase: publish: wrote assets/kirokit-showcase.gif … and assets/kirokit-showcase.mp4 …` and exit code 0.
3. Check length and size: `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4` (30–60) and `stat -f%z assets/kirokit-showcase.*` (GIF ≤ 8388608, MP4 ≤ 4194304).
4. Look at one frame: `ffmpeg -ss 18 -i assets/kirokit-showcase.mp4 -frames:v 1 /tmp/f.png`; the rounded frame should have one clean border with rounded bottom corners. Today a second edge crosses it near the bottom (D1).
5. Prove the privacy gate: `echo "$HOME" > /tmp/h.txt && scripts/showcase/checks.sh privacy /tmp/h.txt`; look for `showcase: privacy: home path in /tmp/h.txt` and exit 1.
6. Open the README on github.com after pushing; the GIF loops under the intro.

**Result:** fail

### Iteration 2

**Assignment:** de0cd579-eed6-41e0-8214-89ffbdd2a182
**Judge:** fresh helper

- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `go test ./...` · exit 0 · 3.8s
- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `golangci-lint run` · exit 0 · 0.7s
- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `go vet ./...` · exit 0 · 0.1s
- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `task build` · exit 0 · 0.4s
- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `node scripts/sync-agent-assets.mjs --check` · exit 0 · 0.2s
- [PASS] check: recorded — all · 7f5cb3fedbdee2d689087f7462d42927a38091e7db75c28a2960cd93fea5479a · `lefthook run pre-commit` · exit 0 · 0.1s
- [USER] done 1: Does the GIF loop by itself right under the intro on github.com, with both screenshots below? `README.md:10-11` is unchanged since iteration 1 and sits above `## Quickstart`; the re-recorded GIF is 36.4 s with a NETSCAPE loop block — `evidence/it2-scenario-5.txt`, `evidence/it2-scenario-1.txt`; push the branch and open the repository page to confirm.
- [PASS] done 2: the re-recorded GIF shows `kirokit install platform-team` printing `installed platform-team into /tmp/kirokit-demo/kiro`, `kiro-cli chat --agent devops --v3` starting, and `/jir` listing `/jira-search` and `/jira-triage (file prompt)` — `evidence/it2-scenario-1-step-4-install.png`, `evidence/it2-scenario-1-step-4-jir.png`
- [PASS] done 3: the committed media show the real app (Wails dev) clone, **Install** and the dock's `installed platform-team into` inside one complete rounded frame; sampled every 0.5 s from 2.5 to 32.5 s, the bottom, left and right border pixels are always the frame colour (≈`#3a3450` after H.264), never clip content as with D1; a fresh run's `frame.png` has all four borders and a rounded bottom-left corner — `evidence/it2-done-3-border-scan.txt`, `evidence/it2-scenario-1-step-5-installed.png`, `evidence/it2-fresh-run.txt`
- [USER] done 4: Does each caption stay long enough to read without sound? The re-recorded cut has four numbered captions plus title and end cards; caption 2 starts at 5.27 s of a 20.18 s CLI clip and caption 4 at 6.59 s of an 11.41 s app clip, and every caption render is now complete (20 of 20) — `evidence/it2-scenario-1-contact.png`, `evidence/it2-d1-regression.txt`; watch the GIF and judge the pacing.
- [PASS] done 5: `strings` on both committed files finds no home path, user name or git email; the fresh run's privacy check on its cast, dock text and captions exits 0; the frames show only `/tmp/kirokit-demo` paths and the public GitHub slug — `evidence/it2-scenario-2.txt`, `evidence/it2-fresh-run.txt`
- [PASS] done 6: a fresh `record.sh --keep` of the judged tree (exported with `git archive`) recorded the real CLI and app and rebuilt both files, exit 0; CONTRIBUTING's "Showcase video" names `task showcase` — `evidence/it2-fresh-run.txt`, `evidence/it2-scenario-5.txt`
- [PASS] done 7: the committed GIF is 2453580 bytes (gate 8 MiB, well under ~10 MB), 1000×700, 12 fps, loops forever — `evidence/it2-scenario-1.txt`
- [PASS] done 8: `assets/kirokit-showcase.mp4` is 825273 bytes, H.264 `yuv420p` 1280×896 at 30 fps, 36.4 s; the GIF is made from that MP4 in the same run — `evidence/it2-scenario-1.txt`
- [PASS] scenario 1: all 6 steps as expected. Step 1 ran `scripts/showcase/record.sh --keep` (what `task showcase` runs) in a `git archive` copy of the judged tree, so the checkout's `assets/` stayed as committed; exit 0, and the log names every step. Steps 2–6: 36.4 s, GIF 2453580 B and MP4 825273 B, the CLI install and `/jir` frames, the Clone and `installed platform-team into` frames in a complete rounded frame, and no `/tmp/kirokit-demo` — `evidence/it2-scenario-1.txt`, `evidence/it2-fresh-run.txt`, `evidence/it2-scenario-1-step-5-clone.png`, `evidence/it2-scenario-1-step-5-installed.png`
- [PASS] scenario 2: steps 1–3 rerun because `checks.sh` changed: the home path and git email exit 1 naming the file, and the public slugs exit 0. Step 4: the fresh run's kept `cli.cast`, `app-dock.txt` and `captions.txt` exit 0. Probe: an empty file fails `checks.sh length` with `cannot read the length of …` — `evidence/it2-scenario-2.txt`, `evidence/it2-fresh-run.txt`
- [PASS] scenario 3: the run started over a kept folder whose pid was dead, then got SIGINT to its process group during the app step: exit 130, `/tmp/kirokit-demo` gone, no wails/vite process and no 34115/5174 listener, and the key's inode and mtime and the `assets/` shasums are unchanged — `evidence/it2-scenario-3.txt`
- [PASS] scenario 4: in an isolated temp repo with the judged `lefthook.yml`, the staged GIF, MP4 and an evidence file with trailing spaces and no final newline keep SHA-1 `6b6eac5…`/`b1b4950…`/`baf0923…` through `lefthook run pre-commit` (exit 0); iteration 1's control still shows the pre-change hooks rewriting media — `evidence/it2-scenario-4.txt`
- [PASS] scenario 5: steps 1–3 as expected: the GIF (width 800, linked to the MP4) sits between the intro and `## Quickstart`, both screenshots remain, the CLI sentence names `jira-search` as the skill and `jira-triage` as a prompt, and CONTRIBUTING has "Showcase video" plus the new `node --test` line; step 4 is done 1 — `evidence/it2-scenario-5.txt`
- [PASS] dod T1: carried from iteration 1; `package.json`, `package-lock.json`, `.gitignore` and `tmux.conf` are unchanged since 3fe8fee — `evidence/it1-dod-keep2.txt`
- [PASS] dod T2: carried from iteration 1; `cli-drive.sh` is unchanged since 3fe8fee, and its timeout test passes again — `evidence/it2-recorder-tests.txt`
- [PASS] dod T3: a live pid in `/tmp/kirokit-demo/pid` fails preflight with `another showcase run (pid …) is using /tmp/kirokit-demo` and leaves the folder; a dead run's folder is taken over; the busy-port test passes; without `--keep` the folder is gone (Scenario 3) — `evidence/it2-truth-1-guard.txt`, `evidence/it2-recorder-tests.txt`
- [PASS] dod T4: carried from iteration 1; `app-drive.mjs` is unchanged since 3fe8fee, and the welcome-step test passes again — `evidence/it2-recorder-tests.txt`, `evidence/it1-dod-t4-replace.txt`
- [PASS] dod T5: the fresh kept run leaves `app.webm`, `app-dock.txt` (holds `installed platform-team into`) and `app-marks.json`; after SIGINT no dev server survives, the link is gone and the key is untouched — `evidence/it2-fresh-run.txt`, `evidence/it2-scenario-3.txt`
- [PASS] dod T6: with a fresh page per image, 20 of 20 renders paint all four frame borders, a clear centre and every caption, and the fresh run's captions are 1280×96 with an opaque centre; the regression test fails on the pre-fix `caption.mjs` (`run 2: frame bottom border`) — `evidence/it2-d1-regression.txt`, `evidence/it2-fresh-run.txt`
- [PASS] dod T7: the fresh `showcase.mp4` is 1280×896 `yuv420p`, 36.3 s, and the GIF is 1000 px wide; the caption cues keep every caption on screen at least 2.5 s — `evidence/it2-fresh-run.txt`
- [PASS] dod T8: the privacy steps pass as in Scenario 2, and the size, length and unreadable-file tests pass — `evidence/it2-scenario-2.txt`, `evidence/it2-recorder-tests.txt`
- [PASS] dod T9: both fixers exclude `**/*.gif` and `**/*.mp4` (and now `docs/plans/*/evidence/**`); the staged media keep their SHA-1 (Scenario 4) — `evidence/it2-scenario-4.txt`
- [PASS] dod T10: the re-recorded GIF is 2.45 MB and the MP4 0.83 MB, both 36.4 s; frames at each caption show the CLI install, `/jir` with `jira-search`, the Clone form and the dock's `installed platform-team into` — `evidence/it2-scenario-1.txt`
- [PASS] dod T11: the README is unchanged since iteration 1 (GIF under the intro, MP4 link, `jira-search` wording, both screenshots); CONTRIBUTING has "Showcase video"; `grep -c` gives 2 — `evidence/it2-scenario-5.txt`
- [PASS] truth oracle: the committed GIF is 36.4 s, and its frames show `kirokit install platform-team`, `/jir` listing `jira-search` and the dock's `installed platform-team into …`, with no home path or user name in frames or text (rendering on github.com is done 1) — `evidence/it2-scenario-1-step-5-installed.png`, `evidence/it2-scenario-1-step-4-jir.png`
- [PASS] truth 1: an interrupted run left `assets/` byte-identical, removed `/tmp/kirokit-demo` and left the age key's inode and mtime unchanged; a second run refused by preflight leaves the live run's folder intact — `evidence/it2-scenario-3.txt`, `evidence/it2-truth-1-second-run.txt`, `evidence/it2-truth-1-guard.txt`
- [PASS] review: D1 is fixed (one fresh page per image, sized at creation, waiting for load, fonts and two frames); the run-ownership guard lets cleanup touch nothing until this run owns the folder; the length check now fails with its own message; no swallowed errors, no stack drift; the lefthook evidence exclude is a recorded deviation; the trust boundaries are unchanged. The suite has 26 passing and 2 skipped (a foreign dev server holds port 5173), exit 0 — `evidence/it2-recorder-tests.txt`; tests: 2 tests added · docs: no doc impact
- [NOTE] review: Two recorder tests broke when an unrelated dev server held port 5173. Their port probe only bound 127.0.0.1, so a vite on `[::1]:5173` slipped past it and `record.sh`'s own lsof preflight then failed them; the probe now asks lsof the same way `record.sh` does, and the two tests skip instead (test-only, unstaged) — `evidence/it2-new-tests.txt`
- [NOTE] review: The GIF still shows a Kiro session resume id. The CLI clip ends on `Resume with: kiro-cli --resume-id sess_…` for an empty chat session. It is not a credential, but the privacy check does not consider it — `evidence/it2-cli-end.png`
- [NOTE] review: The recorder tests are still outside `task ci` and CI. They run with `node --test scripts/showcase/`, as CONTRIBUTING says.

**Pictures:** `evidence/it2-scenario-1-contact.png`, `evidence/it2-scenario-1-step-4-install.png`, `evidence/it2-scenario-1-step-4-jir.png`, `evidence/it2-scenario-1-step-5-clone.png`, `evidence/it2-scenario-1-step-5-installed.png`, `evidence/it2-scenario-1-mp4-16.png`, `evidence/it2-scenario-1-mp4-31.png`, `evidence/it2-fresh-run-cli.png`, `evidence/it2-fresh-run-installed.png`, `evidence/it2-cli-end.png`

**Try it:**

1. Run the recorder's own tests: `node --test scripts/showcase/`; look for 28 tests and no failures (about 60 s; the two `record.sh` folder tests skip while anything listens on port 34115 or 5173).
2. Record the showcase: `task showcase` (no other dev server on 5173); look for `showcase: publish: wrote assets/kirokit-showcase.gif … and assets/kirokit-showcase.mp4 …` and exit code 0.
3. Check length and size: `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4` (30–60) and `stat -f%z assets/kirokit-showcase.*` (GIF ≤ 8388608, MP4 ≤ 4194304).
4. Look at one frame: `ffmpeg -ss 31 -i assets/kirokit-showcase.mp4 -frames:v 1 /tmp/f.png`; the app sits in one clean rounded border with rounded bottom corners, and the dock shows `installed platform-team into /tmp/kirokit-demo/home/.kiro`.
5. Prove the privacy gate: `echo "$HOME" > /tmp/h.txt && scripts/showcase/checks.sh privacy /tmp/h.txt`; look for `showcase: privacy: home path in /tmp/h.txt` and exit 1.
6. Open the README on github.com after pushing; the GIF loops under the intro.

**Result:** pass
