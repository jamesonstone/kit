package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func auditV2SupportGuidance(projectRoot string) []reconcileFinding {
	expectations := map[string][]string{
		"docs/agents/README.md": {
			"## Runtime Routing",
			"load only the linked doc needed for the current decision",
			"Stop loading once the decision is supported",
		},
		"docs/agents/RLM.md": {
			"## Runtime Loop",
			"identify the immediate decision",
			"load the smallest relevant artifact",
			"stop loading once the decision is supported",
			"## Context Budget Rules",
			"specific section over full file",
			"repo-local docs before global model/vendor instructions",
		},
		"docs/agents/WORKFLOWS.md": {
			"Authority order:",
			"Execution order for feature work:",
			"`SPEC.md` controls requirements, plan, tasks, validation, reflection, delivery, and evidence",
			"`BRAINSTORM.md`, `PLAN.md`, and `TASKS.md` are non-binding historical context in v2",
		},
		"docs/agents/GUARDRAILS.md": {
			"Never claim tests passed unless they ran",
			"Never claim files were inspected unless they were inspected",
			"If validation cannot run, state why",
		},
	}

	var findings []reconcileFinding
	for relativePath, snippets := range expectations {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			continue
		}
		body := string(content)
		for _, snippet := range snippets {
			if strings.Contains(body, snippet) {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("v2 instruction support document is missing required guidance %q", snippet),
				templateSource(projectRoot),
				"refresh the v2 docs tree with `kit scaffold agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold agents --version 2 --append-only",
					fmt.Sprintf("rg -n %q %s", snippet, absolutePath),
				},
			))
			break
		}
		if containsVendorToolRequirement(body) {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"v2 instruction support document requires a vendor-specific coding tool",
				constitutionSource(projectRoot),
				"rewrite the guidance as agent-agnostic instructions",
				[]string{fmt.Sprintf("sed -n '1,180p' %s", absolutePath)},
			))
		}
	}

	return findings
}

func auditV3SupportGuidance(projectRoot string) []reconcileFinding {
	expectations := map[string][]string{
		"docs/agents/README.md": {
			"## Runtime Routing",
			"Native agent planning owns research, clarification, design, and plan formation",
			"V1 and V2 artifacts remain supported legacy inputs",
		},
		"docs/agents/WORKFLOWS.md": {
			"## Native Planning To Repository Memory",
			"Before code, assess whether the work contains material rationale",
			"Never mechanically rewrite a V2 spec into V3",
			"`kit dispatch` supports post-plan execution topology",
		},
		"docs/agents/RLM.md": {
			"## Runtime Loop",
			"identify the immediate decision",
			"stop loading once the decision is supported",
			"## Context Budget Rules",
		},
		"docs/agents/TOOLING.md": {
			"Use `kit dispatch` after native planning",
			"accepted plan needs a safe multi-lane execution topology",
		},
		"docs/agents/GUARDRAILS.md": {
			"## Repository Memory Completion Gate",
			"Create or adopt a spec before code when material rationale exists",
			"Every implementation final response must include `Repository Memory`",
		},
	}

	var findings []reconcileFinding
	for relativePath, snippets := range expectations {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			continue
		}
		body := string(content)
		for _, snippet := range snippets {
			if strings.Contains(body, snippet) {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("V3 instruction support document is missing required guidance %q", snippet),
				templateSource(projectRoot),
				"refresh the V3 docs tree with `kit scaffold agents --version 3 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold agents --version 3 --append-only",
					fmt.Sprintf("rg -n %q %s", snippet, absolutePath),
				},
			))
			break
		}
		if containsVendorToolRequirement(body) {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"V3 instruction support document requires a vendor-specific coding tool",
				constitutionSource(projectRoot),
				"rewrite the guidance as agent-agnostic instructions",
				[]string{fmt.Sprintf("sed -n '1,180p' %s", absolutePath)},
			))
		}
	}
	return findings
}

func auditInstructionPromptEntrypoints(projectRoot string, cfg *config.Config, version int) []reconcileFinding {
	if repoKnowledgeEntrypointPath(projectRoot, cfg) != "" {
		return nil
	}

	path := filepath.Join(projectRoot, "docs", "agents", "README.md")
	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		path,
		fmt.Sprintf("generated prompt routing cannot find the version %d repo-local entrypoint", version),
		templateSource(projectRoot),
		"restore `docs/agents/README.md` so prompts can use just-in-time context loading",
		[]string{fmt.Sprintf("kit scaffold agents --version %d --append-only", version)},
	)}
}

func auditAlwaysLoadedCoreDocs(projectRoot string) []reconcileFinding {
	var findings []reconcileFinding
	for _, relativePath := range []string{
		"docs/agents/core.md",
		"docs/agents/CORE.md",
	} {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		if !document.Exists(absolutePath) {
			continue
		}
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			absolutePath,
			"unsupported always-loaded monolithic instruction file exists",
			templateSource(projectRoot),
			"remove the monolithic instruction file and route agents through `docs/agents/README.md` plus just-in-time linked docs",
			[]string{fmt.Sprintf("sed -n '1,180p' %s", absolutePath)},
		))
	}

	return findings
}

func countLines(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

func containsAny(content string, snippets []string) bool {
	for _, snippet := range snippets {
		if strings.Contains(content, snippet) {
			return true
		}
	}
	return false
}

func containsVendorToolRequirement(content string) bool {
	lower := strings.ToLower(content)
	for _, snippet := range vendorToolRequirementSnippets {
		if strings.Contains(lower, snippet) {
			return true
		}
	}
	return false
}
