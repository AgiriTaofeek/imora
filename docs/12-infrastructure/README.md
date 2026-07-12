# Infrastructure

## Docker

> Status: Image and container conventions for every service in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams), underpinning both the [Docker Compose](README.md#docker-compose) and [Kubernetes](README.md#kubernetes) profiles.

---

### Image Distribution Must Follow the Air-Gapped Update Pattern, Not `docker pull`

A Dockerfile that pulls its base image from a public registry at build time, or a Compose/Kubernetes setup that expects `docker pull` to reach a registry at deploy time, silently breaks for the air-gapped deployments [System Context](../03-architecture/README.md#system-context) requires Imora to support. Container images are exactly the kind of artifact [Deployment Model](../03-architecture/README.md#deployment-model)'s signed-bundle update mechanism already covers — staged and signed outside the air gap, transferred via approved removable media, verified locally. **Imora ships as a set of pre-built, signed images distributed through that same mechanism**, not as Dockerfiles requiring a live registry at deploy time. Connected deployments may still pull from a registry for convenience; air-gapped ones use the bundle path — the images themselves are identical either way, only the distribution channel differs.

---

### Build and Runtime Conventions

- **Multi-stage builds.** Build tooling, dependency caches, and source never ship in the final image — only compiled/bundled output and its runtime dependencies. This shrinks both image size and attack surface.
- **Minimal base images** (distroless or slim variants, not full OS images) for the same reason — fewer packages in the final image means fewer components to patch and fewer things an attacker who gains container access can use.
- **Non-root user by default.** Every service container runs as an unprivileged user; no service needs root to do its job, and running as root is the kind of default-permissive choice [Business Rules](../02-domain/README.md#business-rules) BR-7's "deny-by-default" philosophy argues against generally, applied here to infrastructure rather than data capture.
- **Read-only root filesystem where the service allows it** — a direct, concrete mitigation for the Tampering threats [Threat Model](../07-security/README.md#threat-model) already identified: a compromised `ingestion` or `query-api` container with a read-only filesystem can't persist a modified binary or write a backdoor to disk, even with an initial foothold. Services that need a writable scratch directory (temp files, caches) get an explicitly mounted, narrowly-scoped writable volume — not a writable root.

---

### What's Deliberately Not Modeled Here

- Specific base image tags/versions — a maintenance decision, updated over time, not a one-time architectural choice.
- Image vulnerability scanning as a pipeline gate — `12-infrastructure/README.md#cicd`.
- The signed-bundle packaging format itself — already specified at the update-mechanism level in [Deployment Model](../03-architecture/README.md#deployment-model); this document only establishes that container images travel through it.

---

### What This Feeds Next

`docs/12-infrastructure/README.md#kubernetes` should carry these same conventions into the cluster profile, plus the MinIO-initialization-ordering requirement from [Docker Compose](README.md#docker-compose) reimplemented as an init container or Job rather than a Compose dependency.

---

## Docker Compose

> Status: Research-based, current as of July 2026. The single-machine deployment profile from [Deployment Model](../03-architecture/README.md#deployment-model), made concrete — the eight services from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) plus ClickHouse, PostgreSQL, Redis, and MinIO on one host, no message queue, per that document's single-machine design.

---

### The Finding: MinIO's Object Lock Setup Order Is Not Fixable Later

[Storage](../05-data/README.md#storage) specified Compliance-mode Object Lock for EvidenceExport blobs — but a default lock configuration **never applies retroactively**, only to objects created after it's set, even on current MinIO releases that otherwise relaxed the old "must enable at bucket creation" requirement. **The bucket must be created with versioning enabled and the default Compliance-mode retention configured before `workers` generates its first EvidenceExport — not as a follow-up hardening step.** An export generated before that configuration lands is silently unprotected by the WORM guarantee for its entire lifetime, with no way to retroactively apply it. This makes bucket initialization a strict ordering dependency in the Compose startup sequence, not an independent service:

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

`workers` (the only service that generates EvidenceExport blobs, per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)) explicitly depends on `minio-init` completing successfully — not just on MinIO being reachable — so there is no window where `workers` could run against an unversioned, unlocked bucket.

---

### Service Composition

Matches [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)'s single-machine profile and [Deployment Model](../03-architecture/README.md#deployment-model)'s sizing exactly:

| Service | Resource allocation | Notes |
|---|---|---|
| clickhouse | 4-core CPU, 16GB RAM, SSD volume | The dominant resource consumer, per [Deployment Model](../03-architecture/README.md#deployment-model). |
| postgres | Modest (default limits) | Small-cardinality relational data only. |
| redis | 256MB | Cache only — no persistent volume needed; rebuildable from Postgres per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams). |
| minio | Sized to EvidenceExport volume, not baseline | Grows with usage, not deployment size. |
| gateway, ingestion, query-api, alert-engine, workers, notification-service | Default limits, tunable | Stateless application services. |
| dashboard | Static asset serving | No state. |

**No message queue** in this profile — `ingestion` writes directly to ClickHouse, and `alert-engine`/`workers` poll or subscribe directly, per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)'s explicit single-machine simplification.

---

### Startup Ordering

1. `postgres`, `clickhouse`, `redis`, `minio` — the data layer, health-checked before anything else starts.
2. `minio-init` — the bucket-versioning-and-lock step above, gated on MinIO being healthy.
3. Schema migrations (Postgres tables from [Postgres Schema](../05-data/README.md#postgres-schema), ClickHouse tables from [ClickHouse Schema](../05-data/README.md#clickhouse-schema)) — a one-shot init job, not a long-running service.
4. `gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`, `dashboard` — gated on both the data layer and migrations completing.

This ordering is what makes [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)'s "under 1 hour, unassisted" time-to-first-value target achievable: a correct dependency graph means `docker compose up` produces a working instance without manual intervention between steps, rather than a runbook of "wait for X, then manually run Y."

---

### Secrets and Local-Profile Key Management

Per [PII Redaction](../07-security/README.md#pii-redaction)'s single-machine KEK design: the root key is generated on first startup and mounted as a file outside the application's data volumes, not baked into the Compose file or checked into version control. API tokens and database credentials follow the same pattern — generated on first run, persisted to a secrets file the Compose file references but never contains inline.

---

### What's Deliberately Not Modeled Here

- The complete Compose file for all eight application services — this document shows the pattern (the MinIO ordering dependency, the data-layer-first sequencing); the full file is an implementation artifact, not a design decision.
- Exact health-check command/interval tuning — an operational detail downstream of this design.
- Local development conveniences (hot-reload, debug ports) — those are `11-engineering/` concerns, not part of the production single-machine profile this document specifies.

---

Sources: [Object Locking and Immutability — MinIO AIStor Documentation](https://docs.min.io/aistor/administration/object-locking-and-immutability/), [Objects and Versioning — MinIO AIStor Documentation](https://docs.min.io/aistor/administration/objects-and-versioning/), [MinIO discussion #19338 on object locking and versioning behavior](https://github.com/minio/minio/discussions/19338).

### What This Feeds Next

`docs/12-infrastructure/README.md#docker` should specify the per-service Dockerfile conventions (base images, non-root users) this Compose file assumes. `docs/12-infrastructure/README.md#kubernetes` should carry the same MinIO-initialization-ordering requirement into the cluster profile as an init container or Job, not just a Compose-specific pattern.

---

## Kubernetes

> Status: Research-based, current as of July 2026. The cluster deployment profile from [Deployment Model](../03-architecture/README.md#deployment-model) and [Scaling](../03-architecture/README.md#scaling), made concrete.

---

### The Finding: `workers` Cannot Be a Single Horizontally-Scaled Deployment

[Business Rules](../02-domain/README.md#business-rules)'s Conflict Precedence Summary and [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)'s `LegalHoldChecker` → `DeletionExecutor` ordering both depend on the retention sweep touching shared, mutable state (which records are past their retention clock, which are held) in a coordinated way. Naively scaling `workers` to multiple replicas, all running the same `RetentionSweepScheduler` logic independently, risks exactly the race condition [Business Rules](../02-domain/README.md#business-rules) BR-2 was designed to prevent — two replicas evaluating overlapping candidate sets at slightly different times against a hold cache that's changing underneath them.

**Resolution: `workers` splits into two distinct Kubernetes workload types, not one Deployment:**

- **`RetentionSweepScheduler` runs as a `CronJob` with `concurrencyPolicy: Forbid`** — Kubernetes' own documented recommendation for exactly this situation: a scheduled task that modifies shared state should never run overlapping instances. If a sweep is still running when the next scheduled time arrives, the new run is skipped rather than racing the old one.
- **`EvidenceExportGenerator` and per-request UNMASK-adjacent processing run as an ordinary scalable `Deployment`** — each request is independently partitioned by its own export/unmask ID, so horizontal scaling here doesn't touch the shared-state problem above at all. This is genuinely a different workload shape hiding inside one bounded context, and the Kubernetes manifests should reflect that split rather than forcing both into one Deployment for organizational convenience.

---

### Cluster Topology

| Component | Kubernetes primitive | Notes |
|---|---|---|
| gateway, ingestion, query-api, alert-engine, notification-service | `Deployment` + `HorizontalPodAutoscaler` | Stateless, scale independently per their distinct load shapes (write-heavy `ingestion` vs. read-latency-sensitive `query-api`), per [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s write/read separation. |
| `workers` (scheduled portion) | `CronJob`, `concurrencyPolicy: Forbid` | Per the finding above. |
| `workers` (on-demand portion) | `Deployment` + `HorizontalPodAutoscaler` | Per the finding above. |
| `dashboard` | `Deployment` (static assets) | No state. |
| ClickHouse, PostgreSQL | `StatefulSet`, multi-node | Per [Deployment Model](../03-architecture/README.md#deployment-model)'s cluster profile — this is also where the message queue between `ingestion` and its consumers gets introduced, absent in the single-machine profile. |
| MinIO bucket initialization | `Job` (not a Compose-style `depends_on`) | Carries [Docker Compose](README.md#docker-compose)'s versioning-and-lock ordering requirement into the cluster profile — the Job must complete successfully before the `workers` Deployment's pods are allowed to become ready, enforced via an init container or a readiness gate, not just documentation. |

---

### Network Policy as Defense in Depth

A `NetworkPolicy` restricting `dashboard`'s egress to `query-api` only — never directly to ClickHouse, PostgreSQL, or MinIO — enforces [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s Conformist relationship at the network layer, not just in application code. This is the same "enforce it structurally, not procedurally" principle [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) applied to the audit-log wrapper and [Docker](README.md#docker) applied to read-only filesystems, applied here at the network layer: even a compromised or misconfigured `dashboard` pod cannot reach the data stores directly, regardless of what its application code does or doesn't enforce.

---

### What's Deliberately Not Modeled Here

- Exact HPA scaling thresholds (CPU/memory targets, custom metrics) — operational tuning, not architecture.
- The message queue's specific configuration (partitioning, consumer group setup) — an implementation detail of the cluster-profile ingestion path.
- StatefulSet replica counts and ClickHouse sharding key selection — deployment-specific, sized to the customer's actual traffic per [Scaling](../03-architecture/README.md#scaling).

---

Sources: [CronJob — Kubernetes Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/), [How to Configure CronJob concurrencyPolicy for Allow, Forbid, and Replace](https://oneuptime.com/blog/post/2026-02-09-cronjob-concurrency-policy-allow-forbid/view).

### What This Feeds Next

`docs/12-infrastructure/README.md#cicd` should specify how images get built and signed for both the Compose and Kubernetes profiles. `docs/12-infrastructure/README.md#observability` should specify monitoring for Imora's own infrastructure — distinct from the product's own observability features, the same scope distinction [Audit Logging](../07-security/README.md#audit-logging) drew for logging.

---

## CI/CD

> Status: Research-based, current as of July 2026. Specifies how the images [Docker](README.md#docker) requires to be "pre-built and signed" actually get built and signed, and how the two distribution paths from [Deployment Model](../03-architecture/README.md#deployment-model) (registry pull for connected, signed bundle for air-gapped) come from the same pipeline run rather than drifting apart.

---

### The Finding: Keyless Signing Doesn't Work for Air-Gapped Verification

The current default for container image signing is Sigstore's **keyless** signing: Cosign authenticates via OIDC, Fulcio issues a short-lived certificate, and Rekor (a public transparency log) records the signature. This is genuinely the modern standard — but verifying a keyless signature means checking it against Rekor, which is public Sigstore infrastructure reachable only over the internet. That's the same shape of problem [Encryption](../07-security/README.md#encryption)'s cloud-KMS rejection and [Authentication](../07-security/README.md#authentication)'s SSO-IdP-must-be-internal finding already solved elsewhere in this doc set: a verification step depending on a public external service silently breaks the air-gapped requirement from [System Context](../03-architecture/README.md#system-context).

**Resolution: Imora signs images with a traditional Cosign key pair, not keyless.** The private key stays with the build pipeline; the public key ships embedded in every deployment bundle (connected or air-gapped) and in the Compose/Kubernetes manifests themselves. Verification — whether at a connected deployment's registry pull or an air-gapped deployment applying a transferred bundle — checks the signature against that embedded public key locally, with zero dependency on Rekor, Fulcio, or any other public Sigstore service. This is a small deviation from the "modern default," made for the same reason every other air-gapped-compatibility decision in this doc set was made.

---

### Pipeline Stages

1. **Build** — multi-stage builds per [Docker](README.md#docker)'s conventions, producing minimal, non-root, read-only-filesystem-compatible images.
2. **Test**, including a specific gate this document adds: **automated verification of the structural guarantees this doc set has been establishing**, not just functional tests. Concretely: a test that attempts `DELETE`/`ALTER` against `access_audit_events` using the `ingestion`/`query-api` service account credentials and asserts it fails, per [Threat Model](../07-security/README.md#threat-model)'s GRANT-restriction finding; a test confirming containers actually run as non-root with a read-only root filesystem, per [Docker](README.md#docker). This turns "should be true" claims made throughout `02-domain/`, `05-data/`, and `07-security/` into CI-verified facts, not just design intentions that could silently drift as the codebase changes.
3. **SBOM generation** — Syft generating SPDX or CycloneDX format, attached to the image as a signed Cosign attestation. Worth doing regardless of any specific mandate, given [Target Users](../00-overview/README.md#target-users) includes government agencies as a target sector, where software supply-chain provenance is an increasingly standard procurement expectation.
4. **Sign** — the key-pair Cosign signing from the finding above.
5. **Publish, two paths from one build:**
   - **Registry push**, for connected deployments' `docker pull`/Kubernetes image pull.
   - **Signed bundle packaging**, for the air-gapped transfer mechanism [Deployment Model](../03-architecture/README.md#deployment-model) already specified — staged, then moved via approved removable media. Producing both from the identical build artifact is what guarantees a connected customer and an air-gapped customer are running the exact same code, not two paths that could quietly diverge.

---

### What's Deliberately Not Modeled Here

- Specific CI platform (GitHub Actions, GitLab CI, etc.) — a tooling choice, not an architecture decision.
- Exact test framework/coverage thresholds — `11-engineering/README.md#testing`.
- Key-pair rotation schedule for the Cosign signing key itself — follows the same versioned-key principle as [Encryption](../07-security/README.md#encryption)'s KEK rotation, not repeated here.

---

Sources: [Signing Containers — Sigstore Docs](https://docs.sigstore.dev/cosign/signing/signing_with_containers/), [Container Supply Chain Security With Sigstore and Cosign](https://devopsil.com/articles/2026-03-21-supply-chain-security-sigstore-cosign), [How to Sign an SBOM with Cosign — Chainguard Academy](https://edu.chainguard.dev/open-source/sigstore/cosign/how-to-sign-an-sbom-with-cosign/).

### What This Feeds Next

`docs/12-infrastructure/README.md#observability` is the last file in this folder — monitoring Imora's own infrastructure, distinct in scope from the product's own observability features it sells to customers.

---

## Observability

> Status: The last file in `12-infrastructure/` — how Imora monitors its own infrastructure. Scoped deliberately: this is monitoring of Imora's *own* services (gateway, ingestion, query-api, alert-engine, workers, notification-service, dashboard), distinct from the product features Imora sells to customers for monitoring *their* applications — the same scope boundary [Audit Logging](../07-security/README.md#audit-logging) drew for operational versus product logging.

---

### Why Not Dogfood Imora on Itself

Using Imora's own product to monitor Imora's own infrastructure is tempting — and creates a circular blast-radius problem: if the thing that's broken is Imora's own observability pipeline, self-monitoring with itself means the tool you'd reach for to diagnose the outage is the thing that's down. **Infrastructure self-monitoring runs on a separate, standard stack** (Prometheus-style metrics collection, structured logs, a simple alerting path independent of `notification-service`) — simpler than the product, and specifically not dependent on any of the eight bounded contexts being healthy to report that they aren't.

---

### Closing Three Gaps This Doc Set Left Open

This is the third time the same shape of problem has come up: a mechanism gets specified as existing, but nothing was ever assigned to actually watch it. [Threat Model](../07-security/README.md#threat-model) flagged two of these directly; a third was sitting unaddressed in [Scaling](../03-architecture/README.md#scaling). All three get an owner here, in infrastructure monitoring specifically, because none of the three are product features Imora exposes to customers — they're operational signals for whoever runs the deployment:

1. **Sequence-number gap detection** ([Threat Model](../07-security/README.md#threat-model), Repudiation) — the periodic `workers` integrity-check job's output is surfaced as an infrastructure alert, not just a `notification-service` message that depends on that same job's health to deliver.
2. **UNMASK-frequency review** ([Threat Model](../07-security/README.md#threat-model), Elevation of Privilege) — the per-actor unmask-frequency report becomes a standing infrastructure dashboard, visible to Platform Operator role continuously, not just a periodic push.
3. **Scaling threshold monitoring** — [Scaling](../03-architecture/README.md#scaling) specified that a cluster migration should be planned once accumulated storage reaches roughly 50% of the single-machine SSD allocation, but never specified anything that watches for it. **This document closes that gap:** infrastructure monitoring tracks accumulated ClickHouse storage against that 50% threshold directly and alerts Priya's role when it's approached — turning [Scaling](../03-architecture/README.md#scaling)'s calculation into an operational trigger instead of a number a human has to remember to check.

---

### What Gets Monitored

- **Service health** — uptime, error rate, latency per service, standard practice for the eight bounded contexts.
- **CronJob execution** — specifically whether `RetentionSweepScheduler` (per [Kubernetes](README.md#kubernetes)'s `concurrencyPolicy: Forbid` design) is actually completing on schedule, or silently getting skipped run after run because a prior run never finishes — a Forbid policy that's always skipping is a real operational problem masquerading as a working safeguard.
- **Data store resource utilization** — ClickHouse, PostgreSQL, MinIO, Redis, per the sizing in [Deployment Model](../03-architecture/README.md#deployment-model), feeding the scaling-threshold alert above.
- **The three gap-closing signals** above.

---

### What's Deliberately Not Modeled Here

- Specific tool selection (Prometheus/Grafana vs. an alternative stack) — implementation choice, not architecture.
- Alert routing/on-call configuration — an operational runbook concern, downstream of this design.
- Log retention for infrastructure logs themselves — distinct from and much shorter than [Retention](../05-data/README.md#retention)'s product-data retention clocks, since infrastructure logs carry no regulatory retention obligation.

---

### What This Closes Out

This is the last file in `docs/12-infrastructure/`. All five files — [Docker Compose](README.md#docker-compose), [Docker](README.md#docker), [Kubernetes](README.md#kubernetes), [CI/CD](README.md#cicd), and this one — are now internally consistent. `docs/11-engineering/` is next — team conventions and the ADR pattern already scaffolded there, which should retroactively document several of the load-bearing decisions made across this entire doc set as formal ADRs (AGPLv3 licensing, the dual ClickHouse/Postgres store split, key-pair over keyless signing, among others).

