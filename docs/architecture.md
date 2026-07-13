# Imora — Software Architecture, Tech Stack & System Flow

> Build-ready architecture reference. Every decision below has a documented rationale, trade-off analysis, and (where relevant) threat-model stress-test in [`research/`](../research/README.md) — this file states the resulting design; `research/` states why it isn't some other design. Section headers link to the source document.

---

## 1. System Overview

Imora is one system, built from **eight bounded contexts** split along a write path and a read path, with a background context that owns everything compliance-critical:

```
Write path:   browser-sdk → gateway → ingestion → [ClickHouse, PostgreSQL]
Read path:    dashboard → gateway → query-api → [ClickHouse, PostgreSQL, MinIO]
Background:   workers (retention sweeps, legal-hold enforcement, evidence export)
Async:        alert-engine (grouping, regression detection) → notification-service
```

Two topology profiles, not one: a single Docker Compose host (Milestone 1, 2–3 person teams) and a Kubernetes cluster (Milestone 3, large-enterprise scale). **Air-gapping is an orthogonal setting on either profile** — not a third variant — and every Parity/Wedge capability must work identically with zero outbound network access.

Full C4-model diagrams (context → container → component → sequence): [`research/03-architecture/README.md`](../research/03-architecture/README.md) and [`research/03-architecture/diagrams.md`](../research/03-architecture/diagrams.md).

---

## 2. Tech Stack

| Layer | Technology | Why |
|---|---|---|
| **Backend services** (gateway, ingestion, query-api, alert-engine, workers, notification-service) | **Go** | Compiles to a static binary → minimal distroless container images; low memory footprint (matters on a single 16GB host where ClickHouse already claims most of the budget); ecosystem alignment (ClickHouse's Go client, most Kubernetes-native tooling is Go-first). |
| **browser-sdk** | **TypeScript**, framework-agnostic core (`@imora/core`) + thin per-framework wrappers (`@imora/react`, `@imora/vue`, `@imora/angular`) | Same pattern as Sentry's SDK family — a new framework wrapper is a thin adapter, not a parallel implementation. |
| **dashboard** | **TypeScript**, [TanStack Start](https://tanstack.com/start) (SSR, runs on Node via Nitro — not a static SPA) | Consumes `query-api` exclusively, from both server-rendered loaders and client-side navigation — no domain logic in the frontend, and server functions/loaders are a proxy to `query-api`'s REST API, never a direct store connection. |
| **High-volume event store** | **ClickHouse** | SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, AccessAuditEvent — all append-only, time-series-shaped. Column store built for exactly this write/query pattern. |
| **Relational store** | **PostgreSQL** | Session summaries, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata, users/RBAC — small-cardinality, transactionally-updated, needs ACID guarantees ClickHouse doesn't provide. |
| **Object storage** | **MinIO** (S3-compatible, self-hosted) | EvidenceExport frozen blobs, using **Object Lock in Compliance mode** — true WORM immutability, not just an application-level promise. |
| **Cache** | **Redis** | Active-LegalHold lookup cache (for the retention-sweep hold check) and gateway rate-limiting. Never a source of truth — fully rebuildable from Postgres. |
| **Message queue (cluster profile only)** | Kafka or equivalent | Absent in the single-machine profile entirely — `ingestion` writes directly to ClickHouse there. Introduced only at cluster scale to buffer write bursts and enable reprocessing. |
| **Container images** | Multi-stage, distroless/slim base, non-root, read-only root filesystem where possible | Standard hardening — also a direct mitigation against the Tampering threats identified in the threat model. |
| **Image signing** | **Cosign, traditional key pair — not keyless** | Keyless (Sigstore/Fulcio/Rekor) verification requires reaching public infrastructure, which breaks air-gapped verification. A key pair embedded in every deployment bundle verifies locally, zero external dependency. |
| **SBOM** | Syft (SPDX/CycloneDX), signed as a Cosign attestation | Supply-chain provenance — relevant given government agencies are a target sector. |

Full reasoning per choice: [`research/11-engineering/README.md#coding-standards`](../research/11-engineering/README.md#coding-standards) (language), [`research/03-architecture/diagrams.md#container-diagrams`](../research/03-architecture/diagrams.md#container-diagrams) (data stores), [`research/12-infrastructure/README.md`](../research/12-infrastructure/README.md) (Docker/CI-CD/signing).

---

## 3. The Eight Services

| Service | Responsibility | Owns (writes) | Reads |
|---|---|---|---|
| **browser-sdk** | Client-side capture-time masking + rrweb-style event serialization + Core Web Vitals measurement. The *only* place capture-time masking can be enforced — by the time data leaves the browser it must already be safe. | Nothing server-side — emits to `gateway` | — |
| **gateway** | AuthN/authZ, rate limiting, actor-context stamping. The single chokepoint every read/write passes through. | Nothing — annotates requests | Redis (rate limits) |
| **ingestion** | The write path. Accepts and persists SessionEvent/ErrorEvent/PerformanceMetric/SecurityEvent/TraceLink, append-only. | ClickHouse (events), Postgres (Session/Release summary rows) | — |
| **query-api** | The read path. Serves replay/error/metric queries. **Generates the AccessAuditEvent for every VIEW/EXPORT/UNMASK — structurally, not by convention** (see Section 5). | AccessAuditEvent (ClickHouse) | ClickHouse, Postgres, MinIO (export metadata) |
| **alert-engine** | ErrorGroup grouping (write-time, not display-time) and release-attributed Core Web Vitals regression detection. | ErrorGroup (Postgres) | ClickHouse (events), Postgres |
| **workers** | The compliance-critical background context: RetentionPolicy sweeps, legal-hold check-before-destroy, selective GDPR-erasure purging, EvidenceExport generation. | RecordDeleted/DeletionSkipped/EvidenceExportGenerated (ClickHouse), export blobs (MinIO) | Redis (hold cache), Postgres, ClickHouse |
| **notification-service** | Delivery only — translates an alert signal into email/Slack/webhook. Deliberately decoupled from `alert-engine`'s grouping logic. | Nothing domain-relevant | Consumes from `alert-engine` |
| **dashboard** | Presentation only, [TanStack Start](https://tanstack.com/start) (SSR). **Zero domain entities, zero audit-log authority** — calls `query-api` exclusively, never a data store directly, including from server functions/loaders. | Nothing | `query-api` |

**The one rule every service boundary above exists to enforce:** AccessAuditEvent generation belongs exclusively to `query-api` (reads) and `workers` (retention/hold/export actions) — never to `dashboard` or any presentation layer. A UI convention generating audit events would mean anyone calling the API directly bypasses the entire wedge. This applies to `dashboard`'s server-side code (route loaders, server functions) exactly as it applies to its client-side code — SSR changes *where* the HTTP call to `query-api` happens, never *what* it's allowed to call.

Full bounded-context modeling (Shared Kernel / Customer-Supplier / Conformist relationships): [`research/02-domain/README.md#bounded-contexts`](../research/02-domain/README.md#bounded-contexts).

---

## 4. Domain Model, Condensed

**Parity entities:** Session (aggregate root), SessionEvent (rrweb-style stream: FullSnapshot/DOMMutation/MouseMove/Click/Scroll/FormInput/ViewportChange), ErrorEvent, ErrorGroup (deduplicated at write time), Release, PerformanceMetric (LCP/INP/CLS at p75), TraceLink.

**Environment** is a free-text tag (`production`/`staging`/`development`, or a team's own values), set once at SDK `init()` alongside `release` and denormalized onto every event — **a query dimension, not a managed entity.** It doesn't change what's captured or how rigorously it's protected (compliance rigor is identical across environments, deliberately — non-prod is routinely seeded with real production data in practice, so "relax the rules for staging" is a trap, not a convenience). The one place it's load-bearing rather than cosmetic: `Release.priorReleaseId` and regression-detection comparisons are always scoped **within one environment** — the same version deployed to `staging` then `production` is two separate `Release` rows, and comparing across them would produce a nonsense signal.

**Wedge entities:**
- **AccessAuditEvent** — append-only: `{eventId, actorUserId, action, targetRecordType, targetRecordId, timestamp, sourceIp/device, sequenceNumber}`. `action` ∈ {VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED, CONFIG_CHANGED}.
- **RetentionPolicy** — per data category, not per record: `{dataCategory, retentionPeriod, regulatoryBasis}`.
- **LegalHold** — `{holdId, appliedBy, reason, scope, appliedAt, liftedAt}` — `scope` is a **re-evaluated query predicate**, not a fixed record list, so newly-created matching records are automatically covered.
- **EvidenceExport** — `{exportId, requestedBy, incidentReference, generatedAt, contentHash, format}` — a **frozen, self-contained copy** at generation time, never live references, so a later retention purge can't invalidate an already-generated export.
- **SecurityEvent** — optionally tied to a Session, for incident-timeline correlation.

**Five invariants that must hold regardless of implementation:**
1. AccessAuditEvent is append-only — enforced at the storage layer (DB-level GRANTs), not just application code.
2. A retention-driven deletion must check for an active LegalHold **immediately before executing**, not at scheduling time (race-condition prevention).
3. EvidenceExport is frozen at generation time — immune to any later purge or erasure.
4. PII/PHI masking happens at capture time, not query time — an unmasked value must never exist at rest for an unclassified field.
5. Every UNMASK requires a non-empty reason and produces its own audit event.

Full entity/relationship detail, business rules (BR-1 through BR-7), and the full event catalog: [`research/02-domain/README.md`](../research/02-domain/README.md).

---

## 5. The Structural Guarantee (Why the Audit Trail Can't Be Skipped)

This is the single most important implementation detail in the whole system. Stating "every read must be audited" as a requirement doesn't make it true — engineers forget, or a new endpoint gets added without the audit call. The fix: **`AuditedQueryHandler` is the only way to register a read route in `query-api` at all.** There is no code path to reach ClickHouse without passing through `AccessAuditWriter` first — the same principle as a build failing when a required wiring seam is absent, applied to routing instead of compilation.

```
gateway (actor identity) → RequestContext → AuditedQueryHandler
                                                 ├─ SessionQueryHandler
                                                 ├─ ErrorQueryHandler
                                                 ├─ PerformanceQueryHandler
                                                 ├─ TraceCorrelationHandler
                                                 ├─ AuditQueryHandler (reads of the audit trail itself)
                                                 └─ UnmaskEscalationHandler (requires non-empty reason)
                                                       ↓
                                                 AccessAuditWriter → ClickHouse
```

Same pattern in `workers`: `DeletionExecutor` has no incoming path that bypasses `LegalHoldChecker` — the interaction order (`RetentionSweepScheduler` → `LegalHoldChecker` → `SelectivePurgeExecutor` → `DeletionExecutor`) is enforced by the component wiring itself, not a runbook step someone has to remember.

**Two-tier masking, resolved:** hard redaction (unrecognized field — never captured, anywhere, ever, nothing to unmask) vs. soft masking with escalation (known PHI/PII field — real value captured into an encrypted `SecureFieldVault`, unmaskable only through the audited `UnmaskEscalationHandler`).

Full component diagrams and the four traced sequence flows (Session Capture, DSAR Query, Retention Sweep Hitting a Hold, Evidence Export): [`research/03-architecture/diagrams.md#component-diagrams`](../research/03-architecture/diagrams.md#component-diagrams) and [`#sequence-diagrams`](../research/03-architecture/diagrams.md#sequence-diagrams).

---

## 6. Data Architecture

### ClickHouse (high-volume, append-only)

Every table below also carries a denormalized `environment` column (`LowCardinality(String)`) alongside `release_id` — same reasoning as `release_id` itself: avoid a join in a column store.

| Table | Order By | TTL |
|---|---|---|
| `session_events` | `(session_id, occurred_at)` | Per-category clock, conditional on legal hold via dictionary lookup |
| `error_events` | `(stack_trace_fingerprint, occurred_at)` | Same pattern |
| `performance_metrics` | `(environment, release_id, metric_type, occurred_at)` | Same pattern — `environment` leads here specifically because regression detection's dominant query compares releases *within* an environment, unlike the session/error tables where a single session is already environment-scoped by construction. |
| `security_events` | `(session_id, occurred_at)` | Same pattern |
| `access_audit_events` | `(target_record_id, occurred_at)` | **Longest applicable clock across all categories** — never younger than what it audits |

**The key technical resolution:** ClickHouse's TTL clause supports conditional `WHERE` expressions — `TTL occurredAt + INTERVAL 6 YEAR DELETE WHERE dictGetOrDefault('active_legal_holds', 'is_held', sessionId, 0) = 0`. This lets per-category retention (regulatory requirement) and per-row legal-hold enforcement (also a regulatory requirement) coexist without expensive row-level UPDATEs. Partitioning is **monthly** (not yearly) specifically to limit how much data one active legal hold can pin out of the cheap partition-drop path.

Engine: plain `MergeTree` (genuinely append-only, no "latest version wins" semantics needed).

### PostgreSQL (relational, small-cardinality)

Tables: `sessions` (carries its own `environment` column), `releases` (**composite PK `(id, environment)`** — the same version string deployed to `staging` then `production` is two rows, not one, and `prior_release_id` resolves within the same environment), `error_groups`, `retention_policies`, `legal_holds` (scope stored as a structured JSONB predicate — `{"type": "session_ids"|"data_subject"|"date_range"|"incident_reference", ...}` — re-evaluated on every ClickHouse dictionary refresh, never a raw SQL string), `evidence_exports` (metadata + `object_storage_path` pointer), `users` (role: engineer | compliance_officer | platform_operator | admin).

**No cross-store foreign keys** — ClickHouse can't reference Postgres. `ingestion`/`alert-engine` are responsible for writing the Postgres row before or atomically with the ClickHouse write.

### Object Storage (MinIO)

EvidenceExport blobs only, at `imora-evidence-exports/{yyyy}/{mm}/{export_id}.tar.gz`, with **Object Lock in Compliance mode set at bucket creation** (this cannot be applied retroactively — bucket init must run before the first export, enforced as a hard dependency in the deployment startup order).

Full schema definitions, event-schema versioning rule ("additive-only, forever" — a stored audit record must remain readable for up to 7 years), and the storage-layer rationale: [`research/05-data/README.md`](../research/05-data/README.md).

---

## 7. API Surface

**REST** (`/v1/...`, URL-path versioned, ≥2 concurrent versions supported, 6-month deprecation notice):

| Endpoint | Method | Purpose |
|---|---|---|
| `/v1/sessions` | GET | Session search — filterable by `environment`/`release`/user/URL/timeframe. `environment` defaults to `production`, never "all," so a search never silently mixes non-prod noise into a real investigation. |
| `/v1/sessions/{id}` | GET | Session + replay |
| `/v1/sessions/{id}/audit-trail` | GET | Access history for that session |
| `/v1/data-subjects/{id}/sessions` | GET | DSAR query surface — sessions + access history, one lookup |
| `/v1/sessions/{id}/unmask` | POST | UNMASK escalation, `reason` required at the schema level |
| `/v1/legal-holds` | POST / DELETE | Apply / lift a hold |
| `/v1/retention-policies/{category}` | GET / PUT | Per-category retention (PUT produces a `CONFIG_CHANGED` audit event) |
| `/v1/evidence-exports` | POST / GET | Trigger / retrieve an export |
| `/v1/erasure-requests` | POST | GDPR erasure intake |

**Every read through this API — dashboard or third-party integration — goes through the same `AuditedQueryHandler`.** There is no lower-friction API path around the audit guarantee.

**SDK** (`browser-sdk` public surface): `init(config)` — takes `projectKey`, `release`, and `environment` (all free-text tags, none of which change capture/masking behavior), PII/PHI classification config (`data-imora-safe`/`data-imora-mask` attributes or programmatic selectors), `identify(userId, traits)`, `captureException(error, context)`, `addBreadcrumb()`/`setContext()`, automatic Core Web Vitals capture. Performance budget: match or beat Sentry's ~20KB gzipped post-optimization figure — the SDK meant to catch a Core Web Vitals regression cannot itself cause one.

**Webhooks** (outbound, customer-configured): `AlertTriggered`, `RegressionDetected`, `EvidenceExportGenerated`, `ConfigurationChanged`. HMAC-SHA256 signed, SSRF-protected on every delivery attempt (not just at config time — DNS rebinding defense). **Payloads never carry sensitive content — identifiers and links back to the authenticated REST API only.**

Full endpoint-by-endpoint detail and the OpenAPI spec: [`research/06-api/README.md`](../research/06-api/README.md) and [`research/06-api/openapi.yaml`](../research/06-api/openapi.yaml).

---

## 8. Security Model

| Concern | Design |
|---|---|
| **Local auth** | Argon2id password hashing (m=19456, t=2, p=1), TOTP MFA (not SMS — no external dependency, works air-gapped). Required baseline on every tier, including Enterprise. |
| **SSO** | SAML 2.0 / OIDC, Enterprise-tier only, gated by an offline signed license file. Works air-gapped by pointing at an IdP *inside* the same isolated network (self-hosted Keycloak, ADFS) — air-gapped ≠ no internal services. |
| **Authorization** | RBAC baseline (4 roles) for routing + **one ABAC boundary**: the UNMASK action specifically, using the break-the-glass pattern (immediate access, mandatory logged reason, no approval-workflow delay). No role is exempt, including Admin. |
| **Encryption** | Two layers: disk/volume-level (outer, uniform) + field-level envelope encryption via `SecureFieldVault` (inner, AES-256-GCM, only for soft-masked PHI/PII fields). TLS everywhere. KEK rotation is lazy/versioned — no bulk re-encryption on rotation. |
| **PII/PHI classification** | Three-input decision at capture time: explicit safe-allowlist → capture as-is; explicit PHI marker or regex-backstop match → soft-mask into vault; neither → hard-redact (fail closed). |
| **Database-level tamper resistance** | `ingestion`/`query-api` service accounts get **INSERT + SELECT only** on `access_audit_events` — no application credential, including `workers`', ever holds DELETE/ALTER on that table. |
| **Operational audit logging** | Distinct from AccessAuditEvent: `CONFIG_CHANGED` covers policy edits, field-reclassification, role grants — closes the "who changed the rules" gap a pure data-access log misses. |

Full threat model (STRIDE-based, including two non-obvious findings — the DB-GRANT gap and LegalHold-as-DoS-vector against the scaling math) and the full authorization matrix: [`research/07-security/README.md`](../research/07-security/README.md).

---

## 9. Deployment

### Single-Machine Profile (Milestone 1)

All eight services as Docker Compose containers on **one 4-core / 16GB RAM host with SSD storage** (ClickHouse is the dominant resource consumer). No message queue — `ingestion` writes directly to ClickHouse.

Startup order is a hard dependency chain, not a suggestion: data layer (Postgres/ClickHouse/Redis/MinIO, health-checked) → `minio-init` (bucket versioning + Compliance-mode lock — **must run before the first EvidenceExport, cannot be applied retroactively**) → schema migrations → application services. A correct dependency graph here is literally what makes the "under 1 hour, unassisted" onboarding target achievable.

### Cluster Profile (Milestone 3)

Kubernetes. Services scale independently per load shape (write-heavy `ingestion` vs. read-latency-sensitive `query-api`). ClickHouse/Postgres move to multi-node `StatefulSet`s. A message queue is introduced between `ingestion` and its consumers.

**`workers` is not one Deployment** — it splits into a `CronJob` (`concurrencyPolicy: Forbid`, for `RetentionSweepScheduler` — prevents two replicas racing the same hold-check-before-delete logic) and an ordinary scalable `Deployment` (for `EvidenceExportGenerator`, independently partitioned per request).

`NetworkPolicy` restricts `dashboard`'s egress to `query-api` only, at the network layer — even a compromised dashboard pod can't reach a data store directly.

### Air-Gapped (orthogonal to both profiles above)

No SSO, no outbound notifications, no live backend correlation — but full-strength audit trail, masking, retention enforcement regardless. Updates and license activation both use the same pattern: **signed bundle, staged outside the air gap, transferred via approved removable media, verified locally against an embedded public key** — no phone-home, ever.

### Scaling Trigger

**Retention-driven accumulated storage, not throughput.** At ~100,000 sessions/month under a 6-year HIPAA floor, accumulated storage lands in the 2.8–4.2TB range — well past a comfortable single-machine ceiling — while ClickHouse's actual write-throughput ceiling (hundreds of thousands to millions of rows/sec) is nowhere close to being tested. **Plan cluster migration at ~50% of the single-machine SSD allocation**, roughly the 50,000–70,000 sessions/month mark for a 6-year-retention org. Migration changes deployment topology only — the domain model, business rules, and event catalog are identical on both profiles.

Full sizing numbers, the MinIO ordering dependency, Kubernetes manifests-level detail, and the full scaling math: [`research/03-architecture/README.md#deployment-model`](../research/03-architecture/README.md#deployment-model), [`#scaling`](../research/03-architecture/README.md#scaling), and [`research/12-infrastructure/README.md`](../research/12-infrastructure/README.md).

---

## 10. Repository Structure

Monorepo, single root `LICENSE` (AGPLv3) applying uniformly — **no per-directory license overrides, ever.** This is a direct, structural fix for the PostHog `ee/`-directory anti-pattern (PostHog-FOSS is literally "a mirror with proprietary code removed" — a monorepo doesn't prevent that pattern by itself, the license-per-root rule does).

```
imora/
├── LICENSE                  # AGPLv3, root-only
├── services/                 # gateway, ingestion, query-api, alert-engine, workers, notification-service
├── sdk/browser-sdk/           # independently versioned (npm semver) — still root-licensed
├── dashboard/                 # TanStack Start (SSR) — server functions/loaders proxy to query-api only
├── packages/
│   ├── domain-types/          # the Shared Kernel: Session, SessionEvent, AccessAuditEvent, etc.
│   └── event-schemas/         # event payload shapes, shared by producers and consumers
├── deploy/
│   ├── compose/                # single-machine manifests
│   └── kubernetes/             # cluster manifests
├── research/                      # research/rationale (this doc's source material)
└── docs/                      # this folder — build-ready spec
```

Milestone 3 commercial features (SSO, multi-region tooling) live **inside** `services/gateway/` and `deploy/kubernetes/`, gated by a runtime license-file check — never a separate directory or license.

Full reasoning: [`research/03-architecture/README.md#repository-structure`](../research/03-architecture/README.md#repository-structure).

---

## 11. Engineering Practices

- **Branching:** trunk-based. Feature branches live hours to a couple of days; incomplete work ships behind a flag rather than staying on a long-lived branch — deliberately strict here because `packages/domain-types` is a literal shared kernel across three services, and a long-lived branch touching it accumulates drift risk across all three at once.
- **Testing:** unit (business-rule logic in isolation), integration (against **real infrastructure, not mocks** — e.g. actually attempting a `DELETE` against `access_audit_events` with the `ingestion` service account and asserting it fails), end-to-end (the four sequence-diagram flows *are* the e2e spec, not a separate scenario set to invent).
- **CI gates worth calling out specifically:** a JSON-schema diff check that **fails the build on any breaking change** to an event schema (matching the "additive-only, forever" rule); an OpenAPI diff check with a *looser* policy (breaking changes allowed, gated on a major version + deprecation notice) — same tooling category, two different policies, because a 7-year-old stored record and a request/response contract have different compatibility guarantees.
- **Release:** semver for the software itself, independent of event `schemaVersion` (additive-only) and the API's `/v1/` path version (can break at a major boundary) — three separate version axes, not one number wearing three hats. Every release dual-publishes from one signed build artifact: registry push (connected) + signed bundle (air-gapped).
- **Code-level and architectural conventions for the Go backend:** [`coding-standards.md`](coding-standards.md) (error handling, structured logging, interfaces, concurrency, testing) and [`design-system.md`](design-system.md) (hexagonal layering, package boundaries, dependency injection) — both build-ready, both in this folder.

Full detail: [`research/11-engineering/README.md`](../research/11-engineering/README.md).

---

## 12. What's Deliberately Not Decided Here

Consistent with the rest of this doc set's discipline about not inventing detail ahead of need:
- Exact ORM/migration tooling, specific KMS/HSM product for cluster-scale key management, CI platform (GitHub Actions vs. GitLab CI), specific test framework.
- Build tooling for the monorepo (Nx, Turborepo, Bazel, or none).
- Exact HPA scaling thresholds, ClickHouse sharding key selection at cluster scale.

These are implementation choices downstream of the design above, not gaps in the design itself — make them when the code that needs them actually gets written.

---

## What This Feeds

Implementation. This is the last "planning" document in the chain — `prd.md` → `user-stories.md` → `architecture.md` → code.
