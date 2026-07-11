# Coding Standards

> Status: Resolves a decision left implicit everywhere else in this doc set — what language backend services are actually written in — and states the reasoning, since leaving it unspecified indefinitely would eventually force an arbitrary choice under time pressure instead of a deliberate one.

---

## Language: Go for Backend Services, TypeScript for Client-Facing Code

**Backend services** (`gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`): Go. Not a neutral pick — it directly serves three constraints already established elsewhere in this doc set:

- **[docker.md](../09-infrastructure/docker.md)'s minimal-image goal** — a Go service compiles to a single static binary, which fits a distroless/scratch base image about as tightly as multi-stage Docker builds get, with no runtime/interpreter to include.
- **[deployment-model.md](../04-architecture/deployment-model.md)'s single-machine resource budget** — ClickHouse already claims the 4-core/16GB floor; a lower-memory-footprint runtime for the six backend services leaves more of that budget for the part that actually needs it.
- **Ecosystem alignment** — the closest architectural comparator (Uptrace) is written in Go, ClickHouse's own official client library ecosystem is Go-first, and most Kubernetes-native tooling (client libraries, operators) is Go, which matters directly for [kubernetes.md](../09-infrastructure/kubernetes.md)'s cluster profile.

**`browser-sdk` and `dashboard`:** TypeScript, for reasons that aren't really a choice — `browser-sdk` is a browser library by definition, and `dashboard` is a web frontend consuming [rest-api.md](../07-api/rest-api.md).

This does mean two languages across the codebase, not one — accepted deliberately rather than defaulting to a single-language monorepo for its own sake, since Node's resource footprint would work against the single-machine budget for no real benefit on the backend side.

## Formatting and Linting

- Go: `gofmt`, non-negotiable — it's not a style preference, it's what the Go toolchain already enforces by convention.
- TypeScript: ESLint + Prettier, configured once in `packages/` and inherited by `browser-sdk` and `dashboard` rather than duplicated per package.

## What's Deliberately Not Modeled Here

- Detailed style-guide specifics (naming conventions, file organization within a service) — a team-calibrated document that should evolve, not a one-time architectural decision like the language choice above.
- Comment/documentation-string conventions — downstream of whatever the team finds actually gets maintained versus goes stale.
