package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

type specAnswers struct {
	Problem        string
	Goals          string
	NonGoals       string
	Users          string
	Requirements   string
	Acceptance     string
	EdgeCases      string
	DeliveryIntent string
}

const (
	specDeliveryIntentIdeaOnly           = "idea_only"
	specDeliveryIntentIssueBranchPRLater = "issue_branch_pr_later"
	specDeliveryIntentContinueCurrent    = "continue_current"
)

func normalizeSpecAnswer(raw string) string {
	return strings.TrimSpace(raw)
}

func repoInstructionContextRows(projectRoot string, cfg *config.Config) [][]string {
	rows := make([][]string, 0, len(repoInstructionPaths(projectRoot, cfg))+6)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		label := filepath.Base(path)
		if strings.Contains(path, ".github/copilot-instructions.md") {
			label = "COPILOT"
		}
		rows = append(rows, []string{label, path})
	}
	if version == config.InstructionScaffoldVersionTOC {
		rows = append(rows,
			[]string{"AGENTS DOCS", filepath.Join(projectRoot, "docs", "agents", "README.md")},
			[]string{"WORKFLOWS", filepath.Join(projectRoot, "docs", "agents", "WORKFLOWS.md")},
			[]string{"RLM", filepath.Join(projectRoot, "docs", "agents", "RLM.md")},
			[]string{"TOOLING", filepath.Join(projectRoot, "docs", "agents", "TOOLING.md")},
			[]string{"GUARDRAILS", filepath.Join(projectRoot, "docs", "agents", "GUARDRAILS.md")},
			[]string{"REFERENCES", filepath.Join(projectRoot, "docs", "references", "README.md")},
		)
	}

	return rows
}

func appendRepoInstructionContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	for _, row := range repoInstructionContextRows(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", row[0], row[1]))
	}
}

func specSkillDiscoveryContextRows(projectRoot string, cfg *config.Config) [][]string {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()
	rows := [][]string{
		{"Canonical Skills Root", fmt.Sprintf("%s/*/SKILL.md", skillsDir)},
	}
	if detectInstructionScaffoldVersion(projectRoot, cfg) == config.InstructionScaffoldVersionTOC {
		rows = append(rows,
			[]string{"Repo Agents Entry", filepath.Join(projectRoot, "docs", "agents", "README.md")},
			[]string{"Repo References Entry", filepath.Join(projectRoot, "docs", "references", "README.md")},
		)
	}
	rows = append(rows,
		[]string{"Claude Global", globalInputs[0]},
		[]string{"Codex Global AGENTS", globalInputs[1]},
		[]string{"Codex Global Instructions", globalInputs[2]},
		[]string{"Codex Global Skills", globalInputs[3]},
	)
	return rows
}

func appendSpecSkillDiscoveryContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	for _, row := range specSkillDiscoveryContextRows(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", row[0], row[1]))
	}
}

func repoInstructionReadStepText(projectRoot string, cfg *config.Config) string {
	if detectInstructionScaffoldVersion(projectRoot, cfg) == config.InstructionScaffoldVersionTOC {
		lines := []string{
			"Read the repository entrypoint files first, then route through the repo-local docs tree:",
		}
		for _, path := range repoInstructionPaths(projectRoot, cfg) {
			lines = append(lines, fmt.Sprintf("- entrypoint: `%s`", path))
		}
		lines = append(lines,
			fmt.Sprintf("- repo-local entrypoint: `%s`", filepath.Join(projectRoot, "docs", "agents", "README.md")),
			"- from there, read only the relevant docs under `docs/agents/*`, `docs/specs/*`, and `docs/references/*`",
		)
		return strings.Join(lines, "\n")
	}

	lines := []string{"Read the repository instruction files first:"}
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		lines = append(lines, fmt.Sprintf("- `%s`", path))
	}
	return strings.Join(lines, "\n")
}

func appendRepoInstructionReadStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
) int {
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(repoInstructionReadStepText(projectRoot, cfg), "\n", "\n   ")))
	return step + 1
}

func specSkillDiscoveryStepText(projectRoot string, cfg *config.Config, specPath string) string {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()
	lines := []string{"Perform a skills discovery phase before treating SPEC.md as complete:"}
	if detectInstructionScaffoldVersion(projectRoot, cfg) == config.InstructionScaffoldVersionTOC {
		lines = append(lines,
			fmt.Sprintf("- start at `%s` and load only the relevant linked docs for this feature", filepath.Join(projectRoot, "docs", "agents", "README.md")),
			fmt.Sprintf("- use `%s` when full-context loading would be noisy or wasteful", filepath.Join(projectRoot, "docs", "agents", "RLM.md")),
		)
	}
	lines = append(lines, fmt.Sprintf("- inspect repo-local canonical skills under `%s/*/SKILL.md`", skillsDir))
	if detectInstructionScaffoldVersion(projectRoot, cfg) == config.InstructionScaffoldVersionTOC {
		lines = append(lines, "- inspect documented global inputs only after repo-local docs are exhausted:")
	} else {
		lines = append(lines, "- inspect documented global inputs:")
	}
	for i, path := range globalInputs {
		if i == 0 {
			lines = append(lines, fmt.Sprintf("  - `%s`", path))
			continue
		}
		lines = append(lines, fmt.Sprintf("  - `%s`", path))
	}
	lines = append(lines,
		"- choose the minimal relevant set of skills for this feature",
		fmt.Sprintf("- write the selected skills into canonical front matter `skills` in `%s`; use the legacy `## SKILLS` table only when front matter is absent", specPath),
		"- if no additional skills apply, leave front matter skills empty or keep the legacy `none | n/a | n/a | no additional skills required | no` row in documents without front matter",
		"- do not use `.claude/skills` as canonical discovery input",
	)
	return strings.Join(lines, "\n")
}

func appendSpecSkillDiscoveryStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
	specPath string,
) int {
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(specSkillDiscoveryStepText(projectRoot, cfg, specPath), "\n", "\n   ")))
	return step + 1
}

func specDependencyInventoryStepText(specPath, brainstormPath string, hasBrainstorm bool) string {
	lines := []string{
		fmt.Sprintf("Populate or refresh canonical front matter `references` in `%s` before sign-off:", specPath),
	}
	if hasBrainstorm {
		lines = append(lines, fmt.Sprintf("- carry forward still-relevant references from `%s`", brainstormPath))
	}
	lines = append(lines,
		"- keep `skills` focused on execution-time agent skills and track broader supporting inputs in `references`",
		"- include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shaped the feature definition",
		"- use `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`",
		"- add a stable `id` when the reference may need to be updated later",
		"- `selector_type` must be one of `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id` when `selector` is set",
		"- `relation` describes the referenced target's role relative to the source artifact, such as `constrains`, `guides`, `informs`, `implements`, `verifies`, or `uses`",
		"- `read_policy` must be one of `must`, `conditional`, `evidence`, or `skip`",
		"- `status` must be one of `active`, `optional`, or `stale`",
		"- for Figma or MCP-driven design references, store the exact design URL or file/node reference in `target` and use stable selectors when needed",
		"- if a reference influenced decisions but is no longer current, keep it with `status: stale` and `read_policy: skip`",
		"- if no additional references apply, leave front matter references empty and keep the body `## DEPENDENCIES` section prose-only",
	)
	return strings.Join(lines, "\n")
}

func appendSpecDependencyInventoryStep(
	sb *strings.Builder,
	step int,
	specPath string,
	brainstormPath string,
	hasBrainstorm bool,
) int {
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(specDependencyInventoryStepText(specPath, brainstormPath, hasBrainstorm), "\n", "\n   ")))
	return step + 1
}

func specRelationshipsStepText(specPath string) string {
	return strings.Join([]string{
		fmt.Sprintf("Populate or refresh canonical front matter `relationships` in `%s` before sign-off, using the legacy `## RELATIONSHIPS` section only when front matter is absent:", specPath),
		"- omit relationships or use `none` only in legacy body metadata when this feature does not build on an existing feature",
		"- otherwise record one entry per explicit feature relationship",
		"- supported front matter types are `builds_on`, `depends_on`, and `related_to`; supported legacy labels are `builds on`, `depends on`, and `related to`",
		"- use canonical feature directory identifiers such as `0007-catchup-command`",
	}, "\n")
}

func appendSpecRelationshipsStep(sb *strings.Builder, step int, specPath string) int {
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(specRelationshipsStepText(specPath), "\n", "\n   ")))
	return step + 1
}

func relatedFeatureContextStepText(projectRoot, currentDocPath string) string {
	featureDir := filepath.Dir(currentDocPath)
	featureID := filepath.Base(featureDir)
	specsDir := filepath.ToSlash(filepath.Dir(featureDir))
	summaryPath := filepath.ToSlash(filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"))

	lines := []string{
		fmt.Sprintf("Use an RLM-style just-in-time prior-work pass over `%s` before broad repository reads:", specsDir),
		"- must-read inputs stay small: the current task or section plus explicit relationships and references already in scope",
		fmt.Sprintf("- use indices first: start with `kit map %s` and `%s` to shortlist candidate prior features", featureID, summaryPath),
		fmt.Sprintf("- if `kit map` is unavailable, inspect `%s` directly and use the current feature's canonical front matter relationships and references", specsDir),
		"- prior feature docs, repo references, and secondary global inputs are conditional reads only",
		"- open a prior feature doc only if it affects a shared interface or contract, overlapping files or modules, migrations or data shape, acceptance criteria, or an explicit relationship or reference link",
		"- inspect at most 5 prior feature directories before narrowing further or asking a clarifying question",
		"- extract only the concrete facts that change the current feature's requirements, strategy, interfaces, refactor surface, or tests",
		"- do not paraphrase entire prior docs into chat or copy irrelevant historical context into the current artifact",
	}

	return strings.Join(lines, "\n")
}
