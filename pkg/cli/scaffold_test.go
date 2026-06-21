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

func TestScaffoldWorkflowSubcommandCreatesV2SpecScaffold(t *testing.T) {
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
	assertFileExists(t, filepath.Join(projectRoot, "docs", "notes", "0001-sample-feature", ".gitkeep"))
	if _, err := os.Stat(filepath.Join(featurePath, "PLAN.md")); !os.IsNotExist(err) {
		t.Fatalf("scaffold spec should not create PLAN.md, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(featurePath, "TASKS.md")); !os.IsNotExist(err) {
		t.Fatalf("scaffold spec should not create TASKS.md, stat err=%v", err)
	}
}

func TestScaffoldCommandRegistersWorkflowSubcommands(t *testing.T) {
	for _, args := range [][]string{
		{"scaffold", "spec"},
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
	for _, args := range [][]string{
		{"scaffold", "brainstorm"},
		{"scaffold", "plan"},
		{"scaffold", "tasks"},
	} {
		cmd, _, err := rootCmd.Find(args)
		if err == nil && cmd != nil && cmd.CommandPath() == "kit "+strings.Join(args, " ") {
			t.Fatalf("expected %v to be removed from primary scaffold namespace", args)
		}
	}
}
