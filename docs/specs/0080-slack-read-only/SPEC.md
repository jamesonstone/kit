---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: "deliver"
feature:
  id: "0080"
  slug: slack-read-only
  dir: 0080-slack-read-only
relationships:
  - type: related_to
    target: 0065-deletion-safety
  - type: related_to
    target: 0068-human-authorship
references:
  - id: registry-adoption
    name: Downstream ruleset adoption coverage
    type: code
    target: pkg/cli/init_refresh_ruleset_adoption_test.go
    relation: supports
    read_policy: must
    used_for: refresh, health, and reconcile propagation
    status: active
  - id: instruction-templates
    name: Managed instruction templates
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: always-loaded Slack write-approval gate
    status: active
  - id: issue
    name: Slack read-only default ruleset
    type: external
    target: https://github.com/jamesonstone/kit/issues/211
    relation: supports
    read_policy: must
    used_for: accepted scope and delivery identity
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Treat Slack as read-only by default in every Kit-managed project. Agents may
read and search Slack without extra approval, and must not send, reply, react,
edit, delete, or otherwise modify Slack unless the human explicitly authorizes
that specific action after seeing the complete final content.

## CONTEXT

- Slack tools are commonly available to coding agents. Drafting, investigating,
  or being asked to "respond" is easy to confuse with authorization to send.
- Kit distributes active downstream rulesets from GitHub `main` through
  `kit init --refresh`, `kit health`, and `kit reconcile --include-files`.
  Missing local rulesets are created as managed registry artifacts.
- A pointer-only rule is insufficient: an agent can send before loading it.
  Deletion-safety is the matching pattern: a concise always-loaded hard gate
  plus a full pointer-loaded ruleset.
- Topology: single-lane, because tightly coupled ruleset, template, instruction,
  and test edits require continuous design judgment.
- Landing plan: repository `jamesonstone/kit`; issue #211; branch `GH-211`;
  worktree `~/worktrees/jamesonstone/kit/GH-211`; protected base `main`;
  create one ready PR assigned to jamesonstone. Merge is not authorized.

## REQUIREMENTS

- Add an active downstream ruleset named `slack-read-only` with
  `read_policy_default: must` and `registry_scope: downstream`.
- Preserve the user's contract: read/search without approval; read the entire
  thread for a message link and inspect the containing channel when useful;
  draft-only for write/improve/respond requests; never send without explicit
  single-use, message-specific approval such as "send it," "send this," or
  "yes, send that message."
- Apply the same explicit-approval requirement to every Slack mutation, including
  posts, replies, edits, deletes, reactions, forwards, channel changes, and any
  other Slack state change.
- When uncertain, do not write to Slack; ask.
- Compose a concise always-loaded hard gate into generated provider instructions
  and GUARDRAILS. Keep the full protocol in the ruleset.
- Route load timing through RLM, the references index, Constitution baseline,
  and optional implementation-delivery evidence.
- Install the ruleset through the mandatory downstream adoption inventory and
  health/reconcile managed-safety coverage.
- Do not bump immutable `kit instructions` versions solely for this rule.
- No new command, flag, or JSON schema.

### Observable Acceptance

- The rule parses as a valid mandatory downstream ruleset.
- Refresh/health/`kit reconcile --include-files` adopt it as managed when missing.
- Generated and checked-in V1-V3 instructions and GUARDRAILS carry the hard gate.
- RLM, references index, Constitution baseline, and implementation-delivery
  route to the ruleset.
- Focused tests cover semantics, routing, adoption, and generated alignment.

## ACCEPTED PLAN

1. Add `docs/references/rules/slack-read-only.md` as the canonical downstream
   ruleset and this living spec.
2. Add one shared Slack hard gate and compose it into every supported generated
   provider instruction and generated GUARDRAILS surface.
3. Route the rule through RLM, references indices, Constitution baseline,
   implementation-delivery optional evidence, adoption inventory, managed-safety
   stubs, and reconcile expectations.
4. Sync checked-in generated artifacts, add focused tests, validate, curate
   repository memory, and deliver one ready PR from `GH-211`. Stop before merge.

## DECISIONS

- Accepted: always-loaded concise hard gate plus full ruleset, matching
  deletion-safety rather than pointer-only human-authorship, because Slack send
  tools are available before RLM would load a conditional file.
- Accepted: `read_policy_default: must` so linked or selected Slack work loads
  the full protocol; keep implementation-delivery selection optional because the
  hard gate already states the standing invariant.
- Accepted: publish from `jamesonstone/kit` so GitHub `main` registry fetch
  installs the rule downstream. A project-local copy would not enter the registry.
- Accepted: no new `kit instructions` version. Managed AGENTS/CLAUDE/Copilot
  refresh plus the registry ruleset is the distribution path.
- Recorded topology: single-lane, because tightly coupled wording and precedence.

## DISCOVERIES

- Registry fetch is `https://api.github.com/repos/jamesonstone/kit/contents/docs/references/rules?ref=main`. Downstream adoption is live only after this rule lands on `main`.
- `TestRunInitRefresh_InstallsMandatoryDownstreamRules` is the explicit
  regression inventory for rulesets that refresh must install.
- Checked-in `AGENTS.md`, `CLAUDE.md`, Copilot instructions, `GUARDRAILS.md`,
  and `RLM.md` must equal V3 generator output.

## VALIDATION

- PASS: `go test ./internal/templates ./pkg/cli -count=1` and `go test ./... -count=1`.
- PASS: `go vet ./...`; `gofmt -l` clean on affected Go files; `git diff --check` clean.
- PASS: affected handwritten Go source/test files are at or under 300 lines; whole-project reconcile source-file-size audit reported 404 eligible files, 0 above 300 lines.
- PASS: `kit check 0080-slack-read-only`.
- FAIL (pre-existing, not introduced here): `kit check --project` and `kit reconcile --all --output-only` report feature number `0077` duplicated by `0077-auto-assign-initiator` and `0077-unstructured-completion-output`.
- PASS: focused adoption, health/reconcile managed-safety, and stale-guidance tests install and detect `slack-read-only`.

## OUTCOME

- `slack-read-only` is an active downstream registry ruleset. After this lands on GitHub `main`, `kit init --refresh`, `kit health`, and `kit reconcile --include-files` install it as managed content in Kit-managed repositories that do not already have a local copy.
- Generated V1-V3 instructions and GUARDRAILS carry the always-loaded hard gate. RLM, references, Constitution baseline, and implementation-delivery route to the full protocol.
- Ready PR pending for GH-211. Merge is not authorized; live registry fetch stays on `main` until merge.

## REPOSITORY MEMORY

- Artifacts: `docs/specs/0080-slack-read-only/SPEC.md`; `docs/references/rules/slack-read-only.md`; generated instruction templates and checked-in AGENTS/CLAUDE/Copilot/GUARDRAILS/RLM mirrors; Constitution baseline bullet; references index; implementation-delivery optional evidence; adoption and reconcile tests.
- Decision: Constitution update required. Rationale: Slack write-approval is a project-wide agent invariant for every Kit-managed project, not feature-local rationale. The Kit-managed baseline now requires read-only Slack by default and explicit, message-specific approval before any Slack write.
- Curation: baseline bullet added to `docs/CONSTITUTION.md`, `internal/templates/templates.go`, and `pkg/cli/init_refresh.go`; reusable practice lives in the new ruleset.
