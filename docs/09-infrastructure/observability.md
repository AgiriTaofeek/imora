# Observability

> Status: The last file in `09-infrastructure/` — how Imora monitors its own infrastructure. Scoped deliberately: this is monitoring of Imora's *own* services (gateway, ingestion, query-api, alert-engine, workers, notification-service, dashboard), distinct from the product features Imora sells to customers for monitoring *their* applications — the same scope boundary [audit-logging.md](../08-security/audit-logging.md) drew for operational versus product logging.

---

## Why Not Dogfood Imora on Itself

Using Imora's own product to monitor Imora's own infrastructure is tempting — and creates a circular blast-radius problem: if the thing that's broken is Imora's own observability pipeline, self-monitoring with itself means the tool you'd reach for to diagnose the outage is the thing that's down. **Infrastructure self-monitoring runs on a separate, standard stack** (Prometheus-style metrics collection, structured logs, a simple alerting path independent of `notification-service`) — simpler than the product, and specifically not dependent on any of the eight bounded contexts being healthy to report that they aren't.

---

## Closing Three Gaps This Doc Set Left Open

This is the third time the same shape of problem has come up: a mechanism gets specified as existing, but nothing was ever assigned to actually watch it. [threat-model.md](../08-security/threat-model.md) flagged two of these directly; a third was sitting unaddressed in [scaling.md](../04-architecture/scaling.md). All three get an owner here, in infrastructure monitoring specifically, because none of the three are product features Imora exposes to customers — they're operational signals for whoever runs the deployment:

1. **Sequence-number gap detection** ([threat-model.md](../08-security/threat-model.md), Repudiation) — the periodic `workers` integrity-check job's output is surfaced as an infrastructure alert, not just a `notification-service` message that depends on that same job's health to deliver.
2. **UNMASK-frequency review** ([threat-model.md](../08-security/threat-model.md), Elevation of Privilege) — the per-actor unmask-frequency report becomes a standing infrastructure dashboard, visible to Platform Operator role continuously, not just a periodic push.
3. **Scaling threshold monitoring** — [scaling.md](../04-architecture/scaling.md) specified that a cluster migration should be planned once accumulated storage reaches roughly 50% of the single-machine SSD allocation, but never specified anything that watches for it. **This document closes that gap:** infrastructure monitoring tracks accumulated ClickHouse storage against that 50% threshold directly and alerts Priya's role when it's approached — turning [scaling.md](../04-architecture/scaling.md)'s calculation into an operational trigger instead of a number a human has to remember to check.

---

## What Gets Monitored

- **Service health** — uptime, error rate, latency per service, standard practice for the eight bounded contexts.
- **CronJob execution** — specifically whether `RetentionSweepScheduler` (per [kubernetes.md](kubernetes.md)'s `concurrencyPolicy: Forbid` design) is actually completing on schedule, or silently getting skipped run after run because a prior run never finishes — a Forbid policy that's always skipping is a real operational problem masquerading as a working safeguard.
- **Data store resource utilization** — ClickHouse, PostgreSQL, MinIO, Redis, per the sizing in [deployment-model.md](../04-architecture/deployment-model.md), feeding the scaling-threshold alert above.
- **The three gap-closing signals** above.

---

## What's Deliberately Not Modeled Here

- Specific tool selection (Prometheus/Grafana vs. an alternative stack) — implementation choice, not architecture.
- Alert routing/on-call configuration — an operational runbook concern, downstream of this design.
- Log retention for infrastructure logs themselves — distinct from and much shorter than [retention.md](../06-data/retention.md)'s product-data retention clocks, since infrastructure logs carry no regulatory retention obligation.

---

## What This Closes Out

This is the last file in `docs/09-infrastructure/`. All five files — [compose.md](compose.md), [docker.md](docker.md), [kubernetes.md](kubernetes.md), [ci-cd.md](ci-cd.md), and this one — are now internally consistent. `docs/10-engineering/` is next — team conventions and the ADR pattern already scaffolded there, which should retroactively document several of the load-bearing decisions made across this entire doc set as formal ADRs (AGPLv3 licensing, the dual ClickHouse/Postgres store split, key-pair over keyless signing, among others).
