# Business Rules

> Status: Research-based, current as of July 2026. Expands [domain-model.md](domain-model.md)'s five invariants and [bounded-contexts.md](bounded-contexts.md)'s AccessAuditEvent-ownership rule into full specifications, including the edge cases a regulated deployment will actually hit. Each rule states what happens, why, and which entity/context from the prior two documents enforces it.

---

## Rule Set A — Retention and Deletion

### BR-1: Retention period is assigned per data category, and the longest applicable period wins when categories overlap

A Session, ErrorEvent, or SecurityEvent may fall under more than one regulation simultaneously — e.g., a healthcare organization's session data is subject to both HIPAA (6-year floor) and, if the patient is an EU resident, GDPR. When retention requirements conflict, the strictest (longest) applicable period governs, and the specific regulatory basis for that choice is recorded on the RetentionPolicy itself — not left implicit. This is standard practice across regulated data categories: where multiple regimes apply, one policy set to the strictest requirement satisfies all of them simultaneously, rather than maintaining parallel conflicting policies per regulation.

**Enforced by:** `workers` (RetentionPolicy execution), per [bounded-contexts.md](bounded-contexts.md).

### BR-2: A scheduled deletion must check for an active LegalHold immediately before executing, not at scheduling time

Restates [domain-model.md](domain-model.md) Invariant 2 with the specific failure mode it prevents: checking for a hold only when a deletion job is *scheduled*, rather than immediately before it *executes*, creates a race condition — a hold applied after scheduling but before execution would be silently missed. The check-before-destroy step must be the last thing that happens before a delete, not a precondition evaluated earlier in the pipeline. When the check finds an active hold, the deletion is skipped and the skip is logged as an AccessAuditEvent; when a hold is later lifted, the record re-enters the normal retention schedule on its next scheduled evaluation — lifting a hold does not trigger an immediate retroactive deletion.

**Enforced by:** `workers`, immediately preceding every deletion execution.

### BR-3: A GDPR erasure request is honored except where a specific legal obligation or active legal hold requires retention — and even then, only what's strictly necessary is retained

This is the sharpest conflict Adaeze (the DPO persona) will actually hit: a data subject's erasure request against a session that Marcus's hospital is independently required to retain for HIPAA's 6-year floor. GDPR Article 17(3) provides exactly this exception — erasure isn't required where retention is necessary for compliance with a legal obligation (17(3)(b)) or the establishment, exercise, or defense of legal claims (17(3)(e)). **But the exception is not a blanket refusal**: where it's possible to satisfy the legal obligation while still deleting everything not strictly necessary for it, the exception only covers the narrower set. The correct behavior is **selective, field-level purging** — anonymize or delete the fields GDPR requires erased, while preserving only the minimum structure the overriding legal obligation actually requires (e.g., a HIPAA-required audit trail entry can often survive with the subject's identifying fields anonymized, rather than the whole record being retained or the whole record being deleted).

Every partial refusal must be logged with the specific regulatory basis cited (Article 17(3)(b) vs (e), or the specific HIPAA/SOX/PCI-DSS clause), so Adaeze can produce that justification to a regulator or the data subject directly — an unexplained "cannot delete" is not sufficient.

**Enforced by:** `workers` (executes the selective purge) and `query-api`/AccessAuditEvent (the refusal-with-basis is itself logged).

### BR-4: EvidenceExport is immune to both BR-1 and BR-3 once generated

Restates [domain-model.md](domain-model.md)'s resolution: an EvidenceExport is a frozen, self-contained copy at generation time. Neither a later retention purge (BR-1) nor a subsequent erasure request (BR-3) may alter or invalidate an already-generated export — the export's `contentHash` exists specifically so this immutability is independently verifiable, not just asserted.

**Enforced by:** `workers` (export generation), verified by `contentHash`.

---

## Rule Set B — Access, Audit, and Masking

### BR-5: Every VIEW, EXPORT, UNMASK, and DELETE against a sensitive record produces exactly one AccessAuditEvent

Restates [domain-model.md](domain-model.md) Invariant 1 with the ownership rule from [bounded-contexts.md](bounded-contexts.md): this happens in `query-api` (VIEW, EXPORT, UNMASK) and `workers` (DELETE, including the BR-2 skip case), never in `dashboard`. A read that completes without producing an AccessAuditEvent is a defect, not an acceptable fast path — including for internal/admin tooling, which is a common gap where audit trails silently don't apply.

### BR-6: UNMASK requires a non-empty, human-readable reason, and the reason is part of the audit record, not metadata about it

Restates [domain-model.md](domain-model.md) Invariant 5. The reason field exists so a HIPAA risk assessment or DSAR response can show *why* PHI was unmasked, not just that it was — "debugging" is an acceptable reason; a blank field is not.

### BR-7: PII/PHI masking is evaluated at capture time in browser-sdk; a field with no matching allow-list rule is masked by default

Restates [domain-model.md](domain-model.md) Invariant 4 with the failure mode it exists to prevent: a new form field shipped without an explicit masking rule must render masked, not unmasked-until-someone-notices. This is deny-by-default specifically because the alternative (block-list: mask known-sensitive fields, capture everything else) is the pattern that produced the PII-leak fact patterns behind Cost Driver 2 in [problem-statement.md](../00-overview/problem-statement.md).

---

## Rule Set C — Conflict Precedence Summary

When more than one rule above could apply to the same action, precedence resolves in this order, most authoritative first:

1. **Active LegalHold** (BR-2) — blocks deletion outright, regardless of what any retention policy or erasure request says.
2. **Legal obligation requiring retention** (BR-3's exception) — permits refusing full erasure, but only for the minimum data the obligation actually requires.
3. **Longest applicable RetentionPolicy** (BR-1) — governs ordinary scheduled deletion where no hold or overriding legal obligation applies.
4. **Erasure/deletion request** (BR-3's default case) — honored in full once nothing above overrides it.

This ordering is itself a business rule, not just a reading convenience: a `workers` implementation that evaluates these checks out of order (e.g., applying BR-1's retention clock before checking BR-2's hold) would produce the exact race condition BR-2 exists to prevent.

---

Sources: [Art. 17 GDPR – Right to erasure](https://gdpr-info.eu/art-17-gdpr/), [What Is GDPR Article 17 and 4 Ways to Achieve Compliance — Exabeam](https://www.exabeam.com/explainers/gdpr-compliance/what-is-gdpr-article-17-right-to-erasure-and-4-ways-to-achieve-compliance/), [Compliance Log Retention Requirements by Regulation](https://claudiasop.com/blog/compliance-log-retention-requirements.html), [Medical Record Retention: State Mandates vs. Federal Law](https://www.complydome.com/compliance-resources/state-mandates-vs-federal-law-a-small-practice-guide-to-which-record-retention-rule-wins-cms-hipaa-state-laws), [Legal Hold 101: Data Retention and Destruction](https://www.daymarksi.com/information-technology-navigator-blog/legal-hold-101-data-retention-and-destruction), [Defensible Data Deletion After a Legal Hold — Onna](https://www.onna.com/resources/blog/defensible-data-deletion-after-a-legal-hold).

## What This Feeds Next

`docs/02-domain/event-catalog.md` should enumerate the concrete AccessAuditEvent actions (VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED per BR-2) and SessionEvent subtypes already named across [domain-model.md](domain-model.md) and this document. `docs/08-security/pii-redaction.md` and `docs/06-data/retention.md` should implement BR-1 through BR-7 directly rather than re-deriving them.
