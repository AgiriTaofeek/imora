# Diagrams

> Status: Decision recorded here rather than a growing set of exported image files.

Every diagram referenced throughout `docs/04-architecture/` (system context, container, component, and sequence diagrams) is maintained as **inline Mermaid, embedded directly in the markdown file it illustrates** — not as a separate source file exported here and linked in. See [system-context.md](../docs/04-architecture/system-context.md), [container-diagrams.md](../docs/04-architecture/container-diagrams.md), [component-diagrams.md](../docs/04-architecture/component-diagrams.md), and [sequence-diagrams.md](../docs/04-architecture/sequence-diagrams.md).

## Why Inline, Not a Separate Export

- **It renders in place.** GitHub, GitLab, and most documentation tooling render Mermaid natively — a reader sees the diagram next to the prose explaining it, not a link to a separate file that may or may not still match.
- **It stays in sync by construction.** A diagram exported as an image has to be manually regenerated every time the architecture changes; Mermaid text in the same file as the prose gets updated in the same edit, or the mismatch is visible in the same diff.
- **It's diffable.** Reviewing an architecture change means reviewing a text diff of the diagram itself, not an opaque binary image replacement.

This folder exists for anything that doesn't fit that pattern — a genuinely complex diagram Mermaid can't reasonably express, or a visual design asset for [11-design/](../docs/11-design/) — not as the default home for architecture diagrams.
