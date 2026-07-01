package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

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

func TestRunInitRefresh_UpdatesManagedAutoAssignWorkflowWhenAssigneesChange(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	oldAssignees := []string{"old-owner"}
	newAssignees := []string{"jamesonstone", "octocat"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &newAssignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), templates.BuildAutoAssignWorkflow(oldAssignees))

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{`"jamesonstone"`, `"octocat"`} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected refreshed workflow to contain %q, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "old-owner") {
		t.Fatalf("managed workflow kept stale assignee:\n%s", content)
	}
}

func TestRunInitRefresh_DoesNotOverwriteCustomAutoAssignWorkflowWithoutForceTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &assignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	custom := "name: Custom auto assign\n"
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), custom)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if content != custom {
		t.Fatalf("custom workflow was overwritten:\n%s", content)
	}
}

func TestRunInitRefresh_FileForceOverwritesCustomAutoAssignWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &assignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), "name: Custom auto assign\n")

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{autoAssignWorkflowPath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if !strings.Contains(content, `"jamesonstone"`) || !strings.Contains(content, "# Kit-managed auto-assignment workflow.") {
		t.Fatalf("expected force-targeted refresh to write managed workflow, got:\n%s", content)
	}
}

func TestRunInitRefresh_AddsManagedReadmeBadgesAndStaysIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/widget"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, ".github", "workflows", "ci.yml"), "name: CI\n")
	writeFile(t, filepath.Join(tempDir, readmePath), "```text\nWIDGET\n\n                         useful widget service\n```\n\nWidget runs useful jobs for the Acme platform.\n\n## Install\n\nRun it.\n")

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	first := readFile(t, filepath.Join(tempDir, readmePath))
	for _, check := range []string{
		readmeBadgeStart,
		"img.shields.io/github/last-commit/acme/widget",
		"img.shields.io/github/issues/acme/widget",
		"img.shields.io/github/issues-pr/acme/widget",
		"github.com/acme/widget/actions/workflows/ci.yml/badge.svg",
		"img.shields.io/github/v/release/acme/widget",
		readmeBadgeEnd,
	} {
		if !strings.Contains(first, check) {
			t.Fatalf("expected README to contain %q, got:\n%s", check, first)
		}
	}
	if strings.Contains(strings.ToLower(first), "license") {
		t.Fatalf("README badges should not include a license badge, got:\n%s", first)
	}
	if !strings.Contains(first, "Widget runs useful jobs for the Acme platform.\n\n"+readmeBadgeStart+"\n") {
		t.Fatalf("expected badge block after opening paragraph, got:\n%s", first)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("second runInit() error = %v", err)
			}
		})
	})

	second := readFile(t, filepath.Join(tempDir, readmePath))
	if second != first {
		t.Fatalf("README refresh was not idempotent:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestRunInitRefresh_CreatesReadmeStarterWithManagedBadges(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/background-worker"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, readmePath))
	for _, check := range []string{
		"```text\nBACKGROUND WORKER",
		readmeStarterTagline,
		"Background Worker is a Kit-managed project.",
		"img.shields.io/github/last-commit/acme/background-worker",
		"img.shields.io/github/v/release/acme/background-worker",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected README starter to contain %q, got:\n%s", check, content)
		}
	}
}

func TestRunInitRefresh_RejectsUnsupportedFileTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initRefreshFiles = []string{"NOT_MANAGED.md"}

		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "not a Kit-managed refresh target") {
			t.Fatalf("expected unsupported target error, got %v", err)
		}
	})
}

func TestRunInitRefresh_DryRunDiffPrintsPlannedChangesWithoutWriting(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := "source_env .custom\n"
	writeFile(t, filepath.Join(tempDir, envrcPath), existing)

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initDiff = true
		initForce = true
		initRefreshFiles = []string{envrcPath}

		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	for _, check := range []string{
		"diff --git a/.envrc b/.envrc",
		"--- a/.envrc",
		"+++ b/.envrc",
		"-source_env .custom",
		"+dotenv_if_exists",
		"Dry run complete. Planned Created: 0, Updated: 1, Merged: 0, Skipped: 0",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected dry-run diff output to contain %q, got:\n%s", check, output)
		}
	}

	content, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", envrcPath, err)
	}
	if string(content) != existing {
		t.Fatalf("dry run wrote %s; content = %q, want %q", envrcPath, content, existing)
	}
}

func TestRunInitRefresh_DryRunDoesNotWritePlannedRefresh(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, agentsMDPath), templates.LegacyAgentsMD)

	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	unchanged, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if unchanged.InstructionScaffoldVersion != config.InstructionScaffoldVersionVerbose {
		t.Fatalf("dry run updated config version = %d", unchanged.InstructionScaffoldVersion)
	}

	agentsContent, err := os.ReadFile(filepath.Join(tempDir, agentsMDPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(agentsContent) != templates.LegacyAgentsMD {
		t.Fatalf("dry run updated %s", agentsMDPath)
	}
	assertFileDoesNotExist(t, filepath.Join(tempDir, "docs", "agents", "README.md"))
}
