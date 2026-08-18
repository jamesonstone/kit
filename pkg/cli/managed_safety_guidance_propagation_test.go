package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHealthAndReconcileInstallManagedSafetyGuidance(t *testing.T) {
	for _, tt := range []struct {
		name  string
		apply func(*testing.T)
	}{
		{
			name: "health",
			apply: func(t *testing.T) {
				cmd := healthCommandForTest(t, "--json")
				cmd.SetOut(io.Discard)
				if err := runHealth(cmd, nil); err != nil {
					t.Fatalf("runHealth() error = %v", err)
				}
			},
		},
		{
			name: "reconcile include files",
			apply: func(t *testing.T) {
				resetReconcileFlags(t)
				reconcileIncludeFiles = true
				reconcileOutputOnly = true

				cmd := &cobra.Command{}
				cmd.Flags().Bool("output-only", true, "")
				addPromptOnlyFlag(cmd)
				cmd.SetContext(context.Background())
				cmd.SetOut(io.Discard)
				_ = captureStdout(t, func() {
					if err := runReconcile(cmd, nil); err != nil {
						t.Fatalf("runReconcile() error = %v", err)
					}
				})
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := setupManagedSafetyGuidanceProject(t)
			stubManagedSafetyRulesetRegistry(t)
			setWorkingDirectory(t, projectRoot)

			tt.apply(t)
			assertManagedSafetyGuidance(t, projectRoot)
		})
	}
}

func TestAuditWorkLaneShorthandGuidanceFindsExistingSectionDrift(t *testing.T) {
	for _, tt := range []struct {
		path    string
		snippet string
	}{
		{path: "AGENTS.md", snippet: "remaining text is supplemental lane instructions"},
		{path: "CLAUDE.md", snippet: "remaining text is supplemental lane instructions"},
		{path: ".github/copilot-instructions.md", snippet: "remaining text is supplemental lane instructions"},
		{path: "docs/agents/GUARDRAILS.md", snippet: "shorthand is the primary lane choice"},
		{path: "AGENTS.md", snippet: "`new lane`, `new work lane`, `new worklane`, and `new worktree`"},
		{path: "CLAUDE.md", snippet: "human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan"},
		{path: ".github/copilot-instructions.md", snippet: "human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan"},
		{path: "docs/agents/GUARDRAILS.md", snippet: "human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan"},
	} {
		t.Run(tt.path, func(t *testing.T) {
			projectRoot, _ := setupLifecycleTestProject(t)
			removeGuidanceSnippet(t, projectRoot, tt.path, tt.snippet)

			findings := auditWorkLaneShorthandGuidance(projectRoot)
			absolutePath := filepath.Join(projectRoot, filepath.FromSlash(tt.path))
			for _, finding := range findings {
				if finding.FilePath == absolutePath && strings.Contains(finding.Issue, tt.snippet) {
					return
				}
			}
			t.Fatalf("no shorthand drift finding for %s: %#v", tt.path, findings)
		})
	}
}

func setupManagedSafetyGuidanceProject(t *testing.T) string {
	t.Helper()
	projectRoot, _ := setupLifecycleTestProject(t)
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	for _, relativePath := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
	} {
		if err := os.Remove(filepath.Join(projectRoot, filepath.FromSlash(relativePath))); err != nil {
			t.Fatalf("remove fixture artifact %s: %v", relativePath, err)
		}
	}
	return projectRoot
}

func stubManagedSafetyRulesetRegistry(t *testing.T) {
	t.Helper()
	workLane := readRepositoryFile(t, "docs/references/rules/work-lane-gating.md")
	deletionSafety := readRepositoryFile(t, "docs/references/rules/deletion-safety.md")
	stubRulesetRegistry(
		t,
		registryRulesetWithContentForTest("work-lane-gating", workLane, "test-work-lane"),
		registryRulesetWithContentForTest("deletion-safety", deletionSafety, "test-deletion-safety"),
	)
}

func assertManagedSafetyGuidance(t *testing.T, projectRoot string) {
	t.Helper()
	for _, relativePath := range []string{
		"AGENTS.md",
		"CLAUDE.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
	} {
		content := readFile(t, filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
		for _, snippet := range []string{
			"case-insensitively",
			"`c` means continue existing",
			"`n` or `y` means new lane",
			"shorthand is the primary lane choice",
			"`new lane`, `new work lane`, `new worklane`, and `new worktree`",
			"human-assigned GitHub issue, exact `GH-<issue-number>` branch, canonical non-primary worktree, and ready pull-request plan",
		} {
			if !strings.Contains(content, snippet) {
				t.Errorf("%s does not contain %q", relativePath, snippet)
			}
		}
	}

	for _, relativePath := range []string{
		"docs/references/rules/work-lane-gating.md",
		"docs/references/rules/deletion-safety.md",
	} {
		content := readFile(t, filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
		if !strings.Contains(content, "registry_scope: downstream") {
			t.Errorf("%s is not marked for downstream propagation", relativePath)
		}
	}
}
