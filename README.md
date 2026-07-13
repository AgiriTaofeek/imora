# Imora

> Status: Planning/documentation phase — no product code yet. Everything below describes what is designed and decided, not what is built.

Imora is a **self-hosted alternative to Sentry, Datadog RUM, LogRocket, and FullStory**, built for regulated industries (banks, insurers, hospitals, government) that can't send customer session data to a third party. It aims to match those tools on everyday debugging — error tracking, session replay, performance monitoring — and win on three capabilities none of them ship: an audit trail of who viewed a customer's session, retention mapped to actual regulatory clocks, and one-click evidence export for auditors.

## Where to Start

**[research/README.md](research/README.md)** — start with its "Start Here" section, which includes a 10-minute reading path. The full documentation set covers vision through infrastructure across 13 numbered folders, ordered to follow the actual planning flow (vision → problem → PRD → architecture & design → roadmap → feature specs → implementation → testing/rollout/observe), with load-bearing decisions recorded as ADRs in [research/11-engineering/architecture-decisions/](research/11-engineering/architecture-decisions/).

## Repository Layout

- [research/](research/) — the complete planning/design documentation set (also reusable as a documentation framework for other projects; see its README).
- [diagrams/](diagrams/) — architecture diagrams are maintained inline as Mermaid inside `research/03-architecture/`; this folder's README explains why.
- [ROADMAP.md](ROADMAP.md) → [research/08-roadmap/feature-roadmap.md](research/08-roadmap/feature-roadmap.md) — three milestones, no fabricated dates.
- [CHANGELOG.md](CHANGELOG.md) / [CONTRIBUTING.md](CONTRIBUTING.md) — placeholders until there are releases and code to contribute to.

## License

Planned: AGPLv3 for the entire product, including all compliance capabilities — no feature-gated enterprise edition. Rationale in the [Licensing section](research/01-product/README.md#licensing) of `research/01-product/README.md` and [ADR 0001](research/11-engineering/architecture-decisions/0001-agplv3-licensing.md).
