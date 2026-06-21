package cli

import "github.com/jamesonstone/kit/internal/promptdoc"

type finalResponseContractSection struct {
	Heading string
	Items   []string
}

func addFinalResponseContract(doc *promptdoc.Document, sections ...finalResponseContractSection) {
	doc.Heading(2, "Final Response Contract")
	doc.Paragraph("End your final response with these exact headings, in this order. Keep it concise and use repo-relative paths for every file or artifact path.")
	for _, section := range sections {
		doc.Heading(3, section.Heading)
		doc.BulletList(section.Items...)
	}
}

func brainstormFinalResponseContract(featureSlug string) []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Summary",
			Items:   []string{"1-2 sentences describing the research outcome and current confidence."},
		},
		{
			Heading: "Artifacts Updated",
			Items:   []string{"List repo-relative paths for BRAINSTORM.md and any notes or design artifacts materially updated."},
		},
		{
			Heading: "Key Decisions",
			Items:   []string{"Bullets for decisions, recommended defaults, or tradeoffs resolved during brainstorming."},
		},
		{
			Heading: "Open Questions",
			Items:   []string{"List unresolved questions, or write `none` when unresolved assumptions are zero."},
		},
		{
			Heading: "Next Step",
			Items:   []string{"State the next Kit command, usually `kit spec " + featureSlug + "`."},
		},
	}
}

func specFinalResponseContract(featureSlug string) []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Summary",
			Items:   []string{"1-2 sentences describing the current SPEC.md phase, feature state, and confidence."},
		},
		{
			Heading: "Artifacts Updated",
			Items:   []string{"List repo-relative paths for SPEC.md and any supporting artifacts updated."},
		},
		{
			Heading: "Workflow State",
			Items:   []string{"Bullets for the current v2 phase, readiness gate status, acceptance criteria status, validation status, and delivery decision."},
		},
		{
			Heading: "Acceptance And Evidence",
			Items:   []string{"Bullets for acceptance criteria, their mapped validation evidence, and any remaining gaps."},
		},
		{
			Heading: "Open Questions",
			Items:   []string{"List unresolved questions, or write `none` when unresolved assumptions are zero."},
		},
		{
			Heading: "Next Step",
			Items:   []string{"State whether the next action is more clarification, implementation, validation, reflection, delivery, or completion inside `SPEC.md`; regenerate with `kit spec " + featureSlug + " --prompt-only` when needed."},
		},
	}
}

func planFinalResponseContract(featureSlug string) []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Summary",
			Items:   []string{"1-2 sentences describing the chosen implementation approach and current confidence."},
		},
		{
			Heading: "Artifacts Updated",
			Items:   []string{"List repo-relative paths for PLAN.md and any supporting artifacts updated."},
		},
		{
			Heading: "Design Decisions",
			Items:   []string{"Bullets for the main implementation decisions and tradeoffs now fixed in PLAN.md."},
		},
		{
			Heading: "Implementation Risks",
			Items:   []string{"Bullets for remaining risks and mitigations, or write `none`."},
		},
		{
			Heading: "Next Step",
			Items:   []string{"State the next Kit command, usually `kit legacy tasks " + featureSlug + "`."},
		},
	}
}

func tasksFinalResponseContract(featureSlug string) []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Summary",
			Items:   []string{"1-2 sentences describing the generated execution plan and current confidence."},
		},
		{
			Heading: "Artifacts Updated",
			Items:   []string{"List repo-relative paths for TASKS.md and any supporting artifacts updated."},
		},
		{
			Heading: "Task Breakdown",
			Items:   []string{"Summarize task count, ordering, and dependency shape without repeating the full task file."},
		},
		{
			Heading: "Blocked Items",
			Items:   []string{"List blockers or missing decisions, or write `none`."},
		},
		{
			Heading: "Next Step",
			Items:   []string{"State the next Kit command, usually `kit legacy implement " + featureSlug + "`."},
		},
	}
}

func implementFinalResponseContract() []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Work Done",
			Items:   []string{"Concise bullets for completed behavior and task IDs, without restating the whole plan."},
		},
		{
			Heading: "Files Changed",
			Items:   []string{"List repo-relative paths with a one-line reason for each file or grouped file set."},
		},
		{
			Heading: "Validation",
			Items:   []string{"State exactly what ran and the result; if not run, write `not run` and the reason."},
		},
		{
			Heading: "How To Test",
			Items:   []string{"Give practical commands or manual checks the user can run to verify the change."},
		},
		{
			Heading: "How To View",
			Items:   []string{"Explain how to open or exercise user-visible behavior; if no visible surface exists, write `not applicable`."},
		},
		{
			Heading: "Docs/Tasks Updated",
			Items:   []string{"List TASKS.md, PROJECT_PROGRESS_SUMMARY.md, or other docs updated; write `none` if none changed."},
		},
		{
			Heading: "Follow-ups",
			Items:   []string{"List remaining blockers, risks, or next work; write `none` if no follow-up is needed."},
		},
	}
}

func reflectFinalResponseContract() []finalResponseContractSection {
	return []finalResponseContractSection{
		{
			Heading: "Changeset",
			Items:   []string{"List repo-relative changed files and tight bullets for the key diffs."},
		},
		{
			Heading: "Verification",
			Items:   []string{"State exactly what ran, including run IDs when available; if not run, write `not run` and the reason."},
		},
		{
			Heading: "Review Findings",
			Items:   []string{"List findings fixed during reflection and any remaining risks; write `none` when clean."},
		},
		{
			Heading: "Doc Trace",
			Items:   []string{"Report pass/fail notes for BRAINSTORM, SPEC, PLAN, and TASKS when feature-scoped; otherwise write `not applicable`."},
		},
		{
			Heading: "Final Status",
			Items:   []string{"State whether reflection is complete, whether TASKS.md was marked with REFLECTION_COMPLETE, and whether project refresh was needed."},
		},
		{
			Heading: "Follow-ups",
			Items:   []string{"List follow-up work, or write `none`."},
		},
	}
}
