package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/instructions"
)

type reconcileSeverity string

const (
	reconcileSeverityError   reconcileSeverity = "error"
	reconcileSeverityWarning reconcileSeverity = "warning"
)

type reconcileFinding struct {
	Severity          reconcileSeverity
	FilePath          string
	Issue             string
	ContractSource    string
	UpdateInstruction string
	SearchHints       []string
}

type reconcileReport struct {
	ProjectRoot string
	Feature     *feature.Feature
	Findings    []reconcileFinding
	NeedsRollup bool
}

func (r *reconcileReport) cleanResult() string {
	if r.Feature != nil {
		return fmt.Sprintf("No reconciliation needed for feature %s.", r.Feature.Slug)
	}
	return "No reconciliation needed. Kit-managed documents already match the current contract for this scope."
}

func buildReconcileReport(projectRoot string, cfg *config.Config, feat *feature.Feature) (*reconcileReport, error) {
	report := &reconcileReport{
		ProjectRoot: projectRoot,
		Feature:     feat,
	}

	features, err := feature.ListFeaturesWithState(cfg.SpecsPath(projectRoot), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}

	targets := make(map[string]bool, len(features))
	for _, item := range features {
		targets[item.DirName] = true
	}

	if feat == nil {
		report.Findings = append(report.Findings, auditDuplicateFeatureNumbers(cfg.SpecsPath(projectRoot), projectRoot, features)...)
		report.Findings = append(report.Findings, auditConstitution(projectRoot)...)
		report.Findings = append(report.Findings, auditProjectProgressSummary(projectRoot, features)...)
		for i := range features {
			report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, &features[i], targets)...)
		}
		report.Findings = append(report.Findings, auditInstructionFiles(projectRoot, cfg)...)
	} else {
		report.Findings = append(report.Findings, auditFeatureDocuments(projectRoot, feat, targets)...)
		report.Findings = append(report.Findings, auditFeatureRollupCoverage(projectRoot, feat)...)
	}

	for _, finding := range report.Findings {
		if strings.Contains(finding.UpdateInstruction, "`kit rollup`") {
			report.NeedsRollup = true
			break
		}
	}

	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].Severity != report.Findings[j].Severity {
			return report.Findings[i].Severity < report.Findings[j].Severity
		}
		if report.Findings[i].FilePath != report.Findings[j].FilePath {
			return report.Findings[i].FilePath < report.Findings[j].FilePath
		}
		return report.Findings[i].Issue < report.Findings[j].Issue
	})

	return report, nil
}

func auditConstitution(projectRoot string) []reconcileFinding {
	path := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	if !document.Exists(path) {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			"missing Kit-managed root document `CONSTITUTION.md`",
			templateSource(projectRoot),
			"create `docs/CONSTITUTION.md` and populate the current Kit sections before reconciling feature docs",
			[]string{
				fmt.Sprintf("sed -n '1,240p' %s", templateSource(projectRoot)),
				fmt.Sprintf("sed -n '1,240p' %s", initProjectSource(projectRoot)),
			},
		)}
	}

	return auditStructuredDocument(path, document.TypeConstitution, projectRoot, nil)
}

func auditProjectProgressSummary(projectRoot string, features []feature.Feature) []reconcileFinding {
	path := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	if !document.Exists(path) {
		return []reconcileFinding{newFinding(
			reconcileSeverityError,
			path,
			"missing `PROJECT_PROGRESS_SUMMARY.md`",
			templateSource(projectRoot),
			"create the progress summary, then run `kit rollup` to bring the project summary up to date",
			[]string{
				fmt.Sprintf("sed -n '1,220p' %s", templateSource(projectRoot)),
				"kit rollup",
			},
		)}
	}

	findings := auditStructuredDocument(path, document.TypeProgressSummary, projectRoot, nil)
	content, err := os.ReadFile(path)
	if err != nil {
		return append(findings, newFinding(
			reconcileSeverityError,
			path,
			"failed to read `PROJECT_PROGRESS_SUMMARY.md`",
			templateSource(projectRoot),
			"fix file readability before reconciliation can continue",
			[]string{fmt.Sprintf("sed -n '1,240p' %s", path)},
		))
	}

	body := string(content)
	for i := range features {
		findings = append(findings, auditFeatureRollupCoverageFromContent(projectRoot, body, &features[i])...)
	}
	return findings
}

func auditFeatureRollupCoverage(projectRoot string, feat *feature.Feature) []reconcileFinding {
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil
	}
	return auditFeatureRollupCoverageFromContent(projectRoot, string(content), feat)
}

func auditDuplicateFeatureNumbers(specsPath, projectRoot string, features []feature.Feature) []reconcileFinding {
	duplicates := feature.DuplicateNumberGroups(features)
	if len(duplicates) == 0 {
		return nil
	}

	var findings []reconcileFinding
	for number, group := range duplicates {
		names := make([]string, 0, len(group))
		for _, feat := range group {
			names = append(names, feat.DirName)
		}
		findings = append(findings, newFinding(
			reconcileSeverityError,
			specsPath,
			fmt.Sprintf("feature number `%04d` is duplicated by %s", number, strings.Join(names, ", ")),
			initProjectSource(projectRoot),
			"renumber or merge the conflicting feature directories so each numeric prefix is unique across `docs/specs/`",
			[]string{
				fmt.Sprintf("ls %s", specsPath),
				fmt.Sprintf("rg -n \"^# (BRAINSTORM|SPEC|PLAN|TASKS)\" %s", specsPath),
			},
		))
	}

	return findings
}

func auditFeatureRollupCoverageFromContent(projectRoot, content string, feat *feature.Feature) []reconcileFinding {
	summaryPath := filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	rowSnippet := fmt.Sprintf("| %04d | %s |", feat.Number, feat.Slug)
	headingSnippet := fmt.Sprintf("### %s\n", feat.Slug)
	var findings []reconcileFinding

	if !strings.Contains(content, rowSnippet) {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			summaryPath,
			fmt.Sprintf("progress summary is missing the feature-table row for `%s`", feat.DirName),
			templateSource(projectRoot),
			"refresh `PROJECT_PROGRESS_SUMMARY.md` after reconciling feature docs, typically with `kit rollup`",
			[]string{
				fmt.Sprintf("rg -n \"^\\| %04d \\| %s \\|\" %s", feat.Number, feat.Slug, summaryPath),
				fmt.Sprintf("ls %s", filepath.Join(projectRoot, "docs", "specs")),
			},
		))
	}

	if !strings.Contains(content, headingSnippet) {
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			summaryPath,
			fmt.Sprintf("progress summary is missing the feature summary heading for `%s`", feat.DirName),
			templateSource(projectRoot),
			"refresh `PROJECT_PROGRESS_SUMMARY.md` after reconciliation so every current feature has a summary section",
			[]string{
				fmt.Sprintf("rg -n \"^### %s$\" %s", feat.Slug, summaryPath),
				"kit rollup",
			},
		))
	}

	return findings
}

func auditFeatureDocuments(projectRoot string, feat *feature.Feature, relationshipTargets map[string]bool) []reconcileFinding {
	paths := map[string]string{
		"brainstorm": filepath.Join(feat.Path, "BRAINSTORM.md"),
		"spec":       filepath.Join(feat.Path, "SPEC.md"),
		"plan":       filepath.Join(feat.Path, "PLAN.md"),
		"tasks":      filepath.Join(feat.Path, "TASKS.md"),
	}

	var findings []reconcileFinding
	specExists := document.Exists(paths["spec"])
	planExists := document.Exists(paths["plan"])
	tasksExists := document.Exists(paths["tasks"])

	if !specExists && (planExists || tasksExists) {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			paths["spec"],
			"missing `SPEC.md` even though later-phase feature artifacts exist",
			templateSource(projectRoot),
			"create `SPEC.md` and backfill the current feature contract before keeping later artifacts",
			genericFeatureSearchHints(projectRoot, feat, paths["spec"], "SPEC"),
		))
	}
	if !planExists && tasksExists {
		findings = append(findings, newFinding(
			reconcileSeverityError,
			paths["plan"],
			"missing `PLAN.md` even though `TASKS.md` exists",
			templateSource(projectRoot),
			"create `PLAN.md` and restore the implementation approach before keeping the task list",
			genericFeatureSearchHints(projectRoot, feat, paths["plan"], "PLAN"),
		))
	}

	if document.Exists(paths["brainstorm"]) {
		findings = append(findings, auditStructuredDocument(paths["brainstorm"], document.TypeBrainstorm, projectRoot, relationshipTargets)...)
	}
	if specExists {
		findings = append(findings, auditStructuredDocument(paths["spec"], document.TypeSpec, projectRoot, relationshipTargets)...)
	}
	if planExists {
		findings = append(findings, auditStructuredDocument(paths["plan"], document.TypePlan, projectRoot, relationshipTargets)...)
	}
	if tasksExists {
		findings = append(findings, auditStructuredDocument(paths["tasks"], document.TypeTasks, projectRoot, relationshipTargets)...)
		findings = append(findings, auditTaskAlignment(paths["tasks"], projectRoot)...)
	}

	return findings
}

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
				"inspect the file manually and add the missing Kit-managed sections, or use `kit scaffold-agents --force` only if overwrite is acceptable",
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
				"prefer `kit scaffold-agents --append-only` to create the missing file without replacing existing instruction files",
				[]string{"kit scaffold-agents --append-only"},
			))
		case instructionFileMerged:
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				"repository instruction file is missing current Kit-managed sections",
				templateSource(projectRoot),
				"prefer `kit scaffold-agents --append-only` to append the missing Kit-managed sections, then review the result",
				[]string{
					"kit scaffold-agents --append-only",
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
				"restore the thin ToC docs tree, typically with `kit scaffold-agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold-agents --version 2 --append-only",
					"kit scaffold-agents --version 2 --force",
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
				"remove the leftover v2 docs-tree artifact or rerun `kit scaffold-agents --version 1 --force` to finish the downgrade",
				[]string{
					"kit scaffold-agents --version 1 --force",
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
					"restore the thin v2 root files with `kit scaffold-agents --version 2 --append-only`",
					[]string{"kit scaffold-agents --version 2 --append-only"},
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
				"restore the thin routing entrypoint with `kit scaffold-agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold-agents --version 2 --append-only",
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

func instructionFileSet(paths []string) map[string]bool {
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		set[filepath.ToSlash(path)] = true
	}
	return set
}

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
			"`TASKS.md` controls next action",
			"`BRAINSTORM.md` is non-binding research context",
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
				"refresh the v2 docs tree with `kit scaffold-agents --version 2 --append-only` or `--force` if a full refresh is acceptable",
				[]string{
					"kit scaffold-agents --version 2 --append-only",
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

func auditV2PromptEntrypoints(projectRoot string, cfg *config.Config) []reconcileFinding {
	if repoKnowledgeEntrypointPath(projectRoot, cfg) != "" {
		return nil
	}

	path := filepath.Join(projectRoot, "docs", "agents", "README.md")
	return []reconcileFinding{newFinding(
		reconcileSeverityWarning,
		path,
		"generated prompt routing cannot find the v2 repo-local entrypoint",
		templateSource(projectRoot),
		"restore `docs/agents/README.md` so prompts can use just-in-time context loading",
		[]string{"kit scaffold-agents --version 2 --append-only"},
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
