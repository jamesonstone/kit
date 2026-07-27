package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

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
		"Delivery of command-created files:",
		"exact command-owned path snapshot",
		"explicitly stage only the captured paths (including deleted paths)",
		"restore each captured root path to its exact pre-command state",
		"create or update the ready pull request",
		"no documentation updates needed",
	} {
		if !strings.Contains(copied, want) {
			t.Fatalf("expected copied prompt to contain %q, got:\n%s", want, copied)
		}
	}
}
