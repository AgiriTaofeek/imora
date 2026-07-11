# Performance Monitoring

> Status: The C1-driven workflow — what Chidi sees when a Core Web Vitals metric moves but nothing threw an error, distinct from [error-investigation.md](error-investigation.md)'s exception-triggered path.

---

## The Scenario

Per [user-personas.md](../01-product/user-personas.md)'s Chidi scenario: a release shipped last night, and Chidi wants to know whether LCP, INP, or CLS moved for a given page or flow — with no error, no page, nothing broken in the conventional sense.

## The Workflow

1. **The threshold is the industry bar, not an arbitrary internal number.** LCP < 2.5s, INP < 200ms, CLS < 0.1, evaluated at the 75th percentile of real sessions — Google's own methodology, per [target-users.md](../00-overview/target-users.md), not a metric Imora invented. A regression is measured against that bar, so "is this actually bad" isn't a judgment call Chidi has to make from scratch.
2. **p75, not average.** A regression affecting the slowest quarter of real users can hide entirely inside a mean that's dominated by fast connections — [event-schema.md](../06-data/event-schema.md)'s `PerformanceMetricRecorded` event and [clickhouse-schema.md](../06-data/clickhouse-schema.md)'s query shape are built around percentile evaluation specifically so this doesn't happen silently.
3. **Release attribution is automatic**, per story C1 and [event-catalog.md](../02-domain/event-catalog.md)'s `RegressionDetected` event — a statistically significant change (per the trend-detection approach in [target-users.md](../00-overview/target-users.md), matching Sentry's own regression-detection pattern) tied to the specific release that introduced it. Chidi doesn't bisect deploys by hand to find out which one moved the metric.
4. **Drill into the session level.** From the flagged regression, open representative affected sessions — the same replay capability [error-investigation.md](error-investigation.md) uses, applied here to "why is this page slow" instead of "why did this throw."

## Why This Workflow Existing at All Is the Point

Per [target-users.md](../00-overview/target-users.md)'s framing of Persona 6: this is a workflow with no compliance angle whatsoever, deliberately. If Chidi only ever opens the product during an incident or an audit, the product has already failed its own adoption thesis, per [prd.md](../01-product/prd.md)'s Goals. This workflow — checking on a Wednesday whether a routine deploy regressed anything — is what daily use actually looks like.

## What This Feeds Next

[alerting.md](alerting.md) covers how a regression like the one described here actually reaches Chidi or Jon in the first place, rather than requiring someone to go looking.
