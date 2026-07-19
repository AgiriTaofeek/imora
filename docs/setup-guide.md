# Imora — Monorepo Setup Guide

> A step-by-step, run-it-yourself guide to bootstrapping the repository skeleton this doc set has been designing: [`architecture.md`](architecture.md)'s tech stack, [`design-system.md`](design-system.md)'s package boundaries, and [`coding-standards.md`](coding-standards.md)'s tooling, all made concrete on disk. Every command below was checked against each tool's current documentation while writing this guide — versions and exact flags may drift after that point, so treat a version number here as "known-good at time of writing," not a permanent pin. Where a real decision was made (not just "the correct command"), the reasoning is inline, not deferred elsewhere.
>
> This guide assumes you're running the commands yourself, in order, checking each step's result before moving to the next.

---

## The Journey, in Plain Terms

Eleven sections, but really four movements. Knowing which movement you're in tells you *why* the current step exists, not just what to type:

1. **Lay the ground (§1–§2).** An empty repo, a `LICENSE`, and a folder tree with nothing in it yet — just the shape [`design-system.md`](design-system.md) already decided, made real on disk.
2. **Install the two toolchains (§3–§4).** Go tooling for the six backend services, TypeScript tooling for `browser-sdk`/`dashboard`. Nothing service-specific happens yet — this is "make sure `go build`, `golangci-lint`, `pnpm`, and Biome all work" before there's any real code for them to check.
3. **Build the one thing everything else depends on (§5).** The shared schema package is deliberately done *before* any service — every Go service and every TypeScript package downstream imports from it, so it has to exist and actually generate correctly first, or every later step is building on sand.
4. **Fill in the services (§6–§9).** Go services, `browser-sdk`, `dashboard`, then wire up testing across all of it. This is the only movement where you're touching six-plus separate directories, which is exactly where it's easiest to lose track of what's actually done — §10's checklist and the "Checkpoint" call-out at the end of each section below exist specifically to catch that.

If you stop partway through and come back later, the fastest way to reorient is: which movement was I in, and what does that section's checkpoint say I should already have?

---

## 0. Prerequisites

| Tool | Version | Check |
|---|---|---|
| Go | 1.24+ | `go version` |
| Node.js | **22+ required** | `node --version` |
| pnpm | **11+ required** | `pnpm --version` (install via `corepack enable` if missing — ships with Node 20+) |
| Docker + Docker Compose | current | `docker compose version` |
| golangci-lint | v2.x | installed in step 3 |

**Node 22+ and pnpm 11+ are hard requirements, not "latest is nice to have."** pnpm 11 itself requires Node 22+ to run at all — and separately, pnpm 11 changed two defaults worth knowing about *before* you hit them rather than mid-install: postinstall/build scripts now error unless explicitly allowlisted (`allowBuilds` in `pnpm-workspace.yaml`), and by default it refuses to install any package version published less than 24 hours ago (`minimumReleaseAge`, a supply-chain guard — see the note in §4 for what to do when this blocks something real, which it will, on a genuinely fresh install).

---

## 1. Repository Root

```bash
mkdir imora && cd imora
git init
```

Root-level files, before anything language-specific:

```bash
touch LICENSE README.md .gitignore .editorconfig
```

**`LICENSE`** — the full AGPLv3 text ([gnu.org/licenses/agpl-3.0.txt](https://www.gnu.org/licenses/agpl-3.0.txt)), per [ADR 0001](../research/11-engineering/architecture-decisions/0001-agplv3-licensing.md). This is the **only** LICENSE file in the repository — per [Repository Structure](../research/03-architecture/README.md#repository-structure), no subdirectory ever gets its own.

**`.gitignore`** (root):

```gitignore
# Go
/bin/
*.test
*.out

# Node / pnpm
node_modules/
.turbo/
dist/
.vite/

# Generated (regenerate via codegen, per §5 — don't hand-edit, but do commit unless your team decides otherwise)
# packages/*/gen/ is intentionally NOT ignored here — see the note in §5.

# Env / secrets — never committed
.env
.env.local
*.pem
*.key

# OS/editor noise
.DS_Store
.idea/
*.swp
```

**`.editorconfig`** — keeps `gofmt`'s tabs and Biome's spaces from fighting your editor's defaults:

```ini
root = true

[*]
end_of_line = lf
insert_final_newline = true
charset = utf-8
trim_trailing_whitespace = true

[*.go]
indent_style = tab

[*.{ts,tsx,js,jsx,json,yaml,yml,md}]
indent_style = space
indent_size = 2
```

**✅ Checkpoint — before moving to §2, confirm all four files actually have content, not just exist:** `touch` creates an empty file even if you forget to paste anything into it, and an empty `LICENSE` is easy to miss until someone asks "wait, is this actually AGPL?" months later.

```bash
test -s LICENSE && test -s README.md && test -s .gitignore && test -s .editorconfig && echo "✓ all four have content" || echo "✗ something is still empty"
```

---

## 2. Directory Skeleton

This section does exactly one thing: turn [Repository Structure](../research/03-architecture/README.md#repository-structure)'s and [`design-system.md`](design-system.md)'s already-decided folder layout into real, empty directories. Nothing here is a new decision — if you're wondering "why does `gateway` get an `internal/store/` but not a `internal/api/`," that answer lives in those two documents, not here. This section is just typing `mkdir` enough times to match what they already specified.

```bash
mkdir -p services/{gateway,ingestion,query-api,alert-engine,workers,notification-service}
mkdir -p sdk/browser-sdk/packages/{core,react,vue,angular}
mkdir -p dashboard
mkdir -p packages/{domain-types,event-schemas}
mkdir -p deploy/{compose,kubernetes}
mkdir -p tools
```

`docs/` and `research/` already exist (this doc set). Each `services/*` directory gets the same four-folder shape, per [`design-system.md §2`](design-system.md) — `cmd/` for the entry point, `internal/domain` for business logic, `internal/handlers` for HTTP adapters, `internal/store` for the repository/adapter implementations. One loop instead of typing this out six times by hand:

```bash
for svc in gateway ingestion query-api alert-engine workers notification-service; do
  mkdir -p "services/$svc/cmd" "services/$svc/internal/domain" "services/$svc/internal/handlers" "services/$svc/internal/store"
done
```

**✅ Checkpoint** — you should now have an entirely empty skeleton, 24 subdirectories under `services/` (4 per service × 6 services) plus the top-level folders above. Nothing in this tree has a single file in it yet, and that's correct — §3 onward is what starts putting real files inside it.

```bash
find services -type d | wc -l   # expect 30 (6 services + 24 subdirectories)
```

---

## 3. Go: Single Module, golangci-lint, Base Tooling

This is the first section that installs and configures anything — everything before this was just directories. The goal here is narrow on purpose: get `go build` and `golangci-lint` both working cleanly against an *empty* module before any service has real code in it. That way, if either one breaks later while you're actually writing a handler, you already know the tooling itself was fine at the start, and the problem is in what you just wrote — not in some setup step three weeks ago you've forgotten about.

Per the earlier decision to use **one `go.mod` at the repo root** (not Go workspaces) — the six services and `packages/domain-types` are one module, one dependency graph, one `go.sum`. This matches [ADR 0003](../research/11-engineering/architecture-decisions/0003-monorepo-structure.md)'s reasoning directly: there's already one shared deploy/release cadence for all backend services, so the extra indirection of `go.work` + six separate `go.mod` files buys nothing here.

```bash
go mod init github.com/<your-org>/imora   # replace with your actual module path — this string is baked into every internal import path, changing it later means rewriting every import
```

### golangci-lint (v2)

```bash
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2
golangci-lint --version   # confirm v2.x
```

**Corrected config** — note this supersedes the `.golangci.yml` snippet in [`coding-standards.md`](coding-standards.md), which was written before golangci-lint's v2 config format (`version: "2"`, `linters.default`) shipped. This is the accurate version:

```yaml
# .golangci.yml — repo root
version: "2"

run:
  timeout: 5m

linters:
  default: standard   # errcheck, govet, staticcheck, unused, ineffassign
  enable:
    - bodyclose
    - contextcheck
    - containedctx
    - wrapcheck
    - gosec
  settings:
    wrapcheck:
      ignoreSigs:
        - context.Canceled
        - context.DeadlineExceeded
```

```bash
golangci-lint run ./...   # should pass cleanly on an empty module
```

**✅ Checkpoint** — you should now have `go.mod` at the repo root (one line: `module github.com/<your-org>/imora`, plus a `go` directive) and `.golangci.yml` at the repo root. `go build ./...` and `golangci-lint run ./...` both succeed trivially, because there's still no actual Go code anywhere to fail on — that's expected at this point, not a sign something's missing.

```bash
test -f go.mod && test -f .golangci.yml && echo "✓ both present" || echo "✗ one is missing"
```

---

## 4. TypeScript: pnpm Workspaces + Turborepo

Same idea as §3, mirrored for the other language: get `pnpm`, Turborepo, and Biome all working against an empty workspace before any package has real code. By the end of this section you'll have a root `package.json`, `pnpm-workspace.yaml`, `turbo.json`, `tsconfig.json`, and `biome.json` — five files, none of which belong to any single package, all of which every TypeScript package in the monorepo inherits from.

```bash
corepack enable
pnpm --version   # confirm 11+
```

**`pnpm-workspace.yaml`** (repo root):

```yaml
packages:
  - "sdk/browser-sdk/packages/*"
  - "dashboard"
  - "packages/event-schemas"   # the TS side of the shared schema package, per §5
```

**About that `minimumReleaseAge` note from §0:** the first time `pnpm install` tries to pull in a package version published less than 24 hours ago, it will fail with `ERR_PNPM_MINIMUM_RELEASE_AGE_VIOLATION` — this is not a sign anything is broken, it's pnpm 11's supply-chain guard doing exactly what it's designed to do. Don't reach for `minimumReleaseAge: 0` (that disables the check project-wide, forever). Add a **narrow, named exception** instead — the same "keep the guardrail, scope the exception" pattern `wrapcheck`'s `ignoreSigs` already uses in `.golangci.yml`:

```yaml
minimumReleaseAgeExclude:
  - "js-base64@3.9.0"   # example — pnpm auto-appends the real offending package@version here when it blocks one
```

You often won't have to write these by hand — the first time `pnpm install` blocks a package, it auto-adds a placeholder entry for you in `pnpm-workspace.yaml`, which you then just confirm or remove once that specific version has aged past 24 hours.

**pnpm 11's other new default worth knowing before §7:** postinstall/build scripts now error unless explicitly allowlisted. The first time you `pnpm add -D tsup` (§7), expect `[ERR_PNPM_IGNORED_BUILDS] Ignored build scripts: esbuild@x.x.x` — `esbuild`'s postinstall fetches its platform-specific native binary and isn't optional for `tsup`/`vite`/`vitest` to function. Allow it explicitly rather than disabling the guard globally:

```yaml
allowBuilds:
  esbuild: true   # required for tsup/vite/vitest — its postinstall builds the native binary, not optional
```

Then re-run `pnpm install` once to actually execute the now-approved postinstall script.

**Root `package.json`**:

```json
{
  "name": "imora",
  "private": true,
  "license": "AGPL-3.0-only",
  "scripts": {
    "build": "turbo run build",
    "test": "turbo run test",
    "lint": "biome check ."
  },
  "devDependencies": {
    "turbo": "^2.10.5",
    "typescript": "^6.0.3"
  },
  "packageManager": "pnpm@11.13.0"
}
```

**Pinned to TypeScript 6.x, deliberately not 7 — this needs revisiting, not a permanent decision.** TypeScript 7.0 (GA July 8, 2026) rewrote the compiler in Go for an 8–12x build-speed improvement, and there's no reason not to want that eventually. It doesn't ship a programmatic compiler API yet, though — only compiled `tsc`/`tsserver` binaries — which breaks any tool that embeds the compiler directly rather than shelling out to it. That specifically includes `vue-tsc` (`sdk/browser-sdk/packages/vue`) and Angular's editor/language-service tooling (`sdk/browser-sdk/packages/angular`'s dev experience, though its CLI build-time checking is fine on TS 7 already). Rather than split the monorepo across two TypeScript versions for a partial win, everything stays on 6.x until **TypeScript 7.1** closes this gap — Microsoft's own timeline puts that around October 2026, based on their typical release cadence. Revisit this pin then.

```bash
pnpm install
pnpm add turbo --save-dev --workspace-root
```

**`turbo.json`** (repo root) — the `^build` dependency graph means `dashboard`'s build waits on `packages/event-schemas`'s build first, since it imports the generated Zod schemas from it. **No `lint` task here, deliberately** — Turborepo's task model is for orchestrating *per-package* scripts, but Biome checks the whole workspace as one command (§4 above), so routing it through `turbo run lint` finds zero per-package `"lint"` scripts to run and silently does nothing (`0 successful, 0 total`, exit code 0 — looks like a pass, isn't one). Root `package.json`'s `"lint"` script calls `biome check .` directly instead, bypassing Turborepo entirely for this one task — confirmed by actually hitting the silent-no-op and fixing it, not assumed:

```json
{
  "$schema": "https://turborepo.dev/schema.json",
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "test": {
      "dependsOn": ["build"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    }
  }
}
```

**Root `tsconfig.json`** (the shared base every package extends), per [`coding-standards.md §16`](coding-standards.md):

```jsonc
{
  "compilerOptions": {
    // Required for tsup's dts-bundling step to succeed on TypeScript 6.x — without it,
    // `pnpm build` on any browser-sdk package fails with TS5101 ("baseUrl is deprecated")
    // even though nothing in this file sets baseUrl; tsup's internal dts build implicitly
    // triggers the check. Confirmed by actually hitting the failure and fixing it, not assumed.
    "ignoreDeprecations": "6.0",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noPropertyAccessFromIndexSignature": true,
    "useUnknownInCatchVariables": true,
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "isolatedModules": true,
    "skipLibCheck": true
  }
}
```

**[Biome](https://biomejs.dev)** — one tool, both linting and formatting, replacing the ESLint+Prettier pair entirely:

```bash
pnpm add -D -w -E @biomejs/biome   # -E: exact version, per Biome's own recommendation, avoids format/lint drift between contributors
pnpm exec biome init                # scaffolds a starting biome.json
```

**Root `biome.json`** — the `recommended` rule set plus `noFloatingPromises` (still in Biome's `nursery` group, but worth enabling explicitly per [`coding-standards.md §16`](coding-standards.md)'s note on why):

Run `pnpm exec biome init` rather than hand-writing this file — it detects your actual installed Biome version and generates a config in that version's real schema (e.g. `assist.actions.source.organizeImports` instead of a top-level `organizeImports` key, `rules.preset` instead of `rules.recommended` — this shifted between Biome versions, so trust what `init` generates over a hand-copied example that may already be one version behind by the time you read this). Take its output as-is, then layer in the two changes this project actually needs on top of it:

```jsonc
{
  // ...whatever `biome init` generated stays as-is...
  "formatter": {
    "enabled": true,
    "indentStyle": "space"   // biome init may default to "tab" — must match .editorconfig's space rule for *.{ts,tsx,js,jsx,json,yaml,yml,md}, or Biome and your editor fight on every save
  },
  "linter": {
    "enabled": true,
    "rules": {
      "preset": "recommended",   // or whatever key your generated config used for this
      "nursery": { "noFloatingPromises": "error" }   // the one rule this project explicitly wants on, per coding-standards.md §16
    }
  }
}
```

Each TypeScript package (`sdk/browser-sdk/packages/*`, `dashboard`, `packages/event-schemas`) that needs package-specific overrides gets its own `biome.json` with `"root": false` and `"extends": "//"` (Biome's monorepo-inheritance syntax) rather than duplicating the root config — mirrors the same "one shared config, no per-package drift" rule [`coding-standards.md §1`](coding-standards.md) already applies to `golangci-lint`.

```bash
pnpm biome check .   # lint + format check, across the whole workspace
pnpm biome check --write .   # same, but auto-fixes what it safely can
```

**✅ Checkpoint** — five new files at the repo root: `package.json`, `pnpm-workspace.yaml`, `turbo.json`, `tsconfig.json`, `biome.json`. `pnpm install` and `pnpm biome check .` both succeed cleanly.

```bash
for f in package.json pnpm-workspace.yaml turbo.json tsconfig.json biome.json; do test -f "$f" && echo "✓ $f" || echo "✗ $f MISSING"; done
```

**With both toolchains verified, §3–§4's job is done — everything from here on is building actual packages inside them, not more setup-of-the-setup.**

---

## 5. The Shared Schema Package: JSON Schema as the Single Source of Truth

**Why this comes before any of the six services or `browser-sdk` get touched:** every one of them is going to import from this package. Building it now, and proving the generation pipeline actually works end to end (both the Go side and the TypeScript side), means §6 onward is "import a type that already exists and already works" rather than "import a type while also debugging the thing that generates it." This section is also the longest one in the guide, and for a specific reason: it's the one place two different tools' rough edges actually surfaced while writing it (an experimental Zod API that fails on a real type error, `quicktype` needing its output directory to already exist) — those aren't hypothetical warnings, they're what happened when this was actually run, which is exactly why they're documented in this much detail instead of glossed over.

This is the resolved answer to a real gap in the earlier design docs: [Bounded Contexts](../research/02-domain/README.md#bounded-contexts) says `browser-sdk` (TypeScript) and `ingestion`/`query-api` (Go) share "the same entity definitions" — which a Go package literally cannot provide to TypeScript code. The mechanism: **JSON Schema files are the actual source of truth, checked in at `packages/event-schemas/schemas/*.json`; Go structs and TypeScript/Zod schemas are both generated (or derived) from them, never hand-duplicated.**

```bash
mkdir -p packages/event-schemas/schemas
```

Example schema, matching [Event Schema](../research/05-data/README.md#event-schema)'s `SessionEventCaptured` definition:

```json
// packages/event-schemas/schemas/session-event-captured.schema.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "session-event-captured.schema.json",
  "title": "SessionEventCaptured",
  "type": "object",
  "required": ["sessionId", "subtype", "payload", "releaseId", "environment"],
  "properties": {
    "sessionId": { "type": "string", "format": "uuid" },
    "subtype": { "enum": ["FullSnapshot", "DOMMutation", "MouseMove", "Click", "Scroll", "FormInput", "ViewportChange"] },
    "payload": { "type": "object" },
    "releaseId": { "type": "string" },
    "environment": { "type": "string" }
  }
}
```

### Go side: codegen via `go-jsonschema`

```bash
go install github.com/atombender/go-jsonschema@latest
```

Run this once, from the repo root, to confirm it works — open the resulting file afterward and skim it:

```bash
go-jsonschema \
  --package eventschemas \
  --struct-name-from-title \
  --capitalization ID \
  packages/event-schemas/schemas/*.schema.json \
  > packages/domain-types/gen_event_schemas.go
```

**Both extra flags matter, not just style:** without `--struct-name-from-title`, go-jsonschema names the struct after the filename with a `SchemaJson` suffix (`SessionEventCapturedSchemaJson`) instead of using the schema's own `"title"` field — noisier than it needs to be. Without `--capitalization ID`, every `*Id` field (`SessionId`, `ReleaseId`) is generated with the wrong casing per [Naming Conventions](../research/11-engineering/README.md#coding-standards) (`SessionID`, not `SessionId` — initialisms keep one consistent case). Confirmed by actually running both versions and diffing the output while writing this guide, not assumed.

**Now wire this into `go generate` instead of retyping the command every time you touch a schema.** The problem being solved: nothing stops you from editing a schema file and forgetting to regenerate — the Go structs silently drift out of sync with the JSON Schema that's supposed to be their source of truth, with no error until something downstream breaks. `go generate` is Go's built-in fix for exactly this: you write the regeneration command **once**, as a specially-formatted comment (`//go:generate ...`) at the top of a Go file. That comment does *nothing* by itself — `go build`, `go run`, `go test` all ignore it completely. It only executes when you explicitly type `go generate`, at which point Go's tooling scans for every `//go:generate` line in the target package(s) and runs each one as a command. So the whole pattern is: pay the cost of writing the full command out once, and from then on, regenerating is just `go generate` — no flags to remember, no path to get wrong.

**Create** `packages/domain-types/doc.go` (it doesn't exist yet) with exactly this content:

```go
//go:generate sh -c "go-jsonschema --package eventschemas --struct-name-from-title --capitalization ID ../event-schemas/schemas/*.schema.json > gen_event_schemas.go"
package eventschemas
```

Two things about this exact line, both verified by actually running it while writing this guide, not assumed:

- **The schema path is relative to `packages/domain-types/`, not the repo root** (`../event-schemas/schemas/*.schema.json`, not `packages/event-schemas/...`) — `go generate` sets its working directory to wherever the file containing the `//go:generate` comment lives, so the path has to resolve from *there*.
- **The whole command is wrapped in `sh -c "..."`.** This is not optional. `go generate` does not invoke commands through a shell — it splits the line into tokens and executes the first as a program, passing the rest as literal arguments. Without the `sh -c` wrapper, the `*.schema.json` glob and the `>` redirect are both passed through *unexpanded* to `go-jsonschema`, which then fails trying to resolve a schema file literally named `*.schema.json` (confirmed — this is a real failure mode I hit and fixed while writing this guide, not a hypothetical one). Wrapping in `sh -c "..."` forces actual shell interpretation of the glob and the redirect, exactly like running the command directly in a terminal.

**To actually run it:** `go generate ./...` from the repo root (scans the whole module for `//go:generate` comments and runs them all) or `go generate` from inside `packages/domain-types/` (just this one). Either way, it re-runs the exact command above and overwrites `gen_event_schemas.go` — no hand-typing required, and nothing runs unless you explicitly ask for it.

### TypeScript side: codegen via `quicktype`'s `typescript-zod` target

An earlier draft of this guide pointed at Zod v4's native `z.fromJSONSchema()` for this. **Don't use that** — it's confirmed experimental (explicitly not part of Zod's stable API, likely to change), and it fails in practice: importing a `.json` file in TypeScript widens `"$schema": "https://json-schema.org/draft/2020-12/schema"` from its literal value to plain `string`, which doesn't satisfy `fromJSONSchema`'s type signature (it demands one of three exact literal strings) — a real type error, not a hypothetical one, hit while working through this guide.

`quicktype` is the better fit, for the same reason `go-jsonschema` is the right call on the Go side above: a mature, single-purpose codegen tool producing a checked-in file, not a runtime conversion leaning on an unstable API. It has a dedicated `typescript-zod` output target, actively used for exactly this JSON-Schema-to-Zod use case.

```bash
pnpm add -D -w -E quicktype   # -E: exact version, same reason as Biome above — its output is checked into git,
                                # so an unpinned bump could produce a generated-code diff with no real schema change behind it
```

**Create the output directory first — `quicktype`'s `-o` flag doesn't create missing intermediate directories the way `mkdir -p` does, and fails with `Error: ENOENT: no such file or directory` if `src/gen/` doesn't exist yet** (confirmed by actually hitting this, not assumed):

```bash
mkdir -p packages/event-schemas/src/gen
```

Then run it against every schema file at once — like `go-jsonschema` above, this merges multiple schemas into one output file with one exported Zod schema + inferred type per entity:

```bash
pnpm exec quicktype -s schema -l typescript-zod \
  packages/event-schemas/schemas/*.schema.json \
  -o packages/event-schemas/src/gen/event-schemas.ts
```

**Verified by actually running this against the two schemas above, not assumed.** The real output for `SessionEventCaptured`:

```typescript
import * as z from "zod";

export const SubtypeSchema = z.enum([
    "Click", "DOMMutation", "FormInput", "FullSnapshot", "MouseMove", "Scroll", "ViewportChange",
]);
export type Subtype = z.infer<typeof SubtypeSchema>;

export const SessionEventCapturedSchema = z.object({
    "environment": z.string(),
    "payload": z.record(z.string(), z.any()),
    "releaseId": z.string(),
    "sessionId": z.string(),
    "subtype": SubtypeSchema,
});
export type SessionEventCaptured = z.infer<typeof SessionEventCapturedSchema>;
```

**One real gap, confirmed by inspecting the generated output, not assumed away:** `"format": "date-time"` correctly becomes `z.coerce.date()`, but `"format": "uuid"` does **not** become `z.string().uuid()` — it's generated as a bare `z.string()`. So `sessionId`/`eventId`/`actorUserId` pass Zod validation with any string, not just a well-formed UUID. Two ways to handle this, pick one rather than silently accepting the gap:
- **Accept it** — the Go side's structural guarantees (BR-5's audit-event enforcement, the domain layer's own UUID parsing on write) are the actual source of correctness here; the TypeScript-side check exists to catch a malformed *response shape* from `query-api`, not to be the sole line of defense against a malformed UUID.
- **Patch specific fields post-generation** — a small script that runs after quicktype and upgrades named fields (`sessionId`, `eventId`, `actorUserId`, any `*Id` field) from `z.string()` to `z.string().uuid()`, similar in spirit to the `--capitalization ID` fix already applied on the Go side. Worth doing if a genuinely malformed ID reaching `dashboard` would cause a confusing failure downstream rather than an obviously-wrong one.

**Wire into a `package.json` script**, the TypeScript-side equivalent of `go generate`:

```json
{
  "scripts": {
    "generate": "mkdir -p src/gen && quicktype -s schema -l typescript-zod schemas/*.schema.json -o src/gen/event-schemas.ts"
  }
}
```

```bash
cd packages/event-schemas && pnpm init && pnpm add zod@^4
pnpm generate   # runs the script above, writes src/gen/event-schemas.ts
```

```typescript
// packages/event-schemas/src/index.ts — re-export the generated file; nothing hand-written here
export * from "./gen/event-schemas";
```

`browser-sdk` and `dashboard` import `SessionEventCaptured` (the Zod schema, for runtime validation) and its inferred type from this one package — never redefine the shape by hand, per [`coding-standards.md §18`](coding-standards.md).

**Verify both sides agree:** after generating, spot-check that the Go struct's JSON tags and the Zod schema's keys match the schema file exactly — this is what "same entity definition, not translated copies" actually means in practice, and it's worth a one-time manual diff the first time this pipeline runs.

**✅ Checkpoint** — at least one `.schema.json` file exists under `packages/event-schemas/schemas/`, and both generated outputs exist and are non-empty: `packages/domain-types/gen_event_schemas.go` (Go) and `packages/event-schemas/src/gen/event-schemas.ts` (TypeScript). `packages/event-schemas/src/index.ts` exists and re-exports the generated file — **this specific file is easy to lose track of**, since nothing else in this section creates it automatically and no build step will complain about a missing export until something downstream actually tries to import it.

```bash
ls packages/event-schemas/schemas/*.schema.json \
   packages/domain-types/gen_event_schemas.go \
   packages/event-schemas/src/gen/event-schemas.ts \
   packages/event-schemas/src/index.ts 2>&1
```

If any line above prints "No such file or directory" instead of the path, that's the thing to fix before going any further — everything from §6 on assumes this pipeline is real and working, not stubbed.

---

## 6. Go Services

The shared schema package now exists and generates real types — this section is where those types actually get used for the first time. Each of the six services gets minimal-but-real content in the four folders §2 already created for it: enough that `go build` compiles something meaningful, not just an empty directory. This section also brings in the two data-access tools ([Database Access](../research/11-engineering/README.md#coding-standards) already decided both of these; this is where they're actually installed) — `sqlc` for the Postgres-backed services, the `clickhouse-go` client for the ClickHouse-backed ones.

### `packages/domain-types`

```bash
mkdir -p packages/domain-types
```

This holds the generated event-schema types (§5) plus any hand-written domain types that aren't wire-format events (e.g. `LegalHold`, `RetentionPolicy` — internal domain objects per [Domain Model](../research/02-domain/README.md#domain-model), not things `browser-sdk` ever sees). Only the six backend services import this package; `browser-sdk`/`dashboard` only ever import `packages/event-schemas`' TypeScript output.

### Per-service `go.mod` entries — there are none

Since this is a single-module repo (§3), there's no per-service `go.mod`. Each service is just a directory whose `cmd/main.go` is a build target:

```bash
# from repo root, for each service
go build -o bin/gateway ./services/gateway/cmd
```

### sqlc (Postgres access, per [Database Access](../research/11-engineering/README.md#coding-standards))

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
mkdir -p services/query-api/internal/store/postgres/{queries,migrations}
```

```yaml
# services/query-api/internal/store/postgres/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries"
    schema: "migrations"
    gen:
      go:
        package: "postgresgen"
        out: "postgresgen"
        sql_package: "pgx/v5"   # modern default — supports pointer types for nullable columns, per sqlc's own PostgreSQL-specific guidance
```

```bash
cd services/query-api/internal/store/postgres && sqlc generate
```

### ClickHouse client (`ingestion`, `query-api`, `workers`, per [Database Access](../research/11-engineering/README.md#coding-standards))

```bash
go get github.com/ClickHouse/clickhouse-go/v2
```

No codegen here — per [Database Access](../research/11-engineering/README.md#coding-standards), ClickHouse access is hand-written against `clickhouse-go`'s native API directly (`sqlc` doesn't target ClickHouse).

### HTTP routing (`gateway`, `query-api`)

```bash
go get github.com/go-chi/chi/v5
```

Per [HTTP Handlers and Middleware](../research/11-engineering/README.md#coding-standards): [chi](https://github.com/go-chi/chi) — 100% `net/http`-compatible middleware and handlers, no framework-specific context type. `gateway` (the routing chokepoint, per [Bounded Contexts](../research/02-domain/README.md#bounded-contexts)) and `query-api` (the REST surface, per [REST API](../research/06-api/README.md#rest-api)) are the two services that actually route HTTP requests; the other four (`ingestion`, `alert-engine`, `workers`, `notification-service`) don't need it yet.

### Test dependencies (all services)

```bash
go get github.com/stretchr/testify
go get golang.org/x/sync/errgroup
```

**✅ Checkpoint** — `go.sum` now exists (it didn't after §3, since nothing had been `go get`'d yet). `services/query-api/internal/store/postgres/sqlc.yaml` exists, and running `sqlc generate` inside that directory produces a `postgresgen/` folder. `gateway` and `query-api` now have a minimal but real `main.go` — a chi router, the base middleware stack, a `/healthz` endpoint, and graceful shutdown (per [Graceful Shutdown](coding-standards.md#8-graceful-shutdown)) — genuinely running, not a stub; the other four services keep the empty `main()` from §6's start, since they have no HTTP surface to wire yet.

```bash
go build ./... && golangci-lint run ./...   # should still pass cleanly with real router code in the mix
```

```bash
test -f go.sum && test -f services/query-api/internal/store/postgres/sqlc.yaml && echo "✓ both present" || echo "✗ check which is missing"
```

---

## 7. `browser-sdk`

With the schema package generating real types, `browser-sdk`'s four packages (`core`, `react`, `vue`, `angular`) can now actually import something real instead of a placeholder. This section only sets up `core` explicitly — the other three follow the exact same `pnpm init` + `tsup` pattern, so once `core` is done, repeat the same two commands inside `react/`, `vue/`, and `angular/` rather than treating this as a one-off.

```bash
cd sdk/browser-sdk/packages/core && pnpm init
pnpm add -D tsup vitest
```

**`tsup.config.ts`** (each package under `sdk/browser-sdk/packages/*`):

```typescript
import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  clean: true,
  minify: true,   // bundle-size budget, per architecture.md's ~20KB gzipped target
});
```

`@imora/react`, `@imora/vue`, `@imora/angular` each `pnpm add @imora/core` (via the pnpm workspace, so it resolves to the local package, not a published one) and depend on nothing else beyond their framework's peer dependency, per [`design-system.md §11`](design-system.md).

**✅ Checkpoint** — all four packages under `sdk/browser-sdk/packages/*` have their own `package.json` and `tsup.config.ts`; the three framework wrappers each list `@imora/core` as a dependency. `pnpm build` (from the repo root) builds all four without error, even though `src/index.ts` in each one is still just a stub at this point.

```bash
for p in core react vue angular; do test -f "sdk/browser-sdk/packages/$p/package.json" && echo "✓ $p" || echo "✗ $p MISSING"; done
```

---

## 8. `dashboard`

This is the only section in the guide where a single CLI command does most of the work — `dashboard` isn't hand-assembled piece by piece the way `browser-sdk`'s packages were, because [`design-system.md §12`](design-system.md) already decided TanStack Start is the framework, and TanStack's own scaffolding tool wires its Vite config, Nitro server build, and router together correctly in one shot. Hand-assembling that config yourself would just be re-deriving what the CLI already knows, with more chances to get a wiring detail wrong.

Scaffold via the official CLI rather than hand-assembling TanStack Start's Vite/Nitro config from scratch:

```bash
cd dashboard
npx @tanstack/cli create . --add-ons tanstack-query
```

This wires TanStack Router, TanStack Start's SSR/server-function build, and TanStack Query together — matching [`design-system.md §12`](design-system.md)'s decision. `vitest`, `jsdom`, and `@testing-library/react` come pre-wired into the scaffold's `package.json` already — no need to add those three again.

**Two things to fix immediately after scaffolding, confirmed by actually hitting both:**

1. **The CLI installs via `npm`, not `pnpm`, regardless of which package manager the rest of your workspace uses** — it leaves behind a `package-lock.json` and a local `node_modules/`, both of which conflict with this workspace's single `pnpm-lock.yaml`. Delete them and let the workspace-level `pnpm install` take over:
   ```bash
   rm -f package-lock.json && rm -rf node_modules
   ```
2. **The scaffold writes a `"pnpm": { "onlyBuiltDependencies": [...] }` field into `dashboard/package.json`** (for `esbuild` and `lightningcss`, Tailwind v4's CSS engine — both have legitimate native-binary postinstall scripts). That's the **pnpm 10-and-earlier** config location — pnpm 11 (§0) silently ignores it. Delete that field from `dashboard/package.json` entirely, and add the same two packages to `allowBuilds` in the root `pnpm-workspace.yaml` instead, next to `esbuild` from §7:
   ```yaml
   allowBuilds:
     esbuild: true
     lightningcss: true   # Tailwind v4's CSS engine, native binary via postinstall — same reasoning as esbuild
   ```

Then install from the repo root (not inside `dashboard/`) so pnpm properly links it into the workspace, and add the remaining pieces §12 of the design doc calls for:

```bash
cd ..   # back to repo root
pnpm install
cd dashboard
pnpm add zustand
pnpm add @imora/event-schemas --workspace   # the Zod schemas from §5
pnpm add -D @testing-library/user-event    # the one testing dep the scaffold doesn't already include
pnpm build   # confirms the scaffold + your additions still produce a working client + SSR server bundle
```

Route structure per [`design-system.md §12`](design-system.md) — `src/routes/sessions.search.tsx`, `src/routes/legal-holds.tsx`, etc., mirroring [`user-stories.md`](user-stories.md)'s flows. **Correction, confirmed against the actual scaffold output, not assumed:** this CLI version scaffolds routes/components/server code under `src/` (`src/routes/`, `src/components/`, `src/integrations/`), not `app/` — an earlier draft of this guide (and of [`design-system.md`](design-system.md)) assumed an `app/` convention from TanStack Start's own docs that doesn't match what `@tanstack/cli` actually generates. Server functions live under `src/server/` and, per that same section's rule, **only ever call `query-api`'s REST API** — never a database driver.

**✅ Checkpoint** — `dashboard/` now has real content instead of being empty: `package.json`, `vite.config.ts`, `tsr.config.json`, and a `src/` directory containing `routes/`, `components/`, `integrations/`, `router.tsx`, `routeTree.gen.ts` (auto-generated — never hand-edit it, and it's excluded from Biome per the note below), and `styles.css`. `pnpm dev` (from inside `dashboard/`) starts a dev server without crashing. If `src/server/` doesn't exist yet, that's fine — it gets created the first time you actually write a server function, not as part of scaffolding.

**Three more things worth fixing right after scaffolding, all confirmed by actually running `pnpm lint` against the fresh scaffold:**
- **Enable Tailwind v4 parsing in Biome** (§4's `biome.json`) — without it, `dashboard/src/styles.css`'s `@theme` block fails to parse at all: `"css": { "parser": { "tailwindDirectives": true } }`.
- **Exclude the generated route tree from Biome** — `src/routeTree.gen.ts` uses `as any` internally by design (TanStack Router's own codegen, regenerated by `tsr generate` on every build); add `"!**/routeTree.gen.ts"` to `biome.json`'s `files.includes`, the same "don't lint what you don't control" treatment as any other generated file.
- **The scaffold's own `src/router.tsx` has three unused imports** (`QueryClient`, `ReactNode`, and the default `TanstackQueryProvider` export — only `getContext` from that last module is actually used) — remove them; `pnpm lint` flags all three.

```bash
test -f dashboard/package.json && echo "✓ dashboard scaffolded" || echo "✗ still empty — re-run the npx command above"
```

---

## 9. Testing Wiring

Every piece up to this point has its own build/lint tooling, but nothing yet ties "run every test in the monorepo with one command" together — that's this section's entire job, for both languages.

**Go:** no extra config — `go test ./...` from the repo root already works once `testify` is imported, per [`coding-standards.md §13`](coding-standards.md).

**TypeScript:** Vitest's **`projects`** field (the current mechanism — the older `vitest.workspace.ts` file is deprecated as of Vitest 3.2, replaced by this in-config option):

```typescript
// vitest.config.ts — repo root
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    passWithNoTests: true,   // see the note below — without this, zero test files is a hard failure
    projects: [
      "sdk/browser-sdk/packages/*",
      "dashboard",
    ],
  },
});
```

```bash
pnpm add -D -w vitest
pnpm test   # via turbo run test, per §4's turbo.json
```

**Two real problems here, both confirmed by actually running `pnpm test`, not assumed:**

1. **Vitest exits with failure when a package has zero test files, by default.** `passWithNoTests: true` (above) fixes this — without it, `pnpm test` fails on every package until real tests exist, which is backwards for a fresh scaffold.
2. **`turbo run test` doesn't invoke the root config's `projects` runner the way you'd expect.** It dispatches to *each package's own* `"test": "vitest run"` script independently — and a package with no `vitest.config.ts` of its own walks up the directory tree and finds the *root* config. That's already a surprise, but the real bug is worse: the root config's `projects` paths (`"dashboard"`, etc.) then get resolved **relative to the package that found it**, not relative to the root — so running `vitest run` from inside `sdk/browser-sdk/packages/core` fails with `Projects definition references a non-existing file or a directory: .../packages/core/dashboard`. `dashboard` itself doesn't hit this, only because TanStack's scaffold already gives it its own `vite.config.ts`, which stops the upward search before it reaches the root config.

   **Fix: give every package that doesn't already have one its own local Vitest config**, so `turbo run test`'s per-package invocation is always self-contained and never walks up:

   ```typescript
   // sdk/browser-sdk/packages/{core,react,vue,angular}/vitest.config.ts — identical in all four
   import { defineConfig } from "vitest/config";

   export default defineConfig({
     test: { passWithNoTests: true },
   });
   ```

   `dashboard` already has a `vite.config.ts` (from the TanStack scaffold) — add the same `test: { passWithNoTests: true }` directly into its existing `defineConfig({...})` call rather than creating a second config file.

   The root `vitest.config.ts` with `projects` is still worth keeping — it's what lets you run `vitest` once from the repo root and get every package's tests in one process, a genuinely different and useful invocation from `turbo run test`'s per-package dispatch. The two are complementary, not redundant, once each package is self-contained.

**✅ Checkpoint** — root `vitest.config.ts` exists, every `browser-sdk` package plus `dashboard` has its own `passWithNoTests: true`, `vitest` is in the root `package.json`'s devDependencies, and `pnpm test` reports **10 successful, 10 total** (5 builds + 5 tests, all trivially passing since there are no real tests yet — "trivially passing" now actually means passing, not erroring). This is the last new file this guide introduces — everything from here is running what already exists, not creating anything new.

```bash
pnpm test 2>&1 | tail -5   # expect "Tasks: N successful, N total" with zero failures
```

---

## 10. Verification Checklist

This is the "did the whole journey actually work, end to end" pass — not a new step, just re-running the highest-value command from each of §3 through §9 in sequence, now that everything downstream of §5's schema package actually exists. If you've been running each section's own checkpoint as you went, this should be a formality; if you skipped ahead or came back after a break, this is what actually tells you whether it's safe to start writing real service logic. Run these in order; each should complete without error before moving to the next:

- [ ] `go build ./...` — every service compiles (even with empty `main()` bodies)
- [ ] `golangci-lint run ./...` — clean on the empty skeleton
- [ ] `go generate ./...` — regenerates `packages/domain-types`'s event-schema types from the JSON Schema files
- [ ] `pnpm install` — resolves the workspace with no errors
- [ ] `pnpm build` (via Turborepo) — builds `packages/event-schemas` before `dashboard`/`browser-sdk`, per the `^build` dependency graph in §4
- [ ] `pnpm lint` (`biome check .`) — clean on the scaffolded packages
- [ ] `pnpm test` / `go test ./...` — both pass (trivially, on empty packages)
- [ ] Manually diff one generated Go struct against its JSON Schema source and the Zod schema's inferred TS type — confirm all three actually agree, per §5's note

---

## 11. What's Deliberately Not Covered Here

- **Docker/Docker Compose manifests** (`deploy/compose/`) — per [`architecture.md §9`](architecture.md), these come after there's real service code to containerize, not before. The directory exists (§2); the manifests are a follow-up guide.
- **CI pipeline configuration** — per [Release Process](../research/11-engineering/README.md#release-process) and [CI/CD](../research/12-infrastructure/README.md#cicd), the pipeline stages are specified; the actual GitHub Actions/GitLab CI YAML is an implementation detail once this skeleton exists to run CI against.
- **Cosign key generation for image signing** — a release-time concern, not a repo-bootstrap one; revisit when the first image actually needs signing.
- **Database migrations content** — `sqlc`'s `migrations/` directory exists (§6) but is empty until [Postgres Schema](../research/05-data/README.md#postgres-schema)'s tables are actually written as SQL.

---

## What This Feeds

Real code. This is the last document in the planning chain — `prd.md` → `user-stories.md` → `architecture.md` → `coding-standards.md`/`design-system.md` → this guide → services with actual logic in them.
