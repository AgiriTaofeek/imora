# CI/CD

> Status: Research-based, current as of July 2026. Specifies how the images [docker.md](docker.md) requires to be "pre-built and signed" actually get built and signed, and how the two distribution paths from [deployment-model.md](../04-architecture/deployment-model.md) (registry pull for connected, signed bundle for air-gapped) come from the same pipeline run rather than drifting apart.

---

## The Finding: Keyless Signing Doesn't Work for Air-Gapped Verification

The current default for container image signing is Sigstore's **keyless** signing: Cosign authenticates via OIDC, Fulcio issues a short-lived certificate, and Rekor (a public transparency log) records the signature. This is genuinely the modern standard — but verifying a keyless signature means checking it against Rekor, which is public Sigstore infrastructure reachable only over the internet. That's the same shape of problem [encryption.md](../08-security/encryption.md)'s cloud-KMS rejection and [authentication.md](../08-security/authentication.md)'s SSO-IdP-must-be-internal finding already solved elsewhere in this doc set: a verification step depending on a public external service silently breaks the air-gapped requirement from [system-context.md](../04-architecture/system-context.md).

**Resolution: Imora signs images with a traditional Cosign key pair, not keyless.** The private key stays with the build pipeline; the public key ships embedded in every deployment bundle (connected or air-gapped) and in the Compose/Kubernetes manifests themselves. Verification — whether at a connected deployment's registry pull or an air-gapped deployment applying a transferred bundle — checks the signature against that embedded public key locally, with zero dependency on Rekor, Fulcio, or any other public Sigstore service. This is a small deviation from the "modern default," made for the same reason every other air-gapped-compatibility decision in this doc set was made.

---

## Pipeline Stages

1. **Build** — multi-stage builds per [docker.md](docker.md)'s conventions, producing minimal, non-root, read-only-filesystem-compatible images.
2. **Test**, including a specific gate this document adds: **automated verification of the structural guarantees this doc set has been establishing**, not just functional tests. Concretely: a test that attempts `DELETE`/`ALTER` against `access_audit_events` using the `ingestion`/`query-api` service account credentials and asserts it fails, per [threat-model.md](../08-security/threat-model.md)'s GRANT-restriction finding; a test confirming containers actually run as non-root with a read-only root filesystem, per [docker.md](docker.md). This turns "should be true" claims made throughout `02-domain/`, `06-data/`, and `08-security/` into CI-verified facts, not just design intentions that could silently drift as the codebase changes.
3. **SBOM generation** — Syft generating SPDX or CycloneDX format, attached to the image as a signed Cosign attestation. Worth doing regardless of any specific mandate, given [target-users.md](../00-overview/target-users.md) includes government agencies as a target sector, where software supply-chain provenance is an increasingly standard procurement expectation.
4. **Sign** — the key-pair Cosign signing from the finding above.
5. **Publish, two paths from one build:**
   - **Registry push**, for connected deployments' `docker pull`/Kubernetes image pull.
   - **Signed bundle packaging**, for the air-gapped transfer mechanism [deployment-model.md](../04-architecture/deployment-model.md) already specified — staged, then moved via approved removable media. Producing both from the identical build artifact is what guarantees a connected customer and an air-gapped customer are running the exact same code, not two paths that could quietly diverge.

---

## What's Deliberately Not Modeled Here

- Specific CI platform (GitHub Actions, GitLab CI, etc.) — a tooling choice, not an architecture decision.
- Exact test framework/coverage thresholds — `10-engineering/testing.md`.
- Key-pair rotation schedule for the Cosign signing key itself — follows the same versioned-key principle as [encryption.md](../08-security/encryption.md)'s KEK rotation, not repeated here.

---

Sources: [Signing Containers — Sigstore Docs](https://docs.sigstore.dev/cosign/signing/signing_with_containers/), [Container Supply Chain Security With Sigstore and Cosign](https://devopsil.com/articles/2026-03-21-supply-chain-security-sigstore-cosign), [How to Sign an SBOM with Cosign — Chainguard Academy](https://edu.chainguard.dev/open-source/sigstore/cosign/how-to-sign-an-sbom-with-cosign/).

## What This Feeds Next

`docs/09-infrastructure/observability.md` is the last file in this folder — monitoring Imora's own infrastructure, distinct in scope from the product's own observability features it sells to customers.
