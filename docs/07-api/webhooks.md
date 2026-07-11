# Webhooks

> Status: Research-based, current as of July 2026. Resolves what [event-catalog.md](../02-domain/event-catalog.md) deferred: which events get exposed externally, and the delivery mechanics (signing, retry, SSRF protection) that make outbound delivery to a customer-configured endpoint safe.

---

## Which Events Are Exposed, and Why This List Grew by One

[event-catalog.md](../02-domain/event-catalog.md) flagged AlertTriggered, RegressionDetected, and EvidenceExportGenerated as likely candidates. Confirmed, plus one addition:

- **AlertTriggered** — the obvious case; customers already route alerts to their own on-call tooling.
- **RegressionDetected** — story C1's release-attributed regression, useful to pipe into a customer's own deploy-gating tooling.
- **EvidenceExportGenerated** — notifies that an export completed, without carrying its contents (see below).
- **ConfigurationChanged**, added here — [threat-model.md](../08-security/threat-model.md) identified that gap-detection and unmask-review mechanisms exist internally but nothing actively surfaces them externally. A webhook on `ConfigurationChanged` is what lets a customer's own SIEM alert in real time the moment someone weakens a RetentionPolicy or changes a field's PHI classification — an external, independent check on exactly the tampering scenario [threat-model.md](../08-security/threat-model.md) was worried about, not just an internal `notification-service` alert that a compromised insider could also suppress.

**DeletionSkippedDueToHold and other AccessAuditEvent variants are deliberately not exposed as webhooks** — that volume belongs in the REST API's audit-trail query surface ([rest-api.md](rest-api.md)), not a push mechanism; webhooks are for events a customer needs to react to immediately, not a firehose of every access.

---

## Delivery Security

- **HMAC-SHA256 signing**, per-endpoint secret, in a signature header verified before the receiver does anything with the payload — the baseline every major webhook provider uses. (Ephemeral, short-lived signing keys via a JWKS-style endpoint are a legitimate future hardening step, not required for the initial design.)
- **SSRF protection on every delivery attempt, not just at configuration time:** an explicit allowlist plus DNS-resolution check immediately before each fetch — resolve the hostname, refuse delivery if it resolves to a private, loopback, or link-local IP. Checking only when the customer first configures the URL is insufficient; DNS rebinding means a hostname can resolve safely at setup time and to an internal address at actual delivery time.
- **Retry with exponential backoff and jitter**, spanning roughly 3 days, capped at a bounded number of attempts — enough to survive a receiver's temporary outage without becoming an unbounded retry storm against a permanently broken endpoint.

---

## Payload Content: Metadata and References Only, Never Sensitive Content Inline

**No webhook payload ever carries session, PHI, or export content directly — only identifiers and a link back to the access-controlled REST API.** `EvidenceExportGenerated`'s payload is `{exportId, incidentReference, contentHash, generatedAt}`, not the export itself; retrieving the actual content requires `GET /v1/evidence-exports/{id}` through [rest-api.md](rest-api.md)'s authenticated, audited path.

This resolves a question worth stating explicitly rather than leaving implicit: **does a webhook delivery need its own AccessAuditEvent?** No — because it never carries anything sensitive, there's nothing to audit at delivery time. The actual access event fires when someone uses the webhook's reference to fetch the real content through the REST API, which already produces `RecordExported`, per [rest-api.md](rest-api.md). A customer-configured webhook endpoint is outside Imora's own access controls by definition, so it's the wrong place to ever put content that controls are supposed to govern.

---

## Air-Gapped Consistency

Webhooks are outbound-only and customer-configured — consistent with [system-context.md](../04-architecture/system-context.md)'s "optional convenience, never load-bearing" rule for external systems. An air-gapped deployment can still use webhooks pointed at another system inside the same isolated network (an internal SIEM, an internal ticketing system) — the same "the external system just has to be inside the air gap too" logic already established for SSO in [authentication.md](../08-security/authentication.md).

---

## What's Deliberately Not Modeled Here

- Exact JSON payload schemas per event — `07-api/openapi.yaml`.
- Webhook management UI/API (registering, testing, viewing delivery history) — a `dashboard`/`rest-api.md` concern once this design is accepted.
- Ephemeral signing-key rotation mechanics — a future hardening step noted above, not specified further here.

---

Sources: [Webhook Security: HMAC, Retries, Idempotency](https://didit.me/blog/webhook-security-patterns/), [Webhook Security Guide: HMAC Signatures & Replay Protection — Hooklistener](https://www.hooklistener.com/learn/webhook-security-fundamentals), [Standard Webhooks Specification](https://github.com/standard-webhooks/standard-webhooks/blob/main/spec/standard-webhooks.md).

## What This Feeds Next

`docs/07-api/openapi.yaml` is the last file in `07-api/` — it should formalize [rest-api.md](rest-api.md)'s resource surface and this document's payload shapes into an actual spec.
