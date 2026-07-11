# Encryption

> Status: Research-based, current as of July 2026. Resolves the key-rotation mechanics [pii-redaction.md](pii-redaction.md) and [authorization.md](authorization.md) both deferred here, and states the defense-in-depth layering underneath the SecureFieldVault.

---

## Two Layers, Not One

Per [pii-redaction.md](pii-redaction.md)'s "defense in depth" framing:

1. **Disk/volume-level encryption** — the outer layer, applied uniformly to ClickHouse, PostgreSQL, and object-storage volumes regardless of what's in them. This is the baseline every deployment gets, protecting against physical media loss or a storage-layer breach.
2. **Field-level envelope encryption (the SecureFieldVault)** — the inner layer, applied only to the soft-masked PHI/PII fields from [pii-redaction.md](pii-redaction.md)'s Tier 2 classification. This is what survives even if the outer layer is somehow compromised — the actual PHI value stays encrypted independent of disk access.

**Encryption in transit:** TLS on every connection in [container-diagrams.md](../04-architecture/container-diagrams.md)'s topology — browser-sdk to gateway, and every internal service-to-service call. This isn't a separate design decision; it's the uncontroversial baseline every layer above assumes.

---

## Key Rotation: Versioned KEKs, Not Bulk Re-Encryption

Rotating the vault's KEK (per [pii-redaction.md](pii-redaction.md)'s envelope-encryption design) does **not** require re-encrypting the underlying field data, and — more specifically — doesn't even require re-wrapping every existing DEK immediately. The standard, lower-overhead pattern:

- A new KEK version is generated and marked primary. **All new writes** wrap their DEK under the new version.
- **Old KEK versions are retained, active for decryption only.** Each encrypted field's metadata records which KEK version wrapped its DEK, so decryption always uses the correct historical key — there's no ambiguity and no need to know in advance which version applies.
- **No bulk re-wrap job runs on rotation.** This is the deliberate choice for Imora specifically: an eager approach that re-wraps every historical DEK on each rotation would be a real operational burden at the scale [scaling.md](../04-architecture/scaling.md) already identified as the binding constraint (retention-driven accumulation, not throughput) — running a bulk key-migration job against years of accumulated vault entries directly conflicts with [deployment-model.md](../04-architecture/deployment-model.md)'s single-machine, 2–3-person operability requirement. Lazy, versioned retention achieves the same security property (a compromised key stops being usable for *new* data immediately) without that cost.

**Old KEK version retention is governed by the same rule already established for AccessAuditEvent** in [clickhouse-schema.md](../06-data/clickhouse-schema.md): a KEK version must never be destroyed before the longest applicable regulatory retention clock in the deployment has elapsed for every field it might still be protecting. Destroying an old KEK version early would make historical PHI permanently undecryptable — including PHI a legitimate, audited UNMASK request (story M2) might still need to reach within its retention window. This isn't a new rule invented for encryption specifically; it's the same "never younger than what it protects" principle applied to a different kind of record.

**Rotation triggers:** scheduled rotation (annually is a reasonable default cadence) and emergency rotation on any suspected key compromise — the latter happening immediately, not waiting for the next scheduled cycle.

---

## What's Deliberately Not Modeled Here

- Specific KMS/HSM product selection for the cluster-profile KEK source — [pii-redaction.md](pii-redaction.md) already specifies HashiCorp Vault as the reference option; this document doesn't re-litigate that choice.
- Exact TLS cipher suite / protocol version pinning — a `09-infrastructure/` configuration detail, not a design decision.
- The disk-encryption mechanism itself (LUKS, cloud-provider volume encryption, etc.) — deployment-specific, downstream of the requirement stated above.

---

Sources: [Key Rotation Strategies — Replacing Cryptographic Keys Without Downtime](https://www.qcecuring.com/education/key-management/key-rotation-strategies), [Key Rotation in KMS: What Really Happens to Your Encrypted Data?](https://medium.com/@madhurajayashanka/key-rotation-in-aws-and-gcp-kms-what-really-happens-to-your-encrypted-data-7d2a12b07303), [Envelope encryption — Google Cloud KMS Docs](https://docs.cloud.google.com/kms/docs/envelope-encryption).

## What This Feeds Next

`docs/08-security/audit-logging.md` and `docs/08-security/threat-model.md` are the two remaining files in this folder — the former should specify operational logging (distinct from the AccessAuditEvent product feature already fully specified in `02-domain/` and `06-data/`), and the latter should stress-test the assumptions this whole folder has been making.
