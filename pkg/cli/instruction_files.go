package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

const (
	agentsMDPath            = "AGENTS.md"
	claudeMDPath            = "CLAUDE.md"
	copilotInstructionsPath = ".github/copilot-instructions.md"
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
	files := make([]string, 0, len(cfg.Agents)+1)
	for _, file := range cfg.Agents {
		files = appendInstructionFile(files, file)
	}
	files = appendInstructionFile(files, copilotInstructionsPath)
	return files
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

	return writeInstructionFileWithMode(projectRoot, relativePath, mode)
}

func writeInstructionFileWithMode(projectRoot, relativePath string, mode instructionFileWriteMode) (instructionFileWriteResult, error) {
	plan, err := planInstructionFileWrite(projectRoot, relativePath, mode)
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
	plans := make([]instructionFileWritePlan, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		plan, err := planInstructionFileWrite(projectRoot, relativePath, mode)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func planInstructionFileWrite(projectRoot, relativePath string, mode instructionFileWriteMode) (instructionFileWritePlan, error) {
	absolutePath := filepath.Join(projectRoot, relativePath)
	existed := document.Exists(absolutePath)
	content := templates.InstructionFile(relativePath)

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

func applyInstructionFileWritePlan(plan instructionFileWritePlan) (instructionFileWriteResult, error) {
	if plan.result == instructionFileSkipped {
		return instructionFileSkipped, nil
	}

	if err := document.Write(plan.absolutePath, plan.content); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", plan.relativePath, err)
	}

	return plan.result, nil
}
