package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBrainstormPrepareCreatesScaffoldWithoutPrompt(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreFlags := setBrainstormFlagState(false, "", false, false, false, false)
	defer restoreFlags()
	brainstormPrepare = true

	cmd := newBrainstormTestCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runBrainstorm(cmd, []string{"sample-feature"}); err != nil {
		t.Fatalf("runBrainstorm() error = %v", err)
	}

	output := out.String()
	for _, check := range []string{
		"♻️ brainstorm directory and files empty scaffolding created.",
		"Please prepare your notes, documents, images, and examples for the brainstorm phase",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected prepare output to contain %q, got %q", check, output)
		}
	}
	if strings.Contains(output, "/plan") {
		t.Fatalf("expected prepare output not to include brainstorm prompt, got %q", output)
	}

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample-feature")
	assertFileExists(t, filepath.Join(featurePath, "BRAINSTORM.md"))
	assertFileExists(t, filepath.Join(projectRoot, "docs", "notes", "0001-sample-feature", ".gitkeep"))

	content, err := os.ReadFile(filepath.Join(featurePath, "BRAINSTORM.md"))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "target: docs/notes/0001-sample-feature") {
		t.Fatalf("expected brainstorm metadata to reference notes directory, got %q", string(content))
	}
}

func TestScaffoldWorkflowSubcommandsCreatePhaseFiles(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	specResult, err := scaffoldSpecWorkflow("sample-feature")
	if err != nil {
		t.Fatalf("scaffoldSpecWorkflow() error = %v", err)
	}
	if specResult.Feature.DirName != "0001-sample-feature" {
		t.Fatalf("unexpected feature dir %q", specResult.Feature.DirName)
	}
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-sample-feature")
	assertFileExists(t, filepath.Join(featurePath, "SPEC.md"))

	if _, err := scaffoldPlanWorkflow("sample-feature"); err != nil {
		t.Fatalf("scaffoldPlanWorkflow() error = %v", err)
	}
	assertFileExists(t, filepath.Join(featurePath, "PLAN.md"))

	if _, err := scaffoldTasksWorkflow("sample-feature"); err != nil {
		t.Fatalf("scaffoldTasksWorkflow() error = %v", err)
	}
	assertFileExists(t, filepath.Join(featurePath, "TASKS.md"))
}

func TestScaffoldPlanRequiresSpec(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	if err := os.MkdirAll(filepath.Join(projectRoot, "docs", "specs", "0001-sample"), 0755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	_, err = scaffoldPlanWorkflow("sample")
	if err == nil || !strings.Contains(err.Error(), "SPEC.md not found") {
		t.Fatalf("expected missing SPEC.md error, got %v", err)
	}
}

func TestScaffoldTasksRequiresPlan(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	if _, err := scaffoldSpecWorkflow("sample"); err != nil {
		t.Fatalf("scaffoldSpecWorkflow() error = %v", err)
	}

	_, err = scaffoldTasksWorkflow("sample")
	if err == nil || !strings.Contains(err.Error(), "PLAN.md not found") {
		t.Fatalf("expected missing PLAN.md error, got %v", err)
	}
}

func TestScaffoldCommandRegistersWorkflowSubcommands(t *testing.T) {
	for _, args := range [][]string{
		{"scaffold", "brainstorm"},
		{"scaffold", "spec"},
		{"scaffold", "plan"},
		{"scaffold", "tasks"},
		{"scaffold", "agents"},
	} {
		cmd, _, err := rootCmd.Find(args)
		if err != nil {
			t.Fatalf("rootCmd.Find(%v) error = %v", args, err)
		}
		if cmd == nil || cmd.Hidden {
			t.Fatalf("expected visible command for %v", args)
		}
	}
}
