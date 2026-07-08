package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
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
