# Security

## Authentication

> Status: Research-based, current as of July 2026. Specifies how an actor's identity is established before [Authorization](README.md#authorization)'s role/permission model can apply — the mechanism `gateway` uses to populate the RequestContext every downstream AccessAuditEvent depends on, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams).

---

### Local Authentication — the Required Baseline

Per [Pricing](../01-product/README.md#pricing), Community and Team tier run without SSO, and per [System Context](../03-architecture/README.md#system-context), no Parity or Wedge capability may depend on an external system — so local username/password authentication is the baseline every deployment has, including Enterprise ones, not a fallback for smaller tiers only.

- **Password hashing: Argon2id**, the current OWASP-recommended default (parameters m=19456, t=2, p=1), not bcrypt — bcrypt remains acceptable only for legacy systems already using it at cost≥12, not a reasonable choice for a new system in 2026.
- **MFA: TOTP, not SMS.** Time-based one-time passwords operate entirely from a shared secret and a clock — no external SMS gateway, no dependency that would break in an air-gapped deployment. SMS-based OTP would silently fail the air-gapped requirement from [System Context](../03-architecture/README.md#system-context); TOTP doesn't, so it's the only MFA mechanism that belongs in the required path.

---

### SSO — the Enterprise-Tier Layer, Including for Air-Gapped Deployments

A common misconception worth ruling out explicitly: **air-gapped does not mean "no SSO."** It means no path to the public internet — it does not mean no internal network services. The standard, well-established pattern for exactly this situation is deploying the Identity Provider itself *inside* the same isolated network as the application: Keycloak, Auth0 Private Cloud, and ForgeRock all support fully on-premises deployment, with every part of the authentication flow — login screens, token issuance, session verification — happening entirely within the air-gapped boundary. A large air-gapped bank (per [Target Users](../00-overview/README.md#target-users)'s 300+-employee band) plausibly already runs an internal IdP (self-hosted Keycloak, or Active Directory Federation Services) for its other internal systems — Imora's SSO integration talks to *that*, never to a cloud identity provider.

- **Protocol:** SAML 2.0 and OIDC, the two protocols every self-hostable IdP (Keycloak, Auth0 Private Cloud, ForgeRock) supports.
- **Role mapping:** SSO group/claim assertions map to the four roles from [Authorization](README.md#authorization) (engineer, compliance_officer, platform_operator, admin) at login time — the mapping configuration itself is an Admin-role action, closing the loop with [Authorization](README.md#authorization)'s permission table.
- **Gating mechanism:** per [Pricing](../01-product/README.md#pricing), SSO is gated by the offline signed license file, not a phone-home check — consistent with every other Enterprise-tier feature in this doc set.

---

### Session and Token Handling

`gateway` issues a session token on successful authentication (local or SSO) and populates the RequestContext every downstream component consumes, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) — this is the identity that ends up in every `actorUserId` field across [Event Schema](../05-data/README.md#event-schema)'s AccessAuditEvent variants. A request with no valid session cannot reach `query-api`'s `AuditedQueryHandler` or `workers`' actions at all — authentication failure is a hard stop before authorization is even evaluated, not a soft-fail that falls through to a default permission level.

---

### What's Deliberately Not Modeled Here

- Exact token format (JWT vs. opaque reference token) and session lifetime values — implementation detail downstream of this design.
- Keycloak/IdP deployment configuration specifics — `12-infrastructure/`.
- Password reset and account-recovery flows — a product/UX concern once this mechanism is accepted.

---

Sources: [Air-Gapped Deployment Single Sign-On — hoop.dev](https://hoop.dev/blog/air-gapped-deployment-single-sign-on/), [OpenID Connect in Air-Gapped Environments — hoop.dev](https://hoop.dev/blog/openid-connect-in-air-gapped-environments/), [Keycloak SSO Integration — Glasswall (offline/on-premises deployment guide)](https://docs.glasswall.com/research/keycloak-sso-integration), [Password Hashing in 2026: bcrypt vs Argon2 vs scrypt vs PBKDF2 — Security Boulevard](https://securityboulevard.com/2026/06/bcrypt-vs-argon2-vs-scrypt-vs-pbkdf2-a-2026-decision-framework/).

### What This Feeds Next

`research/07-security/README.md#encryption` is the last piece [PII Redaction](README.md#pii-redaction) and [Authorization](README.md#authorization) both deferred — key rotation and the disk-level baseline. `research/07-security/README.md#audit-logging` and `research/07-security/README.md#threat-model` round out this folder.

---

## Authorization

> Status: Research-based, current as of July 2026. Resolves what [Postgres Schema](../05-data/README.md#postgres-schema), [Storage](../05-data/README.md#storage), and [PII Redaction](README.md#pii-redaction) all deferred here: the actual permission model behind the coarse `role` column, and who can trigger the sensitive actions this doc set has been describing since [Domain Model](../02-domain/README.md#domain-model).

---

### Why Role Alone Isn't Enough

[Postgres Schema](../05-data/README.md#postgres-schema) already noted that `users.role` is "enough to route a request, not enough to decide whether a specific field should be masked." That's not a gap to patch later — it's a direct consequence of HIPAA's **minimum necessary standard**, which requires limiting PHI access to the least needed for a specific purpose. In practice, the standard bodies of guidance on this treat RBAC as the *primary* mechanism (roles map to job function, which handles most access decisions correctly) with **attribute-based conditions layered on top for the genuinely sensitive cases** — exactly the hybrid shape Imora needs: coarse role for routing, contextual checks for anything touching masked PHI/PII.

---

### RBAC Baseline

The four roles from [Postgres Schema](../05-data/README.md#postgres-schema), mapped to what they can do:

| Action | Engineer | Compliance Officer | Platform Operator | Admin |
|---|---|---|---|---|
| View session replay / errors / performance (masked view) | Yes | Yes | No — see Separation of Duties below | Yes |
| View access-audit-trail (SessionViewed, RecordDeleted, etc.) | No | Yes | No | Yes |
| Apply / lift a LegalHold | No | Yes | No | Yes |
| Configure RetentionPolicy | No | Yes | Yes (infra-level defaults only) | Yes |
| Generate an EvidenceExport | No | Yes | No | Yes |
| Configure field classification (safe-allow-list vs. PHI-marked, per [PII Redaction](README.md#pii-redaction)) | Yes (developer-facing config) | Yes | No | Yes |
| Trigger an UNMASK request | Yes | Yes | No | Yes |
| Deploy / operate infrastructure | No | No | Yes | Yes |

This is deliberately coarse — it answers "can this role ever do this," not "should this specific request be allowed." The second question is where ABAC comes in.

---

### Where ABAC Applies: Exactly One Boundary, Not Session-Level Filtering

It would be possible to layer attribute checks onto *which sessions* an Engineer can view (e.g., "only sessions from accounts they're assigned to") — but that would directly undermine Chidi's persona requirement from [Product Requirements Document (PRD)](../01-product/README.md#product-requirements-document-prd): effective debugging depends on broad visibility into masked session data, and restricting that by assignment would degrade the parity experience the whole product depends on for daily adoption. **The attribute-based control is concentrated at exactly one boundary: the UNMASK action.**

This is the **break-the-glass** pattern — the standard term in healthcare access control for exactly what story M2 and [Business Rules](../02-domain/README.md#business-rules) BR-6 already describe: access beyond a role's default scope, granted immediately, but logged with a mandatory reason rather than gated behind a slow approval workflow. At the UNMASK boundary specifically:

- The requesting actor's role must permit UNMASK at all (per the RBAC table above).
- A non-empty `reason` is required (BR-6).
- **No role is exempt from this, including Compliance Officer and Admin.** A DPO investigating her own organization's compliance posture still has to state why she's viewing an unmasked field, and that access is still logged. Seniority or job function is not a substitute for the audit trail — the entire wedge collapses if any role gets a quiet bypass.

---

### Separation of Duties: Platform Operator Gets Infrastructure, Not Content

Per the RBAC table, Platform Operator (Priya's persona) can deploy, configure infrastructure-level retention defaults, and operate the system — but has no default path to view session content, the audit trail, or trigger exports. This is a deliberate separation-of-duties boundary, not an oversight: the person who can restart the database and the person who can view a patient's masked session should not, by default, be the same permission grant, even if in a small organization (per [Target Users](../00-overview/README.md#target-users)'s org-size variants) they're literally the same person wearing two hats. The two roles being held by one human doesn't collapse the two *permissions* into one — that person authenticates and acts under whichever role context the action requires, and both are logged under their identity either way.

---

### What's Deliberately Not Modeled Here

- SSO group-to-role mapping mechanics — [Authentication](README.md#authentication).
- The actual policy-engine implementation (a permissions table, OPA/Rego, or an in-process check) — downstream of this design, not part of it.
- UI-level permission affordances (what a Compliance Officer sees vs. an Engineer in the dashboard) — a product/design concern, not an authorization-model one; the enforcement point is `query-api`/`workers`, per [Bounded Contexts](../02-domain/README.md#bounded-contexts), regardless of what the UI shows.

---

Sources: [Implementing the HIPAA Minimum Necessary Standard — AccountableHQ](https://www.accountablehq.com/post/implementing-the-hipaa-minimum-necessary-standard-best-practices-and-policy-examples), [HIPAA Access Control Requirements Explained — Censinet](https://censinet.com/perspectives/hipaa-access-control-requirements-explained), [RBAC vs ABAC: Access Control for Sensitive Data — Skyflow](https://www.skyflow.com/post/rbac-vs-abac-access-control-for-sensitive-data), [RBAC vs. ABAC — Splunk](https://www.splunk.com/en_us/blog/learn/rbac-vs-abac.html).

### What This Feeds Next

`research/07-security/README.md#authentication` should specify how an actor's identity and role are established in the first place (local accounts vs. SSO, per [Pricing](../01-product/README.md#pricing)'s tier split). `research/07-security/README.md#encryption` should specify the key-rotation mechanics [PII Redaction](README.md#pii-redaction) deferred.

---

## Encryption

> Status: Research-based, current as of July 2026. Resolves the key-rotation mechanics [PII Redaction](README.md#pii-redaction) and [Authorization](README.md#authorization) both deferred here, and states the defense-in-depth layering underneath the SecureFieldVault.

---

### Two Layers, Not One

Per [PII Redaction](README.md#pii-redaction)'s "defense in depth" framing:

1. **Disk/volume-level encryption** — the outer layer, applied uniformly to ClickHouse, PostgreSQL, and object-storage volumes regardless of what's in them. This is the baseline every deployment gets, protecting against physical media loss or a storage-layer breach.
2. **Field-level envelope encryption (the SecureFieldVault)** — the inner layer, applied only to the soft-masked PHI/PII fields from [PII Redaction](README.md#pii-redaction)'s Tier 2 classification. This is what survives even if the outer layer is somehow compromised — the actual PHI value stays encrypted independent of disk access.

**Encryption in transit:** TLS on every connection in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)'s topology — browser-sdk to gateway, and every internal service-to-service call. This isn't a separate design decision; it's the uncontroversial baseline every layer above assumes.

---

### Key Rotation: Versioned KEKs, Not Bulk Re-Encryption

Rotating the vault's KEK (per [PII Redaction](README.md#pii-redaction)'s envelope-encryption design) does **not** require re-encrypting the underlying field data, and — more specifically — doesn't even require re-wrapping every existing DEK immediately. The standard, lower-overhead pattern:

- A new KEK version is generated and marked primary. **All new writes** wrap their DEK under the new version.
- **Old KEK versions are retained, active for decryption only.** Each encrypted field's metadata records which KEK version wrapped its DEK, so decryption always uses the correct historical key — there's no ambiguity and no need to know in advance which version applies.
- **No bulk re-wrap job runs on rotation.** This is the deliberate choice for Imora specifically: an eager approach that re-wraps every historical DEK on each rotation would be a real operational burden at the scale [Scaling](../03-architecture/README.md#scaling) already identified as the binding constraint (retention-driven accumulation, not throughput) — running a bulk key-migration job against years of accumulated vault entries directly conflicts with [Deployment Model](../03-architecture/README.md#deployment-model)'s single-machine, 2–3-person operability requirement. Lazy, versioned retention achieves the same security property (a compromised key stops being usable for *new* data immediately) without that cost.

**Old KEK version retention is governed by the same rule already established for AccessAuditEvent** in [ClickHouse Schema](../05-data/README.md#clickhouse-schema): a KEK version must never be destroyed before the longest applicable regulatory retention clock in the deployment has elapsed for every field it might still be protecting. Destroying an old KEK version early would make historical PHI permanently undecryptable — including PHI a legitimate, audited UNMASK request (story M2) might still need to reach within its retention window. This isn't a new rule invented for encryption specifically; it's the same "never younger than what it protects" principle applied to a different kind of record.

**Rotation triggers:** scheduled rotation (annually is a reasonable default cadence) and emergency rotation on any suspected key compromise — the latter happening immediately, not waiting for the next scheduled cycle.

---

### What's Deliberately Not Modeled Here

- Specific KMS/HSM product selection for the cluster-profile KEK source — [PII Redaction](README.md#pii-redaction) already specifies HashiCorp Vault as the reference option; this document doesn't re-litigate that choice.
- Exact TLS cipher suite / protocol version pinning — a `12-infrastructure/` configuration detail, not a design decision.
- The disk-encryption mechanism itself (LUKS, cloud-provider volume encryption, etc.) — deployment-specific, downstream of the requirement stated above.

---

Sources: [Key Rotation Strategies — Replacing Cryptographic Keys Without Downtime](https://www.qcecuring.com/education/key-management/key-rotation-strategies), [Key Rotation in KMS: What Really Happens to Your Encrypted Data?](https://medium.com/@madhurajayashanka/key-rotation-in-aws-and-gcp-kms-what-really-happens-to-your-encrypted-data-7d2a12b07303), [Envelope encryption — Google Cloud KMS Docs](https://docs.cloud.google.com/kms/research/envelope-encryption).

### What This Feeds Next

`research/07-security/README.md#audit-logging` and `research/07-security/README.md#threat-model` are the two remaining files in this folder — the former should specify operational logging (distinct from the AccessAuditEvent product feature already fully specified in `02-domain/` and `05-data/`), and the latter should stress-test the assumptions this whole folder has been making.

---

## Audit Logging

> Status: Research-based, current as of July 2026. Scoped deliberately narrow: this is **operational** audit logging — who changed a policy, who granted a role, who authenticated and how — as distinct from AccessAuditEvent, which is the product-facing data-access trail already fully specified in [Domain Model](../02-domain/README.md#domain-model), [Event Catalog](../02-domain/README.md#event-catalog), and [ClickHouse Schema](../05-data/README.md#clickhouse-schema). This document identifies a real gap in that existing work and closes it, rather than re-describing what's already built.

---

### The Gap: Configuration Changes Were Never Actually Logged

NIST 800-53's AU-2 control requires logging "security or privacy attribute changes" and "changes to user privileges" — not just access to data, but changes to the *rules governing* that access. Checking what's already specified against that standard surfaces a real hole: [Event Catalog](../02-domain/README.md#event-catalog) and [ClickHouse Schema](../05-data/README.md#clickhouse-schema) log every VIEW, EXPORT, UNMASK, DELETE, and DELETION_SKIPPED against data — but **nothing currently logs who changed a RetentionPolicy's duration, who reclassified a field from PHI to safe, or who granted a role.**

This isn't a cosmetic gap. [Business Rules](../02-domain/README.md#business-rules) BR-1's entire guarantee depends on the policy configuration itself being trustworthy — if someone can quietly shorten SessionEvent's retention from 6 years to 30 days with no record of having done so, the audit trail downstream of that change is worthless, and nobody would know to distrust it. The fix has to close this gap in the *same* audit mechanism already built, not bolt on a second one.

### The Fix: Extend AccessAuditEvent, Don't Build a Second Log

Rather than a parallel logging system, **`AccessAuditEvent`'s `action` enum gains one more value: `CONFIG_CHANGED`**, covering RetentionPolicy edits, field-classification changes (per [PII Redaction](README.md#pii-redaction)), and role/permission grants (per [Authorization](README.md#authorization)). This keeps [Business Rules](../02-domain/README.md#business-rules) BR-5 ("every sensitive action produces exactly one AccessAuditEvent") true without qualification, and means story M1's "one audit report" answers both "who viewed what" and "who changed the rules" from the same query — a HIPAA risk assessment shouldn't require reconciling two separate logs to get the full picture. Payload: `targetRecordType = RetentionPolicy | FieldClassification | UserRole`, `targetRecordId` the specific policy/field/user affected, and — following BR-6's existing pattern for UNMASK — the change itself (old value → new value) recorded in the event payload, not just the fact that *a* change happened.

I've made this concrete by adding `CONFIG_CHANGED` to the action enum in [Event Catalog](../02-domain/README.md#event-catalog), [Event Schema](../05-data/README.md#event-schema), and [ClickHouse Schema](../05-data/README.md#clickhouse-schema) directly, rather than leaving it as a proposal in this document that those files never catch up to.

---

### What Doesn't Fit AccessAuditEvent's Shape: Pre-Authentication Events

AccessAuditEvent assumes a resolved `actorUserId` and a `targetRecordId` — but a **failed login attempt** has neither: authentication didn't succeed, so there's no actor identity yet, and there's no record being accessed, just an attempt. Forcing this into AccessAuditEvent's shape would mean inventing a placeholder actor and a placeholder target, which defeats the purpose of that entity's precision.

**Failed authentication attempts reuse the existing `SecurityEvent` entity instead** (`signalType = "failed_authentication"`, `severity` scaling with repeated failures, `sessionId = null`) — no new entity required. This is the same entity [Event Catalog](../02-domain/README.md#event-catalog) already defined for security signals correlated into an incident timeline (story D2), and a repeated-failed-login pattern is exactly the kind of signal that timeline is meant to surface.

---

### System Lifecycle Events

Service start/stop and deployment events (per [Deployment Model](../03-architecture/README.md#deployment-model)'s two topology profiles) are operational telemetry, not compliance-relevant by the standard set above — they don't touch data access or policy configuration. These belong in ordinary infrastructure logging (`12-infrastructure/README.md#observability`), not the AccessAuditEvent/SecurityEvent audit mechanisms this document governs.

---

### What's Deliberately Not Modeled Here

- The exact UI/API surface for triggering a `CONFIG_CHANGED`-producing action — a `06-api/` concern.
- Alerting thresholds on repeated `failed_authentication` SecurityEvents (e.g., account lockout policy) — a product/security-policy decision downstream of this design.

---

Sources: [AU-2: Event Logging — CSF Tools (NIST SP 800-53 Rev. 5)](https://csf.tools/reference/nist-sp-800-53/r5/au/au-2/), [Understanding and Implementing NIST SP 800-53 AU-2 Logging Requirements — SecureStrux](https://securestrux.com/resources/cyber-advisory-center/understanding-and-implementing-nist-sp-800-53-au-2-logging-requirements-for-defense-industrial-base-systems/).

### What This Feeds Next

[Threat Model](README.md#threat-model) stress-tests the assumptions across all six files in this folder — including confirming that role grants producing a `CONFIG_CHANGED` event closes the elevation-of-privilege loop this document opened.

---

## PII Redaction

> Status: Research-based, current as of July 2026. Resolves two mechanisms [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) and [Domain Model](../02-domain/README.md#domain-model) deferred to this document: how a field gets classified into hard-redaction vs. soft-masking, and how the SecureFieldVault actually encrypts what it holds.

---

### Classification: Three Inputs, Two Outcomes

[Component Diagrams](../03-architecture/diagrams.md#component-diagrams)'s two-tier masking model (hard redaction vs. soft masking with escalation) describes the *outcomes*. This document specifies the *decision* that routes a field into one of them, evaluated in `browser-sdk` at capture time, in this order:

1. **Explicitly allow-listed as safe** (a developer-applied marker, e.g. a `data-imora-safe` attribute or equivalent CSS-class convention, the same mechanism Sentry's `sentry-mask`/`sentry-block` classes use) → captured as-is. This is the only path to Tier 1 (Parity) capture.
2. **Explicitly marked as known-PHI/PII** (`data-imora-mask="phi"` or equivalent) **or matched by a regex backstop** for common structural PII — email, US Social Security Number, credit card number, phone number, the same pattern set standard PII-detection tooling (e.g., Microsoft Presidio's rule-based layer) uses for structural entities — → soft-masked into the SecureFieldVault, unmaskable only via the audited escalation path from story M2.
3. **Neither of the above** → hard-redacted. Never captured, never stored, anywhere. This is the fail-closed default from [Business Rules](../02-domain/README.md#business-rules) BR-7 — a field nobody thought to classify gets the safest outcome, not the most convenient one.

The regex backstop in step 2 exists specifically for the failure mode named in [Problem Statement](../00-overview/README.md#problem-statement): a developer ships a new form field and forgets to classify it. Regex catches *structurally recognizable* PII (an SSN-shaped string) even on an unmarked field; it does not catch unstructured sensitive data (a diagnosis written in free text, a salary figure) — those still rely on step 1/2's explicit markers, and fall through to hard redaction by default if nobody added one. Contextual detection (NLP-based entity recognition for names, addresses) is a legitimate future enhancement beyond regex, not a day-one requirement — regex plus fail-closed default is the defensible baseline.

---

### The SecureFieldVault: Envelope Encryption, Sized to Deployment Profile

Standard practice for field-level encryption is envelope encryption: a data encryption key (DEK) encrypts the actual field value with AES-256-GCM, and a key-encryption key (KEK) — typically held in a KMS (AWS KMS, Azure Key Vault, HashiCorp Vault) — encrypts the DEK. This is the right shape for the vault, but the standard KMS options don't fit Imora's own constraints unmodified:

- **A cloud KMS is not viable as a required dependency.** Per [System Context](../03-architecture/README.md#system-context), no Parity or Wedge capability may depend on an external system in the required path, and unmasking a PHI field is squarely a Wedge capability (story M2).
- **A full separate KMS product (e.g., a dedicated HashiCorp Vault deployment) is disproportionate for the single-machine profile** Priya's story P1 depends on — it's another stateful service a 2–3 person team has to run, back up, and reason about in a security review, on top of everything else in [Deployment Model](../03-architecture/README.md#deployment-model).

The resolution follows the same two-profile pattern already established there, rather than inventing a third approach:

- **Single-machine profile:** the KEK is a locally-managed root key, generated at install time, stored as a file outside the application's own data volumes and protected by disk-level encryption (the "defense in depth" layering — application-level field encryption, plus disk-level encryption as the fallback — that field-encryption best practice recommends). No separate KMS service to operate.
- **Cluster/Enterprise profile:** the same envelope-encryption design supports swapping in a proper self-hostable KMS (HashiCorp Vault, self-managed) as the KEK source once the operational overhead is justified by scale — an upgrade path, not a day-one requirement, consistent with [Pricing](../01-product/README.md#pricing)'s Enterprise-tier feature set.

Decryption happens exclusively through `query-api`'s `UnmaskEscalationHandler`, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) — the vault itself never exposes a raw-decrypt path outside that audited flow, which is what makes BR-6's "every unmask is logged with a reason" actually true rather than bypassable through a lower-level API.

---

### What's Deliberately Not Modeled Here

- The exact attribute/CSS-class naming convention — an SDK implementation detail once this design is accepted.
- Key rotation schedule and procedure — `07-security/README.md#encryption`.
- IAM for who can configure which fields are allow-listed vs. PHI-marked — `07-security/README.md#authorization`.

---

Sources: [Real-Time PII Masking in Session Replay — hoop.dev](https://hoop.dev/blog/real-time-pii-masking-in-session-replay/), [PII Detection and Handling in Event Streams — Conduktor](https://www.conduktor.io/glossary/pii-detection-and-handling-in-event-streams), [Field-Level Encryption — CyberArk](https://www.cyberark.com/what-is/field-level-encryption/), [An Overview of Client-Side Field Level Encryption in MongoDB — Severalnines](https://severalnines.com/blog/overview-client-side-field-level-encryption-mongodb/).

### What This Feeds Next

`research/07-security/README.md#encryption` should specify key rotation and the disk-level encryption baseline this document assumes. `research/07-security/README.md#authorization` should specify who can configure field classifications and who can trigger an unmask request.

---

## Threat Model

> Status: Research-based, current as of July 2026. The last file in `07-security/` — stress-tests the five prior documents in this folder using STRIDE, against Imora's actual architecture rather than a generic checklist. Two findings here are new, not restatements: a database-level gap in how immutability is actually enforced, and a way the compliance mechanism itself (LegalHold) could become a denial-of-service vector against [Scaling](../03-architecture/README.md#scaling)'s own storage math.

---

### Spoofing

**Threat:** an unauthenticated client injects fake SessionEvents into `ingestion`, or forges a session token to assume another actor's identity.
**Mitigation, already specified:** `gateway` authenticates every request before it reaches `ingestion` or `query-api` ([Bounded Contexts](../02-domain/README.md#bounded-contexts)); TLS on every connection ([Encryption](README.md#encryption)); Argon2id-hashed local credentials and TOTP MFA with no external dependency ([Authentication](README.md#authentication)).

### Tampering

**Threat, newly identified here:** [Domain Model](../02-domain/README.md#domain-model) Invariant 1 requires immutability "enforced at the storage layer, not just in code" — but no prior document actually specified the database-level GRANTs that would make this true. Without them, "immutable" means only "the application doesn't expose a delete button," which anyone with direct ClickHouse access (a compromised host, an over-privileged service account, an insider with database credentials) can trivially bypass.

**Mitigation, specified here:** ClickHouse natively supports roles scoped to INSERT-only, with no DELETE or ALTER privilege — this isn't a workaround, it's a documented, first-class capability. The `ingestion` and `query-api` service accounts get **INSERT + SELECT only** on `access_audit_events`; no application service account, including `workers`', ever holds DELETE or ALTER on that specific table, even though `workers` legitimately deletes rows from other tables per BR-1/BR-2. Anyone needing to touch `access_audit_events` beyond insert/select requires a separate database-admin credential entirely outside the application's normal operation — which per [Authorization](README.md#authorization)'s separation-of-duties principle, should itself be a break-glass action logged outside the system it would otherwise be modifying (an out-of-band record, e.g., in infrastructure change logs, not `access_audit_events` itself, since an attacker capable of directly tampering with that table could tamper with its own tamper-log too).

**A confirming strength, not a gap:** [PII Redaction](README.md#pii-redaction) specified AES-256-GCM for the SecureFieldVault. GCM is an authenticated encryption mode — it provides tamper-evidence as an inherent property, not an add-on. A modified ciphertext fails to decrypt rather than silently decrypting to garbage, so vault tampering is self-detecting without any additional mechanism.

### Repudiation

**Threat, newly identified here:** [Domain Model](../02-domain/README.md#domain-model) specified `sequenceNumber` on every AccessAuditEvent "to enable gap detection" — but no document specified anything that actually *performs* that detection. A monotonic sequence number that nothing monitors is inert; an actor able to delete a record and its audit entry together (via the Tampering threat above, absent its mitigation) leaves a detectable gap only if something is actively looking for one.

**Mitigation, specified here:** `workers` gains a periodic integrity-check job — scanning `access_audit_events` for `sequenceNumber` discontinuities and alerting via `notification-service` if found. This is a small addition to an existing context, not a new one, consistent with [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s ownership model.

### Information Disclosure

**Threat:** SecureFieldVault key compromise exposing all historically vault-encrypted PHI at once.
**Mitigation, already specified:** envelope encryption with versioned KEKs, disk-level encryption as a second layer, key rotation on schedule and on suspected compromise ([Encryption](README.md#encryption)).

**Threat, newly identified here:** [PII Redaction](README.md#pii-redaction) doesn't specify *how* a masked placeholder is rendered — if masking preserves the original field's length (e.g., replacing "John Smith" with ten asterisks), an observer can infer information about the underlying value across many sessions even without ever unmasking it. **Mitigation:** masked placeholders should be fixed-format (a constant-length or generic label like `[masked]`), never length- or shape-preserving — a detail worth carrying into the SDK implementation this document doesn't otherwise specify.

### Denial of Service

**Threat, newly identified here — the sharpest finding in this document:** [Domain Model](../02-domain/README.md#domain-model) specified `LegalHold.scope` as a re-evaluated query, deliberately broad enough to cover new matching records automatically (per [Postgres Schema](../05-data/README.md#postgres-schema)'s design). But that same breadth is exactly what [Scaling](../03-architecture/README.md#scaling) identified as the thing to watch: an overly broad hold — `{"type": "data_subject", "value": "*"}` or a date range spanning years, applied carelessly or maliciously by a Compliance Officer-role account — pins an unbounded, growing set of records out of `ttl_only_drop_parts`' cheap-partition-drop path entirely, forcing exactly the accumulated-storage growth [Scaling](../03-architecture/README.md#scaling)'s math treats as the trigger for cluster migration. **A single overly broad LegalHold can silently convert Imora's own compliance mechanism into a storage-exhaustion vector against itself.**

**Mitigation, specified here:** a LegalHold scope predicate above a breadth threshold (no session-ID list, no bounded date range — effectively unbounded) requires a second approver (Admin, not just Compliance Officer) at creation time, and surfaces as a flagged item in the periodic retention-sweep report from `workers`. This doesn't weaken BR-2's protection — a legitimately broad hold for an active, large-scope investigation still goes through — it just makes an unbounded hold a deliberate, reviewed decision rather than a default nobody had to think twice about.

**Other DoS threats, already mitigated:** ingestion flood — rate limiting at `gateway` via the Redis cache ([Container Diagrams](../03-architecture/diagrams.md#container-diagrams)); unbounded dashboard queries — pagination/result limits at `query-api`, an implementation detail this document doesn't further specify.

### Elevation of Privilege

**Threat:** a compromised Engineer-role account self-grants Compliance Officer permissions to gain UNMASK access.
**Mitigation, already specified:** role grants require existing Admin/Compliance Officer permission ([Authorization](README.md#authorization)'s RBAC table), and — as of [Audit Logging](README.md#audit-logging) — every role grant now produces its own `CONFIG_CHANGED` AccessAuditEvent, closing the loop this document's own predecessor started.

**Threat, newly identified here:** an UNMASK reason field satisfies BR-6's "non-empty" requirement with a junk value ("asdf", "debug") that no one ever reviews. A logged-but-unreviewed reason is compliance theater, not oversight. **Mitigation:** the same periodic `workers` job that checks sequence gaps should also surface an UNMASK-frequency report per actor to Compliance Officer role — not blocking any individual unmask (that would defeat the break-the-glass pattern's speed), but making abuse patterns visible after the fact, which is the same trade-off break-the-glass access already makes everywhere else.

---

### What's Deliberately Not Modeled Here

- Formal penetration testing or a specific CVE-tracking process — an operational security practice, not an architecture decision.
- The exact breadth threshold that triggers the LegalHold second-approval requirement — a policy tuning decision downstream of this design, not part of it.
- Physical security of self-hosted deployment hardware — genuinely out of this document set's scope; it's the deploying organization's own facility.

---

Sources: [Access control and account management — ClickHouse Docs](https://clickhouse.com/research/operations/access-rights), [How to Use GRANT and REVOKE in ClickHouse](https://oneuptime.com/blog/post/2026-03-31-clickhouse-grant-revoke/view).

### What This Closes Out

This is the last file in `research/07-security/`. All six files — [Authentication](README.md#authentication), [Authorization](README.md#authorization), [Encryption](README.md#encryption), [PII Redaction](README.md#pii-redaction), [Audit Logging](README.md#audit-logging), and this one — are now internally consistent, and this document's two new findings (database-level GRANT restrictions, the LegalHold-as-DoS-vector) should be treated as binding additions to [ClickHouse Schema](../05-data/README.md#clickhouse-schema) and [Business Rules](../02-domain/README.md#business-rules) respectively when those are next revisited, not optional hardening.

