---
type: outline
---

# Outline: Showcase video of kirokit CLI and app for the README

## Summary

Three slices and eleven tasks: first the CLI clip records itself from one script, then the app clip, captions, composition and the safety checks turn it into the finished GIF and MP4, and last the README embeds it and the hooks stop corrupting committed media. The risky parts are the outside tools the recording drives (Kiro's full-screen chat, a GitHub clone, headless Chrome) and keeping private details out of a public file.

## How the build runs

<!-- ql:waves -->
- Wave 1 · Slice 1 (The CLI part records itself from one script) · first
- Wave 2 · Slice 2 (One command makes the checked GIF and MP4) · on its own
- Wave 3 · Slice 3 (The README opens with the showcase and commits keep media intact) · on its own

3 waves · longest path 11 of 11 tasks

A fresh agent runs the scenarios after slice 2 (drives a GitHub clone and headless Chrome, and links the person's age key into a throwaway home.) before the build goes on.
<!-- /ql:waves -->

## How each point is proved

<!-- ql:proof-map -->
- Done means 1 · A 30–60 second showcase autoplays at the top of the GitHub README → "The README opens with the showcase" · runs after slice 3
- Done means 2 · The CLI part shows a team's config installed and its agent starting → "One command records the real CLI and app into both files" · runs after slice 2
- Done means 3 · The app part shows the same job done with buttons → "One command records the real CLI and app into both files" · runs after slice 2
- Done means 4 · A viewer can follow every step without sound → "The README opens with the showcase" · runs after slice 3
- Done means 5 · No private details are on screen → "Private details stop the run before anything is published" · runs after slice 2; "An interrupted run leaves no trace" · runs after slice 2
- Done means 6 · The video can be made again with one command when the UI changes → "One command records the real CLI and app into both files" · runs after slice 2; "An interrupted run leaves no trace" · runs after slice 2; "The README opens with the showcase" · runs after slice 3
- Done means 7 · The README GIF is small enough for GitHub to show it reliably → "One command records the real CLI and app into both files" · runs after slice 2; "Committing the media keeps its bytes" · runs after slice 3
- Done means 8 · A sharp MP4 of the same showcase sits beside the GIF → "One command records the real CLI and app into both files" · runs after slice 2; "Committing the media keeps its bytes" · runs after slice 3
<!-- /ql:proof-map -->

## Oracle

- **Oracle:** after `task showcase`, `assets/kirokit-showcase.gif` plays 30-60 s on the rendered GitHub README. Frames show `kirokit install platform-team`, Kiro's `/jir` completion listing `jira-search`, and the app's dock printing `installed platform-team into …`, and none show the person's home path or user name.
- **Truth:** a failed or interrupted run leaves `assets/` byte-identical, removes `/tmp/kirokit-demo`, and leaves the real age key file untouched.

## Decided while outlining

- The privacy scan removes the two public project slugs `mavogel/kirokit` and `mavogel/tap/` before it looks for `$USER`, because the approved clone address and end card contain the GitHub owner, which is also the local user name; every other occurrence still fails the run (recorded with `plan amend`).
- The privacy, size and length checks live in `scripts/showcase/checks.sh`, sourced by `record.sh` and runnable on their own (`checks.sh privacy <file>…`), so the failure path can be proved without a full recording (recorded with `plan amend`).
- `caption.mjs` also computes the caption cue times (`cues.env`, sourced by `record.sh`): it replays `cli.cast` with the same 3 s idle limit `agg` applies, so cue times match the rendered CLI clip, and reads `app-marks.json` for the app clip.
- The app clip cuts out the clone wait beyond 1.5 s using `app-marks.json`, so a slow network cannot stretch the cut.

## Slice 1: The CLI part records itself from one script

After this slice, `scripts/showcase/record.sh --keep` builds the CLI, types the install-and-chat story into a real Kiro session and leaves `cli.cast` and `cli.gif` in `/tmp/kirokit-demo/out`. A missing tool or a changed Kiro screen fails loudly with the step name.

```diff
 scripts/
+└── showcase/
+    ├── package.json        + private module, playwright-core 1.63.0 pinned
+    ├── package-lock.json   + lockfile for npm ci
+    ├── .gitignore          + node_modules/
+    ├── tmux.conf           + status off, UTF-8, no prompt noise
+    ├── cli-drive.sh        + types the CLI story, waits on screen text
+    └── record.sh           + preflight, trap, build, CLI recording
```

- [x] T1 The showcase toolkit installs from a lockfile
  - **Objective:** add the `scripts/showcase/` package (private ES module, `playwright-core` 1.63.0 as a dev dependency, committed lockfile), its `.gitignore` and the `tmux.conf` the CLI session uses.
  - **Design:** none
  - **Files:** `scripts/showcase/package.json`, `scripts/showcase/package-lock.json`, `scripts/showcase/.gitignore`, `scripts/showcase/tmux.conf`
  - **Trivial:** `npm ci --prefix scripts/showcase` installs exactly the locked `playwright-core` and `git check-ignore scripts/showcase/node_modules` matches.
  - **Definition of done:**
    - `git status` shows no `node_modules` after the install.
    - verify: `npm ci --prefix scripts/showcase && npm ls --prefix scripts/showcase playwright-core@1.63.0`

- [x] T2 `cli-drive.sh` types the CLI story and fails on a changed screen
  - **Objective:** drive a tmux session through install, chat start, trust, `/jir`, Esc, `/quit` one character at a time, waiting on screen text before each step, and never pressing Enter on a chat message.
  - **Design:** "`cli-drive.sh` contract"
  - **Files:** `scripts/showcase/cli-drive.sh`
  - **Test first:** start a tmux session whose `PATH` has no `kirokit`, run `cli-drive.sh` against it; it must exit 1 with `timeout waiting for installed platform-team` (today the script does not exist).
  - **Definition of done:**
    - Against a session with the fresh build and a throwaway `KIRO_HOME`, it reaches `Session ended` and exits 0.
    - A missing text exits 1 naming the awaited text within the 30 s limit.
    - verify: `bash -n scripts/showcase/cli-drive.sh && shellcheck scripts/showcase/cli-drive.sh`

- [x] T3 `record.sh` checks its tools, builds the CLI and records the CLI clip
  - **Objective:** the one entry script: preflight (tools, age key, free ports 34115 and 5173), the `EXIT` trap that stops only its own PIDs and removes `/tmp/kirokit-demo`, the two `go build`s, and the asciinema recording rendered by `agg`; `--keep` keeps the work folder.
  - **Design:** "`task showcase` runs one script that leaves only the two asset files behind", "`cli-drive.sh` contract"
  - **Files:** `scripts/showcase/record.sh`
  - **Test first:** with `agg` hidden from `PATH`, `record.sh` must exit non-zero with `showcase: preflight: agg not found` and leave no `/tmp/kirokit-demo` (today the script does not exist).
  - **Definition of done:**
    - `record.sh --keep` leaves `/tmp/kirokit-demo/out/cli.cast` and `cli.gif`; the cast holds `installed platform-team` and `jira-search` and no model reply.
    - A busy port 34115 fails preflight with `close the running wails dev first`.
    - Without `--keep`, `/tmp/kirokit-demo` is gone after the run.
    - verify: `shellcheck scripts/showcase/record.sh && scripts/showcase/record.sh --keep`

**Validation:** automated: `shellcheck scripts/showcase/*.sh`, `scripts/showcase/record.sh --keep` · manual: none

## Slice 2: One command makes the checked GIF and MP4

**Depends on:** Slice 1

**Risky:** drives a GitHub clone and headless Chrome, and links the person's age key into a throwaway home.

After this slice, `task showcase` records the app in Wails dev mode too, renders captions and frames, composes the MP4 and the GIF, and only writes them to `assets/` when the privacy, size and length checks pass. Any failure leaves `assets/` untouched.

```diff
 scripts/showcase/
+├── app-drive.mjs   + clicks Clone and Install in Chrome, records video, dock text, marks
+├── caption.mjs     + caption and frame PNGs, cue times
+├── checks.sh       + privacy, size and length checks
 └── record.sh       ~ app step, compose, checks, move to assets/
 Taskfile.yml        ~ showcase
```

- [x] T4 `app-drive.mjs` clicks the app story and records it
  - **Objective:** in headless Chrome at 1180×760, type the clone address, press Clone, press Install on `platform-team` and in the setup dialog, wait for the `installed` badge, and write `app.webm`, `app-dock.txt` and `app-marks.json`.
  - **Design:** "`app-drive.mjs` contract"
  - **Files:** `scripts/showcase/app-drive.mjs`
  - **Test first:** `node scripts/showcase/app-drive.mjs <tmp>` with no dev server running must exit non-zero naming the first step (`welcome`), because nothing serves `http://localhost:34115`.
  - **Definition of done:**
    - Against a running dev server it writes the three files; `app-dock.txt` holds `installed platform-team into`.
    - Any step timeout exits non-zero naming the step.
    - A replace dialog, shown when another team is installed, is confirmed rather than hanging.
    - verify: `node --check scripts/showcase/app-drive.mjs`

- [x] T5 `record.sh` runs the app in Wails dev mode under a throwaway home
  - **Objective:** start `go tool wails dev` in `desktop/` with `HOME=/tmp/kirokit-demo/home` and the real Go and npm caches, link (never copy) the age key, wait for the DevServer URL, run `app-drive.mjs`, then stop that PID.
  - **Design:** "`task showcase` runs one script that leaves only the two asset files behind"
  - **Files:** `scripts/showcase/record.sh`
  - **Test first:** `record.sh --keep` today stops after the CLI clip, so `/tmp/kirokit-demo/out/app.webm` is missing.
  - **Definition of done:**
    - `--keep` leaves `app.webm`, `app-dock.txt`, `app-marks.json`.
    - After the run no `wails dev` process started by the script survives, the link is gone with the folder, and the real key file's inode and modification time are unchanged.
    - A dev server that never prints the URL fails with `showcase: app: …` after 180 s.
    - verify: `scripts/showcase/record.sh --keep && ls /tmp/kirokit-demo/out/app.webm`

- [x] T6 `caption.mjs` renders captions, frames and cue times
  - **Objective:** render the four captions plus the title and end cards as transparent PNGs, the 1180×760 rounded mask and border, and `cues.env` from `cli.cast` (with the 3 s idle limit) and `app-marks.json`.
  - **Design:** "Composition contract", "Captions sit in their own band below each clip and are rendered by Chrome"
  - **Files:** `scripts/showcase/caption.mjs`
  - **Test first:** `node scripts/showcase/caption.mjs <out>` on the kept recordings must write `cap-1.png`…`cap-4.png`, `mask.png`, `border.png` and `cues.env`; today it does not exist.
  - **Definition of done:**
    - Each caption PNG is 1280×96 with transparency; the cue for caption 2 falls after `installed platform-team` in the CLI clip.
    - A missing input file exits non-zero naming it.
    - verify: `node scripts/showcase/caption.mjs /tmp/kirokit-demo/out && ffprobe -v error -show_entries stream=width,height -of csv=p=0 /tmp/kirokit-demo/out/cap-1.png`

- [x] T7 `record.sh` composes the MP4 and the GIF
  - **Objective:** join title card, framed CLI clip, 0.5 s crossfade, framed app clip with the clone wait trimmed, and end card into a 1280×896 H.264 MP4 with captions in their own band, then derive the 12 fps GIF.
  - **Design:** "Composition contract", "Captions sit in their own band below each clip and are rendered by Chrome"
  - **Files:** `scripts/showcase/record.sh`
  - **Test first:** `record.sh --keep` today produces no `/tmp/kirokit-demo/out/showcase.mp4`.
  - **Definition of done:**
    - `showcase.mp4` is 1280×896, `yuv420p`, 30-60 s; `showcase.gif` is about 1000 px wide.
    - Each caption is on screen at least 2.5 s.
    - verify: `ffprobe -v error -show_entries stream=width,height,pix_fmt:format=duration -of default=nw=1 /tmp/kirokit-demo/out/showcase.mp4`

- [x] T8 Checks gate the move into `assets/`, and `task showcase` runs it all
  - **Objective:** `checks.sh` scans the cast, dock text and caption text for `$HOME`, `$USER` (after removing the public slugs) and the git email, and checks GIF ≤ 8 MB, MP4 ≤ 4 MB and length 30-60 s; `record.sh` moves the files into `assets/` only after all pass; `Taskfile.yml` gains `showcase`.
  - **Design:** "The script refuses to publish private details, oversize files or a wrong length", "`task showcase` runs one script that leaves only the two asset files behind"
  - **Files:** `scripts/showcase/checks.sh`, `scripts/showcase/record.sh`, `Taskfile.yml`
  - **Test first:** `scripts/showcase/checks.sh privacy <file holding $HOME>` must exit 1 with `showcase: privacy: home path in <file>`; today the script does not exist.
  - **Definition of done:**
    - A file with only `github.com/mavogel/kirokit.git` and `mavogel/tap/kirokit` passes; one with `$USER` elsewhere, `$HOME` or the git email fails naming the file.
    - An oversize or wrong-length file fails `checks.sh size` / `checks.sh length` and leaves `assets/` unchanged.
    - verify: `shellcheck scripts/showcase/*.sh && task showcase`

**Validation:** automated: `shellcheck scripts/showcase/*.sh`, `node --check scripts/showcase/app-drive.mjs`, `node --check scripts/showcase/caption.mjs`, `task showcase` · manual: none

## Slice 3: The README opens with the showcase and commits keep media intact

**Depends on:** Slice 2

After this slice, the README shows the looping GIF right under the intro with a link to the MP4, CONTRIBUTING names `task showcase`, and the pre-commit fixers leave GIF and MP4 bytes alone.

```diff
 assets/
+├── kirokit-showcase.gif   + README showcase
+└── kirokit-showcase.mp4   + H.264 master
 lefthook.yml               ~ fixers skip *.gif, *.mp4
 README.md                  ~ showcase under the intro, jira-search wording
 CONTRIBUTING.md            ~ Showcase video
```

- [x] T9 The whitespace fixers skip GIF and MP4 files
  - **Objective:** add `**/*.gif` and `**/*.mp4` next to `**/*.png` in both fixer jobs and their header comment.
  - **Design:** "Committed media is protected from the whitespace hooks, and the README names the skill correctly"
  - **Files:** `lefthook.yml`
  - **Test first:** stage a scratch GIF and run `lefthook run pre-commit`; today its SHA-1 changes because both fixers rewrite it.
  - **Definition of done:**
    - With the excludes, the staged GIF's and MP4's SHA-1 are unchanged after the hook.
    - verify: `uvx --from pre-commit-hooks==6.0.0 check-yaml lefthook.yml`

- [x] T10 The recorded showcase is committed in `assets/`
  - **Objective:** run `task showcase` on the final scripts and keep the resulting `assets/kirokit-showcase.gif` and `assets/kirokit-showcase.mp4`.
  - **Design:** "`task showcase` runs one script that leaves only the two asset files behind"
  - **Files:** `assets/kirokit-showcase.gif`, `assets/kirokit-showcase.mp4`
  - **Qualification:** the files come from a passing `task showcase` run; frames extracted at each caption show the CLI install, the `/jir` list with `jira-search`, the app's Clone and the dock's `installed platform-team into`.
  - **Definition of done:**
    - GIF ≤ 8 MB, MP4 ≤ 4 MB, both 30-60 s.
    - verify: `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4`

- [x] T11 README and CONTRIBUTING show and explain the showcase
  - **Objective:** embed the GIF under the intro (above Quickstart) with `width="800"` linking to the MP4, correct the `jira-triage` "example skill" wording and alt text to `jira-search`, and add a "Showcase video" subsection to CONTRIBUTING naming the tools and `task showcase`.
  - **Design:** "Committed media is protected from the whitespace hooks, and the README names the skill correctly"
  - **Files:** `README.md`, `CONTRIBUTING.md`, `AGENTS.md`
  - **Trivial:** `grep -n 'kirokit-showcase.gif' README.md` sits between the intro paragraph and `## Quickstart`, and `grep -n 'task showcase' CONTRIBUTING.md` matches.
  - **Definition of done:**
    - The two existing screenshots stay in place.
    - verify: `grep -c 'assets/kirokit-showcase' README.md`

**Validation:** automated: `lefthook run pre-commit`, `go test ./...`, `node scripts/sync-agent-assets.mjs --check` · manual: the GIF loops on the rendered GitHub README and the pacing feels right

## Program design

The Plan's contracts hold the design.

## Shape of the code

| Concept | Lives in | Reuses |
|---|---|---|
| Recording entry, trap, compose | `scripts/showcase/record.sh` | `scripts/desktop-build.sh` shape (`set -euo pipefail`, guarded trap) |
| CLI driver | `scripts/showcase/cli-drive.sh` | spike 1 tmux route |
| App driver | `scripts/showcase/app-drive.mjs` | app selectors in `desktop/frontend/src/pages/` |
| Captions, frames, cues | `scripts/showcase/caption.mjs` | spike 3 Chrome screenshot of an HTML pill |
| Publish checks | `scripts/showcase/checks.sh` | none |

## End-to-end scenarios

### Scenario 1: One command records the real CLI and app into both files

**Runs after:** slice 2

**Proves:** Done means 2, 3, 6, 7, 8

| Step | Action | Expected |
|---|---|---|
| 1 | Run `task showcase` from the repository root | exits 0; the log names each step reached |
| 2 | Run `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4` | a value between 30 and 60 |
| 3 | Run `stat -f%z assets/kirokit-showcase.gif assets/kirokit-showcase.mp4` | GIF ≤ 8388608 bytes, MP4 ≤ 4194304 bytes |
| 4 | Extract a frame at the caption 1 and 2 cues with `ffmpeg -ss <t> -i assets/kirokit-showcase.gif -frames:v 1` | the CLI frame shows `kirokit install platform-team`, then the `/jir` list with `jira-search` and `jira-triage` |
| 5 | Extract frames at the caption 3 and 4 cues | the app frame shows the Clone form, then the dock printing `installed platform-team into` inside a rounded frame |
| 6 | Run `ls /tmp/kirokit-demo` | no such file or directory |

### Scenario 2: Private details stop the run before anything is published

**Runs after:** slice 2

**Proves:** Done means 5

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Write a scratch file holding the output of `echo "$HOME"` and run `scripts/showcase/checks.sh privacy <that file>` | exits 1 with `showcase: privacy: home path in <file>` |
| 2 | Write a scratch file holding only `git clone https://github.com/mavogel/kirokit.git` and `brew install --cask mavogel/tap/kirokit`, and run the same check | exits 0 |
| 3 | Write a scratch file holding the output of `git config user.email` and run the check | exits 1 naming the file |
| 4 | After Scenario 1, run the privacy check on the kept `cli.cast`, `app-dock.txt` and caption text from a `record.sh --keep` run | exits 0 |

### Scenario 3: An interrupted run leaves no trace

**Runs after:** slice 2

**Proves:** Done means 5, 6

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Note `stat -f '%i %m'` of the age key file and `shasum assets/*` | values recorded |
| 2 | Start `scripts/showcase/record.sh` in the background and send it SIGINT while the app step runs | the script exits non-zero |
| 3 | Run `ls /tmp/kirokit-demo`, `pgrep -f 'wails dev'` and repeat step 1 | folder gone, no surviving dev server, key inode and time unchanged, `assets/` unchanged |

### Scenario 4: Committing the media keeps its bytes

**Runs after:** slice 3

**Proves:** Done means 7, 8

**Screens:** none

| Step | Action | Expected |
|---|---|---|
| 1 | Record `shasum assets/kirokit-showcase.gif assets/kirokit-showcase.mp4`, stage both and run `lefthook run pre-commit` | the hook passes |
| 2 | Run the same `shasum` again | identical values |

### Scenario 5: The README opens with the showcase

**Runs after:** slice 3

**Proves:** Done means 1, 4, 6

| Step | Action | Expected |
|---|---|---|
| 1 | Open `README.md` | the showcase GIF (width 800, linked to the MP4) sits under the intro and above `## Quickstart`; both screenshots remain |
| 2 | Read the sentence under the CLI screenshot | it names `jira-search` as the skill and `jira-triage` as a prompt |
| 3 | Open `CONTRIBUTING.md` | a "Showcase video" subsection names the tools and `task showcase` |
| 4 | View the README on github.com after pushing (only you can confirm) | the GIF loops by itself, each caption is readable before it changes |

## Commands the build runs

<!-- ql:commands -->
- `shellcheck scripts/showcase/*.sh` · slice 1, slice 2
- `scripts/showcase/record.sh --keep` · slice 1
- `node --check scripts/showcase/app-drive.mjs` · slice 2, T4
- `node --check scripts/showcase/caption.mjs` · slice 2
- `task showcase` · slice 2
- `lefthook run pre-commit` · slice 3, project checks
- `go test ./...` · slice 3, project checks
- `node scripts/sync-agent-assets.mjs --check` · slice 3, project checks
- `npm ci --prefix scripts/showcase && npm ls --prefix scripts/showcase playwright-core@1.63.0` · T1
- `bash -n scripts/showcase/cli-drive.sh && shellcheck scripts/showcase/cli-drive.sh` · T2
- `shellcheck scripts/showcase/record.sh && scripts/showcase/record.sh --keep` · T3
- `scripts/showcase/record.sh --keep && ls /tmp/kirokit-demo/out/app.webm` · T5
- `node scripts/showcase/caption.mjs /tmp/kirokit-demo/out && ffprobe -v error -show_entries stream=width,height -of csv=p=0 /tmp/kirokit-demo/out/cap-1.png` · T6
- `ffprobe -v error -show_entries stream=width,height,pix_fmt:format=duration -of default=nw=1 /tmp/kirokit-demo/out/showcase.mp4` · T7
- `shellcheck scripts/showcase/*.sh && task showcase` · T8
- `uvx --from pre-commit-hooks==6.0.0 check-yaml lefthook.yml` · T9
- `ffprobe -v error -show_entries format=duration -of csv=p=0 assets/kirokit-showcase.mp4` · T10
- `grep -c 'assets/kirokit-showcase' README.md` · T11
- `golangci-lint run` · project checks
- `go vet ./...` · project checks
- `task build` · project checks
<!-- /ql:commands -->
