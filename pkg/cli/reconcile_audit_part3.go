package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
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
				"inspect the file manually and add the missing Kit-managed sections, or use `kit scaffold agents --force` only if overwrite is acceptable",
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
				"prefer `kit scaffold agents --append-only` to create the missing file without replacing existing instruction files",
				[]string{"kit scaffold agents --append-only"},
			))
		case instructionFileMerged:
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"repository instruction file is missing current Kit-managed sections",
				templateSource(projectRoot),
				"prefer `kit scaffold agents --append-only` to append the missing Kit-managed sections, then review the result",
				[]string{
					"kit scaffold agents --append-only",
					fmt.Sprintf("sed -n '1,240p' %s", absolutePath),
				},
			))
		}
	}

	for _, support := range instructions.SupportDocs(config.InstructionScaffoldVersionTOC) {
		absolutePath := filepath.Join(projectRoot, support.RelativePath)
		exists := document.Exists(absolutePath)
		switch version {
		case config.InstructionScaffoldVersionTOC:
			if exists {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"missing v2 repo-local instruction support document",
				templateSource(projectRoot),
				"restore the thin ToC docs tree, typically with `kit scaffold agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold agents --version 2 --append-only",
					"kit scaffold agents --version 2 --force",
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
				"remove the leftover v2 docs-tree artifact or rerun `kit scaffold agents --version 1 --force` to finish the downgrade",
				[]string{
					"kit scaffold agents --version 1 --force",
					fmt.Sprintf("sed -n '1,240p' %s", absolutePath),
				},
			))
		}
	}

	if version == config.InstructionScaffoldVersionTOC {
		findings = append(findings, auditV2InstructionEntrypoints(projectRoot, instructionFileSet(instructionFiles(cfg)))...)
		findings = append(findings, auditV2SupportGuidance(projectRoot)...)
		findings = append(findings, auditV2PromptEntrypoints(projectRoot, cfg)...)
		findings = append(findings, auditAlwaysLoadedCoreDocs(projectRoot)...)
	}

	return findings
}

const v2RootInstructionMaxLines = 90

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

func auditV2InstructionEntrypoints(projectRoot string, alreadyAudited map[string]bool) []reconcileFinding {
	var findings []reconcileFinding
	for _, relativePath := range v2RequiredRootInstructionPaths {
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
					"missing v2 root instruction entrypoint",
					templateSource(projectRoot),
					"restore the thin v2 root files with `kit scaffold agents --version 2 --append-only`",
					[]string{"kit scaffold agents --version 2 --append-only"},
				))
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"failed to read v2 root instruction entrypoint",
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
				"v2 root instruction file does not route through `docs/agents/README.md`",
				templateSource(projectRoot),
				"restore the thin routing entrypoint with `kit scaffold agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold agents --version 2 --append-only",
					fmt.Sprintf("rg -n \"docs/agents/README.md\" %s", absolutePath),
				},
			))
		}
		if countLines(body) > v2RootInstructionMaxLines || containsAny(body, v2ManualDuplicateSnippets) {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"v2 root instruction file duplicates the full workflow manual instead of staying thin",
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
				"v2 root instruction file requires a vendor-specific coding tool",
				constitutionSource(projectRoot),
				"rewrite the instruction as agent-agnostic guidance and keep vendor-specific files as optional entrypoints only",
				[]string{fmt.Sprintf("sed -n '1,160p' %s", absolutePath)},
			))
		}
		if strings.Contains(strings.ToLower(body), "core.md") {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"v2 root instruction file references an unsupported always-loaded `core.md`",
				templateSource(projectRoot),
				"remove the monolithic core reference and route through `docs/agents/README.md` instead",
				[]string{fmt.Sprintf("rg -n \"core\\.md|docs/agents/README\\.md\" %s", absolutePath)},
			))
		}
	}

	return findings
}
