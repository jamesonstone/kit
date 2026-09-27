package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/instructions"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const (
	agentsMDPath            = instructions.AgentsMDPath
	claudeMDPath            = instructions.ClaudeMDPath
	copilotInstructionsPath = instructions.CopilotInstructionsPath
)

type instructionFileWriteResult string

type instructionFileWriteMode string

const (
	instructionFileCreated instructionFileWriteResult = "created"
	instructionFileUpdated instructionFileWriteResult = "updated"
	instructionFileMerged  instructionFileWriteResult = "merged"
	instructionFileSkipped instructionFileWriteResult = "skipped"
	// instructionFileRemoved deletes a retired Kit-owned file that Git can restore.
	instructionFileRemoved instructionFileWriteResult = "removed"

	// instructionFileWriteModeSkipExisting only creates missing files.
	instructionFileWriteModeSkipExisting instructionFileWriteMode = "skip-existing"
	// instructionFileWriteModeConverge migrates existing entry files onto the
	// managed contract block and leaves the project-owned testing reference alone.
	instructionFileWriteModeConverge instructionFileWriteMode = "converge"
)

type instructionFileWritePlan struct {
	relativePath string
	absolutePath string
	content      string
	result       instructionFileWriteResult
	// converged is false when an entry file was left without a current block.
	converged bool
	note      string
}

func instructionFiles(cfg *config.Config) []string {
	return instructions.InstructionRelativePaths(cfg)
}

// instructionArtifactPaths lists every instruction artifact Kit generates: the
// agent entry files and the testing reference.
func instructionArtifactPaths(cfg *config.Config) []string {
	return append(instructionFiles(cfg), templates.TestingReferencePath)
}

func instructionArtifactContent(relativePath string) string {
	if filepath.ToSlash(relativePath) == templates.TestingReferencePath {
		return templates.TestingReference
	}
	return templates.InstructionEntryFile(relativePath)
}

func writeInstructionFileWithMode(projectRoot, relativePath string, mode instructionFileWriteMode) (instructionFileWriteResult, error) {
	plan, err := planInstructionArtifactWrite(projectRoot, relativePath, mode, false)
	if err != nil {
		return "", err
	}
	if plan.result == instructionFileSkipped {
		return instructionFileSkipped, nil
	}
	if err := document.Write(plan.absolutePath, plan.content); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", plan.relativePath, err)
	}
	return plan.result, nil
}

func planInstructionArtifactWrite(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	force bool,
) (instructionFileWritePlan, error) {
	plan := instructionFileWritePlan{
		relativePath: relativePath,
		absolutePath: filepath.Join(projectRoot, filepath.FromSlash(relativePath)),
		converged:    true,
	}
	if !document.Exists(plan.absolutePath) {
		plan.content = instructionArtifactContent(relativePath)
		plan.result = instructionFileCreated
		return plan, nil
	}
	plan.result = instructionFileSkipped
	if mode == instructionFileWriteModeSkipExisting || filepath.ToSlash(relativePath) == templates.TestingReferencePath {
		return plan, nil
	}
	data, err := os.ReadFile(plan.absolutePath)
	if err != nil {
		return instructionFileWritePlan{}, fmt.Errorf("failed to read %s: %w", relativePath, err)
	}
	migration := migrateEntryFile(filepath.ToSlash(relativePath), string(data), force)
	plan.converged = migration.converged
	plan.note = migration.note
	if migration.content != string(data) {
		plan.content = migration.content
		plan.result = instructionFileUpdated
	}
	return plan, nil
}
