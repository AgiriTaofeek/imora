# User Personas

> Status: Research-based, current as of July 2026. Builds directly on [target-users.md](../00-overview/target-users.md) — that document established *which roles* carry Imora's three cost drivers; this one grounds each role in a concrete scenario, real deadlines, and real litigation outcomes, so the personas describe lived pressure rather than an abstract title.

Per [vision.md](../00-overview/vision.md)'s Positioning section, Imora is an alternative to tools these personas already have opinions about — Sentry, Datadog, LogRocket, FullStory — not a new category. Personas 1–5 (below) are why they'd switch; Persona 6 is why the switch sticks.

---

## Persona 1 — Dara, CISO at a Mid-Size Regional Bank

**Snapshot:** ~250-employee bank, ~40-person engineering org. Dara personally owns incident-response command per the org's size band, per [target-users.md](../00-overview/target-users.md)'s org-size variant findings.

**The scenario she's afraid of, because it already happened to peers in her sector:** in April 2026, **Sutter Health** — a healthcare organization, not a retailer — agreed to a **$21.5 million** class-action settlement over claims it ran third-party tracking tools on its website that captured California visitors' data without consent. Separately, Bloomingdale's was sued (Mikulsky v. Bloomingdale's) specifically because its session-replay vendor "recorded and transmitted [a visitor's] interactions with the website — including mouse movements, keystrokes and page views — to a third-party vendor without her consent." That is the exact architecture of every SaaS session-replay product in [competitive-analysis.md](../00-overview/competitive-analysis.md) Category 1.

**Why she can't just tell legal "we're fine":** the legal theory is genuinely unsettled, not a slam dunk against her, which makes it harder to dismiss, not easier. In Torres v. Prudential Financial — a financial-services case, her own sector — a court found session-replay vendors don't "read" data "in transit" and dismissed a CIPA §631 claim on that theory. But plaintiffs' firms responded by stacking a second claim under §638.51 (pen registers) specifically to survive that defense, and a Ninth Circuit reversal in a related case expanded potential liability rather than narrowing it. Dara's outside counsel cannot promise her this risk is closed — only that it is actively moving.

**What she needs from Imora:** the ability to state factually, in a board deck or a regulator inquiry, that session data was never transmitted to a third party — removing the fact pattern being litigated, rather than betting on which side of an unsettled circuit split her company lands on.

Sources: [Consumer Privacy Lawsuit Roundup 2026](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa), [Court Grants Summary Judgment: Session Replay Data "In Transit" — Inside Privacy](https://www.insideprivacy.com/data-privacy/court-grants-summary-judgment-website-vendor-cannot-read-session-replay-data-in-transit-under-cipa/), [Ninth Circuit Revives Session Replay Tracking Suit — Reed Smith](https://www.reedsmith.com/our-insights/blogs/viewpoints/102ksuo/ninth-circuit-revives-session-replay-tracking-suit/), [Website Wiretapping Roundup: 2025 Decisions](https://www.insideclassactions.com/2026/01/27/2025-website-wiretapping-roundup/).

---

## Persona 2 — Adaeze, Data Protection Officer at a National Insurer

**Snapshot:** appointed under GDPR Article 37 because the insurer conducts large-scale systematic monitoring of policyholders. Reports directly to the board, cannot be instructed by engineering or the CISO on how to do her job — a legal independence most of the other personas here don't have.

**Her clock, literally:** if a breach touching EU personal data occurs, she has **72 hours** from the moment the organization becomes aware of it to notify the supervisory authority — with partial, phased disclosure allowed only if the initial notification isn't delayed. Separately, any policyholder can file a Data Subject Access Request, and she has **one month** to produce a complete answer to "what data do you have on me, and who has looked at it."

**Why today's tools don't help her hit either deadline:** none of the products reviewed in [competitive-analysis.md](../00-overview/competitive-analysis.md) can answer "which employees viewed this specific person's session recording" — that data either doesn't exist, or exists only as a generic "user logged in" system event that doesn't name the record they viewed. She currently answers DSARs about session data by asking engineering to manually search logs, which routinely eats a meaningful fraction of her one-month window before she's even confirmed what exists.

**What she needs from Imora:** an access-audit-trail she can query herself, without depending on engineering's cooperation under time pressure, plus retention that expires data on GDPR's storage-limitation clock instead of a platform-wide TTL she has to argue engineering into changing.

Sources: [GDPR Data Breach Notification: 72-Hour Rule — Recording Law](https://www.recordinglaw.com/world-laws/world-data-privacy-laws/eu-data-privacy-laws/gdpr-breach-notification-72-hour-rule/), [Key GDPR Breach Notification Requirements](https://www.reform.app/blog/gdpr-breach-notification-requirements), [What are the responsibilities of a DPO? — European Commission](https://commission.europa.eu/law/law-topic/data-protection/rules-business-and-organisations/obligations/data-protection-officers/what-are-responsibilities-data-protection-officer-dpo_en).

---

## Persona 3 — Marcus, HIPAA Security Officer at a Regional Hospital Network

**Snapshot:** owns the annual documented risk assessment covering ePHI systems and third parties/business associates, mandated by the HIPAA Security Rule — distinct from the hospital's Privacy Officer, who owns policy and patient-facing disclosure.

**His clock:** if a breach affects 500 or more patients, he has **60 days** from discovery to notify both HHS and every affected individual. Sutter Health's $21.5M settlement — a healthcare peer, sued over the exact third-party-tracking pattern his own patient portal uses — is the scenario his board now asks him about directly.

**Why the annual risk assessment keeps flagging the same gap:** the Security Rule requires him to document administrative, physical, and technical safeguards — access controls and audit logging chief among them — for anything touching ePHI. His current session-replay vendor (Category 1 or 2 in [competitive-analysis.md](../00-overview/competitive-analysis.md)) can tell him *that* a support engineer viewed a patient's portal session, if he's lucky, but not produce that as a standing, queryable audit log he can hand an assessor without a custom export job.

**What he needs from Imora:** field-level access control and an audit log over *who viewed which patient's session*, structured so it satisfies the annual risk assessment as a built-in report, not a one-off favor from engineering before the audit deadline.

Sources: [HIPAA Security Officer — 2026 Update](https://www.hipaajournal.com/hipaa-security-officer/), [What are the notification requirements after a breach? — HIPAA Times](https://hipaatimes.com/what-are-the-notification-requirements-after-a-breach), [Consumer Privacy Lawsuit Roundup 2026](https://cookie-script.com/news/consumer-privacy-lawsuit-roundup-2026-from-cipa-to-coppa).

---

## Persona 4 — Priya, Head of Platform Engineering at a Mid-Stage Fintech

**Snapshot:** ~300 product engineers, which puts her platform group at the upper end of the typical **10–30 person** band mid-stage US fintechs run for that size of engineering org — roughly the standard **1:8–12** platform-to-product-engineer ratio.

**What's actually on her desk:** her on-call rotation needs a minimum of **8–10 engineers** just to cover the baseline security and reliability load for a regulated fintech — and every one of those engineers currently has to know four or five different tools (error tracker, session replay, APM, security/WAF dashboard) to investigate a single incident, none of which share an access model or a retention policy.

**Why she's the hardest sell, not the easiest:** she is the one directly evaluating whether her team can *operate* Imora long-term, not just whether it's compliant. She has already seen self-hosted tools become a second full-time job for someone on her team. She will reject anything that solves the CISO's and DPO's problems by making her team's operational burden worse.

**What she needs from Imora:** the "Operational Simplicity" commitment in [vision.md](../00-overview/vision.md) to be real in practice — deployable by 2–3 people at minimum viable scale, without requiring her to grow platform headcount just to run the observability stack.

Sources: [Platform Engineering Team Size 2026](https://platformengineeringcost.com/team-structure), [At what company size should you adopt Platform Engineering? — SRE School](https://sreschool.com/forum/d/300-at-what-company-size-should-you-adopt-platform-engineering).

---

## Persona 5 — Jon, Incident Commander / Senior SRE on Priya's Team

**Snapshot:** rotates as Incident Commander on the 8–10-person on-call roster Priya staffs. Owns the reliability contract day to day — SLOs, capacity, and the observability tooling itself.

**A Tuesday, 2 a.m.:** a release regresses checkout for a subset of users. Jon's first twenty minutes aren't spent diagnosing — they're spent opening four separate tools, none of which share a timeline, to work out whether this is a frontend bug, a backend regression, or someone actively probing the checkout flow for fraud. If the incident turns out to involve exposed customer data, he now also owns **chain-of-custody documentation**, because a production incident touching customer data can become litigation, and the evidence has to hold up months later — not be reconstructed from Slack threads after the fact.

**What he needs from Imora:** replay, errors, performance, and security signal correlated into one timeline from the start, so the first twenty minutes go to diagnosis instead of reconstruction — and an evidence export that's already defensible, instead of a task he has to remember to do carefully at 2 a.m.

Source: [Overview of Incident Lifecycle in SRE — Squadcast](https://www.squadcast.com/blog/overview-of-incident-lifecycle-in-sre), platform/SRE role research cited under Persona 4.

---

## Persona 6 — Chidi, Senior Frontend Engineer on Priya's Team

**Snapshot:** no compliance mandate, no incident, no auditor in the room. Chidi opens Imora on an ordinary Wednesday because a deploy went out last night and something feels slower, or a support ticket says "checkout is broken" with no other detail.

**A normal Wednesday:** last night's release shipped. This morning, Chidi wants to know whether LCP or INP moved for the checkout flow — against the thresholds the entire industry already treats as the bar (LCP under 2.5s, INP under 200ms, CLS under 0.1, evaluated at the 75th percentile of real users, per Google's own Core Web Vitals methodology). If it did regress, Chidi expects the tool to already know which release caused it — the way Sentry's regression detection ties a statistically significant change in endpoint or function duration back to a specific deploy — rather than making Chidi bisect it by hand.

**Why noise is the thing that actually drives Chidi away:** a 2025 Catchpoint study found **62% of on-call engineers have ignored a critical alert because it was buried in noise.** If Imora pages Chidi once per affected user instead of once per root cause, Chidi tunes it out within a month, and every other capability in this document — audit trails, evidence export, retention policy — becomes irrelevant, because the person who was supposed to use the tool daily has stopped opening it.

**Why Chidi is in this document at all:** Personas 1–5 explain why Dara, Adaeze, Marcus, and Priya sign the contract. None of them explain why anyone *opens the product on a day nothing is wrong*. That's Chidi's job, and it's the actual adoption and retention driver — a tool CISOs mandate but engineers route around is a well-known failure mode, not a hypothetical one.

**What Chidi needs from Imora:** error grouping and performance regression detection that hold up against the category leaders on their own terms, with no compliance framing attached — because Chidi will never open a "compliance" tab, only a "why is this broken" one.

Sources: [How the Core Web Vitals metrics thresholds were defined — web.dev](https://web.dev/articles/defining-core-web-vitals-thresholds), [Alert Fatigue in SRE and DevOps — Sensu](https://sensu.io/blog/alert-fatigue-in-sre-and-devops), [Sentry Endpoint Regression docs](https://docs.sentry.io/product/issues/issue-details/performance-issues/endpoint-regressions/).

---

## What Changed From target-users.md to Here

target-users.md established the *role* and its regulatory obligations. This document adds the parts that make each persona feel like a specific person under specific pressure rather than a job description:

- **Real dollar figures from real settlements** (Sutter Health $21.5M, Bloomingdale's, LA Times $3.85M, Fandom/GameSpot $1.2M) — not hypothetical breach-cost averages.
- **Real deadlines** (GDPR 72-hour breach notification, GDPR 1-month DSAR, HIPAA 60-day breach notification) that turn "we should build an audit trail" into "she has one month and no way to answer the question today."
- **Real team-size constraints** (8–10 person on-call minimum, 1:8–12 platform ratio) that make Persona 4's operational-simplicity requirement concrete instead of a vague preference.
- **Persona 6 (Chidi)**, added deliberately after a review of the first five: Personas 1–5 explain why a regulated org buys Imora, but they're all compliance- or incident-driven, and none of them explain daily use. Without a persona whose job has nothing to do with compliance, the roadmap risks optimizing entirely for the buyer and neglecting the person who determines whether the product gets used at all.

This should feed directly into `docs/01-product/user-stories.md` as JTBD statements per persona, and into `docs/01-product/prd.md` once user stories exist to prioritize against.
