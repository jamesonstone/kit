---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0054"
  slug: "source-file-line-enforcement"
  dir: "0054-source-file-line-enforcement"
relationships:
  - type: builds_on
    target: 0017-reconcile-command
  - type: builds_on
    target: 0047-kit-health-maintenance
references:
  - id: reconcile-audit
    name: Reconcile project audit
    type: code
    target: pkg/cli/reconcile_audit.go
    relation: implements
    read_policy: must
    used_for: whole-project source-file findings
    status: active
  - id: reconcile-prompt
    name: Reconcile agent prompt
    type: code
    target: pkg/cli/reconcile_prompt.go
    relation: implements
    read_policy: must
    used_for: behavior-preserving remediation instructions
    status: active
  - id: guardrails-template
    name: Managed guardrails template
    type: code
    target: internal/templates/instruction_support_templates.go
    relation: implements
    read_policy: must
    used_for: new and reconciled project instructions
    status: active
  - id: source-file-size-rule
    name: Source file size rule
    type: rule
    target: docs/references/rules/source-file-size.md
    relation: constrains
    read_policy: must
    used_for: line limit, scope, exclusions, splits, and verification
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: constrains
    read_policy: must
    used_for: behavior-preserving validation after source splits
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: create and verify issue 116 and its pull request
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the validated issue branch as a ready pull request
    required: true
delivery_intent: issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make the existing 300-line source-file policy consistently visible,
machine-audited, and periodically repairable in new and existing Kit-managed
projects instead of relying on an agent to notice subjective prose.

## CONTEXT

- Current guardrails say to keep implementation/source files “around 300
  lines” only “when splitting improves clarity.” The threshold, scope, required
  audit, and unresolved-exception behavior are not binary.
- V3 root instructions route to `docs/agents/GUARDRAILS.md` but do not expose a
  source-file-size gate before implementation or delivery. Copilot has the
  same omission.
- `kit reconcile --all` audits Kit-managed documents, scaffold artifacts,
  rulesets, specs, progress summaries, and instruction drift. It never
  enumerates source or test files, so it can report a clean semantic audit
  while oversized code remains.
- The weekly Kit Health automation records structural-refresh and semantic-doc
  evidence only. Its current product-code boundary prevents the automation
  from applying a behavior-preserving split even when code hygiene is the
  actionable finding.
- A repository-wide audit on issue branch `GH-116` found six tracked Go files
  over 300 physical lines, despite the completed GH-91 cleanup. This is direct
  evidence that a one-time refactor does not preserve the invariant.
- `kit health` remains limited to Kit-managed files and project validation.
  Source remediation belongs to the coding-agent workflow produced by
  `kit reconcile`, not to hidden CLI writes.
- GitHub issue #116 and exact branch `GH-116` own this change.

## REQUIREMENTS

- Define one mandatory downstream `source-file-size` ruleset with an exact
  maximum of 300 physical lines for version-control-eligible handwritten
  implementation/source and test files.
- Exclude documentation, `docs/**`, `.kit/**`, `.kit.yaml`, ignored files,
  vendored dependency trees, generated files with recognized markers, and
  non-source artifacts.
- Require semantic responsibility splits that preserve behavior and stable
  public entry points. Arbitrary numbered chunks, minification, compacting,
  or removing useful whitespace do not establish compliance.
- Require normal implementation work to audit its complete affected
  source/test scope before delivery. Whole-project reconcile and scheduled
  maintenance must audit the complete repository scope.
- Add a whole-project `kit reconcile` audit that deterministically enumerates
  cached and untracked non-ignored Git files, with a filesystem fallback for a
  project not yet initialized as a Git repository.
- Recognize common implementation, test, script, style, template, schema, and
  query source extensions plus extensionless executable shebang scripts.
- Report one exact warning per oversized file with its observed line count,
  contract source, safe remediation, and verification shortcuts.
- When a Git repository exists but version-control-eligible file enumeration
  fails, report the audit as unavailable rather than claiming a clean result.
- Keep feature-scoped reconcile focused on the feature documents; the complete
  source-file audit runs only for whole-project reconcile.
- Make generated V3 repository instructions and Copilot guidance expose a
  mandatory source-file-size gate and route to the detailed ruleset.
- Make reconcile guidance expectations detect customized existing instruction
  trees that still carry the older subjective rule.
- Allow reconcile prompts containing source findings to authorize only
  behavior-preserving responsibility splits and directly required tests/docs;
  retain the documentation-only boundary for document-only findings.
- Update the weekly Kit Health automation in place. Add source-file-size as a
  third independent evidence dimension, require it for no-op classification,
  and permit only reconcile-identified behavior-preserving source/test splits.
- Keep source-changing maintenance ineligible for documentation/Kit-only CI
  skip behavior.
- Bring Kit's own version-control-eligible handwritten source/test files to
  300 lines or fewer as part of the validated outcome.

### Non-goals

- Automatically splitting or rewriting product code inside `kit reconcile` or
  `kit health`.
- Treating documentation, generated output, vendored dependencies, or Kit
  local state as source-file-size violations.
- Guessing that an oversized file is acceptable and silently suppressing it.
- Changing product behavior while repairing line-limit findings.
- Adding a universal formatter, language parser, or per-language AST splitter.

## ACCEPTED PLAN

1. Add the canonical downstream rule and strengthen Constitution, guardrail,
   V3 root-instruction, Copilot, and reference-index templates plus checked-in
   artifacts.
2. Implement a focused source-file enumerator and physical-line audit in a new
   reconcile source file, including exact scope/exclusion helpers and unit
   tests for Git, non-Git, generated, vendored, documentation, extensionless,
   unterminated-final-line, and enumeration-failure cases.
3. Append whole-project findings to the existing reconcile report and adapt
   prompt/category/summary behavior so code findings authorize only safe
   semantic splits while document-only runs remain documentation-only.
4. Add generator-alignment, stale-guidance, ruleset-validity, prompt, and
   end-to-end reconcile tests.
5. Split the six currently oversized Kit Go files by responsibility without
   changing behavior or public entry points, then verify no in-scope file
   exceeds 300 physical lines.
6. Update `weekly-kit-health` in place with the third evidence dimension,
   bounded remediation authority, explicit no-op gate, validation, CI-skip,
   issue, and final-report requirements.
7. Run focused/full tests, formatting, vet, race checks, build, changed-lines
   lint, Kit feature/project checks, built-binary reconcile smoke tests, full
   diff review, and secret scanning; curate repository memory and deliver a
   ready pull request for issue #116.

## DECISIONS

- Use `kit reconcile`, not `kit check --project`, as the enforcement surface.
  Reconcile is the existing reviewed maintenance workflow; turning the line
  limit into an immediate project-check failure would break downstream CI
  before repositories have a repair lane.
- Keep the CLI read-only with respect to source files. It identifies exact
  violations and emits a bounded coding-agent prompt; the normal issue,
  worktree, review, and validation gates own edits.
- Audit Git cached plus untracked non-ignored files so both existing source and
  newly created implementation files are covered. Use a filesystem fallback
  only before Git initialization.
- Treat a generated marker or vendored path as an exclusion, but do not infer
  generated status merely from file size or an ambiguous directory name.
- Make 300 physical lines an exact threshold. If a safe split is genuinely
  unavailable, the finding remains visible and is reported as blocked rather
  than silently accepted.
- Keep scheduled source edits narrower than general product work: only files
  listed by the current reconcile audit may be split, behavior may not change,
  and all repository-native validation remains required.

## DISCOVERIES

- The first `kit spec source-file-line-enforcement --output-only` invocation
  allocated feature 0054 but unexpectedly entered the deprecated V2 editor
  path. No thesis was submitted and no implementation began, so the empty
  generated placeholder was semantically replaced with this accepted V3 spec.
- `buildReconcilePrompt` currently emits a documentation-only rule and a
  command-owned-file transfer section for every finding. Both assumptions
  must become conditional before source-file findings can be safely repaired.
- The completed GH-91 refactor left no oversized Go files at that time, but
  subsequent development produced six current violations. Scheduled auditing
  is therefore part of the acceptance outcome, not optional follow-up.
- A finding-only interface was insufficient for automation because an older
  Kit binary and a clean current audit produced indistinguishable output. The
  whole-project report now emits literal `source-file-size audit: complete`
  evidence with candidate, eligible-file, and violation counts; incomplete
  enumeration prohibits a clean result.
- Generated markers must appear as recognizable header comments. Treating an
  arbitrary marker string anywhere in the first bytes as generated would let
  handwritten tests containing fixture text escape the audit.
- The six existing Kit violations separated cleanly into responsibility-named
  production or test files. The resulting whole-project audit checks 502
  eligible handwritten source/test files and reports zero violations; the
  largest eligible Go files contain exactly 300 physical lines.
- The existing `weekly-kit-health` automation was updated in place, retaining
  its Wednesday 13:00 schedule, active status, model, reasoning effort, and
  deterministic repository allowlist. It now requires both capability-level
  support and literal complete audit evidence before classifying this
  dimension as clean.

## VALIDATION

- PASS: `make fmt` and `git diff --check`.
- PASS: focused source-audit, generated-marker, prompt-boundary, ruleset,
  template-alignment, instruction-refresh, capability, and reconcile tests.
- PASS: `go test ./... -count=1`.
- PASS: `go vet ./...`.
- PASS: `go test -race ./internal/templates ./internal/worktree ./pkg/cli
  -count=1`.
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero issues.
- PASS: `go build ./cmd/...` and a dedicated `/tmp/kit-gh116` binary build.
- PASS: `/tmp/kit-gh116 rules view source-file-size`.
- PASS: `/tmp/kit-gh116 check 0054-source-file-line-enforcement` and
  `/tmp/kit-gh116 check --all` across 51 features.
- PASS: `/tmp/kit-gh116 reconcile --all --output-only` emitted literal complete
  evidence for 799 version-control-eligible candidates, 502 eligible
  handwritten source/test files, and zero files above 300 physical lines.
- PASS: `/tmp/kit-gh116 reconcile --all --include-files --dry-run --diff`
  remained mutation-free, planned no managed refresh, and retained the same
  literal complete source-audit evidence required by scheduled maintenance.
- PASS: direct whole-tree Go line-count review found no eligible Go file above
  300 lines.
- PASS: the persisted `weekly-kit-health` automation contains the source audit
  capability gate, literal evidence gate, bounded split authority, no-op
  requirement, validation, CI-skip exclusion, and final-report fields while
  preserving its prior schedule and runtime settings.
- PASS: after feature completion and the scoped progress-rollup update,
  `/tmp/kit-gh116 check --project` reported a coherent project and
  `/tmp/kit-gh116 reconcile --all --output-only` reported no reconciliation
  findings plus literal complete source-audit evidence.
- PASS: post-validation Constitution review was recorded with
  `/tmp/kit-gh116 project refresh --now`; the follow-up status reports the
  review is not due for another five completed features or 30 days.

## OUTCOME

Kit now distributes an exact downstream 300-physical-line rule through new and
refreshed project instructions, audits the entire version-control-eligible
source/test scope during whole-project reconcile, emits machine-checkable
complete or incomplete evidence, and authorizes only bounded semantic repair
guidance. Kit's own six violations were split without behavior changes, and
the existing weekly maintenance task now treats source size as a fail-closed
third evidence dimension.

## REPOSITORY MEMORY

Decision: created

Rationale: The exact scope, exclusions, reconcile-versus-check boundary,
scheduled remediation authority, and behavior-preserving split contract are
durable cross-project decisions that code and tests alone cannot explain.

Artifacts:

- `docs/specs/0054-source-file-line-enforcement/SPEC.md`
- `docs/references/rules/source-file-size.md`
- `docs/CONSTITUTION.md`
- `AGENTS.md`
- `CLAUDE.md`
- `.github/copilot-instructions.md`
