package templates

import "strings"

func memoryInstructionSupportContent(relativePath string) string {
	switch relativePath {
	case "docs/agents/README.md":
		return memoryAgentsREADME
	case "docs/references/testing.md":
		return strings.ReplaceAll(referencesTesting, "Validation Map and Evidence sections", "VALIDATION and OUTCOME sections")
	case "docs/references/README.md":
		return memoryReferencesREADME()
	case "docs/references/worktrees.md":
		return referencesWorktrees
	default:
		return ""
	}
}

// memoryAgentsREADME maps the instruction architecture without restating rules.
const memoryAgentsREADME = `# Agents Docs

## Purpose

- Universal contract: the Kit-managed block in ` + "`AGENTS.md`" + `, rendered identically into ` + "`CLAUDE.md`" + ` and ` + "`.github/copilot-instructions.md`" + `. Kit regenerates the block; keep project guidance outside it.
- Contextual rules: ` + "`docs/references/rules/<slug>.md`" + `, read only when the contract's trigger applies.
- Project memory: ` + "`docs/CONSTITUTION.md`" + `, ` + "`docs/specs/<feature>/SPEC.md`" + `, and ` + "`docs/references/`" + `.
- Workflow evidence lists: ` + "`docs/references/workflows/<slug>.md`" + `, available through ` + "`kit context resolve --workflow <slug> --json`" + `.
`
