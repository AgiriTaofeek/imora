# Authorization

> Status: Research-based, current as of July 2026. Resolves what [postgres-schema.md](../06-data/postgres-schema.md), [storage.md](../06-data/storage.md), and [pii-redaction.md](pii-redaction.md) all deferred here: the actual permission model behind the coarse `role` column, and who can trigger the sensitive actions this doc set has been describing since [domain-model.md](../02-domain/domain-model.md).

---

## Why Role Alone Isn't Enough

[postgres-schema.md](../06-data/postgres-schema.md) already noted that `users.role` is "enough to route a request, not enough to decide whether a specific field should be masked." That's not a gap to patch later — it's a direct consequence of HIPAA's **minimum necessary standard**, which requires limiting PHI access to the least needed for a specific purpose. In practice, the standard bodies of guidance on this treat RBAC as the *primary* mechanism (roles map to job function, which handles most access decisions correctly) with **attribute-based conditions layered on top for the genuinely sensitive cases** — exactly the hybrid shape Imora needs: coarse role for routing, contextual checks for anything touching masked PHI/PII.

---

## RBAC Baseline

The four roles from [postgres-schema.md](../06-data/postgres-schema.md), mapped to what they can do:

| Action | Engineer | Compliance Officer | Platform Operator | Admin |
|---|---|---|---|---|
| View session replay / errors / performance (masked view) | Yes | Yes | No — see Separation of Duties below | Yes |
| View access-audit-trail (SessionViewed, RecordDeleted, etc.) | No | Yes | No | Yes |
| Apply / lift a LegalHold | No | Yes | No | Yes |
| Configure RetentionPolicy | No | Yes | Yes (infra-level defaults only) | Yes |
| Generate an EvidenceExport | No | Yes | No | Yes |
| Configure field classification (safe-allow-list vs. PHI-marked, per [pii-redaction.md](pii-redaction.md)) | Yes (developer-facing config) | Yes | No | Yes |
| Trigger an UNMASK request | Yes | Yes | No | Yes |
| Deploy / operate infrastructure | No | No | Yes | Yes |

This is deliberately coarse — it answers "can this role ever do this," not "should this specific request be allowed." The second question is where ABAC comes in.

---

## Where ABAC Applies: Exactly One Boundary, Not Session-Level Filtering

It would be possible to layer attribute checks onto *which sessions* an Engineer can view (e.g., "only sessions from accounts they're assigned to") — but that would directly undermine Chidi's persona requirement from [prd.md](../01-product/prd.md): effective debugging depends on broad visibility into masked session data, and restricting that by assignment would degrade the parity experience the whole product depends on for daily adoption. **The attribute-based control is concentrated at exactly one boundary: the UNMASK action.**

This is the **break-the-glass** pattern — the standard term in healthcare access control for exactly what story M2 and [business-rules.md](../02-domain/business-rules.md) BR-6 already describe: access beyond a role's default scope, granted immediately, but logged with a mandatory reason rather than gated behind a slow approval workflow. At the UNMASK boundary specifically:

- The requesting actor's role must permit UNMASK at all (per the RBAC table above).
- A non-empty `reason` is required (BR-6).
- **No role is exempt from this, including Compliance Officer and Admin.** A DPO investigating her own organization's compliance posture still has to state why she's viewing an unmasked field, and that access is still logged. Seniority or job function is not a substitute for the audit trail — the entire wedge collapses if any role gets a quiet bypass.

---

## Separation of Duties: Platform Operator Gets Infrastructure, Not Content

Per the RBAC table, Platform Operator (Priya's persona) can deploy, configure infrastructure-level retention defaults, and operate the system — but has no default path to view session content, the audit trail, or trigger exports. This is a deliberate separation-of-duties boundary, not an oversight: the person who can restart the database and the person who can view a patient's masked session should not, by default, be the same permission grant, even if in a small organization (per [target-users.md](../00-overview/target-users.md)'s org-size variants) they're literally the same person wearing two hats. The two roles being held by one human doesn't collapse the two *permissions* into one — that person authenticates and acts under whichever role context the action requires, and both are logged under their identity either way.

---

## What's Deliberately Not Modeled Here

- SSO group-to-role mapping mechanics — [authentication.md](authentication.md).
- The actual policy-engine implementation (a permissions table, OPA/Rego, or an in-process check) — downstream of this design, not part of it.
- UI-level permission affordances (what a Compliance Officer sees vs. an Engineer in the dashboard) — a product/design concern, not an authorization-model one; the enforcement point is `query-api`/`workers`, per [bounded-contexts.md](../02-domain/bounded-contexts.md), regardless of what the UI shows.

---

Sources: [Implementing the HIPAA Minimum Necessary Standard — AccountableHQ](https://www.accountablehq.com/post/implementing-the-hipaa-minimum-necessary-standard-best-practices-and-policy-examples), [HIPAA Access Control Requirements Explained — Censinet](https://censinet.com/perspectives/hipaa-access-control-requirements-explained), [RBAC vs ABAC: Access Control for Sensitive Data — Skyflow](https://www.skyflow.com/post/rbac-vs-abac-access-control-for-sensitive-data), [RBAC vs. ABAC — Splunk](https://www.splunk.com/en_us/blog/learn/rbac-vs-abac.html).

## What This Feeds Next

`docs/08-security/authentication.md` should specify how an actor's identity and role are established in the first place (local accounts vs. SSO, per [pricing.md](../01-product/pricing.md)'s tier split). `docs/08-security/encryption.md` should specify the key-rotation mechanics [pii-redaction.md](pii-redaction.md) deferred.
