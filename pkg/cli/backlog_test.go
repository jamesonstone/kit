package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestRunBrainstormBacklog_CreatesPausedBacklogItem(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-current-endpoint", "SPEC.md", "# SPEC\n")

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need to refactor the legacy endpoint to share normalization.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(true, "", false, false, false, false)
	defer restoreFlags()

	cmd := newBrainstormTestCommand()
	if err := runBrainstorm(cmd, []string{"legacy-endpoint-refactor"}); err != nil {
		t.Fatalf("runBrainstorm() error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if !updated.IsFeaturePaused("0002-legacy-endpoint-refactor") {
		t.Fatal("expected backlog item to persist as paused")
	}

	brainstormPath := filepath.Join(specsDir, "0002-legacy-endpoint-refactor", "BRAINSTORM.md")
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", brainstormPath, err)
	}
	if !strings.Contains(string(content), "Need to refactor the legacy endpoint to share normalization.") {
		t.Fatalf("expected thesis in brainstorm, got %q", string(content))
	}
	doc := document.Parse(string(content), brainstormPath, document.TypeBrainstorm)
	relationships, warnings := doc.Relationships()
	if len(warnings) > 0 || len(relationships) != 1 || relationships[0].Type != "related to" || relationships[0].Target != "0001-current-endpoint" {
		t.Fatalf("expected front matter relationship to active feature, got relationships=%#v warnings=%#v content=%q", relationships, warnings, string(content))
	}
	references := doc.References()
	if len(references) == 0 || references[0].Name != featureNotesReferenceName || references[0].Target != "docs/notes/0002-legacy-endpoint-refactor" {
		t.Fatalf("expected feature notes reference, got %#v content=%q", references, string(content))
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "docs", "notes", "0002-legacy-endpoint-refactor", ".gitkeep")); err != nil {
		t.Fatalf("expected backlog notes .gitkeep, got %v", err)
	}

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	summary, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", summaryPath, err)
	}
	if !strings.Contains(string(summary), "| 0002 | legacy-endpoint-refactor | `docs/specs/0002-legacy-endpoint-refactor` | brainstorm | yes |") {
		t.Fatalf("expected paused brainstorm row in rollup, got %q", string(summary))
	}
}

func TestRunBacklog_ListsDeferredItems(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-current-endpoint", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0002-legacy-endpoint-refactor", "BRAINSTORM.md", `# BRAINSTORM

## SUMMARY

Need to refactor the legacy endpoint to share normalization.

## USER THESIS

same
`)
	cfg.SetFeaturePaused("0002-legacy-endpoint-refactor", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreFlags := setBacklogFlagState(false, false, false)
	defer restoreFlags()

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runBacklog(cmd, nil); err != nil {
		t.Fatalf("runBacklog() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "Feature") || !strings.Contains(content, "Description") {
		t.Fatalf("expected fixed-width backlog headers, got %q", content)
	}
	if strings.Contains(content, "| feature | description |") {
		t.Fatalf("expected fixed-width table instead of markdown, got %q", content)
	}
	if !strings.Contains(content, "legacy-endpoint-refactor") {
		t.Fatalf("expected backlog feature row, got %q", content)
	}
	if !strings.Contains(content, "Need to refactor the legacy endpoint to share normalization.") {
		t.Fatalf("expected backlog row, got %q", content)
	}
}

func TestRunBacklogPickup_ClearsPausedState(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-legacy-endpoint-refactor", "BRAINSTORM.md", `# BRAINSTORM

## SUMMARY

Need to refactor the legacy endpoint to share normalization.

## USER THESIS

same
`)
	cfg.SetFeaturePaused("0001-legacy-endpoint-refactor", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreFlags := setBacklogFlagState(true, false, true)
	defer restoreFlags()
	restoreStdout := silenceStdout(t)
	defer restoreStdout()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}

	if err := runBacklog(cmd, []string{"legacy-endpoint-refactor"}); err != nil {
		t.Fatalf("runBacklog() error = %v", err)
	}

	updated, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.IsFeaturePaused("0001-legacy-endpoint-refactor") {
		t.Fatal("expected backlog pickup to clear paused state")
	}
}

func TestRunStatus_NoActiveFeatureWhenOnlyBacklogItemsExist(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureFile(t, specsDir, "0001-legacy-endpoint-refactor", "BRAINSTORM.md", "# BRAINSTORM\n")
	cfg.SetFeaturePaused("0001-legacy-endpoint-refactor", true)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := runStatus(cmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "No active feature in progress") {
		t.Fatalf("expected no-active message, got %q", content)
	}
	if !strings.Contains(content, "kit backlog") {
		t.Fatalf("expected backlog guidance, got %q", content)
	}
}
