# Repository Structure

> Status: Research-based, current as of July 2026. The last file in `04-architecture/` — translates the eight bounded contexts from [bounded-contexts.md](../02-domain/bounded-contexts.md) into an actual codebase layout, constrained by [licensing.md](../01-product/licensing.md)'s rule that no wedge capability may ever be split into a separately-licensed directory.

---

## Monorepo, Not Polyrepo

The three closest architectural comparators (PostHog, OpenReplay, and Sentry) are all monorepos, but that's supporting evidence, not the actual reason. The real reason is [bounded-contexts.md](../02-domain/bounded-contexts.md)'s **Shared Kernel** relationships: `browser-sdk`↔`ingestion` and `ingestion`↔`query-api` were deliberately modeled as operating on the *same* entity definitions from [domain-model.md](../02-domain/domain-model.md), not translated copies. A polyrepo split would force those definitions to live in separately-versioned packages published to a registry, reintroducing exactly the drift risk Shared Kernel was chosen to prevent — the general research tradeoff (monorepo: low coordination cost, higher blast radius; polyrepo: the reverse) resolves in monorepo's favor specifically because coordination cost on shared types is the risk that matters most here, not blast radius.

---

## The Finding This Document Exists to Act On

Checking how the closest comparator actually organizes its monorepo surfaced a direct warning, not just a template to copy: **PostHog-FOSS exists as "a read-only mirror of PostHog, with all proprietary code removed."** That means PostHog's own monorepo contains a directory boundary (`ee/`) that *is* the license-gating anti-pattern [licensing.md](../01-product/licensing.md) already ruled out — a monorepo doesn't prevent that pattern, it's literally how PostHog implements it. Imora's layout has to be designed so there is no natural place to put a commercially-gated directory at all, not just an unused one.

**Concretely: one `LICENSE` file at the repository root, applying uniformly to every directory. No per-directory or per-package license overrides, anywhere, ever.** If a future contributor proposes an `enterprise/` or `ee/` directory, that proposal contradicts this document and [licensing.md](../01-product/licensing.md) both, and needs to be resolved by revisiting those documents, not by adding an exception here.

---

## Layout

```
imora/
├── LICENSE                     # AGPLv3, root-only, applies to everything below
├── services/                   # the six backend bounded contexts (gateway, ingestion, query-api,
│   │                           # alert-engine, workers, notification-service) — one directory each,
│   │                           # all under the same root license, no directory is "more open" than another
│   ├── gateway/
│   ├── ingestion/
│   ├── query-api/
│   ├── alert-engine/
│   ├── workers/
│   └── notification-service/
├── sdk/
│   └── browser-sdk/            # independently versioned (see below) — still root-licensed
├── dashboard/                  # frontend SPA, consumes query-api only per bounded-contexts.md's
│                                # Conformist relationship
├── packages/                   # the Shared Kernel, made literal
│   ├── domain-types/           # Session, SessionEvent, AccessAuditEvent, and every entity from
│   │                           # domain-model.md — the single source browser-sdk, ingestion, and
│   │                           # query-api all import directly, per their Shared Kernel relationship
│   └── event-schemas/          # the event-catalog.md payload shapes, shared by producers and consumers
├── deploy/
│   ├── compose/                # single-machine profile manifests, per deployment-model.md
│   └── kubernetes/             # cluster profile manifests
├── docs/                       # this documentation tree
└── tools/                      # dev/build tooling, not shipped
```

Milestone 3 commercial features (SSO/SAML, multi-region orchestration tooling, per [feature-roadmap.md](../01-product/feature-roadmap.md) and [pricing.md](../01-product/pricing.md)) live **inside** `services/gateway/` and `deploy/kubernetes/` respectively, gated at runtime by the offline signed license file from [pricing.md](../01-product/pricing.md) — not in a separate directory, and not under a separate license. The gate is a feature flag checked at startup, never a build-time or repository-structure-level split. This is the concrete, structural answer to the finding above.

---

## One Deliberate Exception: browser-sdk's Independent Release Cadence

`browser-sdk` ships as an npm package that customer applications import directly — its release cadence (semver, changelogs, compatibility guarantees to external code) is necessarily decoupled from the backend services' internal deploy cadence, even though it lives in the same repository and shares `packages/domain-types` with them. This is standard monorepo practice for public SDKs (independent package versioning within a shared repository, not a shared release train) and doesn't reopen the licensing question above — it's a *versioning* exception, not a *licensing* one; `sdk/browser-sdk/` still has no `LICENSE` file of its own.

---

## What's Deliberately Not Modeled Here

- Build tooling choice (Nx, Turborepo, Bazel, or a simpler approach) — an implementation detail downstream of this layout decision, not an architectural one.
- CI/CD pipeline structure for this layout — `09-infrastructure/ci-cd.md`.
- Actual package/module names or internal directory structure within any single service — that's each service's own concern once `05-services/*.md` is filled in.

---

Sources: [PostHog monorepo layout](https://github.com/PostHog/posthog/blob/master/docs/internal/monorepo-layout.md), [PostHog-FOSS — read-only mirror with proprietary code removed](https://github.com/PostHog/posthog-foss), [Monorepo vs. Polyrepo (Multi-repo): What's the Difference? — Spacelift](https://spacelift.io/blog/monorepo-vs-polyrepo), [Monorepo vs. Polyrepo for Microservices: Decision Framework](https://www.developers.dev/tech-talk/monorepo-vs-polyrepo-an-engineering-decision-framework-for-microservices-at-scale.html).

## What This Closes Out

This is the last file in `docs/04-architecture/`. All eight files — [system-context.md](system-context.md), [container-diagrams.md](container-diagrams.md), [component-diagrams.md](component-diagrams.md), [sequence-diagrams.md](sequence-diagrams.md), [deployment-model.md](deployment-model.md), [scaling.md](scaling.md), [overview.md](overview.md), and this one — are now internally consistent and cross-referenced, alongside `00-overview/`, `01-product/`, and `02-domain/`. `06-data/` (schemas for the ClickHouse/Postgres split this document assumes) and `08-security/` (the SecureFieldVault mechanism from [component-diagrams.md](component-diagrams.md)) are the natural next tiers.
