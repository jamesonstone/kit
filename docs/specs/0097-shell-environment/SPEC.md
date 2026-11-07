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

- Resolve the project's command environment and the current user's approved host configuration before shell-dependent commands.
- Keep shared instructions portable and credentials private. Keep personal shell choices in user-level instructions or untracked local configuration.
- Preserve script interpreters, noninteractive CI/container environments, startup-file ownership, and permission boundaries.
- Verify shell mode and tool resolution with harmless probes. Report unavailable capabilities accurately.
- Install the rule through init and supported-generation reconcile. Preserve project customizations and reconciliation idempotence.
- Do not change personal/global instructions, startup files, security settings, credentials, runtime launchers, or downstream projects.

## ACCEPTED PLAN

1. Add a default downstream `shell-environment` contextual rule and universal-contract route.
2. Use the existing testing reference's Environment Preflights section as the project-owned record or pointer to existing environment documentation.
3. Regenerate the three checked-in entry blocks and add migration, preservation, and portability regression tests.
4. Run Kit's full local validation and a fresh-context review, then publish a ready-for-review pull request through the issue/worktree lane.

## DECISIONS

- Guidance, not a shell launcher: the execution host controls tools and permissions. Kit cannot guarantee host compliance.
- Reuse the existing project-owned reference rather than add a competing environment schema or copy startup files.
- Separate shell startup, project activation, authentication, and network policy. A wrapper is not a denied-action fallback.
- Describe shell examples conditionally. Do not assume zsh for every user, CI runner, or operating system.
- The contract's 1,500-word budget remains fixed. Shorten introductory prose without changing existing invariants to add the route.

## DISCOVERIES

- `kit status` and `kit reconcile --dry-run --diff` also propose pre-existing registry/badge refresh and Mint release-workflow upgrades. Those unrelated changes are left out of this shell-guidance PR.
- The rule registry discovers embedded Markdown automatically; no new CLI or registry schema is necessary.

## VALIDATION

- Passed focused shell-rule, universal-contract, managed-surface, default-installation, and invariant Go tests (`-count=1`).
- Passed `gofmt -l .` (no output), `make vet`, and `make lint` (zero issues).
- Passed `make test` with the existing usage-disable override absent; the full unit/integration suite ran on Go 1.25.12.
- Passed `make build`, `./bin/kit check --project`, and `./bin/kit check 0097-shell-environment`.
- The project check retains one nonblocking compatibility advisory for the customized references index.
- Passed `git diff --check`. A separate fresh-context reviewer compared all 14 intended files to issue #278 and reported no actionable findings.
- Coverage includes fresh init, current and historical-generation reconcile, idempotence, preservation of edited rules/references and project guidance outside managed blocks, and portability markers. The contract remains within its 1,500-word budget at 1,498 words.
- Not validated: live downstream rollout, execution under every shell/operating system, or causation of an execution-host release regression. Hosted checks must be assessed on the published PR head; local results do not imply hosted success.

## OUTCOME

Implemented the default contextual rule, canonical routing, project-reference guidance, and regression coverage on GH-278 for pull-request review. New projects receive the rule after a release through `kit init`; existing projects require an upgraded binary and reviewed `kit reconcile`. Personal/global instructions and startup files remain unchanged. No merge, release, deployment, or downstream adoption is included.

## REPOSITORY MEMORY

Created this spec and the shared shell rule. Updated the contract, reference index, testing starter, and usage documentation. Existing Constitution ownership and host-neutrality constraints remain sufficient.
