# 0010. Fail-open vs. fail-closed when Project Key validation's cache is unreachable

> Status: Proposed — not yet decided, and partly contingent on [0009](0009-project-key-format.md). Context captured so the tradeoff isn't lost.

## Context

If Project Key validation depends on a cache lookup (the "opaque random string" option in [0009](0009-project-key-format.md)), `gateway` has to decide what happens to the entire SDK ingest path — every customer, every website, simultaneously — if that cache (Redis) is unreachable:

- **Fail closed** (reject all writes until the cache recovers): safe by default — no unvalidated event is ever accepted — but turns a cache outage into a total ingestion outage across every customer at once. For a product whose whole pitch includes "genuinely operable by a 2–3 person team," a dependency whose failure silently takes down data capture for everyone is a serious operational risk.
- **Fail open** (accept writes without validation while the cache is down, reconcile/flag later): keeps ingestion available through a transient cache blip, but means an invalid or already-revoked Project Key could write events during that window — a real, if narrow, security gap, and one that's hard to reason about after the fact ("how many unvalidated events did we accept, and were any of them from a revoked key").

Note this question may partly dissolve depending on [0009](0009-project-key-format.md)'s outcome — a self-verifying signed token wouldn't depend on cache availability for the common case at all, which would make this ADR narrower in scope (revocation-check-specific) rather than about the whole validation path.

## Decision

_Not yet made._

## Alternatives Considered

_To be filled in once a decision is made._

## Consequences

_Depends on the decision above._
