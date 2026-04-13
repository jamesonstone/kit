package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
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
	cfg, _ := loadRepoInstructionContext(projectRoot)
	repoAgentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	repoReferencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Mine a reusable skill for feature: %s", feat.Slug))
		doc.Heading(2, "Context Docs")
		doc.Table(
			[]string{"File", "Path"},
			skillMineContextRows(
				constitutionPath,
				summaryPath,
				brainstormPath,
				specPath,
				planPath,
				tasksPath,
				skillsDir,
				claudeSkillsDir,
				repoAgentsPath,
				repoReferencesPath,
				skillPath,
				claudeSkillPath,
				projectRoot,
			),
		)
		doc.Heading(2, "Your Task")
		doc.OrderedList(1, skillMineTaskSteps(
			constitutionPath,
			summaryPath,
			brainstormPath,
			skillsDir,
			claudeSkillsDir,
			repoAgentsPath,
			repoReferencesPath,
			skillPath,
			claudeSkillPath,
		)...)
		doc.Heading(2, "Skill Bundle Format")
		doc.CodeBlock("markdown", skillBundleFormatTemplate())
		doc.Heading(2, "SKILL AUDIT")
		doc.Paragraph("Run after writing any new skill, or if no new skill was written")
		doc.OrderedList(1, skillAuditSteps(skillsDir)...)
		doc.Heading(2, "Skill Audit Summary Format")
		doc.CodeBlock("markdown", skillAuditSummaryTemplate())
		doc.Heading(2, "Rules")
		doc.BulletList(skillMineRules(skillsDir, claudeSkillsDir, projectRoot)...)
	})
}

func skillMineContextRows(
	constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath string,
	skillsDir, claudeSkillsDir, repoAgentsPath, repoReferencesPath, skillPath, claudeSkillPath, projectRoot string,
) [][]string {
	rows := [][]string{
		{"CONSTITUTION", constitutionPath},
	}
	if repoAgentsPath != "" {
		rows = append(rows, []string{"AGENTS DOCS", repoAgentsPath})
	}
	if repoReferencesPath != "" {
		rows = append(rows, []string{"REFERENCES", repoReferencesPath})
	}
	rows = append(rows,
		[]string{"PROJECT_PROGRESS_SUMMARY", summaryPath},
	)
	if document.Exists(brainstormPath) {
		rows = append(rows, []string{"BRAINSTORM", brainstormPath})
	}
	rows = append(rows,
		[]string{"SPEC", specPath},
		[]string{"PLAN", planPath},
		[]string{"TASKS", tasksPath},
		[]string{"Canonical Skills Root", skillsDir},
		[]string{"Claude Mirror Root", claudeSkillsDir},
		[]string{"Canonical Draft Output", skillPath},
		[]string{"Claude Mirror Output", claudeSkillPath},
		[]string{"Project Root", projectRoot},
	)

	return rows
}

func skillMineTaskSteps(
	constitutionPath, summaryPath, brainstormPath, skillsDir, claudeSkillsDir, repoAgentsPath, repoReferencesPath, skillPath, claudeSkillPath string,
) []string {
	steps := []string{
		fmt.Sprintf("Read `CONSTITUTION.md` first at `%s`", constitutionPath),
	}
	if repoAgentsPath != "" {
		steps = append(steps, "Read `docs/agents/README.md` and only the linked docs relevant to this feature's reusable workflow patterns")
	}
	if repoReferencesPath != "" {
		steps = append(steps, "Read `docs/references/README.md` only when a repo-wide reference materially shaped the feature")
	}
	steps = append(steps,
		fmt.Sprintf(
			"Read `PROJECT_PROGRESS_SUMMARY.md` at `%s` to understand cross-feature themes, what has been consistently hard, and what has been consistently smooth",
			summaryPath,
		),
		fmt.Sprintf("Read the feature's spec pipeline in order:\n%s", skillMineSpecPipelineList(brainstormPath)),
		"Run `git diff main` to capture what actually changed during implementation; if `main` does not exist, run `git diff master`",
		fmt.Sprintf(
			"Read all existing canonical skill bundles under `%s/*/SKILL.md` to avoid duplicating patterns that are already captured",
			skillsDir,
		),
		"Analyze the delta between what the spec pipeline planned and what the git diff shows was actually implemented; this divergence is the highest-signal source of reusable pattern content",
		"Extract patterns that are:\n- reusable across features or projects\n- not already covered by an existing skill in the skills directory\n- concrete and actionable rather than vague or project-specific",
		"Beyond pattern extraction, derive novel insights by synthesizing across multiple signals:\n\nA) SPEC DELTA ANALYSIS\n- Compare what was originally specified in SPEC.md to what was actually built per git diff and TASKS.md completion state\n- Divergences are high-signal: they reveal where the spec was wrong, where implementation discovered something better, or where constraints changed mid-flight\n- Capture these divergences as insights, not just patterns\n\nB) FEATURE PROGRESSION ANALYSIS\n- Read PROJECT_PROGRESS_SUMMARY.md to understand the arc of the project: which features are complete, which are in flight, what has been consistently hard or consistently smooth\n- Look for recurring themes across multiple features - these are systemic insights, not one-off patterns\n- A theme that appears in 2+ features is a strong skill candidate\n\nC) CONSTITUTION ALIGNMENT\n- Read CONSTITUTION.md and identify where the implementation reinforced, challenged, or refined the stated principles\n- If the work revealed a principle that should exist but does not, that is a novel insight worth capturing as a skill\n\nD) EMERGENT WORKFLOW INSIGHTS\n- Look for implicit workflows the team has developed that are not yet formalized anywhere in the spec pipeline\n- These are the highest-value skills: the things the team does that work well but have never been written down\n\nFor each novel insight found, ask: \"Would a new coding agent working on this project benefit from knowing this?\" If yes, it is a skill candidate.",
		fmt.Sprintf("Write the canonical skill bundle to `%s`", skillPath),
		fmt.Sprintf(
			"Duplicate the full skill directory into the Claude mirror at `%s` so Claude Code can discover the same skill bundle",
			filepath.Dir(claudeSkillPath),
		),
		"Use the exact transferable skill-bundle format shown in the `Skill Bundle Format` section below",
		"The `description` frontmatter field is critical - it must describe when the skill should trigger, not what it does. Model it on the trigger-condition descriptions in the available skills list from the system prompt",
		"If no reusable patterns or novel insights are found, say so explicitly and write nothing to the canonical or mirrored skills roots",
		"After writing the draft, print a one-paragraph summary of what pattern or insight was captured and why it was considered reusable",
	)

	return steps
}

func skillMineSpecPipelineList(brainstormPath string) string {
	docs := []string{}
	if document.Exists(brainstormPath) {
		docs = append(docs, "- `BRAINSTORM.md`")
	}
	docs = append(docs,
		"- `SPEC.md`",
		"- `PLAN.md`",
		"- `TASKS.md`",
	)

	return strings.Join(docs, "\n")
}

func skillBundleFormatTemplate() string {
	return `<skill-name>/
  SKILL.md
  scripts/        # optional
  references/     # optional
  assets/         # optional

---
name: <slug>
description: <one sentence: when to trigger this skill>
---

# <Title>

<procedural knowledge - what to do, in what order, with what constraints>`
}

func skillAuditSteps(skillsDir string) []string {
	return []string{
		fmt.Sprintf("Read every existing canonical skill at `%s/*/SKILL.md`", skillsDir),
		"For each existing skill, evaluate against four criteria:\n- ACCURACY - does the procedural guidance still match how the codebase actually works? If the code has changed in a way that makes the skill's instructions wrong, the skill is stale.\n- RELEVANCE - does the pattern the skill describes still appear in active development? If the feature or workflow it describes has been removed, superseded, or replaced, the skill is stale.\n- COVERAGE - is this skill now fully subsumed by a newer, broader skill? If so, the narrower one is redundant.\n- TRIGGER CONDITION - are the name and description frontmatter fields still valid triggering conditions? If the trigger condition no longer matches real usage, the skill will fire incorrectly.",
		"For each skill that fails any criterion:\n- State which criterion failed and why in one sentence\n- Add the skill to the `Removed` section as a removal candidate\n- Log the exact reason before asking for approval\n- Do NOT delete the canonical or Claude mirror directories yet",
		"For each skill that passes all criteria:\n- Mark it reviewed in the audit summary\n- Do NOT modify the canonical skill bundle unless it fails a criterion",
		fmt.Sprintf(
			"After the audit summary, ask for explicit approval before deleting any removal candidate:\n- if approved, delete the canonical skill directory with `rm -rf %s/<skill-name>/`\n- if approved, delete the Claude mirror directory with `rm -rf .claude/skills/<skill-name>/`\n- if approval is not granted, stop after the audit summary and leave the skill directories untouched",
			skillsDir,
		),
		"Output the summary format shown in the `Skill Audit Summary Format` section below at the end",
	}
}

func skillAuditSummaryTemplate() string {
	return `## Skill Audit Summary

### Created
- <skill-name>: <one sentence on what insight it captures>

### Removed
- <skill-name>: <one sentence on why it was removed>

### Retained
- <skill-name>: reviewed, still accurate

### No action
- <reason if no skills were created or removed>`
}

func skillMineRules(skillsDir, claudeSkillsDir, projectRoot string) []string {
	return []string{
		"output a SKILL.md draft only if a genuinely reusable pattern or insight exists",
		"do NOT write project-specific implementation details as skills",
		"do NOT duplicate patterns already present in existing canonical skills",
		"keep skill content procedural and agent-executable, not descriptive",
		"the description frontmatter must be a triggering condition, not a summary",
		"skill content must be dense: no fluff, no preamble, no obvious context",
		"if the git diff is empty or unavailable, rely on spec pipeline only",
		"use the canonical skills root as the source of truth and the Claude root only as a duplicated discovery mirror",
		"signal priority for insight derivation (highest to lowest):\n  1. spec-vs-implementation divergence (SPEC.md vs git diff)\n  2. recurring themes across 2+ features (PROJECT_PROGRESS_SUMMARY.md)\n  3. implicit workflows not yet in any document\n  4. single-feature reusable patterns\n  5. constitution alignment gaps",
		"a skill derived from signal 1 or 2 is always worth writing",
		"a skill derived from signal 4 or 5 requires a stronger reusability case",
		"always run the skill audit, even if no new skill was written",
		"never modify a canonical skill that passes all four audit criteria",
		"treat stale-skill cleanup as approval-gated: identify removal candidates first, then ask before deleting anything",
		"if approval is granted, remove the full directory, not just SKILL.md",
		"log every removal candidate with a reason before asking for approval",
		"a skill that is merely incomplete is not stale - only delete if wrong, irrelevant, redundant, or mis-triggered",
		"if all existing skills pass audit and no new skill is warranted, output \"No skill changes - audit complete\" and stop",
		fmt.Sprintf("skills_dir: %s", skillsDir),
		fmt.Sprintf("claude mirror root: %s", claudeSkillsDir),
		fmt.Sprintf("project root: %s", projectRoot),
	}
}
