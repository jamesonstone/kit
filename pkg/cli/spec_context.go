package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

type specAnswers struct {
	Problem      string
	Goals        string
	NonGoals     string
	Users        string
	Requirements string
	Acceptance   string
	EdgeCases    string
}

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
			fmt.Sprintf("- use `%s` when the work is repository-scale and needs broad discovery", filepath.Join(projectRoot, "docs", "agents", "RLM.md")),
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
		fmt.Sprintf("- write the selected skills into the `## SKILLS` table in `%s`", specPath),
		"- if no additional skills apply, keep the required `none | n/a | n/a | no additional skills required | no` row",
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
		fmt.Sprintf("Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:", specPath),
	}
	if hasBrainstorm {
		lines = append(lines, fmt.Sprintf("- carry forward still-relevant dependencies from `%s`", brainstormPath))
	}
	lines = append(lines,
		"- keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`",
		"- include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shaped the feature definition",
		"- use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`",
		"- `Status` must be one of `active`, `optional`, or `stale`",
		"- for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"- if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`",
		"- if no additional dependencies apply, keep the default `none` row",
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
		fmt.Sprintf("Populate or refresh the `## RELATIONSHIPS` section in `%s` before sign-off:", specPath),
		"- use `none` when this feature does not build on an existing feature",
		"- otherwise record one bullet per explicit feature relationship",
		"- supported labels are `builds on`, `depends on`, and `related to`",
		"- use canonical feature directory identifiers such as `0007-catchup-command`",
	}, "\n")
}

func appendSpecRelationshipsStep(sb *strings.Builder, step int, specPath string) int {
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(specRelationshipsStepText(specPath), "\n", "\n   ")))
	return step + 1
}
