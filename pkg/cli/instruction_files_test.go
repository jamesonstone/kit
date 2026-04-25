package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
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

func TestRunScaffoldAgentsForce_CancelLeavesFilesUnchanged(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	agentsPath := filepath.Join(tempDir, agentsMDPath)
	original := "# AGENTS\n\n## Source of truth\n\ncustom source\n"
	writeFile(t, agentsPath, original)

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsForce = true
		scaffoldAgentsAgentsMD = true

		cmd := &cobra.Command{}
		promptOutput := &bytes.Buffer{}
		cmd.SetOut(promptOutput)
		cmd.SetIn(strings.NewReader("n\n"))

		if err := runScaffoldAgents(cmd, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}

		if !strings.Contains(promptOutput.String(), agentsMDPath) {
			t.Fatalf("expected overwrite prompt to mention %s, got %q", agentsMDPath, promptOutput.String())
		}
		if !strings.Contains(promptOutput.String(), "Proceed? [y/N]") {
			t.Fatalf("expected overwrite confirmation prompt, got %q", promptOutput.String())
		}
	})

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(content) != original {
		t.Fatalf("expected %s to remain unchanged after cancellation", agentsMDPath)
	}
}

func TestRunScaffoldAgentsForceYes_OverwritesWithoutPrompt(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	agentsPath := filepath.Join(tempDir, agentsMDPath)
	writeFile(t, agentsPath, "custom instructions\n")

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsForce = true
		scaffoldAgentsYes = true
		scaffoldAgentsAgentsMD = true

		cmd := &cobra.Command{}
		promptOutput := &bytes.Buffer{}
		cmd.SetOut(promptOutput)

		if err := runScaffoldAgents(cmd, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}

		if promptOutput.Len() != 0 {
			t.Fatalf("expected no overwrite prompt with --yes, got %q", promptOutput.String())
		}
	})

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(content) != templates.InstructionFileForVersion(agentsMDPath, config.InstructionScaffoldVersionVerbose) {
		t.Fatalf("expected %s to be overwritten with the scaffold template", agentsMDPath)
	}
}

func TestRunScaffoldAgentsAppendOnly_MergesMissingSectionsWithoutOverwritingExistingContent(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	agentsPath := filepath.Join(tempDir, agentsMDPath)
	original := `# AGENTS

intro

## Source of truth

custom source

## Custom Notes

keep this note

## Communication Style

custom style
`
	writeFile(t, agentsPath, original)

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsAppendOnly = true
		scaffoldAgentsAgentsMD = true

		if err := runScaffoldAgents(&cobra.Command{}, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}
	})

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	merged := string(content)

	for _, check := range []string{
		"custom source",
		"keep this note",
		"custom style",
		"## Change Classification (Required First Step)",
		"## Document Completeness",
	} {
		if !strings.Contains(merged, check) {
			t.Fatalf("expected merged file to contain %q, got %q", check, merged)
		}
	}

	if !(strings.Index(merged, "## Source of truth") < strings.Index(merged, "## Custom Notes") &&
		strings.Index(merged, "## Custom Notes") < strings.Index(merged, "## Change Classification (Required First Step)") &&
		strings.Index(merged, "## Change Classification (Required First Step)") < strings.Index(merged, "## Communication Style")) {
		t.Fatalf("expected missing Kit sections to be inserted before the next recognized section, got %q", merged)
	}
}

func TestRunScaffoldAgentsAppendOnly_MergesSupportDocs(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	rlmPath := filepath.Join(tempDir, "docs", "agents", "RLM.md")
	writeFile(t, rlmPath, `# RLM

## Purpose

custom purpose
`)

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsAppendOnly = true
		scaffoldAgentsVersion = config.InstructionScaffoldVersionTOC

		if err := runScaffoldAgents(&cobra.Command{}, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}
	})

	content, err := os.ReadFile(rlmPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", rlmPath, err)
	}
	merged := string(content)
	for _, check := range []string{
		"custom purpose",
		"## Runtime Loop",
		"## Context Budget Rules",
	} {
		if !strings.Contains(merged, check) {
			t.Fatalf("expected merged support doc to contain %q, got %q", check, merged)
		}
	}
}

func TestRunScaffoldAgentsAppendOnly_FailsBeforeAnyWritesWhenAFileIsNotMergeable(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	agentsPath := filepath.Join(tempDir, agentsMDPath)
	original := "# AGENTS\n\ncompletely custom instructions\n"
	writeFile(t, agentsPath, original)
	claudePath := filepath.Join(tempDir, claudeMDPath)

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsAppendOnly = true
		scaffoldAgentsAgentsMD = true
		scaffoldAgentsClaude = true

		err := runScaffoldAgents(&cobra.Command{}, nil)
		if err == nil || !strings.Contains(err.Error(), "no recognizable Kit-managed sections") {
			t.Fatalf("expected append-only anchor error, got %v", err)
		}
	})

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(content) != original {
		t.Fatalf("expected %s to remain unchanged after append-only failure", agentsMDPath)
	}
	assertFileDoesNotExist(t, claudePath)
}

func TestRunScaffoldAgents_DefaultModeSuggestsAppendOnlyWhenSkippingExistingFiles(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	writeFile(t, filepath.Join(tempDir, copilotInstructionsPath), "custom copilot instructions\n")

	output := captureStdout(t, func() {
		withScaffoldAgentFlags(t, func() {
			scaffoldAgentsCopilot = true
			if err := runScaffoldAgents(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runScaffoldAgents() error = %v", err)
			}
		})
	})

	if !strings.Contains(output, "--append-only") || !strings.Contains(output, "--force") {
		t.Fatalf("expected output to suggest append-only and force, got %q", output)
	}
}

func TestRunScaffoldAgents_RejectsAppendOnlyWithForce(t *testing.T) {
	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsForce = true
		scaffoldAgentsAppendOnly = true
		err := runScaffoldAgents(&cobra.Command{}, nil)
		if err == nil || !strings.Contains(err.Error(), "--append-only cannot be used with --force") {
			t.Fatalf("expected flag validation error, got %v", err)
		}
	})
}

func TestRunScaffoldAgents_RejectsYesWithoutForce(t *testing.T) {
	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsYes = true
		err := runScaffoldAgents(&cobra.Command{}, nil)
		if err == nil || !strings.Contains(err.Error(), "--yes requires --force") {
			t.Fatalf("expected --yes validation error, got %v", err)
		}
	})
}

func TestRunScaffoldAgents_VersionChangeRequiresForce(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, agentsMDPath), templates.AgentsMD)
	writeFile(t, filepath.Join(tempDir, claudeMDPath), templates.ClaudeMD)
	writeFile(t, filepath.Join(tempDir, copilotInstructionsPath), templates.CopilotInstructionsMD)
	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		writeFile(t, filepath.Join(tempDir, support.RelativePath), support.Content)
	}

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsVersion = config.InstructionScaffoldVersionVerbose

		err := runScaffoldAgents(&cobra.Command{}, nil)
		if err == nil || !strings.Contains(err.Error(), "requires --force") {
			t.Fatalf("expected version-change force error, got %v", err)
		}
	})
}

func TestRunScaffoldAgents_Version1ForceRemovesV2DocsTree(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, agentsMDPath), templates.AgentsMD)
	writeFile(t, filepath.Join(tempDir, claudeMDPath), templates.ClaudeMD)
	writeFile(t, filepath.Join(tempDir, copilotInstructionsPath), templates.CopilotInstructionsMD)
	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		writeFile(t, filepath.Join(tempDir, support.RelativePath), support.Content)
	}

	withScaffoldAgentFlags(t, func() {
		scaffoldAgentsVersion = config.InstructionScaffoldVersionVerbose
		scaffoldAgentsForce = true
		scaffoldAgentsYes = true

		if err := runScaffoldAgents(&cobra.Command{}, nil); err != nil {
			t.Fatalf("runScaffoldAgents() error = %v", err)
		}
	})

	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		assertFileDoesNotExist(t, filepath.Join(tempDir, support.RelativePath))
	}

	agentsContent, err := os.ReadFile(filepath.Join(tempDir, agentsMDPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(agentsContent) != templates.InstructionFileForVersion(agentsMDPath, config.InstructionScaffoldVersionVerbose) {
		t.Fatalf("expected %s to revert to the verbose template", agentsMDPath)
	}

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.InstructionScaffoldVersion != config.InstructionScaffoldVersionVerbose {
		t.Fatalf("expected version 1 after downgrade, got %d", updated.InstructionScaffoldVersion)
	}
}

func TestRenderScaffoldAgentsHelp_IncludesVersionTable(t *testing.T) {
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.Flags().AddFlagSet(scaffoldAgentsCmd.LocalFlags())
	cmd.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())
	cmd.Use = scaffoldAgentsCmd.Use
	cmd.Short = scaffoldAgentsCmd.Short
	cmd.Long = scaffoldAgentsCmd.Long
	cmd.Aliases = scaffoldAgentsCmd.Aliases

	if err := renderScaffoldAgentsHelp(cmd); err != nil {
		t.Fatalf("renderScaffoldAgentsHelp() error = %v", err)
	}

	content := out.String()
	for _, check := range []string{"Version Models", "verbose", "toc/rlm", "--version int"} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected help output to contain %q, got %q", check, content)
		}
	}
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
	originalYes := scaffoldAgentsYes
	originalAppendOnly := scaffoldAgentsAppendOnly
	originalVersion := scaffoldAgentsVersion

	t.Cleanup(func() {
		scaffoldAgentsForce = originalForce
		scaffoldAgentsCopilot = originalCopilot
		scaffoldAgentsClaude = originalClaude
		scaffoldAgentsAgentsMD = originalAgentsMD
		scaffoldAgentsYes = originalYes
		scaffoldAgentsAppendOnly = originalAppendOnly
		scaffoldAgentsVersion = originalVersion
	})

	scaffoldAgentsForce = false
	scaffoldAgentsCopilot = false
	scaffoldAgentsClaude = false
	scaffoldAgentsAgentsMD = false
	scaffoldAgentsYes = false
	scaffoldAgentsAppendOnly = false
	scaffoldAgentsVersion = 0

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
