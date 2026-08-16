package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

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
