# Imora — Software Design System

> Build-ready architectural conventions for how a package, a service boundary, and a dependency graph are shaped — the layer above [`coding-standards.md`](coding-standards.md): that file governs a line/function, this one governs a package/service. **Part A** covers the six Go backend services; **Part B** covers `browser-sdk` and `dashboard`. Full reasoning and how each rule ties to a decision already made elsewhere in this project: [`research/11-engineering/README.md#software-design-system`](../research/11-engineering/README.md#software-design-system).

---

# Part A — Go (Backend Services)

## 1. The One Rule Everything Else Follows

**Hexagonal architecture, domain at the center.** Business logic knows nothing about HTTP, SQL, or ClickHouse. Infrastructure depends on the domain; the domain never depends on infrastructure.

```
        ┌─────────────────────────────────────┐
        │   Adapters (HTTP handlers, Postgres/  │
        │   ClickHouse/MinIO clients, gateway)  │
        │                                        │
        │   ┌───────────────────────────────┐   │
        │   │  Ports (consumer-declared,     │   │
        │   │  narrow interfaces)            │   │
        │   │                                 │   │
        │   │   ┌───────────────────────┐    │   │
        │   │   │  Domain                │    │   │
        │   │   │  (packages/domain-types│    │   │
        │   │   │  Session, LegalHold,   │    │   │
        │   │   │  BR-1 through BR-7)    │    │   │
        │   │   └───────────────────────┘    │   │
        │   └───────────────────────────────┘   │
        └─────────────────────────────────────┘
              dependencies point inward only
```

**The concrete proof this is already real, not aspirational:** `query-api`'s `AuditedQueryHandler` is a use-case-layer type sitting between the HTTP adapter and the domain — it's *why* "every sensitive read produces an audit event" is a compile-time-enforced property, not a convention a reviewer has to remember to check. **The actual thesis: prefer a design where the wrong thing doesn't compile, over a design where the wrong thing merely gets caught in review.** Every rule below exists in service of that one sentence.

## 2. Package Boundaries

```
services/query-api/
├── cmd/
│   └── main.go              # §6 Service Bootstrap — the only place adapters and domain meet
├── internal/
│   ├── domain/               # pure business logic — no imports outside packages/domain-types
│   ├── handlers/              # HTTP adapters — thin, per §5 Coding Standards
│   └── store/                 # port implementations — postgres_legal_hold_store.go, clickhouse_audit_writer.go
└── go.mod
```

| Directory | Rule |
|---|---|
| `internal/` (per service) | Everything specific to one bounded context. Compiler-enforced unimportable from outside — this is the actual mechanism that keeps `dashboard` calling `query-api`'s API instead of reaching into its internals. |
| `packages/domain-types`, `packages/event-schemas` | **The only cross-service Go import surface.** Nothing else crosses a service boundary as a Go import. |
| Depth | Flat — one or two levels (`internal/domain`, `internal/handlers`, `internal/store`). Not `internal/domain/entities/session/v2`. |

If two services need to share logic beyond the domain types: either it belongs in `packages/` (genuinely domain-level), or the services talk over an API/event — never a shared internal package.

## 3. Ports and Adapters, Concretely

- **Port** = the interface the domain/use-case layer declares for what it needs (`LegalHoldStore`, not `PostgresClient`) — narrow, 1–3 methods, declared where it's consumed.
- **Adapter** = the concrete implementation (`PostgresLegalHoldStore`, `ClickHouseAccessAuditWriter`) — lives in `internal/store/`, never in `packages/domain-types`.

```go
// internal/domain/legal_hold.go — the port, declared by its consumer
type LegalHoldStore interface {
    IsHeld(ctx context.Context, recordID uuid.UUID) (bool, error)
}

// internal/store/postgres_legal_hold_store.go — the adapter, sqlc-backed
type PostgresLegalHoldStore struct {
    q *sqlcgen.Queries
}

func (s *PostgresLegalHoldStore) IsHeld(ctx context.Context, recordID uuid.UUID) (bool, error) {
    return s.q.CheckActiveLegalHold(ctx, recordID)
}
```

A service swap (e.g. Redis-backed hold cache → something else at cluster scale) is a one-file change: replace the adapter, the port and every consumer of it stay untouched.

## 4. The Repository Pattern Across Two Stores

A repository is a port that hides *which store* backs it from the domain layer — necessarily different technology per entity here, given the dual-store split.

```go
// packages/domain-types — the shape every repository returns; never a store-specific row type
type LegalHold struct {
    ID        uuid.UUID
    Scope     ScopePredicate
    AppliedBy uuid.UUID
    AppliedAt time.Time
}

// query-api's port for the high-volume, ClickHouse-backed side
type AccessAuditEventWriter interface {
    Write(ctx context.Context, event AccessAuditEvent) error
}

// workers' port for the small-cardinality, Postgres-backed side
type LegalHoldStore interface {
    IsHeld(ctx context.Context, recordID uuid.UUID) (bool, error)
    Apply(ctx context.Context, hold LegalHold) error
}
```

Neither consumer imports `sqlc`'s generated package or `clickhouse-go` directly — each depends on its own narrow port. The translation from store row to domain type happens inside the adapter, never leaks past it.

## 5. Dependency Injection: Explicit, No Framework

**Manual constructor wiring in each service's `main.go`. No `wire`, no `fx`, no reflection-based container.**

Why, specifically for this project (not a generic preference): every service's dependency graph is small and static — Operational Simplicity means a 2–3 person team debugging a startup failure at 2am needs a plain stack trace through a function call, not through a DI container's reflection layer. Revisit `wire` (compile-time codegen, not `fx`'s runtime reflection) only if a service's graph grows large enough that hand-wiring becomes actually error-prone — not pre-emptively.

## 6. Service Bootstrap and Lifecycle

Every service's `main.go` follows the same five-step shape:

```go
func main() {
    // 1. Load and validate config — fail fast, specific error, before anything else
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config: %v", err) // the one permitted log.Fatal, per coding-standards.md §8
    }

    // 2. Construct adapters
    db := postgres.Connect(cfg.PostgresDSN)
    holdStore := store.NewPostgresLegalHoldStore(db)
    auditWriter := store.NewClickHouseAuditWriter(cfg.ClickHouseDSN)

    // 3. Construct the domain/use-case layer, injected through ports
    checker := domain.NewLegalHoldChecker(holdStore)
    sweeper := domain.NewRetentionSweepScheduler(checker, auditWriter)

    // 4. Wire the delivery mechanism around the use-case layer
    cronHandler := handlers.NewSweepCronHandler(sweeper)

    // 5. Start, then block on the shutdown signal — see coding-standards.md §8
    runWithGracefulShutdown(cronHandler)
}
```

Wiring, configuration, and graceful shutdown are not each service's independent invention — a developer moving from `workers` to `ingestion` finds the same shape.

## 7. Error Contracts at a Service Boundary

Internal wrapped errors (`%w`) stay internal. **At a true boundary — an HTTP response, a webhook payload — translate to a small, stable set of API error codes.**

```go
// internal/handlers — the translation happens here, at the boundary
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
    session, err := h.useCase.GetSession(r.Context(), sessionID)
    switch {
    case errors.Is(err, domain.ErrNotFound):
        writeAPIError(w, http.StatusNotFound, "session_not_found") // stable code, not err.Error()
    case err != nil:
        logctx.From(r.Context()).Error("get session failed", "error", err) // full detail logged internally
        writeAPIError(w, http.StatusInternalServerError, "internal_error") // never leaked to the caller
    default:
        writeJSON(w, session)
    }
}
```

A leaked `pgconn.PgError` string in a JSON response is the same class of problem as an unmasked PHI field: implementation detail exposed to a caller who shouldn't see it, just lower-stakes.

## 8. Idempotency

Structural, not "the retry logic is careful":

| Component | Idempotent? | Why |
|---|---|---|
| `EvidenceExportGenerator` | Yes, by construction | `contentHash` is derived from the frozen input set — retrying "the same" export produces the same hash. `exportId` should be caller-supplied or derived deterministically, not server-generated per attempt. |
| `RetentionSweepScheduler` per-record processing | Yes, by construction | Re-running hold-check-then-delete against an already-deleted record is a no-op — the record's absence *is* the terminal state. |
| `ingestion` event writes | **No — deliberately not.** | De-duplicating high-volume telemetry is real cost for a failure mode (an occasional duplicate replay frame) that doesn't threaten the audit-trail guarantee. Compliance-critical writes get idempotency; high-volume telemetry doesn't. |

## 9. What's Deliberately Not Specified Here

- Specific port/adapter interface names beyond the examples above — defined as each service is actually built, not invented ahead of the code that needs them.
- Build tooling for the monorepo (Nx/Turborepo/Bazel/none) — see [`architecture.md §12`](architecture.md).
- Event/message-bus internal wiring at cluster scale — downstream of [`architecture.md §9`](architecture.md)'s cluster profile once that's being built.

---

# Part B — TypeScript (`browser-sdk`, `dashboard`)

## 10. Two Different Shapes, Sharing Only the Compiler/Testing Conventions

`browser-sdk` is a library embedded in someone else's page. `dashboard` is Imora's own product surface. They share [`coding-standards.md`](coding-standards.md)'s Part B (strict TypeScript, Zod at boundaries, Vitest) — not the same structure.

## 11. `browser-sdk`: Framework-Agnostic Core, Thin Wrappers

```
sdk/browser-sdk/
├── packages/
│   ├── core/            # @imora/core — capture, masking, transport. Zero framework dependency.
│   ├── react/            # @imora/react — thin wrapper, re-exports core's full API
│   ├── vue/               # @imora/vue
│   └── angular/           # @imora/angular
```

Each framework package's only job is a lifecycle adapter (a React error boundary, Vue's `errorCaptured`) around `@imora/core` — no capability exists in one wrapper and not another. A wrapper that accumulates independent logic beyond the lifecycle hook is a design smell, per [SDK API](../research/06-api/README.md#sdk-api)'s parity guarantee across frameworks.

**Build tool: [`tsup`](https://tsup.egoist.dev)** — wraps esbuild with library-authoring defaults (dual ESM+CJS, generated `.d.ts`), near-zero config. Escalate to Rollup only if a specific `tsup` build misses the [~20KB gzipped bundle budget](architecture.md) and Rollup's tree-shaking is demonstrably the fix — not the default choice.

## 12. `dashboard`: TanStack Start, File-Based Routes, Split State

**Why [TanStack Start](https://tanstack.com/start), not a plain React SPA or Next.js:** `browser-sdk` already ships a first-class React wrapper, so React itself was already the low-risk choice — TanStack Start extends that reasoning one layer deeper, since its router (TanStack Router) and server-state layer (TanStack Query) are one integrated stack from the same team, not three libraries wired together by hand. It also deploys to plain Node/Docker with no platform-specific dependency (its Nitro build target), which fits this project's air-gapped/self-hosted constraint better than Next.js's more Vercel-optimized feature set.

```
dashboard/
├── app/
│   ├── routes/
│   │   ├── sessions.search.tsx        # mirrors user-stories.md's Flow B
│   │   ├── legal-holds.tsx             # mirrors Flow G
│   │   └── evidence-export.tsx         # mirrors Flow H
│   ├── server/                          # server functions — see the rule below
│   └── shared/{components,hooks,lib}/   # generic, reusable across routes only
└── app.config.ts
```

### The One Rule TanStack Start Adds: Server Functions Are a Proxy, Never a Data-Access Path

A plain SPA could only ever call `query-api` from the browser. TanStack Start's server functions and route loaders execute **on the server** — which makes it technically possible to import a database driver and query Postgres/ClickHouse directly from `dashboard`, bypassing `query-api`'s `AuditedQueryHandler` entirely. **This must never happen.** Every server function's only external call is HTTP to `query-api`'s REST API — identical in shape to what client-side code would call, just executing server-side for the initial render.

```typescript
// app/server/sessions.ts — correct: proxies to query-api, holds no store credentials
export const getSession = createServerFn({ method: "GET" })
  .validator((id: string) => id)
  .handler(async ({ data: id }) => {
    const res = await fetch(`${process.env.QUERY_API_URL}/v1/sessions/${id}`, {
      headers: { Authorization: `Bearer ${getServerSessionToken()}` },
    });
    return SessionSchema.parse(await res.json()); // Zod validation per coding-standards.md §18
  });

// ❌ never this — a server function is not a repository
// const session = await db.query("SELECT * FROM sessions WHERE id = $1", [id]);
```

**Server state (TanStack Query) and client-only UI state (Zustand) are never mirrored into each other**, and a route loader prefetches into the same TanStack Query cache the client reuses on navigation — one cache, two execution contexts, not two caches to keep in sync:

```typescript
// route loader — SSR prefetch, populates the cache the client-side hook below reuses
export const Route = createFileRoute("/sessions/search")({
  loader: ({ context: { queryClient }, deps: { environment, filters } }) =>
    queryClient.ensureQueryData({
      queryKey: ["sessions", environment, filters],
      queryFn: () => getSessions({ data: { environment, filters } }), // calls the server function above
    }),
});

// client-side hook — same query key, cache hit if the loader already ran
function useSessions(environment: string, filters: SessionFilters) {
  return useQuery({ queryKey: ["sessions", environment, filters], queryFn: () => getSessions({ data: { environment, filters } }) });
}

// client-only state — no provider, no boilerplate
const useEnvironmentStore = create<{ environment: string; setEnvironment: (e: string) => void }>((set) => ({
  environment: "production", // per user-stories.md's global pattern
  setEnvironment: (environment) => set({ environment }),
}));
```

Zustand, not Redux — this codebase's actual client-only state (environment selector, modal state, in-progress form values) is the small/minimal-boilerplate case; Redux's justification (large team, complex normalized state, time-travel debugging) doesn't describe this project.

## 13. What's Deliberately Not Specified Here (TypeScript)

- Visual design system / component library choice — a product-design decision, not an architecture one; see [`research/10-design/README.md`](../research/10-design/README.md).
- Message-passing/event architecture between `browser-sdk` and a host page beyond what [SDK API](../research/06-api/README.md#sdk-api) already specifies.

---

## What This Feeds

Implementation — this and [`coding-standards.md`](coding-standards.md) are the last two planning documents before code; nothing downstream of these two is left to invent ahead of the code that needs it.
