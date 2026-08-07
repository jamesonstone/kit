---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0056"
  slug: "codex-browser-policy"
  dir: "0056-codex-browser-policy"
relationships:
  - type: builds_on
    target: 0053-testing-and-environment-validation
  - type: builds_on
    target: 0055-codex-thread-initialization
references:
  - id: agents-template
    name: Managed AGENTS template
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: generated Codex browser policy
    status: active
  - id: testing-rule
    name: Testing and environment validation rule
    type: rule
    target: docs/references/rules/testing-and-environment-validation.md
    relation: constrains
    read_policy: must
    used_for: explicit external-browser authorization and lifecycle cleanup
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make Codex interactive browser work default to its isolated built-in
`@Browser`, prohibit silent external Chrome or browser-automation fallback,
and require verified cleanup whenever the user explicitly authorizes an
external browser.

## CONTEXT

- V3 `AGENTS.md` is generated from Kit's provider-specific instruction
  template and is the correct always-loaded surface for Codex browser-tool
  selection.
- The existing testing rules currently default automated browser work to
  Playwright-managed Chromium or Chrome for Testing. That guidance can cause
  a coding task to launch external browser processes even though Codex's
  built-in browser is the user's required interactive default.
- Browser lifecycle guidance already distinguishes task-owned resources from
  user-owned browser processes and requires scoped cleanup. The new policy
  should preserve that ownership model for explicitly authorized external
  browser runs.
- GitHub issue #129 and branch `GH-129` own this Kit-only change. Updating
  downstream repositories or global Codex configuration is outside this lane.

## REQUIREMENTS

- Add the requested `Browser policy` section to generated and checked-in V3
  `AGENTS.md` files.
- For interactive browser work, require Codex's built-in `@Browser`.
- Prohibit `@Chrome`, control of the user's active Chrome profile, and external
  Chrome or Chromium launched through Playwright, Selenium, Cypress, or
  browser MCP tools unless the user explicitly requests it.
- Report `@Browser` unavailability instead of silently falling back.
- When the user explicitly authorizes an external browser, terminate and
  verify all task-owned browser and automation processes before completion.
- Keep this browser-selection policy Codex-specific. Do not add it to generated
  `CLAUDE.md` or GitHub Copilot instructions.
- Preserve the existing testing rule's task-ownership and cleanup protections,
  but remove its contradictory default external-browser launcher choice.
- Make V3 semantic reconciliation detect a missing or weakened Browser policy
  in customized existing `AGENTS.md` files.
- Add focused tests for exact generated content, provider isolation,
  checked-in alignment, ruleset alignment, and stale-guidance detection.

### Non-goals

- Adding a browser launcher, process supervisor, or cleanup implementation to
  Kit.
- Controlling, terminating, or modifying a user-owned Chrome profile.
- Claiming repository instructions can make `@Browser` available when the host
  does not expose it.
- Changing V1 or V2 provider-neutral instruction files.

## ACCEPTED PLAN

1. Add one Codex-specific Browser policy generator beside the existing Codex
   thread-initialization gate and include it only in V3 `AGENTS.md`.
2. Align the checked-in `AGENTS.md` and canonical testing rule with the new
   external-browser authorization boundary.
3. Extend V3 reconcile expectations and focused tests so generated, checked-in,
   and customized managed instructions retain the policy.
4. Run focused and complete Go validation, lint, build, generated-output,
   reconcile, diff, and source-size checks.
5. Curate repository memory and deliver issue #129 through `GH-129` as a ready
   pull request.

## DECISIONS

- Keep the policy in a dedicated Codex-only generator composed immediately
  after the thread-initialization gate. The shared V3 instruction body remains
  provider-neutral.
- Treat external-browser authorization as bounded to the requested task-owned
  run. It does not authorize control of the user's Chrome profile or a silent
  fallback on later tasks.
- Preserve the 90-line thin-entrypoint baseline while allowing each exact
  generated instruction contract to define a larger provider-specific minimum.
  This keeps the new canonical `AGENTS.md` valid without globally weakening
  manual-duplication detection.

## DISCOVERIES

- The V3 generator already has a provider-specific composition point for
  Codex thread initialization; browser policy can use the same boundary
  without duplicating the provider-neutral instruction body.
- The canonical testing rule is the only Kit-managed source that actively
  recommends launching Playwright-managed Chrome or Chromium.
- The new Browser policy increased generated `AGENTS.md` beyond the validator's
  fixed 90-line limit. Existing integration tests exposed the false positive;
  deriving the provider/version limit from the current generated contract
  preserves both generated validity and thin-file enforcement.

## VALIDATION

- PASS: focused template, provider-isolation, checked-in alignment, init,
  ruleset, reconcile-guidance, and stale-guidance tests.
- PASS: regression tests for V3 project validation, Kit health refresh, and the
  generated-contract-aware thin-entrypoint limit.
- PASS: `go test ./... -count=1` after the validator repair.
- PASS: `go vet ./...`.
- PASS: `golangci-lint run --new-from-rev origin/main ./...` with zero issues.
- PASS: `make build`.
- PASS: affected handwritten source and test files remain at or below 300
  physical lines.
- PASS: built-binary `check 0056-codex-browser-policy` and `check --project`.
- PASS: built-binary whole-project reconcile reported no reconciliation needed
  and a complete source-file-size audit of 820 version-control-eligible
  candidates, 513 eligible handwritten source/test files, and zero files above
  300 physical lines.
- PASS: isolated built-binary `kit init --output-only` generated a V3
  `AGENTS.md` byte-identical to the checked-in file; generated Claude and
  Copilot instructions contained no Codex Browser policy.

## OUTCOME

- New and refreshed V3 `AGENTS.md` files contain the requested Codex Browser
  policy. Interactive work defaults to built-in `@Browser`; external browser
  control requires explicit authorization and retains scoped ownership,
  unconditional cleanup, and exit verification. Customized V3 instructions
  are semantically audited, provider-neutral instructions remain unchanged,
  and exact generated contracts remain valid thin entrypoints.

## REPOSITORY MEMORY

Decision: created

Rationale: Browser-provider selection, explicit authorization, user-versus-task
process ownership, and the relationship between always-loaded Codex guidance
and the canonical testing rule are durable decisions that code and tests alone
do not fully explain. No Constitution change is required because this is a
Codex-specific feature boundary rather than a new project-wide invariant.

Artifacts:

- `docs/specs/0056-codex-browser-policy/SPEC.md`
- `AGENTS.md`
- `docs/references/rules/testing-and-environment-validation.md`
- `docs/PROJECT_PROGRESS_SUMMARY.md`
