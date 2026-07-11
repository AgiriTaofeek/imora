# Glossary

> Status: The single canonical glossary for the whole project — consolidates terms precisely defined across `00-overview/`, `01-product/`, and `02-domain/` into one reference every stakeholder can use, not just readers of any one folder. Every definition here traces back to a specific prior document — this file doesn't introduce new meaning, it collects it.

---

## Positioning Terms

**Parity** — a capability necessary to be a credible alternative to Sentry, Datadog RUM, LogRocket, FullStory, OpenReplay, or PostHog at all. Defined in [vision.md](vision.md)'s Positioning section; the full checklist is in [competitive-analysis.md](competitive-analysis.md).

**Wedge** — a capability none of those alternatives, self-hosted or SaaS, currently ship. The three uncontested wedge gaps (access-audit-trail, regulatory-clock retention, evidence export) are ranked in [competitive-analysis.md](competitive-analysis.md)'s Synthesis.

**Alternative** — Imora's own positioning, per [vision.md](vision.md): not a new category, a swap-in replacement for tools regulated teams already use, plus the wedge.

---

## Personas (shorthand used from [user-personas.md](../01-product/user-personas.md) onward)

| Name | Role | Carries |
|---|---|---|
| Dara | CISO, regional bank | Breach cost + CIPA litigation exposure |
| Adaeze | DPO, national insurer | DSAR/breach-notification deadlines, GDPR storage limitation |
| Marcus | HIPAA Security Officer, hospital network | Annual risk assessment, ePHI audit controls |
| Priya | Head of Platform Engineering, fintech | Deployment/operational burden |
| Jon | Incident Commander / SRE, Priya's team | Fragmentation tax, chain-of-custody |
| Chidi | Senior Frontend Engineer, Priya's team | Daily adoption — the parity check, not a cost driver |

Full definitions in [target-users.md](target-users.md) (roles) and [user-personas.md](../01-product/user-personas.md) (grounded scenarios).

---

## Domain Entities (full definitions in [domain-model.md](../02-domain/domain-model.md))

- **Session** — one user's browsing session; the aggregate root for replay data.
- **SessionEvent** — one entry in a Session's rrweb-style capture stream (FullSnapshot, DOMMutation, MouseMove, Click, Scroll, FormInput, ViewportChange).
- **ErrorEvent / ErrorGroup** — a captured exception, and the deduplicated root-cause it's grouped under at write time.
- **Release** — a deployed version identifier used for regression attribution.
- **PerformanceMetric** — an LCP, INP, or CLS measurement tied to a Session and Release.
- **TraceLink** — the shared session/trace identifier correlating a Session to backend spans.
- **AccessAuditEvent** — the append-only log entry produced for every VIEW, EXPORT, UNMASK, DELETE, DELETION_SKIPPED, or CONFIG_CHANGED against sensitive data or its governing configuration. See [event-catalog.md](../02-domain/event-catalog.md) for the full set of named variants (SessionViewed, FieldUnmasked, RecordExported, RecordDeleted, DeletionSkippedDueToHold, ConfigurationChanged).
- **RetentionPolicy** — a per-data-category retention rule mapped to a regulatory clock (PCI-DSS 12mo, HIPAA 6yr, GDPR purpose-bound, SOX 7yr).
- **LegalHold** — a directive that overrides scheduled deletion for records matching a scope query, without creating a separate copy of the data.
- **EvidenceExport** — a frozen, self-contained, hash-verifiable copy of records generated for an incident, immune to later retention or erasure actions.
- **SecurityEvent** — a security signal optionally correlated into a Session's incident timeline.

---

## Business Rule Shorthand (full rules in [business-rules.md](../02-domain/business-rules.md))

- **BR-1 through BR-7** — referenced by number throughout later architecture and data docs. BR-1 (longest-retention-wins), BR-2 (check-before-destroy), BR-3 (GDPR erasure vs. legal obligation, selective purging), BR-4 (export immutability), BR-5 (audit on every sensitive access), BR-6 (unmask requires reason), BR-7 (capture-time masking).
- **Selective purging** — anonymizing or deleting only the fields a competing regulation doesn't require, rather than choosing between full retention and full deletion. The resolution mechanism for BR-3.
- **Deny-by-default capture** — masking any field with no explicit allow-list rule, so an unredacted new field fails closed. The mechanism behind BR-7.

---

## Regulatory Terms

- **DSAR** — Data Subject Access Request; a GDPR right to ask what data an organization holds and who has accessed it. One-month response deadline (extendable to three for complex requests).
- **DPO** — Data Protection Officer; mandatory under GDPR Article 37 for public authorities and large-scale data processors. Legally independent from the organizations that employ them.
- **BAA** — Business Associate Agreement; the HIPAA-required contract that makes a vendor's handling of PHI compliant — a contractual promise about a third party's environment, not organizational control, per [competitive-analysis.md](competitive-analysis.md).
- **CIPA** — California Invasion of Privacy Act; the wiretapping statute (§631, plus §638.51 pen-register claims) behind the session-replay litigation wave in [problem-statement.md](problem-statement.md).
- **PAM** — Privileged Access Management; the adjacent tooling category (BeyondTrust, Delinea) where the access-audit-trail pattern already exists, just not applied to frontend session data.

---

## Architecture Terms (full definitions in [bounded-contexts.md](../02-domain/bounded-contexts.md))

- **Bounded Context** — one of the eight owning service boundaries (gateway, ingestion, query-api, alert-engine, workers, browser-sdk, dashboard, notification-service).
- **Shared Kernel** — two contexts operating directly on the same entity definitions from [domain-model.md](../02-domain/domain-model.md) (browser-sdk↔ingestion, ingestion↔query-api).
- **Customer-Supplier** — an upstream context whose downstream customer can influence its priorities, but the dependency runs one direction (e.g., gateway → query-api).
- **Conformist** — a downstream context that accepts the upstream's model as-is, with no translation authority (alert-engine → notification-service, query-api → dashboard).

---

## Why This Lives in `00-overview/`, Not `02-domain/`

A glossary spanning positioning, personas, regulatory, and architecture terms is useful to every reader from day one — a DPO evaluating the product needs "DSAR" and "Wedge" defined long before they'd ever open `02-domain/`. Keeping it in the overview folder, rather than buried in the domain-modeling folder where only engineers would naturally look, is what makes it a single canonical reference instead of one more file competing with a near-duplicate elsewhere.
