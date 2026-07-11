# Release Process

> Status: How a change becomes a version customers actually run, across both distribution paths from [ci-cd.md](../09-infrastructure/ci-cd.md).

---

## Versioning

Semantic versioning (MAJOR.MINOR.PATCH), independent of two other version-like concepts already defined elsewhere that are easy to conflate with it: [event-schema.md](../06-data/event-schema.md)'s per-event `schemaVersion` (which only ever increments additively, per that document's governing rule) and [rest-api.md](../07-api/rest-api.md)'s `/v1/` API path version (which can break at a major boundary with migration notice). A software MAJOR release does not imply an event-schema or API-version bump, and vice versa — three independent axes, not one number wearing three hats.

## Pipeline

1. Version bump, changelog entry appended to the root `CHANGELOG.md`.
2. Build, test (including the compliance-guarantee suite from [ci-cd.md](../09-infrastructure/ci-cd.md) — the GRANT-restriction and non-root-container assertions), SBOM generation, key-pair signing (per ADR [0007](architecture-decisions/0007-keypair-signing-over-keyless.md)).
3. **Dual publish**, from the identical signed artifact: registry push for connected deployments; signed bundle packaging for air-gapped transfer, per [deployment-model.md](../04-architecture/deployment-model.md).
4. Git tag.

---

## The Gap: How Does an Air-Gapped Customer Even Learn a Release Exists?

Every prior document addressing air-gapped updates assumed the customer already knows they need one and initiates the transfer. Nothing yet addresses the step before that — **Imora has no path to notify an air-gapped customer that a release, especially a critical security patch, exists at all**, for exactly the same reason it can't push the update directly: no reachability into their network. This has to be an out-of-band channel, not a product feature: release notes and security advisories published somewhere a Platform Operator is expected to check periodically (a security mailing list, a published advisory page, an email to a registered contact) — never a webhook, never a phone-home version check, since both would violate the same air-gapped constraint this entire release mechanism was built around. Connected deployments can additionally get an in-product "update available" notice through the registry-pull path; air-gapped ones fundamentally cannot, and no future design should try to route around that by weakening the air-gap for the sake of update convenience.

---

## What's Deliberately Not Modeled Here

- Exact changelog format/automation tooling — implementation detail.
- Security advisory publication mechanics (CVE assignment, disclosure timeline) — a security-process concern downstream of this document, not part of the release pipeline itself.

## What This Feeds Next

`docs/10-engineering/testing.md` should specify the compliance-guarantee test suite referenced above in full. `docs/10-engineering/branching.md` and `coding-standards.md` round out this folder.
