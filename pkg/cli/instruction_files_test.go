package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestSelectedInstructionFiles_ReturnsConfiguredFilesWithoutExplicitSelection(t *testing.T) {
	cfg := &config.Config{Agents: []string{agentsMDPath, claudeMDPath}}

	got := selectedInstructionFiles(cfg, instructionFileSelection{})
	want := []string{agentsMDPath, claudeMDPath, copilotInstructionsPath}

	if !slices.Equal(got, want) {
		t.Fatalf("selectedInstructionFiles() = %v, want %v", got, want)
	}
}

func TestSelectedInstructionFiles_ReturnsOnlyExplicitTargets(t *testing.T) {
	cfg := &config.Config{Agents: []string{"GEMINI.md"}}
	selection := instructionFileSelection{agentsMD: true, copilot: true}

	got := selectedInstructionFiles(cfg, selection)
	want := []string{agentsMDPath, copilotInstructionsPath}

	if !slices.Equal(got, want) {
		t.Fatalf("selectedInstructionFiles() = %v, want %v", got, want)
	}
}

func TestRunInit_CreatesRepositoryInstructionFiles(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := runInit(initCmd, nil); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	for _, relativePath := range []string{"AGENTS.md", "CLAUDE.md", copilotInstructionsPath} {
		absolutePath := filepath.Join(tempDir, relativePath)
		if _, err := os.Stat(absolutePath); err != nil {
			t.Fatalf("expected %s to exist: %v", relativePath, err)
		}
	}
	agentsContent, err := os.ReadFile(filepath.Join(tempDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentsContent), "`git worktree add ~/worktrees/<repo>-<branch> <branch>`") {
		t.Fatalf("AGENTS.md did not contain the flat worktree guidance")
	}

	claudeContent, err := os.ReadFile(filepath.Join(tempDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeContent), "## Workflow: Plan → Act → Reflect (Spec-Driven Track)") {
		t.Fatalf("CLAUDE.md did not contain the comprehensive workflow template")
	}
	if !strings.Contains(string(claudeContent), "`~/worktrees/`") {
		t.Fatalf("CLAUDE.md did not contain the flat worktree guidance")
	}

	copilotContent, err := os.ReadFile(filepath.Join(tempDir, copilotInstructionsPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", copilotInstructionsPath, err)
	}
	if !strings.HasPrefix(string(copilotContent), "# GitHub Copilot Repository Instructions\n\n") {
		t.Fatalf("%s did not contain the expected heading", copilotInstructionsPath)
	}
	if !strings.Contains(string(copilotContent), "`~/worktrees/`") {
		t.Fatalf("%s did not contain the flat worktree guidance", copilotInstructionsPath)
	}
}

func TestRunScaffoldAgents_SkipsExistingCopilotInstructionsWithoutForce(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	originalForce := scaffoldAgentsForce
	scaffoldAgentsForce = false
	t.Cleanup(func() {
		scaffoldAgentsForce = originalForce
	})

	customContent := "custom copilot instructions\n"
	copilotPath := filepath.Join(tempDir, copilotInstructionsPath)
	if err := os.MkdirAll(filepath.Dir(copilotPath), 0755); err != nil {
		t.Fatalf("failed to create .github directory: %v", err)
	}
	if err := os.WriteFile(copilotPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", copilotInstructionsPath, err)
	}

	if err := runScaffoldAgents(scaffoldAgentsCmd, nil); err != nil {
		t.Fatalf("runScaffoldAgents() error = %v", err)
	}

	content, err := os.ReadFile(copilotPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", copilotInstructionsPath, err)
	}
	if got := string(content); got != customContent {
		t.Fatalf("expected existing %s to be preserved, got %q", copilotInstructionsPath, got)
	}
}

func TestRunScaffoldAgents_TargetedSelectionScaffoldsOnlyRequestedFiles(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsAgentsMD = true

		if err := runScaffoldAgents(scaffoldAgentsCmd, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}
	})

	assertFileExists(t, filepath.Join(tempDir, agentsMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, claudeMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, copilotInstructionsPath))
}

func TestRunScaffoldAgents_TargetedSelectionFallsBackToCurrentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsAgentsMD = true

		if err := runScaffoldAgents(scaffoldAgentsCmd, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}
	})

	assertFileExists(t, filepath.Join(tempDir, agentsMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, claudeMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, copilotInstructionsPath))
}

func TestScaffoldAgentsCmd_IncludesSingularAlias(t *testing.T) {
	if !slices.Contains(scaffoldAgentsCmd.Aliases, "scaffold-agent") {
		t.Fatalf("expected scaffold-agents to include scaffold-agent alias")
	}
}

func withScaffoldAgentFlags(t *testing.T, run func()) {
	t.Helper()

	originalForce := scaffoldAgentsForce
	originalCopilot := scaffoldAgentsCopilot
	originalClaude := scaffoldAgentsClaude
	originalAgentsMD := scaffoldAgentsAgentsMD

	t.Cleanup(func() {
		scaffoldAgentsForce = originalForce
		scaffoldAgentsCopilot = originalCopilot
		scaffoldAgentsClaude = originalClaude
		scaffoldAgentsAgentsMD = originalAgentsMD
	})

	scaffoldAgentsForce = false
	scaffoldAgentsCopilot = false
	scaffoldAgentsClaude = false
	scaffoldAgentsAgentsMD = false

	run()
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist, got err = %v", path, err)
	}
}

func setWorkingDirectory(t *testing.T, dir string) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}
