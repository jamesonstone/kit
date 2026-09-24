package cli

// Superseded guidance that reconcile flags when it survives in a managed
// instruction file. Each entry names wording that a later contract replaced.

const (
	legacyOperatorActionTableHeader = "| Type | Action required | Why | Continue with |"
	legacyStatusHeading             = "# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>"
	legacyPrioritizedActionList     = "prioritized action list ordered Blocker, Incomplete, Next, Optional, then None"
	legacyNoneBulletMandate         = "Use one `**None.**` bullet when there are no deviations"
	legacyNestedEvidenceMandate     = "Use at most one nested evidence layer and state each fact once"
	legacyThreeSectionMandate       = "emit exactly `## What happened`, `## Deviations`, and `## Next steps` in that order"
	legacyThreeSectionRestatement   = "`## What happened`, `## Deviations`, and `## Next steps`, in that order"
	legacyStatusTokenMandate        = "**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**"
	legacyDensityBudgetMandate      = "target twelve rendered lines or fewer"
)

func v3ForbiddenGuidance() map[string][]string {
	completion := []string{
		legacyOperatorActionTableHeader,
		legacyStatusHeading,
		legacyPrioritizedActionList,
		legacyNoneBulletMandate,
		legacyNestedEvidenceMandate,
		legacyThreeSectionMandate,
		legacyThreeSectionRestatement,
		legacyStatusTokenMandate,
		legacyDensityBudgetMandate,
	}
	workLane := []string{
		"Before I make any repository changes, should I create a new GitHub issue",
		"`c` means continue existing",
		"Wait for the explicit choice",
	}
	consent := []string{
		"never imply merge consent",
		"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
		"Obtain one explicit user confirmation for the complete bounded batch",
	}
	threadLifecycle := []string{
		"## Conversation Naming",
		"kit instructions naming",
		"kit instructions title",
		"## Codex Thread Initialization Hard Gate",
		"set_thread_pinned",
		"set_thread_title",
	}
	return map[string][]string{
		"AGENTS.md":                       append(append(append(completion, workLane...), consent...), threadLifecycle...),
		"CLAUDE.md":                       threadLifecycle,
		".github/copilot-instructions.md": append(append(completion, workLane...), consent...),
		"docs/agents/GUARDRAILS.md":       append(append(completion, workLane...), consent...),
	}
}
