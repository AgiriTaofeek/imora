# Engineering

## Branching Strategy

> Status: Trunk-based development — not a default choice, a specific fit for the monorepo decision in ADR [0003](architecture-decisions/0003-monorepo-structure.md).

---

### Trunk-Based Development

DORA's research on this is specific, not a general preference: elite performers who meet their reliability targets are 2.3x more likely to use trunk-based development, correlated specifically with three-or-fewer active branches, merging to trunk at least daily, and no code-freeze/integration phases.

**Why this matters more than usual for this repository specifically:** `packages/domain-types` is the literal Shared Kernel that `browser-sdk`, `ingestion`, and `query-api` all depend on directly, per [Bounded Contexts](../02-domain/README.md#bounded-contexts) and ADR [0003](architecture-decisions/0003-monorepo-structure.md). A long-lived feature branch touching that package accumulates drift risk across three consuming services simultaneously, not just within itself — trunk-based development's short-lived-branch discipline is what keeps that drift window small.

### Practice

- Feature branches live hours to a couple of days, not weeks — work that can't fit in that window gets built behind a feature flag and merged incomplete-but-inert, rather than kept on a long-lived branch.
- No permanent `develop` or `release` branch. Releases are tags cut directly from trunk, per [Release Process](README.md#release-process).
- **Hotfixes:** for a critical patch against an already-shipped version (the scenario [Release Process](README.md#release-process)'s air-gapped notification gap describes), branch from the affected tag, cherry-pick the fix, tag a patch release, delete the branch. This is a short-lived, purpose-specific branch, not a standing release-maintenance branch.

### What's Deliberately Not Modeled Here

- PR review requirements/approval counts — a team-process decision, not an architectural one.
- Commit message conventions — downstream tooling choice (e.g., whether commit messages drive changelog automation from [Release Process](README.md#release-process)).

---

## Coding Standards

> Status: Resolves a decision left implicit everywhere else in this doc set — what language backend services are actually written in — and states the reasoning, since leaving it unspecified indefinitely would eventually force an arbitrary choice under time pressure instead of a deliberate one.

---

### Language: Go for Backend Services, TypeScript for Client-Facing Code

**Backend services** (`gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`): Go. Not a neutral pick — it directly serves three constraints already established elsewhere in this doc set:

- **[Docker](../12-infrastructure/README.md#docker)'s minimal-image goal** — a Go service compiles to a single static binary, which fits a distroless/scratch base image about as tightly as multi-stage Docker builds get, with no runtime/interpreter to include.
- **[Deployment Model](../03-architecture/README.md#deployment-model)'s single-machine resource budget** — ClickHouse already claims the 4-core/16GB floor; a lower-memory-footprint runtime for the six backend services leaves more of that budget for the part that actually needs it.
- **Ecosystem alignment** — the closest architectural comparator (Uptrace) is written in Go, ClickHouse's own official client library ecosystem is Go-first, and most Kubernetes-native tooling (client libraries, operators) is Go, which matters directly for [Kubernetes](../12-infrastructure/README.md#kubernetes)'s cluster profile.

**`browser-sdk` and `dashboard`:** TypeScript, for reasons that aren't really a choice — `browser-sdk` is a browser library by definition, and `dashboard` is a web frontend consuming [REST API](../06-api/README.md#rest-api).

This does mean two languages across the codebase, not one — accepted deliberately rather than defaulting to a single-language monorepo for its own sake, since Node's resource footprint would work against the single-machine budget for no real benefit on the backend side.

### Formatting and Linting

- Go: `gofmt`, non-negotiable — it's not a style preference, it's what the Go toolchain already enforces by convention.
- TypeScript: ESLint + Prettier, configured once in `packages/` and inherited by `browser-sdk` and `dashboard` rather than duplicated per package.

### What's Deliberately Not Modeled Here

- Detailed style-guide specifics (naming conventions, file organization within a service) — a team-calibrated document that should evolve, not a one-time architectural decision like the language choice above.
- Comment/documentation-string conventions — downstream of whatever the team finds actually gets maintained versus goes stale.

---

## Testing

> Status: Specifies the compliance-guarantee test suite [CI/CD](../12-infrastructure/README.md#cicd) and [Release Process](README.md#release-process) both referenced without defining, and turns [Event Schema](../05-data/README.md#event-schema)'s "additive-only, forever" rule into an actual CI gate rather than a convention.

---

### The Finding: Schema Compatibility Can Be a CI Gate, Not Just a Rule Someone Has to Remember

[Event Schema](../05-data/README.md#event-schema) states that schema evolution must never rename, retype, or remove an existing field — stated as policy, twice now, but nothing yet checks it automatically. Real, existing tooling closes this gap directly: a JSON-schema diff check (`json-schema-diff-validator` or equivalent) runs in CI on every change to `README.md#event-schema`'s field definitions, and **the build fails on any breaking change**, not just a naming-convention violation. [REST API](../06-api/README.md#rest-api)'s OpenAPI spec gets the equivalent check via `oasdiff` — but with deliberately different strictness, matching that document's own looser rule: breaking changes are *allowed* there, gated on a major version bump and deprecation notice, whereas the event schema check allows no breaking changes ever, full stop. Same tooling category, two different policies, because the two documents specify two different compatibility guarantees for two different reasons (7-year-old stored records vs. a request/response contract with real clients who can migrate).

---

### Test Pyramid

- **Unit** — business-rule logic in isolation: BR-1's longest-clock resolution, BR-3's selective-purge field-level decision, [PII Redaction](../07-security/README.md#pii-redaction)'s three-input classification logic. Fast, no real infrastructure.
- **Integration** — against real infrastructure, deliberately not mocked, because mocking would defeat the point of verifying actual enforcement:
  - The DB-level GRANT restrictions from [Threat Model](../07-security/README.md#threat-model) — attempt `DELETE`/`ALTER` against `access_audit_events` using the `ingestion`/`query-api` service account and assert it fails.
  - MinIO Object Lock in Compliance mode — attempt to delete a locked EvidenceExport blob, including with elevated credentials, and assert it fails, per [Storage](../05-data/README.md#storage).
  - The legal-hold check-before-destroy ordering (BR-2) — apply a hold mid-sweep and assert records not yet processed are protected while already-deleted records aren't retroactively flagged as a bug, per [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) Flow C.
  - `CONFIG_CHANGED` firing correctly on RetentionPolicy/role/field-classification changes, with `oldValue`/`newValue` populated, per [Audit Logging](../07-security/README.md#audit-logging).
  - Container conventions from [Docker](../12-infrastructure/README.md#docker) — non-root user, read-only root filesystem, verified against the actual built image, not asserted in a Dockerfile comment.
- **End-to-end** — the four flows in [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) (Session Capture, DSAR Query, Retention Sweep Hitting a Legal Hold, Evidence Export Generation) **are the e2e test spec already**, not a separate scenario set to invent — that document was written at exactly the right altitude to double as test-case documentation.

---

### What's Deliberately Not Modeled Here

- Specific test framework/runner choice — a tooling decision, not a strategy one.
- Coverage percentage targets — a team-calibrated number, not an architectural constraint.
- Load/performance testing methodology — downstream of [Scaling](../03-architecture/README.md#scaling)'s thresholds once real traffic data exists to test against.

### What This Feeds Next

`docs/11-engineering/README.md#branching-strategy` and `README.md#coding-standards` round out `11-engineering/`.

---

## Release Process

> Status: How a change becomes a version customers actually run, across both distribution paths from [CI/CD](../12-infrastructure/README.md#cicd).

---

### Versioning

Semantic versioning (MAJOR.MINOR.PATCH), independent of two other version-like concepts already defined elsewhere that are easy to conflate with it: [Event Schema](../05-data/README.md#event-schema)'s per-event `schemaVersion` (which only ever increments additively, per that document's governing rule) and [REST API](../06-api/README.md#rest-api)'s `/v1/` API path version (which can break at a major boundary with migration notice). A software MAJOR release does not imply an event-schema or API-version bump, and vice versa — three independent axes, not one number wearing three hats.

### Pipeline

1. Version bump, changelog entry appended to the root `CHANGELOG.md`.
2. Build, test (including the compliance-guarantee suite from [CI/CD](../12-infrastructure/README.md#cicd) — the GRANT-restriction and non-root-container assertions), SBOM generation, key-pair signing (per ADR [0007](architecture-decisions/0007-keypair-signing-over-keyless.md)).
3. **Dual publish**, from the identical signed artifact: registry push for connected deployments; signed bundle packaging for air-gapped transfer, per [Deployment Model](../03-architecture/README.md#deployment-model).
4. Git tag.

---

### The Gap: How Does an Air-Gapped Customer Even Learn a Release Exists?

Every prior document addressing air-gapped updates assumed the customer already knows they need one and initiates the transfer. Nothing yet addresses the step before that — **Imora has no path to notify an air-gapped customer that a release, especially a critical security patch, exists at all**, for exactly the same reason it can't push the update directly: no reachability into their network. This has to be an out-of-band channel, not a product feature: release notes and security advisories published somewhere a Platform Operator is expected to check periodically (a security mailing list, a published advisory page, an email to a registered contact) — never a webhook, never a phone-home version check, since both would violate the same air-gapped constraint this entire release mechanism was built around. Connected deployments can additionally get an in-product "update available" notice through the registry-pull path; air-gapped ones fundamentally cannot, and no future design should try to route around that by weakening the air-gap for the sake of update convenience.

---

### What's Deliberately Not Modeled Here

- Exact changelog format/automation tooling — implementation detail.
- Security advisory publication mechanics (CVE assignment, disclosure timeline) — a security-process concern downstream of this document, not part of the release pipeline itself.

### What This Feeds Next

`docs/11-engineering/README.md#testing` should specify the compliance-guarantee test suite referenced above in full. `docs/11-engineering/README.md#branching-strategy` and `README.md#coding-standards` round out this folder.

