# `gateway` — Design Doc

> Status: In progress. This is the working design for `gateway` before its real implementation begins — filled in section by section, per the [no-ai-slop pre-code checklist](../../.agents/skills/no-ai-slop/SKILL.md#how-to-think-before-writing-any-code). Sections marked _TBD_ are genuinely undecided, not placeholder text to skim past.

---

## Problem / Goal

`gateway` is the single chokepoint every request passes through (per [`docs/architecture.md`](../../docs/architecture.md)) — but its internal design was never actually specified beyond that one-line responsibility. This doc exists to settle that before any handler gets written, because `gateway` turned out to be serving two structurally different callers that a single undifferentiated auth pipeline would handle badly.

## Context (settled elsewhere — linked, not restated)

- Overall responsibility, position in the write/read paths: [`docs/architecture.md` §3](../../docs/architecture.md).
- RBAC baseline (4 roles) and the one ABAC boundary (UNMASK, owned by `query-api`, not `gateway`): [`research/07-security/README.md#authorization`](../../research/07-security/README.md#authorization).
- M0 scope for auth: local auth only (Argon2id + TOTP), no SSO: [`docs/prd.md` §5](../../docs/prd.md).
- General service shape (hexagonal, repository pattern, explicit constructor DI) applies here same as every other service — not re-litigated in this doc: [`docs/design-system.md`](../../docs/design-system.md).

## Trust Boundaries & Actors

Two categorically different callers, currently both routed through the same word ("gateway") in the architecture diagrams:

| | **Domain A — Project/Ingest** | **Domain B — Human/Dashboard** |
|---|---|---|
| Caller | `browser-sdk`, embedded in a *customer's own* website, acting on behalf of anonymous visitors | A real, individually-identified human (engineer, compliance officer, admin) logged into `dashboard` |
| Credential | Project Key — public by design (ships in client-side JS), write-scoped only | Session/token from local auth — private, individually attributable, RBAC-scoped |
| Volume | Potentially very high — bounded by a customer's own website traffic | Low — bounded by how many humans a customer employs |
| Trust level | Low per-request (anyone can extract and replay the key) | High (identity matters — it's what the audit trail is *for*) |
| Downstream | `ingestion` | `query-api` (reads), plus `gateway`-owned actions (project/user management) |

Design stance: one physical `gateway` process, both domains enter through it (matches the existing architecture diagrams), but internal auth logic branches hard by which credential type a request carries — a cheap/fast path for Domain A, a full session+RBAC path for Domain B. See [Data Flow](#data-flow) below.

## Data Flow

_TBD — needs real sequence diagrams (Mermaid, matching the convention already used in [`research/03-architecture/diagrams.md`](../../research/03-architecture/diagrams.md)) for both:_
- _Domain A: SDK event → gateway fast-path validation → ingestion_
- _Domain B: dashboard login → session established → subsequent authenticated request → query-api, with actor context attached_

## Decisions

Each of these is a real, contested, not-yet-resolved decision — tracked as its own ADR rather than decided inline here, since each is independently consequential:

- **Session model for Domain B** (server-side store vs. JWT) — [ADR 0008](../../research/11-engineering/architecture-decisions/0008-gateway-session-model.md), _Proposed, undecided_.
- **Project Key format** (opaque + lookup vs. self-verifying signed token) — [ADR 0009](../../research/11-engineering/architecture-decisions/0009-project-key-format.md), _Proposed, undecided_.
- **Cache-outage failure mode** for Project Key validation (fail open vs. fail closed) — [ADR 0010](../../research/11-engineering/architecture-decisions/0010-gateway-cache-failure-mode.md), _Proposed, undecided_.

## Data Model

_TBD — first cut needed for:_
- _`Project` (key, owner, created-at, active/revoked state — not yet specified anywhere in the domain model)_
- _Session/token record shape (depends on ADR 0008)_

## API Surface

_TBD — endpoint list not yet enumerated. Known so far: login, logout, first-run admin creation (Setup Wizard, per [`docs/user-stories.md` Flow A](../../docs/user-stories.md)), project creation. Ownership of project CRUD (is it `gateway`'s, or does it belong elsewhere?) is itself still open._

## Failure Modes & Abuse Cases

_TBD — needs at least:_
- _Project Key leaked/scraped — detection and revocation path_
- _Cache unreachable (see ADR 0010)_
- _`gateway` itself down — blast radius on both domains_
- _Brute-force / credential-stuffing against Domain B login_

## Open Questions

- Everything marked _TBD_ above.
- Does `gateway` own `Project` CRUD structurally, or is that a separate concern?

## Out of Scope (don't re-litigate here)

- The UNMASK ABAC boundary — that's `query-api`'s design, not `gateway`'s.
- RBAC role definitions themselves — already settled in `research/07-security/README.md`.
- SSO — explicitly deferred past M0.
