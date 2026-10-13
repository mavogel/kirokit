# Deploy Conventions (test fixture)

Fictional sample content for kirokit's own tests. It stands in for the kind of
engineering conventions a team would index as an agent's knowledge base.

## Environments

| Environment | Purpose | Deploys |
| --- | --- | --- |
| dev | Integration and experiments | Automatically on every merge to `main` |
| staging | Production-like verification | Automatically after dev is green |
| prod | Customers | Promoted from staging by the same pipeline, after approval |

The artifact promoted to prod is the one that passed staging. It is never
rebuilt between environments.

## Branching and commits

- Trunk-based: short-lived branches, merged to `main` through a reviewed pull request.
- Commits follow Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, ...).
  A tool reads them to compute the next version and the changelog.
- Releases are tagged `vMAJOR.MINOR.PATCH`. `fix:` bumps the patch, `feat:` the
  minor, and a breaking change (`!` or a `BREAKING CHANGE:` footer) the major.

## The pipeline

1. Build and unit test.
2. Lint, type-check and dependency / secret scanning. A finding blocks the merge.
3. Package the artifact (container image or binary) and sign it.
4. Deploy to dev, run smoke tests, then staging, then prod.

A pull request needs one approving review and a green pipeline. Nobody pushes
directly to `main`.

## Release windows

- No production deploys on Fridays after 14:00 CET or on the day before a public holiday.
- A hotfix for a P1 or P2 may ship at any time with the incident commander's approval.
- Roll out gradually (canary, then a percentage ramp) and watch the dashboard
  between steps; deployments are marked on every graph.

## Rollback

- Roll back by redeploying the previous tagged artifact, not by reverting on the fly.
- Database changes are backward compatible for one release (expand, migrate,
  contract) so the previous version still runs against the new schema.
- If a rollback is needed, say so in the incident channel and open a follow-up ticket.

## Configuration and secrets

- Secrets never live in the repository in plaintext. They are encrypted per team
  and injected when the process starts.
- Non-secret configuration is versioned with the service and overridden per environment.
- Pin tool and dependency versions; bumps are their own reviewed change.
