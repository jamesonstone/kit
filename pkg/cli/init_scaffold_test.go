package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

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

func TestRunInit_CreatesMakefileStarter(t *testing.T) {
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

		content, err := os.ReadFile(filepath.Join(tempDir, makefilePath))
		if err != nil {
			t.Fatalf("expected %s to be created: %v", makefilePath, err)
		}
		if string(content) != templates.Makefile {
			t.Fatalf("%s content = %q, want %q", makefilePath, content, templates.Makefile)
		}
	})
}

func TestRunInit_PreservesExistingMakefile(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := ".PHONY: dev\n\ndev:\n\tnpm run dev\n"
	if err := os.WriteFile(filepath.Join(tempDir, makefilePath), []byte(existing), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", makefilePath, err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})

		content, err := os.ReadFile(filepath.Join(tempDir, makefilePath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", makefilePath, err)
		}
		if string(content) != existing {
			t.Fatalf("%s content = %q, want %q", makefilePath, content, existing)
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
