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

type instructionFileSelection struct {
	agentsMD bool
	claude   bool
	copilot  bool
}

const (
	instructionFileCreated instructionFileWriteResult = "created"
	instructionFileUpdated instructionFileWriteResult = "updated"
	instructionFileSkipped instructionFileWriteResult = "skipped"
)

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
	absolutePath := filepath.Join(projectRoot, relativePath)
	if document.Exists(absolutePath) && !overwrite {
		return instructionFileSkipped, nil
	}

	existed := document.Exists(absolutePath)
	content := templates.InstructionFile(relativePath)
	if err := document.Write(absolutePath, content); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", relativePath, err)
	}

	if existed {
		return instructionFileUpdated, nil
	}

	return instructionFileCreated, nil
}
