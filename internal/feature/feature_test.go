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
			name: "v2 spec phase clarify",
			files: map[string]string{
				"SPEC.md": validV2FeatureSpec("0001-alpha", "clarify"),
			},
			wantPhase: PhaseClarify,
		},
		{
			name: "v2 spec phase deliver overrides legacy tasks",
			files: map[string]string{
				"SPEC.md":  validV2FeatureSpec("0001-alpha", "deliver"),
				"PLAN.md":  "# PLAN\n",
				"TASKS.md": "# TASKS\n\n- [x] T001: done\n" + ReflectionCompleteMarker + "\n",
			},
			wantPhase: PhaseDeliver,
		},
		{
			name: "v3 spec phase validate",
			files: map[string]string{
				"SPEC.md": validV3FeatureSpec("0001-alpha", "validate"),
			},
			wantPhase: PhaseValidate,
		},
		{
			name: "v2 spec without phase defaults to clarify",
			files: map[string]string{
				"SPEC.md": validV2FeatureSpec("0001-alpha", ""),
			},
			wantPhase: PhaseClarify,
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

func validV3FeatureSpec(dirName string, phase string) string {
	return "---\n" +
		"kit_metadata_version: 1\n" +
		"artifact: spec\n" +
		"workflow_version: 3\n" +
		"phase: " + phase + "\n" +
		"feature:\n" +
		"  dir: " + dirName + "\n" +
		"---\n" +
		"# SPEC\n\n" +
		"## PURPOSE\n\npurpose\n\n" +
		"## CONTEXT\n\ncontext\n\n" +
		"## REQUIREMENTS\n\nrequirements\n\n" +
		"## ACCEPTED PLAN\n\nplan\n\n" +
		"## DECISIONS\n\nnone\n\n" +
		"## DISCOVERIES\n\nnone\n\n" +
		"## VALIDATION\n\npending\n\n" +
		"## OUTCOME\n\npending\n\n" +
		"## REPOSITORY MEMORY\n\npending\n"
}

func validV2FeatureSpec(dirName string, phase string) string {
	phaseLine := ""
	if phase != "" {
		phaseLine = "phase: " + phase + "\n"
	}
	return "---\n" +
		"kit_metadata_version: 1\n" +
		"artifact: spec\n" +
		"workflow_version: 2\n" +
		phaseLine +
		"feature:\n" +
		"  dir: " + dirName + "\n" +
		"---\n" +
		"# SPEC\n\n" +
		"## THESIS\n\nthesis\n\n" +
		"## CONTEXT\n\ncontext\n\n" +
		"## CLARIFICATIONS\n\nnone\n\n" +
		"## REQUIREMENTS\n\nrequirements\n\n" +
		"## ASSUMPTIONS\n\nnone\n\n" +
		"## ACCEPTANCE CRITERIA\n\ncriteria\n\n" +
		"## IMPLEMENTATION PLAN\n\nplan\n\n" +
		"## TASK CHECKLIST\n\n- [ ] task\n\n" +
		"## VALIDATION MAP\n\nmap\n\n" +
		"## REFLECTION NOTES\n\nnotes\n\n" +
		"## DOCUMENTATION UPDATES\n\nupdates\n\n" +
		"## DELIVERY DECISION\n\ndecision\n\n" +
		"## EVIDENCE\n\nevidence\n"
}
