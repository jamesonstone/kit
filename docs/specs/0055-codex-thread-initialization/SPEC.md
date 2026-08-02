---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0055"
  slug: "codex-thread-initialization"
  dir: "0055-codex-thread-initialization"
relationships:
  - type: builds_on
    target: 0017-reconcile-command
  - type: builds_on
    target: 0044-versioned-agent-instructions
references:
  - id: agents-template
    name: Managed AGENTS template
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: generated Codex thread initialization gate
    status: active
  - id: reconcile-guidance
    name: Reconcile guidance expectations
    type: code
    target: pkg/cli/reconcile_guidance_expectations.go
    relation: implements
    read_policy: must
    used_for: stale existing-project instruction detection
    status: active
  - id: thread-initialization-rule
    name: Codex thread initialization rule
    type: rule
    target: docs/references/rules/codex-thread-initialization.md
    relation: constrains
    read_policy: must
    used_for: ordered first-response behavior and fail-visible fallback
    status: active
skills:
  - name: github:github
    source: GitHub plugin
    path: github:github
    trigger: create and verify issue 118 on the existing pull-request lane
    required: true
  - name: github:yeet
    source: GitHub plugin
    path: github:yeet
    trigger: publish the validated issue-scoped commit to pull request 117
    required: true
delivery_intent: existing_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make Codex thread renaming and pinning mandatory, ordered, pre-response
initialization actions for newly created tasks, while making unavailable or
failed host operations immediately visible instead of silently weakening the
contract.

## CONTEXT

- The current global instruction says to “attempt” renaming and pinning, then
  says never to stop or delay work when session management is unavailable.
  That combination frames both actions as optional best effort and provides no
  ordering, verification, or observable failure requirement.
- Repository `AGENTS.md` is the provider-specific instruction entry point read
  by Codex, but its V3 template currently begins with general routing. It does
  not reinforce thread initialization or let `kit reconcile` detect weakened
  startup wording in an existing project.
- `AGENTS.md` and `CLAUDE.md` currently share one body generator. The new gate
  is Codex-specific and must not leak into Claude or Copilot instructions.
- Repository instructions cannot manufacture host capabilities or override a
  higher-priority tool boundary. The strongest enforceable contract is a
  required ordered invocation, returned-state verification where available,
  and fail-visible first commentary when an operation cannot succeed.
- GitHub issue #118 owns this additional scope on existing branch `GH-116` and
  ready pull request #117, as explicitly requested by the user.

## REQUIREMENTS

- Put a self-contained `Codex Thread Initialization Hard Gate` first in
  generated and checked-in V3 `AGENTS.md`, immediately after the document
  title and before routing or repository inspection guidance.
- Apply the gate to newly created Codex tasks. Continued tasks retain their
  current title and pin state unless either state is missing or the user asks
  to change it.
- Before first commentary, planning, repository inspection, shell/network
  work, or other task actions, allow only the minimal capability lookup needed
  to find session-title and session-pin operations.
- Require the session-title operation first and the session-pin operation
  second. Derive `[<project>] <description>` from supplied cwd/repository and
  user-request context without inspecting the repository first; keep the
  description lowercase and at most four words.
- Verify success from returned operation state when available. Never defer a
  supported operation to a later interaction.
- If either operation is unsupported, unavailable, or fails, continue the
  requested work but make its exact status visible in the first commentary;
  do not silently skip it or retry indefinitely.
- Keep the gate Codex-only. Do not add it to generated `CLAUDE.md` or GitHub
  Copilot instructions.
- Register a downstream must-read ruleset that preserves the complete ordered
  contract for new and refreshed Kit-managed projects.
- Teach V3 reconciliation expectations to detect a missing or weakened gate in
  customized existing `AGENTS.md` files, while normal managed refresh remains
  responsible for replacing exact stale generated content.
- Add focused tests for ordering, provider isolation, generated-artifact
  alignment, ruleset validity, init/refresh adoption, and stale-guidance
  findings.

### Non-goals

- Claiming that repository prose can guarantee operations absent from the
  Codex host.
- Blocking the user's substantive task when a host operation is unavailable.
- Renaming or pinning continued tasks on every interaction.
- Adding Codex session behavior to Claude or Copilot guidance.
- Creating a second branch or pull request for issue #118.

## ACCEPTED PLAN

1. Add a provider-specific first-section gate to the V3 AGENTS generator and
   align the checked-in AGENTS artifact without changing Claude or Copilot.
2. Add the canonical downstream rule and reference-index routes, then register
   it for new-project initialization and managed refresh.
3. Add AGENTS-specific V3 reconcile expectations that pin the gate, ordered
   operations, and fail-visible first-commentary fallback.
4. Add focused generator, reconciliation, ruleset, and adoption regression
   tests, then run complete repository validation and a built-binary reconcile
   smoke test.
5. Curate the completed rationale, commit the additional scope against issue
   #118 on `GH-116`, update pull request #117 to close both issues, push, and
   verify exact local, remote, pull-request, and hosted-check state.

## DECISIONS

- Use a hard gate rather than stronger adjectives around “attempt.” Reliability
  comes from explicit ordering, a pre-response boundary, verification, and an
  observable fallback, not from repeating “always” or “must.”
- Permit a minimal capability lookup before rename/pin because some hosts lazy
  load thread operations. Repository inspection is unnecessary: the supplied
  cwd and request provide enough information for the short title.
- Continue substantive work after a fail-visible status because an unavailable
  host feature must not deadlock unrelated implementation. Silent continuation
  is prohibited because it made the prior rule unaccountable.
- Generate the gate only for `AGENTS.md`. The shared instruction body remains
  provider-neutral for Claude, and Copilot does not own Codex task state.
- Reconcile exact semantic anchors in addition to template management so a
  customized existing AGENTS document cannot retain only the weak “attempt”
  behavior while appearing structurally current.

## DISCOVERIES

- The first scaffold command allocated feature 0055 but entered the deprecated
  V2 editor flow despite using the current built binary. No thesis was
  submitted and no implementation began, so the empty placeholder was
  semantically replaced with this accepted V3 spec.
- The V3 AGENTS and CLAUDE artifacts shared one instruction-body function.
  A title-gated prefix keeps the new startup behavior Codex-only without
  duplicating the provider-neutral repository-memory body.
- Structural managed-file refresh alone was insufficient for customized
  existing AGENTS documents. Adding exact V3 semantic expectations makes
  reconcile report missing order, pre-response, and fail-visible anchors even
  when an existing section prevents append-only repair.
- The scheduled task needed a capability-level version gate as well as prompt
  wording. It now refuses repository mutations unless `kit capabilities
  reconcile --json` advertises this audit, preventing an older binary from
  treating absence of the new finding as clean evidence.

## VALIDATION

- PASS: `make fmt` and `git diff --check`.
- PASS: focused template, provider-isolation, reconcile-guidance, ruleset,
  downstream-adoption, and capability tests.
- PASS: complete `go test ./internal/templates ./pkg/cli -count=1`.
- PASS: `go test ./... -count=1`.
- PASS: `go vet ./...`.
- PASS: `go test -race ./internal/templates ./internal/worktree ./pkg/cli
  -count=1`.
- PASS: `golangci-lint run --new-from-rev=origin/main ./...` with zero issues.
- PASS: `make build` and a dedicated `/tmp/kit-gh116-current` binary build.
- PASS: built-binary `rules view codex-thread-initialization`, `capabilities
  reconcile --json`, and `check 0055-codex-thread-initialization` smoke tests.
- PASS: built-binary whole-project reconcile reached the new semantic audit
  and reported complete source-file-size evidence for 810 candidates, 504
  eligible handwritten source/test files, and zero violations; its only
  findings were the expected pre-completion progress-summary entries for 0055.
- PASS: persisted `weekly-kit-health` retains its active Wednesday 13:00
  schedule, `gpt-5.6-sol` model, high reasoning effort, local execution, Kit
  project target, and cwd while adding the Codex capability gate, fourth
  evidence dimension, no-op gate, repair rule, and final-report field.
- PASS: post-validation Constitution curation found no new project-wide
  invariant; the provider-specific durable contract is fully routed through
  this spec and the downstream ruleset.

## OUTCOME

New and refreshed V3 `AGENTS.md` files now begin with an ordered Codex hard
gate that requires rename before pin and prohibits substantive work before
both results resolve. Successful host results are verified when available;
unsupported, unavailable, or failed operations must be reported in the first
commentary before work continues. Claude and Copilot remain free of Codex-only
task-state instructions. Whole-project reconcile detects weakened existing
AGENTS semantics, capability metadata exposes that behavior to automation, and
Weekly Kit Health treats the gate as independent fail-closed evidence.

## REPOSITORY MEMORY

Decision: created

Rationale: The provider boundary, first-response ordering, capability lookup
exception, verification behavior, and fail-visible fallback are durable
workflow decisions that code and tests alone cannot fully explain.

Artifacts:

- `docs/specs/0055-codex-thread-initialization/SPEC.md`
- `docs/references/rules/codex-thread-initialization.md`
- `AGENTS.md`
