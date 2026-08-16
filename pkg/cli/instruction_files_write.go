package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func instructionArtifactContent(relativePath string, version int) (string, bool, error) {
	for _, support := range templates.InstructionSupportFiles(version) {
		if support.RelativePath == relativePath {
			return support.Content, true, nil
		}
	}

	if !config.IsInstructionScaffoldVersionSupported(version) {
		return "", false, fmt.Errorf("unsupported instruction scaffold version %d", version)
	}

	return templates.InstructionFileForVersion(relativePath, version), false, nil
}

func applyInstructionFileWritePlan(plan instructionFileWritePlan) (instructionFileWriteResult, error) {
	if plan.result == instructionFileSkipped {
		return instructionFileSkipped, nil
	}

	if err := document.Write(plan.absolutePath, plan.content); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", plan.relativePath, err)
	}

	return plan.result, nil
}
