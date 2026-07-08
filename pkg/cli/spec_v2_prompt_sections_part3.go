package cli

import "github.com/jamesonstone/kit/internal/promptdoc"

func addSpecV2PromptValidationAndDelivery(doc *promptdoc.Document) {
	doc.Heading(2, "Validation And Verification Phase")
	doc.BulletList(
		"Validation is an explicit post-implementation phase inside this v2 `kit spec` workflow.",
		"Map validation 1:1 to Acceptance Criteria in `SPEC.md`; every criterion must have evidence or a documented blocker.",
		"Use the smallest relevant checks that prove the behavior: tests, linters, typechecks, build commands, runtime inspection, manual UI verification, docs review, or targeted scripts.",
		"Record concise evidence inline in `SPEC.md` and link detailed artifacts under `.kit/runs/...` or other stable local evidence locations when available.",
		"Use at least one read-only verification subagent by default after implementation unless the change is documentation-only, trivial, tightly coupled, `--single-agent` is active, or the active runtime cannot spawn subagents.",
		"When verification stays single-lane, record the exception and substitute review method in `SPEC.md`.",
		"Verification agents must compare `SPEC.md` against the actual diff, tests, runtime behavior, documentation updates, and evidence.",
		"Read-only verification lanes must not edit files, stage changes, mark criteria complete, or close their own findings.",
		"For each verifier gap, record `gap id -> acceptance criterion id -> Source Map id -> fix diff area -> rerun evidence -> verifier closure` in `SPEC.md`.",
		"After each verifier gap fix, rerun the relevant validation and update the Acceptance Coverage Audit before reflection or delivery.",
		"If validation is impossible, record reason, risk, substitute evidence, user-visible impact, owner or next action, and whether delivery is blocked.",
		"Skipped or impossible validation does not satisfy an acceptance criterion unless the substitute evidence proves the same behavior or the user accepts a scope change.",
		"If any acceptance criterion is unproven, partially implemented, contradicted by tests, or missing docs, the workflow returns to implementation before reflection or delivery.",
		"Never claim validation passed unless the checks actually ran or the evidence was directly inspected.",
	)

	doc.Heading(2, "Reflection Phase")
	doc.BulletList(
		"After validation passes, reflect on whether the final implementation still matches the thesis, requirements, implementation plan, and acceptance criteria in `SPEC.md`.",
		"Review for regressions, hidden scope creep, dead code, unused public surfaces, missing error handling, missing tests, and documentation drift.",
		"`kit loop workflow` writes runtime-owned `REFLECT.json` verdict evidence next to `SPEC.md` when the reflect stage completes; agents record human reflection notes in `SPEC.md` but must not fabricate or self-report verdict values.",
		"Record reflection notes, remaining risks, skipped validations, and any follow-up recommendations in `SPEC.md`.",
		"If reflection finds a correctness gap, route it back through implementation and validation before delivery.",
	)

	doc.Heading(2, "Zero-Error Completion Gate")
	doc.BulletList(
		"No known errors remain in the implementation, tests, runtime behavior, documentation, validation evidence, or delivery state.",
		"All acceptance criteria pass, or any blocked criterion has an explicit user-accepted scope change recorded in `SPEC.md`.",
		"All relevant tests, checks, reviews, and runtime validations pass, or each impossible validation has a recorded reason, impact, and replacement evidence.",
		"No verifier gap, reflection gap, documentation gap, or delivery-contract gap remains open.",
		"The actual diff, `SPEC.md`, documentation updates, validation evidence, and final response all describe the same completed work.",
		"Do not claim `complete` or zero-error status when any known defect, unproven criterion, stale doc, failed command, or unexplained skipped validation remains.",
	)

	doc.Heading(2, "Documentation Update Rules")
	doc.BulletList(
		"Update affected project documentation when behavior, commands, user workflows, configuration, public APIs, validation rules, or delivery instructions change.",
		"Keep `SPEC.md` current before and after code changes; it is the durable feature state.",
		"Keep `PROJECT_PROGRESS_SUMMARY.md`, `CONSTITUTION.md`, repo instruction docs, capabilities metadata, and relevant references current when this feature changes their contract.",
		"Do not collapse project-level docs into `SPEC.md`; project contracts remain project-level documents.",
		"Record every documentation update or explicit no-op decision in the Documentation Updates section of `SPEC.md`.",
	)

	doc.Heading(2, "Delivery Intent And Hard Gate")
	doc.BulletList(
		"Use the Delivery Decision section to record whether the user wants existing in-flight changes or a later issue/branch/PR lane.",
		"The pre-execution delivery decision is intent only. Do not create or mutate issues, branches, commits, pushes, PRs, review threads, or labels before implementation and validation are stable.",
		"Before any Git or GitHub mutation, load repo-local delivery rules and produce the required Delivery Contract.",
		"Stop if any delivery-contract field is unknown, ambiguous, missing, stale, or conflicts with repo-local rules.",
		"Never substitute generic global defaults for repo-local Kit delivery rules.",
	)

	doc.Heading(2, "SPEC.md Update Requirements")
	doc.BulletList(
		"Update Thesis with the original idea and any refined framing.",
		"Update Context with repo-grounded findings, relevant files, references, and constraints.",
		"Update Context `### Source Map` whenever a material claim is added, invalidated, confirmed, or mapped to acceptance, tasks, validation, or evidence.",
		"Update Clarifications after every user question batch and after repo research resolves an ambiguity.",
		"Update front matter `phase` immediately when a phase transition occurs.",
		"Update Requirements, Assumptions, and stable-ID Acceptance Criteria before implementation begins.",
		"Update Implementation Plan and Task Checklist before and during implementation.",
		"Update Validation Map and Evidence during validation.",
		"Update Evidence with exact command summaries, runtime review notes, verifier findings, and `.kit/runs/...` references when available.",
		"Update Reflection Notes after validation and after any follow-up fixes.",
		"Update Documentation Updates when docs are changed or intentionally left unchanged.",
		"Update Delivery Decision before delivery and after any delivery mutation succeeds or is blocked.",
		"Set phase to `blocked` when progress cannot continue without user input or external state. Set phase to `complete` only when acceptance, validation, reflection, documentation, and delivery decisions are fully represented.",
	)

	doc.Heading(2, "Response Scope")
	doc.BulletList(
		"Clarification-loop replies should use numbered questions with defaults, assumptions, uncertainty, confidence, and unresolved count; do not append the full final-response contract to each clarification batch.",
		"Use the Final Response Contract only when returning a phase result, blocker, delivery result, or completed workflow summary.",
	)
}
