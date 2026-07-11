# PR-FAQ: Imora

> Amazon's "Working Backwards" format — write the launch press release and the hard internal FAQ before any technical design begins. If the press release doesn't describe something meaningfully better than what exists today, the idea isn't ready to build. This document is the test Imora has to pass, not a marketing artifact.

---

## FOR IMMEDIATE RELEASE

### Imora Launches: The Frontend Observability Platform Banks, Hospitals, and Government Agencies Can Actually Use

*A self-hosted alternative to Sentry, Datadog, and LogRocket that never sends customer data outside your own servers — and proves it.*

**[City], [Date]** — Today marks the launch of Imora, a self-hosted platform that helps engineering teams find and fix bugs, track performance, and monitor security on their websites — built specifically for organizations that legally cannot send customer data to an outside company.

For years, banks, hospitals, insurers, and government agencies have faced an impossible choice: use a powerful tool like Sentry or Datadog and accept that customer data leaves their servers, or give up that visibility entirely to stay compliant. Making it worse, a wave of recent lawsuits — including a $21.5 million settlement paid by a major healthcare provider — has targeted companies specifically for recording website visitors' activity through third-party tools without clear proof of consent.

Imora runs entirely inside a customer's own infrastructure. Engineers get the same error tracking, session replay, and performance monitoring they already expect from tools like Sentry — and compliance teams get something none of those tools offer: a permanent record of exactly who on their own team viewed a customer's data, and when, plus a timestamped record of what consent basis existed when that data was captured in the first place. Data is kept only as long as the law actually requires, instead of one blanket setting.

"We spent years telling our board we couldn't get the same visibility into bugs and performance that every other tech company takes for granted, because we handle patient data," said a hospital IT director. "Imora is the first tool that didn't ask us to choose."

Getting started takes under an hour: install Imora on a single server, add one line of code to a website, and the first recorded session appears immediately — fully masked, with an audit trail already running, by default.

Imora is available now, self-hosted, under an open-source license (AGPLv3), with paid support and enterprise features available for larger deployments.

---

## FAQ

### External (customer-facing) questions

**Q: How is this different from just using Sentry or Datadog?**
Nothing changes about the day-to-day debugging experience — same error tracking, same session replay. The difference is where the data lives (your servers, always) and what happens when someone asks "who looked at this."

**Q: We don't have a big DevOps team. Can we actually run this?**
Yes — it's designed to be installed and running by two or three people on a single server within an hour, not a distributed system requiring a dedicated platform team.

**Q: Does this work if we have no internet access at all (air-gapped)?**
Yes, fully — nothing about the core product depends on reaching an outside server, ever.

**Q: What does the audit trail actually show us?**
For any customer session, exactly which employee viewed it, when, and why (if they had to unmask any private field).

**Q: Can we prove a specific user actually agreed to being recorded?**
Yes — every recorded session carries a record of what consent basis applied at the moment it started, so "did we have permission" has an actual answer, not just a policy that says it should.

**Q: What does it cost?**
The full product is free and open-source, including the compliance features. Paid tiers cover managed hosting, support, and enterprise login integration — never the core functionality.

### Internal (harder) questions

**Q: Why would a company switch from what they use today?**
They're not just switching tools — they're removing specific legal exposure (the lawsuits described above) that exists today regardless of how happy they are with their current tool.

**Q: Why hasn't an existing company already built this?**
The closest competitors (Sentry, PostHog, OpenReplay) are all either SaaS-first (data leaves the customer) or self-hosted without the compliance layer (PostHog gates similar features behind a paid tier). Nobody has combined true self-hosting with an audit trail and consent record built in from day one.

**Q: What's the biggest risk this doesn't work?**
That engineers find the day-to-day product worse than what they're replacing. The compliance story gets the contract signed; a mediocre debugging experience is what gets it quietly abandoned six months later. See `design-doc.md`'s MVP scope — parity has to be real, not an afterthought behind the compliance pitch.

**Q: What's the smallest version of this worth shipping first?**
See `design-doc.md`'s MVP section — that answer is detailed enough to actually build against, not restated here.
