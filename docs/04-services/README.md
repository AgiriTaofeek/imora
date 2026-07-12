# Services

## Alert Engine

> Status: Deferred — responsibility, technology shape, and reads/writes are already specified in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)'s container table. Reserved for implementation-level detail (config, env vars, runbook) once real code exists, per [docs/README.md](../README.md)'s folder table.

Service that evaluates alert rules against incoming data and triggers notifications.

### Overview

_TBD_

---

## Browser SDK

> Status: Deferred at the service-implementation level, but not undocumented — the public API surface is fully specified in [SDK API](../06-api/README.md#sdk-api), and capture-time masking behavior in [PII Redaction](../07-security/README.md#pii-redaction). This file is reserved for implementation-level detail (build tooling, package structure) once real code exists, per [docs/README.md](../README.md)'s folder table.

Client-side SDK for capturing errors, performance, and session data in the browser.

### Overview

_TBD_

---

## Dashboard

> Status: Deferred at the service-implementation level, but not undocumented — its Conformist relationship to `query-api` (no domain entities, no audit-log authority) is specified in [Bounded Contexts](../02-domain/README.md#bounded-contexts), and its interaction behavior in [Interaction Patterns](../10-design/README.md#interaction-patterns). Reserved for implementation-level detail once real code exists, per [docs/README.md](../README.md)'s folder table.

Web application used to visualize and investigate data collected by Imora.

### Overview

_TBD_

---

## Gateway

> Status: Deferred — responsibility (authN/authZ, rate limiting, actor-context stamping) is already specified in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) and [Authentication](../07-security/README.md#authentication). Reserved for implementation-level detail once real code exists, per [docs/README.md](../README.md)'s folder table.

Edge/API gateway responsible for routing and protecting inbound traffic.

### Overview

_TBD_

---

## Ingestion Service

> Status: Deferred — responsibility, technology shape, and its Shared Kernel relationship to `browser-sdk`/`query-api` are already specified in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) and [Bounded Contexts](../02-domain/README.md#bounded-contexts). Reserved for implementation-level detail once real code exists, per [docs/README.md](../README.md)'s folder table.

Service responsible for receiving and validating incoming events from SDKs.

### Overview

_TBD_

---

## Notification Service

> Status: Deferred — responsibility and its Conformist relationship to `alert-engine` are already specified in [Container Diagrams](../03-architecture/diagrams.md#container-diagrams) and [Bounded Contexts](../02-domain/README.md#bounded-contexts); delivery mechanics (signing, retry, SSRF protection) are specified in [Webhooks](../06-api/README.md#webhooks). Reserved for implementation-level detail once real code exists, per [docs/README.md](../README.md)'s folder table.

Service responsible for delivering alerts and notifications to configured channels.

### Overview

_TBD_

---

## Query API

> Status: Deferred at the service-implementation level, but not undocumented — this is one of the two containers with full internal component detail already specified in [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) (`AuditedQueryHandler`, `AccessAuditWriter`, `UnmaskEscalationHandler`), plus the full REST surface in [REST API](../06-api/README.md#rest-api). Reserved for implementation-level detail (actual code structure) once real code exists, per [docs/README.md](../README.md)'s folder table.

API for querying stored events, metrics, and sessions.

### Overview

_TBD_

---

## Workers

> Status: Deferred at the service-implementation level, but not undocumented — this is the other container with full internal component detail already specified in [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) (`LegalHoldChecker`, `SelectivePurgeExecutor`, `EvidenceExportGenerator`, `DeletionExecutor`), plus the CronJob/Deployment split in [Kubernetes](../12-infrastructure/README.md#kubernetes). Reserved for implementation-level detail (actual code structure) once real code exists, per [docs/README.md](../README.md)'s folder table.

Background workers that process, enrich, and route ingested events.

### Overview

_TBD_

