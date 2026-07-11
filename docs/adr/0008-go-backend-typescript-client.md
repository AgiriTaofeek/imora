# 0008. Go for backend services, TypeScript for client-facing code

> Status: Accepted. Condensed context in [../design-doc.md](../design-doc.md); full original reasoning is preserved in git history from before the doc-set consolidation.

## Context

Backend service language was left unstated across earlier design work — a decision that would eventually get made arbitrarily under time pressure if not decided deliberately. Two constraints matter most: the single-machine MVP deployment needs a small resource footprint (no separate infra budget to spare), and container images should be minimal for security reasons.

## Decision

Backend services (gateway, ingestion, query-api, alert-engine): Go. Client-facing code (browser-sdk, dashboard): TypeScript, which isn't really a choice — browser-sdk is a browser library by definition, dashboard is a web frontend.

## Alternatives Considered

- **Node/TypeScript for the whole stack (one language everywhere):** rejected — Node's memory footprint works against the single-machine resource budget for no real benefit on the backend side, where a single static binary (Go's output) fits a minimal container image more tightly than a runtime-dependent language.

## Consequences

- Two languages across the codebase, accepted deliberately rather than optimizing for a single-language monorepo.
- Go's static-binary output is what makes a minimal, distroless container image straightforward — directly serves the "small resource footprint" requirement from the MVP deployment target.
- Closest architectural comparators in this space (Uptrace) are also Go-based backends — not the deciding factor, but a signal the choice is well-trodden for this kind of system.
