package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunInitRefresh_ForceIsIdempotentAfterConvergence(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	registry := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	stubRulesetRegistry(t, registry)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("first force refresh error = %v", err)
			}
		})
	})

	beforeConfig, err := os.ReadFile(filepath.Join(tempDir, config.ConfigFileName))
	if err != nil {
		t.Fatalf("failed to read config before second refresh: %v", err)
	}

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initForce = true

		previous := clipboardCopyFunc
		defer func() {
			clipboardCopyFunc = previous
		}()
		clipboardCopyFunc = func(text string) error {
			return nil
		}

		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("second force refresh error = %v", err)
			}
		})
	})

	afterConfig, err := os.ReadFile(filepath.Join(tempDir, config.ConfigFileName))
	if err != nil {
		t.Fatalf("failed to read config after second refresh: %v", err)
	}
	if string(afterConfig) != string(beforeConfig) {
		t.Fatalf("second force refresh rewrote config:\nbefore:\n%s\nafter:\n%s", beforeConfig, afterConfig)
	}
	if !strings.Contains(output, "Created: 0, Updated: 0, Merged: 0") {
		t.Fatalf("expected converged force refresh to report no writes, got:\n%s", output)
	}
	if !strings.Contains(output, "No Kit-managed project changes needed.") {
		t.Fatalf("expected converged force refresh to report no changes, got:\n%s", output)
	}
}

func TestRunInitRefreshForceCopiesDocumentationPrompt(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	var copied string
	withInitFlags(t, func() {
		initRefresh = true
		initForce = true

		previous := clipboardCopyFunc
		defer func() {
			clipboardCopyFunc = previous
		}()
		clipboardCopyFunc = func(text string) error {
			copied = text
			return nil
		}

		output := captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("force refresh error = %v", err)
			}
		})

		for _, want := range []string{
			"Documentation refresh prompt:",
			"Copied the prepared text to the clipboard.",
			"Paste the copied prompt into your agent to review semantic project documentation updates",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("expected force refresh output to contain %q, got:\n%s", want, output)
			}
		}
		if strings.Contains(output, "Post Init Refresh Documentation Review") {
			t.Fatalf("expected force refresh output not to print raw prompt, got:\n%s", output)
		}
	})

	for _, want := range []string{
		"## Post Init Refresh Documentation Review",
		"docs/CONSTITUTION.md",
		"docs/agents",
		"docs/references",
		"kit check --project",
		"no documentation updates needed",
	} {
		if !strings.Contains(copied, want) {
			t.Fatalf("expected copied prompt to contain %q, got:\n%s", want, copied)
		}
	}
}

func TestRunInitRefresh_MigratesGeneratedVerboseInstructionsAndCreatesManagedFiles(t *testing.T) {
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
	if updated.InstructionScaffoldVersion != config.InstructionScaffoldVersionTOC {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", updated.InstructionScaffoldVersion, config.InstructionScaffoldVersionTOC)
	}

	agentsContent, err := os.ReadFile(filepath.Join(tempDir, agentsMDPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(agentsContent) != templates.AgentsMD {
		t.Fatalf("expected generated v1 %s to migrate to v2 template", agentsMDPath)
	}
	assertFileExists(t, filepath.Join(tempDir, "docs", "agents", "README.md"))
	assertFileExists(t, filepath.Join(tempDir, "docs", "references", "README.md"))

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
		"Do not apply the code-file size guideline to documentation files",
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
