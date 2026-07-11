# Testing

> Status: Specifies the compliance-guarantee test suite [ci-cd.md](../09-infrastructure/ci-cd.md) and [release-process.md](release-process.md) both referenced without defining, and turns [event-schema.md](../06-data/event-schema.md)'s "additive-only, forever" rule into an actual CI gate rather than a convention.

---

## The Finding: Schema Compatibility Can Be a CI Gate, Not Just a Rule Someone Has to Remember

[event-schema.md](../06-data/event-schema.md) states that schema evolution must never rename, retype, or remove an existing field — stated as policy, twice now, but nothing yet checks it automatically. Real, existing tooling closes this gap directly: a JSON-schema diff check (`json-schema-diff-validator` or equivalent) runs in CI on every change to `event-schema.md`'s field definitions, and **the build fails on any breaking change**, not just a naming-convention violation. [rest-api.md](../07-api/rest-api.md)'s OpenAPI spec gets the equivalent check via `oasdiff` — but with deliberately different strictness, matching that document's own looser rule: breaking changes are *allowed* there, gated on a major version bump and deprecation notice, whereas the event schema check allows no breaking changes ever, full stop. Same tooling category, two different policies, because the two documents specify two different compatibility guarantees for two different reasons (7-year-old stored records vs. a request/response contract with real clients who can migrate).

---

## Test Pyramid

- **Unit** — business-rule logic in isolation: BR-1's longest-clock resolution, BR-3's selective-purge field-level decision, [pii-redaction.md](../08-security/pii-redaction.md)'s three-input classification logic. Fast, no real infrastructure.
- **Integration** — against real infrastructure, deliberately not mocked, because mocking would defeat the point of verifying actual enforcement:
  - The DB-level GRANT restrictions from [threat-model.md](../08-security/threat-model.md) — attempt `DELETE`/`ALTER` against `access_audit_events` using the `ingestion`/`query-api` service account and assert it fails.
  - MinIO Object Lock in Compliance mode — attempt to delete a locked EvidenceExport blob, including with elevated credentials, and assert it fails, per [storage.md](../06-data/storage.md).
  - The legal-hold check-before-destroy ordering (BR-2) — apply a hold mid-sweep and assert records not yet processed are protected while already-deleted records aren't retroactively flagged as a bug, per [sequence-diagrams.md](../04-architecture/sequence-diagrams.md) Flow C.
  - `CONFIG_CHANGED` firing correctly on RetentionPolicy/role/field-classification changes, with `oldValue`/`newValue` populated, per [audit-logging.md](../08-security/audit-logging.md).
  - Container conventions from [docker.md](../09-infrastructure/docker.md) — non-root user, read-only root filesystem, verified against the actual built image, not asserted in a Dockerfile comment.
- **End-to-end** — the four flows in [sequence-diagrams.md](../04-architecture/sequence-diagrams.md) (Session Capture, DSAR Query, Retention Sweep Hitting a Legal Hold, Evidence Export Generation) **are the e2e test spec already**, not a separate scenario set to invent — that document was written at exactly the right altitude to double as test-case documentation.

---

## What's Deliberately Not Modeled Here

- Specific test framework/runner choice — a tooling decision, not a strategy one.
- Coverage percentage targets — a team-calibrated number, not an architectural constraint.
- Load/performance testing methodology — downstream of [scaling.md](../04-architecture/scaling.md)'s thresholds once real traffic data exists to test against.

## What This Feeds Next

`docs/10-engineering/branching.md` and `coding-standards.md` round out `10-engineering/`.
