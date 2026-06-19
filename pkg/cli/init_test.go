package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunInit_DefaultCopiesConstitutionPromptAndShowsPasteStep(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	withInitFlags(t, func() {
		previous := clipboardCopyFunc
		defer func() {
			clipboardCopyFunc = previous
		}()

		var copied string
		clipboardCopyFunc = func(text string) error {
			copied = text
			return nil
		}

		output := captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		const constitutionPath = "docs/CONSTITUTION.md"
		if !strings.Contains(copied, "Please update "+filepath.Join(cwd, constitutionPath)+" with all patterns") {
			t.Fatalf("expected copied prompt to target %s, got %q", filepath.Join(cwd, constitutionPath), copied)
		}
		if strings.Contains(copied, "Copy this section to the Agent:") {
			t.Fatalf("expected clipboard payload to contain only the prompt body, got %q", copied)
		}
		if !strings.Contains(output, "Copied the prepared text to the clipboard.") {
			t.Fatalf("expected stdout to acknowledge clipboard copy, got %q", output)
		}
		if !strings.Contains(output, "1. Paste the copied prompt into your agent to draft docs/CONSTITUTION.md") {
			t.Fatalf("expected stdout to include the paste guidance, got %q", output)
		}
		if strings.Contains(output, "Please update "+filepath.Join(cwd, constitutionPath)) {
			t.Fatalf("expected default output to avoid printing the raw prompt, got %q", output)
		}
	})
}

func TestRunInit_OutputOnlyPrintsRawPromptAndSkipsDefaultCopy(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	withInitFlags(t, func() {
		initOutputOnly = true

		previous := clipboardCopyFunc
		defer func() {
			clipboardCopyFunc = previous
		}()

		copied := false
		clipboardCopyFunc = func(text string) error {
			copied = true
			return nil
		}

		output := captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		if copied {
			t.Fatalf("expected --output-only to skip default clipboard copy")
		}
		if !strings.HasPrefix(output, "Please update "+filepath.Join(cwd, "docs/CONSTITUTION.md")) {
			t.Fatalf("expected raw prompt output, got %q", output)
		}
		if strings.Contains(output, "Initializing Kit project") {
			t.Fatalf("expected --output-only to suppress init status output, got %q", output)
		}
		if strings.Contains(output, "Next steps") {
			t.Fatalf("expected --output-only to suppress numbered next steps, got %q", output)
		}
	})
}

func TestRunInit_OutputOnlyAndCopyDoesBoth(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	withInitFlags(t, func() {
		initOutputOnly = true
		initCopy = true

		previous := clipboardCopyFunc
		defer func() {
			clipboardCopyFunc = previous
		}()

		var copied string
		clipboardCopyFunc = func(text string) error {
			copied = text
			return nil
		}

		output := captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		if copied == "" {
			t.Fatalf("expected --output-only --copy to copy the prompt")
		}
		if output != copied {
			t.Fatalf("expected stdout and clipboard payload to match, stdout = %q, copied = %q", output, copied)
		}
	})
}

func TestRunInit_PopulatesGlobalConfig(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		configPath := filepath.Join(homeDir, ".config", "kit", config.ConfigFileName)
		if _, err := os.Stat(configPath); err != nil {
			t.Fatalf("expected global config at %s: %v", configPath, err)
		}

		cfg, found, err := config.LoadGlobal()
		if err != nil {
			t.Fatalf("config.LoadGlobal() error = %v", err)
		}
		if !found {
			t.Fatal("config.LoadGlobal() found = false, want true")
		}
		if cfg.GoalPercentage != 95 {
			t.Fatalf("GoalPercentage = %d, want 95", cfg.GoalPercentage)
		}
		if cfg.InstructionScaffoldVersion != config.DefaultInstructionScaffoldVersion {
			t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, config.DefaultInstructionScaffoldVersion)
		}
	})
}

func TestRunInit_CreatesLoopReviewAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	created, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	assertDefaultInitLoopAgent(t, created)
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
