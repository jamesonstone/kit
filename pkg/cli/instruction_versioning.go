package cli

import (
	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/instructions"
)

const instructionScaffoldVersionUnknown = instructions.UnknownVersion

type instructionRemovalPlan struct {
	relativePath string
	absolutePath string
}

func detectInstructionScaffoldVersion(projectRoot string, cfg *config.Config) int {
	return instructions.DetectVersion(projectRoot, cfg)
}
