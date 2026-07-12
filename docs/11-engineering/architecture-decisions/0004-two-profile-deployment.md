# 0004. Two deployment profiles (single-machine, cluster), air-gapped as an orthogonal setting on either

> Status: Accepted. Full reasoning in [Deployment Model](../../03-architecture/README.md#deployment-model) and [Scaling](../../03-architecture/README.md#scaling).

## Context

Story P1 requires a 2–3 person team to deploy a working instance; large enterprises (per [Target Users](../../00-overview/README.md#target-users)'s 300+-employee band) need to scale across clusters and regions. A full distributed topology (independently-scaled services, Kafka, clustered ClickHouse/Postgres) by default would fail P1 outright for small teams. Separately, [System Context](../../03-architecture/README.md#system-context) requires every capability to work with zero external systems present for air-gapped customers — a requirement orthogonal to deployment scale, not a third scale tier.

## Decision

Two topology profiles: single-machine (Docker Compose, no message queue, ClickHouse/Postgres/Redis/MinIO as sibling containers, sized to 4-core/16GB as the floor) and cluster (Kubernetes, independently-scaled services, a message queue introduced between `ingestion` and its consumers, multi-node stores). Air-gapping applies to either profile as a setting — the set of optional external systems present (SSO, notifications, backend trace correlation), not a separate deployment mode.

## Alternatives Considered

- **One-size-fits-all distributed topology:** rejected — directly fails the P1 story for small teams.
- **Air-gapped as a third profile:** rejected — would require maintaining parallel single-machine-air-gapped and cluster-air-gapped variants instead of one orthogonal setting on two profiles.

## Consequences

- Migration between profiles changes only physical deployment shape — domain model, business rules, and event catalog stay identical (see [Scaling](../../03-architecture/README.md#scaling)'s explicit statement of this as a design constraint, not just an outcome).
- The scaling trigger between profiles turned out to be retention-driven accumulated storage, not ingestion throughput — a genuinely counter-intuitive finding specific to Imora's multi-year regulatory retention obligations (ADR-worthy in its own right; see [Scaling](../../03-architecture/README.md#scaling)).
- `workers` had to split into a `CronJob` (Forbid concurrency) and a scalable `Deployment` in the cluster profile specifically because naive horizontal scaling of the retention-sweep logic would race against itself — a direct consequence of this ADR, detailed in [Kubernetes](../../12-infrastructure/README.md#kubernetes).
