# Competitive Analysis

> Status: Research-based, current as of July 2026. Every claim below is sourced from public docs, changelogs, or GitHub issues — not assumption. Where a product's roadmap may have moved since, the source is linked so the claim can be re-verified rather than taken on faith.

This document backs the claims made in [vision.md](vision.md)'s "Gap in Today's Landscape" section. It exists so those claims stay honest as the market moves — re-check the sources here before repeating a claim from memory.

---

## Method

Eleven products were researched across three questions that matter specifically to regulated buyers (finance, healthcare, insurance, government):

1. Can the org run the full stack — including the control plane, not just a "dedicated" tenant — inside its own infrastructure?
2. Does session capture default to safe (deny-by-default / aggressive masking), or does a missed configuration leak PII?
3. Does the product answer "who on our team viewed this customer's data" — and does retention map to actual regulatory clocks (GDPR, HIPAA, PCI-DSS, SOX) rather than one global TTL?

No product answers all three. That gap is the basis for Imora's positioning, not a marketing assumption.

## How This Document Is Used

Per [vision.md](vision.md)'s Positioning section, Imora is scoped as **an alternative to the products below, not a new category** — so every product here is read for two things: what it takes to be a *credible* alternative (parity), and what none of them do (the wedge). The tables below are written to answer both questions directly, and the Synthesis at the end sorts findings into that split explicitly.

---

## Category 1 — SaaS Incumbents (deep observability, no true self-hosting)

| Product | Self-hostable? | Session replay masking | Compliance posture |
|---|---|---|---|
| **Datadog RUM** | No — cloud-only; session replay is browser-only, no on-prem option found | Configurable privacy controls; data stored on Datadog-managed cloud, encrypted at rest | HIPAA-eligible *with a signed BAA*, but customers must restrict workloads to eligible services and disable non-covered features |
| **New Relic Browser** | No — FedRAMP Moderate and HIPAA enablement exist, but as account-level compliance programs on New Relic's cloud, not a self-hosted deployment | Not the focus of research; New Relic markets "on-premises and cloud" support at the infrastructure-monitoring layer, not specifically for Browser/RUM self-hosting | FedRAMP Authorized (Moderate, partial service scope), HIPAA account enablement available |
| **LogRocket** | No — SaaS only, data lands on US-based servers by default | Strong: PII Labeling API, inline/blur masking, passwords never captured by default | SOC 2 Type II; markets HIPAA/GDPR/CCPA support, but data residency requires explicit cross-border transfer mechanisms |
| **FullStory** | No — SaaS only | Best-in-class default: "Private by Default" captures no text unless explicitly allow-listed; Exclude/Mask/Unmask tiers | SOC 2 Type 2, ISO 27001/42001, SOC 1/2/3 — strong certification story, but no confirmed HIPAA BAA |
| **Sentry (Cloud)** | Cloud product is not self-hosted by definition | Aggressive default masking — all text/images redacted client-side before leaving the browser | Standard SaaS compliance program |

**Takeaway:** this tier has the deepest product capability and, in FullStory's and Sentry's case, genuinely good default-safe capture. None of them solve the actual constraint regulated buyers have: the data and control plane living outside the organization's perimeter. A BAA or a FedRAMP badge is a contractual promise about a third party's environment, not organizational control over the data.

Sources: [Datadog HIPAA compliance](https://docs.datadoghq.com/data_security/hipaa_compliance/), [Datadog HIPAA-eligible services](https://www.datadoghq.com/legal/hipaa-eligible-services/), [Datadog RUM data security](https://docs.datadoghq.com/data_security/real_user_monitoring/), [New Relic HIPAA](https://docs.newrelic.com/docs/security/security-privacy/compliance/certificates-standards-regulations/hipaa/), [New Relic FedRAMP](https://docs.newrelic.com/docs/security/security-privacy/compliance/certificates-standards-regulations/fedramp/), [LogRocket Privacy docs](https://docs.logrocket.com/docs/privacy), [LogRocket Security docs](https://docs.logrocket.com/docs/security), [FullStory Private by Default](https://help.fullstory.com/hc/en-us/articles/360044349073-Fullstory-Private-by-Default), [FullStory Trust Center](https://trust.fullstory.com/), [Sentry: Protecting User Privacy in Session Replay](https://docs.sentry.io/security-legal-pii/scrubbing/protecting-user-privacy/).

---

## Category 2 — Open-Source, Self-Hosted Session-Replay Tools

These are the closest existing comparators to what Imora is building on the "session intelligence" axis.

| Product | Self-hostable? | Default PII masking | Notable caveat |
|---|---|---|---|
| **OpenReplay** | Yes — fully self-hosted including the ingestion/processing pipeline, "the only digital experience platform that can be fully self-hosted" per their own docs | Good: sanitizes at the tracker (browser) level before data ever reaches the server; emails auto-obscured by default | Own docs explicitly caveat that self-hosting alone ≠ GDPR compliance — still requires consent flows and DSAR handling on top |
| **PostHog** | Yes — open-source self-hosted is "the same exact product" as PostHog Cloud, per their own FAQ | Input elements masked by default | Session recording volume/retention is more constrained on self-hosted OSS than on Cloud; some enterprise controls are cloud/paid-tier only |
| **Highlight.io** | Yes, historically — Docker self-host, regex-based PII obfuscation and CSS-class masking by default | Reasonable default: obfuscates inputs and common PII regex patterns (SSNs, phone numbers, addresses) out of the box | **Acquired by LaunchDarkly in March 2025.** The standalone hosted product shuts down February 28, 2026, with existing accounts migrated into LaunchDarkly's observability platform. Not a safe long-term reference point — cite with this caveat every time. |

**Takeaway:** these products already do reasonably safe default capture — this is not the open gap it might look like from the outside, and Imora shouldn't market "default-safe capture" as if it invented the idea. The actual gap is one layer up: none of these three answer *"who on our team viewed this session,"* none map retention to a specific regulation's clock, and none produce an audit-ready evidence package. They solve "can I run this myself," not "can I prove to an auditor what happened and who saw it."

Sources: [OpenReplay Sanitize Data docs](https://docs.openreplay.com/en/sdk/sanitize-data/), [OpenReplay product page](https://openreplay.com/), [PostHog self-host docs](https://posthog.com/docs/self-host), [PostHog open-source disclaimer](https://posthog.com/docs/self-host/open-source/disclaimer), [Highlight.io GitHub](https://github.com/highlight/highlight), [Highlight.io privacy docs](https://www.highlight.io/docs/getting-started/client-sdk/replay-configuration/privacy), [Bugsink: a self-hosted alternative to Highlight.io](https://www.bugsink.com/a-self-hosted-alternative-to-highlight-io/), [7 Best Self-Hosted Session Replay Tools 2026](https://temps.sh/blog/best-self-hosted-session-replay-tools-2026).

---

## Category 3 — Adjacent Self-Hosted Tools (not real comparators, but worth naming so we don't overclaim)

| Product | What it actually is | Why it's not a session-intelligence competitor |
|---|---|---|
| **SigNoz** | Open-source, OpenTelemetry-native APM: logs, traces, metrics, exceptions | No native frontend RUM or session replay. It's a backend observability competitor to Datadog/New Relic, not to LogRocket/Highlight — there is nothing here to redact because there's no session capture in the first place. |
| **Grafana Faro** | Web SDK for RUM (errors, web vitals, traces) that feeds a self-run Loki/Tempo collector | No built-in session replay. Security writeups about Faro focus on *how to avoid* leaking PII into custom event attributes, not on redacting a replay — because there is no replay to redact. |
| **GlitchTip** | Lightweight, self-hosted, Sentry-SDK-compatible error tracker (Django, runs on 1 vCPU / 1GB RAM) | Explicitly error-tracking only by design. Calls to Sentry's Session Replay or Profiling APIs against a GlitchTip backend fail silently. |

**Takeaway:** it would be inaccurate to describe these as "self-hosted alternatives that get compliance wrong" — they simply don't operate in the session-replay/PII space at all. They're evidence of a different point: even within the self-hosted world, "frontend session intelligence" and "backend telemetry ownership" are two separate product categories that rarely combine, which is itself part of the fragmentation problem in [vision.md](vision.md).

Sources: [SigNoz](https://signoz.io/), [SigNoz GitHub](https://github.com/signoz/signoz), [Grafana Faro Web SDK](https://github.com/grafana/faro-web-sdk), [Frontend RUM Security: Grafana Faro](https://www.systemshardening.com/articles/observability/frontend-rum-security-grafana-faro/), [GlitchTip](https://glitchtip.com/), [GlitchTip vs Sentry vs Bugsink](https://www.bugsink.com/blog/glitchtip-vs-sentry-vs-bugsink/).

---

## Where "Sentry" specifically stands (a nuance worth keeping straight)

Sentry is source-available (Business Source License, not OSI open source) and can be self-hosted for free — but as of 2026, **Session Replay is still not reliably available on self-hosted Sentry.** The feature has an open tracking issue from the Sentry team going back multiple years, and users report replays either 404ing after 24 hours or landing in Kafka/ClickHouse without ever rendering in the UI. So the one SaaS incumbent with genuinely good default-safe masking (Sentry Cloud) doesn't let a regulated org self-host that specific capability at all today.

Sources: [Sentry self-hosted issue #1873 — Include Session Replay in Self Hosted](https://github.com/getsentry/self-hosted/issues/1873), [Sentry self-hosted issue #3963 — Replay Not Found](https://github.com/getsentry/self-hosted/issues/3963), [Sentry self-hosted issue #3274 — Replay detail 404](https://github.com/getsentry/self-hosted/issues/3274).

---

## Where the "access audit trail" pattern already exists (just not here)

The idea of logging *who viewed a specific recording* is not novel — it's standard in Privileged Access Management (PAM) tooling, where session recordings of infrastructure access are treated as auditable evidence: BeyondTrust, Delinea, and Keeper all log who accessed a recorded session, when, and why, and gate playback with role-based access control.

No frontend session-replay or observability product researched in Category 1 or 2 applies this same pattern to *customer* session data. That's a borrowed idea, not an invented one — which makes it a safer bet: it is a proven pattern in an adjacent, mature category (PAM), not a hypothetical.

Sources: [Audit Logs and Session Replay — hoop.dev](https://hoop.dev/blog/audit-logs-and-session-replay-the-powerful-duo-for-debugging-security-and-compliance), [BeyondTrust: Audit Recorded Sessions](https://www.beyondtrust.com/docs/privileged-identity/app-launcher-and-recording/audit.htm), [Delinea: Session Recording](https://delinea.com/products/secret-server/features/session-recording).

---

## Regulatory retention clocks (the actual numbers, for reference)

Every product reviewed above exposes a single, global retention TTL. The regulations regulated buyers actually answer to don't work that way:

| Regulation | Retention requirement |
|---|---|
| **PCI-DSS** (Requirement 10.7) | Minimum 12 months of audit trail history, with at least 3 months immediately available ("hot") for analysis |
| **HIPAA** (45 CFR §164.316(b)(2)(i) / §164.530(j)) | Minimum 6 years from creation or last-effective date for policies, procedures, and audit logs |
| **GDPR** (Article 5(1)(e)) | No fixed term — data must not be kept longer than necessary for its stated processing purpose ("storage limitation") |
| **SOX** | 7 years for relevant financial records and supporting audit documentation |

A platform with one global TTL forces a regulated org to either over-retain everything to satisfy the strictest regime (cost and privacy risk) or under-retain and fail an audit. Policy-per-data-category, mapped to the regulation that actually applies, is the fix — and it's a data-model and product decision, not a marketing claim.

Sources: [Compliance Log Retention Requirements by Regulation](https://claudiasop.com/blog/compliance-log-retention-requirements.html), [IT Log Retention: Complete Compliance Guide 2026](https://techjacksolutions.com/security/it-log-and-record-retention-requirements/).

---

## The Parity Checklist — What Makes Imora a Credible Alternative at All

Pulled directly from the tables above: this is what a regulated team's engineers already expect from Sentry, Datadog RUM, LogRocket, FullStory, OpenReplay, or PostHog, and what Imora has to match before the wedge below matters to anyone.

| Capability | Bar set by |
|---|---|
| Error tracking with grouping/deduplication | Sentry (category standard) |
| Session replay, production-fidelity | LogRocket, FullStory, OpenReplay |
| Default-safe PII masking (not opt-in) | FullStory ("Private by Default"), Sentry (aggressive default masking) |
| Performance monitoring against Core Web Vitals with release-based regression detection | Sentry Releases/regression issues, Datadog RUM |
| Self-hosted deployment (Docker/Kubernetes, single-machine to cluster) | OpenReplay, PostHog, GlitchTip |
| Framework-agnostic SDKs | All products reviewed |

None of this is differentiation — it's the entry price. A product that skips any row here isn't a credible alternative to what these teams already run, regardless of what it adds on top.

---

## Synthesis — The Wedge, Ranked by How Uncontested It Is

Everything below is the reason to pick Imora over *any* of the products above — including the other self-hosted ones — not just over the SaaS incumbents:

1. **Access-audit-trail for session data** (who viewed this customer's replay, when, why) — zero frontend products found doing this; proven pattern exists in PAM, uncontested to claim.
2. **Retention mapped to regulatory clocks**, not one global TTL — zero products found doing this; requirements are public and verifiable (table above), uncontested to claim.
3. **One-click, cross-signal evidence export** for auditors (replay + errors + security signal + access log as one timestamped package) — zero products found doing this; slightly softer claim since "evidence export" as a category is harder to search for than the other two, so treat as *strong* rather than *certain*.

What is **not** a safe wedge claim, and belongs in the parity checklist instead: "default-safe / deny-by-default capture." FullStory and Sentry already do this well. Imora should meet that bar, not market it as novel.

The parity checklist and this wedge ranking should directly inform `docs/01-product/feature-roadmap.md` and `docs/01-product/prd.md` — parity defines the MVP surface area, the wedge defines why a regulated buyer picks Imora over the closest self-hosted alternative once that surface area exists.
