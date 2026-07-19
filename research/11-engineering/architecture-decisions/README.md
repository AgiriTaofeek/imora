# Architecture Decision Records

> Status: Living index — add one file per decision, numbered sequentially, following the Context / Decision / Alternatives Considered / Consequences shape used below. Each ADR distills reasoning that lives in full elsewhere in `research/`; the ADR is a searchable decision log, not a duplicate of that reasoning.

| # | Decision | Status |
|---|---|---|
| [0001](0001-agplv3-licensing.md) | License the core product, wedge included, under AGPLv3 | Accepted |
| [0002](0002-clickhouse-postgres-dual-store.md) | Split storage across ClickHouse (append-only) and PostgreSQL (relational) | Accepted |
| [0003](0003-monorepo-structure.md) | Monorepo, with no directory that could become a license-gated carve-out | Accepted |
| [0004](0004-two-profile-deployment.md) | Two deployment profiles (single-machine, cluster), air-gapped as an orthogonal setting | Accepted |
| [0005](0005-unified-access-audit-event.md) | One AccessAuditEvent entity with an action enum, not separate logs per action type | Accepted |
| [0006](0006-two-tier-pii-masking.md) | Two-tier masking: hard redaction vs. vault-backed soft masking | Accepted |
| [0007](0007-keypair-signing-over-keyless.md) | Container images signed with a Cosign key pair, not Sigstore keyless signing | Accepted |
| [0008](0008-gateway-session-model.md) | Session model for human/dashboard authentication in `gateway` | Accepted |
| [0009](0009-project-key-format.md) | Project Key format for SDK/ingest authentication | Accepted |
| [0010](0010-gateway-cache-failure-mode.md) | Bounded grace-period local cache for Project Key validation during Redis outages | Accepted |

## When to Write a New ADR

Not every decision — only ones that were genuinely contested (a real alternative existed and was rejected for a stated reason) and would be costly to reverse or easy to accidentally undo without a record of why they were made this way. If a future change contradicts one of these, that's a signal to revisit the ADR explicitly, not silently drift past it.
