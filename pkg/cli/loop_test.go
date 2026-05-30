package cli

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/jamesonstone/kit/internal/verify"
)

func TestResolveStrictLoopStage(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	featurePath := filepath.Join(specsDir, "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	feat := &feature.Feature{Slug: "alpha", DirName: "0001-alpha", Path: featurePath}

	state := resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageSpec || len(state.Diagnostics) == 0 {
		t.Fatalf("missing SPEC stage = %#v, want spec diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName)))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageSpec || len(state.Diagnostics) == 0 {
		t.Fatalf("placeholder SPEC stage = %#v, want spec diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validSpecWithRelationships("none\n"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStagePlan || len(state.Diagnostics) == 0 {
		t.Fatalf("valid SPEC stage = %#v, want plan diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "PLAN.md"), validPlan())
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageTasks || len(state.Diagnostics) == 0 {
		t.Fatalf("valid PLAN stage = %#v, want tasks diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "TASKS.md"), validTasks())
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageImplement || state.TasksTotal != 1 || state.TasksDone != 0 {
		t.Fatalf("incomplete TASKS stage = %#v, want implement 0/1", state)
	}

	writeFile(t, filepath.Join(featurePath, "TASKS.md"), completedTasksWithoutReflection())
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageReflect || state.TasksTotal != 1 || state.TasksDone != 1 {
		t.Fatalf("complete TASKS stage = %#v, want reflect 1/1", state)
	}

	writeFile(t, filepath.Join(featurePath, "TASKS.md"), completedTasksWithoutReflection()+"\n"+feature.ReflectionCompleteMarker+"\n")
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageComplete {
		t.Fatalf("reflection marker stage = %#v, want complete", state)
	}
}

func TestLoopCommandSuppressesDuplicateErrorOutput(t *testing.T) {
	if !loopCmd.SilenceUsage {
		t.Fatal("loopCmd.SilenceUsage = false, want true")
	}
	if !loopCmd.SilenceErrors {
		t.Fatal("loopCmd.SilenceErrors = false, want true")
	}
}

func TestResolveLoopFeatureWithoutArgShowsSelector(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	writeFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), validSpecWithRelationships("none\n"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "SPEC.md"), validSpecWithRelationships("none\n"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "PLAN.md"), validPlan())
	writeFile(t, filepath.Join(specsDir, "0002-beta", "TASKS.md"), completedTasksWithoutReflection()+"\n"+feature.ReflectionCompleteMarker+"\n")
	cfg := config.Default()

	var selected *feature.Feature
	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			var err error
			selected, err = resolveLoopFeature(specsDir, cfg, nil)
			if err != nil {
				t.Fatalf("resolveLoopFeature() error = %v", err)
			}
		})
	})

	if selected == nil || selected.DirName != "0001-alpha" {
		t.Fatalf("selected feature = %#v, want 0001-alpha", selected)
	}
	if !strings.Contains(output, "Select a feature to loop:") {
		t.Fatalf("expected selector output, got:\n%s", output)
	}
	if strings.Contains(output, "0002-beta") {
		t.Fatalf("complete feature should be omitted from loop selector, got:\n%s", output)
	}
}

func TestLoopFeatureCandidatesExcludeCompleteFeatures(t *testing.T) {
	projectRoot := t.TempDir()
	specsDir := filepath.Join(projectRoot, "docs", "specs")
	writeFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), validSpecWithRelationships("none\n"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "SPEC.md"), validSpecWithRelationships("none\n"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "PLAN.md"), validPlan())
	writeFile(t, filepath.Join(specsDir, "0002-beta", "TASKS.md"), completedTasksWithoutReflection()+"\n"+feature.ReflectionCompleteMarker+"\n")

	candidates, err := loopFeatureCandidates(specsDir, config.Default())
	if err != nil {
		t.Fatalf("loopFeatureCandidates() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].DirName != "0001-alpha" {
		t.Fatalf("candidates = %#v, want only 0001-alpha", candidates)
	}
}

func TestExecuteLoopRunsConfiguredAgentUntilComplete(t *testing.T) {
	projectRoot := setupLoopProject(t, "agent.sh", loopAgentScript(99, true, true))
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	feat, err := feature.Resolve(cfg.SpecsPath(projectRoot), "alpha")
	if err != nil {
		t.Fatalf("feature.Resolve() error = %v", err)
	}

	report, err := executeLoop(context.Background(), loopOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Feature:     feat,
		Until:       loopStageComplete,
		Agent:       cfg.Loop.Agent,
	})
	if err != nil {
		t.Fatalf("executeLoop() error = %v\nreport=%#v", err, report)
	}
	if report.Status != "complete" {
		t.Fatalf("Status = %q, want complete", report.Status)
	}
	if len(report.Iterations) != 5 {
		t.Fatalf("Iterations = %d, want 5", len(report.Iterations))
	}
	if report.ArtifactDir == "" {
		t.Fatal("ArtifactDir is empty")
	}
	if _, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(report.ArtifactDir), "run.json")); err != nil {
		t.Fatalf("expected run artifact: %v", err)
	}
	if state := resolveStrictLoopStage(projectRoot, feat); state.Stage != loopStageComplete {
		t.Fatalf("final stage = %#v, want complete", state)
	}
}

func TestExecuteLoopStopsOnLowConfidence(t *testing.T) {
	projectRoot := setupLoopProject(t, "agent.sh", loopAgentScript(80, true, true))
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	feat, err := feature.Resolve(cfg.SpecsPath(projectRoot), "alpha")
	if err != nil {
		t.Fatalf("feature.Resolve() error = %v", err)
	}

	report, err := executeLoop(context.Background(), loopOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Feature:     feat,
		Until:       loopStageComplete,
		Agent:       cfg.Loop.Agent,
	})
	if err == nil || !strings.Contains(err.Error(), "confidence 80") {
		t.Fatalf("expected low confidence error, got %v", err)
	}
	if report.Status != "stopped" {
		t.Fatalf("Status = %q, want stopped", report.Status)
	}
}

func TestExecuteLoopStopsOnMalformedAgentResult(t *testing.T) {
	projectRoot := setupLoopProject(t, "agent.sh", loopAgentScript(99, false, true))
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	feat, err := feature.Resolve(cfg.SpecsPath(projectRoot), "alpha")
	if err != nil {
		t.Fatalf("feature.Resolve() error = %v", err)
	}

	report, err := executeLoop(context.Background(), loopOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Feature:     feat,
		Until:       loopStageComplete,
		Agent:       cfg.Loop.Agent,
	})
	if err == nil || !strings.Contains(err.Error(), "KIT_LOOP_RESULT") {
		t.Fatalf("expected malformed result error, got %v", err)
	}
	if report.Status != "stopped" {
		t.Fatalf("Status = %q, want stopped", report.Status)
	}
}

func TestExecuteLoopDryRunDoesNotRequireAgentConfig(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	feat := &feature.Feature{Slug: "alpha", DirName: "0001-alpha", Path: featurePath}

	report, err := executeLoop(context.Background(), loopOptions{
		ProjectRoot: projectRoot,
		Config:      cfg,
		Feature:     feat,
		Until:       loopStageComplete,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("executeLoop() dry-run error = %v", err)
	}
	if report.Status != "dry_run" {
		t.Fatalf("Status = %q, want dry_run", report.Status)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".kit", "loops")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write loop artifacts, stat err=%v", err)
	}
}

func TestStopOnFailedVerificationUsesCurrentLoopEvidence(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	feat := &feature.Feature{Slug: "alpha", DirName: "0001-alpha", Path: featurePath}
	oldStartedAt := time.Now().Add(-time.Hour).UTC()
	oldRun := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         verify.NewRunID(oldStartedAt),
		Feature:       verify.FeatureRefFromDir(featurePath),
		Status:        verify.RunStatusFail,
		StartedAt:     oldStartedAt,
		EndedAt:       oldStartedAt.Add(time.Second),
	}
	if err := runstore.Write(projectRoot, &oldRun); err != nil {
		t.Fatalf("runstore.Write(old) error = %v", err)
	}
	if err := stopOnFailedVerification(projectRoot, feat, time.Now().UTC()); err != nil {
		t.Fatalf("stale failed verification should not stop loop: %v", err)
	}

	newStartedAt := time.Now().UTC()
	newRun := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         verify.NewRunID(newStartedAt),
		Feature:       verify.FeatureRefFromDir(featurePath),
		Status:        verify.RunStatusFail,
		StartedAt:     newStartedAt,
		EndedAt:       newStartedAt.Add(time.Second),
	}
	if err := runstore.Write(projectRoot, &newRun); err != nil {
		t.Fatalf("runstore.Write(new) error = %v", err)
	}
	err := stopOnFailedVerification(projectRoot, feat, newStartedAt.Add(-time.Second))
	if err == nil || !strings.Contains(err.Error(), newRun.RunID) {
		t.Fatalf("expected current failed verification error, got %v", err)
	}
}

func setupLoopProject(t *testing.T, agentName, agentScript string) string {
	t.Helper()
	projectRoot := t.TempDir()
	agentPath := filepath.Join(projectRoot, agentName)
	if err := os.WriteFile(agentPath, []byte(agentScript), 0755); err != nil {
		t.Fatalf("WriteFile(agent) error = %v", err)
	}
	cfg := config.Default()
	cfg.Loop.MinConfidence = 95
	cfg.Loop.MaxIterations = 10
	cfg.Loop.Agent = config.LoopAgentConfig{Command: agentPath}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll(feature) error = %v", err)
	}
	return projectRoot
}

func loopAgentScript(confidence int, emitResult bool, mutate bool) string {
	result := ""
	if emitResult {
		result = `printf 'KIT_LOOP_RESULT: {"stage":"%s","status":"done","confidence":` + fmtInt(confidence) + `,"blockers":[]}\n' "$KIT_LOOP_STAGE"`
	} else {
		result = `echo "done without loop result"`
	}
	mutations := ""
	if mutate {
		mutations = `case "$KIT_LOOP_STAGE" in
  spec)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validSpecWithRelationships("none\n") + `DOC
    ;;
  plan)
    cat > docs/specs/0001-alpha/PLAN.md <<'DOC'
` + validPlan() + `DOC
    ;;
  tasks)
    cat > docs/specs/0001-alpha/TASKS.md <<'DOC'
` + validTasks() + `DOC
    ;;
  implement)
    cat > docs/specs/0001-alpha/TASKS.md <<'DOC'
` + completedTasksWithoutReflection() + `DOC
    ;;
  reflect)
    printf '\n` + feature.ReflectionCompleteMarker + `\n' >> docs/specs/0001-alpha/TASKS.md
    ;;
esac
`
	}
	return `#!/bin/sh
set -eu
cat >/dev/null
` + mutations + result + `
`
}

func completedTasksWithoutReflection() string {
	return strings.ReplaceAll(
		strings.Replace(validTasks(), "| T001 | sample task | todo | agent | |", "| T001 | sample task | done | agent | |", 1),
		"- [ ] T001: sample task",
		"- [x] T001: sample task",
	)
}

func fmtInt(value int) string {
	return strconv.Itoa(value)
}
