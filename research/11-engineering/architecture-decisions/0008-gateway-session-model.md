# 0008. Session model for human/dashboard authentication in `gateway`

> Status: Accepted.

## Context

`gateway` handles two categorically different kinds of caller (see [`services/gateway/DESIGN.md`](../../../services/gateway/DESIGN.md)): Project Key traffic (machine-to-machine, from `browser-sdk`) and human/dashboard traffic (individual identity, RBAC, the actor context `query-api`'s `AuditedQueryHandler` depends on for the audit trail).

For the human/dashboard side specifically, the open question is how a logged-in session is represented and validated on every subsequent request:

- **Server-side session store** (a session ID in a cookie, the actual session data — user ID, role, expiry — held in Redis/Postgres): revocation is immediate (delete the row, the session is dead everywhere instantly) — important given this product's own compliance posture (an admin needs to be able to kill a compromised session right away, not wait for a token to expire). Costs a lookup on every authenticated request.
- **Stateless JWT**: no lookup needed to validate — `gateway` (or any service) can verify the signature alone. But revocation before natural expiry requires an explicit blocklist, which reintroduces the same lookup-per-request cost this option was chosen to avoid, just for the revocation path instead of the common path.

## Decision

**Server-side session store, backed by Redis.** The cookie holds only a signed, opaque session ID — `HttpOnly`, `Secure`, `SameSite=Strict` — never the session data itself. `gateway` resolves the session ID to `{userId, role, expiry}` via a Redis lookup on every authenticated request to `dashboard`/`query-api`-bound traffic (Domain B only — Project Key validation on the ingest path, Domain A, is a separate question, tracked in [ADR 0009](0009-project-key-format.md)).

## Alternatives Considered

- **Stateless JWT**: rejected. The deciding factor is revocation, not raw cryptographic security — OWASP's own guidance states a properly-configured session cookie is at least as secure as a JWT for a same-origin app, and strictly easier to revoke. Imora's compliance posture makes immediate revocation a product requirement, not a nice-to-have (an admin killing a compromised session right now is part of what Dara/Adaeze/Marcus are actually buying). A JWT without a denylist can't do that; a JWT with a denylist has reintroduced a server-side lookup anyway, at which point it's paid JWT's complexity cost for none of its statelessness benefit.
- **Hybrid access-token + refresh-token pattern** (the 2026 industry default for identity platforms like Auth0/Clerk/Okta): rejected as over-engineered for this specific case. That pattern earns its complexity solving a problem Imora doesn't have — one identity provider serving many client applications across different origins. `dashboard` is same-origin relative to `gateway`/`query-api`; there's no multi-origin token-portability need to justify the added rotation/reuse-detection machinery.
- **Precedent**: Sentry's own self-hosted backend (Django) uses server-side sessions for its dashboard, not JWT — confirmed against Sentry's own configuration docs. Worth noting since Sentry is the exact product Imora is positioned against.

## Consequences

- Redis is already a mandatory dependency in this architecture (rate limiting, legal-hold cache lookups per `docs/architecture.md`) — this decision adds no new infrastructure, just another use of an already-required component.
- `gateway` is stateful with respect to Domain B traffic (session lookups), which is fine — it was never going to be horizontally trivial-stateless for the human-auth path anyway, given RBAC resolution has to happen somewhere.
- MFA (TOTP) fits naturally: the session is only established after both factors succeed, and the session row itself is the single thing that gets revoked to kill access, regardless of how it was originally established.
- Session-store availability becomes a dependency for all dashboard/query-api-bound human traffic — same category of question as [ADR 0010](0010-gateway-cache-failure-mode.md) for the Project Key path, but not the same failure mode: a session-store outage blocks *logging in and continuing to browse*, not the SDK's *data capture*, which is the more severe failure to avoid per `docs/prd.md`'s Operational Simplicity constraint.
