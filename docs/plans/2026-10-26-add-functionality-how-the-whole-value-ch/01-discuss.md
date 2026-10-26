---
answers: app
title: Cross-role agent harness with feedback loop
request: Add functionality how the whole value chain of builders works together. build a proper harness and think if feedback from non-engineers to e.g. skills
type: feature
depth: product
level: engineering
ui: true
overrides: []
stage: verify
iterations: 0
verification_failures: 0
gate: null
created: 1793012565552
startup_owner: a1b59536b4676690
layout: seven
profile: full
status: approved
approved: true
review_digest: f6a6ff58a30518e6aa1e36b50f8da5edd3eb82f1832141f9d99d9853cfc97a88
review_basis: f04128d75568416222744db8abf20eb943051d413c233a213bf5bb1cb4a54276
branch: feat/add-functionality-how-whole-value
base: 3dcf20fa1669f956c07530487954be10184cee27
base_ref: main
---

# Cross-role agent harness with feedback loop

## Your request

> Add functionality how the whole value chain of builders works together. build a proper harness and think if feedback from non-engineers to e.g. skills


## Problem

kirokit shares a team's Kiro setup (agents, skills, steerings, hooks, MCP servers), but its example teams only model engineers: each agent works alone, and the setup has no notion of the whole value chain of builders (product owner, engineer, customer, QA, UX/UI designer) working together. The talk that prompted this ("Every builder ships", attached transcript) shows what teams want instead: role agents that consult each other, review gates in the spec workflow, and a harness that the whole team owns and improves as a product.

Today, a product owner or designer who finds a skill or agent unhelpful has no way back to whoever maintains it. The feedback stays in a chat, the harness does not improve, and maintaining it stays an engineering-only concern.

## Success measure

A non-engineer can report a problem with a skill or agent in under a minute, and the report reaches the person who owns it, with enough context to act on it.

## Done means

1. **A team can install a ready example of cooperating role agents.** `kirokit install <example team>` gives five role agents (product owner, engineer, customer, QA, UX/UI designer), and in `kiro-cli chat` the product owner agent consults the customer agent as a subagent and reports its answer.
2. **The spec workflow has review gates by role.** With the example team installed, a Kiro spec's requirements are reviewed by the product owner agent and its design by the engineer agent, as a steering describes.
3. **Every agent and skill of a team names its owner.** The team's `manifest.yaml` names a team owner as a GitHub handle (e.g. `@jane`), and any agent or skill can name its own instead; vendored skills fall back to the team owner. The desktop app's Agents and Skills pages show each owner.
4. **A non-engineer can send feedback on a skill or agent from where they work.** From the Kiro chat (a prompt) and from the desktop app (a button on the agent or skill), they get a ready feedback report naming the skill or agent.
5. **The feedback reaches the owner.** The report opens as a prefilled GitHub issue in the repository the team's `manifest.yaml` names for feedback (e.g. `feedback: { github: org/kiro-config }`), labelled with the team and the agent or skill, @-mentioning the owner so GitHub notifies them; nothing is stored or sent without the person submitting it. (only you can confirm: the issue reads well to an owner)
6. **A session-analysis skill suggests harness improvements from a Kiro chat.** Given a chat, the skill reports concrete recommendations (a hook to add, a steering to harden, a missing permission, an agent prompt to sharpen), each naming the agent, skill or steering and its owner, then offers to open a prefilled issue per owner the same way as feedback.
7. **The README explains the role setup and the feedback loop.** A reader learns how to use the example team, declare owners, route feedback and analyse a session.

## Scope

### In this task

- The role-agent example team, its steerings and prompts, rendered and golden-tested like the other teams.
- Owners in `manifest.yaml`, validated and shown by kirokit.
- The feedback entry points and the prefilled issue they produce.
- The session-analysis skill.

### Not in this task

- Kiro hooks that fire on spec phases (Kiro offers none today; the gates use a steering, which instructs the agent but cannot force the review).
- Sending feedback without the person (no tokens, no automatic issue creation).

## Decided with you

<!-- ql:decided -->
| # | Question | Your answer |
|---|---|---|
| 1 | Which parts of the talk's setup ship first? | Role agents consulting each other + Spec review gates steering + Named owner per agent/skill + Non-engineer feedback to owners + Session-analysis skill |
| 2 | Where do the role agents live? | New example team |
| 3 | Where does a non-engineer give feedback? | Both prompt and button |
| 4 | Where should the feedback arrive? | GitHub issue for the owner |
| 5 | Which role agents does the example team ship? | Product owner + Engineer + Customer + QA + UX/UI designer |
| 6 | How is an owner written down? | GitHub handle |
| 7 | Must every agent and skill name an owner? | Team default, per-item overrides |
| 8 | What does session analysis do with its recommendations? | Report, then offer owner issues |
| 9 | How does kirokit know which repository receives feedback issues? | manifest.yaml feedback field |
<!-- /ql:decided -->

## First release

All five parts of the talk's setup: role agents that consult each other, spec review gates, an owner per agent and skill, feedback from non-engineers routed to that owner, and a session-analysis skill that turns a Kiro chat into harness improvements.

## What the user sees

- **A product owner in Kiro** starts `kiro-cli chat --agent product-owner`, asks about a feature idea, and the agent asks the customer agent and reports back. When the chat goes wrong, they type `/feedback`, and the agent drafts a report naming the skill or agent, which opens as a prefilled GitHub issue.
- **A designer in the desktop app** opens Agents or Skills, sees who owns each one, and presses **Feedback** to open the same kind of prefilled issue.
- **An owner** gets an issue labelled with the team and the agent or skill, describing what went wrong.

## Not doing

- Automatic edits to skills from feedback; the owner decides.
- Collecting usage or credit metrics.

<!-- optional section here, only when there is content: "Worth it when" -->

<!-- optional section here, only when there is content: "Reference" -->

<!-- optional section here, only when there is content: "Added while building" -->

## Context pointers

- Talk transcript: `/Users/mavogel/.qualitylayer/runs/242568bc-3c5f-4452-a846-1861f2b046d8/launch-files/files/kirokit-next.txt`
- Slides (screenshots): `Screenshot 2026-10-26 at 11.56.12.png` (the five roles), `11.56.46.png` (agents, skills, steering, hooks, MCP tools), `11.56.53.png` (roles cloned as product owner, engineer and customer agents), `11.57.12.png` (agent-centric workflow: ideate, plan, refine, checkpoint, build in worktrees), `11.57.18.png` (takeaways: harness as a product, own yours, codebase ready for agents), all in `/Users/mavogel/.qualitylayer/runs/242568bc-3c5f-4452-a846-1861f2b046d8/launch-files/files/`

## First look at the code

- Team layout and manifest: `teams/platform-team/manifest.yaml:1`, `internal/manifest/manifest.go:16` (`Team`, `GlobalSteerings`, `ExternalSkills`).
- Agent templates: `teams/platform-team/agents/devops/agent.json.tmpl:1` (tools, permissions, resources, SOPS MCP), `teams/platform-team/agents/platformdev/agent.json.tmpl:1`.
- Hooks: `teams/platform-team/hooks/lint-on-save.json.tmpl:1`; valid triggers in `internal/kiroconfig/validate.go:15`.
- Prompts become slash commands: `teams/platform-team/prompts/create-steering.md.tmpl`.
- Onboard starter: `internal/scaffold/scaffold.go:28`, `internal/scaffold/starter/`.
- CLI commands: `cmd/kirokit/commands/install.go:14`, `cmd/kirokit/commands/onboard.go:13`, `cmd/kirokit/commands/sync.go:19`, `cmd/kirokit/commands/setup.go:28`.
- Desktop catalog and pages: `desktop/catalog.go:25` (`Catalog`, `Team`, `Agent:42`, `Skill:54`), `desktop/frontend/src/pages/agents.js:4`, `desktop/frontend/src/pages/skills.js`, `desktop/frontend/src/api.js:1`.
- Golden tests: `testdata/golden/<team>/`, `internal/golden`.

*Untraced areas:* traced by Research since: Kiro's `subagent` tool and `toolsSettings.subagent.availableAgents`, Kiro's session files under `~/.kiro/sessions/`, the desktop app's missing URL opening (Wails `BrowserOpenURL` unused). Still open: what the `subagent` permission rule's `match` compares against, and whether Kiro CLI v3 runs the spec workflow in a chat (docs say yes; not yet run).
