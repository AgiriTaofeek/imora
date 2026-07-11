# 0001. License the core product, wedge included, under AGPLv3

> Status: Accepted. Full reasoning in [licensing.md](../../01-product/licensing.md).

## Context

Imora's positioning depends on true self-hosting, including the compliance capabilities (access-audit-trail, retention policy, evidence export) that are the entire reason a regulated buyer picks it over an alternative. Two comparators show two ways licensing can quietly undermine that: PostHog gates audit logs specifically behind a commercial `ee/` license; Sentry's Functional Source License (source-available, not OSI-approved) promises feature parity between SaaS and self-hosted but doesn't guarantee it in practice — self-hosted Sentry still lacks reliable session replay.

## Decision

License the entire product — parity and wedge alike — under AGPLv3. No `ee/`-style directory, no commercial license required for any capability. AGPL's network-use clause specifically protects against a cloud provider taking Imora's code and reselling it as a competing SaaS without contributing back, which matters because Imora explicitly competes with SaaS incumbents.

## Alternatives Considered

- **MIT/Apache (permissive):** rejected — creates the exact "a well-funded competitor re-hosts this for free" risk AGPL's network-use clause exists to close.
- **Source-available (Sentry's FSL, HashiCorp's old BSL):** rejected — not OSI-approved, and procurement teams at regulated orgs actively screen license type as an evaluation criterion, per [target-users.md](../../00-overview/target-users.md).
- **Open-core with wedge features gated (PostHog's model):** rejected outright — replicates the exact "pay us or you don't get compliance" dependency regulated buyers are trying to escape by self-hosting in the first place.

## Consequences

- Revenue has to come from Milestone 3 (managed hosting, support SLAs, SSO, multi-region tooling) per [pricing.md](../../01-product/pricing.md) — never from gating core functionality.
- A Community-tier deployment past 50 employees that never signs a contract is a permitted consequence of this choice, not an oversight to patch later.
- The license must be adopted before the first public commit — relicensing later, after a community forms under different expectations, is the costliest version of this decision to get wrong (see HashiCorp Terraform → OpenTofu).
