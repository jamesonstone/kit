package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

func addSpecV2PromptAgentTeam(doc *promptdoc.Document, agentTeamModeBullets []string, goalPct int) {
	doc.Heading(2, "Dirty Worktree And Ownership Gate")
	doc.BulletList(
		"Before implementation, run or inspect `git status --short` and record the summary in `SPEC.md`.",
		"Classify existing changes as user-owned, in-scope, unrelated, or unknown. Treat unknown as user-owned until proven otherwise.",
		"Do not overwrite, reformat, stage, commit, claim, or summarize user-owned or unrelated changes as your own work.",
		"If a file with existing changes must be edited, inspect the diff first, preserve user changes, and record why the file is safe to touch.",
		"If ownership or scope is unclear and the change is not safely separable, stop and ask before editing.",
	)

	doc.Heading(2, "Source Map Mechanics")
	doc.Paragraph("Keep the Source Map inside the Context section of `SPEC.md` as `### Source Map`. It tracks material claims, not every file touched.")
	doc.BulletList(
		"Use stable IDs such as `SRC-001`.",
		"Required columns: ID, Source, Selector, Claim / Fact, Used For, Maps To, Status.",
		"Sources may be exact files, headings, symbols, commands, docs, tests, APIs, design nodes, evidence artifacts, or explicit user decisions.",
		"Each claim must state the fact being relied on; a file path alone is not enough.",
		"Map every acceptance criterion to at least one Source Map entry or explicit user decision.",
		"Map every planned touched file to an acceptance criterion, Source Map fact, task, or validation requirement.",
		"Mark stale, inferred, or unverified claims explicitly instead of treating them as confirmed.",
		"If implementation contradicts a Source Map claim, stop, update `SPEC.md`, rerun the relevant readiness gate, and continue only after the contradiction is resolved.",
		"Verification agents must audit the Source Map against the actual diff, validation evidence, and documentation updates.",
	)

	doc.Heading(2, "Objective Readiness Gates")
	doc.BulletList(
		fmt.Sprintf("Clarification gate: `clarification.status: ready`, `clarification.confidence >= %d`, `clarification.unresolved_questions: 0`, accepted assumptions are explicit, rejected assumptions are removed, and all open risks have a mitigation or blocker.", goalPct),
		"Source Map gate: material claims have stable `SRC-###` IDs, exact sources, claim text, status, and mappings to acceptance criteria, tasks, touched files, or validation.",
		"Acceptance gate: every acceptance criterion has a stable `AC-###` ID, is binary-verifiable, has an owner or lane, traces to a Source Map entry or user decision, and maps to at least one validation method.",
		"Planning gate: implementation approach, touched files, rollback strategy, sequencing, and lane overlap risks are recorded in `SPEC.md`.",
		"Dirty-worktree gate: existing changes are inspected and classified before any implementation edit.",
		"Task gate: `SPEC.md` contains a concise task checklist with task status, lane, acceptance mapping, and expected evidence.",
		"Delivery gate: delivery intent is recorded before execution, but no Git or GitHub mutation occurs until after validation and the repo-local delivery hard gate pass.",
		"Implementation may begin only after all gates pass. If any gate fails, update `SPEC.md`, ask the needed questions, or report the blocker.",
	)

	doc.Heading(2, "Acceptance Criteria Discipline")
	doc.BulletList(
		"Use stable acceptance criterion IDs such as `AC-001`, `AC-002`, and `AC-003`.",
		"Each acceptance criterion must be binary-verifiable and must map to Source Map entries, tasks, validation evidence, and verifier status.",
		"After readiness, do not silently split, merge, renumber, remove, or reword acceptance criteria.",
		"If an acceptance criterion changes after readiness, update the Source Map, Validation Map, Task Checklist, reflection notes, and acceptance history in `SPEC.md`; ask the user when the change affects scope or behavior.",
	)

	doc.Heading(2, "Phase Transition Rules")
	doc.BulletList(
		"`clarify`: default phase while questions, assumptions, acceptance criteria, validation mapping, delivery intent, or rollback are unresolved.",
		"`ready`: set only after every readiness gate passes, the Agent Team Plan is recorded, and implementation has not started.",
		"`implement`: set when implementation begins and the task checklist is actively changing.",
		"`validate`: set when implementation tasks are complete enough to gather evidence against every acceptance criterion.",
		"`reflect`: set only after validation passes with no unclosed acceptance, verifier, documentation, or evidence gaps.",
		"`deliver`: set when reflection and documentation sync are complete and the only remaining work is the delivery hard gate or delivery mutation.",
		"`complete`: set only when acceptance criteria, validation evidence, reflection notes, documentation state, and delivery decision are fully represented with no known errors remaining.",
		"`blocked`: set when the workflow cannot progress without user input or external state; record the blocker, attempts made, owner, and next needed input.",
		"Do not skip phases. If an exceptional case compresses phases, record the reason and the prior gate evidence in `SPEC.md` before continuing.",
	)

	doc.Heading(2, "Agent Team Model")
	agentTeamBullets := []string{
		"The supervisor agent owns `SPEC.md`, clarification, scope, acceptance criteria, lane assignment, integration, validation synthesis, delivery gating, and final response.",
		"When present, `docs/references/rules/agent-team-orchestration.md` is the durable rule for deciding whether work uses specialist subagents, read-only verification agents, or a recorded single supervisor lane.",
	}
	agentTeamBullets = append(agentTeamBullets, agentTeamModeBullets...)
	agentTeamBullets = append(agentTeamBullets,
		"Use dynamic lanes with a fixed supervisor contract.",
		"Create specialist lanes only when work separates into low-overlap files, packages, services, UI/backend areas, docs, tests, or validation surfaces.",
		"Treat planned lanes as logical work routing until separate agents are actually spawned.",
		"If no subagents or verification agents actually ran, report exactly: `single supervisor lane; no specialist or verification agents spawned`.",
		"Do not describe logical lanes as agents, spawned lanes, or verification agents unless separate agents actually ran.",
		"Default max concurrent lanes: 3.",
		"Hard ceiling: 4, only when predicted file overlap is clearly low.",
		"Do not use \"as many agents as possible.\"",
		"Verification lanes are read-only by default.",
		"Verification agents review `SPEC.md`, especially Acceptance Criteria, against the diff, tests, runtime behavior, documentation updates, and evidence.",
		"Verification agents record gaps; the supervisor routes fixes back to implementation lanes.",
		"The supervisor must update `SPEC.md` after each phase: clarified decisions, task status, validation evidence, reflection notes, and delivery state.",
	)
	doc.BulletList(agentTeamBullets...)

	doc.Heading(2, "Agent Team Plan")
	doc.Paragraph("Before implementation, output an Agent Team Plan and persist the durable parts in `SPEC.md`. In default mode, the plan must identify the subagents that will actually be spawned unless a single-lane exception applies.")
	doc.BulletList(
		"supervisor responsibilities",
		"proposed lanes",
		"implementation lanes that will actually be spawned as subagents",
		"read-only verification lanes that will actually be spawned as subagents",
		"logical-only lanes that will not be spawned",
		"intentionally omitted lanes",
		"reason for each omitted implementation or verification subagent",
		"predicted touched files per lane",
		"overlap risks",
		"max concurrency",
		"serialized work",
		"validation/review lanes",
	)
}

func addSpecV2PromptExecution(doc *promptdoc.Document) {
	doc.Heading(2, "Implementation Rules")
	doc.BulletList(
		"Implement only after the readiness gates pass.",
		"Execute the task checklist in `SPEC.md`; update task status as work moves from pending to in progress to complete or blocked.",
		"Prefer existing repo patterns, explicit code, minimal production-ready changes, and narrow diffs.",
		"Do not invent abstractions, new public surfaces, or broad refactors unless an acceptance criterion requires them.",
		"Inspect files before editing. Never guess file contents, APIs, flags, tests, or command behavior.",
		"Do not edit files that have unclassified existing changes.",
		"Every touched file must map to an acceptance criterion, task, Source Map fact, or validation requirement.",
		"Keep the supervisor responsible for integrating lane work into one coherent diff and resolving conflicts.",
		"After each implementation lane completes, update `SPEC.md` with changed files, task status, risks discovered, rollback notes, and validation still needed.",
		"If implementation reveals a requirement gap, stop implementation for that area, update `SPEC.md`, run the clarification/readiness gate again, then continue.",
	)

	doc.Heading(2, "Acceptance Coverage Audit")
	doc.Paragraph("Before reflection, delivery, or the final response, update a coverage table in `SPEC.md` that proves every acceptance criterion has been addressed.")
	doc.BulletList(
		"Each acceptance criterion row must include criterion id, implementation evidence, validation command or review evidence, documentation impact, verifier result, and gap status.",
		"Every row status must be pass, fail, or blocked. `Not applicable` requires a recorded rationale and supervisor approval.",
		"Any failed, blocked, partially proven, or verifier-disputed row routes back to implementation or clarification before reflection or delivery.",
		"The final response must summarize this audit instead of making a general claim that acceptance criteria are complete.",
	)

	addSpecV2PromptValidationAndDelivery(doc)
}
