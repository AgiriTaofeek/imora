# Design

## Interaction Patterns

> Status: Behavior, not visual layout — each pattern below is synthesized from a decision already made elsewhere in this doc set, not invented fresh. This is what makes it tractable in a text-only medium where [Dashboard Wireframes](README.md#dashboard-wireframes) genuinely isn't.

---

### Masked Field: View, Then Escalate

**Trigger:** viewing any field classified as soft-masked PHI/PII, per [PII Redaction](../07-security/README.md#pii-redaction)'s two-tier model.

1. Default render shows a fixed-format placeholder — never length- or shape-preserving, per [Threat Model](../07-security/README.md#threat-model)'s Information Disclosure finding (a placeholder that reveals the underlying value's length defeats the point of masking it).
2. An "Unmask" affordance is always visible next to a masked field — not hidden behind a menu — because the break-the-glass pattern from [Authorization](../07-security/README.md#authorization) is meant to be fast, not buried.
3. Selecting it opens a single required field: **reason**. The action is blocked, at the UI layer and the API schema layer ([openapi.yaml](../06-api/openapi.yaml)'s `minLength: 1`), until this is filled — not a placeholder default that satisfies the field technically while defeating BR-6's intent.
4. On submit, the real value renders inline, and a `FieldUnmasked` AccessAuditEvent exists immediately — visible in that same session's audit-trail view without navigating away, so the person unmasking can see their own action was logged, not just trust that it was.

**No role gets a shortened version of this flow** — including Compliance Officer and Admin, per [Authorization](../07-security/README.md#authorization)'s explicit no-exemptions rule. The interaction is identical regardless of who's performing it.

---

### Session Discovery: Search Before You Have an ID

**Trigger:** the [Session Replay](../09-workflows/README.md#session-replay) scenario — a vague bug report with no session ID attached.

1. Search accepts user identifier, URL, timeframe, or error/rage-click signal as alternative entry points, per story C3 — not just an ID lookup box.
2. Results surface in seconds, ranked by relevance to the search signal (a rage-click on the reported page ranks above an unrelated session from the same user).
3. Selecting a result opens the replay directly — no intermediate "are you sure" step, since viewing already produces its own `SessionViewed` audit event regardless of how the session was found.

---

### Replay-to-Trace: One Click, Not a Manual Correlation

**Trigger:** viewing a replay where a backend call occurred, per story J1.

Hovering or clicking a network-request moment in the replay timeline surfaces a direct link to the corresponding backend trace/span, via the shared session identifier from [Component Diagrams](../03-architecture/diagrams.md#component-diagrams). This is deliberately not a separate "find related trace" search step — the correlation already exists at write time ([Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) Flow A), so the interaction is navigation, not a query.

---

### Legal Hold: Friction Scales With Scope Breadth

**Trigger:** applying a LegalHold, per [REST API](../06-api/README.md#rest-api)'s `POST /v1/legal-holds`.

1. Scope is entered as one of the structured predicate types from [Postgres Schema](../05-data/README.md#postgres-schema) (session IDs, data subject, date range, incident reference) — never a free-text query, so the UI can evaluate breadth before submission.
2. A narrow, bounded scope (a specific incident reference, a tight date range) submits immediately on Compliance Officer approval alone.
3. An unbounded-looking scope (no date bound, no session-ID list) surfaces a second-approval requirement inline, per [Threat Model](../07-security/README.md#threat-model)'s finding — not a silent rejection, and not a silent success either. The person applying the hold sees exactly why it needs a second approver, framed as the same storage-exhaustion risk [Scaling](../03-architecture/README.md#scaling) identified, not an opaque policy block.

---

### Alerts: Severity Determines Channel, Not Just Color

**Trigger:** an `AlertTriggered` event reaching a human, per [Alerting](../09-workflows/README.md#alerting).

A grouped error (story C2) and a security-correlated incident (story D2) render with different visual severity treatment *and* route to different channels by default — not just a different badge color on the same notification path. This is the interaction-layer expression of [Alerting](../09-workflows/README.md#alerting)'s point that treating every alert as equally urgent is what produces the 62%-ignored-alerts statistic in the first place.

---

### What's Deliberately Not Modeled Here

- Visual design (color, spacing, typography) for any of the above — [Design System](README.md#design-system).
- Exact component library/framework choice for `dashboard`'s implementation — downstream of [Coding Standards](../11-engineering/README.md#coding-standards), not part of interaction design.

### What This Feeds Next

[Design System](README.md#design-system) should specify the visual language these patterns render in. [Dashboard Wireframes](README.md#dashboard-wireframes) is the one file in this doc set genuinely better served by an actual design tool than prose — worth flagging rather than forcing a text-only substitute.

---

## Design System

> Status: Deferred — visual tokens (color, typography, spacing) are genuinely better authored in a design tool than prose; forcing a text substitute here would produce something worse than useful. [Interaction Patterns](README.md#interaction-patterns) already specifies the *behavior* this file will eventually give a visual language to. See [docs/README.md](../README.md)'s folder table.

Imora's design tokens, components, and visual language.

### Overview

_TBD_

---

## Dashboard Wireframes

> Status: Deferred — visual layout is genuinely better authored in an actual design tool (Figma or equivalent) than prose; a text description of screen layout would be a worse artifact than no artifact. The workflows these screens support are already fully specified in [09-workflows/](../09-workflows/) and [Interaction Patterns](README.md#interaction-patterns). See [docs/README.md](../README.md)'s folder table.

Wireframes for key dashboard screens and flows.

### Overview

_TBD_

