# Documentation Framework

This `docs/` tree is a general-purpose framework for planning and documenting a software project end to end — from why it should exist through how it runs in production. It was built out for **Imora** (a self-hosted frontend observability platform), so the content inside each file is Imora-specific, but the 13-folder structure itself is meant to be reused as-is for other projects.

**Folders are numbered in the order a team actually moves through planning and building a project, not by topic category:**

1. **Vision & Problem** (`00-overview/`) — why this should exist, and for whom.
2. **PRD** (`01-product/`) — what's being built, scoped and justified.
3. **Architecture & Design** (`02-domain/` through `07-security/`) — the business logic, system design, services, data, API, and security that follow from the PRD.
4. **Roadmap** (`08-roadmap/`) — how that design gets sequenced into shippable milestones.
5. **Feature Specs** (`09-workflows/`, `10-design/`) — the per-feature/per-journey detail a milestone actually needs before implementation starts.
6. **Implementation** (`11-engineering/`) — the team conventions and recorded decisions that govern how code actually gets written.
7. **Testing, Rollout & Observe** (`12-infrastructure/`) — how it's deployed, released, and watched once it's running.

Each folder holds a single `README.md` covering its whole topic, with one `##` section per sub-topic (e.g. `07-security/README.md` has one section each for Authentication, Authorization, Encryption, Audit Logging, PII Redaction, and Threat Model) — not one file per sub-topic. `03-architecture/` is the one exception, split into `README.md` (prose: overview, system context, deployment model, scaling, repository structure) and `diagrams.md` (the three Mermaid-heavy diagram sections), since combining all eight into one file made the diagrams hard to find. `11-engineering/architecture-decisions/` also stays one-file-per-decision — ADRs are meant to be an immutable log, not sections in a living document.

Copy the skeleton, keep the folder purposes below, replace the content.

---

## Start Here

**What Imora is, in three sentences:** a self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory, built for regulated industries (banks, insurers, hospitals, government) that can't send customer session data to a third party. It aims to match those tools on ordinary debugging (error tracking, session replay, performance monitoring) and then win the deal on three things none of them do: an audit trail of who on your team viewed a customer's session, data retention mapped to actual regulatory clocks instead of one global setting, and one-click evidence export for a regulator or auditor. The shorthand used everywhere in these docs is **Parity** (what it takes to be a credible alternative at all) and **Wedge** (the three things above, the actual reason to pick it over any other alternative) — full definition in the Glossary section of [00-overview/README.md](00-overview/README.md#glossary).

**If you have 10 minutes**, read these four sections, in order, and skip everything else for now:

1. [00-overview/README.md § Vision](00-overview/README.md#vision) — the Positioning section only; you can stop once you've read that.
2. [00-overview/README.md § Glossary](00-overview/README.md#glossary) — skim it once so Parity/Wedge/BR-1 through BR-7/AccessAuditEvent aren't unfamiliar shorthand the first time they appear elsewhere.
3. [00-overview/README.md § Target Users](00-overview/README.md#target-users) — who this is for, in six roles.
4. [01-product/README.md § PRD](01-product/README.md#product-requirements-document-prd) — Goals and Non-Goals sections only.

**If you have an hour**, add [02-domain/README.md § Domain Model](02-domain/README.md#domain-model), [03-architecture/README.md § Architecture Overview](03-architecture/README.md#architecture-overview) (now with rendered diagrams in `03-architecture/diagrams.md`), and [07-security/README.md § Threat Model](07-security/README.md#threat-model) — that last one doubles as a fast tour of everything the rest of `07-security/README.md` established, since it stress-tests all of it in one place.

Everything past that point is depth, not orientation — come back to it when you need the specific answer, not before.

---

## The Framework

**How to read this table:** *Stage* is the planning-flow stage the folder belongs to (see the numbered list above) — several folders can share a stage when that stage naturally produces more than one kind of document. *Scope* tells you whether a folder is foundational for any software project or only makes sense once the project has grown a certain shape. *Adapt per project* flags the places where this instance's content encodes a specific decision (a database, a deployment tool) rather than a generic category — rename or replace those, don't treat them as required section names. *Status* is Imora-specific — how much of this folder is actually written for this project, not part of the reusable framework itself.

| Folder | Stage | Purpose | Scope | Adapt per project | Status |
|---|---|---|---|---|---|
| `00-overview/` | Vision & Problem | Why the project exists: vision, the problem being solved, competitive landscape, who it's for, and the project-wide glossary. | **Always.** Every project needs this, even a one-page version. | Nothing — this folder's shape is generic. | Complete |
| `01-product/` | PRD | What's being built and why it's scoped that way: requirements, personas, user stories, and (if relevant) pricing/licensing. | **Always**, though pricing/licensing only apply to products with a commercial model — drop those sections for an internal tool or a pure open-source library with no dual-licensing question. | Nothing structural. | Complete |
| `02-domain/` | Architecture & Design | The business logic modeled formally: core entities, service/module boundaries, business rules, and the event vocabulary. Foundational input to the architecture that follows. | **Scales with complexity.** A CRUD app or a simple script doesn't need four separate sections here — a single lightweight domain-model section may be enough. A project with real regulatory, financial, or multi-party business logic (like Imora) benefits from the full breakdown. | Nothing structural — the DDD-flavored section set (domain model, bounded contexts, business rules, event catalog) applies to any moderately complex domain, not just observability. | Complete |
| `03-architecture/` | Architecture & Design | How the system is built: system context, container/component diagrams, sequence diagrams, deployment model, scaling triggers, and repository layout. | **Always**, though depth scales with system complexity — a small project might only need the Overview and System Context sections. | Nothing structural — the C4-model-derived shape (context → container → component) applies to any architecture style. | Complete — diagrams rendered inline as Mermaid, see [diagrams/README.md](../diagrams/README.md) |
| `04-services/` | Architecture & Design | One section per independently deployable/buildable unit, with its specific responsibilities and technology shape. | **Only if the project actually has multiple deployable units.** Skip entirely for a monolith, a single library, or a single-binary tool. | The 8 sections here are Imora's actual services (gateway, ingestion, etc.) — for your project, this might be `apps/`, `modules/`, or not exist at all. | **Deferred.** Container/component-level detail already lives in `03-architecture/`; per-service implementation detail (config, env vars, runbooks) is intentionally left until real code exists to document, rather than invented ahead of it. |
| `05-data/` | Architecture & Design | Schemas, event formats, retention/lifecycle policy, and storage layout for whatever data the system persists. | **Always for anything stateful.** | The ClickHouse/Postgres schema sections are Imora's actual datastore choices — rename to match your own stack (a MongoDB or SQLite section, or just one generic "Schema" section for a single store). | Complete |
| `06-api/` | Architecture & Design | The contract the system exposes — REST/SDK/webhook surface and any formal spec files (OpenAPI, etc.). | **Only if the project exposes an API.** Skip for an internal-only tool with no external interface. | Nothing structural beyond the datastore point above — add or remove sections per which interfaces your project actually has. | Complete |
| `07-security/` | Architecture & Design | Authentication, authorization, encryption, audit logging, data protection, and threat modeling. | **Always**, though depth scales with sensitivity — a low-stakes internal tool needs a lighter version than a system handling regulated data. | Nothing structural — every project has *some* answer to each of these six questions, even if the answer is short. | Complete |
| `08-roadmap/` | Roadmap | How the PRD's scope and the architecture's constraints get sequenced into shippable milestones, with exit criteria per milestone. | **Always**, though a solo/small project may need only one milestone. | Nothing structural — the milestone/thesis/exit-criteria shape is generic. Task-level breakdown for a milestone belongs in a project-management tool, not a new docs file. | Complete — 3 milestones |
| `09-workflows/` | Feature Specs | Step-by-step user-facing journeys through the actual product — one section per key workflow. | **Always for products with users**; skip for pure infrastructure/libraries with no end-user journey. | The individual sections (Onboarding, Session Replay, etc.) are Imora's actual features — rename them to your own project's workflows. | Complete |
| `10-design/` | Feature Specs | Design system, wireframes, and interaction patterns. | **Only for projects with a UI.** Skip entirely for a CLI tool, a library, or a backend-only service. | Nothing structural. | **Partial.** The Interaction Patterns section is done (behavior is text-tractable). Design System and Dashboard Wireframes are still stubs — visual layout and token design are genuinely better served by an actual design tool than prose; forcing a text substitute would produce something worse than useful. |
| `11-engineering/` | Implementation | Team conventions: branching, coding standards, release process, testing strategy, and an `architecture-decisions/` subfolder for ADRs (one file per significant decision, e.g. `0001-use-clickhouse-for-events.md`). | **Always** — even a solo project benefits from writing these down, if only for future-you. | Nothing structural — the ADR pattern in particular is broadly reusable and worth keeping on every project. | Complete — 7 ADRs recorded |
| `12-infrastructure/` | Testing, Rollout & Observe | How the system is deployed and operated: environments, CI/CD, observability of the system itself (not to be confused with `09-workflows/` if your product's *feature* is observability, as Imora's is). | **Always**, depth scales with deployment complexity. | The Docker/Compose/Kubernetes sections are Imora's actual deployment tooling — rename or consolidate per your own stack (a serverless project might just need one "Environments" section). | Complete |

---

## Suggested Fill-In Order

The folder numbering now doubles as the write order — that's the point of this restructuring. Work through the stages top to bottom:

1. `00-overview/` and `01-product/` first — you can't design a system before agreeing on what problem it solves and for whom.
2. `02-domain/` next — model the business logic before designing the architecture that will implement it.
3. `03-architecture/` and `04-services/` — now that the domain is settled, design how it's built.
4. `05-data/`, `06-api/`, `07-security/` — these follow directly from the architecture decisions above.
5. `08-roadmap/` — now that the design surface is known, sequence it into milestones with real exit criteria.
6. `09-workflows/` and `10-design/` — write the per-feature spec for whatever the current milestone actually requires, just before building it.
7. `11-engineering/` and `12-infrastructure/` — team conventions and deploy/observe practices round these out; they don't block the stages above and can be written in parallel once the project has real code to operate.

**A note specific to reading (not writing) `02-domain/` through `12-infrastructure/`:** a lot of sections past this point open with something like "The Finding" or "The Tension This Document Has to Resolve" — that's deliberate; each one is closing a gap or resolving a conflict between two earlier decisions, not introducing a topic from scratch. Reading them out of order means hitting the middle of an argument. If a section's opening doesn't make sense, the fix is almost always to read the specific document it links to first (usually the [Business Rules](02-domain/README.md#business-rules) or [Domain Model](02-domain/README.md#domain-model) section of `02-domain/README.md`), not to push through — these documents assume you've read what they cite, the same way they'd expect a reviewer to actually follow the links rather than take the claim on faith.

## Conventions Worth Keeping

- **A one-line status header at the top of every file** (`> Status: Draft` or `> Status: Research-based, current as of <date>`) — makes it immediately obvious which docs are aspirational scaffolding versus decisions that have actually been made.
- **Every claim traces to something** — a source link, a prior document, or explicit reasoning — rather than being asserted. Docs that read as authoritative but aren't sourced are worse than no docs at all, because they get trusted anyway.
- **Cross-reference relentlessly.** Every section in this tree links to the specific prior sections its claims depend on, so a reader (or a future editor) can verify a decision instead of taking it on faith. When you copy this framework to a new project, keep that habit even more than you keep the folder names.
- **Diagrams live inline as Mermaid, not as exported images referenced from elsewhere.** They render in place, stay in sync with the prose they illustrate because they're edited in the same diff, and are actually diffable in review. See [diagrams/README.md](../diagrams/README.md) for the full reasoning.
