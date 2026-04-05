package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func outputCompiledPrompt(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	answers *specAnswers,
	outputOnly bool,
) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You MUST update the specification file at:
**File**: %s
**Feature**: %s
**Project Root**: %s

## Context Provided by User

`, specPath, featureSlug, projectRoot))

	if answers.Problem != "" {
		sb.WriteString(fmt.Sprintf("**PROBLEM**: %s\n\n", answers.Problem))
	}
	if answers.Goals != "" {
		sb.WriteString(fmt.Sprintf("**GOALS**: %s\n\n", answers.Goals))
	}
	if answers.NonGoals != "" {
		sb.WriteString(fmt.Sprintf("**NON-GOALS**: %s\n\n", answers.NonGoals))
	}
	if answers.Users != "" {
		sb.WriteString(fmt.Sprintf("**USERS**: %s\n\n", answers.Users))
	}
	if answers.Requirements != "" {
		sb.WriteString(fmt.Sprintf("**REQUIREMENTS**: %s\n\n", answers.Requirements))
	}
	if answers.Acceptance != "" {
		sb.WriteString(fmt.Sprintf("**ACCEPTANCE**: %s\n\n", answers.Acceptance))
	}
	if answers.EdgeCases != "" {
		sb.WriteString(fmt.Sprintf("**EDGE-CASES**: %s\n\n", answers.EdgeCases))
	}

	hasContext := answers.Problem != "" || answers.Goals != "" || answers.NonGoals != "" ||
		answers.Users != "" || answers.Requirements != "" || answers.Acceptance != "" ||
		answers.EdgeCases != ""

	sb.WriteString("## Context Docs (read first)\n")
	sb.WriteString("| File | Purpose |\n")
	sb.WriteString("|------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	appendRepoInstructionContextRows(&sb, projectRoot, cfg)
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	appendSpecSkillDiscoveryContextRows(&sb, projectRoot, cfg)
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))

	sb.WriteString("## Your Task\n\n")
	sb.WriteString(fmt.Sprintf("1. Read CONSTITUTION.md (file: %s) to understand project constraints and principles\n", constitutionPath))
	step := appendRepoInstructionReadStep(&sb, 2, projectRoot, cfg)
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("%d. Read BRAINSTORM.md (file: %s) and treat it as upstream research context\n", step, brainstormPath))
		step++
	}
	sb.WriteString(fmt.Sprintf("%d. Read the current SPEC.md (file: %s) and understand the required sections\n", step, specPath))
	step++
	sb.WriteString(fmt.Sprintf("%d. Analyze the codebase at %s to understand existing patterns\n", step, projectRoot))

	questionStart := step + 1
	if hasContext {
		sb.WriteString(fmt.Sprintf(
			"%d. **IMMEDIATELY write all context above into the SPEC.md file at %s** — do NOT ask questions before doing this\n",
			questionStart,
			specPath,
		))
		questionStart++
	}

	questionStart = appendSpecSkillDiscoveryStep(&sb, questionStart, projectRoot, cfg, specPath)
	questionStart = appendSpecRelationshipsStep(&sb, questionStart, specPath)
	questionStart = appendSpecDependencyInventoryStep(&sb, questionStart, specPath, brainstormPath, hasBrainstorm)
	questionStart = appendNumberedSteps(
		&sb,
		questionStart,
		clarificationLoopSteps(
			goalPct,
			fmt.Sprintf(
				"Reassess, save your updates to %s, and continue with additional "+
					"batches of up to 10 questions until the specification is precise "+
					"enough to produce a correct, production-quality solution",
				specPath,
			),
		),
	)

	if hasContext {
		sb.WriteString(fmt.Sprintf("%d. Continue refining each section of SPEC.md as you learn more:\n", questionStart))
	} else {
		sb.WriteString(fmt.Sprintf("%d. **Write your findings directly to %s** as you fill in each section:\n", questionStart, specPath))
	}

	if hasBrainstorm {
		sb.WriteString("   - Carry forward validated findings from BRAINSTORM.md into SPEC.md\n")
	}

	sb.WriteString(fmt.Sprintf(`   - PROBLEM: What problem does this feature solve?
   - GOALS: What are the measurable outcomes?
   - NON-GOALS: What is explicitly out of scope?
   - USERS: Who will use this feature?
   - SKILLS: Which documented skills should the coding agent use for this feature, where do they live, and when should each one trigger?
   - RELATIONSHIPS: Which prior features does this work build on, depend on, or relate to, if any?
   - DEPENDENCIES: Which supporting docs, tools, design refs, APIs, libraries, datasets, assets, and other inputs shaped this specification?
   - REQUIREMENTS: What must be true for this feature to be complete?
   - ACCEPTANCE: How do we verify the feature works?
   - EDGE-CASES: What unusual scenarios must be handled?

Do NOT treat SPEC.md as complete until confidence reaches ≥%d%% and unresolved assumptions = 0.

## SUMMARY Section (MANDATORY)
Once you reach ≥%d%% confidence, write a SUMMARY section at the top of SPEC.md:
- 1-2 sentences maximum
- Information-dense: include the core problem, solution approach, and key constraint
- Written for a coding agent who needs to quickly understand the feature
- Example: "Adds CSV export for user data with streaming support for large datasets (>100k rows). Must complete in <5s and handle Unicode."

## IMPORTANT: File Update Requirement
All specification content MUST be written to: %s
This file is the single source of truth for this feature. Do not leave content only in chat — persist everything to the file.

## Rules
- Keep language precise
- Avoid implementation details (focus on WHAT, not HOW)
- the ## SKILLS section is mandatory and must be populated before sign-off
- the ## RELATIONSHIPS section is mandatory and must be set to none or explicit bullets using canonical feature identifiers before sign-off
- the ## DEPENDENCIES section must be current before sign-off and must keep exact locations for external design inputs
- use repo instruction files, repo-local canonical skills, and documented global inputs during the skills discovery phase
- keep the selected skill set minimal and actionable
- do not use .claude/skills as canonical discovery input
- Spec gate: unresolved assumptions = 0 before sign-off; if unresolved assumptions remain, stop and resolve before marking SPEC complete
- Ensure the spec respects constraints defined in CONSTITUTION.md
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, goalPct, goalPct, specPath))
	appendNonEmptySectionRules(&sb, "`SPEC.md`")

	if err := outputPromptWithClipboardDefault(sb.String(), outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied prompt into your coding agent",
			"Work with the agent to refine the specification",
			fmt.Sprintf("Run 'kit plan %s' to create the implementation plan", featureSlug),
		})
	}

	return nil
}
