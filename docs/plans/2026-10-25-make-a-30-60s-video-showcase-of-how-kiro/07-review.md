# Showcase video of kirokit CLI and app for the README

<!-- ql:links -->
Task: `docs/plans/2026-10-25-make-a-30-60s-video-showcase-of-how-kiro/`
<!-- /ql:links -->

## Why the change

Visitors to the repository had to read a long README to understand kirokit; a 35-second looping GIF at the top now shows the CLI and the desktop app doing the job, and `task showcase` records it again from the real tools whenever the UI changes.

## What a reviewer must know

- Recording runs only on a maintainer's Mac (signed-in `kiro-cli`, the SOPS age key, Google Chrome, network access to github.com), and every run produces new media bytes. That is why `task showcase` is not a project check or a CI job; `node --test scripts/showcase/` covers the recorder's logic without recording.
- `lefthook.yml` now keeps the whitespace fixers off `*.gif`, `*.mp4` and `docs/plans/*/evidence/**`: without that, committing the media or captured proof silently rewrote their bytes.
- Follow-ups from the reviewer, not done here: the recorder tests are not part of `task ci`/CI; the CLI clip ends on Kiro's `Resume with: kiro-cli --resume-id sess_…` line for an empty session (not a credential, not scanned by the privacy check); two recorder tests skip while another dev server holds port 5173.

## What changed

The recorder is one entry script that drives two real tools and composes their clips; everything runs in a throwaway `/tmp/kirokit-demo` the script owns and removes:

```text
task showcase -> scripts/showcase/record.sh [--keep]
  1 preflight     tools, age key, free ports 34115/5173, no other live run
  2 build         kirokit CLI -> /tmp/kirokit-demo/bin and desktop/build/dev-bin
  3 CLI clip      tmux + cli-drive.sh (types; waits on screen text) -> asciinema -> agg -> cli.gif
  4 app clip      wails dev (HOME=/tmp/kirokit-demo/home, age key symlinked)
                  + app-drive.mjs (headless Chrome: Clone, Install, Install) -> app.webm, app-dock.txt, app-marks.json
  5 captions      caption.mjs -> cap-1..4.png, title/end cards, frame.png, cues.env
  6 compose       ffmpeg: title | CLI clip | 0.5 s crossfade | app clip (clone wait cut) | end card
                  -> showcase.mp4 (1280x896 H.264) -> showcase.gif (12 fps, 1000 px)
  7 checks        checks.sh privacy (home path, user name, git email) | size (8 MB / 4 MB) | length (30-60 s)
  8 publish       copy + rename into assets/ only after every check passed
```

What a reader of the repository sees:

```diff
 README.md
+  intro -> showcase GIF (width 800, links to the MP4) -> Quickstart
~  "/jir" lists the skill `jira-search` and the prompt `jira-triage` (was: "the example skill jira-triage")
 CONTRIBUTING.md
+  "Showcase video": tools, `task showcase`, what it refuses to publish, `node --test scripts/showcase/`
 assets/
+  kirokit-showcase.gif   2.3 MB, 35 s
+  kirokit-showcase.mp4   0.8 MB, 35 s
```

## Built, slice by slice

<!-- ql:slices -->
3 slices · 11 tasks · 13 commits · 17 files · +1214 −3

**Slice 1 · The CLI part records itself from one script** · T1, T2, T3 · 7 files · +805 −0

- `scripts/showcase/.gitignore` · T1 · new · +1 −0
- `scripts/showcase/cli-drive.sh` · T2 · new · +66 −0
- `scripts/showcase/package-lock.json` · T1 · new · +26 −0
- `scripts/showcase/package.json` · T1 · new · +9 −0
- `scripts/showcase/record.sh` · T3, T5, T7, T8 · new · +226 −0
- `scripts/showcase/showcase.test.mjs` · T3, T5, T6, T7, T8, T11 · new · +472 −0
- `scripts/showcase/tmux.conf` · T1 · new · +5 −0

**Slice 2 · One command makes the checked GIF and MP4** · T4, T5, T6, T7, T8 · 4 files · +380 −0

- `Taskfile.yml` · T8 · +5 −0
- `scripts/showcase/app-drive.mjs` · T4 · new · +134 −0
- `scripts/showcase/caption.mjs` · T6 · new · +176 −0
- `scripts/showcase/checks.sh` · T8 · new · +65 −0

**Slice 3 · The README opens with the showcase and commits keep media intact** · T9, T10, T11 · 6 files · +29 −3

- `AGENTS.md` · T11 · +1 −0
- `CONTRIBUTING.md` · T11 · +13 −0
- `README.md` · T11 · +5 −2
- `assets/kirokit-showcase.gif` · T10 · new · +0 −0
- `assets/kirokit-showcase.mp4` · T10 · new · +0 −0
- `lefthook.yml` · T9 · +10 −1
<!-- /ql:slices -->

## What was asked for, and how each point was proved

<!-- ql:proof -->
- **1. A 30–60 second showcase autoplays at the top of the GitHub README.** Only you can confirm: it looks right on the rendered GitHub page · waits for your confirmation
- **2. The CLI part shows a team's config installed and its agent starting.** Scenario 1: One command records the real CLI and app into both files, after slice 2 · passed · `evidence/it2-scenario-1-step-4-install.png`
- **3. The app part shows the same job done with buttons.** Scenario 1: One command records the real CLI and app into both files, after slice 2 · passed · `evidence/it2-done-3-border-scan.txt`
- **4. A viewer can follow every step without sound.** Only you can confirm: the pacing feels right · waits for your confirmation
- **5. No private details are on screen.** Scenario 2: Private details stop the run before anything is published, after slice 2 · Scenario 3: An interrupted run leaves no trace, after slice 2 · passed · `evidence/it2-scenario-2.txt`
- **6. The video can be made again with one command when the UI changes.** Scenario 1: One command records the real CLI and app into both files, after slice 2 · Scenario 3: An interrupted run leaves no trace, after slice 2 · Scenario 5: The README opens with the showcase, after slice 3 · passed · `evidence/it2-fresh-run.txt`
- **7. The README GIF is small enough for GitHub to show it reliably.** Scenario 1: One command records the real CLI and app into both files, after slice 2 · Scenario 4: Committing the media keeps its bytes, after slice 3 · passed · `evidence/it2-scenario-1.txt`
- **8. A sharp MP4 of the same showcase sits beside the GIF.** Scenario 1: One command records the real CLI and app into both files, after slice 2 · Scenario 4: Committing the media keeps its bytes, after slice 3 · passed · `evidence/it2-scenario-1.txt`
<!-- /ql:proof -->

## Deliberately not changed

<!-- ql:not-doing -->
- Recording the native macOS window or adding Screen Recording/Accessibility permissions.
- Voice-over, music, or posting the video anywhere; uploading the MP4 as a GitHub attachment stays your choice (`gh … --attach` can do it later).
- A CI job that re-records; recording needs a logged-in `kiro-cli` and a SOPS key, so it runs on a maintainer's Mac.
- Any change to how the `kirokit` CLI or the desktop app behaves.
- Git LFS for the media.
<!-- /ql:not-doing -->

## Known limits

<!-- ql:limits -->
None recorded.
<!-- /ql:limits -->

## How it was verified

<!-- ql:checked -->
Security review, Project checks, Code review and Fix rounds ran: 6 of 6 points of Done means passed; 2 are yours to confirm.
<!-- /ql:checked -->

## Try it yourself

### Try it

#### 1. The showcase plays and reads well
- Do: open `assets/kirokit-showcase.gif` (or the MP4) and watch it once without sound.
- Run: `open assets/kirokit-showcase.gif`
- Expect: about 35 s: title card; `kirokit install platform-team`; Kiro starting and `/jir` listing `/jira-search` and `/jira-triage`; the app cloning the repo and installing `platform-team` with the dock printing `installed platform-team into /tmp/kirokit-demo/home/.kiro`; the end card with both `brew install` lines. Every clip sits in a complete rounded frame with a numbered caption below.
- Saw: frames at each caption match, and the frame border is present at every 0.5 s sample — `evidence/it2-scenario-1-contact.png`, `evidence/it2-done-3-border-scan.txt`

#### 2. One command records it again
- Do: close any `task desktop:dev` (ports 34115 and 5173 must be free), then record.
- Run: `task showcase`
- Expect: about 50 s of `showcase: <step>: …` lines ending in `wrote assets/kirokit-showcase.gif (…) and assets/kirokit-showcase.mp4 (…)`; `/tmp/kirokit-demo` is gone afterwards.
- Saw: a fresh run of the judged tree published both files after all checks — `evidence/it2-fresh-run.txt`

#### 3. The recorder refuses to publish private details
- Do: put your home path into a file and run the privacy check on it.
- Run: `echo "$HOME" > /tmp/p.txt && scripts/showcase/checks.sh privacy /tmp/p.txt`
- Expect: exit 1 with `showcase: privacy: home path in /tmp/p.txt`.
- Saw: home path, git email and user name each fail; the public clone address and brew tap pass — `evidence/it2-scenario-2.txt`

### Not verified

- How the GIF looks on the rendered github.com README and whether the pacing feels right: only you can confirm, after the branch is pushed.
- Two recorder tests (busy-port preflight and dead-run takeover) skipped in the last suite run because another session's vite held port 5173; the takeover test passed earlier while the port was free — `evidence/it2-new-tests.txt`.

### Docs

- `README.md`: showcase under the intro; `/jir` sentence and screenshot alt text name `jira-search` (skill) and `jira-triage` (prompt).
- `CONTRIBUTING.md`: new "Showcase video" subsection.
- `AGENTS.md`: one layout line for `scripts/showcase/`.

## Accepted with open items

<!-- ql:accepted -->
None: the build passed its checks.
<!-- /ql:accepted -->

## Decided while building

<!-- ql:made-for-you -->
- `task showcase` is the one command, and `scripts/showcase/` holds everything it runs — it sits beside the existing `scripts/desktop-build.sh` and `task desktop:*` pattern. (Plan)
- File names `assets/kirokit-showcase.gif` and `assets/kirokit-showcase.mp4` — next to the existing `kirokit-cli.png` and `kirokit-desktop.png`. (Plan)
- `playwright-core` 1.63.0 pinned in `scripts/showcase/package.json` with a lockfile, `channel: 'chrome'` — uses the installed Chrome, so no browser download; 1.63.0 is older than two weeks, and Renovate keeps it current. (Plan)
- Terminal look: 110×32, `agg` theme `github-dark`, font size 16, typing at 70 ms per character, a neutral `$ ` prompt — readable at GIF size and free of the person's prompt. (Plan)
- Canvas 1280×800 plus a caption band for the MP4 (H.264, `crf` about 24, `+faststart`), and a GIF at 12 fps, about 1000 px wide, 128 colours — the spike's 37 s cut was 0.7 MB / 1.7 MB, so there is room for a sharper GIF. (Plan)
- The clone in the app uses the public `https://github.com/mavogel/kirokit.git`, the address the README tells users to use. (Plan)
- The GIF is shown with `width="800"` and links to the MP4, matching the README's centred `<p align="center"><img …>` style. (Plan)
- The approved verify command imported `playwright-core` from the repository root, where Node cannot resolve it (exit 1 above); it now runs `npm ls --prefix scripts/showcase playwright-core@1.63.0`, run by hand with exit 0 because QualityLayer only runs commands frozen at approval; nothing you approved changes. (slice 1)
- `tmux capture-pane` drops trailing spaces, so the driver waits for the prompt as `$`, and the session sets `BASH_SILENCE_DEPRECATION_WARNING=1` so macOS bash prints no zsh banner; nothing you approved changes. (slice 1)
- `agg` uses a 3 s idle limit instead of the spike's 2 s and the driver holds the `/jir` list 3.5 s, so the `jira-search` list stays readable for 3 s; nothing you approved changes. (slice 1)
- The privacy scan removes the public slugs `mavogel/kirokit` and `mavogel/tap/` before it looks for `$USER`, because the approved clone address and end card contain the GitHub owner, which equals the local user name; every other occurrence still fails; nothing you approved changes. (slice 2)
- `app-dock.txt` is sampled every 200 ms over the whole run (page text, dock rows, input values) instead of read once at the end, because dock lines such as the `git clone` line scroll away; half-typed copies of the public clone address are skipped, since they would otherwise trip the user-name check; nothing you approved changes. (slice 2)
- The dev server runs with `KIRO_HOME`, `KIROKIT_REPO`, `SOPS_AGE_KEY_FILE` and the `XDG_*` folders unset, so neither the app nor its setup dialog can reach or show a real path; nothing you approved changes. (slice 2)
- One `frame.png` (clear inside the rounded rectangle, `#13111f` in the corners, 1 px `#3a3450` border) laid over each clip replaces the `alphamerge` mask plus separate border, giving the same picture with one `overlay`; cue times go to a shell-sourceable `cues.env` instead of `cues.json`; nothing you approved changes. (slice 2)
- The title card fades into the CLI clip and the app clip into the end card with 0.4 s fades, besides the approved 0.5 s crossfade between the clips; the CLI clip is padded to the frame with the terminal's own background `#171b21` instead of being scaled; nothing you approved changes. (slice 2)
- `record.sh` calls `checks.sh` as a command rather than sourcing it, so each check stays runnable on its own with the same messages; nothing you approved changes. (slice 2)
- T11 also adds a one-line scripts/showcase pointer to AGENTS.md's layout, so agents find task showcase. (slice 3)
- Task showcase leaves the project check list: every run records new media into assets/, so QualityLayer records it as 'files changed while the check ran' and never as fresh proof. It stays proved by the T8 slice check and Scenarios 1-3. (the Plan, while building)
- Verification round 1 is rerun because the files that changed after it are its fix (fresh-page rendering in `caption.mjs`, the run-ownership guard in `record.sh`, the clearer `checks.sh length` message, the reviewer's tests plus one regression test, and the media re-recorded without the torn frame), which can change its D1, Done 3 and Scenario 1 verdicts; the pre-commit whitespace fixers also now skip `docs/plans/*/evidence/` so captured proof is committed byte-for-byte; nothing you approved changes. (the Plan, while building)
- Verification round 2 was recorded while 1 file changed, 1 of them code the run exercised; rerun it or keep it and say why here; nothing you approved changes. (the Plan, while building)
<!-- /ql:made-for-you -->

## Settled with you

<!-- ql:review-items -->
- Still awaiting your answer: Two recorder tests broke when an unrelated dev server held port 5173. Their port probe only bound 127.0.0.1, so a vite on `[::1]:5173` slipped past it and `record.sh`'s own lsof preflight then failed them; the probe now asks lsof the same way `record.sh` does, and the two tests skip instead (test-only, unstaged).
- Still awaiting your answer: The GIF still shows a Kiro session resume id. The CLI clip ends on `Resume with: kiro-cli --resume-id sess_…` for an empty chat session. It is not a credential, but the privacy check does not consider it.
- Still awaiting your answer: The recorder tests are still outside `task ci` and CI. They run with `node --test scripts/showcase/`, as CONTRIBUTING says.
- Still awaiting your confirmation: The pacing feels right.
- Still awaiting your confirmation: It looks right on the rendered GitHub page.
<!-- /ql:review-items -->

## Ship

<!-- ql:ship -->
Not shipped yet. Branch `feat/make-30-60s-video-how`.
<!-- /ql:ship -->
