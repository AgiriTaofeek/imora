# Event Schema

> Status: Research-based, current as of July 2026. Field-level schemas for every event in [event-catalog.md](../02-domain/event-catalog.md). One governing rule constrains all of them, stated first because it isn't optional.

---

## The Governing Rule: Schema Evolution Must Be Additive-Only, Forever

Per [business-rules.md](../02-domain/business-rules.md) BR-1, AccessAuditEvent and related records must remain correctly interpretable for up to 7 years (SOX) or 6 years (HIPAA). A schema change that renames, retypes, or removes a field would make every record written before that change either unreadable or silently misread by whatever code exists years later — which for an audit trail is indistinguishable from data loss, per the standard already set in [deployment-model.md](../04-architecture/deployment-model.md)'s backup requirement.

The fix, per established schema-evolution practice: **every event carries an explicit `schemaVersion` integer, and evolution is restricted to adding new optional fields with default values — never renaming, retyping, or removing an existing field.** If a field is genuinely wrong, it's deprecated (left in place, no longer populated) and superseded by a new field under a new schema version; it is never edited in place. This is a stricter rule than most event-driven systems need, because most event-driven systems don't have a 7-year legal obligation to keep old events meaningfully readable.

---

## Shared Envelope

Every event below carries this envelope in addition to its own fields:

| Field | Type | Required | Notes |
|---|---|---|---|
| `eventId` | UUID | Yes | |
| `schemaVersion` | integer | Yes | Starts at 1; incremented only on additive changes, per the rule above. |
| `eventType` | string | Yes | e.g., `SessionEventCaptured`, matching [event-catalog.md](../02-domain/event-catalog.md)'s names exactly. |
| `occurredAt` | timestamp (UTC) | Yes | |
| `producerContext` | string | Yes | Which bounded context emitted it, per [bounded-contexts.md](../02-domain/bounded-contexts.md). |

---

## Capture and Ingestion Events

### SessionEventCaptured
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `subtype` | enum: FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange | Yes | One event type, per [event-catalog.md](../02-domain/event-catalog.md)'s decision not to multiply event names. |
| `payload` | JSON | Yes | Already masked per BR-7 at capture time — hard-redacted fields are absent entirely, not null; soft-masked fields contain a SecureFieldVault reference, never the raw value. |
| `releaseId` | string | Yes | |

### ErrorEventCaptured
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | No | Null for backend-originated errors correlated only via TraceLink. |
| `stackTraceFingerprint` | string | Yes | The write-time grouping key. |
| `releaseId` | string | Yes | |
| `message` | string | Yes | Subject to the same two-tier masking as SessionEvent payloads. |

### ErrorGrouped
| Field | Type | Required | Notes |
|---|---|---|---|
| `errorGroupId` | UUID | Yes | |
| `errorEventId` | UUID | Yes | |
| `isNewGroup` | boolean | Yes | |

### PerformanceMetricRecorded
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `metricType` | enum: LCP, INP, CLS | Yes | Per [target-users.md](../00-overview/target-users.md)'s thresholds. |
| `value` | float (ms or unitless per metric) | Yes | |
| `releaseId` | string | Yes | |

### RegressionDetected
| Field | Type | Required | Notes |
|---|---|---|---|
| `subjectType` | enum: metric, errorGroup | Yes | |
| `subjectId` | UUID | Yes | |
| `baselineReleaseId` / `newReleaseId` | string | Yes | |
| `previousValue` / `newValue` | float | Yes | |

### ReleaseDeployed
| Field | Type | Required | Notes |
|---|---|---|---|
| `releaseId` | string | Yes | |
| `priorReleaseId` | string | No | Null for the first release. |

### SecuritySignalReceived
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | No | Optional per [domain-model.md](../02-domain/domain-model.md)'s SecurityEvent cardinality. |
| `signalType` | string | Yes | |
| `severity` | enum: low, medium, high, critical | Yes | |

### TraceLinked
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `backendTraceId` / `backendSpanId` | string | Yes | Opaque identifiers — Imora doesn't interpret backend trace internals, per [prd.md](../01-product/prd.md)'s Non-Goals. |

---

## Access and Audit Events

All five share one base shape — a discriminated union on `action`, not five separately-structured events, since [domain-model.md](../02-domain/domain-model.md) defines them as AccessAuditEvent variants:

| Field | Type | Required | Notes |
|---|---|---|---|
| `actorUserId` | UUID | Yes | From `gateway`'s RequestContext, per [component-diagrams.md](../04-architecture/component-diagrams.md). |
| `action` | enum: VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED | Yes | |
| `targetRecordType` / `targetRecordId` | string / UUID | Yes | Polymorphic, per [domain-model.md](../02-domain/domain-model.md)'s relationship table. |
| `sourceIp` / `sourceDevice` | string | Yes | |
| `sequenceNumber` | monotonic integer | Yes | Per the event-sourcing pattern in [domain-model.md](../02-domain/domain-model.md) — enables gap detection, which itself matters for a 7-year audit trail. |
| `reason` | string | Conditionally required | **Required and non-empty when `action = UNMASK`**, per BR-6; absent otherwise. |
| `holdId` | UUID | Conditionally required | Present only when `action = DELETION_SKIPPED`, per [event-catalog.md](../02-domain/event-catalog.md)'s DeletionSkippedDueToHold. |
| `regulatoryBasis` | string | No | Present when a DELETE was governed by a specific RetentionPolicy clock. |

---

## Retention and Compliance Events

### LegalHoldApplied / LegalHoldLifted
| Field | Type | Required | Notes |
|---|---|---|---|
| `holdId` | UUID | Yes | |
| `actorUserId` | UUID | Yes | |
| `scopeQuery` | string | Yes | A query expression, not a fixed record list, per [domain-model.md](../02-domain/domain-model.md). |
| `reason` | string | Yes | |

### RetentionPolicyEvaluated
| Field | Type | Required | Notes |
|---|---|---|---|
| `dataCategory` | string | Yes | |
| `evaluatedCount` / `deletedCount` / `skippedCount` | integer | Yes | Operational aggregate, not per-record. |

### EvidenceExportGenerated
| Field | Type | Required | Notes |
|---|---|---|---|
| `exportId` | UUID | Yes | |
| `actorUserId` | UUID | Yes | |
| `incidentReference` | string | Yes | |
| `contentHash` | string (SHA-256) | Yes | Per BR-4 — the independently verifiable integrity proof. |
| `includedRecordTypes` | array of string | Yes | |

### ErasureRequestResolved
| Field | Type | Required | Notes |
|---|---|---|---|
| `requestId` | UUID | Yes | |
| `dataSubjectId` | string | Yes | |
| `outcome` | enum: FULL_ERASURE, SELECTIVE_PURGE, DENIED | Yes | |
| `regulatoryBasis` | string | Conditionally required | **Required whenever `outcome != FULL_ERASURE`**, per BR-3 — a partial refusal without a cited basis is not a valid record. |
| `fieldsRetained` | array of string | No | Present only for SELECTIVE_PURGE. |

---

## Notification Events

### AlertTriggered
| Field | Type | Required | Notes |
|---|---|---|---|
| `alertType` | string | Yes | |
| `sourceEventId` | UUID | Yes | References the triggering ErrorGrouped or RegressionDetected event. |
| `severity` | enum | Yes | |

### NotificationSent
| Field | Type | Required | Notes |
|---|---|---|---|
| `channel` | string | Yes | |
| `deliveryStatus` | enum: sent, failed, retrying | Yes | |
| `sourceAlertId` | UUID | Yes | |

---

Sources: [Simple patterns for events schema versioning — Event-Driven.io](https://event-driven.io/en/simple_events_versioning_patterns/), [Event versioning strategies for event-driven architectures — theburningmonk.com](https://theburningmonk.com/2025/04/event-versioning-strategies-for-event-driven-architectures/), [Best Practices for Evolving Schemas in Schema Registry — Solace](https://docs.solace.com/Schema-Registry/schema-registry-best-practices.htm).

## What This Feeds Next

`docs/06-data/clickhouse-schema.md` should turn the Capture/Ingestion and Access/Audit event groups above into actual table definitions — both belong in ClickHouse per [container-diagrams.md](../04-architecture/container-diagrams.md). `docs/06-data/postgres-schema.md` covers the entities these events reference but don't duplicate (Session summary, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata).
