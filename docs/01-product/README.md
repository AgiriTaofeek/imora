# Product

## Product Requirements Document (PRD)

> Status: Research-based, current as of July 2026. Consolidates [Vision](../00-overview/README.md#vision), [Competitive Analysis](../00-overview/README.md#competitive-analysis), [Target Users](../00-overview/README.md#target-users), [User Personas](README.md#user-personas), and [User Stories](README.md#user-stories) into a single scope decision. Nothing here should introduce a requirement that doesn't trace back to one of those documents — this is a consolidation exercise, not a new source of claims.

---

### Product Summary

Imora is a self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory for regulated industries. It is not a new category — per [Vision](../00-overview/README.md#vision)'s Positioning section, it has to clear the same parity bar as those tools and the closest self-hosted alternatives (OpenReplay, PostHog) to be adopted at all, and it wins the deal specifically on three capabilities none of them ship: an access-audit-trail over session data, retention mapped to regulatory clocks, and auditor-ready evidence export.

Every requirement below is tagged **[PARITY]** or **[WEDGE]**, per the convention established in [User Stories](README.md#user-stories).

---

### Goals

1. Be a credible, drop-in replacement for the observability tooling a regulated org's engineers already use daily (parity).
2. Be the only alternative on the list — self-hosted or SaaS — that a CISO, DPO, or HIPAA Security Officer can approve without a compensating control bolted on afterward (wedge).
3. Ship both from day one. Per [User Stories](README.md#user-stories)'s summary table, the split is roughly 7 parity / 6 wedge stories — this is deliberately not sequenced as "observability first, compliance later," because a compliance-only v1 has nothing for [Chidi](README.md#user-personas) to open daily, and an observability-only v1 gives a regulated buyer no reason to choose Imora over OpenReplay.

### Non-Goals

Explicit scope boundaries, so "add compliance features" doesn't silently expand into a different product:

- **Not a general backend APM/infrastructure observability tool.** SigNoz, Grafana Faro, and Datadog Infrastructure already own that; Imora's traces exist to be correlated with frontend sessions, not to replace a backend APM.
- **Not a full GRC/policy-management suite.** The wedge is audit trail, retention, and evidence export *for frontend telemetry Imora itself holds* — not a general compliance-workflow product like Drata, Vanta, or OneTrust. Adaeze and Marcus still need those tools for the rest of their compliance program.
- **Not a SIEM.** Security monitoring in [Vision](../00-overview/README.md#vision)'s "What We Are Building" means correlating security signal into the same incident timeline as replay/errors/performance — not ingesting arbitrary log sources or replacing a SOC's SIEM.
- **Not mobile-first.** Every persona and every researched pain point (CIPA cases, breach-cost data, HIPAA portal examples) is web-frontend-specific. Mobile SDKs are a future expansion, not MVP scope.
- **Not competing on AI-driven anomaly detection as the primary value.** It may be a later differentiator, but none of the three wedge gaps in [Competitive Analysis](../00-overview/README.md#competitive-analysis) are AI-shaped — building this first would be solving a problem nobody in [User Personas](README.md#user-personas) actually raised.

---

### MVP Scope

#### Parity requirements (the entry price — from the Parity Checklist in [Competitive Analysis](../00-overview/README.md#competitive-analysis))

| Requirement | Source story | Bar to clear |
|---|---|---|
| Error tracking with grouping/deduplication by root cause | C2 | Sentry |
| Session replay, production-grade fidelity | C3 | LogRocket, FullStory, OpenReplay |
| Default-safe PII masking (deny-by-default, not opt-in) | — (Guiding Principle) | FullStory "Private by Default," Sentry aggressive default masking |
| Performance monitoring against Core Web Vitals (LCP < 2.5s, INP < 200ms, CLS < 0.1 at p75) with release-attributed regression detection | C1 | Sentry Releases/regression issues |
| Replay-to-backend-trace correlation via shared session identifier | J1 | Sentry + OpenTelemetry pattern |
| Self-hosted deployment: single-machine path for evaluation/small teams, cluster path for scale, with no session data leaving the deployment's network boundary | P1, D1 | OpenReplay, PostHog |
| Framework-agnostic browser SDK | — (Guiding Principle) | Category-wide standard |
| Feature-parity checklist trackable against the specific tool categories being replaced, so a team can actually decommission old tools | P2 | — |

#### Wedge requirements (the reason to choose Imora — from the Synthesis in [Competitive Analysis](../00-overview/README.md#competitive-analysis))

Sequenced by build tractability, not just by importance — the audit trail is the most contained engineering problem of the three and ships first; retention-clock policy and evidence export depend on the audit trail existing first, so they follow rather than ship in parallel:

| Requirement | Source story | MVP or fast-follow |
|---|---|---|
| Access-audit-trail: log of user ID, event/action, record ID, timestamp, source IP for every view of a session/replay — the field set implementations converge on for HIPAA §164.312(b) | M1, A1 | **MVP** — foundational; A2/M2/J2 depend on this existing |
| Field-level access control with an audited "unmask" escalation path for masked PII | M2 | **MVP** — same underlying access-control system as the audit trail |
| Per-data-category retention policy (session replay, errors, security signal each configurable), not a single global TTL, including legal-hold override on scheduled deletion | A2 | Fast-follow — needs the audit trail's event log as its deletion-proof mechanism. Legal hold specifically fulfills the commitment in [Vision](../00-overview/README.md#vision)'s "Compliance Is a Workflow" principle, which previously had no story tracking it. |
| One-click, cross-signal evidence export (replay + errors + security signal + access log, timestamped, immutable once generated) | J2 | Fast-follow — depends on retention/legal-hold existing so an export can't be invalidated by a policy purging its sources mid-export |
| Security-signal correlation into the same incident timeline as replay/errors/performance | D2 | Fast-follow — requires a security-event ingestion path that doesn't yet exist in parity scope |

**MVP wedge is deliberately narrow: the audit trail and the access-control system underneath it.** This is the one wedge capability with no dependency on anything else on this list, and per [Competitive Analysis](../00-overview/README.md#competitive-analysis) it's also the most uncontested gap — proven in PAM tooling, absent from every frontend product reviewed. Retention-clock policy, evidence export, and security correlation are real requirements, not deferred indefinitely, but they're sequenced after MVP because each depends on the audit trail existing first.

---

### Success Metrics

#### Activation (Chidi's side — parity has to actually land)

- **Time-to-first-value target: under 1 hour** from deployment to first captured session + first grouped error, matching the benchmark that leading B2B PLG tools hit under-1-hour time-to-value, against a general SaaS average closer to 1.5 days.
- **Activation rate target: meaningfully above the 37.5% average** reported across 62 B2B SaaS companies (the median company loses roughly two-thirds of signups before they ever reach core value) — because Imora's actual buyers (Personas 1–3) are technical evaluators running a structured POC, not self-serve signups who churn silently.

#### Proof-of-Concept success (Dara/Adaeze/Marcus's side — the wedge has to survive a real evaluation)

Enterprise security/compliance POCs typically run 2–4 weeks for standard evaluations, extending to 6–12 weeks when compliance review is part of the scope — which it will be for every Imora POC by definition. Applying the standard POC discipline (success criteria defined and shared before the first demo, each criterion tied to a number/format/binary condition, at least half of test sessions run without the vendor present) gives concrete, falsifiable MVP success criteria:

- A named compliance stakeholder (DPO or HIPAA Security Officer persona) can, unassisted, pull an access-audit-trail report for a specific session and correctly identify every internal viewer, within the POC window.
- A named engineer (Priya/Jon/Chidi persona) can deploy a working single-machine instance and capture a first session without vendor support present — directly testing the Operational Simplicity principle from [Vision](../00-overview/README.md#vision).

#### North Star Metric

**Weekly Sessions with Full Audit Coverage** — the count of frontend sessions per week that are simultaneously (a) captured with default-safe masking and (b) have a complete, queryable access-audit-trail. This is deliberately not "sessions captured" alone (that's parity, and Sentry/LogRocket already win on volume) and not "audit reports generated" alone (that's a compliance-only metric Chidi never touches) — it's the one number that only goes up when parity and wedge are both working on the same session at the same time, which is the actual product thesis.

---

### Open Questions and Risks

- **Scope creep risk on the wedge.** "Evidence export" and "retention policy" are exactly the kind of requirements that expand into a full GRC product if unconstrained (see Non-Goals). Every wedge feature request should be tested against: does this apply specifically to frontend telemetry Imora holds, or is this actually Adaeze/Marcus's broader compliance-program problem that belongs in a different tool?
- **Operational Simplicity vs. the audit-trail system.** Priya's requirement (P1) is a 2–3 person deployable instance. An access-control-and-audit-trail system is real additional operational surface area (more state, more to back up, more to reason about in a security review). This tension needs to be resolved in `docs/03-architecture/README.md#architecture-overview`, not assumed away here.
- **Licensing model** — *resolved since this was first flagged:* AGPLv3 for the full product, wedge included, with no `ee/`-style commercial split, per [Licensing](README.md#licensing) and ADR [0001](../11-engineering/architecture-decisions/0001-agplv3-licensing.md). Kept in this list rather than deleted because the original concern explains *why* the license was a PRD prerequisite at all: a source-available license with SaaS-only premium features (the Sentry model) would have undermined the parity claim that a regulated org can run the full stack, control plane included.

---

### What This Feeds Next

The MVP/fast-follow split above is a scope decision, not a schedule — `docs/08-roadmap/feature-roadmap.md` should turn it into actual milestones with the licensing question in `docs/01-product/README.md#licensing` resolved first, since it constrains what "self-hosted" is even allowed to mean.

---

## User Personas

> Status: Research-based, current as of July 2026. Builds directly on [Target Users](../00-overview/README.md#target-users) — that document established *which roles* carry Imora's three cost drivers; this one grounds each role in a concrete scenario, real deadlines, and real litigation outcomes, so the personas describe lived pressure rather than an abstract title.

Per [Vision](../00-overview/README.md#vision)'s Positioning section, Imora is an alternative to tools these personas already have opinions about — Sentry, Datadog, LogRocket, FullStory — not a new category. Personas 1–5 (below) are why they'd switch; Persona 6 is why the switch sticks.

---

### Persona 1 — Dara, CISO at a Mid-Size Regional Bank

**Snapshot:** ~250-employee bank, ~40-person engineering org. Dara personally owns incident-response command per the org's size band, per [Target Users](../00-overview/README.md#target-users)'s org-size variant findings.

**The scenario she's afraid of, because it already happened to peers in her sector:** in April 2026, **Sutter Health** — a healthcare organization, not a retailer — agreed to a **$21.5 million** class-action settlement over claims it ran third-party tracking tools on its website that captured California visitors' data without consent. Separately, Bloomingdale's was sued (Mikulsky v. Bloomingdale's) specifically because its session-replay vendor "recorded and transmitted [a visitor's] interactions with the website — including mouse movements, keystrokes and page views — to a third-party vendor without her consent." That is the exact architecture of every SaaS session-replay product in [Competitive Analysis](../00-overview/README.md#competitive-analysis) Category 1.

**Why she can't just tell legal "we're fine":** the legal theory is genuinely unsettled, not a slam dunk against her, which makes it harder to dismiss, not easier. In Torres v. Prudential Financial — a financial-services case, her own sector — a court found session-replay vendors don't "read" data "in transit" and dismissed a CIPA §631 claim on that theory. But plaintiffs' firms responded by stacking a second claim under §638.51 (pen registers) specifically to survive that defense, and a Ninth Circuit reversal in a related case expanded potential liability rather than narrowing it. Dara's outside counsel cannot promise her this risk is closed — only that it is actively moving.

**What she needs from Imora:** the ability to state factually, in a board deck or a regulator inquiry, that session data was never transmitted to a third party — removing the fact pattern being litigated, rather than betting on which side of an unsettled circuit split her company lands on.

Sources: [Consumer Privacy Lawsuit Roundup 2026](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa), [Court Grants Summary Judgment: Session Replay Data "In Transit" — Inside Privacy](https://www.insideprivacy.com/data-privacy/court-grants-summary-judgment-website-vendor-cannot-read-session-replay-data-in-transit-under-cipa/), [Ninth Circuit Revives Session Replay Tracking Suit — Reed Smith](https://www.reedsmith.com/our-insights/blogs/viewpoints/102ksuo/ninth-circuit-revives-session-replay-tracking-suit/), [Website Wiretapping Roundup: 2025 Decisions](https://www.insideclassactions.com/2026/01/27/2025-website-wiretapping-roundup/).

---

### Persona 2 — Adaeze, Data Protection Officer at a National Insurer

**Snapshot:** appointed under GDPR Article 37 because the insurer conducts large-scale systematic monitoring of policyholders. Reports directly to the board, cannot be instructed by engineering or the CISO on how to do her job — a legal independence most of the other personas here don't have.

**Her clock, literally:** if a breach touching EU personal data occurs, she has **72 hours** from the moment the organization becomes aware of it to notify the supervisory authority — with partial, phased disclosure allowed only if the initial notification isn't delayed. Separately, any policyholder can file a Data Subject Access Request, and she has **one month** to produce a complete answer to "what data do you have on me, and who has looked at it."

**Why today's tools don't help her hit either deadline:** none of the products reviewed in [Competitive Analysis](../00-overview/README.md#competitive-analysis) can answer "which employees viewed this specific person's session recording" — that data either doesn't exist, or exists only as a generic "user logged in" system event that doesn't name the record they viewed. She currently answers DSARs about session data by asking engineering to manually search logs, which routinely eats a meaningful fraction of her one-month window before she's even confirmed what exists.

**What she needs from Imora:** an access-audit-trail she can query herself, without depending on engineering's cooperation under time pressure, plus retention that expires data on GDPR's storage-limitation clock instead of a platform-wide TTL she has to argue engineering into changing.

Sources: [GDPR Data Breach Notification: 72-Hour Rule — Recording Law](https://www.recordinglaw.com/world-laws/world-data-privacy-laws/eu-data-privacy-laws/gdpr-breach-notification-72-hour-rule/), [Key GDPR Breach Notification Requirements](https://www.reform.app/blog/gdpr-breach-notification-requirements), [What are the responsibilities of a DPO? — European Commission](https://commission.europa.eu/law/law-topic/data-protection/rules-business-and-organisations/obligations/data-protection-officers/what-are-responsibilities-data-protection-officer-dpo_en).

---

### Persona 3 — Marcus, HIPAA Security Officer at a Regional Hospital Network

**Snapshot:** owns the annual documented risk assessment covering ePHI systems and third parties/business associates, mandated by the HIPAA Security Rule — distinct from the hospital's Privacy Officer, who owns policy and patient-facing disclosure.

**His clock:** if a breach affects 500 or more patients, he has **60 days** from discovery to notify both HHS and every affected individual. Sutter Health's $21.5M settlement — a healthcare peer, sued over the exact third-party-tracking pattern his own patient portal uses — is the scenario his board now asks him about directly.

**Why the annual risk assessment keeps flagging the same gap:** the Security Rule requires him to document administrative, physical, and technical safeguards — access controls and audit logging chief among them — for anything touching ePHI. His current session-replay vendor (Category 1 or 2 in [Competitive Analysis](../00-overview/README.md#competitive-analysis)) can tell him *that* a support engineer viewed a patient's portal session, if he's lucky, but not produce that as a standing, queryable audit log he can hand an assessor without a custom export job.

**What he needs from Imora:** field-level access control and an audit log over *who viewed which patient's session*, structured so it satisfies the annual risk assessment as a built-in report, not a one-off favor from engineering before the audit deadline.

Sources: [HIPAA Security Officer — 2026 Update](https://www.hipaajournal.com/hipaa-security-officer/), [What are the notification requirements after a breach? — HIPAA Times](https://hipaatimes.com/what-are-the-notification-requirements-after-a-breach), [Consumer Privacy Lawsuit Roundup 2026](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa).

---

### Persona 4 — Priya, Head of Platform Engineering at a Mid-Stage Fintech

**Snapshot:** ~300 product engineers, which puts her platform group at the upper end of the typical **10–30 person** band mid-stage US fintechs run for that size of engineering org — roughly the standard **1:8–12** platform-to-product-engineer ratio.

**What's actually on her desk:** her on-call rotation needs a minimum of **8–10 engineers** just to cover the baseline security and reliability load for a regulated fintech — and every one of those engineers currently has to know four or five different tools (error tracker, session replay, APM, security/WAF dashboard) to investigate a single incident, none of which share an access model or a retention policy.

**Why she's the hardest sell, not the easiest:** she is the one directly evaluating whether her team can *operate* Imora long-term, not just whether it's compliant. She has already seen self-hosted tools become a second full-time job for someone on her team. She will reject anything that solves the CISO's and DPO's problems by making her team's operational burden worse.

**What she needs from Imora:** the "Operational Simplicity" commitment in [Vision](../00-overview/README.md#vision) to be real in practice — deployable by 2–3 people at minimum viable scale, without requiring her to grow platform headcount just to run the observability stack.

Sources: [Platform Engineering Team Size 2026](https://platformengineeringcost.com/team-structure), [At what company size should you adopt Platform Engineering? — SRE School](https://sreschool.com/forum/d/300-at-what-company-size-should-you-adopt-platform-engineering).

---

### Persona 5 — Jon, Incident Commander / Senior SRE on Priya's Team

**Snapshot:** rotates as Incident Commander on the 8–10-person on-call roster Priya staffs. Owns the reliability contract day to day — SLOs, capacity, and the observability tooling itself.

**A Tuesday, 2 a.m.:** a release regresses checkout for a subset of users. Jon's first twenty minutes aren't spent diagnosing — they're spent opening four separate tools, none of which share a timeline, to work out whether this is a frontend bug, a backend regression, or someone actively probing the checkout flow for fraud. If the incident turns out to involve exposed customer data, he now also owns **chain-of-custody documentation**, because a production incident touching customer data can become litigation, and the evidence has to hold up months later — not be reconstructed from Slack threads after the fact.

**What he needs from Imora:** replay, errors, performance, and security signal correlated into one timeline from the start, so the first twenty minutes go to diagnosis instead of reconstruction — and an evidence export that's already defensible, instead of a task he has to remember to do carefully at 2 a.m.

Source: [Overview of Incident Lifecycle in SRE — Squadcast](https://www.squadcast.com/blog/overview-of-incident-lifecycle-in-sre), platform/SRE role research cited under Persona 4.

---

### Persona 6 — Chidi, Senior Frontend Engineer on Priya's Team

**Snapshot:** no compliance mandate, no incident, no auditor in the room. Chidi opens Imora on an ordinary Wednesday because a deploy went out last night and something feels slower, or a support ticket says "checkout is broken" with no other detail.

**A normal Wednesday:** last night's release shipped. This morning, Chidi wants to know whether LCP or INP moved for the checkout flow — against the thresholds the entire industry already treats as the bar (LCP under 2.5s, INP under 200ms, CLS under 0.1, evaluated at the 75th percentile of real users, per Google's own Core Web Vitals methodology). If it did regress, Chidi expects the tool to already know which release caused it — the way Sentry's regression detection ties a statistically significant change in endpoint or function duration back to a specific deploy — rather than making Chidi bisect it by hand.

**Why noise is the thing that actually drives Chidi away:** a 2025 Catchpoint study found **62% of on-call engineers have ignored a critical alert because it was buried in noise.** If Imora pages Chidi once per affected user instead of once per root cause, Chidi tunes it out within a month, and every other capability in this document — audit trails, evidence export, retention policy — becomes irrelevant, because the person who was supposed to use the tool daily has stopped opening it.

**Why Chidi is in this document at all:** Personas 1–5 explain why Dara, Adaeze, Marcus, and Priya sign the contract. None of them explain why anyone *opens the product on a day nothing is wrong*. That's Chidi's job, and it's the actual adoption and retention driver — a tool CISOs mandate but engineers route around is a well-known failure mode, not a hypothetical one.

**What Chidi needs from Imora:** error grouping and performance regression detection that hold up against the category leaders on their own terms, with no compliance framing attached — because Chidi will never open a "compliance" tab, only a "why is this broken" one.

Sources: [How the Core Web Vitals metrics thresholds were defined — web.dev](https://web.dev/articles/defining-core-web-vitals-thresholds), [Alert Fatigue in SRE and DevOps — Sensu](https://sensu.io/blog/alert-fatigue-in-sre-and-devops), [Sentry Endpoint Regression docs](https://docs.sentry.io/product/issues/issue-details/performance-issues/endpoint-regressions/).

---

### What Changed From README.md#target-users to Here

README.md#target-users established the *role* and its regulatory obligations. This document adds the parts that make each persona feel like a specific person under specific pressure rather than a job description:

- **Real dollar figures from real settlements** (Sutter Health $21.5M, Bloomingdale's, LA Times $3.85M, Fandom/GameSpot $1.2M) — not hypothetical breach-cost averages.
- **Real deadlines** (GDPR 72-hour breach notification, GDPR 1-month DSAR, HIPAA 60-day breach notification) that turn "we should build an audit trail" into "she has one month and no way to answer the question today."
- **Real team-size constraints** (8–10 person on-call minimum, 1:8–12 platform ratio) that make Persona 4's operational-simplicity requirement concrete instead of a vague preference.
- **Persona 6 (Chidi)**, added deliberately after a review of the first five: Personas 1–5 explain why a regulated org buys Imora, but they're all compliance- or incident-driven, and none of them explain daily use. Without a persona whose job has nothing to do with compliance, the roadmap risks optimizing entirely for the buyer and neglecting the person who determines whether the product gets used at all.

This should feed directly into `docs/01-product/README.md#user-stories` as JTBD statements per persona, and into `docs/01-product/README.md#product-requirements-document-prd` once user stories exist to prioritize against.

---

## User Stories

> Status: Research-based, current as of July 2026. Translates [User Personas](README.md#user-personas) into Jobs-to-Be-Done stories with acceptance criteria grounded in the actual technical/regulatory mechanics each persona is held to — not generic "as a user" filler. Each story cites the specific requirement its acceptance criteria are derived from.

Format: **When** [situation], **I want to** [capability], **so I can** [outcome]. Acceptance criteria follow each story.

Per [Vision](../00-overview/README.md#vision)'s Positioning section, every story is tagged **[PARITY]** (necessary to be a credible alternative to Sentry/Datadog/LogRocket/FullStory/OpenReplay/PostHog at all) or **[WEDGE]** (the specific capability none of those alternatives have). Parity stories define the MVP surface area; wedge stories are why a regulated buyer picks Imora over the closest self-hosted alternative once that surface area exists. See the summary table at the end.

---

### Dara (CISO) — Stories

#### D1. Prove session data never left the perimeter — [PARITY]

*Self-hosting itself matches OpenReplay/PostHog; this story is table stakes for the alternative claim, not the wedge.*

**When** the board or a regulator asks whether customer session data has ever been transmitted to a third party, **I want to** point to an architecture where session capture, storage, and processing all run inside our own infrastructure, **so I can** state it as fact rather than a vendor's contractual promise.

**Acceptance criteria:**
- No session-replay payload (DOM mutations, keystrokes, network calls) leaves the deployment's own network boundary at any stage of capture, storage, or processing.
- This claim is independently verifiable — e.g., via network traffic inspection during a security review — not just documented in a privacy policy.
- Directly closes the fact pattern litigated in Mikulsky v. Bloomingdale's (see [User Personas](README.md#user-personas)), where liability turned on transmission "to a third-party vendor."

#### D2. Get an incident timeline before the war room, not during it — [WEDGE]

*Replay+errors+performance correlation is parity (Sentry/OpenTelemetry pattern); folding security signal into that same timeline is what no competitor does.*

**When** a P1 incident is declared, **I want to** see a single correlated timeline of replay, errors, performance, and security signal for the affected sessions, **so I can** brief the board on scope and cause within the hour, not after a multi-tool reconstruction.

**Acceptance criteria:**
- Session replay and backend traces/errors share a common session identifier propagated from browser to backend (via header or trace baggage), so a given session's frontend replay and backend spans/errors are queryable as one object — the standard mechanism used for OpenTelemetry-based frontend/backend correlation.
- Time-to-first-coherent-timeline is measured and reported, addressing the 20–40% incident-resolution-time tax from tool fragmentation cited in [Problem Statement](../00-overview/README.md#problem-statement).

---

### Adaeze (DPO) — Stories

#### A1. Answer a DSAR within the one-month window, reliably — [WEDGE]

**When** a policyholder files a Data Subject Access Request, **I want to** query, within minutes, what session and error data exists for that person and who on staff has viewed it, **so I can** respond inside GDPR's one-month deadline (extendable by two months only for genuinely complex requests) without a manual engineering search eating that window.

**Acceptance criteria:**
- Query by data-subject identifier returns all session records tied to that person, plus a log of every internal viewer of those records, in a single lookup.
- Export is delivered in a commonly used, non-proprietary electronic format (CSV/JSON/XML), matching GDPR's format requirement for DSAR responses — not a PDF screenshot or a proprietary dashboard link.
- This directly targets the reported outcome from automated DSR platforms: reducing per-request handling time from multiple hours to roughly 15 minutes.

#### A2. Enforce storage-limitation retention without depending on engineering to remember — [WEDGE]

**When** session or error data has outlived its documented processing purpose, **I want to** have it automatically deleted or anonymized on a policy I control, **so I can** demonstrate GDPR Article 5(1)(e) storage-limitation compliance without filing a ticket against engineering's backlog every quarter.

**Acceptance criteria:**
- Retention policy is configurable per data category (session replay, error events, security signal), not a single global TTL — matching the gap identified in [Competitive Analysis](../00-overview/README.md#competitive-analysis).
- Deletion is logged as an auditable event itself, so the DPO can prove *when* data was purged, not just that a TTL exists somewhere in config.
- A **legal hold** flag can be applied to specific records to override scheduled deletion when an investigation or litigation requires preservation — the mechanic behind the "Legal hold support" commitment in [Vision](../00-overview/README.md#vision)'s Guiding Principles, which otherwise has no story defining it. Holds are themselves logged (who applied it, when, why), so a hold can't be used to silently retain data past its policy window without a record of that decision.

---

### Marcus (HIPAA Security Officer) — Stories

#### M1. Produce the annual audit-control evidence without a custom export job — [WEDGE]

**When** the annual HIPAA risk assessment is due, **I want to** generate a report of every access event against ePHI-containing sessions, **so I can** satisfy 45 CFR §164.312(b)'s audit-controls requirement as a standing report rather than a one-off engineering favor.

**Acceptance criteria:**
- Every access-to-sensitive-record event logs, at minimum: user ID, event/action type (view, export, delete), the record identifier accessed, timestamp, and source IP/device — the field set standard implementations of §164.312(b) converge on, since the regulation itself specifies "record and examine activity" without prescribing exact fields.
- Logs are retained a minimum of six years from creation, matching the HIPAA documentation floor, and are reviewable on a defined schedule, not just stored.
- A written, exportable policy statement accompanies the logs describing what is logged, retention length, and who can access the logs themselves — assessors ask for the policy, not just the data.

#### M2. Redact PHI from a replay without losing the ability to debug it — [WEDGE]

*Default masking is parity (FullStory/Sentry already do this); the audited, reason-logged unmask escalation is the wedge.*

**When** a support engineer needs to investigate a bug in the patient portal, **I want to** let them view a redacted replay with PHI fields masked by default, escalatable to unmasked only with a logged, justified access request, **so I can** balance debuggability against minimum-necessary-access without blocking engineering entirely.

**Acceptance criteria:**
- Default view masks any field matching configured PHI patterns (name, MRN, diagnosis codes, DOB) at render time, not just at rest.
- An "unmask" action requires a reason field and is itself an audited access event, captured in the M1 log.

---

### Priya (Head of Platform Engineering) — Stories

#### P1. Deploy a working instance without growing headcount — [PARITY]

**When** evaluating whether to adopt Imora, **I want to** stand up a production-representative instance with a team of 2–3 platform engineers, **so I can** validate operational cost before committing budget — matching the minimum-viable platform team size her org already runs at.

**Acceptance criteria:**
- Single-machine deployment path exists for evaluation/small-scale production, separate from the multi-region/cluster path for scale — both documented as first-class, not "cluster-only" with a single-node deployment left as an afterthought, addressing the "Operational Simplicity" principle in [Vision](../00-overview/README.md#vision).
- No component requires bespoke, undocumented operational knowledge to keep running — a named on-call engineer outside the original deployer can operate it from the runbook alone.

#### P2. Retire redundant tools without losing coverage — [PARITY]

**When** consolidating onto Imora, **I want to** confirm that error tracking, session replay, performance monitoring, and security signal are all present at parity with the 4–5 tools currently stitched together, **so I can** actually decommission those tools rather than running Imora as a sixth dashboard.

**Acceptance criteria:**
- Feature parity checklist against the specific tool categories named in [Competitive Analysis](../00-overview/README.md#competitive-analysis) (error tracking, session replay, RUM/performance, security signal) is explicit and trackable, not assumed.
- Cost/tool-count reduction is measurable post-migration, directly addressing the $100K–$400K/year tool-sprawl figure cited in [Problem Statement](../00-overview/README.md#problem-statement).

---

### Jon (Incident Commander / SRE) — Stories

#### J1. Jump from a replay straight to the backend trace for that click — [PARITY]

**When** investigating a user-reported failure, **I want to** click a moment in a session replay and land directly on the backend trace/error for the exact API call happening at that moment, **so I can** skip manually correlating timestamps across tools.

**Acceptance criteria:**
- A shared session/trace identifier is attached as an attribute on every backend span associated with that session, propagated from the browser via request headers or trace baggage — the same mechanism OpenTelemetry-based frontend/backend correlation already uses in practice.
- Navigating from replay to trace and back is a single action, not a manual search by timestamp.

#### J2. Produce a defensible evidence package without extra effort at 2 a.m. — [WEDGE]

**When** an incident is confirmed to involve exposed or potentially exposed customer data, **I want to** export a single, timestamped package containing the relevant replay, errors, security signal, and access log for that incident, **so I can** hand it to legal or a regulator without reconstructing it from memory or Slack threads afterward — the chain-of-custody burden already on this role, per [User Personas](README.md#user-personas).

**Acceptance criteria:**
- Export is generated in one action from the incident view, not assembled by hand from four separate tools.
- Export is immutable/timestamped once generated, so its contents can't be silently altered after the fact — a baseline requirement for anything offered as litigation evidence.

---

### Chidi (Senior Frontend Engineer) — Stories

These stories carry no compliance framing on purpose — they're the counterweight to the rest of this document, addressing the rebalancing concern raised after the first draft (see [User Personas](README.md#user-personas) Persona 6): a tool that only wins on compliance risks becoming one engineers route around.

#### C1. Know immediately when a release regresses Core Web Vitals — [PARITY]

**When** a release ships, **I want to** see whether LCP, INP, or CLS moved for affected pages, evaluated at the 75th percentile the way Google's own methodology does, **so I can** catch a performance regression the same day, not after it shows up in a support queue.

**Acceptance criteria:**
- Regressions are flagged against the standard "good" thresholds (LCP < 2.5s, INP < 200ms, CLS < 0.1) at the 75th-percentile of real user sessions, not a mean that hides tail degradation.
- The regression is automatically attributed to the release that introduced it, using trend detection over a statistically meaningful window — the approach the category standard (Sentry Releases/regression issues) already uses — rather than requiring manual bisection across deploys.

#### C2. Get one alert per root cause, not one per affected user — [PARITY]

**When** an error spikes in production, **I want to** be notified once for the underlying issue, with all affected sessions grouped under it, **so I can** triage instead of drowning — directly countering the documented pattern where 62% of on-call engineers admit to having ignored a critical alert buried in noise.

**Acceptance criteria:**
- Errors sharing a root cause (same stack trace, same failing endpoint) are grouped into a single actionable issue, not one notification per occurrence.
- Alert volume per real incident is a tracked product metric — if grouping quality regresses, it's visible before Chidi starts tuning the tool out.

#### C3. Reproduce a badly-described bug from a support ticket in minutes — [PARITY]

**When** a support ticket says "checkout is broken" with no repro steps, **I want to** find the matching session replay by user, timeframe, or page, and watch exactly what happened, **so I can** reproduce and fix the bug without asking the reporter to redo it while being screen-shared.

**Acceptance criteria:**
- Session search by user identifier, URL, timeframe, and error/rage-click signal returns candidate sessions in seconds, not a manual log grep.
- Replay fidelity (DOM state, network calls, console errors) is sufficient to reproduce the reported behavior without needing to also read raw logs side by side — matching the baseline UX quality of the category leaders this persona is used to.

---

### Parity vs. Wedge Summary

| Story | Persona | Tag |
|---|---|---|
| D1 — Prove session data never left the perimeter | Dara | PARITY |
| D2 — Incident timeline before the war room | Dara | WEDGE |
| A1 — Answer a DSAR within one month | Adaeze | WEDGE |
| A2 — Enforce storage-limitation retention | Adaeze | WEDGE |
| M1 — Annual audit-control evidence | Marcus | WEDGE |
| M2 — Redact PHI with audited unmask | Marcus | WEDGE |
| P1 — Deploy without growing headcount | Priya | PARITY |
| P2 — Retire redundant tools at parity | Priya | PARITY |
| J1 — Replay-to-trace jump | Jon | PARITY |
| J2 — Defensible evidence package | Jon | WEDGE |
| C1 — Release regression detection | Chidi | PARITY |
| C2 — One alert per root cause | Chidi | PARITY |
| C3 — Reproduce a bug from a vague ticket | Chidi | PARITY |

**7 parity, 6 wedge.** Roughly even, which is the point: this is not a compliance tool with an observability veneer, and it's not an observability tool with compliance bolted on — the two are meant to ship as one product from the start. Note that the wedge stories cluster almost entirely on Dara, Adaeze, and Marcus (the buyers), while parity stories cluster on Priya, Jon, and Chidi (the daily users) — confirming the split first identified in [Target Users](../00-overview/README.md#target-users): parity earns adoption, the wedge earns the contract.

### What This Feeds Next

These stories, especially the acceptance criteria tied to specific regulatory field/format/retention requirements (M1, A1, A2), are detailed enough to seed `docs/01-product/README.md#product-requirements-document-prd` directly. The parity/wedge split above should determine MVP sequencing: parity stories are the entry price for `docs/08-roadmap/feature-roadmap.md`'s earliest milestones, and the correlation mechanism described in D2/J1 (shared session ID propagated into trace context) is a real architectural decision that belongs in `docs/03-architecture/README.md#architecture-overview` once product scope is fixed.

---

## Licensing

> Status: Research-based, current as of July 2026. Flagged as a prerequisite in [Product Requirements Document (PRD)](README.md#product-requirements-document-prd)'s Open Questions section — the license determines whether "self-hosted" actually delivers on the promise in [Vision](../00-overview/README.md#vision), so it has to be decided before [feature-roadmap.md](../08-roadmap/feature-roadmap.md) sequences anything.

---

### The Decision This Document Has to Get Right

Per [Vision](../00-overview/README.md#vision)'s Positioning section, Imora's entire pitch is: parity with Sentry/Datadog/LogRocket/FullStory, plus wedge capabilities (access-audit-trail, regulatory retention, evidence export) that no alternative — SaaS or self-hosted — currently ships. The license has to protect that pitch, not quietly undermine it. Two ways licensing commonly undermines exactly this kind of pitch, found directly in the products already researched for [Competitive Analysis](../00-overview/README.md#competitive-analysis):

1. **Feature-gating the wedge behind a commercial tier.** PostHog's core is MIT-licensed and fully functional self-hosted — but SSO enforcement, RBAC, advanced permissions, and **audit logs specifically** live in its `ee/` directory under a separate PostHog Enterprise License requiring a commercial agreement for production use. If Imora did this to the audit-trail wedge, a regulated org would face the same problem self-hosting was supposed to solve: needing an ongoing commercial relationship just to get the compliance capability that is the entire reason they picked Imora over OpenReplay. This is the single most important thing this document has to rule out.
2. **License-restricted self-hosting that isn't actually equivalent to the SaaS product.** Sentry's Functional Source License states there's no intended feature difference between SaaS and self-hosted — but in practice, self-hosted Sentry still doesn't reliably support Session Replay (documented in [Competitive Analysis](../00-overview/README.md#competitive-analysis)'s "Where Sentry specifically stands" section). The lesson: a license clause promising parity doesn't guarantee engineering delivers it. This document controls the legal terms; [feature-roadmap.md](../08-roadmap/feature-roadmap.md) has to actually ship parity for the promise to mean anything.

---

### What Comparable Products Actually Do

| Product | License | Consequence |
|---|---|---|
| **Sentry** | Functional Source License (FSL) — source-available, not OSI-approved. Prohibits reselling self-hosted Sentry as a competing offering or using the code to build a direct competitor. Converts to Apache 2.0 after a 2-year grace period. | No feature-gating by license text, but self-hosted deployments lag SaaS in practice (session replay). Source-available licenses also carry community-trust risk — see HashiCorp below. |
| **PostHog** | MIT core + separate "PostHog Enterprise License" (source-available, commercial-agreement-required) for `ee/` features. | Fully functional OSS core, but SSO, RBAC, and **audit logs** require a paid agreement — exactly the anti-pattern this document exists to avoid replicating. |
| **OpenReplay** | AGPLv3 core + a separate non-open enterprise license for some parts. | Same open-core shape as PostHog, one license family more protective against re-hosting (see AGPL below). |
| **HashiCorp Terraform** | Switched from MPL (permissive) to BSL after building years of community trust under an open license. | Triggered a community fork (OpenTofu). The lesson for Imora: decide the license model now, before a community forms around different expectations — relicensing later is the highest-trust-cost path. |

Source: [Sentry Licensing](https://open.sentry.io/licensing/), [Introducing the Functional Source License — Sentry Blog](https://blog.sentry.io/introducing-the-functional-source-license-freedom-without-free-riding/), [PostHog Enterprise License — GitHub](https://github.com/PostHog/posthog/blob/master/ee/LICENSE), [PostHog Open Source docs](https://posthog.com/docs/posthog-code/open-source), [Open Source Licenses Explained: AGPL, MIT, GPL, Apache 2.0](https://www.opensourcealternatives.to/blog/open-source-license-guide).

---

### The Recommendation

#### Core product — including the full wedge — under AGPLv3, not MIT and not a source-available/BSL-family license

- **AGPLv3, not MIT/Apache:** a permissive license lets any cloud provider — including a Datadog or a well-funded competitor — take Imora's code, host it, and sell it as a competing SaaS product without contributing anything back. Since Imora explicitly competes with SaaS incumbents (per [Vision](../00-overview/README.md#vision)'s Positioning), this isn't a hypothetical: it's the exact business risk a permissive license creates for a product whose whole model is "we compete with the SaaS players." AGPL's network-use clause closes this — anyone running a modified version as a network service must release their modifications under the same terms. This is the same reasoning Nextcloud and Mastodon apply to their own self-hosted-competitor-to-SaaS products.
- **Not a source-available/BSL-family license (Sentry's FSL, HashiCorp's old BSL):** these exist specifically to prevent competitors from reselling the product, which protects revenue but isn't OSI-approved open source — and per [Target Users](../00-overview/README.md#target-users), procurement teams at regulated orgs actively screen license type as an evaluation criterion. A non-OSI license is a harder sell to exactly the buyers (Dara, Adaeze) this product depends on, for a benefit (anti-resale protection) that AGPL already provides via the network-use clause.
- **No `ee/`-style split that puts the wedge behind a commercial license.** The access-audit-trail, retention-clock policy, and evidence export — the entire reason a regulated org picks Imora per [Product Requirements Document (PRD)](README.md#product-requirements-document-prd)'s MVP scope — ship under the same AGPLv3 terms as everything else, self-hosted, no commercial agreement required. This is the direct fix for the PostHog anti-pattern identified above, and it's a non-negotiable constraint on [feature-roadmap.md](../08-roadmap/feature-roadmap.md): no milestone may move a wedge capability into a paid tier.

#### What is legitimately fine to monetize

Consistent with the principle above — monetize things that are not the reason a regulated buyer chose Imora over the alternatives:

- **Managed hosting** for organizations that want Imora's wedge capabilities without operating the infrastructure themselves — this doesn't touch the self-hosting promise because it's optional, not the only way to get compliance features.
- **Premium support / SLA-backed response times** — Priya's persona ([User Personas](README.md#user-personas)) explicitly worries about operating this without growing headcount; paid support is a legitimate answer to that, not a gate on functionality.
- **SSO/SAML enterprise auth integrations** — unlike audit logs, enterprise SSO is genuinely a nice-to-have convenience feature, not one of the three wedge capabilities identified in [Competitive Analysis](../00-overview/README.md#competitive-analysis). Gating this is a defensible parallel to PostHog's model, since it doesn't touch the compliance promise.
- **Multi-region/HA orchestration tooling and professional services** for air-gapped or complex deployment topologies — operational complexity, not core product capability.

---

### What This Rules Out for feature-roadmap.md

- No milestone may ship the audit-trail, retention-policy, or evidence-export wedge features as "Enterprise" or license-gated. If a future roadmap draft does this, it contradicts this document and needs to come back here first.
- The license decision (AGPLv3) should be adopted before the first public commit, per the HashiCorp/Terraform lesson above — changing it later, after a community has formed under different expectations, is the costliest version of this decision to get wrong.

---

## Pricing

> Status: Research-based, current as of July 2026. Builds on [Licensing](README.md#licensing)'s monetization surface (managed hosting, support SLA, SSO/SAML, multi-region tooling) and constraint (no wedge capability may ever be paywalled). This document decides *how* those specific things get priced.

---

### The Constraint This Model Has to Satisfy

Per [Licensing](README.md#licensing), the core product — full parity plus the entire compliance wedge (access-audit-trail, retention-clock policy, evidence export) — ships AGPLv3, self-hosted, free, no commercial agreement required. Pricing only applies to the Milestone 3 surface from [feature-roadmap.md](../08-roadmap/feature-roadmap.md): managed hosting, premium support/SLA, SSO/SAML, and multi-region/HA tooling. Nothing below may fund itself by gating anything on that list.

---

### Why the Category-Standard Model Doesn't Fit

Both direct comparators price on metered usage: **PostHog** charges per event/session/replay with step-down volume tiers (roughly $0.00005/event at 1–2M events/month, decreasing to $0.000009/event above 250M). **Sentry** charges per error/trace/replay processed, with tiered per-unit rates once a plan's included volume is exceeded. This is the pricing model most engineers evaluating Imora will already expect.

It doesn't work for Imora's actual buyers, for two independent reasons found in research, not just a stylistic preference:

1. **Metering requires reporting usage back to the vendor.** PostHog and Sentry can meter accurately because they host the product themselves (SaaS) — the events never leave their infrastructure to begin with. A self-hosted, air-gapped Imora deployment has no outbound path to report usage at all, by design, per the Operational Simplicity principle in [Vision](../00-overview/README.md#vision). Metered pricing would either require breaking the air-gap (unacceptable to Dara and Adaeze) or estimating usage on trust, which isn't really metered pricing anymore.
2. **Regulated procurement structurally resists metered billing regardless.** Enterprise agreements are overwhelmingly flat-fee, because large buyers' budget processes require a committed number and procurement won't sign a contract where the invoice changes quarter to quarter based on consumption they can't fully control. Where a usage component exists at enterprise scale, it gets negotiated into a fixed annual commitment or a capped ceiling — not left as a fluctuating metered bill. This isn't specific to air-gapped buyers; it's how Dara's and Adaeze's procurement processes work regardless of deployment model.

Copying PostHog or Sentry's pricing model here would repeat the same mistake flagged in [Licensing](README.md#licensing) about copying PostHog's feature-gating — importing a pattern that fits the comparator's business model but actively works against Imora's actual buyers.

---

### The Recommendation

#### Self-hosted core — always free, no license key at all

Parity and wedge capabilities require no licensing mechanism whatsoever — not a free tier with a key, an actual absence of gating. This is the only way to make the "no compensating control bolted on afterward" claim from [Product Requirements Document (PRD)](README.md#product-requirements-document-prd)'s Goals literally true.

#### Milestone 3 commercial add-ons — flat annual pricing, tiered by seat/deployment band, not metered

Per the research above, and consistent with the finding that per-seat models are a natural fit for regulated industries because they already need auditable user counts for compliance reasons anyway:

| Tier | Who it's for (per [Target Users](../00-overview/README.md#target-users) org-size variants) | What's priced |
|---|---|---|
| **Community** | Under ~50 employees — the CTO-wears-every-hat org | $0. Full AGPLv3 core, community support only. |
| **Team** | ~50–300 employees | Flat annual fee, tiered by a self-declared seat-count band agreed at signing — not metered. Adds premium support/SLA. |
| **Enterprise** | 300+ employees, per-persona split intact (CISO, DPO, HIPAA Security Officer, Platform lead as separate buyers) | Flat annual contract, negotiated per deployment scale. Adds SSO/SAML, multi-region/HA tooling, dedicated support, and offline license activation for air-gapped environments. |

Seat/deployment bands are declared at contract signing and renewal — an auditable, contractual fact regulated buyers already produce for their own compliance programs — not measured by telemetry Imora's own architecture doesn't collect.

#### SSO/multi-region gating — offline signed license files, not phone-home activation

For the one thing in Milestone 3 that does need a technical gate (SSO/SAML, multi-region tooling), the standard pattern for government, military, healthcare, and financial air-gapped deployments is a cryptographically signed offline license file (Ed25519 or RSA-2048), generated in a connected environment and transferred into the air-gapped network manually (USB, or a QR-code-based exchange), rather than a server that periodically phones home to validate. This works identically for a connected Team-tier customer and a fully air-gapped Enterprise-tier one — the same mechanism serves both, so there's no separate "air-gapped SKU" to maintain.

#### Managed hosting — the one place usage-based pricing is legitimate

If Imora offers managed hosting as part of Milestone 3, metered/usage-based pricing is fine there, for the same reason it works for PostHog Cloud and Sentry Cloud: Imora would be running the infrastructure directly, so metering doesn't require anything to phone home across an air-gap — there is no air-gap in that deployment model. This is the one SKU where copying the category-standard pricing pattern is actually correct, precisely because it's the one SKU that isn't self-hosted.

---

### Open Questions and Risks

- **Self-declared seat bands are a trust model, not a metered one.** At Enterprise scale this is standard (regulated buyers already self-attest headcount for their own compliance programs), but it means Imora's revenue at that tier depends on contract terms and audit rights, not technical enforcement — a deliberate tradeoff for the air-gap constraint, not an oversight.
- **Community tier at $0 with no seat ceiling could be used by orgs well past 50 employees who just don't sign a contract.** Per [Licensing](README.md#licensing), this is a permitted consequence of choosing AGPLv3 over a restrictive license — the tradeoff for procurement-friendly, non-OSI-avoidance licensing is that Team/Enterprise tiers have to sell on support and features the Community tier lacks (SLA, SSO, multi-region), not on artificial scarcity of the core product.
- **Whether a "Team" tier is worth maintaining at all**, versus going straight from free Community to negotiated Enterprise, is a real open question that depends on actual demand data this document can't produce — flagged here rather than guessed at.

Sources: [PostHog pricing](https://posthog.com/pricing), [Sentry pricing](https://sentry.io/pricing/), [Air-Gapped License Activation — LicenseSpring](https://docs.licensespring.com/license-entitlements/activation-types/air-gapped), [How to Implement an Offline Licensing Model — Keygen](https://keygen.sh/docs/choosing-a-licensing-model/offline-licenses/), [Enterprise SaaS Pricing: List Price vs Negotiated Deals](https://softwarepricing.com/blog/enterprise-saas-pricing/), [Enterprise SaaS Pricing Models Compared — m3ter](https://www.m3ter.com/blog/enterprise-saas-pricing-models-enterprise-pricing-strategy).

