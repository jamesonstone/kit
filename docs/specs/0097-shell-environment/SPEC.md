---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0097"
  slug: "shell-environment"
  dir: "0097-shell-environment"
---
# SPEC

## PURPOSE

Make shell-dependent agent commands use the intended project environment for each team member without embedding personal workstation settings in shared instructions (GH-278).

## CONTEXT

Kit ships contextual rules, renders one contract into three agent entry files, and preserves project-owned testing references. No rule currently directs shell-dependent commands through a verified project/host environment. A login shell does not necessarily load interactive configuration. Execution-policy denials are a separate failure.

## REQUIREMENTS

- Support an optional `.kit.yaml` `shell.command` argv prefix. Omitted/empty means the current user's initialized interactive-login environment; an explicit prefix selects the project invocation.
- Validate and preserve configuration, report its effective source, and test default/override/error/migration behavior without starting a shell from Kit.
- Document that functions execute inside their defining shell and only exported environment variables reach child processes.

- Resolve the project's command environment and the current user's approved host configuration before shell-dependent commands.
- Keep shared instructions portable and credentials private. Keep personal shell choices in user-level instructions or untracked local configuration.
- Preserve script interpreters, noninteractive CI/container environments, startup-file ownership, and permission boundaries.
- Verify shell mode and tool resolution with harmless probes. Report unavailable capabilities accurately.
- Install the rule through init and supported-generation reconcile. Preserve project customizations and reconciliation idempotence.
- Do not change personal/global instructions, startup files, security settings, credentials, runtime launchers, or downstream projects.

## ACCEPTED PLAN

1. Add a default downstream `shell-environment` contextual rule and universal-contract route.
2. Use the existing testing reference's Environment Preflights section as the project-owned record or pointer to existing environment documentation.
3. Add validated shell configuration and config-check reporting, then test defaults, exact overrides, invalid argv, and preservation.
4. Regenerate the three checked-in entry blocks and add migration, preservation, and portability regression tests.
5. Run Kit's full local validation and a fresh-context review, then publish a ready-for-review pull request through the issue/worktree lane.

## DECISIONS

- Added at the user's request: `shell.command` stores an executable/options prefix, not a command string or environment snapshot. Defaults resolve on the current executor; Kit never expands variables or runs the prefix.
- Configuration directs agent behavior and is validated by `kit config check`; it does not change Kit subprocesses or override explicit script/CI contracts. The source field in config-check output does not claim runtime readiness.

- Guidance, not a shell launcher: the execution host controls tools and permissions. Kit cannot guarantee host compliance.
- Reuse the existing project-owned reference rather than add a competing environment schema or copy startup files.
- Separate shell startup, project activation, authentication, and network policy. A wrapper is not a denied-action fallback.
- Describe shell examples conditionally. Do not assume zsh for every user, CI runner, or operating system.
- The contract's 1,500-word budget remains fixed. Shorten introductory prose without changing existing invariants to add the route.

## DISCOVERIES

- `kit status` and `kit reconcile --dry-run --diff` also propose pre-existing registry/badge refresh and Mint release-workflow upgrades. Those unrelated changes are left out of this shell-guidance PR.
- The rule registry discovers embedded Markdown automatically; no new CLI command or registry schema is necessary.
- YAML decoding can coerce numeric/boolean argv entries and drop null members. Raw-node validation rejects these cases before accepting the configuration; whole-command null remains the documented default.

## VALIDATION

All checks below were rerun after the shell-configuration extension and raw-argument validation fix.

- Passed focused shell-rule, universal-contract, managed-surface, default-installation, and invariant Go tests (`-count=1`).
- Passed `gofmt -l .` (no output), `make vet`, and `make lint` (zero issues).
- Passed `make test` with the existing usage-disable override absent; the full unit/integration suite ran on Go 1.25.12.
- Passed `make build`, `./bin/kit config check --json`, `./bin/kit check --project`, and `./bin/kit check 0097-shell-environment`. Config-check output reports the default as `shell: user` and contains no findings.
- The project check retains one nonblocking compatibility advisory for the customized references index.
- Passed `git diff --check`. Fresh-context reviews covered the initial guidance and the complete configuration extension. A malformed-YAML argv finding was fixed and rechecked; no actionable findings remain.
- Coverage includes fresh init, current and historical-generation reconcile, idempotence, preservation of edited rules/references and project guidance outside managed blocks, portability markers, omitted/null/empty defaults, exact override round trips, malformed argv, valid aliases, read-only config checks, argument redaction, and override preservation across reconciliation. The contract remains within its 1,500-word budget at 1,498 words.
- Not validated: live downstream rollout, execution under every shell/operating system, or causation of an execution-host release regression. Hosted checks must be assessed on the published PR head; local results do not imply hosted success.

## OUTCOME

Implemented the default contextual rule, canonical routing, project-reference guidance, validated shell configuration, config-check reporting, and regression coverage on GH-278 for pull-request review. New projects receive the rule after a release through `kit init`; existing projects require an upgraded binary and reviewed `kit reconcile`. Personal/global instructions and startup files remain unchanged. No merge, release, deployment, or downstream adoption is included.

## REPOSITORY MEMORY

Created this spec and the shared shell rule. Updated the contract, reference index, testing starter, and usage documentation. Existing Constitution ownership and host-neutrality constraints remain sufficient.
