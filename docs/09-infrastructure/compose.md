# Docker Compose

> Status: Research-based, current as of July 2026. The single-machine deployment profile from [deployment-model.md](../04-architecture/deployment-model.md), made concrete — the eight services from [container-diagrams.md](../04-architecture/container-diagrams.md) plus ClickHouse, PostgreSQL, Redis, and MinIO on one host, no message queue, per that document's single-machine design.

---

## The Finding: MinIO's Object Lock Setup Order Is Not Fixable Later

[storage.md](../06-data/storage.md) specified Compliance-mode Object Lock for EvidenceExport blobs — but a default lock configuration **never applies retroactively**, only to objects created after it's set, even on current MinIO releases that otherwise relaxed the old "must enable at bucket creation" requirement. **The bucket must be created with versioning enabled and the default Compliance-mode retention configured before `workers` generates its first EvidenceExport — not as a follow-up hardening step.** An export generated before that configuration lands is silently unprotected by the WORM guarantee for its entire lifetime, with no way to retroactively apply it. This makes bucket initialization a strict ordering dependency in the Compose startup sequence, not an independent service:

```yaml
services:
  minio-init:
    image: minio/mc:latest
    depends_on:
      minio:
        condition: service_healthy
    entrypoint: >
      /bin/sh -c "
      mc alias set local http://minio:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD &&
      mc mb --with-versioning local/imora-evidence-exports &&
      mc retention set --default COMPLIANCE 6y local/imora-evidence-exports
      "
    restart: "no"

  workers:
    depends_on:
      minio-init:
        condition: service_completed_successfully
      # ...
```

`workers` (the only service that generates EvidenceExport blobs, per [container-diagrams.md](../04-architecture/container-diagrams.md)) explicitly depends on `minio-init` completing successfully — not just on MinIO being reachable — so there is no window where `workers` could run against an unversioned, unlocked bucket.

---

## Service Composition

Matches [container-diagrams.md](../04-architecture/container-diagrams.md)'s single-machine profile and [deployment-model.md](../04-architecture/deployment-model.md)'s sizing exactly:

| Service | Resource allocation | Notes |
|---|---|---|
| clickhouse | 4-core CPU, 16GB RAM, SSD volume | The dominant resource consumer, per [deployment-model.md](../04-architecture/deployment-model.md). |
| postgres | Modest (default limits) | Small-cardinality relational data only. |
| redis | 256MB | Cache only — no persistent volume needed; rebuildable from Postgres per [container-diagrams.md](../04-architecture/container-diagrams.md). |
| minio | Sized to EvidenceExport volume, not baseline | Grows with usage, not deployment size. |
| gateway, ingestion, query-api, alert-engine, workers, notification-service | Default limits, tunable | Stateless application services. |
| dashboard | Static asset serving | No state. |

**No message queue** in this profile — `ingestion` writes directly to ClickHouse, and `alert-engine`/`workers` poll or subscribe directly, per [container-diagrams.md](../04-architecture/container-diagrams.md)'s explicit single-machine simplification.

---

## Startup Ordering

1. `postgres`, `clickhouse`, `redis`, `minio` — the data layer, health-checked before anything else starts.
2. `minio-init` — the bucket-versioning-and-lock step above, gated on MinIO being healthy.
3. Schema migrations (Postgres tables from [postgres-schema.md](../06-data/postgres-schema.md), ClickHouse tables from [clickhouse-schema.md](../06-data/clickhouse-schema.md)) — a one-shot init job, not a long-running service.
4. `gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`, `dashboard` — gated on both the data layer and migrations completing.

This ordering is what makes [prd.md](../01-product/prd.md)'s "under 1 hour, unassisted" time-to-first-value target achievable: a correct dependency graph means `docker compose up` produces a working instance without manual intervention between steps, rather than a runbook of "wait for X, then manually run Y."

---

## Secrets and Local-Profile Key Management

Per [pii-redaction.md](../08-security/pii-redaction.md)'s single-machine KEK design: the root key is generated on first startup and mounted as a file outside the application's data volumes, not baked into the Compose file or checked into version control. API tokens and database credentials follow the same pattern — generated on first run, persisted to a secrets file the Compose file references but never contains inline.

---

## What's Deliberately Not Modeled Here

- The complete Compose file for all eight application services — this document shows the pattern (the MinIO ordering dependency, the data-layer-first sequencing); the full file is an implementation artifact, not a design decision.
- Exact health-check command/interval tuning — an operational detail downstream of this design.
- Local development conveniences (hot-reload, debug ports) — those are `10-engineering/` concerns, not part of the production single-machine profile this document specifies.

---

Sources: [Object Locking and Immutability — MinIO AIStor Documentation](https://docs.min.io/aistor/administration/object-locking-and-immutability/), [Objects and Versioning — MinIO AIStor Documentation](https://docs.min.io/aistor/administration/objects-and-versioning/), [MinIO discussion #19338 on object locking and versioning behavior](https://github.com/minio/minio/discussions/19338).

## What This Feeds Next

`docs/09-infrastructure/docker.md` should specify the per-service Dockerfile conventions (base images, non-root users) this Compose file assumes. `docs/09-infrastructure/kubernetes.md` should carry the same MinIO-initialization-ordering requirement into the cluster profile as an init container or Job, not just a Compose-specific pattern.
