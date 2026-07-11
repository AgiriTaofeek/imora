# 0002. Split storage across ClickHouse (append-only, high-volume) and PostgreSQL (relational, transactional)

> Status: Accepted. Condensed context in ../design-doc.md; full original reasoning is preserved in git history from before the doc-set consolidation.

## Context

Imora's data falls into two genuinely different shapes: high-volume, append-only, time-series data (SessionEvent, ErrorEvent, PerformanceMetric, SecurityEvent, AccessAuditEvent) and small-cardinality, transactionally-updated relational data (Session summaries, Release, ErrorGroup, RetentionPolicy, LegalHold, EvidenceExport metadata). A single store optimized for one shape is a poor fit for the other — this is also the pattern the closest architectural comparator (Uptrace) uses.

## Decision

ClickHouse holds the five high-volume append-only entities. PostgreSQL holds the six relational entities. No foreign-key constraint exists across the store boundary — `ingestion` and `alert-engine` are responsible for writing both sides consistently, since ClickHouse can't constrain against another engine.

## Alternatives Considered

- **Single relational store for everything:** rejected — PostgreSQL doesn't scale to the write throughput or query shape of session-replay-volume event data.
- **Single column store for everything, including config-like data:** rejected — LegalHold, RetentionPolicy, and role grants need ACID guarantees and referential integrity ClickHouse doesn't provide, and are read/written far less frequently than they'd need to justify giving up transactional correctness.

## Consequences

- Cross-store consistency is an application-level responsibility, explicitly documented rather than assumed — see postgres-schema.md's "Common Design Decisions" section.
- Retention/legal-hold enforcement (BR-1/BR-2) has to bridge the two stores: LegalHold scope lives in Postgres as a re-evaluated predicate, refreshed into a ClickHouse Dictionary that TTL expressions reference — a direct consequence of this split, not an independent design choice.
- The single-machine deployment profile (deployment-model.md) has to run both stores on one host, which is the dominant resource-sizing driver for that profile (ClickHouse at 4-core/16GB is the floor).
