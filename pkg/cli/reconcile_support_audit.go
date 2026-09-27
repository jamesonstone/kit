package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
)

func auditV2SupportGuidance(projectRoot string) []reconcileFinding {
	expectations := v2GuidanceExpectations()

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
				fmt.Sprintf(
					"integrate the missing V2 guidance manually, or preview a targeted generated replacement with `kit reconcile --include-files --force --dry-run --diff --file %s` before overwriting customized content",
					relativePath,
				),
				[]string{
					fmt.Sprintf("kit reconcile --include-files --force --dry-run --diff --file %s", relativePath),
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
	expectations := v3GuidanceExpectations()
	forbidden := v3ForbiddenGuidance()
	paths := make(map[string]bool, len(expectations)+len(forbidden))
	for relativePath := range expectations {
		paths[relativePath] = true
	}
	for relativePath := range forbidden {
		paths[relativePath] = true
	}

	var findings []reconcileFinding
	for relativePath := range paths {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			continue
		}
		body := string(content)
		for _, snippet := range expectations[relativePath] {
			if strings.Contains(body, snippet) {
				continue
			}
			findings = append(findings, v3GuidanceFinding(projectRoot, absolutePath, relativePath,
				fmt.Sprintf("V3 instruction support document is missing required guidance %q", snippet),
				"integrate the missing V3 guidance manually", snippet))
			break
		}
		for _, snippet := range forbidden[relativePath] {
			if !strings.Contains(body, snippet) {
				continue
			}
			findings = append(findings, v3GuidanceFinding(projectRoot, absolutePath, relativePath,
				fmt.Sprintf("V3 instruction support document still contains forbidden guidance %q", snippet),
				"remove the superseded guidance", snippet))
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

func v3GuidanceFinding(projectRoot, absolutePath, relativePath, issue, action, snippet string) reconcileFinding {
	return newFinding(
		reconcileSeverityWarning,
		absolutePath,
		issue,
		templateSource(projectRoot),
		fmt.Sprintf(
			"%s, or preview a targeted generated replacement with `kit reconcile --include-files --force --dry-run --diff --file %s` before overwriting customized content",
			action,
			relativePath,
		),
		[]string{
			fmt.Sprintf("kit reconcile --include-files --force --dry-run --diff --file %s", relativePath),
			fmt.Sprintf("rg -n %q %s", snippet, absolutePath),
		},
	)
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
		[]string{"kit reconcile --include-files --dry-run --diff --file docs/agents/README.md"},
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
