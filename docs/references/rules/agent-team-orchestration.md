---
kind: ruleset
slug: agent-team-orchestration
description: Controls capability-aware supervisor, child-agent, continuity, verification, and truthful degradation policy.
status: active
registry_scope: downstream
applies_to:
  - coding-agent
  - workflow
  - dispatch
  - subagent
  - verification
read_policy_default: conditional
---

# Ruleset: agent-team-orchestration

## Purpose

Use one accountable root supervisor with capability-matched child agents only
when that topology improves correctness or throughput.

This ruleset controls execution topology. It does not relax workflow phase
gates, source-of-truth rules, dirty-worktree ownership, validation, reflection,
or GitHub delivery gates.

Kit remains a prompt and repository-evidence adapter. Kit commands do not
inspect a host's live agent roster, select a model or effort, launch or wait for
an agent, supervise execution, or claim an agent result. The active coding
agent applies this contract using capabilities the host actually exposes.

## Applies When

- A coding agent plans implementation, validation, review, or repair work.
- Work may split into backend, frontend, CLI, tests, docs, data, security,
  compatibility, validation, or repository-research lanes.
- `kit dispatch`, `kit pr fix`, a prompt-library dispatch prompt, CI dispatch,
  or another subagent-capable workflow is used.
- A living or historical feature workflow needs an Agent Team Plan.

Use one supervisor lane, and record the reason, when the work is trivial,
tightly coupled, high-overlap, requires continuous design judgment, the user
requested single-agent execution, or the active host does not confirm separate
execution. An explicit requirement for actual children or independent
verification remains unsatisfied when the host cannot provide it; logical
decomposition is not a substitute.

## Rules

### Lifecycle And Root Accountability

Use this execution lifecycle:

```text
SCOPED -> CAPABILITY_NEGOTIATING -> MAPPING -> SYNTHESIZING
  -> DRILL_DOWN -> SYNTHESIZING -> PLAN_READY -> EXECUTION_READY
  -> IMPLEMENTING -> INTEGRATING -> VERIFYING -> REPAIRING -> VERIFYING
  -> SOURCE_VERIFIED -> PR_READY -> MERGE_AUTHORIZED -> MERGED
  -> RELEASE_VERIFIED -> PROVENANCE_PR_READY -> COMPLETE
```

- `SCOPED` means goals, non-goals, acceptance criteria, material unknowns, and
  mutation boundaries are explicit enough to negotiate capabilities.
- The `SYNTHESIZING -> DRILL_DOWN -> SYNTHESIZING` loop repeats while mapping
  results conflict or material questions remain. Reuse the responsible stable
  agent reference for focused drill-down whenever continuity is confirmed.
- `PLAN_READY` requires the evidence-backed convergence gate below.
  `EXECUTION_READY` additionally requires a current Capability Manifest, lane
  assignments, overlap controls, and every applicable repository mutation gate.
- The `VERIFYING -> REPAIRING -> VERIFYING` loop repeats until findings are
  closed with rerun evidence or accepted as explicit residual risk.
  `SOURCE_VERIFIED` requires the integrated source, tests, documentation, and
  evidence to agree; implementation completion alone cannot enter that state.
- `PR_READY` never implies `MERGE_AUTHORIZED`. That transition requires a
  direct user request or accepted bounded merge plan for the exact current
  pull-request set plus revalidated head, base, actor, repository policy,
  reviews, checks, dependencies, method, and material effects.
- `MERGED` requires observed evidence that the exact authorized pull request
  merged. Merge success is not release evidence.
- `RELEASE_VERIFIED` requires evidence tying the expected merged source to the
  exact release tag, hosted workflow, and artifacts; it is not inferred from a
  merge, tag name, or local build.
- `PROVENANCE_PR_READY` requires a separately authorized and owned
  issue/branch/worktree/pull-request lane whose ready-PR evidence records the
  verified release provenance. It never implies merge authorization for that
  separate pull request.
- `COMPLETE` requires evidence for every applicable prior state. A workflow
  that does not own merge, release, or provenance reports those later states as
  outside its observed outcome; it must not fabricate them or skip through them
  to claim lifecycle completion.
- Missing required evidence or authorization produces a literal blocked task
  outcome at the current state; it never advances the lifecycle optimistically.
- The root supervisor owns the user request, active durable artifact, scope,
  non-goals, assumptions, acceptance criteria, implementation plan, lane
  assignment, integration, conflict resolution, validation, evidence,
  delivery gating, and final response.
- At most one optional `orchestrator` advisor may assess depth, decomposition,
  overlap, and profile fit. The advisor is read-only and advisory; it never
  shares supervisor accountability.
- Only the root supervisor may launch children. Delegation depth is exactly
  one: advisors, mappers, specialists, precision agents, and verifiers must not
  launch descendants even when the host permits nesting.
- A follow-up sent to an existing child remains at the same delegation depth.
  A host-managed orchestration primitive is fully conformant only when its
  evidence establishes separate results and the required one-level boundary.

### CAPABILITY_NEGOTIATING

Enter `CAPABILITY_NEGOTIATING` after `SCOPED` and before `MAPPING`. Build a
Capability Manifest from host documentation, currently exposed
controls, returned stable references, and other direct runtime evidence. Do
not launch a sacrificial child merely to discover capacity.

Record each capability as `confirmed`, `unavailable`, or `unknown`, with its
evidence basis:

```yaml
state: CAPABILITY_NEGOTIATING
evidence_basis: []
separate_execution: confirmed | unavailable | unknown
parallel_execution: confirmed | unavailable | unknown
stable_agent_references: confirmed | unavailable | unknown
same_agent_follow_up: confirmed | unavailable | unknown
model_selection: confirmed | unavailable | unknown
effort_selection: confirmed | unavailable | unknown
fresh_verification: confirmed | unavailable | unknown
wait_status_controls: confirmed | unavailable | unknown
effective_capacity: <host-confirmed value> | host-managed | unavailable | unknown
selected_topology: single-supervisor | root-with-children | host-managed
delegation_depth: zero | one | unknown
degradations: []
```

- Preserve `unknown` literally in evidence and final reporting, but treat it
  as unavailable for routing. Never infer parallelism, continuity, model
  control, or verifier independence from a generic agent label.
- Count an actual agent only when the host creates separate execution and
  returns a separate result. A role prompt, task list, editor mode, handoff, or
  manually opened conversation is a logical lane.
- Refresh the manifest when the host roster, controls, effective capacity,
  scope, or selected topology materially changes, then return to
  `CAPABILITY_NEGOTIATING` before remapping affected lanes.

### Semantic Capability Profiles

Profiles describe required behavior, not vendor products or fixed model IDs:

| Profile | Required capability |
| --- | --- |
| `architect` | Deep planning, cross-boundary design, tradeoff analysis, and final synthesis. |
| `orchestrator` | Read-only complexity assessment, decomposition, overlap analysis, and profile routing. |
| `mapper` | Read-heavy repository discovery, pattern identification, and evidence-backed source mapping. |
| `specialist` | Balanced implementation and focused validation in a well-bounded logical area. |
| `precision` | Strongest justified reasoning for high-risk, ambiguous, security-sensitive, or compatibility work. |
| `verifier` | Fresh, read-only, adversarial validation of requirements, diff, tests, and evidence. |

- Map each requested profile against the active host's live roster, model and
  effort controls, and lane risk. Do not encode vendor or version names in this
  normative rule.
- Record both `requested_profile` and `effective_profile` for every actual or
  logical lane. Record the evidence supporting the effective mapping; use
  `unknown` when the host does not reveal it.
- Never silently substitute a different configuration for an exact user model
  or effort pin. Obtain explicit relaxation or report the pin unsatisfied.

### Profile Fallback Order

For a semantic profile request that the host cannot satisfy exactly, use this
order and record every fallback:

1. an equal-or-stronger eligible configuration for the lane's relevant risks;
2. a narrowed, low-risk lane paired with stronger verification;
3. a runtime-selected and explicitly unverified configuration; then
4. `BLOCKED` when no acceptable configuration remains.

Do not equate price, recency, branding, or a marketing tier with capability.
An exact user pin does not enter this fallback chain unless the user explicitly
relaxes it.

### Required Agent Team And Lane Manifest

Before implementation, produce an Agent Team Plan containing the Capability
Manifest plus a Lane Manifest for every proposed, actual, logical, or omitted
lane:

```yaml
lane_id: stable-logical-id
objective: bounded outcome
execution_kind: actual_agent | logical_lane | omitted
requested_profile: profile
effective_profile: profile | runtime-selected | unknown
profile_evidence: []
stable_agent_ref: host-reference | unavailable | not-applicable
acceptance_criteria: []
source_evidence: []
predicted_files: []
areas_to_avoid: []
overlap_and_serialization: []
validation: []
degradations: []
```

The plan must also identify supervisor and optional advisor responsibilities,
host-confirmed effective capacity, selected topology, serialized work,
intentionally omitted implementation or verification lanes, the reason for
each omission, and final integration ownership.

### Host-Owned Capacity And File Overlap

- The active host owns admission, scheduling, available slots, and effective
  capacity. Kit defines no default or ceiling for concurrent agents.
- Start parallel work only when the host confirms parallel execution and
  current capacity, and predicted file and interface overlap is low. When
  capacity is `unknown`, serialize work or use one supervisor lane.
- Refresh host capacity before a new execution wave. Never request as many
  agents as possible or probe capacity by spawning disposable work.
- If lanes would edit the same file or unstable interface, prefer serial
  execution, assign one implementation owner and a later read-only reviewer,
  or split responsibilities differently. Parallel overlap requires an
  explicit low-risk integration plan.

### Stable Continuity And Rebriefing

- Preserve each `lane_id` and its stable host agent reference across discovery,
  drill-down, implementation, repair, and focused follow-up when same-agent
  continuity is confirmed.
- Ask focused follow-up questions of that same agent instead of spawning a
  replacement merely because the first result needs deeper research.
- If the reference is unavailable, expires, or cannot accept follow-up, keep
  the `lane_id`, create a replacement only when justified, and provide a full
  rebrief: objective, scope and non-goals, acceptance criteria, established
  decisions, source evidence, file ownership, prior results and findings,
  remaining questions, and validation expectations.
- Record the old and replacement references when exposed, the reason, rebrief
  evidence, and `continuity_loss`. Never describe a replacement as the same
  agent.

### Lane Assignment And Mutation Boundaries

Each child receives a clear objective, relevant acceptance criteria, source
evidence, expected files or packages, areas to avoid, validation expectations,
output format, requested and effective profiles, and an instruction to report
blockers instead of guessing.

Children must not independently expand scope. They must not create or switch
worktrees, create branches, stage files, commit, push, open pull requests,
resolve review threads, merge, queue a merge, or mark the whole workflow
complete unless that exact mutation is explicitly assigned, authorized, and
allowed by the active repository rules.

### Merge-Wave Authority

- Only the accountable supervisor owns merge-wave decisions, authorization
  reconciliation, the global ready frontier, and global gate advancement.
- A participant may merge only specifically assigned pull-request nodes from
  the exact authorized `MERGE_READY` frontier. Subagent assignment alone never
  creates merge authority.
- A participant must not expand the approved pull-request set, change another
  node's dependency or ownership, bypass a gate, or advance program-wide state.
- Read-only verification agents can never merge or queue a merge.
- Independent authorized ready nodes may merge concurrently only when their
  repository and failure boundaries are independent. Dependency chains and
  same-base-sensitive operations remain serialized.
- Every merge participant must load `github-pr-merge`; cross-repository
  programs also reconcile the approved set and ready frontier from the
  canonical ledger before each wave.

### Verification, Repair, And Degradation

- After nontrivial implementation, use a fresh independent `verifier` when the
  host confirms that capability. The verifier must have a distinct execution
  and result, must not be the implementer or advisor, and remains read-only.
- Verification agents must not edit files, stage changes, commit, push, close
  findings, mark acceptance criteria complete, resolve review threads, or
  mutate issue, branch, pull-request, merge, or merge-queue state.
- When fresh verification is unavailable or unknown, the root supervisor must
  perform a distinct read-only self-review and report
  `verification_independent: false`. This is degraded conformance; it is
  unsatisfied when independent verification was explicitly required.
- Return a repair to the same stable implementation agent when continuity is
  confirmed. Otherwise use the replacement and full-rebrief contract.
- Each finding records a gap ID, related acceptance criterion, evidence
  inspected, actual and expected behavior, risk, recommended fix area, whether
  delivery is blocked, rerun evidence, and verifier closure.
- Do not proceed to reflection or delivery with open verification gaps unless
  the user explicitly accepts the residual risk.

### Evidence-Backed Goal Convergence

- Read the top-level project `goal_percentage` as the `PLAN_READY` threshold.
  Default to `95` only when it is absent. A configured value that fails Kit's
  config validation blocks readiness.
- Derive auditable goal coverage from explicit goals, acceptance criteria, and
  material questions. Unless the durable artifact defines weights, count each
  item once; do not subdivide items to inflate coverage.
- Mark an item covered only when repository, host, user, or validation evidence
  resolves it. Evidence-backed coverage is not subjective model confidence.
- Enter `PLAN_READY` only when coverage meets the configured threshold and
  there are zero unresolved material questions. Continue discovery,
  negotiation, or focused same-agent follow-up otherwise.
- Keep the threshold, coverage calculation, evidence links, remaining gaps,
  and readiness decision in native task context, not specification front
  matter or an agent transcript.

### Two-Axis Final Reporting

Follow `agent-completion-output` for terminal reporting. Map `task_outcome` to
the overall status heading and retain the fields below in the coordination or
evidence list. Never let a successful task heading hide degraded or
unsatisfied orchestration conformance.

Report task success separately from orchestration compliance:

```yaml
task_outcome: workflow-native outcome
orchestration_conformance: full | degraded | unsatisfied
goal_percentage_threshold: configured value
goal_coverage: evidence-backed value
capability_unknowns: []
actual_agents: []
logical_lanes: []
omitted_lanes: []
requested_and_effective_profiles: []
continuity_replacements: []
verification_independent: true | false | unknown
degradations: []
```

- `full` means all applicable topology obligations and explicit requirements
  were met with evidence.
- `degraded` means an allowed fallback completed with every limitation
  reported and no explicit mandatory topology requirement was misrepresented.
- `unsatisfied` means an explicit actual-child, independent-verification,
  exact-pin, continuity, or other mandatory orchestration requirement was not
  met, regardless of ordinary task success.
- A successful logical single-agent fallback may have a successful
  `task_outcome` while orchestration is `degraded` or `unsatisfied`.
- Report actual agents and their stable references separately from logical and
  omitted lanes. If no separate agents actually ran, state exactly:

```text
single supervisor lane; no specialist or verification agents spawned
```

## Anti-Patterns

- Hardcoding provider, model, version, effort, or numeric concurrency policy in
  this rule.
- Asking Kit to inspect host capabilities, choose a model, launch an agent,
  wait for results, or supervise execution.
- Inferring capability from a provider label, treating `unknown` as available,
  or probing capacity with a disposable child.
- Claiming a role prompt, task list, handoff, or manually opened conversation
  is an actual agent.
- Allowing a child to launch descendants or share root accountability.
- Replacing a stable agent without a full rebrief or hiding continuity loss.
- Silently substituting an unavailable exact model or effort pin.
- Calling supervisor self-review independent verification.
- Treating task success as proof of full orchestration conformance.
- Copying transient capability or convergence state into spec front matter.
- Treating participant assignment as merge authority or allowing a participant
  to expand the authorized pull-request set or advance a global gate.

## Verification

- Confirm `CAPABILITY_NEGOTIATING` occurred before lane mapping and the
  Capability Manifest records evidence, unknowns, selected topology, effective
  capacity, delegation depth, and degradations.
- Confirm actual agents and logical lanes are distinguished and every lane
  records requested and effective profiles with evidence.
- Confirm capacity came from the host without a Kit numeric default or ceiling,
  and overlapping work was serialized or explicitly integrated.
- Confirm only the root launched children, delegation stayed one level deep,
  and the optional advisor remained read-only.
- Confirm stable references were reused for follow-up or replacements received
  full rebriefs and reported continuity loss.
- Confirm a fresh read-only verifier ran when supported, or supervisor
  self-review and lost independence were reported literally.
- Confirm `PLAN_READY` used the project `goal_percentage`, evidence-backed
  coverage, and zero unresolved material questions.
- Confirm fallbacks followed the required order and exact user pins were not
  silently substituted.
- Confirm final reporting separates `task_outcome` from
  `orchestration_conformance` and reports actual, logical, omitted, continuity,
  verification, and degradation evidence.
- For merge waves, confirm only the supervisor decided the frontier, every
  participant received exact authorized nodes, verifiers remained read-only,
  and dependency or same-base-sensitive operations were serialized.

## Examples

Confirmed child execution with continuity:

```text
The host confirms separate execution, same-agent follow-up, and current
capacity. Map low-overlap lanes to semantic profiles, preserve each returned
stable reference, and send focused drill-down or repair questions to that same
agent.
```

Continuity loss:

```text
The original lane reference no longer accepts follow-up. Preserve the lane ID,
record continuity loss, fully rebrief a justified replacement, and report both
references when the host exposes them.
```

No confirmed child primitive:

```text
Keep one supervisor lane, preserve child capability as unknown or unavailable,
perform a distinct read-only self-review, and report task outcome separately
from degraded or unsatisfied orchestration conformance.
```
