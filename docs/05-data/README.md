# Data

## Storage

> Status: Research-based, current as of July 2026. The last file in `05-data/` — specifies the object-storage layer for EvidenceExport blobs, referenced by `evidence_exports.object_storage_path` in [Postgres Schema](README.md#postgres-schema), and ties together the three-store picture (ClickHouse, PostgreSQL, object storage) from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams).

---

### The Finding: Object-Lock Compliance Mode Makes BR-4 a Storage-Layer Guarantee, Not Just an Application Promise

[Business Rules](../02-domain/README.md#business-rules) BR-4 requires that an EvidenceExport be immutable once generated — no later retention purge or erasure action may alter it. Everything specified so far ([Domain Model](../02-domain/README.md#domain-model), [Event Catalog](../02-domain/README.md#event-catalog), `contentHash` in [Postgres Schema](README.md#postgres-schema)) enforces this at the application layer: services simply don't write code paths that modify a generated export. That's real, but it's a weaker guarantee than it sounds — it depends on every future engineer never writing that code path.

MinIO (the self-hosted S3-compatible object store from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)) supports **Object Lock in Compliance mode**: true WORM (write-once-read-many) immutability where a locked object **cannot be deleted or modified by anyone, including the root user, for the duration of the lock** — a materially stronger guarantee than Governance mode, which privileged users can still override. **Every EvidenceExport blob is written with Object Lock in Compliance mode**, set for the same duration as its retention period (per [Retention](README.md#retention)'s rule that AccessAuditEvent-adjacent data never expires before the longest applicable clock). This turns BR-4 from a coding convention into a property the storage system itself enforces.

**One useful, narrower side-effect worth noting separately:** MinIO's Object Lock feature also supports an indefinite legal-hold sub-mode, distinct from duration-based retention. For an EvidenceExport blob specifically, applying a [Domain Model](../02-domain/README.md#domain-model) LegalHold could set this native flag directly, rather than relying purely on the application-level check. **This does not replace the dictionary-based mechanism in [ClickHouse Schema](README.md#clickhouse-schema)** — MinIO's object-lock model operates on whole objects in object storage, not rows inside a ClickHouse table, so the two mechanisms cover different parts of the system: object-lock for the comparatively small number of frozen EvidenceExport blobs, the dictionary lookup for the high-volume SessionEvent/ErrorEvent/AccessAuditEvent rows.

---

### Bucket and Key Layout

```
imora-evidence-exports/
  {yyyy}/{mm}/{export_id}.tar.gz
```

Partitioned by year/month of `generated_at`, even though exports are typically held under Object Lock for years — the date partitioning is for eventual lifecycle/storage-tier management (below), not for any deletion logic, since deletion of a locked object is refused by the storage layer regardless of its path.

- **Object content:** a single compressed archive containing the frozen copy of every referenced Session, ErrorEvent, SecurityEvent, and AccessAuditEvent, per [Domain Model](../02-domain/README.md#domain-model)'s EvidenceExport definition.
- **`contentHash`** (stored in `evidence_exports.content_hash`, per [Postgres Schema](README.md#postgres-schema)) is computed over this exact archive — verifying it means re-hashing the object and comparing, independent of Object Lock, so integrity is checkable even by a party who doesn't trust MinIO's lock enforcement alone.

---

### Tiered Storage — Deferred, Not Designed Away

[Retention](README.md#retention) flagged hot/warm/cold tiering as a valid future optimization without specifying it. For object storage specifically: MinIO supports lifecycle transition rules to move objects to cheaper storage classes after an age threshold, which is compatible with Object Lock (the lock governs deletion, not storage class). This is a legitimate later optimization once real export volume exists to justify it — not a day-one requirement, since [Scaling](../03-architecture/README.md#scaling) already identified accumulated *session-replay* storage in ClickHouse as the actual scaling trigger, not evidence-export blob volume, which is comparatively rare (generated per-incident, not per-session).

---

### Air-Gapped Consistency

MinIO runs self-hosted within the deployment boundary in both topology profiles from [Deployment Model](../03-architecture/README.md#deployment-model) — no cloud object-storage dependency, consistent with [System Context](../03-architecture/README.md#system-context)'s requirement that every Parity and Wedge capability, including evidence export, works with zero external systems present.

---

### What's Deliberately Not Modeled Here

- IAM/bucket policy specifics for who can initiate an export or read a generated one — `07-security/README.md#authorization`.
- Exact archive format (tar.gz vs. a structured container) — an implementation detail once this design is accepted.
- Lifecycle transition rule configuration — deferred per the Tiered Storage section above, pending real usage data.

---

Sources: [Object Locking and Immutability — MinIO Documentation](https://docs.min.io/aistor/administration/object-locking-and-immutability/), [Object Locking, Versioning, Legal Holds and Modes in MinIO](https://blog.min.io/object-locking-versioning-and-holds-in-minio/), [MinIO Object Locking — MinIO Object Storage (AGPLv3)](https://docs.min.io/community/minio-object-store/administration/object-management/object-retention.html).

### What This Closes Out

This is the last file in `docs/05-data/`. All five files — [Event Schema](README.md#event-schema), [ClickHouse Schema](README.md#clickhouse-schema), [Postgres Schema](README.md#postgres-schema), [Retention](README.md#retention), and this one — are now internally consistent, and the object-storage layer's Compliance-mode lock is a concrete answer to the immutability question `README.md#deployment-model`'s backup section raised without fully resolving. `docs/07-security/` (authorization, encryption, audit-logging) and `docs/06-api/` are the natural next tiers — the former specifies who can trigger the actions this folder's schemas record.

---

## Event Schema

> Status: Research-based, current as of July 2026. Field-level schemas for every event in [Event Catalog](../02-domain/README.md#event-catalog). One governing rule constrains all of them, stated first because it isn't optional.

---

### The Governing Rule: Schema Evolution Must Be Additive-Only, Forever

Per [Business Rules](../02-domain/README.md#business-rules) BR-1, AccessAuditEvent and related records must remain correctly interpretable for up to 7 years (SOX) or 6 years (HIPAA). A schema change that renames, retypes, or removes a field would make every record written before that change either unreadable or silently misread by whatever code exists years later — which for an audit trail is indistinguishable from data loss, per the standard already set in [Deployment Model](../03-architecture/README.md#deployment-model)'s backup requirement.

The fix, per established schema-evolution practice: **every event carries an explicit `schemaVersion` integer, and evolution is restricted to adding new optional fields with default values — never renaming, retyping, or removing an existing field.** If a field is genuinely wrong, it's deprecated (left in place, no longer populated) and superseded by a new field under a new schema version; it is never edited in place. This is a stricter rule than most event-driven systems need, because most event-driven systems don't have a 7-year legal obligation to keep old events meaningfully readable.

---

### Shared Envelope

Every event below carries this envelope in addition to its own fields:

| Field | Type | Required | Notes |
|---|---|---|---|
| `eventId` | UUID | Yes | |
| `schemaVersion` | integer | Yes | Starts at 1; incremented only on additive changes, per the rule above. |
| `eventType` | string | Yes | e.g., `SessionEventCaptured`, matching [Event Catalog](../02-domain/README.md#event-catalog)'s names exactly. |
| `occurredAt` | timestamp (UTC) | Yes | |
| `producerContext` | string | Yes | Which bounded context emitted it, per [Bounded Contexts](../02-domain/README.md#bounded-contexts). |

---

### Capture and Ingestion Events

#### SessionEventCaptured
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `subtype` | enum: FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange | Yes | One event type, per [Event Catalog](../02-domain/README.md#event-catalog)'s decision not to multiply event names. |
| `payload` | JSON | Yes | Already masked per BR-7 at capture time — hard-redacted fields are absent entirely, not null; soft-masked fields contain a SecureFieldVault reference, never the raw value. |
| `releaseId` | string | Yes | |

#### ErrorEventCaptured
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | No | Null for backend-originated errors correlated only via TraceLink. |
| `stackTraceFingerprint` | string | Yes | The write-time grouping key. |
| `releaseId` | string | Yes | |
| `message` | string | Yes | Subject to the same two-tier masking as SessionEvent payloads. |

#### ErrorGrouped
| Field | Type | Required | Notes |
|---|---|---|---|
| `errorGroupId` | UUID | Yes | |
| `errorEventId` | UUID | Yes | |
| `isNewGroup` | boolean | Yes | |

#### PerformanceMetricRecorded
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `metricType` | enum: LCP, INP, CLS | Yes | Per [Target Users](../00-overview/README.md#target-users)'s thresholds. |
| `value` | float (ms or unitless per metric) | Yes | |
| `releaseId` | string | Yes | |

#### RegressionDetected
| Field | Type | Required | Notes |
|---|---|---|---|
| `subjectType` | enum: metric, errorGroup | Yes | |
| `subjectId` | UUID | Yes | |
| `baselineReleaseId` / `newReleaseId` | string | Yes | |
| `previousValue` / `newValue` | float | Yes | |

#### ReleaseDeployed
| Field | Type | Required | Notes |
|---|---|---|---|
| `releaseId` | string | Yes | |
| `priorReleaseId` | string | No | Null for the first release. |

#### SecuritySignalReceived
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | No | Optional per [Domain Model](../02-domain/README.md#domain-model)'s SecurityEvent cardinality. |
| `signalType` | string | Yes | |
| `severity` | enum: low, medium, high, critical | Yes | |

#### TraceLinked
| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionId` | UUID | Yes | |
| `backendTraceId` / `backendSpanId` | string | Yes | Opaque identifiers — Imora doesn't interpret backend trace internals, per [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)'s Non-Goals. |

---

### Access and Audit Events

All six share one base shape — a discriminated union on `action`, not six separately-structured events, since [Domain Model](../02-domain/README.md#domain-model) defines them as AccessAuditEvent variants:

| Field | Type | Required | Notes |
|---|---|---|---|
| `actorUserId` | UUID | Yes | From `gateway`'s RequestContext, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams). |
| `action` | enum: VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED, CONFIG_CHANGED | Yes | `CONFIG_CHANGED` added per [Audit Logging](../07-security/README.md#audit-logging) — closes the gap where policy/role/classification changes went unaudited. |
| `targetRecordType` / `targetRecordId` | string / UUID | Yes | Polymorphic, per [Domain Model](../02-domain/README.md#domain-model)'s relationship table. For `CONFIG_CHANGED`, `targetRecordType` is one of RetentionPolicy, FieldClassification, or UserRole. |
| `sourceIp` / `sourceDevice` | string | Yes | |
| `sequenceNumber` | monotonic integer | Yes | Per the event-sourcing pattern in [Domain Model](../02-domain/README.md#domain-model) — enables gap detection, which itself matters for a 7-year audit trail. |
| `reason` | string | Conditionally required | **Required and non-empty when `action = UNMASK`**, per BR-6; absent otherwise. |
| `holdId` | UUID | Conditionally required | Present only when `action = DELETION_SKIPPED`, per [Event Catalog](../02-domain/README.md#event-catalog)'s DeletionSkippedDueToHold. |
| `regulatoryBasis` | string | No | Present when a DELETE was governed by a specific RetentionPolicy clock. |
| `oldValue` / `newValue` | string (JSON) | Conditionally required | **Required when `action = CONFIG_CHANGED`** — the change itself, not just the fact that one occurred, per [Audit Logging](../07-security/README.md#audit-logging). |

---

### Retention and Compliance Events

#### LegalHoldApplied / LegalHoldLifted
| Field | Type | Required | Notes |
|---|---|---|---|
| `holdId` | UUID | Yes | |
| `actorUserId` | UUID | Yes | |
| `scopeQuery` | string | Yes | A query expression, not a fixed record list, per [Domain Model](../02-domain/README.md#domain-model). |
| `reason` | string | Yes | |

#### RetentionPolicyEvaluated
| Field | Type | Required | Notes |
|---|---|---|---|
| `dataCategory` | string | Yes | |
| `evaluatedCount` / `deletedCount` / `skippedCount` | integer | Yes | Operational aggregate, not per-record. |

#### EvidenceExportGenerated
| Field | Type | Required | Notes |
|---|---|---|---|
| `exportId` | UUID | Yes | |
| `actorUserId` | UUID | Yes | |
| `incidentReference` | string | Yes | |
| `contentHash` | string (SHA-256) | Yes | Per BR-4 — the independently verifiable integrity proof. |
| `includedRecordTypes` | array of string | Yes | |

#### ErasureRequestResolved
| Field | Type | Required | Notes |
|---|---|---|---|
| `requestId` | UUID | Yes | |
| `dataSubjectId` | string | Yes | |
| `outcome` | enum: FULL_ERASURE, SELECTIVE_PURGE, DENIED | Yes | |
| `regulatoryBasis` | string | Conditionally required | **Required whenever `outcome != FULL_ERASURE`**, per BR-3 — a partial refusal without a cited basis is not a valid record. |
| `fieldsRetained` | array of string | No | Present only for SELECTIVE_PURGE. |

---

### Notification Events

#### AlertTriggered
| Field | Type | Required | Notes |
|---|---|---|---|
| `alertType` | string | Yes | |
| `sourceEventId` | UUID | Yes | References the triggering ErrorGrouped or RegressionDetected event. |
| `severity` | enum | Yes | |

#### NotificationSent
| Field | Type | Required | Notes |
|---|---|---|---|
| `channel` | string | Yes | |
| `deliveryStatus` | enum: sent, failed, retrying | Yes | |
| `sourceAlertId` | UUID | Yes | |

---

Sources: [Simple patterns for events schema versioning — Event-Driven.io](https://event-driven.io/en/simple_events_versioning_patterns/), [Event versioning strategies for event-driven architectures — theburningmonk.com](https://theburningmonk.com/2025/04/event-versioning-strategies-for-event-driven-architectures/), [Best Practices for Evolving Schemas in Schema Registry — Solace](https://docs.solace.com/Schema-Registry/schema-registry-best-practices.htm).

### What This Feeds Next

`docs/05-data/README.md#clickhouse-schema` should turn the Capture/Ingestion and Access/Audit event groups above into actual table definitions — both belong in ClickHouse per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams). `docs/05-data/README.md#postgres-schema` covers the entities these events reference but don't duplicate (Session summary, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata).

---

## ClickHouse Schema

> Status: Research-based, current as of July 2026. Table definitions for the five ClickHouse-resident entities from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams): SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, and AccessAuditEvent — all high-volume, append-only, per [Event Schema](README.md#event-schema)'s payload shapes.

---

### The Tension This Document Has to Resolve First

ClickHouse is an OLAP columnar store optimized for inserts and aggregate reads, not row-level deletes — the efficient retention pattern is TTL-based expiry aligned to the partition key, which drops entire partitions as a cheap metadata operation (`ttl_only_drop_parts`). But [Business Rules](../02-domain/README.md#business-rules) BR-2 requires checking each *individual* record against active legal holds immediately before it's deleted — a per-row concern, not a per-partition one. A naive implementation would force a choice between "efficient retention" and "correct legal-hold enforcement." It doesn't have to.

**Resolution:** ClickHouse's TTL clause natively supports conditional `WHERE` expressions — `TTL occurredAt + INTERVAL 6 YEAR DELETE WHERE <condition>` is documented, standard syntax. Rather than a mutable per-row flag column (which would require expensive row-level UPDATEs — precisely the operation this whole problem is trying to avoid), the hold check is expressed as a **dictionary lookup**: a ClickHouse Dictionary of currently-active hold scopes, refreshed periodically from Postgres's LegalHold table (the source of truth per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)), queried inline in the TTL expression via `dictGetOrDefault('active_legal_holds', 'is_held', sessionId, 0) = 0`. This is the standard ClickHouse mechanism for exactly this shape of problem — checking every row against a small, frequently-changing reference set without paying for a join or an update on every row.

**One operational consequence worth stating plainly:** a single held row inside a partition prevents `ttl_only_drop_parts` from cheaply dropping that whole partition — the partition falls back to slower row-level TTL evaluation until the hold lifts. This is exactly why the partitioning scheme below uses **monthly**, not yearly, partitions: it limits the "blast radius" of what one legal hold can pin from cheap expiry to a month's worth of data, not a year's.

---

### Common Design Decisions

- **Engine family:** MergeTree (specifically plain `MergeTree`, not `ReplacingMergeTree` — these tables are genuinely append-only per [Domain Model](../02-domain/README.md#domain-model)'s event-sourcing pattern; there is no "latest version wins" semantic to support).
- **Partitioning:** monthly, by `occurredAt`, on every table below — balances the cheap-partition-drop benefit against the legal-hold blast-radius concern above.
- **Deletion mechanism by scenario:** scheduled RetentionPolicy sweeps (BR-1) use TTL-with-dictionary-lookup, above. A prompt, individual GDPR erasure request (BR-3, story A2) that can't wait for the next background merge uses a targeted **lightweight DELETE** instead — 10–100x faster than a full mutation for a single-record removal, at the cost of requiring an eventual background merge to reclaim the space, which is an acceptable tradeoff for a rare, individual action rather than the routine sweep path.

---

### Table Definitions

#### session_events
| Column | Type | Notes |
|---|---|---|
| `session_id` | UUID | |
| `subtype` | Enum8 | FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange, per [Event Schema](README.md#event-schema). |
| `occurred_at` | DateTime64 | Partition and TTL basis. |
| `payload` | String (JSON) | Already masked per BR-7 at capture time. |
| `release_id` | String | |
| **ORDER BY** | `(session_id, occurred_at)` | Replay reconstruction is always "give me this session's events in order" — the dominant query shape. |
| **TTL** | `occurred_at + INTERVAL <category clock> DELETE WHERE dictGetOrDefault('active_legal_holds', 'is_held', session_id, 0) = 0` | Category clock per BR-1's RetentionPolicy for this data category. |

#### error_events
| Column | Type | Notes |
|---|---|---|
| `error_event_id` | UUID | |
| `session_id` | Nullable(UUID) | Per [Event Schema](README.md#event-schema). |
| `stack_trace_fingerprint` | String | Write-time grouping key. |
| `release_id` | String | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(stack_trace_fingerprint, occurred_at)` | ErrorGroup lookups are the dominant read pattern. |
| **TTL** | Same pattern as session_events, keyed to error-data's RetentionPolicy category. |

#### performance_metrics
| Column | Type | Notes |
|---|---|---|
| `session_id` | UUID | |
| `metric_type` | Enum8 | LCP, INP, CLS |
| `value` | Float64 | |
| `release_id` | String | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(release_id, metric_type, occurred_at)` | Regression detection (story C1) queries by release and metric first. |
| **TTL** | Same pattern, own category clock — performance data typically has a shorter regulatory floor than session replay, per BR-1. |

#### security_events
| Column | Type | Notes |
|---|---|---|
| `security_event_id` | UUID | |
| `session_id` | Nullable(UUID) | |
| `signal_type` | String | |
| `severity` | Enum8 | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(session_id, occurred_at)` | Incident-timeline correlation (story D2) is the dominant read shape. |
| **TTL** | Same conditional pattern. |

#### access_audit_events
| Column | Type | Notes |
|---|---|---|
| `event_id` | UUID | |
| `sequence_number` | UInt64 | Monotonic, per [Domain Model](../02-domain/README.md#domain-model)'s gap-detection requirement. |
| `actor_user_id` | UUID | |
| `action` | Enum8 | VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED, CONFIG_CHANGED |
| `target_record_type` | String | For `CONFIG_CHANGED`: RetentionPolicy, FieldClassification, or UserRole, per [Audit Logging](../07-security/README.md#audit-logging). |
| `target_record_id` | UUID | |
| `source_ip` | String | |
| `reason` | Nullable(String) | Required non-null when `action = UNMASK`, enforced at the `query-api`/`workers` write path (BR-6) — ClickHouse itself doesn't enforce conditional nullability, the writing service does. |
| `old_value` / `new_value` | Nullable(String) | Required non-null when `action = CONFIG_CHANGED` — same enforced-at-write-path pattern as `reason` above. |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(target_record_id, occurred_at)` | "Who viewed this specific record" (stories M1, A1) is the dominant read shape — this ordering is what makes Adaeze's DSAR query and Marcus's risk-assessment report both fast. |
| **TTL** | **Longest applicable regulatory clock across all categories the deployment operates under** (per BR-1's longest-wins rule) — this table's retention floor is never shorter than any other table's, since an audit record about a since-deleted SessionEvent may still need to prove the deletion happened correctly. |

**A note on `access_audit_events` never being younger than what it audits:** if SessionEvent is deleted under a 12-month PCI-DSS clock but its corresponding `RecordDeleted` AccessAuditEvent were also deleted on that same clock, the system would lose the ability to prove the deletion was lawful — exactly the audit-control failure [Business Rules](../02-domain/README.md#business-rules) exists to prevent. `access_audit_events`' own TTL is therefore set to the *longest* clock in use, not the clock of whatever it's auditing.

---

### What's Deliberately Not Modeled Here

- Exact `CREATE DICTIONARY` syntax for `active_legal_holds` and its refresh interval — an implementation detail once this design is accepted, not an architecture decision.
- Compression codecs, index granularity tuning, or materialized views for common aggregations — performance tuning downstream of this schema, not part of it.
- The Postgres-side LegalHold, RetentionPolicy, Session, Release, ErrorGroup, and EvidenceExport tables — `README.md#postgres-schema`, next.

---

Sources: [Manage data with TTL — ClickHouse Docs](https://clickhouse.com/docs/guides/developer/ttl), [Support of WHERE and GROUP BY in TTL expressions — ClickHouse GitHub PR #10537](https://github.com/ClickHouse/ClickHouse/pull/10537), [Lightweight delete — ClickHouse Docs](https://clickhouse.com/docs/guides/developer/lightweight-delete), [What Is the Difference Between Mutations and Lightweight Deletes in ClickHouse](https://oneuptime.com/blog/post/2026-03-31-clickhouse-mutations-vs-lightweight-deletes/view).

### What This Feeds Next

`docs/05-data/README.md#postgres-schema` covers the relational side. `docs/05-data/README.md#retention` should turn BR-1's per-category regulatory clocks into the actual `INTERVAL` values used in the TTL clauses above.

---

## Postgres Schema

> Status: Table definitions for the relational side of the dual-store split from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) — Session summaries, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata, and user/RBAC config. Small-cardinality, transactionally-updated data, as opposed to [ClickHouse Schema](README.md#clickhouse-schema)'s high-volume append-only side.

---

### Closing the Loop with README.md#clickhouse-schema: LegalHold Scope Must Be a Re-Evaluated Predicate, Not a Snapshot

[Domain Model](../02-domain/README.md#domain-model) specified `LegalHold.scope` as a query, not a fixed record list, deliberately — a hold on "all sessions tied to incident X" should cover sessions created *after* the hold was applied, if the same compromised account or ongoing incident keeps generating them. A join table (`legal_hold_id, target_record_id`) populated once at hold-creation time would silently miss those later records — a real compliance gap, not a theoretical one.

The resolution: `scope` is stored as a **structured JSONB predicate**, not a fixed list and not a raw SQL string (which would be both a SQL-injection surface and something [ClickHouse Schema](README.md#clickhouse-schema)'s dictionary refresh job has no safe way to evaluate). The predicate is **re-evaluated against current data on every dictionary refresh cycle**, producing the current resolved set of held `target_record_id`s at refresh time — which is what makes new matching records get picked up automatically, and is the concrete mechanism behind [ClickHouse Schema](README.md#clickhouse-schema)'s `active_legal_holds` dictionary.

Supported predicate types, kept deliberately small rather than an open-ended query language:

```json
{"type": "session_ids", "values": ["uuid1", "uuid2"]}
{"type": "data_subject", "value": "user-identifier"}
{"type": "date_range", "start": "2026-01-01", "end": "2026-06-30"}
{"type": "incident_reference", "value": "INC-1234"}
```

A structured, closed set of predicate types — rather than arbitrary SQL — keeps the refresh job's evaluation logic bounded and auditable, which matters for a mechanism gating deletion of compliance-critical data.

---

### Common Design Decisions

- **Referential integrity is enforced within Postgres**, standard foreign keys between the tables below.
- **Referential integrity is *not* enforced across the store boundary.** `error_events.error_group_id` (ClickHouse) references `error_groups.id` (Postgres), and `session_events.session_id` (ClickHouse) references `sessions.id` (Postgres) — neither is a database-level constraint, since ClickHouse doesn't support foreign keys to another engine. This is a deliberate consistency-model tradeoff of the dual-store architecture from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams), not an oversight: `alert-engine` and `ingestion` are responsible for creating the Postgres row before or atomically with the corresponding ClickHouse write, per the Shared Kernel relationship in [Bounded Contexts](../02-domain/README.md#bounded-contexts).

---

### Table Definitions

#### sessions
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | Matches `session_events.session_id` in ClickHouse. |
| `user_or_anonymous_id` | String | |
| `started_at` / `ended_at` | Timestamp | `ended_at` nullable until SessionEnded fires. |
| `release_id` | FK → releases.id | |
| `device_metadata` | JSONB | |

#### releases
| Column | Type | Notes |
|---|---|---|
| `id` | String (PK) | |
| `deployed_at` | Timestamp | |
| `prior_release_id` | FK → releases.id, nullable | Regression baseline comparison, story C1. |

#### error_groups
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `fingerprint` | String, unique | The write-time grouping key `alert-engine` matches against. |
| `first_seen_at` / `last_seen_at` | Timestamp | |
| `occurrence_count` | Integer | Denormalized counter, updated by `alert-engine` on each `ErrorGrouped` — avoids a ClickHouse aggregate query on every dashboard load for story C2's grouping display. |

#### retention_policies
| Column | Type | Notes |
|---|---|---|
| `data_category` | String (PK) | SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, AccessAuditEvent. |
| `retention_period` | Interval | The `INTERVAL` value [ClickHouse Schema](README.md#clickhouse-schema)'s TTL clauses read. |
| `regulatory_basis` | String | Which clock governs — PCI-DSS/HIPAA/GDPR/SOX, per BR-1. |

#### legal_holds
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `applied_by` | FK → users.id | |
| `scope` | JSONB | The re-evaluated predicate, above. |
| `reason` | String | |
| `applied_at` / `lifted_at` | Timestamp | `lifted_at` nullable while active — **this column, not a separate boolean, is what the dictionary refresh job filters on** (`WHERE lifted_at IS NULL`), so there's exactly one place "is this hold active" is determined. |

#### evidence_exports
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `requested_by` | FK → users.id | |
| `incident_reference` | String | |
| `content_hash` | String (SHA-256) | Per BR-4 — verifies the object-storage blob hasn't been altered. |
| `object_storage_path` | String | Pointer to the frozen blob in MinIO, per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) — the export's actual content lives in object storage, not Postgres. |
| `generated_at` | Timestamp | |
| `included_record_types` | JSONB array | |

#### users
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `role` | Enum: engineer, compliance_officer, platform_operator, admin | Maps to the four actor types in [System Context](../03-architecture/README.md#system-context). |
| `sso_subject_id` | String, nullable | Populated only for Enterprise-tier SSO logins, per [Pricing](../01-product/README.md#pricing) — null for Community/Team tier local accounts. |

Fine-grained field-level permissions (as opposed to the coarse role above) are a `07-security/README.md#authorization` concern, not this table's — `role` here is enough to route a request, not enough to decide whether a specific field should be masked.

---

### What's Deliberately Not Modeled Here

- RBAC permission matrices and field-level access rules — `07-security/README.md#authorization`.
- Actual migration tooling or ORM choice — an implementation decision.
- Index tuning beyond the primary/foreign keys implied above — performance work downstream of this schema.

---

### What This Feeds Next

`docs/05-data/README.md#retention` should formalize the `retention_period`/`regulatory_basis` values in `retention_policies` against BR-1's actual regulatory clocks. `docs/05-data/README.md#storage` should specify the object-storage (MinIO) layout that `evidence_exports.object_storage_path` points into.

---

## Retention

> Status: Research-based, current as of July 2026. Assigns an actual retention period to each of the five data categories in `retention_policies` ([Postgres Schema](README.md#postgres-schema)), feeding the `INTERVAL` values in [ClickHouse Schema](README.md#clickhouse-schema)'s TTL clauses. Restating BR-1's regulatory table isn't enough — not every category has a regulatory floor, and treating them as if they all do would itself violate GDPR's storage-limitation principle from the opposite direction.

---

### The Split This Document Has to Make Explicit

[Business Rules](../02-domain/README.md#business-rules) BR-1 covers what happens when regulations *conflict*. It doesn't establish that every data category is regulated in the first place. SessionEvent and AccessAuditEvent plausibly contain or reference PII/PHI and are squarely regulation-driven. But **ErrorEvent and PerformanceMetric have no comparable regulatory floor** — nobody is legally required to keep a stack trace for six years, and industry practice for diagnostic/error data runs 30–90 days by default, well short of any compliance clock. Applying a blanket multi-year retention to categories nothing requires it for isn't caution, it's the same storage-limitation violation GDPR Article 5(1)(e) prohibits in the other direction — keeping data longer than necessary for its actual purpose — and it directly worsens the accumulated-storage scaling problem identified in [Scaling](../03-architecture/README.md#scaling).

---

### Category Assignments

| Category | Driver | Default | Configurable? |
|---|---|---|---|
| **SessionEvent** | Regulation | Set to the strictest regulation the deployment's own industry requires — HIPAA's 6-year floor for healthcare, PCI-DSS's 12-month floor for payment-touching flows, or GDPR's purpose-bound limit otherwise (see below) | Yes — this is the per-category configurability that is the entire point of the wedge, per [Competitive Analysis](../00-overview/README.md#competitive-analysis) |
| **ErrorEvent** | Operational | 90 days by default, matching industry-standard diagnostic-log retention — nobody debugs a 6-year-old stack trace | Yes, but a shorter default than the regulated categories is the deliberate starting point, not an oversight |
| **PerformanceMetric** | Operational | 13 months by default — long enough for year-over-year Core Web Vitals comparison (the same period Datadog defaults to for general telemetry), short enough to avoid unnecessary accumulation | Yes |
| **SecurityEvent** | Regulation (usually) | Aligned to the same clock as SessionEvent when correlated to a session (story D2's incident-timeline requirement means a security signal is only as useful as the session context it's tied to); PCI-DSS's 12-month floor as a baseline when uncorrelated | Yes |
| **AccessAuditEvent** | Regulation, and never shorter than any other category | The **longest** clock among all categories the deployment operates under, per [ClickHouse Schema](README.md#clickhouse-schema)'s finding — this table proves what happened to every other table, so it must outlive all of them | Not independently below the computed longest-clock floor — this one constraint isn't optional |

---

### Translating "GDPR Has No Fixed Term" Into an Actual TTL Value

GDPR's Article 5(1)(e) storage-limitation principle doesn't specify a duration — but [ClickHouse Schema](README.md#clickhouse-schema)'s TTL mechanism needs a concrete `INTERVAL`, not an open-ended "as long as necessary." This isn't a contradiction: **GDPR doesn't prohibit setting a fixed enforced ceiling — it requires that ceiling be justified by actual processing necessity, not that no ceiling exist.** In practice, a deployment operating primarily under GDPR (no HIPAA/PCI-DSS/SOX floor applying) sets a concrete SessionEvent retention value — commonly 12–24 months for a customer-support/debugging purpose — and records the justification in `retention_policies.regulatory_basis`, per [Postgres Schema](README.md#postgres-schema). The DPO persona (Adaeze, per [User Personas](../01-product/README.md#user-personas)) is the one who sets and can justify that number to a regulator; the system enforces whatever she configures, it doesn't invent the number for her.

---

### Why Getting the Operational Defaults Right Matters Beyond Compliance Hygiene

Per [Scaling](../03-architecture/README.md#scaling)'s finding, accumulated retention is Imora's actual scaling trigger — not ingestion throughput. A deployment that defaults ErrorEvent and PerformanceMetric to the same 6-year HIPAA floor as SessionEvent, out of an abundance of caution, would inflate its accumulated-storage multiplier well past the 2–3× estimate that calculation assumed, pulling the single-machine-to-cluster migration threshold forward by years for no compliance benefit at all. The category split above isn't just correctness — it's what keeps [Deployment Model](../03-architecture/README.md#deployment-model)'s single-machine promise to Priya's persona intact for as long as possible.

---

### What's Deliberately Not Modeled Here

- The actual configuration UI/workflow a deployment operator uses to set these values per organization — a product concern, not a data-architecture one.
- Tiered hot/warm/cold storage (moving aging data to cheaper storage before its TTL expires) — a valid optimization industry practice supports, but an implementation decision downstream of the retention *periods* this document sets, not a change to them.

---

Sources: [What is Log Retention? — LogicMonitor](https://www.logicmonitor.com/blog/what-is-log-retention), [Log Retention: Policies, Best Practices & Tools — Last9](https://last9.io/blog/log-retention/), [Log Retention Policies Explained — Groundcover](https://www.groundcover.com/learn/logging/log-retention-policies).

### What This Feeds Next

`docs/05-data/README.md#storage` is the last file in `05-data/` — it should specify the object-storage layout for EvidenceExport blobs, and note how tiered storage (mentioned above but not specified) would interact with the retention periods this document sets.

