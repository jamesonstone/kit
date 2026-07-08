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

func TestScaffoldAgentsCmd_UsesScaffoldNamespace(t *testing.T) {
	if scaffoldAgentsCmd.Use != "agents" {
		t.Fatalf("expected scaffold agents subcommand use, got %q", scaffoldAgentsCmd.Use)
	}
	if slices.Contains(scaffoldAgentsCmd.Aliases, "scaffold-agent") {
		t.Fatal("expected legacy scaffold-agent alias to be removed")
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
