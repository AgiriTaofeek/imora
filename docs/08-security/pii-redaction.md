# PII Redaction

> Status: Research-based, current as of July 2026. Resolves two mechanisms [component-diagrams.md](../04-architecture/component-diagrams.md) and [domain-model.md](../02-domain/domain-model.md) deferred to this document: how a field gets classified into hard-redaction vs. soft-masking, and how the SecureFieldVault actually encrypts what it holds.

---

## Classification: Three Inputs, Two Outcomes

[component-diagrams.md](../04-architecture/component-diagrams.md)'s two-tier masking model (hard redaction vs. soft masking with escalation) describes the *outcomes*. This document specifies the *decision* that routes a field into one of them, evaluated in `browser-sdk` at capture time, in this order:

1. **Explicitly allow-listed as safe** (a developer-applied marker, e.g. a `data-imora-safe` attribute or equivalent CSS-class convention, the same mechanism Sentry's `sentry-mask`/`sentry-block` classes use) → captured as-is. This is the only path to Tier 1 (Parity) capture.
2. **Explicitly marked as known-PHI/PII** (`data-imora-mask="phi"` or equivalent) **or matched by a regex backstop** for common structural PII — email, US Social Security Number, credit card number, phone number, the same pattern set standard PII-detection tooling (e.g., Microsoft Presidio's rule-based layer) uses for structural entities — → soft-masked into the SecureFieldVault, unmaskable only via the audited escalation path from story M2.
3. **Neither of the above** → hard-redacted. Never captured, never stored, anywhere. This is the fail-closed default from [business-rules.md](../02-domain/business-rules.md) BR-7 — a field nobody thought to classify gets the safest outcome, not the most convenient one.

The regex backstop in step 2 exists specifically for the failure mode named in [problem-statement.md](../00-overview/problem-statement.md): a developer ships a new form field and forgets to classify it. Regex catches *structurally recognizable* PII (an SSN-shaped string) even on an unmarked field; it does not catch unstructured sensitive data (a diagnosis written in free text, a salary figure) — those still rely on step 1/2's explicit markers, and fall through to hard redaction by default if nobody added one. Contextual detection (NLP-based entity recognition for names, addresses) is a legitimate future enhancement beyond regex, not a day-one requirement — regex plus fail-closed default is the defensible baseline.

---

## The SecureFieldVault: Envelope Encryption, Sized to Deployment Profile

Standard practice for field-level encryption is envelope encryption: a data encryption key (DEK) encrypts the actual field value with AES-256-GCM, and a key-encryption key (KEK) — typically held in a KMS (AWS KMS, Azure Key Vault, HashiCorp Vault) — encrypts the DEK. This is the right shape for the vault, but the standard KMS options don't fit Imora's own constraints unmodified:

- **A cloud KMS is not viable as a required dependency.** Per [system-context.md](../04-architecture/system-context.md), no Parity or Wedge capability may depend on an external system in the required path, and unmasking a PHI field is squarely a Wedge capability (story M2).
- **A full separate KMS product (e.g., a dedicated HashiCorp Vault deployment) is disproportionate for the single-machine profile** Priya's story P1 depends on — it's another stateful service a 2–3 person team has to run, back up, and reason about in a security review, on top of everything else in [deployment-model.md](../04-architecture/deployment-model.md).

The resolution follows the same two-profile pattern already established there, rather than inventing a third approach:

- **Single-machine profile:** the KEK is a locally-managed root key, generated at install time, stored as a file outside the application's own data volumes and protected by disk-level encryption (the "defense in depth" layering — application-level field encryption, plus disk-level encryption as the fallback — that field-encryption best practice recommends). No separate KMS service to operate.
- **Cluster/Enterprise profile:** the same envelope-encryption design supports swapping in a proper self-hostable KMS (HashiCorp Vault, self-managed) as the KEK source once the operational overhead is justified by scale — an upgrade path, not a day-one requirement, consistent with [pricing.md](../01-product/pricing.md)'s Enterprise-tier feature set.

Decryption happens exclusively through `query-api`'s `UnmaskEscalationHandler`, per [component-diagrams.md](../04-architecture/component-diagrams.md) — the vault itself never exposes a raw-decrypt path outside that audited flow, which is what makes BR-6's "every unmask is logged with a reason" actually true rather than bypassable through a lower-level API.

---

## What's Deliberately Not Modeled Here

- The exact attribute/CSS-class naming convention — an SDK implementation detail once this design is accepted.
- Key rotation schedule and procedure — `08-security/encryption.md`.
- IAM for who can configure which fields are allow-listed vs. PHI-marked — `08-security/authorization.md`.

---

Sources: [Real-Time PII Masking in Session Replay — hoop.dev](https://hoop.dev/blog/real-time-pii-masking-in-session-replay/), [PII Detection and Handling in Event Streams — Conduktor](https://www.conduktor.io/glossary/pii-detection-and-handling-in-event-streams), [Field-Level Encryption — CyberArk](https://www.cyberark.com/what-is/field-level-encryption/), [An Overview of Client-Side Field Level Encryption in MongoDB — Severalnines](https://severalnines.com/blog/overview-client-side-field-level-encryption-mongodb/).

## What This Feeds Next

`docs/08-security/encryption.md` should specify key rotation and the disk-level encryption baseline this document assumes. `docs/08-security/authorization.md` should specify who can configure field classifications and who can trigger an unmask request.
