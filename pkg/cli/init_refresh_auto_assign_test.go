package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

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
