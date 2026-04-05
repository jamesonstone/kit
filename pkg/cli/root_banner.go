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
	result += "                                   " + dim + "General-Purpose Thought-Work Harness" + reset + "\n"
	return result
}

func flowDiagram() string {
	lines := []string{
		whiteBold + "🧱 Project Initialization" + reset + dim + " (run once, update as needed):" + reset,
		gray + "┌──────────────┐" + reset,
		gray + "│ " + constitution + "Constitution" + reset + gray + " │  ← " + reset + dim + "global constraints, principles, priors" + reset,
		gray + "└──────────────┘" + reset,
		"",
		whiteBold + "🧠 Optional Research Step" + reset + dim + ":" + reset,
		gray + "  ┌────────────┐" + reset,
		gray + "  │ " + brainstorm + "Brainstorm" + reset + gray + " │  ← " + reset + dim + "research, framing, options, affected artifacts" + reset,
		gray + "  └─────┬──────┘" + reset,
		gray + "        │" + reset,
		gray + "        ▼" + reset,
		"",
		whiteBold + "🔁 Core Development Loop" + reset + dim + ":" + reset,
		gray + "┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐" + reset,
		gray + "│ " + spec + "Specification" + reset + gray + " │ ─▶ │ " + plan + "Plan" + reset + gray + " │ ─▶ │ " + tasks + "Tasks" + reset + gray + " │ ─▶ │ " + implement + "Implementation" + reset + gray + " │ ─▶ │ " + reflect + "Reflection" + reset + gray + " │ ─┐" + reset,
		gray + "└───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │" + reset,
		gray + "       ▲                                                                            │" + reset,
		gray + "       └────────────────────────────────────────────────────────────────────────────┘" + reset,
		"",
		whiteBold + "🗂️ Structured Engine: Artifact Pipeline" + reset,
		"  1. " + constitution + "Constitution" + reset + dim + "   — strategy, patterns, long-term vision (kept updated)" + reset,
		"  2. " + brainstorm + "Brainstorm" + reset + dim + "     — optional research and framing before the spec" + reset,
		"  3. " + spec + "Specification" + reset + dim + "  — what is being built and why" + reset,
		"  4. " + plan + "Plan" + reset + dim + "           — how it will be built" + reset,
		"  5. " + tasks + "Tasks" + reset + dim + "          — executable work units" + reset,
		"  6. " + implement + "Implementation" + reset + dim + " — execution begins after the readiness gate" + reset,
		"  7. " + reflect + "Reflection" + reset + dim + "     — verify correctness, refine understanding" + reset,
	}

	return strings.Join(lines, "\n")
}
