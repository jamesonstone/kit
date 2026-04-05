# SPEC

## SUMMARY

- Add a new `kit reconcile [feature]` command that audits Kit-managed project documents against the current Kit document contract and outputs a prompt for an agent to reconcile stale or missing documentation.
- The command must default to whole-project reconciliation, stay prompt-only in v1, and emit exact file targets, update instructions, and codebase search guidance instead of editing docs directly.

## PROBLEM

- Kit's document contract evolves over time, but older projects can drift away from current expectations for sections, tables, workflow semantics, and instruction-file structure.
- Existing commands cover validation (`check`), feature catch-up (`catchup`), and handoff preparation (`handoff`), but none are designed to migrate a project's docs forward to newer Kit semantics.
- Users currently need to discover document drift manually, decide which canonical source defines the current contract, and invent search strategies for filling missing content.

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

## NON-GOALS

- Automatically editing project documents in v1.
- Automatically filling missing document content from code in v1.
- Producing machine-readable JSON, SARIF, or migration reports in v1.
- Changing product code as part of reconciliation instructions.
- Replacing `kit check`, `kit handoff`, `kit catchup`, or `kit scaffold-agents`.

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
- If reconciliation finds no issues, the command must print a concise success result and must not emit a prompt body or copy anything to the clipboard.
- If reconciliation finds issues, the command must emit a prompt that begins with `/plan`.
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
- The prompt must tell the agent when to use `kit scaffold-agents --append-only` instead of manual instruction-file edits.
- The prompt must require verification after documentation changes with:
  - `kit check --all` for whole-project mode or `kit check <feature>` for feature mode
  - `kit rollup` when reconciled changes affect `PROJECT_PROGRESS_SUMMARY.md`

## ACCEPTANCE

- Running `kit reconcile` audits the whole project and either reports clean success or outputs a reconciliation prompt.
- Running `kit reconcile --all` produces the same project-wide behavior as `kit reconcile`.
- Running `kit reconcile <feature>` audits only the selected feature plus related rollup context.
- `kit reconcile <feature> --all` fails with an actionable error.
- When findings exist, the prompt starts with `/plan`.
- The prompt is documentation-scoped, includes exact file paths, and forbids unrelated code changes.
- The raw prompt stays compact by grouping findings by file, deduplicating search shortcuts, and avoiding repeated boilerplate.
- The default prompt explicitly tells the coding agent to use subagents and queue work according to overlapping file changes, without conflicting with `--single-agent`.
- Missing `RELATIONSHIPS`, malformed dependency tables, and mismatched task IDs are surfaced as findings.
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

## OPEN-QUESTIONS

- none
