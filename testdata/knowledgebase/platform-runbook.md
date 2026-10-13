# Platform Runbook (test fixture)

Sample knowledge-base content for kirokit's own tests. `kirokit install ... --kb-source`
symlinks a folder like this one into `$KIRO_HOME` and points an agent's
`knowledgeBase` resource at it; the tests use this folder as that source.

- Escalate P1/P2 incidents to the on-call rotation immediately.
- Postmortems live in Confluence under the platform space.
