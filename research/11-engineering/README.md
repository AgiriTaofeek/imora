# Engineering

## Branching Strategy

> Status: Trunk-based development — not a default choice, a specific fit for the monorepo decision in ADR [0003](architecture-decisions/0003-monorepo-structure.md).

---

### Trunk-Based Development

DORA's research on this is specific, not a general preference: elite performers who meet their reliability targets are 2.3x more likely to use trunk-based development, correlated specifically with three-or-fewer active branches, merging to trunk at least daily, and no code-freeze/integration phases.

**Why this matters more than usual for this repository specifically:** `packages/domain-types` is the literal Shared Kernel that `browser-sdk`, `ingestion`, and `query-api` all depend on directly, per [Bounded Contexts](../02-domain/README.md#bounded-contexts) and ADR [0003](architecture-decisions/0003-monorepo-structure.md). A long-lived feature branch touching that package accumulates drift risk across three consuming services simultaneously, not just within itself — trunk-based development's short-lived-branch discipline is what keeps that drift window small.

### Practice

- Feature branches live hours to a couple of days, not weeks — work that can't fit in that window gets built behind a feature flag and merged incomplete-but-inert, rather than kept on a long-lived branch.
- No permanent `develop` or `release` branch. Releases are tags cut directly from trunk, per [Release Process](README.md#release-process).
- **Hotfixes:** for a critical patch against an already-shipped version (the scenario [Release Process](README.md#release-process)'s air-gapped notification gap describes), branch from the affected tag, cherry-pick the fix, tag a patch release, delete the branch. This is a short-lived, purpose-specific branch, not a standing release-maintenance branch.

### What's Deliberately Not Modeled Here

- PR review requirements/approval counts — a team-process decision, not an architectural one.
- Commit message conventions — downstream tooling choice (e.g., whether commit messages drive changelog automation from [Release Process](README.md#release-process)).

---

## Coding Standards

> Status: Resolves a decision left implicit everywhere else in this doc set — what language backend services are actually written in — and states the reasoning, since leaving it unspecified indefinitely would eventually force an arbitrary choice under time pressure instead of a deliberate one.

---

### Language: Go for Backend Services, TypeScript for Client-Facing Code

**Backend services** (`gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`): Go. Not a neutral pick — it directly serves three constraints already established elsewhere in this doc set:

- **[Docker](../12-infrastructure/README.md#docker)'s minimal-image goal** — a Go service compiles to a single static binary, which fits a distroless/scratch base image about as tightly as multi-stage Docker builds get, with no runtime/interpreter to include.
- **[Deployment Model](../03-architecture/README.md#deployment-model)'s single-machine resource budget** — ClickHouse already claims the 4-core/16GB floor; a lower-memory-footprint runtime for the six backend services leaves more of that budget for the part that actually needs it.
- **Ecosystem alignment** — the closest architectural comparator (Uptrace) is written in Go, ClickHouse's own official client library ecosystem is Go-first, and most Kubernetes-native tooling (client libraries, operators) is Go, which matters directly for [Kubernetes](../12-infrastructure/README.md#kubernetes)'s cluster profile.

**`browser-sdk` and `dashboard`:** TypeScript, for reasons that aren't really a choice — `browser-sdk` is a browser library by definition, and `dashboard` is a web frontend consuming [REST API](../06-api/README.md#rest-api).

This does mean two languages across the codebase, not one — accepted deliberately rather than defaulting to a single-language monorepo for its own sake, since Node's resource footprint would work against the single-machine budget for no real benefit on the backend side.

### Formatting and Linting

- **Go: `gofmt` plus `golangci-lint`, both non-negotiable, both CI-gated.** `gofmt` isn't a style preference — it's what the Go toolchain already enforces by convention, and a diff that isn't `gofmt`-clean fails the build, not just code review. `golangci-lint` runs the linter set every backend service shares from one root config (not per-service configs that drift): `errcheck` (every returned error is either handled or explicitly discarded with `_ =`, never silently dropped), `govet`, `staticcheck`, `unused`, `ineffassign` (the default set), plus `bodyclose` (every `http.Response.Body` is closed — a real leak source in long-running services), `contextcheck`/`containedctx` (catches a context stored in a struct field instead of threaded as the first parameter, and a context passed to a function that doesn't inherit from a request's context — both break cancellation propagation silently), and `wrapcheck` (an error returned from an external package must be wrapped with call-site context before it crosses a service boundary, per the Error Handling section below).
- Go baseline: **1.24+**, tracking current stable rather than pinning to an old minor — this repository has no legacy-compatibility constraint forcing an older baseline, so there's no reason to forgo `go fix`-modernized idioms, the stabilized `testing/synctest` package (useful directly against [Testing](README.md#testing)'s concurrent-flow assertions), or the per-iteration loop variable semantics Go 1.22 made the language default (a `for _, h := range holds { go check(h) }` closure capturing the wrong iteration's value was a real, common bug class before 1.22 — it's structurally gone now, not just discouraged by convention).
- TypeScript: ESLint + Prettier, configured once in `packages/` and inherited by `browser-sdk` and `dashboard` rather than duplicated per package.

### Naming Conventions

Per [Effective Go](https://go.dev/doc/effective_go) and [Google's Go Style Decisions](https://google.github.io/styleguide/go/decisions.html) — the two documents that have defined idiomatic Go naming since before this project existed, restated here only where this codebase's own vocabulary makes a specific choice concrete:

- **`MixedCaps`, never underscores**, for every multi-word identifier — `RetentionPolicy`, not `retention_policy` — including the JSON-facing struct field *names in Go*, with the wire-format casing handled entirely by struct tags (see JSON & API Contract Conventions below), not by renaming the Go identifier itself.
- **Initialisms keep a single, consistent case for the whole word**: `ID` not `Id`, `HTTP` not `Http`, `URL` not `Url` — `SessionID`, `LegalHoldID`, `ClickHouseURL`. This is a `golangci-lint`(`staticcheck`/`revive`)-checked rule, not a manually-remembered one.
- **Receiver names are short (1–3 letters), an abbreviation of the type, and consistent across every method on that type** — `func (h *AuditedQueryHandler) ServeHTTP(...)` uses `h` everywhere that type has a method, never `handler` in one method and `h` in another.
- **Package names are short, lowercase, no underscores, and never a plural** — `store`, not `stores`; `auth`, not `authentication_helpers`. A package name doubles as its import qualifier (`store.Postgres...`), so a name that reads naturally at the call site is the actual test, not a name that reads well in isolation.
- **Exported identifiers this codebase's domain vocabulary already fixed elsewhere are not renamed at the code layer** — [Domain Model](../02-domain/README.md#domain-model)'s `Session`, `AccessAuditEvent`, `LegalHold`, `RetentionPolicy` are the Go type names verbatim, so a reader moving between this document, [Event Catalog](../02-domain/README.md#event-catalog), and the actual source never has to mentally translate between a "domain name" and a "code name" for the same concept.

### Error Handling

Per [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and the standard library's own `errors`/`fmt` conventions (Go 1.13+), applied specifically to where this codebase's error paths are compliance-relevant, not just correctness-relevant:

- **Every error is handled exactly once — never both logged and returned.** A function that does `log.Error(err); return err` produces the same failure logged twice by the time it reaches `main()`, once per layer that touched it, which is noise that actively degrades the [Observability](../12-infrastructure/README.md#observability) signal this doc set already depends on for real operational alerts. Handle an error at the one layer that has enough context to either recover from it or translate it into a caller-meaningful response — everywhere else, wrap and propagate.
- **Wrap with `fmt.Errorf("doing X: %w", err)` when the error crosses a package boundary a caller might reasonably want to inspect; wrap with a plain `%v` (no `Unwrap`) when it doesn't.** Per the standard library's own guidance, wrapping with `%w` makes the underlying error part of your function's effective API surface — a caller can `errors.Is`/`errors.As` through it. That's the right choice at a service's internal layer boundaries (e.g., `workers`' `LegalHoldChecker` returning a wrapped Postgres error so `DeletionExecutor` can distinguish "hold found" from "database unreachable" via `errors.Is`), and the wrong choice at a true external boundary (an HTTP handler must not leak a wrapped `*pgconn.PgError` into a JSON response body — translate to a stable API error code there instead, per the Software Design System section below).
- **Sentinel errors (`var ErrNotFound = errors.New(...)`) for conditions a caller branches on; `fmt.Errorf` for everything else.** A caller that needs to ask "was this specifically a not-found, or some other failure?" needs `errors.Is(err, ErrNotFound)` to work, which requires a declared sentinel or a custom type — not a dynamically-formatted string a caller would have to substring-match.
- **`errors.Join` (Go 1.20+), not a hand-rolled multi-error type, for the one place this codebase genuinely accumulates several independent failures at once:** [Kubernetes](../12-infrastructure/README.md#kubernetes)'s `workers` CronJob retention sweep processing a batch of candidate records, where one record's deletion failing shouldn't stop the sweep from continuing through the rest — `errors.Join` preserves every individual error, and `errors.Is`/`errors.As` still walk through a joined error to find a match, so callers don't lose the ability to check for a specific failure buried in the batch.
- **`panic` is not error handling.** Per Uber's guide, production code returns an error and lets the caller decide, full stop — the one narrow exception already established by this doc set's own architecture is a genuine programmer-error invariant violation inside `query-api`'s `AuditedQueryHandler` wrapper (e.g., a new route registered without going through it, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)), which should fail loudly at startup via a compile-time or init-time check, not a runtime panic reachable by request traffic.
- **Type assertions always use the "comma ok" form** (`v, ok := x.(T)`), never the bare single-return form, which panics on a mismatch — this is a lint-enforced rule (`staticcheck` flags the bare form), not a style suggestion left to memory.
- **`log.Fatal`/`os.Exit` appear at most once, in `main()`, in each service's binary.** Every other function — including deep in `workers`' retention-sweep logic — returns an error up the call stack; only the entry point decides that an error is fatal to the process.

### Structured Logging

**`log/slog`** (Go 1.21+ standard library) — no third-party logging dependency, consistent with [Deployment Model](../03-architecture/README.md#deployment-model)'s general bias toward fewer moving parts on the single-machine profile.

- **Request-scoped logger, threaded through `context.Context`, never a package-level global.** `gateway` constructs one `*slog.Logger` per inbound request, pre-populated with the fields every downstream log line in that request's call chain needs — `request_id`, `actor_user_id` (once `RequestContext` resolves it, per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)), and, per [Domain Model](../02-domain/README.md#domain-model)'s Environment note, `environment` — and passes it downstream via context, not as an extra function parameter threaded through every call site by hand. A goroutine spawned to handle part of that request inherits the same context and therefore the same logger fields automatically.
- **The message describes what happened; attributes describe the context it happened in.** `slog.Error("legal hold check failed", "session_id", id, "error", err)`, never a message string with the session ID string-formatted into it — attributes are what make logs queryable/filterable in whatever log aggregation [Observability](../12-infrastructure/README.md#observability) runs, and a value baked into the message text defeats that.
- **`slog.LogAttrs` (the allocation-efficient form) in genuinely hot paths only** — `ingestion`'s per-event write path is the one place in this codebase where this actually matters; everywhere else, the more readable `logger.Info("msg", "key", val)` form is the right default, and micro-optimizing it elsewhere is solving a problem this system doesn't have.
- **What never appears in a log line, full stop:** any field [PII Redaction](../07-security/README.md#pii-redaction) classifies as soft-masked-or-harder — logging an unmasked PHI/PII value defeats every guarantee the rest of this doc set builds around capture-time masking, regardless of how useful it would be for debugging. Log the record identifier, never the record's sensitive content.
- **Operational logs (this section) and `AccessAuditEvent`s (per [Audit Logging](../07-security/README.md#audit-logging)) are two different systems, on purpose** — a `slog` line about "legal hold check failed" is diagnostic telemetry for whoever runs the deployment; an `AccessAuditEvent` is compliance-critical, append-only, and retained per [Retention](../05-data/README.md#retention)'s regulatory clocks. Never route one into the other's storage or retention policy.

### Interface Design

Per Go's own standard-library convention and the community heuristic it codifies — [**"accept interfaces, return structs."**](https://google.github.io/styleguide/go/best-practices.html) A function that needs a dependency asks for the narrowest interface that satisfies what it actually calls, not the concrete type or a broad interface copied from the producer's package; a function that constructs something returns the concrete type, so its caller sees exactly what they got and doesn't pay an unnecessary indirection cost.

- **Interfaces are declared by the consumer, at the point of use — never pre-emptively in the producer's package "in case something needs to mock it."** `workers`' `LegalHoldChecker` component declares the one- or two-method interface it needs from whatever backs the active-hold lookup (per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)'s Redis cache); it does not import a broad `RedisClient` interface that exposes every Redis command Imora doesn't use. This is what keeps a swap (e.g., replacing the Redis-backed cache with something else at cluster scale) a one-file change instead of a repository-wide refactor.
- **Interfaces stay small — one to three methods is the practical ceiling** before it's a sign the abstraction is actually two responsibilities wearing one name. `AuditedQueryHandler`'s design (per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)) is the concrete example already built into this system: it isn't one large interface every query handler implements in full, it's a wrapper *type* that each narrow handler (`SessionQueryHandler`, `ErrorQueryHandler`, etc.) is constructed through — composition standing in for what a fatter interface would otherwise have to express.
- **A `New`-prefixed constructor function is Go's constructor**, since the language has no dedicated syntax for one. Where a type has more than two or three optional configuration values — `workers`' `RetentionSweepScheduler`, which per [Kubernetes](../12-infrastructure/README.md#kubernetes) needs a configurable sweep interval, a concurrency policy, and (in tests) an injectable clock — use the [functional options pattern](https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md) (`NewRetentionSweepScheduler(store, opts ...Option)`) rather than a constructor with five positional parameters or a half-populated config struct passed by convention.

### Generics

Go's generics (stable since 1.18, with type-alias parameterization completed in 1.24) are for **eliminating real, repeated type-specific duplication** — a generic `Paginate[T any](items []T, cursor string) (Page[T], error)` used identically by `query-api`'s Session, ErrorGroup, and AccessAuditEvent list endpoints is a legitimate use, since [REST API](../06-api/README.md#rest-api)'s cursor-pagination shape is genuinely identical across all of them. Generics are **not** a default tool reached for to make a single call site "more flexible" against a hypothetical future need — per this doc set's own recurring discipline of not inventing structure ahead of a real, current requirement, an unconstrained `T any` on a function with exactly one call site is very likely solving a problem that doesn't exist yet.

### Concurrency

- **`errgroup.Group`, not a bare `sync.WaitGroup`, wherever spawned goroutines can fail** — a `WaitGroup` alone leaves error propagation, first-error capture, and shared-context cancellation as three separate things the caller has to hand-roll; `errgroup` provides all three. This is directly relevant to `workers`' retention-sweep fan-out and any parallel per-record processing in `query-api`.
- **`errgroup.Group.SetLimit(n)`** (Go 1.20+) for bounded fan-out — the simplest production-safe way to cap concurrent goroutines against a batch of unknown size (e.g., a legal-hold scope resolving to thousands of candidate records) without a hand-built worker-pool/channel construction.
- **Every `context.WithCancel`/`WithTimeout`/`WithDeadline` is paired with `defer cancel()` at the same call site that created it** — an un-deferred cancel function is a leak, full stop, and `golangci-lint`'s `lostcancel` check (part of `govet`) catches the common cases in CI, but the discipline is the actual rule, the linter is the backstop.
- **A `context.Context` is always the first parameter, named `ctx`, never stored in a struct field** — storing a context on a struct (what `containedctx` flags) breaks the one property that makes context cancellation actually work: that it flows down through explicit call chains, not sideways through shared mutable state. `query-api`'s `RequestContext` (per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)) carries request-scoped *values* — actor identity, the request-scoped logger above — which is exactly what `context.Value` is for; it is not where cancellation semantics live, and it is threaded as a parameter through every handler, not stashed on a long-lived object.

### Graceful Shutdown

Every one of the six backend services is a long-running process managed by [Docker Compose](../12-infrastructure/README.md#docker-compose) or a [Kubernetes](../12-infrastructure/README.md#kubernetes) `Deployment`/`CronJob` — both send `SIGTERM` and wait a bounded grace period (Kubernetes' default is 30 seconds) before `SIGKILL`. A service that doesn't handle `SIGTERM` deliberately drops in-flight requests and, for `workers` specifically, risks the exact "the audit trail is what proves what happened" guarantee this doc set is built around if a retention-sweep or evidence-export is killed mid-operation rather than allowed to finish or cleanly abort.

**The pattern (Go 1.16+ standard library, no third-party dependency):**

- **Two separate contexts, not one** — a signal context that fires the moment `SIGTERM`/`SIGINT` arrives, and a bounded shutdown-timeout context derived from it. Using the already-canceled signal context to bound `Shutdown()` itself would abort immediately instead of draining in-flight work; the second context is what actually gives in-flight requests (or, in `workers`, the current retention-sweep record) time to finish.
- **`signal.NotifyContext`'s returned `stop()` function is always called** (typically via `defer`) — this restores default OS signal handling after the first signal, so an operator's second `Ctrl+C`/second `SIGTERM` forcibly terminates a shutdown that's stuck, rather than being silently absorbed by a handler that already fired once.
- **Every long-running loop — `ingestion`'s write loop, `alert-engine`'s consumer loop, `workers`' `RetentionSweepScheduler` — selects on the shutdown context, not just on its own work channel.** A loop that only checks its work source and never the shutdown signal is a goroutine leak (and, worse here, a process that never actually exits within Kubernetes' grace period, guaranteeing a `SIGKILL` mid-operation).
- **`workers` specifically: a `SIGTERM` arriving mid-sweep does not abort the *current* record's check-before-destroy sequence.** Per [Business Rules](../02-domain/README.md#business-rules) BR-2's ordering guarantee, `LegalHoldChecker` → `DeletionExecutor` for one record is treated as a short, uninterruptible unit — the shutdown context is checked *between* records, never partway through the hold-check-then-delete sequence for one, so a shutdown can never be the reason a deletion executes without its hold check having completed.

Sources: [Graceful Shutdown in Go: Practical Patterns — VictoriaMetrics](https://victoriametrics.com/blog/go-graceful-shutdown/), [Graceful Shutdown in Go: Patterns Every Production Service Needs](https://dev.to/young_gao/graceful-shutdown-in-go-patterns-every-production-service-needs-3l9c).

### HTTP Handlers and Middleware

`gateway`'s entire responsibility (authN/authZ, rate limiting, actor-context stamping, per [Bounded Contexts](../02-domain/README.md#bounded-contexts)) is expressed as a middleware chain over the standard library's `http.Handler` — **no web framework** (Gin/Echo/Fiber), consistent with this document's broader preference for standard-library-first, framework-only-when-the-standard-library-genuinely-can't (per [Structured Logging](README.md#structured-logging)'s identical reasoning for `log/slog` over a third-party logger).

- **A middleware is `func(http.Handler) http.Handler`** — wraps a handler, optionally acts before/after calling it, and the chain is composed once at startup, not per-request:
  ```go
  type Middleware func(http.Handler) http.Handler

  func Chain(h http.Handler, mw ...Middleware) http.Handler {
      for i := len(mw) - 1; i >= 0; i-- {
          h = mw[i](h)
      }
      return h
  }

  // gateway/main.go
  handler := Chain(routes,
      RequestIDMiddleware,   // outermost: assigns request_id first
      LoggingMiddleware,     // builds the request-scoped slog.Logger
      AuthMiddleware,        // resolves RequestContext's actor identity
      RateLimitMiddleware,   // per Redis, per Container Diagrams
  )
  ```
- **Order is load-bearing, not incidental** — `RequestIDMiddleware` and `LoggingMiddleware` must run before `AuthMiddleware` so an authentication failure is still logged with a request ID; `AuthMiddleware` must run before any handler that generates an `AccessAuditEvent`, since [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)'s `RequestContext` (actor identity, source IP) has to exist before `AuditedQueryHandler` can populate it.
- **Handlers are thin — they parse, call into the use-case/domain layer, and serialize the response.** Business logic (BR-1 through BR-7, masking decisions) never lives in an `http.HandlerFunc` body; per the Software Design System section below, a handler is an *adapter*, and an adapter's job is translation at the boundary, not decision-making.

### Database Access

**Postgres: [`sqlc`](https://sqlc.dev), not an ORM.** You write the SQL; `sqlc` generates typed Go from it at build time — no runtime reflection, no query-builder DSL standing between the code and the actual SQL running against `retention_policies`/`legal_holds`/`evidence_exports`. This is the direct extension of this document's core thesis (prefer a design where the wrong thing doesn't compile): a query or schema change that breaks a Go caller fails the build immediately, not at runtime in production. An ORM (GORM and similar) is a legitimate choice for CRUD-heavy, development-speed-first applications — genuinely not what this codebase is; every Postgres table here (per [Postgres Schema](../05-data/README.md#postgres-schema)) has compliance-relevant, precisely-specified query patterns (the composite `releases` key, the `legal_holds.scope` JSONB predicate), which is exactly the case where hand-written, generated-typed SQL beats an ORM's abstraction.

**ClickHouse: the official `clickhouse-go` client directly** — `sqlc` doesn't target ClickHouse, and the write/read shapes there (batch inserts, TTL-governed tables per [ClickHouse Schema](../05-data/README.md#clickhouse-schema)) are different enough from Postgres's transactional CRUD that a shared data-access abstraction across both stores would be forcing two genuinely different access patterns into one interface for no real benefit — consistent with [Postgres Schema](../05-data/README.md#postgres-schema)'s own "no cross-store foreign keys, no shared consistency model" decision.

**Transactions:** every multi-statement Postgres write that must be atomic (e.g., `alert-engine` creating an `ErrorGroup` row and updating `occurrence_count` together) wraps in `db.BeginTx(ctx, nil)` with `defer tx.Rollback()` immediately after — a no-op if `tx.Commit()` already succeeded, and the only way to guarantee a partial write never survives an early return or a panic recovery.

Sources: [sqlc vs GORM vs sqlx: Go Database Libraries Compared](https://reintech.io/blog/sqlc-vs-gorm-vs-sqlx-go-database-libraries-compared-2026), [Comparing database/sql, GORM, sqlx, and sqlc — JetBrains](https://blog.jetbrains.com/go/2023/04/27/comparing-db-packages/).

### JSON and API Contract Conventions

- **Struct tags carry the wire-format name; the Go identifier stays `MixedCaps` per the Naming Conventions above** — `SessionID string `json:"session_id"`` (snake_case on the wire, matching [Event Schema](../05-data/README.md#event-schema)'s and [REST API](../06-api/README.md#rest-api)'s existing field-naming convention; `MixedCaps` in Go).
- **`omitempty` only on fields that are genuinely optional in the domain, never as a way to hide a zero value that's actually meaningful** — per [Event Schema](../05-data/README.md#event-schema)'s conditionally-required fields (`reason` required only when `action = UNMASK`, `oldValue`/`newValue` required only when `action = CONFIG_CHANGED`), a field that's sometimes-required-sometimes-forbidden is modeled with an explicit `*string` (nil means absent, distinguishable from an empty string) plus request-level validation — never a bare `string` with `omitempty` silently swallowing the distinction between "not provided" and "empty."
- **A masked field never has a `json.Marshaler` that could accidentally serialize its real value.** Per [PII Redaction](../07-security/README.md#pii-redaction)'s two-tier model, a soft-masked field's Go representation on the read path is a placeholder-carrying type, not the raw string with a marshaling hook that "usually" masks it — the type itself should make an accidental unmasked serialization a compile-time impossibility, the same "make the wrong thing not compile" principle this document keeps returning to, applied to serialization instead of routing.
- **Every response body from `query-api`'s REST surface is a stable, versioned shape** — per [REST API](../06-api/README.md#rest-api)'s `/v1/` path versioning, a Go struct backing a `v1` response is never field-renamed or retyped in place; a breaking shape change is a new struct under a new path version, mirroring [Event Schema](../05-data/README.md#event-schema)'s additive-only discipline but with that document's explicitly *looser* rule (breaking changes allowed at a major version boundary, not never).

### Security-Specific Go Conventions

- **`crypto/rand`, never `math/rand`/`math/rand/v2`, for anything security-sensitive** — API token generation, session token generation, any value an attacker gaining knowledge of would matter. `math/rand`'s output is fully determined by its seed and is not safe for this regardless of how "random-looking" it appears; this is a `golangci-lint`(`gosec`)-checked rule, not left to memory. Non-security randomness (e.g., jitter on a retry backoff) is exactly where `math/rand/v2` is the right, faster choice — the distinction is the use, not a blanket rule against `math/rand` everywhere.
- **Every SQL statement is parameterized** (`sqlc`-generated queries are parameterized by construction, per Database Access above) — string-concatenated SQL does not exist anywhere in this codebase, including in `workers`' dynamically-scoped `LegalHold` predicate evaluation, where the structured-JSONB-predicate design in [Postgres Schema](../05-data/README.md#postgres-schema) exists specifically so a hold's `scope` is never interpolated into a raw query string.
- **Password hashing is Argon2id, exactly as specified in [Authentication](../07-security/README.md#authentication)** — this document doesn't re-derive that decision, it enforces it hasn't drifted: any code path that hashes a credential outside the one shared `internal/auth` package (per Package Boundaries below) is a review-blocking finding, not a style nitpick.
- **Comparing a secret value (a token, a webhook signature) uses `crypto/subtle.ConstantTimeCompare`, never `==` or `bytes.Equal`** — a variable-time comparison on a secret is a timing side-channel, however impractical to actually exploit over a network; the constant-time form costs nothing and removes the question entirely, which is the same "make the risk structurally absent, not just unlikely" bias this document applies everywhere else.

Sources: [Secure Randomness in Go 1.22 — The Go Programming Language](https://go.dev/blog/chacha8rand), [Math/rand random number generation is insecure — Datadog](https://docs.datadoghq.com/security/code_security/static_analysis/static_analysis_rules/go-security/math-rand-insecure/).

### Unit Test Style

**Table-driven tests as the default shape**, using `t.Run(tc.name, func(t *testing.T) {...})` per case so a failing row identifies itself by name in CI output rather than a bare line number — this is [the Go community's own settled convention](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) for exactly the reason [Testing](README.md#testing) below already leans on: BR-1 through BR-7's edge cases are naturally table-shaped (a matrix of category × regulatory basis × hold-state), not naturally expressed as one test function per case. This resolves the "specific test framework/runner choice" [Testing](README.md#testing) deliberately left open — the strategy (unit/integration/e2e pyramid, what belongs at each layer) is that document's job; which library and test shape implements the unit layer is this one's.

**Standard library `testing` plus `testify`'s `assert`/`require`, not a heavier framework** (Ginkgo/Gomega and similar BDD-style tools are a legitimate choice elsewhere, but add a second testing vocabulary and a slower feedback loop for no benefit this codebase's test shapes actually need). `require` (which stops the test on failure) for setup/precondition assertions where continuing would just cascade into confusing follow-on failures; `assert` (which continues and reports every failure) for the actual assertions under test, so one table-driven case reports all of its mismatches at once instead of only the first.

### TypeScript Coding Conventions (`browser-sdk`, `dashboard`)

> Everything above this point in [Coding Standards](README.md#coding-standards) governs the six Go backend services. This section is the equivalent depth for the two TypeScript surfaces — deliberately kept as one dense subsection rather than interleaved with the Go rules above, so a frontend contributor doesn't have to read backend-specific content to find their own conventions.

**Compiler baseline: `strict: true`, plus four flags the community has converged on as catching more than `strict` alone.** Per [TypeScript's own strict-mode documentation](https://www.typescriptlang.org/tsconfig/strict.html) and the current production-config consensus:

```jsonc
// tsconfig.json — shared base, extended per package
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,     // arr[i] is T | undefined, not T — catches the actual #1 source of "cannot read property of undefined"
    "exactOptionalPropertyTypes": true,   // `reason?: string` means "absent", not "present and undefined" — matters directly for UnmaskRequest's conditionally-required reason field
    "noPropertyAccessFromIndexSignature": true,
    "useUnknownInCatchVariables": true,   // a caught error is `unknown`, not `any` — forces a type check before use, same "narrow the type before trusting it" discipline as Go's comma-ok assertion
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "isolatedModules": true,
    "skipLibCheck": true
  }
}
```

**`typescript-eslint`'s `recommended-type-checked` + `stylistic-type-checked` configs**, not the non-type-checked `recommended` alone — the type-checked rules catch the class of bug this project's compliance surface can least afford (e.g. `no-floating-promises`, which would otherwise let an `await`-less `fetch` to `query-api` fail silently, exactly the kind of dropped-write that would undermine an audit-trail guarantee if it happened on a write path).

**`type` over `interface` by default; `interface` only for a public, extensible object contract (rare in this codebase) or a class contract.** Per the current community consensus, `type` is the only syntax that expresses union types, so any type modeling a set of states — most concretely, [Domain Model](../02-domain/README.md#domain-model)'s `AccessAuditEvent.action` enum, or a UI state that's genuinely one-of-several-shapes — has to be a discriminated union, which forces the choice for the type it's part of.

```typescript
// discriminated union — the TypeScript-side mirror of the Go `action` enum
type AccessAuditEvent =
  | { action: "VIEW"; targetRecordId: string }
  | { action: "UNMASK"; targetRecordId: string; reason: string } // reason only exists on this variant — not optional-on-everything
  | { action: "CONFIG_CHANGED"; targetRecordId: string; oldValue: string; newValue: string };

function describe(event: AccessAuditEvent): string {
  switch (event.action) {
    case "VIEW": return `Viewed ${event.targetRecordId}`;
    case "UNMASK": return `Unmasked ${event.targetRecordId}: ${event.reason}`;
    case "CONFIG_CHANGED": return `Changed ${event.targetRecordId}`;
    default: {
      const _exhaustive: never = event; // compile error if a variant is ever added and this switch isn't updated
      throw new Error(`unhandled action: ${JSON.stringify(_exhaustive)}`);
    }
  }
}
```

This is the direct TypeScript-side expression of this document's recurring thesis (make the wrong thing not compile) — the `never`-typed exhaustiveness check means adding a new `action` variant to [Event Schema](../05-data/README.md#event-schema) without updating every switch over it is a compile error in `dashboard`, not a runtime gap discovered later.

**`unknown`, never `any`, at any boundary where a value's shape isn't yet proven** — an API response from `query-api` is `unknown` until it's validated (see below), a caught error is `unknown` (per `useUnknownInCatchVariables` above), and a narrowing check (a type guard, a schema parse) is what turns it into a trusted type. `any` anywhere in this codebase is a `typescript-eslint`(`no-explicit-any`)-flagged finding, not a style nitpick.

**Runtime schema validation at every network boundary, using [Zod](https://zod.dev)** — TypeScript's type system is erased at compile time, so a `query-api` response typed as `Session` without a runtime check is trusting the network, not verifying it. Every fetch from `dashboard` and every event `browser-sdk` accepts from host-page config parses through a Zod schema before the rest of the code ever sees it as a typed value — the same "verify the boundary, trust the interior" principle [Software Design System](README.md#software-design-system)'s Error Contracts section already applies to the Go backend's API responses.

```typescript
const SessionSchema = z.object({
  sessionId: z.string().uuid(),
  environment: z.string(),
  releaseId: z.string(),
});
type Session = z.infer<typeof SessionSchema>; // the type is derived from the schema, not hand-duplicated

async function fetchSession(id: string): Promise<Session> {
  const res = await fetch(`/v1/sessions/${id}`);
  return SessionSchema.parse(await res.json()); // throws a descriptive error on shape mismatch, never silently trusts it
}
```

**Testing: [Vitest](https://vitest.dev), not Jest, for both `browser-sdk` and `dashboard`.** Native ESM, Vite-native (so `dashboard`'s dev and test configuration share one transform pipeline instead of two), and materially faster on this codebase's realistic size — the same "prefer the standard-adjacent, lower-ceremony tool" bias [Coding Standards](README.md#coding-standards)' choice of stdlib `testing`+`testify` over a BDD framework already applies on the Go side. [React Testing Library](https://testing-library.com/react) for `dashboard` component tests, querying by role/label the way a user actually interacts with a screen — never by CSS class or a `data-testid` when an accessible query exists, since an accessible query is also, incidentally, the closest thing to an automated accessibility check this codebase gets for free.

Sources: [TypeScript strict mode: the 6 tsconfig options that actually matter in production](https://dev.to/jtorchia/typescript-strict-mode-the-6-tsconfig-options-that-actually-matter-in-production-and-when-to-446d), [typescript-eslint: Linting with Type Information](https://typescript-eslint.io/getting-started/typed-linting/), [Types vs. interfaces in TypeScript — LogRocket](https://blog.logrocket.com/types-vs-interfaces-typescript/), [Vitest vs Jest — Speakeasy](https://www.speakeasy.com/blog/vitest-vs-jest/).

### What's Deliberately Not Modeled Here

- File organization *within* a single package beyond the Package Boundaries rules in the Software Design System section below — genuinely team-calibrated, evolves with the codebase.
- Comment/documentation-string conventions beyond standard `godoc` form (exported identifiers get a doc comment starting with the identifier's name) — downstream of whatever the team finds actually gets maintained versus goes stale.
- IDE/editor tooling setup (gopls configuration, save-on-format hooks) — a developer-environment convenience, not a codebase-wide standard.

---

## Software Design System

> Status: The architectural counterpart to [Coding Standards](README.md#coding-standards) above — not "how to write a line of Go," but "how a package, a service boundary, and a dependency graph are supposed to be shaped." Every pattern below is either already load-bearing somewhere else in this doc set (named explicitly, not just implied) or a direct consequence of a constraint established elsewhere — this document doesn't introduce new architectural philosophy, it names the one this project has already been building to and makes it explicit enough to hold a new contributor to.

---

### The Governing Principle: Hexagonal, With the Domain at the Center

Per the [ports-and-adapters / hexagonal pattern](https://threedots.tech/post/introducing-clean-architecture/) — business logic sits at the center, knows nothing about HTTP, SQL, or ClickHouse, and infrastructure depends on the domain, never the reverse. This is not a new decision for this project — it's the explicit name for what [Domain Model](../02-domain/README.md#domain-model) through [Component Diagrams](../03-architecture/diagrams.md#component-diagrams) already built:

- **The domain layer** (`packages/domain-types`, per [Repository Structure](../03-architecture/README.md#repository-structure)) — `Session`, `AccessAuditEvent`, `RetentionPolicy`, `LegalHold`, the five invariants from [Domain Model](../02-domain/README.md#domain-model) — has zero import-time dependency on ClickHouse, Postgres, or any HTTP framework. It is pure Go types and the business rules that govern them (BR-1 through BR-7).
- **Ports** are the interfaces the domain/use-case layer declares for what it needs from the outside world — per the Interface Design rules above, these are consumer-declared, narrow, and named for what they do (`LegalHoldStore`, not `PostgresClient`).
- **Adapters** are the concrete implementations — a `PostgresLegalHoldStore`, a `ClickHouseAccessAuditWriter` — that satisfy a port and live in each service's own package, never in `packages/domain-types`.

**The concrete, already-built proof this isn't aspirational:** `query-api`'s `AuditedQueryHandler` (per [Component Diagrams](../03-architecture/diagrams.md#component-diagrams)) is a use-case-layer wrapper type sitting between the HTTP adapter (`gateway`'s routing) and the domain — it is *why* [Business Rules](../02-domain/README.md#business-rules) BR-5 ("every sensitive read produces exactly one audit event") is a compile-time-enforced property instead of a convention a future engineer could accidentally skip. That pattern — **make the correct dependency direction the only one the type system allows, not the one a code reviewer has to remember to check** — is the design system's actual thesis, restated in one sentence: prefer a design where the wrong thing doesn't compile over a design where the wrong thing merely gets caught in review.

### Package Boundaries

Per [the standard Go project-layout convention](https://github.com/golang-standards/project-layout) and [Repository Structure](../03-architecture/README.md#repository-structure)'s monorepo layout, applied inside each service:

- **`internal/`** holds everything specific to one bounded context — `services/query-api/internal/domain`, `internal/handlers`, `internal/store` — and is compiler-enforced unimportable from outside that module, which is the actual mechanism (not just a naming convention) that keeps `dashboard` from ever reaching into `query-api`'s internals instead of calling its published API, per [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s Conformist relationship.
- **`packages/`** (already established in [Repository Structure](../03-architecture/README.md#repository-structure)) is the *only* cross-service-import surface — `domain-types` and `event-schemas` — and nothing else crosses a service boundary as a Go import. A `query-api` internal package is never imported directly by `alert-engine`; if two services need to share logic beyond the domain types, that logic either belongs in `packages/` (if it's genuinely domain-level) or the services talk to each other over an API/event, not a shared internal Go package.
- **Flat within a service, one or two levels deep** — `internal/domain`, `internal/handlers`, `internal/store`, not `internal/domain/entities/session/v2/types`. Per the research consensus above, deep nesting is a common Go anti-pattern that adds navigation cost without adding real structure; a service small enough to fit in one bounded context (per [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s eight-context split, each already scoped narrowly) rarely needs more than that.

### The Repository Pattern, Applied to This Project's Two-Store Split

A repository is a port (per the Governing Principle above) that gives the domain/use-case layer record-level access to a store, without that layer knowing which store or query technology backs it. This project's dual-store architecture (per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams)) means a repository interface is often backed by different technology depending on which entity it serves, which is precisely the case this pattern exists to hide from callers:

- `workers`' `LegalHoldStore` port is satisfied by a `sqlc`-generated Postgres adapter (per Database Access above) — `legal_holds` is small-cardinality, transactional data.
- `query-api`'s `AccessAuditEventReader`/`Writer` ports are satisfied by a `clickhouse-go`-backed adapter — high-volume, append-only, per [ClickHouse Schema](../05-data/README.md#clickhouse-schema).
- Neither consuming component (`LegalHoldChecker`, `AuditedQueryHandler`) imports `sqlc`'s generated package or `clickhouse-go` directly — they depend on the port interface declared in their own package, per the Interface Design rules in [Coding Standards](README.md#coding-standards). Swapping the ClickHouse client library, or moving `legal_holds` to a different store at some future scale, is contained to one adapter file per port, never a change visible to the domain layer.
- **A repository returns domain types (`LegalHold`, `AccessAuditEvent` from `packages/domain-types`), never a store-specific row type** (`sqlc`'s generated struct, a raw ClickHouse driver row) — the translation from store shape to domain shape happens inside the adapter, which is the actual boundary the Governing Principle's dependency-direction rule is drawn at.

### Dependency Injection: Explicit Constructors, No Framework

**Manual, constructor-based wiring in each service's `main.go` — no `wire`, no `fx`, no reflection-based DI container.** This is a deliberate rejection of the "senior teams use a DI framework" instinct, for a reason specific to this project rather than a generic preference:

- [Deployment Model](../03-architecture/README.md#deployment-model)'s Operational Simplicity requirement means every piece of runtime "magic" a 2–3 person team has to understand to debug a startup failure is a real cost. A stack trace through `fx`'s reflection-based container is strictly harder to read at 2 a.m. than a stack trace through a plain function call in `main.go` — and per [Onboarding](../09-workflows/README.md#onboarding)'s "no manual intervention between steps" requirement, startup-time failures need to be immediately legible.
- Each service's dependency graph is small and static (per [Container Diagrams](../03-architecture/diagrams.md#container-diagrams), no service depends on more than a handful of stores/clients) — exactly the case the research above identifies as "wiring fits on one screen, manual DI is the right choice," not the large-dynamic-graph case `wire`/`fx` exist for.
- `wire`'s compile-time code generation would be a defensible middle ground if a service's graph grew large enough that hand-wiring became error-prone — worth revisiting *if* that happens, not adopted pre-emptively against a problem this codebase doesn't have yet.

### Service Bootstrap and Lifecycle

Every one of the six backend services' `main.go` follows the same shape, so a developer moving from `workers` to `ingestion` finds the entry point structured identically — wiring (this section), configuration, and the Graceful Shutdown pattern from [Coding Standards](README.md#coding-standards) are not each service's independent invention:

1. **Load and validate configuration** (environment variables — per [Docker Compose](../12-infrastructure/README.md#docker-compose)'s secrets-as-mounted-files convention, nothing sensitive as a bare env var) into a typed config struct, failing fast with a specific, actionable error if a required value is missing — this is the "startup failure must be immediately legible" requirement from Dependency Injection above, applied to configuration specifically.
2. **Construct adapters** (repository implementations, the `slog.Logger`) — each adapter's constructor takes only the config values it actually needs, never the whole config struct, per the Interface Design "narrow, consumer-declared" bias.
3. **Construct the domain/use-case layer**, injecting the adapters constructed in step 2 through their port interfaces — this is the one place in the codebase where a concrete adapter type and its port interface are both visible in the same function, by necessity.
4. **Wire the delivery mechanism** — `gateway`'s HTTP middleware chain (per [Coding Standards](README.md#coding-standards)), or `workers`' `CronJob`/on-demand split (per [Kubernetes](../12-infrastructure/README.md#kubernetes)) — around the use-case layer from step 3.
5. **Start, then block on the Graceful Shutdown signal context** from [Coding Standards](README.md#coding-standards) — `main()`'s last responsibility, and the only place `log.Fatal`/`os.Exit` is permitted to appear, per that section's rule.

### Error Contracts Across a Service Boundary

Per the Error Handling section above, a wrapped internal error (`%w`) is appropriate *within* a service. **At a true boundary — `query-api`'s HTTP responses, any event a webhook carries — internal errors are translated into a small, stable set of API-level error codes, never leaked as internal type names or wrapped database errors.** This is the same "boundary" the [Webhooks](../06-api/README.md#webhooks) document already applies to payload *content* ("metadata and references only, never sensitive content inline") — applied here to error *detail* instead: a `pgconn.PgError` string leaking into a JSON error response is an information-disclosure surface (schema/implementation detail exposed to a caller) for the same reason an unmasked PHI field would be, just a lower-stakes version of it.

### Idempotency, Wherever an Operation Can Legitimately Run Twice

A `SIGKILL` after Graceful Shutdown's grace period expires, a Kubernetes pod reschedule mid-operation, or a retry after a transient database error can all cause the same logical operation to execute more than once. Per this document's core thesis, the fix is structural, not "retry logic tries to be careful":

- **`EvidenceExportGenerator` is naturally idempotent by construction** — per [Storage](../05-data/README.md#storage), an export's `contentHash` is computed from its frozen input set, so generating "the same" export twice produces the same hash; a retry after a partial failure re-runs safely rather than producing a second, divergent export under the same `exportId`. The `exportId` itself should be caller-supplied (or derived deterministically from the request, e.g. a hash of `incidentReference` + requested record set), not server-generated per attempt, specifically so a client's retry after a timeout lands on the same operation instead of creating a duplicate.
- **`RetentionSweepScheduler`'s per-record processing is naturally idempotent** — per [Business Rules](../02-domain/README.md#business-rules) BR-2, the hold-check-then-delete sequence for one record, re-run against a record that's already been deleted, is a no-op (the record's absence *is* the terminal state); this is why the Graceful Shutdown section above only needs to protect the sequence for the *current* record, not the sweep as a whole — a resumed sweep simply re-evaluates records from the top, and already-processed ones cost a cheap existence check, not incorrect double-processing.
- **`ingestion`'s writes are not naturally idempotent, and are not made so** — a duplicate `SessionEventCaptured` delivery (e.g. a client retry after a timeout that actually succeeded) produces a duplicate row in [ClickHouse Schema](../05-data/README.md#clickhouse-schema)'s append-only tables. This is a deliberate, accepted tradeoff rather than a gap: de-duplicating high-volume event ingestion (via an idempotency-key table, exactly-once delivery semantics, or similar) is real engineering cost for a failure mode (an occasional duplicate replay frame) that doesn't threaten the audit-trail guarantee the way a duplicated or dropped `AccessAuditEvent` would — the compliance-critical writes (audit events, evidence exports, retention deletions) get idempotency; the high-volume telemetry writes don't, because the cost/benefit genuinely differs between them.

### TypeScript Design System (`browser-sdk` and `dashboard`)

Two genuinely different architectural shapes, sharing the compiler/testing conventions from [Coding Standards](README.md#coding-standards)' TypeScript section but not the same structure — `browser-sdk` is a library embedded in someone else's page; `dashboard` is Imora's own product surface.

**`browser-sdk`: framework-agnostic core, thin wrappers, no runtime dependencies beyond what the bundle-size budget already accounts for.** Already specified at the product level in [SDK API](../06-api/README.md#sdk-api) — this section fixes the *build* side of that decision. **[`tsup`](https://tsup.egoist.dev)**, not Rollup or a hand-rolled esbuild config: it wraps esbuild with sensible library-authoring defaults (dual ESM+CJS output, generated `.d.ts` files) and needs close to zero configuration to publish cleanly, matching this document's general bias toward the lower-ceremony tool wherever the ceremony doesn't buy something this project actually needs. Rollup remains the fallback if `browser-sdk`'s bundle-size budget from [SDK API](../06-api/README.md#sdk-api) is ever missed on a `tsup` build and its more thorough tree-shaking is the difference — not the default, a targeted escalation.

```
sdk/browser-sdk/
├── packages/
│   ├── core/            # @imora/core — capture, masking, transport. Zero framework dependency.
│   ├── react/            # @imora/react — thin wrapper, re-exports core's full API
│   ├── vue/               # @imora/vue
│   └── angular/           # @imora/angular
```

Each framework package's only job is a lifecycle adapter (React error boundary, Vue's `errorCaptured`) around `@imora/core` — per [SDK API](../06-api/README.md#sdk-api), no capability exists in one wrapper and not another, so a wrapper that accumulates its own logic beyond the lifecycle hook is a design smell, not a feature.

**`dashboard`: [TanStack Start](https://tanstack.com/start), TanStack Query for server state, Zustand for client-only UI state.** TanStack Start is a full-stack React framework (SSR, file-based type-safe routing via TanStack Router, server functions) built by the same team as TanStack Query — the same "shares one ecosystem's mental model" reasoning that already justified React over Vue/Angular for `dashboard` (per `browser-sdk` shipping a first-class React wrapper) extends one level deeper: `dashboard`'s router, data-loading, and server-state-cache layers are one integrated stack, not three separately-chosen libraries wired together by hand. It also fits this project's deployment constraint better than the leading alternative: TanStack Start's Nitro-based build deploys to plain Node in Docker with no platform-specific dependency, where Next.js's feature set is comparatively Vercel-optimized — consistent with [System Context](../03-architecture/README.md#system-context)'s "no Parity or Wedge capability may depend on an external system" constraint, applied here to the build tooling rather than a product feature.

- **Route structure mirrors [`docs/user-stories.md`](../../docs/user-stories.md)'s flow boundaries directly** — TanStack Router's file-based routes (`routes/sessions.search.tsx`, `routes/legal-holds.tsx`) are close enough to a 1:1 mapping with that document's Flow B, Flow G, etc. that a contributor can go from "which flow" to "which file" without an intermediate folder-naming decision.
- **Server functions and route loaders are a proxy to `query-api`'s REST API — never a direct store connection.** This is the one genuinely new risk SSR introduces that a static SPA never had: a server function *can* import a database driver and query Postgres/ClickHouse directly, which would silently reintroduce a second, unaudited data-access path around `query-api`'s `AuditedQueryHandler` — exactly the failure mode [Bounded Contexts](../02-domain/README.md#bounded-contexts)'s Conformist relationship exists to prevent. The rule is structural intent, not (yet) a compiler-enforced one: every server function and loader body's only external call is an HTTP request to `query-api`, identical in shape to what client-side code would make, just executing on the server for the initial render.
- **Server state (anything that ultimately comes from `query-api`) lives in TanStack Query, never mirrored into Zustand/Context.** Conflating server state and client state is what produces stale-cache bugs and unnecessary re-render cascades, per the current consensus. The environment selector's *current value* is client state (Zustand, since it has to persist across navigation per [`docs/user-stories.md`](../../docs/user-stories.md)'s global interaction pattern); the *session list for that environment* is server state (TanStack Query, keyed on `["sessions", environment, ...filters]`, prefetched in a route loader via `queryClient.ensureQueryData` for the SSR path and re-used client-side on navigation — one cache, populated from two different execution contexts, not two separate caches to keep in sync).
- **Zustand, not Redux, for the client-state slice that remains** — this codebase's actual client-only state (environment selector, modal open/closed, in-progress form values before submission) is exactly the small-to-medium, minimal-boilerplate case the current research consensus assigns to Zustand rather than Redux Toolkit; Redux's justification (large team, complex normalized state, time-travel debugging) doesn't describe this project's actual shape.

Sources: [tsup vs Rollup vs esbuild 2026](https://www.pkgpulse.com/guides/tsup-vs-rollup-vs-esbuild-2026), [TanStack Start Overview](https://tanstack.com/start/latest/docs/framework/react/overview), [Server Functions — TanStack Start docs](https://tanstack.com/start/latest/docs/framework/react/guide/server-functions), [TanStack Router — type-safe loaders and TanStack Query integration](https://tanstack.com/router/latest/docs/overview), [TanStack Start vs Next.js 2026](https://www.alexcloudstar.com/blog/tanstack-start-vs-nextjs-2026/), [TanStack Query in 2026](https://blog.codercops.com/blog/tanstack-query-server-state-2026), [Redux vs Zustand vs Context API in 2026](https://medium.com/@sparklewebhelp/redux-vs-zustand-vs-context-api-in-2026-7f90a2dc3439).

### What This Feeds

[`docs/coding-standards.md`](../../docs/coding-standards.md) and [`docs/design-system.md`](../../docs/design-system.md) restate this and the section above at build-ready altitude — the rules an engineer actually follows day to day, without re-deriving the reasoning each time.

---

## Testing

> Status: Specifies the compliance-guarantee test suite [CI/CD](../12-infrastructure/README.md#cicd) and [Release Process](README.md#release-process) both referenced without defining, and turns [Event Schema](../05-data/README.md#event-schema)'s "additive-only, forever" rule into an actual CI gate rather than a convention.

---

### The Finding: Schema Compatibility Can Be a CI Gate, Not Just a Rule Someone Has to Remember

[Event Schema](../05-data/README.md#event-schema) states that schema evolution must never rename, retype, or remove an existing field — stated as policy, twice now, but nothing yet checks it automatically. Real, existing tooling closes this gap directly: a JSON-schema diff check (`json-schema-diff-validator` or equivalent) runs in CI on every change to `README.md#event-schema`'s field definitions, and **the build fails on any breaking change**, not just a naming-convention violation. [REST API](../06-api/README.md#rest-api)'s OpenAPI spec gets the equivalent check via `oasdiff` — but with deliberately different strictness, matching that document's own looser rule: breaking changes are *allowed* there, gated on a major version bump and deprecation notice, whereas the event schema check allows no breaking changes ever, full stop. Same tooling category, two different policies, because the two documents specify two different compatibility guarantees for two different reasons (7-year-old stored records vs. a request/response contract with real clients who can migrate).

---

### Test Pyramid

- **Unit** — business-rule logic in isolation: BR-1's longest-clock resolution, BR-3's selective-purge field-level decision, [PII Redaction](../07-security/README.md#pii-redaction)'s three-input classification logic. Fast, no real infrastructure.
- **Integration** — against real infrastructure, deliberately not mocked, because mocking would defeat the point of verifying actual enforcement:
  - The DB-level GRANT restrictions from [Threat Model](../07-security/README.md#threat-model) — attempt `DELETE`/`ALTER` against `access_audit_events` using the `ingestion`/`query-api` service account and assert it fails.
  - MinIO Object Lock in Compliance mode — attempt to delete a locked EvidenceExport blob, including with elevated credentials, and assert it fails, per [Storage](../05-data/README.md#storage).
  - The legal-hold check-before-destroy ordering (BR-2) — apply a hold mid-sweep and assert records not yet processed are protected while already-deleted records aren't retroactively flagged as a bug, per [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) Flow C.
  - `CONFIG_CHANGED` firing correctly on RetentionPolicy/role/field-classification changes, with `oldValue`/`newValue` populated, per [Audit Logging](../07-security/README.md#audit-logging).
  - Container conventions from [Docker](../12-infrastructure/README.md#docker) — non-root user, read-only root filesystem, verified against the actual built image, not asserted in a Dockerfile comment.
- **End-to-end** — the four flows in [Sequence Diagrams](../03-architecture/diagrams.md#sequence-diagrams) (Session Capture, DSAR Query, Retention Sweep Hitting a Legal Hold, Evidence Export Generation) **are the e2e test spec already**, not a separate scenario set to invent — that document was written at exactly the right altitude to double as test-case documentation.

---

### What's Deliberately Not Modeled Here

- Test framework/runner choice — resolved in [Coding Standards](README.md#coding-standards)' Unit Test Style section (standard library `testing` + `testify`, table-driven) — this document's job is the pyramid strategy above, not the tooling.
- Coverage percentage targets — a team-calibrated number, not an architectural constraint.
- Load/performance testing methodology — downstream of [Scaling](../03-architecture/README.md#scaling)'s thresholds once real traffic data exists to test against.

### What This Feeds Next

`research/11-engineering/README.md#branching-strategy` and `README.md#coding-standards` round out `11-engineering/`.

---

## Release Process

> Status: How a change becomes a version customers actually run, across both distribution paths from [CI/CD](../12-infrastructure/README.md#cicd).

---

### Versioning

Semantic versioning (MAJOR.MINOR.PATCH), independent of two other version-like concepts already defined elsewhere that are easy to conflate with it: [Event Schema](../05-data/README.md#event-schema)'s per-event `schemaVersion` (which only ever increments additively, per that document's governing rule) and [REST API](../06-api/README.md#rest-api)'s `/v1/` API path version (which can break at a major boundary with migration notice). A software MAJOR release does not imply an event-schema or API-version bump, and vice versa — three independent axes, not one number wearing three hats.

### Pipeline

1. Version bump, changelog entry appended to the root `CHANGELOG.md`.
2. Build, test (including the compliance-guarantee suite from [CI/CD](../12-infrastructure/README.md#cicd) — the GRANT-restriction and non-root-container assertions), SBOM generation, key-pair signing (per ADR [0007](architecture-decisions/0007-keypair-signing-over-keyless.md)).
3. **Dual publish**, from the identical signed artifact: registry push for connected deployments; signed bundle packaging for air-gapped transfer, per [Deployment Model](../03-architecture/README.md#deployment-model).
4. Git tag.

---

### The Gap: How Does an Air-Gapped Customer Even Learn a Release Exists?

Every prior document addressing air-gapped updates assumed the customer already knows they need one and initiates the transfer. Nothing yet addresses the step before that — **Imora has no path to notify an air-gapped customer that a release, especially a critical security patch, exists at all**, for exactly the same reason it can't push the update directly: no reachability into their network. This has to be an out-of-band channel, not a product feature: release notes and security advisories published somewhere a Platform Operator is expected to check periodically (a security mailing list, a published advisory page, an email to a registered contact) — never a webhook, never a phone-home version check, since both would violate the same air-gapped constraint this entire release mechanism was built around. Connected deployments can additionally get an in-product "update available" notice through the registry-pull path; air-gapped ones fundamentally cannot, and no future design should try to route around that by weakening the air-gap for the sake of update convenience.

---

### What's Deliberately Not Modeled Here

- Exact changelog format/automation tooling — implementation detail.
- Security advisory publication mechanics (CVE assignment, disclosure timeline) — a security-process concern downstream of this document, not part of the release pipeline itself.

### What This Feeds Next

`research/11-engineering/README.md#testing` should specify the compliance-guarantee test suite referenced above in full. `research/11-engineering/README.md#branching-strategy` and `README.md#coding-standards` round out this folder.

