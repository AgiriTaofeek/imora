# 0009. The audit trail is enforced structurally, not by convention

> Status: Accepted. Condensed context in [../design-doc.md](../design-doc.md); full original reasoning is preserved in git history from before the doc-set consolidation.

## Context

The entire compliance pitch (see `pr-faq.md` and `design-doc.md`'s Positioning section) rests on one guarantee: every view of a customer session is logged, no exceptions. Stating "every read must produce an audit event" as a rule that engineers are expected to remember is exactly the failure mode that produces gaps — a defect happens precisely when correctness depends on someone remembering to call a function.

## Decision

`query-api` exposes exactly one way to register a route that reads session data: a wrapper type (`AuditedQueryHandler`) that writes the audit event as an inseparable part of serving the read. There is no code path to register a read handler that skips this wrapper — a future engineer adding a new query endpoint cannot forget the audit step, because there's no way to add the endpoint without it. This applies uniformly regardless of caller — the REST API, the dashboard, and any future integration all produce the same audit event for the same read, because they all go through the same handler.

## Alternatives Considered

- **A convention/code-review checklist ("remember to call the audit logger"):** rejected — this is precisely the failure mode described above. Conventions get missed under deadline pressure; a missing audit entry is discovered only after the fact, if at all.
- **Audit logging enforced in the presentation layer (dashboard):** rejected — anyone calling the underlying API directly would bypass it entirely, making the guarantee true only for one UI, not true for the system.

## Consequences

- Every new read endpoint on session/audit data has to go through `AuditedQueryHandler` by construction — this is a hard constraint on how `query-api` is built, not a style preference.
- This is the single decision that turns "we log access" from a policy statement into a structural guarantee — worth treating as close to immutable; a future refactor that routes around this wrapper for convenience should be treated as a regression, not a cleanup.
