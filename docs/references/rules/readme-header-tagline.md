---
kind: ruleset
slug: readme-header-tagline
description: Standardizes top-level README openings with a Flowcore-style header, tagline, and concise product paragraph.
status: active
registry_scope: downstream
applies_to:
  - readme
  - documentation
  - branding
  - repo-onboarding
  - coding-agent
read_policy_default: conditional
---

# Ruleset: readme-header-tagline

## Purpose

- Keep top-level project READMEs immediately recognizable, skimmable, and useful to humans and coding agents.
- Preserve a consistent Kit-managed opening pattern across repositories.
- Replace one-off README style instructions with durable repo-local guidance.

## Applies When

- Creating or materially updating a top-level `README.md`.
- Initializing or refreshing a Kit-managed project that needs README style guidance.
- A user asks for a README to follow the same header and tagline pattern as Flowcore.
- A coding agent is deciding how to structure the first screen of a repository README.

## Rules

- When creating or rewriting the top-level `README.md` opening, use this exact order:
  - fenced `text` header block
  - one blank line
  - one plain-language product paragraph
  - compact badge/status-badge line or block when badges already exist, are requested, or are managed by Kit
  - subsequent sections such as boundaries, setup, API, architecture, or contribution guidance
  - final `## Maintainers` section
- The fenced header block is required and must be the first substantial project-identity content.
- The header block must use this shape:

````markdown
```text
<PROJECT WORDMARK OR COMPACT BANNER>

                         <short tagline>
```
````

- The wordmark should be an ASCII-art banner when a suitable one is already available or easy to produce cleanly.
- If an ASCII-art banner is not appropriate, use a compact uppercase text wordmark using the project or product name.
- The tagline must be the final non-empty line inside the fenced block.
- The tagline should be short enough to scan at a glance, usually 3 to 10 words.
- Visually offset the tagline to the right when the wordmark is wide, matching the Flowcore pattern.
- Do not put the tagline in a separate H1, paragraph, badge, or blockquote.
- Follow the header block with one plain-language paragraph that starts with the project or product name and explains what it is, who it serves, and what it does.
- Keep the opening paragraph concrete. Prefer product, service, workflow, and ownership nouns over generic adjectives.
- Badges and status badges are allowed near the top of the README when they are useful and compact.
- For Kit-managed public repositories, include the default out-of-the-box badge set unless the user explicitly declines it:
  - Last commit: `https://img.shields.io/github/last-commit/<owner>/<repo>`
  - Open issues: `https://img.shields.io/github/issues/<owner>/<repo>`
  - Pull requests: `https://img.shields.io/github/issues-pr/<owner>/<repo>`
  - Release: `https://img.shields.io/github/v/release/<owner>/<repo>`
  - CI: `https://github.com/<owner>/<repo>/actions/workflows/<workflow>.yml/badge.svg` only when a conventional CI workflow exists, such as `ci.yml`, `ci.yaml`, `test.yml`, `test.yaml`, `build.yml`, or `build.yaml`
- For Kit-managed private repositories, do not include public Shields GitHub metadata badges because public Shields cannot authenticate to private GitHub repository metadata. Include only the native GitHub Actions workflow badge when a conventional workflow exists.
- Do not include a License badge in the default Kit-managed badge set.
- When Kit-managed refresh updates badges, preserve the Kit marker comments around the badge block and update only the block between those markers.
- Keep installation details, long architecture notes, changelogs, and contribution guidance after the header, tagline, and opening paragraph.
- Every top-level `README.md` must end with a `## Maintainers` section as the last H2 header.
- The default Kit-managed maintainer copy is:

```markdown
## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
```

- Use `## Maintainers`, not `## Maintainer`, even when there is currently only one maintainer.
- Preserve existing stronger brand guidance when a repository already has an intentional, current README identity.
- Do not invent product claims, operational capabilities, compliance posture, customer names, or integrations that are not supported by repository evidence.

## Anti-Patterns

- Letting badges, status badges, installation commands, or a generic H1 replace the header, tagline, or concrete opening paragraph.
- Treating badges as a substitute for the fenced identity block or product paragraph.
- Adding a default License badge when the project did not explicitly request one.
- Adding public Shields GitHub metadata badges to private repositories.
- Editing outside the Kit-managed badge marker block when the only requested change is refreshing default badges.
- Leaving `## Maintainer` singular or any other H2 section after `## Maintainers`.
- Using a decorative banner without a tagline.
- Putting the tagline outside the fenced `text` block.
- Writing a tagline that is vague, marketing-only, or disconnected from what the repository actually does.
- Burying the project purpose below setup instructions or implementation details.
- Copying Flowcore-specific domain language into unrelated repositories.
- Replacing a repository's intentional brand system with this pattern when the repo already has a current, deliberate README style.

## Verification

Before completing README work, verify:

- `README.md` makes a fenced `text` block the first substantial content.
- The fenced block contains the project name or wordmark.
- The fenced block includes a concise tagline on its final non-empty line.
- The tagline is visually offset inside the fenced block when the wordmark is wide.
- The first paragraph after the block starts with the project or product name and explains the repository in concrete terms.
- Badges and status badges, when present, do not replace or obscure the identity opening.
- Kit-managed public-repository badge blocks include Last commit, Open issues, Pull requests, Release, and a CI badge only when a conventional CI workflow exists.
- Kit-managed private-repository badge blocks exclude public Shields GitHub metadata badges and include only a native CI workflow badge when a conventional workflow exists.
- Kit-managed badge blocks do not include a License badge by default.
- `README.md` ends with `## Maintainers` as the last H2 header.
- The Maintainers section includes [Jameson](https://github.com/jamesonstone) and the `jamesonstone` GitHub username.
- The opening does not claim unsupported capabilities, integrations, users, or status.
- Setup, architecture, and contribution sections remain below the identity opening.

## Examples

Preferred ASCII-art opening:

````markdown
```text
EEEE  X   X  AAA  M   M PPPP  L     EEEE
E      X X  A   A MM MM P   P L     E
EEE     X   AAAAA M M M PPPP  L     EEE
E      X X  A   A M   M P     L     E
EEEE  X   X A   A M   M P     LLLL  EEEE

SSSS  EEEE RRRR  V   V III  CCCC EEEE
S     E    R   R V   V  I  C     E
SSSS  EEE  RRRR  V   V  I  C     EEE
   S  E    R  R   V V   I  C     E
SSSS  EEEE R   R   V   III  CCCC EEEE

                     event intake and routing for partner integrations
```

Example Service is the Go service that receives partner events, validates payloads, records delivery facts, and routes accepted work to downstream processors.

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/acme/example-service)](https://github.com/acme/example-service/commits) [![Open issues](https://img.shields.io/github/issues/acme/example-service)](https://github.com/acme/example-service/issues) [![Pull requests](https://img.shields.io/github/issues-pr/acme/example-service)](https://github.com/acme/example-service/pulls) [![CI](https://github.com/acme/example-service/actions/workflows/ci.yml/badge.svg)](https://github.com/acme/example-service/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/acme/example-service)](https://github.com/acme/example-service/releases)
<!-- END KIT-MANAGED README BADGES -->

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
````

Compact wordmark variant:

````markdown
```text
ACME WORKER

                         background processing for customer account jobs
```

Acme Worker runs account maintenance jobs, reconciles partner state, and emits operational facts for the Acme platform.

## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).
````
