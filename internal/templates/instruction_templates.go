package templates

import "strings"

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

func repositoryInstructionDocument(title string) string {
	return `# ` + title + `

` + sharedRepositoryInstructions
}

// AgentPointer returns the comprehensive instruction template for agent files.
func AgentPointer(agentName string) string {
	return repositoryInstructionDocument(agentName)
}

// AgentsMD is the comprehensive AGENTS.md template with full workflow and standards.
var AgentsMD = repositoryInstructionDocument("AGENTS")

// ClaudeMD is the comprehensive CLAUDE.md template with full workflow and standards.
var ClaudeMD = repositoryInstructionDocument("CLAUDE")

// CopilotInstructionsMD is the comprehensive repository-wide Copilot instructions template.
var CopilotInstructionsMD = `# GitHub Copilot Repository Instructions

` + copilotQuickRules + `
` + sharedRepositoryInstructions

// InstructionFile returns scaffold content for supported instruction file paths.
func InstructionFile(path string) string {
	cleanPath := strings.ReplaceAll(path, "\\", "/")

	switch {
	case strings.HasSuffix(cleanPath, "/AGENTS.md") || cleanPath == "AGENTS.md":
		return AgentsMD
	case strings.HasSuffix(cleanPath, "/CLAUDE.md") || cleanPath == "CLAUDE.md":
		return ClaudeMD
	case strings.HasSuffix(cleanPath, "/copilot-instructions.md") || cleanPath == "copilot-instructions.md":
		return CopilotInstructionsMD
	default:
		base := cleanPath
		if idx := strings.LastIndex(cleanPath, "/"); idx >= 0 {
			base = cleanPath[idx+1:]
		}
		return AgentPointer(strings.TrimSuffix(base, ".md"))
	}
}
