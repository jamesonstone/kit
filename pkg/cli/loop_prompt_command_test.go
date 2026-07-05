package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestLoopPromptCommandIsRegisteredUnderLoop(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"loop", "prompt"})
	if err != nil {
		t.Fatalf("rootCmd.Find(loop prompt) error = %v", err)
	}
	if cmd == nil || cmd.Name() != "prompt" {
		t.Fatalf("expected loop prompt command, got %#v", cmd)
	}
	if flag := cmd.Flags().Lookup("output-only"); flag == nil {
		t.Fatal("expected loop prompt to expose --output-only")
	}
	if flag := cmd.Flags().Lookup("copy"); flag == nil {
		t.Fatal("expected loop prompt to expose --copy")
	}
}

func TestBuildLoopEngineeringPromptIncludesExecutionCycleAndDelivery(t *testing.T) {
	prompt := buildLoopEngineeringPrompt(loopPromptInput{
		Title:      "the billing workflow",
		Source:     "`docs/specs/0007-billing/SPEC.md`",
		Scope:      "every phase and acceptance criterion",
		Validation: "go test ./... and OpenAPI validation",
		Docs:       "OpenAPI and README updates",
		Delivery:   "create an issue-number branch and ready PR after validation",
	})

	checks := []string{
		"## Loop Goal",
		"Treat `docs/specs/0007-billing/SPEC.md` as the implementation source of truth.",
		"Continue until this scope is completed and validated",
		"## Phase Execution Cycle",
		"Re-read the current requirements, acceptance criteria, validation map, task checklist",
		"Run relevant regression checks for previously completed phases",
		"Re-review the immediately previous completed phase",
		"## Phase Cycle Outputs",
		"tests and validation commands run",
		"## Validation And Regression Contract",
		"go test ./... and OpenAPI validation",
		"## Final Integration Review",
		"Review all completed phases together",
		"## GitHub Delivery Boundary",
		"Create or reuse the correct GitHub issue before branching",
		"Refresh the base with fetch-only behavior before branching",
		"Stage explicitly with `git add <file>` only",
		"Create the PR ready for review",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", check, prompt)
		}
	}
}

func TestBuildFeatureLoopPromptUsesExistingSpecState(t *testing.T) {
	projectRoot := setupLoopProject(t, "agent.sh", loopAgentScript(99, true, false))
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), validV2SpecWithPhase("0001-alpha", "implement"))
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	prompt, feat, err := buildFeatureLoopPrompt(projectRoot, cfg, "alpha")
	if err != nil {
		t.Fatalf("buildFeatureLoopPrompt() error = %v", err)
	}
	if feat.Slug != "alpha" {
		t.Fatalf("feature slug = %q, want alpha", feat.Slug)
	}
	for _, check := range []string{
		"## Feature Context",
		"Feature: `alpha`",
		"Directory: `0001-alpha`",
		"Current phase: `implement`",
		"SPEC.md:",
		"every remaining phase, task-checklist item, acceptance criterion",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected feature prompt to contain %q, got:\n%s", check, prompt)
		}
	}
}

func TestReadAdHocLoopPromptInputDefaults(t *testing.T) {
	var out bytes.Buffer
	input, err := readAdHocLoopPromptInput(&out, strings.NewReader("\n\n\n\n\n\n"))
	if err != nil {
		t.Fatalf("readAdHocLoopPromptInput() error = %v", err)
	}
	if input.Title != "complete the requested work end to end" {
		t.Fatalf("Title = %q", input.Title)
	}
	if input.Source != "the current user request plus repo-local instructions" {
		t.Fatalf("Source = %q", input.Source)
	}
	if input.Delivery == "" {
		t.Fatal("Delivery default is empty")
	}
	if !strings.Contains(out.String(), "Loop goal") {
		t.Fatalf("expected intake prompts, got %q", out.String())
	}
}
