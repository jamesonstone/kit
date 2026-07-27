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

func TestRunInitRefresh_CreatesMissingMakefile(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{makefilePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, makefilePath))
	if content != templates.Makefile {
		t.Fatalf("%s content = %q, want %q", makefilePath, content, templates.Makefile)
	}
}

func TestRunInitRefresh_PrintsManagedFileDeliveryStepsAfterWrite(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runInitRefresh(tempDir, initRefreshOptions{
			files: []string{makefilePath},
		}); err != nil {
			t.Fatalf("runInitRefresh() error = %v", err)
		}
	})

	for _, check := range []string{
		"`Makefile` (create; pre-command absent; expected sha256:",
		"abort if a captured destination path has staged, working-tree, or untracked changes",
		"restore each captured root path to its exact pre-command state",
		"create or update the ready pull request",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected refresh delivery guidance to contain %q, got:\n%s", check, output)
		}
	}
}

func TestRunInitRefresh_CreatesV3WorktreeReference(t *testing.T) {
	tempDir := t.TempDir()
	relativePath := "docs/references/worktrees.md"

	if err := runInitRefresh(tempDir, initRefreshOptions{
		files:      []string{relativePath},
		outputOnly: true,
	}); err != nil {
		t.Fatalf("runInitRefresh() error = %v", err)
	}

	var expected string
	for _, support := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionMemory) {
		if support.RelativePath == relativePath {
			expected = support.Content
			break
		}
	}
	if expected == "" {
		t.Fatal("V3 instruction support files do not include the worktree reference")
	}
	if content := readFile(t, filepath.Join(tempDir, relativePath)); content != expected {
		t.Fatalf("%s content did not match V3 support template", relativePath)
	}
}

func TestRunInitRefresh_FileForcePreservesExistingMakefile(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := ".PHONY: dev\n\ndev:\n\tgo run ./cmd/server\n"
	writeFile(t, filepath.Join(tempDir, makefilePath), existing)

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{makefilePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, makefilePath))
	if content != existing {
		t.Fatalf("%s content = %q, want custom content %q", makefilePath, content, existing)
	}
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
