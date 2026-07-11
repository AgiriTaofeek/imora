# REST API

> Status: Research-based, current as of July 2026. The programmatic surface for `query-api` and `workers` — what Adaeze scripts against to hit GDPR's one-month DSAR deadline (story A1), what a third-party tool would integrate against, and the API `dashboard` itself calls, per [bounded-contexts.md](../02-domain/bounded-contexts.md)'s Conformist relationship.

---

## Versioning: URL Path, Not Headers

**`/v1/...`** — URL path versioning, the recommended default for a REST API: it lets `gateway` route by version prefix without parsing headers, consistent with `gateway`'s existing role as the single routing chokepoint ([bounded-contexts.md](../02-domain/bounded-contexts.md)), and it caches cleanly, which header-based versioning doesn't without extra `Vary` configuration.

**Backward compatibility discipline:** at least two versions supported concurrently (current + previous), with deprecation announced at least 6 months ahead via response headers and documentation before a version is retired. This is a *looser* rule than [event-schema.md](../06-data/event-schema.md)'s "additive-only, forever" — that rule exists because stored events must remain readable for up to 7 years; a request/response API can break at a major version boundary as long as clients get real migration time, which stored data can't be given.

---

## Authentication: Separate API Tokens, Not Interactive Session Tokens

Per [authentication.md](../08-security/authentication.md), interactive logins (local or SSO) produce a session token scoped to a browser session. Programmatic access — the kind Adaeze needs to actually script a DSAR response within her one-month window — uses a **separate, long-lived API token**, issued per-user and revocable independently of that user's interactive session. Both token types resolve to the same `RequestContext` and the same `actorUserId` on every downstream AccessAuditEvent — the distinction is lifecycle and purpose, not identity or privilege.

---

## The Guarantee This Channel Doesn't Get to Skip

Every read through this API is still served by `query-api`'s `AuditedQueryHandler`, per [component-diagrams.md](../04-architecture/component-diagrams.md) — **the REST API is another caller of the same enforcement layer, not a separate path around it.** A third-party tool integrating against `GET /v1/sessions/{id}` produces exactly the same `SessionViewed` AccessAuditEvent a dashboard click would. This is worth stating in the API reference itself, not just the architecture docs — it's the concrete answer to "does using the API instead of the dashboard weaken the audit trail," and the answer is no, structurally, not by policy.

---

## Resource Surface

| Endpoint | Method | Maps to |
|---|---|---|
| `/v1/sessions/{id}` | GET | Session + replay, per [domain-model.md](../02-domain/domain-model.md) |
| `/v1/sessions/{id}/audit-trail` | GET | AccessAuditEvent history for that session — stories M1, A1 |
| `/v1/data-subjects/{id}/sessions` | GET | The DSAR query surface — "all sessions and access history for this person," story A1's core scriptable endpoint |
| `/v1/sessions/{id}/unmask` | POST | UNMASK escalation; request body requires the non-empty `reason` field per BR-6 — the API rejects the request at the schema level if `reason` is missing, not just at the audit-logging level |
| `/v1/legal-holds` | POST | Apply a LegalHold (scope predicate, per [postgres-schema.md](../06-data/postgres-schema.md)) |
| `/v1/legal-holds/{id}` | DELETE | Lift a LegalHold |
| `/v1/retention-policies/{category}` | GET, PUT | RetentionPolicy per data category — a PUT here is exactly the action that produces a `CONFIG_CHANGED` AccessAuditEvent per [audit-logging.md](../08-security/audit-logging.md) |
| `/v1/evidence-exports` | POST | Trigger EvidenceExport generation, story J2 |
| `/v1/evidence-exports/{id}` | GET | Retrieve a generated export's metadata and `contentHash` |
| `/v1/erasure-requests` | POST | Intake for BR-3's erasure-vs-legal-obligation evaluation |

---

## Response Format

JSON by default. `/v1/data-subjects/{id}/sessions` and `/v1/evidence-exports/{id}` additionally support content negotiation for CSV/XML via `Accept`, per story A1's requirement that DSAR responses use a "commonly used, non-proprietary electronic format" — GDPR's actual wording, not just JSON-only convenience.

**List endpoints are cursor-paginated**, not offset-based — offset pagination degrades on the large, append-only tables this API queries against (per [clickhouse-schema.md](../06-data/clickhouse-schema.md)'s `ORDER BY` design), and a cursor is stable under concurrent writes in a way an offset isn't.

**Rate limiting** is enforced at `gateway` via the Redis cache from [container-diagrams.md](../04-architecture/container-diagrams.md) — uniform across dashboard-originated and directly-scripted requests, since both terminate at the same chokepoint.

---

## What's Deliberately Not Modeled Here

- Exact request/response JSON schemas per endpoint — `07-api/openapi.yaml`, the formal spec this document's decisions feed into.
- Webhook delivery for asynchronous events (AlertTriggered, EvidenceExportGenerated) — `07-api/webhooks.md`.
- Specific rate-limit thresholds — an operational tuning value, not an architectural one.

---

Sources: [REST API Versioning: Definition, Strategies & Best Practices — DigitalAPI](https://www.digitalapi.ai/blogs/rest-versioning-definition-best-practices-pros-cons-and-when-to-use), [API Versioning Strategies 2026 — Askantech](https://www.askantech.com/api-versioning-strategies-rest-header-url-deprecation-guide/).

## What This Feeds Next

`docs/07-api/webhooks.md` should specify the asynchronous/push side this document deliberately left out. `docs/07-api/openapi.yaml` should formalize the resource surface above into an actual spec.
