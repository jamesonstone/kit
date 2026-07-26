package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func TestRunScaffoldAgentsAppendOnly_MergesMissingSectionsWithoutOverwritingExistingContent(t *testing.T) {
	tempDir := t.TempDir()
	setWorkingDirectory(t, tempDir)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	agentsPath := filepath.Join(tempDir, agentsMDPath)
	original := strings.Replace(templates.MemoryAgentsMD, "## Constraints", "## Custom Notes\n\nkeep this note\n\n## Constraints", 1)
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
		"keep this note",
		"## Repository Memory Gate",
		"## Final Response Contract",
	} {
		if !strings.Contains(merged, check) {
			t.Fatalf("expected merged file to contain %q, got %q", check, merged)
		}
	}

	if strings.Count(merged, "keep this note") != 1 {
		t.Fatalf("expected custom section to remain exactly once, got %q", merged)
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
