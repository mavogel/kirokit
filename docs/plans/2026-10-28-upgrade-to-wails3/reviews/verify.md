# Verification review by Codex (gpt-6.1-sol · high)

REVISE: 1 item

1. [.github/workflows/desktop.yaml:100](/Users/mavogel/Developer/projects/kirokit/.github/workflows/desktop.yaml:100): Check `gh release view`’s exit status before treating empty output as no installer. Authentication or network failures currently reach `exit 0`, passing CI without exercising the Windows upgrade check.

   Answer: Taken · the step now asks `gh release list` for the latest published release and `gh release view <tag>` for its assets, and throws when either exits non-zero; only an empty release list (or a release without a Windows setup program) skips the check with a notice. Both queries return `v0.8.1` and `kirokit-desktop_0.8.1_windows_amd64-setup.exe` against mavogel/kirokit; `actionlint` passes.

*Not checked:* Rendered `app-menu.html`; the browser tool refused local file URLs.
