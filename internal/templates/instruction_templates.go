package templates

import (
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
)

const sharedRepositoryInstructions = sharedRepositoryInstructionsCore + sharedRepositoryInstructionsStandards

const copilotQuickRules = `## Fast rules for chat and code review

- classify every request first
  - **spec-driven**: ` + "`kit brainstorm`" + ` / ` + "`kit spec`" + ` work, new capability, substantial behavioral or architectural change, existing spec-covered feature work, or cross-component/public-interface changes
  - **ad hoc**: contained bug fix, review, refactor, dependency update, config change, or small refinement
- for spec-driven work:
  - read ` + "`BRAINSTORM.md`" + ` when present, then ` + "`SPEC.md`" + ` → ` + "`PLAN.md`" + ` → ` + "`TASKS.md`" + `
  - ask numbered clarification questions until you reach ≥95% confidence
  - include a recommended default, proposed solution, or assumption for every question
  - accept approvals via ` + "`yes`" + ` / ` + "`y`" + `, partial approvals via ` + "`yes 3, 4`" + `, and overrides via ` + "`no 2: <answer>`" + `
  - run the implementation readiness gate before writing code; if it fails, update docs first
  - implement tasks in order and update docs first if implementation changes behavior, requirements, or approach
- for ad hoc work:
  - follow understand → implement → verify
  - update existing spec docs when the change alters behavior, requirements, or approach
- always:
  - never mix multiple features in one ` + "`docs/specs/<feature>/`" + ` directory
  - keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with canonical docs
  - populate every section in ` + "`BRAINSTORM.md`" + `, ` + "`SPEC.md`" + `, ` + "`PLAN.md`" + `, and ` + "`TASKS.md`" + `; replace placeholder-only sections with ` + "`not applicable`" + `, ` + "`not required`" + `, or ` + "`no additional information required`" + ` instead of leaving HTML TODO comments
  - when creating a ` + "`git worktree`" + `, use the flat ` + "`~/worktrees/`" + ` root with repo-prefixed leaf directories such as ` + "`~/worktrees/<repo>-<branch>`" + `
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

// InstructionFile returns default scaffold content for supported instruction file paths.
func InstructionFile(path string) string {
	return InstructionFileForVersion(path, config.DefaultInstructionScaffoldVersion)
}

// InstructionFileForVersion returns scaffold content for supported instruction file paths.
func InstructionFileForVersion(path string, version int) string {
	cleanPath := strings.ReplaceAll(path, "\\", "/")
	useLegacy := version == config.InstructionScaffoldVersionVerbose

	switch {
	case strings.HasSuffix(cleanPath, "/AGENTS.md") || cleanPath == "AGENTS.md":
		if useLegacy {
			return LegacyAgentsMD
		}
		return AgentsMD
	case strings.HasSuffix(cleanPath, "/CLAUDE.md") || cleanPath == "CLAUDE.md":
		if useLegacy {
			return LegacyClaudeMD
		}
		return ClaudeMD
	case strings.HasSuffix(cleanPath, "/copilot-instructions.md") || cleanPath == "copilot-instructions.md":
		if useLegacy {
			return LegacyCopilotInstructionsMD
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
		return AgentPointer(strings.TrimSuffix(base, ".md"))
	}
}

func InstructionSupportFiles(version int) []ScaffoldFile {
	files := make([]ScaffoldFile, 0, len(instructions.SupportDocs(version)))
	for _, doc := range instructions.SupportDocs(version) {
		files = append(files, ScaffoldFile{
			RelativePath: doc.RelativePath,
			Content:      instructionSupportContent(doc.RelativePath),
		})
	}

	return files
}

func instructionSupportContent(relativePath string) string {
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
