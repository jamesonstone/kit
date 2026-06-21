package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func parseLoopStage(value string) (loopStage, error) {
	switch loopStage(strings.ToLower(strings.TrimSpace(value))) {
	case "spec", loopStageClarify:
		return loopStageClarify, nil
	case loopStageReady:
		return loopStageReady, nil
	case loopStageImplement:
		return loopStageImplement, nil
	case loopStageValidate:
		return loopStageValidate, nil
	case loopStageReflect:
		return loopStageReflect, nil
	case loopStageDeliver:
		return loopStageDeliver, nil
	case loopStageComplete:
		return loopStageComplete, nil
	default:
		return "", fmt.Errorf("invalid --until stage %q", value)
	}
}

func loopTargetComplete(current, until loopStage) bool {
	return loopStageRank(current) > loopStageRank(until)
}

func loopStageRank(stage loopStage) int {
	switch stage {
	case loopStageClarify:
		return 1
	case loopStageReady:
		return 2
	case loopStageImplement:
		return 3
	case loopStageValidate:
		return 4
	case loopStageReflect:
		return 5
	case loopStageDeliver:
		return 6
	case loopStageComplete:
		return 7
	case loopStageBlocked:
		return -1
	default:
		return 0
	}
}

func effectiveLoopMinConfidence(cfg *config.Config, override int) int {
	if override > 0 {
		return clampPercentage(override)
	}
	if cfg != nil && cfg.Loop.MinConfidence > 0 {
		return clampPercentage(cfg.Loop.MinConfidence)
	}
	if cfg != nil && cfg.GoalPercentage > 0 {
		return clampPercentage(cfg.GoalPercentage)
	}
	return 95
}

func effectiveLoopMaxIterations(cfg *config.Config, override int) int {
	if override > 0 {
		return override
	}
	if cfg != nil && cfg.Loop.MaxIterations > 0 {
		return cfg.Loop.MaxIterations
	}
	return 10
}

func clampPercentage(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
