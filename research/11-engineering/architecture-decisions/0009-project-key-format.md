# 0009. Project Key format for SDK/ingest authentication

> Status: Proposed — not yet decided. Context captured so the tradeoff isn't lost; the Decision section is deliberately empty until this is actually resolved.

## Context

A Project Key (the Sentry-DSN-equivalent credential embedded in a customer's own client-side code, per [`services/gateway/DESIGN.md`](../../../services/gateway/DESIGN.md)) authenticates the write path — potentially the highest-volume traffic in the whole system, since it's driven by every visitor to every customer's website, not by Imora's own user count. It is not a secret in the traditional sense: it ships inside public JavaScript and can always be extracted by anyone who views the page source. Its job is narrowly "prove this write belongs to project X," not "prove you're trusted."

Two ways to construct and validate it:

- **Opaque random string + lookup**: `gateway` checks a cache (Redis) on every request — "does this key exist, is it active." Simple, and revocation is immediate (delete the cache/DB entry). Costs a lookup per event, at whatever the SDK's send volume is.
- **Self-verifying signed token** (e.g., HMAC- or JWT-style, encoding the project ID and signed by a key only the backend holds): `gateway` can validate with zero lookup — pure computation, no cache/DB round-trip at all, which is the cheapest possible fast path. But revoking a specific key before it naturally expires needs its own mechanism (a blocklist, or short-lived tokens re-issued frequently), which reintroduces state and lookup cost for that one case.

## Decision

_Not yet made._

## Alternatives Considered

_To be filled in once a decision is made._

## Consequences

_Depends on the decision above — also determines what [0010](0010-gateway-cache-failure-mode.md) is even about, since a signed-token approach may not depend on a cache being up at all._
