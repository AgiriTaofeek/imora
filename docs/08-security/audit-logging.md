# Audit Logging

> Status: Research-based, current as of July 2026. Scoped deliberately narrow: this is **operational** audit logging — who changed a policy, who granted a role, who authenticated and how — as distinct from AccessAuditEvent, which is the product-facing data-access trail already fully specified in [domain-model.md](../02-domain/domain-model.md), [event-catalog.md](../02-domain/event-catalog.md), and [clickhouse-schema.md](../06-data/clickhouse-schema.md). This document identifies a real gap in that existing work and closes it, rather than re-describing what's already built.

---

## The Gap: Configuration Changes Were Never Actually Logged

NIST 800-53's AU-2 control requires logging "security or privacy attribute changes" and "changes to user privileges" — not just access to data, but changes to the *rules governing* that access. Checking what's already specified against that standard surfaces a real hole: [event-catalog.md](../02-domain/event-catalog.md) and [clickhouse-schema.md](../06-data/clickhouse-schema.md) log every VIEW, EXPORT, UNMASK, DELETE, and DELETION_SKIPPED against data — but **nothing currently logs who changed a RetentionPolicy's duration, who reclassified a field from PHI to safe, or who granted a role.**

This isn't a cosmetic gap. [business-rules.md](../02-domain/business-rules.md) BR-1's entire guarantee depends on the policy configuration itself being trustworthy — if someone can quietly shorten SessionEvent's retention from 6 years to 30 days with no record of having done so, the audit trail downstream of that change is worthless, and nobody would know to distrust it. The fix has to close this gap in the *same* audit mechanism already built, not bolt on a second one.

## The Fix: Extend AccessAuditEvent, Don't Build a Second Log

Rather than a parallel logging system, **`AccessAuditEvent`'s `action` enum gains one more value: `CONFIG_CHANGED`**, covering RetentionPolicy edits, field-classification changes (per [pii-redaction.md](pii-redaction.md)), and role/permission grants (per [authorization.md](authorization.md)). This keeps [business-rules.md](../02-domain/business-rules.md) BR-5 ("every sensitive action produces exactly one AccessAuditEvent") true without qualification, and means story M1's "one audit report" answers both "who viewed what" and "who changed the rules" from the same query — a HIPAA risk assessment shouldn't require reconciling two separate logs to get the full picture. Payload: `targetRecordType = RetentionPolicy | FieldClassification | UserRole`, `targetRecordId` the specific policy/field/user affected, and — following BR-6's existing pattern for UNMASK — the change itself (old value → new value) recorded in the event payload, not just the fact that *a* change happened.

I've made this concrete by adding `CONFIG_CHANGED` to the action enum in [event-catalog.md](../02-domain/event-catalog.md), [event-schema.md](../06-data/event-schema.md), and [clickhouse-schema.md](../06-data/clickhouse-schema.md) directly, rather than leaving it as a proposal in this document that those files never catch up to.

---

## What Doesn't Fit AccessAuditEvent's Shape: Pre-Authentication Events

AccessAuditEvent assumes a resolved `actorUserId` and a `targetRecordId` — but a **failed login attempt** has neither: authentication didn't succeed, so there's no actor identity yet, and there's no record being accessed, just an attempt. Forcing this into AccessAuditEvent's shape would mean inventing a placeholder actor and a placeholder target, which defeats the purpose of that entity's precision.

**Failed authentication attempts reuse the existing `SecurityEvent` entity instead** (`signalType = "failed_authentication"`, `severity` scaling with repeated failures, `sessionId = null`) — no new entity required. This is the same entity [event-catalog.md](../02-domain/event-catalog.md) already defined for security signals correlated into an incident timeline (story D2), and a repeated-failed-login pattern is exactly the kind of signal that timeline is meant to surface.

---

## System Lifecycle Events

Service start/stop and deployment events (per [deployment-model.md](../04-architecture/deployment-model.md)'s two topology profiles) are operational telemetry, not compliance-relevant by the standard set above — they don't touch data access or policy configuration. These belong in ordinary infrastructure logging (`09-infrastructure/observability.md`), not the AccessAuditEvent/SecurityEvent audit mechanisms this document governs.

---

## What's Deliberately Not Modeled Here

- The exact UI/API surface for triggering a `CONFIG_CHANGED`-producing action — a `07-api/` concern.
- Alerting thresholds on repeated `failed_authentication` SecurityEvents (e.g., account lockout policy) — a product/security-policy decision downstream of this design.

---

Sources: [AU-2: Event Logging — CSF Tools (NIST SP 800-53 Rev. 5)](https://csf.tools/reference/nist-sp-800-53/r5/au/au-2/), [Understanding and Implementing NIST SP 800-53 AU-2 Logging Requirements — SecureStrux](https://securestrux.com/resources/cyber-advisory-center/understanding-and-implementing-nist-sp-800-53-au-2-logging-requirements-for-defense-industrial-base-systems/).

## What This Feeds Next

[threat-model.md](threat-model.md) stress-tests the assumptions across all six files in this folder — including confirming that role grants producing a `CONFIG_CHANGED` event closes the elevation-of-privilege loop this document opened.
