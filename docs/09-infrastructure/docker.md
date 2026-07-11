# Docker

> Status: Image and container conventions for every service in [container-diagrams.md](../04-architecture/container-diagrams.md), underpinning both the [compose.md](compose.md) and [kubernetes.md](kubernetes.md) profiles.

---

## Image Distribution Must Follow the Air-Gapped Update Pattern, Not `docker pull`

A Dockerfile that pulls its base image from a public registry at build time, or a Compose/Kubernetes setup that expects `docker pull` to reach a registry at deploy time, silently breaks for the air-gapped deployments [system-context.md](../04-architecture/system-context.md) requires Imora to support. Container images are exactly the kind of artifact [deployment-model.md](../04-architecture/deployment-model.md)'s signed-bundle update mechanism already covers — staged and signed outside the air gap, transferred via approved removable media, verified locally. **Imora ships as a set of pre-built, signed images distributed through that same mechanism**, not as Dockerfiles requiring a live registry at deploy time. Connected deployments may still pull from a registry for convenience; air-gapped ones use the bundle path — the images themselves are identical either way, only the distribution channel differs.

---

## Build and Runtime Conventions

- **Multi-stage builds.** Build tooling, dependency caches, and source never ship in the final image — only compiled/bundled output and its runtime dependencies. This shrinks both image size and attack surface.
- **Minimal base images** (distroless or slim variants, not full OS images) for the same reason — fewer packages in the final image means fewer components to patch and fewer things an attacker who gains container access can use.
- **Non-root user by default.** Every service container runs as an unprivileged user; no service needs root to do its job, and running as root is the kind of default-permissive choice [business-rules.md](../02-domain/business-rules.md) BR-7's "deny-by-default" philosophy argues against generally, applied here to infrastructure rather than data capture.
- **Read-only root filesystem where the service allows it** — a direct, concrete mitigation for the Tampering threats [threat-model.md](../08-security/threat-model.md) already identified: a compromised `ingestion` or `query-api` container with a read-only filesystem can't persist a modified binary or write a backdoor to disk, even with an initial foothold. Services that need a writable scratch directory (temp files, caches) get an explicitly mounted, narrowly-scoped writable volume — not a writable root.

---

## What's Deliberately Not Modeled Here

- Specific base image tags/versions — a maintenance decision, updated over time, not a one-time architectural choice.
- Image vulnerability scanning as a pipeline gate — `09-infrastructure/ci-cd.md`.
- The signed-bundle packaging format itself — already specified at the update-mechanism level in [deployment-model.md](../04-architecture/deployment-model.md); this document only establishes that container images travel through it.

---

## What This Feeds Next

`docs/09-infrastructure/kubernetes.md` should carry these same conventions into the cluster profile, plus the MinIO-initialization-ordering requirement from [compose.md](compose.md) reimplemented as an init container or Job rather than a Compose dependency.
