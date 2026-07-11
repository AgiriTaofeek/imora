# Vision

## Project Name

**Imora**

A self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory — built for regulated industries, with compliance capabilities none of them, or their self-hosted counterparts, ship.

---

# Positioning

Imora is not a new category. It is an alternative to the frontend observability tools engineering teams already use and already trust the workflow of — error tracking, session replay, performance monitoring — running entirely inside the organization's own infrastructure instead of a vendor's cloud.

Being "self-hosted" alone isn't the pitch — OpenReplay, PostHog, SigNoz, and GlitchTip already offer that, and are covered in detail in [competitive-analysis.md](competitive-analysis.md). The pitch is: **parity with the tools regulated teams already know, plus the specific compliance capabilities that neither the SaaS incumbents nor the existing self-hosted alternatives have** — an access-audit-trail over who viewed a customer's session, retention mapped to actual regulatory clocks, and evidence export built for an auditor rather than an engineer.

Every claim in this document is checked against that bar: is this parity (necessary to be a credible alternative at all) or wedge (the specific reason to choose Imora over every other alternative on the list)? Both matter. A product that's only compliant and mediocre at debugging won't get adopted daily; a product that's a great debugger with no compliance story doesn't solve the problem regulated buyers actually have.

---

# Vision Statement

To be the frontend observability platform regulated organizations reach for instead of Sentry, Datadog RUM, LogRocket, or FullStory — not because it does something entirely different, but because it does the same job without requiring customer data to leave the organization's infrastructure, and because it answers the questions a compliance team asks that none of those tools are built to answer.

We believe organizations should not have to choose between deep visibility into their frontend applications and complete ownership of their customer data.

Our mission is to give engineering, security, and compliance teams a single system of record for frontend incidents — one that keeps telemetry inside their infrastructure and produces evidence they can hand to an auditor, not just a dashboard they can screenshot.

---

# Why We Exist

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

# The Gap in Today's Landscape

This is not an unserved market — it is a market where every existing option makes a different compromise. Naming that gap precisely is what makes Imora's parity-plus-wedge positioning credible rather than a marketing claim, and it's covered in full detail, with sources, in [competitive-analysis.md](competitive-analysis.md).

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

# Our Belief

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

# What We Are Building

## Parity — what Imora has to match to be a credible alternative at all

The baseline every SaaS incumbent and self-hosted alternative in [competitive-analysis.md](competitive-analysis.md) already offers, and what a regulated org's engineers expect on day one:

- Error tracking with grouping/deduplication, not one alert per occurrence.
- Session replay with production-grade fidelity and default-safe PII masking.
- Performance monitoring against Core Web Vitals, with release-based regression detection.
- Framework-agnostic instrumentation across any modern frontend stack.

## Wedge — what none of the alternatives, self-hosted or SaaS, currently do

The specific capabilities identified as absent across every product reviewed in [competitive-analysis.md](competitive-analysis.md):

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

# Who We Serve

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

# Guiding Principles

## Data Ownership First — Parity

Organizations should fully own their telemetry, including the control plane — not just a dedicated instance running inside a vendor's cloud.

Customer data should remain within the organization's infrastructure unless explicitly configured otherwise, and it should be exportable in an open, documented schema so self-hosting does not become a new form of lock-in.

---

## Security by Default, Not Opt-In Redaction — Parity, with a Wedge Edge

Security must not be an enterprise add-on, and PII protection must not depend on a team remembering to configure a selector. The best tools in this space (Sentry, FullStory) already mask aggressively by default — that is the baseline we should meet, not a differentiator on its own.

Every deployment should default to:

- Deny-by-default session capture — rendering only an explicit allow-list of safe fields, so an unredacted new field fails closed, not open.
- Encryption at rest and in transit.
- Fine-grained, role-based access control down to the field level.
- Audit logging of access to sensitive records, not just system events — who viewed this customer's data, and when. This is the part no frontend observability tool ships today, which makes it Wedge rather than Parity.

---

## Compliance Is a Workflow, Not a Checkbox — Wedge

Most platforms treat compliance as a marketing bullet. We treat it as an operational feature with real mechanics:

- Retention policies mapped to regulatory clocks — GDPR's purpose-bound storage limitation, HIPAA's 6-year documentation floor, PCI-DSS's 12-month minimum audit trail retention, SOX's 7-year requirement — rather than one global TTL.
- Legal hold support that overrides normal retention when an investigation requires it.
- One-click evidence export — a defensible, timestamped incident package combining replay, errors, security signal, and access logs — built for handing to an auditor or regulator, not just an engineer.

---

## Investigation Over Metrics — Parity, with a Wedge Edge

Metrics are useful. Answers are better.

The platform should correlate session replay, errors, performance, and security signals into a single incident timeline, rather than requiring engineers to manually cross-reference separate dashboards to reconstruct what happened. Correlating replay with errors and performance is table stakes among the better tools already; correlating security signal into that same timeline is not, and is where this principle earns its Wedge status.

---

## Open Standards — Parity

Whenever possible, the platform should embrace open standards and interoperability.

Organizations should not become dependent on proprietary protocols, closed ecosystems, or an undocumented storage format that makes migrating away from Imora itself a compliance risk.

---

## Operational Simplicity — Parity

Deployment and operation should be straightforward.

Small teams should be able to deploy the platform on a single machine. Large enterprises should be able to scale it across clusters and regions, including fully air-gapped environments with no outbound dependency for core function. The operational model should remain predictable at every stage.

---

## Framework Agnostic — Parity

The platform should support any modern frontend technology stack.

Observability should not depend on whether a team uses React, Vue, Angular, Svelte, Solid, or plain JavaScript.

---

# Long-Term Vision

We envision a future where organizations can fully understand the health, behavior, reliability, and security of their frontend systems without compromising privacy, compliance, or control.

Our goal is not to invent a new category. It is to be the alternative a regulated organization's CISO, DPO, and engineering lead can all agree on — because it's a credible replacement for the observability tool their engineers already know how to use, and because it's the only alternative on the list that also answers the questions their compliance team is legally required to ask.

Being "self-hosted Sentry" or "self-hosted LogRocket" is necessary but not sufficient — that only matches the parity bar every other self-hosted alternative already clears. The reason to choose Imora specifically is the wedge: visibility, ownership, and provable compliance in one product, which today requires stitching together an observability tool and a set of compliance processes that don't talk to each other.
