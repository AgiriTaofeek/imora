# Imora

> Status: Planning phase — no product code yet.

Imora is a **self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory**, built for regulated industries (banks, insurers, hospitals, government) that can't send customer session data to a third party. It matches those tools on everyday debugging — error tracking, session replay, performance monitoring — and wins on what none of them ship: an audit trail of who viewed a customer's session, and a consent record proving what basis existed when it was captured.

## Where to Start

- **[docs/pr-faq.md](docs/pr-faq.md)** — the press release and hard FAQ, written before any technical design, per Amazon's "Working Backwards" method: if this doesn't describe something meaningfully better than what exists, it isn't ready to build.
- **[docs/design-doc.md](docs/design-doc.md)** — the one design document: problem, positioning, MVP scope, architecture, key trade-offs. Read start to finish in about 20 minutes.
- **[docs/adr/](docs/adr/)** — architecture decision records, one per genuinely contested, hard-to-reverse decision. Added one at a time as real decisions come up, not written speculatively ahead of need.

This is deliberately lean. A large multi-file specification was drafted and discarded in favor of this — the tested pattern used at Google (design docs), Amazon (PR-FAQ), and Uber/Stripe (RFCs + ADRs), all of which plan with a handful of documents, not a full upfront specification. That earlier material still exists in git history if a specific fact or citation is needed later; it's no longer the primary artifact.

## Repository Layout

- [docs/](docs/) — pr-faq.md, design-doc.md, adr/.
- [ROADMAP.md](ROADMAP.md) → `docs/design-doc.md`'s MVP / Beyond MVP sections.
- [CHANGELOG.md](CHANGELOG.md) / [CONTRIBUTING.md](CONTRIBUTING.md) — placeholders until there are releases and code to contribute to.

## License

Planned: AGPLv3 for the entire product, including all compliance capabilities — no feature-gated enterprise edition. See [docs/adr/0001](docs/adr/0001-agplv3-licensing.md).
