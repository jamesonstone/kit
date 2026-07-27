package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInit_DefaultCopiesBootstrapPromptAndShowsPasteStep(t *testing.T) {
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
		const agentsEntrypointGuidance = "Read docs/agents/README.md before inspecting repository evidence or modifying project memory"
		if !strings.Contains(copied, "Treat the exact generated starter at "+filepath.Join(cwd, constitutionPath)+" as a valid bootstrap Constitution") {
			t.Fatalf("expected copied prompt to target %s, got %q", filepath.Join(cwd, constitutionPath), copied)
		}
		for _, check := range []string{
			agentsEntrypointGuidance,
			"Do not ask the user to explain the entire project",
			"When evidence is insufficient, leave the project-specific starter sections unchanged",
			"Follow docs/references/rules/constitution-curation.md",
			"Populate " + filepath.Join(cwd, makefilePath) + " with a canonical project command interface",
			"Leave the safe starter Makefile unchanged when the repository has no verified",
			"Expose `make dev` when the repository has a verified local development or run workflow",
			"Do not leave TODO recipes, echo-only placeholders, guessed commands, or duplicated build logic",
			"Run `make help` and each added target that is safe to execute",
		} {
			if !strings.Contains(copied, check) {
				t.Fatalf("expected copied prompt to contain %q, got %q", check, copied)
			}
		}
		entrypointIndex := strings.Index(copied, agentsEntrypointGuidance)
		bootstrapIndex := strings.Index(copied, "Treat the exact generated starter")
		evidenceIndex := strings.Index(copied, "Inspect implemented behavior")
		if entrypointIndex < 0 || bootstrapIndex < 0 || evidenceIndex < 0 || entrypointIndex >= bootstrapIndex || entrypointIndex >= evidenceIndex {
			t.Fatalf("expected repository entrypoint guidance before Constitution inspection, got %q", copied)
		}
		if strings.Contains(copied, "Copy this section to the Agent:") {
			t.Fatalf("expected clipboard payload to contain only the prompt body, got %q", copied)
		}
		if !strings.Contains(output, "Copied the prepared text to the clipboard.") {
			t.Fatalf("expected stdout to acknowledge clipboard copy, got %q", output)
		}
		if !strings.Contains(output, "1. Paste the copied prompt into your agent to review repository evidence and populate only verified Makefile targets") {
			t.Fatalf("expected stdout to include the paste guidance, got %q", output)
		}
		if strings.Contains(copied, "Please update "+filepath.Join(cwd, constitutionPath)+" with all patterns") {
			t.Fatalf("expected prompt not to demand exhaustive Constitution drafting, got %q", copied)
		}
		if strings.Contains(output, "Initialize project memory and verified command entrypoints") {
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
		if !strings.HasPrefix(output, "Initialize project memory and verified command entrypoints for the repository at "+cwd+".") {
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
