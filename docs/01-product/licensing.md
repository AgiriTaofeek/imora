# Licensing

> Status: Research-based, current as of July 2026. Flagged as a prerequisite in [prd.md](prd.md)'s Open Questions section — the license determines whether "self-hosted" actually delivers on the promise in [vision.md](../00-overview/vision.md), so it has to be decided before [feature-roadmap.md](feature-roadmap.md) sequences anything.

---

## The Decision This Document Has to Get Right

Per [vision.md](../00-overview/vision.md)'s Positioning section, Imora's entire pitch is: parity with Sentry/Datadog/LogRocket/FullStory, plus wedge capabilities (access-audit-trail, regulatory retention, evidence export) that no alternative — SaaS or self-hosted — currently ships. The license has to protect that pitch, not quietly undermine it. Two ways licensing commonly undermines exactly this kind of pitch, found directly in the products already researched for [competitive-analysis.md](../00-overview/competitive-analysis.md):

1. **Feature-gating the wedge behind a commercial tier.** PostHog's core is MIT-licensed and fully functional self-hosted — but SSO enforcement, RBAC, advanced permissions, and **audit logs specifically** live in its `ee/` directory under a separate PostHog Enterprise License requiring a commercial agreement for production use. If Imora did this to the audit-trail wedge, a regulated org would face the same problem self-hosting was supposed to solve: needing an ongoing commercial relationship just to get the compliance capability that is the entire reason they picked Imora over OpenReplay. This is the single most important thing this document has to rule out.
2. **License-restricted self-hosting that isn't actually equivalent to the SaaS product.** Sentry's Functional Source License states there's no intended feature difference between SaaS and self-hosted — but in practice, self-hosted Sentry still doesn't reliably support Session Replay (documented in [competitive-analysis.md](../00-overview/competitive-analysis.md)'s "Where Sentry specifically stands" section). The lesson: a license clause promising parity doesn't guarantee engineering delivers it. This document controls the legal terms; [feature-roadmap.md](feature-roadmap.md) has to actually ship parity for the promise to mean anything.

---

## What Comparable Products Actually Do

| Product | License | Consequence |
|---|---|---|
| **Sentry** | Functional Source License (FSL) — source-available, not OSI-approved. Prohibits reselling self-hosted Sentry as a competing offering or using the code to build a direct competitor. Converts to Apache 2.0 after a 2-year grace period. | No feature-gating by license text, but self-hosted deployments lag SaaS in practice (session replay). Source-available licenses also carry community-trust risk — see HashiCorp below. |
| **PostHog** | MIT core + separate "PostHog Enterprise License" (source-available, commercial-agreement-required) for `ee/` features. | Fully functional OSS core, but SSO, RBAC, and **audit logs** require a paid agreement — exactly the anti-pattern this document exists to avoid replicating. |
| **OpenReplay** | AGPLv3 core + a separate non-open enterprise license for some parts. | Same open-core shape as PostHog, one license family more protective against re-hosting (see AGPL below). |
| **HashiCorp Terraform** | Switched from MPL (permissive) to BSL after building years of community trust under an open license. | Triggered a community fork (OpenTofu). The lesson for Imora: decide the license model now, before a community forms around different expectations — relicensing later is the highest-trust-cost path. |

Source: [Sentry Licensing](https://open.sentry.io/licensing/), [Introducing the Functional Source License — Sentry Blog](https://blog.sentry.io/introducing-the-functional-source-license-freedom-without-free-riding/), [PostHog Enterprise License — GitHub](https://github.com/PostHog/posthog/blob/master/ee/LICENSE), [PostHog Open Source docs](https://posthog.com/docs/posthog-code/open-source), [Open Source Licenses Explained: AGPL, MIT, GPL, Apache 2.0](https://www.opensourcealternatives.to/blog/open-source-license-guide).

---

## The Recommendation

### Core product — including the full wedge — under AGPLv3, not MIT and not a source-available/BSL-family license

- **AGPLv3, not MIT/Apache:** a permissive license lets any cloud provider — including a Datadog or a well-funded competitor — take Imora's code, host it, and sell it as a competing SaaS product without contributing anything back. Since Imora explicitly competes with SaaS incumbents (per [vision.md](../00-overview/vision.md)'s Positioning), this isn't a hypothetical: it's the exact business risk a permissive license creates for a product whose whole model is "we compete with the SaaS players." AGPL's network-use clause closes this — anyone running a modified version as a network service must release their modifications under the same terms. This is the same reasoning Nextcloud and Mastodon apply to their own self-hosted-competitor-to-SaaS products.
- **Not a source-available/BSL-family license (Sentry's FSL, HashiCorp's old BSL):** these exist specifically to prevent competitors from reselling the product, which protects revenue but isn't OSI-approved open source — and per [target-users.md](../00-overview/target-users.md), procurement teams at regulated orgs actively screen license type as an evaluation criterion. A non-OSI license is a harder sell to exactly the buyers (Dara, Adaeze) this product depends on, for a benefit (anti-resale protection) that AGPL already provides via the network-use clause.
- **No `ee/`-style split that puts the wedge behind a commercial license.** The access-audit-trail, retention-clock policy, and evidence export — the entire reason a regulated org picks Imora per [prd.md](prd.md)'s MVP scope — ship under the same AGPLv3 terms as everything else, self-hosted, no commercial agreement required. This is the direct fix for the PostHog anti-pattern identified above, and it's a non-negotiable constraint on [feature-roadmap.md](feature-roadmap.md): no milestone may move a wedge capability into a paid tier.

### What is legitimately fine to monetize

Consistent with the principle above — monetize things that are not the reason a regulated buyer chose Imora over the alternatives:

- **Managed hosting** for organizations that want Imora's wedge capabilities without operating the infrastructure themselves — this doesn't touch the self-hosting promise because it's optional, not the only way to get compliance features.
- **Premium support / SLA-backed response times** — Priya's persona ([user-personas.md](user-personas.md)) explicitly worries about operating this without growing headcount; paid support is a legitimate answer to that, not a gate on functionality.
- **SSO/SAML enterprise auth integrations** — unlike audit logs, enterprise SSO is genuinely a nice-to-have convenience feature, not one of the three wedge capabilities identified in [competitive-analysis.md](../00-overview/competitive-analysis.md). Gating this is a defensible parallel to PostHog's model, since it doesn't touch the compliance promise.
- **Multi-region/HA orchestration tooling and professional services** for air-gapped or complex deployment topologies — operational complexity, not core product capability.

---

## What This Rules Out for feature-roadmap.md

- No milestone may ship the audit-trail, retention-policy, or evidence-export wedge features as "Enterprise" or license-gated. If a future roadmap draft does this, it contradicts this document and needs to come back here first.
- The license decision (AGPLv3) should be adopted before the first public commit, per the HashiCorp/Terraform lesson above — changing it later, after a community has formed under different expectations, is the costliest version of this decision to get wrong.
