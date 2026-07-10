# Postgres Schema

> Status: Table definitions for the relational side of the dual-store split from [container-diagrams.md](../04-architecture/container-diagrams.md) — Session summaries, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata, and user/RBAC config. Small-cardinality, transactionally-updated data, as opposed to [clickhouse-schema.md](clickhouse-schema.md)'s high-volume append-only side.

---

## Closing the Loop with clickhouse-schema.md: LegalHold Scope Must Be a Re-Evaluated Predicate, Not a Snapshot

[domain-model.md](../02-domain/domain-model.md) specified `LegalHold.scope` as a query, not a fixed record list, deliberately — a hold on "all sessions tied to incident X" should cover sessions created *after* the hold was applied, if the same compromised account or ongoing incident keeps generating them. A join table (`legal_hold_id, target_record_id`) populated once at hold-creation time would silently miss those later records — a real compliance gap, not a theoretical one.

The resolution: `scope` is stored as a **structured JSONB predicate**, not a fixed list and not a raw SQL string (which would be both a SQL-injection surface and something [clickhouse-schema.md](clickhouse-schema.md)'s dictionary refresh job has no safe way to evaluate). The predicate is **re-evaluated against current data on every dictionary refresh cycle**, producing the current resolved set of held `target_record_id`s at refresh time — which is what makes new matching records get picked up automatically, and is the concrete mechanism behind [clickhouse-schema.md](clickhouse-schema.md)'s `active_legal_holds` dictionary.

Supported predicate types, kept deliberately small rather than an open-ended query language:

```json
{"type": "session_ids", "values": ["uuid1", "uuid2"]}
{"type": "data_subject", "value": "user-identifier"}
{"type": "date_range", "start": "2026-01-01", "end": "2026-06-30"}
{"type": "incident_reference", "value": "INC-1234"}
```

A structured, closed set of predicate types — rather than arbitrary SQL — keeps the refresh job's evaluation logic bounded and auditable, which matters for a mechanism gating deletion of compliance-critical data.

---

## Common Design Decisions

- **Referential integrity is enforced within Postgres**, standard foreign keys between the tables below.
- **Referential integrity is *not* enforced across the store boundary.** `error_events.error_group_id` (ClickHouse) references `error_groups.id` (Postgres), and `session_events.session_id` (ClickHouse) references `sessions.id` (Postgres) — neither is a database-level constraint, since ClickHouse doesn't support foreign keys to another engine. This is a deliberate consistency-model tradeoff of the dual-store architecture from [container-diagrams.md](../04-architecture/container-diagrams.md), not an oversight: `alert-engine` and `ingestion` are responsible for creating the Postgres row before or atomically with the corresponding ClickHouse write, per the Shared Kernel relationship in [bounded-contexts.md](../02-domain/bounded-contexts.md).

---

## Table Definitions

### sessions
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | Matches `session_events.session_id` in ClickHouse. |
| `user_or_anonymous_id` | String | |
| `started_at` / `ended_at` | Timestamp | `ended_at` nullable until SessionEnded fires. |
| `release_id` | FK → releases.id | |
| `device_metadata` | JSONB | |

### releases
| Column | Type | Notes |
|---|---|---|
| `id` | String (PK) | |
| `deployed_at` | Timestamp | |
| `prior_release_id` | FK → releases.id, nullable | Regression baseline comparison, story C1. |

### error_groups
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `fingerprint` | String, unique | The write-time grouping key `alert-engine` matches against. |
| `first_seen_at` / `last_seen_at` | Timestamp | |
| `occurrence_count` | Integer | Denormalized counter, updated by `alert-engine` on each `ErrorGrouped` — avoids a ClickHouse aggregate query on every dashboard load for story C2's grouping display. |

### retention_policies
| Column | Type | Notes |
|---|---|---|
| `data_category` | String (PK) | SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, AccessAuditEvent. |
| `retention_period` | Interval | The `INTERVAL` value [clickhouse-schema.md](clickhouse-schema.md)'s TTL clauses read. |
| `regulatory_basis` | String | Which clock governs — PCI-DSS/HIPAA/GDPR/SOX, per BR-1. |

### legal_holds
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `applied_by` | FK → users.id | |
| `scope` | JSONB | The re-evaluated predicate, above. |
| `reason` | String | |
| `applied_at` / `lifted_at` | Timestamp | `lifted_at` nullable while active — **this column, not a separate boolean, is what the dictionary refresh job filters on** (`WHERE lifted_at IS NULL`), so there's exactly one place "is this hold active" is determined. |

### evidence_exports
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `requested_by` | FK → users.id | |
| `incident_reference` | String | |
| `content_hash` | String (SHA-256) | Per BR-4 — verifies the object-storage blob hasn't been altered. |
| `object_storage_path` | String | Pointer to the frozen blob in MinIO, per [container-diagrams.md](../04-architecture/container-diagrams.md) — the export's actual content lives in object storage, not Postgres. |
| `generated_at` | Timestamp | |
| `included_record_types` | JSONB array | |

### users
| Column | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `role` | Enum: engineer, compliance_officer, platform_operator, admin | Maps to the four actor types in [system-context.md](../04-architecture/system-context.md). |
| `sso_subject_id` | String, nullable | Populated only for Enterprise-tier SSO logins, per [pricing.md](../01-product/pricing.md) — null for Community/Team tier local accounts. |

Fine-grained field-level permissions (as opposed to the coarse role above) are a `08-security/authorization.md` concern, not this table's — `role` here is enough to route a request, not enough to decide whether a specific field should be masked.

---

## What's Deliberately Not Modeled Here

- RBAC permission matrices and field-level access rules — `08-security/authorization.md`.
- Actual migration tooling or ORM choice — an implementation decision.
- Index tuning beyond the primary/foreign keys implied above — performance work downstream of this schema.

---

## What This Feeds Next

`docs/06-data/retention.md` should formalize the `retention_period`/`regulatory_basis` values in `retention_policies` against BR-1's actual regulatory clocks. `docs/06-data/storage.md` should specify the object-storage (MinIO) layout that `evidence_exports.object_storage_path` points into.
