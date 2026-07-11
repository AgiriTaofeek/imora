# Threat Model

> Status: Research-based, current as of July 2026. The last file in `08-security/` — stress-tests the five prior documents in this folder using STRIDE, against Imora's actual architecture rather than a generic checklist. Two findings here are new, not restatements: a database-level gap in how immutability is actually enforced, and a way the compliance mechanism itself (LegalHold) could become a denial-of-service vector against [scaling.md](../04-architecture/scaling.md)'s own storage math.

---

## Spoofing

**Threat:** an unauthenticated client injects fake SessionEvents into `ingestion`, or forges a session token to assume another actor's identity.
**Mitigation, already specified:** `gateway` authenticates every request before it reaches `ingestion` or `query-api` ([bounded-contexts.md](../02-domain/bounded-contexts.md)); TLS on every connection ([encryption.md](encryption.md)); Argon2id-hashed local credentials and TOTP MFA with no external dependency ([authentication.md](authentication.md)).

## Tampering

**Threat, newly identified here:** [domain-model.md](../02-domain/domain-model.md) Invariant 1 requires immutability "enforced at the storage layer, not just in code" — but no prior document actually specified the database-level GRANTs that would make this true. Without them, "immutable" means only "the application doesn't expose a delete button," which anyone with direct ClickHouse access (a compromised host, an over-privileged service account, an insider with database credentials) can trivially bypass.

**Mitigation, specified here:** ClickHouse natively supports roles scoped to INSERT-only, with no DELETE or ALTER privilege — this isn't a workaround, it's a documented, first-class capability. The `ingestion` and `query-api` service accounts get **INSERT + SELECT only** on `access_audit_events`; no application service account, including `workers`', ever holds DELETE or ALTER on that specific table, even though `workers` legitimately deletes rows from other tables per BR-1/BR-2. Anyone needing to touch `access_audit_events` beyond insert/select requires a separate database-admin credential entirely outside the application's normal operation — which per [authorization.md](authorization.md)'s separation-of-duties principle, should itself be a break-glass action logged outside the system it would otherwise be modifying (an out-of-band record, e.g., in infrastructure change logs, not `access_audit_events` itself, since an attacker capable of directly tampering with that table could tamper with its own tamper-log too).

**A confirming strength, not a gap:** [pii-redaction.md](pii-redaction.md) specified AES-256-GCM for the SecureFieldVault. GCM is an authenticated encryption mode — it provides tamper-evidence as an inherent property, not an add-on. A modified ciphertext fails to decrypt rather than silently decrypting to garbage, so vault tampering is self-detecting without any additional mechanism.

## Repudiation

**Threat, newly identified here:** [domain-model.md](../02-domain/domain-model.md) specified `sequenceNumber` on every AccessAuditEvent "to enable gap detection" — but no document specified anything that actually *performs* that detection. A monotonic sequence number that nothing monitors is inert; an actor able to delete a record and its audit entry together (via the Tampering threat above, absent its mitigation) leaves a detectable gap only if something is actively looking for one.

**Mitigation, specified here:** `workers` gains a periodic integrity-check job — scanning `access_audit_events` for `sequenceNumber` discontinuities and alerting via `notification-service` if found. This is a small addition to an existing context, not a new one, consistent with [bounded-contexts.md](../02-domain/bounded-contexts.md)'s ownership model.

## Information Disclosure

**Threat:** SecureFieldVault key compromise exposing all historically vault-encrypted PHI at once.
**Mitigation, already specified:** envelope encryption with versioned KEKs, disk-level encryption as a second layer, key rotation on schedule and on suspected compromise ([encryption.md](encryption.md)).

**Threat, newly identified here:** [pii-redaction.md](pii-redaction.md) doesn't specify *how* a masked placeholder is rendered — if masking preserves the original field's length (e.g., replacing "John Smith" with ten asterisks), an observer can infer information about the underlying value across many sessions even without ever unmasking it. **Mitigation:** masked placeholders should be fixed-format (a constant-length or generic label like `[masked]`), never length- or shape-preserving — a detail worth carrying into the SDK implementation this document doesn't otherwise specify.

## Denial of Service

**Threat, newly identified here — the sharpest finding in this document:** [domain-model.md](../02-domain/domain-model.md) specified `LegalHold.scope` as a re-evaluated query, deliberately broad enough to cover new matching records automatically (per [postgres-schema.md](../06-data/postgres-schema.md)'s design). But that same breadth is exactly what [scaling.md](../04-architecture/scaling.md) identified as the thing to watch: an overly broad hold — `{"type": "data_subject", "value": "*"}` or a date range spanning years, applied carelessly or maliciously by a Compliance Officer-role account — pins an unbounded, growing set of records out of `ttl_only_drop_parts`' cheap-partition-drop path entirely, forcing exactly the accumulated-storage growth [scaling.md](../04-architecture/scaling.md)'s math treats as the trigger for cluster migration. **A single overly broad LegalHold can silently convert Imora's own compliance mechanism into a storage-exhaustion vector against itself.**

**Mitigation, specified here:** a LegalHold scope predicate above a breadth threshold (no session-ID list, no bounded date range — effectively unbounded) requires a second approver (Admin, not just Compliance Officer) at creation time, and surfaces as a flagged item in the periodic retention-sweep report from `workers`. This doesn't weaken BR-2's protection — a legitimately broad hold for an active, large-scope investigation still goes through — it just makes an unbounded hold a deliberate, reviewed decision rather than a default nobody had to think twice about.

**Other DoS threats, already mitigated:** ingestion flood — rate limiting at `gateway` via the Redis cache ([container-diagrams.md](../04-architecture/container-diagrams.md)); unbounded dashboard queries — pagination/result limits at `query-api`, an implementation detail this document doesn't further specify.

## Elevation of Privilege

**Threat:** a compromised Engineer-role account self-grants Compliance Officer permissions to gain UNMASK access.
**Mitigation, already specified:** role grants require existing Admin/Compliance Officer permission ([authorization.md](authorization.md)'s RBAC table), and — as of [audit-logging.md](audit-logging.md) — every role grant now produces its own `CONFIG_CHANGED` AccessAuditEvent, closing the loop this document's own predecessor started.

**Threat, newly identified here:** an UNMASK reason field satisfies BR-6's "non-empty" requirement with a junk value ("asdf", "debug") that no one ever reviews. A logged-but-unreviewed reason is compliance theater, not oversight. **Mitigation:** the same periodic `workers` job that checks sequence gaps should also surface an UNMASK-frequency report per actor to Compliance Officer role — not blocking any individual unmask (that would defeat the break-the-glass pattern's speed), but making abuse patterns visible after the fact, which is the same trade-off break-the-glass access already makes everywhere else.

---

## What's Deliberately Not Modeled Here

- Formal penetration testing or a specific CVE-tracking process — an operational security practice, not an architecture decision.
- The exact breadth threshold that triggers the LegalHold second-approval requirement — a policy tuning decision downstream of this design, not part of it.
- Physical security of self-hosted deployment hardware — genuinely out of this document set's scope; it's the deploying organization's own facility.

---

Sources: [Access control and account management — ClickHouse Docs](https://clickhouse.com/docs/operations/access-rights), [How to Use GRANT and REVOKE in ClickHouse](https://oneuptime.com/blog/post/2026-03-31-clickhouse-grant-revoke/view).

## What This Closes Out

This is the last file in `docs/08-security/`. All six files — [authentication.md](authentication.md), [authorization.md](authorization.md), [encryption.md](encryption.md), [pii-redaction.md](pii-redaction.md), [audit-logging.md](audit-logging.md), and this one — are now internally consistent, and this document's two new findings (database-level GRANT restrictions, the LegalHold-as-DoS-vector) should be treated as binding additions to [clickhouse-schema.md](../06-data/clickhouse-schema.md) and [business-rules.md](../02-domain/business-rules.md) respectively when those are next revisited, not optional hardening.
