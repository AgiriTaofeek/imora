# Alerting

> Status: How `AlertTriggered` (per [event-catalog.md](../02-domain/event-catalog.md)) actually reaches a human, and how the routing decisions here are what determine whether Chidi tunes the product out — the concrete stakes named in [user-personas.md](../01-product/user-personas.md).

---

## The Stakes, Restated Concretely

A 2025 Catchpoint study found 62% of on-call engineers have ignored a critical alert because it was buried in noise, per [target-users.md](../00-overview/target-users.md) and [user-stories.md](../01-product/user-stories.md) story C2. Every design choice below exists to keep Imora on the right side of that statistic, not the wrong one.

## The Workflow

1. **One alert per root cause is a data-model fact, not a UI filter.** Per [domain-model.md](../02-domain/domain-model.md), `ErrorGroup` assignment happens at write time in `alert-engine` — by the time an `AlertTriggered` event exists, deduplication has already happened. There is no "smart grouping" layer trying to suppress noise after the fact; there's nothing to suppress because the noise was never generated.
2. **Delivery via [webhooks.md](../07-api/webhooks.md), routed to wherever the team already works** — Slack, email, or a webhook into existing on-call tooling (PagerDuty-style), per `notification-service`'s Conformist relationship to `alert-engine` in [bounded-contexts.md](../02-domain/bounded-contexts.md). Imora doesn't become a new place to check; it feeds the places already being checked.
3. **Severity is real, not decorative.** [webhooks.md](../07-api/webhooks.md)'s `AlertTriggered` payload carries a severity field derived from the underlying signal — a Core Web Vitals regression (story C1) and an active security correlation (story D2) are not routed identically, since treating every alert as equally urgent is exactly the pattern that produces the ignored-alert statistic above.

## Alert Volume as a Tracked Product Metric

Per story C2's acceptance criteria: alert volume per real incident is itself something the product tracks and surfaces, not just an outcome hoped for. If grouping quality regresses — more alerts firing per actual root cause than before — that's visible as a metric before Chidi starts tuning the tool out, rather than discovered only when adoption quietly drops.

## What This Feeds Next

[session-replay.md](session-replay.md) and [security-monitoring.md](security-monitoring.md) cover the two workflows an alert from this document most often leads into.
