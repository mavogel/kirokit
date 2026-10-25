---
answers: app
title: Showcase video of kirokit CLI and app for the README
request: "make a 30-60s video /showcase of how kirokit (cli and app) works and how to use it, which i can use in the github readme"
type: feature
depth: light
level: engineering
ui: false
overrides: []
stage: review
iterations: 2
verification_failures: 0
gate: null
created: 1792920764667
startup_owner: 05d9f7946b007a3e
layout: seven
profile: full
status: approved
approved: true
review_digest: c80b954b03a0e6af52851af7bc4cdc4f035bb837053fca444ffef100e3d785f6
review_basis: eb80c35fbca3aaa62e52feebf61d64e10151fec40679bd9c8c517a0d01da67c0
branch: feat/make-30-60s-video-how
base: 87afcb06ba72f010cc64b27864d0edcf93c65ae1
base_ref: main
judge_runs: 0
---

# Showcase video of kirokit CLI and app for the README

## Your request

> make a 30-60s video /showcase of how kirokit (cli and app) works and how to use it, which i can use in the github readme


## Problem

Someone landing on the kirokit GitHub page today has to read a long README and two still screenshots to understand what kirokit does and how to use it. There is no quick way to *see* the CLI and the desktop app working, so visitors who would try it in a minute leave before reaching the Quickstart.

## Done means

1. **A 30–60 second showcase autoplays at the top of the GitHub README.** Opening the repository on github.com shows an animated GIF right under the intro, above Quickstart, and it loops by itself; the two existing screenshots stay where they are (only you can confirm: it looks right on the rendered GitHub page).
2. **The CLI part shows a team's config installed and its agent starting.** The viewer sees `kirokit install platform-team` run into a temporary Kiro home, then `kiro-cli chat --agent devops --v3` start; typing `/jir` lists the team's `jira-search` skill and `jira-triage` prompt.
3. **The app part shows the same job done with buttons.** The viewer sees the desktop app clone the config repository, press **Install** on a team, and the terminal dock run `kirokit install`; it is the app's real UI and backend (Wails dev mode) recorded in a browser tab inside a plain rounded frame.
4. **A viewer can follow every step without sound.** A short caption names each step, and nothing passes too fast to read (only you can confirm: the pacing feels right).
5. **No private details are on screen.** No email address, real home path, token or decrypted secret appears in any frame.
6. **The video can be made again with one command when the UI changes.** The recording scripts are in the repository, record the real CLI and app, and the README or CONTRIBUTING guide names the one command that rebuilds both files.
7. **The README GIF is small enough for GitHub to show it reliably.** It stays under about 10 MB and loads on the repository page.
8. **A sharp MP4 of the same showcase sits beside the GIF.** `assets/` holds a committed MP4 (about 4 MB at most) made from the same recording, ready for release notes, social posts or a later upload through GitHub's web editor.

## Scope

### In this task

- The showcase as a README GIF plus an MP4 master, both in `assets/`.
- The scripts and steps that record the CLI and the app, and stitch them into one video.
- The README change that embeds the video, and a note on how to re-record it.

### Not in this task

- Any change to how the `kirokit` CLI or the desktop app behaves.
- Voice-over or music.
- Posting the video to other sites, or uploading it through GitHub's web editor on your behalf.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | Which file format should the README show? | Both: GIF in the README, MP4 kept as the sharp master |
| 2 | How should the video be produced? | Scripted real recordings, re-recordable with one command |
| 3 | Which story should the 30-60 seconds tell? | User path: CLI install of a team, agent starts in Kiro, then the same in the app |
| 4 | Where in the README should it go? | At the top, under the intro; keep the two screenshots |
| 5 | Text on screen? | Short captions per step (e.g. "1 · install a team") |
<!-- /ql:decided -->

<!-- optional section here, only when there is content: "First release" -->

<!-- optional section here, only when there is content: "What the user sees" -->

<!-- optional section here, only when there is content: "Launch" -->

<!-- optional section here, only when there is content: "Not doing" -->

<!-- optional section here, only when there is content: "Worth it when" -->

<!-- optional section here, only when there is content: "Reference" -->

<!-- optional section here, only when there is content: "Added while building" -->

## Context pointers

- "which i can use in the github readme" — `README.md` on github.com/mavogel/kirokit.

## First look at the code

- `README.md:10-71` — Quickstart: the CLI flow (`kirokit setup`, `kirokit install platform-team`, `kiro-cli chat --agent devops --v3`) and the desktop flow (Clone, Install, terminal dock); the story the video should follow.
- `README.md:43` and `README.md:59` — the two existing stills, `assets/kirokit-cli.png` (1411×663) and `assets/kirokit-desktop.png` (2358×1576).
- `cmd/kirokit/commands/setup.go:24` and `cmd/kirokit/commands/install.go:11` — the commands the CLI part runs.
- `desktop/main.go:19-23` — app window is 1180×760 by default, minimum 720×480; sets the capture frame.
- `desktop/frontend/src/pages/welcome.js:3` (first-launch Clone), `desktop/frontend/src/pages/teams.js:22` (Install button), `desktop/frontend/src/terminal.js` (terminal dock) — the screens the app part shows.
- `Taskfile.yml` — `desktop:dev` runs the app from source via `go tool wails dev`; a `demo:*` task would sit next to it.
- `demo-runbook-v1.md` (gitignored, repo root) — the existing live-demo script with a scratch `KIRO_HOME`; a source of safe demo steps.
- Tools on this machine: `asciinema`, `agg`, `ffmpeg` installed; `vhs` not installed (available via Homebrew); `kiro-cli` and `sops` installed and logged in.

*Untraced areas:* how the app behaves on first launch from a clean settings folder, what the Install prerequisite check shows without a SOPS key, and how to capture the native Wails window from a script.
