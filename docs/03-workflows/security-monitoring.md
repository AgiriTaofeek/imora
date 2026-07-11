# Security Monitoring

> Status: Adaeze's and Marcus's day, tying stories A1, A2, M1, M2, and J2 together into one workflow — the compliance-driven counterpart to [error-investigation.md](error-investigation.md) and [session-replay.md](session-replay.md)'s engineering-driven ones, built on the same underlying data.

---

## Scenario 1: A DSAR Arrives

Adaeze's 72-hour breach-notification clock and one-month DSAR clock, per [user-personas.md](../01-product/user-personas.md), both start the moment this scenario begins:

1. Query `GET /v1/data-subjects/{id}/sessions` ([rest-api.md](../07-api/rest-api.md)) — every session tied to the data subject, plus the full access history for those sessions, in one lookup. Per story A1, this is minutes, not the multi-hour manual log search it replaces.
2. Export the response in a non-proprietary format (CSV/XML via content negotiation, per [rest-api.md](../07-api/rest-api.md)) — GDPR's actual format requirement, not a PDF screenshot.
3. If the response involves a field currently masked, the same audited UNMASK path applies — Adaeze doesn't get a quieter audit trail than Chidi does, per [authorization.md](../08-security/authorization.md)'s explicit no-exemptions rule.

## Scenario 2: An Investigation Requires Preservation

1. Apply a LegalHold via `POST /v1/legal-holds` ([rest-api.md](../07-api/rest-api.md)), scoped by a structured predicate — session IDs, a data subject, a date range, or an incident reference, per [postgres-schema.md](../06-data/postgres-schema.md) — not a fixed snapshot list, so records created after the hold that still match the scope are automatically covered.
2. Per [threat-model.md](../08-security/threat-model.md)'s finding, an unbounded-scope hold (no date bound, no session-ID list) requires a second approver — a deliberate friction point on the one action in this document that could otherwise become a storage-exhaustion vector against [scaling.md](../04-architecture/scaling.md)'s own math, not an oversight.

## Scenario 3: An Incident Needs a Defensible Record

Per story J2, and picking up directly from where [error-investigation.md](error-investigation.md) leaves off when an incident turns out to involve customer data:

1. `POST /v1/evidence-exports` with the incident reference and relevant session IDs.
2. The result is a frozen, hash-verifiable package — immune to any later retention purge or erasure action touching its source data, per [storage.md](../06-data/storage.md)'s MinIO Object Lock enforcement. This is the artifact that gets handed to legal or a regulator, not a set of screenshots assembled from memory.

## Scenario 4: Annual HIPAA Risk Assessment

Marcus's recurring, not incident-driven, workflow: `GET /v1/sessions/{id}/audit-trail` and its aggregate equivalents produce the standing report §164.312(b) requires — a built-in report against data already being collected, per [authentication.md](../08-security/authentication.md) and [audit-logging.md](../08-security/audit-logging.md), not a custom export job assembled once a year under deadline pressure.

## What Ties These Four Scenarios Together

Every one of them queries the same `AccessAuditEvent` entity from [domain-model.md](../02-domain/domain-model.md) — a DSAR, a HIPAA risk assessment, and an incident evidence package are different views over one underlying audit trail, not four separate systems Adaeze and Marcus have to reconcile by hand. That's the concrete, day-to-day payoff of ADR [0005](../10-engineering/architecture-decisions/0005-unified-access-audit-event.md)'s decision to keep AccessAuditEvent unified.
