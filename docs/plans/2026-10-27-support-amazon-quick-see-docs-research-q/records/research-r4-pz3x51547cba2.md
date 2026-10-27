---
type: research
---

# How kirocrew reads ~/.kiro, what kirokit install writes, and the repo's per-user files, showcase and README

## Findings

### kirocrew resolves one Kiro home, honours KIRO_HOME, and reads agents, skills, settings from it
**Answers:** 7

- `kiro_home()` returns `Path.home()/.kiro` unless `KIRO_HOME` is set (`kiro_crew/config/paths.py:1217-1224`). A `KIRO_HOME` that resolves to a root or system directory is ignored with a warning (`paths.py:1221-1223`). Paths below are inside `/Applications/KiroCrew.app/Contents/Resources/backend-dist/kirocrew-backend-arm64/lib/python3.12/site-packages/`.
- The docstring says `KIRO_HOME` is only partly followed today: the agents directory follows it, but session transcripts and the `settings/mcp.json` registry are read from the host `~/.kiro` in "most" readers, so it "is not yet a supported way to isolate an instance" (`kiro_crew/config/paths.py:1201-1210`). This conflicts with `mcp_cleanup.py:249-251`, which does use `kiro_home() / "settings" / "mcp.json"`, and `dashboard/handlers/mcp.py:120`, which hard-codes `Path.home()/.kiro/settings/mcp.json`. The code in each module decides at runtime, so MCP settings are read from different homes depending on the caller (fact; no single winner).
- Agents: user-level dir is `<kiro home>/agents` (`paths.py:1348`, `1351-1378`). `list_agents` scans it (`kiro_crew/agent_discovery.py:2297`, `2341-2350`). A project's `<project>/.kiro/agents` shadows same-named user agents, with no upward walk (`paths.py:1381-1393`; `agent_discovery.py:2278-2285`).
- Skills: only `skill://` entries in an agent's `resources` count as that agent's mapping (`agent_discovery.py:1257-1260`). A `skill://~/...` URI expands with `Path.home()`, not `kiro_home()` (`agent_discovery.py:1331-1332`), and `skill:///abs/...` is used as is (`1333-1334`). A custom agent with no `skill://` entry inherits `<kiro home>/skills/*/SKILL.md` plus `<project>/.kiro/skills/*/SKILL.md`, unless `chat.disableInheritingDefaultResources` is set (`agent_discovery.py:1595-1601`; `acp/skill_projection.py:1929-1953`, which reads `<kiro home>/settings/cli.json`).
- Steering: kirocrew does not parse steering files. When an agent inherits defaults, its generated view gets the resources `file://<kiro home>/steering/**/*.md`, `file://.kiro/steering/**/*.md` and `file://AGENTS.md` (`acp/skill_projection.py:3911-3919`).
- MCP: `<kiro home>/settings/mcp.json` is treated as belonging to kiro-cli. Writes are refused when `KIROCREW_POD` is set (`paths.py:731-751`). Per-agent MCP server names come from the spec's `mcpServers` keys (`agent_discovery.py:2262-2265`).

### Agents kirocrew did not create are listed, kept as the author wrote them, and given a derived view
**Answers:** 7

- Ownership is by exact filename: `OWNED_KIRO_AGENT_FILES` lists `kirocrew.json`, `kirocrew-lite.json`, `kirocrew-conductor.json` etc. (`kiro_crew/agent_files.py:93-108`; the comment at `agent_files.py:89-92` says user files with the same prefix are never touched). Any other file in `~/.kiro/agents` gets `kirocrew_owned=False` and source `"builtin"`, or `"package"` when the filename is `<package>-<name>.json` (`agent_discovery.py:2209-2237`). The directory is described as shared with "ACP adapters and IDE plugins" (`agent_discovery.py:212-213`).
- Fields in foreign specs are coerced rather than rejected (`agent_discovery.py:234-251`). A spec that cannot be parsed is skipped (`agent_discovery.py:2346-2358`).
- For a session, kirocrew builds a copy of each spec as an alias file (name prefix `kirocrew-skill-view-`) in the same agents folder. The copy drops `skill://` resources, replaces `description` with "Kiro Crew managed view of ...", and adds the `kirocrew-core` MCP server and `skill_search` tool when `can_inherit_skills` allows it (`acp/skill_projection.py:3773-3797`, `3837-3860`; description prefix at `skill_projection.py:278-290`). The source spec file is only read, not rewritten (`skill_projection.py:3763-3773`).
- Agents that exclude `skill_search`, declare `tools: []`, or have no `kirocrew-core` entry with `tools: "*"` or missing get no inherited skills and keep their spawn (`agent_discovery.py:1514-1567`). A mapped agent that withholds `skill_search` is refused a view (`skill_projection.py:3797-3801`).
- Deleting a leftover view needs a sidecar record under kirocrew's data home, or the legacy name pattern plus a positive mark (the `kirocrew-core` server or the steering resource) (`skill_projection.py:2118-2167`). "An operator's own agent could in principle carry this name", so name alone is not enough (`skill_projection.py:2138-2145`).
- kirokit's agents in `~/.kiro/agents` (for example `engineer.json`, `qa.json`, `customer.json`) therefore appear as foreign rows. Observed in `~/.kiro/agents/`: they sit beside `kirocrew*.json` files; kirocrew's own data lives in `~/.kiro/crew` (`paths.py:41-49`).
- kirocrew data home defaults to `~/.kiro/crew` and a recovery pointer is kept outside `~/.kiro` (`paths.py:41-65`).

### `kirokit install` renders with the real KIRO_HOME and writes the rendered tree plus per-agent secrets
**Answers:** 8

- `ResolveKiroHome` uses `$KIRO_HOME` if non-empty, else `~/.kiro` (`internal/install/install.go:49-58`).
- `Install` renders into a scratch folder with `KiroHome` set (slashes normalised) (`install.go:92-102`, `75-77`), plans every rendered file at `<kiroHome>/<relative path>` (`install.go:319-351`), and adds `kirokit-secrets/<team>/<agent>/secrets.enc.yaml` for each agent that has `env/secrets.enc.yaml` (`install.go:353-383`).
- Folders seen in a rendered team: `agents`, `hooks`, `prompts`, `skills`, `steering` (`testdata/golden/data-team/rendered/`). Global steerings go to `kirokit-steerings/<team>/` (`internal/render/globalsteerings.go:19`); team `steerings/` map to `steering/` (`internal/render/render.go:383-386`). README names the backed-up folders as `agents`, `prompts`, `skills`, `steering`, `hooks`, `kirokit-steerings`, `kirokit-secrets` (`README.md:281`).
- Files are written `0600`, or `0700` when the rendered file has the owner-execute bit (`install.go:340-343`); existing file modes are reset with chmod (`install.go:451-454`).
- Rendered agent JSON points at Kiro home absolute paths, for example `skill:///kiro-home/skills/jira-search/SKILL.md` and `file:///kiro-home/kirokit-steerings/platform-team/security-baseline.md` in the golden (`testdata/golden/platform-team/rendered/agents/devops.json`, sentinel `/kiro-home`; `install` replaces it with the real home).
- Optional `--kb-*` flags add a `knowledgeBase` resource to one agent and a symlink at `--kb-mount`, which is also recorded in the manifest (`install.go:120-142`, `213-257`; flags `cmd/kirokit/commands/install.go:50-52`).

### The record is `<KIRO_HOME>/.kirokit-installed/<team>.json`, a list of files and symlinks; uninstall and sync remove only those
**Answers:** 8

- Path: `installManifestPath` = `<kiroHome>/.kirokit-installed/<team>.json` (`install.go:290-292`). Shape: `{"files": [...], "symlinks": [...]}` (`install.go:42-45`), mode `0600` (`install.go:303`). The record file itself is listed in `files` (`install.go:144-145`). It is the install record, distinct from the team's `manifest.yaml` in `internal/manifest`, which parses `team`, `owner`, `owners`, `feedback`, `globalSteerings`, `externalSkills` (`internal/manifest/manifest.go:19-28`).
- `Uninstall` removes recorded symlinks, then files, ignoring ones already gone; missing record is an error (`install.go:262-288`).
- `Sync` installs, then removes paths in the previous record that the new record lacks and removes empty parent folders up to the Kiro home (`internal/install/sync.go:16-62`).
- `Installed` lists `*.json` records sorted by team with file counts and mod-time (`internal/install/installed.go:22-54`); `UninstallOthers` removes every installed team except the one being installed (`installed.go:59-75`). The CLI runs `install.Check` first (render into scratch, nothing written), then `uninstallOthers`, then `Install` (`cmd/kirokit/commands/install.go:36-44`); `sync` does the same with `install.Sync` (`cmd/kirokit/commands/sync.go:111-120`). So one team per Kiro home.
- Observed on this machine: `~/.kiro/.kirokit-installed/` holds `platform-team/` and `product-team.json` (directory listing).
- Backups: before writing, each top-level folder containing a file whose content would change is copied to `<kiroHome>/YYYYMMDD-HHmmss-<folder>.backup`; unchanged folders get none; backups are not in the record (`install.go:402-438`; naming and counter in `internal/backup/backup.go:23-35`, `51`, `120`). Observed: `~/.kiro/20261018-105014-agents.backup` and `20261026-195053-agents.backup`.
- Hand-added files in a folder are not recorded, so uninstall and sync never remove them; files written by `install` over an existing file replace it after the backup (`install.go:441-457`).

### The desktop app chooses the Kiro home as saved Settings, else the login shell's KIRO_HOME, else ~/.kiro, and passes it to the CLI as KIRO_HOME
**Answers:** 8

- `App.kiroHome()` returns `Settings.KiroHome` when saved, else `defaultKiroHome` (captured login-shell `KIRO_HOME` once ready, else `~/.kiro`) (`desktop/settings.go:135-155`). Settings live in `settings.json` in the app data folder (`desktop/settings.go:42`, `107-133`); a saved value must be absolute and not the root or home folder itself (`settings.go:157-166`).
- Every terminal and button command gets `KIROKIT_REPO=<repoDir>` and `KIRO_HOME=<kiroHome>` (`desktop/shell.go:117-118`); the zsh wrapper re-asserts them after the user's startup files (`shell.go:144-146`).
- A poller reads the install records of that home every second and emits `installs:changed` (`desktop/installs.go:11`, `53-68`, `88-93`); `Installs()` serves the list (`installs.go:97-103`).
- The app's data directory is `kirokit-desktop` in the user's config folder; default clone `<dataDir>/repo` (`desktop/app.go:55-60`, `133-143`).

### The repo ignores build output and tool state; paths are overridden by flag or environment variable; the only example-file pattern is the scaffold's `*.tmpl` starter
**Answers:** 9

- `.gitignore` ignores `/dist/`, `/kirokit`, `/tmp/`, `.claude/*` (except `.claude/rules/` and `.claude/skills/`), `demo-runbook-*`, and un-ignores `/docs/plans/*/design/` (`.gitignore:1-12`). The checkout's `.git/info/exclude` also lists `/docs/plans/*/design/` and `/docs/plans/*/artifacts/` (QualityLayer entries; `.git/info/exclude` tail). `.gitleaks.toml` and `.sops.yaml` exist at the root.
- `KIROKIT_REPO`: read in the root command's pre-run, made absolute, stored in `repoBase`; `repoPath(rel)` joins it onto `teams`, `global/steerings`, `env/mcp-versions.yaml`, `testdata/golden`; unset means the working directory (`cmd/kirokit/commands/root.go:12-31`, `44-54`).
- `kirokit sync --repo <dir>` overrides it: the order is flag, then `KIROKIT_REPO`, then `.` (`cmd/kirokit/commands/sync.go:77-89`, `128`). `setup` prints `export KIROKIT_REPO=<dir>` (`cmd/kirokit/commands/setup.go:70`, `85`).
- Other environment variables read in Go: `KIRO_HOME` (`internal/install/install.go:50`), `KIROKIT_DESKTOP_CLI` (`desktop/shell.go:41`), `KIROKIT_DESKTOP_SMOKE` (`desktop/main.go:35`), `SHELL`, and from the captured login environment `SOPS_AGE_KEY_FILE`, `APPDATA`, `XDG_CONFIG_HOME`, `HOME` (`desktop/platform.go:32-46`). The desktop app's own overrides are its saved Settings (`RepoURL`, `KiroHome`, `RepoDir`, `Ref`) (`desktop/settings.go:18-23`).
- No tracked `*.example`, `.env.example` or `*.local` files exist (`git ls-files` filtered for example/sample/local/.env gave only `internal/scaffold/starter/agents/example-agent/{agent.json.tmpl,env/mcp.env.tmpl}`). Those are the `kirokit onboard` starter embedded in the binary (`internal/scaffold/`), not a per-user file pattern. Secrets files are `*.enc.yaml` committed and SOPS-encrypted (`AGENTS.md` repository rules). The SOPS age key is read from `SOPS_AGE_KEY_FILE` or the default path (`scripts/showcase/record.sh:62`).
- Ignored local files in this checkout: `kirokit` binary, `tmp/`, `demo-runbook-v1.md` (confirmed with `git check-ignore -v`).

### The README showcase is recorded by `task showcase` in an isolated /tmp folder from the real CLI and app, then checked and published atomically
**Answers:** 10

- `Taskfile.yml` target `showcase` runs `scripts/showcase/record.sh` (`Taskfile.yml:51-54`). Preflight requires tmux, asciinema, agg, ffmpeg, ffprobe, node, npm, go, git, kiro-cli, lsof, Google Chrome, the SOPS age key, and free ports 34115 and 5173 (`record.sh:58-68`). It works in `/tmp/kirokit-demo`, created `mkdir -m 700` with a pid file, removed on exit unless `--keep` (`record.sh:22`, `70-78`, `32-50`).
- Scene 1 (CLI, `cli-drive.sh`): tmux session with `KIRO_HOME=$demo/kiro` and `KIROKIT_REPO=<repo root>` (`record.sh:89-91`). It types `kirokit install platform-team` and waits for `installed platform-team`, then `kiro-cli chat --agent devops --v3`, Enter on "Trust this workspace", waits for "ask a question", types `/jir`, waits for `jira-search`, Esc, `/quit`, waits for "Session ended"; it never presses Enter on a chat message (`scripts/showcase/cli-drive.sh:1-8`, `50-66`). Recorded with asciinema, rendered to a GIF with agg (`record.sh:92-108`).
- Scene 2 (app, `app-drive.mjs`): `wails dev` with a throwaway `HOME`, `KIRO_HOME` and `KIROKIT_REPO` unset (`record.sh:125-132`); headless Chrome via playwright-core types `https://github.com/mavogel/kirokit.git` into `#clone-address`, clicks Clone, waits for the `platform-team` Install button, clicks it, handles an optional "Replace the installed team?" dialog, clicks `#setup-install`, waits for the `.badge.ok` badge and the dock line `installed platform-team into` (`scripts/showcase/app-drive.mjs:64-115`). It samples all page and terminal text every 200 ms into `app-dock.txt` for the privacy check and writes `app-marks.json` (`app-drive.mjs:44-62`, `118-122`).
- Captions (`caption.mjs`): four captions "Install your team's Kiro config", "Kiro starts with the team's agent and skills", "Or click: clone the config repo", "Install the team, see every command"; a title card and an end card with `brew install --cask mavogel/tap/kirokit` and `...kirokit-desktop` and `github.com/mavogel/kirokit` (`caption.mjs:28-40`). It computes cue times from the cast (idle limit 3 s) and the app marks, cuts the clone wait to 1.5 s, and fails if a caption would show under 2.5 s (`caption.mjs:23-25`, `81-109`).
- Compose: ffmpeg joins title, CLI clip, app clip and end card with fades into `showcase.mp4`, then a 12 fps 1000 px-wide GIF (`record.sh:158-210`).
- Checks (`checks.sh`): `privacy` rejects `$HOME`, the git email and `$USER` in `cli.cast`, `app-dock.txt`, `captions.txt`, except the public slugs `mavogel/kirokit` and `mavogel/tap/`; `size` limits GIF to 8 MB and MP4 to 4 MB; `length` needs 30 to 60 s (`checks.sh:8-60`, `record.sh:215-217`). Publish copies to a temp name in `assets/` then renames to `assets/kirokit-showcase.gif` and `.mp4` (`record.sh:219-226`).
- CONTRIBUTING "Showcase video" says to re-record on a Mac when CLI output or app UI changes, lists the brew packages, the other requirements, and `node --test scripts/showcase/` (about 40 s) (`CONTRIBUTING.md:101-112`). `AGENTS.md:11` points to the same section. The README embeds the GIF and links the MP4 (`README.md:10-11`).
- Last 90 days: 2 commits touched `scripts/showcase` (`git log --oneline --since=90.days -- scripts/showcase`).

### The README has 15 H2/H3 sections; other docs point at it by file or by section name, and no test reads its text
**Answers:** 11

- Headings (`README.md`): H1 `kirokit` (3); `Quickstart` (13) with `using the cli` (15) and `using the desktop app` (70); `Directory Layout` (94); `manifest.yaml Reference` (122); `Role agents and feedback` (161); `External Skills` (177); `Central env/mcp-versions.yaml` (199); `Secrets via SOPS` (211); `Testing and verify by rendering upfront` (247); `Onboarding a New Team` (257); `Install / Uninstall` (265) with `Existing files are backed up` (278); `Setup (initial download)` (289); `Sync (automatic updates)` (303); `Desktop app` (338). Lines starting `#` inside code blocks (32-65) are shell comments, not headings. The file is 359 lines; 28 commits in 90 days.
- Anchor links inside the README: `[reference](#manifestyaml-reference)` (`README.md:172`), `[Role agents and feedback](#role...` (`README.md:352`), `[Desktop app](#desktop-app)` (`README.md:92`).
- `CONTRIBUTING.md` links the README by file only (`CONTRIBUTING.md:81`, `152`) and mentions it at `103`; its own in-page anchors are `#development-setup`, `#checks`, `#commit-messages` (`CONTRIBUTING.md:31-34`). `AGENTS.md` mentions `README.md` at lines 4 and 33 and the showcase at 11, with no anchors.
- Section names are quoted in prose elsewhere: "Secrets via SOPS" in `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9` (rendered into new teams) and `global/steerings/security-baseline.md:9`; "External Skills" in `teams/platform-team/manifest.yaml:12`; the README for `owners` and `feedback` in `teams/platform-team/manifest.yaml:4`; `teams/data-team/steerings/data-conventions.md.tmpl:12`; `tmp/demo-runbook-v0.md:247,264` (ignored folder) names "Directory Layout", "Golden Tests" (no such heading today), "Secrets via SOPS", "External Skills", "Sync (automatic updates)".
- The desktop app and its frontend contain no README references (grep of `desktop/` found only a Wails package homepage URL). The only test hit is `internal/kiroconfig/validate_test.go:11`, which uses `file://README.md` as sample data, not the repo README. `scripts/sync-agent-assets.mjs:286` skips files named `README.md` while syncing skills. `Taskfile.yml:52`, `scripts/showcase/record.sh:2` and `package.json:5` mention it in descriptions.
- Teams' rendered golden files can embed the heading name through the scaffold template, so a rename would show in `kirokit onboard` output (`internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9`); goldens under `testdata/golden/` are not generated from that starter (inference from the starter living in `internal/scaffold`, not `teams/`).
- Rule: update README in the same change when a command, flag, manifest field, or template function changes (`AGENTS.md:33`, `CONTRIBUTING.md:152`).

## How it is tested today

- Install and backup logic: Go tests beside the code, using temp dirs (`internal/install/install_test.go`, `installed_test.go`, `sync_test.go`, `install_sops_test.go`; `internal/backup/backup_test.go`; `internal/manifest/manifest_test.go`; `cmd/kirokit/commands/sync_test.go`, `root_test.go`). One file: `go test ./internal/install -run TestName`. Suite: `go test ./...` from the repo root, or `task ci` (build, vet, lint, test) (`AGENTS.md` Commands). Tests needing a Unix shell or execute bits skip on Windows (`internal/testutil`).
- Desktop app: own module; `cd desktop && go test ./...` (for example `desktop/settings_test.go`, `installs_test.go`); `task desktop:test` runs it including `node --test` for `desktop/frontend` (`AGENTS.md`). Not run by me.
- Showcase: `node --test scripts/showcase/` runs `scripts/showcase/showcase.test.mjs` (472 lines) for the privacy, size and length checks, caption timing and error paths of `caption.mjs`, `app-drive.mjs`, `cli-drive.sh` and `record.sh` preflight; tests needing Chrome, ffmpeg or tmux skip without them. From the root, one file is `node --test scripts/showcase/showcase.test.mjs`. It never records anything (`CONTRIBUTING.md:112`).
- kirocrew: the installed app ships no tests that I opened; nothing about it was executed. All kirocrew findings are from reading the installed source.
- README: no automated test reads it; `kirokit test` and `go test ./internal/golden` compare only `testdata/golden/**`.
- None of the test commands above were run by me; this research was read-only.

## Open questions

Question 7: Which of the two `settings/mcp.json` paths kirocrew uses at runtime when `KIRO_HOME` is set depends on the calling module (`mcp_cleanup.py:249-251` follows `kiro_home()`; `dashboard/handlers/mcp.py:120` is fixed to `~/.kiro` at import time). I did not trace every other reader.
Question 7: I did not read kirocrew's behaviour for steering beyond the generated `resources` lines, nor how Kiro IDE (`/Applications/Kiro.app`) reads the same folders; that was not in the files named.
Question 11: Whether `docs/` pages (excluded from my search) link to README anchors; the only hits there were embedded copies in evidence files.

## Code references

Exhaustive for the files named in the questions; selective for kirocrew (a 5,450-line pair plus targeted greps of the package).

- kirocrew (installed): `kiro_crew/config/paths.py:41-65, 731-793, 1190-1224, 1348-1393`; `kiro_crew/agent_discovery.py:186-257, 1249-1260, 1306-1336, 1514-1602, 2199-2265, 2268-2358`; `kiro_crew/agent_files.py:20-26, 93-108`; `kiro_crew/acp/skill_projection.py:1-8, 255-290, 1929-1953, 2118-2167, 3760-3860, 3911-3919`; `kiro_crew/mcp_cleanup.py:249-251`; `kiro_crew/dashboard/handlers/mcp.py:120-123`. Base: `/Applications/KiroCrew.app/Contents/Resources/backend-dist/kirocrew-backend-arm64/lib/python3.12/site-packages/`.
- Install and records: `internal/install/install.go`, `installed.go`, `sync.go`; `internal/backup/backup.go`; `internal/render/globalsteerings.go:19`, `render.go:341-386`; `internal/manifest/manifest.go:19-28`; `cmd/kirokit/commands/install.go`, `sync.go`, `root.go:12-54`, `setup.go:70,85`.
- Desktop: `desktop/app.go:55-143`, `desktop/settings.go:18-166`, `desktop/installs.go`, `desktop/shell.go:107-146`, `desktop/platform.go:32-46`.
- Ignored files: `.gitignore:1-12`, `.git/info/exclude`, `.gitattributes`, `internal/scaffold/starter/`.
- Showcase: `scripts/showcase/{record.sh,cli-drive.sh,app-drive.mjs,caption.mjs,checks.sh,showcase.test.mjs,package.json}`, `Taskfile.yml:51-54`, `CONTRIBUTING.md:101-112`, `AGENTS.md:11`.
- README links: `README.md` headings (list above), `CONTRIBUTING.md:31-34,81,103,152`, `AGENTS.md:4,11,33`, `internal/scaffold/starter/agents/example-agent/env/mcp.env.tmpl:9`, `global/steerings/security-baseline.md:9`, `teams/platform-team/manifest.yaml:4,12`.
