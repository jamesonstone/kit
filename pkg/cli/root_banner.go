package cli

import "strings"

const (
	reset        = "\033[0m"
	dim          = "\033[38;5;245m"
	whiteBold    = "\033[1;37m"
	gray         = "\033[38;5;240m"
	constitution = "\033[38;5;220m"
	brainstorm   = "\033[38;5;117m"
	spec         = "\033[38;5;39m"
	plan         = "\033[38;5;82m"
	tasks        = "\033[38;5;213m"
	implement    = "\033[38;5;208m"
	reflect      = "\033[38;5;141m"
)

func rootLong(style humanOutputStyle) string {
	return rootBanner(style) + `
Kit v2 is a general-purpose harness for disciplined thought work.
Its strongest engine is a document-first, spec-driven workflow, but the
harness also supports ad hoc execution, catch-up, handoff, summarization,
review, and orchestration.

The current command surface is packaged around repository and software
workflows, but the underlying harness patterns generalize to research,
strategy, operations, writing, policy, and other structured fields.

` + flowDiagram(style)
}

func rootBanner(style humanOutputStyle) string {
	colors := []string{
		"\033[38;5;213m",
		"\033[38;5;177m",
		"\033[38;5;134m",
		"\033[38;5;97m",
		"\033[38;5;60m",
		"\033[38;5;238m",
	}

	lines := []string{
		"██╗  ██╗██╗████████╗",
		"██║ ██╔╝██║╚══██╔══╝",
		"█████╔╝ ██║   ██║   ",
		"██╔═██╗ ██║   ██║   ",
		"██║  ██╗██║   ██║   ",
		"╚═╝  ╚═╝╚═╝   ╚═╝   ",
	}

	var result string
	for i, line := range lines {
		result += "                                        " + rootColor(style, colors[i], line) + "\n"
	}
	result += "\n"
	result += "                                      " + rootMuted(style, "Kit v2 Thought-Work Harness") + "\n"
	return result
}

func flowDiagram(style humanOutputStyle) string {
	idea := rootColor(style, brainstorm, "Idea")
	specCommand := rootColor(style, spec, "kit spec <feature>")
	clarify := rootColor(style, brainstorm, "Clarifying Loop")
	planLane := rootColor(style, plan, "Supervisor + Agent Team Plan")
	implementLane := rootColor(style, implement, "Subagent Implementation")
	reflectLane := rootColor(style, reflect, "Subagent Reflection")
	validateLane := rootColor(style, tasks, "Subagent Validation / Verification")
	evidence := rootColor(style, constitution, "Evidence + Delivery Gate")

	lines := []string{
		rootHeading(style, "🧱 Project Initialization") + rootMuted(style, " (run once, update as needed):"),
		rootColor(style, gray, "┌──────────────┐"),
		rootColor(style, gray, "│ ") + rootColor(style, constitution, "Constitution") + rootColor(style, gray, " │  ← ") + rootMuted(style, "global constraints, principles, priors"),
		rootColor(style, gray, "└──────────────┘"),
		"",
		rootHeading(style, "🔁 V2 Feature Workflow") + rootMuted(style, " (default):"),
		"  " + idea + rootMuted(style, " / input"),
		"    " + rootArrow(style),
		"  " + specCommand + rootMuted(style, " creates/updates one durable SPEC.md"),
		"    " + rootArrow(style),
		"  " + clarify + rootMuted(style, " → questions, source map, binary acceptance criteria"),
		"    " + rootArrow(style),
		"  " + planLane + rootMuted(style, " → supervisor owns scope, lanes, touched files"),
		"    " + rootArrow(style),
		"  " + implementLane,
		"    " + rootArrow(style),
		"  " + reflectLane,
		"    " + rootArrow(style),
		"  " + validateLane + rootMuted(style, " → each criterion proved or routed back"),
		"    " + rootArrow(style),
		"  " + evidence + rootMuted(style, " → SPEC.md evidence, docs sync, complete"),
		"",
		rootHeading(style, "🗂️ Durable Artifacts"),
		"  1. " + rootColor(style, constitution, "CONSTITUTION.md") + rootMuted(style, " — project contract and invariants"),
		"  2. " + rootColor(style, spec, "SPEC.md") + rootMuted(style, "         — v2 feature artifact and workflow state"),
		"  3. " + rootColor(style, brainstorm, "BRAINSTORM/PLAN/TASKS") + rootMuted(style, " — legacy v1 artifacts, historical unless using kit legacy"),
	}

	return strings.Join(lines, "\n")
}

func rootHeading(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return whiteBold + text + reset
}

func rootMuted(style humanOutputStyle, text string) string {
	if !style.enabled {
		return text
	}
	return dim + text + reset
}

func rootColor(style humanOutputStyle, color string, text string) string {
	if !style.enabled {
		return text
	}
	return color + text + reset
}

func rootArrow(style humanOutputStyle) string {
	return rootColor(style, gray, "│") + "\n    " + rootColor(style, gray, "▼")
}
