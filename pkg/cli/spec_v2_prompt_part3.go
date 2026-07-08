package cli

import (
	"fmt"
	"strings"
)

func specV2UserContext(answers *specAnswers) string {
	if answers == nil || (answers.Problem == "" && answers.Goals == "" && answers.NonGoals == "" &&
		answers.Users == "" && answers.Requirements == "" && answers.Acceptance == "" &&
		answers.EdgeCases == "" && answers.DeliveryIntent == "") {
		return strings.TrimSpace(`<!-- Derive these values from the current SPEC.md and repo context; ask clarification questions instead of inventing missing details. -->

**THESIS**:
<!-- Read the current SPEC.md Thesis section. -->

**CONTEXT**:
<!-- Infer from SPEC.md, repo research, Source Map entries, and clarification answers. -->

**REQUIREMENTS**:
<!-- Clarify before implementation; do not invent unstated requirements. -->

**ACCEPTANCE CRITERIA**:
<!-- Create stable binary-verifiable AC-### entries during clarification. -->

**DELIVERY INTENT**:
<!-- Read SPEC.md front matter delivery_intent and the Delivery Decision section. -->

**NON-GOALS / EXCLUSIONS**:
<!-- Clarify explicitly before implementation. -->`)
	}

	var items []string
	appendAnswer := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		items = append(items, fmt.Sprintf("**%s**: %s", label, value))
	}
	appendAnswer("THESIS", answers.Problem)
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

func specV2AgentTeamModeBullets(singleAgentMode bool) []string {
	if singleAgentMode {
		return []string{
			"`--single-agent` is active. Keep execution in one supervisor lane and do not require implementation or verification subagents.",
			"Even in single-agent mode, record logical work lanes in the Agent Team Plan when they clarify sequencing, validation, or risk.",
			"In final responses, state `single supervisor lane; no specialist or verification agents spawned` and cite `--single-agent` as the exception.",
		}
	}

	return []string{
		"Default to a subagent team for implementation and verification.",
		"Use a single supervisor lane only when the work is trivial, tightly coupled, the active runtime cannot spawn subagents, or `--single-agent` is explicitly active.",
		"If implementation or verification stays single-lane, record the exact exception in `SPEC.md` before implementation or validation begins.",
		"When the active coding-agent runtime supports subagents, spawn implementation and verification subagents according to the Agent Team Plan; do not keep work single-lane merely because subagents were not explicitly re-requested.",
	}
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
			Items:   []string{"Summarize actual lanes used, subagents actually spawned, lanes intentionally omitted, verification lanes, concurrency, and overlap decisions. If no separate agents actually ran, write `single supervisor lane; no specialist or verification agents spawned`, state the exception that justified single-lane execution, and do not present logical planning lanes as spawned agents."},
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
