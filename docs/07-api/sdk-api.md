# SDK API

> Status: Research-based, current as of July 2026. The public interface of `browser-sdk`, per [container-diagrams.md](../04-architecture/container-diagrams.md) — what a customer's engineering team actually integrates on day one, per story P1. Two constraints shape everything below: framework-agnostic support ([vision.md](../00-overview/vision.md)'s Guiding Principle) and a self-imposed performance budget, because Imora's own story C1 promises same-day Core Web Vitals regression detection — it would be a direct credibility failure if the SDK meant to catch that regression caused one.

---

## Architecture: Framework-Agnostic Core, Thin Wrappers

Following the proven pattern from the closest comparator (Sentry's JavaScript SDK): a single framework-agnostic core package (`@imora/core`) provides capture, masking, and transport; thin per-framework packages (`@imora/react`, `@imora/vue`, `@imora/angular`) wrap the core with framework-specific lifecycle hooks — error boundaries in React, `errorCaptured` in Vue — and re-export everything the core exposes. This is exactly how Sentry's `@sentry/vue` and `@sentry/react` are built: wrappers around `@sentry/browser`, not independent reimplementations. It's what makes [vision.md](../00-overview/vision.md)'s "any modern frontend technology stack" claim an architectural property rather than a marketing line — a new framework wrapper is a thin adapter over already-tested core logic, not a parallel SDK to maintain.

---

## Performance Budget, Stated as a Constraint

Sentry's session-replay SDK, even after a dedicated 35% bundle-size reduction effort, still adds roughly 19–29KB minified+gzipped to the host page. Imora's SDK carries the same rrweb-class capture burden, so the same techniques are load-bearing, not optional polish:

- **Dynamic import for the replay-recording module** — loaded on-demand when `init()` actually runs, not bundled into the host application's initial critical-path load.
- **Tree-shakeable feature flags** — iframe and shadow-DOM recording (both real bundle-size contributors) are opt-in, not default-on, for applications that don't need them.
- **A stated budget, not just an aspiration:** core + minimum viable replay capture should match or beat Sentry's post-optimization figure (~20KB gzipped) — Imora doesn't get to ask a customer to trust its Core Web Vitals regression detection while quietly causing one.

---

## Public API Surface

### `init(config)`
Project key/token (authenticates to `gateway`, per [authentication.md](../08-security/authentication.md)), environment, and `release` — the tag propagated onto every SessionEvent/ErrorEvent/PerformanceMetric per [event-schema.md](../06-data/event-schema.md), typically injected at build time from CI (e.g., the git SHA) rather than hand-maintained.

### PII/PHI classification config
The `data-imora-safe` / `data-imora-mask="phi"` HTML attributes from [pii-redaction.md](../08-security/pii-redaction.md) cover static markup, but not every sensitive field is reachable that way — server-rendered fragments, dynamically generated component trees. `init(config)` accepts an equivalent **programmatic classification config** (CSS selectors, or a callback function evaluated per-field) so a team can classify fields it can't easily hand-annotate. Both paths feed the same capture-time decision in [pii-redaction.md](../08-security/pii-redaction.md) — there is one classification pipeline with two ways to configure it, not two separate mechanisms.

### `identify(userId, traits)`
Associates the current session with a user/data-subject identifier — the field Adaeze's DSAR query (story A1) resolves against.

### `captureException(error, context)`
Manual error capture, supplementing automatic `window.onerror`/`unhandledrejection` hooks — for errors an application catches and handles itself but still wants recorded.

### `addBreadcrumb(event)` / `setContext(key, data)`
Custom context attached to the current session, surfaced alongside the replay when Chidi reproduces a bug from a vague support ticket (story C3) — richer context at capture time is less reconstruction work later.

### Core Web Vitals capture
Automatic by default, per story C1 — no manual API needed for the common case; `init(config)` exposes an opt-out per-metric, not a required setup step.

---

## What's Deliberately Not Modeled Here

- Exact TypeScript type signatures and package versioning/publishing process — `10-engineering/release-process.md`.
- The wire format `init()`'s captured events are serialized into before transport to `gateway` — `07-api/rest-api.md` and [event-schema.md](../06-data/event-schema.md) already define the shape; this document only specifies what triggers it.
- Server-side/backend SDK equivalents (for TraceLink propagation from backend instrumentation) — a separate SDK surface, out of scope for `browser-sdk` specifically.

---

Sources: [Sentry JavaScript SDK — GitHub](https://github.com/getsentry/sentry-javascript), [@sentry/vue — npm](https://www.npmjs.com/package/@sentry/vue), [How We Reduced Replay SDK Bundle Size by 35% — Sentry Engineering Blog](https://sentry.engineering/blog/session-replay-sdk-bundle-size-optimizations).

## What This Feeds Next

`docs/07-api/rest-api.md` should specify the backend surface this SDK talks to (via `gateway`) and the separate programmatic API for DSAR-style queries. `docs/03-workflows/sdk-installation.md` should walk through this API from a first-time integrator's perspective.
