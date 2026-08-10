# Agent Tooling

## Kit-owned operations

- `kit contract resolve`: deterministic local contract selection; no network,
  write, Git, subprocess, or model activity.
- `kit init`: validated canonical repository bootstrap and evidence-gathering
  prompt; it never reads `.env` or infers project truth.
- `kit reconcile`: read-only drift and migration planning by default; explicit
  apply for conflict-free changes.
- `kit registry list|view|add|status`: remote catalog inspection and typed
  artifact administration.
- `kit rules list|view|add`: ruleset-filtered registry administration.
- `kit pr fix`: explicit GitHub feedback intake, exact writable-head lane
  preparation, supervisor-prompt output, and confirmed named-thread resolution;
  it never runs the repair or launches an agent.

Use command help and [`docs/commands.md`](../commands.md) for the exact current
surface. Do not infer removed Kit 1.x commands from historical specs.

## Agent-owned operations

Native coding-agent and repository tooling owns research, clarification,
planning, file edits, test execution, Git/GitHub delivery, review, CI repair,
cloud operations, and deployment. Follow selected repository rules for those
actions.

For `pr-feedback-repair`, prefer status, review, or comment host wakeups.
`kit pr fix --wait` is the bounded `gh` fallback with no model turns while
sleeping; an immediate invocation handles late or human feedback.

## Worktrees

Use native Git worktrees as the policy authority. The separate `git-wt` binary
is an optional unchanged convenience; it is not a Kit subcommand or a contract
dependency.
