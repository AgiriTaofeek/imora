# Error Investigation

> Status: What an engineer actually sees and does, using the mechanisms already specified in [event-catalog.md](../02-domain/event-catalog.md) and [sequence-diagrams.md](../04-architecture/sequence-diagrams.md) — this document is the narrated, human-facing counterpart to those, not a restatement of the component-level flow.

---

## The Scenario

Jon (Incident Commander) gets paged. Chidi (per [user-personas.md](../01-product/user-personas.md)'s Wednesday scenario) is already looking at the same thing without having been paged at all — an error spiked after last night's deploy.

## What They See, Step by Step

1. **One alert, not a flood.** Per story C2 and [event-catalog.md](../02-domain/event-catalog.md)'s `ErrorGrouped` event, the alert is for the root cause — one `AlertTriggered` per `ErrorGroup`, not one per affected session. This is the concrete payoff of write-time grouping: Jon isn't triaging a hundred near-identical pages at 2 a.m.
2. **"Which release did this" is already answered.** [event-catalog.md](../02-domain/event-catalog.md)'s `RegressionDetected` event has already attributed the spike to last night's release, per story C1 — nobody bisects deploys by hand.
3. **Open one affected session's replay.** Watching what the user actually did leading up to the error — the parity capability every comparator in [competitive-analysis.md](../00-overview/competitive-analysis.md) also offers.
4. **Jump straight to the backend trace for that exact moment.** Story J1's replay-to-trace correlation, via the shared session identifier from [component-diagrams.md](../04-architecture/component-diagrams.md) — one click from "here's what the user saw" to "here's what the backend was doing at that instant," instead of manually correlating timestamps across two separate tools.
5. **If it's ambiguous whether this is a bug or an attack:** [event-catalog.md](../02-domain/event-catalog.md)'s `SecuritySignalReceived` events for that session are already correlated into the same timeline, per story D2 — the wedge capability that answers "is this frontend logic, or is someone probing checkout for fraud" without switching to a separate security tool.

## If It Turns Out to Involve Customer Data

Per [user-personas.md](../01-product/user-personas.md)'s Jon scenario, an incident touching customer data carries a chain-of-custody burden most generic incident-response guidance doesn't cover. The path from here is [security-monitoring.md](security-monitoring.md) (if a security signal was involved) or directly to generating an EvidenceExport (story J2) — a single action producing a frozen, hash-verifiable package rather than screenshots assembled under pressure.

---

## What Makes This Workflow Different From a Parity-Only Tool

Steps 1–3 are what any comparator offers. **Step 4 (replay-to-trace) is parity at the category-leading end, and step 5 (security correlation in the same timeline) is wedge** — per the tags established in [user-stories.md](../01-product/user-stories.md). The workflow reads as one continuous investigation specifically because those two tiers were designed to ship together from Milestone 1, per [feature-roadmap.md](../01-product/feature-roadmap.md), not because the wedge was bolted on afterward.

## What This Feeds Next

[performance-monitoring.md](performance-monitoring.md) covers the C1-driven regression-detection workflow in its own right, for cases where nothing broke outright but a metric moved. [security-monitoring.md](security-monitoring.md) covers the D2 correlation path in depth.
