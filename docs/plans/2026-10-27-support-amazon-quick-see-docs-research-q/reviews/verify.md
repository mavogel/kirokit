# Verification review by Codex (gpt-6.1-sol · high)

REVISE: 2 items

1. `internal/quick/quick.go:93` — Distinct skill folders can share a frontmatter `name`, receive the same id, and produce indistinguishable agent pins. Reject duplicate frontmatter names before publishing; add a regression test.

   Answer: Taken · `quick.RenderTeam` now fails before publishing with `skills "<a>" and "<b>" share the name "<n>"; Quick matches skills by name`; the case `two folders, one name` in `TestRenderTeam_Errors` failed first and passes now, and `docs/amazon-quick.md` names the error (9b6ff3b).

2. `internal/quick/quick.go:158` — Copying every file as `0600` strips executable bits from vendored scripts that the Kiro renderer preserves. Preserve owner-executable mode and test an executable script fixture.

   Answer: Taken · `writeSkill` keeps the owner-execute bit (0700) as `writeExternalSkills` does, other files stay 0600; `TestRenderTeam_VendoredScriptStaysExecutable` (a vendored `scripts/run.sh` at 0755) failed first and passes now (9b6ff3b).

*Not checked:* Rendered mockup layout; the browser tool rejected its local-file URL. Source inspected; 42 frontend tests and the documentation link check passed.
