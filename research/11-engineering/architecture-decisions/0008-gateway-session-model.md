# 0008. Session model for human/dashboard authentication in `gateway`

> Status: Proposed — not yet decided. Context captured so the tradeoff isn't lost; the Decision section is deliberately empty until this is actually resolved.

## Context

`gateway` handles two categorically different kinds of caller (see [`services/gateway/DESIGN.md`](../../../services/gateway/DESIGN.md)): Project Key traffic (machine-to-machine, from `browser-sdk`) and human/dashboard traffic (individual identity, RBAC, the actor context `query-api`'s `AuditedQueryHandler` depends on for the audit trail).

For the human/dashboard side specifically, the open question is how a logged-in session is represented and validated on every subsequent request:

- **Server-side session store** (a session ID in a cookie, the actual session data — user ID, role, expiry — held in Redis/Postgres): revocation is immediate (delete the row, the session is dead everywhere instantly) — important given this product's own compliance posture (an admin needs to be able to kill a compromised session right away, not wait for a token to expire). Costs a lookup on every authenticated request.
- **Stateless JWT**: no lookup needed to validate — `gateway` (or any service) can verify the signature alone. But revocation before natural expiry requires an explicit blocklist, which reintroduces the same lookup-per-request cost this option was chosen to avoid, just for the revocation path instead of the common path.

## Decision

_Not yet made._

## Alternatives Considered

_To be filled in once a decision is made — both options above are real candidates, not a strawman/preferred-answer setup._

## Consequences

_Depends on the decision above._
