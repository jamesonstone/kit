---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: "0081"
  slug: retire-thread-naming
  dir: 0081-retire-thread-naming
relationships:
  - type: related_to
    target: 0079-thread-naming
  - type: related_to
    target: 0055-codex-thread-initialization
references:
  - id: naming-spec
    name: Historical conversation naming spec
    type: spec
    target: docs/specs/0079-thread-naming/SPEC.md
    relation: supersedes
    read_policy: evidence
    used_for: original naming policy being retired
    status: active
  - id: pin-spec
    name: Historical Codex thread initialization spec
    type: spec
    target: docs/specs/0055-codex-thread-initialization/SPEC.md
    relation: supersedes
    read_policy: evidence
    used_for: original pin and rename gate being retired
    status: active
  - id: issue
    name: Remove conversation naming and Codex thread pin/init
    type: external
    target: https://github.com/jamesonstone/kit/issues/213
    relation: supports
    read_policy: must
    used_for: accepted scope and delivery identity
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Stop Kit from telling coding agents to rename or pin conversations. New and
refreshed scaffolds must not install a conversation-naming policy, Codex
thread pin/init gate, or the `instructions naming` / `instructions title`
helpers.

## CONTEXT

- Feature 0079 installed a shared `[scope] domain / objective` policy, always-
  loaded Conversation Naming pointers, `docs/references/thread-naming.md`, and
  read-only CLI helpers.
- Feature 0055 installed a Codex-only ordered pre-response gate: rename with
  `set_thread_title`, then pin with `set_thread_pinned`. 0079 later pointed
  that gate at the shared naming policy.
- The user asked to remove the thread-naming rule, then also the thread
  pinning rule, on one new worklane and PR. Together those are the always-
  loaded Conversation Naming section and the Codex Thread Initialization Hard
  Gate.
- Topology: single-lane, because templates, generated instructions, CLI,
  frozen instruction version, rules catalog, and tests must stay consistent.
- Landing plan: repository `jamesonstone/kit`; issue #213; branch `GH-213`;
  worktree `~/worktrees/jamesonstone/kit/GH-213`; protected base `main`;
  create one ready PR assigned to jamesonstone. Merge is not authorized.
- Historical instruction versions stay immutable. Current global instructions
  become v17 without naming or session pin/init.

## REQUIREMENTS

- Remove Conversation Naming from generated `AGENTS.md` and `CLAUDE.md` and
  from v1/v2/v3 scaffolds.
- Remove the Codex Thread Initialization Hard Gate from generated `AGENTS.md`.
- Stop installing `docs/references/thread-naming.md`. Delete its canonical
  source and this repository's projection.
- Remove `kit instructions naming` and `kit instructions title` entirely, not
  as hidden aliases. Update `kit capabilities`, commandset, and Constitution.
- Stop distributing `docs/references/rules/codex-thread-initialization.md`.
  Remove it from this repository's registry state and references index.
- Publish frozen `v17` as the current `kit instructions` version. v16 remains
  the last version that required naming and Codex pin/init.
- Keep Codex browser policy and capability-aware subagent binding.
- Keep 0055 and 0079 specs as history. Do not rewrite v1-v16 instruction
  files.
- Append-only refresh must not re-inject the retired sections. Leftover
  sections in existing projects may remain until operators delete them;
  reconcile should treat the retired always-loaded wording as forbidden.
- Non-goals: changing host products' own rename/pin UIs; deleting historical
  specs or notes; merge.

## ACCEPTED PLAN

1. Expand issue #213 and SPEC 0081 to cover naming plus Codex pin/init.
2. Drop the always-loaded pointers and Codex init gate from instruction
   templates; sync checked-in `AGENTS.md` and `CLAUDE.md`.
3. Remove the naming CLI, `internal/threadtitle`, and support-doc install
   path.
4. Delete the naming policy and Codex initialization ruleset. Update host
   adapters, README, references index, Constitution, and capabilities.
5. Add `v17` without the naming/init section. Invert tests to prove absence.
6. Validate with focused package tests, `go test ./...`, format, vet, and a
   source-size audit of touched Go files. Deliver one ready PR for GH-213.

## DECISIONS

- Remove the whole Codex Thread Initialization Hard Gate, not pin-only.
  Naming is already being removed, so a title-only leftover gate would still
  be a thread-naming rule.
- Do not mechanically strip leftover sections from downstream projects.
  Append-only refresh never deletes user-owned sections; forbidden-guidance
  reconcile findings flag leftovers.
- Keep Codex browser and subagent adapter text. Those are unrelated to
  rename/pin.

## DISCOVERIES

`gh` in this environment has an invalid token. GitHub mutations use the
authenticated GitHub plugin as `jamesonstone`. Git SSH push still works.

## VALIDATION

- PASS: `go test ./...`
- PASS: `go vet ./...`
- PASS: `gofmt` on touched Go files
- PASS: `git diff --check`
- PASS: source-file-size audit of version-control-eligible handwritten source and test files: complete, 0 violations. Largest touched Go file is `pkg/cli/reconcile_guidance_expectations.go` at 292 lines.
- PASS: `kit check 0081-retire-thread-naming`, `kit check 0055-codex-thread-initialization`, `kit check 0079-thread-naming`
- PASS: `kit context resolve --workflow implementation-delivery --feature 0081-retire-thread-naming` returned `blocked: false`
- NEVER-RUN: hosted GitHub Actions checks. They exist only after the ready PR is opened.
- NEVER-RUN: `golangci-lint`. This repository's GitHub workflows do not run it.

## OUTCOME

Local implementation is complete. Frozen `v17` is the current `kit instructions` version. New and refreshed scaffolds omit Conversation Naming and the Codex Thread Initialization Hard Gate. `kit instructions naming` and `kit instructions title` are gone. The naming policy and Codex initialization ruleset are deleted. Historical v1-v16 files, 0055, and 0079 remain. GH-213 is the delivery lane; merge is not authorized.

## REPOSITORY MEMORY

This spec retains the retirement rationale: drop the whole Codex pin/init gate rather than pin-only, keep leftover downstream sections until operators delete them, and flag those leftovers through forbidden-guidance reconcile. Constitution already records that historical specs remain evidence when implementations retire; no new constitution invariant was added.
