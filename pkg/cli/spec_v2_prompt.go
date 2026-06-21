package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

type specV2PromptInput struct {
	SpecPath       string
	BrainstormPath string
	FeatureSlug    string
	ProjectRoot    string
	Config         *config.Config
	Answers        *specAnswers
	PromptOnly     bool
}

func buildSpecV2SupervisorPrompt(input specV2PromptInput) string {
	cfg := input.Config
	if cfg == nil {
		cfg = config.Default()
	}

	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(input.ProjectRoot, "docs", "CONSTITUTION.md")
	projectProgressPath := filepath.Join(input.ProjectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	rlmPath := filepath.Join(input.ProjectRoot, "docs", "agents", "RLM.md")
	rulesPath := filepath.Join(input.ProjectRoot, "docs", "references", "rules")
	featureDir := filepath.Base(filepath.Dir(input.SpecPath))
	notesPath := featureNotesPath(input.ProjectRoot, featureDir)
	designPath := featureDesignMaterialsPath(input.ProjectRoot, featureDir)
	hasBrainstorm := document.Exists(input.BrainstormPath)
	useRLM := specNeedsRLM(input.FeatureSlug, input.SpecPath, input.BrainstormPath, input.Answers)

	durableRows := [][]string{
		{"SPEC", fmt.Sprintf("%s - single durable feature artifact and workflow state", input.SpecPath)},
		{"CONSTITUTION", fmt.Sprintf("%s - durable repository facts: project-wide constraints, invariants, and development contract", constitutionPath)},
		{"PROJECT PROGRESS", fmt.Sprintf("%s - durable repository facts: highest completed artifact and prior-feature index", projectProgressPath)},
		{"KIT MANAGED RULESETS", fmt.Sprintf("%s - pointer-loaded durable repo-local rulesets managed by Kit", rulesPath)},
	}
	instructionRows := [][]string{
		{"RLM", fmt.Sprintf("%s - Kit's just-in-time context-routing pattern for progressive disclosure", rlmPath)},
	}
	instructionTargets := map[string]struct{}{
		filepath.Clean(rlmPath): struct{}{},
	}
	for _, row := range repoInstructionContextRows(input.ProjectRoot, cfg) {
		if len(row) < 2 {
			continue
		}
		cleaned := filepath.Clean(row[1])
		if _, exists := instructionTargets[cleaned]; exists {
			continue
		}
		instructionTargets[cleaned] = struct{}{}
		instructionRows = append(instructionRows, []string{
			row[0],
			fmt.Sprintf("%s - repo-local agent routing and safety guidance", row[1]),
		})
	}
	supportingRows := [][]string{
		{"FEATURE NOTES", fmt.Sprintf("%s - optional reference material supplied before or during spec work", notesPath)},
		{"DESIGN MATERIALS", fmt.Sprintf("%s - optional screenshots, references, and design inputs when relevant", designPath)},
		{"PROJECT ROOT", input.ProjectRoot},
	}
	if hasBrainstorm {
		supportingRows = append(supportingRows, []string{
			"LEGACY BRAINSTORM",
			fmt.Sprintf("%s - historical v1 research context; carry forward only still-relevant facts into SPEC.md", input.BrainstormPath),
		})
	}
	for _, row := range specSkillDiscoveryContextRows(input.ProjectRoot, cfg) {
		if len(row) < 2 {
			continue
		}
		switch row[0] {
		case "Repo Agents Entry", "Repo References Entry":
			continue
		}
		supportingRows = append(supportingRows, row)
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf(
			"You are the supervisor coding agent for the Kit v2 `kit spec` workflow for feature `%s`.",
			input.FeatureSlug,
		))
		doc.Paragraph(fmt.Sprintf(
			"`SPEC.md` is the single durable feature artifact for this workflow. You MUST keep `%s` current as you clarify, plan, implement, validate, reflect, update docs, and prepare delivery. Do not leave durable decisions only in chat.",
			input.SpecPath,
		))

		doc.Heading(2, "Context Provided By User")
		doc.Raw(specV2UserContext(input.Answers))

		doc.Heading(2, "Durable Repository Facts")
		doc.Table([]string{"Input", "Purpose"}, durableRows)

		doc.Heading(2, "Instruction Entrypoints")
		doc.Table([]string{"Input", "Purpose"}, instructionRows)

		doc.Heading(2, "Supporting Inputs")
		doc.Table([]string{"Input", "Purpose"}, supportingRows)
		if useRLM {
			doc.Heading(1, "Use RLM Pattern")
			doc.BulletList(strings.Split(rlmSpecGuidanceStepText(input.SpecPath), "\n")...)
		}

		doc.Heading(2, "Source Of Truth Precedence")
		doc.BulletList(
			"Safety, permission, and system constraints override all workflow instructions.",
			"The current user request overrides stale repository docs when the conflict is explicit and safe.",
			fmt.Sprintf("`%s` controls project invariants, durable development rules, and constraints.", constitutionPath),
			fmt.Sprintf("`%s` controls this feature's requirements, plan, task checklist, validation map, reflection, delivery decision, and evidence.", input.SpecPath),
			"Referenced repo rules, repo docs, skills, APIs, and Source Map entries constrain the specific decisions they cover.",
			"Historical v1 artifacts are non-binding context unless the user explicitly chooses a legacy staged command.",
			"Repo conventions fill gaps only when they do not conflict with higher-priority sources.",
			"When sources conflict, stop, record the conflict in `SPEC.md`, and ask only if repo research cannot resolve it safely.",
		)

		doc.Heading(2, "Repo Routing, References, And Skills")
		doc.BulletList(
			repoInstructionReadStepText(input.ProjectRoot, cfg),
			relatedFeatureContextStepText(input.ProjectRoot, input.SpecPath),
			"Use repo-local docs and canonical skills before secondary global inputs. Load only the specific docs or skills that materially affect the current decision.",
			fmt.Sprintf("Treat `%s` and `%s` as durable repository facts: use the Constitution for project invariants and the progress summary as the current feature index before relying on stale chat context.", constitutionPath, projectProgressPath),
			fmt.Sprintf("Use `%s` for RLM routing when the context set is broad; keep must-read inputs small and record material references instead of copying broad history into `SPEC.md`.", rlmPath),
			fmt.Sprintf("Treat Kit-managed rulesets under `%s` as pointer-loaded durable rules. Load only rulesets whose references, read policy, applies-to text, or current delivery/command decision make them relevant.", rulesPath),
			fmt.Sprintf("Keep front matter `references` in `%s` current for repo docs, skills, MCP tools, APIs, design inputs, datasets, assets, validation evidence, and other resources that shape the feature.", input.SpecPath),
			"Each material reference must include exact target, relation, read policy, used-for text, and status. Prefer stable selectors over fragile line references.",
			"Keep front matter `relationships` explicit when this feature builds on, depends on, or relates to prior feature directories.",
			"Keep front matter `skills` focused on execution-time skills the agent should actually use. If no special skill applies, leave the list empty.",
			"Do not use `.claude/skills` as canonical discovery input.",
		)

		doc.Heading(2, "SPEC.md Contract")
		doc.BulletList(
			fmt.Sprintf("`%s` is the only durable feature workflow artifact for v2 feature work.", input.SpecPath),
			"Use this fixed section order: Thesis, Context, Clarifications, Requirements, Assumptions, Acceptance Criteria, Implementation Plan, Task Checklist, Validation Map, Reflection Notes, Documentation Updates, Delivery Decision, Evidence.",
			"Inside Context, maintain a `### Source Map` subsection for material repo facts, rules, API behavior, command behavior, user decisions, and existing patterns that affect implementation, validation, delivery, or user-visible behavior.",
			"Keep compact front matter current: `workflow_version: 2`, `phase`, `references`, `relationships`, `skills`, and delivery-related routing metadata when present.",
			"Valid phases are `clarify`, `ready`, `implement`, `validate`, `reflect`, `deliver`, `complete`, and `blocked`.",
			"Use front matter for cheap routing metadata and the Markdown body for human-readable decisions, tasks, validation, reflection, and evidence.",
			"Historical v1 artifacts such as `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` may exist in upgraded projects. Treat them as read-only historical context unless the user explicitly asks to work through a legacy staged command surface.",
			"Do not create, rewrite, delete, or move v1 artifacts as part of the v2 workflow. Carry forward still-relevant facts into `SPEC.md` instead.",
		)

		doc.Heading(2, "Supervisor Responsibilities")
		doc.BulletList(
			"Own `SPEC.md` from start to finish and keep it synchronized after every phase.",
			"Ground clarification in repo evidence before implementation begins.",
			"Resolve scope, assumptions, acceptance criteria, delivery intent, rollback strategy, and validation mapping before implementation.",
			"Create and maintain the implementation plan and concise task checklist inside `SPEC.md`.",
			"Assign dynamic lanes only when the work naturally separates with low file or interface overlap.",
			"Integrate lane output, resolve conflicts, and keep the final diff coherent.",
			"Run or coordinate validation, synthesize read-only verifier findings, route gaps back to implementation, and update evidence.",
			"Enforce the delivery hard gate before any issue, branch, commit, push, PR, or review-thread mutation.",
			"Produce the final response only after acceptance criteria, validation evidence, reflection, documentation state, and delivery state are represented in `SPEC.md`.",
		)

		doc.Heading(2, "Prompt-Only And V1 Compatibility")
		promptOnlyBullets := []string{
			"`kit spec` is prompt-producing by default. This prompt is the workflow contract; Kit itself has not directly invoked agents unless an explicit future `--run` or `--loop` mode is used.",
			"If this prompt came from `kit spec --prompt-only`, do not assume Kit made any adoption writes. Inspect current files, then update `SPEC.md` only when the user has asked you to execute the workflow.",
			"Existing Kit-owned projects may contain v1 files. Preserve them non-disruptively and use them only as optional evidence.",
			"If a user asks to finish a v1 flow through old artifact stages, tell them to use the explicit legacy surface instead of reintroducing v1 stage assumptions into v2 `kit spec`.",
			"Do not tell the user that the next normal step is `kit legacy brainstorm`, `kit legacy plan`, `kit legacy tasks`, `kit legacy implement`, `kit legacy reflect`, or a standalone verification command. In v2, the normal feature workflow remains inside `kit spec` and `SPEC.md`.",
		}
		if input.PromptOnly {
			promptOnlyBullets = append([]string{
				"This prompt was generated by `kit spec --prompt-only`. Treat it as inspection-safe output: Kit did not add v2 metadata, missing v2 sections, feature notes directories, rollup updates, or other adoption writes for this prompt.",
			}, promptOnlyBullets...)
		}
		doc.BulletList(promptOnlyBullets...)

		doc.Heading(2, "First-Action Checklist")
		doc.BulletList(
			fmt.Sprintf("Read `%s` and identify current front matter, phase, acceptance criteria, validation map, delivery decision, and evidence state.", input.SpecPath),
			"Read repository instruction entrypoints and route through only the relevant repo-local docs.",
			"Run or inspect `git status --short` before implementation and record the ownership classification in `SPEC.md`.",
			"Classify existing changes as user-owned, in-scope, unrelated, or unknown before editing any touched file.",
			"Identify predicted touched files, packages, commands, validation surfaces, and rollback checkpoint.",
			"Inventory acceptance criteria and convert them to stable IDs such as `AC-001` before implementation.",
			"Build or update the `### Source Map` for every material claim that could change implementation, validation, delivery, or user-visible behavior if wrong.",
			"List expected validation commands and any validations that may need environment, credentials, services, fixtures, or manual review.",
			"Ask clarification questions instead of editing when any readiness, Source Map, dirty-worktree, ownership, or validation gate fails.",
		)

		doc.Heading(2, "Pre-Instruction Report")
		doc.Paragraph("Before asking clarification questions or implementing, output a short pre-instruction report. Persist durable decisions in `SPEC.md`; do not leave workflow state only in chat.")
		doc.BulletList(
			"current `SPEC.md` path, workflow version, and phase",
			"repo instruction docs, feature references, and skills loaded or intentionally omitted",
			"user-provided thesis, requirements, acceptance criteria, and delivery intent known so far",
			"confidence percentage, unresolved question count, and whether any readiness gate blocks implementation",
			"accepted, rejected, defaulted, and still-unverified assumptions",
			"initial acceptance criteria inventory and any criteria that are not yet binary-verifiable",
			"predicted touched files, packages, commands, and validation surfaces",
			"first validation strategy, rollback checkpoint, and evidence locations",
			"dirty-worktree status and ownership classification for existing changes",
			"Source Map status, including unverified or stale claims",
			"whether the work should stay single-lane or requires a later Agent Team Plan",
		)

		doc.Heading(2, "Clarification Loop")
		doc.BulletList(
			fmt.Sprintf("Do not implement until confidence is at least %d%%, unresolved questions are 0, and every assumption is accepted, removed, or converted into a blocker.", goalPct),
			"Start by reading repo-local instructions and the current `SPEC.md`; then inspect only the smallest relevant repo context needed for the immediate decision.",
			"Use numbered clarification batches. Each question must include a recommended default, the assumption behind it, and the uncertainty it resolves.",
			"Accept `yes` or `y` as approval for all recommended defaults in the current batch. Accept `yes 3, 4` or `y 3, 4` for selected defaults. Accept `no 2: <answer>` or `n 2: <answer>` as an override.",
			"Before each clarification batch, state confidence, unresolved question count, and the gate that prevents implementation.",
			"If the user approves defaults, record the exact accepted defaults in `SPEC.md`; do not rely on chat-only approval.",
			"After each batch, update the Clarifications, Requirements, Assumptions, Acceptance Criteria, Delivery Decision, and confidence state in `SPEC.md`.",
			"Research ambiguity against the actual repository before asking the user when the answer is discoverable locally.",
			"Stop and ask when a decision materially changes scope, data shape, external behavior, delivery lane, rollback, or validation strategy and cannot be safely inferred.",
		)

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
			"Clarification gate: unresolved questions = 0; accepted assumptions are explicit; rejected assumptions are removed; all open risks have a mitigation or blocker.",
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
		doc.BulletList(
			"The supervisor agent owns `SPEC.md`, clarification, scope, acceptance criteria, lane assignment, integration, validation synthesis, delivery gating, and final response.",
			"Use dynamic lanes with a fixed supervisor contract.",
			"Create specialist lanes only when work separates into low-overlap files, packages, services, UI/backend areas, docs, tests, or validation surfaces.",
			"Default max concurrent lanes: 3.",
			"Hard ceiling: 4, only when predicted file overlap is clearly low.",
			"Do not use \"as many agents as possible.\"",
			"Verification lanes are read-only by default.",
			"Verification agents review `SPEC.md`, especially Acceptance Criteria, against the diff, tests, runtime behavior, documentation updates, and evidence.",
			"Verification agents record gaps; the supervisor routes fixes back to implementation lanes.",
			"The supervisor must update `SPEC.md` after each phase: clarified decisions, task status, validation evidence, reflection notes, and delivery state.",
		)

		doc.Heading(2, "Agent Team Plan")
		doc.Paragraph("Before implementation, output an Agent Team Plan and persist the durable parts in `SPEC.md`.")
		doc.BulletList(
			"supervisor responsibilities",
			"proposed lanes",
			"intentionally omitted lanes",
			"predicted touched files per lane",
			"overlap risks",
			"max concurrency",
			"serialized work",
			"validation/review lanes",
		)

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

		doc.Heading(2, "Validation And Verification Phase")
		doc.BulletList(
			"Validation is an explicit post-implementation phase inside this v2 `kit spec` workflow.",
			"Map validation 1:1 to Acceptance Criteria in `SPEC.md`; every criterion must have evidence or a documented blocker.",
			"Use the smallest relevant checks that prove the behavior: tests, linters, typechecks, build commands, runtime inspection, manual UI verification, docs review, or targeted scripts.",
			"Record concise evidence inline in `SPEC.md` and link detailed artifacts under `.kit/runs/...` or other stable local evidence locations when available.",
			"Assign one or more read-only verification lanes when the change has enough surface area to justify independent review.",
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

		addFinalResponseContract(doc, specV2FinalResponseContract()...)
	})
}

func specV2UserContext(answers *specAnswers) string {
	if answers == nil || (answers.Problem == "" && answers.Goals == "" && answers.NonGoals == "" &&
		answers.Users == "" && answers.Requirements == "" && answers.Acceptance == "" &&
		answers.EdgeCases == "" && answers.DeliveryIntent == "") {
		return strings.TrimSpace(`<!-- Fill this section before sending the prompt when useful. Leave blanks only if the agent should derive the content from existing SPEC.md and repo context. -->

**THESIS**:
<!-- Original idea, problem statement, or feature request. -->

**CONTEXT**:
<!-- Known repo, product, user, design, or operational context. -->

**REQUIREMENTS**:
<!-- Initial requirements or constraints. -->

**ACCEPTANCE CRITERIA**:
<!-- Initial binary-verifiable acceptance criteria. -->

**DELIVERY INTENT**:
<!-- Existing in-flight changes, or later issue/branch/PR after validation. -->

**NON-GOALS / EXCLUSIONS**:
<!-- What should not be changed. -->`)
	}

	var items []string
	appendAnswer := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		items = append(items, fmt.Sprintf("**%s**: %s", label, value))
	}
	appendAnswer("THESIS / PROBLEM", answers.Problem)
	appendAnswer("GOALS", answers.Goals)
	appendAnswer("NON-GOALS", answers.NonGoals)
	appendAnswer("USERS", answers.Users)
	appendAnswer("REQUIREMENTS", answers.Requirements)
	appendAnswer("ACCEPTANCE CRITERIA", answers.Acceptance)
	appendAnswer("EDGE CASES", answers.EdgeCases)
	if strings.TrimSpace(answers.DeliveryIntent) == "" {
		items = append(items, "**DELIVERY INTENT**: clarify before implementation; record existing in-flight changes or later issue/branch/PR intent in SPEC.md before execution")
	} else {
		appendAnswer("DELIVERY INTENT", answers.DeliveryIntent)
	}
	return strings.Join(items, "\n\n")
}

func specV2FinalResponseContract() []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Summary",
			Items:   []string{"State what changed and whether the workflow phase is complete, blocked, or ready for delivery."},
		},
		{
			Heading: "SPEC.md State",
			Items:   []string{"Report the current phase, confidence, unresolved questions count, and the sections materially updated."},
		},
		{
			Heading: "Acceptance Coverage",
			Items:   []string{"Map stable acceptance criteria IDs to Source Map IDs, implementation evidence, validation evidence, and verifier status, or state the blocker for any gap."},
		},
		{
			Heading: "Validation Evidence",
			Items:   []string{"List exact commands, checks, runtime reviews, documentation reviews, evidence artifact links, and validation-impossible rubrics used."},
		},
		{
			Heading: "Zero-Error Gate",
			Items:   []string{"State whether no known errors remain across implementation, validation, verification, reflection, documentation, and delivery state; if not, mark the workflow blocked and list the exact remaining errors."},
		},
		{
			Heading: "Agent Team",
			Items:   []string{"Summarize lanes used, lanes intentionally omitted, verification lanes, concurrency, and overlap decisions."},
		},
		{
			Heading: "Delivery",
			Items:   []string{"State delivery intent, delivery hard-gate status, and any issue/branch/PR actions taken or still blocked."},
		},
		{
			Heading: "Open Items",
			Items:   []string{"List remaining blockers, skipped validation, follow-ups, or write `none`."},
		},
	}
}
