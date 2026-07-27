---
kit_metadata_version: 1
artifact: "spec"
feature:
  id: "0017"
  slug: "reconcile-command"
  dir: "0017-reconcile-command"
relationships:
  - type: "related_to"
    target: "0007-catchup-command"
  - type: "related_to"
    target: "0011-handoff-document-sync"
  - type: "builds_on"
    target: "0013-scaffold-agents-safe-merge"
references:
  - name: constitution contract
    type: doc
    target: docs/CONSTITUTION.md
    relation: informs
    read_policy: conditional
    used_for: canonical workflow and document invariants
    status: active
  - name: init project spec
    type: doc
    target: docs/specs/0000_INIT_PROJECT.md
    relation: informs
    read_policy: conditional
    used_for: canonical document-model details
    status: active
  - name: current templates
    type: code
    target: internal/templates/templates.go
    relation: implements
    read_policy: conditional
    used_for: current section and table contract
    status: active
  - name: document parser
    type: code
    target: internal/document/document.go
    relation: implements
    read_policy: conditional
    used_for: section parsing and placeholder validation
    status: active
  - name: check command
    type: code
    target: pkg/cli/check.go
    relation: implements
    read_policy: conditional
    used_for: baseline validation behavior and gaps
    status: active
  - name: handoff prompt flow
    type: code
    target: pkg/cli/handoff.go
    relation: implements
    read_policy: conditional
    used_for: project-vs-feature prompt structure
    status: active
  - name: scaffold agents
    type: code
    target: pkg/cli/instruction_files.go
    relation: implements
    read_policy: conditional
    used_for: instruction-file drift planning
    status: active
  - name: rollup generator
    type: code
    target: internal/rollup/rollup.go
    relation: implements
    read_policy: conditional
    used_for: rollup drift expectations
    status: active
  - name: GitHub delivery rule
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: must
    used_for: safe root-checkout transfer and issue-worktree delivery
    status: active
---
# SPEC

## SUMMARY

- Add a `kit reconcile [feature]` command that audits Kit-managed project documents against the current Kit document contract and outputs a prompt for an agent to reconcile stale or missing documentation.
- The original v1 command defaults to whole-project reconciliation and emits exact file targets, update instructions, and codebase search guidance instead of editing docs directly.
- GH-52 extends project-wide `kit reconcile` into the reviewed refresh surface for Kit-managed files, rulesets, and follow-up coding-agent prompts.

## PROBLEM

- Kit's document contract evolves over time, but older projects can drift away from current expectations for sections, tables, workflow semantics, and instruction-file structure.
- Existing commands cover validation (`check`), feature catch-up (`catchup`), and handoff preparation (`handoff`), but none are designed to migrate a project's docs forward to newer Kit semantics.
- Users currently need to discover document drift manually, decide which canonical source defines the current contract, and invent search strategies for filling missing content.
- Managed-file commands can write instruction files into a protected root checkout before the issue worktree exists. Follow-up guidance that ignores those unstaged files can leave the generated contract stale on the default branch and absent from the resulting pull request.

## GOALS

- Add a root command `kit reconcile [feature]`.
- Default to repo-wide document reconciliation when no feature argument is provided.
- Support feature-scoped reconciliation when a feature argument is provided.
- Detect missing documents, missing required sections, placeholder-only required sections, malformed required tables, safe structural truncation, and bounded semantic drift.
- Audit feature docs, `docs/CONSTITUTION.md`, `PROJECT_PROGRESS_SUMMARY.md`, and repository instruction files managed by Kit.
- Reuse existing instruction-file append-only planning to detect repository-instruction drift without mutating files.
- Output a clipboard-first reconciliation prompt when findings exist.
- Output a short clean result and no prompt when no reconciliation is needed.
- Tell the agent exactly how to update each stale document and how to search the codebase for the missing evidence.
- Keep v1 strictly prompt-only and documentation-scoped.
- Keep the raw prompt concise enough to stay readable for repo-wide audits.
- In project-wide interactive mode, ask whether to include managed files, force changes, and output the coding-agent prompt.
- When managed files are included, reuse the init refresh planner to refresh `.kit.yaml`, README managed sections, rulesets, init scaffold artifacts, and instruction docs.
- Carry command-created instruction-file changes into the canonical issue worktree and resulting pull request instead of leaving them unstaged in the protected root checkout.

## NON-GOALS

- Automatically editing project documents in v1 documentation-audit mode.
- Automatically filling missing document content from code in v1.
- Producing machine-readable JSON, SARIF, or migration reports in v1.
- Changing product code as part of reconciliation instructions.
- Replacing `kit check`, `kit handoff`, `kit catchup`, or `kit scaffold agents`.

## USERS

- Maintainers bringing older Kit projects up to the current document contract.
- Contributors inheriting repos whose feature docs predate newer Kit sections, tables, or workflow semantics.
- Coding agents that need a precise, bounded prompt for documentation reconciliation without drifting into implementation.

## SKILLS

| SKILL | SOURCE | PATH | TRIGGER | REQUIRED |
| ----- | ------ | ---- | ------- | -------- |
| none | n/a | n/a | no additional skills required | no |

## RELATIONSHIPS

- related to: 0007-catchup-command
- related to: 0011-handoff-document-sync
- builds on: 0013-scaffold-agents-safe-merge

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| constitution contract | doc | `docs/CONSTITUTION.md` | canonical workflow and document invariants | active |
| init project spec | doc | `docs/specs/0000_INIT_PROJECT.md` | canonical document-model details | active |
| current templates | code | `internal/templates/templates.go` | current section and table contract | active |
| document parser | code | `internal/document/document.go` | section parsing and placeholder validation | active |
| check command | code | `pkg/cli/check.go` | baseline validation behavior and gaps | active |
| handoff prompt flow | code | `pkg/cli/handoff.go` | project-vs-feature prompt structure | active |
| scaffold agents | code | `pkg/cli/instruction_files.go` | instruction-file drift planning | active |
| rollup generator | code | `internal/rollup/rollup.go` | rollup drift expectations | active |

## REQUIREMENTS

- Expose a new root command `kit reconcile [feature]`.
- The command must accept at most one feature argument.
- When no feature argument is provided, the command must audit the whole project by default.
- The command must support `--all` as an explicit alias for whole-project mode.
- Passing both a feature argument and `--all` must fail with an actionable error.
- The command must support `--copy`, `--output-only`, and `--prompt-only`, and must use the shared clipboard-first prompt-output helper.
- In project-wide interactive terminals, the command must ask `include files?`, `force these changes?`, and `output coding-agent prompt too?` unless flags already make those choices explicit.
- The command must support `--include-files`, `--force`, `--dry-run`, `--diff`, and repeatable `--file` for managed-file refreshes.
- `--diff` must require `--dry-run`.
- Feature-scoped reconciliation and `--prompt-only` must remain prompt/audit-only and must not prompt for managed-file refresh choices.
- If reconciliation finds no issues in documentation-audit mode, the command must print a concise success result and must not emit a prompt body or copy anything to the clipboard.
- If managed files were included and the user requested the coding-agent prompt, a clean documentation audit may still emit the post-refresh documentation review prompt.
- If reconciliation finds issues, the command must emit a concise prompt that uses the host agent's native planning capability without embedding a `/plan` trigger.
- The raw prompt body must stay plain text and concise in both normal and `--output-only` modes.
- Human-readable non-`--output-only` terminal output may add compact graphical summaries such as ASCII tables or boxed sections, but must not change the raw prompt payload.
- The prompt must keep the agent focused on documentation reconciliation only and must explicitly forbid unrelated code changes.
- In the default orchestration path, the prompt must explicitly tell the coding agent to use subagents and queue work according to overlapping file changes.
- When `--single-agent` is set, that subagent-specific instruction must be omitted from the raw prompt.
- The prompt must include exact project file paths for every finding.
- The prompt must cite the canonical source order once, using:
  - current embedded template
  - `docs/CONSTITUTION.md`
  - `docs/specs/0000_INIT_PROJECT.md`
  - feature specs that introduced the newer rule when applicable
- Repo-wide prompts may omit per-finding source repetition when the grouped file summary stays unambiguous.
- The audit must detect missing required documents for the selected scope when lifecycle evidence shows those docs should exist.
- The audit must detect missing required sections and placeholder-only required sections in Kit-managed docs.
- The audit must detect malformed required tables in:
  - `## SKILLS`
  - `## DEPENDENCIES`
  - `## PROGRESS TABLE`
- The audit must detect safe structural truncation signals, including:
  - required section exists but has no meaningful body
  - required table has headers but no data rows
  - `TASK DETAILS` is missing entries referenced by `TASK LIST` or `PROGRESS TABLE`
  - a required block is provably incomplete from current parser-visible structure
- The audit must include bounded semantic-drift checks for newer Kit workflow semantics, including:
  - `RELATIONSHIPS` requirements in brainstorm/spec docs
  - dependency-table expectations in brainstorm/spec/plan docs
  - readiness-gate and related workflow wording where Kit-managed docs are missing those semantics
  - stale repository-instruction-file structure detectable via append-only planning
- The audit must include cross-document consistency checks, including:
  - `TASKS.md` ID alignment across progress table, task list, and task details
  - relationship targets that reference nonexistent feature directories
  - rollup content that is missing current features or current feature summaries
- Feature-scoped reconciliation must audit:
  - `BRAINSTORM.md` when present
  - `SPEC.md`
  - `PLAN.md` when present
  - `TASKS.md` when present
  - `PROJECT_PROGRESS_SUMMARY.md` for drift related to the selected feature
- Whole-project reconciliation must audit:
  - `docs/CONSTITUTION.md`
  - `PROJECT_PROGRESS_SUMMARY.md`
  - every feature under `docs/specs/`
  - Kit-managed repository instruction files
- Repository instruction-file drift detection must reuse the append-only planning surface instead of inventing a separate merge engine.
- The generated prompt must group findings compactly by file for repo-wide audits instead of rendering every finding as a full paragraph block.
- Feature-scoped prompts may include slightly more detail than repo-wide prompts, but must still avoid repeated path, source, and search-plan boilerplate.
- The prompt must deduplicate search guidance and include at most 1 to 3 search shortcuts per file or issue category.
- The generated prompt must include a compact fixed response contract with these sections:
  - `Findings`
  - `Updates`
  - `Verification`
- The prompt must tell the agent when to use `kit scaffold agents --append-only` instead of manual instruction-file edits.
- Reconcile and adjacent managed-instruction-file guidance must require the agent to inventory command-created unstaged and untracked files, create or reuse the human-assigned issue and exact `GH-<issue>` worktree, move the in-scope files into that worktree, and include them in the explicitly staged commit and ready pull request.
- The transfer guidance must require destination verification before removing transferred source state from the root checkout, preserve unrelated dirty files, exclude secrets and machine-local or ignored files, and forbid stash, reset, clean, bulk staging, protected-branch commits, and silent overwrite of worktree content.
- The prompt must require verification after documentation changes with:
  - `kit check --all` for whole-project mode or `kit check <feature>` for feature mode
  - `kit rollup` when reconciled changes affect `PROJECT_PROGRESS_SUMMARY.md`

## ACCEPTANCE

- Running `kit reconcile` audits the whole project and either reports clean success or outputs a reconciliation prompt.
- Running `kit reconcile --all` produces the same project-wide behavior as `kit reconcile`.
- Running `kit reconcile <feature>` audits only the selected feature plus related rollup context.
- `kit reconcile <feature> --all` fails with an actionable error.
- When findings exist, the prompt remains compatible with native agent planning and does not embed a `/plan` trigger.
- The prompt is documentation-scoped, includes exact file paths, and forbids unrelated code changes.
- Command-created instruction files are present in the issue worktree and resulting pull request, while the protected root checkout no longer retains the transferred unstaged state.
- Unrelated root-checkout and worktree changes remain untouched throughout transfer and delivery.
- The raw prompt stays compact by grouping findings by file, deduplicating search shortcuts, and avoiding repeated boilerplate.
- The default prompt explicitly tells the coding agent to use subagents and queue work according to overlapping file changes, without conflicting with `--single-agent`.
- Missing `RELATIONSHIPS`, malformed front matter references, and mismatched task IDs are surfaced as findings.
- Instruction-file drift is surfaced without mutating instruction files.
- Interactive terminal output may show a compact graphical audit summary before the clipboard acknowledgement, while `--output-only` stays plain compact text.
- A clean project prints a short success result and does not emit or copy a prompt.
- Help and README document the new command distinctly from `check`, `catchup`, `handoff`, and `scaffold-agents`.

## EDGE-CASES

- The repo has no features under `docs/specs/`.
- A feature has only `BRAINSTORM.md`.
- A feature has `SPEC.md` but no `PLAN.md` or `TASKS.md`.
- A feature has `TASKS.md` sections with inconsistent task IDs.
- A required table exists but contains only headers.
- The project lacks one or more Kit-managed repository instruction files.
- Append-only planning for an instruction file fails because the file cannot be merged safely.
- `PROJECT_PROGRESS_SUMMARY.md` exists but is missing a current feature row or summary heading.
- The selected feature name resolves by slug or numeric prefix.
- The command creates or updates tracked instruction files in the protected root checkout before an issue worktree exists.
- The root checkout contains unrelated dirty files alongside the command-created instruction files.
- The exact issue branch or worktree already exists and must be reused rather than duplicated.

## OPEN-QUESTIONS

- none

## VALIDATION

- Focused reconcile, init, refresh, scaffold-agents, health, delivery-ruleset, and capability tests passed in `pkg/cli`.
- `make fmt`, `make vet`, `go test ./... -count=1`, `golangci-lint run --new-from-rev=origin/main ./...`, and `go test -race ./pkg/cli -count=1` passed.
- `make build` passed with an isolated temporary `GIT_WT_PREFIX`.
- `kit check 0017-reconcile-command`, `kit check --project`, and `kit check --all` passed; the project refresh cadence was not due.
- Built-binary capability checks confirmed that `reconcile` and `init` document the root-checkout transfer and exact issue-worktree delivery guidance while retaining no direct Git or GitHub mutation.

## OUTCOME

- Reconcile, initial project setup, forced refresh review, ordinary refresh output, scaffold-agents completion, and health next actions now share one safe delivery handoff for command-created files.
- The handoff inventories only version-control-eligible in-scope files, creates or reuses the human-assigned issue and exact issue worktree, verifies the destination before clearing transferred root state, and includes the files in an explicitly staged commit and ready pull request.
- The guidance excludes secrets, ignored files, and machine-local configuration and preserves unrelated root-checkout and worktree changes.
- Kit commands continue to perform no hidden Git or GitHub delivery mutation; the generated prompt or human-readable next actions own the explicit follow-up workflow.

## REPOSITORY MEMORY

Decision: updated

Rationale: Safe transfer of command-created instruction files from a protected root checkout into the exact issue worktree is a durable cross-command workflow boundary that code and tests alone do not explain fully.

Artifacts:

- `docs/specs/0017-reconcile-command/SPEC.md`
- `docs/CONSTITUTION.md`
- `docs/references/rules/github-pr-delivery.md`
- `docs/commands.md`
