# Configuration

How a config repository is laid out, what a team's `manifest.yaml` holds, how MCP versions are pinned, and how to add a team. Back to the [README](../README.md).

## Directory Layout

This repository has the configs, the `cli` and desktop app. It can be used as a baseline for the configuration as follows:

```
teams/<team>/manifest.yaml              # team identity + owners + feedback repository + global steering opt-ins + externalSkills + quick
teams/<team>/agents/<agent>/agent.json.tmpl
teams/<team>/agents/<agent>/env/secrets.enc.yaml   # optional, SOPS-encrypted
teams/<team>/hooks/*.json.tmpl          # shared across the team's agents
teams/<team>/prompts/*.md.tmpl
teams/<team>/skills/<skill>/SKILL.md.tmpl
teams/<team>/steerings/*.md.tmpl
teams/<team>/external/skills/<name>/    # vendored external skills (generated, never hand-edit)
teams/<team>/external/skills.lock.yaml  # generated

global/steerings/*.md                   # shared steerings teams opt into
env/mcp-versions.yaml                   # central MCP version pins
quick-ids.yaml                          # your Amazon Quick skill ids (written by kirokit quick ids, gitignored)

testdata/golden/<team>/rendered/...     # committed golden output tree
testdata/golden/<team>/quick/...        # committed golden Amazon Quick bundle (teams with a quick section)
testdata/golden/<team>/context.json     # committed golden template-context snapshot
testdata/golden/quick-ids.yaml          # made-up Quick ids the Quick goldens render with
```

- Template sources use a double extension `<ext>.tmpl`; the renderer globs `**/*.tmpl` and strips `.tmpl`.
- Agent source `agents/<agent>/agent.json.tmpl` renders to flat `agents/<agent>.json` (Kiro CLI v3 layout).
- Paths `teams/`, `global/steerings/`, `env/mcp-versions.yaml`, `quick-ids.yaml` and `testdata/golden/` resolve against the cwd, or against `KIROKIT_REPO` when that environment variable is set (a relative value is made absolute once at start; empty means unset). With `KIROKIT_REPO=<checkout>` every command (`render`, `test`, `update`, `install`, `sync`, `onboard`, `skills fetch`, `quick ids`) runs from any folder. There is no repo-root discovery.
- Example teams: `platform-team` (`devops`, `platformdev`; SOPS MCP servers; opted into `security-baseline`), `data-team` (`javadev` with SOPS Snowflake MCP, `godev` with plain MCP; no global steerings) and `product-team` (five role agents that consult each other, spec review gates, `/feedback`, `analyze-session` and `user-story`; no MCP servers or secrets; exports `user-story` and `product-owner` for Amazon Quick; see [Role agents and feedback](agents-and-skills.md#role-agents-and-feedback)). Copy one or use `kirokit onboard`; duplication between teams is expected.
- Prompts are slash commands: Kiro v3 lists every file in `$KIRO_HOME/prompts/` (`@name` on other engines) even if no agent uses it as `prompt`; `$ARGUMENTS` / `${1}` are filled from user input. Example: `teams/platform-team/prompts/create-steering.md.tmpl` becomes `/create-steering [focus]`.

## manifest.yaml Reference

```yaml
team: my-team
owner: "@my-org/my-team"   # optional; GitHub handle (@user or @org/team) owning every agent and skill
owners:                    # optional; per-item owners instead of the team owner
  agents:
    devops: "@jane"
  skills:
    jira-search: "@nick"
feedback:                  # optional; where feedback issues go (owner/repo on GitHub)
  github: my-org/kiro-config
globalSteerings:
  - security-baseline   # must match global/steerings/<slug>.md; [] opts out (see teams/data-team)
externalSkills:         # optional, see docs/agents-and-skills.md "External Skills"
  - source: https://github.com/anthropics/skills.git   # git URL, no credentials
    ref: <commit-sha>   # tag, branch or commit
    paths:              # skill directories inside the repo; each is named after its last element
      - skills/pdf
      - skills/xlsx
  - source: https://github.com/anthropics/skills.git
    ref: <commit-sha>
    paths:
      - skills/frontend-design
    name: design        # optional; only with a single path
  - source: https://github.com/anthropics/skills.git
    ref: <commit-sha>
    paths:
      - skills          # directory of skills
    all: true           # vendor each subdirectory containing SKILL.md
    exclude:            # optional; skip these skill directories (single path only)
      - docx
quick:                  # optional; the Amazon Quick bundle, see docs/amazon-quick.md
  skills: [jira-search] # team or vendored skills exported for Quick
  agents:
    - name: devops      # an agent of this team
      welcomeMessage: "Which ticket should we look at?"
      starterPrompts: ["Triage my open tickets"]
      skills: [jira-search]   # pinned to the agent in Quick; each must be in quick.skills
```

- Each opted-in global steering renders to `$KIRO_HOME/kirokit-steerings/<team>/<slug>.md` and is appended (created if missing) to the `resources` of every agent of the team as `file://$KIRO_HOME/kirokit-steerings/<team>/<slug>.md`. It is not written to Kiro's auto-loaded `steering/` dir, so only opted-in teams load it. Start the file with steering front matter (e.g. `inclusion: always`).
- `{{ globalSteering "<slug>" }}` inlines a steering, front matter included; prefer the automatic resource.
- Every agent of every team also gets `file://.kiro/steering/**/*.md` and `skill://.kiro/skills/**/SKILL.md` in `resources`, after global steerings. Relative paths load the steering/skills of the workspace Kiro starts in. `skill://` loads metadata at startup, the rest on demand. Existing identical entries are not duplicated.
- `owner`, `owners` and `feedback` are optional; a team without them renders as before. Handles must look like `@user` or `@org/team` and `feedback.github` like `owner/repo`, else loading fails naming the field (`owners.agents.qa: "jane" is not a GitHub handle (write @jane)`). An `owners` key naming an agent or skill the team does not have fails render. Vendored skills have the team owner unless `owners.skills` names them. See [Role agents and feedback](agents-and-skills.md#role-agents-and-feedback).
- `quick` is optional; a team without it has no Quick bundle and renders as before. Unlike the rest of the manifest it is read strictly: an unknown key (`quick.agents[0].welcome`), a duplicate name or an agent without `name` fails loading. Whether each listed skill and agent exists is checked by `kirokit render --quick`. See [Amazon Quick](amazon-quick.md).

## Central env/mcp-versions.yaml

Flat map of MCP server name to pinned version:

```yaml
jira-mcp: "1.4.2"
confluence-mcp: "2.1.0"
```

- Templates read it with `{{ mcpVersion "jira-mcp" }}`. An undefined key fails the render naming the key and file.
- `context.json` snapshots the entire resolved map, so bumping any entry changes every team's golden context; run `kirokit test` to see the affected output.

## Onboarding a New Team

```bash
kirokit onboard <team>
```

Scaffolds `teams/<team>/` from an embedded starter: `manifest.yaml` with the team name, one example agent (`agents/example-agent/`), empty `hooks/`, `prompts/`, `skills/`, `steerings/`. Errors if `teams/<team>/` exists. Copying another team's directory works too.
