package cli

import (
	"fmt"
	"strings"
)

const (
	approvalSyntaxSummary = "`yes`/`y` approves the full batch, `yes 3, 4, 5`/`y 3, 4, 5` approves specific items, and `no`/`n` forms reject defaults or provide overrides"
	approvalAllDefaults   = "\"yes\" or \"y\" approves all recommended defaults in the batch"
	approvalSomeDefaults  = "\"yes 3, 4, 5\" or \"y 3, 4, 5\" approves only those numbered defaults in the batch"
	overrideDefault       = "\"no 2: <answer>\" or \"n 2: <answer>\" rejects question 2's recommended default and provides the override"
	rejectAllDefaults     = "\"no\" or \"n\" rejects all recommended defaults in the batch and requires explicit replacements before proceeding"
)

func clarificationLoopSteps(goalPct int, continueInstruction string) []string {
	return []string{
		fmt.Sprintf(
			"Ask clarifying questions until you reach ≥%d%% confidence that you "+
				"understand the problem and desired solution",
			goalPct,
		),
		"Use numbered lists",
		"Ask questions in batches of up to 10",
		"For every question, include your current best recommended default, " +
			"proposed solution, or assumption",
		"State uncertainties",
		fmt.Sprintf(
			"Accept lean approvals from the user for the current batch: %s\n"+
				"   - %s\n"+
				"   - %s\n"+
				"   - %s\n"+
				"   - %s",
			approvalSyntaxSummary,
			approvalAllDefaults,
			approvalSomeDefaults,
			overrideDefault,
			rejectAllDefaults,
		),
		"If the user approves only specific question numbers, treat all other " +
			"questions in that batch as unresolved",
		"After each batch of up to 10 questions, output your current percentage " +
			"understanding so the user can see progress",
		continueInstruction,
	}
}

func appendNumberedSteps(
	sb *strings.Builder,
	startStep int,
	steps []string,
) int {
	step := startStep
	for _, instruction := range steps {
		sb.WriteString(fmt.Sprintf("%d. %s\n", step, instruction))
		step++
	}

	return step
}
