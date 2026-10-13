# Platform Domain Knowledge (test fixture)

Fictional sample content for kirokit's own tests. It stands in for the kind of
domain context a team would index as an agent's knowledge base.

## What the platform team owns

- The CI/CD pipelines and shared build runners.
- The Kubernetes platform: clusters, ingress, service mesh, cluster add-ons.
- The observability stack: metrics, logs, traces, alerting, on-call routing.
- Shared libraries and golden-path service templates other teams start from.

The data team owns the warehouse and its pipelines. They consume the platform,
they do not operate it.

## Glossary

| Term | Meaning |
| --- | --- |
| Golden path | The supported, pre-approved way to build and ship a service. Leaving it is allowed but unsupported. |
| Service owner | The single team accountable for a service's reliability. Listed in the service catalogue. |
| Error budget | The unreliability a service may spend per quarter before feature work pauses. |
| Toil | Manual, repetitive operational work that grows with the service. Tracked and reduced, not absorbed. |
| Runbook | A step-by-step procedure for one alert or failure mode. Every paging alert links to one. |
| Blast radius | How much breaks if a change goes wrong. Deploys are sized to keep it small. |

## Incident severity

| Severity | Definition | Response |
| --- | --- | --- |
| P1 | Customer-facing outage or data loss | Page on-call immediately, open an incident channel, update status every 30 minutes |
| P2 | Major degradation, workaround exists | Page on-call, update status hourly |
| P3 | Minor degradation or a single customer affected | Handle in business hours |
| P4 | Cosmetic or no customer impact | Backlog |

Escalate P1/P2 to the on-call rotation before doing anything else. Confirm the
severity and component labels on the ticket first, then search for duplicates.

## How work is tracked

- Tickets live in Jira; the board is triaged each morning.
- Decisions and postmortems live in Confluence under the platform space.
- A postmortem is written within five working days of any P1 or P2, blameless
  in tone, with owners and due dates on every follow-up.
