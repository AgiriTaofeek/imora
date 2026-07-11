# Documentation Framework

This `docs/` tree is a general-purpose framework for planning and documenting a software project end to end — from why it should exist through how it runs in production. It was built out for **Imora** (a self-hosted frontend observability platform), so the content inside each file is Imora-specific, but the 12-folder structure itself is meant to be reused as-is for other projects. Copy the skeleton, keep the folder purposes below, replace the content.

---

## Start Here

**What Imora is, in three sentences:** a self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory, built for regulated industries (banks, insurers, hospitals, government) that can't send customer session data to a third party. It aims to match those tools on ordinary debugging (error tracking, session replay, performance monitoring) and then win the deal on three things none of them do: an audit trail of who on your team viewed a customer's session, data retention mapped to actual regulatory clocks instead of one global setting, and one-click evidence export for a regulator or auditor. The shorthand used everywhere in these docs is **Parity** (what it takes to be a credible alternative at all) and **Wedge** (the three things above, the actual reason to pick it over any other alternative) — full definition in [00-overview/glossary.md](00-overview/glossary.md).

**If you have 10 minutes**, read these four, in order, and skip everything else for now:

1. [00-overview/vision.md](00-overview/vision.md) — the Positioning section only; you can stop once you've read that.
2. [00-overview/glossary.md](00-overview/glossary.md) — skim it once so Parity/Wedge/BR-1 through BR-7/AccessAuditEvent aren't unfamiliar shorthand the first time they appear elsewhere.
3. [00-overview/target-users.md](00-overview/target-users.md) — who this is for, in six roles.
4. [01-product/prd.md](01-product/prd.md) — Goals and Non-Goals sections only.

**If you have an hour**, add [02-domain/domain-model.md](02-domain/domain-model.md), [04-architecture/overview.md](04-architecture/overview.md) (now with rendered diagrams in the four files it summarizes), and [08-security/threat-model.md](08-security/threat-model.md) — that last one doubles as a fast tour of everything the rest of `08-security/` established, since it stress-tests all of it in one place.

Everything past that point is depth, not orientation — come back to it when you need the specific answer, not before.

---

## The Framework

**How to read this table:** *Scope* tells you whether a folder is foundational for any software project or only makes sense once the project has grown a certain shape. *Adapt per project* flags the places where this instance's filenames encode a specific decision (a database, a deployment tool) rather than a generic category — rename or replace those, don't treat them as required filenames. *Status* is Imora-specific — how much of this folder is actually written for this project, not part of the reusable framework itself.

| Folder | Purpose | Scope | Adapt per project | Status |
|---|---|---|---|---|
| `00-overview/` | Why the project exists: vision, the problem being solved, competitive landscape, who it's for, and the project-wide glossary. | **Always.** Every project needs this, even a one-page version. | Nothing — this folder's shape is generic. | Complete |
| `01-product/` | What's being built and why it's scoped that way: requirements, personas, user stories, roadmap, and (if relevant) pricing/licensing. | **Always**, though pricing/licensing only apply to products with a commercial model — delete those two files for an internal tool or a pure open-source library with no dual-licensing question. | Nothing structural. | Complete |
| `02-domain/` | The business logic modeled formally: core entities, service/module boundaries, business rules, and the event vocabulary. | **Scales with complexity.** A CRUD app or a simple script doesn't need five separate files here — a single lightweight `domain-model.md` may be enough. A project with real regulatory, financial, or multi-party business logic (like Imora) benefits from the full breakdown. | Nothing structural — the DDD-flavored file set (domain-model, bounded-contexts, business-rules, event-catalog) applies to any moderately complex domain, not just observability. | Complete |
| `03-workflows/` | Step-by-step user-facing journeys through the actual product — one file per key workflow. | **Always for products with users**; skip for pure infrastructure/libraries with no end-user journey. | The individual filenames (`onboarding.md`, `session-replay.md`, etc.) are Imora's actual features — rename them to your own project's workflows. | Complete |
| `04-architecture/` | How the system is built: system context, container/component diagrams, sequence diagrams, deployment model, scaling triggers, and repository layout. | **Always**, though depth scales with system complexity — a small project might only need `overview.md` and `system-context.md`. | Nothing structural — the C4-model-derived shape (context → container → component) applies to any architecture style. | Complete — diagrams rendered inline as Mermaid, see [diagrams/README.md](../diagrams/README.md) |
| `05-services/` | One file per independently deployable/buildable unit, with its specific responsibilities and technology shape. | **Only if the project actually has multiple deployable units.** Skip entirely for a monolith, a single library, or a single-binary tool. | The 8 filenames here are Imora's actual services (gateway, ingestion, etc.) — for your project, this might be `apps/`, `modules/`, or not exist at all. | **Deferred.** Container/component-level detail already lives in `04-architecture/`; per-service implementation detail (config, env vars, runbooks) is intentionally left until real code exists to document, rather than invented ahead of it. |
| `06-data/` | Schemas, event formats, retention/lifecycle policy, and storage layout for whatever data the system persists. | **Always for anything stateful.** | The filenames `clickhouse-schema.md` / `postgres-schema.md` are Imora's actual datastore choices — rename to match your own stack (`mongodb-schema.md`, `sqlite-schema.md`, or just `schema.md` for a single store). | Complete |
| `07-api/` | The contract the system exposes — REST/SDK/webhook surface and any formal spec files (OpenAPI, etc.). | **Only if the project exposes an API.** Skip for an internal-only tool with no external interface. | Nothing structural beyond the datastore point above — add or remove files per which interfaces your project actually has. | Complete |
| `08-security/` | Authentication, authorization, encryption, audit logging, data protection, and threat modeling. | **Always**, though depth scales with sensitivity — a low-stakes internal tool needs a lighter version than a system handling regulated data. | Nothing structural — every project has *some* answer to each of these six questions, even if the answer is short. | Complete |
| `09-infrastructure/` | How the system is deployed and operated: environments, CI/CD, observability of the system itself (not to be confused with `03-workflows/` if your product's *feature* is observability, as Imora's is). | **Always**, depth scales with deployment complexity. | The filenames `docker.md` / `compose.md` / `kubernetes.md` are Imora's actual deployment tooling — rename or consolidate per your own stack (a serverless project might just need one `environments.md`). | Complete |
| `10-engineering/` | Team conventions: branching, coding standards, release process, testing strategy, and an `architecture-decisions/` subfolder for ADRs (one file per significant decision, e.g. `0001-use-clickhouse-for-events.md`). | **Always** — even a solo project benefits from writing these down, if only for future-you. | Nothing structural — the ADR pattern in particular is broadly reusable and worth keeping on every project. | Complete — 7 ADRs recorded |
| `11-design/` | Design system, wireframes, and interaction patterns. | **Only for projects with a UI.** Skip entirely for a CLI tool, a library, or a backend-only service. | Nothing structural. | **Partial.** `interaction-patterns.md` is done (behavior is text-tractable). `design-system.md` and `dashboard-wireframes.md` are still stubs — visual layout and token design are genuinely better served by an actual design tool than prose; forcing a text substitute would produce something worse than useful. |

---

## Suggested Fill-In Order

Not the same as the folder numbering — numbering reflects typical *reading* order for someone new to the project, not the order you should *write* these in:

1. `00-overview/` and `01-product/` first — you can't design a system before agreeing on what problem it solves and for whom.
2. `02-domain/` next — model the business logic before designing the architecture that will implement it.
3. `04-architecture/` and `05-services/` — now that the domain is settled, design how it's built.
4. `06-data/`, `07-api/`, `08-security/` — these follow directly from the architecture decisions above.
5. `03-workflows/`, `09-infrastructure/`, `10-engineering/`, `11-design/` — round these out as the system takes shape; they don't block each other and can be written in parallel.

**A note specific to reading (not writing) `02-domain/` through `09-infrastructure/`:** a lot of files past this point open with something like "The Finding" or "The Tension This Document Has to Resolve" — that's deliberate; each one is closing a gap or resolving a conflict between two earlier decisions, not introducing a topic from scratch. Reading them out of order means hitting the middle of an argument. If a file's opening doesn't make sense, the fix is almost always to read the specific document it links to first (usually [business-rules.md](02-domain/business-rules.md) or [domain-model.md](02-domain/domain-model.md)), not to push through — these documents assume you've read what they cite, the same way they'd expect a reviewer to actually follow the links rather than take the claim on faith.

## Conventions Worth Keeping

- **A one-line status header at the top of every file** (`> Status: Draft` or `> Status: Research-based, current as of <date>`) — makes it immediately obvious which docs are aspirational scaffolding versus decisions that have actually been made.
- **Every claim traces to something** — a source link, a prior document, or explicit reasoning — rather than being asserted. Docs that read as authoritative but aren't sourced are worse than no docs at all, because they get trusted anyway.
- **Cross-reference relentlessly.** Every file in this tree links to the specific prior documents its claims depend on, so a reader (or a future editor) can verify a decision instead of taking it on faith. When you copy this framework to a new project, keep that habit even more than you keep the folder names.
- **Diagrams live inline as Mermaid, not as exported images referenced from elsewhere.** They render in place, stay in sync with the prose they illustrate because they're edited in the same diff, and are actually diffable in review. See [diagrams/README.md](../diagrams/README.md) for the full reasoning.
