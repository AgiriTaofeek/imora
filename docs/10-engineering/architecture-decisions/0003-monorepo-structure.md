# 0003. Monorepo, with no directory that could become a license-gated carve-out

> Status: Accepted. Full reasoning in [repository-structure.md](../../04-architecture/repository-structure.md).

## Context

`browser-sdk`, `ingestion`, and `query-api` share exact entity definitions from [domain-model.md](../../02-domain/domain-model.md) (a deliberate Shared Kernel relationship per [bounded-contexts.md](../../02-domain/bounded-contexts.md)). Checking how the closest comparator (PostHog) organizes its own monorepo surfaced a warning, not just a precedent: PostHog-FOSS is a separate mirror with "all proprietary code removed" — meaning PostHog's monorepo itself contains the `ee/`-directory license-gating pattern ADR 0001 already rejected.

## Decision

Single monorepo. One `LICENSE` file at the root, applying uniformly to every directory — no per-directory or per-package license overrides, ever. Milestone 3 commercial features (SSO, multi-region tooling) live inside the same `services/gateway/` and `deploy/` directories as everything else, gated at runtime by an offline license file, not by repository structure.

## Alternatives Considered

- **Polyrepo, one repo per service:** rejected — would force the Shared Kernel type definitions into independently-versioned published packages, reintroducing the exact drift risk Shared Kernel was chosen to prevent.
- **Monorepo with an `ee/`-style directory for commercial features (PostHog's pattern):** rejected outright — this is structurally the same anti-pattern ADR 0001 rejected at the licensing level, just implemented at the file-layout level instead.

## Consequences

- A future contributor proposing an `enterprise/` or `ee/` directory contradicts both this ADR and ADR 0001 — that proposal needs to come back to both documents, not be resolved locally.
- `browser-sdk` gets an independent release/versioning cadence (it's a published npm package with external compatibility guarantees) despite living in the same repo — a versioning exception, not a licensing one; it still carries no directory-level `LICENSE` override.
