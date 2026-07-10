# Bounded Contexts

> Status: Research-based, current as of July 2026. Assigns each entity in [domain-model.md](domain-model.md) to an owning service boundary. The eight service names below are not proposed here — they're already scaffolded in `docs/05-services/` (gateway, ingestion, query-api, alert-engine, workers, browser-sdk, dashboard, notification-service), which is a real prior decision this document has to honor, not invent around.

---

## Modeling Approach

The context boundaries follow two established patterns rather than an arbitrary split:

- **Write-path / read-path separation**, the standard shape for high-ingest observability systems built on column stores like ClickHouse: ingestion is append-oriented and deliberately minimizes per-row processing, while query serving runs as an independently scaled concern. Conflating the two — having one service both accept high-volume writes and serve latency-sensitive reads — is the architectural mistake this split exists to avoid.
- **DDD context mapping**, specifically three relationship types used below: **Shared Kernel** (contexts that operate directly on the same entity definitions from [domain-model.md](domain-model.md) and must stay in lockstep), **Customer-Supplier** (an upstream context whose downstream customer can influence its priorities, but the dependency direction is one-way), and **Conformist** (a downstream context that just accepts the upstream's model as-is, doing no translation of its own).

---

## The Eight Contexts

### browser-sdk

**Responsibility:** client-side capture — the only place [domain-model.md](domain-model.md)'s Invariant 4 (masking at capture time, not query time) can actually be enforced, since by the time data leaves the browser it must already be safe. Owns the rrweb-style event serialization (FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange) and Core Web Vitals measurement at the point of occurrence.

**Entities produced:** SessionEvent (pre-masked), PerformanceMetric (raw measurement, not yet attributed to a Release).

**Context relationship:** Shared Kernel with `ingestion` on the SessionEvent/PerformanceMetric shape — the wire format browser-sdk emits and the format ingestion accepts must be the same entity definition, not two shapes translated between, or masking guarantees applied in the SDK could be silently lost in translation.

---

### gateway

**Responsibility:** authentication, request routing, and field-level access control enforcement — the chokepoint every read and write passes through before reaching a domain context. This is also where actor identity and source IP get attached to any subsequent AccessAuditEvent, since it's the only context that has verified who's asking.

**Entities produced:** none directly — gateway annotates requests with authenticated-actor context that `query-api` and `workers` consume when they generate AccessAuditEvent entries.

**Context relationship:** Customer-Supplier, upstream of both `ingestion` and `query-api` — both depend on gateway's authenticated identity, but gateway doesn't depend on either's internal model.

---

### ingestion

**Responsibility:** the write path. Accepts SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, and TraceLink writes from browser-sdk (and backend instrumentation, for TraceLink) and persists them append-only, deliberately avoiding read-serving concerns — per the write/read separation pattern above.

**Entities owned:** Session, SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, TraceLink (all write-side).

**Context relationship:** Shared Kernel with `browser-sdk` (as above) and with `query-api` (both operate on the same Session/ErrorEvent/PerformanceMetric definitions from [domain-model.md](domain-model.md) — a divergence here would mean query-api serving a different shape than ingestion wrote, which is exactly the kind of split a shared kernel is meant to prevent).

---

### query-api

**Responsibility:** the read path — serving Session, replay, ErrorGroup, and PerformanceMetric queries to `dashboard`, and correlating TraceLink lookups (story J1). **This is also the only place AccessAuditEvent may be generated for read access**, and that's a load-bearing architectural decision, not a detail: per [domain-model.md](domain-model.md)'s modeling approach, an audit trail enforced anywhere else — in `dashboard`, for instance — is bypassable by anyone calling the API directly, which would make the entire wedge (stories M1, A1) a UI convention instead of a guarantee. Every VIEW of a Session, replay, or masked field must produce an AccessAuditEvent as an inseparable part of serving that read, not an optional side effect a caller could skip.

**Entities owned:** read-side of Session/SessionEvent/ErrorGroup/PerformanceMetric; **owns AccessAuditEvent generation for all VIEW and UNMASK actions.**

**Context relationship:** Shared Kernel with `ingestion` (as above); Customer-Supplier downstream of `gateway`; upstream supplier to `dashboard`.

---

### alert-engine

**Responsibility:** ErrorGroup grouping/deduplication (story C2 — grouping happens here, not in the UI, per [domain-model.md](domain-model.md)'s note that grouping is write-time/processing-time, not display-time) and release-attributed regression detection against Core Web Vitals thresholds (story C1).

**Entities owned:** ErrorGroup, Release-regression evaluation logic. Reads ErrorEvent/PerformanceMetric/Release from the ingestion-owned store.

**Context relationship:** Customer-Supplier, downstream of `ingestion`'s written data; upstream supplier to `notification-service`.

---

### workers

**Responsibility:** the background/async context — the only place two of the sharpest wedge invariants actually execute: RetentionPolicy-driven scheduled deletion (which must check LegalHold before every deletion, per Invariant 2) and EvidenceExport generation (which must freeze a self-contained copy at generation time, per Invariant 3 and the resolved open question in [domain-model.md](domain-model.md)). Both are exactly the kind of correctness-critical, no-user-watching logic that belongs in a dedicated worker context rather than inline in a request path.

**Entities owned:** RetentionPolicy execution, LegalHold evaluation, EvidenceExport generation.

**Context relationship:** Customer-Supplier, downstream consumer of `ingestion`'s stored data and `gateway`'s actor context (an EvidenceExport request or a LegalHold application still needs to know who requested it, for the AccessAuditEvent it produces).

---

### notification-service

**Responsibility:** delivery only — translating an alert-engine-produced signal into an email, Slack message, or webhook. Deliberately separated from `alert-engine` so grouping/regression logic isn't coupled to delivery mechanism; a new notification channel shouldn't require touching alerting logic at all.

**Entities owned:** none from [domain-model.md](domain-model.md) — this context's data (delivery channels, message templates) is orthogonal to the core domain.

**Context relationship:** Conformist, downstream of `alert-engine` — it accepts whatever shape alert-engine produces without attempting to reinterpret or enrich it.

---

### dashboard

**Responsibility:** presentation only. Renders what `query-api` and `alert-engine` (via `notification-service` or directly) return. Has no domain entities of its own and, critically, **must not be where any invariant is enforced** — masking, audit logging, and access control all have to hold even if a caller bypasses the dashboard UI entirely and calls query-api or gateway directly.

**Entities owned:** none.

**Context relationship:** Conformist, downstream of `query-api` — takes the served shape as-is. This is a deliberate choice, not a limitation: giving dashboard any translation authority over sensitive fields would create a second place PII-masking logic could drift from `browser-sdk`'s capture-time enforcement.

---

## Context Map Summary

| Upstream | Downstream | Relationship | Why |
|---|---|---|---|
| browser-sdk | ingestion | Shared Kernel | Same SessionEvent/PerformanceMetric shape, capture-time masking must survive the handoff intact |
| gateway | ingestion, query-api | Customer-Supplier | Both depend on gateway's authenticated actor context |
| ingestion | query-api | Shared Kernel | Same Session/ErrorEvent/PerformanceMetric definitions on write and read sides |
| ingestion | alert-engine | Customer-Supplier | Alert-engine reads what ingestion wrote |
| ingestion, gateway | workers | Customer-Supplier | Retention/legal-hold/export jobs need both stored data and actor identity |
| alert-engine | notification-service | Conformist | Delivery accepts the alert shape as-is |
| query-api | dashboard | Conformist | Presentation must not have translation authority over sensitive fields |

---

## The One Rule This Document Adds to domain-model.md

**AccessAuditEvent generation belongs exclusively to `query-api` (for reads) and `workers` (for retention/legal-hold/export actions) — never to `dashboard` or any presentation layer.** This follows directly from [domain-model.md](domain-model.md)'s Invariant 1 (audit log immutability enforced at the storage layer, not in application code): if a UI convention were the only thing generating audit events, calling the API directly would bypass the entire wedge. Placing this responsibility in the same layer that already owns data access, rather than the layer that owns presentation, is what makes M1 and A1 actual guarantees instead of best-effort logging.

---

## What's Deliberately Not Modeled Here

- Internal service implementation, deployment topology, or scaling behavior — that's `04-architecture/` (container-diagrams.md, deployment-model.md, scaling.md) and each service's own `05-services/*.md`.
- The exact wire protocol between browser-sdk and ingestion — that's `07-api/`.
- Sequencing of a specific request/response flow (e.g., what happens step-by-step when a DSAR query runs) — that's `04-architecture/sequence-diagrams.md`.

---

Sources: [ClickStack: High-Performance Open Source Observability — ClickHouse](https://clickhouse.com/clickstack), [What is observability in 2026? — ClickHouse](https://clickhouse.com/resources/engineering/what-is-observability), [Context Mapping in Domain-Driven Design — Software Patterns Lexicon](https://softwarepatternslexicon.com/java/domain-driven-design-ddd-patterns/strategic-patterns/context-mapping/), [Relationships Between Bounded Contexts in DDD](https://medium.com/@iamprovidence/relationships-between-bounded-contexts-in-ddd-ce5cfe3aaa04).

## What This Feeds Next

`docs/02-domain/business-rules.md` should expand [domain-model.md](domain-model.md)'s five invariants plus this document's AccessAuditEvent-ownership rule into full specifications. `docs/02-domain/event-catalog.md` can enumerate the concrete event types already named here (SessionEvent subtypes, AccessAuditEvent actions). `docs/04-architecture/container-diagrams.md` and `system-context.md` are the direct next step once the domain layer is settled — they turn this context map into an actual architecture diagram.
