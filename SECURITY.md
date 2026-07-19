# Security Policy

Imora is built for organizations that can't send customer session data to a third party —
security issues here are taken as seriously as the product's core premise depends on them being.

## Supported Versions

Imora is in early development; there are no tagged releases yet (see
[`research/08-roadmap/feature-roadmap.md`](research/08-roadmap/feature-roadmap.md)). Once
releases begin, this section will list which versions receive security patches, per the
versioning policy in
[`research/11-engineering/README.md#release-process`](research/11-engineering/README.md#release-process).
Until then, only `main` is supported.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for a security vulnerability.**

Instead, use GitHub's private vulnerability reporting: go to the
[Security tab](https://github.com/AgiriTaofeek/imora/security) → "Report a vulnerability". This
opens a private advisory visible only to maintainers until a fix is ready.

If you'd rather not use GitHub, email **<!-- REPLACE WITH YOUR PREFERRED SECURITY CONTACT -->**
with details. Please include:

- What you found and why it's a security issue, not just a bug
- Steps to reproduce, or a proof of concept if you have one
- The affected component (`services/*`, `browser-sdk`, `dashboard`, etc.)

## What to Expect

This is currently a solo-maintained project — there's no dedicated security team and no formal
SLA. As a good-faith target: an acknowledgment within a few days, and a fix or mitigation plan
before any public disclosure. If you don't hear back in a reasonable time, a follow-up nudge is
fair.

## Scope

Relevant background if you're evaluating this project's security surface:
[`research/07-security/README.md`](research/07-security/README.md) covers the authentication,
authorization, encryption, audit logging, and threat model this project is designed against.
Findings that contradict a documented guarantee there (e.g. the audit-log GRANT restrictions, PII
redaction boundaries) are especially high-value reports.
