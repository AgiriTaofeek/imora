---
name: no-ai-slop
description: >
  A disciplined engineering workflow for using AI to accelerate software development
  without sacrificing maintainability, correctness, security, or architectural quality.
  Use before starting any non-trivial AI-assisted feature or change (to design before
  coding, break work into small reviewable tasks, and gather the right context), during
  implementation (to verify dependencies/APIs actually exist, keep functions small, delay
  abstractions), and before merging (to run the Senior Engineer Checklist and catch
  AI-slop red flags — unverified dependencies, oversized diffs, missing tests, doc drift).
license: Internal — Imora project skill
metadata:
  version: 1.0.0
  author: Agiri Taofeek
  category: engineering-standards
  domain: software-engineering
  updated: 2026-07-19
  tags: [ai, code-review, engineering-standards, workflow, quality, security, dependencies]
---

# Building Software Without AI Slop

> **Purpose:** A disciplined engineering workflow for using AI to accelerate software development without sacrificing maintainability, correctness, security, or architectural quality.
>
> **Core Principle:** AI is an implementation assistant — not the architect, product owner, or senior engineer. Humans remain responsible for every technical decision.

---

# Philosophy

The difference between senior engineers and inexperienced AI-assisted developers is **not** whether they use AI.

Senior engineers use AI extensively.

The difference is that they have **engineering standards** that every AI-generated change must satisfy before it reaches production.

**Golden Rule**

> Never outsource thinking.
> Outsource typing.

AI should increase your velocity, **not replace your engineering judgment**.

**What "slop" actually means here:** not "code an AI wrote," but code that was *merged without anyone forming an opinion about it* — untested, unexplained, over-abstracted, or silently wrong in a way no one caught because no one was really looking. The goal of this skill is to make sure a human always forms that opinion first.

---

# Engineering Principles

## 1. Design Before Code

Never ask AI to start writing code immediately.

Every feature begins with answering:

- What problem are we solving?
- Who are the users?
- What are the functional requirements?
- What are the non-functional requirements?
- What assumptions exist?
- What constraints exist?
- What failure scenarios exist?
- What metrics define success?
- What is explicitly out of scope?

Create a lightweight design document first.

Example:

```text
Goal:
Build subscription payment processing.

Requirements:
- Stripe only
- Idempotent
- Retry failed webhooks
- PostgreSQL
- Horizontal scaling
- Audit logging
- P95 latency <300ms
```

Only after the design is complete should implementation begin.

---

## 2. Humans Own Architecture

Never let AI invent system architecture.

Architecture includes:

- System boundaries
- Services
- Modules
- APIs
- Data models
- Domain models
- Event flows
- Database design
- Ownership
- Security boundaries

AI may suggest improvements.

Humans make the final decisions.

---

## 3. Break Work Into Small Tasks

Large prompts create large mistakes.

Instead of:

> Build an authentication system.

Break it down.

Example roadmap:

```
Database schema
↓
Repository
↓
Business logic
↓
JWT implementation
↓
HTTP handlers
↓
Tests
↓
Documentation
```

Each task should be reviewable in less than five minutes.

---

## 4. Give AI Your Context, Not Just Your Prompt

AI defaults to generic conventions unless told otherwise — its own naming style, its own error-handling idioms, its own folder layout. On an existing codebase, that mismatch is a bigger source of slop than any single bad function.

Before asking for code, provide:

- The relevant existing files, not just a description of them
- Your style guide or lint config
- A pointer to a similar, already-approved piece of code to mirror
- Naming and error-handling conventions already in use

Ask AI to match what's there before asking it to write what's new. A perfectly good function that ignores your codebase's conventions is still slop — it just looks clean in isolation.

---

## 5. Every Generated Line Has an Owner

If AI writes 1,000 lines,

You own 1,000 lines.

You should be able to explain:

- Why it exists
- Why this algorithm
- Why this package
- Complexity
- Failure modes
- Tradeoffs
- Security implications

If you cannot explain it,

Do not merge it.

---

## 6. Read More Than You Generate

Reading code creates better engineers.

For every AI-generated file ask:

- Can this be simpler?
- Can this be deleted?
- Is naming clear?
- Is the control flow obvious?
- Does this duplicate logic?
- Is this abstraction justified?

Treat AI output as a pull request.

---

## 7. Prefer Boring Technology

Choose technology with a long operational history.

Prefer:

- PostgreSQL
- Redis
- Go Standard Library
- React
- Nginx
- Docker
- Kubernetes

Avoid introducing dependencies because AI suggested them.

Ask:

> Can the standard library solve this?

---

## 8. Verify Every Dependency Before You Install It

AI regularly recommends packages that don't exist, are abandoned, or are typosquats of real packages — a known attack pattern ("slopsquatting") that relies on people trusting AI-suggested package names without checking. AI has no way to know if a package it names is real, maintained, or safe; it's pattern-matching plausible names.

Before running `install`:

- Confirm the package actually exists on the real registry (npm, PyPI, crates.io, etc.), not just that AI said so
- Check maintenance activity, download counts, and open security advisories
- Check the license is compatible with your project
- Prefer packages you or a teammate has used before over ones only AI has "heard of"

This is a stricter version of "minimize dependencies" (below) — it's specifically about never trusting AI as the source of truth for whether a package is safe to install.

---

## 9. Delay Abstractions

One implementation does not need an interface.

Avoid:

```
PaymentProvider
RepositoryFactory
NotificationManager
AbstractService
```

until multiple implementations actually exist.

Follow the Rule of Three:

Only abstract after seeing repeated patterns three times.

---

## 10. Tests Are the Specification

Whenever possible:

Write tests before implementation.

Prompt AI with:

```
These are the tests.

Implement the code that satisfies them.
```

This keeps implementation aligned with expected behavior.

---

## 11. Make AI Explain Itself

Never stop after code generation.

Ask AI:

- Why this algorithm?
- Time complexity?
- Space complexity?
- Edge cases?
- Failure scenarios?
- Race conditions?
- Alternative implementations?
- Why not use another approach?
- Is any part of this a guess — an API, a config key, a method signature you're not fully certain exists?

Weak explanations often reveal weak implementations. A confident-sounding explanation for a hallucinated API is still a hallucination — verify against real docs, don't just accept a fluent answer.

---

## 12. Prefer Explicit Code

Code should optimize for readability.

Avoid clever abstractions that save a few lines while increasing cognitive load.

Future engineers should understand the code within minutes.

---

## 13. Minimize Dependencies

Every dependency introduces:

- Security risks
- Upgrade costs
- Breaking changes
- Supply chain attacks
- Maintenance burden

Ask:

> Can I implement this in under 50 lines?

before adding another package.

---

## 14. Design for Deletion

Every feature should be removable.

Good modularity means features can disappear without affecting unrelated parts of the system.

Loose coupling is a feature.

---

## 15. Measure Before Optimizing

Never optimize because AI predicts a bottleneck.

Instead:

Measure.

Profile.

Benchmark.

Optimize only proven bottlenecks.

Remember:

Premature optimization is still one of the largest sources of unnecessary complexity.

---

## 16. Use AI as a Reviewer

AI often provides better value reviewing code than writing it.

Prompt:

```
Review this code like a Staff Engineer.

Focus on:

- Security
- Maintainability
- Performance
- Scalability
- Concurrency
- API design

Do not rewrite it.
Only identify problems.
```

---

## 17. Documentation Comes First

Before implementation create:

- README
- Architecture diagrams
- API contracts
- Database schema
- ADRs (Architecture Decision Records)
- Sequence diagrams
- Deployment plan

Documentation reduces implementation mistakes.

Watch for **doc drift**: AI will happily generate documentation that describes what the code *should* do rather than what it *actually* does, especially after a quick follow-up edit. Treat AI-generated docs as a draft to verify against the real implementation, not a trusted record of it.

---

## 18. Define "Done"

A feature is not complete until it includes:

- Implementation
- Unit tests
- Integration tests
- Logging
- Metrics
- Error handling
- Documentation
- Security review
- Performance review
- Code review
- Rollback strategy

Working code is not finished code.

---

## 19. Make Changes Reversible in Production

Every AI-assisted change should ship with a way to undo it without a redeploy.

- Put new behavior behind a feature flag when risk is non-trivial
- Roll out to a small percentage of traffic before 100%
- Know, before you ship, exactly what "roll back" means for this change — flag flip, revert commit, or data migration undo
- Treat "we can turn it off in seconds" as part of the definition of done, not an afterthought

AI-generated code is not inherently riskier at runtime than human-written code, but it's usually *less battle-tested by the person shipping it* — which makes fast reversibility more valuable, not less.

---

## 20. Keep Humans Responsible

AI is responsible for:

- Boilerplate
- Refactoring
- Test generation
- Documentation drafts
- Code explanation
- Bug finding

Humans remain responsible for:

- Product decisions
- Architecture
- Security
- Performance
- Tradeoffs
- Code ownership

---

## 21. Optimize for Simplicity

Simple systems survive.

Simple code:

- Has fewer bugs.
- Is easier to test.
- Is easier to delete.
- Is easier to extend.
- Is easier to onboard new engineers to.

Complexity must always justify itself.

---

## 22. Favor Composition Over Inheritance

Prefer small, composable components over deep inheritance hierarchies.

Good software grows by combining simple pieces rather than extending rigid class trees.

Examples:

- Middleware pipelines
- Small interfaces
- Functional composition
- Dependency injection

---

## 23. Keep Functions Small and Focused

A function should generally do one thing well.

Warning signs:

- Many nested `if` statements
- Multiple unrelated responsibilities
- Excessive parameters
- Difficult naming

Small functions are easier to test, reuse, and reason about.

---

## 24. Make Illegal States Unrepresentable

Design types and APIs so invalid states cannot exist.

Examples:

- Use enums instead of strings where possible.
- Model domain concepts explicitly.
- Validate input at system boundaries.
- Avoid nullable values when they should never be null.

Push correctness into the type system whenever practical.

---

## 25. Fail Fast

Detect invalid input or impossible conditions as early as possible.

Avoid silently ignoring errors.

Good systems fail loudly during development rather than producing hidden bugs in production.

---

## 26. Design for Observability

Every production system should answer:

- What happened?
- When?
- Why?
- Who was affected?
- How often?

Include:

- Structured logging
- Metrics
- Tracing
- Health checks
- Correlation IDs
- Meaningful error messages

If you cannot observe a system, you cannot operate it.

---

## 27. Consider Security From Day One

Never bolt security onto finished software.

Review every feature for:

- Authentication
- Authorization
- Input validation
- Output encoding
- Rate limiting
- Secret management
- Dependency vulnerabilities
- Least privilege
- Audit logging

Security is a design concern, not a final checklist.

**AI-specific security checks worth adding explicitly:**

- Scan for secrets AI may have echoed back from context (API keys, tokens pasted into a prompt earlier in the session)
- Never let AI choose crypto primitives or auth logic unreviewed — these are exactly the places where plausible-looking code is most dangerous
- Re-check dependency licenses and provenance (see Principle 8) as part of the security pass, not a separate one

---

## 28. Check Licensing and Provenance

AI models are trained on large amounts of code, some of it under licenses (GPL, non-commercial, etc.) that don't match your project's license.

- Be wary of AI output that reproduces a distinctive, non-trivial block near-verbatim — that's a sign it's closely mirroring a specific source rather than generating fresh code
- For anything that looks unusually specific or stylistically distinct from the rest of the codebase, ask AI directly whether it recognizes the pattern from a known project or library
- This is a bigger risk for large, distinctive algorithms than for boilerplate or standard patterns

Not a substitute for legal review on anything commercially sensitive — just a first-pass habit.

---

## 29. Small, Atomic, Well-Described Commits

Small tasks (Principle 3) only help if they turn into small commits.

- One logical change per commit
- Commit message explains *why*, not just *what* — tie it back to the design doc or acceptance criteria
- Keep PRs small enough that a reviewer can hold the whole diff in their head
- Never let AI batch multiple unrelated changes into one commit because it was efficient to generate them together

A large, AI-generated diff is exactly where slop hides — reviewers skim large diffs and read small ones.

---

## 30. Match Your Caution to the Codebase

AI is not equally safe everywhere.

- **Greenfield code**: lower risk. There's no existing behavior to accidentally break, and conventions are still being established.
- **Brownfield / legacy code**: higher risk. AI has no tacit knowledge of *why* a piece of code is the way it is — a workaround for a bug fixed elsewhere, a compliance requirement, a scar from a past incident. It will confidently "clean up" code that was ugly for a reason.

In legacy systems, ask AI to explain what a piece of code currently does and why it might be structured that way *before* asking it to change it. Treat any AI suggestion to delete or simplify old code as a hypothesis to verify with the code's history (git blame, related tickets), not a fact.

---

## 31. Measure Whether the AI Workflow Itself Is Working

"Measure before optimizing" (Principle 15) applies to your AI-assisted process too, not just runtime performance.

Track, informally or formally:

- Revert / rollback rate on AI-assisted changes vs. human-only changes
- Bug density in AI-authored code after it ships
- How long review actually takes on AI-generated PRs vs. hand-written ones

If AI-assisted work is reverted more often or takes longer to review than it saves to write, that's a signal to tighten the process (smaller tasks, more context, stricter review) — not a reason to stop using AI, but a reason to stop assuming the workflow is working just because it feels faster.

---

# Practical AI Workflow

For every feature:

1. Understand the business problem.
2. Write a design document.
3. Identify constraints.
4. Design architecture.
5. Define interfaces.
6. Design the database.
7. Define API contracts.
8. Write acceptance criteria.
9. Write tests.
10. Gather relevant existing code, conventions, and style guides to give AI as context.
11. Ask AI to implement one small task.
12. Read every generated line.
13. Verify any dependency, API, or library AI referenced actually exists as described.
14. Run tests.
15. Review with AI.
16. Refactor.
17. Commit a small, well-described change.
18. Repeat.

---

# AI Prompting Guidelines

Good prompts include:

- Context
- Constraints
- Existing architecture
- Coding standards
- Acceptance criteria
- Performance requirements
- Security requirements
- Expected output

Bad prompts:

```
Build a payment service.
```

Good prompts:

```
Implement the PaymentRepository interface.

Requirements:

- PostgreSQL
- Context-aware
- Connection pooling
- Unit-testable
- No ORM
- Parameterized SQL
- Handle duplicate payments
- Return domain errors
```

Specific prompts produce predictable code.

---

# Red Flags of AI Slop

Be suspicious when AI generates:

- Giant files (>500 lines)
- Deep inheritance
- Generic managers everywhere
- Over-engineered abstractions
- Many unnecessary packages
- Duplicate logic
- Poor naming
- Hidden side effects
- Missing tests
- Missing error handling
- No documentation
- Magic numbers
- Global mutable state
- A dependency, API, or config key you haven't independently verified exists
- Code that ignores the conventions already in the surrounding files
- A commit or PR too large for anyone to meaningfully review

---

# Senior Engineer Checklist

Before merging, ask:

- Do I understand every line?
- Can I explain every decision?
- Is this the simplest solution?
- Is it tested?
- Is it observable?
- Is it secure?
- Is it maintainable?
- Have I verified every dependency and API call actually exists as described?
- Do I know how to roll this back if it's wrong?
- Can another engineer understand it quickly?
- Can I delete it later?
- Would I proudly defend this in a code review?

If the answer to any of these is "no", the work is not ready.

---

# Core Mindset

> AI should amplify engineering excellence, not compensate for its absence.

The goal is not to generate more code.

The goal is to build software that remains understandable, maintainable, secure, and reliable years after it is written.

Good engineers build systems that survive.
Great engineers build systems that remain simple enough for others to improve.
