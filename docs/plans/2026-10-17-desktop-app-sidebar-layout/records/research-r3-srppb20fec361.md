---
type: research
---

# Visual identity: a logo and a README, no design system

## Findings

### The only visual identity is a violet rounded-square logo

**Answers:** 7

The project ships one mark, drawn as a 512x512 SVG (`assets/logo.svg:1`). A rounded square (`rx="112"`) is filled with a diagonal gradient from `#8b5cf6` to `#4c1d95` (`assets/logo.svg:4-9`). Two faint brace shapes in `#ede9fe` at 55% opacity suggest template braces (`assets/logo.svg:10-14`). A white prompt chevron and a short cursor bar sit on top (`assets/logo.svg:15-19`). The logo reads as "templates plus a terminal prompt".

| Element | Value | Source |
|---|---|---|
| Canvas | 512x512, corner radius 112 | `assets/logo.svg:1`, `assets/logo.svg:9` |
| Background | linear gradient `#8b5cf6` to `#4c1d95` (top-left to bottom-right) | `assets/logo.svg:4-7` |
| Braces | stroke `#ede9fe`, opacity .55, width 22, round caps | `assets/logo.svg:11-14` |
| Chevron and cursor | stroke `#fff`, width 34, round caps | `assets/logo.svg:16-18` |

`assets/logo.png` is a 512x512 RGBA PNG (checked with `file`); fact: it is not referenced anywhere in the repository, only the SVG is (`README.md:1`). Inference: the PNG is a raster export of the SVG; I did not compare pixels.

```
assets/
├── logo.svg   (963 bytes, vector source, used by README)
└── logo.png   (32.4 KB, 512x512 raster, unreferenced)
```

### The README is the only place the identity appears

**Answers:** 7

The README opens with the logo centred at 120px width and alt text "kirokit" (`README.md:1`), then a plain `# kirokit` heading (`README.md:3`). No badges, no custom colours, no images other than the logo were found (grep for `logo|img|badge|shields` over the README hit only line 1). Sections are standard Markdown H2s: Quickstart, Taskfile, Directory Layout and others (`README.md:7`, `README.md:41`, `README.md:53`). The PRD contains no colour, type or design guidance; its only "design" hits concern MCP configuration (`docs/prd/2026-09-28-kiro-cli-kit-multi-team.md:96`).

### The CLI has no styled output and the desktop-app design is an empty stub

**Answers:** 7

The only design page is `.qualitylayer/designs/desktop-app/index.html`. It declares a "sidebar layout" purpose in its metadata (`.qualitylayer/designs/desktop-app/design.json:3-4`, `.qualitylayer/designs/desktop-app/index.html:6`). Its body is empty. The stylesheet sets only `color-scheme: light dark` and body `margin: 0; padding: 48px; font: 15px/1.5 system-ui, sans-serif` (`.qualitylayer/designs/desktop-app/index.html:8-11`). So the stated type is the system UI font stack at 15px/1.5, and it does not use the logo's violet. There are no tokens, components, sidebar markup or layout conventions defined anywhere. The repository also contains no Go source that uses a colour or terminal-styling library (grep for lipgloss, fatih/color, termenv and ANSI escapes found no files), so the CLI emits unstyled text. No CSS, HTML (other than the stub), or web assets are tracked besides the logo files.

```
.qualitylayer/designs/desktop-app/
├── design.json   {v:1, title, purpose, entry}
└── index.html    empty body, system-ui 15px
```

Change surface: `assets/` and `.qualitylayer/designs` were touched by 2 commits in the last 90 days (`dabb3eb`, `8e7c6dc`, per `git log --since=90.days`). `README.md` changed in 20 commits in that window. The logo has exactly one consumer (`README.md:1`), so a later change to it affects one file.

## How it is tested today

There are no tests for visual assets. The Go suite runs with `go test ./...` from the repository root, and `task ci` runs build, vet, lint and test (`AGENTS.md:13`). A single package runs with `go test ./internal/golden`. None of these touch `assets/` or `.qualitylayer/designs/`. The agent-asset parity script (`node scripts/sync-agent-assets.mjs --check`) concerns skills, not visuals.

## Open questions

Question 7: Whether `assets/logo.png` is pixel-identical to the SVG, and whether any font, brand name or palette is documented outside the files I read, was not checked.

## Code references

- `assets/logo.svg:1-20` logo geometry and colours
- `assets/logo.png` raster logo (binary, unreferenced)
- `README.md:1-3` the only use of the logo
- `.qualitylayer/designs/desktop-app/design.json:1-5`, `.qualitylayer/designs/desktop-app/index.html:1-14` the empty design stub
- `docs/prd/2026-09-28-kiro-cli-kit-multi-team.md` searched for design terms, none found

Exhaustive for the visual assets listed in the question; the repository was searched for logo, colour-library and CSS/HTML mentions.
