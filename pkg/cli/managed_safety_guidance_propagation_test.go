package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/templates"
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
			setWorkingDirectory(t, projectRoot)

			tt.apply(t)
			assertManagedSafetyGuidance(t, projectRoot)
		})
	}
}

func setupManagedSafetyGuidanceProject(t *testing.T) string {
	t.Helper()
	projectRoot, cfg := setupLifecycleTestProject(t)
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionMemory
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatal(err)
	}
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

func assertManagedSafetyGuidance(t *testing.T, projectRoot string) {
	t.Helper()
	block := templates.UniversalContractBlock()
	for _, relativePath := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		if content := readFile(t, filepath.Join(projectRoot, filepath.FromSlash(relativePath))); !strings.Contains(content, block) {
			t.Errorf("%s does not carry the universal contract block", relativePath)
		}
	}
	for _, slug := range []string{"delivery", "github-pr-merge", "deletion-safety", "slack-read-only", "infrastructure-change-approval"} {
		if !document.Exists(filepath.Join(projectRoot, filepath.FromSlash(rulesetTarget(slug)))) {
			t.Errorf("safety rule %s was not installed", slug)
		}
	}
}
