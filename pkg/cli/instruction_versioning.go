package cli

import (
	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/instructions"
)

const instructionScaffoldVersionUnknown = instructions.UnknownVersion

type instructionRemovalPlan struct {
	relativePath string
	absolutePath string
}

func detectInstructionScaffoldVersion(projectRoot string, cfg *config.Config) int {
	return instructions.DetectVersion(projectRoot, cfg)
}
