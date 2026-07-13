# Imora — Product Requirements Document

> This is the build-ready PRD — everything a designer or engineer needs to start work without reading the full research tree. Every claim here is backed by sourced research in [`research/`](../research/README.md); this file states conclusions, `research/` states the evidence and reasoning behind them. If a number or claim here looks surprising, the "why" is one click away via the links.

---

## 1. Product Summary

**Imora** is a self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory, built for regulated industries (banks, insurers, hospitals, government) that cannot send customer session data to a third party.

It is not a new product category. To be adopted at all, it has to match what engineers already get from Sentry/Datadog/LogRocket/FullStory/OpenReplay/PostHog — error tracking, session replay, performance monitoring, self-hosted deployment. That's **Parity**. It wins the deal specifically on three things none of those tools do — an audit trail of who on the team viewed a customer's session, retention mapped to actual regulatory clocks instead of one global setting, and one-click evidence export built for an auditor. That's the **Wedge**.

Full positioning and market research: [`research/00-overview/README.md`](../research/00-overview/README.md).

---

## 2. The Problem, Condensed

Regulated organizations get deep observability *or* data ownership, never both, today:

| Cost driver | The number | Source |
|---|---|---|
| Regulated industries pay the highest breach costs in the economy | Healthcare: $7.42M avg breach cost, 279 days to detect. Financial services: $5.56M avg. | IBM Cost of a Data Breach Report 2025 |
| Third-party session replay is an active, current litigation risk (independent of any breach) | ~1,500 CIPA wiretapping lawsuits filed in 18 months; $5,000 statutory damages per violation, arguable per-session | See [`research/00-overview/README.md`](../research/00-overview/README.md#problem-statement) |
| Tool fragmentation has a measurable incident-response tax | 20–40% added incident-resolution time; $100K–$400K/year in tool sprawl for a typical regulated org | See [`research/00-overview/README.md`](../research/00-overview/README.md#problem-statement) |

**✓ Verified 2026-07-12.** The four load-bearing citations were independently re-checked against live sources:
- **Sutter Health $21.5M settlement** — confirmed real (third-party tracking tech — Google Analytics, Meta pixel — on the MyHealthOnline patient portal; ~1.6M-member class). The underlying mechanism is pixel/analytics tracking, not session-replay software specifically — a nuance worth keeping straight if this case is cited as a session-replay precedent rather than a broader third-party-tracking one.
- **Mikulsky v. Bloomingdale's** — confirmed. Ninth Circuit reversed the district court's dismissal on June 20, 2025, expanding CIPA §631(a) to cover session-replay-reconstructed website interactions; Bloomingdale's filed a settlement notice in November 2025.
- **Torres v. Prudential Financial** — confirmed. Summary judgment for the defense granted April 17, 2025 (N.D. Cal.), on the "not readable until reassembled, therefore not intercepted in transit" theory — exactly as characterized in `research/00-overview/README.md`.
- **IBM Cost of a Data Breach Report 2025** — confirmed. Healthcare $7.42M average (down from $9.77M the prior year), 279-day mean time to detect; US national average $10.22M (up 9.2%); global average $4.44M (first decline in 5 years). One small correction worth making if cited again: healthcare has been the costliest industry for **14** consecutive years per this year's report, not 15.
- **Not independently re-verified:** the specific "~1,500 CIPA lawsuits in 18 months" count — the broader litigation-wave trend is corroborated by multiple 2025–2026 law-firm analyses, but that exact figure wasn't traced to a primary source. Treat it as directionally right, not a number to repeat verbatim in front of a legal audience without its own citation.

Full problem research with sources: [`research/00-overview/README.md#problem-statement`](../research/00-overview/README.md#problem-statement).

---

## 3. Goals

1. Be a credible, drop-in replacement for the observability tooling a regulated org's engineers already use daily. **(Parity)**
2. Be the only alternative — self-hosted or SaaS — that a CISO, DPO, or HIPAA Security Officer can approve without a compensating control bolted on afterward. **(Wedge)**
3. Ship parity and wedge together, not sequentially. A compliance-only v1 has nothing for an engineer to open daily; an observability-only v1 gives a regulated buyer no reason to pick Imora over a plain self-hosted alternative.

## Non-Goals

- **Not a general backend APM/infrastructure observability tool.** Imora correlates with backend traces; it doesn't replace a backend APM.
- **Not a full GRC/policy-management suite.** The wedge is audit trail, retention, and evidence export for the frontend telemetry Imora itself holds — not a general compliance-workflow product.
- **Not a SIEM.** Security monitoring means correlating security signal into Imora's own incident timeline, not ingesting arbitrary log sources.
- **Not mobile-first.** Web frontend only for the scope covered here.
- **Not competing on AI-driven anomaly detection.** None of the three wedge gaps are AI-shaped.

**Standing rule for every future wedge feature request:** does this apply specifically to frontend telemetry Imora itself holds, or is it actually Adaeze's/Marcus's broader compliance-program problem (something a GRC tool like Drata/Vanta/OneTrust should own)? "Evidence export" and "retention policy" are exactly the shape of requirement that drifts into a full GRC product if this test isn't applied every time, not just at initial scoping.

Full reasoning: [`research/01-product/README.md#product-requirements-document-prd`](../research/01-product/README.md#product-requirements-document-prd).

---

## 4. Who This Is For

Six personas, split into **buyers** (why an org signs the contract) and **the daily user** (why the product actually gets opened once it's installed):

| Persona | Role | Carries | Needs from Imora |
|---|---|---|---|
| **Dara** | CISO, regional bank | Breach cost + CIPA litigation exposure | Provable fact that session data never left the perimeter |
| **Adaeze** | DPO, national insurer | 72-hr breach notification, 1-month DSAR clock | Access-audit-trail she can query herself, regulatory-clock retention |
| **Marcus** | HIPAA Security Officer, hospital network | Annual §164.312(b) risk assessment | Field-level access control + audit log as a standing report |
| **Priya** | Head of Platform Engineering, fintech | Operational burden, on-call load | Genuine 2–3 person deployability at small scale |
| **Jon** | Incident Commander / SRE | Fragmentation tax, chain-of-custody | One correlated investigation timeline, defensible evidence export |
| **Chidi** | Senior Frontend Engineer | Nothing — this is the adoption/retention check | Debugging UX at parity with Sentry/LogRocket, no compliance framing |

**Design note:** Personas 1–5 explain why an org buys Imora. Chidi explains whether anyone actually opens it on a day nothing is wrong. If a screen or flow only serves Personas 1–5, ask whether Chidi would ever willingly click into it — that's the adoption risk this product is specifically designed against.

Full persona detail with real scenarios: [`research/01-product/README.md#user-personas`](../research/01-product/README.md#user-personas). Screen-by-screen flows per persona: [`user-stories.md`](user-stories.md) in this folder.

---

## 5. Scope

### Milestone 0 — the actual next build target for a solo/small team

Milestone 1 below is the north star, not the next sprint — it's multi-engineer, multi-quarter scope at the pace comparable products were actually built at. **Milestone 0 is the resolved, concrete first slice:** narrow enough to ship in weeks, and still the one demo that proves the whole thesis — nobody else logs who viewed a session replay.

| Ships | Explicitly deferred (post-M0) |
|---|---|
| Session capture (rrweb-style) via `browser-sdk` core, vanilla JS only — no framework wrappers yet | React/Vue/Angular SDK wrappers |
| Simple capture-time masking: explicit allow-list + hard-redact-by-default (BR-7) | Soft-masking, `SecureFieldVault`, audited UNMASK escalation (story M2) |
| `ingestion` → ClickHouse + Postgres, single-machine only | Cluster profile, message queue, Kubernetes |
| `query-api` session/replay reads, with `AccessAuditEvent` generation structurally enforced (the actual wedge) | Error tracking, `ErrorGroup` deduplication, `alert-engine` |
| `dashboard`: session search + replay viewer + inline audit-trail panel, search defaults to `production` | Performance/Core Web Vitals monitoring, regression detection |
| **Environment tag** (`environment` set at SDK `init()`, carried on every event, filterable in session search) — cheap to include from day one, and the search-default behavior above depends on it existing | Environment-scoped release regression comparison (needs `alert-engine`, deferred anyway) |
| `gateway`: local auth (Argon2id), no SSO | Retention policy engine, legal hold, evidence export (Milestone 2 wedge) |
| Docker Compose, single 4-core/16GB host | Security-signal correlation, replay-to-trace correlation, `notification-service` |

**Exit criterion:** a named engineer can deploy the single-machine instance unassisted, capture a session, view its replay, and see that view already logged in the audit trail — the same proof-of-wedge moment [Onboarding](../research/09-workflows/README.md#onboarding) describes, just without the error-tracking and performance-monitoring halves of full parity attached yet. This is real product, not a prototype to throw away — Milestone 1 extends it, rather than replacing it.

### Parity requirements (the entry price — full Milestone 1 scope)

| Requirement | Bar to clear |
|---|---|
| Error tracking with grouping/deduplication by root cause | Sentry |
| Session replay, production-grade fidelity | LogRocket, FullStory, OpenReplay |
| Default-safe PII masking (deny-by-default, not opt-in) | FullStory "Private by Default," Sentry |
| Core Web Vitals monitoring (LCP < 2.5s, INP < 200ms, CLS < 0.1 at p75) with release-attributed regression detection | Sentry Releases |
| Replay-to-backend-trace correlation via shared session identifier | Sentry + OpenTelemetry pattern |
| Self-hosted: single-machine path for small teams, cluster path for scale, zero session data leaving the network boundary | OpenReplay, PostHog |
| Framework-agnostic browser SDK | Category-wide standard |
| **Environment tagging** (`production`/`staging`/`development`, or a team's own values) on all captured data, with searches defaulting to `production` so non-prod noise never silently mixes into a real investigation | Category-wide standard |
| Feature-parity checklist trackable against the tools being replaced | — |

### Wedge requirements (the reason to choose Imora), sequenced by milestone

| Requirement | Milestone | Depends on |
|---|---|---|
| Access-audit-trail: who viewed which session, when, from where | **M1** | Nothing — foundational |
| Field-level access control + audited "unmask" escalation | **M1** | Same access-control system as the audit trail |
| Per-data-category retention policy + legal-hold override | **M2** | M1's audit-trail event log (as the deletion-proof mechanism) |
| One-click, cross-signal evidence export (replay + errors + security + access log) | **M2** | M1's access-control system + M2's retention/legal-hold |
| Security-signal correlation into the incident timeline | **M2** | A security-event ingestion path that doesn't exist in M1 |
| Managed hosting, premium support, SSO/SAML, multi-region/HA | **M3** | Nothing in M1/M2 — this is the monetizable, non-wedge surface |

Full milestone thesis, exit criteria, and sequencing logic: [`research/08-roadmap/feature-roadmap.md`](../research/08-roadmap/feature-roadmap.md).

**Resolved:** the scope-realism gap between "north star" and "next sprint" is Milestone 0 above — build that first, expand toward the full M1 checklist from a working base rather than trying to ship the whole checklist before anything runs.

---

## 6. Success Metrics

**Activation (parity has to land):**
- Time-to-first-value under 1 hour, deployment to first captured session + first grouped error.
- Activation rate meaningfully above the 37.5% B2B SaaS average — Imora's buyers run structured POCs, not self-serve signups.

**POC success (the wedge has to survive a real evaluation):**
- A named compliance stakeholder can, unassisted, pull an access-audit-trail report and correctly identify every internal viewer of a session, within the POC window.
- A named engineer can deploy a working single-machine instance and capture a first session with no vendor support present.

**North Star Metric:** *Weekly Sessions with Full Audit Coverage* — sessions captured with default-safe masking **and** a complete, queryable access-audit-trail, simultaneously. Not "sessions captured" alone (parity only) and not "audit reports generated" alone (compliance only) — the one number that only moves when both halves of the product are working on the same session at once.

Full metric derivation: [`research/01-product/README.md#product-requirements-document-prd`](../research/01-product/README.md#product-requirements-document-prd).

---

## 7. Binding Constraints

These are not preferences — every screen, endpoint, and architecture decision has to satisfy all four:

1. **AGPLv3, no wedge feature ever gated behind a commercial tier.** The access-audit-trail, retention-policy engine, and evidence export ship free, self-hosted, no commercial agreement required — permanently, not just at launch. Full reasoning: [`research/01-product/README.md#licensing`](../research/01-product/README.md#licensing).
2. **Zero Parity or Wedge capability may depend on an external system in the required path.** An air-gapped deployment with no outbound network access must still get full audit-trail, masking, and retention enforcement. SSO, notifications, and backend trace correlation are the only things allowed to be optional. Source: [`research/03-architecture/README.md#system-context`](../research/03-architecture/README.md#system-context).
3. **Every VIEW, EXPORT, UNMASK, DELETE, and CONFIG_CHANGE against sensitive data produces exactly one audit event, structurally, not by convention.** A read that completes without logging is a defect, not an edge case. Source: [`research/02-domain/README.md#business-rules`](../research/02-domain/README.md#business-rules) (BR-5).
4. **Operational Simplicity is a hard requirement, not an aspiration.** A 2–3 person team must be able to deploy and operate the single-machine profile. If the audit-trail/compliance machinery makes that untrue, the milestone has failed on its own terms even if every feature technically shipped. **Resolved at the architecture level, not just asserted here:** audit-event generation is structural (the `AuditedQueryHandler` pattern in [`architecture.md §5`](architecture.md)), not a second system layered on top for someone to operate — and the concrete sizing in [`architecture.md §9`](architecture.md) (a single 4-core/16GB host runs all eight services plus the compliance machinery) is the falsifiable claim behind this constraint, not a promise. Source: [`research/08-roadmap/feature-roadmap.md`](../research/08-roadmap/feature-roadmap.md).

---

## What This Feeds

[`user-stories.md`](user-stories.md) — screen-by-screen, persona-by-persona detail a UI/UX designer can work from directly.
[`architecture.md`](architecture.md) — system design, tech stack, data flow, and deployment model an engineer can build from directly.
[`coding-standards.md`](coding-standards.md) and [`design-system.md`](design-system.md) — how the Go backend is actually written and structured, day to day.
