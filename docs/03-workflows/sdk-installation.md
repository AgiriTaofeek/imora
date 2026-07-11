# SDK Installation

> Status: Expands [onboarding.md](onboarding.md)'s 20–35 minute installation step into full detail, per framework — direct application of [sdk-api.md](../07-api/sdk-api.md)'s core-plus-wrapper architecture.

---

## Vanilla JS / Framework-Agnostic

```js
import { init } from '@imora/core';

init({
  projectKey: '<project-key>',
  release: process.env.BUILD_SHA,
});
```

Everything else in [sdk-api.md](../07-api/sdk-api.md)'s public surface (`identify`, `captureException`, `addBreadcrumb`) is available from this same import — this is the baseline every framework wrapper builds on, not a separate integration path.

## React

```js
import { init } from '@imora/react';
// same init() call; the wrapper adds an error boundary and route-change tracking automatically
```

## Vue

```js
import { init } from '@imora/vue';
// adds Vue's errorCaptured hook automatically
```

## Angular

```js
import { init } from '@imora/angular';
// adds Angular's ErrorHandler integration automatically
```

Each wrapper re-exports the full `@imora/core` API per [sdk-api.md](../07-api/sdk-api.md) — there is no capability available in one framework's package and not another's; a new framework wrapper is an adapter over already-tested core logic, not a parallel implementation with its own gaps.

---

## Classifying Sensitive Fields at Install Time

This is the step most likely to be skipped under time pressure, and skipping it doesn't fail open — per [pii-redaction.md](../08-security/pii-redaction.md) and business rule BR-7, any field with no classification is hard-redacted by default. That means the practical risk of skipping this step isn't a compliance gap, it's **debugging friction later** — a support engineer investigating a bug six weeks from now finds an unexpectedly redacted field, not exposed PII. Still worth doing at install time rather than discovering the gap under incident pressure:

```js
init({
  projectKey: '<project-key>',
  release: process.env.BUILD_SHA,
  fieldClassification: {
    safe: ['.product-name', '.page-title'],
    phi: ['[data-patient-field]', '.diagnosis-code'],
  },
});
```

Matches [pii-redaction.md](../08-security/pii-redaction.md)'s programmatic classification path — the same mechanism the `data-imora-safe`/`data-imora-mask` HTML attributes feed, for applications where hand-annotating markup isn't practical.

---

## Verifying Installation

The same check [onboarding.md](onboarding.md) specifies: browse the instrumented app, confirm a session appears in `dashboard` within seconds, confirm Core Web Vitals are recorded. If nothing appears, the most common cause is a `projectKey` mismatch or an ad-blocker/CSP rule blocking the SDK's network calls to `gateway` — not a deeper integration problem.

## What This Feeds Next

[error-investigation.md](error-investigation.md) picks up from here — what happens once real errors start arriving.
