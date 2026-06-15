---
kind: ruleset
slug: command-capabilities
description: Requires Kit maintainers to update kit capabilities metadata whenever Kit command behavior changes.
status: active
registry_scope: kit-maintainer
applies_to:
  - kit
  - kit-maintainer
  - cli
  - command
  - capabilities
  - coding-agent
read_policy_default: must
---

# Ruleset: command-capabilities

## Purpose

- Keep `kit capabilities` as the authoritative command-discovery index for Kit's command surface.
- Ensure Kit maintainers update command metadata when Kit commands, subcommands, flags, aliases, or behavior extensions change.
- Prevent downstream Kit-managed projects from inheriting the maintainer obligation to edit Kit source files.

## Applies When

- Work is happening inside the `jamesonstone/kit` source repository.
- Any Kit command, subcommand, flag, alias, prompt surface, or command behavior is added, removed, renamed, deprecated, hidden, or materially changed.
- A Kit command starts or stops reading network data, writing files, executing subprocesses, mutating git, mutating GitHub, opening an editor, copying output, or writing generated state.
- Kit documentation describes a command behavior change that should also be discoverable by coding agents.

This rule is maintainer-only. Downstream Kit-managed projects should use `kit-capabilities-usage` instead.

## Rules

- Every Kit command or command extension must update `kit capabilities` in the same change.
- Update `pkg/cli/capabilities_catalog.go` whenever the command surface changes.
- For every visible canonical command, the capability catalog must describe:
  - command name and category
  - concise summary
  - mutation level
  - network behavior
  - file-write behavior
  - git mutation behavior
  - important flags
  - related commands
  - when to use the command
  - when not to use the command
  - examples
  - caveats when behavior is subtle or risky
- For hidden or deprecated commands, the catalog must mark the command as hidden or deprecated and explain the replacement.
- For aliases, the catalog must either include an alias on the canonical record or include a hidden/deprecated compatibility record, whichever best matches the command behavior.
- For flag-dependent behavior, explicitly distinguish default behavior from behavior enabled by flags.
- For risky flags, include the safety boundary in the flag metadata.
- Keep the JSON schema stable unless a schema change is intentional and tested.
- Human text output should expose enough best-practice guidance for a coding agent to choose the command without requiring a separate README scan.
- Kit's agent-facing docs must instruct agents to use `kit capabilities <command> --json` for targeted command understanding and avoid relying on stale memory.
- Do not export this maintainer rule into downstream projects through the registry; downstream projects should receive usage guidance, not an instruction to edit Kit internals.

## Verification

Before completing a Kit command-surface change, verify:

- The changed command or flag appears in `kit capabilities` output.
- `kit capabilities <command> --json` includes accurate guidance for behavior, mutation, network, file writes, git mutation, flags, related commands, examples, and caveats where applicable.
- `kit capabilities --search <term> --json` can discover the command by its key workflow terms.
- Root help and capabilities metadata agree about visible commands.
- Tests cover command-catalog drift so visible commands cannot be added without capability metadata.
- Downstream registry refresh installs `kit-capabilities-usage` and does not install this maintainer-only rule.

Recommended commands:

```bash
go test ./pkg/cli -run TestCapabilities
go run ./cmd/kit capabilities <command> --json
go run ./cmd/kit capabilities --search <term> --json
```

## Examples

Adding a new command:

```text
Inside the Kit repository, if `kit publish` is added, add a `publish` capability record that documents
whether it reads GitHub, writes files, mutates git, opens an editor, and what
flags change those safety properties.
```

Extending an existing command:

```text
Inside the Kit repository, if `kit dispatch --resolve` is added, update the `dispatch` capability record
with the new flag, its GitHub mutation boundary, required confirmation flags,
examples, caveats, and related-command guidance.
```

## Anti-Patterns

- Do not add a command and leave `kit capabilities` stale.
- Do not rely on README, root help, or generated agent docs as the only command-discovery source.
- Do not describe command behavior in prose that conflicts with capability metadata.
- Do not hide network, file-write, git, GitHub, editor, clipboard, or subprocess side effects.
- Do not add risky flags without documenting their safety boundary.
- Do not tell downstream projects to edit `pkg/cli/capabilities_catalog.go`.
