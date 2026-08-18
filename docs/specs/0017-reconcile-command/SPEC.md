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
- The current reconcile refresh can still create an exact managed-file snapshot in the primary checkout before the generated coding-agent instructions establish a writable lane. The later work-lane tripwire correctly preserves that state, but it also makes the command's own expected refresh look like an ownership violation and prevents the normal reconcile workflow from completing cleanly.
- Append-only instruction refresh considers an existing top-level section preserved even when newer mandatory semantics were added inside that section or its preamble. Older V3 projects can therefore retain obsolete worktree behavior or omit the testing-validation route while a project-wide audit reports clean.
- A managed-file dry-run can print a real structural diff and then end with the unqualified documentation-audit message `No reconciliation needed`, obscuring the distinction between pending file refreshes and semantic document findings.
- Registry refresh currently advances every ruleset artifact's `source_commit` to the registry repository head even when that artifact's normalized content and `installed_hash` are unchanged. Scheduled maintenance therefore produces noisy `.kit.yaml` churn and can disconnect a retained base hash from the commit that actually supplied it.

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
- Preserve the existing reconcile findings, prompt structure, clipboard behavior, raw output contract, refresh semantics, and flags while ensuring a write-capable managed refresh never applies in the primary checkout.
- When a primary-checkout invocation requests a managed refresh, use the existing no-snapshot delivery path so the coding agent establishes the canonical lane and reruns the same write-capable reconciliation there; do not add a replacement reconcile output format or a second synchronization command.
- Detect bounded semantic drift for current testing routes, testing-reference structure, worktree navigation, pull-request annotations, and writable-lane environment-link safety even when append-only section planning finds no missing headings.
- Distinguish a clean semantic documentation audit from pending managed-file changes in dry-run output.
- Keep per-artifact registry provenance stable when a refresh does not change the installed normalized content.

## NON-GOALS

- Automatically editing project documents in v1 documentation-audit mode.
- Automatically filling missing document content from code in v1.
- Producing machine-readable JSON, SARIF, or migration reports in v1.
- Changing product code as part of reconciliation instructions.
- Replacing `kit check`, `kit handoff`, `kit catchup`, or `kit scaffold agents`.
- Automatically overwriting customized instruction or support-document sections merely because their wording differs from the current generated template.
- Distributing Kit-maintainer release, Kit Improve, or other repository-specific GitHub Actions workflows to downstream projects.

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
  - mandatory testing-validation routes in provider instructions, RLM guidance, the references index, and the project testing reference
  - current V3 worktree behavior for default list navigation, fail-soft pull-request annotations, direct branch navigation, `.env` and `.envrc` ownership, collision handling, and safe removal
- Semantic findings inside existing sections must recommend reviewed manual integration or a targeted forced refresh only when generated overwrite is acceptable; they must not claim append-only refresh can replace existing section content.
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
- Reconcile and adjacent managed-instruction-file guidance must carry the command's exact version-control-eligible path, action, pre-command state, and expected result state rather than infer ownership from post-command Git status.
- When an included managed-file dry-run plans version-control-eligible changes but the semantic documentation audit is otherwise clean, the final result must report both facts and must not print the unqualified no-reconciliation-needed result.
- A tracked ruleset artifact must retain its existing `source_commit` when its source repository, branch, path, and installed normalized hash are unchanged.
- A ruleset artifact must advance `source_commit` when a refresh installs changed normalized content from the current registry head.
- A conflict or local-custom result that retains the prior `installed_hash` must also retain the prior `source_commit`; conflict recovery must continue to point at the source checkpoint used as its merge base.
- The transfer guidance must create or reuse the human-assigned issue and exact `GH-<issue>` worktree, abort when a captured destination path does not match its pre-command state or already has staged, working-tree, or untracked changes, and move only the captured command-owned delta into that worktree.
- Created, updated, merged, and removed paths must be verified and explicitly staged; removed paths must remain absent in the destination and be represented as deletions in the index.
- Before clearing source state, the guidance must verify the destination result state and require the index to contain exactly the captured command-owned paths, including deletions.
- The transfer guidance must restore only the captured command-owned source delta to its exact pre-command state, preserve unrelated dirty files, exclude secrets and machine-local or ignored files, and forbid stash, reset, clean, bulk staging, protected-branch commits, and silent overwrite of worktree content.
- Before applying a non-dry-run included refresh, reconcile must inspect the current Git worktree and distinguish the primary checkout from a linked writable worktree.
- A requested non-dry-run refresh in the primary checkout must not call the managed-file apply path. It must retain the refresh intent, force the existing coding-agent-prompt path when necessary, and render the established no-snapshot delivery instructions that require the write-capable command to be rerun in the selected canonical worktree.
- The deferred delivery instructions must render one exact shell-safe rerun command that preserves whole-project or feature scope, force, file filters, reference and verification migrations, prompt profile, single-agent selection, and explicit clipboard copying while using `--output-only` so the coding agent can consume the post-refresh instructions.
- Reconcile must not add new prompt sections, response fields, status formats, or public flags for primary-checkout deferral. Existing prompt-only, output-only, clipboard, summary, and response contracts remain authoritative.
- Included dry-runs remain read-only and available in the primary checkout. Included refreshes in non-primary linked worktrees and non-Git project roots retain their existing direct-apply behavior.
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
- A write-capable reconcile requested from the primary checkout leaves its files and index unchanged, emits the existing no-snapshot delivery workflow, and results in the managed refresh being rerun and captured in the canonical writable worktree.
- The equivalent write-capable reconcile run in a linked non-primary worktree retains the existing managed refresh and rendered output behavior.
- Primary-checkout dry-runs and non-Git refresh fixtures retain their existing behavior.
- Unrelated root-checkout and worktree changes remain untouched throughout transfer and delivery.
- The raw prompt stays compact by grouping findings by file, deduplicating search shortcuts, and avoiding repeated boilerplate.
- The default prompt explicitly tells the coding agent to use subagents and queue work according to overlapping file changes, without conflicting with `--single-agent`.
- Missing `RELATIONSHIPS`, malformed front matter references, and mismatched task IDs are surfaced as findings.
- Instruction-file drift is surfaced without mutating instruction files.
- Older V2 and V3 testing guidance that omits the mandatory testing-validation route is surfaced without overwriting local content.
- Older V3 worktree guidance that omits current navigation, pull-request annotation, or environment-link safety semantics is surfaced without overwriting local content.
- A managed-file dry-run with pending version-control-eligible changes reports that the structural refresh is pending while the semantic documentation audit is clean.
- Refreshing from a newer registry repository head with unchanged ruleset content produces no `.kit.yaml` change for that artifact.
- A mixed refresh advances provenance only for rulesets whose installed normalized content changed.
- Conflict and local-custom refresh outcomes keep provenance aligned with the retained registry checkpoint.
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
- A primary-checkout refresh is requested while the user disables coding-agent prompt output; reconcile must still emit the existing delivery prompt because deferral is the only safe path that preserves mandatory synchronization.
- The root checkout contains unrelated dirty files alongside the command-created instruction files.
- The exact issue branch or worktree already exists and must be reused rather than duplicated.
- Existing support documents contain every expected top-level heading but predate mandatory testing or worktree semantics added within those headings.
- A dry-run has pending managed-file changes but no semantic documentation findings.
- The registry repository head advances because an unrelated ruleset or non-registry file changed.
- A ruleset conflicts locally and remotely while its recorded base content remains unchanged.

## OPEN-QUESTIONS

- none

## PLAN

- Add read-only worktree inspection that identifies the exact primary checkout without requiring a GitHub remote and preserves non-Git project behavior.
- Route primary-checkout non-dry-run refresh requests through the existing no-snapshot prompt path without changing rendered reconcile contracts; keep direct apply for linked worktrees.
- Add focused primary, linked-worktree, dry-run, and non-Git regressions, then run complete Go and Kit validation before delivery.

## VALIDATION

- Focused reconcile, initial init, refresh, scaffold-agents, health, delivery-ruleset, exact-snapshot, alias-normalization, containment, nested-worktree ignore, and removal-only tests passed in `pkg/cli`.
- `make fmt`, `make vet`, `go test ./... -count=1`, `golangci-lint run --new-from-rev=origin/main ./...`, and `go test -race ./pkg/cli -count=1` passed.
- `make build` passed with an isolated temporary `GIT_WT_PREFIX`.
- `kit check 0017-reconcile-command`, `kit check --project`, and `kit check --all` passed; the project refresh cadence was not due.
- Health JSON and generated guidance confirmed that planned paths carry exact action, pre-command SHA-256 or absence, and expected result SHA-256 or absence while excluding `.env`, `.envrc`, Git-ignored paths, and obvious secret material.
- Built-binary capability checks confirmed that `reconcile` and `init` document the root-checkout transfer and exact issue-worktree delivery guidance while retaining no direct Git or GitHub mutation.
- An independent read-only verifier found and rechecked initial-init propagation, path-alias precedence, project-root containment, and nested-worktree ignore handling; the final verification pass reported no findings.
- GH-108 focused tests confirmed that every current V2 and V3 guidance template satisfies the bounded semantic expectations and that stale testing routes, V3 worktree behavior, and unsafe append-only recommendations are detected.
- A live `go run ./cmd/kit reconcile --all --include-files --dry-run --diff --output-only` preview surfaced the repository's pending testing-ruleset registry update and ended with `Managed-file refresh pending for 1 file` while leaving the worktree unchanged.
- GH-108 passed `make fmt`, `make vet`, `go test ./... -count=1`, `golangci-lint run --new-from-rev=origin/main ./...`, `go test -race ./pkg/cli -count=1`, and `make build`.
- The built GH-108 binary passed `kit check 0017-reconcile-command`, `kit check --project`, and `kit check --all`; its `kit capabilities reconcile --json` output documents existing-section semantic drift and distinct pending-refresh status.
- GH-112 focused regressions confirmed that an unchanged ruleset retains its prior `source_commit` without rewriting `.kit.yaml`, changed content or source identity advances provenance, a mixed refresh leaves unrelated provenance stable, and a conflict retains its prior source checkpoint.
- GH-112 passed `make fmt`, `make vet`, `go test ./... -count=1`, `golangci-lint run --new-from-rev=origin/main ./...`, `go test -race ./pkg/cli -count=1`, and an isolated `make build`.
- The built GH-112 binary passed `kit check 0017-reconcile-command`, `kit check --project`, and `kit check --all`.
- GH-160 focused Git fixtures proved that a requested non-dry-run refresh leaves the primary checkout clean, emits the established no-snapshot delivery workflow, and continues to apply managed files plus exact snapshot guidance in a linked worktree. Non-Git and primary dry-run decisions retained direct apply behavior.
- GH-160 passed `make fmt`, `make vet`, `go test ./... -count=1`, `go test -race ./internal/worktreeprep ./pkg/cli -count=1`, `golangci-lint run --new-from-rev=origin/main ./...`, and `make build`.
- The built GH-160 binary passed `kit check 0017-reconcile-command`, `kit check --project`, and `kit check --all`. A live primary-checkout `reconcile --all --include-files --output-only` emitted the existing post-refresh/no-snapshot prompt and left root `main` clean.
- The final GH-160 managed dry-run planned zero Kit-managed changes and reported a complete source-file-size audit of 699 version-control-eligible candidates, 361 eligible handwritten source/test files, and 0 violations.
- A fresh read-only verifier identified lost deferred invocation flags and fail-open nested Git inspection in the initial GH-160 implementation. After repair, the same verifier confirmed exact shell-safe rerun preservation, fail-closed ancestor metadata handling, focused test success, diff cleanliness, and source-size compliance with no remaining actionable findings.
- CodeRabbit identified that an explicit `--copy` choice was omitted from the deferred rerun. GH-160 preserved the flag, added command-builder coverage, aligned the tracking issue with deferred delivery, and reran the complete Go, race, vet, lint, and build validation successfully.

## OUTCOME

- Reconcile, initial project setup, forced refresh review, ordinary refresh output, scaffold-agents completion, and health next actions now share one safe delivery handoff for command-created files.
- The handoff carries an exact snapshot of each version-control-eligible in-scope path, action, pre-command state, and expected result state; creates or reuses the human-assigned issue and exact issue worktree; aborts on destination conflicts; verifies created, updated, merged, and removed results; and requires the index to contain exactly the captured paths before clearing transferred root state.
- Initial init captures whole-command baselines before any write, refresh planners propagate exact snapshots through health next actions, init-refresh documentation prompts, and reconcile prompts, and scaffold version cleanup contributes explicit removal entries with removal-only regression coverage.
- Snapshot paths are normalized and confined to the project before reads or rendering; Git ignore detection works from nested project roots and fails closed on in-worktree Git errors.
- The guidance excludes secrets, ignored files, and machine-local configuration and preserves unrelated root-checkout and worktree changes.
- Kit commands continue to perform no hidden Git or GitHub delivery mutation; the generated prompt or human-readable next actions own the explicit follow-up workflow.
- GH-108 adds bounded V2 and V3 semantic expectations for current testing-validation routes and V3 worktree navigation, pull-request annotation, environment-link collision handling, and safe removal behavior.
- Existing customized sections remain preserved: semantic findings now direct reviewed manual integration or a targeted forced dry-run instead of incorrectly claiming append-only refresh can update existing content.
- Managed-file dry-runs use the exact version-control-eligible delivery snapshot to report pending structural refreshes separately from a clean semantic documentation audit.
- Kit-specific improvement and release workflows remain outside downstream reconcile scope; the existing auto-assign workflow remains the only Kit-managed downstream GitHub workflow.
- Ruleset refresh now retains an artifact's prior `source_commit` when its installed normalized hash and source repository, branch, and path remain unchanged, eliminating repository-head-only `.kit.yaml` churn.
- Changed rulesets advance to the current registry commit, while conflict and local-custom outcomes that retain the prior installed hash also retain the prior registry checkpoint for future comparison and merge recovery.
- Reconcile now inspects worktree ownership before a requested non-dry-run managed refresh. Primary-checkout requests retain refresh intent but skip the apply path and emit the existing no-snapshot delivery workflow, while linked worktrees continue through the unchanged refresh, snapshot, audit, and prompt renderers.
- Deferred refresh guidance carries the exact shell-safe write-capable invocation, including force, file filters, feature scope, migration flags, prompt profile, single-agent selection, and explicit clipboard copying, and forces raw prompt output for the coding-agent rerun.
- The primary-checkout deferral adds no public flag, prompt section, response field, Git mutation, or GitHub mutation. Dry-runs and non-Git projects preserve their prior behavior, and the generated workflow remains responsible for establishing the canonical issue lane before the write-capable rerun.

## REPOSITORY MEMORY

Decision: updated

Rationale: The existing reconcile feature spec now preserves both per-artifact refresh provenance and the worktree-aware apply boundary. The behavior is feature-specific and fully demonstrated by code and regression tests; the Constitution already requires primary checkouts to remain read-only, so no new project-wide rule or reusable reference is warranted.

Artifacts:

- `docs/specs/0017-reconcile-command/SPEC.md`
