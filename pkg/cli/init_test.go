package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
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

func TestRunInit_CreatesCodeRabbitConfig(t *testing.T) {
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

		content, err := os.ReadFile(filepath.Join(tempDir, codeRabbitConfigPath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", codeRabbitConfigPath, err)
		}
		if string(content) != templates.CodeRabbitConfig {
			t.Fatalf("%s content = %q, want %q", codeRabbitConfigPath, content, templates.CodeRabbitConfig)
		}
	})
}

func TestRunInit_CreatesPullRequestTemplate(t *testing.T) {
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

		content, err := os.ReadFile(filepath.Join(tempDir, pullRequestTemplatePath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", pullRequestTemplatePath, err)
		}
		if string(content) != templates.PullRequestTemplate {
			t.Fatalf("%s content = %q, want %q", pullRequestTemplatePath, content, templates.PullRequestTemplate)
		}
	})
}

func TestRunInit_CreatesGitignoreWithKitLocalArtifacts(t *testing.T) {
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

		content, err := os.ReadFile(filepath.Join(tempDir, gitignorePath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", gitignorePath, err)
		}
		got := string(content)
		for _, pattern := range kitGitignorePatterns() {
			if !strings.Contains(got, pattern+"\n") {
				t.Fatalf("%s missing pattern %q; content:\n%s", gitignorePath, pattern, got)
			}
		}
		if strings.Contains(got, "\n.kit/\n") || strings.HasPrefix(got, ".kit/\n") {
			t.Fatalf("%s should not ignore all of .kit/; content:\n%s", gitignorePath, got)
		}
	})
}

func TestRunInit_CreatesLocalEnvironmentFiles(t *testing.T) {
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

		envContent, err := os.ReadFile(filepath.Join(tempDir, envPath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", envPath, err)
		}
		if string(envContent) != "" {
			t.Fatalf("%s content = %q, want blank file", envPath, envContent)
		}

		envrcContent, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", envrcPath, err)
		}
		if string(envrcContent) != templates.Envrc {
			t.Fatalf("%s content = %q, want %q", envrcPath, envrcContent, templates.Envrc)
		}
	})
}

func TestRunInit_PreservesExistingLocalEnvironmentFiles(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existingEnv := "CUSTOM=value\n"
	existingEnvrc := "source_env .custom\n"
	if err := os.WriteFile(filepath.Join(tempDir, envPath), []byte(existingEnv), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", envPath, err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, envrcPath), []byte(existingEnvrc), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", envrcPath, err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		envContent, err := os.ReadFile(filepath.Join(tempDir, envPath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", envPath, err)
		}
		if string(envContent) != existingEnv {
			t.Fatalf("%s content = %q, want %q", envPath, envContent, existingEnv)
		}

		envrcContent, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", envrcPath, err)
		}
		if string(envrcContent) != existingEnvrc {
			t.Fatalf("%s content = %q, want %q", envrcPath, envrcContent, existingEnvrc)
		}
	})
}

func TestRunInit_PreservesExistingCodeRabbitConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := "reviews:\n  auto_review:\n    enabled: false\n"
	if err := os.WriteFile(filepath.Join(tempDir, codeRabbitConfigPath), []byte(existing), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", codeRabbitConfigPath, err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		content, err := os.ReadFile(filepath.Join(tempDir, codeRabbitConfigPath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", codeRabbitConfigPath, err)
		}
		if string(content) != existing {
			t.Fatalf("%s content = %q, want %q", codeRabbitConfigPath, content, existing)
		}
	})
}

func TestRunInit_AppendsMissingGitignoreEntries(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := "# Custom ignores\ncustom.log\n.kit/runs/\n"
	if err := os.WriteFile(filepath.Join(tempDir, gitignorePath), []byte(existing), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", gitignorePath, err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		content, err := os.ReadFile(filepath.Join(tempDir, gitignorePath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", gitignorePath, err)
		}
		got := string(content)
		if !strings.HasPrefix(got, existing) {
			t.Fatalf("expected existing content to be preserved, got:\n%s", got)
		}
		for _, pattern := range kitGitignorePatterns() {
			if !strings.Contains(got, pattern+"\n") {
				t.Fatalf("%s missing pattern %q; content:\n%s", gitignorePath, pattern, got)
			}
		}
		if strings.Count(got, ".kit/runs/") != 1 {
			t.Fatalf("expected .kit/runs/ to remain deduplicated, got:\n%s", got)
		}
	})
}

func TestRunInit_PreservesExistingPullRequestTemplate(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := "## Summary\n\nCustom template\n"
	if err := document.Write(filepath.Join(tempDir, pullRequestTemplatePath), existing); err != nil {
		t.Fatalf("failed to seed %s: %v", pullRequestTemplatePath, err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		content, err := os.ReadFile(filepath.Join(tempDir, pullRequestTemplatePath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", pullRequestTemplatePath, err)
		}
		if string(content) != existing {
			t.Fatalf("%s content = %q, want %q", pullRequestTemplatePath, content, existing)
		}
	})
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

func TestRunInitRefresh_FileForceOverwritesOnlySelectedExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	writeFile(t, filepath.Join(tempDir, envrcPath), "source_env .custom\n")
	writeFile(t, filepath.Join(tempDir, codeRabbitConfigPath), "custom coderabbit\n")

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{envrcPath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	envrcContent, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", envrcPath, err)
	}
	if string(envrcContent) != templates.Envrc {
		t.Fatalf("%s content = %q, want %q", envrcPath, envrcContent, templates.Envrc)
	}

	codeRabbitContent, err := os.ReadFile(filepath.Join(tempDir, codeRabbitConfigPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", codeRabbitConfigPath, err)
	}
	if string(codeRabbitContent) != "custom coderabbit\n" {
		t.Fatalf("%s content = %q, want custom content", codeRabbitConfigPath, codeRabbitContent)
	}
	assertFileDoesNotExist(t, filepath.Join(tempDir, agentsMDPath))
}

func TestRunInitRefresh_ForceDoesNotOverwriteExistingScaffoldFilesWithoutFileTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	writeFile(t, filepath.Join(tempDir, envrcPath), "source_env .custom\n")
	writeFile(t, filepath.Join(tempDir, "docs", "agents", "GUARDRAILS.md"), "# Guardrails\n\nold\n")

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	envrcContent, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", envrcPath, err)
	}
	if string(envrcContent) != "source_env .custom\n" {
		t.Fatalf("%s content = %q, want custom content", envrcPath, envrcContent)
	}

	guardrailsContent, err := os.ReadFile(filepath.Join(tempDir, "docs", "agents", "GUARDRAILS.md"))
	if err != nil {
		t.Fatalf("failed to read GUARDRAILS.md: %v", err)
	}
	if string(guardrailsContent) != initTestSupportFileContent("docs/agents/GUARDRAILS.md") {
		t.Fatalf("expected generated docs support file to be overwritten on force, got:\n%s", guardrailsContent)
	}
}

func TestRunInitRefresh_RejectsUnsupportedFileTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initRefreshFiles = []string{"README.md"}

		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "not a Kit-managed refresh target") {
			t.Fatalf("expected unsupported target error, got %v", err)
		}
	})
}

func initTestSupportFileContent(relativePath string) string {
	for _, file := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		if file.RelativePath == relativePath {
			return file.Content
		}
	}
	return ""
}

func withInitFlags(t *testing.T, run func()) {
	t.Helper()

	originalCopy := initCopy
	originalOutputOnly := initOutputOnly
	originalRefresh := initRefresh
	originalForce := initForce
	originalRefreshFiles := initRefreshFiles

	t.Cleanup(func() {
		initCopy = originalCopy
		initOutputOnly = originalOutputOnly
		initRefresh = originalRefresh
		initForce = originalForce
		initRefreshFiles = originalRefreshFiles
	})

	initCopy = false
	initOutputOnly = false
	initRefresh = false
	initForce = false
	initRefreshFiles = nil

	run()
}

func setupInitHome(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	return homeDir
}
