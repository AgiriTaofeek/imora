# Contributing to Imora

Guidelines for contributing code, documentation, and design changes to the Imora project.

## Branching and Development Model

Trunk-based development — one `main`, short-lived branches (hours to a couple of days, not
weeks). Rationale: [`research/11-engineering/README.md#branching-strategy`](research/11-engineering/README.md#branching-strategy).
Don't re-litigate this per PR; it's a settled decision, not a default.

### Branch naming

`<type>/<short-description>`, matching what the change actually is:

| Prefix | For |
|---|---|
| `feat/` | New functionality — e.g. `feat/query-api-session-endpoint` |
| `fix/` | Bug fixes — e.g. `fix/gateway-nil-pointer` |
| `chore/` | Tooling, deps, config — e.g. `chore/bump-turbo` |
| `docs/` | Docs-only changes — e.g. `docs/api-reference` |
| `refactor/` | No behavior change, just structure |

## The Full Workflow

1. **Branch from `main`**, using the naming convention above:
   ```bash
   git checkout main && git pull
   git checkout -b feat/your-change
   ```

2. **Make the change, verify locally before pushing:**
   ```bash
   task check   # lint + test + build, both languages — exactly what CI runs
   ```
   See [`docs/setup-guide.md`](docs/setup-guide.md) if `task` or the toolchain isn't set up yet.

3. **Commit and push:**
   ```bash
   git add <files>
   git commit -m "Add session lookup endpoint to query-api"
   git push -u origin feat/your-change
   ```

4. **Open a PR** (`gh pr create` or the link `git push` prints). The PR template embeds the
   [`no-ai-slop`](.agents/skills/no-ai-slop/SKILL.md) Senior Engineer Checklist — actually work
   through it, don't just leave it unchecked.

5. **What happens automatically:**
   - CI (`.github/workflows/ci.yml`) runs `task check` on the PR branch. The merge button stays
     disabled until it's green.
   - Every review conversation thread must be marked resolved before merging.
   - `CODEOWNERS` routes review to the right person by path once there's more than one
     contributor (see "Solo vs. Team Mode" below for the current state).
   - If `main` moves ahead while your PR is open, GitHub will ask you to update your branch
     before merging (`strict` status checks) — this keeps the merged result actually tested, not
     just the pre-rebase version.

6. **Merge**: "Squash and merge" is the only option — keeps `main` at one commit per logical
   change. `allow_merge_commit`/`allow_rebase_merge` are both disabled repo-wide.

7. **Clean up:**
   ```bash
   git checkout main && git pull
   git branch -d feat/your-change   # GitHub auto-deletes the remote branch on merge
   ```

## Hotfixes (the one exception)

For a critical patch against an already-*released* version (not `main`): branch from that
release's tag, cherry-pick the fix, tag a patch release, delete the branch. Different from the
flow above — see
[`research/11-engineering/README.md#branching-strategy`](research/11-engineering/README.md#branching-strategy)
for why this is a deliberately separate, short-lived exception rather than a standing
release-maintenance branch.

## Solo vs. Team Mode

Branch protection on `main` currently requires: PR (no direct pushes), CI passing, conversation
resolution, squash-only merges — but **not** a human approval, since with one maintainer that
requirement is structurally impossible to satisfy (GitHub never lets a PR author approve their
own PR). The moment a second contributor joins:

1. Replace the relevant `@AgiriTaofeek` placeholders in [`.github/CODEOWNERS`](.github/CODEOWNERS)
   with their real handle, scoped to the paths they own.
2. Re-enable required review — `required_approving_review_count: 1` and
   `require_code_owner_reviews: true` on the branch protection rule (do both together, not
   separately, since routing without a required count — or the reverse — doesn't do anything
   useful on its own).

## Code Review Standards

Every PR should satisfy the [`no-ai-slop`](.agents/skills/no-ai-slop/SKILL.md) Senior Engineer
Checklist embedded in the PR template — this applies equally whether the code was written by a
human, AI-assisted, or both. See [`CLAUDE.md`](CLAUDE.md) for the hard pins (tech stack, license,
tooling) that should never be silently deviated from in a PR without flagging it first.
