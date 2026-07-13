# API

## REST API

> Status: Research-based, current as of July 2026. The programmatic surface for `query-api` and `workers` — what Adaeze scripts against to hit GDPR's one-month DSAR deadline (story A1), what a third-party tool would integrate against, and the API `dashboard` itself calls, per [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s Conformist relationship.

---

### Versioning: URL Path, Not Headers

**`/v1/...`** — URL path versioning, the recommended default for a REST API: it lets `gateway` route by version prefix without parsing headers, consistent with `gateway`'s existing role as the single routing chokepoint ([Bounded Contexts](../02-domain/README.md#bounded-contexts)), and it caches cleanly, which header-based versioning doesn't without extra `Vary` configuration.

**Backward compatibility discipline:** at least two versions supported concurrently (current + previous), with deprecation announced at least 6 months ahead via response headers and documentation before a version is retired. This is a *looser* rule than [Event Schema](../05-data/README.md#event-schema)'s "additive-only, forever" — that rule exists because stored events must remain readable for up to 7 years; a request/response API can break at a major version boundary as long as clients get real migration time, which stored data can't be given.

---

### Authentication: Separate API Tokens, Not Interactive Session Tokens

Per [Authentication](../07-security/README.md#authentication), interactive logins (local or SSO) produce a session token scoped to a browser session. Programmatic access — the kind Adaeze needs to actually script a DSAR response within her one-month window — uses a **separate, long-lived API token**, issued per-user and revocable independently of that user's interactive session. Both token types resolve to the same `RequestContext` and the same `actorUserId` on every downstream AccessAuditEvent — the distinction is lifecycle and purpose, not identity or privilege.

---

### The Guarantee This Channel Doesn't Get to Skip

Every read through this API is still served by `query-api`'s `AuditedQueryHandler`, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) — **the REST API is another caller of the same enforcement layer, not a separate path around it.** A third-party tool integrating against `GET /v1/sessions/{id}` produces exactly the same `SessionViewed` AccessAuditEvent a dashboard click would. This is worth stating in the API reference itself, not just the architecture docs — it's the concrete answer to "does using the API instead of the dashboard weaken the audit trail," and the answer is no, structurally, not by policy.

---

### Resource Surface

| Endpoint | Method | Maps to |
|---|---|---|
| `/v1/sessions` | GET | Session search — the [Session Search](../09-workflows/README.md#session-replay) workflow's backing query. Filterable by `environment`, `release`, user identifier, URL, and timeframe; `environment` defaults to `production` if unset, never to "all environments," so a query never silently mixes non-prod noise into a prod investigation. |
| `/v1/sessions/{id}` | GET | Session + replay, per [Domain Model](../02-domain/README.md#domain-model) |
| `/v1/sessions/{id}/audit-trail` | GET | AccessAuditEvent history for that session — stories M1, A1 |
| `/v1/data-subjects/{id}/sessions` | GET | The DSAR query surface — "all sessions and access history for this person," story A1's core scriptable endpoint |
| `/v1/sessions/{id}/unmask` | POST | UNMASK escalation; request body requires the non-empty `reason` field per BR-6 — the API rejects the request at the schema level if `reason` is missing, not just at the audit-logging level |
| `/v1/legal-holds` | POST | Apply a LegalHold (scope predicate, per [Postgres Schema](../05-data/README.md#postgres-schema)) |
| `/v1/legal-holds/{id}` | DELETE | Lift a LegalHold |
| `/v1/retention-policies/{category}` | GET, PUT | RetentionPolicy per data category — a PUT here is exactly the action that produces a `CONFIG_CHANGED` AccessAuditEvent per [Audit Logging](../07-security/README.md#audit-logging) |
| `/v1/evidence-exports` | POST | Trigger EvidenceExport generation, story J2 |
| `/v1/evidence-exports/{id}` | GET | Retrieve a generated export's metadata and `contentHash` |
| `/v1/erasure-requests` | POST | Intake for BR-3's erasure-vs-legal-obligation evaluation |

---

### Response Format

JSON by default. `/v1/data-subjects/{id}/sessions` and `/v1/evidence-exports/{id}` additionally support content negotiation for CSV/XML via `Accept`, per story A1's requirement that DSAR responses use a "commonly used, non-proprietary electronic format" — GDPR's actual wording, not just JSON-only convenience.

**List endpoints are cursor-paginated**, not offset-based — offset pagination degrades on the large, append-only tables this API queries against (per [ClickHouse Schema](../05-data/README.md#clickhouse-schema)'s `ORDER BY` design), and a cursor is stable under concurrent writes in a way an offset isn't.

**Rate limiting** is enforced at `gateway` via the Redis cache from [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) — uniform across dashboard-originated and directly-scripted requests, since both terminate at the same chokepoint.

---

### What's Deliberately Not Modeled Here

- Exact request/response JSON schemas per endpoint — `06-api/openapi.yaml`, the formal spec this document's decisions feed into.
- Webhook delivery for asynchronous events (AlertTriggered, EvidenceExportGenerated) — `06-api/README.md#webhooks`.
- Specific rate-limit thresholds — an operational tuning value, not an architectural one.

---

Sources: [REST API Versioning: Definition, Strategies & Best Practices — DigitalAPI](https://www.digitalapi.ai/blogs/rest-versioning-definition-best-practices-pros-cons-and-when-to-use), [API Versioning Strategies 2026 — Askantech](https://www.askantech.com/api-versioning-strategies-rest-header-url-deprecation-guide/).

### What This Feeds Next

`research/06-api/README.md#webhooks` should specify the asynchronous/push side this document deliberately left out. `research/06-api/openapi.yaml` should formalize the resource surface above into an actual spec.

---

## SDK API

> Status: Research-based, current as of July 2026. The public interface of `browser-sdk`, per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) — what a customer's engineering team actually integrates on day one, per story P1. Two constraints shape everything below: framework-agnostic support ([Vision](../00-overview/README.md#vision)'s Guiding Principle) and a self-imposed performance budget, because Imora's own story C1 promises same-day Core Web Vitals regression detection — it would be a direct credibility failure if the SDK meant to catch that regression caused one.

---

### Architecture: Framework-Agnostic Core, Thin Wrappers

Following the proven pattern from the closest comparator (Sentry's JavaScript SDK): a single framework-agnostic core package (`@imora/core`) provides capture, masking, and transport; thin per-framework packages (`@imora/react`, `@imora/vue`, `@imora/angular`) wrap the core with framework-specific lifecycle hooks — error boundaries in React, `errorCaptured` in Vue — and re-export everything the core exposes. This is exactly how Sentry's `@sentry/vue` and `@sentry/react` are built: wrappers around `@sentry/browser`, not independent reimplementations. It's what makes [Vision](../00-overview/README.md#vision)'s "any modern frontend technology stack" claim an architectural property rather than a marketing line — a new framework wrapper is a thin adapter over already-tested core logic, not a parallel SDK to maintain.

---

### Performance Budget, Stated as a Constraint

Sentry's session-replay SDK, even after a dedicated 35% bundle-size reduction effort, still adds roughly 19–29KB minified+gzipped to the host page. Imora's SDK carries the same rrweb-class capture burden, so the same techniques are load-bearing, not optional polish:

- **Dynamic import for the replay-recording module** — loaded on-demand when `init()` actually runs, not bundled into the host application's initial critical-path load.
- **Tree-shakeable feature flags** — iframe and shadow-DOM recording (both real bundle-size contributors) are opt-in, not default-on, for applications that don't need them.
- **A stated budget, not just an aspiration:** core + minimum viable replay capture should match or beat Sentry's post-optimization figure (~20KB gzipped) — Imora doesn't get to ask a customer to trust its Core Web Vitals regression detection while quietly causing one.

---

### Public API Surface

#### `init(config)`
Project key/token (authenticates to `gateway`, per [Authentication](../07-security/README.md#authentication)), `environment`, and `release`. `release` is typically injected at build time from CI (e.g., the git SHA); `environment` (`production`, `staging`, `development`, or a team's own value) is usually set per deploy target rather than per build — both are free-text tags, propagated onto every SessionEvent/ErrorEvent/PerformanceMetric per [Event Schema](../05-data/README.md#event-schema), per [Domain Model](../02-domain/README.md#domain-model)'s Environment note. Neither field changes SDK behavior (masking, capture fidelity) — they're query dimensions, not mode switches.

#### PII/PHI classification config
The `data-imora-safe` / `data-imora-mask="phi"` HTML attributes from [PII Redaction](../07-security/README.md#pii-redaction) cover static markup, but not every sensitive field is reachable that way — server-rendered fragments, dynamically generated component trees. `init(config)` accepts an equivalent **programmatic classification config** (CSS selectors, or a callback function evaluated per-field) so a team can classify fields it can't easily hand-annotate. Both paths feed the same capture-time decision in [PII Redaction](../07-security/README.md#pii-redaction) — there is one classification pipeline with two ways to configure it, not two separate mechanisms.

#### `identify(userId, traits)`
Associates the current session with a user/data-subject identifier — the field Adaeze's DSAR query (story A1) resolves against.

#### `captureException(error, context)`
Manual error capture, supplementing automatic `window.onerror`/`unhandledrejection` hooks — for errors an application catches and handles itself but still wants recorded.

#### `addBreadcrumb(event)` / `setContext(key, data)`
Custom context attached to the current session, surfaced alongside the replay when Chidi reproduces a bug from a vague support ticket (story C3) — richer context at capture time is less reconstruction work later.

#### Core Web Vitals capture
Automatic by default, per story C1 — no manual API needed for the common case; `init(config)` exposes an opt-out per-metric, not a required setup step.

---

### What's Deliberately Not Modeled Here

- Exact TypeScript type signatures and package versioning/publishing process — `11-engineering/README.md#release-process`.
- The wire format `init()`'s captured events are serialized into before transport to `gateway` — `06-api/README.md#rest-api` and [Event Schema](../05-data/README.md#event-schema) already define the shape; this document only specifies what triggers it.
- Server-side/backend SDK equivalents (for TraceLink propagation from backend instrumentation) — a separate SDK surface, out of scope for `browser-sdk` specifically.

---

Sources: [Sentry JavaScript SDK — GitHub](https://github.com/getsentry/sentry-javascript), [@sentry/vue — npm](https://www.npmjs.com/package/@sentry/vue), [How We Reduced Replay SDK Bundle Size by 35% — Sentry Engineering Blog](https://sentry.engineering/blog/session-replay-sdk-bundle-size-optimizations).

### What This Feeds Next

`research/06-api/README.md#rest-api` should specify the backend surface this SDK talks to (via `gateway`) and the separate programmatic API for DSAR-style queries. `research/09-workflows/README.md#sdk-installation` should walk through this API from a first-time integrator's perspective.

---

## Webhooks

> Status: Research-based, current as of July 2026. Resolves what [Event Catalog](../02-domain/README.md#event-catalog) deferred: which events get exposed externally, and the delivery mechanics (signing, retry, SSRF protection) that make outbound delivery to a customer-configured endpoint safe.

---

### Which Events Are Exposed, and Why This List Grew by One

[Event Catalog](../02-domain/README.md#event-catalog) flagged AlertTriggered, RegressionDetected, and EvidenceExportGenerated as likely candidates. Confirmed, plus one addition:

- **AlertTriggered** — the obvious case; customers already route alerts to their own on-call tooling.
- **RegressionDetected** — story C1's release-attributed regression, useful to pipe into a customer's own deploy-gating tooling.
- **EvidenceExportGenerated** — notifies that an export completed, without carrying its contents (see below).
- **ConfigurationChanged**, added here — [Threat Model](../07-security/README.md#threat-model) identified that gap-detection and unmask-review mechanisms exist internally but nothing actively surfaces them externally. A webhook on `ConfigurationChanged` is what lets a customer's own SIEM alert in real time the moment someone weakens a RetentionPolicy or changes a field's PHI classification — an external, independent check on exactly the tampering scenario [Threat Model](../07-security/README.md#threat-model) was worried about, not just an internal `notification-service` alert that a compromised insider could also suppress.

**DeletionSkippedDueToHold and other AccessAuditEvent variants are deliberately not exposed as webhooks** — that volume belongs in the REST API's audit-trail query surface ([REST API](README.md#rest-api)), not a push mechanism; webhooks are for events a customer needs to react to immediately, not a firehose of every access.

---

### Delivery Security

- **HMAC-SHA256 signing**, per-endpoint secret, in a signature header verified before the receiver does anything with the payload — the baseline every major webhook provider uses. (Ephemeral, short-lived signing keys via a JWKS-style endpoint are a legitimate future hardening step, not required for the initial design.)
- **SSRF protection on every delivery attempt, not just at configuration time:** an explicit allowlist plus DNS-resolution check immediately before each fetch — resolve the hostname, refuse delivery if it resolves to a private, loopback, or link-local IP. Checking only when the customer first configures the URL is insufficient; DNS rebinding means a hostname can resolve safely at setup time and to an internal address at actual delivery time.
- **Retry with exponential backoff and jitter**, spanning roughly 3 days, capped at a bounded number of attempts — enough to survive a receiver's temporary outage without becoming an unbounded retry storm against a permanently broken endpoint.

---

### Payload Content: Metadata and References Only, Never Sensitive Content Inline

**No webhook payload ever carries session, PHI, or export content directly — only identifiers and a link back to the access-controlled REST API.** `EvidenceExportGenerated`'s payload is `{exportId, incidentReference, contentHash, generatedAt}`, not the export itself; retrieving the actual content requires `GET /v1/evidence-exports/{id}` through [REST API](README.md#rest-api)'s authenticated, audited path.

This resolves a question worth stating explicitly rather than leaving implicit: **does a webhook delivery need its own AccessAuditEvent?** No — because it never carries anything sensitive, there's nothing to audit at delivery time. The actual access event fires when someone uses the webhook's reference to fetch the real content through the REST API, which already produces `RecordExported`, per [REST API](README.md#rest-api). A customer-configured webhook endpoint is outside Imora's own access controls by definition, so it's the wrong place to ever put content that controls are supposed to govern.

---

### Air-Gapped Consistency

Webhooks are outbound-only and customer-configured — consistent with [System Context](../03-architecture/README.md#system-context)'s "optional convenience, never load-bearing" rule for external systems. An air-gapped deployment can still use webhooks pointed at another system inside the same isolated network (an internal SIEM, an internal ticketing system) — the same "the external system just has to be inside the air gap too" logic already established for SSO in [Authentication](../07-security/README.md#authentication).

---

### What's Deliberately Not Modeled Here

- Exact JSON payload schemas per event — `06-api/openapi.yaml`.
- Webhook management UI/API (registering, testing, viewing delivery history) — a `dashboard`/`README.md#rest-api` concern once this design is accepted.
- Ephemeral signing-key rotation mechanics — a future hardening step noted above, not specified further here.

---

Sources: [Webhook Security: HMAC, Retries, Idempotency](https://didit.me/blog/webhook-security-patterns/), [Webhook Security Guide: HMAC Signatures & Replay Protection — Hooklistener](https://www.hooklistener.com/learn/webhook-security-fundamentals), [Standard Webhooks Specification](https://github.com/standard-webhooks/standard-webhooks/blob/main/spec/standard-README.md#webhooks).

### What This Feeds Next

`research/06-api/openapi.yaml` is the last file in `06-api/` — it should formalize [REST API](README.md#rest-api)'s resource surface and this document's payload shapes into an actual spec.

