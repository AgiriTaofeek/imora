# Scaling

> Status: Research-based, current as of July 2026. Answers the question [deployment-model.md](deployment-model.md) left open: the concrete threshold where the single-machine profile stops being viable and cluster migration is warranted.

---

## The Finding: Imora's Scaling Trigger Is Retention-Driven Storage, Not Throughput

For a typical observability product, the scaling story is about ingestion throughput or query concurrency. For Imora specifically, the math below shows that's the wrong thing to watch — **[business-rules.md](../02-domain/business-rules.md) BR-1's multi-year regulatory retention clocks make accumulated storage the binding constraint, usually well before ClickHouse's write throughput or query concurrency become a real concern.** This is a direct, non-obvious consequence of combining three numbers already established elsewhere in this doc set with two pieces of external research.

### The math (assumptions stated, so the reader can substitute their own numbers)

- **Session size:** rrweb-based session replay compresses to roughly 200KB average per session (a 30-minute session runs 1–5MB gzipped; a 5-minute session runs 100–500KB) — this is a property of the capture format from [domain-model.md](../02-domain/domain-model.md), not an Imora-specific choice.
- **Single-machine storage ceiling:** a reasonable single-machine SSD allocation per [deployment-model.md](deployment-model.md)'s profile is roughly 2–4TB — comfortably available on modest self-hosted hardware without moving to specialized storage.
- **Retention multiplier:** per [business-rules.md](../02-domain/business-rules.md) BR-1, healthcare data retained under HIPAA's 6-year floor doesn't get deleted at steady-state volume — it accumulates for the full window before the oldest data ages out.

At **~100,000 sessions/month** (a mid-size regulated org's realistic digital-channel traffic): 100,000 × 200KB ≈ 20GB/month of SessionEvent data alone, ×12 ≈ 240GB/year, **× 6 years of HIPAA retention ≈ 1.4TB** just for session replay — before ErrorEvent, PerformanceMetric, SecurityEvent, and AccessAuditEvent (which, per [event-catalog.md](../02-domain/event-catalog.md), fires on every single read, not just every session) are added on top. A conservative 2–3× multiplier for those combined categories puts total accumulated storage in the **2.8–4.2TB range** — at or past a comfortable single-machine ceiling, for a traffic level well within reach of Dara's or Marcus's organizations per [user-personas.md](../01-product/user-personas.md).

**Compare this to ClickHouse's actual throughput ceiling:** even modest hardware sustains ingestion rates in the hundreds of thousands to low millions of rows per second; a single well-funded ClickHouse Cloud node has demonstrated ~4 million rows/second. No realistic session volume from a single regulated organization comes close to saturating that — 100,000 sessions/month is a rounding error against millions of rows/second. **Throughput was never going to be the trigger; retention math is.**

### Query concurrency — a secondary, rarely-binding consideration

The number of simultaneous investigators is bounded by the org-size bands in [target-users.md](../00-overview/target-users.md) — an 8–10 person on-call rotation (Priya's team) plus a handful of compliance staff (Adaeze, Marcus) is a low double-digit concurrent-query ceiling at most, well within what a single ClickHouse node sized for user-facing analytics can serve. This only becomes a real constraint at the 300+-employee band, where it typically arrives *after* the storage threshold above, not before.

---

## Migration Signal and Threshold

**Plan a cluster migration when accumulated storage (current volume × the applicable regulatory retention multiplier for the strictest category in use, per BR-1) is projected to exceed roughly 50% of the single-machine SSD allocation** — not 100%, because migration itself takes lead time, and running a compliance-critical system to the edge of disk capacity is its own operational risk per [deployment-model.md](deployment-model.md)'s backup/RPO requirement. Using the worked example above, that's roughly the **50,000–70,000 sessions/month** mark for an organization under a 6-year retention floor — organizations under GDPR's shorter purpose-bound retention (per BR-1's per-category clocks) have materially more headroom before the same trigger fires, since their accumulated multiplier is smaller.

---

## What Migration Actually Changes

Per [container-diagrams.md](container-diagrams.md)'s closing claim: **the domain model, business rules, and event catalog do not change between profiles.** Migrating from single-machine to cluster means ClickHouse and PostgreSQL move to multi-node, a message queue is introduced between `ingestion` and its consumers, and `query-api`/`ingestion` scale independently — it does not mean re-deriving retention policy, re-validating BR-2's hold-check ordering, or re-specifying AccessAuditEvent's shape. A migration that touched any of those would indicate the single-machine and cluster profiles had silently drifted into two different products, which [deployment-model.md](deployment-model.md) and this document both exist to prevent.

---

## What's Deliberately Not Modeled Here

- Step-by-step migration runbook (how to actually move a running ClickHouse single-node to a sharded cluster without downtime) — an operational procedure for `09-infrastructure/`, not an architecture decision.
- Auto-scaling policies or specific Kubernetes HPA configuration — `09-infrastructure/kubernetes.md`.
- Cost modeling for cluster-scale infrastructure — out of scope for this doc set entirely; a customer-specific deployment decision.

---

Sources: [How to ingest 1 billion rows per second in ClickHouse — Tinybird](https://www.tinybird.co/blog/1b-rows-per-second-clickhouse), [ClickHouse concurrency: how to size for user-facing analytics](https://clickhouse.com/resources/engineering/high-concurrency-sizing-user-analytics), [Sizing and hardware recommendations — ClickHouse Docs](https://clickhouse.com/docs/guides/sizing-and-hardware-recommendations), [Exploring rrweb: A Session Replay Walkthrough and Best Practices](https://medium.com/@idogolan15/exploring-rrweb-a-session-replay-walkthrough-and-best-practices-47a52f0e2447), [Session Replay: What It Is, How It Works, and When You Need It](https://temps.sh/blog/session-replay-how-it-works).

## What This Feeds Next

`docs/04-architecture/repository-structure.md` and `overview.md` are the remaining files in this folder — `overview.md` in particular should synthesize everything from [system-context.md](system-context.md) through this document into the single narrative entry point the rest of `04-architecture/` currently lacks.
