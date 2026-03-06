package feature

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeterminePhase(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantPhase Phase
	}{
		{
			name:      "empty feature defaults to brainstorm",
			files:     map[string]string{},
			wantPhase: PhaseBrainstorm,
		},
		{
			name: "brainstorm only",
			files: map[string]string{
				"BRAINSTORM.md": "# BRAINSTORM\n",
			},
			wantPhase: PhaseBrainstorm,
		},
		{
			name: "spec only",
			files: map[string]string{
				"SPEC.md": "# SPEC\n",
			},
			wantPhase: PhaseSpec,
		},
		{
			name: "plan present",
			files: map[string]string{
				"SPEC.md": "# SPEC\n",
				"PLAN.md": "# PLAN\n",
			},
			wantPhase: PhasePlan,
		},
		{
			name: "tasks without checkboxes stays tasks",
			files: map[string]string{
				"SPEC.md":  "# SPEC\n",
				"PLAN.md":  "# PLAN\n",
				"TASKS.md": "# TASKS\n\nno checkboxes yet\n",
			},
			wantPhase: PhaseTasks,
		},
		{
			name: "incomplete tasks move to implement",
			files: map[string]string{
				"SPEC.md":  "# SPEC\n",
				"PLAN.md":  "# PLAN\n",
				"TASKS.md": "# TASKS\n\n- [x] T001: done\n- [ ] T002: remaining\n",
			},
			wantPhase: PhaseImplement,
		},
		{
			name: "complete tasks move to reflect",
			files: map[string]string{
				"SPEC.md":  "# SPEC\n",
				"PLAN.md":  "# PLAN\n",
				"TASKS.md": "# TASKS\n\n- [x] T001: done\n- [x] T002: done\n",
			},
			wantPhase: PhaseReflect,
		},
		{
			name: "reflection marker completes feature",
			files: map[string]string{
				"SPEC.md":  "# SPEC\n",
				"PLAN.md":  "# PLAN\n",
				"TASKS.md": "# TASKS\n\n- [x] T001: done\n" + ReflectionCompleteMarker + "\n",
			},
			wantPhase: PhaseComplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			featureDir := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(featureDir, name)
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}

			got := DeterminePhase(featureDir)
			if got != tt.wantPhase {
				t.Fatalf("DeterminePhase() = %q, want %q", got, tt.wantPhase)
			}
		})
	}
}
