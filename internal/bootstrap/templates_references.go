package bootstrap

const referencesReadmeStarter = `# References

This directory holds durable repository-wide context that is broader than one
feature. Keep project truth evidence-backed and place registry-managed rules
under ` + "`rules/`" + ` and workflows under ` + "`workflows/`" + `.

- ` + "`testing.md`" + ` — verified validation commands, environments, and evidence.
- ` + "`tooling.md`" + ` — verified toolchain and command references.
- ` + "`external-systems.md`" + ` — verified recurring integrations.
- ` + "`worktrees.md`" + ` — native Git worktree policy and local conventions.
`

const testingStarter = `# Testing Reference

## Code-Level Validation

| Layer | Command | PR check | Required | Notes |
| --- | --- | --- | --- | --- |
| Unverified | Populate from repository evidence | Unverified | unknown | Do not guess commands. |

## Environments and High-Level Suites

No repository-specific environment or end-to-end command has been verified.

## Credentials and Test Data

List credential names without values only when verified. Never copy secrets.

## Known Gaps

- Repository-bootstrap evidence review is pending.
`

const projectToolingStarter = `# Tooling Reference

## Verified Toolchain

No project-specific toolchain evidence has been recorded yet.

## Verified Commands

No repository-native commands have been recorded yet. Keep ` + "`Makefile`" + ` as
the safe help-only starter until commands are proven.
`

const externalSystemsStarter = `# External Systems Reference

## Verified Integrations

No recurring external integration has been verified yet. Do not guess
credentials, endpoints, owners, environments, or production state.
`

const worktreesStarter = `# Git Worktrees

Use native ` + "`git worktree`" + ` as the policy authority. Keep the primary checkout
on the protected default branch and use one exact writable issue branch per
implementation lane. Detached pull-request worktrees are inspection-only.

Preserve dirty or ambiguous work, never stash/reset/clean implicitly, and read
the resolved work-lane and safety rules before creating or removing a lane.
`
