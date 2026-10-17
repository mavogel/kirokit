Result: needs work

1. Quality · "Needs SOPS" badge on the Agents page, light theme (1440-light). The text is #a16207 on #fef3c7, a contrast ratio of 4.42:1. That is below the 4.5:1 WCAG AA minimum for 12px text. Darken `--warn` in the light theme, for example to #92560a or darker, and re-measure. The same token colours the "Missing" badge in the setup dialog.

Answer: Taken · `--warn` in the light theme is now #854d0e, measured at 6.15:1 on #fef3c7 and 6.85:1 on white; applied in the single-file design `sidebar-layout`, which replaced this folder draft.

Everything else I checked meets the Done means.
- **Layout:** the window, sidebar and main pane are drawn at 1440 in both themes. The 390 captures are the desktop window framed by `.scroll-x`, so I did not count the clipped content as a phone failure.
- **Sidebar:** all five sections are present and each is wired to its own page in the source.
- **Rail toggle:** the Collapse button, Cmd/Ctrl+B and the `.rail` class switch the sidebar to a 60px rail, and the main pane widens.
- **Team install flow:** pressing Install opens a prerequisites dialog covering SOPS, key and Kiro CLI. Install stays disabled until "Check again" finds the key. Then the team is marked Installed.
- **Terminal:** it docks at the bottom, can be hidden and shown, resizes by drag or keyboard, echoes each button press as its `kirokit` command, and prints sample output for typed commands.
- **Real data:** teams, agents, skills and MCP versions match the repository. Teams are platform-team and data-team, agents are devops, platformdev, godev and javadev, and skills are confluence-lookup, jira-search and sql-helper. The MCP versions I compared match `env/mcp-versions.yaml`: mcp-remote 0.15.1, context7-mcp 4.4.2, confluence-mcp 2.1.0, pipeline-mcp 3.0.0, snowflake-mcp 0.9.1, jira-mcp 1.4.2 and example-mcp 1.0.0.
- **Contrast:** the other ratios I computed pass: the OK badge in light is 4.57:1, in dark 8.2:1, and the warn badge in dark is 8.7:1. The primary button is 6.9:1 in dark. Muted text on the sidebar is 6.7:1 in light and 6.8:1 on surface-2 in dark.
- **Code changes:** I did not run a diff or `git status`, so I did not check that kirokit's code is unchanged.

Advisory: the "Install state on this page is sample data" note is visible on the Teams page, which is honest. The design page's `<title>` says "kirokit desktop: sidebar layout".

Capture note: `qualitylayer design shot index` failed with "no design index". The working form was `qualitylayer design shot desktop-app-sidebar --page index ...`. All four pictures were taken and are in `/Users/mavogel/.qualitylayer/projects/kiro-cli-kit-a365619120de/tasks/2026-10-17-desktop-app-sidebar-layout/design-checks/` (index-1440-light, index-1440-dark, index-390-light, index-390-dark).

Not checked: keyboard path and focus order; the rail state and the other four pages, which have no picture; the setup dialog rendered, which I judged from source only; supported narrow desktop widths between 390 and 1180; the Done-means point that only you can confirm, that the colours fit the kirokit logo.
