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

func banner() string {
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
		result += "                                        " + colors[i] + line + reset + "\n"
	}
	result += "\n"
	result += "                                      " + dim + "Kit v2 Thought-Work Harness" + reset + "\n"
	return result
}

func flowDiagram() string {
	lines := []string{
		whiteBold + "🧱 Project Initialization" + reset + dim + " (run once, update as needed):" + reset,
		gray + "┌──────────────┐" + reset,
		gray + "│ " + constitution + "Constitution" + reset + gray + " │  ← " + reset + dim + "global constraints, principles, priors" + reset,
		gray + "└──────────────┘" + reset,
		"",
		whiteBold + "🔁 V2 Feature Workflow" + reset + dim + " (default):" + reset,
		gray + "┌──────────────┐    ┌───────────────────────────────────────────────────────────────┐" + reset,
		gray + "│ " + brainstorm + "Idea / Input" + reset + gray + " │ ─▶ │ " + spec + "kit spec <feature>" + reset + gray + "                                           │" + reset,
		gray + "└──────────────┘    │ " + spec + "SPEC.md" + reset + gray + ": clarify → ready → implement → validate → reflect → deliver │" + reset,
		gray + "                    └───────────────────────────────────────────────────────────────┘" + reset,
		"",
		whiteBold + "🗂️ Durable Artifacts" + reset,
		"  1. " + constitution + "CONSTITUTION.md" + reset + dim + " — project contract and invariants" + reset,
		"  2. " + spec + "SPEC.md" + reset + dim + "         — v2 feature artifact: thesis, context, clarifications, requirements, assumptions," + reset,
		dim + "                    acceptance criteria, plan, task checklist, validation map, reflection, docs, delivery, evidence" + reset,
		"  3. " + brainstorm + "BRAINSTORM/PLAN/TASKS" + reset + dim + " — legacy v1 staged artifacts, preserved as historical context when present" + reset,
	}

	return strings.Join(lines, "\n")
}
