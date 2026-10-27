---
answers: app
title: "Support Quick, Kiro IDE and kirocrew; README refresh"
request: "support amazon quick (see docs/research/quick), kiro-ide and kirocrew (which also read KIRO_HOME). update the readme with a new intro, gif, video and make it more concise like https://github.com/maxritter/qualitylayer/blob/main/README.md"
type: feature
depth: light
level: engineering
ui: true
overrides: []
stage: outline
iterations: 0
verification_failures: 0
gate: null
created: 1793107388975
startup_owner: 0bb4ab9c6cb024b3
layout: seven
profile: full
status: approved
approved: true
review_digest: 20495fed3b72b17390eb79baa9eb7326c5f4411882b1d483c99626ea4bd06b95
review_basis: 65fa65de7534f75025638d586c6fbf7cf36546ac5fc8d11b6ef59a2448372c53
branch: feat/support-amazon-quick-see-docs
base: 9bfb355bc8dae4d2813236764bee22f18102d3f9
base_ref: main
---

# Support Quick, Kiro IDE and kirocrew; README refresh

## Your request

> support amazon quick (see docs/research/quick), kiro-ide and kirocrew (which also read KIRO\_HOME). update the readme with a new intro, gif, video and make it more concise like https://github.com/maxritter/qualitylayer/blob/main/README.md


## Problem

kirokit only speaks to Kiro CLI v3 today. A team that also works in Amazon Quick has to rebuild its skills and agents by hand in the Quick app, with nothing reproducible or reviewed. Kiro IDE and kirocrew read the same `KIRO_HOME` that `kirokit install` writes into, but nobody has checked what they pick up from it, and the README does not say. The README itself has grown to 359 lines of reference material: a newcomer has to read far before seeing what kirokit is, and its intro and showcase speak only about Kiro CLI.

## Done means

1. **A team's chosen skills and agents render into an importable Quick bundle.** `kirokit render --quick <team>` writes one folder per skill the team lists for Quick in `manifest.yaml`, and one JSON file per agent it lists there; each agent's prompt, name and description come from its existing Kiro agent, with Quick-only fields (welcome message, starter prompts, pinned skills) in the manifest. Each file imports through Quick's own "From folder" / "From file" screens. In the desktop app, a button on the team renders the same bundle and opens its folder.
2. **Quick skill ids survive a reimport.** A helper reads the skills exported from Quick and records their ids in a gitignored file; the next render puts each skill's `id:` line back and fills each agent's `skills` list, so overwriting in Quick keeps the ids.
3. **Quick output is checked like the Kiro output.** Golden files for the Quick bundle are compared byte for byte in CI, using a committed fixture ids file and never a real account's ids.
4. **Kiro IDE and kirocrew use what `kirokit install` puts in `KIRO_HOME`.** After an install, the team's configuration shows up in Kiro IDE and in kirocrew as far as each tool supports it; small output gaps that stop a tool from picking it up are fixed, larger ones are listed, and the README says what each one picks up.
5. **The README opens with a new, short intro that names every supported Kiro tool and Quick.** A reader sees what kirokit is, a showcase and how to install it before any reference material; the README is noticeably shorter, in the style of the QualityLayer README. Its reference sections (manifest reference, SOPS, sync, desktop details and the like) move into linked `docs/` pages, so nothing is lost and every link and anchor still resolves.
6. **A new showcase GIF and video show the current product.** The recording covers today's CLI install and desktop app story, `kirokit render --quick` with its bundle, and Kiro IDE and kirocrew picking up the installed team. The GIF sits at the top of the README and links to the sharper video.
7. **The Quick round trip works on a real account.** A rendered skill and agent import into Quick, and a re-exported skill equals the rendered file apart from its `id:` line (only you can confirm: the import in your Quick account).

## Scope

### In this task

- Amazon Quick as a new output of `kirokit render`, plus the ids helper and its gitignored ids file.
- A desktop app button that renders a team's Quick bundle and opens its folder.
- Checking and, where needed, adjusting what Kiro IDE and kirocrew pick up from an installed `KIRO_HOME`.
- A shorter README with a new intro, GIF and video; affected `CONTRIBUTING.md` and `AGENTS.md` lines.

### Not in this task

- Installing into Quick automatically: Quick has no API, CLI or local folder for agents and skills, so import stays manual.
- Quick connectors, spaces, routines, sharing and publishing, and any secret in a Quick file.
- Kiro web.

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | What should support for Kiro IDE and kirocrew mean? | Check, fix small gaps, document |
| 2 | Where do the Quick agents come from? | Derived from the Kiro agents |
| 3 | Which skills go into the Quick bundle? | Opt-in list per team |
| 4 | Should the desktop app get the Quick bundle too? | Also an app button |
| 5 | What should the new GIF and video show? | Also Kiro IDE and kirocrew |
| 6 | Where does the README's reference detail go? | Move to docs/ pages |
<!-- /ql:decided -->

<!-- optional section here, only when there is content: "First release" -->

<!-- optional section here, only when there is content: "What the user sees" -->

<!-- optional section here, only when there is content: "Launch" -->

<!-- optional section here, only when there is content: "Not doing" -->

<!-- optional section here, only when there is content: "Worth it when" -->

<!-- optional section here, only when there is content: "Reference" -->

<!-- optional section here, only when there is content: "Added while building" -->

## Context pointers

- `docs/research/quick/QUICK-TARGET-BRIEF.md`: the Quick research brief (findings, proposed two-pass flow, open questions); evidence files beside it in `docs/research/quick/`.
- https://github.com/maxritter/qualitylayer/blob/main/README.md: the README style to follow.
- https://docs.aws.amazon.com/quick/latest/userguide/agents-desktop.html, https://docs.aws.amazon.com/quick/latest/userguide/skills-desktop.html, https://amazon-quick.github.io/Amazon-Quick-official-catalog/ (from the brief).
- Installed locally: Kiro IDE 1.4.26 (`/Applications/Kiro.app`, `kiro`), kirocrew 0.10.0 (`/Applications/KiroCrew.app`, `kirocrew`), kiro-cli 2.34.2, Amazon Quick desktop 1.1.4053 (from the brief).

## First look at the code

- Render entry: `cmd/kirokit/commands/render.go:42` (`--out` is the only flag today); `internal/render/render.go:45` `RenderTeamWith`, `internal/render/render.go:243` `teamInventory`.
- Skills copied byte for byte, outside the template engine (the path Quick's `{{input}}` placeholders need): `internal/render/externalskills.go:28` `writeExternalSkills`.
- Agent resources and validation: `internal/render/globalsteerings.go:76` `addResources`, `internal/kiroconfig/validate.go:33` `ValidateAgentJSON`.
- Where installs land: `internal/install/install.go:47` `ResolveKiroHome` (`$KIRO_HOME`, else `~/.kiro`); the desktop app's Kiro home field `desktop/app.go:36`.
- kirocrew shares the base: its `config/paths.py` keeps its data in `~/.kiro/crew` and treats `~/.kiro/agents` and `~/.kiro/settings/mcp.json` as kiro-cli's (lines 734, 757 in the installed app); `~/.kiro/agents` already holds kirocrew's own `kirocrew*.json` beside kirokit's agents.
- README and showcase: `README.md` (359 lines, sections from line 13 to 338), `scripts/showcase/record.sh`, `assets/kirokit-showcase.gif` and `.mp4`.

*Untraced areas:* what Kiro IDE reads from `~/.kiro` (agents, skills, steering, prompts, MCP); how kirocrew loads agents and skills; how the golden harness (`internal/golden`) would take a second output tree.
