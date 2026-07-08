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
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	previous := clipboardCopyFunc
	t.Cleanup(func() {
		clipboardCopyFunc = previous
	})
	clipboardCopyFunc = func(text string) error {
		return nil
	}

	if err := runInit(initCmd, nil); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	for _, relativePath := range []string{"AGENTS.md", "CLAUDE.md", copilotInstructionsPath} {
		absolutePath := filepath.Join(tempDir, relativePath)
		if _, err := os.Stat(absolutePath); err != nil {
			t.Fatalf("expected %s to exist: %v", relativePath, err)
		}
	}
	for _, relativePath := range []string{
		"docs/agents/README.md",
		"docs/agents/WORKFLOWS.md",
		"docs/agents/RLM.md",
		"docs/agents/TOOLING.md",
		"docs/agents/GUARDRAILS.md",
		"docs/references/README.md",
	} {
		assertFileExists(t, filepath.Join(tempDir, relativePath))
	}
	agentsContent, err := os.ReadFile(filepath.Join(tempDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentsContent), "`docs/agents/README.md`") {
		t.Fatalf("AGENTS.md did not contain the docs/agents entrypoint guidance")
	}

	claudeContent, err := os.ReadFile(filepath.Join(tempDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeContent), "`docs/agents/WORKFLOWS.md`") {
		t.Fatalf("CLAUDE.md did not contain the workflows entrypoint guidance")
	}
	copilotContent, err := os.ReadFile(filepath.Join(tempDir, copilotInstructionsPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", copilotInstructionsPath, err)
	}
	if !strings.HasPrefix(string(copilotContent), "# GitHub Copilot Repository Instructions\n\n") {
		t.Fatalf("%s did not contain the expected heading", copilotInstructionsPath)
	}
	if !strings.Contains(string(copilotContent), "`docs/agents/README.md`") {
		t.Fatalf("%s did not contain the docs/agents entrypoint guidance", copilotInstructionsPath)
	}

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.InstructionScaffoldVersion != config.DefaultInstructionScaffoldVersion {
		t.Fatalf(
			"expected instruction scaffold version %d, got %d",
			config.DefaultInstructionScaffoldVersion,
			cfg.InstructionScaffoldVersion,
		)
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

		output := captureStdout(t, func() {
			if err := runScaffoldAgents(scaffoldAgentsCmd, nil); err != nil {
				t.Fatalf("runScaffoldAgents() error = %v", err)
			}
		})
		if !strings.Contains(output, "♻️ agents directory and files empty scaffolding created.") {
			t.Fatalf("expected agents scaffold completion wording, got %q", output)
		}
	})

	assertFileExists(t, filepath.Join(tempDir, agentsMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, claudeMDPath))
	assertFileDoesNotExist(t, filepath.Join(tempDir, copilotInstructionsPath))
	assertFileExists(t, filepath.Join(tempDir, "docs/agents/README.md"))
	assertFileExists(t, filepath.Join(tempDir, "docs/references/README.md"))
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
	assertFileExists(t, filepath.Join(tempDir, "docs/agents/README.md"))
	assertFileExists(t, filepath.Join(tempDir, "docs/references/README.md"))
}
