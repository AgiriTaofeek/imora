# ClickHouse Schema

> Status: Research-based, current as of July 2026. Table definitions for the five ClickHouse-resident entities from [container-diagrams.md](../04-architecture/container-diagrams.md): SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, and AccessAuditEvent — all high-volume, append-only, per [event-schema.md](event-schema.md)'s payload shapes.

---

## The Tension This Document Has to Resolve First

ClickHouse is an OLAP columnar store optimized for inserts and aggregate reads, not row-level deletes — the efficient retention pattern is TTL-based expiry aligned to the partition key, which drops entire partitions as a cheap metadata operation (`ttl_only_drop_parts`). But [business-rules.md](../02-domain/business-rules.md) BR-2 requires checking each *individual* record against active legal holds immediately before it's deleted — a per-row concern, not a per-partition one. A naive implementation would force a choice between "efficient retention" and "correct legal-hold enforcement." It doesn't have to.

**Resolution:** ClickHouse's TTL clause natively supports conditional `WHERE` expressions — `TTL occurredAt + INTERVAL 6 YEAR DELETE WHERE <condition>` is documented, standard syntax. Rather than a mutable per-row flag column (which would require expensive row-level UPDATEs — precisely the operation this whole problem is trying to avoid), the hold check is expressed as a **dictionary lookup**: a ClickHouse Dictionary of currently-active hold scopes, refreshed periodically from Postgres's LegalHold table (the source of truth per [container-diagrams.md](../04-architecture/container-diagrams.md)), queried inline in the TTL expression via `dictGetOrDefault('active_legal_holds', 'is_held', sessionId, 0) = 0`. This is the standard ClickHouse mechanism for exactly this shape of problem — checking every row against a small, frequently-changing reference set without paying for a join or an update on every row.

**One operational consequence worth stating plainly:** a single held row inside a partition prevents `ttl_only_drop_parts` from cheaply dropping that whole partition — the partition falls back to slower row-level TTL evaluation until the hold lifts. This is exactly why the partitioning scheme below uses **monthly**, not yearly, partitions: it limits the "blast radius" of what one legal hold can pin from cheap expiry to a month's worth of data, not a year's.

---

## Common Design Decisions

- **Engine family:** MergeTree (specifically plain `MergeTree`, not `ReplacingMergeTree` — these tables are genuinely append-only per [domain-model.md](../02-domain/domain-model.md)'s event-sourcing pattern; there is no "latest version wins" semantic to support).
- **Partitioning:** monthly, by `occurredAt`, on every table below — balances the cheap-partition-drop benefit against the legal-hold blast-radius concern above.
- **Deletion mechanism by scenario:** scheduled RetentionPolicy sweeps (BR-1) use TTL-with-dictionary-lookup, above. A prompt, individual GDPR erasure request (BR-3, story A2) that can't wait for the next background merge uses a targeted **lightweight DELETE** instead — 10–100x faster than a full mutation for a single-record removal, at the cost of requiring an eventual background merge to reclaim the space, which is an acceptable tradeoff for a rare, individual action rather than the routine sweep path.

---

## Table Definitions

### session_events
| Column | Type | Notes |
|---|---|---|
| `session_id` | UUID | |
| `subtype` | Enum8 | FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange, per [event-schema.md](event-schema.md). |
| `occurred_at` | DateTime64 | Partition and TTL basis. |
| `payload` | String (JSON) | Already masked per BR-7 at capture time. |
| `release_id` | String | |
| **ORDER BY** | `(session_id, occurred_at)` | Replay reconstruction is always "give me this session's events in order" — the dominant query shape. |
| **TTL** | `occurred_at + INTERVAL <category clock> DELETE WHERE dictGetOrDefault('active_legal_holds', 'is_held', session_id, 0) = 0` | Category clock per BR-1's RetentionPolicy for this data category. |

### error_events
| Column | Type | Notes |
|---|---|---|
| `error_event_id` | UUID | |
| `session_id` | Nullable(UUID) | Per [event-schema.md](event-schema.md). |
| `stack_trace_fingerprint` | String | Write-time grouping key. |
| `release_id` | String | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(stack_trace_fingerprint, occurred_at)` | ErrorGroup lookups are the dominant read pattern. |
| **TTL** | Same pattern as session_events, keyed to error-data's RetentionPolicy category. |

### performance_metrics
| Column | Type | Notes |
|---|---|---|
| `session_id` | UUID | |
| `metric_type` | Enum8 | LCP, INP, CLS |
| `value` | Float64 | |
| `release_id` | String | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(release_id, metric_type, occurred_at)` | Regression detection (story C1) queries by release and metric first. |
| **TTL** | Same pattern, own category clock — performance data typically has a shorter regulatory floor than session replay, per BR-1. |

### security_events
| Column | Type | Notes |
|---|---|---|
| `security_event_id` | UUID | |
| `session_id` | Nullable(UUID) | |
| `signal_type` | String | |
| `severity` | Enum8 | |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(session_id, occurred_at)` | Incident-timeline correlation (story D2) is the dominant read shape. |
| **TTL** | Same conditional pattern. |

### access_audit_events
| Column | Type | Notes |
|---|---|---|
| `event_id` | UUID | |
| `sequence_number` | UInt64 | Monotonic, per [domain-model.md](../02-domain/domain-model.md)'s gap-detection requirement. |
| `actor_user_id` | UUID | |
| `action` | Enum8 | VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED |
| `target_record_type` | String | |
| `target_record_id` | UUID | |
| `source_ip` | String | |
| `reason` | Nullable(String) | Required non-null when `action = UNMASK`, enforced at the `query-api`/`workers` write path (BR-6) — ClickHouse itself doesn't enforce conditional nullability, the writing service does. |
| `occurred_at` | DateTime64 | |
| **ORDER BY** | `(target_record_id, occurred_at)` | "Who viewed this specific record" (stories M1, A1) is the dominant read shape — this ordering is what makes Adaeze's DSAR query and Marcus's risk-assessment report both fast. |
| **TTL** | **Longest applicable regulatory clock across all categories the deployment operates under** (per BR-1's longest-wins rule) — this table's retention floor is never shorter than any other table's, since an audit record about a since-deleted SessionEvent may still need to prove the deletion happened correctly. |

**A note on `access_audit_events` never being younger than what it audits:** if SessionEvent is deleted under a 12-month PCI-DSS clock but its corresponding `RecordDeleted` AccessAuditEvent were also deleted on that same clock, the system would lose the ability to prove the deletion was lawful — exactly the audit-control failure [business-rules.md](../02-domain/business-rules.md) exists to prevent. `access_audit_events`' own TTL is therefore set to the *longest* clock in use, not the clock of whatever it's auditing.

---

## What's Deliberately Not Modeled Here

- Exact `CREATE DICTIONARY` syntax for `active_legal_holds` and its refresh interval — an implementation detail once this design is accepted, not an architecture decision.
- Compression codecs, index granularity tuning, or materialized views for common aggregations — performance tuning downstream of this schema, not part of it.
- The Postgres-side LegalHold, RetentionPolicy, Session, Release, ErrorGroup, and EvidenceExport tables — `postgres-schema.md`, next.

---

Sources: [Manage data with TTL — ClickHouse Docs](https://clickhouse.com/docs/guides/developer/ttl), [Support of WHERE and GROUP BY in TTL expressions — ClickHouse GitHub PR #10537](https://github.com/ClickHouse/ClickHouse/pull/10537), [Lightweight delete — ClickHouse Docs](https://clickhouse.com/docs/guides/developer/lightweight-delete), [What Is the Difference Between Mutations and Lightweight Deletes in ClickHouse](https://oneuptime.com/blog/post/2026-03-31-clickhouse-mutations-vs-lightweight-deletes/view).

## What This Feeds Next

`docs/06-data/postgres-schema.md` covers the relational side. `docs/06-data/retention.md` should turn BR-1's per-category regulatory clocks into the actual `INTERVAL` values used in the TTL clauses above.
