# 0010. Bounded grace-period local cache for Project Key validation during Redis outages

> Status: Accepted.

## Context

Project Key validation ([ADR 0009](0009-project-key-format.md)) depends on a Redis lookup. `gateway` has to decide what happens to the entire SDK ingest path — every customer, every website, simultaneously — if that cache is unreachable:

- **Fail closed** (reject all writes until Redis recovers): safe by default — no unvalidated event is ever accepted — but turns a Redis outage into a total ingestion outage across every customer at once. For a product whose whole pitch includes "genuinely operable by a 2–3 person team" ([`docs/prd.md`](../../../docs/prd.md) binding constraint #4), a dependency whose failure silently takes down data capture for everyone is a serious operational risk.
- **Fail open** (accept writes without validation while Redis is down): keeps ingestion available through a transient blip, but means an invalid or already-revoked Project Key could write events during that window, unboundedly, for as long as the outage lasts — a real security/compliance gap for a product whose core pitch is provable data control.

Neither extreme actually fits Imora, because it has two binding constraints in tension (Operational Simplicity vs. compliance-first data control), not one clear priority. A third option, based on real precedent, resolves the tension instead of picking a side.

## Decision

**A short-TTL local in-memory cache on `gateway`, populated from successful Redis lookups, with a bounded grace period used only when Redis is unreachable.** Concretely:

- On every successful Redis validation of a Project Key, `gateway` caches `{keyId: validUntil}` locally (in-process), `validUntil` a short duration out (minutes, not hours).
- While Redis is healthy, the local cache is just a performance optimization — Redis remains the source of truth, re-checked normally.
- If Redis becomes unreachable: a key present in the local cache and still within its grace window continues to be accepted. A key not present locally (never seen, or outside the grace window) is rejected.
- This bounds the worst case precisely: a revoked key can remain valid for at most the grace-period duration during an active Redis outage — not indefinitely, and not zero.

Modeled directly on Sentry Relay's own `cache.project_grace_period` mechanism — bounded staleness tolerance, not an unconditional accept/reject.

## Alternatives Considered

- **Pure fail-closed**: rejected — makes Redis a single point of failure for all data capture across every customer, directly conflicting with the Operational Simplicity binding constraint on a single-machine, non-SRE-operated deployment.
- **Pure fail-open**: rejected — an unbounded exposure window during any Redis outage, for however long that outage lasts, is a real compliance-relevant gap this product specifically shouldn't have, given its own pitch is provable data control.
- **Precedent**: Sentry Relay's grace-period cache tolerance for project/key validation during upstream unavailability, and the general industry pattern of local in-memory fallback caches syncing periodically to a central store under high load — both point at bounded local caching as the standard answer to this exact problem, not a binary choice.

## Consequences

- `gateway` needs an in-process cache (map + TTL, or an equivalent lightweight structure) in addition to the Redis lookup — a small but real implementation requirement, not just configuration.
- The grace-period duration is a concrete, tunable number that trades off exposure-window size against availability during an outage — needs an actual default chosen at implementation time (not consequential enough for its own ADR; a reasonable starting point, revisited if real-world Redis reliability data suggests otherwise).
- A newly-created Project Key that has never been successfully validated once is *not* covered by the grace period (nothing to cache yet) — if Redis is down at the exact moment a brand-new key's first event arrives, that event is rejected. Acceptable: this is a narrow, rare edge case, not the common outage scenario (an established customer's already-active key continuing to work through a blip).
- Revocation during an active outage is not instantaneous for keys already in the local cache — a deliberately bounded, sized gap per this ADR, not an oversight.
