package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// Agents and scripts run without a terminal: the prompt goes to stdout and the
// user's clipboard is left alone.
func TestPromptGoesToStdoutWithoutTerminal(t *testing.T) {
	stubStdoutTerminal(t, false)
	copied := ""
	previous := clipboardCopyFunc
	clipboardCopyFunc = func(text string) error { copied = text; return nil }
	t.Cleanup(func() { clipboardCopyFunc = previous })

	out := captureStdout(t, func() {
		if err := writePromptWithClipboardDefault("prompt text", false, false); err != nil {
			t.Fatal(err)
		}
	})
	if copied != "" || !strings.Contains(out, "prompt text") {
		t.Fatalf("copied = %q, stdout = %q", copied, out)
	}
}

func TestInitRefusesInsideExistingKitProject(t *testing.T) {
	setupMigrationEnvironment(t)
	parent := freshInitProject(t)
	child := filepath.Join(parent, "services", "api")
	writeFile(t, filepath.Join(child, "README.md"), "# api\n")
	setWorkingDirectory(t, child)
	withInitFlags(t, func() {
		initOutputOnly = true
		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "inside the Kit project at") || !strings.Contains(err.Error(), filepath.Base(parent)) {
			t.Fatalf("runInit() error = %v, want refusal naming the parent project", err)
		}
	})
	assertFileDoesNotExist(t, filepath.Join(child, ".kit.yaml"))
	assertFileDoesNotExist(t, filepath.Join(child, "AGENTS.md"))
}

// status and registry status agree with health: a clean clone missing only the
// local .env and .envrc is current.
func TestStatusAgreesWithHealthOnCleanClone(t *testing.T) {
	setupMigrationEnvironment(t)
	source := freshInitProject(t)
	clone := filepath.Join(t.TempDir(), "clone")
	runGitForSourceAuditTest(t, source, "clone", "-q", source, clone)
	setWorkingDirectory(t, clone)
	cfg := loadMigratedConfig(t, clone)
	report, err := buildRegistryStatusReport(clone, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if report.State != statusKitManagedStateCurrent || report.PlannedChanges != 0 || len(report.Items) != 0 {
		t.Fatalf("registry status on a current clean clone = %#v", report)
	}
}
