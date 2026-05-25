package cli

import (
	"os"
	"path/filepath"
	stdreflect "reflect"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestWorkflowStageCandidatesUseCurrentLifecyclePhase(t *testing.T) {
	specsDir := filepath.Join(t.TempDir(), "docs", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.MkdirAll(filepath.Join(specsDir, "0001-empty-feature"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	createFeatureFile(t, specsDir, "0002-brainstorm-only", "BRAINSTORM.md", "# BRAINSTORM\n")
	createFeatureFile(t, specsDir, "0003-spec-only", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0004-plan-only", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0004-plan-only", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0005-empty-tasks", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0005-empty-tasks", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0005-empty-tasks", "TASKS.md", "# TASKS\n\nno checkboxes yet\n")
	createFeatureFile(t, specsDir, "0006-implementation-open", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0006-implementation-open", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0006-implementation-open", "TASKS.md", "# TASKS\n\n- [x] T001: done\n- [ ] T002: remaining\n")
	createFeatureFile(t, specsDir, "0007-reflect-ready", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0007-reflect-ready", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0007-reflect-ready", "TASKS.md", "# TASKS\n\n- [x] T001: done\n")
	createFeatureFile(t, specsDir, "0008-complete", "SPEC.md", "# SPEC\n")
	createFeatureFile(t, specsDir, "0008-complete", "PLAN.md", "# PLAN\n")
	createFeatureFile(t, specsDir, "0008-complete", "TASKS.md", "# TASKS\n\n- [x] T001: done\n\n"+feature.ReflectionCompleteMarker+"\n")

	tests := []struct {
		name  string
		stage workflowSelectionStage
		want  []string
	}{
		{
			name:  "spec includes pre-spec features with or without brainstorm",
			stage: workflowSelectionStageSpec,
			want:  []string{"0001-empty-feature", "0002-brainstorm-only"},
		},
		{
			name:  "plan hides completed plans and later phases",
			stage: workflowSelectionStagePlan,
			want:  []string{"0003-spec-only"},
		},
		{
			name:  "tasks hides completed task docs and later phases",
			stage: workflowSelectionStageTasks,
			want:  []string{"0004-plan-only"},
		},
		{
			name:  "implement hides empty tasks, completed implementation, and reflection",
			stage: workflowSelectionStageImplement,
			want:  []string{"0006-implementation-open"},
		},
		{
			name:  "reflect only shows implementation-complete features without reflection marker",
			stage: workflowSelectionStageReflect,
			want:  []string{"0007-reflect-ready"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates, err := workflowStageCandidates(specsDir, tt.stage)
			if err != nil {
				t.Fatalf("workflowStageCandidates() error = %v", err)
			}

			var got []string
			for _, candidate := range candidates {
				got = append(got, candidate.DirName)
			}
			if !stdreflect.DeepEqual(got, tt.want) {
				t.Fatalf("workflowStageCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}
