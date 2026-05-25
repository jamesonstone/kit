package cli

import (
	"path/filepath"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

type workflowSelectionStage string

const (
	workflowSelectionStageSpec      workflowSelectionStage = "spec"
	workflowSelectionStagePlan      workflowSelectionStage = "plan"
	workflowSelectionStageTasks     workflowSelectionStage = "tasks"
	workflowSelectionStageImplement workflowSelectionStage = "implement"
	workflowSelectionStageReflect   workflowSelectionStage = "reflect"
)

func workflowStageCandidates(specsDir string, stage workflowSelectionStage) ([]feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if featureReadyForWorkflowStage(f, stage) {
			candidates = append(candidates, f)
		}
	}

	return candidates, nil
}

func featureReadyForWorkflowStage(f feature.Feature, stage workflowSelectionStage) bool {
	specPath := filepath.Join(f.Path, "SPEC.md")
	planPath := filepath.Join(f.Path, "PLAN.md")
	tasksPath := filepath.Join(f.Path, "TASKS.md")

	hasSpec := document.Exists(specPath)
	hasPlan := document.Exists(planPath)
	hasTasks := document.Exists(tasksPath)

	switch stage {
	case workflowSelectionStageSpec:
		return !hasSpec && f.Phase == feature.PhaseBrainstorm
	case workflowSelectionStagePlan:
		return hasSpec && !hasPlan && f.Phase == feature.PhaseSpec
	case workflowSelectionStageTasks:
		return hasSpec && hasPlan && !hasTasks && f.Phase == feature.PhasePlan
	case workflowSelectionStageImplement:
		return hasSpec && hasPlan && hasTasks && f.Phase == feature.PhaseImplement
	case workflowSelectionStageReflect:
		return hasSpec && hasPlan && hasTasks && f.Phase == feature.PhaseReflect
	default:
		return false
	}
}
