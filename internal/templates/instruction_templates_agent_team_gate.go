package templates

const multiAgentOrchestrationRoutingGate = `## Multi-Agent Orchestration Evaluation Hard Gate

- Before finalizing any native implementation plan for a new feature, a substantial architectural or behavioral change, or a multi-file refactor, load ` + "`docs/references/rules/agent-team-orchestration.md`" + ` and evaluate whether the work benefits from multi-agent or parallel decomposition using that rule's lifecycle and semantic capability profiles.
- A single mechanical edit, a direct question, or read-only research that never forms an implementation plan does not trigger this gate.
- Record the decision before the plan is finalized: either a multi-lane Agent Team Plan, or ` + "`single-lane, because <reason>`" + ` using that rule's single-lane criteria. Never skip the evaluation silently, even when the recorded answer is single-lane.
- This gate fires during plan formation and precedes the Work Lane Mutation Hard Gate below, which fires later, before the first repository mutation.

`

const multiAgentOrchestrationHardGate = `## Multi-Agent Orchestration Evaluation Hard Gate

Before a coding agent finalizes any native implementation plan for a new
feature, a substantial architectural or behavioral change, or a multi-file
refactor, it must:

1. Load ` + "`docs/references/rules/agent-team-orchestration.md`" + ` and
   enter its ` + "`CAPABILITY_NEGOTIATING`" + ` state.
2. Evaluate whether the work benefits from multi-agent or parallel
   decomposition using that rule's lifecycle and semantic capability
   profiles (` + "`architect`" + `, ` + "`orchestrator`" + `, ` + "`mapper`" + `,
   ` + "`specialist`" + `, ` + "`precision`" + `, ` + "`verifier`" + `).
3. Record the decision before the plan is finalized:
   - a multi-lane Agent Team Plan / Lane Manifest, using that rule's
     existing artifact; or
   - ` + "`single-lane, because <reason>`" + `, using that rule's existing
     single-lane criteria: trivial, tightly coupled, high-overlap, requires
     continuous design judgment, the user requested single-agent execution,
     or the active host does not confirm separate execution.

- This gate is mandatory even when the recorded answer is single-lane. A
  single mechanical edit, a direct question, or read-only research that
  never forms an implementation plan does not trigger it.
- This gate fires during native plan formation, before the plan is
  finalized. It precedes the Work Lane Mutation Hard Gate below, which fires
  later, before the first repository mutation.
- Never treat this evaluation as permission to force parallel execution on
  work that does not need it; a recorded single-lane decision remains a
  fully valid outcome.

---

`
