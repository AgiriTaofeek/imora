# System Context

> Status: Research-based, current as of July 2026. C4 model Level 1 — the highest, most abstract view: Imora as a single box, its people, and the external systems it touches. Per C4 convention this is deliberately non-technical (no protocols, no service names — those are [bounded-contexts.md](../02-domain/bounded-contexts.md) and `container-diagrams.md`'s job) and limited to the 2–4 actors and 2–4 external systems that actually matter, not an exhaustive list.

---

## The System

**Imora** — one box. Everything inside it (the eight bounded contexts, the domain entities) is out of scope for this document by design; a system context diagram that shows internal detail has stopped being a system context diagram.

---

## People

Per C4 convention, actors are the human roles that directly interact with the system — which produces one modeling decision worth stating explicitly:

| Actor | Interacts via | What they do |
|---|---|---|
| **Engineer** (Chidi, Jon) | dashboard | Investigates sessions, errors, performance regressions, and correlated incidents. |
| **Compliance Officer** (Adaeze, Marcus) | dashboard, query-api | Queries the access-audit-trail, applies/lifts legal holds, generates evidence exports, resolves DSAR/erasure requests per [business-rules.md](../02-domain/business-rules.md). |
| **Platform Operator** (Priya) | deployment tooling, admin surface | Deploys, configures, and operates the self-hosted instance — the actor evaluating everything in `deployment-model.md` and `scaling.md`. |
| **Data Subject** (the end user whose browsing is captured) | *does not interact with Imora directly* | Uses the regulated organization's own web application, which happens to be instrumented. This person has typically never heard of Imora. Modeling them as a normal actor with an arrow "into" the system would be a mistake — the only path by which they affect the system is indirectly, through a DSAR the Compliance Officer resolves on their behalf. This distinction is precisely why Adaeze's role exists at all. |

Four actors, not five, per the 2–4 guidance — Dara (CISO) is deliberately not a separate row here: her interaction with the system (reviewing evidence exports, board-level reporting) is a subset of the Compliance Officer's queries, not a distinct usage pattern at this level of abstraction.

---

## External Systems

| System | Required or optional | Relationship |
|---|---|---|
| **The organization's own web/mobile application** | Required | Embeds browser-sdk. This is the source of every SessionEvent, ErrorEvent, and PerformanceMetric — there is no Imora deployment without one of these. |
| **The organization's backend services** | Optional | Correlated via TraceLink (story J1) when instrumented with a compatible session/trace identifier. Per [prd.md](../01-product/prd.md)'s Non-Goals, Imora does not own or replace backend tracing — it only consumes a correlation ID from a system it doesn't manage. |
| **Identity Provider (SSO/SAML)** | Optional, Enterprise-tier | Per [pricing.md](../01-product/pricing.md) and [licensing.md](../01-product/licensing.md), enterprise auth integration is a legitimate commercial add-on — never required to use the core (parity + wedge) product. |
| **Notification channels (email, Slack, webhook endpoints)** | Optional | notification-service delivers AlertTriggered outcomes here, per [event-catalog.md](../02-domain/event-catalog.md). |

---

## The Finding This Document Exists to Surface: Two Context Variants, Not One

A standard system context diagram assumes one topology. Imora can't — [vision.md](../00-overview/vision.md)'s Operational Simplicity principle explicitly commits to "fully air-gapped environments with no outbound dependency for core function," which means the *set of external systems present* is itself a deployment-mode decision, not a constant:

- **Connected deployment:** all four external systems above may be present. SSO simplifies Platform Operator/Engineer login; notifications reach Slack/email; backend TraceLink correlation is live.
- **Air-gapped deployment:** only the organization's own web application is present. No SSO, no outbound notifications, no external backend correlation (TraceLink correlation is simply inactive if nothing is configured to receive it) — **and every Parity and Wedge capability must still work at full strength.** An access-audit-trail that degrades, or a legal-hold check that silently no-ops, because an air-gapped instance can't reach an external system would directly contradict Dara's and Adaeze's core requirement in [user-personas.md](../01-product/user-personas.md).

The practical consequence for every document downstream of this one: nothing in `container-diagrams.md`, `component-diagrams.md`, or `deployment-model.md` may place a Parity or Wedge capability behind a call to an external system in the required-path. External systems are additive convenience (Milestone 3, per [feature-roadmap.md](../01-product/feature-roadmap.md)), never load-bearing for the core product — a fact that was implicit across five prior documents but had not been stated as an explicit architectural constraint until this one.

---

## Interaction Summary

| Actor / System | Action | Toward |
|---|---|---|
| Data Subject | Browses the instrumented application (unaware of Imora) | Organization's web application |
| Organization's web application | Streams masked SessionEvents, ErrorEvents, PerformanceMetrics | Imora |
| Engineer | Investigates sessions, errors, and regressions | Imora |
| Compliance Officer | Queries audit trail; applies legal holds; generates evidence exports | Imora |
| Platform Operator | Deploys, configures, and operates | Imora |
| Imora | Correlates via TraceLink (if configured) | Organization's backend services |
| Imora | Authenticates via (if configured) | Identity Provider |
| Imora | Delivers alerts (if configured) | Notification channels |

---

## What's Deliberately Not Modeled Here

- Internal service boundaries — that's [bounded-contexts.md](../02-domain/bounded-contexts.md), already done, and `container-diagrams.md`, next.
- Protocols, ports, or data formats for any interaction above — that's `07-api/` and `container-diagrams.md`.
- Deployment topology detail (single-machine vs. cluster vs. air-gapped specifics) — that's `deployment-model.md`.

---

Sources: [System context diagram — C4 model](https://c4model.com/diagrams/system-context), [C4 System Context Diagram: Beginner's Guide](https://skills.visual-paradigm.com/docs/from-zero-to-c4-beginner-modeling-blueprint/mastering-the-four-levels-of-c4/c4-system-context-diagram-beginner/).

## What This Feeds Next

`docs/04-architecture/container-diagrams.md` is the direct next step — it takes the single "Imora" box here and expands it into the eight bounded contexts from [bounded-contexts.md](../02-domain/bounded-contexts.md), now with the two-variant (connected vs. air-gapped) constraint from this document applied to each one.
