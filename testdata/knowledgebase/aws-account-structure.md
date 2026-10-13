# AWS Account Structure (test fixture)

Fictional sample content for kirokit's own tests. The account IDs below are the
placeholder-style numbers AWS uses in its own documentation; none of them is real.

## Organization layout

```
Root
|-- Security OU
|   |-- security-audit        (111111111111)
|   `-- log-archive           (222222222222)
|-- Infrastructure OU
|   |-- shared-services       (333333333333)
|   `-- network               (444444444444)
|-- Workloads OU
|   |-- Platform OU
|   |   |-- platform-dev      (555555555555)
|   |   |-- platform-staging  (666666666666)
|   |   `-- platform-prod     (777777777777)
|   `-- Data OU
|       |-- data-dev          (888888888888)
|       `-- data-prod         (999999999999)
|-- Sandbox OU                (one account per engineer, auto-expiring)
`-- Suspended OU              (closed accounts awaiting deletion)
```

The management account holds only billing and the organization itself. No
workloads, no engineers' day-to-day access.

## What each account is for

| Account | Purpose |
| --- | --- |
| security-audit | Read-only security tooling, the delegated administrator for GuardDuty, Security Hub and Config |
| log-archive | Immutable store for organization CloudTrail and Config logs |
| shared-services | CI runners, artifact registry, shared tooling |
| network | Transit gateway, shared DNS, egress inspection |
| `<team>-<env>` | One account per team and environment, so a mistake in dev cannot touch prod |

## Naming and tagging

- Accounts are named `<team>-<env>`; resources start with the service name.
- Every resource carries `team`, `env`, `cost-center` and `owner` tags. Untagged
  resources are reported weekly and cleaned up if nobody claims them.

## Access

- People sign in through IAM Identity Center with the corporate identity
  provider. There are no long-lived IAM users and no access keys on laptops.
- Permission sets: `ReadOnly`, `Developer`, `PlatformAdmin` and a monitored
  `BreakGlass`. Prod accounts grant `Developer` read access only; changes go
  through the pipeline.
- Sessions last eight hours. Workloads use IAM roles, never static credentials.

## Regions

- Primary region is `eu-central-1`; `eu-west-1` is the disaster-recovery region.
- A service control policy denies every other region except for a short list of
  global services.

## Guardrails (service control policies)

- Deny leaving the organization, closing accounts, or disabling CloudTrail,
  GuardDuty and Config.
- Deny creating IAM users and access keys outside the break-glass path.
- Require encryption at rest for new storage and block public object storage.

## Networking

- Each workload account gets a /20 from the shared range and attaches to the
  transit gateway in the `network` account.
- Traffic to the internet leaves through the central egress inspection point.
  Private service endpoints are preferred over public ones.
