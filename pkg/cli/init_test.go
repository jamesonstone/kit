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

func TestRunInit_CreatesAutoAssignWorkflowFromGlobalFallback(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	global := config.Default()
	global.GitHub.DefaultAssignees = &assignees
	if _, _, err := config.PopulateGlobalConfig(global); err != nil {
		t.Fatalf("config.PopulateGlobalConfig() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{
		"# Kit-managed auto-assignment workflow.",
		"pull_request_target:",
		"continue-on-error: true",
		`"jamesonstone"`,
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected auto-assign workflow to contain %q, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "actions/checkout") {
		t.Fatalf("auto-assign workflow must not check out PR code:\n%s", content)
	}
}

func TestRunInit_UsesProjectAutoAssignAssigneesBeforeGlobalFallback(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	globalAssignees := []string{"jamesonstone"}
	global := config.Default()
	global.GitHub.DefaultAssignees = &globalAssignees
	if _, _, err := config.PopulateGlobalConfig(global); err != nil {
		t.Fatalf("config.PopulateGlobalConfig() error = %v", err)
	}
	projectAssignees := []string{"octocat", "@hubot"}
	project := config.Default()
	project.GitHub.DefaultAssignees = &projectAssignees
	if err := config.Save(tempDir, project); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{`"octocat"`, `"hubot"`} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected project assignee %q in workflow, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "jamesonstone") {
		t.Fatalf("project assignees should take precedence over global fallback:\n%s", content)
	}
}

func TestRunInit_ExplicitEmptyProjectAutoAssignAssigneesSkipsGlobalFallback(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	globalAssignees := []string{"jamesonstone"}
	global := config.Default()
	global.GitHub.DefaultAssignees = &globalAssignees
	if _, _, err := config.PopulateGlobalConfig(global); err != nil {
		t.Fatalf("config.PopulateGlobalConfig() error = %v", err)
	}
	projectAssignees := []string{}
	project := config.Default()
	project.GitHub.DefaultAssignees = &projectAssignees
	if err := config.Save(tempDir, project); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if strings.Contains(content, "jamesonstone") {
		t.Fatalf("explicit empty project assignees should not fall back to global config:\n%s", content)
	}
	if !strings.Contains(content, "const assignees = [];") {
		t.Fatalf("expected explicit empty project assignees to render a no-op workflow, got:\n%s", content)
	}
}

func TestRunInit_CreatesNonBlockingAutoAssignWorkflowWithoutAssignees(t *testing.T) {
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

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{
		"const assignees = [];",
		"No Kit auto-assignees configured; skipping.",
		"continue-on-error: true",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected no-assignee workflow to contain %q, got:\n%s", check, content)
		}
	}
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

func TestRunInit_InstallsRegistryRulesetsAndState(t *testing.T) {
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

	rulesetContent, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(registry.Slug)))
	if err != nil {
		t.Fatalf("expected registry ruleset to be installed by kit init: %v", err)
	}
	if !strings.Contains(string(rulesetContent), "slug: safety-guardrails") {
		t.Fatalf("unexpected ruleset content:\n%s", rulesetContent)
	}

	created, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := created.RegistryArtifact(rulesetKind, registry.Slug)
	if !ok {
		t.Fatalf("expected registry artifact for %s", registry.Slug)
	}
	if artifact.State != registryArtifactStateManaged || artifact.InstalledHash != registry.NormalizedHash {
		t.Fatalf("artifact = %#v, want managed hash %s", artifact, registry.NormalizedHash)
	}
}

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
