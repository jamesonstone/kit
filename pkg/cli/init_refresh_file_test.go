package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
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
			if err := runInitForTest(initCmd, nil); err != nil {
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
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, makefilePath))
	if content != templates.Makefile {
		t.Fatalf("%s content = %q, want %q", makefilePath, content, templates.Makefile)
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
			if err := runInitForTest(initCmd, nil); err != nil {
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
	writeFile(t, filepath.Join(tempDir, "docs", "references", "testing.md"), "# Testing Reference\n\nold\n")

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
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

	guardrailsContent, err := os.ReadFile(filepath.Join(tempDir, "docs", "references", "testing.md"))
	if err != nil {
		t.Fatalf("failed to read docs/references/testing.md: %v", err)
	}
	// The testing reference is project-owned once created; --force never replaces it.
	if string(guardrailsContent) != "# Testing Reference\n\nold\n" {
		t.Fatalf("expected project testing reference to be preserved on force, got:\n%s", guardrailsContent)
	}
}
