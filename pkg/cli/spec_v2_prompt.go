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
	SingleAgent    bool
}

func buildSpecV2SupervisorPrompt(input specV2PromptInput) string {
	cfg := input.Config
	if cfg == nil {
		cfg = config.Default()
	}

	constitutionPath := filepath.Join(input.ProjectRoot, "docs", "CONSTITUTION.md")
	agentDocsPath := filepath.Join(input.ProjectRoot, "docs", "agents", "README.md")
	rulesPath := filepath.Join(input.ProjectRoot, "docs", "references", "rules")
	progressPath := filepath.Join(input.ProjectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	featureDir := filepath.Base(filepath.Dir(input.SpecPath))
	notesPath := featureNotesPath(input.ProjectRoot, featureDir)

	contextRows := [][]string{
		{"Canonical feature state", input.SpecPath},
		{"Repository constraints", constitutionPath},
		{"Agent routing", agentDocsPath},
		{"Conditional rules", rulesPath},
		{"Project index", progressPath},
		{"Project skills", filepath.Join(input.ProjectRoot, cfg.SkillsDir, "*", "SKILL.md")},
		{"Optional feature notes", notesPath},
	}
	if document.Exists(input.BrainstormPath) {
		contextRows = append(contextRows, []string{"Legacy research context", input.BrainstormPath})
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Supervise feature `%s` through its deprecated V2 compatibility workflow phase.", input.FeatureSlug))
		doc.Paragraph(fmt.Sprintf("`SPEC.md` is the single durable feature artifact. Keep `%s` current, complete the requested scope, and block only on a material decision that repository evidence cannot resolve safely.", input.SpecPath))

		doc.Heading(2, "Goal")
		doc.BulletList(
			"Turn the user's goal into a correct, minimal, production-ready result.",
			"Keep requirements, decisions, tasks, validation, reflection, documentation, delivery state, and evidence synchronized in `SPEC.md`.",
			"Continue through the requested scope without treating an individual task or phase as a stopping point.",
		)

		doc.Heading(2, "User Context")
		doc.Raw(specV2UserContext(input.Answers))

		doc.Heading(2, "Repository Context")
		doc.Table([]string{"Input", "Path"}, contextRows)
		if specNeedsRLM(input.FeatureSlug, input.SpecPath, input.BrainstormPath, input.Answers) {
			doc.Heading(2, "Context Routing")
			doc.Paragraph("The context is broad: use the repository RLM pattern to load the smallest source that resolves the current decision, then stop loading context.")
		}

		doc.Heading(2, "Source And State Contract")
		doc.BulletList(
			"Priority is system and safety constraints, the current user request, repository invariants, this feature's `SPEC.md`, then relevant referenced rules/docs/skills and established repo conventions.",
			"Read repository instruction entrypoints as routing maps. Load only sources relevant to the current decision, and record material sources in front matter `references` or the Context `### Source Map`.",
			"Use stable `SRC-###`, `REQ-###`, `AC-###`, and task IDs where traceability matters. Map every acceptance criterion to implementation and validation evidence.",
			"Keep front matter current: `workflow_version: 2`, `phase`, clarification state, feature identity, relationships, references, skills, and delivery intent.",
			"Keep the body sections in this order: Thesis, Context, Clarifications, Requirements, Assumptions, Acceptance Criteria, Implementation Plan, Task Checklist, Validation Map, Reflection Notes, Documentation Updates, Delivery Decision, Evidence.",
			"Treat legacy `BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` as read-only history unless the user explicitly requested a legacy staged workflow.",
		)

		doc.Heading(2, "Clarification And Autonomy")
		doc.BulletList(
			"During `clarify`, research repository-discoverable facts first. Ask only about material choices that remain non-discoverable; use numbered questions with a recommended default and why the answer changes the result. Keep `clarification.status: open` only while one or more such questions remain.",
			"If questions remain in an explicit clarification turn, update `SPEC.md`, output an `Open Questions` section, and stop before implementation. When none remain, state `Open Questions: none` and move the durable state to `ready`.",
			"Record residual uncertainty that does not require a user decision as an assumption or named risk. Confidence is a reporting signal and does not determine `clarification.status`.",
			"Outside `clarify`, do not re-ask settled questions or request routine permission for safe discovery and in-scope work. Proceed using the user request, `SPEC.md`, and repository evidence.",
			"Ask before a material scope/behavior change, an irreversible or production action, a required external mutation not already authorized, or a choice whose alternatives would produce meaningfully different outcomes.",
			"When implementation exposes a real requirement conflict, record it and return only the affected work to clarification; do not reopen unrelated decisions.",
		)

		doc.Heading(2, "Constraints And Approval Boundaries")
		constraints := []string{
			"Inspect `git status --short` before edits. Preserve user-owned and unrelated changes; inspect an existing diff before touching the same file.",
			"Prefer explicit existing patterns and the smallest coherent diff. Do not invent public surfaces, abstractions, APIs, or broad refactors without an acceptance criterion.",
			"Safe repository reads and reversible in-scope edits need no extra approval. Respect any stricter user, system, sandbox, or repo-local boundary.",
			"Do not create or mutate issues, branches, staging, commits, pushes, PRs, review threads, labels, production systems, or other external state until authorized and the applicable repo-local gate is satisfied.",
			"Before Git/GitHub delivery mutation, load the relevant rules under `docs/references/rules`, establish the Delivery Contract, and stop if any required field is unknown or conflicting.",
		}
		if input.PromptOnly {
			constraints = append(constraints, "This was generated with `--prompt-only`; Kit made no adoption or document writes. Inspect current files before deciding whether the user authorized workflow execution.")
		}
		doc.BulletList(constraints...)

		doc.Heading(2, "Phase Outcomes")
		doc.Table([]string{"Phase", "Required outcome"}, [][]string{
			{"clarify", "Ground scope and decisions; binary acceptance criteria; 1:1 validation map; delivery and rollback known."},
			{"ready", "Readiness gates pass; implementation plan, task checklist, predicted files, ownership, and routing are recorded."},
			{"implement", "Execute the checklist, update focused tests/docs, and keep task/evidence state current."},
			{"validate", "Run mapped checks, close verifier findings, and prove every acceptance criterion or record a genuine blocker."},
			{"reflect", "Audit the integrated diff for correctness, regressions, scope, documentation, and remaining risk; route gaps back."},
			{"deliver", "Run the repository delivery contract and record exact issue/branch/commit/PR/check state."},
			{"complete", "Implementation, acceptance, validation, reflection, docs, evidence, and delivery decision agree with no known gap."},
			{"blocked", "Record the blocker, attempts, owner, impact, and exact input or external change needed."},
		})
		doc.Paragraph("Do not skip a phase gate. The Kit loop validates phase state in code; prompt prose does not replace those checks.")

		doc.Heading(2, "Agent Routing")
		agentBullets := []string{
			"The supervisor owns `SPEC.md`, scope, integration, acceptance, validation synthesis, delivery gating, and the final report.",
			"Use specialist agents only for independent low-overlap work. Predict files/interfaces first, serialize shared surfaces, and keep verification read-only.",
		}
		agentBullets = append(agentBullets, specV2AgentTeamModeBullets(input.SingleAgent)...)
		doc.BulletList(agentBullets...)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			"Every in-scope acceptance criterion has implementation evidence, exact validation evidence, documentation disposition, and verifier status.",
			"Relevant tests, lint/typecheck/build/runtime checks, generated docs, and manual review pass; skipped checks state reason, risk, substitute evidence, and delivery impact.",
			"Reflection finds no unresolved correctness, regression, scope, dead-code, security, documentation, or delivery-contract gap.",
			"The final diff, `SPEC.md`, evidence, docs, and user-facing summary describe the same result. Never claim a check ran when it did not.",
		)

		doc.Heading(2, "Output Contract")
		doc.BulletList(
			"During clarification, report current confidence and unresolved count, then `Open Questions`; do not append an implementation summary while blocked on answers.",
			"During execution, keep commentary concise and put durable decisions/evidence in `SPEC.md` rather than repeating the workflow contract in chat.",
			"At a phase result or final handoff, report outcome first, exact validation, changed artifacts, current phase/delivery state, agent use, and only genuine open items.",
		)
		addFinalResponseContract(doc, specV2FinalResponseContract()...)
	})
}

func specV2UserContext(answers *specAnswers) string {
	if answers == nil || (answers.Problem == "" && answers.Goals == "" && answers.NonGoals == "" &&
		answers.Users == "" && answers.Requirements == "" && answers.Acceptance == "" &&
		answers.EdgeCases == "" && answers.DeliveryIntent == "") {
		return strings.TrimSpace(`**GOAL**: Read the current SPEC.md Thesis and the user's current request.

**CONTEXT**: Ground missing facts in the repository and Source Map.

**CONSTRAINTS**: Read repository invariants, relevant rules, and explicit non-goals.

**SUCCESS**: Use the stable acceptance criteria and validation map in SPEC.md.

**DELIVERY**: Read front matter delivery intent and the Delivery Decision; ask only if a material delivery choice remains non-discoverable.`)
	}

	var items []string
	appendAnswer := func(label, value string) {
		if strings.TrimSpace(value) != "" {
			items = append(items, fmt.Sprintf("**%s**: %s", label, value))
		}
	}
	appendAnswer("THESIS", answers.Problem)
	appendAnswer("GOALS", answers.Goals)
	appendAnswer("NON-GOALS", answers.NonGoals)
	appendAnswer("USERS", answers.Users)
	appendAnswer("REQUIREMENTS", answers.Requirements)
	appendAnswer("ACCEPTANCE", answers.Acceptance)
	appendAnswer("EDGE CASES", answers.EdgeCases)
	appendAnswer("DELIVERY INTENT", answers.DeliveryIntent)
	return strings.Join(items, "\n\n")
}

func specV2AgentTeamModeBullets(singleAgentMode bool) []string {
	if singleAgentMode {
		return []string{
			"`--single-agent` is active: keep execution and verification in one supervisor lane, but record logical sequencing when useful.",
			"Report: `single supervisor lane; no specialist or verification agents spawned`.",
		}
	}
	return []string{
		"For nontrivial separable work, follow `docs/references/rules/agent-team-orchestration.md`, record the plan, and use at least one read-only verifier after implementation.",
		"Stay single-lane for trivial, tightly coupled, ambiguous, or high-overlap work, and record the reason.",
	}
}

func specV2FinalResponseContract() []finalResponseContractSection {
	return []finalResponseContractSection{
		{Heading: "Outcome", Items: []string{"State what changed and whether the workflow is complete, blocked, or ready for its next phase."}},
		{Heading: "Evidence", Items: []string{"List acceptance coverage and exact validation results, including any skipped check and risk."}},
		{Heading: "Artifacts And State", Items: []string{"List material repo-relative paths, current SPEC phase, and delivery state."}},
		{Heading: "Agent Team", Items: []string{"State actual agents used and verification performed. If none ran, write exactly: `single supervisor lane; no specialist or verification agents spawned`."}},
		{Heading: "Open Items", Items: []string{"List only genuine blockers or follow-ups; write `none` when clean."}},
	}
}
