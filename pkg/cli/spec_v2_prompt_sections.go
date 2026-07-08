package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func addSpecV2PromptContext(doc *promptdoc.Document, input specV2PromptInput, durableRows, instructionRows, supportingRows [][]string, useRLM bool) {
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
}

func addSpecV2PromptFoundations(doc *promptdoc.Document, input specV2PromptInput, cfg *config.Config, constitutionPath, projectProgressPath, rlmPath, rulesPath string) {
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
		"The initial `SPEC.md` may contain only the user's thesis/goal plus delivery intent. Treat every other section as work to infer, research, clarify, and validate during the clarification loop before implementation.",
		"Inside Context, maintain a `### Source Map` subsection for material repo facts, rules, API behavior, command behavior, user decisions, and existing patterns that affect implementation, validation, delivery, or user-visible behavior.",
		"Keep compact front matter current: `workflow_version: 2`, `phase`, `clarification.status`, `clarification.confidence`, `clarification.unresolved_questions`, `references`, `relationships`, `skills`, and delivery-related routing metadata when present.",
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
		"`kit spec` is prompt-producing by default. This prompt is the workflow contract; direct agent execution belongs to explicit loop/run surfaces such as `kit loop workflow`.",
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
}

func addSpecV2PromptClarification(doc *promptdoc.Document, input specV2PromptInput, goalPct int) {
	doc.Heading(2, "Clarification-First Operating Model")
	doc.BulletList(
		"Start in Clarification Mode unless `SPEC.md` front matter already has `clarification.status: ready`, `clarification.confidence` at or above the configured goal, and `clarification.unresolved_questions: 0`.",
		"Clarification Mode output is a pre-instruction report plus a numbered question batch with recommended defaults, assumptions, uncertainty, confidence, and unresolved count.",
		"After each clarification batch, update `SPEC.md` front matter clarification state and body sections before doing anything else.",
		"If unresolved questions remain, stop after the question batch. Do not append implementation instructions, do not start coding, and do not claim the workflow is ready.",
		"Execution Mode begins only after clarification state is ready, acceptance criteria are stable `AC-###` entries, validation maps those IDs, rollback and delivery are known, and dirty-worktree ownership is recorded.",
		"Keep the current conversation as live context after clarification completes. `SPEC.md` is durable state, but do not discard chat-derived decisions while continuing implementation in the same thread.",
		"If this prompt is running under `kit loop workflow`, do not guess user intent in the clarify stage. Research what is discoverable, update `SPEC.md`, and block with the exact questions if any user decision remains.",
	)

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
		"When present, load `docs/references/rules/agent-team-orchestration.md` before deciding whether implementation or verification should use subagents.",
		"Ask clarification questions instead of editing when any readiness, Source Map, dirty-worktree, ownership, or validation gate fails.",
	)

	doc.Heading(2, "Pre-Instruction Report")
	doc.Paragraph("Before asking clarification questions or implementing, output a short pre-instruction report. Persist durable decisions in `SPEC.md`; do not leave workflow state only in chat.")
	doc.BulletList(
		"current `SPEC.md` path, workflow version, and phase",
		"`clarification.status`, `clarification.confidence`, and `clarification.unresolved_questions` from front matter",
		"repo instruction docs, feature references, and skills loaded or intentionally omitted",
		"user-provided thesis, requirements, acceptance criteria, and delivery intent known so far",
		"confidence percentage, unresolved question count, and whether any readiness gate blocks implementation",
		"accepted, rejected, defaulted, and still-unverified assumptions",
		"initial acceptance criteria inventory and any criteria that are not yet binary-verifiable",
		"predicted touched files, packages, commands, and validation surfaces",
		"first validation strategy, rollback checkpoint, and evidence locations",
		"dirty-worktree status and ownership classification for existing changes",
		"Source Map status, including unverified or stale claims",
		"subagent team plan: implementation subagents, read-only verification subagents, logical-only lanes, omitted lanes, concurrency, and any single-lane exception",
	)

	doc.Heading(2, "Clarification Loop")
	doc.BulletList(
		fmt.Sprintf("Do not implement until confidence is at least %d%%, unresolved questions are 0, and every assumption is accepted, removed, or converted into a blocker.", goalPct),
		"Maintain front matter `clarification.status` as `open` while questions remain, `ready` only when implementation can begin without guessing, and `blocked` when progress needs user input or external state.",
		"Start by reading repo-local instructions and the current `SPEC.md`; then inspect only the smallest relevant repo context needed for the immediate decision.",
		"Use numbered clarification batches. Each question must include a recommended default, the assumption behind it, and the uncertainty it resolves.",
		"Accept `yes` or `y` as approval for all recommended defaults in the current batch. Accept `yes 3, 4` or `y 3, 4` for selected defaults. Accept `no 2: <answer>` or `n 2: <answer>` as an override.",
		"Before each clarification batch, state confidence, unresolved question count, and the gate that prevents implementation.",
		"If the user approves defaults, record the exact accepted defaults in `SPEC.md`; do not rely on chat-only approval.",
		"After each batch, update the Clarifications, Requirements, Assumptions, Acceptance Criteria, Delivery Decision, and front matter clarification state in `SPEC.md`.",
		"When the gate becomes ready, set `clarification.status: ready`, `clarification.confidence` to the current confidence percentage, and `clarification.unresolved_questions: 0` before moving phase to `ready`.",
		"Research ambiguity against the actual repository before asking the user when the answer is discoverable locally.",
		"Stop and ask when a decision materially changes scope, data shape, external behavior, delivery lane, rollback, or validation strategy and cannot be safely inferred.",
	)
}
