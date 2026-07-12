# Workflows

## Onboarding

> Status: The concrete walkthrough behind [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)'s "under 1 hour, unassisted" time-to-value target and [feature-roadmap.md](../08-roadmap/feature-roadmap.md) Milestone 1's exit criteria ("a named engineer can stand up the single-machine deployment path and capture a first session with no vendor support present"). This document is what makes that target testable, not just aspirational.

---

### The Target

Two things have to be true within one hour of starting, with nobody from Imora involved:

1. A first session is captured, masked correctly, and visible in `dashboard` — the parity check.
2. An access-audit-trail entry exists for viewing that session, visible via the same UI — the wedge check, proven true from minute one, not held back for compliance-heavy customers to discover later.

---

### Walkthrough

#### 0–15 min: Deploy

`docker compose up`, per [Docker Compose](../12-infrastructure/README.md#docker-compose). The startup ordering (data layer → `minio-init` bucket-versioning-and-lock step → schema migrations → application services) runs automatically — there is no runbook step of "wait, then manually run X." This is the concrete test of [Deployment Model](../03-architecture/README.md#deployment-model)'s Operational Simplicity claim: if this step requires a human to intervene between sub-steps, the target has already failed regardless of what happens next.

#### 15–20 min: First login

Local authentication per [Authentication](../07-security/README.md#authentication) — Argon2id-hashed credentials, no SSO configuration required for this path, since SSO is Enterprise-tier per [Pricing](../01-product/README.md#pricing) and this walkthrough is the Community/Team-tier default.

#### 20–35 min: Install browser-sdk

`npm install @imora/core` (or the relevant framework wrapper, per [SDK API](../06-api/README.md#sdk-api)), `init({projectKey, release})` in the target application. Per [SDK API](../06-api/README.md#sdk-api)'s performance budget, this step should not be perceptibly slower than installing any comparable SDK — if bundle size or setup friction is the thing eating the hour, that's a defect in that document's own stated constraint, not an acceptable cost of onboarding.

#### 35–50 min: First data

Browse the instrumented application. A SessionEvent stream, at least one PerformanceMetric (Core Web Vitals), and ideally one triggered/caught error should appear in `dashboard` within seconds of the interaction, per the capture flow in [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) Flow A.

#### 50–60 min: Verify the wedge, not just the parity

Open the session just captured. View its replay — this alone is the parity check, and every SaaS incumbent and self-hosted alternative in [Competitive Analysis](../00-overview/README.md#competitive-analysis) can do this. **Then pull up that session's audit trail** ([REST API](../06-api/README.md#rest-api)'s `GET /v1/sessions/{id}/audit-trail`, or the equivalent dashboard view) and confirm the view just performed is already logged there, with the viewer's identity and timestamp. This is the step that makes the hour actually prove something: not just "I can see a replay," which was never in question, but "the audit trail is real, on by default, from the first session, with nothing to configure to turn it on."

---

### What Failing This Workflow Means

If any step above requires manual intervention, external support, or more than the stated time budget, that's a defect against [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)'s stated success metric — this document is the acceptance test for that metric, not a separate aspiration from it.

---

### What's Deliberately Not Modeled Here

- Exact dashboard screens/visual layout for each step — [Dashboard Wireframes](../10-design/README.md#dashboard-wireframes).
- SDK installation detail beyond the single `init()` call shown above — [SDK Installation](README.md#sdk-installation), next.

### What This Feeds Next

[SDK Installation](README.md#sdk-installation) expands the 20–35 minute step above into full framework-by-framework detail. [Error Investigation](README.md#error-investigation) and the other workflow files cover what happens after this first hour, once real usage begins.

---

## Session Replay

> Status: Story C3's specific workflow — finding the right session among many from a vague description, distinct from [Error Investigation](README.md#error-investigation)'s "I already have an alert pointing at a session" path.

---

### The Scenario

Per [User Stories](../01-product/README.md#user-stories) story C3: a support ticket says "checkout is broken," with no repro steps, no error ID, nothing to jump straight to. This is the harder, more common case — most bug reports don't arrive with a session already identified.

### The Workflow

1. **Search by user, timeframe, page, or signal — not just by ID.** Per story C3's acceptance criteria, candidate sessions surface in seconds from a user identifier, a URL, a timeframe, or an error/rage-click signal, not a manual log grep. This is the parity bar every comparator in [Competitive Analysis](../00-overview/README.md#competitive-analysis) sets — Imora has to match it before anything else here matters.
2. **Replay fidelity has to be sufficient on its own**, per [Domain Model](../02-domain/README.md#domain-model)'s rrweb-based capture (full snapshot plus incremental DOM/interaction events) — reproducing the reported behavior shouldn't require also reading raw logs side by side. If the replay alone can't answer "what did the user actually do," the capture layer has failed regardless of how good the search is.
3. **Masked by default, escalatable if needed.** Per [PII Redaction](../07-security/README.md#pii-redaction)'s two-tier model, the default view is safe to look at without a second thought — PHI/PII fields stay masked unless the specific debugging need requires the real value, in which case the audited UNMASK path from [Authorization](../07-security/README.md#authorization) applies exactly as it would anywhere else. Nothing about "just trying to reproduce a bug" gets a quieter audit trail.

### Why This Is Listed Separately From Error Investigation

[Error Investigation](README.md#error-investigation) starts from an alert that already points at a session. This workflow starts from nothing but a vague human description — and per [User Personas](../01-product/README.md#user-personas), that's Chidi's actual daily reality more often than a clean alert is. A product that's only good at the alert-triggered path but weak at open-ended search would fail the workflow that happens most often.

### What This Feeds Next

[Security Monitoring](README.md#security-monitoring) covers the equivalent search/investigation workflow from Adaeze's and Marcus's side — compliance-driven rather than debugging-driven, but built on the same underlying session data.

---

## Error Investigation

> Status: What an engineer actually sees and does, using the mechanisms already specified in [Event Catalog](../02-domain/README.md#event-catalog) and [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) — this document is the narrated, human-facing counterpart to those, not a restatement of the component-level flow.

---

### The Scenario

Jon (Incident Commander) gets paged. Chidi (per [User Personas](../01-product/README.md#user-personas)'s Wednesday scenario) is already looking at the same thing without having been paged at all — an error spiked after last night's deploy.

### What They See, Step by Step

1. **One alert, not a flood.** Per story C2 and [Event Catalog](../02-domain/README.md#event-catalog)'s `ErrorGrouped` event, the alert is for the root cause — one `AlertTriggered` per `ErrorGroup`, not one per affected session. This is the concrete payoff of write-time grouping: Jon isn't triaging a hundred near-identical pages at 2 a.m.
2. **"Which release did this" is already answered.** [Event Catalog](../02-domain/README.md#event-catalog)'s `RegressionDetected` event has already attributed the spike to last night's release, per story C1 — nobody bisects deploys by hand.
3. **Open one affected session's replay.** Watching what the user actually did leading up to the error — the parity capability every comparator in [Competitive Analysis](../00-overview/README.md#competitive-analysis) also offers.
4. **Jump straight to the backend trace for that exact moment.** Story J1's replay-to-trace correlation, via the shared session identifier from [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) — one click from "here's what the user saw" to "here's what the backend was doing at that instant," instead of manually correlating timestamps across two separate tools.
5. **If it's ambiguous whether this is a bug or an attack:** [Event Catalog](../02-domain/README.md#event-catalog)'s `SecuritySignalReceived` events for that session are already correlated into the same timeline, per story D2 — the wedge capability that answers "is this frontend logic, or is someone probing checkout for fraud" without switching to a separate security tool.

### If It Turns Out to Involve Customer Data

Per [User Personas](../01-product/README.md#user-personas)'s Jon scenario, an incident touching customer data carries a chain-of-custody burden most generic incident-response guidance doesn't cover. The path from here is [Security Monitoring](README.md#security-monitoring) (if a security signal was involved) or directly to generating an EvidenceExport (story J2) — a single action producing a frozen, hash-verifiable package rather than screenshots assembled under pressure.

---

### What Makes This Workflow Different From a Parity-Only Tool

Steps 1–3 are what any comparator offers. **Step 4 (replay-to-trace) is parity at the category-leading end, and step 5 (security correlation in the same timeline) is wedge** — per the tags established in [User Stories](../01-product/README.md#user-stories). The workflow reads as one continuous investigation specifically because those two tiers were designed to ship together from Milestone 1, per [feature-roadmap.md](../08-roadmap/feature-roadmap.md), not because the wedge was bolted on afterward.

### What This Feeds Next

[Performance Monitoring](README.md#performance-monitoring) covers the C1-driven regression-detection workflow in its own right, for cases where nothing broke outright but a metric moved. [Security Monitoring](README.md#security-monitoring) covers the D2 correlation path in depth.

---

## Performance Monitoring

> Status: The C1-driven workflow — what Chidi sees when a Core Web Vitals metric moves but nothing threw an error, distinct from [Error Investigation](README.md#error-investigation)'s exception-triggered path.

---

### The Scenario

Per [User Personas](../01-product/README.md#user-personas)'s Chidi scenario: a release shipped last night, and Chidi wants to know whether LCP, INP, or CLS moved for a given page or flow — with no error, no page, nothing broken in the conventional sense.

### The Workflow

1. **The threshold is the industry bar, not an arbitrary internal number.** LCP < 2.5s, INP < 200ms, CLS < 0.1, evaluated at the 75th percentile of real sessions — Google's own methodology, per [Target Users](../00-overview/README.md#target-users), not a metric Imora invented. A regression is measured against that bar, so "is this actually bad" isn't a judgment call Chidi has to make from scratch.
2. **p75, not average.** A regression affecting the slowest quarter of real users can hide entirely inside a mean that's dominated by fast connections — [Event Schema](../05-data/README.md#event-schema)'s `PerformanceMetricRecorded` event and [ClickHouse Schema](../05-data/README.md#clickhouse-schema)'s query shape are built around percentile evaluation specifically so this doesn't happen silently.
3. **Release attribution is automatic**, per story C1 and [Event Catalog](../02-domain/README.md#event-catalog)'s `RegressionDetected` event — a statistically significant change (per the trend-detection approach in [Target Users](../00-overview/README.md#target-users), matching Sentry's own regression-detection pattern) tied to the specific release that introduced it. Chidi doesn't bisect deploys by hand to find out which one moved the metric.
4. **Drill into the session level.** From the flagged regression, open representative affected sessions — the same replay capability [Error Investigation](README.md#error-investigation) uses, applied here to "why is this page slow" instead of "why did this throw."

### Why This Workflow Existing at All Is the Point

Per [Target Users](../00-overview/README.md#target-users)'s framing of Persona 6: this is a workflow with no compliance angle whatsoever, deliberately. If Chidi only ever opens the product during an incident or an audit, the product has already failed its own adoption thesis, per [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd)'s Goals. This workflow — checking on a Wednesday whether a routine deploy regressed anything — is what daily use actually looks like.

### What This Feeds Next

[Alerting](README.md#alerting) covers how a regression like the one described here actually reaches Chidi or Jon in the first place, rather than requiring someone to go looking.

---

## Alerting

> Status: How `AlertTriggered` (per [Event Catalog](../02-domain/README.md#event-catalog)) actually reaches a human, and how the routing decisions here are what determine whether Chidi tunes the product out — the concrete stakes named in [User Personas](../01-product/README.md#user-personas).

---

### The Stakes, Restated Concretely

A 2025 Catchpoint study found 62% of on-call engineers have ignored a critical alert because it was buried in noise, per [Target Users](../00-overview/README.md#target-users) and [User Stories](../01-product/README.md#user-stories) story C2. Every design choice below exists to keep Imora on the right side of that statistic, not the wrong one.

### The Workflow

1. **One alert per root cause is a data-model fact, not a UI filter.** Per [Domain Model](../02-domain/README.md#domain-model), `ErrorGroup` assignment happens at write time in `alert-engine` — by the time an `AlertTriggered` event exists, deduplication has already happened. There is no "smart grouping" layer trying to suppress noise after the fact; there's nothing to suppress because the noise was never generated.
2. **Delivery via [Webhooks](../06-api/README.md#webhooks), routed to wherever the team already works** — Slack, email, or a webhook into existing on-call tooling (PagerDuty-style), per `notification-service`'s Conformist relationship to `alert-engine` in [Bounded Contexts](../02-domain/README.md#bounded-contexts). Imora doesn't become a new place to check; it feeds the places already being checked.
3. **Severity is real, not decorative.** [Webhooks](../06-api/README.md#webhooks)'s `AlertTriggered` payload carries a severity field derived from the underlying signal — a Core Web Vitals regression (story C1) and an active security correlation (story D2) are not routed identically, since treating every alert as equally urgent is exactly the pattern that produces the ignored-alert statistic above.

### Alert Volume as a Tracked Product Metric

Per story C2's acceptance criteria: alert volume per real incident is itself something the product tracks and surfaces, not just an outcome hoped for. If grouping quality regresses — more alerts firing per actual root cause than before — that's visible as a metric before Chidi starts tuning the tool out, rather than discovered only when adoption quietly drops.

### What This Feeds Next

[Session Replay](README.md#session-replay) and [Security Monitoring](README.md#security-monitoring) cover the two workflows an alert from this document most often leads into.

---

## Security Monitoring

> Status: Adaeze's and Marcus's day, tying stories A1, A2, M1, M2, and J2 together into one workflow — the compliance-driven counterpart to [Error Investigation](README.md#error-investigation) and [Session Replay](README.md#session-replay)'s engineering-driven ones, built on the same underlying data.

---

### Scenario 1: A DSAR Arrives

Adaeze's 72-hour breach-notification clock and one-month DSAR clock, per [User Personas](../01-product/README.md#user-personas), both start the moment this scenario begins:

1. Query `GET /v1/data-subjects/{id}/sessions` ([REST API](../06-api/README.md#rest-api)) — every session tied to the data subject, plus the full access history for those sessions, in one lookup. Per story A1, this is minutes, not the multi-hour manual log search it replaces.
2. Export the response in a non-proprietary format (CSV/XML via content negotiation, per [REST API](../06-api/README.md#rest-api)) — GDPR's actual format requirement, not a PDF screenshot.
3. If the response involves a field currently masked, the same audited UNMASK path applies — Adaeze doesn't get a quieter audit trail than Chidi does, per [Authorization](../07-security/README.md#authorization)'s explicit no-exemptions rule.

### Scenario 2: An Investigation Requires Preservation

1. Apply a LegalHold via `POST /v1/legal-holds` ([REST API](../06-api/README.md#rest-api)), scoped by a structured predicate — session IDs, a data subject, a date range, or an incident reference, per [Postgres Schema](../05-data/README.md#postgres-schema) — not a fixed snapshot list, so records created after the hold that still match the scope are automatically covered.
2. Per [Threat Model](../07-security/README.md#threat-model)'s finding, an unbounded-scope hold (no date bound, no session-ID list) requires a second approver — a deliberate friction point on the one action in this document that could otherwise become a storage-exhaustion vector against [Scaling](../03-architecture/README.md#scaling)'s own math, not an oversight.

### Scenario 3: An Incident Needs a Defensible Record

Per story J2, and picking up directly from where [Error Investigation](README.md#error-investigation) leaves off when an incident turns out to involve customer data:

1. `POST /v1/evidence-exports` with the incident reference and relevant session IDs.
2. The result is a frozen, hash-verifiable package — immune to any later retention purge or erasure action touching its source data, per [Storage](../05-data/README.md#storage)'s MinIO Object Lock enforcement. This is the artifact that gets handed to legal or a regulator, not a set of screenshots assembled from memory.

### Scenario 4: Annual HIPAA Risk Assessment

Marcus's recurring, not incident-driven, workflow: `GET /v1/sessions/{id}/audit-trail` and its aggregate equivalents produce the standing report §164.312(b) requires — a built-in report against data already being collected, per [Authentication](../07-security/README.md#authentication) and [Audit Logging](../07-security/README.md#audit-logging), not a custom export job assembled once a year under deadline pressure.

### What Ties These Four Scenarios Together

Every one of them queries the same `AccessAuditEvent` entity from [Domain Model](../02-domain/README.md#domain-model) — a DSAR, a HIPAA risk assessment, and an incident evidence package are different views over one underlying audit trail, not four separate systems Adaeze and Marcus have to reconcile by hand. That's the concrete, day-to-day payoff of ADR [0005](../11-engineering/architecture-decisions/0005-unified-access-audit-event.md)'s decision to keep AccessAuditEvent unified.

---

## SDK Installation

> Status: Expands [Onboarding](README.md#onboarding)'s 20–35 minute installation step into full detail, per framework — direct application of [SDK API](../06-api/README.md#sdk-api)'s core-plus-wrapper architecture.

---

### Vanilla JS / Framework-Agnostic

```js
import { init } from '@imora/core';

init({
  projectKey: '<project-key>',
  release: process.env.BUILD_SHA,
});
```

Everything else in [SDK API](../06-api/README.md#sdk-api)'s public surface (`identify`, `captureException`, `addBreadcrumb`) is available from this same import — this is the baseline every framework wrapper builds on, not a separate integration path.

### React

```js
import { init } from '@imora/react';
// same init() call; the wrapper adds an error boundary and route-change tracking automatically
```

### Vue

```js
import { init } from '@imora/vue';
// adds Vue's errorCaptured hook automatically
```

### Angular

```js
import { init } from '@imora/angular';
// adds Angular's ErrorHandler integration automatically
```

Each wrapper re-exports the full `@imora/core` API per [SDK API](../06-api/README.md#sdk-api) — there is no capability available in one framework's package and not another's; a new framework wrapper is an adapter over already-tested core logic, not a parallel implementation with its own gaps.

---

### Classifying Sensitive Fields at Install Time

This is the step most likely to be skipped under time pressure, and skipping it doesn't fail open — per [PII Redaction](../07-security/README.md#pii-redaction) and business rule BR-7, any field with no classification is hard-redacted by default. That means the practical risk of skipping this step isn't a compliance gap, it's **debugging friction later** — a support engineer investigating a bug six weeks from now finds an unexpectedly redacted field, not exposed PII. Still worth doing at install time rather than discovering the gap under incident pressure:

```js
init({
  projectKey: '<project-key>',
  release: process.env.BUILD_SHA,
  fieldClassification: {
    safe: ['.product-name', '.page-title'],
    phi: ['[data-patient-field]', '.diagnosis-code'],
  },
});
```

Matches [PII Redaction](../07-security/README.md#pii-redaction)'s programmatic classification path — the same mechanism the `data-imora-safe`/`data-imora-mask` HTML attributes feed, for applications where hand-annotating markup isn't practical.

---

### Verifying Installation

The same check [Onboarding](README.md#onboarding) specifies: browse the instrumented app, confirm a session appears in `dashboard` within seconds, confirm Core Web Vitals are recorded. If nothing appears, the most common cause is a `projectKey` mismatch or an ad-blocker/CSP rule blocking the SDK's network calls to `gateway` — not a deeper integration problem.

### What This Feeds Next

[Error Investigation](README.md#error-investigation) picks up from here — what happens once real errors start arriving.

