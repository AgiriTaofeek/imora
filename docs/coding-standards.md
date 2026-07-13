# Imora — Coding Standards

> Build-ready rules for both language surfaces: **Part A** covers the six Go backend services (`gateway`, `ingestion`, `query-api`, `alert-engine`, `workers`, `notification-service`); **Part B** covers the two TypeScript surfaces (`browser-sdk`, `dashboard`). The reasoning, sourcing, and how each rule ties back to this project's actual constraints lives in [`research/11-engineering/README.md#coding-standards`](../research/11-engineering/README.md#coding-standards) — this file is the checklist and the snippets you actually work from, not the argument for why.
>
> Baseline: **Go 1.24+**, **TypeScript 5.6+**.

---

# Part A — Go (Backend Services)

## 1. Formatting & Linting (CI-gated, not optional)

- `gofmt` — every diff must be `gofmt`-clean. This fails the build, not code review.
- `golangci-lint`, one root config shared by all six services. Enabled: `errcheck`, `govet`, `staticcheck`, `unused`, `ineffassign` (default set) + `bodyclose`, `contextcheck`, `containedctx`, `wrapcheck`, `gosec`, `lostcancel` (via `govet`).
- No per-service lint config drift. If a service needs an exception, it's a documented, reviewed override in the shared config — not a local `.golangci.yml`.

```yaml
# .golangci.yml — one file, root of the monorepo
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - ineffassign
    - bodyclose
    - contextcheck
    - containedctx
    - wrapcheck
    - gosec

linters-settings:
  wrapcheck:
    ignoreSigs:
      - context.Canceled
      - context.DeadlineExceeded
```

## 2. Naming Conventions

| Rule | Example |
|---|---|
| `MixedCaps`, never underscores | `RetentionPolicy`, not `retention_policy` |
| Initialisms: one consistent case | `SessionID`, `ClickHouseURL`, `LegalHoldID` — never `SessionId`, `ClickhouseUrl` |
| Receiver: short, consistent per type | `func (h *AuditedQueryHandler) ServeHTTP(...)` — `h` on every method of that type |
| Package names: short, lowercase, singular | `store`, `auth` — never `stores`, `authentication_helpers` |
| Domain vocabulary carries through unchanged | `Session`, `AccessAuditEvent`, `LegalHold` are the Go type names verbatim — no code-layer rename of a term already fixed in [Domain Model](../research/02-domain/README.md#domain-model) |

## 3. Error Handling

| Rule | Do | Don't |
|---|---|---|
| Handle once | Handle an error at the one layer with enough context to act on it | `log.Error(err); return err` — double-logs the same failure |
| Wrap at internal boundaries | `fmt.Errorf("checking legal hold: %w", err)` when a caller might `errors.Is`/`errors.As` through it | Wrap with `%w` at a true external boundary (leaks internals into an API response) |
| Sentinel errors | `var ErrNotFound = errors.New(...)` for conditions callers branch on | A dynamically formatted string a caller has to substring-match |
| Batch failures | `errors.Join` when one operation legitimately accumulates several independent errors (e.g. a retention-sweep batch) | A hand-rolled multi-error slice type |
| Panics | Never in reachable request/business logic — return an error | `panic` as a substitute for an error return |
| Type assertions | Always `v, ok := x.(T)` | Bare `v := x.(T)` (panics on mismatch; `staticcheck`-flagged) |
| Fatal exits | `log.Fatal`/`os.Exit` at most once, only in `main()` | Any other function terminating the process directly |

```go
var ErrLegalHoldActive = errors.New("legal hold active for target record")

// internal boundary — wrap, caller can errors.Is through it
func (c *legalHoldChecker) Check(ctx context.Context, recordID uuid.UUID) error {
    held, err := c.store.IsHeld(ctx, recordID)
    if err != nil {
        return fmt.Errorf("checking legal hold for %s: %w", recordID, err)
    }
    if held {
        return ErrLegalHoldActive
    }
    return nil
}

// caller branches on the sentinel
if err := checker.Check(ctx, id); err != nil {
    if errors.Is(err, ErrLegalHoldActive) {
        return workers.SkipDeletion(ctx, id) // BR-2's DeletionSkippedDueToHold path
    }
    return fmt.Errorf("retention sweep: %w", err) // unexpected — propagate
}
```

```go
// batch failures — retention sweep continues past one bad record
func (s *RetentionSweepScheduler) sweep(ctx context.Context, records []Record) error {
    var errs error
    for _, r := range records {
        if err := s.processOne(ctx, r); err != nil {
            errs = errors.Join(errs, fmt.Errorf("record %s: %w", r.ID, err))
            continue // one failure doesn't stop the sweep
        }
    }
    return errs // nil if every record succeeded; errors.Is/As still work through it
}
```

## 4. Structured Logging (`log/slog`, stdlib, no third-party logger)

```go
// gateway — one request-scoped logger, built once, carried via context
func LoggingMiddleware(base *slog.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger := base.With(
                "request_id", middleware.RequestIDFromContext(r.Context()),
                "environment", os.Getenv("IMORA_ENVIRONMENT"),
            )
            ctx := logctx.WithLogger(r.Context(), logger) // e.g. via a small ctx-key helper
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// deep in a call chain — pull the logger back out, never a new one
func (h *AuditedQueryHandler) recordView(ctx context.Context, sessionID uuid.UUID) {
    logger := logctx.From(ctx) // already carries request_id, actor_user_id, environment
    if err := h.auditWriter.Write(ctx, SessionViewedEvent(sessionID)); err != nil {
        logger.Error("audit write failed", "session_id", sessionID, "error", err)
    }
}
```

- Message = what happened. Attributes = the context — never a value string-formatted into the message.
- `slog.LogAttrs` only in genuinely hot paths (`ingestion`'s per-event write path). Everywhere else, the plain `Info("msg", "k", v)` form.
- **Never log anything [PII Redaction](../research/07-security/README.md#pii-redaction) classifies as soft-masked-or-harder.** Log the record ID, never its content.
- Operational logs (this section) and `AccessAuditEvent`s ([Audit Logging](../research/07-security/README.md#audit-logging)) are different systems — never merge them.

## 5. Interfaces & Constructors

```go
// consumer-declared, narrow, at point of use — not imported from a producer package
type legalHoldStore interface {
    IsHeld(ctx context.Context, recordID uuid.UUID) (bool, error)
}

type legalHoldChecker struct {
    store legalHoldStore // accepts the interface
}

func NewLegalHoldChecker(store legalHoldStore) *legalHoldChecker { // returns the concrete struct
    return &legalHoldChecker{store: store}
}
```

```go
// functional options — 3+ optional config values
type SweepOption func(*RetentionSweepScheduler)

func WithInterval(d time.Duration) SweepOption {
    return func(s *RetentionSweepScheduler) { s.interval = d }
}
func WithClock(c clock.Clock) SweepOption { // injectable in tests
    return func(s *RetentionSweepScheduler) { s.clock = c }
}

func NewRetentionSweepScheduler(store legalHoldStore, opts ...SweepOption) *RetentionSweepScheduler {
    s := &RetentionSweepScheduler{store: store, interval: 24 * time.Hour, clock: clock.Real{}}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

- **Accept interfaces, return structs. Interfaces stay 1–3 methods.** More than that is usually two responsibilities.

## 6. Generics

```go
// legitimate: identical shape across Session/ErrorGroup/AccessAuditEvent list endpoints
type Page[T any] struct {
    Items      []T
    NextCursor string
}

func Paginate[T any](items []T, cursor string, pageSize int) (Page[T], error) {
    // cursor decode, slice, encode next cursor — identical for every T
}
```

Don't reach for `T any` on a function with one call site "for flexibility" — that's solving a problem that doesn't exist yet.

## 7. Concurrency

```go
// errgroup, not bare WaitGroup — first error wins, shared context cancels the rest
func (s *RetentionSweepScheduler) sweepBatch(ctx context.Context, records []Record) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // bounded fan-out — a legal-hold scope can resolve to thousands of records

    for _, r := range records {
        g.Go(func() error {
            return s.processOne(ctx, r) // per-iteration loop var, safe since Go 1.22
        })
    }
    return g.Wait()
}
```

- Every `context.WithCancel`/`WithTimeout`/`WithDeadline` paired with `defer cancel()` at the same call site.
- `ctx context.Context` always the first parameter, named `ctx`. Never stored on a struct field (`containedctx`-checked).

## 8. Graceful Shutdown

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop() // restores default signal handling — a 2nd SIGTERM force-kills a stuck shutdown

    srv := buildServer(ctx) // wiring per §12 Service Bootstrap
    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            logger.Error("server error", "error", err)
        }
    }()

    <-ctx.Done() // blocks until SIGTERM/SIGINT
    stop()

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second) // under k8s's 30s grace period
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Error("shutdown did not complete cleanly", "error", err)
    }
}
```

- `workers`' retention sweep: the shutdown context is checked **between records**, never mid-sequence inside one record's hold-check-then-delete — a shutdown can never be the reason a deletion executes without its hold check completing.

## 9. HTTP Handlers and Middleware

No web framework — `net/http` plus a middleware chain.

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mw ...Middleware) http.Handler {
    for i := len(mw) - 1; i >= 0; i-- {
        h = mw[i](h)
    }
    return h
}

// gateway/main.go — order is load-bearing
handler := Chain(routes,
    RequestIDMiddleware,   // outermost: assigns request_id first
    LoggingMiddleware,     // builds the request-scoped slog.Logger
    AuthMiddleware,        // resolves RequestContext's actor identity — must precede audit-writing handlers
    RateLimitMiddleware,   // per Redis, per Container Diagrams
)
```

Handlers parse, call the use-case layer, serialize the response. Business logic never lives in an `http.HandlerFunc` body — a handler is an adapter (see [`design-system.md`](design-system.md)).

## 10. Database Access

- **Postgres: [`sqlc`](https://sqlc.dev), not an ORM.** Write SQL, get typed Go at build time — no runtime reflection, no query-builder DSL.
- **ClickHouse: the official `clickhouse-go` client directly.** `sqlc` doesn't target it; the batch-insert/TTL-governed access pattern is different enough from Postgres CRUD that forcing a shared abstraction across both stores would cost more than it saves.

```go
// sqlc-generated (Postgres) — a query/schema mismatch fails the build, not production
func (r *postgresLegalHoldStore) IsHeld(ctx context.Context, recordID uuid.UUID) (bool, error) {
    return r.q.CheckActiveLegalHold(ctx, recordID) // r.q is sqlc-generated, typed
}

// transactions — atomic multi-statement writes
func (r *postgresErrorGroupStore) RecordOccurrence(ctx context.Context, groupID uuid.UUID) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // no-op if Commit already succeeded

    if err := r.q.WithTx(tx).IncrementOccurrenceCount(ctx, groupID); err != nil {
        return fmt.Errorf("increment occurrence count: %w", err)
    }
    if err := r.q.WithTx(tx).UpdateLastSeenAt(ctx, groupID); err != nil {
        return fmt.Errorf("update last seen: %w", err)
    }
    return tx.Commit()
}
```

## 11. JSON & API Contracts

```go
type Session struct {
    ID          uuid.UUID `json:"session_id"`   // MixedCaps in Go, snake_case on the wire
    Environment string    `json:"environment"`
    Release     string    `json:"release_id"`
}

// conditionally-required field — explicit pointer, never a bare string with omitempty
type UnmaskRequest struct {
    FieldID uuid.UUID `json:"field_id"`
    Reason  *string   `json:"reason"` // nil = missing; empty string is a distinct, still-invalid case
}
```

A masked field's Go type should make an accidental unmasked serialization a compile-time impossibility — a placeholder-carrying type on the read path, never the raw string with a marshaler that "usually" masks it.

## 12. Security-Specific Conventions

```go
// tokens/secrets — crypto/rand, never math/rand
func GenerateAPIToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil { // crypto/rand
        return "", fmt.Errorf("generating token: %w", err)
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// comparing a secret — constant-time, never == or bytes.Equal
func VerifyWebhookSignature(received, computed []byte) bool {
    return subtle.ConstantTimeCompare(received, computed) == 1
}
```

- Every SQL statement is parameterized (`sqlc`-generated queries are parameterized by construction — no string-concatenated SQL, anywhere, including `LegalHold.scope` evaluation).
- Password hashing is Argon2id, exactly as specified in [Authentication](../research/07-security/README.md#authentication) — enforced to one shared `internal/auth` package, not re-implemented per service.

## 13. Testing

```go
func TestRetentionPolicy_LongestClockWins(t *testing.T) {
    tests := []struct {
        name     string
        policies []RetentionPolicy
        want     time.Duration
    }{
        {"HIPAA beats PCI-DSS", []RetentionPolicy{hipaaSixYear, pciTwelveMonth}, sixYears},
        {"single policy", []RetentionPolicy{gdprPurposeBound}, purposeBoundDefault},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := ResolveRetention(tc.policies)
            require.NoError(t, got.Err) // setup precondition — stop here if it fails
            assert.Equal(t, tc.want, got.Duration) // the actual assertion under test
        })
    }
}
```

- **Table-driven, `t.Run(tc.name, ...)` per case** — a failing row names itself in CI output.
- **Standard library `testing` + `testify`** (`assert`/`require`), not a BDD framework.
- The compliance-guarantee integration suite and the e2e suite are specified in [`research/11-engineering/README.md#testing`](../research/11-engineering/README.md#testing) — strategy lives there; this section is unit-test style only.

## 14. Package-Local Conventions

- **`internal/` for everything specific to one service** (`domain`, `handlers`, `store`) — compiler-enforced, not just a naming convention.
- **Flat, one or two levels deep.** `internal/domain`, not `internal/domain/entities/session/v2`.
- Exported identifiers get a `godoc` comment starting with the identifier's name. Nothing more elaborate is required by default.

## 15. What's Deliberately Not Specified Here

- File organization *within* a package beyond the rules above — team-calibrated, evolves.
- IDE/editor setup — developer convenience, not a codebase standard.
- Coverage targets, load-testing methodology — see [`research/11-engineering/README.md#testing`](../research/11-engineering/README.md#testing).

---

# Part B — TypeScript (`browser-sdk`, `dashboard`)

## 16. Compiler Configuration (CI-gated)

```jsonc
// tsconfig.json — shared base, extended per package
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,     // arr[i] is T | undefined — catches the #1 real source of "cannot read property of undefined"
    "exactOptionalPropertyTypes": true,   // `reason?: string` means absent, not present-and-undefined
    "noPropertyAccessFromIndexSignature": true,
    "useUnknownInCatchVariables": true,   // caught errors are `unknown`, forcing a check before use
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "isolatedModules": true,
    "skipLibCheck": true
  }
}
```

`typescript-eslint`'s `recommended-type-checked` + `stylistic-type-checked` configs — not the non-type-checked `recommended` alone. The type-checked rules catch `no-floating-promises`, which matters directly here: an `await`-less call to `query-api` would otherwise fail silently, exactly the kind of dropped write this project's compliance surface can't tolerate.

## 17. Naming & Type Conventions

- `PascalCase` for types/components/interfaces, `camelCase` for variables/functions, `SCREAMING_SNAKE_CASE` for true module-level constants.
- **`type` by default; `interface` only for a public extensible contract or a class contract.** `type` is the only syntax for unions — and every state-shaped value in this codebase (an `AccessAuditEvent`'s `action`, a UI load/error/success state) should be a discriminated union, which settles the choice for the type it belongs to.
- **`unknown`, never `any`, at any boundary where a value's shape isn't yet proven.** `any` is a lint-blocked finding (`no-explicit-any`), not a style nitpick.

```typescript
// discriminated union — mirrors the Go-side AccessAuditEvent.action enum
type AccessAuditEvent =
  | { action: "VIEW"; targetRecordId: string }
  | { action: "UNMASK"; targetRecordId: string; reason: string }
  | { action: "CONFIG_CHANGED"; targetRecordId: string; oldValue: string; newValue: string };

function describe(event: AccessAuditEvent): string {
  switch (event.action) {
    case "VIEW": return `Viewed ${event.targetRecordId}`;
    case "UNMASK": return `Unmasked ${event.targetRecordId}: ${event.reason}`;
    case "CONFIG_CHANGED": return `Changed ${event.targetRecordId}`;
    default: {
      const _exhaustive: never = event; // compile error if a variant is added without updating this switch
      throw new Error(`unhandled action: ${JSON.stringify(_exhaustive)}`);
    }
  }
}
```

## 18. Runtime Validation at Every Network Boundary

TypeScript's types are erased at compile time — an API response typed as `Session` without a runtime check is trusting the network, not verifying it. Every fetch in `dashboard`, and every host-page config `browser-sdk` accepts, parses through **[Zod](https://zod.dev)** before the rest of the code sees it as a typed value.

```typescript
const SessionSchema = z.object({
  sessionId: z.string().uuid(),
  environment: z.string(),
  releaseId: z.string(),
});
type Session = z.infer<typeof SessionSchema>; // derived from the schema, never hand-duplicated

async function fetchSession(id: string): Promise<Session> {
  const res = await fetch(`/v1/sessions/${id}`);
  return SessionSchema.parse(await res.json()); // throws a descriptive error on shape mismatch
}
```

## 19. `browser-sdk`-Specific Conventions

- **Zero runtime dependencies beyond what the [~20KB gzipped bundle budget](architecture.md) already accounts for.** Every dependency addition is checked against that budget before it's checked against convenience.
- **Framework-agnostic core; wrappers add only a lifecycle hook, nothing else.** `@imora/react`'s entire job is a React error boundary calling into `@imora/core` — accumulating independent logic in a wrapper is a design smell, since [SDK API](../research/06-api/README.md#sdk-api) guarantees identical capability across every framework package.
- **Never throw into host-page code paths the host doesn't control.** A masking failure or a capture error inside `browser-sdk` degrades gracefully (drop the event, log to the SDK's own internal diagnostic channel) — it never becomes an uncaught exception in a customer's production application.

## 20. `dashboard`-Specific Conventions

`dashboard` is [TanStack Start](https://tanstack.com/start) (SSR, on Node via Nitro) — not a static SPA. File-based routes mirror [`user-stories.md`](user-stories.md)'s flow boundaries: `routes/sessions.search.tsx`, `routes/legal-holds.tsx`, `routes/evidence-export.tsx`.

**The one rule everything else here follows: a server function or route loader's only external call is HTTP to `query-api`. Never a direct database import.** SSR makes it *possible* for `dashboard` to hold its own Postgres/ClickHouse credentials — doing so would bypass `query-api`'s audit-event guarantee entirely, per [`design-system.md §12`](design-system.md). Code review should treat a database driver import anywhere under `app/server/` as a blocking finding, not a style comment.

```typescript
// app/server/sessions.ts — correct: proxies to query-api
export const getSession = createServerFn({ method: "GET" })
  .validator((id: string) => id)
  .handler(async ({ data: id }) => {
    const res = await fetch(`${process.env.QUERY_API_URL}/v1/sessions/${id}`, {
      headers: { Authorization: `Bearer ${getServerSessionToken()}` },
    });
    return SessionSchema.parse(await res.json());
  });
```

**Server state in [TanStack Query](https://tanstack.com/query), client-only UI state in [Zustand](https://github.com/pmndrs/zustand) — never mirrored between the two.** A route loader prefetches into the same cache the client reuses on navigation.

```typescript
// route loader — SSR prefetch
export const Route = createFileRoute("/sessions/search")({
  loader: ({ context: { queryClient }, deps: { environment, filters } }) =>
    queryClient.ensureQueryData({
      queryKey: ["sessions", environment, filters],
      queryFn: () => getSessions({ data: { environment, filters } }),
    }),
});

// client-side — same query key, cache hit if the loader already ran
function useSessions(environment: string, filters: SessionFilters) {
  return useQuery({
    queryKey: ["sessions", environment, filters],
    queryFn: () => getSessions({ data: { environment, filters } }),
  });
}

// client-only state — Zustand, no provider, no boilerplate
const useEnvironmentStore = create<{ environment: string; setEnvironment: (e: string) => void }>((set) => ({
  environment: "production", // per user-stories.md's global pattern: defaults to production, always
  setEnvironment: (environment) => set({ environment }),
}));
```

- **Query by role/label in tests, never by CSS class or an unnecessary `data-testid`** — an accessible query doubles as a lightweight accessibility check.

## 21. Testing (TypeScript)

**[Vitest](https://vitest.dev)**, not Jest — native ESM, Vite-native, faster on this codebase's realistic size, the same lower-ceremony bias as Go's stdlib-`testing`-over-BDD-framework choice. **[React Testing Library](https://testing-library.com/react)** for `dashboard` components — test behavior a user would see, not internal state.

```typescript
test("unmask requires a non-empty reason before submitting", async () => {
  render(<MaskedField value="[masked]" onUnmask={vi.fn()} />);
  await userEvent.click(screen.getByRole("button", { name: /unmask/i }));
  await userEvent.click(screen.getByRole("button", { name: /submit/i }));
  expect(screen.getByText(/reason is required/i)).toBeInTheDocument();
});
```

## 22. What's Deliberately Not Specified Here (TypeScript)

- Specific component library / design token choice for `dashboard` — a product-design decision, not a coding standard; see [`design-system.md`](design-system.md) and [`research/10-design/README.md`](../research/10-design/README.md).
- CSS approach (CSS Modules, Tailwind, vanilla-extract) — implementation detail downstream of whatever `dashboard`'s actual visual design work settles on.

---

## What This Feeds

[`design-system.md`](design-system.md) — the architectural layer these rules sit inside: package boundaries, repository pattern, dependency direction, and service bootstrap for Go; feature-folder structure and the server-state/client-state split for TypeScript.
