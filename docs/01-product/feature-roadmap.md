# Feature Roadmap

> Status: Research-based, current as of July 2026. Sequences [prd.md](prd.md)'s MVP scope into three milestones, constrained by [licensing.md](licensing.md)'s rule that no wedge capability may be moved behind a commercial gate. This document assigns *what ships when and why*; `docs/12-milestones/milestone-1.md` through `milestone-3.md` are where each phase gets broken into trackable tasks once work starts.

No calendar dates appear in this document. Committing to quarters without a real team size and velocity to base them on would be fabricated precision — per [user-personas.md](user-personas.md), Priya's team size ranges from a 2–3 person minimum-viable platform team to a 10–30 person mid-stage fintech platform group, and that range alone changes any date estimate by multiples. Sequencing and exit criteria are fixed; timing is not.

---

## Milestone 1 — Credible Alternative

**Thesis:** per [prd.md](prd.md)'s Goals, a compliance-only v1 gives Chidi nothing to open daily, and an observability-only v1 gives Dara no reason to pick Imora over OpenReplay. Milestone 1 has to ship both halves at once, or it isn't a viable v1 by either persona's standard.

**Ships:**
- Full **Parity checklist** from [prd.md](prd.md): error grouping/deduplication, production-fidelity session replay, default-safe PII masking, Core Web Vitals monitoring with release-attributed regression detection, replay-to-backend-trace correlation, single-machine and cluster self-hosted deployment paths, framework-agnostic browser SDK.
- The narrow **MVP wedge**: access-audit-trail (who viewed which session, when) and the field-level access-control system underneath it, including the audited "unmask" escalation from story M2.
- AGPLv3 licensing applied from the first public commit, per [licensing.md](licensing.md) — not retrofitted later.

**Explicitly deferred to Milestone 2:** retention-clock policy engine (A2), evidence export (J2), security-signal correlation into the incident timeline (D2). [prd.md](prd.md) already establishes why: each depends on the audit-trail's event log existing first, so building them in parallel with Milestone 1 risks rework.

**Exit criteria** (from [prd.md](prd.md)'s Success Metrics, made falsifiable):
- Time-to-first-value under 1 hour: a new deployment captures its first session and first grouped error within an hour of setup, unassisted.
- A named DPO/HIPAA-Security-Officer-persona tester can pull an access-audit-trail report for a specific session and correctly identify every internal viewer, without engineering's help, within a simulated POC window.
- A named engineer can stand up the single-machine deployment path and capture a first session with no vendor support present — the concrete test of the Operational Simplicity principle in [vision.md](../00-overview/vision.md).
- Feature-parity checklist (story P2) is complete enough that a pilot team could plausibly decommission Sentry or LogRocket in favor of Imora, not just run it alongside.

**What would mean Milestone 1 failed even if every feature shipped:** if the audit-trail system measurably compromises Operational Simplicity — if standing it up requires more than the 2–3 person minimum-viable team Priya represents — the milestone has failed on its own terms, per the tension flagged in [prd.md](prd.md)'s Open Questions. That tension gets resolved in `docs/04-architecture/overview.md`, not assumed away here.

---

## Milestone 2 — The Wedge, Complete

**Thesis:** Milestone 1 proves Imora is a credible alternative and has one real compliance advantage. Milestone 2 is what makes that advantage comprehensive enough that Dara, Adaeze, and Marcus don't need a second tool alongside Imora to close out an audit.

**Ships:**
- **Retention policy mapped to regulatory clocks** (story A2): per-data-category retention (session replay, errors, security signal configured separately), not one global TTL — built on Milestone 1's audit-trail event log as its deletion-proof mechanism. Includes legal-hold override, fulfilling the "Compliance Is a Workflow" principle in [vision.md](../00-overview/vision.md).
- **One-click, cross-signal evidence export** (story J2): replay + errors + security signal + access log as one timestamped, immutable package — built on Milestone 1's access-control system so an export's contents can be trusted.
- **Security-signal correlation into the incident timeline** (story D2): the last piece of "Investigation Over Metrics" from [vision.md](../00-overview/vision.md) — requires a security-event ingestion path that doesn't exist until this milestone.

**Exit criteria:**
- A DSAR-style query (story A1) — "what data exists for this person, and who has viewed it" — resolves in minutes, not hours, consistent with the ~15-minute benchmark cited in [user-stories.md](user-stories.md) for automated DSR platforms.
- An evidence export generated mid-incident remains valid even if a retention policy purges its underlying source data afterward — the specific ordering risk flagged in [prd.md](prd.md).
- All three Milestone 2 capabilities ship under the same AGPLv3 terms as Milestone 1, per [licensing.md](licensing.md) — this is a gate, not a preference; a roadmap draft that license-gates any of this needs to go back to licensing.md first.

---

## Milestone 3 — Sustainable and Scalable

**Thesis:** everything that is legitimately fine to monetize per [licensing.md](licensing.md), because none of it is the reason a regulated buyer chose Imora over the alternatives in the first place — it's what makes running Imora at scale sustainable for both the organizations operating it and the project building it.

**Ships:**
- Managed hosting option for organizations that want the wedge without operating the infrastructure.
- Premium support / SLA-backed response times, addressing the operational-burden concern Priya's persona raises directly.
- SSO/SAML enterprise auth integrations — a legitimate commercial gate, unlike Milestone 1/2 capabilities, because it isn't one of the three wedge gaps identified in [competitive-analysis.md](../00-overview/competitive-analysis.md).
- Multi-region/HA orchestration tooling and professional services for air-gapped or complex deployment topologies, extending the Operational Simplicity principle to large-enterprise scale per [vision.md](../00-overview/vision.md).

**Exit criteria:**
- A large-enterprise deployment (300+ employees, per the org-size variant in [target-users.md](../00-overview/target-users.md)) can run Imora across clusters/regions with the same predictable operational model Milestone 1 established for a single machine.
- Commercial offerings in this milestone generate revenue without a single feature from Milestone 1 or 2 being repositioned behind a paywall — the standing constraint from [licensing.md](licensing.md) applies permanently, not just at launch.

---

## Sequencing Logic, Summarized

| Milestone | Answers | Fails without |
|---|---|---|
| 1 — Credible Alternative | Why would anyone switch to this at all? | Chidi's daily-use case (parity) |
| 2 — The Wedge, Complete | Why does this replace, not supplement, a compliance workflow? | The audit-trail foundation from Milestone 1 |
| 3 — Sustainable and Scalable | How does this stay funded without breaking the promise in [vision.md](../00-overview/vision.md)? | A monetization surface that doesn't touch Milestones 1–2 |

This table, and the exit criteria above, are what `docs/12-milestones/milestone-1.md` through `milestone-3.md` should expand into task-level detail next.
