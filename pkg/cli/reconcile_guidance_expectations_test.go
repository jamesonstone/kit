package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestReconcileGuidanceExpectationsMatchCurrentTemplates(t *testing.T) {
	tests := []struct {
		name         string
		version      int
		expectations map[string][]string
		audit        func(string) []reconcileFinding
	}{
		{
			name:         "V2",
			version:      config.InstructionScaffoldVersionTOC,
			expectations: v2GuidanceExpectations(),
			audit:        auditV2SupportGuidance,
		},
		{
			name:         "V3",
			version:      config.InstructionScaffoldVersionMemory,
			expectations: v3GuidanceExpectations(),
			audit:        auditV3SupportGuidance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := writeCurrentReconcileGuidanceFixture(t, tt.version)
			for relativePath, snippets := range tt.expectations {
				content := readFile(t, filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
				for _, snippet := range snippets {
					if !strings.Contains(content, snippet) {
						t.Fatalf("%s does not contain required guidance %q", relativePath, snippet)
					}
				}
			}

			if findings := tt.audit(projectRoot); len(findings) != 0 {
				t.Fatalf("current templates produced reconcile findings: %#v", findings)
			}
		})
	}
}

func TestAuditV2SupportGuidanceFindsStaleTestingSemantics(t *testing.T) {
	tests := []struct {
		path    string
		snippet string
	}{
		{
			path:    "docs/agents/RLM.md",
			snippet: "`docs/references/rules/testing-and-environment-validation.md`",
		},
		{
			path:    "docs/references/README.md",
			snippet: "`rules/testing-and-environment-validation.md`",
		},
		{
			path:    "docs/references/testing.md",
			snippet: "## High-Level Suites",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			projectRoot := writeCurrentReconcileGuidanceFixture(
				t,
				config.InstructionScaffoldVersionTOC,
			)
			removeGuidanceSnippet(t, projectRoot, tt.path, tt.snippet)

			assertStaleGuidanceFinding(
				t,
				projectRoot,
				tt.path,
				tt.snippet,
				auditV2SupportGuidance(projectRoot),
			)
		})
	}
}

func TestAuditV3SupportGuidanceFindsStaleSessionBrowserTestingAndWorktreeSemantics(t *testing.T) {
	tests := []struct {
		path    string
		snippet string
	}{
		{
			path:    "AGENTS.md",
			snippet: "First, call the available thread-title operation (`set_thread_title` when available)",
		},
		{
			path:    "AGENTS.md",
			snippet: "For interactive browser work, use Codex's built-in browser through `@Browser`.",
		},
		{
			path:    "AGENTS.md",
			snippet: "unless I explicitly request it.",
		},
		{
			path:    "AGENTS.md",
			snippet: "When I explicitly authorize an external browser, terminate and verify all",
		},
		{
			path:    ".github/copilot-instructions.md",
			snippet: "Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md`",
		},
		{
			path:    "docs/agents/RLM.md",
			snippet: "`docs/references/rules/testing-and-environment-validation.md`",
		},
		{
			path:    "docs/agents/TOOLING.md",
			snippet: "Link the primary checkout's `.env` and `.envrc` into writable lanes by default",
		},
		{
			path:    "docs/references/README.md",
			snippet: "`worktrees.md` for the canonical native Git worktree hierarchy",
		},
		{
			path:    "docs/references/testing.md",
			snippet: "## High-Level Suites",
		},
		{
			path:    "docs/references/worktrees.md",
			snippet: "The `PR#` column runs one batched `gh` lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			projectRoot := writeCurrentReconcileGuidanceFixture(
				t,
				config.InstructionScaffoldVersionMemory,
			)
			removeGuidanceSnippet(t, projectRoot, tt.path, tt.snippet)

			assertStaleGuidanceFinding(
				t,
				projectRoot,
				tt.path,
				tt.snippet,
				auditV3SupportGuidance(projectRoot),
			)
		})
	}
}

func TestReconcileCleanResultReportsPendingManagedFileDryRun(t *testing.T) {
	report := &reconcileReport{}

	got := reconcileCleanResult(report, true, true, 2)
	for _, want := range []string{
		"Managed-file refresh pending for 2 files.",
		"The semantic documentation audit is clean for this scope.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("reconcileCleanResult() = %q, want %q", got, want)
		}
	}
	if strings.Contains(got, "No reconciliation needed") {
		t.Fatalf("pending refresh must not report unqualified clean result: %q", got)
	}

	if got := reconcileCleanResult(report, true, true, 0); got != report.cleanResult() {
		t.Fatalf("zero-change dry run = %q, want %q", got, report.cleanResult())
	}
}

func TestReconcileCleanResultIncludesSourceAuditDuringPendingDryRun(t *testing.T) {
	report := &reconcileReport{
		SourceFileAudit: &sourceFileAuditSummary{
			CandidateCount: 12,
			EligibleCount:  7,
			Complete:       true,
		},
	}

	got := reconcileCleanResult(report, true, true, 1)
	for _, want := range []string{
		"Managed-file refresh pending for 1 file.",
		"source-file-size audit: complete",
		"7 eligible handwritten source/test files checked",
		"0 above 300 physical lines",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("reconcileCleanResult() = %q, want %q", got, want)
		}
	}
}

func writeCurrentReconcileGuidanceFixture(t *testing.T, version int) string {
	t.Helper()

	projectRoot := t.TempDir()
	for _, support := range templates.InstructionSupportFiles(version) {
		writeFile(t, filepath.Join(projectRoot, filepath.FromSlash(support.RelativePath)), support.Content)
	}
	if version == config.InstructionScaffoldVersionMemory {
		writeFile(
			t,
			filepath.Join(projectRoot, "AGENTS.md"),
			templates.MemoryAgentsMD,
		)
		writeFile(
			t,
			filepath.Join(projectRoot, ".github", "copilot-instructions.md"),
			templates.MemoryCopilotInstructionsMD,
		)
	}
	return projectRoot
}

func removeGuidanceSnippet(
	t *testing.T,
	projectRoot string,
	relativePath string,
	snippet string,
) {
	t.Helper()

	absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	content := readFile(t, absolutePath)
	if count := strings.Count(content, snippet); count != 1 {
		t.Fatalf("%s contains %q %d times, want exactly once", relativePath, snippet, count)
	}
	writeFile(t, absolutePath, strings.Replace(content, snippet, "", 1))
}

func assertStaleGuidanceFinding(
	t *testing.T,
	projectRoot string,
	relativePath string,
	snippet string,
	findings []reconcileFinding,
) {
	t.Helper()

	absolutePath := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
	for _, finding := range findings {
		if finding.FilePath != absolutePath {
			continue
		}
		if !strings.Contains(finding.Issue, snippet) {
			t.Fatalf("finding issue = %q, want missing snippet %q", finding.Issue, snippet)
		}
		preview := "kit reconcile --include-files --force --dry-run --diff --file " + relativePath
		if !strings.Contains(finding.UpdateInstruction, preview) {
			t.Fatalf("update instruction = %q, want targeted preview %q", finding.UpdateInstruction, preview)
		}
		if strings.Contains(finding.UpdateInstruction, "--append-only") {
			t.Fatalf("existing-section drift must not recommend append-only refresh: %q", finding.UpdateInstruction)
		}
		return
	}

	t.Fatalf("no reconcile finding for stale guidance in %s: %#v", relativePath, findings)
}
