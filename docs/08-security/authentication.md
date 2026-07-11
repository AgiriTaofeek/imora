# Authentication

> Status: Research-based, current as of July 2026. Specifies how an actor's identity is established before [authorization.md](authorization.md)'s role/permission model can apply — the mechanism `gateway` uses to populate the RequestContext every downstream AccessAuditEvent depends on, per [component-diagrams.md](../04-architecture/component-diagrams.md).

---

## Local Authentication — the Required Baseline

Per [pricing.md](../01-product/pricing.md), Community and Team tier run without SSO, and per [system-context.md](../04-architecture/system-context.md), no Parity or Wedge capability may depend on an external system — so local username/password authentication is the baseline every deployment has, including Enterprise ones, not a fallback for smaller tiers only.

- **Password hashing: Argon2id**, the current OWASP-recommended default (parameters m=19456, t=2, p=1), not bcrypt — bcrypt remains acceptable only for legacy systems already using it at cost≥12, not a reasonable choice for a new system in 2026.
- **MFA: TOTP, not SMS.** Time-based one-time passwords operate entirely from a shared secret and a clock — no external SMS gateway, no dependency that would break in an air-gapped deployment. SMS-based OTP would silently fail the air-gapped requirement from [system-context.md](../04-architecture/system-context.md); TOTP doesn't, so it's the only MFA mechanism that belongs in the required path.

---

## SSO — the Enterprise-Tier Layer, Including for Air-Gapped Deployments

A common misconception worth ruling out explicitly: **air-gapped does not mean "no SSO."** It means no path to the public internet — it does not mean no internal network services. The standard, well-established pattern for exactly this situation is deploying the Identity Provider itself *inside* the same isolated network as the application: Keycloak, Auth0 Private Cloud, and ForgeRock all support fully on-premises deployment, with every part of the authentication flow — login screens, token issuance, session verification — happening entirely within the air-gapped boundary. A large air-gapped bank (per [target-users.md](../00-overview/target-users.md)'s 300+-employee band) plausibly already runs an internal IdP (self-hosted Keycloak, or Active Directory Federation Services) for its other internal systems — Imora's SSO integration talks to *that*, never to a cloud identity provider.

- **Protocol:** SAML 2.0 and OIDC, the two protocols every self-hostable IdP (Keycloak, Auth0 Private Cloud, ForgeRock) supports.
- **Role mapping:** SSO group/claim assertions map to the four roles from [authorization.md](authorization.md) (engineer, compliance_officer, platform_operator, admin) at login time — the mapping configuration itself is an Admin-role action, closing the loop with [authorization.md](authorization.md)'s permission table.
- **Gating mechanism:** per [pricing.md](../01-product/pricing.md), SSO is gated by the offline signed license file, not a phone-home check — consistent with every other Enterprise-tier feature in this doc set.

---

## Session and Token Handling

`gateway` issues a session token on successful authentication (local or SSO) and populates the RequestContext every downstream component consumes, per [component-diagrams.md](../04-architecture/component-diagrams.md) — this is the identity that ends up in every `actorUserId` field across [event-schema.md](../06-data/event-schema.md)'s AccessAuditEvent variants. A request with no valid session cannot reach `query-api`'s `AuditedQueryHandler` or `workers`' actions at all — authentication failure is a hard stop before authorization is even evaluated, not a soft-fail that falls through to a default permission level.

---

## What's Deliberately Not Modeled Here

- Exact token format (JWT vs. opaque reference token) and session lifetime values — implementation detail downstream of this design.
- Keycloak/IdP deployment configuration specifics — `09-infrastructure/`.
- Password reset and account-recovery flows — a product/UX concern once this mechanism is accepted.

---

Sources: [Air-Gapped Deployment Single Sign-On — hoop.dev](https://hoop.dev/blog/air-gapped-deployment-single-sign-on/), [OpenID Connect in Air-Gapped Environments — hoop.dev](https://hoop.dev/blog/openid-connect-in-air-gapped-environments/), [Keycloak SSO Integration — Glasswall (offline/on-premises deployment guide)](https://docs.glasswall.com/docs/keycloak-sso-integration), [Password Hashing in 2026: bcrypt vs Argon2 vs scrypt vs PBKDF2 — Security Boulevard](https://securityboulevard.com/2026/06/bcrypt-vs-argon2-vs-scrypt-vs-pbkdf2-a-2026-decision-framework/).

## What This Feeds Next

`docs/08-security/encryption.md` is the last piece [pii-redaction.md](pii-redaction.md) and [authorization.md](authorization.md) both deferred — key rotation and the disk-level baseline. `docs/08-security/audit-logging.md` and `docs/08-security/threat-model.md` round out this folder.
