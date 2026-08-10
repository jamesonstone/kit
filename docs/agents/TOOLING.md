# Agent Tooling

## Kit-owned operations

- `kit contract resolve`: deterministic local contract selection; no network,
  write, Git, subprocess, or model activity.
- `kit init`: validated full materialization for a fresh project.
- `kit reconcile`: read-only drift and migration planning by default; explicit
  apply for conflict-free changes.
- `kit registry list|view|add|status`: remote catalog inspection and typed
  artifact administration.
- `kit rules list|view|add`: ruleset-filtered registry administration.

Use command help and [`docs/commands.md`](../commands.md) for the exact current
surface. Do not infer removed Kit 1.x commands from historical specs.

## Agent-owned operations

Native coding-agent and repository tooling owns research, clarification,
planning, file edits, test execution, Git/GitHub delivery, review, CI repair,
cloud operations, and deployment. Follow selected repository rules for those
actions.

For `pr-feedback-repair`, the host should prefer status, review, or comment
wakeups. Its fallback is one bounded `gh` helper process with no model turns
while sleeping; Kit supplies only the resolved local workflow contract.

## Worktrees

Use native Git worktrees as the policy authority. The separate `git-wt` binary
is an optional unchanged convenience; it is not a Kit subcommand or a contract
dependency.
