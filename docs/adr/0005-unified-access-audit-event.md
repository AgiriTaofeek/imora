# 0005. One AccessAuditEvent entity with an action enum, not separate logs per action type

> Status: Accepted. Condensed context in ../design-doc.md; full original reasoning is preserved in git history from before the doc-set consolidation.

## Context

Stories M1 (HIPAA audit-control evidence) and A1 (DSAR "who viewed this" queries) both need to answer questions that span multiple kinds of access — views, exports, unmasks, deletions. Auditing configuration changes (who changed a RetentionPolicy's duration) surfaced later as a real gap: NIST 800-53's AU-2 requires logging "security or privacy attribute changes," not just data access, and nothing initially covered that.

## Decision

A single AccessAuditEvent entity, append-only, discriminated by an `action` field (`VIEW`, `EXPORT`, `UNMASK`, `DELETE`, `DELETION_SKIPPED`, and later `CONFIG_CHANGED`). One query answers "who did what to this record," across every action type, rather than requiring a compliance officer to reconcile several separate logs. Failed authentication attempts — which have neither a resolved actor nor a target record — deliberately do *not* fit this shape and reuse the existing SecurityEvent entity instead.

## Alternatives Considered

- **Separate log tables per action type (a ViewLog, an ExportLog, an UnmaskLog):** rejected — turns a single "produce an audit report" story into a multi-table reconciliation exercise, and multiplies the places immutability/GRANT restrictions have to be independently enforced.
- **Forcing failed-login events into AccessAuditEvent's shape:** rejected — would require inventing a placeholder actor and target record, defeating the precision of the entity.

## Consequences

- `access_audit_events`' own retention clock must be at least as long as the longest clock among everything it audits — an audit record about a since-deleted SessionEvent still has to prove the deletion was lawful (see clickhouse-schema.md).
- Adding `CONFIG_CHANGED` after the fact required touching three already-written files (event-catalog.md, event-schema.md, clickhouse-schema.md) to stay consistent — a real cost of the unified-entity choice, paid once, rather than an ongoing cost of a fragmented one.
- Every future compliance-relevant action gets evaluated against "does this belong as a new `action` value on this entity" before any alternative is considered.
