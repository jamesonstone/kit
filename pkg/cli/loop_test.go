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
	if state.Stage != loopStageClarify || len(state.Diagnostics) == 0 {
		t.Fatalf("missing SPEC stage = %#v, want clarify diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName)))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageClarify || len(state.Diagnostics) == 0 {
		t.Fatalf("placeholder SPEC stage = %#v, want clarify diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "clarify"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageClarify || len(state.Diagnostics) != 0 {
		t.Fatalf("valid clarify SPEC stage = %#v, want clarify without diagnostics", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "ready"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageReady || len(state.Diagnostics) != 0 {
		t.Fatalf("ready SPEC stage = %#v, want ready", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "implement"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageImplement || len(state.Diagnostics) != 0 {
		t.Fatalf("implement SPEC stage = %#v, want implement", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "validate"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageValidate || len(state.Diagnostics) != 0 {
		t.Fatalf("validate SPEC stage = %#v, want validate", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "reflect"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageReflect || len(state.Diagnostics) != 0 {
		t.Fatalf("reflect SPEC stage = %#v, want reflect", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "deliver"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageDeliver || len(state.Diagnostics) != 0 {
		t.Fatalf("deliver SPEC stage = %#v, want deliver", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "complete"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageComplete || len(state.Diagnostics) != 0 {
		t.Fatalf("complete SPEC stage = %#v, want complete", state)
	}

	writeFile(t, filepath.Join(featurePath, "SPEC.md"), validV2SpecWithPhase("0001-alpha", "blocked"))
	state = resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageBlocked || len(state.Diagnostics) != 0 {
		t.Fatalf("blocked SPEC stage = %#v, want blocked", state)
	}
}

func TestResolveStrictLoopStageRejectsReadyBeforeClarificationReady(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll(feature) error = %v", err)
	}
	feat := &feature.Feature{Slug: "alpha", DirName: "0001-alpha", Path: featurePath}
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), v2SpecWithClarification("0001-alpha", "ready", "open", 80, 2))

	state := resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageClarify {
		t.Fatalf("stage = %#v, want clarify diagnostics", state)
	}
	diagnostics := strings.Join(state.Diagnostics, "\n")
	for _, check := range []string{
		"clarification.status",
		"clarification.confidence",
		"clarification.unresolved_questions",
	} {
		if !strings.Contains(diagnostics, check) {
			t.Fatalf("expected diagnostics to contain %q, got %#v", check, state.Diagnostics)
		}
	}
}

func TestResolveStrictLoopStageRejectsUnmappedAcceptanceCriteria(t *testing.T) {
	projectRoot := t.TempDir()
	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("MkdirAll(feature) error = %v", err)
	}
	feat := &feature.Feature{Slug: "alpha", DirName: "0001-alpha", Path: featurePath}
	spec := strings.Replace(validV2SpecWithPhase("0001-alpha", "ready"), "- AC-001 -> go test ./...", "- AC-999 -> go test ./...", 1)
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), spec)

	state := resolveStrictLoopStage(projectRoot, feat)
	if state.Stage != loopStageClarify {
		t.Fatalf("stage = %#v, want clarify diagnostics", state)
	}
	if diagnostics := strings.Join(state.Diagnostics, "\n"); !strings.Contains(diagnostics, "AC-001") {
		t.Fatalf("expected missing AC-001 validation mapping diagnostic, got %#v", state.Diagnostics)
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
	writeFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), validV2SpecWithPhase("0001-alpha", "clarify"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "SPEC.md"), validV2SpecWithPhase("0002-beta", "complete"))
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
	writeFile(t, filepath.Join(specsDir, "0001-alpha", "SPEC.md"), validV2SpecWithPhase("0001-alpha", "clarify"))
	writeFile(t, filepath.Join(specsDir, "0002-beta", "SPEC.md"), validV2SpecWithPhase("0002-beta", "complete"))

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

	stageContracts := map[loopStage]string{
		loopStageClarify:   "Research repository-discoverable facts",
		loopStageReady:     "Audit readiness",
		loopStageImplement: "Execute the in-scope task checklist",
		loopStageValidate:  "Run the validation map",
		loopStageReflect:   "Review the integrated diff",
		loopStageDeliver:   "Read the repo-local delivery rules",
	}
	for stage, contract := range stageContracts {
		prompt := readFile(t, filepath.Join(projectRoot, ".kit", "captured-prompts", string(stage)+".md"))
		for _, required := range []string{
			"Advance feature `alpha` through the `" + string(stage) + "` phase only.",
			"## Phase Contract",
			contract,
			"## Kit Loop Contract",
			"KIT_LOOP_RESULT:",
		} {
			if !strings.Contains(prompt, required) {
				t.Fatalf("%s prompt missing %q:\n%s", stage, required, prompt)
			}
		}
		if strings.Contains(prompt, "## Phase Outcomes") {
			t.Fatalf("%s prompt reinjected the full lifecycle contract:\n%s", stage, prompt)
		}
		for otherStage, otherContract := range stageContracts {
			if otherStage != stage && strings.Contains(prompt, otherContract) {
				t.Fatalf("%s prompt contains unrelated %s contract %q", stage, otherStage, otherContract)
			}
		}
		if stage != loopStageClarify && strings.Contains(prompt, "Open Questions") {
			t.Fatalf("%s prompt contains clarify-only questions contract", stage)
		}
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
  clarify)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "ready") + `DOC
    ;;
  ready)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "implement") + `DOC
    ;;
  implement)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "validate") + `DOC
    ;;
  validate)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "reflect") + `DOC
    ;;
  reflect)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "deliver") + `DOC
    ;;
  deliver)
    cat > docs/specs/0001-alpha/SPEC.md <<'DOC'
` + validV2SpecWithPhase("0001-alpha", "complete") + `DOC
    ;;
esac
`
	}
	return `#!/bin/sh
set -eu
mkdir -p .kit/captured-prompts
cat > ".kit/captured-prompts/$KIT_LOOP_STAGE.md"
` + mutations + result + `
`
}

func validV2SpecWithPhase(dirName string, phase string) string {
	status := "open"
	confidence := 0
	unresolved := 1
	switch phase {
	case "ready", "implement", "validate", "reflect", "deliver", "complete":
		status = "ready"
		confidence = 95
		unresolved = 0
	case "blocked":
		status = "blocked"
	}
	return v2SpecWithClarification(dirName, phase, status, confidence, unresolved)
}

func v2SpecWithClarification(dirName string, phase string, status string, confidence int, unresolved int) string {
	id, slug, ok := strings.Cut(dirName, "-")
	if !ok {
		id = ""
		slug = dirName
	}
	return `---
kit_metadata_version: 1
artifact: spec
workflow_version: 2
phase: ` + phase + `
clarification:
  status: ` + status + `
  confidence: ` + fmtInt(confidence) + `
  unresolved_questions: ` + fmtInt(unresolved) + `
feature:
  id: "` + id + `"
  slug: ` + slug + `
  dir: ` + dirName + `
---
# SPEC

## THESIS

Thesis for ` + slug + `.

## CONTEXT

Repo-grounded context.

## CLARIFICATIONS

No unresolved clarification questions.

## REQUIREMENTS

- Requirement one.

## ASSUMPTIONS

No blocking assumptions.

## ACCEPTANCE CRITERIA

- AC-001: Binary-verifiable criterion.

## IMPLEMENTATION PLAN

Implement the planned change.

## TASK CHECKLIST

- [x] T001: Maintain v2 workflow state.

## VALIDATION MAP

- AC-001 -> go test ./...

## REFLECTION NOTES

No remaining risks.

## DOCUMENTATION UPDATES

README and command docs are current.

## DELIVERY DECISION

No delivery mutation requested.

## EVIDENCE

Validation evidence recorded.
`
}

func fmtInt(value int) string {
	return strconv.Itoa(value)
}
