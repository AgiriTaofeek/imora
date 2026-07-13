# Imora — User Stories & UI/UX Flow Spec

> Written for a UI/UX designer or frontend engineer to work from directly — every flow below names the actual screens, states, and components involved, not just the business outcome. Business justification for *why* each flow exists lives in [`research/01-product/README.md#user-stories`](../research/01-product/README.md#user-stories) and [`research/09-workflows/README.md`](../research/09-workflows/README.md) — this file answers "what do I design," those answer "why does this exist."
>
> Every flow is tagged **[PARITY]** (has to match the category leaders) or **[WEDGE]** (nobody else does this — the one place Imora's UI gets to be genuinely novel) or **[MIXED]**.

---

## 1. Screen Inventory

The full `dashboard` surface implied by every workflow below. Build order should roughly follow this table top to bottom — it's ordered by dependency, not alphabetically.

| # | Screen | Primary persona(s) | Parity/Wedge | Depends on |
|---|---|---|---|---|
| 1 | **Setup Wizard** (first-run) | Priya | Parity | Nothing — entry point |
| 2 | **Login** (local + SSO) | All | Parity | Setup Wizard |
| 3 | **Project / SDK Install** | Priya, Chidi | Parity | Login |
| 4 | **Session Search** | Chidi, Jon | Parity | SDK capturing data |
| 5 | **Session Replay Detail** | Chidi, Jon | Parity core, Wedge audit-trail panel | Session Search |
| 6 | **Error Groups List** | Chidi, Jon | Parity | Data captured |
| 7 | **Error Group Detail** | Chidi, Jon | Parity | Error Groups List |
| 8 | **Performance / Core Web Vitals Dashboard** | Chidi | Parity | Data captured |
| 9 | **Alert Rules & Channels Config** | Priya, Jon | Parity | Notification setup |
| 10 | **Incident Timeline** (correlated view) | Jon | Mixed | Session Replay + Error + Security data |
| 11 | **Access Audit Trail Viewer** | Adaeze, Marcus, (read-only: Dara) | Wedge | Any prior data access |
| 12 | **Data Subject / DSAR Lookup** | Adaeze | Wedge | Audit Trail Viewer |
| 13 | **Legal Hold Management** | Adaeze, Marcus | Wedge | Session data existing |
| 14 | **Retention Policy Config** | Adaeze, Priya | Wedge | Nothing — configurable day one |
| 15 | **Evidence Export Builder** | Jon, Marcus, Adaeze | Wedge | Audit Trail + Session + Error + Security data |
| 16 | **Compliance Report Generator** (HIPAA-style standing report) | Marcus | Wedge | Audit Trail Viewer |
| 17 | **User & Role Management** | Admin (Priya wears this hat in small orgs) | Parity-adjacent | Login |
| 18 | **Field Classification Config** (PII/PHI markers) | Priya, Chidi | Wedge-adjacent | SDK Install |

---

## 2. Global Interaction Patterns (apply across every screen above)

These aren't per-flow — they're system-wide UI rules. A designer should treat these as constraints on every screen listed above, not optional polish.

### Masked field: view, then escalate
1. Any field classified as soft-masked PHI/PII renders as a **fixed-format placeholder** — never length- or shape-preserving (e.g., always `[masked]`, never `••••••••` sized to the real value's length — that leaks information about the value itself).
2. An **"Unmask" affordance is always visible** next to the masked field, never buried in a menu.
3. Clicking it opens a **single required field: reason** (free text, non-empty — the UI must block submit until filled, mirroring the API's `minLength: 1` validation).
4. On submit: the real value renders inline, **and the audit-trail entry for that exact unmask action is visible immediately**, in the same view, without navigating away — the person unmasking sees their own action logged in real time.
5. **No role gets a shortened version of this flow.** Compliance Officer and Admin go through the identical reason-required modal — design one component, not a "trusted user" fast path.

### Environment selector: persistent, defaults to production, never silently "all"
Every data-bearing screen (Session Search, Error Groups List, Performance Dashboard, Incident Timeline, Alert Rules) carries a **persistent environment selector** — not a per-query filter buried in an advanced-search panel. It defaults to `production` on first load, every time, for every persona; a user has to actively switch to `staging`/`development` to see non-prod data. This is a deliberate default, not a neutral one: it's what stops a `staging` load-test session from ever appearing, unlabeled, in the middle of Chidi's production incident investigation. The selector's current value should be visually persistent (a header chip, not a dropdown that resets on navigation) so it's always obvious which environment a screen is currently scoped to — this matters most on exactly the screens where getting it wrong is costliest (Incident Timeline, Evidence Export Builder).

### Session/record discovery: search before you have an ID
Every list/search screen (Session Search, Error Groups, Audit Trail) accepts **multiple entry points** — user identifier, URL, timeframe, error/rage-click signal — not just an ID lookup box, always scoped by the environment selector above. Results should feel instant (seconds), ranked by relevance to the search signal used.

### Replay-to-trace: one click, not a manual correlation
Inside Session Replay Detail, hovering/clicking a network-request moment in the timeline surfaces a direct link to the backend trace — this is navigation, not a secondary search.

### Legal hold: friction scales with scope breadth
A narrow, bounded hold scope (specific incident reference, tight date range) submits on single approval. An unbounded-looking scope (no date bound, no session-ID list) surfaces a **second-approval requirement inline** — shown as a reason ("this scope is unbounded and needs a second approver"), never a silent block and never a silent success.

### Alerts: severity determines channel, not just badge color
A grouped error and a security-correlated incident must be visually *and* routing-ly distinct — different default channels, not just a different color chip on the same delivery path.

Full interaction rationale: [`research/10-design/README.md#interaction-patterns`](../research/10-design/README.md#interaction-patterns).

---

## 3. Flow-by-Flow Detail

### Flow A — First-Time Setup & Onboarding — Priya — [PARITY]

**Goal:** a working instance, first session captured, audit trail proven live — in under 1 hour, unassisted.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | Terminal (`docker compose up`) | Not a UI screen — but the dashboard's first-load state must handle "backend not ready yet" gracefully (loading/health-check state, not a raw error page) since startup ordering (data layer → schema migration → app services) takes real time. |
| 2 | **Setup Wizard** | First-run only. Admin account creation (username/password, Argon2id hashed), no SSO option shown here (that's Enterprise-tier, config'd later). |
| 3 | **Login** | Local auth form. MFA (TOTP) setup prompt — optional but surfaced, not buried in settings. |
| 4 | **Project / SDK Install** | Shows a project key, a copy-pasteable `init()` snippet per framework (vanilla/React/Vue/Angular tabs), and a **live "waiting for first event" indicator** — this screen should visibly update the moment the first SessionEvent arrives, not require a manual refresh. |
| 5 | **Empty-state Session Search** | Before any data exists: explicit empty state pointing back to the install snippet, not a blank table. |
| 6 | **Session Replay Detail** (first session) | Once data lands: replay renders, and critically — **the Access Audit Trail panel on this same screen already shows one entry**: the view the user is currently performing. This is the proof-of-wedge moment the onboarding target is built around. |

**Acceptance criteria (from [story P1](../research/01-product/README.md#user-stories)):** no step above should require documentation lookup or support contact to complete. If a step needs a tooltip explaining what to do next, that's a UI gap, not an acceptable "assumed knowledge."

---

### Flow B — Session Discovery from a Vague Bug Report — Chidi — [PARITY]

**Goal:** find the right session with nothing but "checkout is broken" to go on.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Session Search** | Environment selector defaults to `production` (Section 2's global pattern). Search bar accepts: user ID, URL/page, timeframe picker, or a toggle for "has error" / "has rage-click." Results list shows session cards with a snippet (page, duration, error count badge) — not just an ID. |
| 2 | **Session Replay Detail** | Full DOM replay, playback controls (play/pause/scrub), timeline strip showing error markers and network-request markers inline. Masked fields render per the global pattern (Section 2). |
| 3 | (same screen) **Access Audit Trail panel** | Visible without navigating away — collapsed by default, expandable — showing who else has viewed this session and when. |

**Edge case to design for:** zero search results. Don't show a blank table — show "no sessions matched — try broadening timeframe or search signal," since a vague ticket often means the first search guess is wrong.

---

### Flow C — Error Investigation After a Deploy — Jon / Chidi — [PARITY → WEDGE handoff]

**Goal:** one alert, not a flood; know which release caused it; jump straight to the backend trace; know immediately if it's security-related.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Error Groups List** | Environment-scoped, defaults to `production`. One row per `ErrorGroup` (not per occurrence) — occurrence count as a column, sorted by recency/severity by default. Each row shows the release tag that introduced it if `RegressionDetected` fired. |
| 2 | **Error Group Detail** | Stack trace, affected session list (click into any → Session Replay Detail), and a **release-comparison strip** showing before/after metric values for the regression. |
| 3 | (from Session Replay Detail, opened from step 2) | **Replay-to-trace link** at the exact moment of the error — one click to backend trace view (external or embedded, per whatever `TraceLink` target exists). |
| 4 | **Incident Timeline** | If a `SecuritySignalReceived` event correlates to the same session — this is the screen that visually merges replay + error + security signal into one scrollable timeline, not three separate tabs. This is the single screen that most directly demonstrates the wedge's "Investigation Over Metrics" principle — worth the most design attention of any screen in this list. |

**If the incident touches customer data:** the natural next action from this screen is a button into Flow H (Evidence Export Builder) — design this as a visible CTA on the Incident Timeline, not something Jon has to know to go find elsewhere.

---

### Flow D — Performance Regression, No Error Involved — Chidi — [PARITY]

**Goal:** "did last night's release regress LCP/INP/CLS," with no exception thrown.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Performance / Core Web Vitals Dashboard** | Environment-scoped (defaults to `production`) — this matters more here than almost anywhere else, since a `staging` load-test session mixed into a production p75 calculation would silently corrupt the metric. Three metric cards (LCP/INP/CLS) evaluated at **p75**, not average — the UI should make the percentile basis visible (a small label, not hidden in a tooltip), since p75 vs. mean is a real trust signal for this persona. Threshold lines at the Google "good" bar (2.5s / 200ms / 0.1) drawn directly on the chart. |
| 2 | (same screen) **Release markers** | Vertical markers on the timeline for each `ReleaseDeployed` event **in the selected environment only** — a regression should visually line up with a marker, not require cross-referencing a separate release list, and should never show a `staging` deploy marker while viewing `production` data. |
| 3 | Drill-down | Clicking a regressed metric/release pair opens a filtered **Session Search** (pre-filtered to sessions from that release, that metric flagged) — reuses Flow B's screen, doesn't invent a new one. |

---

### Flow E — Alert Configuration — Priya / Jon — [PARITY]

**Goal:** set up delivery so Chidi doesn't tune the product out from noise.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Alert Rules & Channels Config** | Each rule is scoped to an environment (typically `production` only — a rule not scoped this way would page on-call for a `staging` regression, exactly the noise this whole flow exists to prevent). Channel setup (Slack/email/webhook) as a simple list + add form. Rule list shows current routing: which severity tier goes to which channel. |
| 2 | (same screen) | A **visible alert-volume metric** — "N alerts fired this week, M were the same root cause" — surfaced here, not buried in an analytics-only view, since this is the concrete counter-signal to alert fatigue the product commits to tracking. |

---

### Flow F — DSAR Response — Adaeze — [WEDGE]

**Goal:** answer "what data exists for this person, and who has viewed it" inside a 1-month legal deadline, in minutes.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Data Subject / DSAR Lookup** | Single input: a data-subject identifier. Submit returns two things on one screen: (a) every session tied to that person, (b) the full access-history log for those sessions — combined, not two separate queries the user has to run and mentally merge. |
| 2 | (same screen) **Export button** | CSV/JSON/XML format picker (per GDPR's non-proprietary-format requirement) — this is a real, prominent action, not a hidden "..." menu item, since this screen's entire reason to exist is producing that export. |
| 3 | If a result includes masked fields | Same global Unmask pattern (Section 2) applies — no shortcut here even though this is a compliance-driven screen. |

**Design note:** this screen should feel *fast and confident*, not like a developer tool bolted on for compliance. Adaeze is not a technical user by default — avoid raw JSON dumps as the primary view; a structured, readable summary first, with raw export as a secondary action.

---

### Flow G — Legal Hold Application — Adaeze / Marcus — [WEDGE]

**Goal:** preserve records for an investigation without silently becoming a storage-exhaustion vector.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Legal Hold Management** | "Apply Hold" form: scope type selector (session ID list / data subject / date range / incident reference — a **closed set of options**, never a free-text query box), reason field (required). |
| 2 | (same screen, conditional) | If the entered scope is unbounded (no date range or session list) — inline banner: "this scope has no bound and requires a second approver" with an approve/request-approval action, per the global pattern in Section 2. |
| 3 | **Hold list view** | Active holds table: scope summary, applied-by, applied-at, lift action. Lifted holds shown in a separate/collapsed "history" section, not deleted from view — holds are audit-relevant even after lifting. |

---

### Flow H — Evidence Export — Jon / Marcus / Adaeze — [WEDGE]

**Goal:** one action, one frozen, hash-verifiable package — no assembling screenshots under pressure.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Evidence Export Builder** | Incident reference field, session ID(s) to include (can be pre-populated if arriving from the Incident Timeline CTA in Flow C). Preview of what will be included (replay + errors + security signal + access log) before generating — a confirmation step, since this is a compliance-weight action. |
| 2 | (same screen) **Generation state** | This is not instant — show a real progress/generating state, not a spinner with no context, since the export freezes potentially large records. |
| 3 | **Export result** | Shows `exportId`, `contentHash`, generated timestamp, and a download/retrieve action. The `contentHash` should be visibly copyable — it's the independently-verifiable integrity proof this screen's entire value proposition depends on; don't bury it in a details accordion. |

---

### Flow I — Annual HIPAA Risk Assessment — Marcus — [WEDGE]

**Goal:** a standing, generate-on-demand report satisfying §164.312(b) — not a custom export job assembled once a year under deadline pressure.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Compliance Report Generator** | Date range picker, report scope (all activity / specific data category). One "Generate Report" action. |
| 2 | Report output | A structured document view (not raw JSON) covering: access event summary, unmask-frequency-by-actor, config-change log, retention-policy state — with an export-to-PDF/print-friendly action, since this is handed to an external assessor. |

---

### Flow J — Retention Policy Configuration — Adaeze / Priya — [WEDGE]

**Goal:** per-category retention, not one global TTL — the thing every comparator lacks.

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **Retention Policy Config** | One row per data category (SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, AccessAuditEvent) — each with its own retention period and a **regulatory-basis field** (free text or a picker: PCI-DSS/HIPAA/GDPR/SOX/Custom) that's required, not optional, when setting a value. |
| 2 | (same screen) | AccessAuditEvent's row should visibly show it can never be set shorter than the longest of the other four — a disabled/computed-minimum state, not a silent validation error after submit. |
| 3 | Every save | Triggers a `CONFIG_CHANGED` audit event — the UI should confirm this ("this change has been logged") so the person editing the policy sees the same accountability principle applied to themselves. |

---

### Flow K — User & Role Management — Admin — [PARITY-adjacent]

| Step | Screen | UI state / what to design |
|---|---|---|
| 1 | **User & Role Management** | User list with role badges (Engineer / Compliance Officer / Platform Operator / Admin). Role change action requires the actor's own confirmation — this triggers `CONFIG_CHANGED` too. |
| 2 | (same screen) | **No self-elevation path.** A user cannot grant themselves a higher role — the UI should not even present that option for the currently-logged-in user's own row. |

---

### Flow L — Field Classification Setup — Priya / Chidi — [WEDGE-adjacent]

Mostly a code-level task (`init()` config or HTML attributes), but the dashboard should still surface a **read-only view of currently classified fields** (safe-listed vs. PHI-marked vs. unclassified-therefore-hard-redacted) so a team can audit their own masking configuration without reading source code — this is the screen that turns "trust the SDK config" into "verify the SDK config."

---

## 4. Cross-Cutting Design Priorities

If forced to rank which screens matter most for the product's actual thesis (parity *and* wedge working together, not two separate products):

1. **Session Replay Detail with the inline Audit Trail panel** (Flow A/B) — this single screen is where a prospective buyer sees the whole pitch in one glance.
2. **Incident Timeline** (Flow C) — the clearest visual expression of "one workspace, not four tools."
3. **DSAR Lookup** (Flow F) — the screen that has to make a non-technical compliance persona feel confident, not like she's using a developer tool.

Everything else is real, necessary, and should still be built well — but these three are where design quality most directly drives the product's central claim.

---

## What This Feeds

[`architecture.md`](architecture.md) — the API endpoints, data model, and services each screen above actually calls into.
