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
				fmt.Sprintf(
					"integrate the missing V3 guidance manually, or preview a targeted generated replacement with `kit reconcile --include-files --force --dry-run --diff --file %s` before overwriting customized content",
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
