# 0007. Container images signed with a Cosign key pair, not Sigstore keyless signing

> Status: Accepted. Full reasoning in [ci-cd.md](../../09-infrastructure/ci-cd.md).

## Context

The current default for container image signing is Sigstore's keyless signing — Cosign authenticates via OIDC, Fulcio issues a short-lived certificate, and Rekor (a public transparency log) records the signature. Verifying a keyless signature means checking it against Rekor, which requires reaching public Sigstore infrastructure over the internet — the same shape of problem already resolved for SSO ([authentication.md](../../08-security/authentication.md)) and the vault's KMS ([encryption.md](../../08-security/encryption.md)/[pii-redaction.md](../../08-security/pii-redaction.md)).

## Decision

Sign images with a traditional Cosign key pair. The private key stays with the build pipeline; the public key ships embedded in every deployment bundle and manifest. Verification — connected or air-gapped — checks against that embedded public key locally, with zero dependency on Rekor, Fulcio, or any other public Sigstore service.

## Alternatives Considered

- **Keyless signing (the modern default):** rejected specifically for air-gapped verification — silently breaks the "no external system in the required path" rule from [system-context.md](../../04-architecture/system-context.md) if adopted without modification.
- **Self-hosted Fulcio/Rekor mirror:** not rejected outright, but deferred — a legitimate future option for the cluster/Enterprise profile, but unnecessary operational overhead for the connected+air-gapped baseline this ADR covers, following the same "don't require infrastructure Priya's team doesn't need" reasoning as ADR 0004.

## Consequences

- One build produces both the registry-published image (connected deployments) and the signed bundle (air-gapped transfer) — verified identically either way, which is what guarantees the two deployment paths never quietly diverge.
- Key-pair rotation for the signing key itself follows the same versioned-key principle as the vault's KEK rotation (ADR-adjacent to [encryption.md](../../08-security/encryption.md)), not a separate scheme invented for this purpose.
