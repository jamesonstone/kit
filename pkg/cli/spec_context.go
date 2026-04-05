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

func appendRepoInstructionContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		label := filepath.Base(path)
		if strings.Contains(path, ".github/copilot-instructions.md") {
			label = "COPILOT"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", label, path))
	}
}

func appendSpecSkillDiscoveryContextRows(sb *strings.Builder, projectRoot string, cfg *config.Config) {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()

	sb.WriteString(fmt.Sprintf("| Canonical Skills Root | %s/*/SKILL.md |\n", skillsDir))
	sb.WriteString(fmt.Sprintf("| Claude Global | %s |\n", globalInputs[0]))
	sb.WriteString(fmt.Sprintf("| Codex Global AGENTS | %s |\n", globalInputs[1]))
	sb.WriteString(fmt.Sprintf("| Codex Global Instructions | %s |\n", globalInputs[2]))
	sb.WriteString(fmt.Sprintf("| Codex Global Skills | %s |\n", globalInputs[3]))
}

func appendRepoInstructionReadStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
) int {
	sb.WriteString(fmt.Sprintf("%d. Read the repository instruction files first:\n", step))
	for _, path := range repoInstructionPaths(projectRoot, cfg) {
		sb.WriteString(fmt.Sprintf("   - `%s`\n", path))
	}
	return step + 1
}

func appendSpecSkillDiscoveryStep(
	sb *strings.Builder,
	step int,
	projectRoot string,
	cfg *config.Config,
	specPath string,
) int {
	skillsDir := cfg.SkillsPath(projectRoot)
	globalInputs := globalSkillDiscoveryInputs()

	sb.WriteString(fmt.Sprintf("%d. Perform a skills discovery phase before treating SPEC.md as complete:\n", step))
	sb.WriteString(fmt.Sprintf("   - inspect repo-local canonical skills under `%s/*/SKILL.md`\n", skillsDir))
	sb.WriteString("   - inspect documented global inputs:\n")
	for _, path := range globalInputs {
		sb.WriteString(fmt.Sprintf("     - `%s`\n", path))
	}
	sb.WriteString("   - choose the minimal relevant set of skills for this feature\n")
	sb.WriteString(fmt.Sprintf("   - write the selected skills into the `## SKILLS` table in `%s`\n", specPath))
	sb.WriteString("   - if no additional skills apply, keep the required `none | n/a | n/a | no additional skills required | no` row\n")
	sb.WriteString("   - do not use `.claude/skills` as canonical discovery input\n")
	return step + 1
}

func appendSpecDependencyInventoryStep(
	sb *strings.Builder,
	step int,
	specPath string,
	brainstormPath string,
	hasBrainstorm bool,
) int {
	sb.WriteString(fmt.Sprintf("%d. Populate or refresh the `## DEPENDENCIES` table in `%s` before sign-off:\n", step, specPath))
	if hasBrainstorm {
		sb.WriteString(fmt.Sprintf("   - carry forward still-relevant dependencies from `%s`\n", brainstormPath))
	}
	sb.WriteString("   - keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`\n")
	sb.WriteString("   - include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shaped the feature definition\n")
	sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
	sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
	sb.WriteString("   - for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
	sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale`\n")
	sb.WriteString("   - if no additional dependencies apply, keep the default `none` row\n")
	return step + 1
}

func appendSpecRelationshipsStep(sb *strings.Builder, step int, specPath string) int {
	sb.WriteString(fmt.Sprintf("%d. Populate or refresh the `## RELATIONSHIPS` section in `%s` before sign-off:\n", step, specPath))
	sb.WriteString("   - use `none` when this feature does not build on an existing feature\n")
	sb.WriteString("   - otherwise record one bullet per explicit feature relationship\n")
	sb.WriteString("   - supported labels are `builds on`, `depends on`, and `related to`\n")
	sb.WriteString("   - use canonical feature directory identifiers such as `0007-catchup-command`\n")
	return step + 1
}
