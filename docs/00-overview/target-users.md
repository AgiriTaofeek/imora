# Target Users

> Status: Research-based, current as of July 2026. Builds on [vision.md](vision.md) and [problem-statement.md](problem-statement.md) — this document identifies *who* carries each of the three cost drivers already quantified, using real regulatory role definitions rather than invented titles.

---

## Why Personas, Not Just "Regulated Industries"

Enterprise security and observability purchases are made by a buying committee, not a single buyer — typically 7–10 stakeholders, split across roles that read the same pitch completely differently: security reads strategic risk, engineering reads technical fit, procurement reads commercial terms. Selling (or designing) for all of them with one undifferentiated message is a documented failure mode. Imora has at least five distinct personas in a regulated org, each of whom owns a different piece of the problem.

Source: [Cybersecurity Buyer Personas: CISO, CIO, and Security Team](https://getgangly.com/blog/cybersecurity-buyer-personas).

Per [vision.md](vision.md)'s Positioning section, Imora is scoped as an alternative to tools these personas already use, plus specific compliance capabilities none of those alternatives have. That split shows up directly in the personas below: Personas 1–5 are the reason a regulated org switches to that alternative at all (the wedge); Persona 6 is the reason engineers keep using it once they have (parity).

---

## Persona 1 — CISO / Head of Security (Risk Owner, Economic Buyer)

**Who they are:** accountable for the organization's overall breach exposure and security posture. In fintechs under ~300 employees, this person frequently also owns incident-response command directly, rather than delegating it.

**What they carry:** the full weight of [problem-statement.md](problem-statement.md) Cost Driver 1 (breach cost — $7.42M average in healthcare, $5.56M in financial services) and Cost Driver 2 (CIPA wiretapping litigation exposure from third-party session-replay vendors) lands on this role's budget and reputation.

**What they need from Imora:** a platform where they can state, truthfully, that customer session data never left the organization's infrastructure — removing the specific fact pattern (a third party intercepting the session) that CIPA plaintiffs are currently winning on, and reducing the blast radius that drives breach-cost figures.

Sources: [Incident Commander: Roles, Responsibilities and Best Practices — Rootly](https://rootly.com/incident-response/incident-commander), [Incident Response Plan Template for Fintech](https://risktemplate.com/blog/2026-03-19-incident-response-plan-template-fintech/).

---

## Persona 2 — Data Protection Officer / Privacy Officer (Compliance Gatekeeper)

**Who they are:** a legally distinct, independent role — not a subset of security. Under GDPR Article 37, appointing a DPO is *mandatory*, not optional, for any organization that is a public authority, conducts large-scale systematic monitoring of individuals, or processes special-category data at scale — which covers essentially every government agency, insurer, and healthcare provider Imora targets. The DPO reports to the organization's highest management level and cannot be instructed by the controller on how to do their job — a legally protected independence most other stakeholders don't have.

**Their responsibilities directly relevant to Imora:** advising on Data Protection Impact Assessments, serving as the contact point for regulators on breach reporting, and handling Data Subject Access Requests (DSARs) — a person asking "what data do you have on me, and who has looked at it."

**What they carry:** GDPR's storage-limitation principle (data retained no longer than necessary) and the DSAR obligation are exactly the workflow gap identified in [competitive-analysis.md](competitive-analysis.md) — no frontend observability product maps retention to this, and none can answer "who on your team viewed this person's session" in the form a DSAR response requires.

**What they need from Imora:** retention policy mapped to actual regulatory clocks, and an access-audit-trail they can hand a regulator or a data subject directly, instead of reconstructing it by hand.

Sources: [What are the responsibilities of a DPO? — European Commission](https://commission.europa.eu/law/law-topic/data-protection/rules-business-and-organisations/obligations/data-protection-officers/what-are-responsibilities-data-protection-officer-dpo_en), [The DPO role under GDPR — GRC Solutions](https://grcsolutions.io/data-protection-officer-dpo-under-the-gdpr/).

---

## Persona 3 — HIPAA Security Officer (Healthcare-Specific Technical Compliance Owner)

**Who they are:** a role mandated by the HIPAA Security Rule, distinct from the HIPAA Privacy Officer. Where the Privacy Officer governs *who is allowed to see* PHI, the Security Officer is responsible for the technical safeguards protecting *electronic* PHI specifically — access controls, audit logging, encryption, and an annual documented risk assessment covering third parties and business associates.

**What they carry:** this persona is the most literal match for the access-audit-trail gap in [competitive-analysis.md](competitive-analysis.md) — "implementing the administrative, physical, and technical safeguards required by the Security Rule" is, in the frontend-observability context, precisely "who on our team viewed this patient's session, and when."

**What they need from Imora:** field-level access control and audit logging that satisfies an annual HIPAA risk assessment out of the box, not as a custom integration the Security Officer has to build against a generic audit log.

Source: [HIPAA Security Officer — 2026 Update](https://www.hipaajournal.com/hipaa-security-officer/), [HIPAA Privacy Officer vs. Security Officer](https://www.foxgrp.com/hipaa-compliance/hipaa-privacy-officer-vs-security-officer/).

---

## Persona 4 — VP Engineering / Head of Platform Engineering / SRE Lead (Technical Champion, Operator)

**Who they are:** owns the decision of what gets deployed and who operates it long-term. This is the persona evaluating self-hosting operational burden directly — Docker vs. Kubernetes, single-machine vs. multi-region, and whether a small platform team can actually run this without becoming a full-time job.

**What they carry:** Cost Driver 3 from [problem-statement.md](problem-statement.md) — the fragmentation tax (20–40% added incident-resolution time, $100K–$400K/year in tool sprawl) is this persona's budget line and their team's on-call quality of life.

**What they need from Imora:** the "Operational Simplicity" guiding principle from [vision.md](vision.md) to actually hold — a single machine for a small team, clusters/regions for a large one, with a predictable operational model at every stage. This persona will reject anything that trades data ownership for an unmaintainable deployment.

---

## Persona 5 — Incident Commander / On-Call Engineer (Day-to-Day User)

**Who they are:** the person actually inside the tool during a live incident. In regulated environments specifically, this role has an added burden most generic SRE guidance doesn't cover: **chain-of-custody documentation**, because a production incident touching customer data can become litigation, and evidence handling has to be defensible after the fact, not reconstructed from memory weeks later.

**What they carry:** this is the persona who feels tool fragmentation first-hand — establishing a shared timeline across four disconnected tools before diagnosis can even start — and the one who has the most to gain from the "Investigation Over Metrics" guiding principle in [vision.md](vision.md).

**What they need from Imora:** one workspace that already correlates replay, errors, performance, and security signal into a single timeline, so the first twenty minutes of an incident aren't spent building that timeline by hand — and evidence export that produces a defensible record without extra work, in case the incident does become litigation.

Source: [Overview of Incident Lifecycle in SRE — Squadcast](https://www.squadcast.com/blog/overview-of-incident-lifecycle-in-sre), incident-response role research cited under Persona 1.

---

## Persona 6 — Senior Frontend Engineer (Daily User, No Compliance Angle)

**Who they are:** the person who opens the tool on an ordinary Tuesday with no incident, no audit, and no regulator involved — investigating why LCP crept up after last night's deploy, triaging a spike in a JS error, or reproducing a bug a support ticket described badly. This persona doesn't appear in [problem-statement.md](problem-statement.md)'s three cost drivers at all, and that is deliberate: the first five personas explain why a regulated org would *buy* Imora, but adoption and daily retention depend on whether this person actually *wants to open it*. A compliance-mandated tool that engineers route around is a common enterprise-software failure mode, and this persona is the check against it.

**What "good" looks like, day to day, grounded in what the category already competes on:**
- **Performance regressions judged against Core Web Vitals** — LCP under 2.5s, INP under 200ms, CLS under 0.1 are the "good" thresholds Google evaluates at the 75th percentile of real user data; this engineer needs to see when a release pushes a page past those thresholds, not just an average that hides it.
- **Signal over noise** — a 2025 Catchpoint study found 62% of on-call engineers have ignored a critical alert because it was buried in noise; this persona needs errors grouped/deduplicated by root cause, not one alert per affected user.
- **"Which release did this" answered automatically** — the category standard (Sentry's regression detection) already ties a performance or error regression to the specific release that introduced it, using statistical trend detection rather than manual bisection; this persona expects that as table stakes, not a differentiator.

**What they need from Imora:** a debugging experience that is at least as good as the best point solution it's replacing (Sentry for errors, Datadog for performance) on its own merits — independent of any compliance capability — because this is the persona whose daily use makes the platform worth having at all.

Sources: [How the Core Web Vitals metrics thresholds were defined — web.dev](https://web.dev/articles/defining-core-web-vitals-thresholds), [Understanding Core Web Vitals — Google Search Central](https://developers.google.com/search/docs/appearance/core-web-vitals), [Alert Fatigue in SRE and DevOps — Sensu](https://sensu.io/blog/alert-fatigue-in-sre-and-devops), [Sentry Endpoint Regression docs](https://docs.sentry.io/product/issues/issue-details/performance-issues/endpoint-regressions/), [Sentry Releases docs](https://docs.sentry.io/product/releases/).

---

## Org-Size Variants — Personas Collapse in Smaller Organizations

Role separation scales with headcount, and Imora needs to work whether these are five people or one:

- **Under ~50 employees (early-stage fintech, small regulated startup):** the CTO typically wears the CISO, VP Engineering, and Incident Commander hats simultaneously. Imora needs to be usable by one technically capable generalist, not just a fully staffed platform team.
- **~50–300 employees:** incident-response ownership typically lands with the CISO or Head of Security, or — if that role doesn't exist yet — the VP of Engineering or Head of Compliance. Personas 1 and 4 are frequently the same person here.
- **300+ employees / large regulated enterprise:** all five buyer-side personas above (1–5; Persona 6 doesn't collapse into other roles the way buyer roles do — an engineer stays an engineer regardless of company size) are typically distinct people with separate reporting lines, and the DPO/HIPAA Security Officer's legally protected independence (Persona 2/3) becomes a real constraint on how the product's access control and audit trail must be designed — those roles need to be able to pull records without depending on engineering's cooperation.

Source: [Incident Response Plan Template: What Every Fintech Needs](https://risktemplate.com/blog/2026-03-19-incident-response-plan-template-fintech/).

---

## Mapping Personas to Cost Drivers

| Persona | Primary Cost Driver Carried | What They Need From Imora |
|---|---|---|
| CISO / Head of Security | Driver 1 (breach cost) + Driver 2 (CIPA litigation) | No third-party interception of session data |
| DPO / Privacy Officer | Driver 1 + Driver 2, plus DSAR/regulatory obligations | Regulatory-clock retention, access-audit-trail, evidence export |
| HIPAA Security Officer | Driver 1 (healthcare-specific) | Field-level access control and audit logging out of the box |
| VP Engineering / Platform Lead | Driver 3 (fragmentation tax) + deployment burden | Genuine operational simplicity at any scale |
| Incident Commander / On-call Engineer | Driver 3 (fragmentation tax) directly, day to day | Single correlated investigation workspace, defensible evidence export |
| Senior Frontend Engineer | None of the three — this is the adoption/retention check, not a cost driver | Debugging UX at parity with the best point solution, on its own merits |

Persona 6 is deliberately included above with "none of the three" rather than left out of the table entirely: Personas 1–5 explain why a regulated org signs the contract, but none of them use the product daily. If the roadmap optimizes only for Personas 1–5, Imora risks becoming a compliance-mandated tool engineers route around rather than one they choose to open — a documented failure mode in enterprise security/compliance software. Persona 6's row exists specifically to keep that risk visible in the table, not just in prose.

This table should directly inform persona sections in `docs/01-product/user-personas.md` and the JTBD framing in `docs/01-product/user-stories.md` — those are the next logical files once this one is in place.
