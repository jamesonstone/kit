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

func TestAuditWorkLaneDefaultGuidanceFindsExistingSectionDrift(t *testing.T) {
	for _, tt := range []struct {
		path    string
		snippet string
	}{
		{path: "AGENTS.md", snippet: "Default to a new worklane without asking"},
		{path: "CLAUDE.md", snippet: "Default to a new worklane without asking"},
		{path: ".github/copilot-instructions.md", snippet: "Default to a new worklane without asking"},
		{path: "docs/agents/GUARDRAILS.md", snippet: "Default to a new worklane without asking"},
		{path: "AGENTS.md", snippet: "Never offer or ask the user to choose between lanes"},
		{path: "CLAUDE.md", snippet: "Never offer or ask the user to choose between lanes"},
		{path: ".github/copilot-instructions.md", snippet: "Never offer or ask the user to choose between lanes"},
		{path: "docs/agents/GUARDRAILS.md", snippet: "Never offer or ask the user to choose between lanes"},
		{path: "AGENTS.md", snippet: "Treat exact existing-PR lifecycle work as continuation"},
		{path: "CLAUDE.md", snippet: "Treat exact existing-PR lifecycle work as continuation"},
		{path: ".github/copilot-instructions.md", snippet: "Treat exact existing-PR lifecycle work as continuation"},
		{path: "docs/agents/GUARDRAILS.md", snippet: "Exact existing pull requests targeted for review repair, CI repair, base"},
	} {
		t.Run(tt.path, func(t *testing.T) {
			projectRoot, _ := setupLifecycleTestProject(t)
			removeGuidanceSnippet(t, projectRoot, tt.path, tt.snippet)

			findings := auditWorkLaneDefaultGuidance(projectRoot)
			absolutePath := filepath.Join(projectRoot, filepath.FromSlash(tt.path))
			for _, finding := range findings {
				if finding.FilePath == absolutePath && strings.Contains(finding.Issue, tt.snippet) {
					return
				}
			}
			t.Fatalf("no default routing drift finding for %s: %#v", tt.path, findings)
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
	completionOutput := readRepositoryFile(t, "docs/references/rules/agent-completion-output.md")
	humanAuthorship := readRepositoryFile(t, "docs/references/rules/human-authorship.md")
	stubRulesetRegistry(
		t,
		registryRulesetWithContentForTest("work-lane-gating", workLane, "test-work-lane"),
		registryRulesetWithContentForTest("deletion-safety", deletionSafety, "test-deletion-safety"),
		registryRulesetWithContentForTest("agent-completion-output", completionOutput, "test-completion-output"),
		registryRulesetWithContentForTest("human-authorship", humanAuthorship, "test-human-authorship"),
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
			"Default to a new worklane without asking",
			"one human-assigned",
			"exact `GH-<issue-number>` branch",
			"canonical non-primary",
			"ready pull-request plan",
			"Continue an existing lane only when the user explicitly directs",
			"Never offer or ask the user to choose between lanes",
			"review repair, CI repair, base",
			"ordered merge coordination",
			"coordination or corrective pull request",
			"bounded in-place-remediation authority",
			"allocate a new lane",
		} {
			if !strings.Contains(content, snippet) {
				t.Errorf("%s does not contain %q", relativePath, snippet)
			}
		}
		for _, snippet := range []string{
			"## Agent Completion Output Contract",
			"Answer ordinary conversational requests naturally",
			"must not receive status tokens, canonical section headings, synthetic None items, task profiles, or repository-memory reporting",
			"Use the structured contract when omitting it could hide a blocker",
			"Do not classify by word count, token count, elapsed time, or tool-call count",
			"emit exactly `## What happened`, `## Deviations`, and `## Next steps` in that order",
			"**Status: PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>.**",
			"Use one `**None.**` bullet when there are no deviations",
			"Use one `**None.**` bullet when no action remains",
		} {
			if !strings.Contains(content, snippet) {
				t.Errorf("%s does not contain %q", relativePath, snippet)
			}
		}
		if strings.Contains(content, legacyOperatorActionTableHeader) {
			t.Errorf("%s still contains leftover operator-action table %q", relativePath, legacyOperatorActionTableHeader)
		}
		for _, forbidden := range []string{
			"# PASS|PARTIAL|BLOCKED|FAIL — <one-sentence outcome>",
			"prioritized action list ordered Blocker, Incomplete, Next, Optional, then None",
		} {
			if strings.Contains(content, forbidden) {
				t.Errorf("%s still contains superseded completion guidance %q", relativePath, forbidden)
			}
		}
	}

	for _, relativePath := range []string{
		"docs/references/rules/work-lane-gating.md",
		"docs/references/rules/deletion-safety.md",
		"docs/references/rules/agent-completion-output.md",
		"docs/references/rules/human-authorship.md",
	} {
		content := readFile(t, filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
		if !strings.Contains(content, "registry_scope: downstream") {
			t.Errorf("%s is not marked for downstream propagation", relativePath)
		}
	}
}
