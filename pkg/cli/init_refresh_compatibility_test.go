package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestRunInitRefresh_PreservesGeneratedVerboseInstructionsAsLegacy(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, agentsMDPath), templates.LegacyAgentsMD)
	writeFile(t, filepath.Join(tempDir, claudeMDPath), templates.LegacyClaudeMD)
	writeFile(t, filepath.Join(tempDir, copilotInstructionsPath), templates.LegacyCopilotInstructionsMD)
	writeFile(t, filepath.Join(tempDir, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n\n## PRINCIPLES\n\ncustom\n\n## CONSTRAINTS\n\ncustom\n\n## NON-GOALS\n\nnone\n\n## DEFINITIONS\n\nnone\n")
	writeFile(t, filepath.Join(tempDir, envrcPath), "source_env .custom\n")

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.InstructionScaffoldVersion != config.InstructionScaffoldVersionVerbose {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", updated.InstructionScaffoldVersion, config.InstructionScaffoldVersionVerbose)
	}

	agentsContent, err := os.ReadFile(filepath.Join(tempDir, agentsMDPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(agentsContent) != templates.LegacyAgentsMD {
		t.Fatalf("expected generated v1 %s to remain supported legacy input", agentsMDPath)
	}
	assertFileDoesNotExist(t, filepath.Join(tempDir, "docs", "agents", "README.md"))
	assertFileDoesNotExist(t, filepath.Join(tempDir, "docs", "references", "README.md"))

	envrcContent, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", envrcPath, err)
	}
	if string(envrcContent) != "source_env .custom\n" {
		t.Fatalf("%s was overwritten during default refresh: %q", envrcPath, envrcContent)
	}

	constitutionContent, err := os.ReadFile(filepath.Join(tempDir, "docs", "CONSTITUTION.md"))
	if err != nil {
		t.Fatalf("failed to read CONSTITUTION.md: %v", err)
	}
	for _, check := range []string{
		"custom",
		"### Kit-Managed Baseline Rules",
		"docs/references/rules/agent-completion-output.md",
		"version-control-eligible handwritten implementation/source and test file at 300 physical lines or less",
		"vendored dependencies, and proven generated files",
	} {
		if !strings.Contains(string(constitutionContent), check) {
			t.Fatalf("expected CONSTITUTION.md to contain %q, got:\n%s", check, constitutionContent)
		}
	}

	rulesetPath := filepath.Join(tempDir, "docs", "references", "rules", "safety-guardrails.md")
	rulesetContent, err := os.ReadFile(rulesetPath)
	if err != nil {
		t.Fatalf("expected registry ruleset to be imported: %v", err)
	}
	if !strings.Contains(string(rulesetContent), "slug: safety-guardrails") {
		t.Fatalf("unexpected ruleset content:\n%s", rulesetContent)
	}
}

func TestRunInit_DiffRequiresDryRun(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initDiff = true

		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "--diff requires --dry-run") {
			t.Fatalf("expected --diff without --dry-run error, got %v", err)
		}
	})
}
