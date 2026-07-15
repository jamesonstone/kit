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
Kit is a repository-memory and specification harness for agent-driven work.
Native agent planning owns research and design; Kit preserves consequential
rationale, validation, and outcomes in canonical repository documents.

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
	result += "                                      " + rootMuted(style, "Kit Repository-Memory Harness") + "\n"
	return result
}

func flowDiagram(style humanOutputStyle) string {
	idea := rootColor(style, brainstorm, "Request")
	nativePlan := rootColor(style, plan, "Native Agent Plan")
	specCommand := rootColor(style, spec, "kit spec <feature>")
	implementLane := rootColor(style, implement, "Implementation")
	validateLane := rootColor(style, tasks, "Validation")
	evidence := rootColor(style, constitution, "Curated Repository Memory")

	lines := []string{
		rootHeading(style, "🧱 Project Initialization") + rootMuted(style, " (run once, update as needed):"),
		rootColor(style, gray, "┌──────────────┐"),
		rootColor(style, gray, "│ ") + rootColor(style, constitution, "Constitution") + rootColor(style, gray, " │  ← ") + rootMuted(style, "global constraints, principles, priors"),
		rootColor(style, gray, "└──────────────┘"),
		"",
		rootHeading(style, "🔁 Native Plan To Repository Memory") + rootMuted(style, " (default):"),
		"  " + idea + rootMuted(style, " / input"),
		"    " + rootArrow(style),
		"  " + nativePlan + rootMuted(style, " → research, clarification, design, accepted plan"),
		"    " + rootArrow(style),
		"  " + specCommand + rootMuted(style, " → create/adopt SPEC.md when material rationale exists"),
		"    " + rootArrow(style),
		"  " + implementLane,
		"    " + rootArrow(style),
		"  " + validateLane + rootMuted(style, " → observable acceptance and exact evidence"),
		"    " + rootArrow(style),
		"  " + evidence + rootMuted(style, " → spec, invariants, references, domain docs, or not required"),
		"",
		rootHeading(style, "🗂️ Durable Artifacts"),
		"  1. " + rootColor(style, constitution, "CONSTITUTION.md") + rootMuted(style, " — project contract and invariants"),
		"  2. " + rootColor(style, spec, "SPEC.md") + rootMuted(style, "         — material feature rationale and living history"),
		"  3. " + rootColor(style, brainstorm, "REFERENCES / DOMAIN DOCS") + rootMuted(style, " — reusable and scope-specific knowledge"),
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
