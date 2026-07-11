# Onboarding

> Status: The concrete walkthrough behind [prd.md](../01-product/prd.md)'s "under 1 hour, unassisted" time-to-value target and [feature-roadmap.md](../01-product/feature-roadmap.md) Milestone 1's exit criteria ("a named engineer can stand up the single-machine deployment path and capture a first session with no vendor support present"). This document is what makes that target testable, not just aspirational.

---

## The Target

Two things have to be true within one hour of starting, with nobody from Imora involved:

1. A first session is captured, masked correctly, and visible in `dashboard` — the parity check.
2. An access-audit-trail entry exists for viewing that session, visible via the same UI — the wedge check, proven true from minute one, not held back for compliance-heavy customers to discover later.

---

## Walkthrough

### 0–15 min: Deploy

`docker compose up`, per [compose.md](../09-infrastructure/compose.md). The startup ordering (data layer → `minio-init` bucket-versioning-and-lock step → schema migrations → application services) runs automatically — there is no runbook step of "wait, then manually run X." This is the concrete test of [deployment-model.md](../04-architecture/deployment-model.md)'s Operational Simplicity claim: if this step requires a human to intervene between sub-steps, the target has already failed regardless of what happens next.

### 15–20 min: First login

Local authentication per [authentication.md](../08-security/authentication.md) — Argon2id-hashed credentials, no SSO configuration required for this path, since SSO is Enterprise-tier per [pricing.md](../01-product/pricing.md) and this walkthrough is the Community/Team-tier default.

### 20–35 min: Install browser-sdk

`npm install @imora/core` (or the relevant framework wrapper, per [sdk-api.md](../07-api/sdk-api.md)), `init({projectKey, release})` in the target application. Per [sdk-api.md](../07-api/sdk-api.md)'s performance budget, this step should not be perceptibly slower than installing any comparable SDK — if bundle size or setup friction is the thing eating the hour, that's a defect in that document's own stated constraint, not an acceptable cost of onboarding.

### 35–50 min: First data

Browse the instrumented application. A SessionEvent stream, at least one PerformanceMetric (Core Web Vitals), and ideally one triggered/caught error should appear in `dashboard` within seconds of the interaction, per the capture flow in [sequence-diagrams.md](../04-architecture/sequence-diagrams.md) Flow A.

### 50–60 min: Verify the wedge, not just the parity

Open the session just captured. View its replay — this alone is the parity check, and every SaaS incumbent and self-hosted alternative in [competitive-analysis.md](../00-overview/competitive-analysis.md) can do this. **Then pull up that session's audit trail** ([rest-api.md](../07-api/rest-api.md)'s `GET /v1/sessions/{id}/audit-trail`, or the equivalent dashboard view) and confirm the view just performed is already logged there, with the viewer's identity and timestamp. This is the step that makes the hour actually prove something: not just "I can see a replay," which was never in question, but "the audit trail is real, on by default, from the first session, with nothing to configure to turn it on."

---

## What Failing This Workflow Means

If any step above requires manual intervention, external support, or more than the stated time budget, that's a defect against [prd.md](../01-product/prd.md)'s stated success metric — this document is the acceptance test for that metric, not a separate aspiration from it.

---

## What's Deliberately Not Modeled Here

- Exact dashboard screens/visual layout for each step — [11-design/dashboard-wireframes.md](../11-design/dashboard-wireframes.md).
- SDK installation detail beyond the single `init()` call shown above — [sdk-installation.md](sdk-installation.md), next.

## What This Feeds Next

[sdk-installation.md](sdk-installation.md) expands the 20–35 minute step above into full framework-by-framework detail. [error-investigation.md](error-investigation.md) and the other workflow files cover what happens after this first hour, once real usage begins.
