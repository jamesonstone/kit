package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func runSpecTemplate(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	outputOnly bool,
) error {
	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	hasBrainstorm := document.Exists(brainstormPath)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Please review and complete the specification at %s.\n\n", specPath))
	sb.WriteString(fmt.Sprintf("This is a new feature: %s\n\n", featureSlug))
	sb.WriteString("## Context Docs (read first)\n")
	sb.WriteString(fmt.Sprintf("|- CONSTITUTION: %s — project-wide constraints, principles, priors\n", constitutionPath))
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("|- REPO INSTRUCTION: %s — active workflow and skill usage rules\n", path))
	}
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("|- BRAINSTORM: %s — upstream research context and codebase findings\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("|- CANONICAL SKILLS: %s/*/SKILL.md — repo-local reusable skills\n", cfg.SkillsPath(projectRoot)))
	for _, path := range globalSkillDiscoveryInputs() {
		sb.WriteString(fmt.Sprintf("|- GLOBAL INPUT: %s — documented global skill or instruction input\n", path))
	}

	sb.WriteString(`

## Context Provided by User
<!-- ⚠️ FILL THIS OUT BEFORE SUBMITTING TO YOUR CODING AGENT -->

**PROBLEM**:
<!-- What problem does this feature solve? -->

**GOALS**:
<!-- What are the measurable outcomes? (comma-separated) -->

**NON-GOALS**:
<!-- What is explicitly out of scope? -->

**USERS**:
<!-- Who will use this feature? -->

**REQUIREMENTS**:
<!-- What must be true for this feature to be complete? -->

**ACCEPTANCE**:
<!-- How do we verify the feature works? -->

**EDGE-CASES**:
<!-- What unusual scenarios must be handled? -->

## Your Task

1. Read CONSTITUTION.md to understand project constraints and principles
`)

	if hasBrainstorm {
		sb.WriteString("2. Read the repository instruction files first\n")
		sb.WriteString("3. Read BRAINSTORM.md and carry forward validated findings\n")
		sb.WriteString("4. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("5. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("6. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString("7. Perform a skills discovery phase before asking sign-off questions:\n")
		sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", cfg.SkillsPath(projectRoot)))
		for _, path := range globalSkillDiscoveryInputs() {
			sb.WriteString(fmt.Sprintf("   - inspect `%s`\n", path))
		}
		sb.WriteString(fmt.Sprintf("   - populate the `## SKILLS` table in `%s`\n", specPath))
		sb.WriteString("   - keep the required `none | n/a | n/a | no additional skills required | no` row if nothing else applies\n")
		sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
		sb.WriteString(fmt.Sprintf("8. Populate or refresh the `## RELATIONSHIPS` section in `%s` before sign-off:\n", specPath))
		sb.WriteString("   - use `none` when this feature does not build on an existing feature\n")
		sb.WriteString("   - otherwise record one bullet per explicit feature relationship\n")
		sb.WriteString("   - supported labels are `builds on`, `depends on`, and `related to`\n")
		sb.WriteString("   - use canonical feature directory identifiers such as `0007-catchup-command`\n")
		sb.WriteString(fmt.Sprintf("9. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", specPath))
		sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", brainstormPath))
		sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
		sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
		sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
		sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
		sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
		sb.WriteString("   - keep the default `none` row only if no additional dependencies apply\n")
		nextStep := appendNumberedSteps(
			&sb,
			10,
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
		sb.WriteString(fmt.Sprintf("%d. Continue refining each section of SPEC.md as you learn more:\n", nextStep))
	} else {
		sb.WriteString("2. Read the repository instruction files first\n")
		sb.WriteString("3. Read the SPEC.md template and understand the required sections\n")
		sb.WriteString(fmt.Sprintf("4. Analyze the codebase at %s to understand existing patterns\n", projectRoot))
		sb.WriteString("5. **IMMEDIATELY update SPEC.md** with the context provided above before asking any questions\n")
		sb.WriteString("6. Perform a skills discovery phase before asking sign-off questions:\n")
		sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", cfg.SkillsPath(projectRoot)))
		for _, path := range globalSkillDiscoveryInputs() {
			sb.WriteString(fmt.Sprintf("   - inspect `%s`\n", path))
		}
		sb.WriteString(fmt.Sprintf("   - populate the `## SKILLS` table in `%s`\n", specPath))
		sb.WriteString("   - keep the required `none | n/a | n/a | no additional skills required | no` row if nothing else applies\n")
		sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
		sb.WriteString(fmt.Sprintf("7. Populate or refresh the `## RELATIONSHIPS` section in `%s` before sign-off:\n", specPath))
		sb.WriteString("   - use `none` when this feature does not build on an existing feature\n")
		sb.WriteString("   - otherwise record one bullet per explicit feature relationship\n")
		sb.WriteString("   - supported labels are `builds on`, `depends on`, and `related to`\n")
		sb.WriteString("   - use canonical feature directory identifiers such as `0007-catchup-command`\n")
		sb.WriteString(fmt.Sprintf("8. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", specPath))
		sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
		sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
		sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
		sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
		sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
		sb.WriteString("   - keep the default `none` row only if no additional dependencies apply\n")
		nextStep := appendNumberedSteps(
			&sb,
			9,
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
		sb.WriteString(fmt.Sprintf("%d. Continue refining each section of SPEC.md as you learn more:\n", nextStep))
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
`, goalPct, goalPct))
	appendNonEmptySectionRules(&sb, "`SPEC.md`")

	if !outputOnly {
		fmt.Println()
		fmt.Println(dim + "⚠️ IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
		fmt.Println(dim + "   with details about your feature. The more context you provide, the" + reset)
		fmt.Println(dim + "   better the agent can help you write the specification." + reset)
		fmt.Println()
		fmt.Println(dim + "   Tip: Run 'kit spec <feature> --interactive' for a guided" + reset)
		fmt.Println(dim + "   editor-first experience, or add '--inline' for terminal multiline entry." + reset)
		fmt.Println()
	}

	if err := outputPromptWithClipboardDefault(sb.String(), outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define the specification", specPath),
			fmt.Sprintf("Run 'kit plan %s' to create the implementation plan", featureSlug),
		})
	}

	return nil
}
