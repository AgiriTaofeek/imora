# Imora — Agent Instructions

## Read This First

Docs are this project's source of truth — there is no PM, so these are authoritative, not
aspirational:

- [`docs/architecture.md`](docs/architecture.md) — system design, tech stack
- [`docs/coding-standards.md`](docs/coding-standards.md) — Go + TypeScript conventions (naming,
  error handling, testing style, etc.) — don't restate these here, read the file
- [`docs/design-system.md`](docs/design-system.md) — package boundaries, service structure
- [`docs/setup-guide.md`](docs/setup-guide.md) — how the monorepo was bootstrapped, and why
- [`research/11-engineering/architecture-decisions/`](research/11-engineering/architecture-decisions/) — ADRs; check here before assuming a technical decision is still open

## Commands

- `task check` — lint + test + build, both languages (the closest thing to a CI script)
- `task build` / `task test` / `task lint` / `task vet` / `task generate`
- `task dev:go SERVICE=<name>` — hot-reload a Go service (`air`); defaults to `gateway`
- `task dev:ts` — dashboard dev server (Vite, HMR)
- `task --list` for the full set

## Hard Pins — Do Not Silently Deviate

These were deliberate decisions, not defaults. If a task seems to require deviating from one,
say so and ask — don't just do it:

- Router: **chi** (`github.com/go-chi/chi/v5`) only — not gin/echo/fiber
- Lint/format (TS): **Biome** only — not ESLint/Prettier
- TypeScript pinned to **6.x** — not 7.x yet (blocked on `vue-tsc`/Angular tooling support;
  revisit ~TS 7.1, see `docs/setup-guide.md`)
- **One `go.mod`** at the repo root — not Go workspaces (`go.work`); see ADR 0003
- **pnpm 11+ / Node 22+** required, not just recommended
- Every package's license: **AGPL-3.0-only** — see ADR 0001
- No fabricated dates or timelines in docs, changelogs, or roadmap content — see `README.md`

## Verify Before Trusting a Diagnostic

IDE-reported errors during an edit can be stale (seen repeatedly on this project: phantom CSS
parse errors, a bogus "illegal return statement," a false duplicate-key error). Before acting on
one, re-check with the real tool — `go build`, `golangci-lint run ./...`, `biome check .` — and
trust that over the squiggly line.

## Engineering Discipline

This project follows the `no-ai-slop` skill (`.agents/skills/no-ai-slop/SKILL.md`) for every
AI-assisted change. Apply it — don't just know it exists.

**Invoke it explicitly** before implementing any non-trivial feature or change: confirm the
problem, constraints, and scope before writing code. Don't start typing on a vague request —
ask, or state the assumptions you're proceeding under.

**Pseudocode before real code, for any non-trivial logic.** Present interfaces and control flow
as commented pseudocode first, get explicit sign-off, then translate to real code. Don't skip to
a finished diff and treat review-after-the-fact as equivalent — the point is catching a wrong
decision (or a missing interface method, a false assumption) before it's real code, not after.

**Hard stops — never skip these, skill invoked or not:**
- Never install or recommend a dependency without confirming it actually exists on the real
  registry, is maintained, and its license is compatible. A plausible-sounding package name is
  not verification.
- Never state that an API, method, or config key exists without checking it against real docs
  or the installed version. If unverified, say so — don't present a guess as fact.
- Never invent architecture (system boundaries, services, data models) unprompted — propose,
  don't decide.
- Keep commits/diffs small and independently reviewable. Explain *why*, not just *what*.
- Before calling anything "done," run the Senior Engineer Checklist from the skill: understood
  every line, tested, observable, secure, has a rollback path, no unverified dependency or API.

If a request conflicts with these (e.g. "just add whatever library gets this working fastest"),
flag the conflict instead of silently complying.
