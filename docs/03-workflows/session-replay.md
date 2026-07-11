# Session Replay

> Status: Story C3's specific workflow — finding the right session among many from a vague description, distinct from [error-investigation.md](error-investigation.md)'s "I already have an alert pointing at a session" path.

---

## The Scenario

Per [user-stories.md](../01-product/user-stories.md) story C3: a support ticket says "checkout is broken," with no repro steps, no error ID, nothing to jump straight to. This is the harder, more common case — most bug reports don't arrive with a session already identified.

## The Workflow

1. **Search by user, timeframe, page, or signal — not just by ID.** Per story C3's acceptance criteria, candidate sessions surface in seconds from a user identifier, a URL, a timeframe, or an error/rage-click signal, not a manual log grep. This is the parity bar every comparator in [competitive-analysis.md](../00-overview/competitive-analysis.md) sets — Imora has to match it before anything else here matters.
2. **Replay fidelity has to be sufficient on its own**, per [domain-model.md](../02-domain/domain-model.md)'s rrweb-based capture (full snapshot plus incremental DOM/interaction events) — reproducing the reported behavior shouldn't require also reading raw logs side by side. If the replay alone can't answer "what did the user actually do," the capture layer has failed regardless of how good the search is.
3. **Masked by default, escalatable if needed.** Per [pii-redaction.md](../08-security/pii-redaction.md)'s two-tier model, the default view is safe to look at without a second thought — PHI/PII fields stay masked unless the specific debugging need requires the real value, in which case the audited UNMASK path from [authorization.md](../08-security/authorization.md) applies exactly as it would anywhere else. Nothing about "just trying to reproduce a bug" gets a quieter audit trail.

## Why This Is Listed Separately From Error Investigation

[error-investigation.md](error-investigation.md) starts from an alert that already points at a session. This workflow starts from nothing but a vague human description — and per [user-personas.md](../01-product/user-personas.md), that's Chidi's actual daily reality more often than a clean alert is. A product that's only good at the alert-triggered path but weak at open-ended search would fail the workflow that happens most often.

## What This Feeds Next

[security-monitoring.md](security-monitoring.md) covers the equivalent search/investigation workflow from Adaeze's and Marcus's side — compliance-driven rather than debugging-driven, but built on the same underlying session data.
