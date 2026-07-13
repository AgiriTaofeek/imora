# Domain

## Domain Model

> Status: Research-based, current as of July 2026. Turns the parity and wedge requirements in [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd) and [User Stories](../01-product/README.md#user-stories) into actual entities, relationships, and invariants. This is deliberately at the DDD entity/aggregate altitude, not a field-level schema — that's `05-data/`'s job once this is settled.

---

### Modeling Approach

Three of the wedge entities below (access-audit-trail, retention/legal-hold, evidence export) aren't modeled from scratch — they follow established patterns from adjacent domains, found by checking how systems that already solve this problem structure it, rather than inventing a shape:

- **The audit trail follows the event-sourcing pattern**, not a mutable "last accessed by" field. Every access is an immutable, appended record — capture at minimum a stable entity key, a monotonic sequence number, event type, actor and source, a timestamp, and a payload — with immutability enforced at the storage layer (insert-only permissions), not just in application code. This is the standard shape for anything claiming to be an audit trail, and it's what makes M1's "produce audit-control evidence" story actually satisfiable.
- **Session capture follows the rrweb pattern** (the library underlying most session-replay tools, including the OpenReplay/PostHog category Imora has to match): one full DOM snapshot at session start, followed by a stream of incremental events — DOM mutations, mouse movement, clicks, scrolls, form input, viewport changes — each timestamped and typed. Session replay is not a video; it's a reconstructable event stream, and modeling it as anything else would break parity with the category.
- **Legal hold follows records-management practice, not a new deletion state.** A hold doesn't move data anywhere or copy it — it's a directive that a scheduled retention-driven deletion job must check against before it's allowed to execute. If a record is held, the deletion is skipped and that skip is itself logged; when the hold lifts, normal retention resumes automatically. Modeling legal hold as a parallel "frozen" copy of the data, instead of an override check on the deletion job, would be over-engineering the exact problem records-management systems already solved simply.

One open question from [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd) gets resolved here as a direct consequence of this modeling: feature-roadmap.md's Milestone 2 exit criteria required that "an evidence export generated mid-incident remains valid even if a retention policy purges its underlying source data afterward." The domain answer is that **EvidenceExport is a frozen, self-contained copy at generation time, not a live set of references** — it has to be, or the event-sourcing/retention-deletion model above would silently invalidate exports whenever a purge job runs. This wasn't decided by policy; it falls out of taking the other two patterns seriously.

---

### Core Entities — Parity

These exist to satisfy the Parity requirements in [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd) — the baseline every credible alternative already has.

- **Session** (aggregate root) — one user's browsing session. Holds a session identifier, user/anonymous identifier, start/end timestamps, the `Release` it occurred under, an `environment` tag, and device/browser metadata.
- **SessionEvent** — a single entry in a Session's rrweb-style event stream: `{type, timestamp, data}`, where `type` is one of FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, or ViewportChange. A Session is its ordered sequence of SessionEvents plus the initial snapshot — this *is* what "session replay" means as a domain concept, not a separate video artifact.
- **ErrorEvent** — a captured exception, tied to the Session and Release it occurred in.
- **ErrorGroup** — the deduplicated root-cause an ErrorEvent belongs to, keyed by stack-trace/fingerprint. This is the entity that makes story C2 ("one alert per root cause, not one per affected user") a data-model fact rather than a UI trick — grouping happens at write time, not display time.
- **Release** — a deployed version identifier, scoped to an `environment` — the same version string typically reaches `staging` before `production`, as two separate Release rows, not one. Sessions, ErrorEvents, and PerformanceMetrics are all tagged with the Release active when they occurred, which is what makes regression attribution (story C1: "which release did this") a query against existing tags rather than a separate bisection process.
- **PerformanceMetric** — a Core Web Vitals measurement (LCP, INP, or CLS) tied to a Session and Release, captured at the percentile Google's own methodology evaluates (p75), per [Target Users](../00-overview/README.md#target-users).
- **TraceLink** — the shared session/trace identifier propagated to backend spans (story J1). Deliberately thin: Imora doesn't own backend tracing (that's a Non-Goal per [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)), it only owns the correlation key that lets a Session's replay jump to backend traces held elsewhere.

### Environment: a Tag, Not an Entity

Every comparator in [Competitive Analysis](../00-overview/README.md#competitive-analysis) lets a team separate `production` traffic from `staging`/`development` noise — this is parity, not wedge, and its absence from earlier drafts of this document was a real gap, not a deliberate omission. **`environment` is modeled as a free-text string tag set at SDK `init()` time (identical mechanism to `release`), not a managed entity with its own lifecycle** — there's no CRUD, no approval workflow, nothing to administer beyond what a team types into their SDK config. It's denormalized directly onto `Session` (and, per [Event Schema](../05-data/README.md#event-schema), onto every event row derived from it) purely so query-api and the ClickHouse column store can filter and group by it without a join, the same reasoning already applied to `release_id`.

**The one place this isn't "just a filter":** regression detection (story C1's `RegressionDetected`) compares a release against its immediate predecessor. That comparison is only meaningful *within the same environment* — comparing `production`'s baseline against a `staging` deploy of the same version would produce a nonsense signal. `Release.priorReleaseId` therefore resolves to the prior release **in that same environment**, not the globally-most-recent row.

**Compliance rigor does not vary by environment.** A masked field, an access-audit-trail entry, and a retention clock apply identically whether the tagged environment is `production` or `development` — deliberately, because non-production environments routinely get seeded with copied production data in practice, and "staging is safe to relax" is a common, dangerous assumption this design refuses to build in. `environment` changes what a query *returns*, never what the system *enforces*.

---

### Core Entities — Wedge

These exist to satisfy the Wedge requirements — the reason a regulated buyer picks Imora over any Parity-only alternative.

- **AccessAuditEvent** — the append-only log entry from the event-sourcing pattern above: `{eventId, actorUserId, action, targetRecordType, targetRecordId, timestamp, sourceIp/device, sequenceNumber}`. `action` includes VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED (a retention deletion blocked by a LegalHold, per BR-2), and CONFIG_CHANGED (a RetentionPolicy, field-classification, or role change — added by [Audit Logging](../07-security/README.md#audit-logging) after NIST 800-53 AU-2 review showed configuration changes were otherwise unaudited). This single entity is what satisfies story M1 (HIPAA §164.312(b) audit-control evidence) and half of story A1 (DSAR "who has viewed this" queries) — both are reads over the same append-only log, not separate systems.
- **UnmaskRequest** — modeled as an AccessAuditEvent with `action = UNMASK` and a **required, non-empty `reason` field**. Not a separate entity family — per story M2, the requirement is that unmasking is audited and justified, and the event-sourcing pattern already gives us an auditable, timestamped record; adding a parallel entity would just be two places to keep an access log instead of one.
- **RetentionPolicy** — scoped per data category (SessionEvent, ErrorEvent, SecurityEvent, AccessAuditEvent itself), not per record: `{dataCategory, retentionPeriod, regulatoryBasis}`. This is the entity that satisfies story A2 and directly implements the regulatory-clock table from [Competitive Analysis](../00-overview/README.md#competitive-analysis) (PCI-DSS 12mo, HIPAA 6yr, GDPR purpose-bound, SOX 7yr) — each category can point at a different clock.
- **LegalHold** — `{holdId, appliedBy, reason, scope, appliedAt, liftedAt}`, where `scope` is a query/filter (e.g., "all Sessions tied to incident X"), not a fixed record list, matching how legal-hold systems actually define custodial scope. Per the modeling approach above, a scheduled deletion driven by RetentionPolicy must check for an active LegalHold covering the target before executing; a held record produces a "deletion skipped" AccessAuditEvent instead of being deleted. This is the entity that fulfills the "legal hold support" commitment in [Vision](../00-overview/README.md#vision)'s Guiding Principles, which had no domain definition until now.
- **EvidenceExport** — `{exportId, requestedBy, incidentReference, generatedAt, contentHash, format}`, holding a **frozen, self-contained copy** of the referenced Sessions, ErrorEvents, SecurityEvents, and AccessAuditEvents at generation time — not live references — per the resolution above. `contentHash` exists specifically so the export's integrity is independently verifiable, satisfying story J2's "immutable once generated" requirement.
- **SecurityEvent** — a security signal (e.g., an anomaly or WAF-style alert) optionally tied to a Session, existing specifically so it can be correlated into the same incident timeline as replay/errors/performance (story D2) rather than living in a separate tool.

---

### Relationships

| From | To | Cardinality | Note |
|---|---|---|---|
| Session | SessionEvent | 1 → * | The replay itself |
| Session | ErrorEvent | 1 → * | Errors that occurred during this session |
| ErrorEvent | ErrorGroup | * → 1 | Grouping is write-time, not display-time |
| Session | Release | * → 1 | Enables regression attribution |
| PerformanceMetric | Session, Release | * → 1 each | |
| Session | TraceLink | 1 → * | One session may correlate to many backend spans |
| AccessAuditEvent | Session \| EvidenceExport \| any targetRecordType | * → 1 (polymorphic) | Every read of sensitive data produces one of these |
| RetentionPolicy | data category (not a record) | 1 → * | Category-scoped, per [Competitive Analysis](../00-overview/README.md#competitive-analysis)'s finding that a single global TTL is the gap |
| LegalHold | Session, ErrorEvent, SecurityEvent, AccessAuditEvent (via scope query) | * → * | Scope is a query, not a fixed list |
| EvidenceExport | Session, ErrorEvent, SecurityEvent, AccessAuditEvent | * → * (frozen copies) | Not live references, by design |
| SecurityEvent | Session | * → 0..1 | Optional — not every security signal ties to a session |

---

### Key Invariants

These are the rules the system must never violate, regardless of which service ends up implementing them — flagged here so `README.md#business-rules` and `README.md#bounded-contexts` inherit them rather than re-deriving them:

1. **AccessAuditEvent is append-only.** No update or delete path may exist for this entity at the storage layer, not just in application code — per the research above, immutability enforced only in code is not a real guarantee.
2. **A RetentionPolicy-driven deletion job must check for an active LegalHold before executing.** A held record is skipped, and the skip itself is logged as an AccessAuditEvent — silence is not an acceptable outcome for a skipped deletion.
3. **EvidenceExport is frozen at generation time.** A later RetentionPolicy purge or LegalHold change must never alter or invalidate an already-generated export.
4. **PII/PHI masking happens at capture time, not query time, by default.** Per [Vision](../00-overview/README.md#vision)'s "Security by Default" principle, a SessionEvent's sensitive fields should be masked before they're written to storage unless explicitly allow-listed — a query-time-only masking design would mean the unmasked value already exists at rest, which fails the deny-by-default commitment even if every read path is correctly filtered.
5. **Every UNMASK action requires a non-empty reason and produces its own AccessAuditEvent** — per story M2, an escalation without a logged justification isn't an escalation, it's a bypass.

---

### What's Deliberately Not Modeled Here

- Field-level schemas (column types, indexes, partitioning) — that belongs in `05-data/README.md#clickhouse-schema`, `05-data/README.md#postgres-schema`, and `05-data/README.md#event-schema`.
- Service ownership of each entity (which microservice writes SessionEvent vs. AccessAuditEvent) — that's `README.md#bounded-contexts`, next.
- The specific mechanism for capture-time masking (regex rules, allow-list config format) — that's a `07-security/README.md#pii-redaction` concern, informed by invariant 4 above but not specified by it.

---

Sources: [How do you enforce immutability and append-only audit trails? — DesignGurus](https://www.designgurus.io/answers/detail/how-do-you-enforce-immutability-and-appendonly-audit-trails), [Event Sourcing Pattern — Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing), [rrweb GitHub](https://github.com/rrweb-io/rrweb), [How does session replay work: Observer — DEV Community](https://dev.to/yuyz0112/how-does-session-replay-work-part2-observer-4jmg), [Legal Hold 101: Data Retention and Destruction](https://www.daymarksi.com/information-technology-navigator-blog/legal-hold-101-data-retention-and-destruction), [Retention Policies vs Litigation Hold vs Archiving](https://www.syscloud.com/saas-data-protection-center/microsoft-365/exchange-online-retention-policy-litigation-hold-archiving/).

### What This Feeds Next

`research/02-domain/README.md#bounded-contexts` should assign each entity above to an owning service boundary, and `research/02-domain/README.md#business-rules` should expand the five invariants into full business-rule specifications with edge cases. `research/02-domain/README.md#event-catalog` can be derived largely from the SessionEvent and AccessAuditEvent type enumerations already defined here.

---

## Bounded Contexts

> Status: Research-based, current as of July 2026. Assigns each entity in [Domain Model](README.md#domain-model) to an owning service boundary. The eight service names below are not proposed here — they're already scaffolded in `research/04-services/` (gateway, ingestion, query-api, alert-engine, workers, browser-sdk, dashboard, notification-service), which is a real prior decision this document has to honor, not invent around.

---

### Modeling Approach

The context boundaries follow two established patterns rather than an arbitrary split:

- **Write-path / read-path separation**, the standard shape for high-ingest observability systems built on column stores like ClickHouse: ingestion is append-oriented and deliberately minimizes per-row processing, while query serving runs as an independently scaled concern. Conflating the two — having one service both accept high-volume writes and serve latency-sensitive reads — is the architectural mistake this split exists to avoid.
- **DDD context mapping**, specifically three relationship types used below: **Shared Kernel** (contexts that operate directly on the same entity definitions from [Domain Model](README.md#domain-model) and must stay in lockstep), **Customer-Supplier** (an upstream context whose downstream customer can influence its priorities, but the dependency direction is one-way), and **Conformist** (a downstream context that just accepts the upstream's model as-is, doing no translation of its own).

---

### The Eight Contexts

#### browser-sdk

**Responsibility:** client-side capture — the only place [Domain Model](README.md#domain-model)'s Invariant 4 (masking at capture time, not query time) can actually be enforced, since by the time data leaves the browser it must already be safe. Owns the rrweb-style event serialization (FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange) and Core Web Vitals measurement at the point of occurrence.

**Entities produced:** SessionEvent (pre-masked), PerformanceMetric (raw measurement, not yet attributed to a Release).

**Context relationship:** Shared Kernel with `ingestion` on the SessionEvent/PerformanceMetric shape — the wire format browser-sdk emits and the format ingestion accepts must be the same entity definition, not two shapes translated between, or masking guarantees applied in the SDK could be silently lost in translation.

---

#### gateway

**Responsibility:** authentication, request routing, and field-level access control enforcement — the chokepoint every read and write passes through before reaching a domain context. This is also where actor identity and source IP get attached to any subsequent AccessAuditEvent, since it's the only context that has verified who's asking.

**Entities produced:** none directly — gateway annotates requests with authenticated-actor context that `query-api` and `workers` consume when they generate AccessAuditEvent entries.

**Context relationship:** Customer-Supplier, upstream of both `ingestion` and `query-api` — both depend on gateway's authenticated identity, but gateway doesn't depend on either's internal model.

---

#### ingestion

**Responsibility:** the write path. Accepts SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, and TraceLink writes from browser-sdk (and backend instrumentation, for TraceLink) and persists them append-only, deliberately avoiding read-serving concerns — per the write/read separation pattern above.

**Entities owned:** Session, SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, TraceLink (all write-side).

**Context relationship:** Shared Kernel with `browser-sdk` (as above) and with `query-api` (both operate on the same Session/ErrorEvent/PerformanceMetric definitions from [Domain Model](README.md#domain-model) — a divergence here would mean query-api serving a different shape than ingestion wrote, which is exactly the kind of split a shared kernel is meant to prevent).

---

#### query-api

**Responsibility:** the read path — serving Session, replay, ErrorGroup, and PerformanceMetric queries to `dashboard`, and correlating TraceLink lookups (story J1). **This is also the only place AccessAuditEvent may be generated for read access**, and that's a load-bearing architectural decision, not a detail: per [Domain Model](README.md#domain-model)'s modeling approach, an audit trail enforced anywhere else — in `dashboard`, for instance — is bypassable by anyone calling the API directly, which would make the entire wedge (stories M1, A1) a UI convention instead of a guarantee. Every VIEW of a Session, replay, or masked field must produce an AccessAuditEvent as an inseparable part of serving that read, not an optional side effect a caller could skip.

**Entities owned:** read-side of Session/SessionEvent/ErrorGroup/PerformanceMetric; **owns AccessAuditEvent generation for all VIEW and UNMASK actions.**

**Context relationship:** Shared Kernel with `ingestion` (as above); Customer-Supplier downstream of `gateway`; upstream supplier to `dashboard`.

---

#### alert-engine

**Responsibility:** ErrorGroup grouping/deduplication (story C2 — grouping happens here, not in the UI, per [Domain Model](README.md#domain-model)'s note that grouping is write-time/processing-time, not display-time) and release-attributed regression detection against Core Web Vitals thresholds (story C1).

**Entities owned:** ErrorGroup, Release-regression evaluation logic. Reads ErrorEvent/PerformanceMetric/Release from the ingestion-owned store.

**Context relationship:** Customer-Supplier, downstream of `ingestion`'s written data; upstream supplier to `notification-service`.

---

#### workers

**Responsibility:** the background/async context — the only place two of the sharpest wedge invariants actually execute: RetentionPolicy-driven scheduled deletion (which must check LegalHold before every deletion, per Invariant 2) and EvidenceExport generation (which must freeze a self-contained copy at generation time, per Invariant 3 and the resolved open question in [Domain Model](README.md#domain-model)). Both are exactly the kind of correctness-critical, no-user-watching logic that belongs in a dedicated worker context rather than inline in a request path.

**Entities owned:** RetentionPolicy execution, LegalHold evaluation, EvidenceExport generation.

**Context relationship:** Customer-Supplier, downstream consumer of `ingestion`'s stored data and `gateway`'s actor context (an EvidenceExport request or a LegalHold application still needs to know who requested it, for the AccessAuditEvent it produces).

---

#### notification-service

**Responsibility:** delivery only — translating an alert-engine-produced signal into an email, Slack message, or webhook. Deliberately separated from `alert-engine` so grouping/regression logic isn't coupled to delivery mechanism; a new notification channel shouldn't require touching alerting logic at all.

**Entities owned:** none from [Domain Model](README.md#domain-model) — this context's data (delivery channels, message templates) is orthogonal to the core domain.

**Context relationship:** Conformist, downstream of `alert-engine` — it accepts whatever shape alert-engine produces without attempting to reinterpret or enrich it.

---

#### dashboard

**Responsibility:** presentation only. Renders what `query-api` and `alert-engine` (via `notification-service` or directly) return. Has no domain entities of its own and, critically, **must not be where any invariant is enforced** — masking, audit logging, and access control all have to hold even if a caller bypasses the dashboard UI entirely and calls query-api or gateway directly.

**Entities owned:** none.

**Context relationship:** Conformist, downstream of `query-api` — takes the served shape as-is. This is a deliberate choice, not a limitation: giving dashboard any translation authority over sensitive fields would create a second place PII-masking logic could drift from `browser-sdk`'s capture-time enforcement.

**A note specific to `dashboard` running [TanStack Start](https://tanstack.com/start), not a static SPA:** server-side rendering and server functions mean `dashboard` now executes code server-side, which makes it *technically capable* of holding its own database credentials — something a build-time-only static SPA never could have done. The Conformist relationship above is unchanged by this: a TanStack Start server function is still just an HTTP client calling `query-api`, executing on a different side of the network than a browser-side `fetch` would, never a second data-access path. Full detail: [Container Diagrams](../03-architecture/diagrams.md#container-diagrams).

---

### Context Map Summary

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

### The One Rule This Document Adds to README.md#domain-model

**AccessAuditEvent generation belongs exclusively to `query-api` (for reads) and `workers` (for retention/legal-hold/export actions) — never to `dashboard` or any presentation layer.** This follows directly from [Domain Model](README.md#domain-model)'s Invariant 1 (audit log immutability enforced at the storage layer, not in application code): if a UI convention were the only thing generating audit events, calling the API directly would bypass the entire wedge. Placing this responsibility in the same layer that already owns data access, rather than the layer that owns presentation, is what makes M1 and A1 actual guarantees instead of best-effort logging.

---

### What's Deliberately Not Modeled Here

- Internal service implementation, deployment topology, or scaling behavior — that's `03-architecture/` (diagrams.md#container-diagrams, README.md#deployment-model, README.md#scaling) and each service's own `04-services/*.md`.
- The exact wire protocol between browser-sdk and ingestion — that's `06-api/`.
- Sequencing of a specific request/response flow (e.g., what happens step-by-step when a DSAR query runs) — that's `03-architecture/diagrams.md#sequence-diagrams`.

---

Sources: [ClickStack: High-Performance Open Source Observability — ClickHouse](https://clickhouse.com/clickstack), [What is observability in 2026? — ClickHouse](https://clickhouse.com/resources/engineering/what-is-observability), [Context Mapping in Domain-Driven Design — Software Patterns Lexicon](https://softwarepatternslexicon.com/java/domain-driven-design-ddd-patterns/strategic-patterns/context-mapping/), [Relationships Between Bounded Contexts in DDD](https://medium.com/@iamprovidence/relationships-between-bounded-contexts-in-ddd-ce5cfe3aaa04).

### What This Feeds Next

`research/02-domain/README.md#business-rules` should expand [Domain Model](README.md#domain-model)'s five invariants plus this document's AccessAuditEvent-ownership rule into full specifications. `research/02-domain/README.md#event-catalog` can enumerate the concrete event types already named here (SessionEvent subtypes, AccessAuditEvent actions). `research/03-architecture/diagrams.md#container-diagrams` and `README.md#system-context` are the direct next step once the domain layer is settled — they turn this context map into an actual architecture diagram.

---

## Business Rules

> Status: Research-based, current as of July 2026. Expands [Domain Model](README.md#domain-model)'s five invariants and [Bounded Contexts](README.md#bounded-contexts)'s AccessAuditEvent-ownership rule into full specifications, including the edge cases a regulated deployment will actually hit. Each rule states what happens, why, and which entity/context from the prior two documents enforces it.

---

### Rule Set A — Retention and Deletion

#### BR-1: Retention period is assigned per data category, and the longest applicable period wins when categories overlap

A Session, ErrorEvent, or SecurityEvent may fall under more than one regulation simultaneously — e.g., a healthcare organization's session data is subject to both HIPAA (6-year floor) and, if the patient is an EU resident, GDPR. When retention requirements conflict, the strictest (longest) applicable period governs, and the specific regulatory basis for that choice is recorded on the RetentionPolicy itself — not left implicit. This is standard practice across regulated data categories: where multiple regimes apply, one policy set to the strictest requirement satisfies all of them simultaneously, rather than maintaining parallel conflicting policies per regulation.

**Enforced by:** `workers` (RetentionPolicy execution), per [Bounded Contexts](README.md#bounded-contexts).

#### BR-2: A scheduled deletion must check for an active LegalHold immediately before executing, not at scheduling time

Restates [Domain Model](README.md#domain-model) Invariant 2 with the specific failure mode it prevents: checking for a hold only when a deletion job is *scheduled*, rather than immediately before it *executes*, creates a race condition — a hold applied after scheduling but before execution would be silently missed. The check-before-destroy step must be the last thing that happens before a delete, not a precondition evaluated earlier in the pipeline. When the check finds an active hold, the deletion is skipped and the skip is logged as an AccessAuditEvent; when a hold is later lifted, the record re-enters the normal retention schedule on its next scheduled evaluation — lifting a hold does not trigger an immediate retroactive deletion.

**Enforced by:** `workers`, immediately preceding every deletion execution.

#### BR-3: A GDPR erasure request is honored except where a specific legal obligation or active legal hold requires retention — and even then, only what's strictly necessary is retained

This is the sharpest conflict Adaeze (the DPO persona) will actually hit: a data subject's erasure request against a session that Marcus's hospital is independently required to retain for HIPAA's 6-year floor. GDPR Article 17(3) provides exactly this exception — erasure isn't required where retention is necessary for compliance with a legal obligation (17(3)(b)) or the establishment, exercise, or defense of legal claims (17(3)(e)). **But the exception is not a blanket refusal**: where it's possible to satisfy the legal obligation while still deleting everything not strictly necessary for it, the exception only covers the narrower set. The correct behavior is **selective, field-level purging** — anonymize or delete the fields GDPR requires erased, while preserving only the minimum structure the overriding legal obligation actually requires (e.g., a HIPAA-required audit trail entry can often survive with the subject's identifying fields anonymized, rather than the whole record being retained or the whole record being deleted).

Every partial refusal must be logged with the specific regulatory basis cited (Article 17(3)(b) vs (e), or the specific HIPAA/SOX/PCI-DSS clause), so Adaeze can produce that justification to a regulator or the data subject directly — an unexplained "cannot delete" is not sufficient.

**Enforced by:** `workers` (executes the selective purge) and `query-api`/AccessAuditEvent (the refusal-with-basis is itself logged).

#### BR-4: EvidenceExport is immune to both BR-1 and BR-3 once generated

Restates [Domain Model](README.md#domain-model)'s resolution: an EvidenceExport is a frozen, self-contained copy at generation time. Neither a later retention purge (BR-1) nor a subsequent erasure request (BR-3) may alter or invalidate an already-generated export — the export's `contentHash` exists specifically so this immutability is independently verifiable, not just asserted.

**Enforced by:** `workers` (export generation), verified by `contentHash`.

---

### Rule Set B — Access, Audit, and Masking

#### BR-5: Every VIEW, EXPORT, UNMASK, and DELETE against a sensitive record produces exactly one AccessAuditEvent

Restates [Domain Model](README.md#domain-model) Invariant 1 with the ownership rule from [Bounded Contexts](README.md#bounded-contexts): this happens in `query-api` (VIEW, EXPORT, UNMASK) and `workers` (DELETE, including the BR-2 skip case), never in `dashboard`. A read that completes without producing an AccessAuditEvent is a defect, not an acceptable fast path — including for internal/admin tooling, which is a common gap where audit trails silently don't apply.

#### BR-6: UNMASK requires a non-empty, human-readable reason, and the reason is part of the audit record, not metadata about it

Restates [Domain Model](README.md#domain-model) Invariant 5. The reason field exists so a HIPAA risk assessment or DSAR response can show *why* PHI was unmasked, not just that it was — "debugging" is an acceptable reason; a blank field is not.

#### BR-7: PII/PHI masking is evaluated at capture time in browser-sdk; a field with no matching allow-list rule is masked by default

Restates [Domain Model](README.md#domain-model) Invariant 4 with the failure mode it exists to prevent: a new form field shipped without an explicit masking rule must render masked, not unmasked-until-someone-notices. This is deny-by-default specifically because the alternative (block-list: mask known-sensitive fields, capture everything else) is the pattern that produced the PII-leak fact patterns behind Cost Driver 2 in [Problem Statement](../00-overview/README.md#problem-statement).

---

### Rule Set C — Conflict Precedence Summary

When more than one rule above could apply to the same action, precedence resolves in this order, most authoritative first:

1. **Active LegalHold** (BR-2) — blocks deletion outright, regardless of what any retention policy or erasure request says.
2. **Legal obligation requiring retention** (BR-3's exception) — permits refusing full erasure, but only for the minimum data the obligation actually requires.
3. **Longest applicable RetentionPolicy** (BR-1) — governs ordinary scheduled deletion where no hold or overriding legal obligation applies.
4. **Erasure/deletion request** (BR-3's default case) — honored in full once nothing above overrides it.

This ordering is itself a business rule, not just a reading convenience: a `workers` implementation that evaluates these checks out of order (e.g., applying BR-1's retention clock before checking BR-2's hold) would produce the exact race condition BR-2 exists to prevent.

---

Sources: [Art. 17 GDPR – Right to erasure](https://gdpr-info.eu/art-17-gdpr/), [What Is GDPR Article 17 and 4 Ways to Achieve Compliance — Exabeam](https://www.exabeam.com/explainers/gdpr-compliance/what-is-gdpr-article-17-right-to-erasure-and-4-ways-to-achieve-compliance/), [Compliance Log Retention Requirements by Regulation](https://claudiasop.com/blog/compliance-log-retention-requirements.html), [Medical Record Retention: State Mandates vs. Federal Law](https://www.complydome.com/compliance-resources/state-mandates-vs-federal-law-a-small-practice-guide-to-which-record-retention-rule-wins-cms-hipaa-state-laws), [Legal Hold 101: Data Retention and Destruction](https://www.daymarksi.com/information-technology-navigator-blog/legal-hold-101-data-retention-and-destruction), [Defensible Data Deletion After a Legal Hold — Onna](https://www.onna.com/resources/blog/defensible-data-deletion-after-a-legal-hold).

### What This Feeds Next

`research/02-domain/README.md#event-catalog` should enumerate the concrete AccessAuditEvent actions (VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED per BR-2) and SessionEvent subtypes already named across [Domain Model](README.md#domain-model) and this document. `research/07-security/README.md#pii-redaction` and `research/05-data/README.md#retention` should implement BR-1 through BR-7 directly rather than re-deriving them.

---

## Event Catalog

> Status: Research-based, current as of July 2026. Enumerates the concrete domain events implied by [Domain Model](README.md#domain-model), [Bounded Contexts](README.md#bounded-contexts), and [Business Rules](README.md#business-rules), named per the EventStorming convention: past tense, noun/verb, describing a fact that happened — not a command or a request. This is the last domain-layer document before `05-data/README.md#event-schema` and `06-api/README.md#webhooks` turn these into wire formats.

Each event lists: **Producer** (owning bounded context, per [Bounded Contexts](README.md#bounded-contexts)), **Trigger**, **Key payload**, **Consumers**.

---

### Capture and Ingestion Events (Parity)

#### SessionStarted
- **Producer:** browser-sdk
- **Trigger:** A new browsing session begins.
- **Key payload:** sessionId, userId/anonymousId, initial Release tag, `environment` tag, device/browser metadata.
- **Consumers:** ingestion (persists), alert-engine (session-count baselines).

#### SessionEventCaptured
- **Producer:** browser-sdk
- **Trigger:** Any rrweb-style incremental capture — DOM mutation, mouse move, click, scroll, form input, or viewport change, per [Domain Model](README.md#domain-model)'s SessionEvent entity. One event type with a `subtype` field, not seven separate event names, since they share identical routing and storage behavior.
- **Key payload:** sessionId, subtype, timestamp, masked event data (per BR-7 — masking already applied before this event exists).
- **Consumers:** ingestion (persists), query-api (later serves as replay).

#### SessionEnded
- **Producer:** browser-sdk
- **Trigger:** Session timeout or explicit page-unload.
- **Key payload:** sessionId, endedAt, total event count.
- **Consumers:** ingestion, workers (becomes eligible for retention-clock evaluation per BR-1 once ended).

#### ErrorEventCaptured
- **Producer:** browser-sdk (client errors) or backend instrumentation via TraceLink (correlated backend errors).
- **Trigger:** An unhandled exception or explicit error report.
- **Key payload:** sessionId, stack trace/fingerprint, Release tag, `environment` tag, timestamp.
- **Consumers:** ingestion (persists), alert-engine (grouping, below).

#### ErrorGrouped
- **Producer:** alert-engine
- **Trigger:** An ErrorEventCaptured event is fingerprint-matched to an existing ErrorGroup, or a new ErrorGroup is created if no match exists — the write-time grouping decision from [Domain Model](README.md#domain-model), satisfying story C2 ("one alert per root cause, not one per affected user").
- **Key payload:** errorGroupId, matched sessionId/errorEventId, isNewGroup flag.
- **Consumers:** notification-service (only notified once per group, not per occurrence).

#### PerformanceMetricRecorded
- **Producer:** browser-sdk
- **Trigger:** An LCP, INP, or CLS measurement completes for a page view.
- **Key payload:** sessionId, metric type, value, Release tag, `environment` tag.
- **Consumers:** ingestion, alert-engine (regression evaluation, below).

#### RegressionDetected
- **Producer:** alert-engine
- **Trigger:** A statistically significant change in a Core Web Vitals metric or error rate is attributed to a specific Release — story C1's "which release did this," evaluated at p75 per [Target Users](../00-overview/README.md#target-users). The baseline/new-value comparison is always **within one `environment`** — per [Domain Model](README.md#domain-model)'s Environment note, comparing across environments would produce a meaningless signal.
- **Key payload:** metric or errorGroupId, `environment`, previous Release baseline, new Release value, releaseId.
- **Consumers:** notification-service.

#### ReleaseDeployed
- **Producer:** ingestion (via deploy-hook or CI integration — not yet specified further; a `06-api/` concern).
- **Trigger:** A new Release identifier is registered **for a specific environment** — the same version string reaching `staging` then `production` fires this event twice, once per environment, not once total.
- **Key payload:** releaseId, `environment`, deployedAt, prior releaseId (for regression baseline comparison, scoped to the same environment).
- **Consumers:** alert-engine.

#### SecuritySignalReceived
- **Producer:** ingestion
- **Trigger:** A SecurityEvent (anomaly, WAF-style signal) arrives, optionally tied to a sessionId — satisfying story D2's correlation requirement.
- **Key payload:** sessionId (optional), signal type, severity, timestamp.
- **Consumers:** query-api (incident timeline correlation), alert-engine.

#### TraceLinked
- **Producer:** ingestion
- **Trigger:** A backend span arrives carrying a propagated session/trace identifier, per story J1's correlation mechanism.
- **Key payload:** sessionId, backend traceId/spanId.
- **Consumers:** query-api (replay-to-trace navigation).

---

### Access and Audit Events (Wedge)

Every event in this section is an AccessAuditEvent variant per [Domain Model](README.md#domain-model) — append-only, produced exclusively by `query-api` or `workers` per [Bounded Contexts](README.md#bounded-contexts)'s ownership rule, never by `dashboard`.

#### SessionViewed
- **Producer:** query-api
- **Trigger:** Any read of a Session's replay or metadata by an authenticated actor.
- **Key payload:** actorUserId, sessionId, timestamp, source IP/device (from gateway's actor context).
- **Consumers:** the audit log itself; surfaces in M1/A1 query responses.

#### FieldUnmasked
- **Producer:** query-api
- **Trigger:** An UNMASK action against a masked field, per BR-6.
- **Key payload:** actorUserId, sessionId, field identifier, **reason (required, non-empty)**.
- **Consumers:** audit log; HIPAA risk-assessment reports (M1).

#### RecordExported
- **Producer:** query-api (ad hoc export) or workers (EvidenceExport generation).
- **Trigger:** Any EXPORT action, including a full EvidenceExport per story J2.
- **Key payload:** actorUserId, exported record set, exportId (if EvidenceExport), contentHash.
- **Consumers:** audit log.

#### RecordDeleted
- **Producer:** workers
- **Trigger:** A scheduled deletion executes under BR-1/BR-2 (no active hold found).
- **Key payload:** targetRecordType, targetRecordId, regulatoryBasis (which RetentionPolicy clock triggered it).
- **Consumers:** audit log.

#### DeletionSkippedDueToHold
- **Producer:** workers
- **Trigger:** BR-2's check-before-destroy finds an active LegalHold covering the target record.
- **Key payload:** targetRecordId, holdId, timestamp.
- **Consumers:** audit log — this is the event that makes a skipped deletion visible rather than silent, per BR-2.

#### ConfigurationChanged
- **Producer:** query-api or workers, wherever the change is made.
- **Trigger:** A RetentionPolicy edit, a field-classification change (per [PII Redaction](../07-security/README.md#pii-redaction)), or a role/permission grant (per [Authorization](../07-security/README.md#authorization)) — the "security or privacy attribute changes" NIST 800-53 AU-2 requires logging, distinct from data access itself. Identified as a gap and closed in [Audit Logging](../07-security/README.md#audit-logging).
- **Key payload:** actorUserId, targetRecordType (RetentionPolicy | FieldClassification | UserRole), targetRecordId, oldValue, newValue, timestamp.
- **Consumers:** audit log — without this, BR-1's retention guarantee depends on trusting an unaudited configuration, which defeats the point.

#### ErasureRequestReceived / ErasureRequestResolved
- **Producer:** workers (resolution), triggered by a request logged wherever DSAR intake happens (a `06-api/` or future workflow concern, not specified here).
- **Trigger:** A data-subject erasure request enters BR-3's precedence evaluation.
- **Key payload:** ErasureRequestResolved specifically carries: outcome (full erasure / selective purge / denied), regulatory basis cited if not fully honored (Article 17(3)(b) or (e), or the specific overriding statute), fields actually purged vs. retained.
- **Consumers:** audit log; this is Adaeze's evidence that a partial refusal was justified, not arbitrary.

---

### Retention and Compliance Events (Wedge)

#### LegalHoldApplied / LegalHoldLifted
- **Producer:** workers, triggered by an authorized actor's request (via gateway's actor context).
- **Trigger:** A hold is placed on or removed from a scope query (per [Domain Model](README.md#domain-model)'s LegalHold entity).
- **Key payload:** holdId, appliedBy/liftedBy, scope query, reason.
- **Consumers:** workers itself (every subsequent BR-2 check evaluates against currently-applied holds); audit log.

#### RetentionPolicyEvaluated
- **Producer:** workers
- **Trigger:** A scheduled sweep evaluates records against BR-1's regulatory clocks — the step immediately preceding a RecordDeleted or DeletionSkippedDueToHold outcome.
- **Key payload:** dataCategory, evaluatedAt, record count evaluated, outcome counts.
- **Consumers:** audit log (operational, not per-record).

#### EvidenceExportGenerated
- **Producer:** workers
- **Trigger:** Story J2's one-click export completes.
- **Key payload:** exportId, incidentReference, frozen record set, contentHash, generatedAt — per BR-4, this event's payload is the permanent, immutable record of what the export contained.
- **Consumers:** RecordExported (audit log entry), the requesting actor.

---

### Notification Events

#### AlertTriggered
- **Producer:** alert-engine
- **Trigger:** ErrorGrouped (new group or threshold crossed) or RegressionDetected.
- **Key payload:** alert type, source event reference, severity.
- **Consumers:** notification-service.

#### NotificationSent
- **Producer:** notification-service
- **Trigger:** AlertTriggered is translated into a delivery (email, Slack, webhook) per its Conformist relationship to alert-engine in [Bounded Contexts](README.md#bounded-contexts).
- **Key payload:** channel, delivery status, sourceAlertId.
- **Consumers:** none within the domain — this is the terminal event in the chain.

---

Sources: [EventStorming Glossary & Cheat Sheet — DDD Crew](https://ddd-crew.github.io/eventstorming-glossary-cheat-sheet/), [Domain events: Design and implementation — Microsoft Learn](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/microservice-ddd-cqrs-patterns/domain-events-design-implementation).

### What This Feeds Next

`research/05-data/README.md#event-schema` should turn each event above into a concrete field-level schema. `research/06-api/README.md#webhooks` should decide which of these (likely AlertTriggered, RegressionDetected, EvidenceExportGenerated) are exposed externally. The terms this document and its predecessors have been using consistently (Session, AccessAuditEvent, LegalHold, etc.) are formally defined in one place in [Glossary](../00-overview/README.md#glossary).

