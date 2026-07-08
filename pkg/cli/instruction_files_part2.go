package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
	"github.com/jamesonstone/kit/internal/templates"
)

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

	if version != config.InstructionScaffoldVersionTOC {
		return relativePaths
	}

	for _, support := range instructions.SupportDocs(version) {
		relativePaths = appendInstructionFile(relativePaths, support.RelativePath)
	}

	return relativePaths
}

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
