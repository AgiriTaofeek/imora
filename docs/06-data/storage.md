# Storage

> Status: Research-based, current as of July 2026. The last file in `06-data/` — specifies the object-storage layer for EvidenceExport blobs, referenced by `evidence_exports.object_storage_path` in [postgres-schema.md](postgres-schema.md), and ties together the three-store picture (ClickHouse, PostgreSQL, object storage) from [container-diagrams.md](../04-architecture/container-diagrams.md).

---

## The Finding: Object-Lock Compliance Mode Makes BR-4 a Storage-Layer Guarantee, Not Just an Application Promise

[business-rules.md](../02-domain/business-rules.md) BR-4 requires that an EvidenceExport be immutable once generated — no later retention purge or erasure action may alter it. Everything specified so far ([domain-model.md](../02-domain/domain-model.md), [event-catalog.md](../02-domain/event-catalog.md), `contentHash` in [postgres-schema.md](postgres-schema.md)) enforces this at the application layer: services simply don't write code paths that modify a generated export. That's real, but it's a weaker guarantee than it sounds — it depends on every future engineer never writing that code path.

MinIO (the self-hosted S3-compatible object store from [container-diagrams.md](../04-architecture/container-diagrams.md)) supports **Object Lock in Compliance mode**: true WORM (write-once-read-many) immutability where a locked object **cannot be deleted or modified by anyone, including the root user, for the duration of the lock** — a materially stronger guarantee than Governance mode, which privileged users can still override. **Every EvidenceExport blob is written with Object Lock in Compliance mode**, set for the same duration as its retention period (per [retention.md](retention.md)'s rule that AccessAuditEvent-adjacent data never expires before the longest applicable clock). This turns BR-4 from a coding convention into a property the storage system itself enforces.

**One useful, narrower side-effect worth noting separately:** MinIO's Object Lock feature also supports an indefinite legal-hold sub-mode, distinct from duration-based retention. For an EvidenceExport blob specifically, applying a [domain-model.md](../02-domain/domain-model.md) LegalHold could set this native flag directly, rather than relying purely on the application-level check. **This does not replace the dictionary-based mechanism in [clickhouse-schema.md](clickhouse-schema.md)** — MinIO's object-lock model operates on whole objects in object storage, not rows inside a ClickHouse table, so the two mechanisms cover different parts of the system: object-lock for the comparatively small number of frozen EvidenceExport blobs, the dictionary lookup for the high-volume SessionEvent/ErrorEvent/AccessAuditEvent rows.

---

## Bucket and Key Layout

```
imora-evidence-exports/
  {yyyy}/{mm}/{export_id}.tar.gz
```

Partitioned by year/month of `generated_at`, even though exports are typically held under Object Lock for years — the date partitioning is for eventual lifecycle/storage-tier management (below), not for any deletion logic, since deletion of a locked object is refused by the storage layer regardless of its path.

- **Object content:** a single compressed archive containing the frozen copy of every referenced Session, ErrorEvent, SecurityEvent, and AccessAuditEvent, per [domain-model.md](../02-domain/domain-model.md)'s EvidenceExport definition.
- **`contentHash`** (stored in `evidence_exports.content_hash`, per [postgres-schema.md](postgres-schema.md)) is computed over this exact archive — verifying it means re-hashing the object and comparing, independent of Object Lock, so integrity is checkable even by a party who doesn't trust MinIO's lock enforcement alone.

---

## Tiered Storage — Deferred, Not Designed Away

[retention.md](retention.md) flagged hot/warm/cold tiering as a valid future optimization without specifying it. For object storage specifically: MinIO supports lifecycle transition rules to move objects to cheaper storage classes after an age threshold, which is compatible with Object Lock (the lock governs deletion, not storage class). This is a legitimate later optimization once real export volume exists to justify it — not a day-one requirement, since [scaling.md](../04-architecture/scaling.md) already identified accumulated *session-replay* storage in ClickHouse as the actual scaling trigger, not evidence-export blob volume, which is comparatively rare (generated per-incident, not per-session).

---

## Air-Gapped Consistency

MinIO runs self-hosted within the deployment boundary in both topology profiles from [deployment-model.md](../04-architecture/deployment-model.md) — no cloud object-storage dependency, consistent with [system-context.md](../04-architecture/system-context.md)'s requirement that every Parity and Wedge capability, including evidence export, works with zero external systems present.

---

## What's Deliberately Not Modeled Here

- IAM/bucket policy specifics for who can initiate an export or read a generated one — `08-security/authorization.md`.
- Exact archive format (tar.gz vs. a structured container) — an implementation detail once this design is accepted.
- Lifecycle transition rule configuration — deferred per the Tiered Storage section above, pending real usage data.

---

Sources: [Object Locking and Immutability — MinIO Documentation](https://docs.min.io/aistor/administration/object-locking-and-immutability/), [Object Locking, Versioning, Legal Holds and Modes in MinIO](https://blog.min.io/object-locking-versioning-and-holds-in-minio/), [MinIO Object Locking — MinIO Object Storage (AGPLv3)](https://docs.min.io/community/minio-object-store/administration/object-management/object-retention.html).

## What This Closes Out

This is the last file in `docs/06-data/`. All five files — [event-schema.md](event-schema.md), [clickhouse-schema.md](clickhouse-schema.md), [postgres-schema.md](postgres-schema.md), [retention.md](retention.md), and this one — are now internally consistent, and the object-storage layer's Compliance-mode lock is a concrete answer to the immutability question `deployment-model.md`'s backup section raised without fully resolving. `docs/08-security/` (authorization, encryption, audit-logging) and `docs/07-api/` are the natural next tiers — the former specifies who can trigger the actions this folder's schemas record.
