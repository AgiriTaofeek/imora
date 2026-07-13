# Overview

## Vision

### Project Name

**Imora**

A self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory — built for regulated industries, with compliance capabilities none of them, or their self-hosted counterparts, ship.

---

## Positioning

Imora is not a new category. It is an alternative to the frontend observability tools engineering teams already use and already trust the workflow of — error tracking, session replay, performance monitoring — running entirely inside the organization's own infrastructure instead of a vendor's cloud.

Being "self-hosted" alone isn't the pitch — OpenReplay, PostHog, SigNoz, and GlitchTip already offer that, and are covered in detail in [Competitive Analysis](README.md#competitive-analysis). The pitch is: **parity with the tools regulated teams already know, plus the specific compliance capabilities that neither the SaaS incumbents nor the existing self-hosted alternatives have** — an access-audit-trail over who viewed a customer's session, retention mapped to actual regulatory clocks, and evidence export built for an auditor rather than an engineer.

Every claim in this document is checked against that bar: is this parity (necessary to be a credible alternative at all) or wedge (the specific reason to choose Imora over every other alternative on the list)? Both matter. A product that's only compliant and mediocre at debugging won't get adopted daily; a product that's a great debugger with no compliance story doesn't solve the problem regulated buyers actually have.

---

## Vision Statement

To be the frontend observability platform regulated organizations reach for instead of Sentry, Datadog RUM, LogRocket, or FullStory — not because it does something entirely different, but because it does the same job without requiring customer data to leave the organization's infrastructure, and because it answers the questions a compliance team asks that none of those tools are built to answer.

We believe organizations should not have to choose between deep visibility into their frontend applications and complete ownership of their customer data.

Our mission is to give engineering, security, and compliance teams a single system of record for frontend incidents — one that keeps telemetry inside their infrastructure and produces evidence they can hand to an auditor, not just a dashboard they can screenshot.

---

## Why We Exist

Modern organizations rely heavily on web applications to deliver critical services.

Banks process transfers through browsers. Insurance companies handle claims through portals. Governments provide citizen services online. Healthcare providers manage patient workflows through web applications.

When these applications fail, engineering teams need answers.

What happened? Who was affected? Why did it happen? Was any regulated data exposed? Who is allowed to know?

Today, most organizations rely on external observability vendors to answer these questions. While these platforms provide valuable capabilities, regulated organizations run into the same wall:

- Regulatory restrictions around data residency.
- Strict security requirements that a shared multi-tenant SaaS product cannot satisfy.
- Customer privacy concerns, especially around session replay capturing PII, PHI, or payment data.
- Compliance obligations that require proving *who accessed what customer data, and when* — not just that the system was secure.
- Limited control over telemetry retention, deletion, and processing.
- Inability to operate in private, on-premise, or air-gapped environments.

As a result, many organizations are forced to sacrifice visibility in order to maintain compliance and security.

We believe this trade-off should not exist.

---

## The Gap in Today's Landscape

This is not an unserved market — it is a market where every existing option makes a different compromise. Naming that gap precisely is what makes Imora's parity-plus-wedge positioning credible rather than a marketing claim, and it's covered in full detail, with sources, in [Competitive Analysis](README.md#competitive-analysis).

**SaaS incumbents (Datadog RUM, New Relic Browser, LogRocket, FullStory, Sentry Cloud) solve depth, not ownership.**
They offer excellent observability — this is the parity bar Imora has to clear — but even their "private" or "dedicated cloud" enterprise tiers run in the vendor's infrastructure under the vendor's control plane. For a bank or hospital, that is not self-hosting — the data still leaves the organization's perimeter, and the customer's compliance team is trusting a third party's SOC 2 report instead of their own controls.

**Open-source self-hosted session-replay tools (OpenReplay, PostHog, and — until its LaunchDarkly acquisition wound the standalone product down — Highlight.io) solve infrastructure ownership, not compliance workflow.**
This is Imora's actual competitive set — the alternatives regulated teams are already choosing between. To their credit, most of these already default to reasonably safe capture: OpenReplay obscures common PII patterns automatically, PostHog masks input elements by default, and Highlight.io obfuscated inputs against common PII regex patterns out of the box — roughly the same bar the best SaaS incumbents (Sentry, FullStory) already clear. Running the software on your own servers, with sane default masking, is necessary but still not sufficient to win a regulated buyer over the others on this list. None of them ship with:

- **Access-to-data audit trails** — a log of *who on your team viewed a specific customer's session replay*, not just a system-level "user logged in" event. This pattern already exists in privileged-access-management tools (BeyondTrust, Delinea) for infrastructure sessions — it has not been adopted by any frontend session-replay tool we found.
- **Retention mapped to regulatory clocks** — every platform we reviewed exposes a single global TTL, not policies tied to the actual numbers regulated orgs must hit: PCI-DSS's 12-month minimum audit trail retention, HIPAA's 6-year documentation floor, GDPR's purpose-bound (not fixed-term) storage limitation, or SOX's 7-year requirement.
- **Evidence export for auditors** — the ability to produce a defensible, timestamped package for an incident (replay + errors + security signal + access log) instead of forcing a compliance team to reconstruct the story from screenshots. We found no frontend observability product, self-hosted or SaaS, that ships this.

These three are Imora's wedge — the reason to pick it over OpenReplay or PostHog specifically, not just over Datadog.

Separately, **backend-focused self-hosted tools (SigNoz, Grafana Faro) solve telemetry ownership but don't do frontend session intelligence at all** — no session replay, no frontend PII capture to govern in the first place. And **GlitchTip**, the lightweight self-hosted Sentry alternative, is error-tracking only by design; replay calls against it fail silently. Neither is a real alternative to Imora — they compete on a narrower slice of the parity bar, not on the wedge.

**Fragmentation itself is a compliance liability, not just an inconvenience.**
When an incident touches security, performance, and user experience, teams today stitch the story together across three or four tools — each with its own access controls, its own retention policy, and its own audit trail. In a regulated environment, that means there is no single, coherent record of what was seen, by whom, and when. That gap, not feature parity alone, is the one Imora is built to close.

---

## Our Belief

Frontend applications have become critical infrastructure. Yet frontend observability remains fragmented.

Teams often need multiple tools to understand a single incident:

- Error monitoring tools.
- Product analytics platforms.
- Session replay solutions.
- Security monitoring systems.
- Performance monitoring tools.
- Backend tracing platforms.

This fragmentation increases operational complexity, slows incident resolution, and — for regulated organizations specifically — breaks the chain of custody an auditor or regulator will eventually ask for.

We believe engineers should be able to investigate an issue from a single workspace, and compliance teams should be able to answer "who saw what, and when" from that same workspace, without a second system.

---

## What We Are Building

### Parity — what Imora has to match to be a credible alternative at all

The baseline every SaaS incumbent and self-hosted alternative in [Competitive Analysis](README.md#competitive-analysis) already offers, and what a regulated org's engineers expect on day one:

- Error tracking with grouping/deduplication, not one alert per occurrence.
- Session replay with production-grade fidelity and default-safe PII masking.
- Performance monitoring against Core Web Vitals, with release-based regression detection.
- Framework-agnostic instrumentation across any modern frontend stack.

### Wedge — what none of the alternatives, self-hosted or SaaS, currently do

The specific capabilities identified as absent across every product reviewed in [Competitive Analysis](README.md#competitive-analysis):

- Session intelligence with an access-audit-trail — who on your team viewed this customer's session, and when.
- Security monitoring correlated into the same investigation timeline as replay, errors, and performance — not a separate tool.
- Retention mapped to regulatory clocks, not a single global TTL.
- Evidence export built for an auditor or regulator, not just an engineer.

The platform will enable organizations to answer questions such as:

- Why did a user experience a failure?
- Which release introduced a regression?
- Which API call caused the issue?
- How did the user reach the failure state?
- Was regulated customer data exposed, and to whom?
- Who on our own team accessed this customer's session, and was that access appropriate?
- Is the issue related to security, performance, or application logic?

Without sending sensitive telemetry outside their environment, and without assembling the answer by hand across four separate tools.

---

## Who We Serve

Our primary audience includes:

- Financial institutions.
- Fintech companies.
- Insurance organizations.
- Government agencies.
- Healthcare providers.
- Enterprise software teams.
- Organizations operating in regulated environments.

These organizations prioritize data ownership, security, compliance, reliability, and operational transparency above convenience — and they are currently choosing between SaaS incumbents that solve depth but not ownership, and self-hosted alternatives that solve ownership but not compliance workflow. Imora exists to be the one alternative that doesn't force that choice.

---

## Guiding Principles

### Data Ownership First — Parity

Organizations should fully own their telemetry, including the control plane — not just a dedicated instance running inside a vendor's cloud.

Customer data should remain within the organization's infrastructure unless explicitly configured otherwise, and it should be exportable in an open, documented schema so self-hosting does not become a new form of lock-in.

---

### Security by Default, Not Opt-In Redaction — Parity, with a Wedge Edge

Security must not be an enterprise add-on, and PII protection must not depend on a team remembering to configure a selector. The best tools in this space (Sentry, FullStory) already mask aggressively by default — that is the baseline we should meet, not a differentiator on its own.

Every deployment should default to:

- Deny-by-default session capture — rendering only an explicit allow-list of safe fields, so an unredacted new field fails closed, not open.
- Encryption at rest and in transit.
- Fine-grained, role-based access control down to the field level.
- Audit logging of access to sensitive records, not just system events — who viewed this customer's data, and when. This is the part no frontend observability tool ships today, which makes it Wedge rather than Parity.

---

### Compliance Is a Workflow, Not a Checkbox — Wedge

Most platforms treat compliance as a marketing bullet. We treat it as an operational feature with real mechanics:

- Retention policies mapped to regulatory clocks — GDPR's purpose-bound storage limitation, HIPAA's 6-year documentation floor, PCI-DSS's 12-month minimum audit trail retention, SOX's 7-year requirement — rather than one global TTL.
- Legal hold support that overrides normal retention when an investigation requires it.
- One-click evidence export — a defensible, timestamped incident package combining replay, errors, security signal, and access logs — built for handing to an auditor or regulator, not just an engineer.

---

### Investigation Over Metrics — Parity, with a Wedge Edge

Metrics are useful. Answers are better.

The platform should correlate session replay, errors, performance, and security signals into a single incident timeline, rather than requiring engineers to manually cross-reference separate dashboards to reconstruct what happened. Correlating replay with errors and performance is table stakes among the better tools already; correlating security signal into that same timeline is not, and is where this principle earns its Wedge status.

---

### Open Standards — Parity

Whenever possible, the platform should embrace open standards and interoperability.

Organizations should not become dependent on proprietary protocols, closed ecosystems, or an undocumented storage format that makes migrating away from Imora itself a compliance risk.

---

### Operational Simplicity — Parity

Deployment and operation should be straightforward.

Small teams should be able to deploy the platform on a single machine. Large enterprises should be able to scale it across clusters and regions, including fully air-gapped environments with no outbound dependency for core function. The operational model should remain predictable at every stage.

---

### Framework Agnostic — Parity

The platform should support any modern frontend technology stack.

Observability should not depend on whether a team uses React, Vue, Angular, Svelte, Solid, or plain JavaScript.

---

## Long-Term Vision

We envision a future where organizations can fully understand the health, behavior, reliability, and security of their frontend systems without compromising privacy, compliance, or control.

Our goal is not to invent a new category. It is to be the alternative a regulated organization's CISO, DPO, and engineering lead can all agree on — because it's a credible replacement for the observability tool their engineers already know how to use, and because it's the only alternative on the list that also answers the questions their compliance team is legally required to ask.

Being "self-hosted Sentry" or "self-hosted LogRocket" is necessary but not sufficient — that only matches the parity bar every other self-hosted alternative already clears. The reason to choose Imora specifically is the wedge: visibility, ownership, and provable compliance in one product, which today requires stitching together an observability tool and a set of compliance processes that don't talk to each other.

---

## Problem Statement

> Status: Research-based, current as of July 2026. Builds on [Vision](README.md#vision) and [Competitive Analysis](README.md#competitive-analysis) — this document exists to quantify the problem those two describe, with sourced numbers instead of assertions.

---

### The Problem, Stated Plainly

Regulated organizations cannot get deep frontend observability without either sending customer data to a third party or giving up most of the visibility that third party would have provided. Every available option — described in detail in [Competitive Analysis](README.md#competitive-analysis) — forces a version of this trade-off. That forced choice has a real, measurable cost, and increasingly a legal one, not just an inconvenience.

---

### Cost Driver 1 — Regulated Industries Already Pay the Highest Breach Costs in the Economy

Per the IBM Cost of a Data Breach Report 2025 (Ponemon Institute, 600 organizations across 17 industries and 16 countries):

| Metric | Figure |
|---|---|
| Average US breach cost (all industries) | **$10.22M** — an all-time high |
| Global average breach cost | $4.44M (first decline in 5 years) |
| **Healthcare** average breach cost | **$7.42M** — highest of any industry, 15 consecutive years running |
| Healthcare — time to identify and contain | **279 days** (vs. 241-day global average) |
| **Financial services** average breach cost | **$5.56M** — second highest industry |

These are exactly Imora's target sectors. A platform that keeps customer session data, error context, and security signal inside the organization's own infrastructure directly reduces the blast radius and disclosure obligations of the sector that already pays the most per incident, and takes the longest to even notice one is happening.

Sources: [IBM Cost of a Data Breach Report 2025](https://www.ibm.com/reports/data-breach), [Cost of a data breach: the healthcare industry — IBM](https://www.ibm.com/think/insights/cost-of-a-data-breach-healthcare-industry), [Average Cost of a Data Breach 2026 — 46 Facts from IBM & Verizon](https://cnicsolutions.com/statistics/data-breaches-research/average-cost-of-a-data-breach-statistics-2026/).

---

### Cost Driver 2 — Third-Party Session Replay Is Now an Active Litigation Risk, Independent of Any Breach

This is a sharper and more current problem than "compliance risk" in the abstract. Since roughly 2022, and accelerating through 2025–2026, plaintiffs' firms have been suing companies for running third-party session-replay and pixel-tracking scripts (FullStory, Hotjar, Microsoft Clarity, Mouseflow, LogRocket are all named in filed complaints) under state wiretapping statutes — most prominently California's Invasion of Privacy Act (CIPA § 631).

The legal theory: capturing a visitor's keystrokes, form entries, and mouse movements and transmitting them to a third party in real time is framed as "electronic eavesdropping" without the visitor's meaningful consent — the same statute originally written for telephone wiretaps.

What makes this acute right now:

- **~1,500 CIPA lawsuits** were filed in the 18 months before August 2025, largely from automated scanning that flags any site running a recognizable session-replay script.
- Plaintiffs' firms (Pacific Trial Attorneys, Swigart Law Group, Kind Law, among others) have started **stacking a second claim** under CIPA § 638.51 (pen registers) alongside § 631, specifically so the case survives even if a court dismisses the wiretapping claim — a defense that used to work reliably no longer does.
- Statutory damages are **$5,000 per violation** — and each individual session can be argued as a separate violation, which is what makes these viable as class actions rather than one-off suits.

**This is a distinct pain point from GDPR/HIPAA/PCI-DSS compliance**, and it applies even to organizations with no breach, no regulator inquiry, and a fully "compliant" SOC 2 vendor relationship. The exposure comes from the act of routing session data through a third party's servers at all — which is exactly what every SaaS incumbent in [Competitive Analysis](README.md#competitive-analysis) requires by design. Self-hosting doesn't automatically eliminate CIPA exposure (consent and disclosure obligations still apply regardless of where the data lives), but it removes the specific fact pattern plaintiffs are currently winning on: *a third-party company is intercepting and storing the communication.*

Sources: [Zimmerman Reed: Session Replay and Pixel Tracking Class Actions](https://captaincompliance.com/education/zimmerman-reed-llp-inside-the-firm-turning-session-replay-and-pixel-tracking-into-billion-dollar-class-actions/), [Kind Law: Session Replay Software and Meta Pixel Digital Wiretapping Lawsuits](https://captaincompliance.com/education/kind-law-how-session-replay-software-and-meta-pixel-are-fueling-a-new-wave-of-digital-wiretapping-lawsuits/), [Consumer Privacy Lawsuit Roundup 2026: CIPA to COPPA](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa), [CIPA Compliance — Secure Privacy](https://secureprivacy.ai/blog/cipa-california-invasion-of-privacy-act-compliance), [California Wiretap Law: Session Replay — Lokker](https://lokker.com/privacy-law/cipa).

---

### Cost Driver 3 — Tool Fragmentation Has a Quantifiable Incident-Response Tax

Separate from legal and breach-cost exposure, running error tracking, session replay, performance monitoring, and security monitoring as separate tools has a measured operational cost during live incidents:

- Context-switching between disconnected tools during an incident adds an estimated **20–40% to incident resolution time**, because responders spend the first period of an incident just establishing a shared timeline across tools rather than diagnosing the issue.
- On a P1 incident costing roughly $5,000/minute in revenue impact — a realistic figure for the transaction-processing systems banks, insurers, and healthcare portals run — that context-switching tax alone is **$60K–$120K per hour** of incident.
- Teams running the typical 4–8 observability tools report spending **$100K–$400K per year** on tooling before any of it has resolved a single incident faster.

This is the operational argument for the "single workspace" claim in [Vision](README.md#vision): it is not a UX nicety, it is a measured tax that scales with incident frequency and severity, and it compounds the regulatory pressure — a slower investigation is also a slower breach-notification timeline, which itself carries penalties under GDPR (72-hour notification) and most US state breach laws.

Sources: [Your MTTR Is High Because Your Observability Is Fragmented](https://opstree.com/blog/mttr-is-high-because-your-observability-is-fragmented/), [The True Cost of Observability Tool Sprawl in 2026](https://oneuptime.com/blog/post/2026-02-28-true-cost-of-observability-tool-sprawl/view), [The Fragmentation Tax: What Multi-Tool Incident Response Is Really Costing You — Selector](https://www.selector.ai/blog/the-fragmentation-tax-what-multi-tool-incident-response-is-really-costing-you/).

---

### Why the Obvious Fixes Don't Fix It

- **"Just buy the enterprise/dedicated-cloud tier"** doesn't solve Cost Driver 1 or 2 — the data and control plane still live outside the organization, and the third party is still the one intercepting the session (Category 1 in [Competitive Analysis](README.md#competitive-analysis)).
- **"Just self-host an open-source tool"** solves the infrastructure-ownership half of Cost Driver 1, and removes the third-party-interception fact pattern behind Cost Driver 2 — but doesn't solve Cost Driver 3, because none of the self-hostable options combine replay, errors, performance, and security in one product (Category 2 and 3 in [Competitive Analysis](README.md#competitive-analysis)), and none of them produce the audit trail or regulatory-clock retention a compliance team needs to actually prove what happened after the fact. This is why "just self-host OpenReplay or PostHog" isn't a full answer either — it clears the parity bar and half of Driver 2, but not Driver 3.
- **"Just add more tooling"** makes Cost Driver 3 worse, not better — it's the thing causing the fragmentation tax in the first place.

---

### Consequence of Inaction

Every quarter a regulated organization runs frontend observability through fragmented, third-party-hosted tools, it is carrying three compounding exposures simultaneously: the highest breach-cost profile in the economy (Driver 1), an active and currently-winning class-action theory that applies regardless of breach status (Driver 2), and a quantifiable, recurring operational tax on every incident it investigates (Driver 3). None of these are solved by adding another point solution — they are solved by changing where the data lives and how many tools an investigation has to cross.

This is why Imora is scoped as an alternative to the tools already in use — Sentry, Datadog RUM, LogRocket, FullStory, and their self-hosted counterparts — rather than a net-new category, per [Vision](README.md#vision)'s Positioning section: the fastest way to zero out these three drivers is a credible swap-in replacement that also closes the gaps none of those alternatives close today.

---

## Target Users

> Status: Research-based, current as of July 2026. Builds on [Vision](README.md#vision) and [Problem Statement](README.md#problem-statement) — this document identifies *who* carries each of the three cost drivers already quantified, using real regulatory role definitions rather than invented titles.

---

### Why Personas, Not Just "Regulated Industries"

Enterprise security and observability purchases are made by a buying committee, not a single buyer — typically 7–10 stakeholders, split across roles that read the same pitch completely differently: security reads strategic risk, engineering reads technical fit, procurement reads commercial terms. Selling (or designing) for all of them with one undifferentiated message is a documented failure mode. Imora has at least five distinct personas in a regulated org, each of whom owns a different piece of the problem.

Source: [Cybersecurity Buyer Personas: CISO, CIO, and Security Team](https://getgangly.com/blog/cybersecurity-buyer-personas).

Per [Vision](README.md#vision)'s Positioning section, Imora is scoped as an alternative to tools these personas already use, plus specific compliance capabilities none of those alternatives have. That split shows up directly in the personas below: Personas 1–5 are the reason a regulated org switches to that alternative at all (the wedge); Persona 6 is the reason engineers keep using it once they have (parity).

---

### Persona 1 — CISO / Head of Security (Risk Owner, Economic Buyer)

**Who they are:** accountable for the organization's overall breach exposure and security posture. In fintechs under ~300 employees, this person frequently also owns incident-response command directly, rather than delegating it.

**What they carry:** the full weight of [Problem Statement](README.md#problem-statement) Cost Driver 1 (breach cost — $7.42M average in healthcare, $5.56M in financial services) and Cost Driver 2 (CIPA wiretapping litigation exposure from third-party session-replay vendors) lands on this role's budget and reputation.

**What they need from Imora:** a platform where they can state, truthfully, that customer session data never left the organization's infrastructure — removing the specific fact pattern (a third party intercepting the session) that CIPA plaintiffs are currently winning on, and reducing the blast radius that drives breach-cost figures.

Sources: [Incident Commander: Roles, Responsibilities and Best Practices — Rootly](https://rootly.com/incident-response/incident-commander), [Incident Response Plan Template for Fintech](https://risktemplate.com/blog/2026-03-19-incident-response-plan-template-fintech/).

---

### Persona 2 — Data Protection Officer / Privacy Officer (Compliance Gatekeeper)

**Who they are:** a legally distinct, independent role — not a subset of security. Under GDPR Article 37, appointing a DPO is *mandatory*, not optional, for any organization that is a public authority, conducts large-scale systematic monitoring of individuals, or processes special-category data at scale — which covers essentially every government agency, insurer, and healthcare provider Imora targets. The DPO reports to the organization's highest management level and cannot be instructed by the controller on how to do their job — a legally protected independence most other stakeholders don't have.

**Their responsibilities directly relevant to Imora:** advising on Data Protection Impact Assessments, serving as the contact point for regulators on breach reporting, and handling Data Subject Access Requests (DSARs) — a person asking "what data do you have on me, and who has looked at it."

**What they carry:** GDPR's storage-limitation principle (data retained no longer than necessary) and the DSAR obligation are exactly the workflow gap identified in [Competitive Analysis](README.md#competitive-analysis) — no frontend observability product maps retention to this, and none can answer "who on your team viewed this person's session" in the form a DSAR response requires.

**What they need from Imora:** retention policy mapped to actual regulatory clocks, and an access-audit-trail they can hand a regulator or a data subject directly, instead of reconstructing it by hand.

Sources: [What are the responsibilities of a DPO? — European Commission](https://commission.europa.eu/law/law-topic/data-protection/rules-business-and-organisations/obligations/data-protection-officers/what-are-responsibilities-data-protection-officer-dpo_en), [The DPO role under GDPR — GRC Solutions](https://grcsolutions.io/data-protection-officer-dpo-under-the-gdpr/).

---

### Persona 3 — HIPAA Security Officer (Healthcare-Specific Technical Compliance Owner)

**Who they are:** a role mandated by the HIPAA Security Rule, distinct from the HIPAA Privacy Officer. Where the Privacy Officer governs *who is allowed to see* PHI, the Security Officer is responsible for the technical safeguards protecting *electronic* PHI specifically — access controls, audit logging, encryption, and an annual documented risk assessment covering third parties and business associates.

**What they carry:** this persona is the most literal match for the access-audit-trail gap in [Competitive Analysis](README.md#competitive-analysis) — "implementing the administrative, physical, and technical safeguards required by the Security Rule" is, in the frontend-observability context, precisely "who on our team viewed this patient's session, and when."

**What they need from Imora:** field-level access control and audit logging that satisfies an annual HIPAA risk assessment out of the box, not as a custom integration the Security Officer has to build against a generic audit log.

Source: [HIPAA Security Officer — 2026 Update](https://www.hipaajournal.com/hipaa-security-officer/), [HIPAA Privacy Officer vs. Security Officer](https://www.foxgrp.com/hipaa-compliance/hipaa-privacy-officer-vs-security-officer/).

---

### Persona 4 — VP Engineering / Head of Platform Engineering / SRE Lead (Technical Champion, Operator)

**Who they are:** owns the decision of what gets deployed and who operates it long-term. This is the persona evaluating self-hosting operational burden directly — Docker vs. Kubernetes, single-machine vs. multi-region, and whether a small platform team can actually run this without becoming a full-time job.

**What they carry:** Cost Driver 3 from [Problem Statement](README.md#problem-statement) — the fragmentation tax (20–40% added incident-resolution time, $100K–$400K/year in tool sprawl) is this persona's budget line and their team's on-call quality of life.

**What they need from Imora:** the "Operational Simplicity" guiding principle from [Vision](README.md#vision) to actually hold — a single machine for a small team, clusters/regions for a large one, with a predictable operational model at every stage. This persona will reject anything that trades data ownership for an unmaintainable deployment.

---

### Persona 5 — Incident Commander / On-Call Engineer (Day-to-Day User)

**Who they are:** the person actually inside the tool during a live incident. In regulated environments specifically, this role has an added burden most generic SRE guidance doesn't cover: **chain-of-custody documentation**, because a production incident touching customer data can become litigation, and evidence handling has to be defensible after the fact, not reconstructed from memory weeks later.

**What they carry:** this is the persona who feels tool fragmentation first-hand — establishing a shared timeline across four disconnected tools before diagnosis can even start — and the one who has the most to gain from the "Investigation Over Metrics" guiding principle in [Vision](README.md#vision).

**What they need from Imora:** one workspace that already correlates replay, errors, performance, and security signal into a single timeline, so the first twenty minutes of an incident aren't spent building that timeline by hand — and evidence export that produces a defensible record without extra work, in case the incident does become litigation.

Source: [Overview of Incident Lifecycle in SRE — Squadcast](https://www.squadcast.com/blog/overview-of-incident-lifecycle-in-sre), incident-response role research cited under Persona 1.

---

### Persona 6 — Senior Frontend Engineer (Daily User, No Compliance Angle)

**Who they are:** the person who opens the tool on an ordinary Tuesday with no incident, no audit, and no regulator involved — investigating why LCP crept up after last night's deploy, triaging a spike in a JS error, or reproducing a bug a support ticket described badly. This persona doesn't appear in [Problem Statement](README.md#problem-statement)'s three cost drivers at all, and that is deliberate: the first five personas explain why a regulated org would *buy* Imora, but adoption and daily retention depend on whether this person actually *wants to open it*. A compliance-mandated tool that engineers route around is a common enterprise-software failure mode, and this persona is the check against it.

**What "good" looks like, day to day, grounded in what the category already competes on:**
- **Performance regressions judged against Core Web Vitals** — LCP under 2.5s, INP under 200ms, CLS under 0.1 are the "good" thresholds Google evaluates at the 75th percentile of real user data; this engineer needs to see when a release pushes a page past those thresholds, not just an average that hides it.
- **Signal over noise** — a 2025 Catchpoint study found 62% of on-call engineers have ignored a critical alert because it was buried in noise; this persona needs errors grouped/deduplicated by root cause, not one alert per affected user.
- **"Which release did this" answered automatically** — the category standard (Sentry's regression detection) already ties a performance or error regression to the specific release that introduced it, using statistical trend detection rather than manual bisection; this persona expects that as table stakes, not a differentiator.

**What they need from Imora:** a debugging experience that is at least as good as the best point solution it's replacing (Sentry for errors, Datadog for performance) on its own merits — independent of any compliance capability — because this is the persona whose daily use makes the platform worth having at all.

Sources: [How the Core Web Vitals metrics thresholds were defined — web.dev](https://web.dev/articles/defining-core-web-vitals-thresholds), [Understanding Core Web Vitals — Google Search Central](https://developers.google.com/search/research/appearance/core-web-vitals), [Alert Fatigue in SRE and DevOps — Sensu](https://sensu.io/blog/alert-fatigue-in-sre-and-devops), [Sentry Endpoint Regression docs](https://docs.sentry.io/product/issues/issue-details/performance-issues/endpoint-regressions/), [Sentry Releases docs](https://docs.sentry.io/product/releases/).

---

### Org-Size Variants — Personas Collapse in Smaller Organizations

Role separation scales with headcount, and Imora needs to work whether these are five people or one:

- **Under ~50 employees (early-stage fintech, small regulated startup):** the CTO typically wears the CISO, VP Engineering, and Incident Commander hats simultaneously. Imora needs to be usable by one technically capable generalist, not just a fully staffed platform team.
- **~50–300 employees:** incident-response ownership typically lands with the CISO or Head of Security, or — if that role doesn't exist yet — the VP of Engineering or Head of Compliance. Personas 1 and 4 are frequently the same person here.
- **300+ employees / large regulated enterprise:** all five buyer-side personas above (1–5; Persona 6 doesn't collapse into other roles the way buyer roles do — an engineer stays an engineer regardless of company size) are typically distinct people with separate reporting lines, and the DPO/HIPAA Security Officer's legally protected independence (Persona 2/3) becomes a real constraint on how the product's access control and audit trail must be designed — those roles need to be able to pull records without depending on engineering's cooperation.

Source: [Incident Response Plan Template: What Every Fintech Needs](https://risktemplate.com/blog/2026-03-19-incident-response-plan-template-fintech/).

---

### Mapping Personas to Cost Drivers

| Persona | Primary Cost Driver Carried | What They Need From Imora |
|---|---|---|
| CISO / Head of Security | Driver 1 (breach cost) + Driver 2 (CIPA litigation) | No third-party interception of session data |
| DPO / Privacy Officer | Driver 1 + Driver 2, plus DSAR/regulatory obligations | Regulatory-clock retention, access-audit-trail, evidence export |
| HIPAA Security Officer | Driver 1 (healthcare-specific) | Field-level access control and audit logging out of the box |
| VP Engineering / Platform Lead | Driver 3 (fragmentation tax) + deployment burden | Genuine operational simplicity at any scale |
| Incident Commander / On-call Engineer | Driver 3 (fragmentation tax) directly, day to day | Single correlated investigation workspace, defensible evidence export |
| Senior Frontend Engineer | None of the three — this is the adoption/retention check, not a cost driver | Debugging UX at parity with the best point solution, on its own merits |

Persona 6 is deliberately included above with "none of the three" rather than left out of the table entirely: Personas 1–5 explain why a regulated org signs the contract, but none of them use the product daily. If the roadmap optimizes only for Personas 1–5, Imora risks becoming a compliance-mandated tool engineers route around rather than one they choose to open — a documented failure mode in enterprise security/compliance software. Persona 6's row exists specifically to keep that risk visible in the table, not just in prose.

This table should directly inform persona sections in `research/01-product/README.md#user-personas` and the JTBD framing in `research/01-product/README.md#user-stories` — those are the next logical files once this one is in place.

---

## Competitive Analysis

> Status: Research-based, current as of July 2026. Every claim below is sourced from public docs, changelogs, or GitHub issues — not assumption. Where a product's roadmap may have moved since, the source is linked so the claim can be re-verified rather than taken on faith.

This document backs the claims made in [Vision](README.md#vision)'s "Gap in Today's Landscape" section. It exists so those claims stay honest as the market moves — re-check the sources here before repeating a claim from memory.

---

### Method

Eleven products were researched across three questions that matter specifically to regulated buyers (finance, healthcare, insurance, government):

1. Can the org run the full stack — including the control plane, not just a "dedicated" tenant — inside its own infrastructure?
2. Does session capture default to safe (deny-by-default / aggressive masking), or does a missed configuration leak PII?
3. Does the product answer "who on our team viewed this customer's data" — and does retention map to actual regulatory clocks (GDPR, HIPAA, PCI-DSS, SOX) rather than one global TTL?

No product answers all three. That gap is the basis for Imora's positioning, not a marketing assumption.

### How This Document Is Used

Per [Vision](README.md#vision)'s Positioning section, Imora is scoped as **an alternative to the products below, not a new category** — so every product here is read for two things: what it takes to be a *credible* alternative (parity), and what none of them do (the wedge). The tables below are written to answer both questions directly, and the Synthesis at the end sorts findings into that split explicitly.

---

### Category 1 — SaaS Incumbents (deep observability, no true self-hosting)

| Product | Self-hostable? | Session replay masking | Compliance posture |
|---|---|---|---|
| **Datadog RUM** | No — cloud-only; session replay is browser-only, no on-prem option found | Configurable privacy controls; data stored on Datadog-managed cloud, encrypted at rest | HIPAA-eligible *with a signed BAA*, but customers must restrict workloads to eligible services and disable non-covered features |
| **New Relic Browser** | No — FedRAMP Moderate and HIPAA enablement exist, but as account-level compliance programs on New Relic's cloud, not a self-hosted deployment | Not the focus of research; New Relic markets "on-premises and cloud" support at the infrastructure-monitoring layer, not specifically for Browser/RUM self-hosting | FedRAMP Authorized (Moderate, partial service scope), HIPAA account enablement available |
| **LogRocket** | No — SaaS only, data lands on US-based servers by default | Strong: PII Labeling API, inline/blur masking, passwords never captured by default | SOC 2 Type II; markets HIPAA/GDPR/CCPA support, but data residency requires explicit cross-border transfer mechanisms |
| **FullStory** | No — SaaS only | Best-in-class default: "Private by Default" captures no text unless explicitly allow-listed; Exclude/Mask/Unmask tiers | SOC 2 Type 2, ISO 27001/42001, SOC 1/2/3 — strong certification story, but no confirmed HIPAA BAA |
| **Sentry (Cloud)** | Cloud product is not self-hosted by definition | Aggressive default masking — all text/images redacted client-side before leaving the browser | Standard SaaS compliance program |

**Takeaway:** this tier has the deepest product capability and, in FullStory's and Sentry's case, genuinely good default-safe capture. None of them solve the actual constraint regulated buyers have: the data and control plane living outside the organization's perimeter. A BAA or a FedRAMP badge is a contractual promise about a third party's environment, not organizational control over the data.

Sources: [Datadog HIPAA compliance](https://docs.datadoghq.com/data_security/hipaa_compliance/), [Datadog HIPAA-eligible services](https://www.datadoghq.com/legal/hipaa-eligible-services/), [Datadog RUM data security](https://docs.datadoghq.com/data_security/real_user_monitoring/), [New Relic HIPAA](https://docs.newrelic.com/research/security/security-privacy/compliance/certificates-standards-regulations/hipaa/), [New Relic FedRAMP](https://docs.newrelic.com/research/security/security-privacy/compliance/certificates-standards-regulations/fedramp/), [LogRocket Privacy docs](https://docs.logrocket.com/research/privacy), [LogRocket Security docs](https://docs.logrocket.com/research/security), [FullStory Private by Default](https://help.fullstory.com/hc/en-us/articles/360044349073-Fullstory-Private-by-Default), [FullStory Trust Center](https://trust.fullstory.com/), [Sentry: Protecting User Privacy in Session Replay](https://docs.sentry.io/security-legal-pii/scrubbing/protecting-user-privacy/).

---

### Category 2 — Open-Source, Self-Hosted Session-Replay Tools

These are the closest existing comparators to what Imora is building on the "session intelligence" axis.

| Product | Self-hostable? | Default PII masking | Notable caveat |
|---|---|---|---|
| **OpenReplay** | Yes — fully self-hosted including the ingestion/processing pipeline, "the only digital experience platform that can be fully self-hosted" per their own docs | Good: sanitizes at the tracker (browser) level before data ever reaches the server; emails auto-obscured by default | Own docs explicitly caveat that self-hosting alone ≠ GDPR compliance — still requires consent flows and DSAR handling on top |
| **PostHog** | Yes — open-source self-hosted is "the same exact product" as PostHog Cloud, per their own FAQ | Input elements masked by default | Session recording volume/retention is more constrained on self-hosted OSS than on Cloud; some enterprise controls are cloud/paid-tier only |
| **Highlight.io** | Yes, historically — Docker self-host, regex-based PII obfuscation and CSS-class masking by default | Reasonable default: obfuscates inputs and common PII regex patterns (SSNs, phone numbers, addresses) out of the box | **Acquired by LaunchDarkly in March 2025.** The standalone hosted product shuts down February 28, 2026, with existing accounts migrated into LaunchDarkly's observability platform. Not a safe long-term reference point — cite with this caveat every time. |

**Takeaway:** these products already do reasonably safe default capture — this is not the open gap it might look like from the outside, and Imora shouldn't market "default-safe capture" as if it invented the idea. The actual gap is one layer up: none of these three answer *"who on our team viewed this session,"* none map retention to a specific regulation's clock, and none produce an audit-ready evidence package. They solve "can I run this myself," not "can I prove to an auditor what happened and who saw it."

Sources: [OpenReplay Sanitize Data docs](https://docs.openreplay.com/en/sdk/sanitize-data/), [OpenReplay product page](https://openreplay.com/), [PostHog self-host docs](https://posthog.com/research/self-host), [PostHog open-source disclaimer](https://posthog.com/research/self-host/open-source/disclaimer), [Highlight.io GitHub](https://github.com/highlight/highlight), [Highlight.io privacy docs](https://www.highlight.io/research/getting-started/client-sdk/replay-configuration/privacy), [Bugsink: a self-hosted alternative to Highlight.io](https://www.bugsink.com/a-self-hosted-alternative-to-highlight-io/), [7 Best Self-Hosted Session Replay Tools 2026](https://temps.sh/blog/best-self-hosted-session-replay-tools-2026).

---

### Category 3 — Adjacent Self-Hosted Tools (not real comparators, but worth naming so we don't overclaim)

| Product | What it actually is | Why it's not a session-intelligence competitor |
|---|---|---|
| **SigNoz** | Open-source, OpenTelemetry-native APM: logs, traces, metrics, exceptions | No native frontend RUM or session replay. It's a backend observability competitor to Datadog/New Relic, not to LogRocket/Highlight — there is nothing here to redact because there's no session capture in the first place. |
| **Grafana Faro** | Web SDK for RUM (errors, web vitals, traces) that feeds a self-run Loki/Tempo collector | No built-in session replay. Security writeups about Faro focus on *how to avoid* leaking PII into custom event attributes, not on redacting a replay — because there is no replay to redact. |
| **GlitchTip** | Lightweight, self-hosted, Sentry-SDK-compatible error tracker (Django, runs on 1 vCPU / 1GB RAM) | Explicitly error-tracking only by design. Calls to Sentry's Session Replay or Profiling APIs against a GlitchTip backend fail silently. |

**Takeaway:** it would be inaccurate to describe these as "self-hosted alternatives that get compliance wrong" — they simply don't operate in the session-replay/PII space at all. They're evidence of a different point: even within the self-hosted world, "frontend session intelligence" and "backend telemetry ownership" are two separate product categories that rarely combine, which is itself part of the fragmentation problem in [Vision](README.md#vision).

Sources: [SigNoz](https://signoz.io/), [SigNoz GitHub](https://github.com/signoz/signoz), [Grafana Faro Web SDK](https://github.com/grafana/faro-web-sdk), [Frontend RUM Security: Grafana Faro](https://www.systemshardening.com/articles/observability/frontend-rum-security-grafana-faro/), [GlitchTip](https://glitchtip.com/), [GlitchTip vs Sentry vs Bugsink](https://www.bugsink.com/blog/glitchtip-vs-sentry-vs-bugsink/).

---

### Where "Sentry" specifically stands (a nuance worth keeping straight)

Sentry is source-available (Business Source License, not OSI open source) and can be self-hosted for free — but as of 2026, **Session Replay is still not reliably available on self-hosted Sentry.** The feature has an open tracking issue from the Sentry team going back multiple years, and users report replays either 404ing after 24 hours or landing in Kafka/ClickHouse without ever rendering in the UI. So the one SaaS incumbent with genuinely good default-safe masking (Sentry Cloud) doesn't let a regulated org self-host that specific capability at all today.

Sources: [Sentry self-hosted issue #1873 — Include Session Replay in Self Hosted](https://github.com/getsentry/self-hosted/issues/1873), [Sentry self-hosted issue #3963 — Replay Not Found](https://github.com/getsentry/self-hosted/issues/3963), [Sentry self-hosted issue #3274 — Replay detail 404](https://github.com/getsentry/self-hosted/issues/3274).

---

### Where the "access audit trail" pattern already exists (just not here)

The idea of logging *who viewed a specific recording* is not novel — it's standard in Privileged Access Management (PAM) tooling, where session recordings of infrastructure access are treated as auditable evidence: BeyondTrust, Delinea, and Keeper all log who accessed a recorded session, when, and why, and gate playback with role-based access control.

No frontend session-replay or observability product researched in Category 1 or 2 applies this same pattern to *customer* session data. That's a borrowed idea, not an invented one — which makes it a safer bet: it is a proven pattern in an adjacent, mature category (PAM), not a hypothetical.

Sources: [Audit Logs and Session Replay — hoop.dev](https://hoop.dev/blog/audit-logs-and-session-replay-the-powerful-duo-for-debugging-security-and-compliance), [BeyondTrust: Audit Recorded Sessions](https://www.beyondtrust.com/research/privileged-identity/app-launcher-and-recording/audit.htm), [Delinea: Session Recording](https://delinea.com/products/secret-server/features/session-recording).

---

### Regulatory retention clocks (the actual numbers, for reference)

Every product reviewed above exposes a single, global retention TTL. The regulations regulated buyers actually answer to don't work that way:

| Regulation | Retention requirement |
|---|---|
| **PCI-DSS** (Requirement 10.7) | Minimum 12 months of audit trail history, with at least 3 months immediately available ("hot") for analysis |
| **HIPAA** (45 CFR §164.316(b)(2)(i) / §164.530(j)) | Minimum 6 years from creation or last-effective date for policies, procedures, and audit logs |
| **GDPR** (Article 5(1)(e)) | No fixed term — data must not be kept longer than necessary for its stated processing purpose ("storage limitation") |
| **SOX** | 7 years for relevant financial records and supporting audit documentation |

A platform with one global TTL forces a regulated org to either over-retain everything to satisfy the strictest regime (cost and privacy risk) or under-retain and fail an audit. Policy-per-data-category, mapped to the regulation that actually applies, is the fix — and it's a data-model and product decision, not a marketing claim.

Sources: [Compliance Log Retention Requirements by Regulation](https://claudiasop.com/blog/compliance-log-retention-requirements.html), [IT Log Retention: Complete Compliance Guide 2026](https://techjacksolutions.com/security/it-log-and-record-retention-requirements/).

---

### The Parity Checklist — What Makes Imora a Credible Alternative at All

Pulled directly from the tables above: this is what a regulated team's engineers already expect from Sentry, Datadog RUM, LogRocket, FullStory, OpenReplay, or PostHog, and what Imora has to match before the wedge below matters to anyone.

| Capability | Bar set by |
|---|---|
| Error tracking with grouping/deduplication | Sentry (category standard) |
| Session replay, production-fidelity | LogRocket, FullStory, OpenReplay |
| Default-safe PII masking (not opt-in) | FullStory ("Private by Default"), Sentry (aggressive default masking) |
| Performance monitoring against Core Web Vitals with release-based regression detection | Sentry Releases/regression issues, Datadog RUM |
| Self-hosted deployment (Docker/Kubernetes, single-machine to cluster) | OpenReplay, PostHog, GlitchTip |
| Framework-agnostic SDKs | All products reviewed |

None of this is differentiation — it's the entry price. A product that skips any row here isn't a credible alternative to what these teams already run, regardless of what it adds on top.

---

### Synthesis — The Wedge, Ranked by How Uncontested It Is

Everything below is the reason to pick Imora over *any* of the products above — including the other self-hosted ones — not just over the SaaS incumbents:

1. **Access-audit-trail for session data** (who viewed this customer's replay, when, why) — zero frontend products found doing this; proven pattern exists in PAM, uncontested to claim.
2. **Retention mapped to regulatory clocks**, not one global TTL — zero products found doing this; requirements are public and verifiable (table above), uncontested to claim.
3. **One-click, cross-signal evidence export** for auditors (replay + errors + security signal + access log as one timestamped package) — zero products found doing this; slightly softer claim since "evidence export" as a category is harder to search for than the other two, so treat as *strong* rather than *certain*.

What is **not** a safe wedge claim, and belongs in the parity checklist instead: "default-safe / deny-by-default capture." FullStory and Sentry already do this well. Imora should meet that bar, not market it as novel.

The parity checklist and this wedge ranking should directly inform `research/08-roadmap/feature-roadmap.md` and `research/01-product/README.md#product-requirements-document-prd` — parity defines the MVP surface area, the wedge defines why a regulated buyer picks Imora over the closest self-hosted alternative once that surface area exists.

---

## Glossary

> Status: The single canonical glossary for the whole project — consolidates terms precisely defined across `00-overview/`, `01-product/`, and `02-domain/` into one reference every stakeholder can use, not just readers of any one folder. Every definition here traces back to a specific prior document — this file doesn't introduce new meaning, it collects it.

---

### Positioning Terms

**Parity** — a capability necessary to be a credible alternative to Sentry, Datadog RUM, LogRocket, FullStory, OpenReplay, or PostHog at all. Defined in [Vision](README.md#vision)'s Positioning section; the full checklist is in [Competitive Analysis](README.md#competitive-analysis).

**Wedge** — a capability none of those alternatives, self-hosted or SaaS, currently ship. The three uncontested wedge gaps (access-audit-trail, regulatory-clock retention, evidence export) are ranked in [Competitive Analysis](README.md#competitive-analysis)'s Synthesis.

**Alternative** — Imora's own positioning, per [Vision](README.md#vision): not a new category, a swap-in replacement for tools regulated teams already use, plus the wedge.

---

### Personas (shorthand used from [User Personas](../01-product/README.md#user-personas) onward)

| Name | Role | Carries |
|---|---|---|
| Dara | CISO, regional bank | Breach cost + CIPA litigation exposure |
| Adaeze | DPO, national insurer | DSAR/breach-notification deadlines, GDPR storage limitation |
| Marcus | HIPAA Security Officer, hospital network | Annual risk assessment, ePHI audit controls |
| Priya | Head of Platform Engineering, fintech | Deployment/operational burden |
| Jon | Incident Commander / SRE, Priya's team | Fragmentation tax, chain-of-custody |
| Chidi | Senior Frontend Engineer, Priya's team | Daily adoption — the parity check, not a cost driver |

Full definitions in [Target Users](README.md#target-users) (roles) and [User Personas](../01-product/README.md#user-personas) (grounded scenarios).

---

### Domain Entities (full definitions in [Domain Model](../02-domain/README.md#domain-model))

- **Session** — one user's browsing session; the aggregate root for replay data.
- **SessionEvent** — one entry in a Session's rrweb-style capture stream (FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange).
- **ErrorEvent / ErrorGroup** — a captured exception, and the deduplicated root-cause it's grouped under at write time.
- **Release** — a deployed version identifier used for regression attribution.
- **PerformanceMetric** — an LCP, INP, or CLS measurement tied to a Session and Release.
- **TraceLink** — the shared session/trace identifier correlating a Session to backend spans.
- **AccessAuditEvent** — the append-only log entry produced for every VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED, or CONFIG_CHANGED against sensitive data or its governing configuration. See [Event Catalog](../02-domain/README.md#event-catalog) for the full set of named variants (SessionViewed, FieldUnmasked, RecordExported, RecordDeleted, DeletionSkippedDueToHold, ConfigurationChanged).
- **RetentionPolicy** — a per-data-category retention rule mapped to a regulatory clock (PCI-DSS 12mo, HIPAA 6yr, GDPR purpose-bound, SOX 7yr).
- **LegalHold** — a directive that overrides scheduled deletion for records matching a scope query, without creating a separate copy of the data.
- **EvidenceExport** — a frozen, self-contained, hash-verifiable copy of records generated for an incident, immune to later retention or erasure actions.
- **SecurityEvent** — a security signal optionally correlated into a Session's incident timeline.

---

### Business Rule Shorthand (full rules in [Business Rules](../02-domain/README.md#business-rules))

- **BR-1 through BR-7** — referenced by number throughout later architecture and data docs. BR-1 (longest-retention-wins), BR-2 (check-before-destroy), BR-3 (GDPR erasure vs. legal obligation, selective purging), BR-4 (export immutability), BR-5 (audit on every sensitive access), BR-6 (unmask requires reason), BR-7 (capture-time masking).
- **Selective purging** — anonymizing or deleting only the fields a competing regulation doesn't require, rather than choosing between full retention and full deletion. The resolution mechanism for BR-3.
- **Deny-by-default capture** — masking any field with no explicit allow-list rule, so an unredacted new field fails closed. The mechanism behind BR-7.

---

### Regulatory Terms

- **DSAR** — Data Subject Access Request; a GDPR right to ask what data an organization holds and who has accessed it. One-month response deadline (extendable to three for complex requests).
- **DPO** — Data Protection Officer; mandatory under GDPR Article 37 for public authorities and large-scale data processors. Legally independent from the organizations that employ them.
- **BAA** — Business Associate Agreement; the HIPAA-required contract that makes a vendor's handling of PHI compliant — a contractual promise about a third party's environment, not organizational control, per [Competitive Analysis](README.md#competitive-analysis).
- **CIPA** — California Invasion of Privacy Act; the wiretapping statute (§631, plus §638.51 pen-register claims) behind the session-replay litigation wave in [Problem Statement](README.md#problem-statement).
- **PAM** — Privileged Access Management; the adjacent tooling category (BeyondTrust, Delinea) where the access-audit-trail pattern already exists, just not applied to frontend session data.

---

### Architecture Terms (full definitions in [Bounded Contexts](../02-domain/README.md#bounded-contexts))

- **Bounded Context** — one of the eight owning service boundaries (gateway, ingestion, query-api, alert-engine, workers, browser-sdk, dashboard, notification-service).
- **Shared Kernel** — two contexts operating directly on the same entity definitions from [Domain Model](../02-domain/README.md#domain-model) (browser-sdk↔ingestion, ingestion↔query-api).
- **Customer-Supplier** — an upstream context whose downstream customer can influence its priorities, but the dependency runs one direction (e.g., gateway → query-api).
- **Conformist** — a downstream context that accepts the upstream's model as-is, with no translation authority (alert-engine → notification-service, query-api → dashboard).

---

### Why This Lives in `00-overview/`, Not `02-domain/`

A glossary spanning positioning, personas, regulatory, and architecture terms is useful to every reader from day one — a DPO evaluating the product needs "DSAR" and "Wedge" defined long before they'd ever open `02-domain/`. Keeping it in the overview folder, rather than buried in the domain-modeling folder where only engineers would naturally look, is what makes it a single canonical reference instead of one more file competing with a near-duplicate elsewhere.

