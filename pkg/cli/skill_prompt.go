package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func buildSkillMinePrompt(
	feat *feature.Feature,
	brainstormPath, specPath, planPath, tasksPath string,
	skillsDir string,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	claudeSkillsDir := filepath.Join(projectRoot, ".claude", "skills")
	skillPath := filepath.Join(skillsDir, feat.Slug, "SKILL.md")
	claudeSkillPath := filepath.Join(claudeSkillsDir, feat.Slug, "SKILL.md")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Mine a reusable skill for feature: %s\n\n", feat.Slug))

	appendSkillMineContextDocs(
		&sb,
		constitutionPath,
		summaryPath,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		skillsDir,
		claudeSkillsDir,
		skillPath,
		claudeSkillPath,
		projectRoot,
	)
	appendSkillMineTask(
		&sb,
		constitutionPath,
		summaryPath,
		brainstormPath,
		skillsDir,
		claudeSkillsDir,
		skillPath,
		claudeSkillPath,
	)
	appendSkillAuditSection(&sb, skillsDir)
	appendSkillMineRules(&sb, skillsDir, claudeSkillsDir, projectRoot)

	return sb.String()
}

func appendSkillMineContextDocs(
	sb *strings.Builder,
	constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath string,
	skillsDir, claudeSkillsDir, skillPath, claudeSkillPath, projectRoot string,
) {
	sb.WriteString("## Context Docs\n")
	sb.WriteString("| File | Path |\n")
	sb.WriteString("|------|------|\n")
	sb.WriteString(fmt.Sprintf("| CONSTITUTION | %s |\n", constitutionPath))
	sb.WriteString(fmt.Sprintf("| PROJECT_PROGRESS_SUMMARY | %s |\n", summaryPath))
	if document.Exists(brainstormPath) {
		sb.WriteString(fmt.Sprintf("| BRAINSTORM | %s |\n", brainstormPath))
	}
	sb.WriteString(fmt.Sprintf("| SPEC | %s |\n", specPath))
	sb.WriteString(fmt.Sprintf("| PLAN | %s |\n", planPath))
	sb.WriteString(fmt.Sprintf("| TASKS | %s |\n", tasksPath))
	sb.WriteString(fmt.Sprintf("| Canonical Skills Root | %s |\n", skillsDir))
	sb.WriteString(fmt.Sprintf("| Claude Mirror Root | %s |\n", claudeSkillsDir))
	sb.WriteString(fmt.Sprintf("| Canonical Draft Output | %s |\n", skillPath))
	sb.WriteString(fmt.Sprintf("| Claude Mirror Output | %s |\n", claudeSkillPath))
	sb.WriteString(fmt.Sprintf("| Project Root | %s |\n\n", projectRoot))
}

func appendSkillMineTask(
	sb *strings.Builder,
	constitutionPath, summaryPath, brainstormPath, skillsDir, claudeSkillsDir, skillPath, claudeSkillPath string,
) {
	sb.WriteString("## Your Task\n")
	sb.WriteString(fmt.Sprintf("1. Read `CONSTITUTION.md` first at `%s`\n", constitutionPath))
	sb.WriteString(
		fmt.Sprintf(
			"2. Read `PROJECT_PROGRESS_SUMMARY.md` at `%s` to understand cross-feature themes, what has been consistently hard, and what has been consistently smooth\n",
			summaryPath,
		),
	)
	sb.WriteString("3. Read the feature's spec pipeline in order:\n")
	if document.Exists(brainstormPath) {
		sb.WriteString("   - `BRAINSTORM.md`\n")
	}
	sb.WriteString("   - `SPEC.md`\n")
	sb.WriteString("   - `PLAN.md`\n")
	sb.WriteString("   - `TASKS.md`\n")
	sb.WriteString(
		"4. Run `git diff main` to capture what actually changed during implementation; if `main` does not exist, run `git diff master`\n",
	)
	sb.WriteString(
		fmt.Sprintf(
			"5. Read all existing canonical skill bundles under `%s/*/SKILL.md` to avoid duplicating patterns that are already captured\n",
			skillsDir,
		),
	)
	sb.WriteString(
		"6. Analyze the delta between what the spec pipeline planned and what the git diff shows was actually implemented; this divergence is the highest-signal source of reusable pattern content\n",
	)
	sb.WriteString("7. Extract patterns that are:\n")
	sb.WriteString("   - reusable across features or projects\n")
	sb.WriteString("   - not already covered by an existing skill in the skills directory\n")
	sb.WriteString("   - concrete and actionable rather than vague or project-specific\n")
	sb.WriteString(
		"8. Beyond pattern extraction, derive novel insights by synthesizing across multiple signals:\n",
	)
	sb.WriteString("\nA) SPEC DELTA ANALYSIS\n")
	sb.WriteString(
		"   - Compare what was originally specified in SPEC.md to what was actually built per git diff and TASKS.md completion state\n",
	)
	sb.WriteString(
		"   - Divergences are high-signal: they reveal where the spec was wrong, where implementation discovered something better, or where constraints changed mid-flight\n",
	)
	sb.WriteString("   - Capture these divergences as insights, not just patterns\n")
	sb.WriteString("\nB) FEATURE PROGRESSION ANALYSIS\n")
	sb.WriteString(
		"   - Read PROJECT_PROGRESS_SUMMARY.md to understand the arc of the project: which features are complete, which are in flight, what has been consistently hard or consistently smooth\n",
	)
	sb.WriteString(
		"   - Look for recurring themes across multiple features - these are systemic insights, not one-off patterns\n",
	)
	sb.WriteString("   - A theme that appears in 2+ features is a strong skill candidate\n")
	sb.WriteString("\nC) CONSTITUTION ALIGNMENT\n")
	sb.WriteString(
		"   - Read CONSTITUTION.md and identify where the implementation reinforced, challenged, or refined the stated principles\n",
	)
	sb.WriteString(
		"   - If the work revealed a principle that should exist but does not, that is a novel insight worth capturing as a skill\n",
	)
	sb.WriteString("\nD) EMERGENT WORKFLOW INSIGHTS\n")
	sb.WriteString(
		"   - Look for implicit workflows the team has developed that are not yet formalized anywhere in the spec pipeline\n",
	)
	sb.WriteString(
		"   - These are the highest-value skills: the things the team does that work well but have never been written down\n",
	)
	sb.WriteString(
		"\nFor each novel insight found, ask: \"Would a new coding agent working on this project benefit from knowing this?\" If yes, it is a skill candidate.\n",
	)
	sb.WriteString(fmt.Sprintf("9. Write the canonical skill bundle to `%s`\n", skillPath))
	sb.WriteString(
		fmt.Sprintf(
			"10. Duplicate the full skill directory into the Claude mirror at `%s` so Claude Code can discover the same skill bundle\n",
			filepath.Join(claudeSkillsDir, filepath.Base(filepath.Dir(claudeSkillPath))),
		),
	)
	sb.WriteString("11. Use this exact transferable skill-bundle format:\n\n")
	sb.WriteString("```markdown\n")
	sb.WriteString("<skill-name>/\n")
	sb.WriteString("  SKILL.md\n")
	sb.WriteString("  scripts/        # optional\n")
	sb.WriteString("  references/     # optional\n")
	sb.WriteString("  assets/         # optional\n\n")
	sb.WriteString("  ---\n")
	sb.WriteString("  name: <slug>\n")
	sb.WriteString("  description: <one sentence: when to trigger this skill>\n")
	sb.WriteString("  ---\n\n")
	sb.WriteString("  # <Title>\n\n")
	sb.WriteString("  <procedural knowledge - what to do, in what order, with what constraints>\n")
	sb.WriteString("```\n")
	sb.WriteString(
		"12. The `description` frontmatter field is critical - it must describe when the skill should trigger, not what it does. Model it on the trigger-condition descriptions in the available skills list from the system prompt\n",
	)
	sb.WriteString(
		"13. If no reusable patterns or novel insights are found, say so explicitly and write nothing to the canonical or mirrored skills roots\n",
	)
	sb.WriteString(
		"14. After writing the draft, print a one-paragraph summary of what pattern or insight was captured and why it was considered reusable\n\n",
	)
}

func appendSkillAuditSection(sb *strings.Builder, skillsDir string) {
	sb.WriteString("## SKILL AUDIT\n")
	sb.WriteString("Run after writing any new skill, or if no new skill was written\n\n")
	sb.WriteString(
		fmt.Sprintf(
			"1. Read every existing canonical skill at `%s/*/SKILL.md`\n",
			skillsDir,
		),
	)
	sb.WriteString("2. For each existing skill, evaluate against four criteria:\n\n")
	sb.WriteString("   ACCURACY - does the procedural guidance still match how the codebase actually works? If the code has changed in a way that makes the skill's instructions wrong, the skill is stale.\n\n")
	sb.WriteString("   RELEVANCE - does the pattern the skill describes still appear in active development? If the feature or workflow it describes has been removed, superseded, or replaced, the skill is stale.\n\n")
	sb.WriteString("   COVERAGE - is this skill now fully subsumed by a newer, broader skill? If so, the narrower one is redundant.\n\n")
	sb.WriteString("   TRIGGER CONDITION - are the name and description frontmatter fields still valid triggering conditions? If the trigger condition no longer matches real usage, the skill will fire incorrectly.\n")
	sb.WriteString("3. For each skill that fails any criterion:\n")
	sb.WriteString("   - State which criterion failed and why in one sentence\n")
	sb.WriteString(
		fmt.Sprintf(
			"   - Delete the canonical skill directory with `rm -rf %s/<skill-name>/`\n",
			skillsDir,
		),
	)
	sb.WriteString("   - Delete the skill directory with `rm -rf .claude/skills/<skill-name>/`\n")
	sb.WriteString("   - Log the deletion in the audit summary before executing it\n")
	sb.WriteString("4. For each skill that passes all criteria:\n")
	sb.WriteString("   - Mark it reviewed in the audit summary\n")
	sb.WriteString("   - Do NOT modify the canonical skill bundle unless it fails a criterion\n")
	sb.WriteString("5. Output this summary format at the end:\n\n")
	sb.WriteString("```markdown\n")
	sb.WriteString("## Skill Audit Summary\n\n")
	sb.WriteString("### Created\n")
	sb.WriteString("- <skill-name>: <one sentence on what insight it captures>\n\n")
	sb.WriteString("### Removed\n")
	sb.WriteString("- <skill-name>: <one sentence on why it was removed>\n\n")
	sb.WriteString("### Retained\n")
	sb.WriteString("- <skill-name>: reviewed, still accurate\n\n")
	sb.WriteString("### No action\n")
	sb.WriteString("- <reason if no skills were created or removed>\n")
	sb.WriteString("```\n\n")
}

func appendSkillMineRules(sb *strings.Builder, skillsDir, claudeSkillsDir, projectRoot string) {
	sb.WriteString("Rules:\n")
	sb.WriteString("- output a SKILL.md draft only if a genuinely reusable pattern or insight exists\n")
	sb.WriteString("- do NOT write project-specific implementation details as skills\n")
	sb.WriteString("- do NOT duplicate patterns already present in existing canonical skills\n")
	sb.WriteString("- keep skill content procedural and agent-executable, not descriptive\n")
	sb.WriteString("- the description frontmatter must be a triggering condition, not a summary\n")
	sb.WriteString("- skill content must be dense: no fluff, no preamble, no obvious context\n")
	sb.WriteString("- if the git diff is empty or unavailable, rely on spec pipeline only\n")
	sb.WriteString("- use the canonical skills root as the source of truth and the Claude root only as a duplicated discovery mirror\n")
	sb.WriteString("- signal priority for insight derivation (highest to lowest):\n")
	sb.WriteString("  1. spec-vs-implementation divergence (SPEC.md vs git diff)\n")
	sb.WriteString("  2. recurring themes across 2+ features (PROJECT_PROGRESS_SUMMARY.md)\n")
	sb.WriteString("  3. implicit workflows not yet in any document\n")
	sb.WriteString("  4. single-feature reusable patterns\n")
	sb.WriteString("  5. constitution alignment gaps\n")
	sb.WriteString("- a skill derived from signal 1 or 2 is always worth writing\n")
	sb.WriteString("- a skill derived from signal 4 or 5 requires a stronger reusability case\n")
	sb.WriteString("- always run the skill audit, even if no new skill was written\n")
	sb.WriteString("- never modify a canonical skill that passes all four audit criteria\n")
	sb.WriteString("- deletion is permanent: remove the full directory, not just SKILL.md\n")
	sb.WriteString("- log every deletion with a reason before executing it\n")
	sb.WriteString("- a skill that is merely incomplete is not stale - only delete if wrong, irrelevant, redundant, or mis-triggered\n")
	sb.WriteString("- if all existing skills pass audit and no new skill is warranted, output \"No skill changes - audit complete\" and stop\n")
	sb.WriteString(fmt.Sprintf("- skills_dir: %s\n", skillsDir))
	sb.WriteString(fmt.Sprintf("- claude mirror root: %s\n", claudeSkillsDir))
	sb.WriteString(fmt.Sprintf("- project root: %s\n", projectRoot))
}
