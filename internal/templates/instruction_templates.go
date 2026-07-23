package templates

import (
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
)

const sharedRepositoryInstructions = sharedRepositoryInstructionsCore + sharedRepositoryInstructionsStandards

const copilotQuickRules = `## Fast rules for chat and code review

- classify every request first
  - **spec-driven**: ` + "`kit spec`" + ` work, explicit legacy staged ` + "`kit legacy`" + ` work, new capability, substantial behavioral or architectural change, existing spec-covered feature work, or cross-component/public-interface changes
  - **ad hoc**: contained bug fix, review, refactor, dependency update, config change, or small refinement
- for spec-driven work:
  - read ` + "`SPEC.md`" + ` first for v2 feature work; it carries requirements, plan, task checklist, validation, reflection, delivery, and evidence
  - treat ` + "`BRAINSTORM.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + ` as legacy staged context unless the user explicitly chooses a legacy staged command
  - resolve repository-discoverable facts first and ask numbered questions only for material non-discoverable choices
  - include a recommended default and impact for every question; proceed without routine approval when no material question remains
  - run the v2 readiness gates before writing code; if any gate fails, update ` + "`SPEC.md`" + ` first
  - implement from the ` + "`SPEC.md`" + ` task checklist and update ` + "`SPEC.md`" + ` first if implementation changes behavior, requirements, approach, validation, reflection, documentation, or delivery state
- for ad hoc work:
  - follow understand → implement → verify
  - update existing spec docs when the change alters behavior, requirements, or approach
- always:
  - never mix multiple features in one ` + "`docs/specs/<feature>/`" + ` directory
  - keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with canonical docs
  - populate every required section in v2 ` + "`SPEC.md`" + `; for legacy staged workflows, populate every required section in the active staged artifact
  - replace placeholder-only sections with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + ` instead of leaving HTML TODO comments
  - keep each lane in its existing checkout or a canonical ` + "`~/worktrees/<owner>/<repository>/<lane>`" + ` worktree; never nest worktrees inside repositories or discard state to create one
  - prefer readable, maintainable code with explicit error handling and focused functions
  - fix all lint and test failures before completion and wait for the user's output before triaging findings they report
  - do NOT run ` + "`coderabbit --prompt-only`" + `, ` + "`git add`" + `, or ` + "`git commit`" + ` without explicit approval

---
`

type ScaffoldFile struct {
	RelativePath string
	Content      string
}

func repositoryInstructionDocument(title string) string {
	return `# ` + title + `

` + sharedRepositoryInstructions
}

func tocInstructionDocument(title string) string {
	return `# ` + title + `

` + tocRepositoryInstructions(title)
}

func memoryInstructionDocument(title string) string {
	return `# ` + title + `

` + memoryRepositoryInstructions(title)
}

// LegacyAgentPointer returns the comprehensive legacy instruction template for agent files.
func LegacyAgentPointer(agentName string) string {
	return repositoryInstructionDocument(agentName)
}

// AgentPointer returns the default instruction template for agent files.
func AgentPointer(agentName string) string {
	return tocInstructionDocument(agentName)
}

// LegacyAgentsMD is the comprehensive AGENTS.md template with full workflow and standards.
var LegacyAgentsMD = repositoryInstructionDocument("AGENTS")

// LegacyClaudeMD is the comprehensive CLAUDE.md template with full workflow and standards.
var LegacyClaudeMD = repositoryInstructionDocument("CLAUDE")

// LegacyCopilotInstructionsMD is the comprehensive repository-wide Copilot instructions template.
var LegacyCopilotInstructionsMD = `# GitHub Copilot Repository Instructions

` + copilotQuickRules + `
` + sharedRepositoryInstructions

// AgentsMD is the default AGENTS.md scaffold content.
var AgentsMD = tocInstructionDocument("AGENTS")

// ClaudeMD is the default CLAUDE.md scaffold content.
var ClaudeMD = tocInstructionDocument("CLAUDE")

// CopilotInstructionsMD is the default Copilot instructions scaffold content.
var CopilotInstructionsMD = tocCopilotInstructions

// MemoryAgentsMD is the v3 native-plan repository-memory scaffold.
var MemoryAgentsMD = memoryInstructionDocument("AGENTS")

// MemoryClaudeMD is the v3 native-plan repository-memory scaffold.
var MemoryClaudeMD = memoryInstructionDocument("CLAUDE")

// MemoryCopilotInstructionsMD is the v3 repository-wide Copilot scaffold.
var MemoryCopilotInstructionsMD = memoryCopilotInstructions

// InstructionFile returns default scaffold content for supported instruction file paths.
func InstructionFile(path string) string {
	return InstructionFileForVersion(path, config.DefaultInstructionScaffoldVersion)
}

// InstructionFileForVersion returns scaffold content for supported instruction file paths.
func InstructionFileForVersion(path string, version int) string {
	cleanPath := strings.ReplaceAll(path, "\\", "/")
	useLegacy := version == config.InstructionScaffoldVersionVerbose
	useMemory := version == config.InstructionScaffoldVersionMemory

	switch {
	case strings.HasSuffix(cleanPath, "/AGENTS.md") || cleanPath == "AGENTS.md":
		if useLegacy {
			return LegacyAgentsMD
		}
		if useMemory {
			return MemoryAgentsMD
		}
		return AgentsMD
	case strings.HasSuffix(cleanPath, "/CLAUDE.md") || cleanPath == "CLAUDE.md":
		if useLegacy {
			return LegacyClaudeMD
		}
		if useMemory {
			return MemoryClaudeMD
		}
		return ClaudeMD
	case strings.HasSuffix(cleanPath, "/copilot-instructions.md") || cleanPath == "copilot-instructions.md":
		if useLegacy {
			return LegacyCopilotInstructionsMD
		}
		if useMemory {
			return MemoryCopilotInstructionsMD
		}
		return CopilotInstructionsMD
	default:
		base := cleanPath
		if idx := strings.LastIndex(cleanPath, "/"); idx >= 0 {
			base = cleanPath[idx+1:]
		}
		if useLegacy {
			return LegacyAgentPointer(strings.TrimSuffix(base, ".md"))
		}
		if useMemory {
			return memoryInstructionDocument(strings.TrimSuffix(base, ".md"))
		}
		return AgentPointer(strings.TrimSuffix(base, ".md"))
	}
}

func InstructionSupportFiles(version int) []ScaffoldFile {
	files := make([]ScaffoldFile, 0, len(instructions.SupportDocs(version)))
	for _, doc := range instructions.SupportDocs(version) {
		files = append(files, ScaffoldFile{
			RelativePath: doc.RelativePath,
			Content:      instructionSupportContent(doc.RelativePath, version),
		})
	}

	return files
}

func instructionSupportContent(relativePath string, version int) string {
	if version == config.InstructionScaffoldVersionMemory {
		if content := memoryInstructionSupportContent(relativePath); content != "" {
			return content
		}
	}
	switch relativePath {
	case "docs/agents/README.md":
		return agentsREADME
	case "docs/agents/WORKFLOWS.md":
		return agentsWorkflows
	case "docs/agents/RLM.md":
		return agentsRLM
	case "docs/agents/TOOLING.md":
		return agentsTooling
	case "docs/agents/GUARDRAILS.md":
		return agentsGuardrails
	case "docs/references/README.md":
		return referencesREADME
	case "docs/references/testing.md":
		return referencesTesting
	case "docs/references/tooling.md":
		return referencesTooling
	case "docs/references/external-systems.md":
		return referencesExternalSystems
	default:
		return ""
	}
}
