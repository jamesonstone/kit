package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/verify"
)

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
		ProjectRoot:   projectRoot,
		Config:        cfg,
		Feature:       feat,
		Until:         loopStageComplete,
		Agent:         cfg.Loop.Agent,
		ReflectRunner: loopReflectRunnerForTest(),
		ReflectNow: func() time.Time {
			return time.Unix(1700001800, 0).UTC()
		},
	})
	if err != nil {
		t.Fatalf("executeLoop() error = %v\nreport=%#v", err, report)
	}
	if report.Status != "complete" {
		t.Fatalf("Status = %q, want complete", report.Status)
	}
	if len(report.Iterations) != 6 {
		t.Fatalf("Iterations = %d, want 6", len(report.Iterations))
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
	var verdict ReflectVerdict
	reflectData, err := os.ReadFile(filepath.Join(feat.Path, reflectVerdictFileName))
	if err != nil {
		t.Fatalf("expected REFLECT.json to be written: %v", err)
	}
	if err := json.Unmarshal(reflectData, &verdict); err != nil {
		t.Fatalf("REFLECT.json is invalid JSON: %v\n%s", err, string(reflectData))
	}
	if !verdict.TestsPass || verdict.ScopeDrift != "none" || verdict.Timestamp == "" {
		t.Fatalf("unexpected reflect verdict: %#v", verdict)
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
