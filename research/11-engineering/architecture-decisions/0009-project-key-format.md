# 0009. Project Key format for SDK/ingest authentication

> Status: Accepted.

## Context

A Project Key (the Sentry-DSN-equivalent credential embedded in a customer's own client-side code, per [`services/gateway/DESIGN.md`](../../../services/gateway/DESIGN.md)) authenticates the write path — potentially the highest-volume traffic in the whole system, since it's driven by every visitor to every customer's website, not by Imora's own user count. It is not a secret in the traditional sense: it ships inside public JavaScript and can always be extracted by anyone who views the page source. Its job is narrowly "prove this write belongs to project X," not "prove you're trusted."

Two ways to construct and validate it:

- **Opaque random string + lookup**: `gateway` checks a cache (Redis) on every request — "does this key exist, is it active." Simple, and revocation is immediate (delete the cache/DB entry). Costs a lookup per event, at whatever the SDK's send volume is.
- **Self-verifying signed token** (e.g., HMAC- or JWT-style, encoding the project ID and signed by a key only the backend holds): `gateway` can validate with zero lookup — pure computation, no cache/DB round-trip at all, which is the cheapest possible fast path. But revoking a specific key before it naturally expires needs its own mechanism (a blocklist, or short-lived tokens re-issued frequently), which reintroduces state and lookup cost for that one case.

## Decision

**Opaque random string, validated via a Redis lookup** — `{keyId} → {projectId, active, createdAt}`. No signature, no self-verification, no embedded "secret" component. Structurally the same shape as Sentry's DSN (public key + project ID, server-side lookup), not a signed-token scheme.

## Alternatives Considered

- **Self-verifying signed token**: rejected, for the same core reason JWT was rejected for sessions in [ADR 0008](0008-gateway-session-model.md) — revocability. A leaked Project Key (expected; it's exposed by design, not by accident) needs to stop working immediately once revoked, not remain valid until a denylist entry is added or the token naturally expires. A signed token without a lookup has no row to delete; a signed token with a revocation-check lookup has paid the complexity of signing/verification for none of the "no round-trip" benefit it was chosen for.
- **A "secret key" component alongside the public identifier** (Sentry's original DSN design): rejected explicitly. Sentry itself deprecated this — a secret shipped inside public client-side JavaScript was never actually secret, and keeping it implied a protection guarantee the design couldn't deliver. Not repeating a mistake the closest comparable product already made and walked back.
- **Precedent**: Sentry's DSN (public key + project ID, server-side lookup, no signature) and Datadog RUM's "client token" (a distinct credential type from their general API key, specifically because API keys can't safely be exposed client-side) — both real, production-scale systems in this exact category, both converged on the same shape this ADR adopts.

## Consequences

- Consistent with ADR 0008's guiding value for this category of decision: revocability outranks avoiding a lookup, applied the same way to both credential types rather than switching philosophy between them.
- The Redis lookup on the write path needs to be genuinely cheap (single key fetch, no joins, cacheable) since this is the highest-volume path in the system — this is a real performance requirement on the implementation, not just a design preference.
- Makes [ADR 0010](0010-gateway-cache-failure-mode.md) (cache-outage failure mode) a fully live, necessary decision rather than one that might have partly dissolved under the signed-token alternative — tackle that next.
- Project Key is its own distinct type, never reused from or convertible to any credential with broader privilege (no read access, no admin capability) — matching Datadog's client-token precedent.
