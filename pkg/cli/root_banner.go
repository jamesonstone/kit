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
It renders one universal agent contract, ships contextual rules with each
release, and validates and converges Kit-managed project state.
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
	initCommand := rootColor(style, brainstorm, "kit init")
	check := rootColor(style, plan, "kit check --project")
	agent := rootColor(style, implement, "Coding Agent")
	reconcile := rootColor(style, constitution, "kit reconcile")

	lines := []string{
		rootHeading(style, "🔁 Agent-First Workflow"),
		"  " + initCommand + rootMuted(style, " → universal contract, core rules, project memory"),
		"    " + rootArrow(style),
		"  " + agent + rootMuted(style, " → plan, implement, validate, and curate memory"),
		"    " + rootArrow(style),
		"  " + check + rootMuted(style, " → validate specs, rules, contract, and source limits"),
		"    " + rootArrow(style),
		"  " + reconcile + rootMuted(style, " → detect and safely curate drift"),
		"",
		rootHeading(style, "🗂️ Canonical Inputs"),
		"  " + rootColor(style, spec, "SPEC.md") + rootMuted(style, " · contextual rules · Constitution · testing reference · source"),
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
