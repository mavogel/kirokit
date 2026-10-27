# Secrets via SOPS

How MCP servers get their secrets without a secret ever appearing in a template, a rendered file or a golden. Back to the [README](../README.md).

Secrets never appear in plaintext in templates, rendered files or goldens.

1. Each agent needing a secret has `teams/<team>/agents/<agent>/env/secrets.enc.yaml` (flat SOPS-encrypted YAML, safe to commit).
2. `agent.json.tmpl` wraps the MCP server in `sops exec-env` via `sopsCommand` / `sopsArgsJSON`:

   ```
   "command": {{ sopsCommand }},
   "args": {{ sopsArgsJSON (printf "%s/kirokit-secrets/<team>/<agent>/secrets.enc.yaml" .KiroHome) "<real-mcp-command>" "<real-arg>" ... }}
   ```

   The secrets path must be absolute and `.KiroHome`-based, never repo-relative: it resolves under the installed `$KIRO_HOME`.
3. `sops exec-env <file> "<command>"` decrypts and launches the command with the keys as env vars at MCP launch time, not render time, and not via Kiro's `${VAR}` substitution (unreliable).
4. `kirokit install` copies each secrets file to `$KIRO_HOME/kirokit-secrets/<team>/<agent>/secrets.enc.yaml`.

**Windows:** sops runs the command with `cmd.exe /C` there instead of `/bin/sh -c`, and `cmd.exe` cannot take quoted arguments. So on Windows `kirokit install` writes the arguments space-separated and unquoted, and refuses any argument with a space, a tab, `"` or one of `& | < > ^ % ! ( )`, naming the team, agent and argument (`render: <team>/<agent>: argument "…" cannot pass through sops exec-env on Windows (cmd.exe): …`). Write such values without spaces, as mcp-remote documents for headers: `Authorization:${VAR}`, with the space inside the secret. Goldens and `render` always use the macOS/Linux form.

**Remote (HTTP) MCP servers** (e.g. Atlassian `https://mcp.atlassian.com/v2/mcp`) run through the pinned `mcp-remote` proxy under the same wrapper. The secret holds the full header value (`Bearer <key>` for a service-account key, `Basic <base64(email:token)>` for a personal token); the template passes `--header Authorization:${ATLASSIAN_AUTH}` (no space, so it also passes `cmd.exe` on Windows). On macOS and Linux the rendered command single-quotes it, which keeps the placeholder literal; `mcp-remote` expands it from the env sops sets. See `teams/platform-team/agents/devops/agent.json.tmpl`.

Atlassian prerequisites:
- An org admin must enable **Allow API token authentication** in [Atlassian Administration](https://admin.atlassian.com/) (Rovo → Rovo MCP server → Authentication); otherwise every call fails with "You don't have permission to connect via API token".
- Create the token on the [MCP token page](https://id.atlassian.com/manage-profile/security/api-tokens?autofillToken&expiryDays=max&appId=mcp-v2&selectedScopes=all); it needs the "… via Agent Interface" scopes (e.g. `search:jira:agent-interface`), which the generic token page does not offer. Details: [Configure authentication via API token](https://support.atlassian.com/atlassian-rovo-mcp-server/docs/configuring-authentication-via-api-token/).
- Jira Cloud rejects unbounded JQL; restrict searches (e.g. `project = KEY`).

Without `SOPS_AGE_KEY_FILE`, sops reads the age key from `keys.txt` in its config folder: `~/Library/Application Support/sops/age/keys.txt` on macOS, `~/.config/sops/age/keys.txt` on Linux (or `$XDG_CONFIG_HOME/sops/age/keys.txt` on either), `%AppData%\sops\age\keys.txt` on Windows. The desktop app's setup check looks in the same place.

Keys: decrypting needs a SOPS key (age, PGP, KMS) listed as a recipient in `.sops.yaml`. Each team has its own `creation_rules` entry with its age public key(s); a team key cannot decrypt another team's secrets. Never commit private keys (e.g. keep them in `~/.config/sops/age/<team>.txt`).

```bash
SOPS_AGE_KEY_FILE=~/.config/sops/age/platform-team.txt sops teams/platform-team/agents/devops/env/secrets.enc.yaml   # edit
age-keygen -o ~/.config/sops/age/my-team.txt                                                                         # new team key
```

Add an engineer: append their age public key (comma-separated) to the team's rule, then `sops updatekeys <file>` per secrets file. Committed secret values are placeholders; replace via `sops <file>`.

Amazon Quick has no equivalent of this: a Quick bundle never carries an MCP server or a secret (see [Amazon Quick](amazon-quick.md)).
