# Architecture Decision Records

> Add one file per decision, numbered sequentially, following the Context / Decision / Alternatives Considered / Consequences shape below. Only write one for a decision that was genuinely contested (a real alternative existed and was rejected for a stated reason) and would be costly to reverse or easy to accidentally undo without a record of why — not every choice made along the way.

| # | Decision |
|---|---|
| [0001](0001-agplv3-licensing.md) | License the core product, wedge included, under AGPLv3 |
| [0002](0002-clickhouse-postgres-dual-store.md) | Split storage across ClickHouse (append-only) and PostgreSQL (relational) |
| [0003](0003-monorepo-structure.md) | Monorepo, with no directory that could become a license-gated carve-out |
| [0004](0004-two-profile-deployment.md) | Two deployment profiles (single-machine, cluster), air-gapped as an orthogonal setting |
| [0005](0005-unified-access-audit-event.md) | One AccessAuditEvent entity with an action enum, not separate logs per action type |
| [0006](0006-two-tier-pii-masking.md) | Two-tier masking: hard redaction vs. vault-backed soft masking |
| [0007](0007-keypair-signing-over-keyless.md) | Container images signed with a Cosign key pair, not Sigstore keyless signing |
| [0008](0008-go-backend-typescript-client.md) | Go for backend services, TypeScript for client-facing code |
| [0009](0009-structural-audit-log-enforcement.md) | The audit trail is enforced structurally, not by convention |

If a future change contradicts one of these, that's a signal to revisit the ADR explicitly, not silently drift past it.
