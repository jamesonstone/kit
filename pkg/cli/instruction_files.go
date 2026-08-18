package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/instructions"
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

func planInstructionFileWrite(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	version int,
) (instructionFileWritePlan, error) {
	return planInstructionArtifactWrite(projectRoot, relativePath, mode, version)
}

func planInstructionArtifactWrite(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	version int,
) (instructionFileWritePlan, error) {
	absolutePath := filepath.Join(projectRoot, relativePath)
	existed := document.Exists(absolutePath)
	content, _, err := instructionArtifactContent(relativePath, version)
	if err != nil {
		return instructionFileWritePlan{}, err
	}

	switch mode {
	case instructionFileWriteModeSkipExisting:
		if existed {
			return instructionFileWritePlan{
				relativePath: relativePath,
				absolutePath: absolutePath,
				result:       instructionFileSkipped,
			}, nil
		}
		return instructionFileWritePlan{
			relativePath: relativePath,
			absolutePath: absolutePath,
			content:      content,
			result:       instructionFileCreated,
		}, nil
	case instructionFileWriteModeOverwrite:
		result := instructionFileCreated
		if existed {
			existingContent, err := readInstructionFile(absolutePath)
			if err != nil {
				return instructionFileWritePlan{}, fmt.Errorf("failed to read %s: %w", relativePath, err)
			}
			if existingContent == content {
				return instructionFileWritePlan{
					relativePath: relativePath,
					absolutePath: absolutePath,
					result:       instructionFileSkipped,
				}, nil
			}
			result = instructionFileUpdated
		}
		return instructionFileWritePlan{
			relativePath: relativePath,
			absolutePath: absolutePath,
			content:      content,
			result:       result,
		}, nil
	case instructionFileWriteModeAppendOnly:
		if !existed {
			return instructionFileWritePlan{
				relativePath: relativePath,
				absolutePath: absolutePath,
				content:      content,
				result:       instructionFileCreated,
			}, nil
		}

		existingContent, err := readInstructionFile(absolutePath)
		if err != nil {
			return instructionFileWritePlan{}, fmt.Errorf("failed to read %s: %w", relativePath, err)
		}

		mergedContent, changed, err := mergeInstructionFileContent(existingContent, content)
		if err != nil {
			return instructionFileWritePlan{}, fmt.Errorf(
				"append-only merge failed for %s: %w. Use --force to overwrite or edit the file manually to add Kit section headings",
				relativePath,
				err,
			)
		}

		if !changed {
			return instructionFileWritePlan{
				relativePath: relativePath,
				absolutePath: absolutePath,
				result:       instructionFileSkipped,
			}, nil
		}

		return instructionFileWritePlan{
			relativePath: relativePath,
			absolutePath: absolutePath,
			content:      mergedContent,
			result:       instructionFileMerged,
		}, nil
	default:
		return instructionFileWritePlan{}, fmt.Errorf("unsupported instruction file write mode %q", mode)
	}
}

func instructionArtifactPaths(
	cfg *config.Config,
	selection instructionFileSelection,
	version int,
	forceFullModel bool,
) []string {
	relativePaths := selectedInstructionFiles(cfg, selection)
	if forceFullModel {
		relativePaths = instructionFiles(cfg)
	}

	if !config.UsesInstructionSupportDocs(version) {
		return relativePaths
	}

	for _, support := range instructions.SupportDocs(version) {
		relativePaths = appendInstructionFile(relativePaths, support.RelativePath)
	}

	return relativePaths
}
