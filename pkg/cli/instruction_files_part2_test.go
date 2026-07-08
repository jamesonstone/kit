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

	if strings.Index(merged, "## Source of truth") >= strings.Index(merged, "## Custom Notes") ||
		strings.Index(merged, "## Custom Notes") >= strings.Index(merged, "## Change Classification (Required First Step)") ||
		strings.Index(merged, "## Change Classification (Required First Step)") >= strings.Index(merged, "## Communication Style") {
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
