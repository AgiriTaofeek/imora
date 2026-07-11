# Product Requirements Document (PRD)

> Status: Research-based, current as of July 2026. Consolidates [vision.md](../00-overview/vision.md), [competitive-analysis.md](../00-overview/competitive-analysis.md), [target-users.md](../00-overview/target-users.md), [user-personas.md](user-personas.md), and [user-stories.md](user-stories.md) into a single scope decision. Nothing here should introduce a requirement that doesn't trace back to one of those documents — this is a consolidation exercise, not a new source of claims.

---

## Product Summary

Imora is a self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory for regulated industries. It is not a new category — per [vision.md](../00-overview/vision.md)'s Positioning section, it has to clear the same parity bar as those tools and the closest self-hosted alternatives (OpenReplay, PostHog) to be adopted at all, and it wins the deal specifically on three capabilities none of them ship: an access-audit-trail over session data, retention mapped to regulatory clocks, and auditor-ready evidence export.

Every requirement below is tagged **[PARITY]** or **[WEDGE]**, per the convention established in [user-stories.md](user-stories.md).

---

## Goals

1. Be a credible, drop-in replacement for the observability tooling a regulated org's engineers already use daily (parity).
2. Be the only alternative on the list — self-hosted or SaaS — that a CISO, DPO, or HIPAA Security Officer can approve without a compensating control bolted on afterward (wedge).
3. Ship both from day one. Per [user-stories.md](user-stories.md)'s summary table, the split is roughly 7 parity / 6 wedge stories — this is deliberately not sequenced as "observability first, compliance later," because a compliance-only v1 has nothing for [Chidi](user-personas.md) to open daily, and an observability-only v1 gives a regulated buyer no reason to choose Imora over OpenReplay.

## Non-Goals

Explicit scope boundaries, so "add compliance features" doesn't silently expand into a different product:

- **Not a general backend APM/infrastructure observability tool.** SigNoz, Grafana Faro, and Datadog Infrastructure already own that; Imora's traces exist to be correlated with frontend sessions, not to replace a backend APM.
- **Not a full GRC/policy-management suite.** The wedge is audit trail, retention, and evidence export *for frontend telemetry Imora itself holds* — not a general compliance-workflow product like Drata, Vanta, or OneTrust. Adaeze and Marcus still need those tools for the rest of their compliance program.
- **Not a SIEM.** Security monitoring in [vision.md](../00-overview/vision.md)'s "What We Are Building" means correlating security signal into the same incident timeline as replay/errors/performance — not ingesting arbitrary log sources or replacing a SOC's SIEM.
- **Not mobile-first.** Every persona and every researched pain point (CIPA cases, breach-cost data, HIPAA portal examples) is web-frontend-specific. Mobile SDKs are a future expansion, not MVP scope.
- **Not competing on AI-driven anomaly detection as the primary value.** It may be a later differentiator, but none of the three wedge gaps in [competitive-analysis.md](../00-overview/competitive-analysis.md) are AI-shaped — building this first would be solving a problem nobody in [user-personas.md](user-personas.md) actually raised.

---

## MVP Scope

### Parity requirements (the entry price — from the Parity Checklist in [competitive-analysis.md](../00-overview/competitive-analysis.md))

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

### Wedge requirements (the reason to choose Imora — from the Synthesis in [competitive-analysis.md](../00-overview/competitive-analysis.md))

Sequenced by build tractability, not just by importance — the audit trail is the most contained engineering problem of the three and ships first; retention-clock policy and evidence export depend on the audit trail existing first, so they follow rather than ship in parallel:

| Requirement | Source story | MVP or fast-follow |
|---|---|---|
| Access-audit-trail: log of user ID, event/action, record ID, timestamp, source IP for every view of a session/replay — the field set implementations converge on for HIPAA §164.312(b) | M1, A1 | **MVP** — foundational; A2/M2/J2 depend on this existing |
| Field-level access control with an audited "unmask" escalation path for masked PII | M2 | **MVP** — same underlying access-control system as the audit trail |
| Per-data-category retention policy (session replay, errors, security signal each configurable), not a single global TTL, including legal-hold override on scheduled deletion | A2 | Fast-follow — needs the audit trail's event log as its deletion-proof mechanism. Legal hold specifically fulfills the commitment in [vision.md](../00-overview/vision.md)'s "Compliance Is a Workflow" principle, which previously had no story tracking it. |
| One-click, cross-signal evidence export (replay + errors + security signal + access log, timestamped, immutable once generated) | J2 | Fast-follow — depends on retention/legal-hold existing so an export can't be invalidated by a policy purging its sources mid-export |
| Security-signal correlation into the same incident timeline as replay/errors/performance | D2 | Fast-follow — requires a security-event ingestion path that doesn't yet exist in parity scope |

**MVP wedge is deliberately narrow: the audit trail and the access-control system underneath it.** This is the one wedge capability with no dependency on anything else on this list, and per [competitive-analysis.md](../00-overview/competitive-analysis.md) it's also the most uncontested gap — proven in PAM tooling, absent from every frontend product reviewed. Retention-clock policy, evidence export, and security correlation are real requirements, not deferred indefinitely, but they're sequenced after MVP because each depends on the audit trail existing first.

---

## Success Metrics

### Activation (Chidi's side — parity has to actually land)

- **Time-to-first-value target: under 1 hour** from deployment to first captured session + first grouped error, matching the benchmark that leading B2B PLG tools hit under-1-hour time-to-value, against a general SaaS average closer to 1.5 days.
- **Activation rate target: meaningfully above the 37.5% average** reported across 62 B2B SaaS companies (the median company loses roughly two-thirds of signups before they ever reach core value) — because Imora's actual buyers (Personas 1–3) are technical evaluators running a structured POC, not self-serve signups who churn silently.

### Proof-of-Concept success (Dara/Adaeze/Marcus's side — the wedge has to survive a real evaluation)

Enterprise security/compliance POCs typically run 2–4 weeks for standard evaluations, extending to 6–12 weeks when compliance review is part of the scope — which it will be for every Imora POC by definition. Applying the standard POC discipline (success criteria defined and shared before the first demo, each criterion tied to a number/format/binary condition, at least half of test sessions run without the vendor present) gives concrete, falsifiable MVP success criteria:

- A named compliance stakeholder (DPO or HIPAA Security Officer persona) can, unassisted, pull an access-audit-trail report for a specific session and correctly identify every internal viewer, within the POC window.
- A named engineer (Priya/Jon/Chidi persona) can deploy a working single-machine instance and capture a first session without vendor support present — directly testing the Operational Simplicity principle from [vision.md](../00-overview/vision.md).

### North Star Metric

**Weekly Sessions with Full Audit Coverage** — the count of frontend sessions per week that are simultaneously (a) captured with default-safe masking and (b) have a complete, queryable access-audit-trail. This is deliberately not "sessions captured" alone (that's parity, and Sentry/LogRocket already win on volume) and not "audit reports generated" alone (that's a compliance-only metric Chidi never touches) — it's the one number that only goes up when parity and wedge are both working on the same session at the same time, which is the actual product thesis.

---

## Open Questions and Risks

- **Scope creep risk on the wedge.** "Evidence export" and "retention policy" are exactly the kind of requirements that expand into a full GRC product if unconstrained (see Non-Goals). Every wedge feature request should be tested against: does this apply specifically to frontend telemetry Imora holds, or is this actually Adaeze/Marcus's broader compliance-program problem that belongs in a different tool?
- **Operational Simplicity vs. the audit-trail system.** Priya's requirement (P1) is a 2–3 person deployable instance. An access-control-and-audit-trail system is real additional operational surface area (more state, more to back up, more to reason about in a security review). This tension needs to be resolved in `docs/04-architecture/overview.md`, not assumed away here.
- **Licensing model** — *resolved since this was first flagged:* AGPLv3 for the full product, wedge included, with no `ee/`-style commercial split, per [licensing.md](licensing.md) and ADR [0001](../10-engineering/architecture-decisions/0001-agplv3-licensing.md). Kept in this list rather than deleted because the original concern explains *why* the license was a PRD prerequisite at all: a source-available license with SaaS-only premium features (the Sentry model) would have undermined the parity claim that a regulated org can run the full stack, control plane included.

---

## What This Feeds Next

The MVP/fast-follow split above is a scope decision, not a schedule — `docs/01-product/feature-roadmap.md` should turn it into actual milestones with the licensing question in `docs/01-product/licensing.md` resolved first, since it constrains what "self-hosted" is even allowed to mean.
