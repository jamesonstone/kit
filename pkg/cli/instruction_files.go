package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

const (
	agentsMDPath            = instructions.AgentsMDPath
	claudeMDPath            = instructions.ClaudeMDPath
	copilotInstructionsPath = instructions.CopilotInstructionsPath
)

type instructionFileWriteResult string

type instructionFileWriteMode string

type instructionFileSelection struct {
	agentsMD bool
	claude   bool
	copilot  bool
}

const (
	instructionFileCreated instructionFileWriteResult = "created"
	instructionFileUpdated instructionFileWriteResult = "updated"
	instructionFileMerged  instructionFileWriteResult = "merged"
	instructionFileSkipped instructionFileWriteResult = "skipped"

	instructionFileWriteModeSkipExisting instructionFileWriteMode = "skip-existing"
	instructionFileWriteModeOverwrite    instructionFileWriteMode = "overwrite"
	instructionFileWriteModeAppendOnly   instructionFileWriteMode = "append-only"
)

type instructionFileWritePlan struct {
	relativePath string
	absolutePath string
	content      string
	result       instructionFileWriteResult
}

func (s instructionFileSelection) any() bool {
	return s.agentsMD || s.claude || s.copilot
}

func instructionFiles(cfg *config.Config) []string {
	return instructions.InstructionRelativePaths(cfg)
}

func selectedInstructionFiles(cfg *config.Config, selection instructionFileSelection) []string {
	if !selection.any() {
		return instructionFiles(cfg)
	}

	files := make([]string, 0, 3)
	if selection.agentsMD {
		files = appendInstructionFile(files, agentsMDPath)
	}
	if selection.claude {
		files = appendInstructionFile(files, claudeMDPath)
	}
	if selection.copilot {
		files = appendInstructionFile(files, copilotInstructionsPath)
	}

	return files
}

func appendInstructionFile(files []string, path string) []string {
	for _, existing := range files {
		if existing == path {
			return files
		}
	}

	return append(files, path)
}

func writeInstructionFile(projectRoot, relativePath string, overwrite bool) (instructionFileWriteResult, error) {
	mode := instructionFileWriteModeSkipExisting
	if overwrite {
		mode = instructionFileWriteModeOverwrite
	}

	return writeInstructionFileWithMode(projectRoot, relativePath, mode, config.DefaultInstructionScaffoldVersion)
}

func writeInstructionFileWithMode(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	version int,
) (instructionFileWriteResult, error) {
	plan, err := planInstructionFileWrite(projectRoot, relativePath, mode, version)
	if err != nil {
		return "", err
	}

	return applyInstructionFileWritePlan(plan)
}

func determineInstructionFileWriteMode(force, appendOnly bool) (instructionFileWriteMode, error) {
	if force && appendOnly {
		return "", fmt.Errorf("--append-only cannot be used with --force")
	}

	if appendOnly {
		return instructionFileWriteModeAppendOnly, nil
	}

	if force {
		return instructionFileWriteModeOverwrite, nil
	}

	return instructionFileWriteModeSkipExisting, nil
}

func existingInstructionFiles(projectRoot string, relativePaths []string) []string {
	var existing []string
	for _, relativePath := range relativePaths {
		if document.Exists(filepath.Join(projectRoot, relativePath)) {
			existing = append(existing, relativePath)
		}
	}

	return existing
}

func planInstructionFileWrites(projectRoot string, relativePaths []string, mode instructionFileWriteMode) ([]instructionFileWritePlan, error) {
	return planInstructionArtifactWrites(projectRoot, relativePaths, mode, config.DefaultInstructionScaffoldVersion)
}

func planInstructionArtifactWrites(
	projectRoot string,
	relativePaths []string,
	mode instructionFileWriteMode,
	version int,
) ([]instructionFileWritePlan, error) {
	plans := make([]instructionFileWritePlan, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		plan, err := planInstructionArtifactWrite(projectRoot, relativePath, mode, version)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func planInstructionFileWrite(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	version int,
) (instructionFileWritePlan, error) {
	return planInstructionArtifactWrite(projectRoot, relativePath, mode, version)
}
