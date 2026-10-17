Result: done

I took and opened all four pictures (1440 and 390, light and dark). None is blank, stale or broken. They are in `/Users/mavogel/.qualitylayer/projects/kiro-cli-kit-a365619120de/tasks/2026-10-17-desktop-app-sidebar-layout/design-checks/`. I found nothing that must change. I changed no files; the only untracked path in git is the plan folder.

Acceptance points, checked against the pictures and source:
1. **Window:** one desktop window shows the sidebar, main pane and docked terminal at 1440 in both themes.

   Answer: Kept · confirmed; nothing to change.
2. **Sections:** all five are listed, each with its own page section in the source. The pictures show only Teams. The other four pages are drawn only in source and were not clicked.

   Answer: Kept · confirmed in Chrome: clicking each of the five sections showed its page and moved `aria-current` to it.
3. **Rail:** the toggle (`.rail` class, 232px to 60px) and the grid column both support a wider main pane. I did not capture the collapsed state.

   Answer: Kept · confirmed in Chrome: the rail settles at 60px, the main pane widens to 1118px, and every icon button keeps its name in the accessibility tree.
4. **Team setup:** the Teams cards show platform-team "Not installed" with an Install button. In the source, Install opens a dialog with SOPS, key and Kiro CLI checks, and the key starts as missing. "Check again" enables Install, and Install marks the team installed on the Teams and Install pages. The dialog is not in any picture.

   Answer: Taken · clicked through in Chrome: the dialog opened with Install disabled, and Check again enabled it. Install then flipped both badges and printed `installed platform-team into /Users/you/.kiro`. The run exposed lost focus after install, because the opener button is re-rendered; focus now moves to that team's new Sync button, confirmed in a second run.
5. **Terminal:** it is docked at the bottom with a resize handle, clear and hide controls, and a typed-command input. Buttons echo `kirokit install`, `sync` and `uninstall` commands with sample output, as the source script shows.

   Answer: Taken · `kirokit version` now prints `kirokit v0.4.0` and `commit: 9a919bf`, matching the Plan's contract; confirmed by typing it in Chrome.
6. **Real data:** I compared against the repo and it matches.
   - Teams: `platform-team` and `data-team`.
   - Agents: devops, platformdev, godev and javadev, with the descriptions from the agent configs.
   - Skills: confluence-lookup, jira-search and sql-helper.
   - Prompts: 3 for platform-team and 2 for data-team.
   - MCP versions: all seven match `env/mcp-versions.yaml`.
   - The install state is labelled as sample data on the page.

   Answer: Kept · confirmed; the steering lines were also corrected (see the advisory below).
7. **Themes:** both read well, with purple taken from the logo mark. Contrast ratios I computed from the tokens:

   | Pair | Ratio |
   |---|---|
   | Light muted text on sidebar | 6.71 |
   | Light accent text on soft accent | 7.57 |
   | White on light accent fill | 7.1 |
   | Light "Installed" badge | 4.57 |
   | Light warning badge | 6.15 |
   | Dark text on dark surfaces | 6.84 to 8.68 |
   | Terminal muted text | 7.21 |

   All are at or above 4.5. Whether the look fits kirokit is yours to confirm.

   Answer: Kept · confirmed; the fit with kirokit stays Done means 7's point for the person to confirm.
8. **No code change:** the working tree adds only `docs/plans/2026-10-17-desktop-app-sidebar-layout/`.

   Answer: Kept · confirmed; nothing to change.

The 390 pictures show the desktop window cropped at the right edge. That is the intended `.scroll-x` framing for a desktop-only screen, not a phone-layout failure.

Advisory: the data-team card says "Steering: none shared", yet `teams/data-team/steerings/data-conventions.md.tmpl` exists. If "shared" means the global steerings, consider wording like "Team steering: data-conventions". Also, typing `kirokit install <team>` in the terminal installs without the prerequisites dialog. That is probably acceptable for a mockup.

Answer: Taken · data-team now reads "Steering: data-conventions" and platform-team "Steering: team-conventions, shared security-baseline", matching `teams/*/steerings/` and the manifest's `globalSteerings`.
Answer: Kept · typed `kirokit install` skips the dialog on purpose: the real CLI has no prerequisites step, and the terminal mirrors the CLI.

Not checked: the rail-collapsed state, the Agents, Install and sync, Skills and MCP versions and Settings pages as rendered, the setup dialog and its steps, the terminal hidden and resized states, and all keyboard and focus paths. The icon-rail accessible names rely on `title` because the label spans are `display:none`. 390 and the supported narrow desktop width (about 1180px window) were judged from the 390 picture only.

Answer: Taken · in the rail, labels are now visually hidden with a clip rule instead of `display: none`, so each icon button keeps its text as its accessible name. I confirmed the rail and the setup dialog with my own Chrome screenshots and accessibility snapshots after the fix. I checked the five pages by clicking each one and reading its visibility, not by screenshot. Full keyboard paths (Tab order, shortcuts, Escape, the resize keys) are left to the build's browser run.
