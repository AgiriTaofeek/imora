# Event Catalog

> Status: Research-based, current as of July 2026. Enumerates the concrete domain events implied by [domain-model.md](domain-model.md), [bounded-contexts.md](bounded-contexts.md), and [business-rules.md](business-rules.md), named per the EventStorming convention: past tense, noun/verb, describing a fact that happened — not a command or a request. This is the last domain-layer document before `06-data/event-schema.md` and `07-api/webhooks.md` turn these into wire formats.

Each event lists: **Producer** (owning bounded context, per [bounded-contexts.md](bounded-contexts.md)), **Trigger**, **Key payload**, **Consumers**.

---

## Capture and Ingestion Events (Parity)

### SessionStarted
- **Producer:** browser-sdk
- **Trigger:** A new browsing session begins.
- **Key payload:** sessionId, userId/anonymousId, initial Release tag, device/browser metadata.
- **Consumers:** ingestion (persists), alert-engine (session-count baselines).

### SessionEventCaptured
- **Producer:** browser-sdk
- **Trigger:** Any rrweb-style incremental capture — DOM mutation, mouse move, click, scroll, form input, or viewport change, per [domain-model.md](domain-model.md)'s SessionEvent entity. One event type with a `subtype` field, not seven separate event names, since they share identical routing and storage behavior.
- **Key payload:** sessionId, subtype, timestamp, masked event data (per BR-7 — masking already applied before this event exists).
- **Consumers:** ingestion (persists), query-api (later serves as replay).

### SessionEnded
- **Producer:** browser-sdk
- **Trigger:** Session timeout or explicit page-unload.
- **Key payload:** sessionId, endedAt, total event count.
- **Consumers:** ingestion, workers (becomes eligible for retention-clock evaluation per BR-1 once ended).

### ErrorEventCaptured
- **Producer:** browser-sdk (client errors) or backend instrumentation via TraceLink (correlated backend errors).
- **Trigger:** An unhandled exception or explicit error report.
- **Key payload:** sessionId, stack trace/fingerprint, Release tag, timestamp.
- **Consumers:** ingestion (persists), alert-engine (grouping, below).

### ErrorGrouped
- **Producer:** alert-engine
- **Trigger:** An ErrorEventCaptured event is fingerprint-matched to an existing ErrorGroup, or a new ErrorGroup is created if no match exists — the write-time grouping decision from [domain-model.md](domain-model.md), satisfying story C2 ("one alert per root cause, not one per affected user").
- **Key payload:** errorGroupId, matched sessionId/errorEventId, isNewGroup flag.
- **Consumers:** notification-service (only notified once per group, not per occurrence).

### PerformanceMetricRecorded
- **Producer:** browser-sdk
- **Trigger:** An LCP, INP, or CLS measurement completes for a page view.
- **Key payload:** sessionId, metric type, value, Release tag.
- **Consumers:** ingestion, alert-engine (regression evaluation, below).

### RegressionDetected
- **Producer:** alert-engine
- **Trigger:** A statistically significant change in a Core Web Vitals metric or error rate is attributed to a specific Release — story C1's "which release did this," evaluated at p75 per [target-users.md](../00-overview/target-users.md).
- **Key payload:** metric or errorGroupId, previous Release baseline, new Release value, releaseId.
- **Consumers:** notification-service.

### ReleaseDeployed
- **Producer:** ingestion (via deploy-hook or CI integration — not yet specified further; a `07-api/` concern).
- **Trigger:** A new Release identifier is registered.
- **Key payload:** releaseId, deployedAt, prior releaseId (for regression baseline comparison).
- **Consumers:** alert-engine.

### SecuritySignalReceived
- **Producer:** ingestion
- **Trigger:** A SecurityEvent (anomaly, WAF-style signal) arrives, optionally tied to a sessionId — satisfying story D2's correlation requirement.
- **Key payload:** sessionId (optional), signal type, severity, timestamp.
- **Consumers:** query-api (incident timeline correlation), alert-engine.

### TraceLinked
- **Producer:** ingestion
- **Trigger:** A backend span arrives carrying a propagated session/trace identifier, per story J1's correlation mechanism.
- **Key payload:** sessionId, backend traceId/spanId.
- **Consumers:** query-api (replay-to-trace navigation).

---

## Access and Audit Events (Wedge)

Every event in this section is an AccessAuditEvent variant per [domain-model.md](domain-model.md) — append-only, produced exclusively by `query-api` or `workers` per [bounded-contexts.md](bounded-contexts.md)'s ownership rule, never by `dashboard`.

### SessionViewed
- **Producer:** query-api
- **Trigger:** Any read of a Session's replay or metadata by an authenticated actor.
- **Key payload:** actorUserId, sessionId, timestamp, source IP/device (from gateway's actor context).
- **Consumers:** the audit log itself; surfaces in M1/A1 query responses.

### FieldUnmasked
- **Producer:** query-api
- **Trigger:** An UNMASK action against a masked field, per BR-6.
- **Key payload:** actorUserId, sessionId, field identifier, **reason (required, non-empty)**.
- **Consumers:** audit log; HIPAA risk-assessment reports (M1).

### RecordExported
- **Producer:** query-api (ad hoc export) or workers (EvidenceExport generation).
- **Trigger:** Any EXPORT action, including a full EvidenceExport per story J2.
- **Key payload:** actorUserId, exported record set, exportId (if EvidenceExport), contentHash.
- **Consumers:** audit log.

### RecordDeleted
- **Producer:** workers
- **Trigger:** A scheduled deletion executes under BR-1/BR-2 (no active hold found).
- **Key payload:** targetRecordType, targetRecordId, regulatoryBasis (which RetentionPolicy clock triggered it).
- **Consumers:** audit log.

### DeletionSkippedDueToHold
- **Producer:** workers
- **Trigger:** BR-2's check-before-destroy finds an active LegalHold covering the target record.
- **Key payload:** targetRecordId, holdId, timestamp.
- **Consumers:** audit log — this is the event that makes a skipped deletion visible rather than silent, per BR-2.

### ConfigurationChanged
- **Producer:** query-api or workers, wherever the change is made.
- **Trigger:** A RetentionPolicy edit, a field-classification change (per [pii-redaction.md](../08-security/pii-redaction.md)), or a role/permission grant (per [authorization.md](../08-security/authorization.md)) — the "security or privacy attribute changes" NIST 800-53 AU-2 requires logging, distinct from data access itself. Identified as a gap and closed in [audit-logging.md](../08-security/audit-logging.md).
- **Key payload:** actorUserId, targetRecordType (RetentionPolicy | FieldClassification | UserRole), targetRecordId, oldValue, newValue, timestamp.
- **Consumers:** audit log — without this, BR-1's retention guarantee depends on trusting an unaudited configuration, which defeats the point.

### ErasureRequestReceived / ErasureRequestResolved
- **Producer:** workers (resolution), triggered by a request logged wherever DSAR intake happens (a `07-api/` or future workflow concern, not specified here).
- **Trigger:** A data-subject erasure request enters BR-3's precedence evaluation.
- **Key payload:** ErasureRequestResolved specifically carries: outcome (full erasure / selective purge / denied), regulatory basis cited if not fully honored (Article 17(3)(b) or (e), or the specific overriding statute), fields actually purged vs. retained.
- **Consumers:** audit log; this is Adaeze's evidence that a partial refusal was justified, not arbitrary.

---

## Retention and Compliance Events (Wedge)

### LegalHoldApplied / LegalHoldLifted
- **Producer:** workers, triggered by an authorized actor's request (via gateway's actor context).
- **Trigger:** A hold is placed on or removed from a scope query (per [domain-model.md](domain-model.md)'s LegalHold entity).
- **Key payload:** holdId, appliedBy/liftedBy, scope query, reason.
- **Consumers:** workers itself (every subsequent BR-2 check evaluates against currently-applied holds); audit log.

### RetentionPolicyEvaluated
- **Producer:** workers
- **Trigger:** A scheduled sweep evaluates records against BR-1's regulatory clocks — the step immediately preceding a RecordDeleted or DeletionSkippedDueToHold outcome.
- **Key payload:** dataCategory, evaluatedAt, record count evaluated, outcome counts.
- **Consumers:** audit log (operational, not per-record).

### EvidenceExportGenerated
- **Producer:** workers
- **Trigger:** Story J2's one-click export completes.
- **Key payload:** exportId, incidentReference, frozen record set, contentHash, generatedAt — per BR-4, this event's payload is the permanent, immutable record of what the export contained.
- **Consumers:** RecordExported (audit log entry), the requesting actor.

---

## Notification Events

### AlertTriggered
- **Producer:** alert-engine
- **Trigger:** ErrorGrouped (new group or threshold crossed) or RegressionDetected.
- **Key payload:** alert type, source event reference, severity.
- **Consumers:** notification-service.

### NotificationSent
- **Producer:** notification-service
- **Trigger:** AlertTriggered is translated into a delivery (email, Slack, webhook) per its Conformist relationship to alert-engine in [bounded-contexts.md](bounded-contexts.md).
- **Key payload:** channel, delivery status, sourceAlertId.
- **Consumers:** none within the domain — this is the terminal event in the chain.

---

Sources: [EventStorming Glossary & Cheat Sheet — DDD Crew](https://ddd-crew.github.io/eventstorming-glossary-cheat-sheet/), [Domain events: Design and implementation — Microsoft Learn](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/microservice-ddd-cqrs-patterns/domain-events-design-implementation).

## What This Feeds Next

`docs/06-data/event-schema.md` should turn each event above into a concrete field-level schema. `docs/07-api/webhooks.md` should decide which of these (likely AlertTriggered, RegressionDetected, EvidenceExportGenerated) are exposed externally. The terms this document and its predecessors have been using consistently (Session, AccessAuditEvent, LegalHold, etc.) are formally defined in one place in [glossary.md](../00-overview/glossary.md).
