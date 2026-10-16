# Security Baseline

Shared security conventions for every team that opts into this steering.

- Never hardcode credentials, API tokens, or connection strings in agent configs, hooks, or prompts. Route them through this kit's SOPS-based secret injection (see the README's "Secrets via SOPS" section).
- Treat every MCP server integration as least-privilege: request only the scopes an agent actually needs.
- Review generated or agent-authored code changes before merging; do not grant blanket auto-approval for write or shell tools without a human in the loop.
