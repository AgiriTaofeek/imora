# Branching Strategy

> Status: Trunk-based development — not a default choice, a specific fit for the monorepo decision in ADR [0003](architecture-decisions/0003-monorepo-structure.md).

---

## Trunk-Based Development

DORA's research on this is specific, not a general preference: elite performers who meet their reliability targets are 2.3x more likely to use trunk-based development, correlated specifically with three-or-fewer active branches, merging to trunk at least daily, and no code-freeze/integration phases.

**Why this matters more than usual for this repository specifically:** `packages/domain-types` is the literal Shared Kernel that `browser-sdk`, `ingestion`, and `query-api` all depend on directly, per [bounded-contexts.md](../02-domain/bounded-contexts.md) and ADR [0003](architecture-decisions/0003-monorepo-structure.md). A long-lived feature branch touching that package accumulates drift risk across three consuming services simultaneously, not just within itself — trunk-based development's short-lived-branch discipline is what keeps that drift window small.

## Practice

- Feature branches live hours to a couple of days, not weeks — work that can't fit in that window gets built behind a feature flag and merged incomplete-but-inert, rather than kept on a long-lived branch.
- No permanent `develop` or `release` branch. Releases are tags cut directly from trunk, per [release-process.md](release-process.md).
- **Hotfixes:** for a critical patch against an already-shipped version (the scenario [release-process.md](release-process.md)'s air-gapped notification gap describes), branch from the affected tag, cherry-pick the fix, tag a patch release, delete the branch. This is a short-lived, purpose-specific branch, not a standing release-maintenance branch.

## What's Deliberately Not Modeled Here

- PR review requirements/approval counts — a team-process decision, not an architectural one.
- Commit message conventions — downstream tooling choice (e.g., whether commit messages drive changelog automation from [release-process.md](release-process.md)).
