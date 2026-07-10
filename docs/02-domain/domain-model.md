# Domain Model

> Status: Research-based, current as of July 2026. Turns the parity and wedge requirements in [prd.md](../01-product/prd.md) and [user-stories.md](../01-product/user-stories.md) into actual entities, relationships, and invariants. This is deliberately at the DDD entity/aggregate altitude, not a field-level schema — that's `06-data/`'s job once this is settled.

---

## Modeling Approach

Three of the wedge entities below (access-audit-trail, retention/legal-hold, evidence export) aren't modeled from scratch — they follow established patterns from adjacent domains, found by checking how systems that already solve this problem structure it, rather than inventing a shape:

- **The audit trail follows the event-sourcing pattern**, not a mutable "last accessed by" field. Every access is an immutable, appended record — capture at minimum a stable entity key, a monotonic sequence number, event type, actor and source, a timestamp, and a payload — with immutability enforced at the storage layer (insert-only permissions), not just in application code. This is the standard shape for anything claiming to be an audit trail, and it's what makes M1's "produce audit-control evidence" story actually satisfiable.
- **Session capture follows the rrweb pattern** (the library underlying most session-replay tools, including the OpenReplay/PostHog category Imora has to match): one full DOM snapshot at session start, followed by a stream of incremental events — DOM mutations, mouse movement, clicks, scrolls, form input, viewport changes — each timestamped and typed. Session replay is not a video; it's a reconstructable event stream, and modeling it as anything else would break parity with the category.
- **Legal hold follows records-management practice, not a new deletion state.** A hold doesn't move data anywhere or copy it — it's a directive that a scheduled retention-driven deletion job must check against before it's allowed to execute. If a record is held, the deletion is skipped and that skip is itself logged; when the hold lifts, normal retention resumes automatically. Modeling legal hold as a parallel "frozen" copy of the data, instead of an override check on the deletion job, would be over-engineering the exact problem records-management systems already solved simply.

One open question from [prd.md](../01-product/prd.md) gets resolved here as a direct consequence of this modeling: feature-roadmap.md's Milestone 2 exit criteria required that "an evidence export generated mid-incident remains valid even if a retention policy purges its underlying source data afterward." The domain answer is that **EvidenceExport is a frozen, self-contained copy at generation time, not a live set of references** — it has to be, or the event-sourcing/retention-deletion model above would silently invalidate exports whenever a purge job runs. This wasn't decided by policy; it falls out of taking the other two patterns seriously.

---

## Core Entities — Parity

These exist to satisfy the Parity requirements in [prd.md](../01-product/prd.md) — the baseline every credible alternative already has.

- **Session** (aggregate root) — one user's browsing session. Holds a session identifier, user/anonymous identifier, start/end timestamps, the `Release` it occurred under, and device/browser metadata.
- **SessionEvent** — a single entry in a Session's rrweb-style event stream: `{type, timestamp, data}`, where `type` is one of FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, or ViewportChange. A Session is its ordered sequence of SessionEvents plus the initial snapshot — this *is* what "session replay" means as a domain concept, not a separate video artifact.
- **ErrorEvent** — a captured exception, tied to the Session and Release it occurred in.
- **ErrorGroup** — the deduplicated root-cause an ErrorEvent belongs to, keyed by stack-trace/fingerprint. This is the entity that makes story C2 ("one alert per root cause, not one per affected user") a data-model fact rather than a UI trick — grouping happens at write time, not display time.
- **Release** — a deployed version identifier. Sessions, ErrorEvents, and PerformanceMetrics are all tagged with the Release active when they occurred, which is what makes regression attribution (story C1: "which release did this") a query against existing tags rather than a separate bisection process.
- **PerformanceMetric** — a Core Web Vitals measurement (LCP, INP, or CLS) tied to a Session and Release, captured at the percentile Google's own methodology evaluates (p75), per [target-users.md](../00-overview/target-users.md).
- **TraceLink** — the shared session/trace identifier propagated to backend spans (story J1). Deliberately thin: Imora doesn't own backend tracing (that's a Non-Goal per [prd.md](../01-product/prd.md)), it only owns the correlation key that lets a Session's replay jump to backend traces held elsewhere.

---

## Core Entities — Wedge

These exist to satisfy the Wedge requirements — the reason a regulated buyer picks Imora over any Parity-only alternative.

- **AccessAuditEvent** — the append-only log entry from the event-sourcing pattern above: `{eventId, actorUserId, action, targetRecordType, targetRecordId, timestamp, sourceIp/device, sequenceNumber}`. `action` includes VIEW, EXPORT, UNMASK, and DELETE. This single entity is what satisfies story M1 (HIPAA §164.312(b) audit-control evidence) and half of story A1 (DSAR "who has viewed this" queries) — both are reads over the same append-only log, not separate systems.
- **UnmaskRequest** — modeled as an AccessAuditEvent with `action = UNMASK` and a **required, non-empty `reason` field**. Not a separate entity family — per story M2, the requirement is that unmasking is audited and justified, and the event-sourcing pattern already gives us an auditable, timestamped record; adding a parallel entity would just be two places to keep an access log instead of one.
- **RetentionPolicy** — scoped per data category (SessionEvent, ErrorEvent, SecurityEvent, AccessAuditEvent itself), not per record: `{dataCategory, retentionPeriod, regulatoryBasis}`. This is the entity that satisfies story A2 and directly implements the regulatory-clock table from [competitive-analysis.md](../00-overview/competitive-analysis.md) (PCI-DSS 12mo, HIPAA 6yr, GDPR purpose-bound, SOX 7yr) — each category can point at a different clock.
- **LegalHold** — `{holdId, appliedBy, reason, scope, appliedAt, liftedAt}`, where `scope` is a query/filter (e.g., "all Sessions tied to incident X"), not a fixed record list, matching how legal-hold systems actually define custodial scope. Per the modeling approach above, a scheduled deletion driven by RetentionPolicy must check for an active LegalHold covering the target before executing; a held record produces a "deletion skipped" AccessAuditEvent instead of being deleted. This is the entity that fulfills the "legal hold support" commitment in [vision.md](../00-overview/vision.md)'s Guiding Principles, which had no domain definition until now.
- **EvidenceExport** — `{exportId, requestedBy, incidentReference, generatedAt, contentHash, format}`, holding a **frozen, self-contained copy** of the referenced Sessions, ErrorEvents, SecurityEvents, and AccessAuditEvents at generation time — not live references — per the resolution above. `contentHash` exists specifically so the export's integrity is independently verifiable, satisfying story J2's "immutable once generated" requirement.
- **SecurityEvent** — a security signal (e.g., an anomaly or WAF-style alert) optionally tied to a Session, existing specifically so it can be correlated into the same incident timeline as replay/errors/performance (story D2) rather than living in a separate tool.

---

## Relationships

| From | To | Cardinality | Note |
|---|---|---|---|
| Session | SessionEvent | 1 → * | The replay itself |
| Session | ErrorEvent | 1 → * | Errors that occurred during this session |
| ErrorEvent | ErrorGroup | * → 1 | Grouping is write-time, not display-time |
| Session | Release | * → 1 | Enables regression attribution |
| PerformanceMetric | Session, Release | * → 1 each | |
| Session | TraceLink | 1 → * | One session may correlate to many backend spans |
| AccessAuditEvent | Session \| EvidenceExport \| any targetRecordType | * → 1 (polymorphic) | Every read of sensitive data produces one of these |
| RetentionPolicy | data category (not a record) | 1 → * | Category-scoped, per [competitive-analysis.md](../00-overview/competitive-analysis.md)'s finding that a single global TTL is the gap |
| LegalHold | Session, ErrorEvent, SecurityEvent, AccessAuditEvent (via scope query) | * → * | Scope is a query, not a fixed list |
| EvidenceExport | Session, ErrorEvent, SecurityEvent, AccessAuditEvent | * → * (frozen copies) | Not live references, by design |
| SecurityEvent | Session | * → 0..1 | Optional — not every security signal ties to a session |

---

## Key Invariants

These are the rules the system must never violate, regardless of which service ends up implementing them — flagged here so `business-rules.md` and `bounded-contexts.md` inherit them rather than re-deriving them:

1. **AccessAuditEvent is append-only.** No update or delete path may exist for this entity at the storage layer, not just in application code — per the research above, immutability enforced only in code is not a real guarantee.
2. **A RetentionPolicy-driven deletion job must check for an active LegalHold before executing.** A held record is skipped, and the skip itself is logged as an AccessAuditEvent — silence is not an acceptable outcome for a skipped deletion.
3. **EvidenceExport is frozen at generation time.** A later RetentionPolicy purge or LegalHold change must never alter or invalidate an already-generated export.
4. **PII/PHI masking happens at capture time, not query time, by default.** Per [vision.md](../00-overview/vision.md)'s "Security by Default" principle, a SessionEvent's sensitive fields should be masked before they're written to storage unless explicitly allow-listed — a query-time-only masking design would mean the unmasked value already exists at rest, which fails the deny-by-default commitment even if every read path is correctly filtered.
5. **Every UNMASK action requires a non-empty reason and produces its own AccessAuditEvent** — per story M2, an escalation without a logged justification isn't an escalation, it's a bypass.

---

## What's Deliberately Not Modeled Here

- Field-level schemas (column types, indexes, partitioning) — that belongs in `06-data/clickhouse-schema.md`, `06-data/postgres-schema.md`, and `06-data/event-schema.md`.
- Service ownership of each entity (which microservice writes SessionEvent vs. AccessAuditEvent) — that's `bounded-contexts.md`, next.
- The specific mechanism for capture-time masking (regex rules, allow-list config format) — that's a `08-security/pii-redaction.md` concern, informed by invariant 4 above but not specified by it.

---

Sources: [How do you enforce immutability and append-only audit trails? — DesignGurus](https://www.designgurus.io/answers/detail/how-do-you-enforce-immutability-and-appendonly-audit-trails), [Event Sourcing Pattern — Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing), [rrweb GitHub](https://github.com/rrweb-io/rrweb), [How does session replay work: Observer — DEV Community](https://dev.to/yuyz0112/how-does-session-replay-work-part2-observer-4jmg), [Legal Hold 101: Data Retention and Destruction](https://www.daymarksi.com/information-technology-navigator-blog/legal-hold-101-data-retention-and-destruction), [Retention Policies vs Litigation Hold vs Archiving](https://www.syscloud.com/saas-data-protection-center/microsoft-365/exchange-online-retention-policy-litigation-hold-archiving/).

## What This Feeds Next

`docs/02-domain/bounded-contexts.md` should assign each entity above to an owning service boundary, and `docs/02-domain/business-rules.md` should expand the five invariants into full business-rule specifications with edge cases. `docs/02-domain/event-catalog.md` can be derived largely from the SessionEvent and AccessAuditEvent type enumerations already defined here.
