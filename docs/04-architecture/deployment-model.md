# Deployment Model

> Status: Research-based, current as of July 2026. Makes the two topology profiles from [container-diagrams.md](container-diagrams.md) concrete — actual hardware numbers, actual orchestration tooling, and the two operational questions those profiles didn't yet answer: how does an air-gapped deployment update itself, and what does backup/restore have to guarantee for compliance data specifically. `09-infrastructure/compose.md` and `kubernetes.md` (already scaffolded under those exact names) turn this into actual manifests.

---

## Two Profiles, Concrete Now

### Single-Machine Profile — Docker Compose

Sizing follows the closest architectural comparator (Uptrace, the same ClickHouse+Postgres+Redis stack) rather than a guess:

| Component | Minimum | Why |
|---|---|---|
| ClickHouse | 4-core CPU, 16GB RAM, SSD storage | The dominant resource consumer — it holds SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, and AccessAuditEvent, all high-volume. 8GB is an absolute floor for basic workloads; 16GB is where a real deployment should start. |
| PostgreSQL | A few hundred MB to low GB | Small-cardinality relational data (Session summaries, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata) — minimal by comparison. |
| Redis | 256MB | Cache only (active-hold lookups, rate limits) — rebuildable from Postgres at any time, never a source of truth. |
| Object storage (MinIO) | Sized to EvidenceExport volume, not baseline | Grows with usage, not with deployment size — a low-traffic deployment can start small. |

This gives Priya's story P1 an actual number instead of a promise: a **single 4-core/16GB-RAM host with SSD storage** is the concrete claim behind "a 2–3 person team can deploy a working instance," per [prd.md](../01-product/prd.md)'s Success Metrics. All eight containers from [container-diagrams.md](container-diagrams.md) run as Docker Compose services on this one host, no Kafka, matching the single-machine profile already established there.

### Cluster Profile — Kubernetes

Containers scale independently per their load shape (write-heavy `ingestion` vs. read-latency-sensitive `query-api`, per the write/read separation rationale in [bounded-contexts.md](../02-domain/bounded-contexts.md)); ClickHouse and Postgres move to multi-node; a message queue is introduced between `ingestion` and its consumers. Manifest-level detail belongs in `09-infrastructure/kubernetes.md`, not here.

---

## Air-Gapped Is an Orthogonal Axis, Not a Third Profile

Per [system-context.md](system-context.md), air-gapped applies to *either* profile — a single-machine air-gapped deployment and a cluster air-gapped deployment are both real configurations, distinguished only by whether the four optional external systems (SSO, notification channels, backend TraceLink correlation, and now: update delivery, below) are reachable at all.

---

## Updates in an Air-Gapped Deployment — the Question Neither Prior Document Answered

[system-context.md](system-context.md) established that air-gapped deployments have no outbound dependency for core function, and [pricing.md](../01-product/pricing.md) already solved an adjacent problem (Enterprise license activation) with signed offline files transferred by hand. Software updates need the same answer, and the standard pattern for exactly this problem — used across government, defense, and healthcare air-gapped patch management — is:

1. **Stage outside the air gap.** Update bundles are prepared, signed (SHA-256 or digital signature, the same cryptographic pattern [licensing.md](../01-product/licensing.md) already specified for license files), and validated in a connected staging environment.
2. **Transfer via approved removable media**, not a network link — organization-managed media, virus-scanned on both sides, with every transfer documented (contents, date, approver, handler). Write-once media is preferred specifically because it leaves an irreversible audit trail of what crossed the boundary and when — a compliance-relevant property, not just a security one.
3. **Apply and verify signature locally**, with no network call back to Imora's own infrastructure required at any step.

**This reuses the exact operational muscle Enterprise customers already have from license activation** — the same signed-bundle-via-removable-media process, not a second unrelated procedure Platform Operators have to learn. That consistency wasn't planned when [pricing.md](../01-product/pricing.md) specified offline license files; it falls out of both problems having the same shape.

---

## Backup and Restore — a Compliance Requirement, Not Just an Operational Nicety

This is worth stating as a hard requirement rather than deferring to general ops best practice: per [business-rules.md](../02-domain/business-rules.md) BR-1, AccessAuditEvent data must survive for up to 7 years (SOX) or 6 years (HIPAA) depending on category. **A lost AccessAuditEvent isn't a data-loss incident — it's an unanswerable audit-control gap under HIPAA §164.312(b)**, the exact requirement Marcus's persona depends on. Backup scope, concretely:

- **ClickHouse and PostgreSQL require backup with an RPO tight enough that no AccessAuditEvent is ever unrecoverable** — a gap in the audit trail is indistinguishable from a compliance failure to an assessor, regardless of the actual cause.
- **Object storage (EvidenceExport blobs) requires backup** for the same reason BR-4 requires immutability: an export that can't survive a disk failure isn't actually the defensible artifact story J2 promises.
- **Redis requires no backup.** It's a cache, rebuildable from Postgres and the active LegalHold set — treating it as durable state would be a design error, not a safety margin.

---

## What's Deliberately Not Modeled Here

- Actual Docker Compose or Kubernetes manifests — `09-infrastructure/compose.md` and `kubernetes.md`.
- CI/CD pipeline and release process for shipping updates in the first place (as opposed to applying them air-gapped) — `09-infrastructure/ci-cd.md` and `10-engineering/release-process.md`.
- Specific backup tooling or schedule — an implementation decision downstream of the RPO requirement stated above.

---

Sources: [Sizing and hardware recommendations — ClickHouse Docs](https://clickhouse.com/docs/guides/sizing-and-hardware-recommendations), [Hardware Requirements — Altinity Knowledge Base](https://kb.altinity.com/altinity-kb-setup-and-maintenance/cluster-production-configuration-guide/hardware-requirements/), [Patch Management in Isolated Networks — SecOps Solution](https://www.secopsolution.com/blog/patch-management-in-isolated-networks-best-practices-for-air-gapped-environments), [Air-gapped deployments for defense software](https://corvusintell.com/blog/secure-cloud/air-gapped-deployment-defense/).

## What This Feeds Next

`docs/04-architecture/scaling.md` should define the concrete threshold at which the single-machine profile stops being viable and a cluster migration is warranted. `docs/09-infrastructure/compose.md` and `kubernetes.md` can now be written directly against the two profiles here.
