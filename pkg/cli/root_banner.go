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
Kit is a repository-local contract and evidence harness for coding agents.
It materializes durable rules, workflows, specifications, strategies, and
implementation patterns, then resolves only the evidence needed for the work.
Kit does not choose a model, infer project truth, or launch an agent.

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
	result += "                                      " + rootMuted(style, "Kit Coding-Agent Contract") + "\n"
	return result
}

func flowDiagram(style humanOutputStyle) string {
	capabilities := rootColor(style, brainstorm, "kit capabilities <command> --json")
	context := rootColor(style, plan, "kit context resolve --workflow <slug> --json")
	agent := rootColor(style, implement, "Coding Agent")
	reconcile := rootColor(style, constitution, "kit reconcile")

	lines := []string{
		rootHeading(style, "🔁 Agent-First Workflow"),
		"  " + capabilities + rootMuted(style, " → establish command safety"),
		"    " + rootArrow(style),
		"  " + context + rootMuted(style, " → select ordered local evidence"),
		"    " + rootArrow(style),
		"  " + agent + rootMuted(style, " → plan, implement, validate, and curate memory"),
		"    " + rootArrow(style),
		"  " + reconcile + rootMuted(style, " → detect and safely curate drift"),
		"",
		rootHeading(style, "🗂️ Canonical Inputs"),
		"  " + rootColor(style, spec, "SPEC.md") + rootMuted(style, " · rules · workflows · Constitution · references · source evidence"),
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
