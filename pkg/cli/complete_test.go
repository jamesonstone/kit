package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func TestEligibleFeaturesForCompletion(t *testing.T) {
	specsDir := filepath.Join(t.TempDir(), "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	createFeatureTasks(t, specsDir, "0001-alpha", "- [x] done\n")
	createFeatureTasks(t, specsDir, "0002-beta", "- [x] done\n\n"+feature.ReflectionCompleteMarker+"\n")
	createFeatureFile(t, specsDir, "0003-gamma", "SPEC.md", "# SPEC\n")
	cfg := config.Default()

	candidates, err := eligibleFeaturesForCompletion(specsDir, cfg)
	if err != nil {
		t.Fatalf("eligibleFeaturesForCompletion() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Slug != "alpha" {
		t.Fatalf("eligibleFeaturesForCompletion() = %+v, want only alpha", candidates)
	}
}

func TestEligibleFeaturesForCompletion_ExcludesPaused(t *testing.T) {
	specsDir := filepath.Join(t.TempDir(), "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	createFeatureTasks(t, specsDir, "0001-alpha", "- [x] done\n")
	createFeatureTasks(t, specsDir, "0002-beta", "- [x] done\n")
	cfg := config.Default()
	cfg.SetFeaturePaused("0002-beta", true)

	candidates, err := eligibleFeaturesForCompletion(specsDir, cfg)
	if err != nil {
		t.Fatalf("eligibleFeaturesForCompletion() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Slug != "alpha" {
		t.Fatalf("eligibleFeaturesForCompletion() = %+v, want only alpha", candidates)
	}
}

func TestMarkFeaturesCompletePreflightPreventsPartialCompletion(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	featureA := createFeatureTasks(t, specsDir, "0001-alpha", "- [x] done\n")
	featureB := createFeatureTasks(t, specsDir, "0002-beta", "- [ ] todo\n")
	cfg := config.Default()

	err := markFeaturesComplete(
		&bytes.Buffer{},
		&bytes.Buffer{},
		[]feature.Feature{featureA, featureB},
		false,
		projectRoot,
		cfg,
	)
	if err == nil {
		t.Fatal("markFeaturesComplete() error = nil, want error")
	}

	for _, tasksPath := range []string{
		filepath.Join(featureA.Path, "TASKS.md"),
		filepath.Join(featureB.Path, "TASKS.md"),
	} {
		data, readErr := os.ReadFile(tasksPath)
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error = %v", tasksPath, readErr)
		}
		if strings.Contains(string(data), feature.ReflectionCompleteMarker) {
			t.Fatalf("expected no reflection marker in %s after failed preflight", tasksPath)
		}
	}
}

func TestMarkFeaturesCompleteAllTargets(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	featureA := createFeatureTasks(t, specsDir, "0001-alpha", "- [x] done\n")
	featureB := createFeatureTasks(t, specsDir, "0002-beta", "- [x] done\n")
	cfg := config.Default()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	if err := markFeaturesComplete(
		out,
		errOut,
		[]feature.Feature{featureA, featureB},
		false,
		projectRoot,
		cfg,
	); err != nil {
		t.Fatalf("markFeaturesComplete() error = %v", err)
	}

	for _, feat := range []feature.Feature{featureA, featureB} {
		data, readErr := os.ReadFile(filepath.Join(feat.Path, "TASKS.md"))
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error = %v", feat.Path, readErr)
		}
		if !strings.Contains(string(data), feature.ReflectionCompleteMarker) {
			t.Fatalf("expected reflection marker in %s", feat.Path)
		}
		if !strings.Contains(out.String(), feat.Slug) {
			t.Fatalf("expected output to mention %s, got %q", feat.Slug, out.String())
		}
	}

	if errOut.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", errOut.String())
	}
	if _, statErr := os.Stat(filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")); statErr != nil {
		t.Fatalf("expected PROJECT_PROGRESS_SUMMARY.md to be written, got %v", statErr)
	}
}

func TestRunComplete_RemovesCompletedFeatureFromActiveStatus(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	createFeatureTasks(t, specsDir, "0001-latest-complete", "- [x] done\n")

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	previousForce, previousAll := completeForce, completeAll
	completeForce = false
	completeAll = false
	defer func() {
		completeForce = previousForce
		completeAll = previousAll
	}()

	completeCmd := &cobra.Command{}
	completeOut := &bytes.Buffer{}
	completeCmd.SetOut(completeOut)
	completeCmd.SetErr(&bytes.Buffer{})

	if err := runComplete(completeCmd, []string{"latest-complete"}); err != nil {
		t.Fatalf("runComplete() error = %v", err)
	}

	statusCmd := &cobra.Command{}
	statusCmd.Flags().Bool("json", false, "")
	statusCmd.Flags().Bool("all", false, "")
	if err := statusCmd.Flags().Set("all", "true"); err != nil {
		t.Fatalf("Flags().Set(all) error = %v", err)
	}
	statusOut := &bytes.Buffer{}
	statusCmd.SetOut(statusOut)

	if err := runStatus(statusCmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	content := statusOut.String()
	if !strings.Contains(content, "Active feature: none in progress") {
		t.Fatalf("expected no active feature after completion, got %q", content)
	}
	if !strings.Contains(content, "0001-latest-complete") {
		t.Fatalf("expected completed feature in status output, got %q", content)
	}
	if !strings.Contains(content, "COMPLETE") {
		t.Fatalf("expected completed state in status output, got %q", content)
	}
	if strings.Contains(content, "ACTIVE") {
		t.Fatalf("expected completed feature row to avoid ACTIVE state, got %q", content)
	}
}

func createFeatureTasks(t *testing.T, specsDir, dirName, tasks string) feature.Feature {
	t.Helper()
	featurePath := filepath.Join(specsDir, dirName)
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, "TASKS.md"), []byte(tasks), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	number, slug, ok := feature.ParseDirName(dirName)
	if !ok {
		t.Fatalf("ParseDirName(%q) failed", dirName)
	}
	return feature.Feature{
		Number:  number,
		Slug:    slug,
		DirName: dirName,
		Path:    featurePath,
		Phase:   feature.DeterminePhase(featurePath),
	}
}

func createFeatureFile(t *testing.T, specsDir, dirName, fileName, content string) {
	t.Helper()
	featurePath := filepath.Join(specsDir, dirName)
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, fileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
