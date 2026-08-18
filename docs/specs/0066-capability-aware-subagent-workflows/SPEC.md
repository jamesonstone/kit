---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: deliver
feature:
  id: 0066
  slug: capability-aware-subagent-workflows
  dir: 0066-capability-aware-subagent-workflows
relationships:
  - type: builds_on
    target: 0012-default-subagent-orchestration
  - type: builds_on
    target: 0042-native-plan-repository-memory
  - type: related_to
    target: 0059-conservative-coding-agent-first
references:
  - id: orchestration-rule
    name: Canonical agent-team orchestration contract
    type: documentation
    target: docs/references/rules/agent-team-orchestration.md
    relation: implements
    read_policy: must
    used_for: capability negotiation, delegation, continuity, and verification
    status: active
  - id: tooling-adapter
    name: Cross-agent tooling adapter
    type: documentation
    target: docs/agents/TOOLING.md
    relation: implements
    read_policy: must
    used_for: host-native capability mapping and provider examples
    status: active
  - id: codex-binding
    name: Managed Codex instruction binding
    type: code
    target: internal/templates/instruction_templates_v3.go
    relation: implements
    read_policy: must
    used_for: conditional Codex-native subagent controls
    status: active
  - id: dispatch-surface
    name: Dispatch command and shared prompt contract
    type: code
    target: pkg/cli/dispatch.go
    relation: implements
    read_policy: must
    used_for: breaking flag removal and provider-neutral orchestration prompts
    status: active
  - id: release-workflow
    name: Main-branch release workflow
    type: code
    target: .github/workflows/release-tag-main.yml
    relation: implements
    read_policy: must
    used_for: v3 tag selection and idempotent publication
    status: active
delivery_intent: new_issue_branch_pr_ready
---
# SPEC

## PURPOSE

Make Kit's existing subagent orchestration rule capability-aware and usable
across coding-agent hosts while optimizing the adapter for Codex's native
model, effort, child-thread, follow-up, waiting, and parallel-execution
controls. Kit remains prompt-only: it defines truthful negotiation and routing
behavior but never selects a model, launches a child, or supervises execution.

Publish the public breaking changes as Kit v3 so users receive valid Go module
semantics, deliberate CLI removals, and release behavior that can recover
idempotently without silently advancing tags.

## CONTEXT

- `agent-team-orchestration` already owns supervisor accountability, lane
  decomposition, overlap, concurrency, verification, and reporting. A second
  ruleset would create competing topology authority.
- Current active guidance hardcodes a default of three concurrent lanes and a
  hard maximum of four. The runtime host, not Kit, is the only reliable source
  for live capacity and scheduling constraints.
- Current `kit dispatch` and `kit pr fix` expose `--max-subagents` and carry the
  configured integer through capability metadata, prompts, review-loop
  plumbing, templates, workflows, tests, and goldens.
- Generated `AGENTS.md`, `CLAUDE.md`, and Copilot instructions are thin entry
  points. Shared provider-neutral runtime guidance belongs in `TOOLING.md`;
  Codex-only bindings can use the existing title-gated V3 template pattern.
- Warp's current project-rule convention consumes `AGENTS.md`; generating a
  parallel `WARP.md` would duplicate authority. Claude and Copilot already
  route through their existing Kit-generated entrypoints.
- Kit's module path is unversioned while the requested release is v3. Go major
  versions above v1 require `/v3` in the module path and exact self-imports.
- Main pushes currently advance a v2 patch tag before the full quality gate.
  The requested workflow must establish v3.0.0, reuse a qualifying same-SHA
  tag, fail on conflicts, validate before tagging, and recover publication
  without allocating another patch.
- The top-level `.kit.yaml` `goal_percentage` is the project convergence
  threshold. The legacy loop confidence setting is not orchestration policy.
- Historical specifications are immutable evidence. This living 0066 spec
  records the superseding v3 behavior without rewriting older feature records.
- Issue #156, branch `GH-156`, and the canonical non-primary worktree are the
  authorized feature-delivery lane. Merge and post-merge provenance changes
  remain separate authorization boundaries.

## REQUIREMENTS

### Canonical Capability-Aware Contract

- Extend only `agent-team-orchestration`; do not add another ruleset, runtime
  API, provider-profile generator, or vendor configuration schema.
- Add `CAPABILITY_NEGOTIATING` between scope and lane mapping. Record only
  host-confirmed facts for separate execution, parallelism, stable references,
  same-agent follow-up, model and effort selection, fresh verification,
  waiting/status controls, effective capacity, selected topology, delegation
  depth, evidence basis, and degradations.
- Treat unknown capabilities as unavailable for routing while preserving the
  literal `unknown` state in the final report. Never launch a sacrificial child
  to probe capacity.
- Count an actual agent only when the host creates separate execution and
  returns a separate result. A role prompt, task list, editor mode, handoff, or
  manually opened conversation is only logical decomposition.
- Keep one accountable root supervisor and at most one optional read-only
  orchestration advisor. Only the root launches children; delegation depth is
  exactly one even when a host permits nesting.
- Preserve `lane_id` to stable host agent reference across discovery,
  drill-down, implementation, and repair when the host supports continuity.
  Otherwise fully rebrief a replacement and report continuity loss.
- Use a fresh independent verifier when supported. Otherwise perform a
  distinct read-only supervisor self-review and report that verification was
  not independent.
- Report ordinary `task_outcome` independently from
  `orchestration_conformance: full | degraded | unsatisfied`. A logically
  successful single-agent fallback cannot satisfy an explicit requirement for
  actual children or independent verification.

### Capability Profiles And Routing

- Define provider-neutral profiles: `architect`, `orchestrator`, `mapper`,
  `specialist`, `precision`, and `verifier`.
- Map profiles against the active host's live roster and controls instead of
  fixed model IDs. Keep vendor/model examples illustrative and outside the
  normative rule.
- Use fallback order: an equal-or-stronger eligible configuration; a narrowed
  low-risk lane with stronger verification; a runtime-selected and explicitly
  unverified configuration; then `BLOCKED`.
- Never silently substitute an unavailable exact user model or effort pin.
- Use the top-level project `goal_percentage` as the evidence-backed
  `PLAN_READY` threshold, defaulting to 95 only when absent. Invalid configured
  values block. Store convergence evidence in native task context, not spec
  front matter.

### Host Adapters

- Add a concise generic adapter to the shared `TOOLING.md` template and its
  checked-in generated artifact. It must direct every host to inspect native
  child, model, effort, follow-up, waiting, capacity, and verifier controls;
  use a single supervisor when no child primitive is confirmed; and never
  claim a logical lane was a child.
- Put current illustrative mappings in `TOOLING.md`: strongest justified Codex
  configuration for architecture and precision, balanced read-heavy mapping,
  fast bounded implementation, and a fresh strong verifier; analogous
  Opus/Sonnet/Haiku classes for Claude; conservative semantic-profile requests
  for Copilot; native orchestration and host-owned parallelism for Warp/Oz.
- Add exactly one small conditional Codex binding to generated `AGENTS.md`
  beside existing Codex-specific hooks. It must use Codex's live roster and
  native spawn, follow-up, wait, model, and effort controls without duplicating
  the canonical rule.
- Keep that binding absent from generated `CLAUDE.md` and Copilot instructions.
  Warp reads `AGENTS.md` but skips the binding unless the active host is Codex.
- Keep the default instruction target set exactly AGENTS, CLAUDE, and Copilot.
  Generate no `WARP.md` and no provider agent-definition directories.

### CLI And Prompt Surface

- Remove `--max-subagents` from `kit dispatch` and `kit pr fix`; passing it must
  fail as an unknown flag.
- Remove associated constants, option fields, validation, help, capability
  metadata, prompt parameters, workflow plumbing, tests, and default-three or
  hard-four wording. Preserve `--single-agent`.
- Keep common prompt suffixes provider-neutral. Prompts must distinguish actual
  agents, logical and omitted lanes, requested and effective profiles,
  confirmed and unconfirmed parallelism, continuity replacements, and
  independent versus self verification.

### Go v3 And Release Behavior

- Change the root module to `github.com/jamesonstone/kit/v3` and update exact
  self-imports, command entrypoints, linker symbol paths, install guidance,
  upgrade behavior, migration notes, and release configuration. Do not rewrite
  GitHub repository, API, registry, or source-origin URLs that remain
  unversioned identities.
- Select v3.0.0 when no v3 tag exists and increment v3 patches afterward.
- Reuse a qualifying tag already attached to the same SHA; fail rather than
  move a conflicting tag.
- Run required quality gates before tag creation and permit idempotent release
  publication recovery without creating a second patch.
- Expose `workflow_dispatch` on the main release workflow so a skipped push
  event can re-enter the same quality-gated Mint and GoReleaser path without
  manual tag creation or a second release implementation. Reject manual
  dispatch from any ref other than `refs/heads/main`.
- Use `jamesonstone/mint@v0.2.1` in the two release workflows for immutable Git
  tag and GitHub Release state. Keep Kit's exact v3 selector, GoReleaser
  artifact and checksum production, and idempotent artifact upload; do not use
  Mint's conventional-commit resolver where it could change Kit's release
  selection policy.
- Serialize release-sensitive main work with `concurrency.queue: max` while
  treating third-party schemas that lag hosted GitHub syntax as a documented
  validation exception, not proof the hosted workflow is invalid.
- Preserve the separate post-merge provenance lifecycle: verify the exact
  merge SHA, v3.0.0 tag and artifacts first, then create a separate issue,
  branch, worktree, and ready PR for `.kit.yaml` provenance. Do not merge either
  PR without direct authorization for that exact PR.

### Observable Acceptance

- Semantic tests prove capability negotiation, profiles, one-level delegation,
  continuity, convergence, degradation, truthful reporting, and absence of
  fixed caps or normative vendor versions in the canonical rule.
- Template tests prove the Codex binding appears exactly once in AGENTS and is
  absent from Claude/Copilot; the generic adapter and illustrative mappings
  match generated artifacts; default instruction targets remain unchanged.
- Runtime fixtures cover full controls, limited routing/continuation,
  host-managed orchestration, unknown capacity, no child primitive,
  unavailable exact pin, and replacement rebrief after continuity loss.
- CLI tests prove the removed flag is absent and rejected while single-agent
  behavior remains intact. Active sources have no obsolete numeric policy.
- Module, import, linker, install, release-transition, tag-conflict, same-SHA,
  and publication-recovery tests pass for the v3 boundary.
- Release workflow tests require both ordinary main-push activation and the
  supported main-only manual recovery trigger.
- Bounded compatibility smokes report literal PASS or PARTIAL for Codex,
  Claude Code, Copilot CLI, and Warp/Oz based on runtime availability and
  authentication; documentation or prompt assertions are not runtime proof.

## ACCEPTED PLAN

1. Extend the canonical orchestration rule with capability negotiation,
   lifecycle states, profiles, continuity, degradation, convergence, and
   truthful two-axis reporting; capture the durable contract in this spec.
2. Add the conditional Codex AGENTS binding and generic TOOLING adapter, retain
   Claude/Copilot routing, and keep Warp on AGENTS without generating WARP.md.
3. Remove max-subagent CLI and prompt plumbing, make common prompts
   provider-neutral, and update workflow/template/golden mirrors.
4. Propagate stabilized policy wording through embedded and checked-in
   generated documents and reconciliation expectations.
5. Migrate exact self-imports across internal code, CLI production code, and
   tests only after semantic files stabilize.
6. Migrate the module/release boundary, including go.mod, linker paths, install
   and upgrade guidance, v3 notes, and tag/publication workflows.
7. Integrate centrally: format, tidy, reconcile generated artifacts, run the
   complete validation matrix, and curate repository memory.
8. Use fresh policy/adapter and Go-release verification passes. Return repairs
   to the lane that owns the affected surface.
9. Explicitly stage only issue #156 changes, commit and push as Jameson Stone,
   and create one ready pull request assigned to Jameson Stone. Stop before
   merge pending separate authorization.

## DECISIONS

- Extend the existing rule instead of creating a second orchestration policy;
  it already owns every relevant topology responsibility.
- Keep the normative contract provider-neutral and put live provider examples
  in TOOLING so roster changes do not rewrite project invariants.
- Optimize Codex through one conditional generated binding, not provider agent
  files or duplicated rule bodies.
- Treat capability evidence and runtime result truth as more important than
  requested topology. Degradation is allowed only when reported literally.
- Remove the public cap flags immediately in v3 rather than retaining a hidden
  compatibility sink, because the major-version migration is the explicit
  breaking boundary.
- Use a real `/v3` Go module rather than publishing a misleading v3 tag on an
  unversioned module path.
- Preserve repository-origin URLs and historical specs; the migration changes
  Go import identities and active release/install contracts only.
- Delegate immutable release-tag and GitHub Release state to Mint v0.2.1 while
  retaining Kit's exact first-v3.0.0-then-patch selector and GoReleaser artifact
  boundary. Keep creation, reuse, build, and upload recovery in the main
  workflow because a workflow-token tag push cannot trigger the tag workflow.
- Keep post-merge provenance in a separately authorized PR so the feature
  branch does not predict a merge SHA or accidentally create a second release.

## DISCOVERIES

- The connected GitHub app could not create issue #156 because its integration
  lacks issue-write permission; authenticated `gh` created the same scoped,
  human-assigned issue after recon confirmed the failed request made no issue.
- Initial context resolution correctly blocked before the spec existed and
  rejected directory evidence hints. Resolution must be repeated using this
  feature and concrete files after the spec is populated.
- No existing open issue matched the task and no `GH-156` branch, worktree, or
  pull request existed before lane creation.
- A tag pushed with the workflow-provided `GITHUB_TOKEN` does not trigger a
  second workflow run. The main release workflow therefore must publish both
  newly created and reused same-SHA tags itself; relying on the tag workflow
  would allow a green v3.0.0 tagging run with no GitHub Release artifacts.
- Mint v0.2.1 provides the required immutable tag reuse/conflict behavior and
  idempotent GitHub Release creation, but its conventional-commit resolver can
  select major or minor bumps. Kit therefore keeps `release-next-tag.sh` as the
  release-policy authority and calls only Mint's `release-tag` and
  `github-release` state operations.
- Go v3 required more than self-import replacement: the module declaration,
  linker symbol paths, installation and upgrade guidance, release selection,
  and publication recovery all had to move together. Repository, API, and
  registry URLs remain unversioned identities.
- Warp's current canonical documentation lives under `/platform/orchestration/`
  and `/agents/capabilities/rules/`; an older slash-command URL returned 404
  and was replaced before delivery.
- Runtime fixtures are useful only when each scenario is structurally complete
  and semantically asserted. Raw string presence across a whole fixture file
  could not prove full controls, safe unknown-capacity serialization, no-child
  fallback, exact-pin blocking, or continuity replacement behavior.
- Release checksums provide integrity verification but are not cryptographic
  signatures. The v3 migration guide now describes the updater as
  checksum-verified rather than signed.
- `concurrency.queue: max` is current hosted GitHub Actions syntax, but the
  locally available third-party action schema does not understand it. Hosted
  semantics and dedicated workflow tests are authoritative for this field.
- Reconciliation correctly proposes registry provenance for the changed
  canonical rule, but feature delivery intentionally leaves `.kit.yaml`
  untouched. Exact post-release provenance remains a separate issue, lane,
  and ready pull request after v3.0.0 is verified.
- The first hosted prompt-system check exposed an exact compatibility phrase
  that narrower Go tests did not cover: the dispatch benchmark requires
  `read-only verification`, while the new prompt initially said `read-only
  verifier`. Restoring the durable phrase in both shared prompt surfaces and
  asserting it directly made all 114 prompt-system assertions pass.
- A later mixed maintenance squash retained `[skip ci]` from an old source
  commit in the synthesized commit body, so GitHub created no main-push release
  run even after the PR title was corrected. The repository had no supported
  way to recover that absent run. `workflow_dispatch` now re-enters the same
  release workflow without bypassing Mint, quality gates, or artifact checks.

## VALIDATION

- Focused orchestration, template, instruction, context, CLI, release-prompt,
  and release-workflow package tests pass, including structural per-scenario
  YAML fixture assertions and generated/checked-in instruction equality.
- Full `go test ./...`, `go test -race ./...`, `go vet ./...`, formatting, and
  changed-path `golangci-lint` validation pass.
- Native, Windows cross-build, explicit v3.0.0 linker build, and GoReleaser
  snapshot validation pass. The snapshot reports the checkout's existing v2
  source tag, while the explicit linker build and release tests prove the v3
  binary identity and first-v3 transition.
- Both release workflows parse as YAML; the tag selector passes syntax,
  first-v3, patch, same-SHA reuse, conflict, and stale-tag tests. Dedicated
  tests require main-workflow publication for both create and reuse modes and
  require the manual recovery trigger.
- Both release workflows use only `jamesonstone/mint@v0.2.1` for tag and
  GitHub Release state, retain Kit's selector and GoReleaser upload boundary,
  and contain no manual tag or GitHub Release creation. A bounded Mint v0.2.1
  smoke created a tag, reused it on the same commit, and rejected moving it to
  a different commit.
- `kit check 0066-capability-aware-subagent-workflows`, `kit check --all`,
  `kit check --project`, source-size audit, reconcile output-only, dry-run
  reconcile review, module tidiness, diff checks, and active-source residue
  searches pass. The cross-repository workflow remains inapplicable and
  correctly blocks without a program ledger for this single-repository task.
- Context resolution is unblocked for implementation delivery and PR feedback
  repair with the updated canonical orchestration-rule digest.
- Codex compatibility: PASS. Actual parallel child execution, stable child
  follow-up, waiting, synthesis, and fresh independent policy and release
  verification were exercised. Requested model and effort were recorded;
  effective child runtime metadata was not independently exposed.
- Claude Code compatibility: PASS with degraded metadata. `CLAUDE.md` routed
  through shared Kit guidance to one actual isolated, non-delegating read-only
  child. Parent model was exposed; child effort and continuation were unknown.
- GitHub Copilot CLI compatibility: PASS with degraded metadata. Repository
  instructions were discovered and one native synchronous child returned a
  separate result; effective effort, continuation, and parallel capacity were
  not exposed.
- Warp/Oz compatibility: PASS with degraded metadata. Warp automatically
  loaded `AGENTS.md`, followed the shared routing chain, launched exactly one
  read-only non-delegating child through native orchestration, and returned an
  independent lifecycle result. Model selection was host-managed and the
  unexercised parallel, continuation, and effort axes remained unknown.
- Fresh policy/adapter and Go-release verifiers both report PASS after repairs.
  Neither verifier mutated repository or GitHub state.
- The deterministic prompt-system suite passes all 24 task runs and all 114
  assertions after the hosted compatibility repair.
- Hosted pull-request checks are recorded separately after the ready pull
  request exists. No merge, GitHub Release, proxy installation, upgrade, or
  post-release provenance claim is made by source validation.

## OUTCOME

Source implementation and independent verification are complete on issue #156
and branch `GH-156`. The canonical rule now negotiates live host capabilities,
maps provider-neutral profiles, preserves or truthfully replaces child
continuity, enforces one-level delegation, and reports task outcome separately
from orchestration conformance. Shared instructions adapt the contract across
Codex, Claude, Copilot, Warp/Oz, and single-agent hosts without generating a
second ruleset or provider agent files.

The CLI cap flags and fixed scheduling policy are removed, common prompts are
provider-neutral, the module is a valid `/v3` module, and the release workflow
can establish and idempotently publish v3.0.0. Mint v0.2.1 now owns immutable
tag and GitHub Release state while Kit retains exact version selection,
GoReleaser builds and checksums, and idempotent artifact upload. The same
quality-gated path is now available through `workflow_dispatch` when a push
workflow was skipped before a run existed. The recovery change is delivered
through issue #164 and its separately authorized pull request; merge and
release verification retain their own evidence gates.

## REPOSITORY MEMORY

- Created this living spec for material cross-provider, CLI, module, and
  release rationale that code and tests cannot preserve alone.
- Updated the existing canonical orchestration rule and shared
  tooling/instruction guidance rather than creating competing policy bodies.
- Added focused v3 migration and release notes for the public module and CLI
  boundary.
- Updated the existing release-workflow rationale for the supported manual
  recovery entrypoint; no parallel tag or release implementation was added.
- Historical specs remain unchanged.
- Constitution curation found no additional change necessary: Kit's prompt-only
  boundary already exists as project-wide truth, while capability profiles,
  provider adapters, and release recovery are feature-specific contracts
  preserved here and in their canonical rule and reference documents.
