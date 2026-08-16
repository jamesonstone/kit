package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
	"github.com/jamesonstone/kit/internal/templates"
)

func auditInstructionFiles(projectRoot string, cfg *config.Config) []reconcileFinding {
	var findings []reconcileFinding
	version := detectInstructionScaffoldVersion(projectRoot, cfg)
	if version == instructionScaffoldVersionUnknown {
		version = config.DefaultInstructionScaffoldVersion
	}

	for _, relativePath := range instructionFiles(cfg) {
		plan, err := planInstructionFileWrite(
			projectRoot,
			relativePath,
			instructionFileWriteModeAppendOnly,
			version,
		)
		absolutePath := filepath.Join(projectRoot, relativePath)
		if err != nil {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"repository instruction file drift cannot be reconciled safely with append-only planning",
				templateSource(projectRoot),
				fmt.Sprintf("inspect the file manually or preview a targeted replacement with `kit reconcile --include-files --force --dry-run --diff --file %s`", relativePath),
				[]string{
					fmt.Sprintf("sed -n '1,240p' %s", absolutePath),
					fmt.Sprintf("sed -n '1,240p' %s", templateSource(projectRoot)),
				},
			))
			continue
		}

		switch plan.result {
		case instructionFileCreated:
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"missing Kit-managed repository instruction file",
				templateSource(projectRoot),
				fmt.Sprintf("preview creation with `kit reconcile --include-files --dry-run --diff --file %s`, then apply only after review", relativePath),
				[]string{fmt.Sprintf("kit reconcile --include-files --dry-run --diff --file %s", relativePath)},
			))
		case instructionFileMerged:
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"repository instruction file is missing current Kit-managed sections",
				templateSource(projectRoot),
				fmt.Sprintf("preview the missing managed sections with `kit reconcile --include-files --dry-run --diff --file %s`, then apply only after review", relativePath),
				[]string{
					fmt.Sprintf("kit reconcile --include-files --dry-run --diff --file %s", relativePath),
					fmt.Sprintf("sed -n '1,240p' %s", absolutePath),
				},
			))
		}
	}

	for _, support := range instructions.SupportDocs(version) {
		absolutePath := filepath.Join(projectRoot, support.RelativePath)
		exists := document.Exists(absolutePath)
		switch version {
		case config.InstructionScaffoldVersionTOC, config.InstructionScaffoldVersionMemory:
			if exists {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"missing repo-local instruction support document",
				templateSource(projectRoot),
				fmt.Sprintf("preview restoration with `kit reconcile --include-files --dry-run --diff --file %s`, using `--force` only after reviewing customized content", support.RelativePath),
				[]string{
					fmt.Sprintf("kit reconcile --include-files --dry-run --diff --file %s", support.RelativePath),
					fmt.Sprintf("kit reconcile --include-files --force --dry-run --diff --file %s", support.RelativePath),
				},
			))
		case config.InstructionScaffoldVersionVerbose:
			if !exists {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"v2 docs-tree artifact is present in a version 1 instruction model",
				templateSource(projectRoot),
				"review and remove the leftover V2 support artifact only when it has no project-owned content",
				[]string{
					fmt.Sprintf("sed -n '1,240p' %s", absolutePath),
				},
			))
		}
	}

	if config.UsesInstructionSupportDocs(version) {
		findings = append(findings, auditInstructionEntrypoints(projectRoot, instructionFileSet(instructionFiles(cfg)), version)...)
		if version == config.InstructionScaffoldVersionMemory {
			findings = append(findings, auditV3SupportGuidance(projectRoot)...)
		} else {
			findings = append(findings, auditV2SupportGuidance(projectRoot)...)
		}
		findings = append(findings, auditInstructionPromptEntrypoints(projectRoot, cfg, version)...)
		findings = append(findings, auditAlwaysLoadedCoreDocs(projectRoot)...)
	}
	findings = append(findings, auditWorkLaneShorthandGuidance(projectRoot)...)
	if version == config.InstructionScaffoldVersionTOC && !exactGeneratedInstructionScaffold(projectRoot, cfg, version) {
		finding := newFinding(
			reconcileSeverityWarning,
			filepath.Join(projectRoot, config.ConfigFileName),
			"customized V2 instruction artifacts are not eligible for automatic V3 migration",
			templateSource(projectRoot),
			"review `kit reconcile --include-files --force --dry-run --diff`; Kit will not overwrite customized V2 instructions automatically",
			[]string{"kit reconcile --include-files --dry-run --diff", "kit reconcile --include-files --force --dry-run --diff"},
		)
		finding.NonBlocking = true
		findings = append(findings, finding)
	}

	return findings
}

const (
	rootInstructionMinimumMaxLines             = 100
	rootInstructionCustomizationAllowanceLines = 20
)

var v2RequiredRootInstructionPaths = []string{
	instructions.AgentsMDPath,
	instructions.ClaudeMDPath,
	instructions.CopilotInstructionsPath,
}

var v2ManualDuplicateSnippets = []string{
	"## Workflow: Plan",
	"## Quality gate policy",
	"## Code Style Standards",
	"## Architecture & Structure",
	"## State Summarization",
	"### Phase 1: PLAN",
	"### Phase 2: ACT",
	"### Phase 3: REFLECT",
}

var vendorToolRequirementSnippets = []string{
	"must use claude",
	"must use copilot",
	"must use codex",
	"requires claude",
	"requires copilot",
	"requires codex",
	"only use claude",
	"only use copilot",
	"only use codex",
}

func auditInstructionEntrypoints(projectRoot string, alreadyAudited map[string]bool, version int) []reconcileFinding {
	var findings []reconcileFinding
	model := fmt.Sprintf("version %d", version)
	for _, relativePath := range v2RequiredRootInstructionPaths {
		scaffoldCommand := fmt.Sprintf("kit reconcile --include-files --dry-run --diff --file %s", relativePath)
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			if os.IsNotExist(err) {
				if alreadyAudited[relativePath] {
					continue
				}
				findings = append(findings, newFinding(
					reconcileSeverityWarning,
					absolutePath,
					fmt.Sprintf("missing %s root instruction entrypoint", model),
					templateSource(projectRoot),
					fmt.Sprintf("restore the thin root files with `%s`", scaffoldCommand),
					[]string{scaffoldCommand},
				))
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("failed to read %s root instruction entrypoint", model),
				templateSource(projectRoot),
				"fix file readability before project validation can inspect instruction drift",
				[]string{fmt.Sprintf("sed -n '1,160p' %s", absolutePath)},
			))
			continue
		}

		body := string(content)
		if !strings.Contains(body, "docs/agents/README.md") {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("%s root instruction file does not route through `docs/agents/README.md`", model),
				templateSource(projectRoot),
				fmt.Sprintf("restore the thin routing entrypoint with `%s` or `--force` if a full refresh is acceptable", scaffoldCommand),
				[]string{
					scaffoldCommand,
					fmt.Sprintf("rg -n \"docs/agents/README.md\" %s", absolutePath),
				},
			))
		}
		if countLines(body) > rootInstructionMaxLines(relativePath, version) || containsAny(body, v2ManualDuplicateSnippets) {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("%s root instruction file duplicates the full workflow manual instead of staying thin", model),
				templateSource(projectRoot),
				"move durable workflow guidance to `docs/agents/*` and keep the root file as a routing table",
				[]string{
					fmt.Sprintf("wc -l %s", absolutePath),
					fmt.Sprintf("sed -n '1,180p' %s", absolutePath),
				},
			))
		}
		if containsVendorToolRequirement(body) {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("%s root instruction file requires a vendor-specific coding tool", model),
				constitutionSource(projectRoot),
				"rewrite the instruction as agent-agnostic guidance and keep vendor-specific files as optional entrypoints only",
				[]string{fmt.Sprintf("sed -n '1,160p' %s", absolutePath)},
			))
		}
		if strings.Contains(strings.ToLower(body), "core.md") {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("%s root instruction file references an unsupported always-loaded `core.md`", model),
				templateSource(projectRoot),
				"remove the monolithic core reference and route through `docs/agents/README.md` instead",
				[]string{fmt.Sprintf("rg -n \"core\\.md|docs/agents/README\\.md\" %s", absolutePath)},
			))
		}
	}

	return findings
}

func rootInstructionMaxLines(relativePath string, version int) int {
	generatedLines := countLines(templates.InstructionFileForVersion(relativePath, version)) +
		rootInstructionCustomizationAllowanceLines
	if generatedLines > rootInstructionMinimumMaxLines {
		return generatedLines
	}
	return rootInstructionMinimumMaxLines
}

func instructionFileSet(paths []string) map[string]bool {
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		set[filepath.ToSlash(path)] = true
	}
	return set
}
