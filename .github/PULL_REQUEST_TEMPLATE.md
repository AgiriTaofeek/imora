## What changed and why

<!-- Not just what — why. Tie back to a design doc, ADR, or issue if one exists. -->

## Senior Engineer Checklist

Per [`no-ai-slop`](../.agents/skills/no-ai-slop/SKILL.md) — every box should be true before this
is ready for review, not just before merge:

- [ ] I understand every line of this diff and can explain every decision in it
- [ ] This is the simplest solution I found, not just the first one that worked
- [ ] It's tested (unit, and integration if it crosses a service boundary)
- [ ] It's observable (structured logging / metrics where relevant)
- [ ] It's secure (auth, input validation, no secrets committed)
- [ ] Every dependency and API call in this diff was independently verified to actually exist —
      not just assumed because an AI suggested it
- [ ] I know how to roll this back if it's wrong
- [ ] The diff is small enough for a reviewer to hold in their head

## Test plan

<!-- How was this actually verified? Commands run, not just "should work." -->
