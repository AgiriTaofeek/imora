# Kubernetes

> Status: Research-based, current as of July 2026. The cluster deployment profile from [deployment-model.md](../04-architecture/deployment-model.md) and [scaling.md](../04-architecture/scaling.md), made concrete.

---

## The Finding: `workers` Cannot Be a Single Horizontally-Scaled Deployment

[business-rules.md](../02-domain/business-rules.md)'s Conflict Precedence Summary and [component-diagrams.md](../04-architecture/component-diagrams.md)'s `LegalHoldChecker` → `DeletionExecutor` ordering both depend on the retention sweep touching shared, mutable state (which records are past their retention clock, which are held) in a coordinated way. Naively scaling `workers` to multiple replicas, all running the same `RetentionSweepScheduler` logic independently, risks exactly the race condition [business-rules.md](../02-domain/business-rules.md) BR-2 was designed to prevent — two replicas evaluating overlapping candidate sets at slightly different times against a hold cache that's changing underneath them.

**Resolution: `workers` splits into two distinct Kubernetes workload types, not one Deployment:**

- **`RetentionSweepScheduler` runs as a `CronJob` with `concurrencyPolicy: Forbid`** — Kubernetes' own documented recommendation for exactly this situation: a scheduled task that modifies shared state should never run overlapping instances. If a sweep is still running when the next scheduled time arrives, the new run is skipped rather than racing the old one.
- **`EvidenceExportGenerator` and per-request UNMASK-adjacent processing run as an ordinary scalable `Deployment`** — each request is independently partitioned by its own export/unmask ID, so horizontal scaling here doesn't touch the shared-state problem above at all. This is genuinely a different workload shape hiding inside one bounded context, and the Kubernetes manifests should reflect that split rather than forcing both into one Deployment for organizational convenience.

---

## Cluster Topology

| Component | Kubernetes primitive | Notes |
|---|---|---|
| gateway, ingestion, query-api, alert-engine, notification-service | `Deployment` + `HorizontalPodAutoscaler` | Stateless, scale independently per their distinct load shapes (write-heavy `ingestion` vs. read-latency-sensitive `query-api`), per [bounded-contexts.md](../02-domain/bounded-contexts.md)'s write/read separation. |
| `workers` (scheduled portion) | `CronJob`, `concurrencyPolicy: Forbid` | Per the finding above. |
| `workers` (on-demand portion) | `Deployment` + `HorizontalPodAutoscaler` | Per the finding above. |
| `dashboard` | `Deployment` (static assets) | No state. |
| ClickHouse, PostgreSQL | `StatefulSet`, multi-node | Per [deployment-model.md](../04-architecture/deployment-model.md)'s cluster profile — this is also where the message queue between `ingestion` and its consumers gets introduced, absent in the single-machine profile. |
| MinIO bucket initialization | `Job` (not a Compose-style `depends_on`) | Carries [compose.md](compose.md)'s versioning-and-lock ordering requirement into the cluster profile — the Job must complete successfully before the `workers` Deployment's pods are allowed to become ready, enforced via an init container or a readiness gate, not just documentation. |

---

## Network Policy as Defense in Depth

A `NetworkPolicy` restricting `dashboard`'s egress to `query-api` only — never directly to ClickHouse, PostgreSQL, or MinIO — enforces [bounded-contexts.md](../02-domain/bounded-contexts.md)'s Conformist relationship at the network layer, not just in application code. This is the same "enforce it structurally, not procedurally" principle [component-diagrams.md](../04-architecture/component-diagrams.md) applied to the audit-log wrapper and [docker.md](docker.md) applied to read-only filesystems, applied here at the network layer: even a compromised or misconfigured `dashboard` pod cannot reach the data stores directly, regardless of what its application code does or doesn't enforce.

---

## What's Deliberately Not Modeled Here

- Exact HPA scaling thresholds (CPU/memory targets, custom metrics) — operational tuning, not architecture.
- The message queue's specific configuration (partitioning, consumer group setup) — an implementation detail of the cluster-profile ingestion path.
- StatefulSet replica counts and ClickHouse sharding key selection — deployment-specific, sized to the customer's actual traffic per [scaling.md](../04-architecture/scaling.md).

---

Sources: [CronJob — Kubernetes Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/), [How to Configure CronJob concurrencyPolicy for Allow, Forbid, and Replace](https://oneuptime.com/blog/post/2026-02-09-cronjob-concurrency-policy-allow-forbid/view).

## What This Feeds Next

`docs/09-infrastructure/ci-cd.md` should specify how images get built and signed for both the Compose and Kubernetes profiles. `docs/09-infrastructure/observability.md` should specify monitoring for Imora's own infrastructure — distinct from the product's own observability features, the same scope distinction [audit-logging.md](../08-security/audit-logging.md) drew for logging.
