# Problem Statement

> Status: Research-based, current as of July 2026. Builds on [vision.md](vision.md) and [competitive-analysis.md](competitive-analysis.md) — this document exists to quantify the problem those two describe, with sourced numbers instead of assertions.

---

## The Problem, Stated Plainly

Regulated organizations cannot get deep frontend observability without either sending customer data to a third party or giving up most of the visibility that third party would have provided. Every available option — described in detail in [competitive-analysis.md](competitive-analysis.md) — forces a version of this trade-off. That forced choice has a real, measurable cost, and increasingly a legal one, not just an inconvenience.

---

## Cost Driver 1 — Regulated Industries Already Pay the Highest Breach Costs in the Economy

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

## Cost Driver 2 — Third-Party Session Replay Is Now an Active Litigation Risk, Independent of Any Breach

This is a sharper and more current problem than "compliance risk" in the abstract. Since roughly 2022, and accelerating through 2025–2026, plaintiffs' firms have been suing companies for running third-party session-replay and pixel-tracking scripts (FullStory, Hotjar, Microsoft Clarity, Mouseflow, LogRocket are all named in filed complaints) under state wiretapping statutes — most prominently California's Invasion of Privacy Act (CIPA § 631).

The legal theory: capturing a visitor's keystrokes, form entries, and mouse movements and transmitting them to a third party in real time is framed as "electronic eavesdropping" without the visitor's meaningful consent — the same statute originally written for telephone wiretaps.

What makes this acute right now:

- **~1,500 CIPA lawsuits** were filed in the 18 months before August 2025, largely from automated scanning that flags any site running a recognizable session-replay script.
- Plaintiffs' firms (Pacific Trial Attorneys, Swigart Law Group, Kind Law, among others) have started **stacking a second claim** under CIPA § 638.51 (pen registers) alongside § 631, specifically so the case survives even if a court dismisses the wiretapping claim — a defense that used to work reliably no longer does.
- Statutory damages are **$5,000 per violation** — and each individual session can be argued as a separate violation, which is what makes these viable as class actions rather than one-off suits.

**This is a distinct pain point from GDPR/HIPAA/PCI-DSS compliance**, and it applies even to organizations with no breach, no regulator inquiry, and a fully "compliant" SOC 2 vendor relationship. The exposure comes from the act of routing session data through a third party's servers at all — which is exactly what every SaaS incumbent in [competitive-analysis.md](competitive-analysis.md) requires by design. Self-hosting doesn't automatically eliminate CIPA exposure (consent and disclosure obligations still apply regardless of where the data lives), but it removes the specific fact pattern plaintiffs are currently winning on: *a third-party company is intercepting and storing the communication.*

Sources: [Zimmerman Reed: Session Replay and Pixel Tracking Class Actions](https://captaincompliance.com/education/zimmerman-reed-llp-inside-the-firm-turning-session-replay-and-pixel-tracking-into-billion-dollar-class-actions/), [Kind Law: Session Replay Software and Meta Pixel Digital Wiretapping Lawsuits](https://captaincompliance.com/education/kind-law-how-session-replay-software-and-meta-pixel-are-fueling-a-new-wave-of-digital-wiretapping-lawsuits/), [Consumer Privacy Lawsuit Roundup 2026: CIPA to COPPA](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa), [CIPA Compliance — Secure Privacy](https://secureprivacy.ai/blog/cipa-california-invasion-of-privacy-act-compliance), [California Wiretap Law: Session Replay — Lokker](https://lokker.com/privacy-law/cipa).

---

## Cost Driver 3 — Tool Fragmentation Has a Quantifiable Incident-Response Tax

Separate from legal and breach-cost exposure, running error tracking, session replay, performance monitoring, and security monitoring as separate tools has a measured operational cost during live incidents:

- Context-switching between disconnected tools during an incident adds an estimated **20–40% to incident resolution time**, because responders spend the first period of an incident just establishing a shared timeline across tools rather than diagnosing the issue.
- On a P1 incident costing roughly $5,000/minute in revenue impact — a realistic figure for the transaction-processing systems banks, insurers, and healthcare portals run — that context-switching tax alone is **$60K–$120K per hour** of incident.
- Teams running the typical 4–8 observability tools report spending **$100K–$400K per year** on tooling before any of it has resolved a single incident faster.

This is the operational argument for the "single workspace" claim in [vision.md](vision.md): it is not a UX nicety, it is a measured tax that scales with incident frequency and severity, and it compounds the regulatory pressure — a slower investigation is also a slower breach-notification timeline, which itself carries penalties under GDPR (72-hour notification) and most US state breach laws.

Sources: [Your MTTR Is High Because Your Observability Is Fragmented](https://opstree.com/blog/mttr-is-high-because-your-observability-is-fragmented/), [The True Cost of Observability Tool Sprawl in 2026](https://oneuptime.com/blog/post/2026-02-28-true-cost-of-observability-tool-sprawl/view), [The Fragmentation Tax: What Multi-Tool Incident Response Is Really Costing You — Selector](https://www.selector.ai/blog/the-fragmentation-tax-what-multi-tool-incident-response-is-really-costing-you/).

---

## Why the Obvious Fixes Don't Fix It

- **"Just buy the enterprise/dedicated-cloud tier"** doesn't solve Cost Driver 1 or 2 — the data and control plane still live outside the organization, and the third party is still the one intercepting the session (Category 1 in [competitive-analysis.md](competitive-analysis.md)).
- **"Just self-host an open-source tool"** solves the infrastructure-ownership half of Cost Driver 1, and removes the third-party-interception fact pattern behind Cost Driver 2 — but doesn't solve Cost Driver 3, because none of the self-hostable options combine replay, errors, performance, and security in one product (Category 2 and 3 in [competitive-analysis.md](competitive-analysis.md)), and none of them produce the audit trail or regulatory-clock retention a compliance team needs to actually prove what happened after the fact. This is why "just self-host OpenReplay or PostHog" isn't a full answer either — it clears the parity bar and half of Driver 2, but not Driver 3.
- **"Just add more tooling"** makes Cost Driver 3 worse, not better — it's the thing causing the fragmentation tax in the first place.

---

## Consequence of Inaction

Every quarter a regulated organization runs frontend observability through fragmented, third-party-hosted tools, it is carrying three compounding exposures simultaneously: the highest breach-cost profile in the economy (Driver 1), an active and currently-winning class-action theory that applies regardless of breach status (Driver 2), and a quantifiable, recurring operational tax on every incident it investigates (Driver 3). None of these are solved by adding another point solution — they are solved by changing where the data lives and how many tools an investigation has to cross.

This is why Imora is scoped as an alternative to the tools already in use — Sentry, Datadog RUM, LogRocket, FullStory, and their self-hosted counterparts — rather than a net-new category, per [vision.md](vision.md)'s Positioning section: the fastest way to zero out these three drivers is a credible swap-in replacement that also closes the gaps none of those alternatives close today.
