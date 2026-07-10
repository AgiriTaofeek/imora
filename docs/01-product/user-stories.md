# User Stories

> Status: Research-based, current as of July 2026. Translates [user-personas.md](user-personas.md) into Jobs-to-Be-Done stories with acceptance criteria grounded in the actual technical/regulatory mechanics each persona is held to — not generic "as a user" filler. Each story cites the specific requirement its acceptance criteria are derived from.

Format: **When** [situation], **I want to** [capability], **so I can** [outcome]. Acceptance criteria follow each story.

Per [vision.md](../00-overview/vision.md)'s Positioning section, every story is tagged **[PARITY]** (necessary to be a credible alternative to Sentry/Datadog/LogRocket/FullStory/OpenReplay/PostHog at all) or **[WEDGE]** (the specific capability none of those alternatives have). Parity stories define the MVP surface area; wedge stories are why a regulated buyer picks Imora over the closest self-hosted alternative once that surface area exists. See the summary table at the end.

---

## Dara (CISO) — Stories

### D1. Prove session data never left the perimeter — [PARITY]

*Self-hosting itself matches OpenReplay/PostHog; this story is table stakes for the alternative claim, not the wedge.*

**When** the board or a regulator asks whether customer session data has ever been transmitted to a third party, **I want to** point to an architecture where session capture, storage, and processing all run inside our own infrastructure, **so I can** state it as fact rather than a vendor's contractual promise.

**Acceptance criteria:**
- No session-replay payload (DOM mutations, keystrokes, network calls) leaves the deployment's own network boundary at any stage of capture, storage, or processing.
- This claim is independently verifiable — e.g., via network traffic inspection during a security review — not just documented in a privacy policy.
- Directly closes the fact pattern litigated in Mikulsky v. Bloomingdale's (see [user-personas.md](user-personas.md)), where liability turned on transmission "to a third-party vendor."

### D2. Get an incident timeline before the war room, not during it — [WEDGE]

*Replay+errors+performance correlation is parity (Sentry/OpenTelemetry pattern); folding security signal into that same timeline is what no competitor does.*

**When** a P1 incident is declared, **I want to** see a single correlated timeline of replay, errors, performance, and security signal for the affected sessions, **so I can** brief the board on scope and cause within the hour, not after a multi-tool reconstruction.

**Acceptance criteria:**
- Session replay and backend traces/errors share a common session identifier propagated from browser to backend (via header or trace baggage), so a given session's frontend replay and backend spans/errors are queryable as one object — the standard mechanism used for OpenTelemetry-based frontend/backend correlation.
- Time-to-first-coherent-timeline is measured and reported, addressing the 20–40% incident-resolution-time tax from tool fragmentation cited in [problem-statement.md](../00-overview/problem-statement.md).

---

## Adaeze (DPO) — Stories

### A1. Answer a DSAR within the one-month window, reliably — [WEDGE]

**When** a policyholder files a Data Subject Access Request, **I want to** query, within minutes, what session and error data exists for that person and who on staff has viewed it, **so I can** respond inside GDPR's one-month deadline (extendable by two months only for genuinely complex requests) without a manual engineering search eating that window.

**Acceptance criteria:**
- Query by data-subject identifier returns all session records tied to that person, plus a log of every internal viewer of those records, in a single lookup.
- Export is delivered in a commonly used, non-proprietary electronic format (CSV/JSON/XML), matching GDPR's format requirement for DSAR responses — not a PDF screenshot or a proprietary dashboard link.
- This directly targets the reported outcome from automated DSR platforms: reducing per-request handling time from multiple hours to roughly 15 minutes.

### A2. Enforce storage-limitation retention without depending on engineering to remember — [WEDGE]

**When** session or error data has outlived its documented processing purpose, **I want to** have it automatically deleted or anonymized on a policy I control, **so I can** demonstrate GDPR Article 5(1)(e) storage-limitation compliance without filing a ticket against engineering's backlog every quarter.

**Acceptance criteria:**
- Retention policy is configurable per data category (session replay, error events, security signal), not a single global TTL — matching the gap identified in [competitive-analysis.md](../00-overview/competitive-analysis.md).
- Deletion is logged as an auditable event itself, so the DPO can prove *when* data was purged, not just that a TTL exists somewhere in config.
- A **legal hold** flag can be applied to specific records to override scheduled deletion when an investigation or litigation requires preservation — the mechanic behind the "Legal hold support" commitment in [vision.md](../00-overview/vision.md)'s Guiding Principles, which otherwise has no story defining it. Holds are themselves logged (who applied it, when, why), so a hold can't be used to silently retain data past its policy window without a record of that decision.

---

## Marcus (HIPAA Security Officer) — Stories

### M1. Produce the annual audit-control evidence without a custom export job — [WEDGE]

**When** the annual HIPAA risk assessment is due, **I want to** generate a report of every access event against ePHI-containing sessions, **so I can** satisfy 45 CFR §164.312(b)'s audit-controls requirement as a standing report rather than a one-off engineering favor.

**Acceptance criteria:**
- Every access-to-sensitive-record event logs, at minimum: user ID, event/action type (view, export, delete), the record identifier accessed, timestamp, and source IP/device — the field set standard implementations of §164.312(b) converge on, since the regulation itself specifies "record and examine activity" without prescribing exact fields.
- Logs are retained a minimum of six years from creation, matching the HIPAA documentation floor, and are reviewable on a defined schedule, not just stored.
- A written, exportable policy statement accompanies the logs describing what is logged, retention length, and who can access the logs themselves — assessors ask for the policy, not just the data.

### M2. Redact PHI from a replay without losing the ability to debug it — [WEDGE]

*Default masking is parity (FullStory/Sentry already do this); the audited, reason-logged unmask escalation is the wedge.*

**When** a support engineer needs to investigate a bug in the patient portal, **I want to** let them view a redacted replay with PHI fields masked by default, escalatable to unmasked only with a logged, justified access request, **so I can** balance debuggability against minimum-necessary-access without blocking engineering entirely.

**Acceptance criteria:**
- Default view masks any field matching configured PHI patterns (name, MRN, diagnosis codes, DOB) at render time, not just at rest.
- An "unmask" action requires a reason field and is itself an audited access event, captured in the M1 log.

---

## Priya (Head of Platform Engineering) — Stories

### P1. Deploy a working instance without growing headcount — [PARITY]

**When** evaluating whether to adopt Imora, **I want to** stand up a production-representative instance with a team of 2–3 platform engineers, **so I can** validate operational cost before committing budget — matching the minimum-viable platform team size her org already runs at.

**Acceptance criteria:**
- Single-machine deployment path exists for evaluation/small-scale production, separate from the multi-region/cluster path for scale — both documented as first-class, not "cluster-only" with a single-node deployment left as an afterthought, addressing the "Operational Simplicity" principle in [vision.md](../00-overview/vision.md).
- No component requires bespoke, undocumented operational knowledge to keep running — a named on-call engineer outside the original deployer can operate it from the runbook alone.

### P2. Retire redundant tools without losing coverage — [PARITY]

**When** consolidating onto Imora, **I want to** confirm that error tracking, session replay, performance monitoring, and security signal are all present at parity with the 4–5 tools currently stitched together, **so I can** actually decommission those tools rather than running Imora as a sixth dashboard.

**Acceptance criteria:**
- Feature parity checklist against the specific tool categories named in [competitive-analysis.md](../00-overview/competitive-analysis.md) (error tracking, session replay, RUM/performance, security signal) is explicit and trackable, not assumed.
- Cost/tool-count reduction is measurable post-migration, directly addressing the $100K–$400K/year tool-sprawl figure cited in [problem-statement.md](../00-overview/problem-statement.md).

---

## Jon (Incident Commander / SRE) — Stories

### J1. Jump from a replay straight to the backend trace for that click — [PARITY]

**When** investigating a user-reported failure, **I want to** click a moment in a session replay and land directly on the backend trace/error for the exact API call happening at that moment, **so I can** skip manually correlating timestamps across tools.

**Acceptance criteria:**
- A shared session/trace identifier is attached as an attribute on every backend span associated with that session, propagated from the browser via request headers or trace baggage — the same mechanism OpenTelemetry-based frontend/backend correlation already uses in practice.
- Navigating from replay to trace and back is a single action, not a manual search by timestamp.

### J2. Produce a defensible evidence package without extra effort at 2 a.m. — [WEDGE]

**When** an incident is confirmed to involve exposed or potentially exposed customer data, **I want to** export a single, timestamped package containing the relevant replay, errors, security signal, and access log for that incident, **so I can** hand it to legal or a regulator without reconstructing it from memory or Slack threads afterward — the chain-of-custody burden already on this role, per [user-personas.md](user-personas.md).

**Acceptance criteria:**
- Export is generated in one action from the incident view, not assembled by hand from four separate tools.
- Export is immutable/timestamped once generated, so its contents can't be silently altered after the fact — a baseline requirement for anything offered as litigation evidence.

---

## Chidi (Senior Frontend Engineer) — Stories

These stories carry no compliance framing on purpose — they're the counterweight to the rest of this document, addressing the rebalancing concern raised after the first draft (see [user-personas.md](user-personas.md) Persona 6): a tool that only wins on compliance risks becoming one engineers route around.

### C1. Know immediately when a release regresses Core Web Vitals — [PARITY]

**When** a release ships, **I want to** see whether LCP, INP, or CLS moved for affected pages, evaluated at the 75th percentile the way Google's own methodology does, **so I can** catch a performance regression the same day, not after it shows up in a support queue.

**Acceptance criteria:**
- Regressions are flagged against the standard "good" thresholds (LCP < 2.5s, INP < 200ms, CLS < 0.1) at the 75th-percentile of real user sessions, not a mean that hides tail degradation.
- The regression is automatically attributed to the release that introduced it, using trend detection over a statistically meaningful window — the approach the category standard (Sentry Releases/regression issues) already uses — rather than requiring manual bisection across deploys.

### C2. Get one alert per root cause, not one per affected user — [PARITY]

**When** an error spikes in production, **I want to** be notified once for the underlying issue, with all affected sessions grouped under it, **so I can** triage instead of drowning — directly countering the documented pattern where 62% of on-call engineers admit to having ignored a critical alert buried in noise.

**Acceptance criteria:**
- Errors sharing a root cause (same stack trace, same failing endpoint) are grouped into a single actionable issue, not one notification per occurrence.
- Alert volume per real incident is a tracked product metric — if grouping quality regresses, it's visible before Chidi starts tuning the tool out.

### C3. Reproduce a badly-described bug from a support ticket in minutes — [PARITY]

**When** a support ticket says "checkout is broken" with no repro steps, **I want to** find the matching session replay by user, timeframe, or page, and watch exactly what happened, **so I can** reproduce and fix the bug without asking the reporter to redo it while being screen-shared.

**Acceptance criteria:**
- Session search by user identifier, URL, timeframe, and error/rage-click signal returns candidate sessions in seconds, not a manual log grep.
- Replay fidelity (DOM state, network calls, console errors) is sufficient to reproduce the reported behavior without needing to also read raw logs side by side — matching the baseline UX quality of the category leaders this persona is used to.

---

## Parity vs. Wedge Summary

| Story | Persona | Tag |
|---|---|---|
| D1 — Prove session data never left the perimeter | Dara | PARITY |
| D2 — Incident timeline before the war room | Dara | WEDGE |
| A1 — Answer a DSAR within one month | Adaeze | WEDGE |
| A2 — Enforce storage-limitation retention | Adaeze | WEDGE |
| M1 — Annual audit-control evidence | Marcus | WEDGE |
| M2 — Redact PHI with audited unmask | Marcus | WEDGE |
| P1 — Deploy without growing headcount | Priya | PARITY |
| P2 — Retire redundant tools at parity | Priya | PARITY |
| J1 — Replay-to-trace jump | Jon | PARITY |
| J2 — Defensible evidence package | Jon | WEDGE |
| C1 — Release regression detection | Chidi | PARITY |
| C2 — One alert per root cause | Chidi | PARITY |
| C3 — Reproduce a bug from a vague ticket | Chidi | PARITY |

**7 parity, 6 wedge.** Roughly even, which is the point: this is not a compliance tool with an observability veneer, and it's not an observability tool with compliance bolted on — the two are meant to ship as one product from the start. Note that the wedge stories cluster almost entirely on Dara, Adaeze, and Marcus (the buyers), while parity stories cluster on Priya, Jon, and Chidi (the daily users) — confirming the split first identified in [target-users.md](../00-overview/target-users.md): parity earns adoption, the wedge earns the contract.

## What This Feeds Next

These stories, especially the acceptance criteria tied to specific regulatory field/format/retention requirements (M1, A1, A2), are detailed enough to seed `docs/01-product/prd.md` directly. The parity/wedge split above should determine MVP sequencing: parity stories are the entry price for `docs/01-product/feature-roadmap.md`'s earliest milestones, and the correlation mechanism described in D2/J1 (shared session ID propagated into trace context) is a real architectural decision that belongs in `docs/04-architecture/overview.md` once product scope is fixed.
