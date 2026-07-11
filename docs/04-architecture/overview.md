# Architecture Overview

> Status: Synthesizes [system-context.md](system-context.md) through [scaling.md](scaling.md) into the single narrative entry point this folder otherwise lacks. Read this first; go to the individual documents for the detail behind any claim here.

---

## The Shape of the Architecture, in One Paragraph

Imora is one system ([system-context.md](system-context.md)) built from eight bounded contexts ([bounded-contexts.md](../02-domain/bounded-contexts.md), made concrete in [container-diagrams.md](container-diagrams.md)) split cleanly along a write path (`browser-sdk` → `gateway` → `ingestion`) and a read path (`query-api` → `dashboard`), with a background context (`workers`) that owns the compliance-critical jobs — retention sweeps, legal-hold enforcement, evidence export — that shouldn't run inline with a user-facing request. Two of those eight containers, `query-api` and `workers`, carry almost all the actual business-rule weight and get their internal structure specified in [component-diagrams.md](component-diagrams.md); the other six are comparatively thin at that altitude. The whole system deploys in one of two topology profiles ([deployment-model.md](deployment-model.md)) — a single Docker Compose host or a Kubernetes cluster — with air-gapping as an orthogonal setting on either profile, not a third variant.

---

## The Core Architectural Bet

Per [vision.md](../00-overview/vision.md), Imora's entire pitch rests on the wedge (access-audit-trail, regulatory-clock retention, evidence export) being real guarantees, not documented intentions. Every document in `04-architecture/` exists to make that bet structurally true rather than procedurally true:

- **AccessAuditEvent generation is architecturally impossible to skip**, not a convention — `query-api`'s `AuditedQueryHandler` is the only way to register a read route at all ([component-diagrams.md](component-diagrams.md)).
- **The legal-hold check runs immediately before every individual deletion**, not once per batch — closing the exact race condition a less careful implementation would reintroduce ([sequence-diagrams.md](sequence-diagrams.md) Flow C).
- **An EvidenceExport is a frozen copy at generation time**, immune by construction to anything that happens to its source data afterward ([sequence-diagrams.md](sequence-diagrams.md) Flow D).
- **Backup RPO for AccessAuditEvent is a compliance requirement, not an ops nicety** — a lost audit record is indistinguishable from a HIPAA §164.312(b) failure to an assessor ([deployment-model.md](deployment-model.md)).

None of these were achievable by writing "must be audited" in a requirements document — each required a specific structural decision at the architecture layer, which is the actual justification for this folder existing as more than a formality.

---

## The Five Findings Worth Remembering

If nothing else from this folder is read, these five are the load-bearing ones:

1. **Two system-context variants, not one** — air-gapped deployments must clear every Parity and Wedge bar with zero external systems present; SSO, notifications, and backend correlation are all optional convenience, never load-bearing ([system-context.md](system-context.md)).
2. **The audit-log guarantee had to be made structural** — a wrapper type that's the only registration path for a read route, not a function callers remember to invoke ([component-diagrams.md](component-diagrams.md)).
3. **Masking is two mechanisms, not one** — hard redaction (unrecognized fields, never captured, never unmaskable) and soft masking with audited escalation (known PHI/PII fields, captured into a vault) — resolving an apparent contradiction between [domain-model.md](../02-domain/domain-model.md)'s capture-time invariant and story M2's unmask requirement.
4. **Air-gapped updates reuse the license-activation pattern** — the same signed-bundle-via-removable-media process from [pricing.md](../01-product/pricing.md) solves software updates too, not a second procedure ([deployment-model.md](deployment-model.md)).
5. **The scaling trigger is retention-driven storage, not throughput** — a genuinely counter-intuitive result for an observability product, and specific to Imora's multi-year regulatory retention obligations that no SaaS-only competitor has to plan around ([scaling.md](scaling.md)).

---

## Reading Order for Someone New to This Folder

1. This document, for orientation.
2. [system-context.md](system-context.md) — who and what touches Imora from outside.
3. [container-diagrams.md](container-diagrams.md) — the eight services and their data stores.
4. [component-diagrams.md](component-diagrams.md) — inside `query-api` and `workers` specifically.
5. [sequence-diagrams.md](sequence-diagrams.md) — four flows traced through all of the above.
6. [deployment-model.md](deployment-model.md) and [scaling.md](scaling.md) — where it runs, and when that stops being enough.

Everything here assumes [domain-model.md](../02-domain/domain-model.md), [bounded-contexts.md](../02-domain/bounded-contexts.md), [business-rules.md](../02-domain/business-rules.md), and [event-catalog.md](../02-domain/event-catalog.md) from `02-domain/` as settled — this folder is what those become once they have to run somewhere.

---

## What's Not Yet Covered

This folder specifies structure and behavior; it deliberately stops short of:

- Actual field-level schemas — `06-data/`.
- Wire protocols and public API shape — `07-api/`.
- The SecureFieldVault's encryption mechanism and other security implementation detail — `08-security/`.
- Actual Compose/Kubernetes manifests and CI/CD — `09-infrastructure/`.
- How the eight bounded contexts map to actual code organization (monorepo vs. polyrepo, folder layout) — [repository-structure.md](repository-structure.md), now written; monorepo with a single root license, per ADR [0003](../10-engineering/architecture-decisions/0003-monorepo-structure.md).

---

## What This Feeds Next

`docs/04-architecture/repository-structure.md` closes out this folder — the last translation step, from bounded contexts to an actual codebase layout. After that, `06-data/` and `08-security/` are the two folders this document leans on most heavily (the SecureFieldVault and the ClickHouse/Postgres schemas), and are the natural next tier.
