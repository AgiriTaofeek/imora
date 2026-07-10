# Pricing

> Status: Research-based, current as of July 2026. Builds on [licensing.md](licensing.md)'s monetization surface (managed hosting, support SLA, SSO/SAML, multi-region tooling) and constraint (no wedge capability may ever be paywalled). This document decides *how* those specific things get priced.

---

## The Constraint This Model Has to Satisfy

Per [licensing.md](licensing.md), the core product — full parity plus the entire compliance wedge (access-audit-trail, retention-clock policy, evidence export) — ships AGPLv3, self-hosted, free, no commercial agreement required. Pricing only applies to the Milestone 3 surface from [feature-roadmap.md](feature-roadmap.md): managed hosting, premium support/SLA, SSO/SAML, and multi-region/HA tooling. Nothing below may fund itself by gating anything on that list.

---

## Why the Category-Standard Model Doesn't Fit

Both direct comparators price on metered usage: **PostHog** charges per event/session/replay with step-down volume tiers (roughly $0.00005/event at 1–2M events/month, decreasing to $0.000009/event above 250M). **Sentry** charges per error/trace/replay processed, with tiered per-unit rates once a plan's included volume is exceeded. This is the pricing model most engineers evaluating Imora will already expect.

It doesn't work for Imora's actual buyers, for two independent reasons found in research, not just a stylistic preference:

1. **Metering requires reporting usage back to the vendor.** PostHog and Sentry can meter accurately because they host the product themselves (SaaS) — the events never leave their infrastructure to begin with. A self-hosted, air-gapped Imora deployment has no outbound path to report usage at all, by design, per the Operational Simplicity principle in [vision.md](../00-overview/vision.md). Metered pricing would either require breaking the air-gap (unacceptable to Dara and Adaeze) or estimating usage on trust, which isn't really metered pricing anymore.
2. **Regulated procurement structurally resists metered billing regardless.** Enterprise agreements are overwhelmingly flat-fee, because large buyers' budget processes require a committed number and procurement won't sign a contract where the invoice changes quarter to quarter based on consumption they can't fully control. Where a usage component exists at enterprise scale, it gets negotiated into a fixed annual commitment or a capped ceiling — not left as a fluctuating metered bill. This isn't specific to air-gapped buyers; it's how Dara's and Adaeze's procurement processes work regardless of deployment model.

Copying PostHog or Sentry's pricing model here would repeat the same mistake flagged in [licensing.md](licensing.md) about copying PostHog's feature-gating — importing a pattern that fits the comparator's business model but actively works against Imora's actual buyers.

---

## The Recommendation

### Self-hosted core — always free, no license key at all

Parity and wedge capabilities require no licensing mechanism whatsoever — not a free tier with a key, an actual absence of gating. This is the only way to make the "no compensating control bolted on afterward" claim from [prd.md](prd.md)'s Goals literally true.

### Milestone 3 commercial add-ons — flat annual pricing, tiered by seat/deployment band, not metered

Per the research above, and consistent with the finding that per-seat models are a natural fit for regulated industries because they already need auditable user counts for compliance reasons anyway:

| Tier | Who it's for (per [target-users.md](../00-overview/target-users.md) org-size variants) | What's priced |
|---|---|---|
| **Community** | Under ~50 employees — the CTO-wears-every-hat org | $0. Full AGPLv3 core, community support only. |
| **Team** | ~50–300 employees | Flat annual fee, tiered by a self-declared seat-count band agreed at signing — not metered. Adds premium support/SLA. |
| **Enterprise** | 300+ employees, per-persona split intact (CISO, DPO, HIPAA Security Officer, Platform lead as separate buyers) | Flat annual contract, negotiated per deployment scale. Adds SSO/SAML, multi-region/HA tooling, dedicated support, and offline license activation for air-gapped environments. |

Seat/deployment bands are declared at contract signing and renewal — an auditable, contractual fact regulated buyers already produce for their own compliance programs — not measured by telemetry Imora's own architecture doesn't collect.

### SSO/multi-region gating — offline signed license files, not phone-home activation

For the one thing in Milestone 3 that does need a technical gate (SSO/SAML, multi-region tooling), the standard pattern for government, military, healthcare, and financial air-gapped deployments is a cryptographically signed offline license file (Ed25519 or RSA-2048), generated in a connected environment and transferred into the air-gapped network manually (USB, or a QR-code-based exchange), rather than a server that periodically phones home to validate. This works identically for a connected Team-tier customer and a fully air-gapped Enterprise-tier one — the same mechanism serves both, so there's no separate "air-gapped SKU" to maintain.

### Managed hosting — the one place usage-based pricing is legitimate

If Imora offers managed hosting as part of Milestone 3, metered/usage-based pricing is fine there, for the same reason it works for PostHog Cloud and Sentry Cloud: Imora would be running the infrastructure directly, so metering doesn't require anything to phone home across an air-gap — there is no air-gap in that deployment model. This is the one SKU where copying the category-standard pricing pattern is actually correct, precisely because it's the one SKU that isn't self-hosted.

---

## Open Questions and Risks

- **Self-declared seat bands are a trust model, not a metered one.** At Enterprise scale this is standard (regulated buyers already self-attest headcount for their own compliance programs), but it means Imora's revenue at that tier depends on contract terms and audit rights, not technical enforcement — a deliberate tradeoff for the air-gap constraint, not an oversight.
- **Community tier at $0 with no seat ceiling could be used by orgs well past 50 employees who just don't sign a contract.** Per [licensing.md](licensing.md), this is a permitted consequence of choosing AGPLv3 over a restrictive license — the tradeoff for procurement-friendly, non-OSI-avoidance licensing is that Team/Enterprise tiers have to sell on support and features the Community tier lacks (SLA, SSO, multi-region), not on artificial scarcity of the core product.
- **Whether a "Team" tier is worth maintaining at all**, versus going straight from free Community to negotiated Enterprise, is a real open question that depends on actual demand data this document can't produce — flagged here rather than guessed at.

Sources: [PostHog pricing](https://posthog.com/pricing), [Sentry pricing](https://sentry.io/pricing/), [Air-Gapped License Activation — LicenseSpring](https://docs.licensespring.com/license-entitlements/activation-types/air-gapped), [How to Implement an Offline Licensing Model — Keygen](https://keygen.sh/docs/choosing-a-licensing-model/offline-licenses/), [Enterprise SaaS Pricing: List Price vs Negotiated Deals](https://softwarepricing.com/blog/enterprise-saas-pricing/), [Enterprise SaaS Pricing Models Compared — m3ter](https://www.m3ter.com/blog/enterprise-saas-pricing-models-enterprise-pricing-strategy).
