---
kind: ruleset
slug: kit-capabilities-usage
description: Teaches downstream Kit-managed projects to use kit capabilities for command discovery without maintaining Kit's internal command catalog.
status: active
registry_scope: downstream
applies_to:
  - kit
  - cli
  - command-discovery
  - coding-agent
read_policy_default: conditional
---

# Ruleset: kit-capabilities-usage

## Purpose

- Help coding agents in Kit-managed projects discover the correct Kit command before acting.
- Keep downstream project instructions small by using `kit capabilities` for targeted command guidance.
- Separate downstream command usage from Kit maintainer work on the internal command catalog.

## Applies When

- A coding agent is working in a Kit-managed downstream project and is unsure which Kit command, subcommand, flag, or alias applies.
- A task involves choosing between Kit commands such as `kit map`, `kit check`, `kit legacy verify`, `kit ci`, `kit pr fix`, `kit dispatch`, `kit loop review`, or `kit rules`.
- Project docs, scripts, or prompts mention Kit command behavior and should avoid stale assumptions.

## Rules

- Use `kit capabilities` for command discovery when Kit command choice, mutation behavior, network use, file writes, or git behavior is uncertain.
- Prefer targeted JSON after narrowing the command:
  - `kit capabilities <command> --json`
  - `kit capabilities rules add --json`
  - `kit capabilities skill mine --json`
- Use `kit capabilities --search <term> --json` for compact workflow discovery.
- Use `kit capabilities --full --json` only when hidden or deprecated compatibility command metadata is specifically needed.
- Treat `kit capabilities` as read-only command metadata: it does not load `.kit.yaml`, write project files, call the network, execute subprocesses, or mutate git.
- Do not maintain Kit's internal command catalog from a downstream project.
- If downstream project guidance conflicts with `kit capabilities`, prefer the live command metadata for command behavior and update the downstream documentation.
- If the Kit command catalog appears stale or wrong, fix it in the `jamesonstone/kit` repository or report it upstream.

## Anti-Patterns

- Do not guess Kit command flags from memory when `kit capabilities <command> --json` can answer the question.
- Do not run `kit capabilities --full --json` repeatedly as persistent context.
- Do not copy Kit command contracts into downstream always-loaded instruction files when a targeted lookup would be enough.
- Do not tell downstream projects to edit `pkg/cli/capabilities_catalog.go`.
- Do not treat `kit capabilities` as a project policy source; use repo-local docs and rulesets for project policy.

## Verification

Before completing downstream instruction or prompt work that references Kit commands, verify:

- `kit capabilities --search <term> --json` discovers the relevant Kit command.
- `kit capabilities <command> --json` accurately describes mutation, network, file-write, git, flag, example, and caveat behavior.
- Downstream project docs use `kit capabilities` for command discovery and do not require editing Kit source files.
- `kit init --refresh` installs or refreshes this usage rule without installing the maintainer-only `command-capabilities` rule.

## Examples

Choosing a command:

```bash
kit capabilities --search "pr review repair" --json
kit capabilities pr fix --json
kit capabilities loop review --json
```

Checking command safety before use:

```bash
kit capabilities dispatch --json
```

Refreshing an existing downstream project:

```bash
kit init --refresh
```
