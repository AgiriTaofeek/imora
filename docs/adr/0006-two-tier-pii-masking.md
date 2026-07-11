# 0006. Two-tier masking: hard redaction vs. vault-backed soft masking, decided at capture time

> Status: Accepted. Condensed context in ../design-doc.md; full original reasoning is preserved in git history from before the doc-set consolidation.

## Context

domain-model.md's Invariant 4 (mask before storage, deny-by-default) and story M2 (PHI fields escalatable to unmasked with justification) appear to contradict each other: if an unrecognized field's value is never captured, there's nothing left to unmask later. Something had to reconcile "fail closed on unknown fields" with "some known-sensitive fields must remain recoverable."

## Decision

Masking is two distinct mechanisms, both decided in `browser-sdk` at capture time:

1. **Hard redaction** — a field matching no explicit allow-list rule and no PHI/PII marker. The real value is never captured anywhere, encrypted or not. Irreversible by design.
2. **Soft masking with escalation** — a field explicitly marked as known-PHI/PII (or matched by the regex backstop for structural patterns like SSNs). The real value is captured into the SecureFieldVault, envelope-encrypted, and recoverable only through the audited `UnmaskEscalationHandler` path.

Only category 2 is ever unmaskable.

## Alternatives Considered

- **Single masking tier (mask everything the same way):** rejected — can't simultaneously satisfy "fail closed on the unknown" and "let a justified, audited unmask recover a known field."
- **Capture everything, mask only at query time:** rejected outright — means the unmasked value already exists at rest, failing deny-by-default even if every read path is correctly filtered; this is also the specific anti-pattern problem-statement.md's "new field ships unredacted" failure mode describes.

## Consequences

- `dashboard` gets zero translation authority over sensitive fields — enforcing this distinction anywhere in the presentation layer would create a second place masking logic could drift from capture-time enforcement.
- The regex backstop only catches structurally recognizable PII (SSN/credit-card/email/phone shapes) — unstructured sensitive data (free-text diagnoses) still depends on explicit marking, and correctly falls through to hard redaction by default if nobody adds one.
- The masked placeholder itself must not be length- or shape-preserving (per threat-model.md's Information Disclosure finding), or it leaks information about the underlying value even without ever being unmasked.
