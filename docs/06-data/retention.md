# Retention

> Status: Research-based, current as of July 2026. Assigns an actual retention period to each of the five data categories in `retention_policies` ([postgres-schema.md](postgres-schema.md)), feeding the `INTERVAL` values in [clickhouse-schema.md](clickhouse-schema.md)'s TTL clauses. Restating BR-1's regulatory table isn't enough — not every category has a regulatory floor, and treating them as if they all do would itself violate GDPR's storage-limitation principle from the opposite direction.

---

## The Split This Document Has to Make Explicit

[business-rules.md](../02-domain/business-rules.md) BR-1 covers what happens when regulations *conflict*. It doesn't establish that every data category is regulated in the first place. SessionEvent and AccessAuditEvent plausibly contain or reference PII/PHI and are squarely regulation-driven. But **ErrorEvent and PerformanceMetric have no comparable regulatory floor** — nobody is legally required to keep a stack trace for six years, and industry practice for diagnostic/error data runs 30–90 days by default, well short of any compliance clock. Applying a blanket multi-year retention to categories nothing requires it for isn't caution, it's the same storage-limitation violation GDPR Article 5(1)(e) prohibits in the other direction — keeping data longer than necessary for its actual purpose — and it directly worsens the accumulated-storage scaling problem identified in [scaling.md](../04-architecture/scaling.md).

---

## Category Assignments

| Category | Driver | Default | Configurable? |
|---|---|---|---|
| **SessionEvent** | Regulation | Set to the strictest regulation the deployment's own industry requires — HIPAA's 6-year floor for healthcare, PCI-DSS's 12-month floor for payment-touching flows, or GDPR's purpose-bound limit otherwise (see below) | Yes — this is the per-category configurability that is the entire point of the wedge, per [competitive-analysis.md](../00-overview/competitive-analysis.md) |
| **ErrorEvent** | Operational | 90 days by default, matching industry-standard diagnostic-log retention — nobody debugs a 6-year-old stack trace | Yes, but a shorter default than the regulated categories is the deliberate starting point, not an oversight |
| **PerformanceMetric** | Operational | 13 months by default — long enough for year-over-year Core Web Vitals comparison (the same period Datadog defaults to for general telemetry), short enough to avoid unnecessary accumulation | Yes |
| **SecurityEvent** | Regulation (usually) | Aligned to the same clock as SessionEvent when correlated to a session (story D2's incident-timeline requirement means a security signal is only as useful as the session context it's tied to); PCI-DSS's 12-month floor as a baseline when uncorrelated | Yes |
| **AccessAuditEvent** | Regulation, and never shorter than any other category | The **longest** clock among all categories the deployment operates under, per [clickhouse-schema.md](clickhouse-schema.md)'s finding — this table proves what happened to every other table, so it must outlive all of them | Not independently below the computed longest-clock floor — this one constraint isn't optional |

---

## Translating "GDPR Has No Fixed Term" Into an Actual TTL Value

GDPR's Article 5(1)(e) storage-limitation principle doesn't specify a duration — but [clickhouse-schema.md](clickhouse-schema.md)'s TTL mechanism needs a concrete `INTERVAL`, not an open-ended "as long as necessary." This isn't a contradiction: **GDPR doesn't prohibit setting a fixed enforced ceiling — it requires that ceiling be justified by actual processing necessity, not that no ceiling exist.** In practice, a deployment operating primarily under GDPR (no HIPAA/PCI-DSS/SOX floor applying) sets a concrete SessionEvent retention value — commonly 12–24 months for a customer-support/debugging purpose — and records the justification in `retention_policies.regulatory_basis`, per [postgres-schema.md](postgres-schema.md). The DPO persona (Adaeze, per [user-personas.md](../01-product/user-personas.md)) is the one who sets and can justify that number to a regulator; the system enforces whatever she configures, it doesn't invent the number for her.

---

## Why Getting the Operational Defaults Right Matters Beyond Compliance Hygiene

Per [scaling.md](../04-architecture/scaling.md)'s finding, accumulated retention is Imora's actual scaling trigger — not ingestion throughput. A deployment that defaults ErrorEvent and PerformanceMetric to the same 6-year HIPAA floor as SessionEvent, out of an abundance of caution, would inflate its accumulated-storage multiplier well past the 2–3× estimate that calculation assumed, pulling the single-machine-to-cluster migration threshold forward by years for no compliance benefit at all. The category split above isn't just correctness — it's what keeps [deployment-model.md](../04-architecture/deployment-model.md)'s single-machine promise to Priya's persona intact for as long as possible.

---

## What's Deliberately Not Modeled Here

- The actual configuration UI/workflow a deployment operator uses to set these values per organization — a product concern, not a data-architecture one.
- Tiered hot/warm/cold storage (moving aging data to cheaper storage before its TTL expires) — a valid optimization industry practice supports, but an implementation decision downstream of the retention *periods* this document sets, not a change to them.

---

Sources: [What is Log Retention? — LogicMonitor](https://www.logicmonitor.com/blog/what-is-log-retention), [Log Retention: Policies, Best Practices & Tools — Last9](https://last9.io/blog/log-retention/), [Log Retention Policies Explained — Groundcover](https://www.groundcover.com/learn/logging/log-retention-policies).

## What This Feeds Next

`docs/06-data/storage.md` is the last file in `06-data/` — it should specify the object-storage layout for EvidenceExport blobs, and note how tiered storage (mentioned above but not specified) would interact with the retention periods this document sets.
